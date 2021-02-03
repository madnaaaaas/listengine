package main

import (
	"github.com/madnaaaaas/listengine"
)

const SOURCEFILENAME = "source.txt"

func main() {
	sl, err := listengine.NewSourceList(SOURCEFILENAME)
	if err != nil {
		return
	}
	listengine.TelegramBot(sl, "YOUR TOKEN FOR TELEGRAM BOT")
}