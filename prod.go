package main

import (
	"fmt"
	"strings"
)

//go:generate lexgen -o lex.yy1 -p yy1 dump.l

type Prod struct {
	Id        int
	Lhs       *Symbol
	Rhs       []*Symbol
	PrecSym   *Symbol
	Semant    string
	Reducible bool
}

const NoSemant = "}{" // this won't be an actual semantic action

func (p *Prod) Dump(prefix string) string {
	if p.Semant != NoSemant {
		r := strings.NewReader(p.Semant)
		w := new(strings.Builder)
		fmt.Fprintf(w, "\n\tcase %d:\n", p.Id)
		l := &yy1Lex{prod: p, wr: w}
		l.Init(r)
		for l.Lex(nil) != 0 {
		}
		return w.String()
	} else if t1 := p.Lhs.Type; t1 != "" {
		t2 := ""
		if len(p.Rhs) > 0 {
			t2 = p.Rhs[0].Type
		}
		if t1 != t2 {
			fmt.Printf("Rule %d: default action may clobber type\n", p.Id)
		}
	}
	return ""
}

type ProdTab struct {
	All []*Prod
}

func (t *ProdTab) NewProd(lhs *Symbol, rhs []*Symbol) *Prod {
	prod := &Prod{
		Id:     len(t.All),
		Lhs:    lhs,
		Rhs:    rhs,
		Semant: NoSemant,
	}
	t.All = append(t.All, prod)
	lhs.LhsProd = append(lhs.LhsProd, prod)
	for _, sym := range rhs {
		sym.RhsProd++
		if !sym.IsNt {
			prod.PrecSym = sym
		}
	}
	return prod
}
