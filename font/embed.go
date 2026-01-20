package font

import (
	"embed"
	"io/fs"
)

// EmbeddedFonts provides access to bundled fallback fonts.
// To use, set this to an embed.FS containing font files.
//
// Example usage in your main package:
//
//	//go:embed fonts/*.ttf fonts/*.otf
//	var embeddedFonts embed.FS
//
//	func init() {
//	    font.EmbeddedFonts = &embeddedFonts
//	}
var EmbeddedFonts *embed.FS

// LoadEmbeddedFonts loads all fonts from the embedded filesystem.
// Returns nil if no embedded fonts are configured.
func LoadEmbeddedFonts() ([]*Font, error) {
	if EmbeddedFonts == nil {
		return nil, nil
	}

	return LoadFromFS(EmbeddedFonts, ".")
}

// LoadFromFS loads all fonts from a filesystem (embed.FS, os.DirFS, etc.).
func LoadFromFS(fsys fs.FS, root string) ([]*Font, error) {
	var fonts []*Font

	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if d.IsDir() {
			return nil
		}

		if !IsFontFile(path) {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil // Skip unreadable files
		}

		loaded, err := LoadFromBytes(data, path)
		if err != nil {
			return nil // Skip unparseable fonts
		}

		fonts = append(fonts, loaded...)
		return nil
	})

	if err != nil {
		return fonts, err
	}

	return fonts, nil
}

// DefaultFallbackFamilies returns a list of common fallback font families
// to try when a requested font is not available.
func DefaultFallbackFamilies() []string {
	return []string{
		// Sans-serif fallbacks
		"Noto Sans",
		"DejaVu Sans",
		"Liberation Sans",
		"Arial",
		"Helvetica",
		"sans-serif",

		// Serif fallbacks
		"Noto Serif",
		"DejaVu Serif",
		"Liberation Serif",
		"Times New Roman",
		"Times",
		"serif",

		// Monospace fallbacks
		"Noto Sans Mono",
		"DejaVu Sans Mono",
		"Liberation Mono",
		"Courier New",
		"Courier",
		"monospace",
	}
}

// GenericFamilyMapping maps generic family names to concrete font families.
var GenericFamilyMapping = map[string][]string{
	"sans-serif": {
		"Noto Sans",
		"DejaVu Sans",
		"Liberation Sans",
		"Arial",
		"Helvetica",
	},
	"serif": {
		"Noto Serif",
		"DejaVu Serif",
		"Liberation Serif",
		"Times New Roman",
		"Times",
	},
	"monospace": {
		"Noto Sans Mono",
		"DejaVu Sans Mono",
		"Liberation Mono",
		"Courier New",
		"Courier",
	},
	"cursive": {
		"Comic Sans MS",
		"Apple Chancery",
		"cursive",
	},
	"fantasy": {
		"Impact",
		"Papyrus",
		"fantasy",
	},
	"system-ui": {
		"SF Pro Text",          // macOS
		"Segoe UI",             // Windows
		"Ubuntu",               // Ubuntu Linux
		"Cantarell",            // GNOME
		"Noto Sans",            // Fallback
	},
	"ui-sans-serif": {
		"SF Pro Text",
		"Segoe UI",
		"system-ui",
	},
	"ui-serif": {
		"New York",
		"Georgia",
		"serif",
	},
	"ui-monospace": {
		"SF Mono",
		"Consolas",
		"monospace",
	},
}

// ExpandGenericFamily expands a generic family name to concrete families.
// Returns the original family in a slice if not a generic family.
func ExpandGenericFamily(family string) []string {
	if families, ok := GenericFamilyMapping[family]; ok {
		return families
	}
	return []string{family}
}

// ExpandFamilies expands a list of families, replacing generic names.
func ExpandFamilies(families []string) []string {
	var expanded []string
	seen := make(map[string]bool)

	for _, family := range families {
		for _, f := range ExpandGenericFamily(family) {
			if !seen[f] {
				seen[f] = true
				expanded = append(expanded, f)
			}
		}
	}

	return expanded
}
