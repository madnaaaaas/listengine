package listengine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Record struct {
	Name string
	Num  int
	Meta Meta
}

func (r Record) Write(w io.Writer) {
	s := fmt.Sprintf("%d;%s;%s\n", r.Num, r.Name, r.Meta.String())
	if _, err := w.Write([]byte(s)); err != nil {
		fmt.Println(err)
	}
}

type SourceList []Record

func (sl *SourceList) Write(w io.Writer) {
	for _, sr := range *sl {
		//fmt.Fprintf(w, "%d;%s\n", sr.num, sr.name)
		sr.Write(w)
	}
}

func (sl *SourceList) AddSource(name string, m Meta) {
	*sl = append(*sl, Record{Num : len(*sl), Name : name, Meta : m})
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
	defer func() {
		if err := r.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	err = sl.Read(r)
	if err != nil {
		return nil, err
	}
	return sl, err
}
