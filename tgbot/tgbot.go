package tgbot

import (
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/madnaaaaas/listengine"
)

const (
	telegramTypeNew = "new"
	telegramTypeMenu = "menu"
	telegramTypeView = "view"
	telegramTypeEdit = "edit"
	telegramTypeMeta = "meta"
	telegramTypeRandom = "random"
	telegramTypeSearch = "search"

	telegramCommandBack = "/back"
	telegramCommandView = "/view"
	telegramCommandEdit = "/edit"
	telegramCommandRandom = "/random"
	telegramCommandSearch = "/search"
	telegramCommandPrev = "/prev"
	telegramCommandNext = "/next"
	telegramCommandSkip = "/skip"
)

func MetaTelegramBot(r listengine.Record, num int, viewed bool) string {
	vs := "(-)"
	if viewed {
		vs = "(+)"
	}
	ret := fmt.Sprintf("%d. %s %s:\n", num, r.Name, vs)
	for k, v := range r.Meta {
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
		ret += fmt.Sprintf("%s: %s\n", k, v)
	}
	return ret
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

func isStop(msg *tgbotapi.Message) bool {
	return msg != nil && msg.Text == "/stop"
}

func isEmptyUpdate(update tgbotapi.Update) bool {
	return update.CallbackQuery == nil &&
		(update.Message == nil ||
			(update.Message.Text != "/start" && update.Message.Text != "/stop"))
}

func TelegramBotGoroutines(sourceFileName string, token string) {
	sl, err := listengine.NewSourceList(sourceFileName)
	if err != nil {
		return
	}
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

	users := make(map[string]*User)

	for update := range updates {
		username := userName(update)
		if username == "" {
			fmt.Println("empty username")
			continue
		}
		user, ok := users[username]
		if isEmptyUpdate(update) {
			fmt.Printf("%s: empty update\n", username)
			if ok && update.Message != nil {
				user.lastMsgID = 0
			}
			continue
		}
		if isStop(update.Message) {
			if ok {
				delete(users, username)
			}
			if _, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "До скорых встреч!")); err != nil {
				fmt.Println(err)
			}
			continue
		}
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