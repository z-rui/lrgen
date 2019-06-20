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
	prefix  string
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

func (yy1lex *yy1Lex) Lex(yy1lval *yy1SymType) int {
	const (
		INITIAL = iota
	)
	BEGIN := func(s int32) int32 {
		yy1lex.Start, s = s, yy1lex.Start
		return s
	}
	_ = BEGIN
	yy1less := func(n int) {
		n += yy1lex.s
		yy1lex.t = n
		yy1lex.r = n
	}
	_ = yy1less
	yy1more := func() { yy1lex.t = yy1lex.s }
	_ = yy1more

yy1S0:
	yy1lex.Pos += yy1lex.t - yy1lex.s
	yy1lex.s = yy1lex.t
	yy1acc := -1
	yy1lex.t = yy1lex.r
	yy1c := yy1lex.Start
	if '\x00' <= yy1c && yy1c <= '\x00' {
		goto yy1S1
	}

	goto yy1fin
yy1S1:
	yy1c = yy1lex.next()
	if yy1c < '%' {
		if yy1c < '$' {
			if '\x00' <= yy1c {
				goto yy1S2
			}
		} else {
			goto yy1S3
		}
	} else if yy1c < '@' {
		goto yy1S2
	} else if yy1c < 'A' {
		goto yy1S3
	} else if yy1c <= '\U0010ffff' {
		goto yy1S2
	}

	goto yy1fin
yy1S2:
	yy1acc = 1
	yy1lex.t = yy1lex.r
	yy1c = yy1lex.next()
	if yy1c < '%' {
		if '\x00' <= yy1c && yy1c <= '#' {
			goto yy1S2
		}
	} else if yy1c < 'A' {
		if yy1c <= '?' {
			goto yy1S2
		}
	} else if yy1c <= '\U0010ffff' {
		goto yy1S2
	}

	goto yy1fin
yy1S3:
	yy1acc = 1
	yy1lex.t = yy1lex.r
	yy1c = yy1lex.next()
	if yy1c < '0' {
		if '$' <= yy1c && yy1c <= '$' {
			goto yy1S4
		}
	} else if yy1c <= '9' {
		goto yy1S5
	}

	goto yy1fin
yy1S4:
	yy1acc = 0
	yy1lex.t = yy1lex.r

	goto yy1fin
yy1S5:
	yy1acc = 0
	yy1lex.t = yy1lex.r
	yy1c = yy1lex.next()
	if '0' <= yy1c && yy1c <= '9' {
		goto yy1S5
	}

	goto yy1fin

yy1fin:
	yy1lex.r = yy1lex.t // put back read-ahead bytes
	yy1text := yy1lex.buf[yy1lex.s:yy1lex.r]
	if len(yy1text) == 0 {
		if yy1lex.err != nil {
			return 0
		}
		panic("scanner is jammed")
	}
	switch yy1acc {
	case 0:
		{
			yy := yy1lex.prefix
			p := yy1lex.prod
			ty := ""

			if yy1text[1] == '$' {
				fmt.Fprintf(yy1lex.wr, "%sval", yy)
				ty = p.Lhs.Type
			} else {
				n, _ := strconv.Atoi(string(yy1text[1:]))
				n--
				fmt.Fprintf(yy1lex.wr, "%sD[%d]", yy, n)
				if 0 <= n && n < len(p.Rhs) && p.Rhs[n].Type != "" {
					ty = p.Rhs[n].Type
				}
			}
			if yy1text[0] == '@' {
				ty = yy + "pos"
			}
			if ty != "" {
				fmt.Fprintf(yy1lex.wr, ".%s", ty)
			}
		}
	case 1:
		{
			yy1lex.wr.Write(yy1text)
		}
	}
	goto yy1S0
}
