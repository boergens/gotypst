package pdf

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWriterAlloc(t *testing.T) {
	w := NewWriter(V1_7)
	ref1 := w.Alloc()
	ref2 := w.Alloc()
	ref3 := w.Alloc()

	if ref1.Num() != 1 {
		t.Errorf("first ref = %d, want 1", ref1.Num())
	}
	if ref2.Num() != 2 {
		t.Errorf("second ref = %d, want 2", ref2.Num())
	}
	if ref3.Num() != 3 {
		t.Errorf("third ref = %d, want 3", ref3.Num())
	}
}

func TestDocumentBasic(t *testing.T) {
	doc := NewDocument(V1_7)
	doc.Info().SetTitle("Test Document")
	doc.Info().SetProducer("GoTypst Test")

	// Add a single page.
	page := doc.AddPage(A4Width, A4Height)
	page.Finish()

	var buf bytes.Buffer
	err := doc.Finish(&buf)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	pdf := buf.String()

	// Check PDF header.
	if !strings.HasPrefix(pdf, "%PDF-1.7\n") {
		t.Errorf("missing PDF header")
	}

	// Check for binary marker.
	if !strings.Contains(pdf, "%\x80\x81\x82\x83") {
		t.Errorf("missing binary marker")
	}

	// Check for trailer.
	if !strings.Contains(pdf, "trailer") {
		t.Errorf("missing trailer")
	}

	// Check for xref.
	if !strings.Contains(pdf, "xref") {
		t.Errorf("missing xref")
	}

	// Check for EOF.
	if !strings.HasSuffix(pdf, "%%EOF\n") {
		t.Errorf("missing EOF marker")
	}

	// Check for catalog.
	if !strings.Contains(pdf, "/Type /Catalog") {
		t.Errorf("missing catalog")
	}

	// Check for page tree.
	if !strings.Contains(pdf, "/Type /Pages") {
		t.Errorf("missing page tree")
	}

	// Check for page.
	if !strings.Contains(pdf, "/Type /Page") {
		t.Errorf("missing page")
	}

	// Check for media box.
	if !strings.Contains(pdf, "/MediaBox") {
		t.Errorf("missing MediaBox")
	}
}

func TestDocumentMultiplePages(t *testing.T) {
	doc := NewDocument(V1_4)

	// Add three pages with different sizes.
	page1 := doc.AddPage(A4Width, A4Height)
	page1.Finish()

	page2 := doc.AddPage(LetterWidth, LetterHeight)
	page2.Finish()

	page3 := doc.AddPage(A3Width, A3Height)
	page3.Finish()

	var buf bytes.Buffer
	err := doc.Finish(&buf)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	pdf := buf.String()

	// Check page count in page tree.
	if !strings.Contains(pdf, "/Count 3") {
		t.Errorf("wrong page count")
	}

	// Check for Kids array.
	if !strings.Contains(pdf, "/Kids [") {
		t.Errorf("missing Kids array")
	}
}

func TestDocumentWithContent(t *testing.T) {
	doc := NewDocument(V1_7)

	// Add a page with content.
	page := doc.AddPage(A4Width, A4Height)
	content := doc.AddContentStream([]byte("BT /F1 12 Tf 100 700 Td (Hello World) Tj ET"))
	page.SetContents(content)
	page.Finish()

	var buf bytes.Buffer
	err := doc.Finish(&buf)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	pdf := buf.String()

	// Check for content stream.
	if !strings.Contains(pdf, "stream") {
		t.Errorf("missing stream")
	}
	if !strings.Contains(pdf, "endstream") {
		t.Errorf("missing endstream")
	}

	// Check for Contents reference in page.
	if !strings.Contains(pdf, "/Contents") {
		t.Errorf("missing Contents reference")
	}
}

func TestDocumentInfo(t *testing.T) {
	doc := NewDocument(V1_7)
	info := doc.Info()
	info.SetTitle("My Title")
	info.SetAuthor("Test Author")
	info.SetSubject("Test Subject")
	info.SetKeywords("test, pdf, gotypst")
	info.SetCreator("Test Creator")
	info.SetProducer("GoTypst")
	info.SetCreationDate(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))

	page := doc.AddPage(A4Width, A4Height)
	page.Finish()

	var buf bytes.Buffer
	err := doc.Finish(&buf)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	pdf := buf.String()

	if !strings.Contains(pdf, "/Title (My Title)") {
		t.Errorf("missing Title")
	}
	if !strings.Contains(pdf, "/Author (Test Author)") {
		t.Errorf("missing Author")
	}
	if !strings.Contains(pdf, "/Producer (GoTypst)") {
		t.Errorf("missing Producer")
	}
}

func TestCatalogOptions(t *testing.T) {
	doc := NewDocument(V1_7)
	cat := doc.Catalog()
	cat.SetLang("en-US")
	cat.SetPageMode(PageModeOutlines)
	cat.SetPageLayout(PageLayoutTwoColumnLeft)
	cat.SetMarked(true)

	page := doc.AddPage(A4Width, A4Height)
	page.Finish()

	var buf bytes.Buffer
	err := doc.Finish(&buf)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	pdf := buf.String()

	if !strings.Contains(pdf, "/Lang (en-US)") {
		t.Errorf("missing Lang")
	}
	if !strings.Contains(pdf, "/PageMode /UseOutlines") {
		t.Errorf("missing PageMode")
	}
	if !strings.Contains(pdf, "/PageLayout /TwoColumnLeft") {
		t.Errorf("missing PageLayout")
	}
}

func TestPageResources(t *testing.T) {
	doc := NewDocument(V1_7)

	// Allocate a font reference (in a real scenario, you'd write the font).
	fontRef := doc.Writer().Alloc()

	page := doc.AddPage(A4Width, A4Height)
	res := page.Resources()
	res.AddFont(Name("F1"), fontRef)
	res.AddProcSet(Name("PDF"))
	res.AddProcSet(Name("Text"))
	page.Finish()

	var buf bytes.Buffer
	err := doc.Finish(&buf)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	pdf := buf.String()

	if !strings.Contains(pdf, "/Font") {
		t.Errorf("missing Font in resources")
	}
	if !strings.Contains(pdf, "/F1") {
		t.Errorf("missing F1 font name")
	}
	if !strings.Contains(pdf, "/ProcSet") {
		t.Errorf("missing ProcSet")
	}
}

func TestPDFVersions(t *testing.T) {
	tests := []struct {
		version Version
		header  string
	}{
		{V1_4, "%PDF-1.4"},
		{V1_5, "%PDF-1.5"},
		{V1_6, "%PDF-1.6"},
		{V1_7, "%PDF-1.7"},
		{V2_0, "%PDF-2.0"},
	}

	for _, tt := range tests {
		doc := NewDocument(tt.version)
		page := doc.AddPage(100, 100)
		page.Finish()

		var buf bytes.Buffer
		err := doc.Finish(&buf)
		if err != nil {
			t.Fatalf("Finish failed for %v: %v", tt.version, err)
		}

		pdf := buf.String()
		if !strings.HasPrefix(pdf, tt.header) {
			t.Errorf("version %v: got header %q, want prefix %q",
				tt.version, pdf[:20], tt.header)
		}
	}
}

func TestXrefTable(t *testing.T) {
	doc := NewDocument(V1_7)

	// Add multiple pages to generate multiple objects.
	for i := 0; i < 5; i++ {
		page := doc.AddPage(100, 100)
		page.Finish()
	}

	var buf bytes.Buffer
	err := doc.Finish(&buf)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	pdf := buf.String()

	// Check xref header format.
	if !strings.Contains(pdf, "xref\n0 ") {
		t.Errorf("xref format incorrect")
	}

	// Check for object 0 (free list head).
	if !strings.Contains(pdf, "0000000000 65535 f ") {
		t.Errorf("missing object 0 in xref")
	}

	// Check startxref.
	if !strings.Contains(pdf, "startxref\n") {
		t.Errorf("missing startxref")
	}
}

func TestEmptyDocument(t *testing.T) {
	doc := NewDocument(V1_7)

	// A document with no pages should still be valid.
	var buf bytes.Buffer
	err := doc.Finish(&buf)
	if err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	pdf := buf.String()

	// Should have header.
	if !strings.HasPrefix(pdf, "%PDF-1.7") {
		t.Errorf("missing header")
	}

	// Should have empty page count.
	if !strings.Contains(pdf, "/Count 0") {
		t.Errorf("expected Count 0 for empty document")
	}
}
