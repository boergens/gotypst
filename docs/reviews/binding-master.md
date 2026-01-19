# Binding System Master Review

**Synthesized by:** obsidian
**Date:** 2026-01-19
**Source Reviews:**
- hq-xx2o: Binding Correctness Review (obsidian)
- hq-92tz: Binding Test Coverage Review (quartz)
- hq-znbk: Binding Integration Review (jasper)

## Executive Summary

All three reviews independently identified the same critical issue: **the codebase contains two destructuring implementations, and the wrong one is being used**. The complete, correct implementation in `eval/binding.go` is dead code, while the incomplete `eval/expr.go` implementation is used by all callers.

This is the single most impactful finding and fixing it will resolve 7 of the 8 correctness issues identified.

## Critical Finding: Wrong Implementation in Use

| Implementation | Location | Status | Features |
|---------------|----------|--------|----------|
| **binding.go** | `Destructure()`, `DestructureAssign()`, `DestructureImpl()` | DEAD CODE | Spread patterns, length checks, dict shorthand, all error cases |
| **expr.go** | `destructure()`, `destructureComplex()`, `destructureAssign()` | IN USE | Basic patterns only, no spreads, no length checks, silent failures |

**Callers using the broken expr.go version:**
- `evalLetBinding` (expr.go:1236)
- `evalDestructAssignment` (expr.go:1258)
- `evalForLoop` (expr.go:1530, 1551, 1577, 1603)
- Function parameter binding (call.go:341)

## Consolidated Issue List

### Priority 1: Critical (Blocks core functionality)

| ID | Issue | Severity | Source |
|----|-------|----------|--------|
| B1 | **Switch to binding.go implementation** | Critical | All reviews |
| B2 | Missing flow state check in evalLetBinding | High | Correctness |
| B3 | Uninitialized let bindings return early instead of binding None | Medium | Correctness |

### Priority 2: High (Affects pattern support)

These are all **resolved by B1** (switching to binding.go):

| ID | Issue | Current Behavior | Correct Behavior |
|----|-------|------------------|------------------|
| P1 | Spread patterns not supported | Ignored | `let (a, ..rest) = arr` works |
| P2 | No length mismatch errors | Silent ignore/skip | Error on too few/many elements |
| P3 | Dict shorthand not supported | Ignored | `let (name) = dict` looks up "name" |
| P4 | Named pattern on array error missing | Silent ignore | Error message |
| P5 | Unnamed pattern on dict error missing | Silent ignore | Error message |
| P6 | Dict destructuring missing from assignment | Silent no-op | `(a, b) = dict` works |
| P7 | Nested patterns limited | Partial | Full recursive support |

### Priority 3: Medium (Test coverage gaps)

| ID | Issue | Impact |
|----|-------|--------|
| T1 | Core destructure functions not unit tested | Regressions possible |
| T2 | No integration tests for complex patterns | End-to-end bugs undetected |
| T3 | Tests simulate behavior rather than test actual code paths | False confidence |

### Priority 4: Low (Optimization)

| ID | Issue | Impact |
|----|-------|--------|
| O1 | Scope capture captures everything | Higher memory for closures |

## Fix Plan

### Fix 1: Switch Destructure Implementations (Resolves B1, P1-P7)

**Scope:** Change all callers to use `binding.go` functions

**Changes Required:**

1. **evalLetBinding** (expr.go ~line 1236):
   ```go
   // Before
   return destructure(vm, pattern, value)

   // After
   if err := Destructure(vm, pattern, value); err != nil {
       return nil, err
   }
   return None, nil
   ```

2. **evalDestructAssignment** (expr.go ~line 1258):
   ```go
   // Before
   return destructureAssign(vm, pattern, value)

   // After
   if err := DestructureAssign(vm, pattern, value); err != nil {
       return nil, err
   }
   return None, nil
   ```

3. **evalForLoop** (expr.go, multiple locations):
   - Replace all `destructure()` calls with `Destructure()`

4. **callClosure parameter binding** (call.go ~line 341):
   - Replace `destructure()` with `Destructure()`

### Fix 2: Add Flow Check to evalLetBinding (Resolves B2)

**Location:** expr.go, evalLetBinding function

```go
value, err := vm.Eval(init)
if err != nil {
    return nil, err
}
// ADD THIS CHECK:
if vm.Flow != nil {
    return None, nil
}
// Continue to destructuring...
```

### Fix 3: Handle Nil Init in evalLetBinding (Resolves B3)

**Location:** expr.go, evalLetBinding function

```go
// Before
if init == nil {
    return None, nil  // Wrong: returns without binding
}

// After
var value Value = None
if init != nil {
    value, err = vm.Eval(init)
    if err != nil {
        return nil, err
    }
}
// Then destructure with value (which is None if no init)
```

### Fix 4: Remove Dead Code

After Fix 1 is complete and tested, remove from expr.go:
- `destructure()` function
- `destructureComplex()` function
- `destructureAssign()` function

### Fix 5: Add Integration Tests (Resolves T1-T3)

Create `tests/fixtures/scripting/destructure.typ` with comprehensive test patterns:
- Array destructuring basic
- Array destructuring with spread
- Dict destructuring shorthand
- Dict destructuring named
- Dict destructuring with spread
- Nested destructuring
- Placeholder patterns
- Error cases (length mismatch, type mismatch)

## Recommended Bead Structure

| Bead | Description | Priority | Depends On |
|------|-------------|----------|------------|
| fix-destructure-switch | Switch to binding.go destructure implementation | P1 | - |
| fix-let-flow-check | Add flow state check to evalLetBinding | P1 | - |
| fix-let-nil-init | Handle nil init in evalLetBinding | P2 | - |
| fix-destructure-cleanup | Remove dead code from expr.go | P2 | fix-destructure-switch |
| add-destructure-tests | Add integration tests for destructuring | P2 | fix-destructure-switch |

## Risk Assessment

| Fix | Risk | Mitigation |
|-----|------|------------|
| Switch implementations | Medium - changes core behavior | Run full test suite; binding.go is already tested |
| Flow check | Low - additive change | Matches Rust behavior exactly |
| Nil init | Low - edge case fix | Add specific test case |
| Remove dead code | Low - after switch verified | Ensure no remaining callers |
| Add tests | None | Pure addition |

## Conclusion

The binding system has a solid foundation in `binding.go`, but it's not being used. The fix is straightforward: wire up the correct implementation and remove the incomplete one. This single change resolves the majority of identified issues.

**Estimated scope:** 4-5 targeted fixes, all within eval package.
