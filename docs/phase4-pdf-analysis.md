# Phase 4: typst-pdf Crate Analysis

## Executive Summary

The `typst-pdf` crate is a relatively thin adapter layer that transforms Typst's `PagedDocument` into PDF bytes. The heavy lifting is delegated to **krilla**, a high-level PDF generation library. This architecture means our Go translation will need to either:
1. Port krilla to Go (significant effort), or
2. Use existing Go PDF libraries as the backend

## 1. Architecture Overview

### 1.1 Source Structure

```
crates/typst-pdf/
├── src/
│   ├── lib.rs        # Entry point: pdf() function
│   ├── convert.rs    # Core conversion logic
│   ├── page.rs       # Page labeling
│   ├── text.rs       # Text/font rendering
│   ├── paint.rs      # Colors, gradients, patterns
│   ├── shape.rs      # Geometric shapes
│   ├── image.rs      # Image embedding (raster, SVG, PDF)
│   ├── link.rs       # Hyperlinks and annotations
│   ├── outline.rs    # Document bookmarks
│   ├── metadata.rs   # PDF metadata
│   ├── attach.rs     # File attachments
│   ├── util.rs       # Utilities
│   └── tags/         # PDF/UA accessibility tagging
│       ├── context/
│       ├── tree/
│       ├── util/
│       ├── groups.rs
│       ├── mod.rs
│       └── resolve.rs
└── Cargo.toml
```

### 1.2 Pipeline Flow

```
PagedDocument + PdfOptions
        │
        ▼
    pdf() [lib.rs]
        │
        ▼
    convert() [convert.rs]
        │
        ├──► GlobalContext (fonts, images, destinations)
        ├──► FrameContext (per-frame state, transforms)
        │
        ▼
    convert_pages()
        │
        ├──► handle_frame() [recursive]
        │       │
        │       ├──► handle_text()   → surface.draw_glyphs()
        │       ├──► handle_shape()  → surface.draw_path()
        │       ├──► handle_image()  → surface.draw_image/svg/pdf
        │       ├──► handle_link()   → collect annotations
        │       └──► handle_group()  → recursive frames
        │
        ▼
    finish() → Vec<u8> (PDF bytes)
```

## 2. Key Data Structures

### 2.1 Public API

```rust
// Main entry point
pub fn pdf(document: &PagedDocument, options: &PdfOptions) -> SourceResult<Vec<u8>>

// Configuration
pub struct PdfOptions {
    pub ident: Smart<Option<String>>,    // Document identifier
    pub timestamp: Option<Datetime>,      // Creation timestamp
    pub page_ranges: Option<PageRanges>,  // Selective export
    pub standards: PdfStandards,          // PDF/A, PDF/UA compliance
    pub tagged: bool,                     // Accessibility tagging
}

// Standards support
pub enum PdfStandard {
    V1_4, V1_5, V1_6, V1_7, V2_0,        // PDF versions
    A_1b, A_2a, A_2b, A_2u, A_3a, A_3b, A_3u,  // PDF/A
    UA_1,                                 // PDF/UA accessibility
}
```

### 2.2 Internal State

```rust
// Document-wide state
struct GlobalContext {
    fonts_forward: HashMap<Font, krilla::Font>,      // Typst → Krilla font mapping
    fonts_backward: HashMap<krilla::Font, Font>,     // Reverse mapping
    image_spans: HashMap<krilla::Image, Span>,       // Image → source location
    named_dests: HashMap<Location, String>,          // Named destinations
    loc_to_page_index: LocationToPageIndex,          // Location mapping
    tags: Option<TagManager>,                        // Accessibility state
}

// Per-frame state
struct FrameContext {
    state: State,                    // Transform stack
    link_annotations: Vec<LinkAnnotation>,
}

struct State {
    transforms: Vec<Transform>,      // Transform stack
    container_transform: Transform,  // Current container
    size: Size,                      // Container dimensions
}
```

## 3. Dependencies

### 3.1 Typst Internal
| Crate | Purpose |
|-------|---------|
| typst-library | Document model, elements |
| typst-syntax | Source locations, spans |
| typst-assets | Asset management |
| typst-timing | Performance timing |
| typst-utils | Shared utilities |
| typst-macros | Procedural macros |

### 3.2 External
| Crate | Purpose | Go Equivalent Needed |
|-------|---------|---------------------|
| **krilla** | PDF generation engine | Core challenge |
| **krilla-svg** | SVG rendering to PDF | SVG handling |
| pdf-writer | Low-level PDF primitives | (via krilla) |
| image | Image processing | Go image stdlib |
| serde | Serialization | encoding/json |
| ecow | Efficient strings/vectors | Go strings/slices |
| comemo | Memoization | Manual caching |
| indexmap | Ordered hash maps | orderedmap pkg |

### 3.3 Krilla Architecture

Krilla is the key dependency - a high-level PDF creation library that handles:

- **Graphics**: Paths, fills, strokes, blend modes, clipping
- **Transforms**: Affine transformations
- **Gradients**: Linear, radial, conic (sweep)
- **Images**: Raster (JPEG, PNG), SVG embedding
- **Fonts**: OpenType support, subsetting, color fonts
- **Text**: Glyph drawing (layout done externally)
- **Accessibility**: PDF/UA-1 tagged PDF
- **Standards**: PDF 1.4-2.0, PDF/A variants

Krilla explicitly does NOT handle:
- Text layout
- Pagination
- High-level document structure

## 4. Go PDF Library Recommendations

### 4.1 Primary Recommendation: **pdfcpu**

**Pros:**
- Pure Go, Apache 2.0 licensed
- Active development, PDF 2.0 support
- Comprehensive manipulation features
- Strong encryption support
- Good performance

**Cons:**
- More focused on processing/editing than generation
- May need augmentation for advanced rendering

### 4.2 Alternative: **gopdf (signintech/gopdf)**

**Pros:**
- Simple, focused PDF generation API
- Unicode/TTF font support
- Basic graphics primitives
- Active maintenance

**Cons:**
- Less feature-rich than pdfcpu
- No PDF/A or PDF/UA support out of box

### 4.3 Comparison Matrix

| Feature | pdfcpu | gopdf | unipdf |
|---------|--------|-------|--------|
| License | Apache 2.0 | MIT | Commercial |
| Pure Go | Yes | Yes | Yes |
| Font subsetting | Yes | Partial | Yes |
| PDF/A support | In progress | No | Yes |
| PDF/UA support | Limited | No | Yes |
| Active | Yes | Yes | Yes |
| Gradients | Limited | Basic | Yes |
| SVG | No | No | No |

### 4.4 Recommended Strategy

**Hybrid approach:**
1. Use **pdfcpu** as the foundation for PDF structure
2. Build custom rendering layer for Typst-specific needs
3. Consider low-level PDF writing for complex features

This mirrors how typst-pdf uses krilla - we need a similar abstraction layer in Go.

## 5. Translation Challenges

### 5.1 High Complexity

| Challenge | Typst Implementation | Go Approach |
|-----------|---------------------|-------------|
| **Font subsetting** | krilla handles via pdf-writer | Need Go font library (sfnt, etc.) |
| **Color fonts** | krilla OpenType support | Limited Go support |
| **PDF/UA tagging** | Complex tree structure | Manual implementation |
| **SVG embedding** | krilla-svg crate | Consider rasterization |
| **Gradient interpolation** | Perceptual color spaces | Manual Oklab implementation |

### 5.2 Medium Complexity

| Challenge | Notes |
|-----------|-------|
| **Transform stacks** | Straightforward matrix math |
| **Path rendering** | Map to Go graphics primitives |
| **Image embedding** | Go image stdlib adequate |
| **Link annotations** | PDF spec implementation |
| **Outline/bookmarks** | Tree structure, manageable |

### 5.3 Low Complexity

| Challenge | Notes |
|-----------|-------|
| **Metadata** | Simple key-value pairs |
| **Page labels** | Direct translation |
| **File attachments** | Binary embedding |

### 5.4 Architectural Considerations

1. **No direct krilla equivalent in Go**: We'll need to build PDF primitives ourselves or heavily extend an existing library.

2. **Memoization pattern** (`#[comemo::memoize]`): Replace with explicit caching in Go.

3. **Error handling**: Rust's `Result<T, E>` → Go's `(T, error)` pattern.

4. **Generics/traits**: Rust traits → Go interfaces.

5. **Memory management**: Rust's ownership → Go's GC (simpler but watch for allocations).

## 6. Implementation Phases

### Phase 4a: Foundation
- Set up Go PDF library integration
- Implement basic document structure
- Simple text rendering

### Phase 4b: Graphics
- Path rendering
- Solid colors
- Basic images (JPEG, PNG)

### Phase 4c: Advanced Rendering
- Gradients (linear, radial, conic)
- Patterns
- Font subsetting

### Phase 4d: Features
- Links and annotations
- Outlines/bookmarks
- Metadata

### Phase 4e: Compliance
- PDF/A support
- PDF/UA accessibility
- Standards validation

## 7. References

- [typst-pdf source](https://github.com/typst/typst/tree/main/crates/typst-pdf)
- [krilla repository](https://github.com/LaurenzV/krilla)
- [pdfcpu](https://github.com/pdfcpu/pdfcpu)
- [gopdf](https://github.com/signintech/gopdf)
- [PDF 2.0 specification (ISO 32000-2)](https://www.iso.org/standard/75839.html)
