package listengine

import (
	"bytes"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"io"
	"strconv"
	"sync"
)

func (r Record) PrintMetaTelegramBot(w io.Writer, num int, viewed bool) {
	vs := "(-)"
	if viewed {
		vs = "(+)"
	}
	fmt.Fprintf(w, "%d. %s %s:\n", num, r.name, vs)
	for k, v := range r.m {
		if k == "kinopoisk_id" {
			k = "kinopoisk"
			v = "https://www.kinopoisk.ru/film/" + v + "/"
		}
		if k == "imdb_id" {
			k = "imdb"
			for len(v) < 7 {
				v = "0" + v
			}
			v = "https://www.imdb.com/title/tt" + v + "/"
		}
		fmt.Fprintf(w, "%s: %s\n", k, v)
	}
}

type State struct {
	l *List
	num int
	t string
}

func (st *State) msg() (tgbotapi.InlineKeyboardMarkup, string) {
	text := ""
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	var row []tgbotapi.InlineKeyboardButton
	switch st.t {
	case "new":
		text = "Выберите список фильмов"
		db := fmt.Sprintf("db (%d)", len(*st.l.sl))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(db, "db"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("wallfilm (150)", "wallfilm"))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case "menu":
		if st.l == nil {
			text = "Пустой список"
			break
		}
		text = fmt.Sprintf("Список: %s (Всего %d, Просмотренно %d)",
			st.l.path, len(st.l.list), st.l.vCount)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Просмотр", "/view"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Редактирование", "/edit"))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Поиск", "/search"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Случайный фильм", "/random"))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case "view":
		total := (len(st.l.list) + 9)/ 10
		text = fmt.Sprintf("Просмотр списка: %s (Страница %d из %d)",
			st.l.path, st.num, total)
		for i := 10 * (st.num - 1); i < 10 * st.num && i < len(st.l.list); i++ {
			name := fmt.Sprintf("%d.%s", i + 1, st.l.GetRecord(i).name)
			if st.l.Check(i) {
				name += " (+)"
			} else {
				name += " (-)"
			}
			data := fmt.Sprintf("%d", i)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name, data))
			keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Предыдущая страница", "/prev"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Следующая страница", "/next"))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case "edit":
		total := (len(st.l.list) + 9)/ 10
		text = fmt.Sprintf("Редактирование списка: %s (Страница %d из %d)",
			st.l.path, st.num, total)
		for i := 10 * (st.num - 1); i < 10 * st.num; i++ {
			name := fmt.Sprintf("%d.%s", i + 1, st.l.GetRecord(i).name)
			if st.l.Check(i) {
				name += " (Удалить из просмотренного)"
			} else {
				name += " (Добавить к просмотренному)"
			}
			data := fmt.Sprintf("%d", i)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name, data))
			keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Предыдущая страница", "/prev"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Следующая страница", "/next"))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case "meta":
		buf := new(bytes.Buffer)
		v := st.l.Check(st.num)
		st.l.GetRecord(st.num).PrintMetaTelegramBot(buf, st.num + 1, v)

		text = buf.String()
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Предыдущий фильм", "/prev"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Следующий фильм", "/next"))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
		s := ""
		if v {
			s = "Не смотрел"
		} else {
			s = "Смотрел"
		}
		data := fmt.Sprintf("%d", st.num)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(s, data))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case "random":
		text = "Случайный фильм:"
		r := st.l.GetRecord(st.num)
		name := fmt.Sprintf("%d.%s", st.num + 1, r.name)
		data := fmt.Sprintf("%d", st.num)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(name, data))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil

		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Пропустить", "/skip"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Другой случайный фильм", "/random"))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case "search":
		text = "Введите ключевые слова для поиска через пробел"
	}
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Назад", "/back"))
	keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil

	return keyboard, text
}

type UserTelegramBot struct {
	states []*State
	username string
	lastMsgID int
	bot *tgbotapi.BotAPI

	lock sync.Mutex
}

func NewUserTelegramBot(username string, sl *SourceList, bot *tgbotapi.BotAPI) (*UserTelegramBot, error) {
	l := NewList(sl)
	if err := l.ReadUser(username); err != nil {
		return nil, err
	}
	user := &UserTelegramBot {
		username: username,
		states: []*State{{l:l,t: "new"}},
		bot: bot,
	}
	return user, nil
}

func (user *UserTelegramBot) getLast() *State {
	if user == nil || user.states == nil {
		return nil
	}

	return user.states[len(user.states) - 1]
}

func (user *UserTelegramBot) send(keyboard tgbotapi.InlineKeyboardMarkup, text string, chatID int64) {
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
		}
		if _, err := user.bot.Send(edk); err != nil {
			fmt.Println(err)
		}
	}
}

func (user *UserTelegramBot) back() {
	if len(user.states) > 1 {
		user.states = user.states[:len(user.states) - 1]
	}
}

func (user *UserTelegramBot) addList(listName string) {
	var l *List
	if listName == "db" {
		l = NewFullList(user.getLast().l)
	} else {
		l, _ = user.states[0].l.SubList(listName)
	}
	user.states = append(user.states[:1], &State{l:l, t: "menu"})
}

func (user *UserTelegramBot) view() {
	user.states = append(user.states, &State{l:user.getLast().l, t: "view", num: 1})
}

func (user *UserTelegramBot) edit() {
	user.states = append(user.states, &State{l:user.getLast().l, t: "edit", num: 1})
}

func (user *UserTelegramBot) prevAndNext(command string) {
	st := user.getLast()
	if command == "/prev" {
		if st.num > 1 {
			st.num--
		}
	} else if command == "/next" {
		var max int
		if st.t == "view" || st.t == "edit" {
			max = (len(st.l.list) + 9)/ 10
		} else if st.t == "meta" {
			max = len(st.l.list)
		}
		if st.num < max {
			st.num++
		}
	}
}

func (user *UserTelegramBot) meta(command string) {
	num, _ := strconv.Atoi(command)
	user.states = append(user.states, &State{l:user.getLast().l, t: "meta", num: num})
}

func (user *UserTelegramBot) skip() {
	st := user.getLast()
	st.l.Skip(st.num)
}

func (user *UserTelegramBot) search(s string) {
	st := user.getLast()
	st.l = st.l.Search(s)
	st.t = "menu"
	user.lastMsgID = 0
}

func (user *UserTelegramBot) prepareSearch() {
	st := user.getLast()
	user.states = append(user.states, &State{l:st.l, t: "search"})
}

func (user *UserTelegramBot) random() {
	num := user.getLast().l.Random()
	if num < 0 {
		return
	}
	st := user.getLast()
	if st.t == "random" {
		st.num = num
	} else {
		user.states = append(user.states, &State{l: st.l, t: "random", num: num})
	}
}

func (user *UserTelegramBot) mark(command string) {
	num, _ := strconv.Atoi(command)
	st := user.getLast()
	v := st.l.Check(num)
	st.l.Mark(num, !v)


	if err := st.l.WriteUser(); err != nil {
		fmt.Println(err)
	}
}

func (user *UserTelegramBot) clean() {
	user.lastMsgID = 0
	user.states = user.states[:1]
}

func (user *UserTelegramBot) UpdateCallback(update tgbotapi.Update) {
	user.lock.Lock()
	defer user.lock.Unlock()

	st := user.getLast()
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		command := update.CallbackQuery.Data
		switch command {
		case "/back":
			user.back()
		case "/view":
			user.view()
		case "/edit":
			user.edit()
		case "/prev", "/next":
			user.prevAndNext(command)
		case "/random":
			user.random()
		case "/skip":
			user.skip()
		case "/search":
			user.prepareSearch()
		default:
			if st.t == "new" {
				user.addList(command)
			} else if st.t == "view" || st.t == "random" {
				user.meta(command)
			} else if st.t == "meta" || st.t == "edit" {
				user.mark(command)
			}
		}
	} else if update.Message != nil && update.Message.Text != "" {
		chatID = update.Message.Chat.ID
		s := update.Message.Text
		if st.t == "search" {
			user.search(s)
		}

		switch s {
		case "/start", "/stop":
			user.clean()
			if s == "/stop" {
				user.bot.Send(tgbotapi.NewMessage(chatID, "До скорых встреч!"))
				return
			}
		}
	}
	if chatID != 0 {
		keyboard, text := user.getLast().msg()
		user.send(keyboard, text, chatID)
	}
}

func userName(update tgbotapi.Update) string {
	if update.Message != nil {
		return update.Message.From.UserName
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.UserName
	}
	return ""
}

func TelegramBotGoroutines(sl *SourceList, token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return
	}

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return
	}

	users := make(map[string]*UserTelegramBot)

	for update := range updates {
		username := userName(update)
		if username == "" {
			fmt.Println("empty username")
			continue
		}
		user, ok := users[username]
		if !ok {
			var err error
			if user, err = NewUserTelegramBot(username, sl, bot); err != nil {
				continue
			}
			users[username] = user
		}
		go user.UpdateCallback(update)
	}
}