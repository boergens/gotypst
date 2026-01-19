# Binding Integration Review

**Reviewer**: jasper
**Date**: 2026-01-19
**Bead**: hq-znbk

## Summary

The binding/destructuring system is well-integrated across the VM, expression evaluator, and closure system. All existing tests pass. The implementation follows a clean callback-based design that allows code reuse between binding creation and reassignment.

## Integration Analysis

### 1. VM Execution Integration

**Location**: `eval/vm.go:51-74`, `eval/scope.go`

The VM provides clean binding operations:
- `Define(name, value)` - creates immutable binding
- `DefineWithSpan(name, value, span)` - creates binding with source location
- `Bind(name, binding)` - inserts pre-made binding
- `Get(name)` / `GetMut(name)` - lookup bindings

**Scope Stack Design**:
- `Scopes` manages a stack with `top` (current scope) + `scopes[]` (outer scopes) + `base` (stdlib)
- `Enter()` / `Exit()` push/pop scopes
- `Get()` searches top → outer → base
- `FlattenToScope()` merges all scopes for closure capture

**Status**: Working correctly. Tests in `vm_test.go` verify scope shadowing and variable visibility.

### 2. Expression Evaluator Integration

**Location**: `eval/expr.go:1201-1352`

**Let Bindings** (`evalLetBinding`):
- Handles closure bindings (`let f(x) = ...`) - defines function name
- Handles plain bindings (`let x = ...`) - calls `destructure()`
- Properly evaluates init expression before destructuring

**Destructure Assignment** (`evalDestructAssignment`):
- For `(a, b) = expr` - updates existing mutable bindings
- Uses `destructureAssign()` which calls `binding.Write()`

**Pattern Matching** (`destructure`, `destructureComplex`):
- Supports `NormalPattern` (simple ident), `PlaceholderPattern` (_), `DestructuringPattern`, `ParenthesizedPattern`
- Array destructuring iterates items and binds values
- Dict destructuring uses named patterns to extract fields

**For Loop Integration** (`eval/expr.go:1523-1627`):
- Enters single scope for entire loop
- Calls `destructure(vm, pattern, elem)` each iteration
- Properly handles break/continue/return flow events

**Status**: Working correctly. Note: `destructureComplex()` is simpler than `binding.go:Destructure` - it doesn't handle spreads or error cases as thoroughly.

### 3. Closure Integration

**Location**: `eval/expr.go:1126-1179`, `eval/call.go:250-374`

**Closure Creation** (`evalClosure`):
1. Evaluates default values for named params
2. Captures scope via `captureVariables()` → `vm.Scopes.FlattenToScope()`
3. Creates `Closure` struct with `Captured` scope
4. Wraps in `Func` with `ClosureFunc` representation

**Closure Calling** (`callClosure`):
1. Saves current scopes
2. Creates new `Scopes` with nil base
3. Sets top to `closure.Captured.Clone()` (isolated copy)
4. Defines function name for recursion
5. Binds parameters via pattern matching loop
6. Evaluates body
7. Handles return flow event
8. Restores original scopes

**Scope Capture Design**:
- Uses full scope flattening (captures everything)
- Comment notes a proper impl would use static analysis for minimal capture
- Clone on call ensures closures are independent

**Parameter Destructuring** (`call.go:333-344`):
- `DestructuringParam` supported in function signatures
- Calls `destructure()` to bind pattern from argument value

**Status**: Working correctly. The "capture everything" approach is safe but not optimal for memory.

### 4. Binding System Core

**Location**: `eval/binding.go`

**Binding Struct**:
- `Value`, `Span`, `Kind` (Normal/Closure/Module), `Mutable`, `Category`
- Immutability enforced via `Write()` check

**Destructure Functions**:
- `Destructure()` - creates new bindings via `vm.DefineWithSpan()`
- `DestructureAssign()` - updates existing via `AccessExpr()` + assignment
- `DestructureImpl()` - callback-based recursive traversal

**Pattern Support**:
- Arrays: indexed access, spread (`..rest`)
- Dicts: named fields, field shorthand, spread for unused keys
- Comprehensive error types for validation

**Status**: Working correctly. Well-tested with comprehensive error messages.

## Potential Issues / Edge Cases

### 1. Duplicate Destructure Implementations

**Issue**: There are two destructure code paths:
- `binding.go:Destructure/DestructureImpl` - comprehensive with spreads
- `expr.go:destructure/destructureComplex` - simpler, no spread support

**Impact**: The simpler `expr.go` version is used in `evalLetBinding`. This means spread patterns in let bindings may not work correctly.

**Recommendation**: Unify to use `binding.go:Destructure` everywhere.

### 2. Scope Capture Captures Everything

**Issue**: `captureVariables()` returns `vm.Scopes.FlattenToScope()` which captures all accessible variables, not just those referenced by the closure.

**Impact**: Memory usage could be higher than necessary for closures in deeply nested scopes.

**Recommendation**: Low priority - current approach is correct, just not optimal.

### 3. No Integration Tests for Complex Patterns

**Issue**: Tests in `binding_test.go` don't actually parse and evaluate Typst code - they test lower-level functions directly or simulate destructuring.

**Impact**: End-to-end behavior with actual syntax patterns isn't verified.

**Recommendation**: Consider adding integration tests that parse real Typst expressions.

### 4. DestructureAssign Limited Implementation

**Issue**: `expr.go:destructureAssign` only handles array destructuring with `NormalPattern` elements. Dict destructuring and nested patterns aren't supported.

**Impact**: `(a, b) = arr` works, but `(a, (b, c)) = nested` may not.

**Recommendation**: Consider using `binding.go:DestructureAssign` instead.

## Test Coverage

All tests pass:
```
ok  github.com/boergens/gotypst/eval  0.172s
```

Binding-specific tests:
- `TestDestructureArray` - array destructuring basics
- `TestDestructureDict` - dict field access
- `TestDestructuringErrors` - error message verification
- `TestBindingKinds` - normal/mutable/closure/module bindings
- `TestBindingCategory` - documentation categories
- `TestBindingClone` - clone independence
- `TestDestructureImplNilPattern` - nil safety

VM/Scope tests:
- `TestScope` - define/get/contains/names
- `TestScopes` - scope stack shadowing
- `TestVm` - VM binding operations

Call tests:
- `TestCallNativeFunc` - native function calls
- `TestCallDepthLimit` - recursion limit

## Conclusion

The binding system integration is solid. The main concerns are:

1. **Duplicate destructure implementations** - functional difference between `binding.go` and `expr.go` versions
2. **Limited DestructureAssign** - doesn't support nested patterns

These are minor issues that don't affect basic functionality. The core integration between VM, expression evaluator, and closures is working correctly.
