package inline

import (
	"testing"

	"github.com/boergens/gotypst/eval"
)

func TestMathRootItem(t *testing.T) {
	t.Run("implements Item interface", func(t *testing.T) {
		item := &MathRootItem{}
		var _ Item = item // Compile-time check
	})

	t.Run("NaturalWidth returns width", func(t *testing.T) {
		item := &MathRootItem{width: 50}
		if got := item.NaturalWidth(); got != 50 {
			t.Errorf("NaturalWidth() = %v, want 50", got)
		}
	})

	t.Run("Textual returns sqrt symbol", func(t *testing.T) {
		item := &MathRootItem{}
		if got := item.Textual(); got != "\u221A" {
			t.Errorf("Textual() = %q, want %q", got, "\u221A")
		}
	})
}

func TestMathFracItem(t *testing.T) {
	t.Run("implements Item interface", func(t *testing.T) {
		item := &MathFracItem{}
		var _ Item = item // Compile-time check
	})

	t.Run("NaturalWidth returns width", func(t *testing.T) {
		item := &MathFracItem{width: 30}
		if got := item.NaturalWidth(); got != 30 {
			t.Errorf("NaturalWidth() = %v, want 30", got)
		}
	})
}

func TestMathAttachItem(t *testing.T) {
	t.Run("implements Item interface", func(t *testing.T) {
		item := &MathAttachItem{}
		var _ Item = item // Compile-time check
	})

	t.Run("NaturalWidth returns width", func(t *testing.T) {
		item := &MathAttachItem{width: 40}
		if got := item.NaturalWidth(); got != 40 {
			t.Errorf("NaturalWidth() = %v, want 40", got)
		}
	})
}

func TestMathDelimitedItem(t *testing.T) {
	t.Run("implements Item interface", func(t *testing.T) {
		item := &MathDelimitedItem{}
		var _ Item = item // Compile-time check
	})

	t.Run("NaturalWidth returns width", func(t *testing.T) {
		item := &MathDelimitedItem{width: 60}
		if got := item.NaturalWidth(); got != 60 {
			t.Errorf("NaturalWidth() = %v, want 60", got)
		}
	})
}

func TestLayoutMathRoot(t *testing.T) {
	fontSize := Abs(12.0)
	ctx := NewMathContext(nil, fontSize)

	t.Run("square root without index", func(t *testing.T) {
		elem := &eval.MathRootElement{
			Index: eval.Content{},
			Radicand: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "x"},
				},
			},
		}

		item := LayoutMathRoot(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathRoot returned nil")
		}

		if item.Index != nil {
			t.Error("Expected Index to be nil for square root")
		}

		if item.width <= 0 {
			t.Errorf("Expected positive width, got %v", item.width)
		}

		if item.height <= 0 {
			t.Errorf("Expected positive height, got %v", item.height)
		}
	})

	t.Run("nth root with index", func(t *testing.T) {
		elem := &eval.MathRootElement{
			Index: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "3"},
				},
			},
			Radicand: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "x"},
				},
			},
		}

		item := LayoutMathRoot(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathRoot returned nil")
		}

		if item.Index == nil {
			t.Error("Expected Index to be non-nil for nth root")
		}
	})

	t.Run("empty radicand", func(t *testing.T) {
		elem := &eval.MathRootElement{
			Index:    eval.Content{},
			Radicand: eval.Content{},
		}

		item := LayoutMathRoot(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathRoot returned nil")
		}

		// Should still have dimensions (for the radical symbol)
		if item.width <= 0 {
			t.Errorf("Expected positive width for empty radicand, got %v", item.width)
		}
	})
}

func TestLayoutMathFrac(t *testing.T) {
	fontSize := Abs(12.0)
	ctx := NewMathContext(nil, fontSize)

	t.Run("simple fraction", func(t *testing.T) {
		elem := &eval.MathFracElement{
			Num: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "1"},
				},
			},
			Denom: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "2"},
				},
			},
		}

		item := LayoutMathFrac(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathFrac returned nil")
		}

		if item.Num == nil {
			t.Error("Expected Num to be non-nil")
		}

		if item.Denom == nil {
			t.Error("Expected Denom to be non-nil")
		}

		if item.width <= 0 {
			t.Errorf("Expected positive width, got %v", item.width)
		}

		if item.height <= 0 {
			t.Errorf("Expected positive height, got %v", item.height)
		}

		// Height should be greater than a single line (num + denom + line)
		minHeight := fontSize * 2
		if item.height < minHeight {
			t.Errorf("Expected height >= %v, got %v", minHeight, item.height)
		}
	})

	t.Run("empty numerator and denominator", func(t *testing.T) {
		elem := &eval.MathFracElement{
			Num:   eval.Content{},
			Denom: eval.Content{},
		}

		item := LayoutMathFrac(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathFrac returned nil")
		}

		// Should still have dimensions
		if item.width <= 0 {
			t.Errorf("Expected positive width, got %v", item.width)
		}
	})
}

func TestLayoutMathAttach(t *testing.T) {
	fontSize := Abs(12.0)
	ctx := NewMathContext(nil, fontSize)

	t.Run("superscript only", func(t *testing.T) {
		elem := &eval.MathAttachElement{
			Base: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "x"},
				},
			},
			Superscript: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "2"},
				},
			},
			Subscript: eval.Content{},
			Primes:    0,
		}

		item := LayoutMathAttach(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathAttach returned nil")
		}

		if item.Superscript == nil {
			t.Error("Expected Superscript to be non-nil")
		}

		if item.Subscript != nil {
			t.Error("Expected Subscript to be nil")
		}
	})

	t.Run("subscript only", func(t *testing.T) {
		elem := &eval.MathAttachElement{
			Base: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "x"},
				},
			},
			Superscript: eval.Content{},
			Subscript: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "i"},
				},
			},
			Primes: 0,
		}

		item := LayoutMathAttach(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathAttach returned nil")
		}

		if item.Subscript == nil {
			t.Error("Expected Subscript to be non-nil")
		}
	})

	t.Run("both subscript and superscript", func(t *testing.T) {
		elem := &eval.MathAttachElement{
			Base: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "x"},
				},
			},
			Superscript: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "2"},
				},
			},
			Subscript: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "i"},
				},
			},
			Primes: 0,
		}

		item := LayoutMathAttach(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathAttach returned nil")
		}

		if item.Superscript == nil {
			t.Error("Expected Superscript to be non-nil")
		}

		if item.Subscript == nil {
			t.Error("Expected Subscript to be non-nil")
		}
	})

	t.Run("with primes", func(t *testing.T) {
		elem := &eval.MathAttachElement{
			Base: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "f"},
				},
			},
			Superscript: eval.Content{},
			Subscript:   eval.Content{},
			Primes:      2,
		}

		item := LayoutMathAttach(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathAttach returned nil")
		}

		if item.Primes != 2 {
			t.Errorf("Expected Primes = 2, got %d", item.Primes)
		}

		// Width should include prime width
		baseOnlyElem := &eval.MathAttachElement{
			Base: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "f"},
				},
			},
			Primes: 0,
		}
		baseOnlyItem := LayoutMathAttach(ctx, baseOnlyElem)

		if item.width <= baseOnlyItem.width {
			t.Error("Expected width with primes to be greater than base only")
		}
	})
}

func TestLayoutMathDelimited(t *testing.T) {
	fontSize := Abs(12.0)
	ctx := NewMathContext(nil, fontSize)

	t.Run("parentheses", func(t *testing.T) {
		elem := &eval.MathDelimitedElement{
			Open:  "(",
			Close: ")",
			Body: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "a + b"},
				},
			},
		}

		item := LayoutMathDelimited(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathDelimited returned nil")
		}

		if item.Open != "(" {
			t.Errorf("Expected Open = '(', got %q", item.Open)
		}

		if item.Close != ")" {
			t.Errorf("Expected Close = ')', got %q", item.Close)
		}

		if item.Body == nil {
			t.Error("Expected Body to be non-nil")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		elem := &eval.MathDelimitedElement{
			Open:  "[",
			Close: "]",
			Body:  eval.Content{},
		}

		item := LayoutMathDelimited(ctx, elem)

		if item == nil {
			t.Fatal("LayoutMathDelimited returned nil")
		}

		// Without a shaping context, delimiter width is 0
		// Just ensure dimensions are non-negative
		if item.width < 0 {
			t.Errorf("Expected non-negative width, got %v", item.width)
		}

		if item.height <= 0 {
			t.Errorf("Expected positive height, got %v", item.height)
		}
	})
}

func TestDefaultMathLayoutConfig(t *testing.T) {
	fontSize := Abs(12.0)
	config := DefaultMathLayoutConfig(fontSize)

	if config.FontSize != fontSize {
		t.Errorf("FontSize = %v, want %v", config.FontSize, fontSize)
	}

	if config.ScriptScale <= 0 || config.ScriptScale >= 1 {
		t.Errorf("ScriptScale should be between 0 and 1, got %v", config.ScriptScale)
	}

	if config.ScriptScriptScale <= 0 || config.ScriptScriptScale >= config.ScriptScale {
		t.Errorf("ScriptScriptScale should be less than ScriptScale, got %v", config.ScriptScriptScale)
	}

	if config.RuleThickness <= 0 {
		t.Errorf("RuleThickness should be positive, got %v", config.RuleThickness)
	}
}

func TestBuildMathRootFrame(t *testing.T) {
	fontSize := Abs(12.0)
	ctx := NewMathContext(nil, fontSize)

	t.Run("basic root", func(t *testing.T) {
		item := &MathRootItem{
			Index: nil,
			Radicand: &MathFrame{
				Width:    20,
				Height:   fontSize * 1.2,
				Baseline: fontSize,
			},
			width:    30,
			height:   fontSize * 1.5,
			baseline: fontSize * 1.2,
		}

		frame := BuildMathRootFrame(item, ctx)

		if frame == nil {
			t.Fatal("BuildMathRootFrame returned nil")
		}

		if frame.Size.Width != item.width {
			t.Errorf("Frame width = %v, want %v", frame.Size.Width, item.width)
		}

		if frame.Size.Height != item.height {
			t.Errorf("Frame height = %v, want %v", frame.Size.Height, item.height)
		}
	})
}

func TestBuildMathFracFrame(t *testing.T) {
	fontSize := Abs(12.0)
	ctx := NewMathContext(nil, fontSize)

	t.Run("basic fraction", func(t *testing.T) {
		item := &MathFracItem{
			Num: &MathFrame{
				Width:    15,
				Height:   fontSize * 1.2,
				Baseline: fontSize,
			},
			Denom: &MathFrame{
				Width:    15,
				Height:   fontSize * 1.2,
				Baseline: fontSize,
			},
			width:    20,
			height:   fontSize * 3,
			baseline: fontSize * 1.5,
		}

		frame := BuildMathFracFrame(item, ctx)

		if frame == nil {
			t.Fatal("BuildMathFracFrame returned nil")
		}

		if frame.Size.Width != item.width {
			t.Errorf("Frame width = %v, want %v", frame.Size.Width, item.width)
		}
	})
}

func TestNestedMathElements(t *testing.T) {
	fontSize := Abs(12.0)
	ctx := NewMathContext(nil, fontSize)

	t.Run("fraction in root", func(t *testing.T) {
		// sqrt(a/b)
		fracElem := &eval.MathFracElement{
			Num: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "a"},
				},
			},
			Denom: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "b"},
				},
			},
		}

		rootElem := &eval.MathRootElement{
			Index: eval.Content{},
			Radicand: eval.Content{
				Elements: []eval.ContentElement{
					fracElem,
				},
			},
		}

		item := LayoutMathRoot(ctx, rootElem)

		if item == nil {
			t.Fatal("LayoutMathRoot returned nil for nested content")
		}

		if item.Radicand == nil {
			t.Fatal("Expected Radicand to be non-nil")
		}

		// Radicand should contain the fraction
		if len(item.Radicand.Items) == 0 {
			t.Error("Expected Radicand to contain items")
		}
	})

	t.Run("superscript with root", func(t *testing.T) {
		// x^sqrt(2)
		rootElem := &eval.MathRootElement{
			Index: eval.Content{},
			Radicand: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "2"},
				},
			},
		}

		attachElem := &eval.MathAttachElement{
			Base: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "x"},
				},
			},
			Superscript: eval.Content{
				Elements: []eval.ContentElement{
					rootElem,
				},
			},
			Subscript: eval.Content{},
			Primes:    0,
		}

		item := LayoutMathAttach(ctx, attachElem)

		if item == nil {
			t.Fatal("LayoutMathAttach returned nil for nested content")
		}

		if item.Superscript == nil {
			t.Fatal("Expected Superscript to be non-nil")
		}
	})
}

func TestMathFrame(t *testing.T) {
	t.Run("empty frame", func(t *testing.T) {
		frame := &MathFrame{
			Items:    nil,
			Width:    0,
			Height:   12,
			Baseline: 10,
		}

		if frame.Width != 0 {
			t.Errorf("Empty frame width should be 0, got %v", frame.Width)
		}
	})

	t.Run("frame with items", func(t *testing.T) {
		frame := &MathFrame{
			Items:    []Item{&TextItem{}},
			Width:    50,
			Height:   14,
			Baseline: 12,
		}

		if len(frame.Items) != 1 {
			t.Errorf("Frame should have 1 item, got %d", len(frame.Items))
		}
	})
}
