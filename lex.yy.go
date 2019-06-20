// Generated from lex.l.  DO NOT EDIT.

package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"unicode/utf8"
	// start condition
)

type yyLex struct {
	Start   int32
	Path    string
	Pos     int // position of current token
	In      io.Reader
	buf     []byte
	linePos []int
	s, t    int // buf[s:t] == token to be flushed
	r, w    int // buf[r:w] == buffered text
	err     error
	Out     io.Writer
	part    int
}

func (l *yyLex) Init(r io.Reader) *yyLex {
	l.Start = 0
	l.Pos = 0
	l.In = r
	l.buf = make([]byte, 4096)
	l.s, l.t, l.r, l.w = 0, 0, 0, 0
	l.err = nil
	return l
}

func (l *yyLex) ErrorAt(pos int, s string, v ...interface{}) {
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

func (l *yyLex) Error(s string) {
	l.ErrorAt(l.Pos, s)
}

func (l *yyLex) fill() {
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

func (l *yyLex) next() rune {
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

func (yylex *yyLex) Lex(yylval *yySymType) int {
	const (
		INITIAL = iota
		copymode
		codefrag
	)
	BEGIN := func(s int32) int32 {
		yylex.Start, s = s, yylex.Start
		return s
	}
	_ = BEGIN
	yyless := func(n int) {
		n += yylex.s
		yylex.t = n
		yylex.r = n
	}
	_ = yyless
	yymore := func() { yylex.t = yylex.s }
	_ = yymore
	level := 0

yyS0:
	yylex.Pos += yylex.t - yylex.s
	yylex.s = yylex.t
	yyacc := -1
	yylex.t = yylex.r
	yyc := yylex.Start
	if yyc < '\x01' {
		if '\x00' <= yyc {
			goto yyS1
		}
	} else if yyc < '\x02' {
		goto yyS2
	} else if yyc <= '\x02' {
		goto yyS3
	}

	goto yyfin
yyS1:
	yyc = yylex.next()
	if yyc < ';' {
		if yyc < '%' {
			if yyc < '\x0e' {
				if yyc < '\t' {
					if '\x00' <= yyc {
						goto yyS4
					}
				} else {
					goto yyS5
				}
			} else if yyc < ' ' {
				goto yyS4
			} else if yyc < '!' {
				goto yyS5
			} else {
				goto yyS4
			}
		} else if yyc < '/' {
			if yyc < '&' {
				goto yyS6
			} else {
				goto yyS4
			}
		} else if yyc < '0' {
			goto yyS7
		} else if yyc < ':' {
			goto yyS4
		} else {
			goto yyS8
		}
	} else if yyc < '_' {
		if yyc < '=' {
			if yyc < '<' {
				goto yyS9
			} else {
				goto yyS10
			}
		} else if yyc < 'A' {
			goto yyS4
		} else if yyc < '[' {
			goto yyS11
		} else {
			goto yyS4
		}
	} else if yyc < '{' {
		if yyc < '`' {
			goto yyS11
		} else if yyc < 'a' {
			goto yyS4
		} else {
			goto yyS11
		}
	} else if yyc < '|' {
		goto yyS12
	} else if yyc < '}' {
		goto yyS13
	} else if yyc <= '\U0010ffff' {
		goto yyS4
	}

	goto yyfin
yyS2:
	yyc = yylex.next()
	if yyc < '\v' {
		if yyc < '\n' {
			if '\x00' <= yyc {
				goto yyS14
			}
		} else {
			goto yyS15
		}
	} else if yyc < '%' {
		goto yyS14
	} else if yyc < '&' {
		goto yyS16
	} else if yyc <= '\U0010ffff' {
		goto yyS14
	}

	goto yyfin
yyS3:
	yyc = yylex.next()
	if yyc < '|' {
		if yyc < '{' {
			if '\x00' <= yyc {
				goto yyS17
			}
		} else {
			goto yyS18
		}
	} else if yyc < '}' {
		goto yyS17
	} else if yyc < '~' {
		goto yyS19
	} else if yyc <= '\U0010ffff' {
		goto yyS17
	}

	goto yyfin
yyS4:
	yyacc = 16
	yylex.t = yylex.r

	goto yyfin
yyS5:
	yyacc = 10
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < ' ' {
		if '\t' <= yyc && yyc <= '\r' {
			goto yyS5
		}
	} else if yyc <= ' ' {
		goto yyS5
	}

	goto yyfin
yyS6:
	yyacc = 16
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '_' {
		if yyc < 'A' {
			if '%' <= yyc && yyc <= '%' {
				goto yyS20
			}
		} else if yyc <= 'Z' {
			goto yyS21
		}
	} else if yyc < 'a' {
		if yyc <= '_' {
			goto yyS21
		}
	} else if yyc <= 'z' {
		goto yyS21
	}

	goto yyfin
yyS7:
	yyacc = 16
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '/' {
		if '*' <= yyc && yyc <= '*' {
			goto yyS22
		}
	} else if yyc <= '/' {
		goto yyS23
	}

	goto yyfin
yyS8:
	yyacc = 1
	yylex.t = yylex.r

	goto yyfin
yyS9:
	yyacc = 2
	yylex.t = yylex.r

	goto yyfin
yyS10:
	yyacc = 16
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '\v' {
		if '\x00' <= yyc && yyc <= '\t' {
			goto yyS24
		}
	} else if yyc < '?' {
		if yyc <= '=' {
			goto yyS24
		}
	} else if yyc <= '\U0010ffff' {
		goto yyS24
	}

	goto yyfin
yyS11:
	yyacc = 4
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '_' {
		if yyc < 'A' {
			if '0' <= yyc && yyc <= '9' {
				goto yyS11
			}
		} else if yyc <= 'Z' {
			goto yyS11
		}
	} else if yyc < 'a' {
		if yyc <= '_' {
			goto yyS11
		}
	} else if yyc <= 'z' {
		goto yyS11
	}

	goto yyfin
yyS12:
	yyacc = 7
	yylex.t = yylex.r

	goto yyfin
yyS13:
	yyacc = 3
	yylex.t = yylex.r

	goto yyfin
yyS14:
	yyacc = 16
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '\n' {
		if '\x00' <= yyc {
			goto yyS25
		}
	} else if yyc < '\v' {
		goto yyS15
	} else if yyc <= '\U0010ffff' {
		goto yyS25
	}

	goto yyfin
yyS15:
	yyacc = 15
	yylex.t = yylex.r

	goto yyfin
yyS16:
	yyacc = 16
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '\v' {
		if yyc < '\n' {
			if '\x00' <= yyc {
				goto yyS25
			}
		} else {
			goto yyS15
		}
	} else if yyc < '%' {
		goto yyS25
	} else if yyc < '&' {
		goto yyS26
	} else if yyc <= '\U0010ffff' {
		goto yyS25
	}

	goto yyfin
yyS17:
	yyacc = 13
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '|' {
		if '\x00' <= yyc && yyc <= 'z' {
			goto yyS17
		}
	} else if yyc < '~' {
		if yyc <= '|' {
			goto yyS17
		}
	} else if yyc <= '\U0010ffff' {
		goto yyS17
	}

	goto yyfin
yyS18:
	yyacc = 11
	yylex.t = yylex.r

	goto yyfin
yyS19:
	yyacc = 12
	yylex.t = yylex.r

	goto yyfin
yyS20:
	yyacc = 0
	yylex.t = yylex.r

	goto yyfin
yyS21:
	yyacc = 5
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '_' {
		if yyc < 'A' {
			if '0' <= yyc && yyc <= '9' {
				goto yyS21
			}
		} else if yyc <= 'Z' {
			goto yyS21
		}
	} else if yyc < 'a' {
		if yyc <= '_' {
			goto yyS21
		}
	} else if yyc <= 'z' {
		goto yyS21
	}

	goto yyfin
yyS22:
	yyc = yylex.next()
	if yyc < '*' {
		if '\x00' <= yyc {
			goto yyS22
		}
	} else if yyc < '+' {
		goto yyS27
	} else if yyc <= '\U0010ffff' {
		goto yyS22
	}

	goto yyfin
yyS23:
	yyacc = 8
	yylex.t = yylex.r
	yyc = yylex.next()
	if yyc < '\v' {
		if '\x00' <= yyc && yyc <= '\t' {
			goto yyS23
		}
	} else if yyc <= '\U0010ffff' {
		goto yyS23
	}

	goto yyfin
yyS24:
	yyc = yylex.next()
	if yyc < '>' {
		if yyc < '\v' {
			if '\x00' <= yyc && yyc <= '\t' {
				goto yyS24
			}
		} else {
			goto yyS24
		}
	} else if yyc < '?' {
		goto yyS28
	} else if yyc <= '\U0010ffff' {
		goto yyS24
	}

	goto yyfin
yyS25:
	yyc = yylex.next()
	if yyc < '\n' {
		if '\x00' <= yyc {
			goto yyS25
		}
	} else if yyc < '\v' {
		goto yyS15
	} else if yyc <= '\U0010ffff' {
		goto yyS25
	}

	goto yyfin
yyS26:
	yyc = yylex.next()
	if yyc < '\n' {
		if '\x00' <= yyc {
			goto yyS25
		}
	} else if yyc < '\v' {
		goto yyS29
	} else if yyc <= '\U0010ffff' {
		goto yyS25
	}

	goto yyfin
yyS27:
	yyc = yylex.next()
	if yyc < '+' {
		if yyc < '*' {
			if '\x00' <= yyc {
				goto yyS22
			}
		} else {
			goto yyS27
		}
	} else if yyc < '/' {
		goto yyS22
	} else if yyc < '0' {
		goto yyS30
	} else if yyc <= '\U0010ffff' {
		goto yyS22
	}

	goto yyfin
yyS28:
	yyacc = 6
	yylex.t = yylex.r

	goto yyfin
yyS29:
	yyacc = 14
	yylex.t = yylex.r

	goto yyfin
yyS30:
	yyacc = 9
	yylex.t = yylex.r

	goto yyfin

yyfin:
	yylex.r = yylex.t // put back read-ahead bytes
	yytext := yylex.buf[yylex.s:yylex.r]
	if len(yytext) == 0 {
		if yylex.err != nil {
			return 0
		}
		panic("scanner is jammed")
	}
	switch yyacc {
	case 0:
		return MARK
	case 1:
		return COLON
	case 2:
		return SEMI
	case 3:
		return PIPE
	case 4:
		{
			yylval.s = string(yytext)
			return IDENT
		}
	case 5:
		{
			yylval.s = string(yytext[1:])
			return KEYWORD
		}
	case 6:
		{
			yylval.s = string(yytext[1 : len(yytext)-1])
			return TYPENAME
		}
	case 7:
		level = 1
		BEGIN(codefrag)
	case 8:
		// inline comment
	case 9:
		/* multi-line comment */
	case 10:
		// white spaces
	case 11:
		level++
		yymore()
	case 12:
		{
			level--
			if level == 0 {
				yylval.s = string(yytext[:len(yytext)-1])
				BEGIN(INITIAL)
				return CODEFRAG
			} else {
				yymore()
			}
		}
	case 13:
		yymore()
	case 14:
		BEGIN(INITIAL)
	case 15:
		yylex.Out.Write(yytext)
	case 16:
		{
			yylex.Error("invalid character")
		}
	}
	goto yyS0
}
