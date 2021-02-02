package main

import (
	"listengine"
)

const USERNAME = "user"
const SOURCEFILENAME = "source.txt"

func main() {
	sl, err := listengine.NewSourceList(SOURCEFILENAME)
	if err != nil {
		return
	}
	l := listengine.NewList(sl)
	err = l.ReadUser(USERNAME)
	if err != nil {
		return
	}
	listengine.Console(l)
}