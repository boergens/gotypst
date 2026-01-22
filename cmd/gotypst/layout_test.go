package main

import (
	"testing"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/realize"
)

func TestLayoutWithRealization(t *testing.T) {
	// Create a minimal world for testing using testdata directory
	world, err := eval.NewFileWorld("testdata", "main.typ")
	if err != nil {
		t.Skipf("cannot create test world: %v", err)
	}

	t.Run("empty content", func(t *testing.T) {
		content := &eval.Content{}

		doc, err := layout(world, content)
		if err != nil {
			t.Fatalf("layout failed: %v", err)
		}

		if doc == nil {
			t.Fatal("layout returned nil document")
		}

		// Empty content should still produce pages
		if len(doc.Pages) == 0 {
			t.Error("expected at least one page for empty content")
		}
	})

	t.Run("nil content", func(t *testing.T) {
		doc, err := layout(world, nil)
		if err != nil {
			t.Fatalf("layout failed with nil content: %v", err)
		}

		if doc == nil {
			t.Fatal("layout returned nil document")
		}
	})

	t.Run("simple text content", func(t *testing.T) {
		content := &eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "Hello, World!"},
			},
		}

		doc, err := layout(world, content)
		if err != nil {
			t.Fatalf("layout failed: %v", err)
		}

		if doc == nil {
			t.Fatal("layout returned nil document")
		}

		if len(doc.Pages) == 0 {
			t.Error("expected at least one page")
		}
	})

	t.Run("content with paragraphs", func(t *testing.T) {
		// Multiple text elements should be grouped into a paragraph by realization
		content := &eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "First text "},
				&eval.TextElement{Text: "second text."},
			},
		}

		doc, err := layout(world, content)
		if err != nil {
			t.Fatalf("layout failed: %v", err)
		}

		if doc == nil {
			t.Fatal("layout returned nil document")
		}

		if len(doc.Pages) == 0 {
			t.Error("expected at least one page")
		}
	})

	t.Run("content with heading", func(t *testing.T) {
		content := &eval.Content{
			Elements: []eval.ContentElement{
				&eval.HeadingElement{
					Depth:   1,
					Content: eval.Content{Elements: []eval.ContentElement{&eval.TextElement{Text: "Title"}}},
				},
				&eval.TextElement{Text: "Some body text."},
			},
		}

		doc, err := layout(world, content)
		if err != nil {
			t.Fatalf("layout failed: %v", err)
		}

		if doc == nil {
			t.Fatal("layout returned nil document")
		}

		if len(doc.Pages) == 0 {
			t.Error("expected at least one page")
		}
	})
}

func TestConvertRealizedContent(t *testing.T) {
	t.Run("empty pairs", func(t *testing.T) {
		content := convertRealizedContent(nil)
		if content == nil {
			t.Fatal("convertRealizedContent returned nil")
		}
		if len(content.Elements) != 0 {
			t.Errorf("expected 0 elements, got %d", len(content.Elements))
		}
	})

	t.Run("single element pair", func(t *testing.T) {
		pairs := []realize.Pair{
			{
				Content: &eval.TextElement{Text: "Test"},
				Styles:  eval.EmptyStyleChain(),
			},
		}
		content := convertRealizedContent(pairs)
		if content == nil {
			t.Fatal("convertRealizedContent returned nil")
		}
		if len(content.Elements) != 1 {
			t.Errorf("expected 1 element, got %d", len(content.Elements))
		}
	})

	t.Run("multiple element pairs", func(t *testing.T) {
		pairs := []realize.Pair{
			{Content: &eval.TextElement{Text: "First"}, Styles: eval.EmptyStyleChain()},
			{Content: &eval.TextElement{Text: "Second"}, Styles: eval.EmptyStyleChain()},
			{Content: &eval.ParagraphElement{}, Styles: eval.EmptyStyleChain()},
		}
		content := convertRealizedContent(pairs)
		if content == nil {
			t.Fatal("convertRealizedContent returned nil")
		}
		if len(content.Elements) != 3 {
			t.Errorf("expected 3 elements, got %d", len(content.Elements))
		}
	})
}

func TestLayoutDocumentReturnsPagedDocument(t *testing.T) {
	world, err := eval.NewFileWorld("testdata", "main.typ")
	if err != nil {
		t.Skipf("cannot create test world: %v", err)
	}

	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.TextElement{Text: "Test content"},
		},
	}

	doc, err := layout(world, content)
	if err != nil {
		t.Fatalf("layout failed: %v", err)
	}

	// Verify the document has the expected structure
	if doc == nil {
		t.Fatal("expected non-nil PagedDocument")
	}

	// Verify pages have valid frames
	for i, page := range doc.Pages {
		if page.Frame.Width() <= 0 {
			t.Errorf("page %d has invalid width: %v", i, page.Frame.Width())
		}
		if page.Frame.Height() <= 0 {
			t.Errorf("page %d has invalid height: %v", i, page.Frame.Height())
		}
		if page.Number < 1 {
			t.Errorf("page %d has invalid number: %d", i, page.Number)
		}
	}
}
