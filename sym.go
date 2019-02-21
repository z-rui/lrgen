package main

type Assoc int8

const (
	UNSPEC Assoc = iota
	LEFT
	RIGHT
	NONASSOC
)

type Symbol struct {
	Id       int
	Name     string
	Type     string
	Nullable bool
	IsNt     bool
	Assoc    Assoc
	Prec     int
	LhsProd  []*Prod // productions on lhs
	RhsProd  int     // # occurrences on rhs
	First    BitSet  // FIRST set
}

func (s *Symbol) String() string {
	return s.Name
}

type SymTab struct {
	All     []*Symbol
	NtBase  int
	nameMap map[string]*Symbol
}

// Lookup looks up a symbol by its name.
// If not exists, add the symbol to the table,
func (t *SymTab) Lookup(name string) *Symbol {
	if t.nameMap == nil {
		t.nameMap = make(map[string]*Symbol)
	}
	if sym, ok := t.nameMap[name]; ok {
		return sym
	}
	sym := &Symbol{
		Id:   len(t.All),
		Name: name,
		IsNt: t.NtBase > 0,
	}
	t.All = append(t.All, sym)
	t.nameMap[name] = sym
	return sym
}

func (t *SymTab) AllT() []*Symbol {
	return t.All[:t.NtBase]
}

func (t *SymTab) AllNt() []*Symbol {
	return t.All[t.NtBase:]
}

func (t *SymTab) StartNt() {
	t.NtBase = t.Count()
}

func (t *SymTab) IsNt(s *Symbol) bool {
	return s.Id >= t.NtBase
}

func (t *SymTab) Count() int {
	return len(t.All)
}

func (t *SymTab) CountNt() int {
	return t.Count() - t.NtBase
}

// GenFirst computes the FIRST set of each symbol
// as well as the Nullable attribute.
func (t *SymTab) GenFirst() {
	for _, sym := range t.All {
		sym.First = t.NewTermSet()
		if !t.IsNt(sym) {
			sym.First.Set(uint(sym.Id))
		}
	}
	for {
		changed := false
		for _, lhs := range t.AllNt() {
		L:
			for _, prod := range lhs.LhsProd {
				for _, sym := range prod.Rhs {
					if lhs.First.Union(sym.First) {
						changed = true
					}
					if !sym.Nullable {
						continue L
					}
				}
				if !lhs.Nullable {
					changed = true
					lhs.Nullable = true
				}
			}
		}
		if !changed {
			break
		}
	}
}

func (t *SymTab) NewTermSet() BitSet {
	return NewBitSet(uint(t.Count()))
}
