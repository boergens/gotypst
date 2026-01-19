# Phase 3: typst-layout Crate Analysis

This document provides a comprehensive analysis of the `typst-layout` crate to prepare for Go translation in Phase 3.

## Overview

The `typst-layout` crate is Typst's layout engine, responsible for converting abstract document content into positioned frames ready for rendering. It handles:

- Page layout and pagination
- Block-level flow layout
- Inline text layout with line breaking
- Mathematical equation layout
- Grid and table layout
- Geometric transformations

## Module Structure

```
typst-layout/src/
├── lib.rs          # Public API: layout_document, layout_fragment, layout_frame
├── flow/           # Block-level flow layout
│   ├── mod.rs      # Entry points and FlowMode enum
│   ├── collect.rs  # Content collection and preprocessing
│   ├── compose.rs  # Frame composition with floats/footnotes
│   └── distribute.rs # Content distribution across regions
├── inline/         # Inline/paragraph layout
│   ├── mod.rs      # Entry point: layout_par, layout_inline
│   ├── collect.rs  # Text collection and item preparation
│   ├── prepare.rs  # BiDi analysis and text shaping
│   ├── linebreak.rs # Line breaking algorithms
│   ├── shaping.rs  # Text shaping via rustybuzz
│   └── finalize.rs # Line finalization to frames
├── math/           # Mathematical expression layout
│   ├── mod.rs      # Entry: layout_equation_inline, layout_equation_block
│   ├── fragment.rs # MathFragment types (Glyph, Frame, Space, etc.)
│   ├── fraction.rs # Fraction layout
│   ├── accent.rs   # Accent positioning
│   └── [others]    # Scripts, radicals, matrices, etc.
├── grid/           # Grid and table layout
│   ├── mod.rs      # Entry: layout_grid, layout_table
│   ├── layouter.rs # GridLayouter with multi-region support
│   ├── lines.rs    # Grid line rendering
│   ├── repeated.rs # Repeated elements (headers/footers)
│   └── rowspans.rs # Rowspan management
├── pages/          # Document-level page layout
│   ├── mod.rs      # Entry: layout_document
│   └── collect.rs  # Page collection and parallelization
├── stack.rs        # Stack layout (horizontal/vertical)
├── lists.rs        # List layout (bulleted, numbered)
├── shapes.rs       # Shape rendering (line, rect, circle, curve)
├── transforms.rs   # Geometric transforms (rotate, scale, skew)
├── modifiers.rs    # Frame modifiers (links, visibility)
├── image.rs        # Image layout
├── pad.rs          # Padding
├── repeat.rs       # Repeat patterns
└── rules.rs        # Rule elements
```

## Layout Pipeline

### 1. Document Layout (`pages/`)

**Entry Point:** `layout_document(engine, content, styles) -> PagedDocument`

**Process:**
1. Realize root content through engine routines
2. Collect content into page items (runs, parity blocks, tags)
3. **Parallelize page layout** - page runs are laid out concurrently
4. Build introspection data for positioned elements
5. Return final `PagedDocument`

**Key Types:**
- `Item` enum: `Run` (content), `Tags` (between pages), `Parity` (odd/even positioning)
- `ManualPageCounter`: Manages logical vs physical page numbers
- `SplitLocator`: Tracks element positions for introspection

### 2. Flow Layout (`flow/`)

**Entry Points:**
- `layout_frame(engine, content, locator, styles, region)` - Single region
- `layout_fragment(engine, content, locator, styles, regions)` - Multiple regions
- `layout_columns(...)` - Column-specific with parent-scoped floats

**Process:**
1. **Collect** children into preprocessed structures
2. **Compose** frames by distributing content into regions
3. Handle **floats** and **footnotes** as out-of-flow elements
4. Apply **sticky blocks** and spacing logic

**Key Types:**
```rust
enum FlowMode { Root, Block, Inline }

struct Work {
    children: Vec<Child>,      // Unprocessed content
    spillover: Vec<Child>,     // From breakable blocks
    floats: VecDeque<Float>,   // Queued floats
    footnotes: Vec<Footnote>,  // Pending footnotes
    skips: HashSet<Location>,  // Processed locations
}

struct Config {
    mode: FlowMode,
    columns: ColumnConfig,
    footnotes: FootnoteConfig,
    line_numbers: LineNumberConfig,
}

enum Child {
    LineChild { frame, leading, spacing, ... },
    SingleChild { frame, fr, sticky, ... },
    MultiChild { layouter, sticky, ... },
    PlacedChild { placed, ... },
    Spacing { ... },
}
```

**Float Handling:**
- Floats checked for fit in available space
- Non-fitting floats queued for subsequent regions
- Top/bottom positioning based on page location
- Relayout triggered when floats placed

**Footnote Handling:**
- Markers discovered within frames
- "Footnote invariant": marker and entry on same page
- Migration logic if entry doesn't fit

### 3. Inline Layout (`inline/`)

**Entry Point:** `layout_par(par, engine, locator, styles, region, expand, situation) -> Fragment`

**Pipeline:**
1. **Collect**: Gather text into single string for BiDi analysis
2. **Prepare**: BiDi analysis, text shaping, build item list
3. **Linebreak**: Segment into lines using selected algorithm
4. **Finalize**: Convert lines to frames

**Key Types:**
```rust
enum Item {
    Text(ShapedText),
    Space(Abs),           // Absolute spacing
    FrSpace(Fr),          // Fractional spacing
    Frame(Frame),         // Inline boxes
    Tag(Tag),
    Skip,                 // Invisible Unicode
}

struct Preparation {
    text: String,
    config: Config,
    bidi: Option<BidiLevels>,  // Only if direction varies
    items: Vec<(Range, Item)>,
    indices: Vec<usize>,       // Byte-to-item mapping
    spans: SpanMapper,
}

struct Config {
    justify: bool,
    linebreaks: Linebreaks,
    first_line_indent: Abs,
    hanging_indent: Abs,
    alignment: Alignment,
    hyphenation: bool,
    costs: LinebreakCosts,
    cjk_spacing: bool,
}
```

### 4. Line Breaking (`inline/linebreak.rs`)

**Two Algorithms:**

#### Simple First-Fit (`linebreak_simple`)
- Greedy approach taking longest possible line
- Fast but suboptimal
- Used for `Linebreaks::Simple` mode

#### Knuth-Plass (`linebreak_optimized`)
- Dynamic programming for optimal line breaks
- Three-phase process:
  1. **Approximate pass**: Compute rough costs as upper bound
  2. **Bounded optimization**: Exact DP with pruning
  3. **Path retracing**: Reconstruct optimal breaks

**Cost Function:**
```
cost = (1 + badness + penalty)²

badness = 100 * |ratio|³  (for justified/shrinking lines)

Penalties:
- Hyphenation base: 135.0
- Short word hyphen: +15% per char < 5 from word edge
- Consecutive hyphens: +135.0
- Runts (isolated words): 100.0
```

### 5. Text Shaping (`inline/shaping.rs`)

**Process:**
1. Segment text by BiDi level and script
2. Call rustybuzz for each segment
3. Apply tracking, spacing, adjustability

**Key Types:**
```rust
struct ShapedText {
    text: String,
    glyphs: Glyphs,
    dir: Dir,
    lang: Lang,
    region: Option<Region>,
    // ... styling
}

struct ShapedGlyph {
    font: Font,
    glyph_id: u16,
    x_advance: Em,
    x_offset: Em,
    adjustability: Adjustability,  // Stretch/shrink
    range: Range<usize>,           // Character cluster
    script: Script,
}

struct Glyphs {
    kept: EcoVec<ShapedGlyph>,
    trimmed: EcoVec<ShapedGlyph>,  // End-of-line whitespace
}
```

**Font Fallback:**
- Primary font selection from families
- Fallback via `book.select_fallback()` if coverage fails
- Tofu glyphs for unrenderable characters

**CJK Features:**
- Punctuation classification (GB/CNS/JIS standards)
- Consecutive punctuation adjustment
- CJK-Latin spacing (0.25em advance)

### 6. Math Layout (`math/`)

**Entry Points:**
- `layout_equation_inline(...)` - Inline equations
- `layout_equation_block(...)` - Display equations with numbering

**Core Architecture:**
```rust
struct MathContext {
    engine: Engine,
    locator: Locator,
    font_stack: Vec<Font>,     // Handle font changes
    fragments: Vec<MathFragment>,
}

enum MathFragment {
    Glyph(GlyphFragment),
    Frame(FrameFragment),
    Space(Abs),
    Linebreak,
    Align,
    Tag(Tag),
}

struct GlyphFragment {
    font: Font,
    size: Abs,
    glyphs: Vec<Glyph>,
    class: MathClass,        // Op, Bin, Rel, etc.
    italics_correction: Abs,
    // ...
}
```

**Layout Dispatch:**
- Spacing and alignment: inline handling
- Glyphs: font metrics and stretching
- Fractions: numerator/denominator positioning
- Accents: attachment point calculation
- Scripts: superscript/subscript positioning
- Radicals: root symbol stretching

**Glyph Stretching:**
1. Try pre-made variants (ordered by size)
2. Assemble from parts if variants insufficient
3. Cap repetition at 1024 repeats

### 7. Grid Layout (`grid/`)

**Entry Points:**
- `layout_grid(grid, engine, locator, styles, regions) -> Fragment`
- `layout_table(table, ...)` - Same underlying layouter

**Key Types:**
```rust
struct GridLayouter {
    grid: Grid,
    regions: Regions,
    rcols: Vec<Abs>,           // Resolved column widths
    rrows: Vec<RowState>,      // Row state
    current: Current,          // Region-specific state
    lrows: Vec<Row>,           // Laid out rows
    // ...
}

struct Current {
    initial_header_height: Abs,
    repeating_header_height: Abs,
    pending_header_height: Abs,
    footer_height: Abs,
}
```

**Algorithm Phases:**

1. **Column Measurement** (`measure_columns`):
   - Resolve fixed and relative widths
   - Measure auto columns by laying out cells
   - Distribute remaining space to fractional columns
   - Fair-share shrinking when space insufficient

2. **Row Layout** (`layout_row`):
   - Check for region breaks
   - Detect rowspans and unbreakable groups
   - Route to: `layout_auto_row`, `layout_relative_row`, or defer fractional

3. **Region Finalization** (`finish_region`):
   - Enforce orphan prevention
   - Remove trailing gutter rows
   - Layout fractional rows
   - Process completed rowspans
   - Prepare next region headers/footers

**Advanced Features:**
- Multi-page rowspans with height tracking
- Repeating headers with orphan prevention
- RTL support (coordinate mirroring)
- Gutter handling between content tracks

## Critical Dependencies

### Internal Typst Crates
- **typst-library**: Element types, styling, content model
- **typst-syntax**: Spans, source locations
- **typst-utils**: Utility functions, data structures
- **typst-macros**: Procedural macros
- **typst-timing**: Performance instrumentation
- **typst-assets**: Embedded resources

### External Libraries

| Library | Purpose | Go Translation Challenge |
|---------|---------|--------------------------|
| **rustybuzz** | Text shaping (HarfBuzz wrapper) | Must use Go binding or pure Go port |
| **ttf-parser** | Font parsing | Need Go font parser |
| **hypher** | Hyphenation | Need hyphenation dictionary support |
| **unicode-bidi** | BiDi algorithm | Use `golang.org/x/text/unicode/bidi` |
| **unicode-segmentation** | Text segmentation | Use `golang.org/x/text/unicode/norm` |
| **icu_segmenter** | Line/word breaking | Use ICU bindings or pure Go |
| **icu_properties** | Unicode properties | Use `golang.org/x/text` |
| **bumpalo** | Arena allocator | Standard Go allocation or pool |
| **ecow** | Copy-on-write strings | Go strings are immutable, different strategy |
| **smallvec** | Small vector optimization | Go slices, potentially with pool |
| **comemo** | Memoization | Custom cache implementation |
| **kurbo** | 2D geometry | Pure Go or use existing geometry lib |

## Go Translation Challenges

### 1. Text Shaping (Critical)

**Challenge:** rustybuzz/HarfBuzz is essential for complex text rendering.

**Options:**
1. **CGo binding to HarfBuzz** - Best quality, adds CGo dependency
2. **Pure Go port** - No dependencies, massive effort
3. **Simpler shaping** - ASCII/basic scripts only, limited support

**Recommendation:** Use CGo binding initially. `github.com/nicholasi/harfbuzz` or custom binding. Plan for optional pure-Go fallback for simple cases.

### 2. Memoization (comemo)

**Challenge:** Typst heavily uses memoization for incremental compilation.

**Go Approach:**
```go
type MemoCache struct {
    mu    sync.RWMutex
    cache map[uint64]interface{}
}

func Memoize[T any](cache *MemoCache, key uint64, compute func() T) T {
    // Check cache, compute if missing
}
```

Consider using input hashing similar to Rust's approach.

### 3. Copy-on-Write (ecow)

**Challenge:** Rust's ecow provides efficient CoW semantics.

**Go Approach:** Go strings are already immutable. For slices, use explicit copying or consider persistent data structures from `github.com/emirpasic/gods`.

### 4. Arena Allocation (bumpalo)

**Challenge:** Rust uses arena allocation for layout performance.

**Go Approach:**
- Use `sync.Pool` for frame/fragment reuse
- Consider arena allocator for hot paths
- Profile to determine if needed

### 5. Parallelization

**Challenge:** Typst parallelizes page layout via rayon.

**Go Approach:**
```go
func layoutPages(items []Item) []Frame {
    results := make([]Frame, len(items))
    var wg sync.WaitGroup
    for i, item := range items {
        wg.Add(1)
        go func(i int, item Item) {
            defer wg.Done()
            results[i] = layoutItem(item)
        }(i, item)
    }
    wg.Wait()
    return results
}
```

### 6. Enum Pattern Translation

**Rust:**
```rust
enum Child {
    LineChild { frame: Frame, leading: Abs },
    SingleChild { frame: Frame, fr: Fr },
    // ...
}
```

**Go:**
```go
type Child interface {
    isChild()
}

type LineChild struct {
    Frame   Frame
    Leading Abs
}
func (LineChild) isChild() {}

type SingleChild struct {
    Frame Frame
    Fr    Fr
}
func (SingleChild) isChild() {}
```

### 7. Error Handling

**Rust:** Uses `SourceResult<T>` with span information.

**Go:**
```go
type SourceError struct {
    Span    Span
    Message string
}

func (e *SourceError) Error() string {
    return fmt.Sprintf("%s: %s", e.Span, e.Message)
}
```

### 8. Generic Size/Point Types

**Rust:**
```rust
struct GenericSize<T> { main: T, cross: T }
```

**Go:**
```go
type Size struct {
    Width  Abs
    Height Abs
}

type GenericSize[T any] struct {
    Main  T
    Cross T
}
```

## Recommended Translation Order

### Phase 3a: Foundation
1. Core types: `Frame`, `Fragment`, `Abs`, `Fr`, `Point`, `Size`
2. Geometry utilities from kurbo
3. Basic transforms

### Phase 3b: Simple Layout
1. `stack.rs` - Stack layout (straightforward)
2. `shapes.rs` - Shape rendering
3. `pad.rs` - Padding
4. `transforms.rs` - Geometric transforms

### Phase 3c: Flow Layout (Partial)
1. `flow/collect.rs` - Content collection
2. `flow/distribute.rs` - Distribution without floats
3. `flow/compose.rs` - Basic composition

### Phase 3d: Inline Layout (Critical Path)
1. `inline/collect.rs` - Text collection
2. `inline/shaping.rs` - **Requires HarfBuzz integration**
3. `inline/prepare.rs` - BiDi analysis
4. `inline/linebreak.rs` - Both algorithms
5. `inline/finalize.rs` - Line finalization

### Phase 3e: Grid Layout
1. `grid/layouter.rs` - Core grid logic
2. `grid/lines.rs` - Line rendering
3. `grid/rowspans.rs` - Rowspan support

### Phase 3f: Math Layout (Complex)
1. `math/fragment.rs` - Math types
2. `math/mod.rs` - Context and dispatch
3. Individual math components

### Phase 3g: Document Layout
1. `pages/collect.rs` - Page collection
2. `pages/mod.rs` - Document assembly
3. Full float/footnote support

## Key Data Structures Summary

| Rust Type | Go Translation | Notes |
|-----------|---------------|-------|
| `Abs` | `type Abs float64` | Absolute length in points |
| `Fr` | `type Fr float64` | Fractional unit |
| `Em` | `type Em float64` | Font-relative unit |
| `Frame` | `struct Frame { Size, Items }` | Container for positioned items |
| `Fragment` | `[]Frame` | Multi-frame result |
| `Point` | `struct Point { X, Y Abs }` | 2D coordinate |
| `Size` | `struct Size { Width, Height Abs }` | 2D dimensions |
| `Axes<T>` | `struct Axes[T] { X, Y T }` | Generic 2D pair |
| `Sides<T>` | `struct Sides[T] { Left, Top, Right, Bottom T }` | Padding/margins |
| `Corners<T>` | `struct Corners[T] { TopLeft, ... T }` | Corner radii |
| `Transform` | `struct Transform { a,b,c,d,e,f float64 }` | 2D affine matrix |
| `Ratio` | `type Ratio float64` | Normalized 0-1 value |
| `Alignment` | `struct Alignment { X, Y Align }` | 2D alignment |

## Testing Strategy

1. **Unit tests** for each module mirroring Rust tests
2. **Golden tests** comparing layout output with Typst reference
3. **Fuzzing** for line breaking and text shaping
4. **Benchmark suite** comparing with Rust performance

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Text shaping complexity | High | Start with CGo HarfBuzz, document limitations |
| Math layout complexity | High | Implement incrementally, test extensively |
| Performance regression | Medium | Profile early, use pools/arenas as needed |
| Unicode edge cases | Medium | Use well-tested golang.org/x/text packages |
| Float layout bugs | Medium | Port tests carefully, add fuzzing |

## Conclusion

The `typst-layout` crate represents the most complex component for Go translation due to:

1. Deep integration with text shaping (HarfBuzz)
2. Sophisticated algorithms (Knuth-Plass line breaking)
3. Complex state management (floats, footnotes, rowspans)
4. Performance-critical code paths

Recommended approach:
- Start with simpler modules (stack, shapes, transforms)
- Tackle text shaping early to validate approach
- Build comprehensive test suite alongside implementation
- Consider CGo dependencies where pure Go is impractical
