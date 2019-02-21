# lrgen

lrgen is an LALR(1) parser generator written in Go and targeting Go.

The grammar file consists of three parts divided by the `%%` markers:

```
Prologue
%%
Definitions
%%
Rules
```

## Prologue

Any Go code will be copied verbatim.  You may want to include a
package declaration and import some packages at the beginning of this.

## Definitions

The following directives are supported

* `%token [type] {identifier}` -- define terminal symbols
* `%left {identifier}` `%right {identifier}` `%nonassoc {identifier}`
-- define precedence and associativity
* `%type type {identifier}` -- specify type for nonterminal symbols
* `%error code` -- code to be executed on syntax error
* `%base identifier|type` -- specify the base type of the parser

## Rules

```
identifier ':' {identifier} [code] { '|' {identifier} [code] } ';'
```

The identifier to the left of `:` is the lhs of the production,
and the identifier list on the right is the rhs, with an optional
semantic action attached.  The rhs can be separated by `|` and
must be terminated by `;`.

## Generated Parser

The generated parser will be called `y.tab.go`.  To use the parser,
create a yyParser object, and pass tokens to its `ParseToken` method.
On the end of input, pass 0 to it.

## Desciption file

A description file `y.output` will be generated.  It contains the state
machine description, and also reports all conflicts.
