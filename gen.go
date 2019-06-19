package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type LRGen struct {
	yyLex
	Token    int
	yylval   yySymType
	StTab
	Stat     io.Writer
	Prefix   string // "yy" in yyParser
	Union    string // fields in yySymType
	pt       ParTab
	currPrec int
}

func (g *LRGen) Fatal(s string) {
	g.yyLex.Error(s)
	os.Exit(1)
}

func (g *LRGen) GenAll() {
	g.sy.GenFirst()
	g.sy.GenReducible(g.pr.All[1].Lhs)
	g.StTab.GenAll()
	g.genParTab()
}

func (g *LRGen) Dump() {
	stat := bufio.NewWriter(g.Stat)
	defer stat.Flush()
	g.StTab.Dump(stat)
	g.dumpStats(stat)
	out := bufio.NewWriter(g.Out)
	defer out.Flush()
	g.dumpSymbols(out)
	g.dumpTable(out)
	g.dumpParser(out)
}

func (g *LRGen) Rules() []*Prod { return g.pr.All }

func (g *LRGen) Run() {
	g.sy.Lookup("$end")
	g.sy.Lookup("error")
	g.sy.Lookup("$unk")
	// augment start symbol
	dollar := g.sy.All[0]
	augment := g.pr.NewProd(dollar, []*Symbol{dollar, dollar})
	g.parse()
	if len(g.pr.All) <= 1 {
		g.Fatal("no rules defined")
	}
	augment.Rhs[0] = g.pr.All[1].Lhs
	g.GenAll()
	g.Dump()
}

func (g *LRGen) dumpStats(w io.Writer) {
	nT := g.sy.NtBase
	nNt := g.sy.CountNt()
	nState := len(g.StTab.All)
	nProd := len(g.pr.All)
	sr, rr := 0, 0 // conflicts
	for _, s := range g.StTab.All {
		for _, conf := range s.Conf {
			if s.Action[conf.Sym.Id] == SHIFT {
				sr++
			} else {
				rr++
			}
		}
	}
	notreduced := 0
	for _, prod := range g.pr.All[1:] {
		if !prod.Reducible {
			if notreduced == 0 {
				fmt.Fprintln(w, "Useless rules:")
			}
			fmt.Fprintf(w, "(%d)\t%v\n", prod.Id, Item{prod, 0})
			notreduced++
		}
	}
	if notreduced > 0 {
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, nT, "terminals,", nNt, "nonterminals")
	fmt.Fprintln(w, nProd, "productions,", nState, "states")
	fmt.Fprintln(w, g.pt.Size(), "entries in parsing table")
	fmt.Fprintln(w, sr, "shift/reduce conflicts")
	fmt.Fprintln(w, rr, "reduce/reduce conflicts")

	// the following reports to stdout
	if total := sr + rr; total > 0 {
		fmt.Println(total, "conflicts.")
	}
	if notreduced > 0 {
		fmt.Println(notreduced, "rule(s) not reduced.")
	}
}

func (g *LRGen) dumpSymbols(w *bufio.Writer) {
	w.WriteString("// Tokens\nconst (\n\t_ = iota + 2 // eof, error, unk\n")
	for _, sym := range g.sy.All[3:g.sy.NtBase] {
		fmt.Fprintf(w, "\t%v\n", sym)
	}
	w.WriteString(")\n\n")
	fmt.Fprintf(w, "var %sName = []string{\n", g.Prefix)
	for _, sym := range g.sy.AllT() {
		fmt.Fprintf(w, "\t%q,\n", sym)
	}
	w.WriteString("}\n\n")
}

func (g *LRGen) dumpTable(w *bufio.Writer) {
	dump := func(name string, arr []int) {
		fmt.Fprintf(w, "var %s%s = [...]int{", g.Prefix, name)
		for i, v := range arr {
			if i%10 == 0 {
				w.WriteString("\n\t")
			} else {
				w.WriteString(" ")
			}
			fmt.Fprintf(w, "%d,", v)
		}
		w.WriteString("\n}\n\n")
	}
	t := g.pt
	fmt.Fprintf(w, "const %sAccept = %d\n", g.Prefix, t.Accept)
	fmt.Fprintf(w, "const %sLast = %d\n\n", g.Prefix, g.sy.NtBase)
	w.WriteString("// Parse tables\n")
	dump("R1", t.R1)
	dump("R2", t.R2)
	dump("Reduce", t.Reduce)
	dump("Goto", t.Goto)
	dump("Action", t.Action)
	dump("Check", t.Check)
	dump("Pact", t.Pact)
	dump("Pgoto", t.Pgoto)
}

func (g *LRGen) dumpParser(w *bufio.Writer) {
	err := yyTmpl.Execute(w, g)
	if err != nil {
		g.Fatal(err.Error())
	}
}

