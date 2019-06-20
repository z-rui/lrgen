package main

//go:generate lexgen lex.l

const (
	KEYWORD = iota + 3
	IDENT
	TYPENAME
	CODEFRAG
	MARK
	COLON
	SEMI
	PIPE
)

type yySymType struct {
	s string
}

func (g *LRGen) getToken() {
	g.Token = g.Lex(&g.yylval)
}

func (g *LRGen) syntaxError() {
	g.Fatal("syntax error")
}

func (g *LRGen) parse() {
	g.yyLex.Start = 1 // start in copy mode
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
			s := g.yylval.s
			g.getToken()
			switch s {
			case "union":
				if g.Token != CODEFRAG {
					g.syntaxError()
				}
				g.Union = g.yylval.s
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
		typename = g.yylval.s
		g.getToken()
	}
	for g.Token == IDENT {
		sym := g.sy.Lookup(g.yylval.s)
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
		sym := g.sy.Lookup(g.yylval.s)
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
			lhs = g.sy.Lookup(g.yylval.s)
			g.getToken()
			if g.Token != COLON {
				g.syntaxError()
			}
			g.getToken()
			if !g.sy.IsNt(lhs) {
				g.Fatal("lhs must be nonterminal")
			}
		case PIPE: // use previous lhs
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
			sym := g.sy.Lookup(g.yylval.s)
			rhs = append(rhs, sym)
			g.getToken()
		}
		prod := g.pr.NewProd(lhs, rhs)
		if g.Token == KEYWORD && g.yylval.s == "prec" {
			g.getToken()
			if g.Token != IDENT {
				g.syntaxError()
			}
			prod.PrecSym = g.sy.Lookup(g.yylval.s)
			g.getToken()
		}
		if g.Token == CODEFRAG {
			prod.Semant = g.yylval.s
			g.getToken()
		}
		if g.Token == SEMI {
			lhs = nil
			g.getToken()
		}
	}
}
