package font

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-text/typesetting/font"
)

// LoadFromFile loads fonts from a file path.
// Returns multiple fonts for TTC (font collection) files.
func LoadFromFile(path string) ([]*Font, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read font file: %w", err)
	}

	return LoadFromBytes(data, path)
}

// LoadFromBytes loads fonts from raw bytes.
// The path parameter is used for metadata (can be empty for embedded fonts).
func LoadFromBytes(data []byte, path string) ([]*Font, error) {
	if len(data) < 4 {
		return nil, errors.New("font data too short")
	}

	// Check if it's a font collection (TTC)
	if isTTC(data) {
		return loadTTC(data, path)
	}

	// Single font (TTF/OTF)
	return loadSingle(data, path, 0)
}

// isTTC checks if the data starts with a TTC header.
func isTTC(data []byte) bool {
	return len(data) >= 4 && string(data[:4]) == "ttcf"
}

// loadTTC loads fonts from a TrueType Collection.
func loadTTC(data []byte, path string) ([]*Font, error) {
	resource := bytes.NewReader(data)
	faces, err := font.ParseTTC(resource)
	if err != nil {
		return nil, fmt.Errorf("parse TTC: %w", err)
	}

	// Keep a copy of the raw TTC data for subsetting
	// Each font in the collection shares this data
	rawData := make([]byte, len(data))
	copy(rawData, data)

	fonts := make([]*Font, 0, len(faces))
	for i, face := range faces {
		info := extractInfo(face)
		fonts = append(fonts, &Font{
			face:    face,
			Info:    info,
			Path:    path,
			Index:   i,
			RawData: rawData, // Shared reference for TTC
		})
	}

	return fonts, nil
}

// loadSingle loads a single font (TTF/OTF).
func loadSingle(data []byte, path string, index int) ([]*Font, error) {
	resource := bytes.NewReader(data)
	face, err := font.ParseTTF(resource)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	// Keep a copy of the raw data for subsetting
	rawData := make([]byte, len(data))
	copy(rawData, data)

	info := extractInfo(face)
	return []*Font{{
		face:    face,
		Info:    info,
		Path:    path,
		Index:   index,
		RawData: rawData,
	}}, nil
}

// extractInfo extracts FontInfo from a font face.
func extractInfo(face *font.Face) FontInfo {
	info := FontInfo{
		Style:   StyleNormal,
		Weight:  WeightNormal,
		Stretch: StretchNormal,
	}

	if face.Font == nil {
		return info
	}

	// Extract from font Description using the Font's Describe method
	desc := face.Font.Describe()

	info.Family = desc.Family
	info.FullName = desc.Family

	// Map style
	switch desc.Aspect.Style {
	case font.StyleItalic:
		info.Style = StyleItalic
	case font.StyleNormal:
		info.Style = StyleNormal
	default:
		info.Style = StyleOblique
	}

	// Map weight
	info.Weight = Weight(desc.Aspect.Weight)
	if info.Weight == 0 {
		info.Weight = WeightNormal
	}

	// Map stretch
	info.Stretch = Stretch(desc.Aspect.Stretch)
	if info.Stretch == 0 {
		info.Stretch = StretchNormal
	}

	return info
}

// IsFontFile checks if a path points to a supported font file.
func IsFontFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".ttf", ".otf", ".ttc", ".otc":
		return true
	default:
		return false
	}
}
