# GoTypst - Project Guide

## Overview

GoTypst is a Go implementation of the [Typst](https://typst.app/) typesetting system. The original Typst is written in Rust and provides a modern alternative to LaTeX for document preparation.

**Current Status**: ~30-40% complete. Strong parsing and evaluation infrastructure, but cannot yet compile a complete `.typ` file end-to-end due to missing standard library functions and incomplete realization bridge.

## IMPORTANT: Follow the Rust Implementation Closely

**This is a port, not a reimagining.** When implementing any feature:

1. **Always read the Rust code first** - The reference implementation is at `typst-reference/`
2. **Match the structure** - Use the same function signatures, type definitions, and algorithms
3. **Match the behavior** - The Go implementation should produce identical output to the Rust version
4. **Match the naming** - Use equivalent names (adjusted for Go conventions like PascalCase for exports)
5. **Don't invent** - If you're unsure how something should work, look at the Rust code

The goal is a faithful Go port that behaves identically to the original Typst compiler.

## Reference Implementation

The original Rust implementation is available at `typst-reference/` for reference. This maps to:

| Rust Crate | Go Package |
|------------|------------|
| `typst-syntax` | `syntax/` |
| `typst-eval` | `eval/` |
| `typst-realize` | `realize/` |
| `typst-layout` | `layout/` |
| `typst-library` | `library/` |
| `typst-pdf` | `pdf/` |
| `typst-html` | `html/` |
| `typst-svg` | `svg/` |
| `typst-kit` | `kit/` |

## Architecture

Compilation pipeline:
```
Source (.typ) → Parse → Evaluate → Realize → Layout → Render (PDF/HTML/SVG)
```

### Key Packages

| Package | Purpose | Status |
|---------|---------|--------|
| `syntax/` | Lexer, parser, AST | ~95% complete |
| `eval/` | VM, evaluation engine, scope management | ~70% complete |
| `realize/` | Bridge between eval and layout (show rules, grouping) | Partial |
| `layout/inline/` | Text shaping, line breaking | ~60% complete |
| `layout/flow/` | Block-level flow layout | ~30% (types only) |
| `layout/pages/` | Document pagination | ~30% (types only) |
| `layout/grid/` | Grid and table layout | Partial |
| `layout/math/` | Mathematical typography | Partial |
| `pdf/` | PDF generation | ~40% complete |
| `library/` | Standard library functions | ~5% (critical gap) |
| `html/`, `svg/` | Export formats | Stubs only |
| `kit/` | World implementation utilities | Partial |

### Entry Point

`cmd/gotypst/main.go` - CLI that compiles `.typ` files to PDF

## Build & Test

```bash
# Build
go build ./cmd/gotypst

# Run
gotypst compile input.typ -o output.pdf

# Test
make test
# or
go test ./...
```

## Key Files

- `gotypst.go` - Public API (World, Document, Page interfaces)
- `cmd/gotypst/main.go` - CLI entry point
- `docs/PROJECT_STATUS.md` - Detailed completion tracking
- `docs/ROADMAP.md` - Development phases

## Critical Gaps

1. **Standard Library (~95% missing)** - Most built-in functions not implemented:
   - Text: `text()`, `par()`, `strong()`, `emph()`
   - Structure: `heading()`, `list()`, `enum()`, `table()`
   - Layout: `page()`, `grid()`, `stack()`, `columns()`

2. **Realization System** - Incomplete bridge between evaluation and layout

3. **Layout Integration** - Flow and pages packages are mostly type definitions

## Development Notes

- When implementing standard library functions, reference `typst-reference/crates/typst-library/`
- The eval package uses a stack-based VM with 28 different value types
- Test fixtures are in `tests/fixtures/` - many fail due to incomplete stdlib
