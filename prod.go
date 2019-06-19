package main

import (
	"fmt"
	"strings"
)

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
		for {
			ch, err := r.ReadByte()
			if err != nil {
				break
			}
			switch ch {
			case '$':
				var n int
				_, err = fmt.Fscan(r, &n)
				if err == nil {
					n--
					fmt.Fprintf(w, "%sD[%d]", prefix, n)
					if 0 <= n && n < len(p.Rhs) && p.Rhs[n].Type != "" {
						fmt.Fprintf(w, ".%s", p.Rhs[n].Type)
					}
				} else {
					ch, err = r.ReadByte()
					if err == nil {
						if ch == '$' {
							fmt.Fprintf(w, "%sval", prefix)
							if p.Lhs.Type != "" {
								fmt.Fprintf(w, ".%s", p.Lhs.Type)
							}
						} else {
							w.WriteByte('$')
							w.WriteByte(ch)
						}
					}
				}
			default:
				w.WriteByte(ch)
			}
		}
		return w.String()
	} else if p.Id != 0 {
		t1 := p.Lhs.Type
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
