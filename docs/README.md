# GoTypst Documentation

Research and design documentation for the GoTypst project - a Go translation of the Typst typesetting system.

## Architecture

Core design patterns and translation strategies.

- [Rust to Go Translation Guide](TRANSLATION_GUIDE.md) - Common patterns for translating Rust to idiomatic Go
- [Error Handling Design](ERROR_HANDLING.md) - Go error patterns that translate Typst's Rust error system

## Phase Analysis

Deep analysis of Typst's pipeline phases.

- [Phase 2: Eval Analysis](phase2-eval-analysis.md) - Typst evaluation crate and tree-walking interpreter
- [Phase 2: Library Analysis](phase2-library-analysis.md) - Typst standard library structure and type system
- [Phase 3: Layout Analysis](phase3-layout-analysis.md) - Layout engine and pagination system
- [Phase 3: Realize Analysis](phase3-realize-analysis.md) - Realization subsystem and show rule application
- [Phase 4: PDF Analysis](phase4-pdf-analysis.md) - PDF generation and krilla integration

## Dependencies

Go library equivalents for Typst's Rust dependencies.

- [Go Library Equivalents](go-library-equivalents.md) - Research findings for Go alternatives to Rust crates
- [Go Font Handling](GO_FONT_HANDLING.md) - Font library evaluation and recommendations

## Reference

Detailed analysis of specific Typst subsystems.

- [Macros Analysis](MACROS_ANALYSIS.md) - Procedural macro system and Go translation strategies
