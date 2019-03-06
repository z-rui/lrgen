package main

type Prod struct {
	Id        int
	Lhs       *Symbol
	Rhs       []*Symbol
	PrecSym   *Symbol
	Semant    string
	Reducible bool
}

const NoSemant = "}{" // this won't be an actual semantic action

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
