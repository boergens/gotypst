package eval

import (
	"testing"
)

func TestTextElement(t *testing.T) {
	t.Run("creates basic text element", func(t *testing.T) {
		elem := NewTextElement("Hello, World!")
		if elem.Text != "Hello, World!" {
			t.Errorf("expected 'Hello, World!', got %q", elem.Text)
		}
		if elem.HasStyling() {
			t.Error("expected no styling on new element")
		}
	})

	t.Run("implements ContentElement", func(t *testing.T) {
		var _ ContentElement = &TextElement{}
	})
}

func TestTextElementWithFont(t *testing.T) {
	elem := NewTextElement("text").WithFont("Arial", "Helvetica")
	if len(elem.Font) != 2 {
		t.Errorf("expected 2 fonts, got %d", len(elem.Font))
	}
	if elem.Font[0] != "Arial" {
		t.Errorf("expected 'Arial', got %q", elem.Font[0])
	}
	if elem.Font[1] != "Helvetica" {
		t.Errorf("expected 'Helvetica', got %q", elem.Font[1])
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementWithSize(t *testing.T) {
	elem := NewTextElement("text").WithSize(12.0)
	if elem.Size != 12.0 {
		t.Errorf("expected size 12.0, got %f", elem.Size)
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementWithWeight(t *testing.T) {
	elem := NewTextElement("text").WithWeight(FontWeightBold)
	if elem.Weight != FontWeightBold {
		t.Errorf("expected FontWeightBold (700), got %d", elem.Weight)
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementWithStyle(t *testing.T) {
	elem := NewTextElement("text").WithStyle(FontStyleItalic)
	if elem.Style != FontStyleItalic {
		t.Errorf("expected FontStyleItalic, got %d", elem.Style)
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementWithFill(t *testing.T) {
	elem := NewTextElement("text").WithFill(Color{R: 255, G: 0, B: 0, A: 255})
	if elem.Fill == nil {
		t.Fatal("expected Fill to be set")
	}
	if elem.Fill.R != 255 || elem.Fill.G != 0 || elem.Fill.B != 0 {
		t.Errorf("expected red color, got %+v", elem.Fill)
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementWithStroke(t *testing.T) {
	stroke := Stroke{
		Paint:     &Color{R: 0, G: 0, B: 0, A: 255},
		Thickness: 1.0,
		Cap:       LineCapRound,
		Join:      LineJoinRound,
	}
	elem := NewTextElement("text").WithStroke(stroke)
	if elem.Stroke == nil {
		t.Fatal("expected Stroke to be set")
	}
	if elem.Stroke.Thickness != 1.0 {
		t.Errorf("expected thickness 1.0, got %f", elem.Stroke.Thickness)
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementWithUnderline(t *testing.T) {
	elem := NewTextElement("text").WithUnderline(NewUnderline())
	if elem.Underline == nil {
		t.Fatal("expected Underline to be set")
	}
	if !elem.Underline.Evade {
		t.Error("expected Evade to be true by default")
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementWithStrikethrough(t *testing.T) {
	elem := NewTextElement("text").WithStrikethrough(NewStrikethrough())
	if elem.Strikethrough == nil {
		t.Fatal("expected Strikethrough to be set")
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementWithOverline(t *testing.T) {
	elem := NewTextElement("text").WithOverline(NewOverline())
	if elem.Overline == nil {
		t.Fatal("expected Overline to be set")
	}
	if !elem.Overline.Evade {
		t.Error("expected Evade to be true by default")
	}
	if !elem.HasStyling() {
		t.Error("expected HasStyling to return true")
	}
}

func TestTextElementClone(t *testing.T) {
	original := NewTextElement("original").
		WithFont("Arial").
		WithSize(12.0).
		WithWeight(FontWeightBold).
		WithStyle(FontStyleItalic).
		WithFill(Color{R: 255, G: 0, B: 0, A: 255}).
		WithUnderline(NewUnderline())

	cloned := original.Clone()

	// Verify values are copied
	if cloned.Text != original.Text {
		t.Errorf("expected text %q, got %q", original.Text, cloned.Text)
	}
	if cloned.Size != original.Size {
		t.Errorf("expected size %f, got %f", original.Size, cloned.Size)
	}
	if cloned.Weight != original.Weight {
		t.Errorf("expected weight %d, got %d", original.Weight, cloned.Weight)
	}
	if cloned.Style != original.Style {
		t.Errorf("expected style %d, got %d", original.Style, cloned.Style)
	}

	// Verify independence
	cloned.Text = "modified"
	if original.Text == cloned.Text {
		t.Error("clone should be independent of original")
	}

	cloned.Font[0] = "Times"
	if original.Font[0] == cloned.Font[0] {
		t.Error("clone font slice should be independent")
	}
}

func TestFontWeightFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected FontWeight
	}{
		{"thin", FontWeightThin},
		{"extralight", FontWeightExtraLight},
		{"extra-light", FontWeightExtraLight},
		{"light", FontWeightLight},
		{"regular", FontWeightRegular},
		{"normal", FontWeightRegular},
		{"medium", FontWeightMedium},
		{"semibold", FontWeightSemiBold},
		{"semi-bold", FontWeightSemiBold},
		{"bold", FontWeightBold},
		{"extrabold", FontWeightExtraBold},
		{"extra-bold", FontWeightExtraBold},
		{"black", FontWeightBlack},
		{"unknown", FontWeightInherit},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FontWeightFromString(tt.input)
			if result != tt.expected {
				t.Errorf("FontWeightFromString(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFontStyleFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected FontStyle
	}{
		{"normal", FontStyleNormal},
		{"italic", FontStyleItalic},
		{"oblique", FontStyleOblique},
		{"unknown", FontStyleInherit},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FontStyleFromString(tt.input)
			if result != tt.expected {
				t.Errorf("FontStyleFromString(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLineCapFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected LineCap
	}{
		{"butt", LineCapButt},
		{"round", LineCapRound},
		{"square", LineCapSquare},
		{"unknown", LineCapButt},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := LineCapFromString(tt.input)
			if result != tt.expected {
				t.Errorf("LineCapFromString(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLineJoinFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected LineJoin
	}{
		{"miter", LineJoinMiter},
		{"round", LineJoinRound},
		{"bevel", LineJoinBevel},
		{"unknown", LineJoinMiter},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := LineJoinFromString(tt.input)
			if result != tt.expected {
				t.Errorf("LineJoinFromString(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTextDecorationDefaults(t *testing.T) {
	t.Run("underline has evade true", func(t *testing.T) {
		deco := NewUnderline()
		if !deco.Evade {
			t.Error("expected Evade to be true for underline")
		}
	})

	t.Run("strikethrough has evade false", func(t *testing.T) {
		deco := NewStrikethrough()
		if deco.Evade {
			t.Error("expected Evade to be false for strikethrough")
		}
	})

	t.Run("overline has evade true", func(t *testing.T) {
		deco := NewOverline()
		if !deco.Evade {
			t.Error("expected Evade to be true for overline")
		}
	})
}

func TestTextDecorationWithStroke(t *testing.T) {
	deco := &TextDecoration{
		Stroke: &Stroke{
			Paint:     &Color{R: 0, G: 0, B: 255, A: 255},
			Thickness: 2.0,
		},
	}
	if deco.Stroke == nil {
		t.Fatal("expected Stroke to be set")
	}
	if deco.Stroke.Thickness != 2.0 {
		t.Errorf("expected thickness 2.0, got %f", deco.Stroke.Thickness)
	}
}

func TestTextDecorationWithOffset(t *testing.T) {
	offset := 3.0
	deco := &TextDecoration{
		Offset: &offset,
	}
	if deco.Offset == nil {
		t.Fatal("expected Offset to be set")
	}
	if *deco.Offset != 3.0 {
		t.Errorf("expected offset 3.0, got %f", *deco.Offset)
	}
}

func TestDashPattern(t *testing.T) {
	dash := DashPattern{
		Array: []float64{5.0, 3.0},
		Phase: 1.0,
	}
	if len(dash.Array) != 2 {
		t.Errorf("expected 2 dash elements, got %d", len(dash.Array))
	}
	if dash.Phase != 1.0 {
		t.Errorf("expected phase 1.0, got %f", dash.Phase)
	}
}
