package tgbot

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/madnaaaaas/listengine"
)

type State struct {
	l *listengine.List
	num int
	t string
}

func (st *State) msg() (tgbotapi.InlineKeyboardMarkup, string) {
	text := ""
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	var row []tgbotapi.InlineKeyboardButton
	switch st.t {
	case telegramTypeNew:
		text = "Выберите список фильмов"
		db := fmt.Sprintf("db (%d)", st.l.SlLen())
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(db, "db"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("wallfilm (150)", "wallfilm"))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case telegramTypeMenu:
		if st.l == nil {
			text = "Пустой список"
			break
		}
		text = fmt.Sprintf("Список: %s (Всего %d, Просмотренно %d)",
			st.l.Path(), st.l.Len(), st.l.ViewedCount())
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Просмотр", telegramCommandView))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Редактирование", telegramCommandEdit))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Поиск", telegramCommandSearch))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Случайный фильм", telegramCommandRandom))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case telegramTypeView:
		total := (st.l.Len() + 9)/ 10
		text = fmt.Sprintf("Просмотр списка: %s (Страница %d из %d)",
			st.l.Path(), st.num, total)
		for i := 10 * (st.num - 1); i < 10 * st.num && i < st.l.Len(); i++ {
			name := fmt.Sprintf("%d.%s", i + 1, st.l.GetRecord(i).Name)
			if st.l.Check(i) {
				name += " (+)"
			} else {
				name += " (-)"
			}
			data := fmt.Sprintf("%d", i)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name, data))
			keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Предыдущая страница", telegramCommandPrev))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Следующая страница", telegramCommandNext))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case telegramTypeEdit:
		total := (st.l.Len() + 9)/ 10
		text = fmt.Sprintf("Редактирование списка: %s (Страница %d из %d)",
			st.l.Path(), st.num, total)
		for i := 10 * (st.num - 1); i < 10 * st.num; i++ {
			name := fmt.Sprintf("%d.%s", i + 1, st.l.GetRecord(i).Name)
			if st.l.Check(i) {
				name += " (Удалить из просмотренного)"
			} else {
				name += " (Добавить к просмотренному)"
			}
			data := fmt.Sprintf("%d", i)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name, data))
			keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Предыдущая страница", telegramCommandPrev))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Следующая страница", telegramCommandNext))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case telegramTypeMeta:
		v := st.l.Check(st.num)
		text = MetaTelegramBot(st.l.GetRecord(st.num), st.num + 1, v)

		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Предыдущий фильм", telegramCommandPrev))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Следующий фильм", telegramCommandNext))
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
	case telegramTypeRandom:
		text = "Случайный фильм:"
		r := st.l.GetRecord(st.num)
		name := fmt.Sprintf("%d.%s", st.num + 1, r.Name)
		data := fmt.Sprintf("%d", st.num)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(name, data))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil

		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Пропустить", telegramCommandSkip))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Другой случайный фильм", telegramCommandRandom))
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	case telegramTypeSearch:
		text = "Введите ключевые слова для поиска через пробел"
	}
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Назад", telegramCommandBack))
	keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil

	return keyboard, text
}
