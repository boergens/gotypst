# Phase 3: typst-realize Crate Analysis

This document analyzes the `typst-realize` crate to understand its role in the Typst rendering pipeline.

## Overview

The `typst-realize` crate implements the **realization subsystem** for Typst. Realization is the process of recursively applying styling and show rules to transform content into well-known elements suitable for further processing (layout and rendering).

**Key insight**: Realize sits between parsing/evaluation and layout. It doesn't produce visual output directly; instead, it transforms the content tree by applying rules and grouping related elements.

## Pipeline Position

```
Source → Parse → Evaluate → REALIZE → Layout → Render
                              ↑
                         (this crate)
```

1. **Input**: Content tree from evaluation with associated style chains
2. **Process**: Apply show rules, group elements, collapse spaces
3. **Output**: Transformed content ready for layout

## Crate Structure

```
typst-realize/
├── Cargo.toml
└── src/
    ├── lib.rs      # Core realization logic
    └── spaces.rs   # Space collapsing algorithm
```

### Dependencies

Internal Typst crates:
- `typst-library` - Core types and content definitions
- `typst-macros` - Procedural macros
- `typst-syntax` - Syntax tree types
- `typst-timing` - Performance instrumentation
- `typst-utils` - Utility functions

External:
- `arrayvec` - Fixed-capacity vectors
- `bumpalo` - Arena allocator for efficient memory management
- `comemo` - Memoization framework
- `ecow` - Efficient copy-on-write strings
- `regex` - Pattern matching for text show rules

## Core Function: `realize()`

```rust
pub fn realize<'a>(
    kind: RealizationKind,
    engine: &mut Engine,
    locator: &mut SplitLocator,
    arenas: &'a Arenas,
    content: &'a Content,
    styles: StyleChain<'a>,
) -> SourceResult<Vec<Pair<'a>>>
```

### Parameters

| Parameter | Type | Purpose |
|-----------|------|---------|
| `kind` | `RealizationKind` | Specifies the realization context |
| `engine` | `&mut Engine` | Layout/evaluation engine state |
| `locator` | `&mut SplitLocator` | Position tracking |
| `arenas` | `&'a Arenas` | Pre-allocated memory arenas |
| `content` | `&'a Content` | Input content tree |
| `styles` | `StyleChain<'a>` | Cascading style information |

### RealizationKind Variants

| Kind | Purpose |
|------|---------|
| `LayoutDocument` | Full document layout preparation |
| `LayoutFragment` | Fragment layout (block/inline detection) |
| `LayoutPar` | Paragraph-specific realization |
| `HtmlDocument` | HTML export preparation |
| `HtmlFragment` | HTML fragment export |
| `Math` | Mathematical content realization |

## Key Data Structures

### State

Maintains mutable context during realization:
- Engine reference
- Locator for position tracking
- Output sink for realized content
- Active groupings (paragraphs, lists, citations)
- Configuration flags

### GroupingRule

Defines how related elements are collected for unified processing:

```rust
struct GroupingRule {
    trigger: /* when to start grouping */,
    inner: /* what elements belong inside */,
    interrupt: /* what elements break the group */,
    finalize: /* how to finalize the group */,
}
```

Grouping rules handle:
- **Paragraphs** - Collecting inline content into paragraph elements
- **Lists** - Grouping list items
- **Citations** - Collecting citations for unified bibliography handling

### SpaceState (spaces.rs)

Categorizes elements for space collapsing:

| State | Description |
|-------|-------------|
| `Invisible` | Elements that don't affect space collapsing (tags) |
| `Destructive` | Elements that discard adjacent spaces |
| `Supportive` | Normal elements requiring spaces on both sides |
| `Space` | Space elements that can collapse with adjacent spaces |

## Show Rules Processing

Show rules are transformations applied to content elements. The system supports:

1. **User-defined recipes** - Custom transformations written in Typst
2. **Built-in rules** - Native element transformations

### Processing Flow

1. **Verdict determination** - Inspect element and styles to decide action
2. **Preparation** - One-time setup (location assignment, synthesis)
3. **Rule execution** - Apply matching show rules
4. **Grouping** - Collect related elements after transformation

### Regex Text Matching

Special handling for regex-based show rules in textual content:
- Collect and merge text representations
- Account for space collapsing
- Find pattern matches across element boundaries

## Space Collapsing Algorithm

The `collapse_spaces()` function in `spaces.rs` handles:

1. Removing unnecessary spaces at content boundaries
2. Collapsing adjacent spaces into single spaces
3. Operating in-place for efficiency (no allocations)

Uses cursor-based position tracking to shift elements leftward as spaces are discarded.

## Integration Points

### With typst-layout

Layout calls realize through the engine routines:

```rust
let children = (engine.routines.realize)(
    RealizationKind::LayoutFragment { kind: &mut kind },
    &mut engine,
    &mut locator,
    &arenas,
    content,
    styles,
)?;
```

The realized children are then passed to layout functions (`layout_flow`, etc.).

### With Renderers

Realize does **not** directly interface with renderers. The data flow is:

```
Content → Realize → Layout → Frame → Render
```

Realize produces transformed content that layout converts to `Frame` structures. Renderers (`typst-render`, `typst-pdf`, `typst-svg`) then process frames.

## Frame and FrameItem (Layout Output)

For context, here's what layout produces (consumed by renderers):

### Frame
A finished layout with items at fixed positions:
- `size` - Dimensions
- `baseline` - Optional vertical position
- `items` - Collection of positioned FrameItems
- `kind` - Soft (follows parent) or Hard (uses own size)

### FrameItem
Building blocks of frames:

| Variant | Description |
|---------|-------------|
| `Group` | Subframes with optional transform/clipping |
| `Text` | Shaped text runs |
| `Shape` | Geometric shapes with fill/stroke |
| `Image` | Images with dimensions |
| `Link` | Navigation destinations |
| `Tag` | Introspectable elements |

## Renderer Overview

### typst-render (Raster Images)

```rust
pub fn render(page: &Page, pixel_per_pt: f32) -> Pixmap
```

Walks frame hierarchy, dispatching to specialized handlers for text, shapes, images.

### typst-pdf

```rust
pub fn pdf(document: &PagedDocument, options: &PdfOptions) -> Vec<u8>
```

Supports PDF standards (1.4-2.0, PDF/A, PDF/UA), tagged PDF for accessibility.

### typst-svg

```rust
pub fn svg(page: &Page) -> String
pub fn svg_frame(frame: &Frame) -> String
pub fn svg_merged(document: &PagedDocument, gap: Abs) -> String
```

Deduplicates glyphs, clip paths, gradients for optimized output.

## Key Observations for Go Implementation

1. **Arena allocation** - Realize uses `bumpalo` for efficient memory management during recursive traversal. Go's GC handles this differently; consider pooling.

2. **Memoization** - Uses `comemo` for caching. Go doesn't have a standard equivalent; consider manual caching or libraries.

3. **Multiple realization kinds** - The enum-based dispatch pattern maps to Go interfaces.

4. **Regex handling** - Rust's `regex` crate has Go equivalent in `regexp` package.

5. **Space collapsing** - The in-place algorithm can be directly translated to Go slices.

6. **Style chains** - The cascading style system needs careful translation to maintain efficiency.

## Summary

The `typst-realize` crate transforms evaluated content into a form suitable for layout by:

1. Applying show rules (user-defined and built-in)
2. Grouping related elements (paragraphs, lists, citations)
3. Collapsing spaces according to typesetting rules
4. Supporting multiple output contexts (document, fragment, HTML, math)

It's a transformation layer, not a rendering layer. The actual visual output is produced by the layout → render pipeline that consumes realized content.
