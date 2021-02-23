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

type Record struct {
	name string
	num  int
	m Meta
}

func (r Record) Write(w io.Writer) {
	s := fmt.Sprintf("%d;%s;%s\n", r.num, r.name, r.m.String())
	w.Write([]byte(s))
}

type SourceList []Record

func (sl *SourceList) Write(w io.Writer) {
	for _, sr := range *sl {
		//fmt.Fprintf(w, "%d;%s\n", sr.num, sr.name)
		sr.Write(w)
	}
}

func (sl *SourceList) AddSource(name string, m Meta) {
	*sl = append(*sl, Record{num : len(*sl), name : name, m : m})
}

func (sl *SourceList) Read(r io.Reader) error {
	bufr := bufio.NewReader(r)
	for {
		line, err := bufr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		line = strings.TrimSpace(line)
		t := strings.Split(line, ";")
		if len(t) != 3 {
			continue
		}
		//num, _ := strconv.Atoi(t[0])
		name := t[1]
		m := NewMeta(t[2])
		sl.AddSource(name, m)
	}
	return nil
}

func NewSourceList(filename string) (*SourceList, error) {
	sl := new(SourceList)
	*sl = make([]Record, 0)
	r, err := os.Open("../sources/" + filename)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	err = sl.Read(r)
	if err != nil {
		return nil, err
	}
	return sl, err
}

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
	defer f.Close()
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
		fmt.Sscanf(line, "%d", &num)
		ret.AddRecord(num)
	}
	ret.path = l.path + ".Sublist(" + name + ")"
	return ret, nil
}

func (l *List) ReadUser(username string) error {
	l.username = username
	l.viewed = make(map[int]bool)
	f, err := os.OpenFile("notes.txt", os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
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
		fmt.Sscanf(line, "%d", &num)
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
		for _, w := range words {
			if strings.Contains(strings.ToUpper((*l.sl)[slId].name), w) {
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
	defer w.Close()

	for slId := range l.viewed {
		fmt.Fprintf(w, "%d\n", slId)
	}
	return nil
}

func (l *List) Check(num int) bool {
	return l.viewed[l.list[num]]
}

func (l *List) GetRecord(num int) Record {
	return (*l.sl)[l.list[num]]
}
