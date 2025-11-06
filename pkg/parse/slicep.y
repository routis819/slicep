%{
package parse

type Node any

type ExprNode struct {}

type IdentifierNode struct {
    Value string
}

type NumberNode struct {
    Value uint
}

type ProcedureCallNode struct {
    Operator Node
    Operands []Node
}
%}
			
%union {
    uintval uint
    sval string
    node Node
    nodelist []Node
}

%type	<node>		program
			command
			expression
			literal
			self_evaluating
			number
			procedure_call
			operator
			operand

%type	<nodelist>	operand_star

%token			LPAREN
			RPAREN

%token	<sval>		IDENT

%token	<uintval>	UINTEGER10

%%

program:	command
		{
		    currentLexer.result = $1
		}
		;

command: expression
		{
		    $$ = $1 // expression이 생성한 AST 노드를 그대로 반환
		}
		;

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

expression:	IDENT
		{
		    $$ = &IdentifierNode{Value: $1}
		}
	|	literal
		{
		    $$ = $1
		}
	|	procedure_call
		{
		    $$ = $1
		}
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
		{
		    $$ = &NumberNode{Value: $1}
		}
	;

/* <procedure call> −→ (<operator> <operand>*) */

procedure_call:	LPAREN operator operand_star RPAREN
		{
		    $$ = &ProcedureCallNode{Operator: $2, Operands: $3}
		}
	;

operator:	expression
	;

operand_star:
		/* empty */
		{
		    $$ = make([]Node, 0)
		}
	|	operand_star operand
		{
		    $$ = append($1, $2)
		}
	;

operand:	expression
	;

%%
