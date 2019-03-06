package main

import (
	"io"
)

func (g *LRGen) copyUntilMark() {
	for {
		line, err := g.In.ReadString('\n')
		g.line++
		if line == "%%\n" || err == io.EOF {
			break
		}
		if err != nil {
			g.Fatal(err.Error())
		}
		io.WriteString(g.Out, line)
	}
}

func (g *LRGen) parse() {
	g.copyUntilMark() // prologue
	g.getToken()
	g.parseTokDef()
	g.parseRuleDef()
	if g.Token != 0 {
		g.syntaxError()
	}
}

func (g *LRGen) parseTokDef() {
	for {
		switch g.Token {
		case KEYWORD:
			s := string(g.Text)
			g.getToken()
			switch s {
			case "union":
				if g.Token != CODEFRAG {
					g.syntaxError()
				}
				g.union = string(g.Text)
				g.getToken()
			case "token":
				if g.sy.NtBase > 0 {
					g.Fatal("%token not allowed after %type")
				}
				g.parseTypes()
			case "type":
				if g.sy.NtBase == 0 {
					g.sy.StartNt()
				}
				g.parseTypes()
			case "left":
				g.parsePrec(LEFT)
			case "right":
				g.parsePrec(RIGHT)
			case "nonassoc":
				g.parsePrec(NONASSOC)
			default:
				g.syntaxError()
			}
		case MARK:
			g.getToken()
			fallthrough
		case 0:
			return
		default:
			g.syntaxError()
		}
	}
}

func (g *LRGen) parseTypes() {
	var typename string
	if g.Token == TYPENAME {
		typename = string(g.Text)
		g.getToken()
	}
	for g.Token == IDENT {
		sym := g.sy.Lookup(string(g.Text))
		sym.Type = typename
		g.getToken()
	}
}

func (g *LRGen) parsePrec(assoc Assoc) {
	if g.sy.NtBase > 0 {
		g.Fatal("%left, %right or %nonassoc not allowed after %type")
	}
	g.currPrec++
	for g.Token == IDENT {
		sym := g.sy.Lookup(string(g.Text))
		sym.Assoc = assoc
		sym.Prec = g.currPrec
		g.getToken()
	}
}

func (g *LRGen) parseRuleDef() {
	if g.sy.NtBase == 0 {
		g.sy.StartNt()
	}
	var lhs *Symbol
	for { // rule definitions
		var rhs []*Symbol
		switch g.Token {
		case IDENT:
			lhs = g.sy.Lookup(string(g.Text))
			g.getToken()
			if g.Token != ':' {
				g.syntaxError()
			}
			g.getToken()
			if !g.sy.IsNt(lhs) {
				g.Fatal("lhs must be nonterminal")
			}
		case '|': // use previous lhs
			if lhs == nil {
				g.syntaxError()
			}
			g.getToken()
		case 0:
			return
		default:
			g.syntaxError()
		}
		for g.Token == IDENT {
			sym := g.sy.Lookup(string(g.Text))
			rhs = append(rhs, sym)
			g.getToken()
		}
		prod := g.pr.NewProd(lhs, rhs)
		if g.Token == KEYWORD && string(g.Text) == "prec" {
			g.getToken()
			if g.Token != IDENT {
				g.syntaxError()
			}
			prod.PrecSym = g.sy.Lookup(string(g.Text))
			g.getToken()
		}
		if g.Token == CODEFRAG {
			prod.Semant = string(g.Text)
			g.getToken()
		}
		if g.Token == ';' {
			lhs = nil
			g.getToken()
		}
	}
}
