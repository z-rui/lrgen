package main

import (
	"text/template"
)

var yyTmpl = template.Must(template.New("yyparse").Parse(`
{{$yy := .Prefix}}
type {{$yy}}SymType struct {
	yys, yyp int
	{{.Union}}
}

var {{$yy}}Debug = 0 // debug info from parser

// {{$yy}}Parse read tokens from {{$yy}}lex and parses input.
// Returns result on success, or nil on failure.
func {{$yy}}Parse(yy *{{$yy}}Lex) *{{$yy}}SymType {
	var (
		yyn, yyt int
		yystate  = 0
		yyerror  = 0
		yymajor  = -1
		yystack  []{{$yy}}SymType
		yyD      []{{$yy}}SymType // rhs of reduction
		yylval   {{$yy}}SymType   // lexcial value from lexer
		yyval    {{$yy}}SymType   // value to be pushed onto stack
	)
	goto yyaction
yystack:
	yyval.yys = yystate
	yystack = append(yystack, yyval)
	yystate = yyn
	if {{$yy}}Debug >= 2 {
		println("\tGOTO state", yyn)
	}
yyaction:
	// look up shift or reduce
	yyn = int({{$yy}}Pact[yystate])
	if yyn == len({{$yy}}Action) && yystate != {{$yy}}Accept { // simple state
		goto yydefault
	}
	if yymajor < 0 {
		yymajor = yy.Lex(&yylval)
		if {{$yy}}Debug >= 1 {
			println("In state", yystate)
		}
		if {{$yy}}Debug >= 2 {
			println("\tInput token", {{$yy}}Name[yymajor])
		}
	}
	if yymajor == 0 && yystate == {{$yy}}Accept {
		if {{$yy}}Debug >= 1 {
			println("\tACCEPT!")
		}
		return &yystack[0]
	}
	yyn += yymajor
	if 0 <= yyn && yyn < len({{$yy}}Action) && int({{$yy}}Check[yyn]) == yymajor {
		yyn = int({{$yy}}Action[yyn])
		if yyn <= 0 {
			yyn = -yyn
			goto yyreduce
		}
		if {{$yy}}Debug >= 1 {
			println("\tSHIFT token", {{$yy}}Name[yymajor])
		}
		if yyerror > 0 {
			yyerror--
		}
		yymajor = -1
		yyval = yylval
		yyval.yyp = yy.Pos
		goto yystack
	}
yydefault:
	yyn = int({{$yy}}Reduce[yystate])
yyreduce:
	if yyn == 0 {
		switch yyerror {
		case 0: // new error
			if {{$yy}}Debug >= 1 {
				println("\tERROR!")
			}
			msg := "unexpected " + {{$yy}}Name[yymajor]
			var expect []int
			if {{$yy}}Reduce[yystate] == 0 {
				yyn = {{$yy}}Pact[yystate] + 3
				for i := 3; i < {{$yy}}Last; i++ {
					if 0 <= yyn && yyn < len({{$yy}}Action) && {{$yy}}Check[yyn] == i && {{$yy}}Action[yyn] != 0 {
						expect = append(expect, i)
						if len(expect) > 4 {
							break
						}
					}
					yyn++
				}
			}
			if n := len(expect); 0 < n && n <= 4 {
				for i, tok := range expect {
					switch i {
					case 0:
						msg += ", expecting "
					case n-1:
						msg += " or "
					default:
						msg += ", "
					}
					msg += {{$yy}}Name[tok]
				}
			}
			yy.Error(msg)
			fallthrough
		case 1, 2: // partially recovered error
			for { // pop states until error can be shifted
				yyn = int({{$yy}}Pact[yystate]) + 1
				if 0 <= yyn && yyn < len({{$yy}}Action) && {{$yy}}Check[yyn] == 1 {
					yyn = {{$yy}}Action[yyn]
					if yyn > 0 {
						break
					}
				}
				if len(yystack) == 0 {
					return nil
				}
				if {{$yy}}Debug >= 2 {
					println("\tPopping state", yystate)
				}
				yystate = yystack[len(yystack)-1].yys
				yystack = yystack[:len(yystack)-1]
			}
			yyerror = 3
			if {{$yy}}Debug >= 1 {
				println("\tSHIFT token error")
			}
			goto yystack
		default: // still waiting for valid tokens
			if yymajor == 0 { // no more tokens
				return nil
			}
			if {{$yy}}Debug >= 1 {
				println("\tDISCARD token", {{$yy}}Name[yymajor])
			}
			yymajor = -1
			goto yyaction
		}
	}
	if {{$yy}}Debug >= 1 {
		println("\tREDUCE rule", yyn)
	}
	yyt = len(yystack) - int({{$yy}}R2[yyn])
	yyD = yystack[yyt:]
	if len(yyD) > 0 { // pop items and restore state
		yyval = yyD[0]
		yystate = yyval.yys
		yystack = yystack[:yyt]
	} else {
		yyval.yyp = yy.Pos
	}
	switch yyn { // Semantic actions
	{{- range .Rules }}{{ .Dump $yy }}{{ end }}
	}
	// look up goto
	yyt = int({{$yy}}R1[yyn]) - {{$yy}}Last
	yyn = int({{$yy}}Pgoto[yyt]) + yystate
	if 0 <= yyn && yyn < len({{$yy}}Action) &&
		int({{$yy}}Check[yyn]) == yystate {
		yyn = int({{$yy}}Action[yyn])
	} else {
		yyn = int({{$yy}}Goto[yyt])
	}
	goto yystack
}
`))
