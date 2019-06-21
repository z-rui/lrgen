package main

import (
	"fmt"
	"strconv"
)

type yy1SymType struct{}

%%
	prod *Prod
	wr io.Writer

%%

[@$]([0-9]+|"$")	{
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

[@$]|[^$@]+	{
	yy.wr.Write(yytext)
}
