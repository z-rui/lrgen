package main

import (
	"fmt"
	"strconv"
)

type yy1SymType struct{}

%%
	prod *Prod
	prefix string
	wr io.Writer

%%

[@$]([0-9]+|"$")	{
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

[@$]|[^$@]+	{
	yy1lex.wr.Write(yy1text)
}
