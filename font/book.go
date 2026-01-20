package font

import (
	"math"
	"sort"
	"strings"
	"sync"
)

// FontBook manages a collection of fonts and provides lookup functionality.
type FontBook struct {
	// fonts is the list of all loaded fonts, indexed by their position.
	fonts []*Font

	// byFamily groups fonts by normalized family name.
	byFamily map[string][]*Font

	mu sync.RWMutex
}

// NewFontBook creates a new empty FontBook.
func NewFontBook() *FontBook {
	return &FontBook{
		fonts:    make([]*Font, 0),
		byFamily: make(map[string][]*Font),
	}
}

// Add adds fonts to the book.
func (b *FontBook) Add(fonts ...*Font) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, f := range fonts {
		b.fonts = append(b.fonts, f)

		// Index by family
		family := normalizeFamily(f.Info.Family)
		b.byFamily[family] = append(b.byFamily[family], f)
	}
}

// Len returns the number of fonts in the book.
func (b *FontBook) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.fonts)
}

// Font returns the font at the given index.
// Returns nil if index is out of bounds.
func (b *FontBook) Font(index int) *Font {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if index < 0 || index >= len(b.fonts) {
		return nil
	}
	return b.fonts[index]
}

// Fonts returns all fonts in the book.
func (b *FontBook) Fonts() []*Font {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]*Font, len(b.fonts))
	copy(result, b.fonts)
	return result
}

// Families returns all unique family names in the book.
func (b *FontBook) Families() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	families := make([]string, 0, len(b.byFamily))
	for family := range b.byFamily {
		families = append(families, family)
	}
	sort.Strings(families)
	return families
}

// Select selects the best matching font for the given criteria.
// Returns nil if no suitable font is found.
func (b *FontBook) Select(families []string, variant Variant) *Font {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Try each family in order
	for _, family := range families {
		normalized := normalizeFamily(family)
		candidates := b.byFamily[normalized]
		if len(candidates) == 0 {
			continue
		}

		// Find best match for variant
		best := selectBestVariant(candidates, variant)
		if best != nil {
			return best
		}
	}

	return nil
}

// SelectWithFallback selects a font with fallback to any available font.
func (b *FontBook) SelectWithFallback(families []string, variant Variant) *Font {
	// Try exact match first
	if f := b.Select(families, variant); f != nil {
		return f
	}

	// Fallback to any font
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.fonts) > 0 {
		return b.fonts[0]
	}

	return nil
}

// FindByFamily returns all fonts matching the given family name.
func (b *FontBook) FindByFamily(family string) []*Font {
	b.mu.RLock()
	defer b.mu.RUnlock()

	normalized := normalizeFamily(family)
	fonts := b.byFamily[normalized]
	if len(fonts) == 0 {
		return nil
	}

	result := make([]*Font, len(fonts))
	copy(result, fonts)
	return result
}

// IndexOf returns the index of the given font in the book.
// Returns -1 if not found.
func (b *FontBook) IndexOf(f *Font) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for i, font := range b.fonts {
		if font == f {
			return i
		}
	}
	return -1
}

// selectBestVariant finds the best matching font for a variant.
func selectBestVariant(fonts []*Font, variant Variant) *Font {
	if len(fonts) == 0 {
		return nil
	}

	var best *Font
	bestScore := math.MaxFloat64

	for _, f := range fonts {
		score := variantDistance(f.Info, variant)
		if score < bestScore {
			bestScore = score
			best = f
		}
	}

	return best
}

// variantDistance calculates the distance between a font's properties and a target variant.
// Lower is better.
func variantDistance(info FontInfo, target Variant) float64 {
	var distance float64

	// Style mismatch is significant
	if info.Style != target.Style {
		distance += 10.0

		// Oblique is somewhat close to italic
		if (info.Style == StyleOblique && target.Style == StyleItalic) ||
			(info.Style == StyleItalic && target.Style == StyleOblique) {
			distance -= 5.0
		}
	}

	// Weight distance (normalize to 0-1 scale for 100-900 range)
	weightDiff := math.Abs(float64(info.Weight-target.Weight)) / 400.0
	distance += weightDiff * 5.0

	// Stretch distance
	stretchDiff := math.Abs(float64(info.Stretch - target.Stretch))
	distance += stretchDiff * 2.0

	return distance
}

// normalizeFamily normalizes a font family name for comparison.
func normalizeFamily(family string) string {
	// Lowercase
	s := strings.ToLower(family)

	// Remove common suffixes
	s = strings.TrimSuffix(s, " regular")
	s = strings.TrimSuffix(s, " normal")

	// Normalize whitespace
	s = strings.Join(strings.Fields(s), " ")

	return s
}

// SystemFontBook creates a FontBook loaded with system fonts.
func SystemFontBook() (*FontBook, error) {
	fonts, err := DiscoverSystemFonts()
	if err != nil {
		return nil, err
	}

	book := NewFontBook()
	book.Add(fonts...)
	return book, nil
}
