package listengine

import (
	"bytes"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"io"
	"os"
	"sort"
	"strings"
)

const HELPTELEGRAM = `Commands:
	/help - help
	/random - get random film
	/skip - skip previous random film
	/clear - clear skip list
	/print - print current list
	/add <numbers> - add films
	/search <keywords> - search films by keywords in names
	/write - write list to user file
	/back - back to previous list
`
func PrintHeaderTelegramBot(w io.Writer) {
	fmt.Fprintf(w, " V| Num| Name\n")
}

func (r *Record) PrintTelegramBot(w io.Writer) {
	v := ""
	if r.viewed {
		v = "+"
	} else {
		v = "-"
	}
	fmt.Fprintf(w, " %s|%4d|%s\n", v, r.sr.num, r.sr.name)
}

func (l *List) PrintTelegramBot(w io.Writer) {
	if len(l.list) == 0 {
		fmt.Fprintf(w, "EMPTY LIST\n")
		return
	}
	fmt.Fprintf(w, "%s:%s (total %d):\n", l.username, l.path, len(l.list))
	tlist := make([]*Record, 0, len(l.list))
	for _, r := range l.list {
		tlist = append(tlist, r)
	}
	sort.Slice(tlist, func(i, j int) bool { return tlist[i].sr.num < tlist[j].sr.num })
	PrintHeaderTelegramBot(w)
	for _, r := range tlist {
		r.PrintTelegramBot(w)
	}
	if l.skip != nil {
		l.SkipList().PrintTelegramBot(w)
	}
}

type TGUserInfo struct {
	l *List
	lastRnd int
}

func TelegramBot(sl *SourceList, token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return
	}

	//bot.Debug = true

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return
	}

	//lastRnd := 0
	//array := []*List{first}

	users := make(map[string][]*TGUserInfo)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		fmt.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)
		var info *TGUserInfo
		user, ok := users[update.Message.From.UserName]
		if ok {
			info = user[len(user) - 1]
		} else {
			users[update.Message.From.UserName] = make([]*TGUserInfo, 1)
			info = &TGUserInfo{l : NewList(sl)}
			info.l.ReadUser(update.Message.From.UserName)
			users[update.Message.From.UserName][0] = info
		}
		s := update.Message.Text
		buf := new(bytes.Buffer)

		switch {
		case s == "/random":
			r := info.l.Random()
			if r != nil {
				PrintHeaderTelegramBot(buf)
				r.PrintTelegramBot(buf)
				info.lastRnd = r.sr.num
			}
		case s == "/skip":
			info.l.Skip(info.lastRnd)
		case s == "/print":
			info.l.PrintTelegramBot(buf)
		case s == "/clear":
			info.l.Clear()
		case s == "/help":
			buf.WriteString(HELPTELEGRAM)
		case strings.HasPrefix(s, "/add "):
			{
				e := strings.Split(strings.TrimPrefix(s, "/add "), " ")

				for _, entry := range e {
					var num int
					fmt.Sscanf(entry, "%d", &num)
					info.l.Mark(num, true)
				}
			}
		case strings.HasPrefix(s, "/search "):
			{
				entry := strings.TrimPrefix(s, "/search ")
				res := info.l.Search(entry)
				if res != nil {
					res.PrintTelegramBot(buf)
					//array = append(array, l)
					users[update.Message.From.UserName] = append(users[update.Message.From.UserName],
						&TGUserInfo{l : res})
					fmt.Fprintf(buf, "%s |-> %s\n", info.l.path, res.path)
				}
			}
		case s == "/back":
			{
				if len(user) != 1 {
					prev := info.l
					user = user[:len(user) - 1]
					users[update.Message.From.UserName] = user
					fmt.Fprintf(buf, "%s |-> %s\n", prev.path, user[len(user) - 1].l.path)
				}
			}
		case s == "/write":
			{
				first := user[0]
				w, err := os.OpenFile("../users/"+first.l.username+".txt", os.O_WRONLY|os.O_CREATE, 0666)
				if err == nil {
					first.l.WriteUser(w)
					w.Close()
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