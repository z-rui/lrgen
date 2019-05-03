package main

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
)

type lex struct {
	*bufio.Reader
	filename string
	lineno   int
}

func (l *lex) Lex(yyval *yySymType) int {
reinput:
	c, _, err := l.ReadRune()
	if err != nil {
		return 0
	}
	switch c {
	case '+':
		return PLUS
	case '-':
		return MINUS
	case '*', '×':
		return TIMES
	case '/', '÷':
		return DIV
	case '(':
		return LPAR
	case ')':
		return RPAR
	case '\n':
		l.lineno++
		return NL
	case ' ', '\t', '\r', '\f', '\v':
		goto reinput
	default:
		var buf []rune
		for err == nil && strings.ContainsRune("0123456789.", c) {
			buf = append(buf, c)
			c, _, err = l.ReadRune()
		}
		if len(buf) > 0 {
			switch err {
			case io.EOF:
			case nil:
				l.UnreadRune()
			default:
				l.Error(err.Error())
			}
			var ok bool
			yyval.num, ok = new(big.Rat).SetString(string(buf))
			if ok {
				return NUM
			}
		}
	}
	return 2 // $unk
}

func (l *lex) Error(s string) {
	fmt.Printf("%s:%d: %s\n", l.filename, l.lineno, s)
}

func main() {
	yyParse(&lex{
		Reader:   bufio.NewReader(os.Stdin),
		filename: "<stdin>",
		lineno:   1,
	})
}

%%

%union {
	num *big.Rat
}

%token <num> NUM
%token LPAR RPAR NL

%left PLUS MINUS
%left TIMES DIV

%type <num> expr

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
|	input error NL { yyerror = 0 }
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
