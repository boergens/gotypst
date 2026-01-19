package pdf

import (
	"strings"
	"testing"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/inline"
)

func TestContentStream_GraphicsState(t *testing.T) {
	cs := NewContentStream()

	cs.SaveState()
	cs.SetFillColorRGB(1, 0, 0)
	cs.SetStrokeColorRGB(0, 0, 1)
	cs.RestoreState()

	want := "q\n1 0 0 rg\n0 0 1 RG\nQ\n"
	if got := cs.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestContentStream_Text(t *testing.T) {
	cs := NewContentStream()

	cs.BeginText()
	cs.SetFont("/F1", 12)
	cs.SetTextMatrixPos(100, 200)
	cs.ShowText("Hello")
	cs.EndText()

	output := cs.String()

	if !strings.Contains(output, "BT") {
		t.Error("missing BT operator")
	}
	if !strings.Contains(output, "/F1 12 Tf") {
		t.Error("missing font setting")
	}
	if !strings.Contains(output, "1 0 0 1 100 200 Tm") {
		t.Error("missing text matrix")
	}
	if !strings.Contains(output, "(Hello) Tj") {
		t.Error("missing text show")
	}
	if !strings.Contains(output, "ET") {
		t.Error("missing ET operator")
	}
}

func TestContentStream_Path(t *testing.T) {
	cs := NewContentStream()

	cs.MoveTo(0, 0)
	cs.LineTo(100, 100)
	cs.Stroke()

	want := "0 0 m\n100 100 l\nS\n"
	if got := cs.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestContentStream_Rectangle(t *testing.T) {
	cs := NewContentStream()

	cs.Rectangle(10, 20, 100, 50)
	cs.Fill()

	want := "10 20 100 50 re\nf\n"
	if got := cs.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestContentStream_RoundedRectangle(t *testing.T) {
	cs := NewContentStream()

	cs.RoundedRectangle(0, 0, 100, 100, 10)
	cs.Stroke()

	output := cs.String()

	// Should contain move, lines, and curves
	if !strings.Contains(output, "m\n") {
		t.Error("missing move operator")
	}
	if !strings.Contains(output, "l\n") {
		t.Error("missing line operator")
	}
	if !strings.Contains(output, "c\n") {
		t.Error("missing curve operator")
	}
	if !strings.Contains(output, "h\n") {
		t.Error("missing close path operator")
	}
	if !strings.Contains(output, "S\n") {
		t.Error("missing stroke operator")
	}
}

func TestContentStream_Transform(t *testing.T) {
	cs := NewContentStream()

	cs.TranslateTransform(10, 20)

	want := "1 0 0 1 10 20 cm\n"
	if got := cs.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestContentStream_DashPattern(t *testing.T) {
	cs := NewContentStream()

	cs.SetDashPattern([]layout.Abs{3, 1}, 0)

	want := "[3 1] 0 d\n"
	if got := cs.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestContentStream_TextPositioning(t *testing.T) {
	cs := NewContentStream()

	cs.BeginText()
	cs.ShowTextWithPositioning([]TextPositionItem{
		TextPositionString("A"),
		TextPositionOffset(-50),
		TextPositionString("B"),
	})
	cs.EndText()

	output := cs.String()

	if !strings.Contains(output, "[(A)-50(B)] TJ") {
		t.Errorf("expected TJ array, got: %s", output)
	}
}

func TestContentStream_StringEscaping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "(hello)"},
		{"hello(world)", "(hello\\(world\\))"},
		{"back\\slash", "(back\\\\slash)"},
		{"new\nline", "(new\\nline)"},
	}

	for _, tt := range tests {
		got := encodeString(tt.input)
		if got != tt.want {
			t.Errorf("encodeString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{1.0, "1"},
		{1.5, "1.5"},
		{1.25, "1.25"},
		{1.2500, "1.25"},
		{0.0, "0"},
		{100.0, "100"},
		{0.1234, "0.1234"},
	}

	for _, tt := range tests {
		got := formatFloat(tt.input)
		if got != tt.want {
			t.Errorf("formatFloat(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestApplyStrokeStyle(t *testing.T) {
	cs := NewContentStream()

	stroke := &inline.FixedStroke{
		Paint:     nil,
		Thickness: 2,
		LineCap:   inline.LineCapRound,
		LineJoin:  inline.LineJoinBevel,
		DashArray: []layout.Abs{4, 2},
		DashPhase: 1,
	}

	cs.ApplyStrokeStyle(stroke)

	output := cs.String()

	if !strings.Contains(output, "2 w") {
		t.Error("missing line width")
	}
	if !strings.Contains(output, "1 J") {
		t.Error("missing line cap (round=1)")
	}
	if !strings.Contains(output, "2 j") {
		t.Error("missing line join (bevel=2)")
	}
	if !strings.Contains(output, "[4 2] 1 d") {
		t.Error("missing dash pattern")
	}
}

func TestContentStream_Color(t *testing.T) {
	cs := NewContentStream()

	c := &Color{R: 255, G: 128, B: 0, A: 255}
	cs.SetFillColor(c)

	output := cs.String()

	// Should have RGB values normalized to 0-1
	if !strings.Contains(output, "1 ") {
		t.Error("expected R=1")
	}
	if !strings.Contains(output, "rg") {
		t.Error("missing rg operator")
	}
}
