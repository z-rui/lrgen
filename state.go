package main

import (
	"fmt"
	"io"
	"sort"
)

type StateItem struct {
	Item
	LookAhead BitSet
	Nullable  bool
	First     BitSet
}

type Action int

const (
	NONE  Action = -iota // undefined
	ERROR                // explicit error
	ACCEPT
	SHIFT
	// positive numbers mean reduce
)

type Conflict struct {
	Sym  *Symbol
	Item Item
}

type State struct {
	Id      int
	Kernel  ItemSet
	Closure []StateItem
	Action  []Action // indexed by syid, terminals
	Goto    []*State // indexed by syid
	CLA     []BitSet // common LA, indexed by syid-ntBase
	Conf    []Conflict
	Default Action // default action
}

func (s *State) genFirst() {
L:
	for i := range s.Closure {
		it := &s.Closure[i]
		if !it.Final() {
			for _, sym := range it.Rhs[it.Pos+1:] {
				it.First.Union(sym.First)
				if !sym.Nullable {
					continue L
				}
			}
		}
		it.Nullable = true
	}
}

func (s *State) dumpItems(w io.Writer, closure bool) {
	for i, it := range s.Closure {
		if !closure && i >= len(s.Kernel) && !it.Final() {
			continue
		}
		fmt.Fprintf(w, "\t%v", it)
		if it.Final() {
			fmt.Fprintf(w, "    (%d)", it.Id)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
}

func (s *State) dumpActions(w io.Writer, sy *SymTab) {
	any := false

	// terminal actions
	for i, sym := range sy.AllT() {
		act := s.Action[i]
		if act == s.Default || act == NONE {
			continue
		}
		fmt.Fprintf(w, "\t%v\t", sym)
		switch act {
		case ERROR:
			fmt.Fprintln(w, "error")
		case ACCEPT:
			fmt.Fprintln(w, "accept")
		case SHIFT:
			fmt.Fprintf(w, "shift\t%d\n", s.Goto[i].Id)
		default:
			fmt.Fprintf(w, "reduce\t%d\n", act)
		}
		any = true
	}

	// conflicts
	for _, conf := range s.Conf {
		var prev string
		switch act := s.Action[conf.Sym.Id]; act {
		case NONE, ERROR:
			panic("BUG: cannot reduce on state")
		case ACCEPT, SHIFT:
			prev = "shift"
		default:
			prev = "reduce"
		}
		fmt.Fprintf(w, "\t%v\treduce\t%d\t**%s/reduce conflict**\n",
			conf.Sym, conf.Item.Id, prev)
	}
	if any {
		fmt.Fprintln(w)
		any = false
	}

	// nonterminal actions
	for i, sym := range sy.AllNt() {
		if dest := s.Goto[i+sy.NtBase]; dest != nil {
			fmt.Fprintf(w, "\t%v\tgoto\t%d\n", sym, dest.Id)
			any = true
		}
	}
	if any {
		fmt.Fprintln(w)
	}

	// default action
	if s.Default == 0 {
		fmt.Fprintf(w, "\t.\terror\n")
	} else {
		fmt.Fprintf(w, "\t.\treduce\t%d\n", s.Default)
	}
}

type StTab struct {
	sy  SymTab
	pr  ProdTab
	All []*State
}

func (t *StTab) GenAll() {
	s := t.newState(ItemSet{{t.pr.All[0], 0}})
	t.genStates(s)
	t.genLookAhead()
	t.genReduce()
}

func (t *StTab) newState(kernel ItemSet) *State {
	sort.Sort(kernel)
	// search for existing state
	for _, s := range t.All {
		if kernel.Equal(s.Kernel) {
			return s
		}
	}
	s := &State{
		Id:     len(t.All),
		Kernel: kernel,
		Action: make([]Action, t.sy.NtBase),
		Goto:   make([]*State, t.sy.Count()),
		CLA:    make([]BitSet, t.sy.CountNt()),
	}
	t.makeClosure(s)
	s.genFirst()
	t.All = append(t.All, s)
	return s
}

func (t *StTab) makeClosure(s *State) {
	var q []*Symbol
	expand := func(it Item) {
		if it.Final() {
			return
		}
		if next := it.Next(); t.sy.IsNt(next) {
			if cla := &s.CLA[next.Id-t.sy.NtBase]; *cla == nil {
				q = append(q, next)
				*cla = t.sy.NewTermSet()
			}
		}
	}
	for _, it := range s.Kernel {
		s.Closure = append(s.Closure, StateItem{
			Item:      it,
			LookAhead: t.sy.NewTermSet(),
			Nullable:  false,
			First:     t.sy.NewTermSet(),
		})
		expand(it)
	}
	var it Item
	// note: q may change
	for i := 0; i < len(q); i++ {
		lhs := q[i]
		la := s.CLA[lhs.Id-t.sy.NtBase]
		for _, prod := range lhs.LhsProd {
			if !prod.Reducible {
				continue // don't bother with useless productions
			}
			it.Prod = prod
			s.Closure = append(s.Closure, StateItem{
				Item:      it,
				LookAhead: la,
				Nullable:  false,
				First:     t.sy.NewTermSet(),
			})
			expand(it)
		}
	}
}

func (t *StTab) genStates(s *State) {
	q := []*State{s}
	for len(q) > 0 {
		s, q = q[0], q[1:]
		nextItems := make([]ItemSet, t.sy.Count())
		for _, it := range s.Closure {
			if it.Final() {
				continue
			}
			next := it.Next()
			if next.Id == 0 {
				s.Action[0] = ACCEPT
				continue // do not shift $end
			}
			nextItems[next.Id] = append(nextItems[next.Id], Item{
				Prod: it.Prod,
				Pos:  it.Pos + 1,
			})
		}
		for i, kernel := range nextItems {
			if len(kernel) == 0 {
				continue
			}
			ns := t.newState(kernel)
			s.Goto[i] = ns
			if i < t.sy.NtBase {
				s.Action[i] = SHIFT
			}
			if ns.Id == t.Count()-1 { // new state
				q = append(q, ns)
			}
		}
	}
}

func (t *StTab) genLookAhead() {
	q := make([]*State, 0, len(t.All))
	inQ := NewBitSet(uint(len(t.All)))
	enQ := func(s *State) {
		if !inQ.Test(uint(s.Id)) {
			q = append(q, s)
			inQ.Set(uint(s.Id))
		}
	}
	deQ := func() (s *State) {
		s = q[0]
		q = q[1:]
		inQ.Clear(uint(s.Id))
		return
	}
	for _, s := range t.All {
		enQ(s)
	}
	for len(q) > 0 {
		s := deQ()
		for _, it := range s.Closure {
			if it.Final() {
				continue
			}
			next := it.Next()
			la := it.LookAhead
			if t.sy.IsNt(next) { // has closure items
				cla := s.CLA[next.Id-t.sy.NtBase]
				if cla.Union(it.First) || (it.Nullable && cla.Union(la)) {
					enQ(s)
				}
			}
			if ns := s.Goto[next.Id]; ns != nil {
				for j := range ns.Kernel {
					jt := ns.Closure[j]
					if it.Prod == jt.Prod && it.Pos+1 == jt.Pos {
						if jt.LookAhead.Union(la) {
							enQ(ns)
						}
					}
				}
			}
		}
	}
}

func (t *StTab) genReduce() {
	actCnt := make([]int, len(t.pr.All))
	for _, s := range t.All {
		for _, it := range s.Closure {
			if !it.Final() {
				continue
			}
			it.LookAhead.Range(func(i uint) {
				switch s.Action[i] {
				case NONE:
					s.Action[i] = Action(it.Id)
				case ERROR:
					return // do not override ERROR
				case ACCEPT, SHIFT: // shift/reduce conflict, try to resolve
					s0 := t.sy.All[i] // prec of lookahead
					s1 := it.PrecSym  // prec of production
					if s0.Assoc != UNSPEC && s1 != nil && s1.Assoc != UNSPEC {
						if s0.Prec == s1.Prec {
							switch s1.Assoc {
							case NONASSOC:
								s.Action[i] = ERROR // explicit error
							case LEFT:
								s.Action[i] = Action(it.Id) // use reduce
							}
						} else if s0.Prec < s1.Prec {
							s.Action[i] = Action(it.Id) // use reduce
						}
						break
					}
					fallthrough // cannot resolve
				default: // conflict
					s.Conf = append(s.Conf, Conflict{t.sy.All[i], it.Item})
					return
				}
				if act := s.Action[i]; act == Action(it.Id) {
					actCnt[act]++
					if actCnt[act] > actCnt[s.Default] {
						s.Default = act
					}
				}
			})
		}
	}
}

func (t *StTab) Dump(w io.Writer) {
	for sId, s := range t.All {
		fmt.Fprintf(w, "state-%d\n", sId)
		s.dumpItems(w, false)
		s.dumpActions(w, &t.sy)
		fmt.Fprintln(w)
	}
}

func (t *StTab) Count() int {
	return len(t.All)
}
