// Package pdf provides PDF export functionality for Typst documents.
package pdf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/font/opentype"
)

// FontRef uniquely identifies a font for PDF embedding.
// Two fonts with the same underlying data should have the same FontRef.
type FontRef struct {
	// ID is a unique identifier assigned during PDF generation.
	ID int
	// Face is the font face this reference points to.
	Face *font.Face
	// Loader provides access to raw font table data.
	Loader *opentype.Loader
	// Source is the original font file bytes (optional, for subsetting).
	Source []byte
}

// FontUsage tracks which glyphs are used from a specific font.
type FontUsage struct {
	// Ref identifies the font.
	Ref FontRef
	// GlyphIDs contains all glyph IDs used from this font.
	// Maps GlyphID to the Unicode codepoint(s) that produced it.
	GlyphIDs map[uint16][]rune
	// Ordered list of glyph IDs for consistent output.
	orderedGlyphs []uint16
}

// AddGlyph records that a glyph was used for a given character.
func (fu *FontUsage) AddGlyph(glyphID uint16, char rune) {
	if fu.GlyphIDs == nil {
		fu.GlyphIDs = make(map[uint16][]rune)
	}
	if chars, ok := fu.GlyphIDs[glyphID]; ok {
		// Check if char already recorded
		for _, c := range chars {
			if c == char {
				return
			}
		}
		fu.GlyphIDs[glyphID] = append(chars, char)
	} else {
		fu.GlyphIDs[glyphID] = []rune{char}
		fu.orderedGlyphs = append(fu.orderedGlyphs, glyphID)
	}
}

// OrderedGlyphIDs returns glyph IDs in insertion order.
func (fu *FontUsage) OrderedGlyphIDs() []uint16 {
	return fu.orderedGlyphs
}

// SortedGlyphIDs returns glyph IDs in sorted order for deterministic output.
func (fu *FontUsage) SortedGlyphIDs() []uint16 {
	sorted := make([]uint16, len(fu.orderedGlyphs))
	copy(sorted, fu.orderedGlyphs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return sorted
}

// FontMap manages fonts used in a PDF document.
type FontMap struct {
	mu     sync.RWMutex
	fonts  map[*font.Face]*FontUsage
	nextID int
}

// NewFontMap creates a new FontMap.
func NewFontMap() *FontMap {
	return &FontMap{
		fonts:  make(map[*font.Face]*FontUsage),
		nextID: 1,
	}
}

// RegisterFont registers a font with its loader and optional source data.
// This should be called before RegisterGlyph to provide font data for embedding.
func (fm *FontMap) RegisterFont(face *font.Face, loader *opentype.Loader, source []byte) FontRef {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	usage, ok := fm.fonts[face]
	if !ok {
		usage = &FontUsage{
			Ref: FontRef{
				ID:     fm.nextID,
				Face:   face,
				Loader: loader,
				Source: source,
			},
			GlyphIDs: make(map[uint16][]rune),
		}
		fm.fonts[face] = usage
		fm.nextID++
	} else {
		// Update loader/source if not set
		if usage.Ref.Loader == nil && loader != nil {
			usage.Ref.Loader = loader
		}
		if usage.Ref.Source == nil && source != nil {
			usage.Ref.Source = source
		}
	}
	return usage.Ref
}

// RegisterGlyph records that a glyph was used from a font.
// Returns the FontRef for the font.
func (fm *FontMap) RegisterGlyph(face *font.Face, glyphID uint16, char rune) FontRef {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	usage, ok := fm.fonts[face]
	if !ok {
		usage = &FontUsage{
			Ref: FontRef{
				ID:   fm.nextID,
				Face: face,
			},
			GlyphIDs: make(map[uint16][]rune),
		}
		fm.fonts[face] = usage
		fm.nextID++
	}
	usage.AddGlyph(glyphID, char)
	return usage.Ref
}

// GetUsage returns the font usage for a face.
func (fm *FontMap) GetUsage(face *font.Face) (*FontUsage, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	usage, ok := fm.fonts[face]
	return usage, ok
}

// AllFonts returns all font usages in ID order.
func (fm *FontMap) AllFonts() []*FontUsage {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	usages := make([]*FontUsage, 0, len(fm.fonts))
	for _, usage := range fm.fonts {
		usages = append(usages, usage)
	}
	sort.Slice(usages, func(i, j int) bool {
		return usages[i].Ref.ID < usages[j].Ref.ID
	})
	return usages
}

// FontEmbedding contains data needed to embed a font in PDF.
type FontEmbedding struct {
	// Ref identifies the font.
	Ref FontRef
	// Subset is the subsetted font data (or full font if not subsetted).
	Subset []byte
	// GlyphMapping maps original GlyphID to new GlyphID in subset.
	// If nil, glyph IDs are unchanged (full font embedded).
	GlyphMapping map[uint16]uint16
	// ToUnicode maps glyph IDs to Unicode codepoints for text extraction.
	ToUnicode map[uint16][]rune
	// Widths maps glyph IDs to glyph widths in PDF units (1/1000 of em).
	Widths map[uint16]int
	// PostScriptName is the font's PostScript name.
	PostScriptName string
	// IsCIDFont indicates if the font should be embedded as a CID font.
	IsCIDFont bool
}

// FontSubsetter handles font subsetting for PDF embedding.
type FontSubsetter struct {
	usage *FontUsage
}

// NewFontSubsetter creates a subsetter for a font usage.
func NewFontSubsetter(usage *FontUsage) *FontSubsetter {
	return &FontSubsetter{usage: usage}
}

// Subset creates a subsetted font containing only the used glyphs.
// For TrueType/OpenType fonts, this creates a valid font file.
func (fs *FontSubsetter) Subset() (*FontEmbedding, error) {
	face := fs.usage.Ref.Face

	// Get font data
	fontData, err := getFontData(&fs.usage.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get font data: %w", err)
	}

	// Get PostScript name
	psName := getPostScriptName(face)

	// Get glyph widths
	widths := make(map[uint16]int)
	upem := getUnitsPerEm(face)
	for glyphID := range fs.usage.GlyphIDs {
		advance := getGlyphAdvance(face, glyphID)
		// Convert to PDF units (1/1000 of em)
		widths[glyphID] = int(float64(advance) * 1000.0 / float64(upem))
	}

	// Determine if we should use CID font (for fonts with >255 glyphs or CJK)
	isCID := len(fs.usage.GlyphIDs) > 255 || hasCJKGlyphs(fs.usage)

	// For now, embed the full font. Full subsetting would require
	// parsing and modifying the font tables (glyf, loca, etc.).
	// This is a common approach for initial implementations.
	embedding := &FontEmbedding{
		Ref:            fs.usage.Ref,
		Subset:         fontData,
		GlyphMapping:   nil, // No remapping - using original glyph IDs
		ToUnicode:      copyToUnicode(fs.usage.GlyphIDs),
		Widths:         widths,
		PostScriptName: psName,
		IsCIDFont:      isCID,
	}

	return embedding, nil
}

// SubsetMinimal creates a minimal subset containing only used glyphs.
// This produces smaller files but requires more complex font table manipulation.
func (fs *FontSubsetter) SubsetMinimal() (*FontEmbedding, error) {
	face := fs.usage.Ref.Face

	fontData, err := getFontData(&fs.usage.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get font data: %w", err)
	}

	// Parse the font to determine format
	format := detectFontFormat(fontData)
	if format == FontFormatUnknown {
		// Fall back to full embedding
		return fs.Subset()
	}

	// Create minimal subset
	subsetData, glyphMapping, err := createMinimalSubset(fontData, fs.usage.SortedGlyphIDs(), format)
	if err != nil {
		// Fall back to full embedding on error
		return fs.Subset()
	}

	psName := getPostScriptName(face)
	widths := make(map[uint16]int)
	upem := getUnitsPerEm(face)

	for glyphID := range fs.usage.GlyphIDs {
		advance := getGlyphAdvance(face, glyphID)
		widths[glyphID] = int(float64(advance) * 1000.0 / float64(upem))
	}

	return &FontEmbedding{
		Ref:            fs.usage.Ref,
		Subset:         subsetData,
		GlyphMapping:   glyphMapping,
		ToUnicode:      copyToUnicode(fs.usage.GlyphIDs),
		Widths:         widths,
		PostScriptName: psName + "-Subset",
		IsCIDFont:      len(fs.usage.GlyphIDs) > 255,
	}, nil
}

// FontFormat represents a font file format.
type FontFormat int

const (
	FontFormatUnknown FontFormat = iota
	FontFormatTTF
	FontFormatOTF
	FontFormatWOFF
	FontFormatWOFF2
)

// detectFontFormat identifies the font format from the file data.
func detectFontFormat(data []byte) FontFormat {
	if len(data) < 4 {
		return FontFormatUnknown
	}

	// Check magic bytes
	switch {
	case bytes.HasPrefix(data, []byte{0x00, 0x01, 0x00, 0x00}):
		return FontFormatTTF // TrueType
	case bytes.HasPrefix(data, []byte("true")):
		return FontFormatTTF // TrueType (Mac)
	case bytes.HasPrefix(data, []byte("OTTO")):
		return FontFormatOTF // OpenType with CFF
	case bytes.HasPrefix(data, []byte("wOFF")):
		return FontFormatWOFF
	case bytes.HasPrefix(data, []byte("wOF2")):
		return FontFormatWOFF2
	default:
		return FontFormatUnknown
	}
}

// createMinimalSubset creates a font subset containing only specified glyphs.
func createMinimalSubset(fontData []byte, glyphIDs []uint16, format FontFormat) ([]byte, map[uint16]uint16, error) {
	switch format {
	case FontFormatTTF:
		return subsetTTF(fontData, glyphIDs)
	case FontFormatOTF:
		return subsetOTF(fontData, glyphIDs)
	default:
		return nil, nil, fmt.Errorf("unsupported font format for subsetting")
	}
}

// subsetTTF creates a TrueType font subset.
func subsetTTF(fontData []byte, glyphIDs []uint16) ([]byte, map[uint16]uint16, error) {
	// Parse font tables
	tables, err := parseFontTables(fontData)
	if err != nil {
		return nil, nil, err
	}

	// Always include glyph 0 (notdef)
	glyphSet := make(map[uint16]bool)
	glyphSet[0] = true
	for _, gid := range glyphIDs {
		glyphSet[gid] = true
	}

	// Add composite glyph components
	if glyfData, ok := tables["glyf"]; ok {
		if locaData, ok := tables["loca"]; ok {
			addCompositeComponents(glyphSet, glyfData, locaData, tables)
		}
	}

	// Create glyph ID mapping (old ID -> new ID)
	sortedGlyphs := make([]uint16, 0, len(glyphSet))
	for gid := range glyphSet {
		sortedGlyphs = append(sortedGlyphs, gid)
	}
	sort.Slice(sortedGlyphs, func(i, j int) bool { return sortedGlyphs[i] < sortedGlyphs[j] })

	mapping := make(map[uint16]uint16)
	for newID, oldID := range sortedGlyphs {
		mapping[oldID] = uint16(newID)
	}

	// Build subset tables
	subsetTables := make(map[string][]byte)

	// Copy required tables unchanged
	for _, name := range []string{"head", "hhea", "maxp", "OS/2", "name", "post", "cvt ", "fpgm", "prep"} {
		if data, ok := tables[name]; ok {
			subsetTables[name] = data
		}
	}

	// Subset glyf and loca tables
	if glyfData, ok := tables["glyf"]; ok {
		if locaData, ok := tables["loca"]; ok {
			newGlyf, newLoca, err := subsetGlyfLoca(glyfData, locaData, sortedGlyphs, tables)
			if err != nil {
				return nil, nil, err
			}
			subsetTables["glyf"] = newGlyf
			subsetTables["loca"] = newLoca
		}
	}

	// Subset hmtx table
	if hmtxData, ok := tables["hmtx"]; ok {
		subsetTables["hmtx"] = subsetHmtx(hmtxData, sortedGlyphs, tables)
	}

	// Update maxp with new glyph count
	if maxpData, ok := subsetTables["maxp"]; ok && len(maxpData) >= 6 {
		maxpCopy := make([]byte, len(maxpData))
		copy(maxpCopy, maxpData)
		binary.BigEndian.PutUint16(maxpCopy[4:6], uint16(len(sortedGlyphs)))
		subsetTables["maxp"] = maxpCopy
	}

	// Build the final font file
	subsetData, err := buildFontFile(subsetTables)
	if err != nil {
		return nil, nil, err
	}

	return subsetData, mapping, nil
}

// subsetOTF creates an OpenType (CFF) font subset.
func subsetOTF(fontData []byte, glyphIDs []uint16) ([]byte, map[uint16]uint16, error) {
	// CFF subsetting is more complex; for now fall back to TrueType logic
	// if the font has a glyf table, otherwise return error
	tables, err := parseFontTables(fontData)
	if err != nil {
		return nil, nil, err
	}

	if _, hasGlyf := tables["glyf"]; hasGlyf {
		return subsetTTF(fontData, glyphIDs)
	}

	// CFF subsetting requires specialized handling
	return nil, nil, fmt.Errorf("CFF font subsetting not yet implemented")
}

// parseFontTables parses the font file and returns table data.
func parseFontTables(data []byte) (map[string][]byte, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("font data too short")
	}

	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	if len(data) < 12+numTables*16 {
		return nil, fmt.Errorf("font data too short for table directory")
	}

	tables := make(map[string][]byte)
	for i := 0; i < numTables; i++ {
		offset := 12 + i*16
		tag := string(data[offset : offset+4])
		tableOffset := int(binary.BigEndian.Uint32(data[offset+8 : offset+12]))
		tableLength := int(binary.BigEndian.Uint32(data[offset+12 : offset+16]))

		if tableOffset+tableLength > len(data) {
			return nil, fmt.Errorf("table %s extends beyond font data", tag)
		}

		tables[tag] = data[tableOffset : tableOffset+tableLength]
	}

	return tables, nil
}

// addCompositeComponents adds glyph IDs referenced by composite glyphs.
func addCompositeComponents(glyphSet map[uint16]bool, glyfData, locaData []byte, tables map[string][]byte) {
	// Determine loca format from head table
	headData, ok := tables["head"]
	if !ok || len(headData) < 52 {
		return
	}
	locaFormat := int(binary.BigEndian.Uint16(headData[50:52]))

	// Get glyph offsets
	offsets := getGlyphOffsets(locaData, locaFormat, len(glyfData))

	// Process each glyph to find composite references
	for gid := range glyphSet {
		if int(gid) >= len(offsets)-1 {
			continue
		}
		start := offsets[gid]
		end := offsets[gid+1]
		if start >= end || start >= len(glyfData) {
			continue
		}
		if end > len(glyfData) {
			end = len(glyfData)
		}
		glyphData := glyfData[start:end]
		if len(glyphData) < 10 {
			continue
		}

		// Check if composite (numberOfContours < 0)
		numContours := int16(binary.BigEndian.Uint16(glyphData[0:2]))
		if numContours >= 0 {
			continue
		}

		// Parse composite glyph
		pos := 10 // Skip header
		for pos+4 <= len(glyphData) {
			flags := binary.BigEndian.Uint16(glyphData[pos : pos+2])
			componentGID := binary.BigEndian.Uint16(glyphData[pos+2 : pos+4])
			glyphSet[componentGID] = true
			pos += 4

			// Determine argument size
			argSize := 2
			if flags&0x0001 != 0 { // ARG_1_AND_2_ARE_WORDS
				argSize = 4
			}
			pos += argSize

			// Handle transform
			if flags&0x0008 != 0 { // WE_HAVE_A_SCALE
				pos += 2
			} else if flags&0x0040 != 0 { // WE_HAVE_AN_X_AND_Y_SCALE
				pos += 4
			} else if flags&0x0080 != 0 { // WE_HAVE_A_TWO_BY_TWO
				pos += 8
			}

			if flags&0x0020 == 0 { // MORE_COMPONENTS
				break
			}
		}
	}
}

// getGlyphOffsets parses the loca table to get glyph offsets.
func getGlyphOffsets(locaData []byte, format int, glyfLength int) []int {
	var offsets []int
	if format == 0 {
		// Short format: offsets are uint16 * 2
		for i := 0; i+2 <= len(locaData); i += 2 {
			offsets = append(offsets, int(binary.BigEndian.Uint16(locaData[i:i+2]))*2)
		}
	} else {
		// Long format: offsets are uint32
		for i := 0; i+4 <= len(locaData); i += 4 {
			offsets = append(offsets, int(binary.BigEndian.Uint32(locaData[i:i+4])))
		}
	}
	return offsets
}

// subsetGlyfLoca creates subsetted glyf and loca tables.
func subsetGlyfLoca(glyfData, locaData []byte, glyphIDs []uint16, tables map[string][]byte) ([]byte, []byte, error) {
	headData, ok := tables["head"]
	if !ok || len(headData) < 52 {
		return nil, nil, fmt.Errorf("missing or invalid head table")
	}
	locaFormat := int(binary.BigEndian.Uint16(headData[50:52]))
	offsets := getGlyphOffsets(locaData, locaFormat, len(glyfData))

	var newGlyf bytes.Buffer
	newOffsets := make([]int, len(glyphIDs)+1)

	for i, gid := range glyphIDs {
		newOffsets[i] = newGlyf.Len()
		if int(gid) >= len(offsets)-1 {
			continue
		}
		start := offsets[gid]
		end := offsets[gid+1]
		if start >= end || start >= len(glyfData) {
			continue
		}
		if end > len(glyfData) {
			end = len(glyfData)
		}
		newGlyf.Write(glyfData[start:end])
	}
	newOffsets[len(glyphIDs)] = newGlyf.Len()

	// Pad glyf to 4-byte boundary
	for newGlyf.Len()%4 != 0 {
		newGlyf.WriteByte(0)
	}

	// Build loca table (use long format for simplicity)
	var newLoca bytes.Buffer
	for _, off := range newOffsets {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(off))
		newLoca.Write(buf[:])
	}

	return newGlyf.Bytes(), newLoca.Bytes(), nil
}

// subsetHmtx creates a subsetted hmtx table.
func subsetHmtx(hmtxData []byte, glyphIDs []uint16, tables map[string][]byte) []byte {
	// Get number of h-metrics from hhea
	hheaData, ok := tables["hhea"]
	if !ok || len(hheaData) < 36 {
		return hmtxData
	}
	numHMetrics := int(binary.BigEndian.Uint16(hheaData[34:36]))

	var newHmtx bytes.Buffer
	for _, gid := range glyphIDs {
		var advanceWidth uint16
		var lsb int16

		if int(gid) < numHMetrics {
			offset := int(gid) * 4
			if offset+4 <= len(hmtxData) {
				advanceWidth = binary.BigEndian.Uint16(hmtxData[offset : offset+2])
				lsb = int16(binary.BigEndian.Uint16(hmtxData[offset+2 : offset+4]))
			}
		} else {
			// Use last advance width, get LSB from extended table
			if numHMetrics > 0 {
				offset := (numHMetrics - 1) * 4
				if offset+2 <= len(hmtxData) {
					advanceWidth = binary.BigEndian.Uint16(hmtxData[offset : offset+2])
				}
			}
			lsbOffset := numHMetrics*4 + (int(gid)-numHMetrics)*2
			if lsbOffset+2 <= len(hmtxData) {
				lsb = int16(binary.BigEndian.Uint16(hmtxData[lsbOffset : lsbOffset+2]))
			}
		}

		var buf [4]byte
		binary.BigEndian.PutUint16(buf[0:2], advanceWidth)
		binary.BigEndian.PutUint16(buf[2:4], uint16(lsb))
		newHmtx.Write(buf[:])
	}

	return newHmtx.Bytes()
}

// buildFontFile assembles tables into a font file.
func buildFontFile(tables map[string][]byte) ([]byte, error) {
	// Get table names in sorted order
	names := make([]string, 0, len(tables))
	for name := range tables {
		names = append(names, name)
	}
	sort.Strings(names)

	numTables := len(names)

	// Calculate search range parameters
	entrySelector := 0
	searchRange := 1
	for searchRange*2 <= numTables {
		searchRange *= 2
		entrySelector++
	}
	searchRange *= 16
	rangeShift := numTables*16 - searchRange

	// Build header
	var buf bytes.Buffer

	// Offset table
	binary.Write(&buf, binary.BigEndian, uint32(0x00010000)) // version
	binary.Write(&buf, binary.BigEndian, uint16(numTables))
	binary.Write(&buf, binary.BigEndian, uint16(searchRange))
	binary.Write(&buf, binary.BigEndian, uint16(entrySelector))
	binary.Write(&buf, binary.BigEndian, uint16(rangeShift))

	// Calculate table offsets
	tableOffset := 12 + numTables*16
	offsets := make([]int, numTables)
	for i, name := range names {
		offsets[i] = tableOffset
		tableLen := len(tables[name])
		// Pad to 4-byte boundary
		tableOffset += (tableLen + 3) &^ 3
	}

	// Write table directory
	for i, name := range names {
		data := tables[name]

		// Tag
		tag := []byte(name)
		for len(tag) < 4 {
			tag = append(tag, ' ')
		}
		buf.Write(tag[:4])

		// Checksum
		checksum := calculateChecksum(data)
		binary.Write(&buf, binary.BigEndian, checksum)

		// Offset
		binary.Write(&buf, binary.BigEndian, uint32(offsets[i]))

		// Length
		binary.Write(&buf, binary.BigEndian, uint32(len(data)))
	}

	// Write table data
	for _, name := range names {
		data := tables[name]
		buf.Write(data)
		// Pad to 4-byte boundary
		for buf.Len()%4 != 0 {
			buf.WriteByte(0)
		}
	}

	return buf.Bytes(), nil
}

// calculateChecksum computes the checksum for a table.
func calculateChecksum(data []byte) uint32 {
	var sum uint32
	// Pad to multiple of 4
	padded := data
	if len(data)%4 != 0 {
		padded = make([]byte, (len(data)+3)&^3)
		copy(padded, data)
	}
	for i := 0; i < len(padded); i += 4 {
		sum += binary.BigEndian.Uint32(padded[i : i+4])
	}
	return sum
}

// Helper functions for font data access.

// getFontData extracts the raw font data from a FontRef.
func getFontData(ref *FontRef) ([]byte, error) {
	// First, try the stored source bytes
	if ref.Source != nil {
		return ref.Source, nil
	}

	// Second, try to reconstruct from loader
	if ref.Loader != nil {
		return reconstructFontFromLoader(ref.Loader)
	}

	return nil, fmt.Errorf("no font data available: provide source bytes or loader")
}

// reconstructFontFromLoader rebuilds a font file from individual tables.
func reconstructFontFromLoader(ld *opentype.Loader) ([]byte, error) {
	// Get list of tables
	tags := ld.Tables()
	if len(tags) == 0 {
		return nil, fmt.Errorf("font has no tables")
	}

	// Read all tables
	tables := make(map[string][]byte)
	for _, tag := range tags {
		data, err := ld.RawTable(tag)
		if err != nil {
			continue // Skip unreadable tables
		}
		tables[tag.String()] = data
	}

	// Build font file
	return buildFontFile(tables)
}

// getPostScriptName returns the PostScript name of the font.
func getPostScriptName(face *font.Face) string {
	if face == nil || face.Font == nil {
		return "Unknown"
	}

	// Try to get name from font tables
	// For now, use a placeholder - proper implementation would
	// read from the name table
	return "Font"
}

// getUnitsPerEm returns the font's units per em.
func getUnitsPerEm(face *font.Face) int {
	if face == nil || face.Font == nil {
		return 1000 // Default
	}
	return int(face.Font.Upem())
}

// getGlyphAdvance returns the advance width of a glyph.
func getGlyphAdvance(face *font.Face, glyphID uint16) int {
	if face == nil {
		return 0
	}

	// Get horizontal advance from face
	advance := face.HorizontalAdvance(font.GID(glyphID))
	return int(advance)
}

// hasCJKGlyphs checks if the font usage includes CJK characters.
func hasCJKGlyphs(usage *FontUsage) bool {
	for _, chars := range usage.GlyphIDs {
		for _, c := range chars {
			if isCJK(c) {
				return true
			}
		}
	}
	return false
}

// isCJK checks if a rune is a CJK character.
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x20000 && r <= 0x2A6DF) || // CJK Extension B
		(r >= 0x2A700 && r <= 0x2B73F) || // CJK Extension C
		(r >= 0x2B740 && r <= 0x2B81F) || // CJK Extension D
		(r >= 0x2B820 && r <= 0x2CEAF) || // CJK Extension E
		(r >= 0x2CEB0 && r <= 0x2EBEF) || // CJK Extension F
		(r >= 0x30000 && r <= 0x3134F) || // CJK Extension G
		(r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana
		(r >= 0xAC00 && r <= 0xD7AF) // Hangul Syllables
}

// copyToUnicode creates a copy of the glyph to Unicode mapping.
func copyToUnicode(m map[uint16][]rune) map[uint16][]rune {
	result := make(map[uint16][]rune, len(m))
	for k, v := range m {
		chars := make([]rune, len(v))
		copy(chars, v)
		result[k] = chars
	}
	return result
}
