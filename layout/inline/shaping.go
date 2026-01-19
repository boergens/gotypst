// Package inline provides inline/paragraph layout including text shaping.
package inline

import (
	"fmt"
	"sync"
	"unicode"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/unicode/bidi"
)

// Constants for special characters.
const (
	SHY       = '\u00ad' // Soft hyphen
	SHYSTR    = "\u00ad"
	Hyphen    = '-'
	HyphenStr = "-"
)

// Dir represents text direction.
type Dir int

const (
	DirLTR Dir = iota // Left to right
	DirRTL            // Right to left
)

// IsPositive returns true for LTR direction.
func (d Dir) IsPositive() bool {
	return d == DirLTR
}

// Lang represents a language tag.
type Lang string

const (
	LangChinese  Lang = "zh"
	LangJapanese Lang = "ja"
	LangKorean   Lang = "ko"
	LangEnglish  Lang = "en"
)

// Region represents a region/territory code.
type Region string

// AsStr returns the region as a string.
func (r Region) AsStr() string {
	return string(r)
}

// Abs represents an absolute length in points (1/72 inch).
type Abs float64

// Zero returns an Abs of zero.
func AbsZero() Abs {
	return 0
}

// Em represents a font-relative unit.
type Em float64

// Zero returns an Em of zero.
func EmZero() Em {
	return 0
}

// One returns 1em.
func EmOne() Em {
	return 1.0
}

// At converts Em to Abs at a given font size.
func (e Em) At(size Abs) Abs {
	return Abs(float64(e) * float64(size))
}

// FromAbs creates Em from Abs at a given font size.
func EmFromAbs(abs Abs, size Abs) Em {
	if size == 0 {
		return 0
	}
	return Em(float64(abs) / float64(size))
}

// Range represents a byte range in text.
type Range struct {
	Start int
	End   int
}

// Contains returns true if the range contains the given index.
func (r Range) Contains(i int) bool {
	return i >= r.Start && i < r.End
}

// Adjustability represents how much a glyph can stretch or shrink.
type Adjustability struct {
	// Stretchability: (left, right) amounts the glyph can stretch.
	Stretchability [2]Em
	// Shrinkability: (left, right) amounts the glyph can shrink.
	Shrinkability [2]Em
}

// ShapedGlyph represents a single shaped glyph with positioning info.
type ShapedGlyph struct {
	Font          *font.Face
	GlyphID       uint16
	XAdvance      Em
	XOffset       Em
	YOffset       Em
	Size          Abs
	Adjustability Adjustability
	Range         Range
	SafeToBreak   bool
	Char          rune
	IsJustifiable bool
	Script        language.Script
}

// IsSpace returns true if the glyph is a whitespace character.
func (g *ShapedGlyph) IsSpace() bool {
	return isSpace(g.Char)
}

// IsCJScript returns true if the glyph is CJK.
func (g *ShapedGlyph) IsCJScript() bool {
	return isCJScript(g.Char, g.Script)
}

// IsCJKPunctuation returns true if the glyph is CJK punctuation.
func (g *ShapedGlyph) IsCJKPunctuation() bool {
	return g.IsCJKLeftAlignedPunctuation(CJKPunctStyleGB) ||
		g.IsCJKRightAlignedPunctuation() ||
		g.IsCJKCenterAlignedPunctuation(CJKPunctStyleGB)
}

// IsCJKLeftAlignedPunctuation checks for CJK left-aligned punctuation.
func (g *ShapedGlyph) IsCJKLeftAlignedPunctuation(style CJKPunctStyle) bool {
	return isCJKLeftAlignedPunctuation(g.Char, g.XAdvance, g.Stretchability(), style)
}

// IsCJKRightAlignedPunctuation checks for CJK right-aligned punctuation.
func (g *ShapedGlyph) IsCJKRightAlignedPunctuation() bool {
	return isCJKRightAlignedPunctuation(g.Char, g.XAdvance, g.Stretchability())
}

// IsCJKCenterAlignedPunctuation checks for CJK center-aligned punctuation.
func (g *ShapedGlyph) IsCJKCenterAlignedPunctuation(style CJKPunctStyle) bool {
	return isCJKCenterAlignedPunctuation(g.Char, style)
}

// IsLetterOrNumber returns true if the glyph is a Latin/Greek/Cyrillic letter or number.
func (g *ShapedGlyph) IsLetterOrNumber() bool {
	switch g.Script {
	case language.Latin, language.Greek, language.Cyrillic:
		return true
	}
	switch g.Char {
	case '#', '$', '%', '&':
		return true
	}
	return g.Char >= '0' && g.Char <= '9'
}

// Stretchability returns the glyph's stretch amounts.
func (g *ShapedGlyph) Stretchability() [2]Em {
	return g.Adjustability.Stretchability
}

// Shrinkability returns the glyph's shrink amounts.
func (g *ShapedGlyph) Shrinkability() [2]Em {
	return g.Adjustability.Shrinkability
}

// ShrinkLeft applies left shrink to the glyph.
func (g *ShapedGlyph) ShrinkLeft(amount Em) {
	g.XOffset -= amount
	g.XAdvance -= amount
	g.Adjustability.Shrinkability[0] -= amount
}

// ShrinkRight applies right shrink to the glyph.
func (g *ShapedGlyph) ShrinkRight(amount Em) {
	g.XAdvance -= amount
	g.Adjustability.Shrinkability[1] -= amount
}

// Glyphs represents a collection of shaped glyphs with trimming support.
type Glyphs struct {
	inner []ShapedGlyph
	kept  Range // Range of kept (non-trimmed) glyphs
}

// NewGlyphsFromSlice creates Glyphs from a slice.
func NewGlyphsFromSlice(glyphs []ShapedGlyph) *Glyphs {
	return &Glyphs{
		inner: glyphs,
		kept:  Range{Start: 0, End: len(glyphs)},
	}
}

// NewGlyphsFromVec creates Glyphs from a vector (ownership transfer).
func NewGlyphsFromVec(glyphs []ShapedGlyph) *Glyphs {
	return &Glyphs{
		inner: glyphs,
		kept:  Range{Start: 0, End: len(glyphs)},
	}
}

// Len returns the number of kept glyphs.
func (g *Glyphs) Len() int {
	return g.kept.End - g.kept.Start
}

// All returns all glyphs including trimmed ones.
func (g *Glyphs) All() []ShapedGlyph {
	return g.inner
}

// Kept returns the kept glyphs.
func (g *Glyphs) Kept() []ShapedGlyph {
	return g.inner[g.kept.Start:g.kept.End]
}

// IsFullyEmpty returns true if there are no glyphs at all.
func (g *Glyphs) IsFullyEmpty() bool {
	return len(g.inner) == 0
}

// Trim removes glyphs matching the predicate from start and end.
func (g *Glyphs) Trim(pred func(*ShapedGlyph) bool) {
	start := g.kept.Start
	end := g.kept.End
	for start < end && pred(&g.inner[start]) {
		start++
	}
	for end > start && pred(&g.inner[end-1]) {
		end--
	}
	g.kept = Range{Start: start, End: end}
}

// At returns the glyph at the given index within kept range.
func (g *Glyphs) At(i int) *ShapedGlyph {
	if i < 0 || i >= g.Len() {
		return nil
	}
	return &g.inner[g.kept.Start+i]
}

// Last returns the last kept glyph.
func (g *Glyphs) Last() *ShapedGlyph {
	if g.Len() == 0 {
		return nil
	}
	return &g.inner[g.kept.End-1]
}

// KeptContains returns true if the index is within the kept range.
func (g *Glyphs) KeptContains(i int) bool {
	return g.kept.Contains(i)
}

// ShapedText represents shaped text with metadata.
type ShapedText struct {
	Base    int         // Base byte offset in original text
	Text    string      // The text that was shaped
	Dir     Dir         // Text direction
	Lang    Lang        // Language
	Region  *Region     // Optional region
	Variant FontVariant // Font variant used
	Glyphs  *Glyphs     // Shaped glyphs
}

// FontVariant describes a font variant (style, weight, stretch).
type FontVariant struct {
	Style   FontStyle
	Weight  FontWeight
	Stretch FontStretch
}

// FontStyle represents italic/normal style.
type FontStyle int

const (
	FontStyleNormal FontStyle = iota
	FontStyleItalic
	FontStyleOblique
)

// FontWeight represents font weight (100-900).
type FontWeight int

const (
	FontWeightThin       FontWeight = 100
	FontWeightExtraLight FontWeight = 200
	FontWeightLight      FontWeight = 300
	FontWeightNormal     FontWeight = 400
	FontWeightMedium     FontWeight = 500
	FontWeightSemiBold   FontWeight = 600
	FontWeightBold       FontWeight = 700
	FontWeightExtraBold  FontWeight = 800
	FontWeightBlack      FontWeight = 900
)

// FontStretch represents font stretch.
type FontStretch int

const (
	FontStretchNormal FontStretch = iota
	FontStretchCondensed
	FontStretchExpanded
)

// Width returns the total width of the shaped text.
func (s *ShapedText) Width() Abs {
	var total Abs
	for _, g := range s.Glyphs.Kept() {
		total += g.XAdvance.At(g.Size)
	}
	return total
}

// Justifiables returns the count of justifiable glyphs.
func (s *ShapedText) Justifiables() int {
	count := 0
	for _, g := range s.Glyphs.Kept() {
		if g.IsJustifiable {
			count++
		}
	}
	return count
}

// CJKJustifiableAtLast returns true if the last glyph is CJK justifiable.
func (s *ShapedText) CJKJustifiableAtLast() bool {
	last := s.Glyphs.Last()
	if last == nil {
		return false
	}
	return last.IsCJScript() || last.IsCJKPunctuation()
}

// Stretchability returns total stretchability.
func (s *ShapedText) Stretchability() Abs {
	var total Abs
	for _, g := range s.Glyphs.Kept() {
		stretch := g.Stretchability()
		total += (stretch[0] + stretch[1]).At(g.Size)
	}
	return total
}

// Shrinkability returns total shrinkability.
func (s *ShapedText) Shrinkability() Abs {
	var total Abs
	for _, g := range s.Glyphs.Kept() {
		shrink := g.Shrinkability()
		total += (shrink[0] + shrink[1]).At(g.Size)
	}
	return total
}

// Empty returns an empty ShapedText with the same metadata.
func (s *ShapedText) Empty() *ShapedText {
	return &ShapedText{
		Base:    s.Base,
		Text:    "",
		Dir:     s.Dir,
		Lang:    s.Lang,
		Region:  s.Region,
		Variant: s.Variant,
		Glyphs:  NewGlyphsFromSlice(nil),
	}
}

// CJKPunctStyle represents CJK punctuation style variants.
type CJKPunctStyle int

const (
	CJKPunctStyleGB  CJKPunctStyle = iota // GB (Simplified Chinese)
	CJKPunctStyleCNS                      // CNS (Traditional Chinese)
	CJKPunctStyleJIS                      // JIS (Japanese)
)

// GetCJKPunctStyle returns the CJK punctuation style for a language/region.
func GetCJKPunctStyle(lang Lang, region *Region) CJKPunctStyle {
	switch lang {
	case LangChinese:
		if region != nil {
			r := region.AsStr()
			if r == "TW" || r == "HK" {
				return CJKPunctStyleCNS
			}
		}
		return CJKPunctStyleGB
	case LangJapanese:
		return CJKPunctStyleJIS
	default:
		return CJKPunctStyleGB
	}
}

// Character classification functions.

func isSpace(c rune) bool {
	return c == ' ' || c == '\u00A0' || c == '\u3000'
}

// IsOfCJScript returns true if the character is CJK.
func IsOfCJScript(c rune) bool {
	return isCJScript(c, getScript(c))
}

func isCJScript(c rune, script language.Script) bool {
	switch script {
	case language.Hiragana, language.Katakana, language.Han:
		return true
	}
	return c == '\u30FC' // Katakana-Hiragana prolonged sound mark
}

func isCJKLeftAlignedPunctuation(c rune, xAdvance Em, stretchability [2]Em, style CJKPunctStyle) bool {
	// U+201D (right double quote) and U+2019 (right single quote) for closing marks
	if (c == '\u201d' || c == '\u2019') && xAdvance+stretchability[1] == EmOne() {
		return true
	}

	if (style == CJKPunctStyleGB || style == CJKPunctStyleJIS) &&
		(c == '，' || c == '。' || c == '．' || c == '、' || c == '：' || c == '；') {
		return true
	}

	if style == CJKPunctStyleGB && (c == '？' || c == '！') {
		return true
	}

	switch c {
	case '》', '）', '』', '」', '】', '〗', '〕', '〉', '］', '｝':
		return true
	}
	return false
}

func isCJKRightAlignedPunctuation(c rune, xAdvance Em, stretchability [2]Em) bool {
	// U+201C (left double quote) and U+2018 (left single quote) for opening marks
	if (c == '\u201c' || c == '\u2018') && xAdvance+stretchability[0] == EmOne() {
		return true
	}
	switch c {
	case '《', '（', '『', '「', '【', '〖', '〔', '〈', '［', '｛':
		return true
	}
	return false
}

func isCJKCenterAlignedPunctuation(c rune, style CJKPunctStyle) bool {
	if style == CJKPunctStyleCNS &&
		(c == '，' || c == '。' || c == '．' || c == '、' || c == '：' || c == '；') {
		return true
	}
	return c == '\u30FB' || c == '\u00B7' // Katakana middle dot, middle dot
}

func isJustifiable(c rune, script language.Script, xAdvance Em, stretchability [2]Em) bool {
	style := CJKPunctStyleGB
	return isSpace(c) ||
		isCJScript(c, script) ||
		isCJKLeftAlignedPunctuation(c, xAdvance, stretchability, style) ||
		isCJKRightAlignedPunctuation(c, xAdvance, stretchability) ||
		isCJKCenterAlignedPunctuation(c, style)
}

// isDefaultIgnorable returns true for Unicode default ignorable characters.
func isDefaultIgnorable(c rune) bool {
	// Default ignorable code points per Unicode
	switch {
	case c == 0x00AD: // Soft hyphen
		return true
	case c >= 0x034F && c <= 0x034F: // Combining grapheme joiner
		return true
	case c >= 0x115F && c <= 0x1160: // Hangul fillers
		return true
	case c >= 0x17B4 && c <= 0x17B5: // Khmer vowel inherent
		return true
	case c >= 0x180B && c <= 0x180E: // Mongolian free variation selectors
		return true
	case c >= 0x200B && c <= 0x200F: // Zero-width characters
		return true
	case c >= 0x202A && c <= 0x202E: // Bidi control characters
		return true
	case c >= 0x2060 && c <= 0x206F: // Word joiner, invisible operators
		return true
	case c >= 0xFE00 && c <= 0xFE0F: // Variation selectors
		return true
	case c == 0xFEFF: // BOM/ZWNBSP
		return true
	case c >= 0xFFF0 && c <= 0xFFF8: // Specials
		return true
	case c >= 0x1D173 && c <= 0x1D17A: // Musical formatting
		return true
	case c >= 0xE0000 && c <= 0xE0FFF: // Tags and variation selectors supplement
		return true
	}
	return false
}

// getScript returns the Unicode script for a character.
func getScript(c rune) language.Script {
	// Simplified script detection - in production, use a proper Unicode library
	if unicode.In(c, unicode.Han) {
		return language.Han
	}
	if unicode.In(c, unicode.Hiragana) {
		return language.Hiragana
	}
	if unicode.In(c, unicode.Katakana) {
		return language.Katakana
	}
	if unicode.In(c, unicode.Latin) {
		return language.Latin
	}
	if unicode.In(c, unicode.Greek) {
		return language.Greek
	}
	if unicode.In(c, unicode.Cyrillic) {
		return language.Cyrillic
	}
	if unicode.In(c, unicode.Arabic) {
		return language.Arabic
	}
	if unicode.In(c, unicode.Hebrew) {
		return language.Hebrew
	}
	return language.Common
}

// CJK punctuation patterns for line breaking.
var (
	// Opening punctuation marks
	BeginPunctPat = []rune{
		'\u201c', '\u2018', // " and ' (curly quotes)
		'《', '〈', '（', '『', '「', '【', '〖', '〔', '［', '｛',
	}
	// Closing punctuation marks
	EndPunctPat = []rune{
		'\u201d', '\u2019', // " and ' (curly quotes)
		'，', '．', '。', '、', '：', '；', '》', '〉', '）', '』', '」', '】',
		'〗', '〕', '］', '｝', '？', '！',
	}
)

// ShapingContext holds context for a shaping operation.
type ShapingContext struct {
	Shaper   *shaping.HarfbuzzShaper
	Faces    []*font.Face  // Font faces for fallback
	Size     Abs           // Font size
	Variant  FontVariant   // Font variant
	Features []shaping.FontFeature
	Fallback bool // Enable font fallback
	Dir      Dir
	glyphs   []ShapedGlyph
	used     []*font.Face
	mu       sync.Mutex
}

// NewShapingContext creates a new shaping context.
func NewShapingContext(faces []*font.Face, size Abs) *ShapingContext {
	return &ShapingContext{
		Shaper:   &shaping.HarfbuzzShaper{},
		Faces:    faces,
		Size:     size,
		Fallback: true,
		glyphs:   make([]ShapedGlyph, 0, 128),
	}
}

// Shape shapes text and returns ShapedText.
func Shape(ctx *ShapingContext, base int, text string, dir Dir, lang Lang, region *Region) *ShapedText {
	if len(text) == 0 {
		return &ShapedText{
			Base:    base,
			Text:    text,
			Dir:     dir,
			Lang:    lang,
			Region:  region,
			Variant: ctx.Variant,
			Glyphs:  NewGlyphsFromSlice(nil),
		}
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.glyphs = ctx.glyphs[:0]
	ctx.used = ctx.used[:0]
	ctx.Dir = dir

	shapeSegment(ctx, base, text)

	trackAndSpace(ctx)
	calculateAdjustability(ctx, lang, region)

	glyphs := make([]ShapedGlyph, len(ctx.glyphs))
	copy(glyphs, ctx.glyphs)

	return &ShapedText{
		Base:    base,
		Text:    text,
		Dir:     dir,
		Lang:    lang,
		Region:  region,
		Variant: ctx.Variant,
		Glyphs:  NewGlyphsFromVec(glyphs),
	}
}

// shapeSegment shapes a text segment using available fonts.
func shapeSegment(ctx *ShapingContext, base int, text string) {
	// Skip if text only contains newlines, tabs, or ignorable characters
	hasContent := false
	for _, c := range text {
		if c != '\n' && c != '\t' && !isDefaultIgnorable(c) {
			hasContent = true
			break
		}
	}
	if !hasContent {
		return
	}

	// Find a suitable font
	var face *font.Face
	for _, f := range ctx.Faces {
		if f != nil && !containsFace(ctx.used, f) {
			face = f
			break
		}
	}

	if face == nil {
		// No font available, create tofu glyphs
		if len(ctx.Faces) > 0 && ctx.Faces[0] != nil {
			shapeTofus(ctx, base, text, ctx.Faces[0])
		}
		return
	}

	ctx.used = append(ctx.used, face)

	// Prepare shaping input
	runes := []rune(text)
	direction := di.DirectionLTR
	if ctx.Dir == DirRTL {
		direction = di.DirectionRTL
	}

	input := shaping.Input{
		Text:         runes,
		RunStart:     0,
		RunEnd:       len(runes),
		Face:         face,
		Size:         toFixed(float64(ctx.Size)),
		Direction:    direction,
		FontFeatures: ctx.Features,
	}

	// Shape the text
	output := ctx.Shaper.Shape(input)

	// Convert shaped glyphs to our format
	byteOffset := 0
	runeIdx := 0

	for i, glyph := range output.Glyphs {
		// Calculate byte range for this cluster
		cluster := glyph.ClusterIndex

		// Find the byte position for this rune
		for runeIdx < cluster && runeIdx < len(runes) {
			byteOffset += len(string(runes[runeIdx]))
			runeIdx++
		}

		start := base + byteOffset

		// Find end of cluster
		endRune := cluster + 1
		if i+1 < len(output.Glyphs) {
			endRune = output.Glyphs[i+1].ClusterIndex
		} else {
			endRune = len(runes)
		}

		endByte := byteOffset
		for r := cluster; r < endRune && r < len(runes); r++ {
			endByte += len(string(runes[r]))
		}
		end := base + endByte

		// Get the character for this cluster
		var c rune
		if cluster < len(runes) {
			c = runes[cluster]
		}

		script := getScript(c)
		xAdvance := Em(float64(glyph.XAdvance) / float64(ctx.Size))

		ctx.glyphs = append(ctx.glyphs, ShapedGlyph{
			Font:          face,
			GlyphID:       uint16(glyph.GlyphID),
			XAdvance:      xAdvance,
			XOffset:       Em(float64(glyph.XOffset) / float64(ctx.Size)),
			YOffset:       Em(float64(glyph.YOffset) / float64(ctx.Size)),
			Size:          ctx.Size,
			Adjustability: Adjustability{},
			Range:         Range{Start: start, End: end},
			SafeToBreak:   true, // Simplified; HarfBuzz provides this info
			Char:          c,
			IsJustifiable: isJustifiable(c, script, xAdvance, [2]Em{0, 0}),
			Script:        script,
		})
	}

	ctx.used = ctx.used[:len(ctx.used)-1]
}

// shapeTofus creates placeholder glyphs for missing characters.
func shapeTofus(ctx *ShapingContext, base int, text string, face *font.Face) {
	xAdvance := Em(0.5) // Default tofu width

	addGlyph := func(cluster int, c rune) {
		start := base + cluster
		end := start + len(string(c))
		script := getScript(c)

		ctx.glyphs = append(ctx.glyphs, ShapedGlyph{
			Font:          face,
			GlyphID:       0,
			XAdvance:      xAdvance,
			XOffset:       0,
			YOffset:       0,
			Size:          ctx.Size,
			Adjustability: Adjustability{},
			Range:         Range{Start: start, End: end},
			SafeToBreak:   true,
			Char:          c,
			IsJustifiable: isJustifiable(c, script, xAdvance, [2]Em{0, 0}),
			Script:        script,
		})
	}

	if ctx.Dir.IsPositive() {
		byteIdx := 0
		for _, c := range text {
			addGlyph(byteIdx, c)
			byteIdx += len(string(c))
		}
	} else {
		// RTL: reverse order
		runes := []rune(text)
		byteIdx := len(text)
		for i := len(runes) - 1; i >= 0; i-- {
			c := runes[i]
			byteIdx -= len(string(c))
			addGlyph(byteIdx, c)
		}
	}
}

// trackAndSpace applies tracking and word spacing.
func trackAndSpace(ctx *ShapingContext) {
	// Simplified tracking - in production, this would read from styles
	tracking := Em(0) // Default no tracking
	spacing := Em(1)  // Default 100% word spacing

	for i := 0; i < len(ctx.glyphs); i++ {
		g := &ctx.glyphs[i]

		// Adjust NBSP to match regular space width
		if g.Char == '\u00A0' {
			// In production, would calculate delta from font metrics
		}

		// Apply word spacing
		if g.IsSpace() {
			g.XAdvance = g.XAdvance * spacing
		}

		// Apply tracking between glyphs (not at end of cluster)
		if i+1 < len(ctx.glyphs) && g.Range.Start != ctx.glyphs[i+1].Range.Start {
			g.XAdvance += tracking
		}
	}
}

// calculateAdjustability computes stretch/shrink values for justification.
func calculateAdjustability(ctx *ShapingContext, lang Lang, region *Region) {
	style := GetCJKPunctStyle(lang, region)

	for i := 0; i < len(ctx.glyphs); i++ {
		g := &ctx.glyphs[i]
		stretchable := i+1 >= len(ctx.glyphs) || g.Range.Start != ctx.glyphs[i+1].Range.Start

		g.Adjustability = baseAdjustability(g, style, stretchable)
	}

	// Apply consecutive CJK punctuation compression
	for i := 0; i < len(ctx.glyphs)-1; i++ {
		g := &ctx.glyphs[i]
		if g.IsCJKPunctuation() && style == CJKPunctStyleCNS {
			continue
		}

		next := &ctx.glyphs[i+1]
		width := g.XAdvance
		delta := width / 2.0

		if g.IsCJKPunctuation() && next.IsCJKPunctuation() {
			totalShrink := g.Shrinkability()[1] + next.Shrinkability()[0]
			if totalShrink >= delta {
				leftDelta := min(g.Shrinkability()[1], delta)
				g.ShrinkRight(leftDelta)
				next.ShrinkLeft(delta - leftDelta)
			}
		}
	}
}

// baseAdjustability computes base adjustability for a glyph.
func baseAdjustability(g *ShapedGlyph, style CJKPunctStyle, stretchable bool) Adjustability {
	width := g.XAdvance
	limited := func(v Em) Em {
		maxV := width * 0.75
		if v > maxV {
			return maxV
		}
		return v
	}

	if g.IsSpace() {
		// Space can stretch/shrink significantly
		return Adjustability{
			Stretchability: [2]Em{0, width * 0.5},
			Shrinkability:  [2]Em{0, limited(width * 0.33)},
		}
	}

	if g.IsCJKLeftAlignedPunctuation(style) {
		return Adjustability{
			Stretchability: [2]Em{0, 0},
			Shrinkability:  [2]Em{0, width / 2.0},
		}
	}

	if g.IsCJKRightAlignedPunctuation() {
		return Adjustability{
			Stretchability: [2]Em{0, 0},
			Shrinkability:  [2]Em{width / 2.0, 0},
		}
	}

	if g.IsCJKCenterAlignedPunctuation(style) {
		return Adjustability{
			Stretchability: [2]Em{0, 0},
			Shrinkability:  [2]Em{width / 4.0, width / 4.0},
		}
	}

	if stretchable {
		// Small tracking adjustment for inter-glyph space
		return Adjustability{
			Stretchability: [2]Em{0, width * 0.02},
			Shrinkability:  [2]Em{0, limited(width * 0.02)},
		}
	}

	return Adjustability{}
}

// ShapeRange shapes a range of text, splitting by bidi level and script.
func ShapeRange(ctx *ShapingContext, text string, base int, start, end int, bidiPara *bidi.Paragraph) []*ShapedText {
	if start >= end {
		return nil
	}

	segment := text[start:end]
	var results []*ShapedText

	// Get bidi ordering
	ordering, err := bidiPara.Order()
	if err != nil {
		// Fallback: shape as single segment
		shaped := Shape(ctx, base+start, segment, DirLTR, "", nil)
		return []*ShapedText{shaped}
	}

	// Process runs in visual order
	for i := 0; i < ordering.NumRuns(); i++ {
		run := ordering.Run(i)
		runStart, runEnd := run.Pos()
		runDir := DirLTR
		if run.Direction() == bidi.RightToLeft {
			runDir = DirRTL
		}

		// Adjust to segment coordinates
		if runStart < start {
			runStart = start
		}
		if runEnd > end {
			runEnd = end
		}
		if runStart >= runEnd {
			continue
		}

		runText := text[runStart:runEnd]
		shaped := Shape(ctx, base+runStart, runText, runDir, "", nil)
		results = append(results, shaped)
	}

	return results
}

// Utility functions.

func containsFace(faces []*font.Face, face *font.Face) bool {
	for _, f := range faces {
		if f == face {
			return true
		}
	}
	return false
}

func toFixed(f float64) fixed.Int26_6 {
	return fixed.Int26_6(f * 64) // 26.6 fixed point
}

func min(a, b Em) Em {
	if a < b {
		return a
	}
	return b
}

// Debug helpers.

func (s *ShapedText) String() string {
	return fmt.Sprintf("ShapedText{text=%q, dir=%v, glyphs=%d}", s.Text, s.Dir, s.Glyphs.Len())
}
