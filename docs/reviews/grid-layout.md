# Grid/Table Layout Translation Analysis

**Reviewer:** obsidian
**Date:** 2026-01-19
**Status:** Translation Roadmap

## Overview

This review analyzes the Typst `typst-layout/src/grid/` module for Go translation. The grid layout system handles both `#grid()` and `#table()` elements through a unified `GridLayouter` that supports multi-region pagination, rowspans, colspans, and repeating headers/footers.

## Source Module Structure

```
typst-layout/src/grid/
├── mod.rs        # Entry points: layout_grid, layout_table, layout_cell
├── layouter.rs   # GridLayouter core algorithm
├── lines.rs      # Grid line rendering (strokes, borders)
├── repeated.rs   # Repeating headers/footers
└── rowspans.rs   # Multi-row cell span handling
```

## Key Data Structures

### GridLayouter

The central struct managing grid layout state:

```rust
struct GridLayouter<'a> {
    // Core data
    grid: &'a Grid,              // Cell grid structure
    regions: Regions<'a>,        // Available layout regions
    rcols: Vec<Abs>,             // Resolved column widths
    width: Abs,                  // Total grid width

    // Row/region state
    rrows: Vec<RowState>,        // Resolved rows by region
    current: Current,            // Active region state
    finished: Vec<Frame>,        // Completed frames
    unbreakable_rows_left: usize,

    // Headers and rowspans
    repeating_headers: Vec<...>,
    pending_headers: Vec<...>,
    rowspans: Vec<Rowspan>,
    finished_header_rows: Vec<FinishedHeaderRowInfo>,

    // Supporting
    cell_locators: HashMap<...>,
    styles: StyleChain<'a>,
    is_rtl: bool,
    row_state: RowState,
}
```

**Go Translation:**
```go
type GridLayouter struct {
    Grid     *Grid
    Regions  *Regions
    RCols    []Abs
    Width    Abs

    RRows              []RowState
    Current            Current
    Finished           []Frame
    UnbreakableRowsLeft int

    RepeatingHeaders    []Header
    PendingHeaders      []Header
    Rowspans            []Rowspan
    FinishedHeaderRows  []FinishedHeaderRowInfo

    CellLocators map[Axes[int]]Locator
    Styles       StyleChain
    IsRTL        bool
    RowState     RowState
}
```

### Rowspan

Tracks cells spanning multiple rows:

```rust
struct Rowspan {
    x: usize,                        // First column
    y: usize,                        // First row
    disambiguator: usize,            // Cell identifier
    rowspan: usize,                  // Rows spanned
    is_effectively_unbreakable: bool,
    dx: Abs,                         // Horizontal offset
    dy: Abs,                         // Vertical offset in first region
    first_region: usize,
    region_full: Abs,                // Full height in first region
    heights: Vec<Abs>,               // Per-region available space
    max_resolved_row: Option<usize>,
    is_being_repeated: bool,
}
```

**Go Translation:**
```go
type Rowspan struct {
    X, Y            int
    Disambiguator   int
    RowspanCount    int
    IsUnbreakable   bool
    DX, DY          Abs
    FirstRegion     int
    RegionFull      Abs
    Heights         []Abs
    MaxResolvedRow  *int  // nil = None
    IsBeingRepeated bool
}
```

### LineSegment

Represents drawable grid lines:

```rust
struct LineSegment {
    stroke: Arc<Stroke<Abs>>,
    offset: Abs,
    length: Abs,
    priority: StrokePriority,
}

enum StrokePriority {
    GridStroke = 0,   // Global grid styling
    CellStroke = 1,   // Per-cell overrides
    ExplicitLine = 2, // User-placed hline/vline
}
```

**Go Translation:**
```go
type LineSegment struct {
    Stroke   *Stroke
    Offset   Abs
    Length   Abs
    Priority StrokePriority
}

type StrokePriority int

const (
    GridStroke StrokePriority = iota
    CellStroke
    ExplicitLine
)
```

## Core Algorithms

### 1. Column Measurement (`measure_columns`)

Three-phase resolution:

**Phase 1: Fixed Columns**
- Relative-sized columns resolved against region width
- Direct percentage/length conversion

**Phase 2: Auto Columns**
- Layout all cells in auto columns
- Capture maximum width per column
- Expensive: requires full cell layout

**Phase 3: Fractional Distribution**
- Remaining space allocated to `fr` columns proportionally
- If insufficient space: fair-share shrinking of auto columns

**Fair-share Shrinking Algorithm:**
```
1. Calculate fair share = remaining / overlarge_count
2. Find columns below fair share threshold
3. Shrink them to their measured size
4. Redistribute remaining to overlarge columns
5. Repeat until stable
```

### 2. Row Layout Strategy

Three sizing types with distinct behaviors:

| Type | Behavior | Can Break? |
|------|----------|------------|
| Auto | Measured per-region | Yes |
| Relative | Fixed height | May force break |
| Fractional | Deferred to region end | No |

**`layout_row_internal` dispatch:**
```
1. Check region capacity
2. Verify rowspan prerequisites
3. Route to: layout_auto_row, layout_relative_row, or defer
```

### 3. Region Finalization (`finish_region`)

Sequential processing:
1. Orphan prevention (remove lonely headers)
2. Strip trailing gutter rows
3. Layout footer if conditions permit
4. Size fractional rows from remaining space
5. Complete pending rowspans
6. Prepare headers for next region

### 4. Repeating Headers Lifecycle

```
pending_headers  ─┬─►  repeating_headers
                 │     (if row placed after)
                 └─►  discarded
                      (if orphan)
```

**Orphan Prevention:** If a header is the only content in a region, it's removed and the region is skipped.

### 5. Line Generation Algorithm

`generate_line_segments()` state machine:
```
for each track orthogonal to line direction:
    if stroke matches current segment:
        extend segment length
    else:
        yield current segment
        start new segment

    interrupt on:
    - merged cells (colspan/rowspan blocking)
    - stroke change
    - priority change
```

**Stroke Priority Resolution:**
1. Explicit user line (highest)
2. Adjacent cells' strokes
3. Global grid stroke (lowest)

## Go Translation Challenges

### Challenge 1: Lifetime Management

**Rust:** `GridLayouter<'a>` borrows grid/styles
**Go:** Cannot express borrowed references

**Solution:** Use pointers with clear ownership documentation:
```go
type GridLayouter struct {
    Grid   *Grid   // Not owned, must outlive layouter
    Styles *Styles // Shared reference
}
```

### Challenge 2: Generic Size Types

**Rust:** `Axes<usize>`, `Sides<Abs>`
**Go:** Limited generics support

**Solution:** Define concrete types:
```go
type AxesInt struct{ X, Y int }
type AxesAbs struct{ X, Y Abs }
type Sides struct{ Left, Top, Right, Bottom Abs }
```

### Challenge 3: Iterator State Machines

**Rust:** `generate_line_segments` returns impl Iterator

**Go Options:**
1. Callback pattern: `func(LineSegment) bool`
2. Slice collection: `func() []LineSegment`
3. Channel (overhead): `func() <-chan LineSegment`

**Recommendation:** Use slice collection for simplicity.

### Challenge 4: Arc<Stroke>

**Rust:** Shared ownership via `Arc<Stroke<Abs>>`
**Go:** No reference counting

**Solution:** Pointer with nil check:
```go
type LineSegment struct {
    Stroke *Stroke  // May be shared, do not mutate
}
```

### Challenge 5: Option<usize> Fields

**Rust:** `max_resolved_row: Option<usize>`
**Go:** No sum types

**Solutions:**
```go
// Option 1: Pointer
MaxResolvedRow *int

// Option 2: Sentinel value
MaxResolvedRow int  // -1 = None

// Option 3: Explicit flag
MaxResolvedRow    int
HasMaxResolvedRow bool
```

**Recommendation:** Use pointer for clarity.

## Dependencies

### Internal (from typst-library)

| Type | Purpose | Go Location |
|------|---------|-------------|
| `Grid` | Cell structure | `library/layout/grid.go` |
| `Cell` | Individual cell | `library/layout/cell.go` |
| `Frame` | Output container | `kit/frame.go` |
| `Regions` | Layout regions | `kit/regions.go` |
| `StyleChain` | Inherited styles | `library/styles.go` |
| `Abs`, `Fr` | Length units | `kit/length.go` |

### External

| Rust | Purpose | Go Equivalent |
|------|---------|---------------|
| `kurbo` | Geometry | `image/geom` or custom |
| `ecow` | CoW strings | Go strings (immutable) |
| `smallvec` | Small vecs | Go slices |

## Test Strategy

### Unit Tests

1. **Column measurement:**
   - Fixed/relative columns
   - Auto columns with varying content
   - Fractional distribution
   - Fair-share shrinking edge cases

2. **Row layout:**
   - Auto row height calculation
   - Relative row forcing breaks
   - Fractional row sizing

3. **Rowspans:**
   - Single-region spans
   - Multi-region spans with pagination
   - Gutter removal at breaks

4. **Headers:**
   - Orphan prevention
   - Multi-level header conflicts
   - Footer positioning

5. **Lines:**
   - Stroke priority resolution
   - Segment generation across merged cells
   - RTL coordinate handling

### Integration Tests

1. Golden tests comparing layout output with Typst reference
2. Complex tables with headers, rowspans, and pagination
3. RTL grid rendering

## Recommended Translation Order

### Phase 1: Foundation (Prerequisite)
- [ ] `Abs`, `Fr`, `Em` types in `kit/length.go`
- [ ] `Frame`, `Fragment` types in `kit/frame.go`
- [ ] `Regions` type in `kit/regions.go`
- [ ] `Grid`, `Cell` types in `library/layout/`

### Phase 2: Core Layouter
- [ ] `GridLayouter` struct
- [ ] `measure_columns` algorithm
- [ ] `layout_row` routing logic
- [ ] `finish_region` basic flow

### Phase 3: Advanced Features
- [ ] `Rowspan` tracking and completion
- [ ] Repeating headers/footers
- [ ] Orphan prevention

### Phase 4: Line Rendering
- [ ] `LineSegment` type
- [ ] `generate_line_segments`
- [ ] Stroke priority resolution

### Phase 5: Entry Points
- [ ] `layout_grid`
- [ ] `layout_table`
- [ ] `layout_cell`

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Rowspan complexity | High | Port tests first, implement incrementally |
| Multi-region pagination | High | Comprehensive test coverage |
| Fair-share algorithm edge cases | Medium | Unit test edge cases |
| RTL handling (missing in lines.rs) | Medium | Add explicit RTL tests |
| Performance regression | Medium | Benchmark against Rust |

## Files for Reference

**Rust Source (typst/typst):**
- `crates/typst-layout/src/grid/mod.rs`
- `crates/typst-layout/src/grid/layouter.rs`
- `crates/typst-layout/src/grid/lines.rs`
- `crates/typst-layout/src/grid/repeated.rs`
- `crates/typst-layout/src/grid/rowspans.rs`

**Go Targets:**
- `layout/grid/layouter.go`
- `layout/grid/lines.go`
- `layout/grid/repeated.go`
- `layout/grid/rowspans.go`

## Conclusion

The grid layout module is moderately complex, with the primary challenges being:

1. **Rowspan multi-region handling** - tracking heights across page breaks with gutter removal
2. **Repeating headers** - orphan prevention and level conflict resolution
3. **Column measurement** - fair-share shrinking algorithm

The module is self-contained with clear interfaces, making it suitable for incremental translation. The recommended approach is to port the column measurement and basic row layout first, then add rowspan and header features, and finally line rendering.

Estimated scope: ~2500 lines of Rust → ~2000 lines of Go (accounting for Go's verbosity and removed lifetime annotations).
