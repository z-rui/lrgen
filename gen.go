package main

import (
	"bytes"
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
	base     string // base type of yyParser
	errCode  string // error handler code
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
	const tmpl1 = `type $$Stack struct {
	s int         // state
	v interface{} // semantic value
}

type $$Parser struct {
`
	const tmpl2 = `
	state int
	errSt int // error state
	stack []$$Stack
}

var $$Debug = 0 // debug info from parser

// ParseToken runs the state machine for a single token.
// Returns true if no error occurs.
func ($$ *$$Parser) ParseToken($$major int, $$minor interface{}) bool {
	if $$Debug >= 1 {
		println("In state", $$.state)
		if $$Debug >= 2 {
			println("\tINPUT token", $$Name[$$major])
		}
	}
	for {
		var $$Val interface{}
		// look up shift or reduce
		$$n := int($$Pact[$$.state])
		if $$major >= 0 {
			$$n += $$major
		} else if $$n < len($$Action) {
			break // cannot decide without lookahead
		}
		if 0 <= $$n && $$n < len($$Action) && int($$Check[$$n]) == $$major {
			$$n = int($$Action[$$n])
		} else {
			$$n = -int($$Reduce[$$.state])
		}
		switch {
		case $$n > 0: // shift
			if $$Debug >= 1 {
				println("\tSHIFT token", $$Name[$$major])
			}
			if $$.errSt > 0 {
				$$.errSt--
			}
			$$Val = $$minor
			$$major = -1
		case $$n < 0: // reduce
			$$n = -$$n
			if $$Debug >= 1 {
				println("\tREDUCE rule", $$n)
			}
			$$t := len($$.stack) - int($$R2[$$n])
			$$D := $$.stack[$$t:]
			if len($$D) > 0 { // pop items and restore state
				$$.state = $$.stack[$$t].s
				$$Val = $$.stack[$$t].v
				$$.stack = $$.stack[:$$t]
			}
			switch $$n { // Semantic actions`
	const tmpl3 = `
			}
			// look up goto
			$$t = int($$R1[$$n]) - $$Last
			$$n = int($$Pgoto[$$t]) + $$.state
			if 0 <= $$n && $$n < len($$Action) &&
				int($$Check[$$n]) == $$.state {
				$$n = int($$Action[$$n])
			} else {
				$$n = int($$Goto[$$t])
			}
		default:
			if $$major == 0 && $$.state == $$Accept {
				if $$Debug >= 1 {
					println("\tACCEPT!")
				}
				return true
			}
			switch $$.errSt {
			case 0: // new error
				if $$Debug >= 1 {
					println("\tERROR! Unexpected", $$Name[yymajor])
				}
`
	const tmpl4 = `
				fallthrough
			case 1, 2: // partially recovered error
				for { // pop states until error can be shifted
					$$n = int($$Pact[$$.state]) + 1
					if 0 <= $$n && $$n < len($$Action) && $$Check[$$n] == 1 {
						$$n = $$Action[$$n]
						if $$n > 0 {
							break
						}
					}
					if len($$.stack) == 0 {
						if $$Debug >= 2 {
							println("\tCannot shift error")
						}
						return false
					}
					if $$Debug >= 2 {
						println("\tPopping state", $$.state)
					}
					$$.state = $$.stack[len($$.stack)-1].s
					$$.stack = $$.stack[:len($$.stack)-1]
				}
				$$.errSt = 3
				if $$Debug >= 2 {
					println("\tSHIFT token error")
				}
				$$Val = nil
			default: // still waiting for valid tokens
				if $$Debug >= 1 {
					println("\tDISCARD token", $$Name[$$major])
				}
				return $$major != 0
			}
		}
		if $$Debug >= 2 {
			println("\tGOTO state", $$n)
		}
		$$.stack = append($$.stack, $$Stack{$$.state, $$Val})
		$$.state = $$n
	}
	return true
}

// Result returns the result on a sucessful parse.
func (p *$$Parser) Result() interface{} {
	if len(p.stack) == 0 {
		return nil
	}
	return p.stack[0].v
}

// ErrOk clears the error state of the parser.
func (p *$$Parser) ErrOk() {
	p.errSt = 0
}
`
	io.WriteString(w, strings.Replace(tmpl1, "$$", g.Prefix, -1))
	io.WriteString(w, g.base)
	io.WriteString(w, strings.Replace(tmpl2, "$$", g.Prefix, -1))
	g.dumpSemant(w)
	io.WriteString(w, strings.Replace(tmpl3, "$$", g.Prefix, -1))
	io.WriteString(w, g.errCode)
	io.WriteString(w, strings.Replace(tmpl4, "$$", g.Prefix, -1))
}

func (g *LRGen) dumpSemant(w io.Writer) {
	dump := func(prod *Prod) {
		var state int
		const (
			OUT = iota
			DOLLAR
			NUM
		)
		nRhs := len(prod.Rhs)
		used := make([]bool, nRhs+1)
		buf := &bytes.Buffer{}
		var n int // the number after $
		putDollar := func(i int) {
			fmt.Fprintf(buf, "%sD%d", g.Prefix, i)
			used[i] = true
		}
		for i := range prod.Semant {
			ch := prod.Semant[i]
			switch state {
			case OUT:
				switch ch {
				case '$':
					state = DOLLAR
				default:
					buf.WriteByte(ch)
				}
			case DOLLAR:
				switch ch {
				case '$':
					putDollar(0)
					state = OUT
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					n = int(ch - '0')
					state = NUM
				default:
					buf.WriteByte('$')
					buf.WriteByte(ch)
					state = OUT
				}
			case NUM:
				switch ch {
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					n = n*10 + int(ch-'0')
					state = NUM
				default:
					putDollar(n)
					buf.WriteByte(ch)
					state = OUT
				}
			}
		}
		switch state {
		case DOLLAR:
			buf.WriteByte('$')
		case NUM:
			putDollar(n)
		}
		if used[0] {
			typ := prod.Lhs.Type
			if typ == "" {
				typ = "interface{}"
			}
			fmt.Fprintf(w, "var %sD0 %s\n", g.Prefix, typ)
		}
		for i := 1; i <= nRhs; i++ {
			if used[i] {
				fmt.Fprintf(w, "%sD%d := %sD[%d].v", g.Prefix, i, g.Prefix, i-1)
				if typ := prod.Rhs[i-1].Type; typ != "" {
					fmt.Fprintf(w, ".(%s)\n", typ)
				} else {
					fmt.Fprintln(w)
				}
			}
		}
		buf.WriteTo(w)
		if used[0] {
			io.WriteString(w, "\nyyVal = yyD0\n")
		}
	}
	for i, prod := range g.pr.All {
		if prod.Semant != NoSemant {
			fmt.Fprintf(w, "\ncase %d:\n", i)
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
}
