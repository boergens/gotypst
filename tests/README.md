# Test Fixtures

Test fixtures extracted from the [Typst test suite](https://github.com/typst/typst/tree/main/tests) for validating our Go translation.

## Directory Structure

```
tests/fixtures/
├── syntax/           # Language syntax parsing tests
│   ├── comment.typ   # Line and block comments
│   ├── numbers.typ   # Number display in text mode
│   ├── escape.typ    # Escape sequences (\n, \u{...}, etc.)
│   ├── shorthand.typ # Unicode shorthands (~, --, ---, etc.)
│   └── embedded.typ  # Embedded expressions (#expr)
├── foundations/      # Core types and built-in functions
│   ├── int.typ       # Integer type, bases, conversion
│   ├── str.typ       # String type and methods
│   └── array.typ     # Array type and methods
└── scripting/        # Language control flow and expressions
    ├── let.typ       # Variable bindings
    ├── if.typ        # Conditional expressions
    ├── for.typ       # For loops
    └── ops.typ       # Binary and unary operators
```

## Test Format

Typst tests use a delimiter syntax:

```typst
--- test-name attr ---
// Test code here
#test(actual, expected)
```

### Test Attributes

- `paged` - Test produces paged output
- `html` - Test produces HTML output

### Error Annotations

Tests can annotate expected errors:

```typst
// Error: 2-7 error message here
// Hint: 2-7 hint message here
#invalid_code
```

The format is `line-column` or `line:column-line:column` for spans.

## Categories

### Syntax (Highest Priority for Phase 1)

Tests for the lexer and parser:
- Comments (line `//` and block `/* */`)
- Numeric literals (integers, floats, hex `0x`, binary `0b`)
- String escape sequences
- Unicode shorthands
- Embedded code expressions

### Foundations

Tests for core types:
- `int` - Integer literals, bases, conversion, methods
- `str` - String literals, methods (len, at, slice, etc.)
- `array` - Array literals, methods (push, pop, map, filter, etc.)

### Scripting

Tests for control flow and expressions:
- `let` - Variable binding patterns
- `if` - Conditional expressions with else-if chains
- `for` - Iteration over arrays, dictionaries, strings
- `ops` - Arithmetic, comparison, boolean, assignment operators

## Original Test Suite

The Typst test suite has 15 categories:
- foundations, html, introspection, layout, loading
- math, model, pdf, pdftags, scripting
- styling, symbols, syntax, text, visualize

We've extracted a representative subset focused on parsing and basic evaluation
for Phase 1 of the Go translation.

## Usage

These fixtures should be used to:
1. Validate the Go lexer produces correct tokens
2. Validate the Go parser produces correct AST nodes
3. Validate basic evaluation matches expected results
4. Test error recovery and error message quality

## Source

Extracted from: https://github.com/typst/typst/tree/main/tests/suite
