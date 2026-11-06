# Agent Guidelines for slicep

## Commit Guidelines

- **Commit Size**: Keep each commit within **300 lines of changes** (excluding test files `*_test.go` and auto-generated files such as `pkg/lex/lex.go`, `pkg/lex/tokentype_string.go`, and `pkg/parse/y.go`) as a general guideline.
  - When implementing a feature or issue that would exceed ~300 lines of non-test/non-generated code, split the work into smaller, logical, self-contained commits that each pass `scripts/pre-commit`.
