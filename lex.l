package main

%%

	Out io.Writer
	part int

id	[A-Z_a-z][A-Z_a-z0-9]*

%x	codefrag copymode

%%
	level := 0

"%%"	{
	yylex.part++
	if yylex.part < 2 {
		return MARK
	}
	BEGIN(copymode)
}
":"	return COLON
";"	return SEMI
"|"	return PIPE

{id}	{
	yylval.s = string(yytext)
	return IDENT
}
%{id}	{
	yylval.s = string(yytext[1:])
	return KEYWORD
}

"<"[^\n>]+">"	{
	yylval.s = string(yytext[1:len(yytext)-1])
	return TYPENAME
}
"{"	level = 1; BEGIN(codefrag)

"//".*				// inline comment
"/*"("*"*[^*/]+)*"*"+"/"	/* multi-line comment */
[ \t\r\v\f\n]+			// white spaces

<codefrag>"{"	level++; yymore()
<codefrag>"}"	{
	level--
	if level == 0 {
		yylval.s = string(yytext[:len(yytext)-1])
		BEGIN(INITIAL)
		return CODEFRAG
	} else {
		yymore()
	}
}
<codefrag>[^{}]+	yymore()

<copymode>(.|\n)*	{
	yylex.Out.Write(yytext)
	return 0
}

<*>.|\n	{
	yylex.Error("invalid character")
}
