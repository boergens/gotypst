# Binding Test Coverage Review

**Reviewer:** quartz
**Date:** 2026-01-19
**Files Reviewed:**
- `eval/binding.go` (implementation)
- `eval/binding_test.go` (unit tests)
- `tests/fixtures/scripting/let.typ` (integration tests)
- `tests/fixtures/scripting/for.typ` (integration tests)

## Executive Summary

The binding/destructuring implementation has **moderate unit test coverage** but **significant gaps in integration testing**. The existing tests primarily verify error message formatting and binding struct behavior rather than actual destructuring logic with parsed patterns.

**Coverage Grade: C+**

## Test Coverage Analysis

### 1. Well-Covered Areas

| Component | Test | Coverage |
|-----------|------|----------|
| `NewBinding()` | `TestBindingKinds` | Full |
| `NewMutableBinding()` | `TestBindingKinds` | Full |
| `NewClosureBinding()` | `TestBindingKinds` | Full |
| `NewModuleBinding()` | `TestBindingKinds` | Full |
| `Binding.Clone()` | `TestBindingClone` | Full |
| `Binding.Category` | `TestBindingCategory` | Full |
| `Binding.Write()` (success) | `TestBindingKinds` | Partial |
| Error message formatting | `TestDestructuringErrors` | Full |

### 2. Coverage Gaps

#### Critical Gaps (High Priority)

| Component | Status | Impact |
|-----------|--------|--------|
| `Destructure()` | **NOT TESTED** | Core let binding functionality untested |
| `DestructureAssign()` | **NOT TESTED** | Assignment destructuring untested |
| `destructureArray()` | **NOT TESTED** | Array pattern matching untested |
| `destructureDict()` | **NOT TESTED** | Dict pattern matching untested |

#### Important Gaps (Medium Priority)

| Component | Status | Notes |
|-----------|--------|-------|
| `Binding.Read()` | NOT TESTED | Simple method, low risk |
| `Binding.ReadChecked()` | NOT TESTED | Future checks mentioned in comments |
| `Binding.Write()` (error case) | NOT TESTED | Immutable binding error path |
| `ImmutableBindingError.Error()` | NOT TESTED | Error message verification |
| `UnknownPatternError.Error()` | NOT TESTED | Error message verification |

#### Pattern Type Coverage Gaps

| Pattern Type | Go Tests | .typ Fixtures |
|--------------|----------|---------------|
| `NormalPattern` (identifier) | Indirect only | Yes |
| `PlaceholderPattern` (`_`) | NOT TESTED | Yes (for.typ:98) |
| `ParenthesizedPattern` (`(x)`) | NOT TESTED | Yes (let.typ:62) |
| `DestructuringPattern` (`(a, b)`) | NOT TESTED | Limited |
| `DestructuringSpread` (`..rest`) | NOT TESTED | NO |
| `DestructuringNamed` (`key: val`) | NOT TESTED | NO |

### 3. Integration Test Coverage (.typ fixtures)

#### Covered Scenarios
- Basic let bindings (`let.typ:5-10`)
- Function sugar (`let.typ:14`)
- For loop with dict destructuring (`for.typ:10-12`)
- For loop with enumerate destructuring (`for.typ:48-49`)
- Placeholder in for loops (`for.typ:98`)
- Error: cannot destructure string (`for.typ:86-89`)

#### Missing Scenarios
- Array destructuring: `let (a, b, c) = (1, 2, 3)`
- Nested destructuring: `let ((a, b), c) = ((1, 2), 3)`
- Spread in arrays: `let (first, ..rest) = (1, 2, 3, 4)`
- Spread in dicts: `let (a: x, ..rest) = (a: 1, b: 2, c: 3)`
- Named dict patterns: `let (name: n, age: a) = (name: "x", age: 1)`
- Destructuring assignment (non-let): `(a, b) = (b, a)`
- Empty array destructuring: `let () = ()`
- Mismatch errors: too few/many elements

## Test Quality Assessment

### Organization (Grade: B)

**Strengths:**
- Clear test function names following Go conventions
- Logical grouping by functionality
- Use of table-driven tests in some areas

**Weaknesses:**
- `TestDestructureArray` and `TestDestructureDict` are misleadingly named - they don't test actual destructuring, only manual binding creation
- Comments acknowledge limitations but don't address them

### Readability (Grade: B+)

**Strengths:**
- Clear setup code
- Good use of subtests with `t.Run()`
- Assertions are straightforward

**Weaknesses:**
- Some tests simulate behavior rather than testing actual code paths
- Missing helper functions for common patterns

### Maintainability (Grade: B-)

**Concerns:**
- Tests are coupled to internal VM implementation
- No test helpers for creating parsed patterns
- Changes to destructuring logic won't be caught by current tests

## Recommended Additional Test Cases

### High Priority

1. **Array Destructuring with Real Patterns**
```go
func TestDestructureArrayIntegration(t *testing.T) {
    // Parse and evaluate: let (a, b, c) = (1, 2, 3)
    // Verify: a=1, b=2, c=3
}
```

2. **Dict Destructuring with Real Patterns**
```go
func TestDestructureDictIntegration(t *testing.T) {
    // Parse and evaluate: let (name, age) = (name: "Alice", age: 30)
    // Verify: name="Alice", age=30
}
```

3. **Spread Patterns**
```go
func TestDestructureSpread(t *testing.T) {
    // Parse and evaluate: let (first, ..rest) = (1, 2, 3, 4)
    // Verify: first=1, rest=(2, 3, 4)
}
```

4. **Named Dict Patterns**
```go
func TestDestructureNamedDict(t *testing.T) {
    // Parse and evaluate: let (name: n, age: a) = (name: "x", age: 1)
    // Verify: n="x", a=1
}
```

### Medium Priority

5. **Placeholder Pattern**
```go
func TestPlaceholderPattern(t *testing.T) {
    // Parse and evaluate: let (a, _, c) = (1, 2, 3)
    // Verify: a=1, c=3, no binding for _
}
```

6. **Nested Destructuring**
```go
func TestNestedDestructuring(t *testing.T) {
    // Parse and evaluate: let ((a, b), c) = ((1, 2), 3)
    // Verify: a=1, b=2, c=3
}
```

7. **Destructuring Assignment**
```go
func TestDestructuringAssignment(t *testing.T) {
    // let a = 1; let b = 2
    // (a, b) = (b, a)
    // Verify: a=2, b=1
}
```

8. **Immutable Binding Error**
```go
func TestImmutableBindingWrite(t *testing.T) {
    b := NewBinding(Int(42), span)
    err := b.Write(Int(100))
    // Verify: err is ImmutableBindingError
}
```

### Low Priority (Edge Cases)

9. **Empty Array Destructuring**
```go
func TestEmptyArrayDestructuring(t *testing.T) {
    // let () = ()
    // Should succeed with no bindings
}
```

10. **Length Mismatch Errors**
```go
func TestDestructureLengthMismatch(t *testing.T) {
    // let (a, b, c) = (1, 2)  // too few
    // let (a, b) = (1, 2, 3)  // too many
}
```

## Recommended .typ Fixture Additions

Add to `tests/fixtures/scripting/destructure.typ`:

```typst
--- destructure-array-basic paged ---
#let (a, b, c) = (1, 2, 3)
#test(a, 1)
#test(b, 2)
#test(c, 3)

--- destructure-array-spread paged ---
#let (first, ..rest) = (1, 2, 3, 4)
#test(first, 1)
#test(rest, (2, 3, 4))

--- destructure-dict-basic paged ---
#let (name, age) = (name: "Alice", age: 30)
#test(name, "Alice")
#test(age, 30)

--- destructure-dict-named paged ---
#let (name: n, age: a) = (name: "Bob", age: 25)
#test(n, "Bob")
#test(a, 25)

--- destructure-dict-spread paged ---
#let (a, ..rest) = (a: 1, b: 2, c: 3)
#test(a, 1)
#test(rest, (b: 2, c: 3))

--- destructure-nested paged ---
#let ((a, b), c) = ((1, 2), 3)
#test(a, 1)
#test(b, 2)
#test(c, 3)

--- destructure-placeholder paged ---
#let (a, _, c) = (1, 2, 3)
#test(a, 1)
#test(c, 3)

--- destructure-swap paged ---
#let a = 1
#let b = 2
// Note: This tests assignment destructuring if supported
// (a, b) = (b, a)

--- destructure-error-too-few paged ---
// Error: not enough elements to destructure
#let (a, b, c) = (1, 2)

--- destructure-error-too-many paged ---
// Error: too many elements to destructure
#let (a, b) = (1, 2, 3)

--- destructure-error-type paged ---
// Error: cannot destructure int
#let (a, b) = 42
```

## Conclusion

The binding implementation has solid foundational tests for binding struct operations but lacks comprehensive coverage of the core destructuring functionality. The tests simulate destructuring behavior rather than testing actual pattern-matched destructuring through the parser and evaluator pipeline.

**Immediate Actions:**
1. Add integration tests that parse and evaluate actual destructuring patterns
2. Create `tests/fixtures/scripting/destructure.typ` with comprehensive patterns
3. Add tests for `Binding.Write()` error case and `ImmutableBindingError`

**Future Considerations:**
- Consider property-based testing for destructuring patterns
- Add fuzzing tests for malformed patterns
- Benchmark destructuring performance with large arrays/dicts
