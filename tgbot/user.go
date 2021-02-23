package tgbot

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/madnaaaaas/listengine"
	"strconv"
	"sync"
)

type User struct {
	states []*State
	username string
	lastMsgID int
	bot *tgbotapi.BotAPI

	lock sync.Mutex
}

func NewUserTelegramBot(username string, sl *listengine.SourceList, bot *tgbotapi.BotAPI) (*User, error) {
	l := listengine.NewList(sl)
	if err := l.ReadUser(username); err != nil {
		return nil, err
	}
	user := &User {
		username: username,
		states: []*State{{l:l,t: telegramTypeNew}},
		bot: bot,
	}
	return user, nil
}

func (user *User) getLast() *State {
	if user == nil || user.states == nil {
		return nil
	}

	return user.states[len(user.states) - 1]
}

func (user *User) send(keyboard tgbotapi.InlineKeyboardMarkup, text string, chatID int64) {
	if user.lastMsgID == 0 {
		msgC := tgbotapi.NewMessage(chatID, text)
		msgC.ReplyMarkup = keyboard
		if msg, err := user.bot.Send(msgC); err != nil {
			fmt.Println(err)
		} else {
			user.lastMsgID = msg.MessageID
		}
	} else {
		edt := tgbotapi.NewEditMessageText(chatID, user.lastMsgID, text)
		edk := tgbotapi.NewEditMessageReplyMarkup(chatID, user.lastMsgID, keyboard)
		if _, err := user.bot.Send(edt); err != nil {
			fmt.Println(err)
		} else if _, err = user.bot.Send(edk); err != nil {
			fmt.Println(err)
		}
	}
}

func (user *User) back() {
	if len(user.states) > 1 {
		user.states = user.states[:len(user.states) - 1]
	}
}

func (user *User) addList(command string) {
	var l *listengine.List
	if command == "db" {
		l = listengine.NewFullList(user.getLast().l)
	} else {
		l, _ = user.states[0].l.SubList(command)
	}
	user.states = append(user.states[:1], &State{l:l, t: telegramTypeMenu})
}

func (user *User) view() {
	user.states = append(user.states, &State{l:user.getLast().l, t: telegramTypeView, num: 1})
}

func (user *User) edit() {
	user.states = append(user.states, &State{l:user.getLast().l, t: telegramTypeEdit, num: 1})
}

func (user *User) prevAndNext(command string) {
	st := user.getLast()
	if command == telegramCommandPrev {
		if st.num > 1 {
			st.num--
		}
	} else if command == telegramCommandNext {
		var max int
		if st.t == telegramTypeView || st.t == telegramTypeEdit {
			max = (st.l.Len() + 9)/ 10
		} else if st.t == telegramTypeMeta {
			max = st.l.Len()
		}
		if st.num < max {
			st.num++
		}
	}
}

func (user *User) meta(command string) {
	num, _ := strconv.Atoi(command)
	user.states = append(user.states, &State{l:user.getLast().l, t: telegramTypeMeta, num: num})
}

func (user *User) skip() {
	st := user.getLast()
	st.l.Skip(st.num)
	user.random()
}

func (user *User) search(s string) {
	st := user.getLast()
	st.l = st.l.Search(s)
	st.t = telegramTypeMenu
}

func (user *User) prepareSearch() {
	st := user.getLast()
	user.states = append(user.states, &State{l:st.l, t: telegramTypeSearch})
}

func (user *User) random() {
	num := user.getLast().l.Random()
	if num < 0 {
		return
	}
	st := user.getLast()
	if st.t == telegramTypeRandom {
		st.num = num
	} else {
		user.states = append(user.states, &State{l: st.l, t: telegramTypeRandom, num: num})
	}
}

func (user *User) mark(command string) {
	num, _ := strconv.Atoi(command)
	st := user.getLast()
	v := st.l.Check(num)
	st.l.Mark(num, !v)


	if err := st.l.WriteUser(); err != nil {
		fmt.Println(err)
	}
}

func (user *User) UpdateCallback(update tgbotapi.Update) {
	user.lock.Lock()
	defer user.lock.Unlock()

	st := user.getLast()
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		command := update.CallbackQuery.Data
		switch command {
		case telegramCommandBack:
			user.back()
		case telegramCommandView:
			user.view()
		case telegramCommandEdit:
			user.edit()
		case telegramCommandPrev, telegramCommandNext:
			user.prevAndNext(command)
		case telegramCommandRandom:
			user.random()
		case telegramCommandSkip:
			user.skip()
		case telegramCommandSearch:
			user.prepareSearch()
		default:
			if st.t == telegramTypeNew{
				user.addList(command)
			} else if st.t == telegramTypeView || st.t == telegramTypeRandom {
				user.meta(command)
			} else if st.t == telegramTypeMeta || st.t == telegramTypeEdit {
				user.mark(command)
			}
		}
	} else if update.Message != nil && update.Message.Text != "" {
		user.lastMsgID = 0

		chatID = update.Message.Chat.ID
		s := update.Message.Text
		if st.t == telegramTypeSearch {
			user.search(s)
			s = ""
		}

		switch s {
		case "/start":
			user.states = user.states[:1]
		}
	}
	st = user.getLast()
	if chatID != 0 {
		keyboard, text := st.msg()
		user.send(keyboard, text, chatID)
		fmt.Printf("%s: %s(%s):%d\n", user.username, st.t, st.l.Path(), st.num)
	}
}
