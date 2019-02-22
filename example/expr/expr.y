package main

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
)

type lex struct {
	*bufio.Reader
	filename  string
	lineno    int
	lookahead rune
	major     int
	minor     interface{}
	text      []rune
}

const (
	Eof     = 0
	Unknown = 2
)

func (l *lex) next() rune {
	var c rune
	var err error
	if l.lookahead == 0 {
		c, _, err = l.ReadRune()
		if err == io.EOF {
			c = Eof
		} else if err != nil {
			l.error(err)
			os.Exit(1)
		}
	} else {
		c = l.lookahead
		l.lookahead = 0
	}
	return c
}

func (l *lex) getToken() {
reinput:
	l.text = l.text[:0]
	c := l.next()
	l.text = append(l.text, c)
	switch c {
	case Eof:
		l.major = Eof
	case '+':
		l.major = PLUS
	case '-':
		l.major = MINUS
	case '*', '×':
		l.major = TIMES
	case '/', '÷':
		l.major = DIV
	case '(':
		l.major = LPAR
	case ')':
		l.major = RPAR
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		l.scanNum(c)
		l.major = NUM
		if r, ok := (&big.Rat{}).SetString(string(l.text)); ok {
			l.minor = r
		} else {
			l.error("bad formed number")
			l.major = Unknown
		}
	case '\n':
		l.lineno++
		l.major = NL
	case ' ', '\t', '\r', '\f', '\v':
		goto reinput
	default:
		l.major = Unknown
	}
}

func (l *lex) scanNum(c rune) {
	for {
		if c = l.next(); '0' <= c && c <= '9' || c == '.' {
			l.text = append(l.text, c)
		} else {
			break
		}
	}
	l.lookahead = c
}

func (l *lex) error(v ...interface{}) {
	fmt.Printf("%s:%d: ", l.filename, l.lineno)
	fmt.Println(v...)
}

func main() {
	var p yyParser
	yyDebug = 1
	p.Reader = bufio.NewReader(os.Stdin)
	p.filename = "<stdin>"
	p.lineno = 1
	for {
		p.getToken()
		if !p.ParseToken(p.major, p.minor) || p.major == Eof {
			break
		}
	}
}

%%

%token <*big.Rat> NUM
%token LPAR RPAR NL

%left PLUS MINUS
%left TIMES DIV

%type <*big.Rat> expr

%base lex
%error {
	yy.error("syntax error near", strconv.Quote(string(yy.text)))
}

%%

input:
	/* epsilon */
|	input expr NL
	{
		var s string
		if $2.IsInt() {
			s = $2.Num().String()
		} else {
			s = $2.String()
		}
		fmt.Println(s)
	}
|	input error NL { yy.ErrOk() }
;

expr:
	NUM
|	LPAR expr RPAR
	{
		$$ = $2
	}
|	PLUS expr
	{
		$$ = $2
	}
|	MINUS expr
	{
		$$ = $2.Neg($2)
	}
|	expr PLUS expr
	{
		$$ = $1.Add($1, $3)
	}
|	expr MINUS expr
	{
		$$ = $1.Sub($1, $3)
	}
|	expr TIMES expr
	{
		$$ = $1.Mul($1, $3)
	}
|	expr DIV expr
	{
		$$ = $1.Quo($1, $3)
	}
;
