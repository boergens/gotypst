# Typst Macros Analysis

Analysis of the `typst-macros` crate to understand Typst's procedural macro system and develop Go translation strategies.

## Table of Contents

1. [Overview](#overview)
2. [Macro Inventory](#macro-inventory)
3. [Detailed Macro Analysis](#detailed-macro-analysis)
4. [Usage Patterns in typst-library](#usage-patterns-in-typst-library)
5. [Go Translation Strategy](#go-translation-strategy)

---

## Overview

The `typst-macros` crate is a procedural macro library that bridges Rust code with Typst's runtime type system. It provides macros that:

- Generate trait implementations for interoperability with Typst's type system
- Create metadata for documentation and introspection
- Handle argument parsing and type conversion
- Reduce boilerplate for elements, functions, and types

**Source:** https://github.com/typst/typst/tree/main/crates/typst-macros

**Dependencies:**
- `heck` - Case conversion (snake_case to kebab-case, etc.)
- `proc-macro2` - Procedural macro utilities
- `quote` - Code generation
- `syn` - Rust syntax parsing

---

## Macro Inventory

| Macro | Type | Purpose |
|-------|------|---------|
| `#[func]` | Attribute | Makes Rust functions usable as Typst functions |
| `#[ty]` | Attribute | Makes Rust types usable as Typst types |
| `#[elem]` | Attribute | Implements element (content node) functionality |
| `#[scope]` | Attribute | Organizes related functions/constants into namespaces |
| `cast!` | Procedural | Implements type conversion traits |
| `#[derive(Cast)]` | Derive | Auto-derives casting for enums to kebab-case strings |
| `#[time]` | Attribute | Records function timing for performance tracing |

---

## Detailed Macro Analysis

### 1. `#[func]` - Function Macro

**Purpose:** Transforms Rust functions into Typst-callable native functions.

**What it generates:**
- `NativeFunc` trait implementation
- Argument parsing wrapper with type conversion
- Parameter metadata (names, types, defaults, documentation)
- Support for special parameters (engine, context, span, args)

**Attributes:**
```rust
#[func(
    scope,                    // Include in type's scope
    contextual,               // Requires runtime context
    constructor,              // Acts as type constructor
    name = "custom-name",     // Override function name
    title = "Display Title",  // Documentation title
    keywords = ["alias"]      // Search keywords
)]
```

**Field attributes:**
- `#[named]` - Keyword argument (not positional)
- `#[default(value)]` - Default value
- `#[variadic]` - Accepts variable arguments
- `#[external]` - Documentation-only parameter

**Example:**
```rust
#[func(title = "Minimum")]
pub fn min(
    span: Span,                      // Special: injected by runtime
    #[variadic] values: Vec<Value>,  // Variadic positional args
) -> SourceResult<Value>
```

**Generated code structure:**
```rust
impl NativeFunc for min {
    const NAME: &'static str = "min";
    fn data() -> &'static NativeFuncData { /* metadata */ }
}
// Plus wrapper closure for argument parsing
```

---

### 2. `#[ty]` - Type Macro

**Purpose:** Exposes Rust types to Typst's type system.

**What it generates:**
- `NativeType` trait implementation
- Type metadata (name, documentation, keywords)
- Optional scope reference for associated functions
- Default `cast!` implementation (unless `cast` attribute specified)

**Attributes:**
```rust
#[ty(
    scope,           // Has associated scope with methods
    cast,            // Custom casting (skip auto-generated cast!)
    name = "name",   // Override type name
    title = "Title"  // Display title
)]
```

**Example:**
```rust
#[ty(cast)]
#[derive(Default, Clone, PartialEq, Hash)]
pub struct Styles(EcoVec<LazyHash<Style>>);
```

---

### 3. `#[elem]` - Element Macro

**Purpose:** Creates document elements (content nodes) with full Typst integration.

**What it generates:**
- Struct definition (with optional Debug, Hash, Clone derives)
- `new()` constructor
- Builder methods (`with_*()`) for optional fields
- Field accessor traits for different field types
- `NativeElement` trait implementation
- `Construct` trait for instantiation from arguments
- `Set` trait for style configuration
- Capability trait implementations (Synthesize, Locatable, etc.)

**Attributes:**
```rust
#[elem(
    scope,                           // Include in scope
    name = "custom",                 // Override element name
    title = "Display Title",         // Documentation title
    keywords = ["search", "terms"],  // Search keywords
    // Capability traits:
    Synthesize,    // Process content during construction
    Locatable,     // Track document position
    Tagged,        // Enable labeling/referencing
    ShowSet,       // Apply default styling
    LocalName,     // Provide localization
    Figurable,     // Can be wrapped in figure
    PlainText,     // Has plain text representation
    Count,         // Participate in counters
    Refable,       // Support cross-references
    Outlinable     // Include in outlines
)]
```

**Field attributes:**
- `#[required]` - Mandatory field
- `#[positional]` - Positional argument
- `#[default(value)]` - Default value
- `#[synthesized]` - Auto-generated field
- `#[fold]` - Foldable style property
- `#[internal]` - Not exposed to users
- `#[ghost]` - Exists in styles but not struct
- `#[external]` - Documentation-only

**Example:**
```rust
#[elem(scope, title = "Raw Text", Synthesize, Locatable, Tagged)]
pub struct RawElem {
    #[required]
    pub text: RawContent,

    #[default(false)]
    pub block: bool,

    pub lang: Option<EcoString>,

    #[synthesized]
    pub lines: Vec<RawLine>,
}
```

---

### 4. `#[scope]` - Scope Macro

**Purpose:** Groups related functions and constants into a namespace.

**What it generates:**
- `NativeScope` trait implementation
- Registration of each item in the scope
- Constructor identification
- Deprecation metadata handling

**Example:**
```rust
#[scope]
impl Array {
    #[func(constructor)]
    pub fn construct(value: ToArray) -> Array { ... }

    #[func(title = "Length")]
    pub fn len(&self) -> usize { ... }

    #[func]
    pub fn first(&self) -> StrResult<Value> { ... }
}
```

---

### 5. `cast!` - Cast Macro

**Purpose:** Implements type conversion between Rust types and Typst values.

**What it generates:**
- `Reflect` trait - Type introspection
- `FromValue` trait - Convert from Typst value
- `IntoValue` trait - Convert to Typst value

**Syntax:**
```rust
cast! {
    TargetType,

    // IntoValue (optional): how to serialize
    self => self.into_inner().into_value(),

    // FromValue patterns: type => conversion expression
    content: Content => Self::Content(content),
    func: Func => Self::Func(func),

    // String patterns: "literal" => expression
    "start" => Self::Start,
    "end" => Self::End,
}
```

**Example:**
```rust
cast! {
    Transformation,
    content: Content => Self::Content(content),
    func: Func => Self::Func(func),
}
```

---

### 6. `#[derive(Cast)]` - Cast Derive Macro

**Purpose:** Auto-derives casting for simple enums, converting variants to kebab-case strings.

**What it generates:**
- `Reflect`, `FromValue`, `IntoValue` implementations
- Kebab-case string conversion for each variant
- Optional custom string mappings via `#[string]` attribute

**Example:**
```rust
#[derive(Debug, Copy, Clone, Eq, PartialEq, Hash, Cast)]
pub enum VerticalFontMetric {
    Ascender,      // -> "ascender"
    CapHeight,     // -> "cap-height"
    XHeight,       // -> "x-height"
    Baseline,      // -> "baseline"
    Descender,     // -> "descender"
}
```

---

### 7. `#[time]` - Timing Macro

**Purpose:** Records function execution time for performance profiling.

**Behavior:** Wraps function body with timing instrumentation. Disabled on `wasm32` targets.

---

## Usage Patterns in typst-library

### Element Definitions

Elements are the primary content building blocks:

```rust
// From raw.rs
#[elem(scope, title = "Raw Text / Code", Synthesize, Locatable, Tagged, ShowSet, LocalName, Figurable, PlainText)]
pub struct RawElem {
    #[required]
    pub text: RawContent,

    #[default(false)]
    pub block: bool,

    pub lang: Option<EcoString>,
}
```

### Function Definitions

Functions exposed to Typst scripting:

```rust
// From calc.rs
#[func(title = "Absolute")]
pub fn abs(value: ToAbs) -> Value

#[func]
pub fn round(
    value: DecNum,
    #[named]
    #[default(0)]
    digits: i64,
) -> StrResult<DecNum>
```

### Type Scopes

Organizing methods on types:

```rust
// From array.rs
#[scope]
impl Array {
    #[func(constructor)]
    pub fn construct(value: ToArray) -> Array

    #[func(title = "Length")]
    pub fn len(&self) -> usize

    #[func]
    pub fn push(&mut self, value: Value)
}
```

---

## Go Translation Strategy

### Overview

Go lacks Rust's procedural macro system. We have three main approaches:

1. **Code Generation** (`go generate`) - Generate code at build time
2. **Manual Expansion** - Write out what macros would generate
3. **Interface-Based Patterns** - Use Go idioms that achieve similar goals

### Recommendation: Hybrid Approach

For gotypst, we recommend a **hybrid approach** combining manual expansion with targeted code generation for repetitive patterns.

---

### Strategy by Macro

#### `#[func]` → Manual + Registry Pattern

**Approach:** Define functions normally, register them in a central registry with metadata.

```go
// Define the function
func calcAbs(args Args) (Value, error) {
    value, err := args.Get("value")
    if err != nil {
        return nil, err
    }
    // Implementation
}

// Register with metadata
func init() {
    RegisterFunc(FuncMeta{
        Name:     "abs",
        Title:    "Absolute",
        Params: []ParamMeta{
            {Name: "value", Type: TypeNum, Required: true},
        },
        Func: calcAbs,
    })
}
```

**Pros:** Explicit, debuggable, no magic
**Cons:** More boilerplate, manual sync between signature and metadata

---

#### `#[ty]` → Interface + Registry Pattern

**Approach:** Types implement a `NativeType` interface and register themselves.

```go
type NativeType interface {
    TypeName() string
    TypeData() *TypeData
}

type TypeData struct {
    Name     string
    Title    string
    Keywords []string
    Doc      string
}

// Implementation
type Styles struct { /* fields */ }

func (Styles) TypeName() string { return "styles" }
func (Styles) TypeData() *TypeData {
    return &TypeData{
        Name:  "styles",
        Title: "Styles",
    }
}
```

---

#### `#[elem]` → Struct + Interface + Code Generation

**Approach:** This is the most complex macro. Use a combination:

1. **Struct definition** - Manual
2. **Field accessors** - Generated or manual based on complexity
3. **Traits/capabilities** - Go interfaces

```go
// Element interface
type Element interface {
    ElemFunc() ElemFunc  // Returns element type identifier
    Fields() []Field     // Introspection
}

// Capabilities as interfaces
type Locatable interface {
    Location() Location
}

type Synthesizable interface {
    Synthesize(engine *Engine, styles Styles) error
}

// Element definition
type RawElem struct {
    Text  RawContent  // required
    Block bool        // default: false
    Lang  *string     // optional
    Lines []RawLine   // synthesized
}

func (e *RawElem) ElemFunc() ElemFunc { return RawElemFunc }

// Constructor
func NewRawElem(text RawContent) *RawElem {
    return &RawElem{Text: text}
}

// Builder methods
func (e *RawElem) WithBlock(v bool) *RawElem {
    e.Block = v
    return e
}
```

**For code generation**, create a YAML/JSON schema and generate:
- Constructor functions
- Builder methods
- Field accessor implementations
- Registration code

---

#### `#[scope]` → Method Organization (No Direct Equivalent)

**Approach:** Go doesn't need scopes - methods are already organized by receiver type. Use a registration pattern for Typst exposure.

```go
// Methods defined normally on the type
func (a *Array) Len() int { return len(a.items) }
func (a *Array) First() (Value, error) { /* ... */ }

// Register scope for Typst runtime
func init() {
    RegisterScope("array", ScopeMeta{
        Constructor: arrayConstruct,
        Methods: map[string]MethodMeta{
            "len":   {Func: (*Array).Len, Title: "Length"},
            "first": {Func: (*Array).First},
        },
    })
}
```

---

#### `cast!` → Type Conversion Functions

**Approach:** Implement conversion as explicit functions.

```go
// Conversion interface
type FromValue interface {
    FromValue(v Value) error
}

type IntoValue interface {
    IntoValue() Value
}

// For enum-like types with multiple source types
func TransformationFromValue(v Value) (Transformation, error) {
    switch val := v.(type) {
    case Content:
        return TransformationContent{val}, nil
    case Func:
        return TransformationFunc{val}, nil
    default:
        return nil, fmt.Errorf("cannot convert %T to Transformation", v)
    }
}
```

---

#### `#[derive(Cast)]` → Code Generation or Constants

**Approach:** For simple string enums, use `go generate` with `stringer` or similar.

```go
//go:generate stringer -type=VerticalFontMetric -linecomment

type VerticalFontMetric int

const (
    Ascender   VerticalFontMetric = iota // ascender
    CapHeight                            // cap-height
    XHeight                              // x-height
    Baseline                             // baseline
    Descender                            // descender
)

// Parsing function
func ParseVerticalFontMetric(s string) (VerticalFontMetric, error) {
    switch s {
    case "ascender":
        return Ascender, nil
    case "cap-height":
        return CapHeight, nil
    // ...
    }
    return 0, fmt.Errorf("unknown metric: %s", s)
}
```

---

### Code Generation Tools

For repetitive patterns, consider creating a code generator:

**Input (YAML/JSON):**
```yaml
elements:
  - name: RawElem
    title: "Raw Text / Code"
    capabilities: [Synthesize, Locatable, Tagged]
    fields:
      - name: text
        type: RawContent
        required: true
      - name: block
        type: bool
        default: "false"
      - name: lang
        type: "*string"
```

**Output:** Generated Go code with constructors, builders, and trait implementations.

**Tool:** Create `cmd/genelem/main.go` that reads definitions and outputs Go code.

---

### Summary Table

| Rust Macro | Go Strategy | Effort |
|------------|-------------|--------|
| `#[func]` | Registry + manual functions | Medium |
| `#[ty]` | Interface + registry | Low |
| `#[elem]` | Struct + interfaces + code gen | High |
| `#[scope]` | Method organization + registry | Low |
| `cast!` | Conversion functions | Medium |
| `#[derive(Cast)]` | `go generate` + stringer | Low |
| `#[time]` | Middleware/decorator pattern | Low |

---

### Recommended Implementation Order

1. **Start with manual expansion** for the first few elements/functions
2. **Identify patterns** that are truly repetitive
3. **Build code generators** only for validated patterns
4. **Keep metadata explicit** rather than relying on reflection

This approach ensures we understand the generated code deeply before automating it.
