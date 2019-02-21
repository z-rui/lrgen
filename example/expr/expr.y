package main

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"os"
)

type lex struct {
	*bufio.Reader
	lookahead byte
	major     int
	minor     interface{}
	text      []byte
}

const (
	Eof     = 0
	Unknown = 2
)

func (l *lex) next() byte {
	var c byte
	var err error
	if l.lookahead == 0 {
		c, err = l.ReadByte()
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
	case '*':
		l.major = TIMES
	case '/':
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
		l.major = NL
	case ' ', '\t', '\r', '\f', '\v':
		goto reinput
	default:
		l.major = Unknown
	}
}

func (l *lex) scanNum(c byte) {
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
	if len(l.text) > 0 {
		fmt.Printf("near %q: ", l.text)
	}
	fmt.Println(v...)
}

func main() {
	var p yyParser
	yyDebug = 0
	p.Reader = bufio.NewReader(os.Stdin)
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
	yyp.lex.error("unexpected token", yyName[yymajor])
}

%%

input:
	/* epsilon */
|	input line NL { yyp.ErrOk() }
;

line:
	expr
	{
		var s string
		if $1.IsInt() {
			s = $1.Num().String()
		} else {
			s = $1.String()
		}
		fmt.Println(s)
	}
|	error
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
