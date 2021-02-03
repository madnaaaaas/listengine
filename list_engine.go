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

const TOTAL = 150

type SourceRecord struct {
	name string
	num  int
}

func (r SourceRecord) Write(w io.Writer) {
	s := fmt.Sprintf("%d;%s\n", r.num, r.name)
	w.Write([]byte(s))
}

type Record struct {
	sr     *SourceRecord
	viewed bool
}

func (r Record) Write(w io.Writer) {
	if r.sr == nil {
		return
	}
	v := ""
	if r.viewed {
		v = "+"
	} else {
		v = "-"
	}
	s := fmt.Sprintf("%d;%s;%s\n", r.sr.num, r.sr.name, v)
	w.Write([]byte(s))
}

type SourceList struct {
	list map[int]*SourceRecord
}

func (sl *SourceList) Write(w io.Writer) {
	for _, sr := range sl.list {
		fmt.Fprintf(w, "%d;%s\n", sr.num, sr.name)
	}
}

func (sl *SourceList) AddSource(num int, name string) {
	if num >= 1 && num <= TOTAL {
		sl.list[num] = &SourceRecord{num: num, name: name}
	}
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
		var num int
		var name string
		fmt.Sscanf(line, "%d;%s", &num, &name)
		sl.AddSource(num, name)
	}
	return nil
}

func NewSourceList(filename string) (*SourceList, error) {
	sl := &SourceList{list: make(map[int]*SourceRecord)}
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
	list     map[int]*Record
	skip     map[int]bool
	v_count  int
	path     string
	username string
}

func (l *List) Skip(n int) {
	if l.skip == nil {
		l.skip = make(map[int]bool)
	}
	l.skip[n] = true
}

func (l *List) Clear() {
	l.skip = nil
}

func (l *List) Mark(num int, v bool) {
	if num >= 1 && num <= TOTAL {
		if _, ok := l.list[num]; !ok {
			return
		}
		if !l.list[num].viewed && v {
			l.v_count++
		}
		if l.list[num].viewed && !v {
			l.v_count--
		}
		l.list[num].viewed = v
	}
}

func NewList(sl *SourceList) *List {
	l := &List{path: "Source", list: make(map[int]*Record)}
	for num, sr := range sl.list {
		l.list[num] = &Record{sr: sr}
	}
	return l
}

func (l *List) Write(w io.Writer) {
	for _, r := range l.list {
		r.Write(w)
	}
}

func (l *List) AddRecord(r *Record) {
	if _, ok := l.list[r.sr.num]; ok {
		if !l.list[r.sr.num].viewed && r.viewed {
			l.v_count++
		}
		if l.list[r.sr.num].viewed && !r.viewed {
			l.v_count--
		}
	} else {
		l.list[r.sr.num] = r
		if r.viewed {
			l.v_count++
		}
	}
}

func (l *List) ReadUser(username string) error {
	f, err := os.Open("../users/" + username + ".txt")
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
		l.Mark(num, true)
	}
	l.username = username
	return nil
}

func (l *List) Random() *Record {
	if l.v_count+len(l.skip) >= TOTAL {
		return nil
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(TOTAL) + 1
	for _, ok := l.list[n]; !ok || (l.list[n].viewed || l.skip[n]); {
		n = (n % TOTAL) + 1
		_, ok = l.list[n]
	}
	return l.list[n]
}

func (l *List) SkipList() *List {
	if len(l.list) == 0 || len(l.skip) == 0 {
		return nil
	}
	ret := &List{username: l.username, list: make(map[int]*Record)}
	for n, _ := range l.skip {
		if r, ok := l.list[n]; ok {
			ret.AddRecord(r)
		}
	}
	ret.path = l.path + "->Skiplist"
	return ret
}

func (l *List) Search(str string) *List {
	ret := &List{username: l.username, list: make(map[int]*Record)}
	words := strings.Split(strings.ToUpper(str), " ")
	for _, r := range l.list {
		for _, w := range words {
			if strings.Contains(strings.ToUpper(r.sr.name), w) {
				ret.AddRecord(r)
				break
			}
		}
	}
	if len(ret.list) == 0 {
		return nil
	}
	ret.path = l.path + "->Search(" + str + ")"
	return ret
}

func (l *List) WriteUser(w io.Writer) {
	for _, r := range l.list {
		if r.viewed {
			fmt.Fprintf(w, "%d\n", r.sr.num)
		}
	}
}
