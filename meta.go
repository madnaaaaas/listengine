package listengine

import "strings"

type Meta map[string]string

func NewMeta(s string) Meta {
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return nil
	}

	s = s[1:len(s) - 1]
	res := make(Meta)
	array := strings.Split(s, "]")
	for _, entry := range array {
		t := strings.Split(entry, "[")
		if len(t) != 2 {
			continue
		}
		res[t[0]] = t[1]
	}

	return res
}

func (m Meta) Add(key, value string) {
	if m == nil {
		return
	}

	str, ok := m[key]
	if ok {
		str += ", "
	}
	m[key] = str + value
}

func (m Meta) String() string {
	if m == nil {
		return "{}"
	}
	res := "{"
	for name, value := range m {
		res += name+"["+value+"]"
	}
	res += "}"

	return res
}
