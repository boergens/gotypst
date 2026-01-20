package font

import (
	"encoding/binary"
	"testing"
)

func TestGlyphSet(t *testing.T) {
	gs := NewGlyphSet()

	// Test empty set
	if gs.Len() != 0 {
		t.Errorf("expected empty set, got len=%d", gs.Len())
	}
	if gs.Contains(0) {
		t.Error("empty set should not contain glyph 0")
	}

	// Add glyphs
	gs.Add(5)
	gs.Add(10)
	gs.Add(3)
	gs.Add(5) // duplicate

	if gs.Len() != 3 {
		t.Errorf("expected 3 glyphs, got %d", gs.Len())
	}

	if !gs.Contains(5) {
		t.Error("set should contain glyph 5")
	}
	if !gs.Contains(10) {
		t.Error("set should contain glyph 10")
	}
	if gs.Contains(1) {
		t.Error("set should not contain glyph 1")
	}

	// Test sorted
	sorted := gs.Sorted()
	expected := []uint16{3, 5, 10}
	if len(sorted) != len(expected) {
		t.Fatalf("expected %d sorted glyphs, got %d", len(expected), len(sorted))
	}
	for i, v := range expected {
		if sorted[i] != v {
			t.Errorf("sorted[%d] = %d, expected %d", i, sorted[i], v)
		}
	}
}

func TestGlyphSetAddAll(t *testing.T) {
	gs := NewGlyphSet()
	gs.AddAll([]uint16{1, 2, 3, 4, 5})

	if gs.Len() != 5 {
		t.Errorf("expected 5 glyphs, got %d", gs.Len())
	}

	for i := uint16(1); i <= 5; i++ {
		if !gs.Contains(i) {
			t.Errorf("set should contain glyph %d", i)
		}
	}
}

func TestFontUsage(t *testing.T) {
	fu := NewFontUsage()

	// Create mock fonts (using nil for face since we don't need it)
	font1 := &Font{Info: FontInfo{Family: "TestFont1"}}
	font2 := &Font{Info: FontInfo{Family: "TestFont2"}}

	// Add glyphs
	fu.AddGlyph(font1, 1)
	fu.AddGlyph(font1, 2)
	fu.AddGlyph(font2, 10)
	fu.AddGlyph(font1, 3)

	// Check font1
	gs1 := fu.GetGlyphSet(font1)
	if gs1 == nil {
		t.Fatal("expected glyph set for font1")
	}
	if gs1.Len() != 3 {
		t.Errorf("expected 3 glyphs for font1, got %d", gs1.Len())
	}

	// Check font2
	gs2 := fu.GetGlyphSet(font2)
	if gs2 == nil {
		t.Fatal("expected glyph set for font2")
	}
	if gs2.Len() != 1 {
		t.Errorf("expected 1 glyph for font2, got %d", gs2.Len())
	}

	// Check fonts list
	fonts := fu.Fonts()
	if len(fonts) != 2 {
		t.Errorf("expected 2 fonts, got %d", len(fonts))
	}
}

func TestParseFontDirectory(t *testing.T) {
	// Create a minimal valid TrueType font header
	data := createMinimalTTFHeader()

	tables, err := parseFontDirectory(data)
	if err != nil {
		t.Fatalf("parseFontDirectory failed: %v", err)
	}

	// We created a minimal font with just a 'test' table
	if len(tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(tables))
	}

	if _, ok := tables["test"]; !ok {
		t.Error("expected 'test' table")
	}
}

func TestParseFontDirectoryInvalid(t *testing.T) {
	// Test too short
	_, err := parseFontDirectory([]byte{0, 1, 0})
	if err == nil {
		t.Error("expected error for short data")
	}

	// Test invalid sfnt version
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:4], 0x12345678) // Invalid version
	_, err = parseFontDirectory(data)
	if err == nil {
		t.Error("expected error for invalid sfnt version")
	}
}

func TestCalculateChecksum(t *testing.T) {
	// Test checksum calculation
	data := []byte{0x00, 0x01, 0x00, 0x00} // sfnt version
	checksum := calculateChecksum(data)
	expected := uint32(0x00010000)
	if checksum != expected {
		t.Errorf("checksum = %08x, expected %08x", checksum, expected)
	}

	// Test with padding needed
	data2 := []byte{0x01, 0x02, 0x03}
	checksum2 := calculateChecksum(data2)
	// Should pad to 4 bytes and calculate
	expectedPadded := uint32(0x01020300)
	if checksum2 != expectedPadded {
		t.Errorf("checksum = %08x, expected %08x", checksum2, expectedPadded)
	}
}

func TestSubsettedFontGlyphMapping(t *testing.T) {
	// Test that glyph mapping is created correctly
	gs := NewGlyphSet()
	gs.Add(0)  // .notdef
	gs.Add(5)
	gs.Add(10)
	gs.Add(20)

	sorted := gs.Sorted()
	glyphMapping := make(map[uint16]uint16)
	cidToGID := make([]uint16, len(sorted))

	for newID, oldID := range sorted {
		glyphMapping[oldID] = uint16(newID)
		cidToGID[newID] = oldID
	}

	// Check mapping
	if glyphMapping[0] != 0 {
		t.Errorf("glyph 0 should map to 0, got %d", glyphMapping[0])
	}
	if glyphMapping[5] != 1 {
		t.Errorf("glyph 5 should map to 1, got %d", glyphMapping[5])
	}
	if glyphMapping[10] != 2 {
		t.Errorf("glyph 10 should map to 2, got %d", glyphMapping[10])
	}
	if glyphMapping[20] != 3 {
		t.Errorf("glyph 20 should map to 3, got %d", glyphMapping[20])
	}

	// Check reverse mapping
	if cidToGID[0] != 0 {
		t.Errorf("CID 0 should map to glyph 0, got %d", cidToGID[0])
	}
	if cidToGID[1] != 5 {
		t.Errorf("CID 1 should map to glyph 5, got %d", cidToGID[1])
	}
}

// Helper function to create a minimal valid TrueType font header
func createMinimalTTFHeader() []byte {
	// Create minimal TTF with just offset table and one table record
	numTables := uint16(1)
	headerSize := 12 + int(numTables)*16
	tableDataOffset := uint32(headerSize)
	tableData := []byte("TEST") // 4 bytes of table data

	data := make([]byte, headerSize+len(tableData))

	// Offset table
	binary.BigEndian.PutUint32(data[0:4], 0x00010000) // sfnt version
	binary.BigEndian.PutUint16(data[4:6], numTables)
	binary.BigEndian.PutUint16(data[6:8], 16)         // searchRange
	binary.BigEndian.PutUint16(data[8:10], 0)         // entrySelector
	binary.BigEndian.PutUint16(data[10:12], 0)        // rangeShift

	// Table record for "test"
	copy(data[12:16], "test")                                                  // tag
	binary.BigEndian.PutUint32(data[16:20], calculateChecksum(tableData))      // checksum
	binary.BigEndian.PutUint32(data[20:24], tableDataOffset)                   // offset
	binary.BigEndian.PutUint32(data[24:28], uint32(len(tableData)))            // length

	// Table data
	copy(data[headerSize:], tableData)

	return data
}

func TestBuildFont(t *testing.T) {
	tables := []struct {
		tag  string
		data []byte
	}{
		{"head", make([]byte, 54)}, // head table is 54 bytes
		{"maxp", make([]byte, 6)},  // minimal maxp
	}

	result, err := buildFont(tables)
	if err != nil {
		t.Fatalf("buildFont failed: %v", err)
	}

	// Verify it's a valid TTF
	if len(result) < 12 {
		t.Fatal("result too short")
	}

	sfntVersion := binary.BigEndian.Uint32(result[0:4])
	if sfntVersion != 0x00010000 {
		t.Errorf("invalid sfnt version: %08x", sfntVersion)
	}

	numTables := binary.BigEndian.Uint16(result[4:6])
	if numTables != 2 {
		t.Errorf("expected 2 tables, got %d", numTables)
	}
}

func TestCanSubset(t *testing.T) {
	// Font without raw data
	f1 := &Font{}
	if f1.CanSubset() {
		t.Error("font without raw data should not be subsettable")
	}

	// Font with raw data
	f2 := &Font{RawData: []byte{0, 1, 2, 3}}
	if !f2.CanSubset() {
		t.Error("font with raw data should be subsettable")
	}
}

func TestNewSubsetter(t *testing.T) {
	// Font without raw data should return nil subsetter
	f1 := &Font{}
	if f1.NewSubsetter() != nil {
		t.Error("expected nil subsetter for font without raw data")
	}

	// Font with raw data should return valid subsetter
	f2 := &Font{RawData: []byte{0, 1, 2, 3}}
	s := f2.NewSubsetter()
	if s == nil {
		t.Error("expected non-nil subsetter for font with raw data")
	}
}
