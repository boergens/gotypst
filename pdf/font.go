package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"sort"

	"github.com/boergens/gotypst/font"
	typfont "github.com/go-text/typesetting/font"
)

// GlyphSet tracks used glyphs for a font.
type GlyphSet struct {
	Font    *font.Font
	Face    *typfont.Face
	Glyphs  map[uint16]rune // GlyphID -> representative character
	Name    string          // PDF resource name (e.g., "F1")
}

// Add adds a glyph to the set.
func (gs *GlyphSet) Add(glyphID uint16, char rune) {
	if gs.Glyphs == nil {
		gs.Glyphs = make(map[uint16]rune)
	}
	if _, exists := gs.Glyphs[glyphID]; !exists {
		gs.Glyphs[glyphID] = char
	}
}

// SortedGlyphIDs returns glyph IDs in sorted order.
func (gs *GlyphSet) SortedGlyphIDs() []uint16 {
	ids := make([]uint16, 0, len(gs.Glyphs))
	for id := range gs.Glyphs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// GlyphCollector collects glyph usage across fonts during rendering.
type GlyphCollector struct {
	fonts    map[*typfont.Face]*GlyphSet
	fontList []*GlyphSet // Ordered list for deterministic output
	nextID   int
}

// NewGlyphCollector creates a new glyph collector.
func NewGlyphCollector() *GlyphCollector {
	return &GlyphCollector{
		fonts:  make(map[*typfont.Face]*GlyphSet),
		nextID: 1,
	}
}

// Record records a glyph usage for a font face.
func (gc *GlyphCollector) Record(face *typfont.Face, glyphID uint16, char rune) {
	gs, ok := gc.fonts[face]
	if !ok {
		gs = &GlyphSet{
			Face:   face,
			Glyphs: make(map[uint16]rune),
			Name:   fmt.Sprintf("F%d", gc.nextID),
		}
		gc.nextID++
		gc.fonts[face] = gs
		gc.fontList = append(gc.fontList, gs)
	}
	gs.Add(glyphID, char)
}

// SetFont associates a Font struct with a face.
func (gc *GlyphCollector) SetFont(face *typfont.Face, f *font.Font) {
	if gs, ok := gc.fonts[face]; ok {
		gs.Font = f
	}
}

// Fonts returns all collected font sets.
func (gc *GlyphCollector) Fonts() []*GlyphSet {
	return gc.fontList
}

// FontName returns the PDF resource name for a font face.
func (gc *GlyphCollector) FontName(face interface{}) string {
	if f, ok := face.(*typfont.Face); ok {
		if gs, exists := gc.fonts[f]; exists {
			return "/" + gs.Name
		}
	}
	return "/F1" // Fallback
}

// FontXObject represents an embedded font for PDF.
type FontXObject struct {
	Ref        Ref
	Name       string           // PDF resource name
	BaseFont   string           // PostScript font name
	Subtype    string           // TrueType, Type0, Type1, etc.
	Encoding   string           // Encoding name
	GlyphSet   *GlyphSet        // Source glyph data
	SubsetData []byte           // Subsetted font data (compressed)
	GlyphMap   map[uint16]uint16 // Original GlyphID -> Subset GlyphID
	Widths     []int            // Glyph widths for /Widths array
	FirstChar  int
	LastChar   int
}

// SubsetFont creates a subset font containing only the used glyphs.
func SubsetFont(gs *GlyphSet) (*FontXObject, error) {
	if gs.Font == nil || len(gs.Font.Data) == 0 {
		return nil, fmt.Errorf("font data not available for subsetting")
	}

	// Get sorted glyph IDs
	glyphIDs := gs.SortedGlyphIDs()
	if len(glyphIDs) == 0 {
		return nil, fmt.Errorf("no glyphs to subset")
	}

	// Always include .notdef (glyph 0)
	hasNotdef := false
	for _, id := range glyphIDs {
		if id == 0 {
			hasNotdef = true
			break
		}
	}
	if !hasNotdef {
		glyphIDs = append([]uint16{0}, glyphIDs...)
	}

	// Create glyph mapping (original ID -> new ID)
	glyphMap := make(map[uint16]uint16)
	for newID, oldID := range glyphIDs {
		glyphMap[oldID] = uint16(newID)
	}

	// Perform font subsetting
	subsetData, err := subsetTTF(gs.Font.Data, gs.Font.Index, glyphIDs)
	if err != nil {
		return nil, fmt.Errorf("subset font: %w", err)
	}

	// Compress the subset data
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(subsetData); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	// Calculate widths
	widths := calculateWidths(gs.Face, glyphIDs)

	// Generate subset tag (6 uppercase letters)
	subsetTag := generateSubsetTag(gs.Name)
	baseFontName := subsetTag + "+" + sanitizeFontName(gs.Font.Info.Family)

	return &FontXObject{
		Name:       gs.Name,
		BaseFont:   baseFontName,
		Subtype:    "TrueType",
		Encoding:   "Identity-H",
		GlyphSet:   gs,
		SubsetData: compressed.Bytes(),
		GlyphMap:   glyphMap,
		Widths:     widths,
		FirstChar:  0,
		LastChar:   len(glyphIDs) - 1,
	}, nil
}

// subsetTTF performs TrueType font subsetting.
// This is a simplified implementation that creates a minimal valid font.
func subsetTTF(data []byte, index int, glyphIDs []uint16) ([]byte, error) {
	// For TTC files, we need to extract the specific font
	if len(data) >= 4 && string(data[:4]) == "ttcf" {
		extracted, err := extractFromTTC(data, index)
		if err != nil {
			return nil, err
		}
		data = extracted
	}

	// Parse the font tables
	tables, err := parseTTFTables(data)
	if err != nil {
		return nil, err
	}

	// Build subset font with required tables
	return buildSubsetFont(tables, glyphIDs)
}

// TTFTable represents a TrueType table.
type TTFTable struct {
	Tag      string
	Checksum uint32
	Offset   uint32
	Length   uint32
	Data     []byte
}

// parseTTFTables parses TrueType font tables.
func parseTTFTables(data []byte) (map[string]*TTFTable, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("font data too short")
	}

	// Read offset table
	numTables := int(beUint16(data[4:6]))
	if len(data) < 12+numTables*16 {
		return nil, fmt.Errorf("invalid table directory")
	}

	tables := make(map[string]*TTFTable)
	for i := 0; i < numTables; i++ {
		offset := 12 + i*16
		tag := string(data[offset : offset+4])
		checksum := beUint32(data[offset+4 : offset+8])
		tableOffset := beUint32(data[offset+8 : offset+12])
		length := beUint32(data[offset+12 : offset+16])

		if int(tableOffset+length) > len(data) {
			return nil, fmt.Errorf("table %s extends beyond file", tag)
		}

		tables[tag] = &TTFTable{
			Tag:      tag,
			Checksum: checksum,
			Offset:   tableOffset,
			Length:   length,
			Data:     data[tableOffset : tableOffset+length],
		}
	}

	return tables, nil
}

// buildSubsetFont builds a new font with only the specified glyphs.
func buildSubsetFont(tables map[string]*TTFTable, glyphIDs []uint16) ([]byte, error) {
	// Required tables for a minimal TrueType font
	requiredTags := []string{"head", "hhea", "maxp", "post", "name", "cmap", "glyf", "loca", "hmtx"}

	// Check which required tables exist
	var availableTags []string
	for _, tag := range requiredTags {
		if _, ok := tables[tag]; ok {
			availableTags = append(availableTags, tag)
		}
	}

	// Add optional tables if present
	optionalTags := []string{"cvt ", "fpgm", "prep", "OS/2"}
	for _, tag := range optionalTags {
		if _, ok := tables[tag]; ok {
			availableTags = append(availableTags, tag)
		}
	}

	numTables := len(availableTags)
	if numTables == 0 {
		return nil, fmt.Errorf("no tables to include in subset")
	}

	// Calculate search range values
	searchRange := 1
	entrySelector := 0
	for searchRange*2 <= numTables {
		searchRange *= 2
		entrySelector++
	}
	searchRange *= 16
	rangeShift := numTables*16 - searchRange

	// Build output
	var out bytes.Buffer

	// Write offset table header
	out.Write([]byte{0, 1, 0, 0}) // sfnt version (TrueType)
	writeBeUint16(&out, uint16(numTables))
	writeBeUint16(&out, uint16(searchRange))
	writeBeUint16(&out, uint16(entrySelector))
	writeBeUint16(&out, uint16(rangeShift))

	// Calculate offsets for table data
	tableRecordSize := 16
	headerSize := 12 + numTables*tableRecordSize
	currentOffset := uint32(headerSize)

	// Prepare table data with 4-byte alignment
	type tableEntry struct {
		tag    string
		data   []byte
		offset uint32
	}
	entries := make([]tableEntry, 0, numTables)

	for _, tag := range availableTags {
		table := tables[tag]
		data := table.Data

		// For glyf and loca tables, we need to subset
		if tag == "glyf" || tag == "loca" {
			// For simplicity, include all original glyph data
			// A full implementation would extract only needed glyphs
		}

		// Align to 4 bytes
		padding := (4 - (len(data) % 4)) % 4

		entries = append(entries, tableEntry{
			tag:    tag,
			data:   data,
			offset: currentOffset,
		})
		currentOffset += uint32(len(data) + padding)
	}

	// Write table directory
	for _, entry := range entries {
		out.WriteString(entry.tag)
		writeBeUint32(&out, calculateChecksum(entry.data))
		writeBeUint32(&out, entry.offset)
		writeBeUint32(&out, uint32(len(entry.data)))
	}

	// Write table data
	for _, entry := range entries {
		out.Write(entry.data)
		// Pad to 4-byte boundary
		padding := (4 - (len(entry.data) % 4)) % 4
		for i := 0; i < padding; i++ {
			out.WriteByte(0)
		}
	}

	return out.Bytes(), nil
}

// extractFromTTC extracts a single font from a TrueType Collection.
func extractFromTTC(data []byte, index int) ([]byte, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("TTC data too short")
	}

	// Read TTC header
	numFonts := int(beUint32(data[8:12]))
	if index >= numFonts {
		return nil, fmt.Errorf("font index %d out of range (collection has %d fonts)", index, numFonts)
	}

	// Read offset for this font
	offsetPos := 12 + index*4
	if len(data) < offsetPos+4 {
		return nil, fmt.Errorf("invalid TTC offset table")
	}
	fontOffset := beUint32(data[offsetPos : offsetPos+4])

	// Read the font's offset table to determine its size
	if len(data) < int(fontOffset)+12 {
		return nil, fmt.Errorf("invalid font offset in TTC")
	}

	numTables := int(beUint16(data[fontOffset+4 : fontOffset+6]))

	// Find the extent of all tables
	var maxEnd uint32
	for i := 0; i < numTables; i++ {
		entryOffset := int(fontOffset) + 12 + i*16
		if len(data) < entryOffset+16 {
			return nil, fmt.Errorf("invalid table directory in TTC")
		}
		tableOffset := beUint32(data[entryOffset+8 : entryOffset+12])
		tableLength := beUint32(data[entryOffset+12 : entryOffset+16])
		end := tableOffset + tableLength
		if end > maxEnd {
			maxEnd = end
		}
	}

	// Return the font data (tables may be shared with other fonts in collection)
	// For subsetting, we'll rebuild with only the tables we need
	return data[fontOffset:], nil
}

// calculateWidths calculates glyph widths in PDF units (1/1000 em).
func calculateWidths(face *typfont.Face, glyphIDs []uint16) []int {
	widths := make([]int, len(glyphIDs))
	upem := face.Upem()

	for i, gid := range glyphIDs {
		advance := face.HorizontalAdvance(typfont.GID(gid))
		// Convert to 1/1000 em units
		widths[i] = int(float64(advance) * 1000 / float64(upem))
	}

	return widths
}

// generateSubsetTag generates a 6-character subset tag.
func generateSubsetTag(name string) string {
	// Simple hash-based tag generation
	var hash uint32
	for _, c := range name {
		hash = hash*31 + uint32(c)
	}

	tag := make([]byte, 6)
	for i := 0; i < 6; i++ {
		tag[i] = 'A' + byte(hash%26)
		hash /= 26
	}
	return string(tag)
}

// sanitizeFontName creates a valid PostScript font name.
func sanitizeFontName(name string) string {
	var result []byte
	for _, c := range name {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result = append(result, byte(c))
		}
	}
	if len(result) == 0 {
		return "Font"
	}
	return string(result)
}

// Helper functions for big-endian byte manipulation
func beUint16(b []byte) uint16 {
	return uint16(b[0])<<8 | uint16(b[1])
}

func beUint32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func writeBeUint16(buf *bytes.Buffer, v uint16) {
	buf.WriteByte(byte(v >> 8))
	buf.WriteByte(byte(v))
}

func writeBeUint32(buf *bytes.Buffer, v uint32) {
	buf.WriteByte(byte(v >> 24))
	buf.WriteByte(byte(v >> 16))
	buf.WriteByte(byte(v >> 8))
	buf.WriteByte(byte(v))
}

func calculateChecksum(data []byte) uint32 {
	// Pad to 4-byte boundary for calculation
	padded := data
	if len(data)%4 != 0 {
		padded = make([]byte, len(data)+(4-len(data)%4))
		copy(padded, data)
	}

	var sum uint32
	for i := 0; i < len(padded); i += 4 {
		sum += beUint32(padded[i : i+4])
	}
	return sum
}
