package main

import (
	"github.com/madnaaaaas/listengine/tgbot"
)

const SOURCEFILENAME = "db.txt"

func main() {
	tgbot.TelegramBotGoroutines(SOURCEFILENAME, "YOUR TELEGRAM BOT TOKEN")
}