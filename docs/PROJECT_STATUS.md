# GoTypst Project Status

*Generated: 2026-01-19*

## Executive Summary

GoTypst is a Go translation of the Typst typesetting system. The project is approximately **30-40% complete**, with strong progress on parsing and evaluation infrastructure, but significant gaps in the standard library and layout integration needed for end-to-end compilation.

**Current state: Cannot compile .typ files end-to-end.**

---

## Codebase Overview

| Package | Files | Lines | Status | Completeness |
|---------|-------|-------|--------|--------------|
| syntax | 34 | ~12,000 | Working | 95% |
| eval | 20 | ~8,700 | Working | 70% |
| layout/inline | 8 | ~3,000 | Working | 60% |
| layout/flow | 3 | ~1,300 | Types only | 30% |
| layout/pages | 7 | ~1,500 | Types only | 30% |
| pdf | 6 | ~2,100 | Partial | 40% |
| library | 4 | ~500 | Minimal | 5% |
| tests | 2 | ~700 | Harness | N/A |
| render/html/svg | 3 | ~15 | Stubs | 0% |

**Total: 92 Go files, ~35,000 lines of code**

---

## What Works

### 1. Syntax Package ✅
The parser is essentially complete:
- Full lexer with all token types
- Parser for markup, code, and math modes
- Complete AST types for all Typst expressions
- Source file handling with spans and error recovery
- Incremental reparsing infrastructure

**Evidence:** All syntax unit tests pass. Test fixtures parse correctly.

### 2. Eval Package 🟡
The evaluation engine structure is complete:
- VM with scope stack, flow control, and context management
- All 28 value types defined (None, Bool, Int, Float, Str, Array, Dict, Content, Func, etc.)
- Expression evaluation dispatch for all AST node types
- Closure capture and module import handling
- Binary and unary operators
- Control flow (if, for, while, break, continue, return)
- Set/show rule evaluation structure

**Limitation:** Cannot fully execute code because standard library functions are missing.

### 3. Layout/Inline Package 🟡
Text shaping and line breaking work:
- Text shaping via go-text/typesetting library
- Knuth-Plass line breaking algorithm
- Inline decoration handling
- Glyph positioning and runs

**Evidence:** Shaping and line breaking tests pass.

### 4. PDF Package 🟡
Basic PDF generation works:
- Content stream generation
- Image embedding (JPEG, PNG)
- Basic text rendering
- Coordinate transformation

**Limitation:** Font embedding incomplete, no advanced features.

---

## What's Missing

### 1. No CLI Entry Point ❌
There is no `main.go` or `cmd/` directory. No way to actually invoke the compiler.

### 2. Standard Library (~95% Missing) ❌

The library package only contains `foundations/ops.go` with basic arithmetic operations. Missing:

| Module | Status | What's Needed |
|--------|--------|---------------|
| foundations/calc | ❌ | 40+ math functions (sin, cos, sqrt, etc.) |
| foundations/str | ❌ | String methods (len, contains, split, etc.) |
| foundations/array | ❌ | Array methods (map, filter, sort, etc.) |
| foundations/dict | ❌ | Dictionary methods |
| foundations/datetime | ❌ | Date/time handling |
| text | ❌ | text(), par(), emph(), strong(), etc. |
| model | ❌ | heading(), list(), enum(), table(), etc. |
| layout | ❌ | page(), grid(), stack(), columns(), etc. |
| math | ❌ | Mathematical typesetting functions |
| visualize | ❌ | line(), rect(), circle(), path(), etc. |
| loading | ❌ | json(), yaml(), csv(), etc. |
| introspection | ❌ | query(), locate(), counter(), state() |

### 3. Realization System ❌

The "realize" step that converts styled content into concrete layout elements is not implemented. This is the bridge between eval and layout.

### 4. Layout Integration ❌

- Flow layout exists as types but cannot process children
- Pages layout cannot call flow layout
- No integration between inline, flow, and pages

### 5. World Implementation ❌

No concrete implementation of the World interface:
- File system access
- Font discovery and loading
- Package management
- Resource resolution

---

## Test Suite Status

### Unit Tests: ✅ Passing
```
ok  github.com/boergens/gotypst/eval         1.210s
ok  github.com/boergens/gotypst/layout/inline 1.463s
ok  github.com/boergens/gotypst/layout/pages  0.990s
ok  github.com/boergens/gotypst/library/foundations 1.739s
ok  github.com/boergens/gotypst/pdf           1.940s
ok  github.com/boergens/gotypst/syntax        2.133s
```

### Fixture Tests: 🟡 Partial
- Syntax fixtures: Pass (lexing/parsing works)
- Foundations fixtures: Fail (runtime errors not produced because code doesn't execute)
- Scripting fixtures: Fail (same reason)

The fixture test harness is built but cannot fully validate behavior until the VM can execute code with a complete standard library.

---

## Path to End-to-End Compilation

To compile even a simple "Hello World" .typ file:

### Phase A: Minimal Working Compiler
1. **World implementation** with file system and basic font loading
2. **Core library functions:**
   - `text()` element
   - `par()` element
   - Basic string/array operations
3. **Realize step** to convert content to layout primitives
4. **Flow layout integration** with pages
5. **CLI entry point**

### Phase B: Practical Compiler
6. **Heading, list, enum** elements
7. **Grid/table** layout
8. **Image** embedding with proper sizing
9. **Font subsetting** for PDF
10. **Math** typesetting

### Phase C: Feature Parity
11. **Introspection** (query, counter, state)
12. **Bibliography** and citations
13. **HTML export**
14. **All standard library functions**

---

## Comparison to Original Typst

Original Typst (Rust) has approximately:
- 150+ source files in typst-syntax
- 50+ source files in typst-eval
- 100+ source files in typst-layout
- 50+ source files in typst-library
- 30+ source files in typst-pdf

GoTypst currently has about **20-25% of the original file count**, but many files are stubs or partial implementations.

---

## Recommendations

### Near-term (Get Something Working)
1. Create a minimal World implementation for local files
2. Implement `text()` and basic content elements
3. Wire up flow layout to pages layout
4. Add a simple CLI that outputs PDF

### Medium-term (Useful Compiler)
5. Implement core standard library functions systematically
6. Add proper font handling with system font discovery
7. Complete the realize step

### Architecture Notes
- The eval/value type system is well-designed and should serve as the foundation
- Consider implementing library functions incrementally, driven by test fixtures
- The inline shaping code is solid; focus on flow/pages integration

---

## Conclusion

GoTypst has a solid foundation with working parsing and evaluation infrastructure. The main blockers to a working compiler are:

1. **Standard library** - No content elements exist
2. **Realization** - No bridge from eval to layout
3. **Flow/pages integration** - Layout packages don't connect
4. **World** - No file/font system

The architecture is sound and follows the original Typst structure well. Progress is methodical but the remaining work is substantial - completing the compiler to parity would likely require 3-4x the current codebase size.
