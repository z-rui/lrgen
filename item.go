package main

type Item struct {
	*Prod
	Pos int
}

func (it Item) Final() bool       { return it.Pos == len(it.Rhs) }
func (it Item) Next() *Symbol     { return it.Rhs[it.Pos] }
func (it Item) Equal(x Item) bool { return it.Prod == x.Prod && it.Pos == x.Pos }
func (it Item) Less(x Item) bool {
	return it.Prod.Id < x.Prod.Id || it.Prod.Id == x.Prod.Id && it.Pos == x.Pos
}

func (it Item) String() string {
	var s []byte
	if it.Lhs.Id == 0 {
		s = append(s, "$acc"...)
	} else {
		s = append(s, it.Lhs.String()...)
	}
	s = append(s, ": "...)
	for i, sym := range it.Rhs {
		if i == it.Pos {
			s = append(s, '.')
		}
		s = append(s, ' ')
		s = append(s, sym.Name...)
	}
	if len(it.Rhs) == it.Pos {
		s = append(s, '.')
	}
	return string(s)
}

type ItemSet []Item

func (s ItemSet) Len() int           { return len(s) }
func (s ItemSet) Less(i, j int) bool { return s[i].Less(s[j]) }
func (s ItemSet) Swap(i, j int)      { t := s[i]; s[i] = s[j]; s[j] = t }
func (s ItemSet) Equal(t ItemSet) bool {
	if len(s) != len(t) {
		return false
	}
	for i, it := range s {
		if !it.Equal(t[i]) {
			return false
		}
	}
	return true
}
