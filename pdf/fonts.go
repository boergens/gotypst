package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"sort"

	"github.com/go-text/typesetting/font"
)

// FontResource represents a PDF font resource ready for embedding.
type FontResource struct {
	// Ref is the indirect object reference for the Type0 font dictionary.
	Ref Ref
	// ResourceName is the PDF resource name (e.g., "F1").
	ResourceName string
	// Objects contains all indirect objects needed for this font.
	Objects []IndirectObject
}

// Type0Font represents a PDF Type 0 (composite) font for CID fonts.
// Type0 fonts use CIDFonts as descendants for CJK and complex scripts.
type Type0Font struct {
	// Face is the font face this represents.
	Face *font.Face
	// Glyphs maps glyph IDs to Unicode code points used in this font.
	Glyphs map[uint16]rune
	// Widths maps glyph IDs to their advance widths in font units.
	Widths map[uint16]int
	// FontName is the PostScript name of the font.
	FontName string
}

// NewType0Font creates a new Type0Font for the given face.
func NewType0Font(face *font.Face) *Type0Font {
	fontName := "Font"
	if face != nil {
		// Try to get a meaningful name from the font
		fontName = fmt.Sprintf("F%p", face)
	}
	return &Type0Font{
		Face:     face,
		Glyphs:   make(map[uint16]rune),
		Widths:   make(map[uint16]int),
		FontName: fontName,
	}
}

// AddGlyph registers a glyph with its Unicode mapping.
func (f *Type0Font) AddGlyph(gid uint16, char rune, width int) {
	f.Glyphs[gid] = char
	f.Widths[gid] = width
}

// ToObjects generates all PDF objects needed for this font.
func (f *Type0Font) ToObjects(fontRef Ref, allocRef func() Ref) []IndirectObject {
	var objects []IndirectObject

	// Allocate refs for sub-objects
	descendantRef := allocRef()
	descriptorRef := allocRef()
	toUnicodeRef := allocRef()
	cidToGIDRef := allocRef()

	// Build width array: [cid [width] cid [width] ...]
	widthArray := f.buildWidthArray()

	// Create CIDSystemInfo
	cidSystemInfo := Dict{
		Name("Registry"):   String("Adobe"),
		Name("Ordering"):   String("Identity"),
		Name("Supplement"): Int(0),
	}

	// Create FontDescriptor
	descriptor := Dict{
		Name("Type"):        Name("FontDescriptor"),
		Name("FontName"):    Name(f.FontName),
		Name("Flags"):       Int(32), // Symbolic
		Name("FontBBox"):    Array{Int(-1000), Int(-500), Int(2000), Int(1500)},
		Name("ItalicAngle"): Int(0),
		Name("Ascent"):      Int(1000),
		Name("Descent"):     Int(-200),
		Name("CapHeight"):   Int(700),
		Name("StemV"):       Int(80),
	}
	objects = append(objects, IndirectObject{Ref: descriptorRef, Object: descriptor})

	// Create CIDToGIDMap (Identity mapping)
	cidToGIDMap := f.buildCIDToGIDMap()
	cidToGIDStream := Stream{
		Dict: Dict{},
		Data: cidToGIDMap,
	}
	cidToGIDStream.Compress()
	objects = append(objects, IndirectObject{Ref: cidToGIDRef, Object: cidToGIDStream})

	// Create CIDFont (DescendantFont)
	cidFont := Dict{
		Name("Type"):           Name("Font"),
		Name("Subtype"):        Name("CIDFontType2"),
		Name("BaseFont"):       Name(f.FontName),
		Name("CIDSystemInfo"):  cidSystemInfo,
		Name("FontDescriptor"): descriptorRef,
		Name("DW"):             Int(1000), // Default width
		Name("W"):              widthArray,
		Name("CIDToGIDMap"):    cidToGIDRef,
	}
	objects = append(objects, IndirectObject{Ref: descendantRef, Object: cidFont})

	// Create ToUnicode CMap
	toUnicodeCMap := f.buildToUnicodeCMap()
	toUnicodeStream := Stream{
		Dict: Dict{},
		Data: toUnicodeCMap,
	}
	toUnicodeStream.Compress()
	objects = append(objects, IndirectObject{Ref: toUnicodeRef, Object: toUnicodeStream})

	// Create Type0 font dictionary (the main font object)
	type0Dict := Dict{
		Name("Type"):            Name("Font"),
		Name("Subtype"):         Name("Type0"),
		Name("BaseFont"):        Name(f.FontName),
		Name("Encoding"):        Name("Identity-H"),
		Name("DescendantFonts"): Array{descendantRef},
		Name("ToUnicode"):       toUnicodeRef,
	}
	objects = append(objects, IndirectObject{Ref: fontRef, Object: type0Dict})

	return objects
}

// buildWidthArray creates the W (widths) array for the CIDFont.
// Format: [cid1 [w1] cid2 [w2] ...] or [cidStart cidEnd w w w ...]
func (f *Type0Font) buildWidthArray() Array {
	if len(f.Widths) == 0 {
		return Array{}
	}

	// Get sorted glyph IDs
	gids := make([]uint16, 0, len(f.Widths))
	for gid := range f.Widths {
		gids = append(gids, gid)
	}
	sort.Slice(gids, func(i, j int) bool { return gids[i] < gids[j] })

	// Build array using [cid [width]] format for simplicity
	var arr Array
	for _, gid := range gids {
		width := f.Widths[gid]
		arr = append(arr, Int(gid), Array{Int(width)})
	}

	return arr
}

// buildCIDToGIDMap creates an Identity CIDToGIDMap.
// This maps CID values directly to glyph IDs (CID = GID).
func (f *Type0Font) buildCIDToGIDMap() []byte {
	if len(f.Glyphs) == 0 {
		return nil
	}

	// Find max glyph ID
	maxGID := uint16(0)
	for gid := range f.Glyphs {
		if gid > maxGID {
			maxGID = gid
		}
	}

	// Create identity mapping: 2 bytes per entry
	mapSize := int(maxGID+1) * 2
	data := make([]byte, mapSize)

	for gid := range f.Glyphs {
		// Big-endian: CID -> GID (identity)
		offset := int(gid) * 2
		if offset+1 < len(data) {
			data[offset] = byte(gid >> 8)
			data[offset+1] = byte(gid)
		}
	}

	return data
}

// buildToUnicodeCMap creates a ToUnicode CMap for text extraction.
func (f *Type0Font) buildToUnicodeCMap() []byte {
	var buf bytes.Buffer

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
	buf.WriteString("1 begincodespacerange\n")
	buf.WriteString("<0000> <FFFF>\n")
	buf.WriteString("endcodespacerange\n")

	// Write character mappings in groups of up to 100
	gids := make([]uint16, 0, len(f.Glyphs))
	for gid := range f.Glyphs {
		gids = append(gids, gid)
	}
	sort.Slice(gids, func(i, j int) bool { return gids[i] < gids[j] })

	for i := 0; i < len(gids); i += 100 {
		end := i + 100
		if end > len(gids) {
			end = len(gids)
		}
		batch := gids[i:end]

		fmt.Fprintf(&buf, "%d beginbfchar\n", len(batch))
		for _, gid := range batch {
			char := f.Glyphs[gid]
			if char <= 0xFFFF {
				fmt.Fprintf(&buf, "<%04X> <%04X>\n", gid, char)
			} else {
				// Surrogate pair for characters outside BMP
				char -= 0x10000
				high := 0xD800 + (char >> 10)
				low := 0xDC00 + (char & 0x3FF)
				fmt.Fprintf(&buf, "<%04X> <%04X%04X>\n", gid, high, low)
			}
		}
		buf.WriteString("endbfchar\n")
	}

	buf.WriteString("endcmap\n")
	buf.WriteString("CMapName currentdict /CMap defineresource pop\n")
	buf.WriteString("end\n")
	buf.WriteString("end\n")

	return buf.Bytes()
}

// FontManager manages PDF fonts for a document.
type FontManager struct {
	// fonts maps font faces to their Type0Font data.
	fonts map[*font.Face]*Type0Font
	// resources maps font faces to their PDF resource names.
	resources map[*font.Face]string
	// nextFontNum is the next font number for resource naming.
	nextFontNum int
}

// NewFontManager creates a new FontManager.
func NewFontManager() *FontManager {
	return &FontManager{
		fonts:       make(map[*font.Face]*Type0Font),
		resources:   make(map[*font.Face]string),
		nextFontNum: 1,
	}
}

// RegisterGlyph registers a glyph for a font face.
// Returns the PDF resource name for the font.
func (m *FontManager) RegisterGlyph(face *font.Face, gid uint16, char rune, widthUnits int) string {
	f, ok := m.fonts[face]
	if !ok {
		f = NewType0Font(face)
		m.fonts[face] = f

		resourceName := fmt.Sprintf("F%d", m.nextFontNum)
		m.nextFontNum++
		m.resources[face] = resourceName
	}

	f.AddGlyph(gid, char, widthUnits)
	return m.resources[face]
}

// FontName returns the PDF resource name for a font face.
func (m *FontManager) FontName(face interface{}) string {
	if f, ok := face.(*font.Face); ok {
		if name, ok := m.resources[f]; ok {
			return "/" + name
		}
	}
	// Fallback for unknown fonts
	return "/F1"
}

// GetFonts returns all registered fonts.
func (m *FontManager) GetFonts() map[*font.Face]*Type0Font {
	return m.fonts
}

// GenerateResources generates all font resources and returns them.
// The allocRef function is used to allocate new object references.
func (m *FontManager) GenerateResources(allocRef func() Ref) []FontResource {
	var resources []FontResource

	for face, t0font := range m.fonts {
		fontRef := allocRef()
		objects := t0font.ToObjects(fontRef, allocRef)

		resources = append(resources, FontResource{
			Ref:          fontRef,
			ResourceName: m.resources[face],
			Objects:      objects,
		})
	}

	return resources
}

// buildCompressedStream creates a compressed PDF stream.
func buildCompressedStream(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
