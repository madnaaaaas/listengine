package main

import (
	"github.com/madnaaaaas/listengine"
)

const SOURCEFILENAME = "db.txt"

func main() {
	sl, err := listengine.NewSourceList(SOURCEFILENAME)
	if err != nil {
		return
	}
	listengine.TelegramBotGoroutines(sl, "YOUR TELEGRAM BOT TOKEN")
}