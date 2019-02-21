package main

import (
	"sort"
)

type Table [][]int

func (t Table) less(i, j int) bool {
	x, y := t[i], t[j]
	lx, ly := len(x), len(y)
	var n int
	if lx < ly {
		n = lx
	} else {
		n = ly
	}
	for k := 0; k < n; k++ {
		if x[k] < y[k] {
			return true
		}
	}
	return lx < ly
}

// Table compaction algorithm
// adapted from Tarjan & Yao, "Storing a sparse table".
func (t Table) Pack() (C, V, r []int) {
	rows := len(t)
	total := 0
	maxCols := 0

	// Step 1 - count nonzero positions
	nz := make([][]int, rows)
	for i, row := range t {
		total += len(row)
		if len(row) > maxCols {
			maxCols = len(row)
		}
		for j, val := range row {
			if val != 0 {
				nz[i] = append(nz[i], j)
			}
		}
	}

	// Step 2 - sort rows
	sorted := make([]int, rows)
	for i := range sorted {
		sorted[i] = i
	}
	sort.Slice(sorted, func(i, j int) bool {
		x, y := sorted[i], sorted[j]
		lx, ly := len(nz[x]), len(nz[y])
		return lx > ly || lx == ly && t.less(x, y)
	})

	// Step 3 - first-fit
	entry := NewBitSet(uint(total + 1))
	used := NewBitSet(uint(total + maxCols))
	min := 0  // first position where entry[min] = false
	max := -1 // last position where entry[max] = true
	r = make([]int, rows)

	prev := -1
	for _, i := range sorted {
		if len(nz[i]) == 0 { // all zero row
			r[i] = max + 1
			continue
		}
		if prev >= 0 && len(nz[prev]) == len(nz[i]) && !t.less(prev, i) { // duplicate row
			r[i] = r[prev]
			continue
		}
		prev = i
		ri := min - nz[i][0]
	check:
		if used.Test(uint(ri + maxCols)) {
			ri++
			goto check
		}
		for _, j := range nz[i] {
			if entry.Test(uint(ri + j)) {
				ri++
				goto check
			}
		}
		// ri is ok
		used.Set(uint(ri + maxCols))
		for _, j := range nz[i] {
			entry.Set(uint(ri + j))
			if ri+j > max {
				max = ri + j
			}
		}
		for entry.Test(uint(min)) {
			min++
		}
		r[i] = ri
	}

	// Generate C, V
	C = make([]int, max+1)
	V = make([]int, max+1)
	for i := range V {
		V[i] = -1
	}
	for i, row := range t {
		ri := r[i]
		for _, j := range nz[i] {
			C[ri+j] = row[j]
			V[ri+j] = j
		}
	}
	return C, V, r
}

type ParTab struct {
	Accept int   // accepting state
	R1, R2 []int // reduction rules
	Reduce []int // default reduce
	Goto   []int // default goto
	Action []int // C of Act+Goto
	Check  []int // V of Act+Goto
	Pact   []int // r of Act
	Pgoto  []int // r of Goto
}

func (t *ParTab) Size() int {
	return len(t.R1) + len(t.R2) +
		len(t.Reduce) + len(t.Goto) +
		len(t.Action) + len(t.Check) +
		len(t.Pact) + len(t.Pgoto)
}

func (g *LRGen) genParTab() {
	nProd := len(g.pr.All)
	nState := len(g.StTab.All)
	nT := g.sy.NtBase
	nNt := g.sy.Count() - nT
	t := ParTab{
		R1:     make([]int, nProd),
		R2:     make([]int, nProd),
		Reduce: make([]int, nState),
		Goto:   make([]int, nNt),
	}
	for i, prod := range g.pr.All {
		t.R1[i] = prod.Lhs.Id
		t.R2[i] = len(prod.Rhs)
	}
	var tab Table
	for sId, s := range g.StTab.All {
		row := make([]int, nT)
		for i, act := range s.Action {
			switch act {
			case NONE:
			case ACCEPT:
				t.Accept = sId
			case ERROR:
				row[i] = -nProd
			case SHIFT:
				row[i] = s.Goto[i].Id
			default:
				if act != s.Default {
					row[i] = int(-act)
				}
			}
		}
		t.Reduce[sId] = int(s.Default)
		tab = append(tab, row)
	}
	for i := nT; i < nT+nNt; i++ {
		row := make([]int, nState)
		defCnt := make([]int, nState)
		def := 0
		for sId, s := range g.StTab.All {
			act := 0
			if ns := s.Goto[i]; ns != nil {
				act = ns.Id
			}
			if act != 0 {
				row[sId] = act
				defCnt[act]++
				if defCnt[act] > defCnt[def] {
					def = act
				}
			}
		}
		t.Goto[i-nT] = def
		if def != 0 {
			for i, v := range row {
				if v == def {
					row[i] = 0
				}
			}
		}
		tab = append(tab, row)
	}
	var r []int
	t.Action, t.Check, r = tab.Pack()
	for i, v := range t.Action {
		if v == -nProd {
			t.Action[i] = 0
		}
	}
	t.Pact = r[:nState]
	t.Pgoto = r[nState:]
	g.pt = t
}
