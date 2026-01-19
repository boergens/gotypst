# Phase 2: typst-eval Deep Analysis

This document provides a comprehensive analysis of Typst's evaluation crate (`typst-eval`) to prepare for Go translation.

## Table of Contents

1. [Overview and Architecture](#overview-and-architecture)
2. [Evaluation Pipeline Flow](#evaluation-pipeline-flow)
3. [Key Data Structures](#key-data-structures)
4. [Expression Evaluation Dispatch Pattern](#expression-evaluation-dispatch-pattern)
5. [Closure Capture Mechanism](#closure-capture-mechanism)
6. [Module/Import System](#moduleimport-system)
7. [Integration Points](#integration-points)
8. [Go Translation Challenges](#go-translation-challenges)

---

## Overview and Architecture

The `typst-eval` crate is Typst's code interpreter, responsible for transforming parsed AST nodes into runtime values. It is a **tree-walking interpreter** that directly traverses the syntax tree during evaluation.

### Source Files Structure

```
typst-eval/src/
├── lib.rs        # Entry points: eval(), eval_string(), Eval trait
├── vm.rs         # Virtual machine state
├── code.rs       # Code block and expression evaluation
├── call.rs       # Function calls and closure evaluation
├── flow.rs       # Control flow (break, continue, return)
├── import.rs     # Module imports and includes
├── access.rs     # Mutable value access
├── binding.rs    # Variable binding and destructuring
├── ops.rs        # Unary and binary operators
├── methods.rs    # Built-in mutating methods
├── markup.rs     # Markup element evaluation
├── math.rs       # Math mode evaluation
└── rules.rs      # Set/show rule evaluation
```

### Dependencies

- **typst-syntax**: Provides AST types (`ast::*`), parsing, and `Span`
- **typst-library**: Provides runtime types (`Value`, `Scope`, `Content`, etc.)
- **comemo**: Memoization framework for caching evaluation results
- **ecow**: Efficient copy-on-write strings/vectors

---

## Evaluation Pipeline Flow

### Entry Point: `eval()`

```rust
pub fn eval(
    routines: &Routines,
    world: Tracked<dyn World + '_>,
    traced: Tracked<Traced>,
    sink: TrackedMut<Sink>,
    route: Tracked<Route>,
    source: &Source,
) -> SourceResult<Module>
```

**Pipeline Steps:**

1. **Cycle Detection**: Check if source file is already in the evaluation route
2. **Engine Preparation**: Create `Engine` with world, introspector, traced, sink, route
3. **VM Initialization**: Create `Vm` with engine, context, scopes, and root span
4. **Error Check**: Report parser errors unless in trace mode
5. **Evaluation**: Cast root to `ast::Markup` and call `eval()`
6. **Flow Check**: Ensure no forbidden control flow (break/continue/return at top level)
7. **Module Assembly**: Create `Module` with name, scope, and content

### Alternative Entry: `eval_string()`

Used for evaluating string snippets with different syntax modes:
- `SyntaxMode::Code` - Evaluates as code expression
- `SyntaxMode::Markup` - Evaluates as markup content
- `SyntaxMode::Math` - Evaluates as math equation

### Core Evaluation Loop

In `code.rs`, the `eval_code()` function processes expression streams:

```rust
fn eval_code<'a>(
    vm: &mut Vm,
    exprs: &mut impl Iterator<Item = ast::Expr<'a>>,
) -> SourceResult<Value> {
    while let Some(expr) = exprs.next() {
        // Handle set/show rules specially (they affect remaining expressions)
        // Join values with ops::join()
        // Check for flow events (break/continue/return)
    }
}
```

**Key insight**: Set and show rules "capture" the remaining expressions in their scope:

```rust
ast::Expr::SetRule(set) => {
    let styles = set.eval(vm)?;
    let tail = eval_code(vm, exprs)?;  // Evaluate rest with styles applied
    Value::Content(tail.display().styled_with_map(styles))
}
```

---

## Key Data Structures

### Vm (Virtual Machine)

```rust
pub struct Vm<'a> {
    pub engine: Engine<'a>,           // Underlying typesetter
    pub flow: Option<FlowEvent>,       // Current control flow event
    pub scopes: Scopes<'a>,           // Stack of variable scopes
    pub inspected: Option<Span>,       // Span being traced (IDE support)
    pub context: Tracked<'a, Context<'a>>,  // Contextual data
}
```

**Key Methods:**
- `define(var, value)` - Bind value to identifier in current scope
- `bind(var, binding)` - Insert binding with metadata
- `trace(value)` - Record value for IDE inspection
- `world()` - Access the world (file system, packages, etc.)

### Scopes (Scope Stack)

```rust
pub struct Scopes<'a> {
    pub top: Scope,                    // Active scope
    pub scopes: Vec<Scope>,           // Lower scopes (stack)
    base: Option<&'a Scope>,          // Standard library scope
}
```

**Operations:**
- `enter()` - Push new scope onto stack
- `exit()` - Pop scope from stack
- `get(&ident)` - Look up binding by identifier
- `get_mut(&ident)` - Get mutable reference to binding

### Scope

```rust
pub struct Scope {
    map: IndexMap<EcoString, Binding>,  // Variable bindings
    deduplicate: bool,                   // Prevent duplicate definitions
    category: Option<Category>,          // Metadata for docs
}
```

### Value (Runtime Values)

```rust
pub enum Value {
    // Primitives
    None, Auto, Bool(bool), Int(i64), Float(f64),

    // Measurements
    Length(Length), Angle(Angle), Ratio(Ratio),
    Relative(Relative<Length>), Fraction(Fraction),

    // Data
    Str(Str), Bytes(Bytes), Label(Label),
    Datetime(Datetime), Duration(Duration), Decimal(Decimal),

    // Visual
    Color(Color), Gradient(Gradient), Tiling(Tiling), Symbol(Symbol),

    // Collections
    Content(Content), Array(Array), Dict(Dict),

    // Callables
    Func(Func), Args(Args), Type(Type), Module(Module),

    // Dynamic
    Dyn(Dynamic),
    Styles(Styles), Version(Version),
}
```

### FlowEvent (Control Flow)

```rust
pub enum FlowEvent {
    Break(Span),
    Continue(Span),
    Return(Span, Option<Value>, bool),  // span, value, is_conditional
}
```

### Closure

```rust
pub struct Closure {
    pub node: ClosureNode,         // AST node (Closure or Context)
    pub defaults: Vec<Value>,       // Default values for named params
    pub captured: Scope,            // Captured variable bindings
    pub num_pos_params: usize,      // Number of positional parameters
}
```

---

## Expression Evaluation Dispatch Pattern

The evaluation system uses Rust's **trait-based dispatch** via the `Eval` trait:

```rust
pub trait Eval {
    type Output;
    fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output>;
}
```

### Implementation Pattern

Each AST node type implements `Eval` with appropriate output type:

```rust
impl Eval for ast::Expr<'_> {
    type Output = Value;

    fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output> {
        match self {
            Self::Text(v) => v.eval(vm).map(Value::Content),
            Self::Ident(v) => v.eval(vm),
            Self::Int(v) => v.eval(vm),
            Self::FuncCall(v) => v.eval(vm),
            // ... 50+ variants
        }
    }
}
```

### Expression Types by Category

**Literals:**
```rust
impl Eval for ast::Int<'_> {
    type Output = Value;
    fn eval(self, _: &mut Vm) -> SourceResult<Self::Output> {
        Ok(Value::Int(self.get()))
    }
}
```

**Identifiers:**
```rust
impl Eval for ast::Ident<'_> {
    type Output = Value;
    fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output> {
        vm.scopes.get(&self).at(span)?.read_checked(...).clone()
    }
}
```

**Binary Operations:**
```rust
impl Eval for ast::Binary<'_> {
    type Output = Value;
    fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output> {
        match self.op() {
            ast::BinOp::Add => apply_binary(self, vm, ops::add),
            ast::BinOp::Assign => apply_assignment(self, vm, |_, b| Ok(b)),
            // ...
        }
    }
}
```

**Function Calls:**
```rust
impl Eval for ast::FuncCall<'_> {
    type Output = Value;
    fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output> {
        // 1. Check call depth
        // 2. Handle field access calls specially (methods)
        // 3. Evaluate callee and arguments
        // 4. Cast callee to Func
        // 5. Call func.call(engine, context, args)
    }
}
```

### Short-Circuit Evaluation

Boolean operators use short-circuit evaluation:

```rust
fn apply_binary(...) -> SourceResult<Value> {
    let lhs = binary.lhs().eval(vm)?;

    // Short-circuit for && and ||
    if (binary.op() == ast::BinOp::And && lhs == false.into_value())
        || (binary.op() == ast::BinOp::Or && lhs == true.into_value())
    {
        return Ok(lhs);
    }

    let rhs = binary.rhs().eval(vm)?;
    op(lhs, rhs).at(binary.span())
}
```

---

## Closure Capture Mechanism

### Capture Visitor

The `CapturesVisitor` performs static analysis to determine which variables a closure needs to capture:

```rust
pub struct CapturesVisitor<'a> {
    external: Option<&'a Scopes<'a>>,  // Scopes to capture from
    internal: Scopes<'a>,               // Locally bound names
    captures: Scope,                    // Collected captures
    capturer: Capturer,                 // Function or Context
}
```

### Capture Algorithm

1. **Visit all AST nodes** in the closure body
2. **For identifiers**: Check if defined internally; if not, capture from external scope
3. **For bindings**: Add to internal scope (let, for, closure params)
4. **For scopes**: Enter/exit scope for code/content blocks

```rust
fn visit(&mut self, node: &SyntaxNode) {
    match node.cast() {
        Some(ast::Expr::Ident(ident)) => self.capture(ident.get(), Scopes::get),

        Some(ast::Expr::Closure(expr)) => {
            // Visit default values BEFORE entering scope
            for param in expr.params().children() {
                if let ast::Param::Named(named) = param {
                    self.visit(named.expr().to_untyped());
                }
            }

            self.internal.enter();
            // Bind parameters
            for param in expr.params().children() {
                self.bind_param(param);
            }
            self.visit(expr.body().to_untyped());
            self.internal.exit();
        }

        // ...
    }
}
```

### Closure Creation

```rust
impl Eval for ast::Closure<'_> {
    type Output = Value;
    fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output> {
        // 1. Evaluate default values for named parameters
        let mut defaults = Vec::new();
        for param in self.params().children() {
            if let ast::Param::Named(named) = param {
                defaults.push(named.expr().eval(vm)?);
            }
        }

        // 2. Collect captured variables
        let captured = {
            let mut visitor = CapturesVisitor::new(Some(&vm.scopes), Capturer::Function);
            visitor.visit(self.to_untyped());
            visitor.finish()
        };

        // 3. Create closure
        let closure = Closure {
            node: ClosureNode::Closure(self.to_untyped().clone()),
            defaults,
            captured,
            num_pos_params: count_pos_params(self),
        };

        Ok(Value::Func(Func::from(closure).spanned(self.params().span())))
    }
}
```

### Closure Invocation

The memoized `eval_closure` function:

```rust
pub fn eval_closure(
    func: &Func,
    closure: &LazyHash<Closure>,
    // ... engine params ...
    mut args: Args,
) -> SourceResult<Value> {
    // 1. Create fresh scopes with captured variables (NOT call-site scopes)
    let mut scopes = Scopes::new(None);
    scopes.top = closure.captured.clone();

    // 2. Create new VM
    let mut vm = Vm::new(engine, context, scopes, body.span());

    // 3. Bind function name for recursion
    if let Some(name) = name {
        vm.define(name, func.clone());
    }

    // 4. Bind parameters from arguments
    for param in params.children() {
        match param {
            ast::Param::Pos(pattern) => {
                destructure(&mut vm, pattern, args.expect("...")?)
            }
            ast::Param::Spread(spread) => { /* collect rest */ }
            ast::Param::Named(named) => {
                let value = args.named(&name)?.unwrap_or(default.clone());
                vm.define(name, value);
            }
        }
    }

    // 5. Evaluate body
    let output = body.eval(&mut vm)?;

    // 6. Handle return flow
    match vm.flow {
        Some(FlowEvent::Return(_, Some(explicit), _)) => Ok(explicit),
        // ...
    }
}
```

---

## Module/Import System

### Import Types

1. **Package imports** (`@package/name:version`)
2. **File imports** (relative paths)
3. **Module imports** (from functions/types with scopes)

### Import Evaluation

```rust
impl Eval for ast::ModuleImport<'_> {
    type Output = Value;
    fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output> {
        let source = source_expr.eval(vm)?;

        match &source {
            Value::Func(func) => { /* import from function scope */ }
            Value::Type(_) => { /* import from type scope */ }
            Value::Module(_) => { /* already a module */ }
            Value::Str(path) => {
                // Resolve path and import
                source = Value::Module(import(&mut vm.engine, path, span)?);
            }
            // ...
        }

        // Handle import variants:
        // - Bare import: `import "file.typ"` -> bind module name
        // - Renamed: `import "file.typ" as x` -> bind as x
        // - Wildcard: `import "file.typ": *` -> bind all exports
        // - Items: `import "file.typ": a, b` -> bind specific items
    }
}
```

### Import Resolution

```rust
pub fn import(engine: &mut Engine, from: &str, span: Span) -> SourceResult<Module> {
    if from.starts_with('@') {
        // Package import
        let spec = from.parse::<PackageSpec>().at(span)?;
        import_package(engine, spec, span)
    } else {
        // File import
        let path = resolve_path(from, span)?;
        import_file(engine, path, span)
    }
}

fn import_file(engine: &mut Engine, id: FileId, span: Span) -> SourceResult<Module> {
    let source = engine.world.source(id).at(span)?;

    // Prevent cyclic imports
    if engine.route.contains(source.id()) {
        bail!(span, "cyclic import");
    }

    // Recursively evaluate the file
    eval(engine.routines, engine.world, ..., &source)
}
```

### Package Resolution

```rust
fn resolve_package(engine: &mut Engine, spec: PackageSpec, span: Span)
    -> SourceResult<(EcoString, FileId)>
{
    // 1. Load typst.toml manifest
    let manifest_id = /* ... */;
    let bytes = engine.world.file(manifest_id).at(span)?;
    let manifest: PackageManifest = toml::from_str(string)?;

    // 2. Validate package spec
    manifest.validate(&spec).at(span)?;

    // 3. Return entry point
    Ok((manifest.package.name, entry_point_id))
}
```

---

## Integration Points

### typst-syntax Integration

**AST Types Used:**
- `ast::Expr` - All expression variants
- `ast::Markup`, `ast::Math`, `ast::Code` - Top-level containers
- `ast::Ident`, `ast::FuncCall`, `ast::Closure`, etc.
- `Span` - Source location tracking
- `SyntaxNode` - Generic node for visitor patterns

**Parsing Functions:**
- `parse()` - Parse markup
- `parse_code()` - Parse code expression
- `parse_math()` - Parse math expression

### typst-library Integration

**Foundation Types:**
- `Value`, `Func`, `Args`, `Module`, `Scope`, `Binding`
- `Content`, `Array`, `Dict`, `Str`, `Bytes`
- `Context`, `Styles`, `Recipe`

**Engine Types:**
- `Engine` - Core typesetting engine
- `World` - File system and package access
- `Route` - Evaluation route for cycle detection
- `Sink` - Warning/error collection
- `Traced` - IDE tracing support

**Math/Markup Elements:**
- All `*Elem` types (TextElem, StrongElem, EquationElem, etc.)

### comemo Integration

Memoization is critical for performance:

```rust
#[comemo::memoize]
pub fn eval(...) -> SourceResult<Module> { ... }

#[comemo::memoize]
pub fn eval_closure(...) -> SourceResult<Value> { ... }
```

This allows caching of:
- Module evaluation (file imports)
- Closure calls with same arguments

---

## Go Translation Challenges

### 1. Trait-Based Dispatch

**Rust:**
```rust
trait Eval {
    type Output;
    fn eval(self, vm: &mut Vm) -> SourceResult<Self::Output>;
}
```

**Go Challenge:** No generics on interfaces (type erasure).

**Solution Options:**
- Use interface with `Eval() (Value, error)` and type assertions
- Use a large switch statement on AST node types
- Use visitor pattern with explicit methods per node type

```go
// Option A: Interface with type assertions
type Evaluator interface {
    Eval(vm *Vm) (Value, error)
}

// Option B: Switch dispatch
func evalExpr(vm *Vm, expr ast.Expr) (Value, error) {
    switch e := expr.(type) {
    case *ast.IntLiteral:
        return evalInt(vm, e)
    case *ast.FuncCall:
        return evalFuncCall(vm, e)
    // ...50+ cases
    }
}
```

### 2. Lifetime Management

**Rust:** Uses `'a` lifetimes extensively:
```rust
pub struct Vm<'a> {
    pub scopes: Scopes<'a>,
    pub context: Tracked<'a, Context<'a>>,
}
```

**Go Challenge:** No lifetimes; GC handles memory.

**Solution:**
- Use explicit ownership or reference counting where needed
- Be careful with closure captures (may keep references alive)
- Consider using `sync.Pool` for VM reuse

### 3. Memoization

**Rust:** `#[comemo::memoize]` provides automatic memoization with tracking.

**Go Challenge:** No equivalent automatic memoization.

**Solution Options:**
- Use explicit caching with `sync.Map` or LRU cache
- Implement content-addressed caching (hash inputs)
- Consider libraries like `golang.org/x/sync/singleflight`

```go
type EvalCache struct {
    modules sync.Map  // FileId -> *Module
}

func (c *EvalCache) EvalFile(id FileId, eval func() (*Module, error)) (*Module, error) {
    if v, ok := c.modules.Load(id); ok {
        return v.(*Module), nil
    }
    result, err := eval()
    if err == nil {
        c.modules.Store(id, result)
    }
    return result, err
}
```

### 4. Sum Types (Enums with Data)

**Rust:** Rich enums with pattern matching:
```rust
enum Value {
    None,
    Bool(bool),
    Int(i64),
    Array(Array),
    Func(Func),
    // ...
}
```

**Go Challenge:** No sum types.

**Solution Options:**

**Option A: Interface with type assertions**
```go
type Value interface {
    valueMarker()
}

type IntValue int64
func (IntValue) valueMarker() {}

type ArrayValue []Value
func (ArrayValue) valueMarker() {}
```

**Option B: Struct with kind tag**
```go
type ValueKind int
const (
    ValueNone ValueKind = iota
    ValueBool
    ValueInt
    // ...
)

type Value struct {
    Kind   ValueKind
    Bool   bool
    Int    int64
    Array  []Value
    Func   *Func
    // ... (wasteful but fast)
}
```

**Option C: Interface with methods**
```go
type Value interface {
    Type() Type
    Display() Content
    // Common operations
}
```

### 5. Error Handling with Source Locations

**Rust:** Rich error types with spans:
```rust
type SourceResult<T> = Result<T, EcoVec<SourceDiagnostic>>;
```

**Go Challenge:** Errors are simple interfaces.

**Solution:**
```go
type SourceError struct {
    Span    Span
    Message string
    Hints   []string
}

func (e *SourceError) Error() string {
    return fmt.Sprintf("%s: %s", e.Span, e.Message)
}

type SourceResult[T any] struct {
    Value  T
    Errors []SourceError
}
```

### 6. Closure Captures

**Rust:** Automatic capture analysis with `CapturesVisitor`.

**Go Challenge:** Go closures capture by reference automatically, but we need explicit control.

**Solution:**
```go
type Closure struct {
    Node        *ast.Closure
    Defaults    []Value
    Captured    *Scope  // Explicitly captured scope
    NumPosParams int
}

func (c *Closure) Call(vm *Vm, args *Args) (Value, error) {
    // Create new scope with captured variables
    newScope := c.Captured.Clone()
    // ...
}
```

### 7. Mutability Patterns

**Rust:** Uses `&mut` for mutable access:
```rust
fn access<'a>(self, vm: &'a mut Vm) -> SourceResult<&'a mut Value>
```

**Go Challenge:** No reference types with lifetime guarantees.

**Solution:**
- Use pointers explicitly
- Be careful with concurrent access
- Consider copy-on-write for immutable-by-default semantics

```go
func (a *ArrayValue) AtMut(index int) (*Value, error) {
    if index < 0 || index >= len(*a) {
        return nil, outOfBoundsError(index)
    }
    return &(*a)[index], nil
}
```

### 8. Pattern Matching and Destructuring

**Rust:** Native pattern matching:
```rust
match expr {
    ast::Expr::FuncCall(call) => call.eval(vm),
    // ...
}
```

**Go:** Type switches are less ergonomic:
```go
switch e := expr.(type) {
case *ast.FuncCall:
    return evalFuncCall(vm, e)
}
```

### 9. Stack Overflow Protection

**Rust:** Uses `stacker` crate for stack growth:
```rust
stacker::maybe_grow(32 * 1024, 2 * 1024 * 1024, f)
```

**Go Challenge:** Go has growable stacks by default, but still need depth limits.

**Solution:**
```go
const MaxCallDepth = 256

func (vm *Vm) CheckCallDepth() error {
    if vm.callDepth >= MaxCallDepth {
        return errors.New("maximum call depth exceeded")
    }
    return nil
}
```

### 10. Tracked/TrackedMut Pattern

**Rust:** Comemo's tracking for incremental computation:
```rust
Tracked<dyn World + '_>
TrackedMut<Sink>
```

**Go Challenge:** No equivalent pattern.

**Solution:**
- Use explicit interfaces
- Implement dirty tracking manually if needed for caching

---

## Recommended Go Architecture

### Package Structure

```
gotypst/
├── eval/
│   ├── vm.go           // Vm struct and methods
│   ├── eval.go         // Main eval functions
│   ├── expr.go         // Expression evaluation
│   ├── call.go         // Function calls
│   ├── flow.go         // Control flow
│   ├── import.go       // Module imports
│   ├── capture.go      // Closure capture analysis
│   ├── access.go       // Mutable access
│   ├── binding.go      // Destructuring
│   ├── ops.go          // Operators
│   ├── markup.go       // Markup evaluation
│   └── math.go         // Math evaluation
├── value/
│   ├── value.go        // Value type and variants
│   ├── scope.go        // Scope and Scopes
│   ├── func.go         // Func and Closure
│   └── ...
└── syntax/
    └── ...             // AST types (from Phase 1)
```

### Key Design Decisions

1. **Value Representation**: Use interface-based approach with common methods
2. **Dispatch**: Use switch statements in eval functions (clearer than reflection)
3. **Error Handling**: Custom error types with span information
4. **Caching**: Explicit LRU caches for modules and closures
5. **Immutability**: Default to immutable; explicit methods for mutation

---

## Summary

The typst-eval crate is a sophisticated tree-walking interpreter with:

- **50+ expression types** dispatched via trait implementations
- **Lexical scoping** with explicit capture for closures
- **Control flow** via flow events (break/continue/return)
- **Module system** with cycle detection and package support
- **Memoization** for performance optimization

Key Go translation challenges center around:
1. Replacing trait dispatch with switches/interfaces
2. Managing lifetimes without compiler support
3. Implementing memoization explicitly
4. Representing sum types idiomatically

The recommended approach is to start with a simple switch-based dispatcher and interface-based values, adding optimizations incrementally.
