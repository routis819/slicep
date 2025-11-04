%{
package parse
%}
			
%token LPAREN RPAREN

%token IDENT

%token UINTEGER10

%%

program:	command
	;

command: expression;

/* <expression> −→ <identifier> */
/* | <literal> */
/* | <procedure call> */
/* | <lambda expression> */
/* | <conditional> */
/* | <assignment> */
/* | <derived expression> */
/* | <macro use> */
/* | <macro block> */
/* | <includer> */

expression: IDENT
	|	literal
	|	procedure_call
		;

/* <literal> −→ <quotation> | <self-evaluating> */

literal:	self_evaluating
	;

/* <self-evaluating> −→ <boolean> | <number> | <vector> */
/* | <character> | <string> | <bytevector> */

self_evaluating:
		number		
	;

number:		UINTEGER10
	;

/* <procedure call> −→ (<operator> <operand>*) */

procedure_call:	LPAREN operator operand_star RPAREN
	;

operator:	expression
	;

operand_star:
		/* empty */
	|	operand_star operand
	;

operand:	expression
	;

%%
