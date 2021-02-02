package listengine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

const HELP = `Commands:
	EXT, ext, exit - exit
	HLP, hlp, help - help
	RND, rnd, random - get random film
	SKP, skp, skip - skip previous random film
	CLR, clr, clear - clear skip list
	PRN, prn, print - print state
	ADD <number> - add film
	SRC <keywords> - search films by keywords in names
	WRT - write list to user file
`

func Blunk(width int) string {
	ret := "_____________"
	for width > 0 {
		ret += "_"
		width -= 1
	}
	return ret
}

func PrintHeader(w io.Writer, width int) {
	name := "Name"
	for width > utf8.RuneCountInString(name) {
		if (width-utf8.RuneCountInString(name))%2 == 0 {
			name += " "
		} else {
			name = " " + name
		}
	}
	fmt.Fprintf(w, "%s\n|Num |%s|Viewed|\n%s\n", Blunk(width), name, Blunk(width))
}

func (r Record) PrintWidthWidth(w io.Writer, width int) {
	if r.sr == nil {
		return
	}
	v := ""
	if r.viewed {
		v = "+"
	} else {
		v = "-"
	}
	name := r.sr.name
	for width > utf8.RuneCountInString(name) {
		name = name + " "
	}
	fmt.Fprintf(w, "|%4d|%s|  %s   |\n%s\n", r.sr.num, name, v, Blunk(width))
}

func (r Record) Print(w io.Writer) {
	if r.sr == nil {
		return
	}
	width := utf8.RuneCountInString(r.sr.name)
	r.PrintWidthWidth(w, width)
}

func (l *List) Print(w io.Writer) {
	if len(l.list) == 0 {
		fmt.Fprintf(w, "EMPTY LIST\n")
		return
	}
	fmt.Fprintf(w, "%s:%s (total %d):\n", l.username, l.path, len(l.list))
	width := 0
	tlist := make([]*Record, 0, len(l.list))
	for _, r := range l.list {
		if utf8.RuneCountInString(r.sr.name) > width {
			width = utf8.RuneCountInString(r.sr.name)
		}
		tlist = append(tlist, r)
	}
	sort.Slice(tlist, func(i, j int) bool { return tlist[i].sr.num < tlist[j].sr.num })
	PrintHeader(w, width)
	for _, r := range tlist {
		r.PrintWidthWidth(w, width)
	}
	if l.skip != nil {
		l.SkipList().Print(w)
	}
}

func Console(l *List) {
	first := l
	lastRnd := 0
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Ready to go. " + HELP)
	array := []*List{first}
	for {
		scanner.Scan()
		s := scanner.Text()
		if len(s) == 0 {
			break
		}
		s = strings.ToUpper(s)
		switch {
		case s == "EXT", s == "EXIT":
			return
		case s == "RND", s == "RANDOM":
			r := l.Random()
			if r != nil {
				PrintHeader(os.Stdout, utf8.RuneCountInString(r.sr.name))
				r.Print(os.Stdout)
				lastRnd = r.sr.num
			}
		case s == "SKP", s == "SKIP":
			l.Skip(lastRnd)
		case s == "PRN", s == "PRINT":
			l.Print(os.Stdout)
		case s == "CLR", s == "CLEAR":
			l.Clear()
		case s == "HLP", s == "HELP":
			fmt.Print(HELP)
		case strings.HasPrefix(s, "ADD "):
			{
				entry := strings.TrimPrefix(s, "ADD ")
				var num int
				fmt.Sscanf(entry, "%d", &num)
				l.Mark(num, true)
			}
		case strings.HasPrefix(s, "SRC "):
			{
				entry := strings.TrimPrefix(s, "SRC ")
				res := l.Search(entry)
				if res != nil {
					res.Print(os.Stdout)
					array = append(array, l)
					fmt.Fprintf(os.Stdout, "%s |-> %s\n", l.path, res.path)
					l = res
				}
			}
		case s == "BCK", s == "BACK":
			{
				if len(array) != 1 {
					prev := l
					l = array[len(array)-1]
					array = array[:len(array)-1]
					fmt.Fprintf(os.Stdout, "%s |-> %s\n", prev.path, l.path)
				}
			}
		case s == "WRT", s == "WRITE":
			{
				w, err := os.OpenFile("../users/"+first.username+".txt", os.O_WRONLY|os.O_CREATE, 0666)
				if err == nil {
					first.WriteUser(w)
					w.Close()
				}
			}
		default:
			fmt.Printf("UNKNOWN COMMAND: %s\n", s)
		}
	}
}