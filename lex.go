package main

import (
	"fmt"
	"io"
	"os"
)

type Reader interface {
	io.Reader
	ReadByte() (byte, error)
	ReadString(byte) (string, error)
	UnreadByte() error
}

type Lexer struct {
	In    Reader
	Token int
	Text  []byte
	line  int
}

func isAlpha(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || c == '_'
}

func isAlnum(c byte) bool {
	return isAlpha(c) || '0' <= c && c <= '9'
}

const (
	KEYWORD = -(iota + 1)
	IDENT
	TYPENAME
	CODEFRAG
	MARK
)

func (l *Lexer) Fatal(s string) {
	fmt.Fprintf(os.Stderr, "%d: %s\n", l.line+1, s)
	os.Exit(1)
}

func (l *Lexer) syntaxError() {
	var s string
	if l.Token > 0 {
		s = fmt.Sprintf("%q", l.Token)
	} else if len(l.Text) > 16 {
		s = string(l.Text[:13]) + "..."
	} else {
		s = string(l.Text)
	}
	switch l.Token {
	case KEYWORD:
		s = "directive %" + s
	case IDENT:
		s = "identifer " + s
	case 0:
		s = "EOF"
	}
	l.Fatal("Syntax error near " + s)
}

func (l *Lexer) getToken() {
	l.Text = l.Text[:0]
reinput:
	c, _ := l.In.ReadByte()
	l.Token = int(c)
	switch c {
	case '%':
		c, _ = l.In.ReadByte()
		if c == '%' {
			l.Token = MARK
		} else if isAlpha(c) {
			l.scanIdent(c)
			l.Token = KEYWORD
		}
	case '<':
		l.scanType()
		l.Token = TYPENAME
	case '{':
		l.scanCode()
		l.Token = CODEFRAG
	case '/':
		c, _ = l.In.ReadByte()
		switch c {
		case '/':
			l.skipComment1()
			goto reinput
		case '*':
			l.skipComment()
			goto reinput
		default:
			l.In.UnreadByte()
		}
	case ' ', '\t', '\r', '\v', '\f':
		goto reinput
	case '\n':
		l.line++
		goto reinput
	default:
		if isAlpha(c) {
			l.scanIdent(c)
			l.Token = IDENT
		}
	}
}

func (l *Lexer) skipComment1() {
	for {
		c, err := l.In.ReadByte()
		if c == '\n' || err != nil {
			break
		}
	}
}

func (l *Lexer) skipComment() {
	for {
		c, err := l.In.ReadByte()
	check:
		if err != nil {
			l.Fatal("comment is not closed")
		}
		if c == '*' {
			c, err = l.In.ReadByte()
			if c == '/' {
				break
			}
			goto check
		}
	}
}

func (l *Lexer) scan(f func(byte) bool) (byte, error) {
	for {
		c, err := l.In.ReadByte()
		if c == '\n' {
			l.line++
		}
		if !f(c) || err != nil {
			return c, err
		}
		l.Text = append(l.Text, c)
	}
}

func (l *Lexer) scanIdent(c byte) {
	l.Text = append(l.Text, c)
	var err error
	c, err = l.scan(isAlnum)
	if err == nil {
		if c != '\n' {
			l.In.UnreadByte()
		}
	} else if err != io.EOF {
		l.Fatal(err.Error())
	}
}

func (l *Lexer) scanType() {
	c, _ := l.scan(func(c byte) bool { return c != '>' })
	if c != '>' {
		l.Fatal("'<' is not closed by '>'")
	}
}

func (l *Lexer) scanCode() {
	level := 1
	c, _ := l.scan(func(c byte) bool {
		switch c {
		case '{':
			level++
		case '}':
			level--
		}
		return level > 0
	})
	if c != '}' {
		l.Fatal("'{' is not closed by '}'")
	}
}
