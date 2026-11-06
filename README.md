# slicep

A very small LISP scripting engine for golang.

It aims to implement a subset of r7rs-small, prioritizing features
based on their utility.

## Syntax & Design Philosophy

Rather than inventing a custom LISP dialect, `slicep` defines its grammar by
**pruning** the formal syntax of [R7RS-small](https://small.r7rs.org/):

- **Strict R7RS-small Subset**: Unused production rules are pruned without altering
  existing grammar structures or non-terminal names, ensuring valid `slicep` scripts
  remain compatible with standard Scheme tooling (syntax highlighters, formatters,
  etc.).
- **Growth by Restoration**: As new features are needed, previously pruned rules from
  the [R7RS-small specification](https://small.r7rs.org/) are restored rather than
  designed from scratch.

See [`docs/syntax.md`](docs/syntax.md) for the complete formal syntax specification
currently targeted by `slicep`.

## Roadmap

This project evolves through three major milestones, starting as a safe, fast embedded
DSL for logic-bearing assets (such as game damage formulas and skill definitions) and
gradually expanding toward full R7RS-small compliance:

1.  **Milestone 1: Core Expression & Asset Scripting (Current Target)**
    *   Support fundamental data types (`boolean`, `int`, `float`, `string`), arithmetic
        and comparison operations, conditionals (`if`), variable and function
        definitions (`define`), and variable mutation (`set!`).
    *   Compile the AST directly into bytecode and execute it on a lightweight
        bytecode Virtual Machine (VM), enforcing **compile-time termination guarantees
        (preventing recursion and infinite loops statically)**.
    *   Provide seamless host interop for embedding logic-heavy assets (e.g., game
        damage formulas, item/skill behaviors) in Go applications.

2.  **Milestone 2: Safe Collection Processing (`expr-lang` Parity)**
    *   Incrementally extend the Milestone 1 bytecode VM with collection types and
        bounded higher-order operations such as `map`, `filter`, and `reduce` while
        preserving static termination guarantees.

3.  **Milestone 3: Toward Full R7RS-small**
    *   **Top Priority — Tail Call Optimization (TCO)**: Implement proper tail recursion
        as the foundational step for enabling general recursion and loops in Scheme.
    *   **Dual Execution Modes (DSL Mode vs. Turing-Complete Mode)**: Even after
        achieving Turing completeness, retain the termination-guaranteed **DSL Mode**
        alongside the **Turing-Complete Mode** so host applications can choose between
        strict safety (for assets/configs) and full expressiveness.
    *   Expand language coverage toward the complete
        [R7RS-small specification](https://small.r7rs.org/), restoring pruned syntax
        and semantics (`lambda`, pairs/lists, macros, etc.) and redesigning the
        compiler/VM architecture as needed.

