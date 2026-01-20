// Package pdf provides font embedding for PDF output.
package pdf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/boergens/gotypst/font"
)

// FontManager handles font subsetting and embedding for PDF output.
type FontManager struct {
	// fonts maps font pointers to their PDF font entries.
	fonts map[*font.Font]*PDFFont
	// fontOrder maintains consistent ordering for font entries.
	fontOrder []*font.Font
}

// NewFontManager creates a new font manager.
func NewFontManager() *FontManager {
	return &FontManager{
		fonts: make(map[*font.Font]*PDFFont),
	}
}

// PDFFont represents a font entry in the PDF.
type PDFFont struct {
	// Font is the original font.
	Font *font.Font
	// Name is the PDF font name (e.g., "F1", "F2").
	Name string
	// Glyphs tracks which glyphs are used.
	Glyphs *font.GlyphSet
	// SubsetPrefix is the 6-character subset tag (e.g., "ABCDEF").
	SubsetPrefix string
	// Ref is the PDF object reference for this font.
	Ref Ref
	// DescendantRef is the reference to the CIDFont descendant.
	DescendantRef Ref
	// DescriptorRef is the reference to the FontDescriptor.
	DescriptorRef Ref
	// FontFileRef is the reference to the embedded font file.
	FontFileRef Ref
	// ToUnicodeRef is the reference to the ToUnicode CMap.
	ToUnicodeRef Ref
	// CIDToGIDMapRef is the reference to the CIDToGIDMap stream.
	CIDToGIDMapRef Ref
	// Subset is the subsetted font data.
	Subset *font.SubsettedFont
}

// GetOrCreateFont returns the PDF font for a given font, creating one if needed.
func (m *FontManager) GetOrCreateFont(f *font.Font) *PDFFont {
	if pdfFont, ok := m.fonts[f]; ok {
		return pdfFont
	}

	pdfFont := &PDFFont{
		Font:         f,
		Name:         fmt.Sprintf("F%d", len(m.fonts)+1),
		Glyphs:       font.NewGlyphSet(),
		SubsetPrefix: generateSubsetPrefix(len(m.fonts)),
	}
	m.fonts[f] = pdfFont
	m.fontOrder = append(m.fontOrder, f)
	return pdfFont
}

// AddGlyph records a glyph being used from a font.
func (m *FontManager) AddGlyph(f *font.Font, glyphID uint16) {
	pdfFont := m.GetOrCreateFont(f)
	pdfFont.Glyphs.Add(glyphID)
}

// Fonts returns all fonts in order.
func (m *FontManager) Fonts() []*PDFFont {
	result := make([]*PDFFont, len(m.fontOrder))
	for i, f := range m.fontOrder {
		result[i] = m.fonts[f]
	}
	return result
}

// HasFonts returns true if any fonts have been recorded.
func (m *FontManager) HasFonts() bool {
	return len(m.fonts) > 0
}

// SubsetFonts creates subsetted versions of all recorded fonts.
func (m *FontManager) SubsetFonts() error {
	for _, pdfFont := range m.fonts {
		if !pdfFont.Font.CanSubset() {
			continue
		}

		subsetter := pdfFont.Font.NewSubsetter()
		if subsetter == nil {
			continue
		}

		subset, err := subsetter.Subset(pdfFont.Glyphs)
		if err != nil {
			return fmt.Errorf("subset font %s: %w", pdfFont.Font.Family(), err)
		}
		pdfFont.Subset = subset
	}
	return nil
}

// generateSubsetPrefix generates a unique 6-character prefix for font subsetting.
// The prefix consists of uppercase letters A-Z.
func generateSubsetPrefix(index int) string {
	prefix := make([]byte, 6)
	for i := 5; i >= 0; i-- {
		prefix[i] = 'A' + byte(index%26)
		index /= 26
	}
	return string(prefix)
}

// BuildFontResources creates the Font resource dictionary for a page.
func (m *FontManager) BuildFontResources() Dict {
	fonts := make(Dict)
	for _, pdfFont := range m.fonts {
		fonts[Name(pdfFont.Name)] = pdfFont.Ref
	}
	return fonts
}

// WriteFontObjects writes all font objects to the PDF writer.
func (m *FontManager) WriteFontObjects(w *Writer) error {
	for _, pdfFont := range m.fonts {
		if err := m.writeFontObject(w, pdfFont); err != nil {
			return err
		}
	}
	return nil
}

// writeFontObject writes a single font's PDF objects.
func (m *FontManager) writeFontObject(w *Writer, pdfFont *PDFFont) error {
	if pdfFont.Subset == nil {
		// No subset available, use a fallback Type1 font
		return m.writeType1Font(w, pdfFont)
	}

	// Allocate all refs
	pdfFont.Ref = w.allocRef()
	pdfFont.DescendantRef = w.allocRef()
	pdfFont.DescriptorRef = w.allocRef()
	pdfFont.FontFileRef = w.allocRef()
	pdfFont.ToUnicodeRef = w.allocRef()

	// Write font file stream
	fontFileStream := Stream{
		Dict: Dict{
			Name("Length1"): Int(len(pdfFont.Subset.Data)),
		},
		Data: pdfFont.Subset.Data,
	}
	if err := fontFileStream.Compress(); err != nil {
		return err
	}
	w.addObjectWithRef(pdfFont.FontFileRef, fontFileStream)

	// Write FontDescriptor
	fontDescriptor := m.buildFontDescriptor(pdfFont)
	w.addObjectWithRef(pdfFont.DescriptorRef, fontDescriptor)

	// Build and write ToUnicode CMap
	toUnicodeData := m.buildToUnicodeCMap(pdfFont)
	toUnicodeStream := Stream{
		Dict: make(Dict),
		Data: toUnicodeData,
	}
	if err := toUnicodeStream.Compress(); err != nil {
		return err
	}
	w.addObjectWithRef(pdfFont.ToUnicodeRef, toUnicodeStream)

	// Build CIDFont (descendant)
	cidFont := m.buildCIDFont(pdfFont)
	w.addObjectWithRef(pdfFont.DescendantRef, cidFont)

	// Build Type0 composite font
	type0Font := m.buildType0Font(pdfFont)
	w.addObjectWithRef(pdfFont.Ref, type0Font)

	return nil
}

// writeType1Font writes a basic Type1 font fallback.
func (m *FontManager) writeType1Font(w *Writer, pdfFont *PDFFont) error {
	pdfFont.Ref = w.allocRef()

	fontDict := Dict{
		Name("Type"):     Name("Font"),
		Name("Subtype"):  Name("Type1"),
		Name("BaseFont"): Name("Helvetica"),
		Name("Encoding"): Name("WinAnsiEncoding"),
	}
	w.addObjectWithRef(pdfFont.Ref, fontDict)
	return nil
}

// buildFontDescriptor builds the FontDescriptor dictionary.
func (m *FontManager) buildFontDescriptor(pdfFont *PDFFont) Dict {
	// Use subset prefix in PostScript name
	psName := pdfFont.SubsetPrefix + "+" + sanitizePostScriptName(pdfFont.Font.Info.PostScriptName)
	if psName == pdfFont.SubsetPrefix+"+" {
		psName = pdfFont.SubsetPrefix + "+" + sanitizePostScriptName(pdfFont.Font.Family())
	}

	// Flags: Symbolic (bit 3) = 4, Serif (bit 2) = 2, etc.
	flags := 4 // Symbolic - use font's built-in encoding

	return Dict{
		Name("Type"):        Name("FontDescriptor"),
		Name("FontName"):    Name(psName),
		Name("Flags"):       Int(flags),
		Name("FontBBox"):    Array{Int(-500), Int(-300), Int(1500), Int(1000)}, // Approximate
		Name("ItalicAngle"): Int(0),
		Name("Ascent"):      Int(800),
		Name("Descent"):     Int(-200),
		Name("CapHeight"):   Int(700),
		Name("StemV"):       Int(80),
		Name("FontFile2"):   pdfFont.FontFileRef,
	}
}

// buildCIDFont builds the CIDFont (descendant) dictionary.
func (m *FontManager) buildCIDFont(pdfFont *PDFFont) Dict {
	psName := pdfFont.SubsetPrefix + "+" + sanitizePostScriptName(pdfFont.Font.Info.PostScriptName)
	if psName == pdfFont.SubsetPrefix+"+" {
		psName = pdfFont.SubsetPrefix + "+" + sanitizePostScriptName(pdfFont.Font.Family())
	}

	// Build W array (widths)
	widths := m.buildWidthsArray(pdfFont)

	cidFont := Dict{
		Name("Type"):           Name("Font"),
		Name("Subtype"):        Name("CIDFontType2"),
		Name("BaseFont"):       Name(psName),
		Name("CIDToGIDMap"):    Name("Identity"),
		Name("FontDescriptor"): pdfFont.DescriptorRef,
		Name("CIDSystemInfo"): Dict{
			Name("Registry"):   String("Adobe"),
			Name("Ordering"):   String("Identity"),
			Name("Supplement"): Int(0),
		},
	}

	if len(widths) > 0 {
		cidFont[Name("W")] = widths
	}

	return cidFont
}

// buildWidthsArray builds the W (widths) array for a CIDFont.
func (m *FontManager) buildWidthsArray(pdfFont *PDFFont) Array {
	if pdfFont.Subset == nil {
		return nil
	}

	// For now, use a default width of 500 for all glyphs
	// In a complete implementation, we'd read widths from the hmtx table
	glyphIDs := pdfFont.Glyphs.Sorted()
	if len(glyphIDs) == 0 {
		return nil
	}

	// Build width array in format: [startCID [width1 width2 ...]]
	widths := make(Array, 0)

	// Simple format: list each glyph's width
	for _, gid := range glyphIDs {
		newGID := pdfFont.Subset.GlyphMapping[gid]
		widths = append(widths, Int(newGID), Array{Int(500)})
	}

	return widths
}

// buildType0Font builds the Type0 composite font dictionary.
func (m *FontManager) buildType0Font(pdfFont *PDFFont) Dict {
	psName := pdfFont.SubsetPrefix + "+" + sanitizePostScriptName(pdfFont.Font.Info.PostScriptName)
	if psName == pdfFont.SubsetPrefix+"+" {
		psName = pdfFont.SubsetPrefix + "+" + sanitizePostScriptName(pdfFont.Font.Family())
	}

	return Dict{
		Name("Type"):            Name("Font"),
		Name("Subtype"):         Name("Type0"),
		Name("BaseFont"):        Name(psName),
		Name("Encoding"):        Name("Identity-H"),
		Name("DescendantFonts"): Array{pdfFont.DescendantRef},
		Name("ToUnicode"):       pdfFont.ToUnicodeRef,
	}
}

// buildToUnicodeCMap builds the ToUnicode CMap for a font.
func (m *FontManager) buildToUnicodeCMap(pdfFont *PDFFont) []byte {
	var buf bytes.Buffer

	// CMap header
	buf.WriteString("/CIDInit /ProcSet findresource begin\n")
	buf.WriteString("12 dict begin\n")
	buf.WriteString("begincmap\n")
	buf.WriteString("/CIDSystemInfo <<\n")
	buf.WriteString("  /Registry (Adobe)\n")
	buf.WriteString("  /Ordering (UCS)\n")
	buf.WriteString("  /Supplement 0\n")
	buf.WriteString(">> def\n")
	buf.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	buf.WriteString("/CMapType 2 def\n")

	// Code space range
	buf.WriteString("1 begincodespacerange\n")
	buf.WriteString("<0000> <FFFF>\n")
	buf.WriteString("endcodespacerange\n")

	// Build char mappings
	// We need to map from CID (new glyph ID) to Unicode
	if pdfFont.Subset != nil {
		// Collect mappings: CID -> Unicode
		type mapping struct {
			cid  uint16
			char rune
		}
		var mappings []mapping

		// Get glyphs from the glyph set
		glyphIDs := pdfFont.Glyphs.Sorted()
		for _, oldGID := range glyphIDs {
			newGID, ok := pdfFont.Subset.GlyphMapping[oldGID]
			if !ok {
				continue
			}
			// We need to track the original character
			// For now, use a placeholder - in production, this would come from shaped text
			mappings = append(mappings, mapping{cid: newGID, char: rune(oldGID)})
		}

		// Sort by CID
		sort.Slice(mappings, func(i, j int) bool {
			return mappings[i].cid < mappings[j].cid
		})

		// Write bfchar entries (up to 100 per block)
		for i := 0; i < len(mappings); i += 100 {
			end := i + 100
			if end > len(mappings) {
				end = len(mappings)
			}
			chunk := mappings[i:end]

			fmt.Fprintf(&buf, "%d beginbfchar\n", len(chunk))
			for _, m := range chunk {
				if m.char < 0x10000 {
					fmt.Fprintf(&buf, "<%04X> <%04X>\n", m.cid, m.char)
				} else {
					// Surrogate pair for characters outside BMP
					high := 0xD800 + ((int(m.char) - 0x10000) >> 10)
					low := 0xDC00 + ((int(m.char) - 0x10000) & 0x3FF)
					fmt.Fprintf(&buf, "<%04X> <%04X%04X>\n", m.cid, high, low)
				}
			}
			buf.WriteString("endbfchar\n")
		}
	}

	// CMap footer
	buf.WriteString("endcmap\n")
	buf.WriteString("CMapName currentdict /CMap defineresource pop\n")
	buf.WriteString("end\n")
	buf.WriteString("end\n")

	return buf.Bytes()
}

// sanitizePostScriptName removes invalid characters from a PostScript name.
func sanitizePostScriptName(name string) string {
	var result bytes.Buffer
	for _, r := range name {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	if result.Len() == 0 {
		return "Font"
	}
	return result.String()
}

// EncodeGlyphString encodes a glyph string for PDF text operators.
// Returns a hex string suitable for use with Tj operator in Identity-H encoding.
func EncodeGlyphString(glyphIDs []uint16) string {
	var buf bytes.Buffer
	buf.WriteByte('<')
	for _, gid := range glyphIDs {
		fmt.Fprintf(&buf, "%04X", gid)
	}
	buf.WriteByte('>')
	return buf.String()
}

// EncodeGlyphID encodes a single glyph ID for PDF.
func EncodeGlyphID(glyphID uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, glyphID)
	return buf
}
