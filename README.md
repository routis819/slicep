# slicep

A very small LISP scripting engine for golang.

It aims to implement a subset of r7rs-small, prioritizing features
based on their utility.

## Roadmap

This project aims to evolve into a practical scripting engine by
following these steps:

1.  **Core Language Implementation**:
    *   Stabilize the evaluation of procedure calls, starting with
        primitive arithmetic functions.
    *   Implement `define` to manage variable bindings and state
        within an environment.
    *   Introduce `lambda` to support closures and first-class
        functions, making the language Turing-complete.

2.  **Performance and Optimization**:
    *   Develop a compiler to translate the AST into a custom bytecode
        format.
    *   Build a bytecode Virtual Machine (VM) for faster execution
        compared to tree-walking interpretation.
    *   Implement Tail Call Optimization (TCO) within the VM to allow
        for efficient, stack-safe recursion.

3.  **Application as a DSL**:
    *   The completed engine will be ideal for use as an embedded
        Domain-Specific Language (DSL).
    *   Its S-expression syntax is perfectly suited for defining
        hierarchical data such as configuration files or game assets.
    *   By exposing host functions (from the Go application) to the
        script environment, it can serve as a powerful tool for
        extending application logic, for instance, defining game
        object behaviors and properties.

