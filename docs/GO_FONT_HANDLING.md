# Go Font Handling Libraries

This document evaluates Go font handling libraries as replacements for Rust's ttf-parser and font ecosystem used by Typst.

## Executive Summary

**Recommendation: Use `github.com/go-text/typesetting` as the primary font handling library.**

This library provides the most comprehensive solution with:
- Pure Go HarfBuzz port for text shaping
- OpenType feature support
- Font fallback handling
- System font discovery
- Production-tested in Fyne, Gio, and Ebitengine

For font subsetting (PDF embedding), use `github.com/cdillond/gdf/subset` which provides both pure Go and HarfBuzz-based subsetting options.

## Library Comparison Matrix

| Feature | golang.org/x/image/font/sfnt | go-text/typesetting | tdewolff/canvas | tdewolff/font |
|---------|------------------------------|---------------------|-----------------|---------------|
| TTF/OTF Parsing | ✅ | ✅ | ✅ | ✅ |
| WOFF/WOFF2 | ❌ | ❌ | ✅ | ✅ |
| Glyph Outlines | ✅ | ✅ | ✅ | ✅ |
| Font Metrics | ✅ | ✅ | ✅ | ✅ |
| OpenType Features (GSUB/GPOS) | ❌ (kern only) | ✅ | Limited | Limited |
| Text Shaping | ❌ | ✅ (HarfBuzz) | Via go-text | Via go-text |
| Font Fallback | ❌ | ✅ | ❌ | ❌ |
| System Font Discovery | ❌ | ✅ | ✅ | ✅ |
| Subsetting | ❌ | ❌ | ❌ | ✅ (basic) |
| License | BSD-3 | BSD-3/Unlicense | MIT | MIT |

## Detailed Library Analysis

### 1. golang.org/x/image/font/sfnt (Official Go)

**Package:** `golang.org/x/image/font/sfnt` (low-level) and `golang.org/x/image/font/opentype` (high-level)

**Strengths:**
- Official Go project, well-maintained
- Clean low-level API for font access
- Good glyph outline extraction via `LoadGlyph`
- Supports TrueType and OpenType (CFF) outlines

**Limitations:**
- No GSUB table support (no ligatures, contextual alternates, etc.)
- Only basic kerning via GPOS
- No text shaping capability
- Decoder only - cannot write/modify fonts
- No colored glyph support (COLR/CBDT)

**Best For:** Basic glyph extraction when full text shaping isn't needed.

```go
import (
    "golang.org/x/image/font/sfnt"
    "golang.org/x/image/math/fixed"
)

// Parse font
f, err := sfnt.Parse(fontBytes)
if err != nil {
    return err
}

// Get metrics
var buf sfnt.Buffer
metrics, err := f.Metrics(&buf, fixed.I(16), font.HintingNone)

// Get glyph outline
glyphIndex, err := f.GlyphIndex(&buf, 'A')
segments, err := f.LoadGlyph(&buf, glyphIndex, fixed.I(16), nil)
for _, seg := range segments {
    switch seg.Op {
    case sfnt.SegmentOpMoveTo:
        // Start new contour
    case sfnt.SegmentOpLineTo:
        // Line to seg.Args[0]
    case sfnt.SegmentOpQuadTo:
        // Quadratic Bézier to seg.Args[1] via seg.Args[0]
    case sfnt.SegmentOpCubeTo:
        // Cubic Bézier to seg.Args[2] via seg.Args[0], seg.Args[1]
    }
}
```

### 2. github.com/go-text/typesetting (Recommended)

**Package:** `github.com/go-text/typesetting`

**Strengths:**
- **Pure Go HarfBuzz port** - full text shaping support
- Complete OpenType feature support (GSUB/GPOS)
- Font fallback with configurable strategies
- System font discovery via `fontscan`
- Script/language-aware shaping
- Production-tested in major Go UI frameworks (Fyne, Gio, Ebitengine)

**Limitations:**
- No font subsetting (use separate library)
- Still v0.x (API may change)

**Best For:** Complete text rendering pipeline with shaping.

```go
import (
    "github.com/go-text/typesetting/font"
    "github.com/go-text/typesetting/fontscan"
    "github.com/go-text/typesetting/shaping"
    "github.com/go-text/typesetting/language"
    "golang.org/x/image/math/fixed"
)

// System font discovery
fontMap := fontscan.NewFontMap(nil)
fontMap.UseSystemFonts("")
fontMap.SetQuery(fontscan.Query{Families: []string{"DejaVu Sans", "sans-serif"}})

// Load specific font
face, err := font.ParseTTF(fontBytes)
if err != nil {
    return err
}

// Shape text
shaper := &shaping.HarfbuzzShaper{}
input := shaping.Input{
    Text:      []rune("Hello, World!"),
    RunStart:  0,
    RunEnd:    13,
    Face:      face,
    Size:      fixed.I(16),
    Script:    language.Latin,
    Direction: di.DirectionLTR,
    // Enable OpenType features
    FontFeatures: []shaping.FontFeature{
        {Tag: mustTag("liga"), Value: 1}, // Enable ligatures
        {Tag: mustTag("kern"), Value: 1}, // Enable kerning
    },
}
output := shaper.Shape(input)

// Process shaped glyphs
for _, glyph := range output.Glyphs {
    // glyph.GlyphID - glyph to render
    // glyph.XAdvance - horizontal advance
    // glyph.XOffset, glyph.YOffset - positioning offset
}

// Font fallback example
faces := []*font.Face{primaryFace, fallbackFace, emojiFace}
segments := shaping.SplitByFontGlyphs(input, faces)
for _, segment := range segments {
    out := shaper.Shape(segment)
    // Render shaped output
}
```

### 3. github.com/tdewolff/canvas

**Package:** `github.com/tdewolff/canvas`

**Strengths:**
- Complete vector graphics library with font support
- Supports TTF, OTF, WOFF, WOFF2, EOT
- Text formatter with rich layout
- Font subsetting for output (PDF, SVG, etc.)
- Gamma correction for better rendering

**Limitations:**
- Text shaping via go-text/typesetting
- Focused on canvas rendering, not standalone font handling

**Best For:** Vector graphics with text rendering.

```go
import "github.com/tdewolff/canvas"

// Load font family
fontFamily := canvas.NewFontFamily("DejaVu")
fontFamily.LoadFontFile("DejaVuSans.ttf", canvas.FontRegular)
fontFamily.LoadFontFile("DejaVuSans-Bold.ttf", canvas.FontBold)

// Create font face at size
face := fontFamily.Face(12.0, canvas.Black, canvas.FontRegular, canvas.FontNormal)

// Draw text
ctx := canvas.NewContext(c)
ctx.DrawText(10, 10, canvas.NewTextLine(face, "Hello, World!", canvas.Left))
```

### 4. github.com/tdewolff/font

**Package:** `github.com/tdewolff/font`

**Strengths:**
- Parses WOFF, WOFF2, EOT and converts to SFNT
- Basic font subsetting via fontcmd tool
- Clean conversion API

**Limitations:**
- Primarily a format converter
- Limited OpenType feature access
- Command-line tool for subsetting, not library API

**Best For:** Converting web fonts to standard formats.

```go
import "github.com/tdewolff/font"

// Convert WOFF2 to TTF
r, err := font.NewSFNTReader(woff2Reader)
// r is now a TTF/OTF reader
```

## Font Subsetting for PDF Embedding

For PDF embedding, fonts must be subsetted to include only used glyphs. Use `github.com/cdillond/gdf/subset`:

```go
import (
    "golang.org/x/image/font/sfnt"
    "github.com/cdillond/gdf/subset"
)

// Parse original font
f, _ := sfnt.Parse(fontBytes)

// Define characters to keep
cutset := map[rune]struct{}{
    'H': {}, 'e': {}, 'l': {}, 'o': {},
}

// Option 1: Pure Go (TrueType with 'glyf' only)
subsetBytes, err := subset.TTFSubset(f, fontBytes, cutset)

// Option 2: HarfBuzz (handles all font types)
subsetBytes, err := subset.HBSubset(fontBytes, cutset)
```

**Limitations of Pure Go Subsetting:**
- Only works with TrueType fonts (glyf tables)
- No variable font support
- No CFF outline support

**HarfBuzz-based Subsetting:**
- Requires `hb-subset` tool or CGo with libharfbuzz
- Handles all font formats
- More robust for production use

## Comparison to Typst's Font Stack

Typst uses:
- **ttf-parser** - Zero-allocation font parsing
- **rustybuzz** - HarfBuzz port for text shaping
- **typst-kit** - Font discovery and management

Equivalent Go stack:
| Typst (Rust) | Go Equivalent |
|--------------|---------------|
| ttf-parser | go-text/typesetting/font |
| rustybuzz | go-text/typesetting/harfbuzz |
| typst-kit fonts | go-text/typesetting/fontscan |
| subsetting (pdf-writer) | github.com/cdillond/gdf/subset |

The go-text/typesetting library is essentially the Go equivalent of Typst's font stack, providing similar capabilities through its HarfBuzz port.

## Integration Architecture

Recommended architecture for gotypst:

```
┌─────────────────────────────────────────────────────────────┐
│                     Text Rendering Pipeline                  │
├─────────────────────────────────────────────────────────────┤
│  1. Font Loading                                             │
│     └─ go-text/typesetting/font.ParseTTF()                  │
│                                                              │
│  2. Font Discovery (optional)                                │
│     └─ go-text/typesetting/fontscan.FontMap                 │
│                                                              │
│  3. Text Segmentation                                        │
│     └─ go-text/typesetting/shaping.Segmenter                │
│        (script, direction, font fallback)                    │
│                                                              │
│  4. Text Shaping                                             │
│     └─ go-text/typesetting/shaping.HarfbuzzShaper           │
│        (GSUB/GPOS, ligatures, kerning)                       │
│                                                              │
│  5. Glyph Rendering                                          │
│     └─ Shaped glyphs → PDF/SVG output                       │
│                                                              │
│  6. Font Subsetting (for PDF)                                │
│     └─ github.com/cdillond/gdf/subset                       │
└─────────────────────────────────────────────────────────────┘
```

## OpenType Feature Support

go-text/typesetting supports all OpenType features via the `FontFeatures` field:

```go
// Common OpenType features
features := []shaping.FontFeature{
    {Tag: mustTag("liga"), Value: 1},  // Standard ligatures (fi, fl)
    {Tag: mustTag("kern"), Value: 1},  // Kerning
    {Tag: mustTag("frac"), Value: 1},  // Fractions (1/2 → ½)
    {Tag: mustTag("smcp"), Value: 1},  // Small capitals
    {Tag: mustTag("onum"), Value: 1},  // Old-style numerals
    {Tag: mustTag("tnum"), Value: 1},  // Tabular numerals
    {Tag: mustTag("ss01"), Value: 1},  // Stylistic set 1
}

func mustTag(s string) ot.Tag {
    tag, _ := ot.NewTag(s)
    return tag
}
```

## Font Fallback Strategies

go-text/typesetting provides two main fallback approaches:

### 1. SplitByFontGlyphs (Simple)
```go
// Split by first font that supports all runes
faces := []*font.Face{latin, cjk, emoji}
segments := shaping.SplitByFontGlyphs(input, faces)
```

### 2. Fontmap (Advanced)
```go
// Custom per-rune font selection
type MyFontmap struct {
    faces map[rune]*font.Face
}

func (m *MyFontmap) ResolveFace(r rune) *font.Face {
    if face, ok := m.faces[r]; ok {
        return face
    }
    return m.defaultFace
}

segments := shaping.SplitByFace(input, &MyFontmap{})
```

## Performance Considerations

1. **Reuse HarfbuzzShaper** - Create once, use for all shaping operations
2. **Cache parsed fonts** - `font.Face` objects are expensive to create
3. **Use fontscan caching** - System font discovery is cached on disk
4. **Buffer reuse** - sfnt.Buffer can be reused across operations

```go
// Good: Reuse shaper
shaper := &shaping.HarfbuzzShaper{}
for _, text := range texts {
    output := shaper.Shape(makeInput(text, face))
}

// Good: Cache font faces
var fontCache = make(map[string]*font.Face)
func getFace(path string) *font.Face {
    if face, ok := fontCache[path]; ok {
        return face
    }
    // Load and cache
}
```

## Recommended Dependencies

```go
// go.mod
require (
    github.com/go-text/typesetting v0.3.2
    github.com/cdillond/gdf v0.1.19
    golang.org/x/image v0.23.0
)
```

## References

- [go-text/typesetting](https://github.com/go-text/typesetting) - Main font/shaping library
- [go-text/typesetting/shaping](https://pkg.go.dev/github.com/go-text/typesetting/shaping) - Shaping API docs
- [go-text/typesetting/fontscan](https://pkg.go.dev/github.com/go-text/typesetting/fontscan) - Font discovery docs
- [golang.org/x/image/font/sfnt](https://pkg.go.dev/golang.org/x/image/font/sfnt) - Low-level font parsing
- [tdewolff/canvas](https://github.com/tdewolff/canvas) - Vector graphics with fonts
- [cdillond/gdf/subset](https://pkg.go.dev/github.com/cdillond/gdf/subset) - Font subsetting
- [OpenType Feature List](https://learn.microsoft.com/en-us/typography/opentype/spec/featurelist) - Feature tags
- [Typst typst-kit](https://docs.rs/typst-kit) - Rust equivalent for comparison
