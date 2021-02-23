package listengine

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

type List struct {
	sl *SourceList
	viewed map[int]bool // sl_id
	skip map[int]bool // id
	vCount int
	list []int // id -> sl_id
	username string
	path string
	lastRandom int
}

func (l *List) Skip(num int) {
	if l.skip == nil {
		l.skip = make(map[int]bool)
	}
	l.skip[num] = true
}

func (l *List) Clear() {
	l.skip = nil
}

func (l *List) Mark(num int, v bool) {
	if num >= 0 && num < len(l.list) {
		slId := l.list[num]
		if !l.viewed[slId] && v {
			l.vCount++
		}
		if l.viewed[slId]  && !v {
			l.vCount--
		}
		if v {
			l.viewed[slId] = true
		} else {
			delete(l.viewed, slId)
		}
	}
}

func NewList(sl *SourceList) *List {
	return &List{
		sl: sl,
		path: "Source",
		list: make([]int, 0),
	}
}

func NewFullList(l *List) *List {
	ret := &List{
		sl: l.sl,
		path: "Source",
		viewed: l.viewed,
	}
	list := make([]int, len(*l.sl))
	for i := range list {
		list[i] = i
		if ret.viewed[i] {
			ret.vCount++
		}
	}
	ret.list = list
	return ret
}

func (l *List) Copy() *List {
	return &List{
		sl: l.sl,
		username: l.username,
		viewed: l.viewed,
		list: make([]int, 0),
	}
}

func (l *List) Write(w io.Writer) {
	for _, id := range l.list {
		(*l.sl)[id].Write(w)
	}
}

func (l *List) AddRecord(slId int) {
	l.list = append(l.list, slId)
	if l.viewed[slId] {
		l.vCount++
	}
}

func (l *List) SubList(name string) (*List, error) {
	f, err := os.Open("../lists/" + name + ".txt")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	bufr := bufio.NewReader(f)
	ret := l.Copy()

	for {
		line, err := bufr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		var num int
		if _, err := fmt.Sscanf(line, "%d", &num); err != nil {
			fmt.Println(err)
		}
		ret.AddRecord(num)
	}
	ret.path = l.path + ".Sublist(" + name + ")"
	return ret, nil
}

func (l *List) ReadUser(username string) error {
	l.username = username
	l.viewed = make(map[int]bool)
	f, err := os.OpenFile("../users/" + username + ".txt", os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	bufr := bufio.NewReader(f)
	for {
		line, err := bufr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		var num int
		if _, err := fmt.Sscanf(line, "%d", &num); err != nil {
			fmt.Println(err)
		}
		if num >= 0 && num < len(*l.sl) {
			l.viewed[num] = true
		}
	}
	return nil
}

func (l *List) Random() int {
	if l.vCount+len(l.skip) >= len(l.list) {
		return -1
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(len(l.list))
	for l.Check(n) || l.skip[n]{
		n = (n + 1) % len(l.list)
	}
	return n
}

func (l *List) SkipList() *List {
	if l.skip == nil || len(l.skip) == 0 {
		return nil
	}
	ret := l.Copy()
	for n  := range l.skip {
		if n >= 0 && n < len(l.list) {
			ret.AddRecord(l.list[n])
		}
	}
	ret.path = l.path + ".Skiplist"
	return ret
}

func (l *List) Search(str string) *List {
	ret := l.Copy()
	words := strings.Split(strings.ToUpper(str), " ")
	for _, slId := range l.list {
		r := (*l.sl)[slId]
		entry := strings.ToUpper(r.Name)
		if r.Meta != nil {
			if s, ok := r.Meta["name_en"]; ok {
				entry += " " + strings.ToUpper(s)
			}
		}
		for _, w := range words {
			if strings.Contains(entry, w) {
				ret.AddRecord(slId)
				break
			}
		}
	}
	if len(ret.list) == 0 {
		return nil
	}
	ret.path = l.path + ".Search(" + str + ")"
	return ret
}

func (l *List) Seen(pred bool) *List {
	ret := l.Copy()
	for _, slId := range l.list {
		if l.viewed[slId] == pred {
			ret.AddRecord(slId)
		}
	}
	if len(ret.list) == 0 {
		return nil
	}
	s := ""
	if pred {
		s = "true"
	} else {
		s = "false"
	}
	ret.path = l.path + ".Seen(" + s + ")"
	return ret
}

func (l *List) WriteUser() error {
	w, err := os.OpenFile("../users/"+l.username+".txt", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for slId := range l.viewed {
		if _, err := fmt.Fprintf(w, "%d\n", slId); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (l *List) Check(num int) bool {
	return l.viewed[l.list[num]]
}

func (l *List) GetRecord(num int) Record {
	return (*l.sl)[l.list[num]]
}

func (l *List) Len() int {
	return len(l.list)
}

func (l *List) SlLen() int {
	return len(*l.sl)
}

func (l *List) Path() string {
	return l.path
}

func (l *List) ViewedCount() int {
	return l.vCount
}
