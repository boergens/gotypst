# gotypst Roadmap

## Current Status (Phases 1-4 Complete)

| Phase | Crate | Status | Notes |
|-------|-------|--------|-------|
| 1 | typst-syntax | ✅ Done | Lexer, parser, AST, source handling |
| 2 | typst-eval | ✅ Done | VM, scopes, bindings, expressions, flow control |
| 3 | typst-layout | ✅ Done | Inline text, line breaking, flow, pages |
| 4 | typst-pdf | ✅ Done | PDF writer, objects, streams, images |

**Estimated completion: 30-40%**

## Critical Gaps for Hello World

1. **No CLI** - No main.go entry point
2. **Standard library ~95% missing** - No text(), par(), heading(), page()
3. **No realization system** - Bridge between eval and layout missing
4. **No World implementation** - File system and font access missing

## Phase 5-9: Remaining Work

### Phase 5: Standard Library Core (`go-1u8t`)
10 tasks for data type methods and calc module:
- calc math functions (basic, trig, rounding, number theory)
- str methods (inspection, pattern matching, transformation)
- array/dict methods
- datetime/duration

### Phase 6: Text and Visual Elements (`go-jeej`)
6 tasks for visual elements:
- Color type and operations
- Color space conversions
- Gradient and tiling
- Symbol definitions
- **Text element** ← Hello World critical
- Raw/code elements

### Phase 7: Document Model (`go-xeqo`)
8 tasks for document structure:
- **Page element** ← Hello World critical
- Heading element
- **Paragraph/parbreak** ← Hello World critical
- List elements
- Stack and alignment
- Grid and columns
- Box and block
- Pad and place

### Phase 8: Realization System (`go-dwbq`)
6 tasks for content→layout bridge:
- **Realize core function** ← Hello World critical
- Show rules processing
- Grouping rules
- Space collapsing
- Regex text matching
- Style chain cascading

### Phase 9: World and CLI (`go-pafo`)
6 tasks for integration:
- **World interface** ← Hello World critical
- **Standard library scope** ← Hello World critical
- Font loading
- File loading functions
- **Compile pipeline** ← Hello World critical
- **CLI entry point** ← Hello World critical

## Hello World Critical Path

Minimum tasks to compile `#"Hello World"` to PDF:

```
go-pafo.1  World interface
    ↓
go-pafo.2  Standard library scope (minimal)
    ↓
go-jeej.5  Text element
go-xeqo.1  Page element
go-xeqo.3  Paragraph/parbreak
    ↓
go-dwbq.1  Realize core function
    ↓
go-pafo.5  Compile pipeline
    ↓
go-pafo.6  CLI entry point
```

**8 tasks** on critical path, parallelizable in pairs.

## Task Summary

| Phase | Epic | Tasks | Est. Time |
|-------|------|-------|-----------|
| 5 | go-1u8t | 10 | 2.5 hrs |
| 6 | go-jeej | 6 | 1.5 hrs |
| 7 | go-xeqo | 8 | 2 hrs |
| 8 | go-dwbq | 6 | 1.5 hrs |
| 9 | go-pafo | 6 | 1.5 hrs |
| **Total** | | **36** | **9 hrs** |

## Recommended Approach

**Hello World First**: Focus on the 8 critical path tasks to get end-to-end compilation working, then expand standard library breadth.
