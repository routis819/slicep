# slicep Formal Syntax

This document specifies the formal syntax of `slicep`, written in an extended BNF.

## Design Principle: Pruning from R7RS-small

`slicep` does not invent a custom LISP dialect. Instead, its syntax specification is
strictly derived by **pruning** the official
[R7RS-small specification](https://small.r7rs.org/).

When maintaining or extending this document, the following rules must be observed:

1. **Preserve Non-terminal Names**: All non-terminal symbols (e.g., `<expression>`,
   `<procedure call>`, `<ureal 10>`) must retain their exact names from the
   [R7RS-small specification](https://small.r7rs.org/).
2. **Simplify by Pruning Only**: Grammar reduction is achieved strictly by removing
   unused production alternatives (`|`) or omitting optional elements that can derive
   `<empty>`. Existing rules are not arbitrarily rewritten.
3. **Extend by Restoration**: When adding new language features to `slicep`, restore
   the corresponding pruned rules from the
   [R7RS-small specification](https://small.r7rs.org/) rather than designing new
   syntax.
4. **Document Any Divergence**: If a deviation from R7RS-small is ever required for
   practical or parser implementation reasons, it must be explicitly documented
   alongside the original R7RS-small rule.

## Overview of Supported Features

- **Data Types**: Boolean (`boolean`), Integer (`int`), Floating-point (`float`), and
  String (`string`) literals
- **Arithmetic, Comparisons & Procedure Calls**: Basic arithmetic (`+`, `-`, `*`,
  `/`), comparisons (`=`, `<`, `>`, `<=`, `>=`), and general procedure application
  via `<procedure call>`
- **Conditionals**: Branching logic via `if` (`<conditional>`)
- **Definitions**: Variable and function declarations via `define`
- **Assignments**: Variable mutation via `set!` (`<assignment>`)

## Notation

All spaces in the grammar are for legibility. Case is not significant in the grammar
of numbers (e.g., `1e5` and `1E5` are equivalent), but is significant in identifiers,
booleans, and string escapes (e.g., `foo` and `Foo` are distinct). `<empty>` stands
for the empty string.

The following extensions to BNF are used:
- `<thing>*` means zero or more occurrences of `<thing>`.
- `<thing>+` means at least one `<thing>`.

---

## Lexical Structure

This section describes how individual tokens (identifiers, booleans, numbers,
strings, etc.) are formed from sequences of characters.

`<intertoken space>` can occur on either side of any token, but not within a token.

Identifiers, numbers, and booleans are terminated by a `<delimiter>` or by the end of
the input.

### Whitespace and Comments

```ebnf
<token> -> <identifier> | <boolean> | <number> | <string>
    | ( | )
<delimiter> -> <whitespace> | <vertical line>
    | ( | ) | " | ;
<intraline whitespace> -> <space or tab>
<whitespace> -> <intraline whitespace> | <line ending>
<vertical line> -> |
<line ending> -> <newline> | <return> <newline>
    | <return>
<comment> -> ; <all subsequent characters up to a line ending>
    | <nested comment>
<nested comment> -> #| <comment text> <comment cont>* |#
<comment text> -> <character sequence not containing #| or |#>
<comment cont> -> <nested comment> <comment text>

<atmosphere> -> <whitespace> | <comment>
<intertoken space> -> <atmosphere>*
```

### Identifiers, Booleans, and Strings

Note that `<infnan>` (`+inf.0`, `-inf.0`, `+nan.0`, `-nan.0`) is an exception to the
`<peculiar identifier>` rule; it is parsed as a number, not an identifier. Built-in
operators (`+`, `-`, `*`, `/`, `=`, `<`, `>`, `<=`, `>=`) and syntax keywords
(`define`, `set!`, `if`, `begin`) follow the standard identifier lexical structure.

```ebnf
<identifier> -> <initial> <subsequent>*
    | <peculiar identifier>
<initial> -> <letter> | <special initial>
<letter> -> a | b | c | ... | z
    | A | B | C | ... | Z
<special initial> -> ! | $ | % | & | * | / | : | < | =
    | > | ? | ^ | _ | ~
<subsequent> -> <initial> | <digit>
    | <special subsequent>
<digit> -> 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9
<explicit sign> -> + | -
<special subsequent> -> <explicit sign> | . | @
<mnemonic escape> -> \a | \b | \t | \n | \r
<peculiar identifier> -> <explicit sign>
    | <explicit sign> <sign subsequent> <subsequent>*
    | <explicit sign> . <dot subsequent> <subsequent>*
    | . <dot subsequent> <subsequent>*
<dot subsequent> -> <sign subsequent> | .
<sign subsequent> -> <initial> | <explicit sign> | @

<boolean> -> #t | #f | #true | #false

<string> -> " <string element>* "
<string element> -> <any character other than " or \>
    | <mnemonic escape> | \" | \\
    | \<intraline whitespace>* <line ending> <intraline whitespace>*
```

### Numbers (`int` and `float`)

`slicep` supports decimal (`R = 10`) integer and floating-point numbers, pruned from
the R7RS `<num 10>` rules:
- **`int`**: Represented by `<sign>` followed by `<uinteger 10>`.
- **`float`**: Represented by `<sign>` followed by `<decimal 10>`, or `<infnan>`.

```ebnf
<number> -> <num 10>
<num 10> -> <real 10>
<real 10> -> <sign> <ureal 10>
    | <infnan>
<ureal 10> ->
    <uinteger 10>
    | <decimal 10>
<decimal 10> ->
    <uinteger 10> <suffix>
    | . <digit 10>+ <suffix>
    | <digit 10>+ . <digit 10>* <suffix>
<uinteger 10> -> <digit 10>+
<infnan> -> +inf.0 | -inf.0 | +nan.0 | -nan.0

<suffix> -> <empty>
    | <exponent marker> <sign> <digit 10>+
<exponent marker> -> e
<sign> -> <empty> | + | -
<digit 10> -> <digit>
```

---

## Expressions

Arithmetic and comparison operations (`+`, `-`, `*`, `/`, `=`, `<`, `>`, etc.) are
evaluated as `<procedure call>` expressions where `<operator>` is the corresponding
identifier.

> **Note on `<operator>` and Static Termination Guarantees**:
> To strictly preserve R7RS-small production rules, `<operator> -> <expression>` is
> retained in the grammar below. However, because `<lambda expression>` is intentionally
> pruned from `<expression>` and functions can only be declared via
> `(define (<identifier> <def formals>) <body>)`, valid operators in practice reduce to
> `<operator> -> <identifier>` (referring to a statically known built-in or previously
> defined function).
>
> This restriction is essential for **preventing recursion and infinite loops at
> compile time** (e.g., disallowing first-class function values, `set!` reassignment of
> functions, or Y-combinator self-application). Unless a dedicated mechanism for
> preventing or safely bounding recursive calls is introduced, `<operator>` must
> effectively remain restricted to `<identifier>`.

```ebnf
<expression> -> <identifier>
    | <literal>
    | <procedure call>
    | <conditional>
    | <assignment>

<literal> -> <self-evaluating>
<self-evaluating> -> <boolean> | <number> | <string>

<procedure call> -> (<operator> <operand>*)
<operator> -> <expression>
<operand> -> <expression>

<body> -> <definition>* <sequence>
<sequence> -> <command>* <expression>
<command> -> <expression>

<conditional> -> (if <test> <consequent> <alternate>)
<test> -> <expression>
<consequent> -> <expression>
<alternate> -> <expression> | <empty>

<assignment> -> (set! <identifier> <expression>)
```

---

## Programs and Definitions

A program consists of one or more commands (expressions) or definitions. `define`
supports both variable bindings and function declarations.

```ebnf
<program> -> <command or definition>+

<command or definition> -> <command>
    | <definition>
    | (begin <command or definition>+)

<definition> -> (define <identifier> <expression>)
    | (define (<identifier> <def formals>) <body>)

<def formals> -> <identifier>*
```
