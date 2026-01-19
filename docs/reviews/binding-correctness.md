# Binding/Destructuring Correctness Review

**Reviewer:** obsidian
**Date:** 2026-01-19
**Status:** Issues Found

## Overview

This review compares the Go binding/destructuring implementation against the Rust Typst source (`typst-eval/src/binding.rs`) to verify translation accuracy.

## Critical Finding: Duplicate Implementations

**Severity: High**

The codebase contains TWO separate destructuring implementations:

1. **`eval/binding.go`** - Complete implementation that closely follows Rust:
   - `Destructure()` - creates new bindings
   - `DestructureAssign()` - updates existing bindings
   - `DestructureImpl()` - core recursive implementation
   - `destructureArray()` - handles array patterns with spread
   - `destructureDict()` - handles dict patterns with named/shorthand syntax

2. **`eval/expr.go`** - Incomplete implementation that is **actually being used**:
   - `destructure()` (lowercase) - simple version
   - `destructureComplex()` - incomplete array/dict handling
   - `destructureAssign()` - array-only, incomplete

**The good implementation in `binding.go` is dead code.** All callers use the incomplete `expr.go` version:
- `evalLetBinding` (expr.go:1236)
- `evalForLoop` (expr.go:1530, 1551, 1577, 1603)
- Function parameter binding (call.go:341)

## Issue Details

### 1. Missing Flow State Check in `evalLetBinding`

**Location:** `eval/expr.go:1201-1237`
**Severity:** High

**Rust behavior:**
```rust
fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output> {
    let value = match self.init() {
        Some(expr) => expr.eval(vm)?,
        None => Value::None,
    };
    if vm.flow.is_some() {  // <-- This check is missing in Go
        return Ok(Value::None);
    }
    // ... destructuring
}
```

**Go behavior:** After evaluating the init expression, Go immediately proceeds to destructuring without checking if a control flow event (return/break/continue) occurred during evaluation. This could cause unexpected binding behavior when the init expression triggers control flow.

### 2. Missing Value for Uninitialized Let Bindings

**Location:** `eval/expr.go:1224-1227`
**Severity:** Medium

**Rust behavior:**
```rust
let value = match self.init() {
    Some(expr) => expr.eval(vm)?,
    None => Value::None,  // Uses None as value, still binds
};
// Then destructures with this value
```

**Go behavior:**
```go
init := e.Init()
if init == nil {
    return None, nil  // Returns early, no binding occurs
}
```

When a let binding has no initializer (`let x`), Rust binds `None` to the pattern. Go returns early without creating any binding.

### 3. Spread Pattern Not Supported in `destructureComplex`

**Location:** `eval/expr.go:1290-1325`
**Severity:** High

**Rust behavior:** Handles `DestructuringItem::Spread` by:
1. Calculating sink size: `(1 + len).checked_sub(items.count())`
2. Extracting slice of remaining elements
3. Binding to sink pattern if present

**Go behavior:** The `destructureComplex` function only handles `DestructuringBinding`, completely ignoring `DestructuringSpread`. Patterns like `let (a, ..rest, b) = arr` will not work correctly.

### 4. No Length Mismatch Errors in Array Destructuring

**Location:** `eval/expr.go:1296-1305`
**Severity:** High

**Rust behavior:**
- Checks `i < len` after iteration
- Calls `bail!(wrong_number_of_elements(...))` if mismatched

**Go behavior:**
```go
for i, item := range items {
    if i < len(v) {  // Just skips if out of bounds
        // ...
    }
}
// No check after loop
```

Arrays with more elements than patterns silently ignore extras. Arrays with fewer elements silently skip bindings. Both should error.

### 5. Dict Shorthand Syntax Not Supported

**Location:** `eval/expr.go:1310-1321`
**Severity:** High

**Rust behavior:** Handles both forms:
```rust
// Shorthand: let (name, age) = dict
DestructuringItem::Pattern(Pattern::Normal(Expr::Ident(ident))) => {
    let v = dict.get(&ident).at(ident.span())?;
    f(vm, Expr::Ident(ident), v.clone())?;
}
// Named: let (n: name, a: age) = dict
DestructuringItem::Named(named) => { ... }
```

**Go behavior:** Only handles `DestructuringNamed`:
```go
for _, item := range items {
    if named, ok := item.(*syntax.DestructuringNamed); ok {
        // Only named patterns work
    }
}
```

The shorthand form `let (name) = dict` (which looks up key "name") does not work.

### 6. Dict Destructuring Missing from `destructureAssign`

**Location:** `eval/expr.go:1328-1352`
**Severity:** Medium

**Rust behavior:** `DestructAssignment` uses `destructure_impl` which handles both arrays and dicts.

**Go behavior:** `destructureAssign` only handles `ArrayValue`:
```go
switch v := value.(type) {
case ArrayValue:
    // ... only array logic
}
// No case for DictValue
```

Reassignment to dict destructuring patterns (`(a, b) = dict`) will silently do nothing.

### 7. Missing Error for Named Patterns on Arrays

**Location:** `eval/expr.go:1296-1304`
**Severity:** Medium

**Rust behavior:**
```rust
DestructuringItem::Named(named) => {
    bail!(named.span(), "cannot destructure named pattern from an array")
}
```

**Go behavior:** No check for `DestructuringNamed` in array context - these are silently ignored.

### 8. Missing Error for Non-Identifier Patterns on Dicts

**Location:** `eval/expr.go:1310-1321`
**Severity:** Medium

**Rust behavior:**
```rust
DestructuringItem::Pattern(expr) => {
    bail!(expr.span(), "cannot destructure unnamed pattern from dictionary");
}
```

**Go behavior:** Non-identifier patterns are silently ignored rather than producing an error.

## Comparison: `binding.go` vs Rust

The **unused** `binding.go` implementation correctly handles:

| Feature | Rust | binding.go | expr.go (used) |
|---------|------|------------|----------------|
| Spread patterns | ✓ | ✓ | ✗ |
| Length mismatch errors | ✓ | ✓ | ✗ |
| Dict shorthand | ✓ | ✓ | ✗ |
| Named on array error | ✓ | ✓ | ✗ |
| Unnamed on dict error | ✓ | ✓ | ✗ |
| Nested patterns | ✓ | ✓ | ✗ |
| Dict reassignment | ✓ | ✓ | ✗ |

## Recommendations

### Immediate Fix

Replace calls to `expr.go` destructure functions with the proper `binding.go` implementations:

1. Change `evalLetBinding` (expr.go:1236):
   ```go
   // Before
   return destructure(vm, pattern, value)

   // After
   if err := Destructure(vm, pattern, value); err != nil {
       return nil, err
   }
   return None, nil
   ```

2. Change `evalDestructAssignment` (expr.go:1258):
   ```go
   // Before
   return destructureAssign(vm, pattern, value)

   // After
   if err := DestructureAssign(vm, patternAsPattern, value); err != nil {
       return nil, err
   }
   return None, nil
   ```

3. Update `evalLetBinding` to add flow check and handle nil init properly.

4. Update all for-loop destructuring calls in `evalForLoop`.

5. Verify call.go parameter binding uses correct implementation.

### After Fix

Remove the dead code from `expr.go`:
- `destructure()`
- `destructureComplex()`
- `destructureAssign()`

## Test Coverage Gap

The test file `eval/binding_test.go` tests the `binding.go` implementation directly using `vm.DefineWithSpan` to simulate destructuring, but doesn't test actual destructuring through parsed syntax. Integration tests parsing `let (a, b) = [1, 2]` would have caught these issues.

## Files Reviewed

- `eval/binding.go` - Good implementation, not used
- `eval/binding_test.go` - Tests binding.go directly
- `eval/expr.go:1201-1352` - Broken implementation in use
- `eval/call.go:335-345` - Parameter binding
- `syntax/pattern.go` - Pattern AST types

## Rust Reference

Source: `github.com/typst/typst/crates/typst-eval/src/binding.rs` (main branch)
