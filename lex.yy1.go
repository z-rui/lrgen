// Generated from dump.l.  DO NOT EDIT.

package main

import (
	"io"
	"os"
	"sort"
	"unicode/utf8"
)

import (
	"fmt"
	"strconv"
)

type yy1SymType struct{}

type yy1Lex struct {
	Start   int32 // start condition
	Path    string
	Pos     int // position of current token
	In      io.Reader
	buf     []byte
	linePos []int
	s, t    int // buf[s:t] == token to be flushed
	r, w    int // buf[r:w] == buffered text
	err     error
	prod    *Prod
	wr      io.Writer
}

func (l *yy1Lex) Init(r io.Reader) *yy1Lex {
	l.Start = 0
	l.Pos = 0
	l.In = r
	l.buf = make([]byte, 4096)
	l.s, l.t, l.r, l.w = 0, 0, 0, 0
	l.err = nil
	return l
}

func (l *yy1Lex) ErrorAt(pos int, s string, v ...interface{}) {
	if len(v) > 0 {
		s = fmt.Sprintf(s, v...)
	}
	lin := sort.SearchInts(l.linePos, pos)
	col := pos
	if lin > 0 {
		col -= l.linePos[lin-1] + 1
	}
	fmt.Fprintf(os.Stderr, "%s:%d:%d: %s\n", l.Path, lin+1, col+1, s)
}

func (l *yy1Lex) Error(s string) {
	l.ErrorAt(l.Pos, s)
}

func (l *yy1Lex) fill() {
	if n := len(l.buf); l.w == n {
		if l.s+l.s <= len(l.buf) {
			// less than half usable, better extend buffer
			if n == 0 {
				n = 4096
			} else {
				n *= 2
			}
			buf := make([]byte, n)
			copy(buf, l.buf[l.s:])
			l.buf = buf
		} else {
			// shift content
			copy(l.buf, l.buf[l.s:])
		}
		l.r -= l.s
		l.w -= l.s
		l.t -= l.s
		l.s = 0
	}
	n, err := l.In.Read(l.buf[l.w:])
	// update newline positions
	for i := l.w; i < l.w+n; i++ {
		if l.buf[i] == '\n' {
			l.linePos = append(l.linePos, l.Pos+(i-l.s))
		}
	}
	l.w += n
	if err != nil {
		l.err = err
	}
}

func (l *yy1Lex) next() rune {
	for l.r+utf8.UTFMax > l.w && !utf8.FullRune(l.buf[l.r:l.w]) && l.err == nil {
		l.fill()
	}
	if l.r == l.w { // nothing is available
		return -1
	}
	c, n := rune(l.buf[l.r]), 1
	if c >= utf8.RuneSelf {
		c, n = utf8.DecodeRune(l.buf[l.r:l.w])
	}
	l.r += n
	return c
}

func (yy *yy1Lex) Lex(yylval *yy1SymType) int {
	const (
		INITIAL = iota
	)
	BEGIN := func(s int32) int32 {
		yy.Start, s = s, yy.Start
		return s
	}
	_ = BEGIN
	yyless := func(n int) {
		n += yy.s
		yy.t = n
		yy.r = n
	}
	_ = yyless
	yymore := func() { yy.t = yy.s }
	_ = yymore

yyS0:
	yy.Pos += yy.t - yy.s
	yy.s, yy.t = yy.t, yy.r
	yyacc := -1
	yyc := yy.Start
	if yyc == '\x00' {
		goto yyS1
	}
	goto yyfin

yyS1:
	yyc = yy.next()
	switch yyc {
	case '$':
		goto yyS3
	case '@':
		goto yyS3
	}
	goto yyS2

yyS2:
	yyacc = 1
	yy.t = yy.r
	yyc = yy.next()
	if yyc < '%' {
		if '\x00' <= yyc && yyc <= '#' {
			goto yyS2
		}
	} else if yyc < 'A' {
		if yyc <= '?' {
			goto yyS2
		}
	} else if yyc <= '\U0010ffff' {
		goto yyS2
	}
	goto yyfin

yyS3:
	yyacc = 1
	yy.t = yy.r
	yyc = yy.next()
	if yyc < '0' {
		if yyc == '$' {
			goto yyS4
		}
	} else if yyc <= '9' {
		goto yyS5
	}
	goto yyfin

yyS4:
	yyacc = 0
	yy.t = yy.r
	goto yyfin

yyS5:
	yyacc = 0
	yy.t = yy.r
	yyc = yy.next()
	if '0' <= yyc && yyc <= '9' {
		goto yyS5
	}
	goto yyfin

yyfin:
	yy.r = yy.t // put back read-ahead bytes
	yytext := yy.buf[yy.s:yy.r]
	yyleng := len(yytext)
	if yyleng == 0 {
		if yy.err != nil {
			return 0
		}
		panic("scanner is jammed")
	}
	switch yyacc {
	case 0:
		{
			p := yy.prod
			ty := ""

			if yytext[1] == '$' {
				fmt.Fprintf(yy.wr, "yyval")
				ty = p.Lhs.Type
			} else {
				n, _ := strconv.Atoi(string(yytext[1:]))
				n--
				fmt.Fprintf(yy.wr, "yyD[%d]", n)
				if 0 <= n && n < len(p.Rhs) && p.Rhs[n].Type != "" {
					ty = p.Rhs[n].Type
				}
			}
			if yytext[0] == '@' {
				ty = "yyp"
			}
			if ty != "" {
				fmt.Fprintf(yy.wr, ".%s", ty)
			}
		}
	case 1:
		{
			yy.wr.Write(yytext)
		}
	}
	goto yyS0
}
