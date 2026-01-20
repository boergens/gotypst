// Package font provides font subsetting capabilities.
//
// Font subsetting reduces font file size by including only the glyphs
// that are actually used in a document. This is essential for PDF
// embedding where full font files would be wasteful.
package font

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
)

// GlyphSet tracks which glyphs are used from a font.
type GlyphSet struct {
	glyphs map[uint16]struct{}
}

// NewGlyphSet creates a new empty glyph set.
func NewGlyphSet() *GlyphSet {
	return &GlyphSet{
		glyphs: make(map[uint16]struct{}),
	}
}

// Add adds a glyph ID to the set.
func (g *GlyphSet) Add(glyphID uint16) {
	g.glyphs[glyphID] = struct{}{}
}

// AddAll adds multiple glyph IDs to the set.
func (g *GlyphSet) AddAll(glyphIDs []uint16) {
	for _, id := range glyphIDs {
		g.glyphs[id] = struct{}{}
	}
}

// Contains returns true if the glyph ID is in the set.
func (g *GlyphSet) Contains(glyphID uint16) bool {
	_, ok := g.glyphs[glyphID]
	return ok
}

// Len returns the number of glyphs in the set.
func (g *GlyphSet) Len() int {
	return len(g.glyphs)
}

// Sorted returns the glyph IDs in sorted order.
func (g *GlyphSet) Sorted() []uint16 {
	ids := make([]uint16, 0, len(g.glyphs))
	for id := range g.glyphs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// FontUsage tracks glyph usage across multiple fonts in a document.
type FontUsage struct {
	usage map[*Font]*GlyphSet
}

// NewFontUsage creates a new font usage tracker.
func NewFontUsage() *FontUsage {
	return &FontUsage{
		usage: make(map[*Font]*GlyphSet),
	}
}

// AddGlyph records a glyph being used from a font.
func (f *FontUsage) AddGlyph(font *Font, glyphID uint16) {
	gs, ok := f.usage[font]
	if !ok {
		gs = NewGlyphSet()
		f.usage[font] = gs
	}
	gs.Add(glyphID)
}

// GetGlyphSet returns the glyph set for a font.
func (f *FontUsage) GetGlyphSet(font *Font) *GlyphSet {
	return f.usage[font]
}

// Fonts returns all fonts with recorded usage.
func (f *FontUsage) Fonts() []*Font {
	fonts := make([]*Font, 0, len(f.usage))
	for font := range f.usage {
		fonts = append(fonts, font)
	}
	return fonts
}

// SubsettedFont contains the subsetted font data and mapping information.
type SubsettedFont struct {
	// Data is the subsetted TrueType font data.
	Data []byte

	// GlyphMapping maps original glyph IDs to new glyph IDs in the subset.
	// The new IDs are contiguous starting from 0.
	GlyphMapping map[uint16]uint16

	// CIDToGID maps CIDs (new glyph indices) to original glyph IDs.
	CIDToGID []uint16

	// OriginalFont is a reference to the original font.
	OriginalFont *Font
}

// Subsetter handles font subsetting operations.
type Subsetter struct {
	font *Font
	data []byte
}

// NewSubsetter creates a new subsetter for a font.
// The rawData should be the original TTF/OTF file bytes.
func NewSubsetter(font *Font, rawData []byte) *Subsetter {
	return &Subsetter{
		font: font,
		data: rawData,
	}
}

// Subset creates a subsetted font containing only the specified glyphs.
// The .notdef glyph (ID 0) is always included.
func (s *Subsetter) Subset(glyphSet *GlyphSet) (*SubsettedFont, error) {
	if len(s.data) < 12 {
		return nil, errors.New("font data too short")
	}

	// Parse the font directory
	tables, err := parseFontDirectory(s.data)
	if err != nil {
		return nil, fmt.Errorf("parse font directory: %w", err)
	}

	// Build the list of glyphs to include (always include .notdef)
	glyphIDs := glyphSet.Sorted()
	if len(glyphIDs) == 0 || glyphIDs[0] != 0 {
		glyphIDs = append([]uint16{0}, glyphIDs...)
	}

	// Create glyph mapping (old ID -> new ID)
	glyphMapping := make(map[uint16]uint16)
	cidToGID := make([]uint16, len(glyphIDs))
	for newID, oldID := range glyphIDs {
		glyphMapping[oldID] = uint16(newID)
		cidToGID[newID] = oldID
	}

	// Create the subsetted font
	subsetData, err := s.createSubset(tables, glyphIDs, glyphMapping)
	if err != nil {
		return nil, fmt.Errorf("create subset: %w", err)
	}

	return &SubsettedFont{
		Data:         subsetData,
		GlyphMapping: glyphMapping,
		CIDToGID:     cidToGID,
		OriginalFont: s.font,
	}, nil
}

// tableRecord represents a TrueType table directory entry.
type tableRecord struct {
	tag      string
	checksum uint32
	offset   uint32
	length   uint32
}

// parseFontDirectory parses the TrueType font directory.
func parseFontDirectory(data []byte) (map[string]tableRecord, error) {
	r := bytes.NewReader(data)

	// Read offset table
	var sfntVersion uint32
	if err := binary.Read(r, binary.BigEndian, &sfntVersion); err != nil {
		return nil, err
	}

	// Check for valid sfnt version
	if sfntVersion != 0x00010000 && sfntVersion != 0x4F54544F { // 1.0 or 'OTTO'
		return nil, fmt.Errorf("unsupported sfnt version: %08x", sfntVersion)
	}

	var numTables uint16
	if err := binary.Read(r, binary.BigEndian, &numTables); err != nil {
		return nil, err
	}

	// Skip searchRange, entrySelector, rangeShift
	r.Seek(6, io.SeekCurrent)

	tables := make(map[string]tableRecord, numTables)
	for i := 0; i < int(numTables); i++ {
		var tagBytes [4]byte
		if _, err := r.Read(tagBytes[:]); err != nil {
			return nil, err
		}

		var checksum, offset, length uint32
		if err := binary.Read(r, binary.BigEndian, &checksum); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.BigEndian, &offset); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}

		tag := string(tagBytes[:])
		tables[tag] = tableRecord{
			tag:      tag,
			checksum: checksum,
			offset:   offset,
			length:   length,
		}
	}

	return tables, nil
}

// createSubset creates the subsetted font data.
func (s *Subsetter) createSubset(tables map[string]tableRecord, glyphIDs []uint16, glyphMapping map[uint16]uint16) ([]byte, error) {
	// Get required tables
	head, ok := tables["head"]
	if !ok {
		return nil, errors.New("missing head table")
	}
	maxp, ok := tables["maxp"]
	if !ok {
		return nil, errors.New("missing maxp table")
	}
	loca, ok := tables["loca"]
	if !ok {
		return nil, errors.New("missing loca table")
	}
	glyf, ok := tables["glyf"]
	if !ok {
		return nil, errors.New("missing glyf table")
	}

	// Read head table to get indexToLocFormat
	headData := s.data[head.offset : head.offset+head.length]
	indexToLocFormat := int16(binary.BigEndian.Uint16(headData[50:52]))

	// Read maxp to get numGlyphs
	maxpData := s.data[maxp.offset : maxp.offset+maxp.length]
	numGlyphs := binary.BigEndian.Uint16(maxpData[4:6])

	// Read loca table
	locaData := s.data[loca.offset : loca.offset+loca.length]
	glyfData := s.data[glyf.offset : glyf.offset+glyf.length]

	// Get glyph offsets
	offsets := make([]uint32, numGlyphs+1)
	if indexToLocFormat == 0 {
		// Short format (16-bit offsets, multiplied by 2)
		for i := 0; i <= int(numGlyphs); i++ {
			offsets[i] = uint32(binary.BigEndian.Uint16(locaData[i*2:i*2+2])) * 2
		}
	} else {
		// Long format (32-bit offsets)
		for i := 0; i <= int(numGlyphs); i++ {
			offsets[i] = binary.BigEndian.Uint32(locaData[i*4 : i*4+4])
		}
	}

	// Build new glyf and loca tables
	var newGlyf bytes.Buffer
	newLoca := make([]uint32, len(glyphIDs)+1)

	for newID, oldID := range glyphIDs {
		newLoca[newID] = uint32(newGlyf.Len())

		if int(oldID) >= int(numGlyphs) {
			// Invalid glyph ID, write empty glyph
			continue
		}

		glyphStart := offsets[oldID]
		glyphEnd := offsets[oldID+1]

		if glyphStart >= glyphEnd {
			// Empty glyph
			continue
		}

		if glyphEnd > uint32(len(glyfData)) {
			continue
		}

		glyphBytes := glyfData[glyphStart:glyphEnd]

		// Check if this is a composite glyph and update component references
		if len(glyphBytes) >= 10 {
			numberOfContours := int16(binary.BigEndian.Uint16(glyphBytes[0:2]))
			if numberOfContours < 0 {
				// Composite glyph - need to update component glyph IDs
				glyphBytes = s.updateCompositeGlyph(glyphBytes, glyphMapping)
			}
		}

		newGlyf.Write(glyphBytes)

		// Pad to 4-byte boundary
		for newGlyf.Len()%4 != 0 {
			newGlyf.WriteByte(0)
		}
	}
	newLoca[len(glyphIDs)] = uint32(newGlyf.Len())

	// Build new loca table data
	var newLocaData bytes.Buffer
	// Use long loca format for simplicity
	for _, offset := range newLoca {
		binary.Write(&newLocaData, binary.BigEndian, offset)
	}

	// Build new maxp table
	newMaxp := make([]byte, len(maxpData))
	copy(newMaxp, maxpData)
	binary.BigEndian.PutUint16(newMaxp[4:6], uint16(len(glyphIDs)))

	// Build new head table (update indexToLocFormat to 1 for long loca)
	newHead := make([]byte, len(headData))
	copy(newHead, headData)
	binary.BigEndian.PutUint16(newHead[50:52], 1) // Long loca format

	// Subset hhea and hmtx tables
	newHhea, newHmtx, err := s.subsetHorizontalMetrics(tables, glyphIDs, numGlyphs)
	if err != nil {
		return nil, err
	}

	// Build the new font
	tablesToInclude := []struct {
		tag  string
		data []byte
	}{
		{"head", newHead},
		{"hhea", newHhea},
		{"maxp", newMaxp},
		{"loca", newLocaData.Bytes()},
		{"glyf", newGlyf.Bytes()},
		{"hmtx", newHmtx},
	}

	// Copy other required tables
	for _, tag := range []string{"cmap", "name", "OS/2", "post", "cvt ", "fpgm", "prep"} {
		if t, ok := tables[tag]; ok {
			tablesToInclude = append(tablesToInclude, struct {
				tag  string
				data []byte
			}{tag, s.data[t.offset : t.offset+t.length]})
		}
	}

	return buildFont(tablesToInclude)
}

// updateCompositeGlyph updates glyph references in a composite glyph.
func (s *Subsetter) updateCompositeGlyph(data []byte, glyphMapping map[uint16]uint16) []byte {
	result := make([]byte, len(data))
	copy(result, data)

	// Skip glyph header (10 bytes: numberOfContours, xMin, yMin, xMax, yMax)
	offset := 10

	for offset+4 <= len(result) {
		flags := binary.BigEndian.Uint16(result[offset:])
		glyphIndex := binary.BigEndian.Uint16(result[offset+2:])

		// Update glyph index if it's in our mapping
		if newID, ok := glyphMapping[glyphIndex]; ok {
			binary.BigEndian.PutUint16(result[offset+2:], newID)
		}

		offset += 4

		// Skip arguments based on flags
		if flags&0x0001 != 0 { // ARG_1_AND_2_ARE_WORDS
			offset += 4
		} else {
			offset += 2
		}

		if flags&0x0008 != 0 { // WE_HAVE_A_SCALE
			offset += 2
		} else if flags&0x0040 != 0 { // WE_HAVE_AN_X_AND_Y_SCALE
			offset += 4
		} else if flags&0x0080 != 0 { // WE_HAVE_A_TWO_BY_TWO
			offset += 8
		}

		if flags&0x0020 == 0 { // MORE_COMPONENTS
			break
		}
	}

	return result
}

// subsetHorizontalMetrics subsets the hhea and hmtx tables.
func (s *Subsetter) subsetHorizontalMetrics(tables map[string]tableRecord, glyphIDs []uint16, numGlyphs uint16) ([]byte, []byte, error) {
	hhea, ok := tables["hhea"]
	if !ok {
		return nil, nil, errors.New("missing hhea table")
	}
	hmtx, ok := tables["hmtx"]
	if !ok {
		return nil, nil, errors.New("missing hmtx table")
	}

	hheaData := s.data[hhea.offset : hhea.offset+hhea.length]
	hmtxData := s.data[hmtx.offset : hmtx.offset+hmtx.length]

	// Get numberOfHMetrics from hhea
	numberOfHMetrics := binary.BigEndian.Uint16(hheaData[34:36])

	// Build new hmtx table
	var newHmtx bytes.Buffer
	for _, glyphID := range glyphIDs {
		var advanceWidth uint16
		var lsb int16

		if glyphID < numberOfHMetrics {
			// Full entry: advanceWidth + lsb
			offset := int(glyphID) * 4
			if offset+4 <= len(hmtxData) {
				advanceWidth = binary.BigEndian.Uint16(hmtxData[offset:])
				lsb = int16(binary.BigEndian.Uint16(hmtxData[offset+2:]))
			}
		} else {
			// Only lsb, use last advanceWidth
			if numberOfHMetrics > 0 {
				offset := int(numberOfHMetrics-1) * 4
				if offset+2 <= len(hmtxData) {
					advanceWidth = binary.BigEndian.Uint16(hmtxData[offset:])
				}
			}
			// lsb entries follow the full metrics
			lsbOffset := int(numberOfHMetrics)*4 + int(glyphID-numberOfHMetrics)*2
			if lsbOffset+2 <= len(hmtxData) {
				lsb = int16(binary.BigEndian.Uint16(hmtxData[lsbOffset:]))
			}
		}

		binary.Write(&newHmtx, binary.BigEndian, advanceWidth)
		binary.Write(&newHmtx, binary.BigEndian, lsb)
	}

	// Update hhea with new numberOfHMetrics
	newHhea := make([]byte, len(hheaData))
	copy(newHhea, hheaData)
	binary.BigEndian.PutUint16(newHhea[34:36], uint16(len(glyphIDs)))

	return newHhea, newHmtx.Bytes(), nil
}

// buildFont constructs a TrueType font file from tables.
func buildFont(tables []struct {
	tag  string
	data []byte
}) ([]byte, error) {
	numTables := len(tables)

	// Calculate table directory size
	headerSize := 12 + numTables*16

	// Calculate search parameters
	searchRange := 1
	entrySelector := 0
	for searchRange*2 <= numTables {
		searchRange *= 2
		entrySelector++
	}
	searchRange *= 16
	rangeShift := numTables*16 - searchRange

	var buf bytes.Buffer

	// Write offset table
	binary.Write(&buf, binary.BigEndian, uint32(0x00010000)) // sfnt version
	binary.Write(&buf, binary.BigEndian, uint16(numTables))
	binary.Write(&buf, binary.BigEndian, uint16(searchRange))
	binary.Write(&buf, binary.BigEndian, uint16(entrySelector))
	binary.Write(&buf, binary.BigEndian, uint16(rangeShift))

	// Calculate table offsets
	offset := uint32(headerSize)
	tableOffsets := make([]uint32, numTables)
	for i, t := range tables {
		tableOffsets[i] = offset
		padded := (len(t.data) + 3) &^ 3 // Pad to 4-byte boundary
		offset += uint32(padded)
	}

	// Write table directory
	for i, t := range tables {
		buf.WriteString(t.tag[:4])
		binary.Write(&buf, binary.BigEndian, calculateChecksum(t.data))
		binary.Write(&buf, binary.BigEndian, tableOffsets[i])
		binary.Write(&buf, binary.BigEndian, uint32(len(t.data)))
	}

	// Write table data
	for _, t := range tables {
		buf.Write(t.data)
		// Pad to 4-byte boundary
		for buf.Len()%4 != 0 {
			buf.WriteByte(0)
		}
	}

	return buf.Bytes(), nil
}

// calculateChecksum calculates the TrueType table checksum.
func calculateChecksum(data []byte) uint32 {
	var sum uint32
	// Pad to 4-byte boundary for checksum calculation
	padded := data
	if len(data)%4 != 0 {
		padded = make([]byte, (len(data)+3)&^3)
		copy(padded, data)
	}

	for i := 0; i < len(padded); i += 4 {
		sum += binary.BigEndian.Uint32(padded[i:])
	}
	return sum
}
