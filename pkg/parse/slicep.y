%{
package parse

import "github.com/routis819/slicep/pkg/lex"
%}

%union {
	pos       lex.Pos
	bval      bool
	ival      int64
	fval      float64
	sval      string
	datum     Datum
	datumlist []Datum
}

%type	<datumlist>	datum_star
%type	<datum>		datum
			simple_datum
			compound_datum
			boolean
			number
			string
			symbol
			list

%token			LPAREN
			RPAREN

%token	<bval>		BOOLEAN
%token	<ival>		INTEGER
%token	<fval>		FLOAT
%token	<sval>		STRING
			IDENT

%%

top:		datum_star
		{
			yylex.(*yyLexerImpl).result = $1
		}
	;

/* <datum> -> <simple datum> | <compound datum> */

datum:		simple_datum
		{
			$$ = $1
		}
	|	compound_datum
		{
			$$ = $1
		}
	;

/* <simple datum> -> <boolean> | <number> | <string> | <symbol> */

simple_datum:	boolean
		{
			$$ = $1
		}
	|	number
		{
			$$ = $1
		}
	|	string
		{
			$$ = $1
		}
	|	symbol
		{
			$$ = $1
		}
	;

boolean:	BOOLEAN
		{
			$$ = &BoolDatum{Value: $1, Pos: $<pos>1}
		}
	;

number:		INTEGER
		{
			$$ = &IntDatum{Value: $1, Pos: $<pos>1}
		}
	|	FLOAT
		{
			$$ = &FloatDatum{Value: $1, Pos: $<pos>1}
		}
	;

string:		STRING
		{
			$$ = &StringDatum{Value: $1, Pos: $<pos>1}
		}
	;

/* <symbol> -> <identifier> */

symbol:		IDENT
		{
			$$ = &SymbolDatum{Value: $1, Pos: $<pos>1}
		}
	;

/* <compound datum> -> <list> */

compound_datum:	list
		{
			$$ = $1
		}
	;

/* <list> -> (<datum>*) */

list:		LPAREN datum_star RPAREN
		{
			$$ = &ListDatum{Elements: $2, Pos: $<pos>1}
		}
	;

datum_star:
		/* empty */
		{
			$$ = make([]Datum, 0)
		}
	|	datum_star datum
		{
			$$ = append($1, $2)
		}
	;

%%
