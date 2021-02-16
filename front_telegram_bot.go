package listengine

import (
	"bytes"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"io"
	"os"
	"strings"
	"sync"
)

const HELPTELEGRAM = `Commands:
	/help - help
	/random - get random film
	/skip - skip previous random film
	/clear - clear skip list
	/print - print current list
	/add <numbers with space separator> - add films
	/search <keywords with space separator> - search films by keywords in names
	/write - write list to user file
	/back - back to previous list
	/seen - show only viewed
	/unseen - show only unseen
	/meta - print info for film by number
`
func PrintHeaderTelegramBot(w io.Writer) {
	fmt.Fprintf(w, " V| Num| Name\n")
}

func (l *List) PrintRecordTelegramBot (id int, w io.Writer) {
	slId := l.list[id]
	v := ""
	if l.viewed[slId] {
		v = "+"
	} else {
		v = "-"
	}
	fmt.Fprintf(w, " %s|%d|%s\n", v, id + 1, (*l.sl)[slId].name)
}

func (r Record) PrintMetaTelegramBot(w io.Writer) {
	fmt.Fprintf(w, "%s:\n", r.name)
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

func (l *List) PrintTelegramBot(w io.Writer) {
	if l == nil || len(l.list) == 0 {
		fmt.Fprintf(w, "EMPTY LIST\n")
		return
	}
	fmt.Fprintf(w, "%s:%s (total %d):\n", l.username, l.path, len(l.list))
	PrintHeaderTelegramBot(w)
	for id, _ := range l.list {
		l.PrintRecordTelegramBot(id, w)
	}
	if l.skip != nil {
		l.SkipList().PrintTelegramBot(w)
	}
}

func TelegramBot(sl *SourceList, token string) {
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

	users := make(map[string][]*List)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		username := update.Message.From.UserName
		fmt.Printf("[%s] %s\n", username, update.Message.Text)
		var l *List
		user, ok := users[username]
		if ok {
			l = user[len(user) - 1]
		} else {
			user = make([]*List, 1)
			l = NewList(sl)
			l.ReadUser(username)
			l, _ = l.SubList("wallfilm")
			user[0] = l
			users[username] = user
		}
		s := update.Message.Text
		buf := new(bytes.Buffer)

		switch {
		case s == "/random":
			id := l.Random()
			if id >= 0 && id < len(l.list) {
				PrintHeaderTelegramBot(buf)
				l.PrintRecordTelegramBot(id, buf)
				l.lastRandom = id
			}
		case s == "/skip":
			l.Skip(l.lastRandom)
		case s == "/print":
			l.PrintTelegramBot(buf)
		case s == "/clear":
			l.Clear()
		case s == "/help":
			buf.WriteString(HELPTELEGRAM)
		case strings.HasPrefix(s, "/add "):
			{
				e := strings.Split(strings.TrimPrefix(s, "/add "), " ")

				for _, entry := range e {
					var num int
					fmt.Sscanf(entry, "%d", &num)
					l.Mark(num, true)
				}
			}
		case strings.HasPrefix(s, "/meta "):
			{
				s := strings.TrimPrefix(s, "/meta ")
				var num int
				fmt.Sscanf(s, "%d", &num)
				if num > 0 && num <= len(l.list) {
					(*l.sl)[l.list[num - 1]].PrintMetaTelegramBot(buf)
				}
			}
		case strings.HasPrefix(s, "/search "):
			{
				entry := strings.TrimPrefix(s, "/search ")
				res := l.Search(entry)
				if res != nil {
					res.PrintTelegramBot(buf)
					users[username] = append(users[username], res)
					fmt.Fprintf(buf, "%s -> %s\n", l.path, res.path)
				}
			}
		case s == "/back":
			{
				if len(user) != 1 {
					newL := user[len(user) - 2]
					user = user[:len(user) - 1]
					users[username] = user
					fmt.Fprintf(buf, "%s -> %s\n", l.path, newL.path)
				}
			}
		case s == "/write":
			{
				first := user[0]
				w, err := os.OpenFile("../users/"+first.username+".txt", os.O_WRONLY|os.O_CREATE, 0666)
				if err == nil {
					first.WriteUser(w)
					w.Close()
				}
			}
		case s == "/seen", s == "/unseen":
			{
				pred := false
				if s == "/seen" {
					pred = true
				}
				res := l.Seen(pred)
				if res != nil {
					res.PrintTelegramBot(buf)
					users[username] = append(users[username], res)
					fmt.Fprintf(buf, "%s -> %s\n", l.path, res.path)
				}
			}
		}

		if buf.Len() != 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, buf.String())
			msg.ReplyToMessageID = update.Message.MessageID
			if _, err := bot.Send(msg); err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Printf("buf is empty\n")
		}
	}
}

type Response struct {
	buf *bytes.Buffer
	msg *tgbotapi.Message
}

type UserTelegramBot struct {
	list []*List

	lock sync.RWMutex
}

func NewUser(username string, sl *SourceList) (*UserTelegramBot, error) {
	l := NewList(sl)
	if err := l.ReadUser(username); err != nil {
		return nil, err
	}
	l, _ = l.SubList("wallfilm")
	user := &UserTelegramBot{
		list: []*List{l},
	}
	return user, nil
}

func (user *UserTelegramBot) getLast() *List {
	if user == nil || user.list == nil {
		return nil
	}

	return user.list[len(user.list) - 1]
}

func (user *UserTelegramBot) random(buf *bytes.Buffer) {
	user.lock.RLock()
	defer user.lock.RUnlock()
	l := user.getLast()
	id := l.Random()
	if id >= 0 && id < len(l.list) {
		PrintHeaderTelegramBot(buf)
		l.PrintRecordTelegramBot(id, buf)
		l.lastRandom = id
	}
}

func (user *UserTelegramBot) skip(buf *bytes.Buffer) {
	user.lock.Lock()
	defer user.lock.Unlock()
	l := user.getLast()
	l.Skip(l.lastRandom)
	fmt.Fprintf(buf, "Skip %d\n", l.lastRandom)
}

func (user *UserTelegramBot) print(buf *bytes.Buffer) {
	user.lock.RLock()
	defer user.lock.RUnlock()
	user.getLast().PrintTelegramBot(buf)
}

func (user *UserTelegramBot) clear(buf *bytes.Buffer) {
	user.lock.Lock()
	defer user.lock.Unlock()
	user.getLast().Clear()
	fmt.Fprintf(buf, "Clear skiplist\n")
}

func (user *UserTelegramBot) back(buf *bytes.Buffer) {
	user.lock.Lock()
	defer user.lock.Unlock()
	if len(user.list) > 1 {
		prevL := user.getLast()
		newL := user.list[len(user.list) - 2]
		user.list = user.list[:len(user.list) - 1]
		fmt.Fprintf(buf, "%s -> %s\n", prevL.path, newL.path)
	}
}

func (user *UserTelegramBot) write(buf *bytes.Buffer) {
	user.lock.Lock()
	defer user.lock.Unlock()
	first := user.list[0]
	w, err := os.OpenFile("../users/"+first.username+".txt", os.O_WRONLY|os.O_CREATE, 0666)
	if err == nil {
		first.WriteUser(w)
		w.Close()
	}
	fmt.Fprintf(buf, "Write userinfo to server\n")
}

func (user *UserTelegramBot) seen(buf *bytes.Buffer, pred bool) {
	user.lock.Lock()
	defer user.lock.Unlock()
	l := user.getLast()
	res := l.Seen(pred)
	if res != nil {
		res.PrintTelegramBot(buf)
		user.list = append(user.list, res)
		fmt.Fprintf(buf, "%s -> %s\n", l.path, res.path)
	}
}

func (user *UserTelegramBot) add(buf *bytes.Buffer, s string) {
	user.lock.Lock()
	defer user.lock.Unlock()
	l := user.getLast()
	e := strings.Split(strings.TrimPrefix(s, "/add "), " ")
	for _, entry := range e {
		var num int
		fmt.Sscanf(entry, "%d", &num)
		l.Mark(num, true)
	}
	fmt.Fprintf(buf, "Added %s\n", strings.TrimPrefix(s, "/add "))
}

func (user *UserTelegramBot) meta(buf *bytes.Buffer, s string) {
	user.lock.RLock()
	defer user.lock.RUnlock()
	l := user.getLast()
	s = strings.TrimPrefix(s, "/meta ")
	var num int
	fmt.Sscanf(s, "%d", &num)
	if num > 0 && num <= len(l.list) {
		(*l.sl)[l.list[num - 1]].PrintMetaTelegramBot(buf)
	}
}

func (user *UserTelegramBot) search(buf *bytes.Buffer, s string) {
	user.lock.Lock()
	defer user.lock.Unlock()
	l := user.getLast()
	s = strings.TrimPrefix(s, "/search ")
	res := l.Search(s)
	if res != nil {
		res.PrintTelegramBot(buf)
		user.list = append(user.list, res)
		fmt.Fprintf(buf, "%s -> %s\n", l.path, res.path)
	}
}

func (user *UserTelegramBot) UpdateCallback(update tgbotapi.Update, out chan Response) {
	s := update.Message.Text
	buf := new(bytes.Buffer)

	switch {
	case s == "/random":
		user.random(buf)
	case s == "/skip":
		user.skip(buf)
	case s == "/print":
		user.print(buf)
	case s == "/clear":
		user.clear(buf)
	case s == "/back":
		user.back(buf)
	case s == "/write":
		user.write(buf)
	case s == "/seen", s == "/unseen":
		pred := false
		if s == "/seen" {
			pred = true
		}
		user.seen(buf, pred)

	case s == "/help":
		buf.WriteString(HELPTELEGRAM)
	case strings.HasPrefix(s, "/add "):
		user.add(buf, s)
	case strings.HasPrefix(s, "/meta "):
		user.meta(buf, s)
	case strings.HasPrefix(s, "/search "):
		user.search(buf, s)
	}

	out <- Response{buf, update.Message}
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

	out := make(chan Response)

	for {
		select {
		case update := <-updates:
			{
				username := update.Message.From.UserName
				user, ok := users[username]
				if !ok {
					var err error
					if user, err = NewUser(username, sl); err != nil {
						continue
					}
					users[username] = user
				}
				go user.UpdateCallback(update, out)
			}
		case r := <-out:
			{
				if r.buf.Len() != 0 {
					msg := tgbotapi.NewMessage(r.msg.Chat.ID, r.buf.String())
					msg.ReplyToMessageID = r.msg.MessageID
					if _, err := bot.Send(msg); err != nil {
						fmt.Println(err)
					}
				} else {
					fmt.Printf("buf is empty\n")
				}
			}

		}
	}
}