package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type LRGen struct {
	Lexer
	StTab
	Out      io.Writer
	Stat     io.Writer
	Prefix   string // "yy" in yyParser
	union    string // fields in yySymType
	pt       ParTab
	currPrec int
}

func (g *LRGen) GenAll() {
	g.sy.GenFirst()
	g.StTab.GenAll()
	g.genParTab()
}

func (g *LRGen) Dump() {
	g.StTab.Dump(g.Stat)
	g.dumpStats(g.Stat)
	g.dumpSymbols(g.Out)
	g.dumpTable(g.Out)
	g.dumpParser(g.Out)
}

func (g *LRGen) Run() {
	g.sy.Lookup("$")
	g.sy.Lookup("error")
	g.sy.Lookup("(unknown)")
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
	fmt.Fprintln(w, nT, "terminals,", nNt, "nonterminals")
	fmt.Fprintln(w, nProd, "productions,", nState, "states")
	fmt.Fprintln(w, g.pt.Size(), "entries in parsing table")
	fmt.Fprintln(w, sr, "shift/reduce conflicts")
	fmt.Fprintln(w, rr, "reduce/reduce conflicts")

	// the following reports to stdout
	if total := sr + rr; total > 0 {
		fmt.Println(total, "conflicts.")
	}
	for _, sym := range g.sy.AllNt() {
		if len(sym.LhsProd) == 0 && sym.RhsProd > 0 {
			fmt.Println(sym, "undefined")
		}
	}
}

func (g *LRGen) dumpSymbols(w io.Writer) {
	io.WriteString(w, "// Tokens\nconst (\n\t_ = iota + 2 // eof, error, unk\n")
	for _, sym := range g.sy.All[3:g.sy.NtBase] {
		fmt.Fprintf(w, "\t%v\n", sym)
	}
	io.WriteString(w, ")\n\n")
	fmt.Fprintf(w, "var %sName = []string{\n", g.Prefix)
	for _, sym := range g.sy.AllT() {
		fmt.Fprintf(w, "\t%q,\n", sym)
	}
	io.WriteString(w, "}\n\n")
}

func (g *LRGen) dumpTable(w io.Writer) {
	dump := func(name string, arr []int) {
		fmt.Fprintf(w, "var %s%s = [...]int{", g.Prefix, name)
		for i, v := range arr {
			if i%10 == 0 {
				io.WriteString(w, "\n\t")
			} else {
				io.WriteString(w, " ")
			}
			fmt.Fprintf(w, "%d,", v)
		}
		io.WriteString(w, "\n}\n\n")
	}
	t := g.pt
	fmt.Fprintf(w, "const %sAccept = %d\n", g.Prefix, t.Accept)
	fmt.Fprintf(w, "const %sLast = %d\n\n", g.Prefix, g.sy.NtBase)
	io.WriteString(w, "// Parse tables\n")
	dump("R1", t.R1)
	dump("R2", t.R2)
	dump("Reduce", t.Reduce)
	dump("Goto", t.Goto)
	dump("Action", t.Action)
	dump("Check", t.Check)
	dump("Pact", t.Pact)
	dump("Pgoto", t.Pgoto)
}

func (g *LRGen) dumpParser(w io.Writer) {
	const tmpl1 = `type $$SymType struct {
	$$s int         // state
`
	const tmpl2 = `
}

type $$Lexer interface {
	Lex(*$$SymType) int
	Error($$state, $$major int, expect []int)
}

var $$Debug = 0 // debug info from parser

// $$Parse read tokens from $$lex and parses input.
// Returns true on success.
func $$Parse($$lex $$Lexer) bool {
	var (
		$$n, $$t int
		$$state  int
		$$error  int
		$$major  int = -1
		$$stack  []$$SymType
		$$D      []$$SymType
		$$val    $$SymType
	)
	goto $$action
$$stack:
	$$val.$$s = $$state
	$$stack = append($$stack, $$val)
	$$state = $$n
	if $$Debug >= 2 {
		println("\tGOTO state", $$n)
	}
$$action:
	// look up shift or reduce
	$$n = int($$Pact[$$state])
	if $$n == len($$Action) { // simple state
		goto $$default
	}
	if $$major < 0 {
		$$major = $$lex.Lex(&$$val)
		if $$Debug >= 1 {
			println("In state", $$state)
		}
		if $$Debug >= 2 {
			println("\tInput token", $$Name[$$major])
		}
	}
	$$n += $$major
	if 0 <= $$n && $$n < len($$Action) && int($$Check[$$n]) == $$major {
		$$n = int($$Action[$$n])
		if $$n <= 0 {
			$$n = -$$n
			goto $$reduce
		}
		if $$Debug >= 1 {
			println("\tSHIFT token", $$Name[$$major])
		}
		if $$error > 0 {
			$$error--
		}
		$$major = -1
		goto $$stack
	}
$$default:
	$$n = int($$Reduce[$$state])
$$reduce:
	if $$n == 0 {
		if $$major == 0 && $$state == $$Accept {
			if $$Debug >= 1 {
				println("\tACCEPT!")
			}
			return true
		}
		switch $$error {
		case 0: // new error
			if $$Debug >= 1 {
				println("\tERROR!")
			}
			var expect []int
			if $$Reduce[$$state] == 0 {
				$$n = $$Pact[$$state] + 3
				for i := 3; i < $$Last; i++ {
					if 0 <= $$n && $$n < len($$Action) && $$Check[$$n] == i && $$Action[$$n] != 0 {
						expect = append(expect, i)
					}
					$$n++
				}
			}
			$$lex.Error($$state, $$major, expect)
			fallthrough
		case 1, 2: // partially recovered error
			for { // pop states until error can be shifted
				$$n = int($$Pact[$$state]) + 1
				if 0 <= $$n && $$n < len($$Action) && $$Check[$$n] == 1 {
					$$n = $$Action[$$n]
					if $$n > 0 {
						break
					}
				}
				if len($$stack) == 0 {
					return false
				}
				if $$Debug >= 2 {
					println("\tPopping state", $$state)
				}
				$$state = $$stack[len($$stack)-1].$$s
				$$stack = $$stack[:len($$stack)-1]
			}
			$$error = 3
			if $$Debug >= 1 {
				println("\tSHIFT token error")
			}
			goto $$stack
		default: // still waiting for valid tokens
			if $$major == 0 { // no more tokens
				return false
			}
			if $$Debug >= 1 {
				println("\tDISCARD token", $$Name[$$major])
			}
			$$major = -1
			goto $$action
		}
	}
	if $$Debug >= 1 {
		println("\tREDUCE rule", $$n)
	}
	$$t = len($$stack) - int($$R2[$$n])
	$$D = $$stack[$$t:]
	if len($$D) > 0 { // pop items and restore state
		$$val = $$D[0]
		$$state = $$val.$$s
		$$stack = $$stack[:$$t]
	}
	switch $$n { // Semantic actions
`
	const tmpl3 = `
	}
	// look up goto
	$$t = int($$R1[$$n]) - $$Last
	$$n = int($$Pgoto[$$t]) + $$state
	if 0 <= $$n && $$n < len($$Action) &&
		int($$Check[$$n]) == $$state {
		$$n = int($$Action[$$n])
	} else {
		$$n = int($$Goto[$$t])
	}
	goto $$stack
}
`
	io.WriteString(w, strings.Replace(tmpl1, "$$", g.Prefix, -1))
	io.WriteString(w, g.union)
	io.WriteString(w, strings.Replace(tmpl2, "$$", g.Prefix, -1))
	g.dumpSemant(w)
	io.WriteString(w, strings.Replace(tmpl3, "$$", g.Prefix, -1))
}

func (g *LRGen) dumpSemant(w io.Writer) {
	buf := bufio.NewWriter(w)
	dump := func(prod *Prod) {
		r := strings.NewReader(prod.Semant)
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
					fmt.Fprintf(buf, "%sD[%d]", g.Prefix, n)
					if 0 <= n && n < len(prod.Rhs) && prod.Rhs[n].Type != "" {
						fmt.Fprintf(buf, ".%s", prod.Rhs[n].Type)
					}
				} else {
					ch, err = r.ReadByte()
					if err == nil {
						if ch == '$' {
							fmt.Fprintf(buf, "%sval", g.Prefix)
							if prod.Lhs.Type != "" {
								fmt.Fprintf(buf, ".%s", prod.Lhs.Type)
							}
						} else {
							buf.WriteByte('$')
							buf.WriteByte(ch)
						}
					}
				}
			default:
				buf.WriteByte(ch)
			}
		}
	}
	for i, prod := range g.pr.All {
		if prod.Semant != NoSemant {
			fmt.Fprintf(buf, "\ncase %d:\n", i)
			dump(prod)
		} else {
			t1 := prod.Lhs.Type
			t2 := ""
			if len(prod.Rhs) > 0 {
				t2 = prod.Rhs[0].Type
			}
			if t1 != "" && t1 != t2 {
				fmt.Printf("Rule %d: default action may clobber type\n", prod.Id)
			}
		}
	}
	buf.Flush()
}
