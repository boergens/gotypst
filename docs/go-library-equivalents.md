# Go Library Equivalents for Typst Dependencies

Research findings for Go alternatives to critical Rust dependencies used by Typst.

## Summary Table

| Rust Crate | Go Equivalent | Maturity | Notes |
|------------|---------------|----------|-------|
| rustybuzz | go-text/typesetting/harfbuzz | High | Direct HarfBuzz port, used by Fyne/Gio/Ebitengine |
| ttf-parser | golang.org/x/image/font/sfnt | High | Official Go library, low-level SFNT parser |
| pdf-writer | gopdf, gofpdf, pdfcpu | High | Multiple options, gopdf is pure Go |
| hypher | speedata/hyphenation | Medium | TeX algorithm port, actively maintained |
| unicode-* | golang.org/x/text + rivo/uniseg | High | Official + UAX#29 segmentation |
| comemo | wcharczuk/go-incr | Medium | Jane Street incremental port |
| kurbo | dominikh/go-curve | Medium | Direct kurbo port |

---

## 1. Text Shaping (rustybuzz replacement)

### Recommended: `github.com/go-text/typesetting/harfbuzz`

**Maturity**: High - Production use in Fyne, Gio, Ebitengine

**Description**: Direct port of HarfBuzz C/C++ library to pure Go. Provides advanced text layout for various scripts and languages with font-aware substitutions and positioning.

**API Compatibility**:
- Input: runes + font
- Output: slice of positioned glyphs with font indices
- Supports font features, BiDi, complex scripts

**Features**:
- Text shaping with OpenType features
- Line wrapping via `shaping.WrapParagraph`
- Font scanning via `fontscan` package
- Letter/word spacing via `AddSpacing`

**Gaps**: None significant - comprehensive HarfBuzz port

**Links**:
- https://github.com/go-text/typesetting
- https://pkg.go.dev/github.com/go-text/typesetting/harfbuzz

---

## 2. Font Parsing (ttf-parser replacement)

### Recommended: `golang.org/x/image/font/sfnt`

**Maturity**: High - Official Go extended library

**Description**: Low-level decoder for TTF and OTF (SFNT) fonts. Does not depend on rasterization packages.

**API Compatibility**:
- Parse single fonts or font collections (TTC/OTC)
- Access glyph outlines, metrics, tables
- No rasterization included (separate concern)

**Features**:
- TTF, OTF, TTC, OTC support
- Font table access
- Glyph metrics and outlines
- Variable font support

**Alternative**: `golang.org/x/image/font/opentype` - Higher-level API with rasterization, implements `font.Face` interface

**Gaps**: None for parsing - full SFNT support

**Links**:
- https://pkg.go.dev/golang.org/x/image/font/sfnt
- https://pkg.go.dev/golang.org/x/image/font/opentype

---

## 3. PDF Generation (pdf-writer replacement)

### Recommended: `github.com/signintech/gopdf`

**Maturity**: High - Pure Go, actively maintained

**Description**: Simple library for generating PDFs. Pure Go with no external dependencies.

**API Compatibility**:
- Page creation and manipulation
- Text rendering with TTF fonts
- Image embedding
- Basic drawing operations

**Features**:
- UTF-8 font support
- Image embedding (JPEG, PNG)
- Links and bookmarks
- Page templates

**Alternatives**:
- **gofpdf**: Classic choice, more features but less maintained
- **pdfcpu**: Full PDF manipulation (read/write/modify), CLI included
- **unipdf**: Commercial, most feature-complete

**Gaps**:
- Low-level PDF primitives may need custom implementation
- No direct equivalent to Rust's streaming write model

**Links**:
- https://github.com/signintech/gopdf
- https://github.com/pdfcpu/pdfcpu
- https://github.com/jung-kurt/gofpdf

---

## 4. Hyphenation (hypher replacement)

### Recommended: `github.com/speedata/hyphenation`

**Maturity**: Medium - Stable, used in production typesetting

**Description**: Port of TeX's hyphenation algorithm (Frank Liang) to Go.

**API Compatibility**:
- Input: word string + pattern file
- Output: array of break point indices
- Configurable left/right minimum lengths

**Features**:
- TeX pattern file support (from CTAN)
- Configurable `Leftmin`/`Rightmin` parameters
- Returns hyphenation points as indices

**Example**:
```go
// "developers" with US English patterns returns [2 5 7 9]
// meaning: de-vel-op-er-s
```

**Gaps**: None - matches TeX hyphenation behavior

**Links**:
- https://github.com/speedata/hyphenation
- https://pkg.go.dev/github.com/speedata/hyphenation

---

## 5. Unicode Handling (unicode-* crates replacement)

### Recommended: `golang.org/x/text` + `github.com/rivo/uniseg`

**Maturity**: High - Official + well-maintained third-party

**Description**:
- `golang.org/x/text`: Official i18n/l10n library with encoding, transformation, locale handling
- `rivo/uniseg`: UAX#29 text segmentation (grapheme clusters, words, sentences)

**API Compatibility**:
- Normalization (NFC, NFD, NFKC, NFKD)
- Text transformations
- Grapheme/word/sentence segmentation
- String width calculation

**Features**:
- `golang.org/x/text/unicode/norm`: Normalization
- `golang.org/x/text/unicode/bidi`: Bidirectional text
- `golang.org/x/text/transform`: Text transformations
- `rivo/uniseg`: UAX#14 line breaking, UAX#29 segmentation, wcwidth

**Alternatives**:
- `blevesearch/segment`: Another UAX#29 implementation
- `clipperhouse/uax29`: Tokenizer based on UAX#29

**Gaps**:
- `x/text` doesn't include text segmentation directly
- Need `uniseg` or similar for grapheme clustering

**Links**:
- https://pkg.go.dev/golang.org/x/text
- https://github.com/rivo/uniseg

---

## 6. Incremental Computation (comemo replacement)

### Recommended: `github.com/wcharczuk/go-incr`

**Maturity**: Medium - Based on Jane Street's incremental

**Description**: Library for building incremental computation graphs. Enables partial recomputation when only a subset of inputs change.

**API Compatibility**:
- Define computation nodes with dependencies
- Automatic dependency tracking
- Parallel stabilization support

**Features**:
- Incremental recomputation
- Parallel stabilization (`ParallelStabilize`)
- Dynamic graph modification
- Observer pattern for outputs

**Conceptual Difference**:
- comemo: Annotation-based (`#[memoize]`, `#[track]`)
- go-incr: Explicit graph construction

**Alternative**: `github.com/kofalt/go-memoize` - Simple memoization without dependency tracking

**Gaps**:
- No annotation/macro system (Go limitation)
- Requires explicit graph construction vs comemo's implicit tracking
- More boilerplate than comemo's declarative approach

**Links**:
- https://github.com/wcharczuk/go-incr
- https://pkg.go.dev/github.com/wcharczuk/go-incr

---

## 7. 2D Geometry (kurbo replacement)

### Recommended: `github.com/dominikh/go-curve`

**Maturity**: Medium - Direct kurbo port, follows upstream

**Description**: Primitives and routines for 2D shapes, curves, and paths. Manual idiomatic Go port of the kurbo Rust crate.

**API Compatibility**:
- Points, vectors, rectangles
- Bezier curves (quadratic, cubic)
- Arcs and arc length computation
- Path operations

**Features**:
- Stroke expansion (novel kurbo algorithm)
- Arc length parameterization
- Affine transformations
- Path simplification

**Alternatives**:
- `llgcode/draw2d`: 2D graphics with multiple backends
- `tidwall/geometry`: Efficient geometry for spatial indexing

**Gaps**:
- Development follows kurbo; may lag upstream features
- Less battle-tested than kurbo

**Links**:
- https://github.com/dominikh/go-curve

---

## Architecture Recommendations

### High Confidence (Direct Equivalents)
1. **Text shaping**: go-text/typesetting - comprehensive, production-ready
2. **Font parsing**: x/image/font/sfnt - official, complete
3. **Hyphenation**: speedata/hyphenation - exact algorithm match

### Medium Confidence (Good Alternatives)
4. **Unicode**: x/text + uniseg - combined coverage matches Rust crates
5. **2D geometry**: go-curve - direct port, actively maintained

### Requires Adaptation
6. **PDF generation**: Multiple options exist but architectural differences from pdf-writer's streaming model
7. **Incremental computation**: go-incr provides capability but API is fundamentally different from comemo's annotation-based approach

### Key Architectural Differences

**comemo vs go-incr**: This is the biggest gap. comemo uses Rust's macro system for transparent memoization with fine-grained dependency tracking. Go lacks macros, so go-incr requires explicit graph construction. A gotypst implementation would need to:
- Design computation graph explicitly
- Wrap tracked values in observer types
- May need custom tracking for document-specific caching

**PDF generation**: Rust's pdf-writer is a low-level streaming writer. Go libraries are higher-level. May need:
- Custom low-level PDF writer
- Or adapter layer over gopdf/pdfcpu
