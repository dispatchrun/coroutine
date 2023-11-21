package types

import "fmt"

type stringid = uint32

type stringmap struct {
	strings []string

	seen map[string]stringid
}

func newStringMap(strings []string) *stringmap {
	return &stringmap{
		strings: strings,
		seen:    map[string]stringid{},
	}
}

func (m *stringmap) Intern(s string) stringid {
	if s == "" {
		return 0
	}
	if id, ok := m.seen[s]; ok {
		return id
	}
	m.strings = append(m.strings, s)
	id := stringid(len(m.strings)) // IDs start at 1
	m.seen[s] = id
	return id
}

func (m *stringmap) Lookup(id stringid) string {
	if id < 0 || int(id) > len(m.strings) {
		panic(fmt.Sprintf("string %d not found", id))
	}
	if id == 0 {
		return ""
	}
	return m.strings[id-1]
}
