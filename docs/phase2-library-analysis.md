# Typst Standard Library Analysis

This document analyzes the `typst-library` crate to understand Typst's standard library structure, function registration, type system integration, and provides Go translation notes.

## 1. Library Structure Overview

The typst-library crate is organized into 9 functional modules plus 5 core source files:

```
crates/typst-library/src/
├── foundations/    # Core types and functions
├── introspection/  # Document introspection and metadata
├── layout/         # Page and layout primitives
├── loading/        # Data file loading (JSON, YAML, etc.)
├── math/           # Mathematical typesetting
├── model/          # Document model (headings, lists, etc.)
├── pdf/            # PDF output handling
├── text/           # Text processing
├── visualize/      # Graphics and colors
├── diag.rs         # Diagnostics
├── engine.rs       # Execution engine
├── lib.rs          # Library assembly
├── routines.rs     # Core routines
└── symbols.rs      # Symbol definitions
```

## 2. Function Categories

### 2.1 Math Functions (`foundations/calc.rs`)

40+ pure math functions organized into categories:

| Category | Functions |
|----------|-----------|
| **Basic** | `abs`, `pow`, `exp`, `sqrt`, `root` |
| **Trig** | `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2` |
| **Hyperbolic** | `sinh`, `cosh`, `tanh` |
| **Logarithmic** | `log`, `ln` |
| **Combinatorial** | `fact`, `perm`, `binom` |
| **Number Theory** | `gcd`, `lcm` |
| **Rounding** | `floor`, `ceil`, `trunc`, `fract`, `round` |
| **Comparison** | `clamp`, `min`, `max` |
| **Parity** | `even`, `odd` |
| **Division** | `rem`, `div_euclid`, `rem_euclid`, `quo` |
| **Constants** | `inf`, `pi`, `tau`, `e` |

**Go Translation Notes:**
- Use `math` package for basic ops (`math.Abs`, `math.Sqrt`, etc.)
- Implement custom `Num` union type to handle int64/float64 dispatch
- All functions are pure - no state management needed
- Return `(result, error)` for functions that can fail (e.g., `sqrt(-1)`)

### 2.2 String Functions (`foundations/str.rs`)

Strings wrap `EcoString` (copy-on-write). Key methods:

| Category | Methods |
|----------|---------|
| **Construction** | `construct`, `from_unicode`, `to_unicode` |
| **Inspection** | `len`, `is_empty`, `first`, `last`, `at`, `slice` |
| **Decomposition** | `clusters`, `codepoints` |
| **Pattern Matching** | `contains`, `starts_with`, `ends_with`, `find`, `position`, `match`, `matches` |
| **Transformation** | `replace`, `split`, `trim`, `normalize`, `rev`, `repeat` |

**Go Translation Notes:**
- Use `string` type directly (Go strings are immutable, similar semantics)
- Implement grapheme cluster iteration via `golang.org/x/text/unicode/norm`
- Pattern matching supports both literal strings and regex
- Negative indices wrap from end - implement helper function

### 2.3 Array Functions (`foundations/array.rs`)

Arrays wrap `EcoVec<Value>` (copy-on-write). Key methods:

| Category | Methods |
|----------|---------|
| **Access** | `len`, `first`, `last`, `at`, `push`, `pop`, `insert`, `remove`, `slice` |
| **Search** | `find`, `position`, `contains`, `filter`, `any`, `all` |
| **Transform** | `map`, `enumerate`, `flatten`, `rev`, `sorted`, `dedup` |
| **Combine** | `zip`, `join`, `intersperse`, `chunks`, `windows`, `fold`, `reduce` |
| **Convert** | `to_dict`, `split`, `range` |

**Go Translation Notes:**
- Use `[]Value` slice type
- Implement copy-on-write semantics for mutation methods
- Functional methods (`map`, `filter`, etc.) need closure support
- `sorted` uses stable sort - use `sort.SliceStable`

### 2.4 Dictionary Functions (`foundations/dict.rs`)

Dictionaries use `Arc<IndexMap<Str, Value>>` for order-preserving storage:

| Category | Methods |
|----------|---------|
| **Access** | `len`, `is_empty`, `get`, `at`, `contains` |
| **Modify** | `insert`, `remove`, `take`, `clear` |
| **Iterate** | `keys`, `values`, `pairs`, `filter`, `map` |

**Go Translation Notes:**
- Use `map[string]Value` for basic implementation
- Order preservation requires custom ordered map or slice-backed structure
- Keys are always strings
- `+` operator merges dicts (later values override)

### 2.5 Date/Time Functions (`foundations/datetime.rs`)

Three internal representations: Date, Time, Datetime:

| Category | Functions/Methods |
|----------|-------------------|
| **Factory** | `datetime`, `datetime.today` |
| **Accessors** | `year`, `month`, `day`, `hour`, `minute`, `second`, `weekday`, `ordinal` |
| **Display** | `display(pattern)` |
| **Arithmetic** | `+ duration`, `- duration`, `datetime - datetime` |

**Go Translation Notes:**
- Use `time.Time` for Datetime
- Implement custom Duration type (Typst duration != Go duration)
- Format patterns need translation layer (`[year]-[month]-[day]` -> Go format)
- Limited timezone support (UTC offset only)

### 2.6 Color Functions (`visualize/color.rs`)

8 color spaces with conversion and manipulation:

| Color Space | Components | Use Case |
|-------------|------------|----------|
| Luma | Lightness | Grayscale |
| RGB | R, G, B, Alpha | sRGB display |
| Linear RGB | R, G, B, Alpha | Color operations |
| Oklab | L, a, b, Alpha | Perceptual uniformity |
| Oklch | L, Chroma, Hue, Alpha | Hue manipulation |
| HSL | H, S, L, Alpha | Intuitive adjustment |
| HSV | H, S, V, Alpha | Artist-friendly |
| CMYK | C, M, Y, K | Print production |

**Methods:** `lighten`, `darken`, `saturate`, `desaturate`, `negate`, `rotate`, `mix`, `transparentize`, `opacify`, `components`, `space`, `to_hex`

**Go Translation Notes:**
- Define Color interface with concrete types per space
- Use float32 for component values
- Implement color space conversion matrix operations
- Consider `github.com/lucasb-eyer/go-colorful` for Oklab/Oklch support

### 2.7 Layout Primitives (`layout/`)

Core layout elements:

| Element | Purpose | Key Parameters |
|---------|---------|----------------|
| **page** | Page layout | paper, width, height, margin, columns, fill, header, footer |
| **columns** | Multi-column | count (default: 2), gutter (default: 4%) |
| **stack** | Linear layout | dir (ltr/rtl/ttb/btt), spacing, children |
| **align** | Alignment | 2D alignment (H x V combinations) |
| **pad** | Padding | top, bottom, left, right, x, y, rest |
| **place** | Positioning | alignment, dx, dy, float |
| **grid** | Grid layout | columns, rows, gutter, alignment |

**Go Translation Notes:**
- Define Element interface with layout methods
- Page element contains frame and margin calculations
- Alignment uses Start/End (direction-aware) + Left/Right/Top/Bottom (fixed)
- Units: Abs (points), Em, Ratio (percentage), Fr (fractional)

## 3. Function Registration and Calling

### 3.1 Function Definition Pattern

Rust uses procedural macros to define functions:

```rust
#[func(title = "Square Root")]
pub fn sqrt(
    /// The number to take the square root of.
    value: f64,
) -> SourceResult<f64> {
    if value < 0.0 {
        bail!(value.span(), "cannot take square root of negative number");
    }
    Ok(value.sqrt())
}
```

### 3.2 Function Types

The `Func` type wraps multiple representations:

| Variant | Description |
|---------|-------------|
| Native | Rust implementation with `NativeFuncData` |
| Element | Content element constructors |
| Closure | User-defined functions |
| Plugin | WebAssembly functions |
| With | Partially applied functions |

### 3.3 Calling Convention

Native function signature:
```rust
Fn(&mut Engine, Tracked<Context>, &mut Args) -> SourceResult<Value>
```

**Go Translation Notes:**
- Define `Func` interface with `Call(engine, ctx, args) (Value, error)`
- Use function registry pattern: `map[string]Func`
- Native functions: implement as Go functions with signature `func(*Engine, *Context, *Args) (Value, error)`
- Closures: capture environment in struct with `Call` method

## 4. Type System Integration

### 4.1 Value Type

The core `Value` enum has 29 variants:

```
Absence: None, Auto
Primitives: Bool, Int, Float, Decimal
Dimensions: Length, Angle, Ratio, Relative, Fraction
Visual: Color, Gradient, Tiling, Symbol
Temporal: Version, Datetime, Duration
Text/Data: Str, Bytes, Label
Structured: Content, Styles, Array, Dict
Callable: Func, Args, Type, Module
Dynamic: Dyn (custom types)
```

### 4.2 Type Coercion

Implicit conversions supported:
- Int -> Float (automatic)
- Length/Ratio -> Relative
- Specific types -> union types (e.g., `Num` for int|float)

**Go Translation Notes:**
- Define `Value` as interface with type switch for operations
- Alternative: tagged union struct with type field
- Implement `Cast[T](Value) (T, error)` for type coercion
- Use reflection or code generation for automatic conversion

### 4.3 Type Metadata

Types have associated metadata:
- Short name (code): `str`
- Long name (diagnostic): `string`
- Constructor function
- Scope (methods and constants)

## 5. Pure vs Side-Effect Functions

### 5.1 Pure Functions (No Side Effects)

The vast majority of library functions are pure:

- **All calc functions** - mathematical computations
- **All string methods** - transformations return new strings
- **All array methods** - return new arrays (copy-on-write)
- **All dict methods** - return new dicts
- **Color operations** - return new colors
- **Type constructors** - create new values

### 5.2 Functions with Context Dependencies

Some functions depend on execution context (but don't mutate global state):

| Function | Dependency |
|----------|------------|
| `datetime.today()` | Current time |
| `counter.*` | Document state |
| `state.*` | Document state |
| `query` | Document content |
| `locate` | Layout position |
| `measure` | Layout engine |

### 5.3 Functions with Side Effects

Limited to I/O and document modification:

| Category | Functions | Effect |
|----------|-----------|--------|
| **File Loading** | `read`, `json`, `yaml`, `toml`, `csv`, `xml`, `cbor` | File system read |
| **Image Loading** | `image` | File system read |
| **Document** | Introspection functions | Modify document tree |

**Go Translation Notes:**
- Pure functions: straightforward translation
- Context-dependent: pass context explicitly
- I/O functions: use Go's `io.Reader` interfaces
- Consider functional options pattern for configuration

## 6. Module System

### 6.1 Global Scope Assembly

The library initializes in this order:
1. Math module instantiated
2. Global scope constructed via `global()` function
3. Individual modules register definitions
4. Module-level definitions added (math, pdf, etc.)
5. Prelude values populate global namespace

### 6.2 Prelude Contents

Pre-imported into every document:
- Color constants (CSS colors + color space functions)
- Direction values: `ltr`, `rtl`, `ttb`, `btt`
- Alignment constants: `left`, `center`, `right`, `top`, `bottom`, `horizon`

## 7. Go Implementation Recommendations

### 7.1 Package Structure

```
pkg/stdlib/
├── calc/       # Math functions
├── str/        # String functions
├── array/      # Array functions
├── dict/       # Dictionary functions
├── datetime/   # Date/time functions
├── color/      # Color functions
├── layout/     # Layout primitives
├── model/      # Document model
├── loading/    # File loading
├── symbols/    # Symbol definitions
└── registry.go # Function registration
```

### 7.2 Key Type Definitions

```go
// Core value type
type Value interface {
    Type() Type
    String() string
}

// Function signature
type NativeFunc func(*Engine, *Context, *Args) (Value, error)

// Function wrapper
type Func struct {
    name   string
    native NativeFunc
    params []ParamInfo
    doc    string
}

// Numeric union for calc functions
type Num struct {
    isInt bool
    i     int64
    f     float64
}
```

### 7.3 Implementation Priorities

1. **Phase 1**: Core types (Value, Func, Type)
2. **Phase 2**: Calc module (pure functions, easy wins)
3. **Phase 3**: String/Array/Dict (data structures)
4. **Phase 4**: Color and Layout (visual types)
5. **Phase 5**: Document model and introspection

### 7.4 Testing Strategy

- Port Typst's test cases where possible
- Property-based testing for math functions
- Fuzzing for string/array edge cases
- Golden file tests for layout primitives
