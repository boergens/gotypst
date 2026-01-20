package pdf

import (
	"strings"
	"testing"
)

func TestType0Font_ToUnicodeCMap(t *testing.T) {
	font := NewType0Font(nil)
	font.AddGlyph(1, 'A', 600)
	font.AddGlyph(2, 'B', 600)
	font.AddGlyph(3, 'C', 600)
	font.AddGlyph(100, '中', 1000) // CJK character

	cmap := font.buildToUnicodeCMap()
	cmapStr := string(cmap)

	// Check for CMap header
	if !strings.Contains(cmapStr, "begincmap") {
		t.Error("CMap should contain begincmap")
	}

	// Check for code space range
	if !strings.Contains(cmapStr, "<0000> <FFFF>") {
		t.Error("CMap should define code space range")
	}

	// Check for character mappings
	if !strings.Contains(cmapStr, "<0001> <0041>") {
		t.Error("CMap should map glyph 1 to 'A' (U+0041)")
	}
	if !strings.Contains(cmapStr, "<0064> <4E2D>") {
		t.Error("CMap should map glyph 100 to '中' (U+4E2D)")
	}
}

func TestType0Font_WidthArray(t *testing.T) {
	font := NewType0Font(nil)
	font.AddGlyph(1, 'A', 600)
	font.AddGlyph(2, 'B', 700)
	font.AddGlyph(10, 'K', 650)

	arr := font.buildWidthArray()

	// Width array should have entries for each glyph
	// Format: [cid [width] cid [width] ...]
	if len(arr) != 6 {
		t.Errorf("Width array should have 6 elements (3 glyphs * 2), got %d", len(arr))
	}
}

func TestFontManager_RegisterGlyph(t *testing.T) {
	manager := NewFontManager()

	// Register first glyph
	name1 := manager.RegisterGlyph(nil, 1, 'A', 600)
	if name1 != "F1" {
		t.Errorf("First font should be named F1, got %s", name1)
	}

	// Register another glyph with same face
	name2 := manager.RegisterGlyph(nil, 2, 'B', 600)
	if name2 != "F1" {
		t.Errorf("Same face should return same font name F1, got %s", name2)
	}
}

func TestFontManager_FontName(t *testing.T) {
	manager := NewFontManager()
	manager.RegisterGlyph(nil, 1, 'A', 600)

	// Test FontName lookup
	name := manager.FontName(nil)
	if name != "/F1" {
		t.Errorf("FontName should return /F1, got %s", name)
	}
}

func TestType0Font_CIDToGIDMap(t *testing.T) {
	font := NewType0Font(nil)
	font.AddGlyph(1, 'A', 600)
	font.AddGlyph(2, 'B', 600)
	font.AddGlyph(255, 'X', 600)

	data := font.buildCIDToGIDMap()

	// Should have 2 bytes per entry, for entries 0..255
	expectedLen := 256 * 2
	if len(data) != expectedLen {
		t.Errorf("CIDToGIDMap should have %d bytes, got %d", expectedLen, len(data))
	}

	// Check identity mapping for glyph 1 (bytes at offset 2, 3)
	if data[2] != 0 || data[3] != 1 {
		t.Errorf("GID 1 should map to CID 1, got %02X%02X", data[2], data[3])
	}
}

func TestTextPositionHex(t *testing.T) {
	cs := NewContentStream()
	items := []TextPositionItem{
		TextPositionHex([]byte{0x00, 0x41}), // Glyph ID 65 ('A')
		TextPositionHex([]byte{0x00, 0x42}), // Glyph ID 66 ('B')
	}
	cs.ShowTextWithPositioning(items)

	result := cs.String()
	if !strings.Contains(result, "<0041>") {
		t.Error("Hex string should contain <0041>")
	}
	if !strings.Contains(result, "<0042>") {
		t.Error("Hex string should contain <0042>")
	}
	if !strings.Contains(result, "TJ") {
		t.Error("Output should contain TJ operator")
	}
}
