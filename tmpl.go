package main

import (
	"text/template"
)

var yyTmpl = template.Must(template.New("yyparse").Parse(`
{{$yy := .Prefix}}
type {{$yy}}SymType struct {
	{{$yy}}s int
	{{$yy}}pos int
	{{.Union}}
}

var {{$yy}}Debug = 0 // debug info from parser

// {{$yy}}Parse read tokens from {{$yy}}lex and parses input.
// Returns result on success, or nil on failure.
func {{$yy}}Parse({{$yy}}lex *{{$yy}}Lex) *{{$yy}}SymType {
	var (
		{{$yy}}n, {{$yy}}t int
		{{$yy}}state  = 0
		{{$yy}}error  = 0
		{{$yy}}major  = -1
		{{$yy}}stack  []{{$yy}}SymType
		{{$yy}}D      []{{$yy}}SymType // rhs of reduction
		{{$yy}}lval   {{$yy}}SymType   // lexcial value from lexer
		{{$yy}}val    {{$yy}}SymType   // value to be pushed onto stack
	)
	goto {{$yy}}action
{{$yy}}stack:
	{{$yy}}val.{{$yy}}s = {{$yy}}state
	{{$yy}}val.{{$yy}}pos = {{$yy}}lex.Pos
	{{$yy}}stack = append({{$yy}}stack, {{$yy}}val)
	{{$yy}}state = {{$yy}}n
	if {{$yy}}Debug >= 2 {
		println("\tGOTO state", {{$yy}}n)
	}
{{$yy}}action:
	// look up shift or reduce
	{{$yy}}n = int({{$yy}}Pact[{{$yy}}state])
	if {{$yy}}n == len({{$yy}}Action) && {{$yy}}state != {{$yy}}Accept { // simple state
		goto {{$yy}}default
	}
	if {{$yy}}major < 0 {
		{{$yy}}major = {{$yy}}lex.Lex(&{{$yy}}lval)
		if {{$yy}}Debug >= 1 {
			println("In state", {{$yy}}state)
		}
		if {{$yy}}Debug >= 2 {
			println("\tInput token", {{$yy}}Name[{{$yy}}major])
		}
	}
	if {{$yy}}major == 0 && {{$yy}}state == {{$yy}}Accept {
		if {{$yy}}Debug >= 1 {
			println("\tACCEPT!")
		}
		return &{{$yy}}stack[0]
	}
	{{$yy}}n += {{$yy}}major
	if 0 <= {{$yy}}n && {{$yy}}n < len({{$yy}}Action) && int({{$yy}}Check[{{$yy}}n]) == {{$yy}}major {
		{{$yy}}n = int({{$yy}}Action[{{$yy}}n])
		if {{$yy}}n <= 0 {
			{{$yy}}n = -{{$yy}}n
			goto {{$yy}}reduce
		}
		if {{$yy}}Debug >= 1 {
			println("\tSHIFT token", {{$yy}}Name[{{$yy}}major])
		}
		if {{$yy}}error > 0 {
			{{$yy}}error--
		}
		{{$yy}}major = -1
		{{$yy}}val = {{$yy}}lval
		goto {{$yy}}stack
	}
{{$yy}}default:
	{{$yy}}n = int({{$yy}}Reduce[{{$yy}}state])
{{$yy}}reduce:
	if {{$yy}}n == 0 {
		switch {{$yy}}error {
		case 0: // new error
			if {{$yy}}Debug >= 1 {
				println("\tERROR!")
			}
			msg := "unexpected " + {{$yy}}Name[{{$yy}}major]
			var expect []int
			if {{$yy}}Reduce[{{$yy}}state] == 0 {
				{{$yy}}n = {{$yy}}Pact[{{$yy}}state] + 3
				for i := 3; i < {{$yy}}Last; i++ {
					if 0 <= {{$yy}}n && {{$yy}}n < len({{$yy}}Action) && {{$yy}}Check[{{$yy}}n] == i && {{$yy}}Action[{{$yy}}n] != 0 {
						expect = append(expect, i)
						if len(expect) > 4 {
							break
						}
					}
					{{$yy}}n++
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
			{{$yy}}lex.Error(msg)
			fallthrough
		case 1, 2: // partially recovered error
			for { // pop states until error can be shifted
				{{$yy}}n = int({{$yy}}Pact[{{$yy}}state]) + 1
				if 0 <= {{$yy}}n && {{$yy}}n < len({{$yy}}Action) && {{$yy}}Check[{{$yy}}n] == 1 {
					{{$yy}}n = {{$yy}}Action[{{$yy}}n]
					if {{$yy}}n > 0 {
						break
					}
				}
				if len({{$yy}}stack) == 0 {
					return nil
				}
				if {{$yy}}Debug >= 2 {
					println("\tPopping state", {{$yy}}state)
				}
				{{$yy}}state = {{$yy}}stack[len({{$yy}}stack)-1].{{$yy}}s
				{{$yy}}stack = {{$yy}}stack[:len({{$yy}}stack)-1]
			}
			{{$yy}}error = 3
			if {{$yy}}Debug >= 1 {
				println("\tSHIFT token error")
			}
			goto {{$yy}}stack
		default: // still waiting for valid tokens
			if {{$yy}}major == 0 { // no more tokens
				return nil
			}
			if {{$yy}}Debug >= 1 {
				println("\tDISCARD token", {{$yy}}Name[{{$yy}}major])
			}
			{{$yy}}major = -1
			goto {{$yy}}action
		}
	}
	if {{$yy}}Debug >= 1 {
		println("\tREDUCE rule", {{$yy}}n)
	}
	{{$yy}}t = len({{$yy}}stack) - int({{$yy}}R2[{{$yy}}n])
	{{$yy}}D = {{$yy}}stack[{{$yy}}t:]
	if len({{$yy}}D) > 0 { // pop items and restore state
		{{$yy}}val = {{$yy}}D[0]
		{{$yy}}state = {{$yy}}val.{{$yy}}s
		{{$yy}}stack = {{$yy}}stack[:{{$yy}}t]
	}
	switch {{$yy}}n { // Semantic actions
	{{- range .Rules }}{{ .Dump $yy }}{{ end }}
	}
	// look up goto
	{{$yy}}t = int({{$yy}}R1[{{$yy}}n]) - {{$yy}}Last
	{{$yy}}n = int({{$yy}}Pgoto[{{$yy}}t]) + {{$yy}}state
	if 0 <= {{$yy}}n && {{$yy}}n < len({{$yy}}Action) &&
		int({{$yy}}Check[{{$yy}}n]) == {{$yy}}state {
		{{$yy}}n = int({{$yy}}Action[{{$yy}}n])
	} else {
		{{$yy}}n = int({{$yy}}Goto[{{$yy}}t])
	}
	goto {{$yy}}stack
}
`))
