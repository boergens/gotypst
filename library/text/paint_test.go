package text

import (
	"testing"

	"github.com/boergens/gotypst/layout/inline"
)

func TestNewColor(t *testing.T) {
	c := NewColor(255, 128, 64, 200)

	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 200 {
		t.Errorf("NewColor = %v, want {255, 128, 64, 200}", c)
	}
}

func TestNewRGB(t *testing.T) {
	c := NewRGB(100, 150, 200)

	if c.R != 100 || c.G != 150 || c.B != 200 || c.A != 255 {
		t.Errorf("NewRGB = %v, want {100, 150, 200, 255}", c)
	}
}

func TestColorFromHex(t *testing.T) {
	tests := []struct {
		hex     string
		want    Color
		wantErr bool
	}{
		// Short form RGB
		{"#f00", NewRGB(255, 0, 0), false},
		{"#0f0", NewRGB(0, 255, 0), false},
		{"#00f", NewRGB(0, 0, 255), false},
		{"abc", NewRGB(170, 187, 204), false},

		// Short form RGBA
		{"#f00f", NewColor(255, 0, 0, 255), false},
		{"#0f08", NewColor(0, 255, 0, 136), false},

		// Long form RGB
		{"#ff0000", NewRGB(255, 0, 0), false},
		{"00ff00", NewRGB(0, 255, 0), false},
		{"#0000ff", NewRGB(0, 0, 255), false},

		// Long form RGBA
		{"#ff000080", NewColor(255, 0, 0, 128), false},

		// Errors
		{"", Color{}, true},
		{"#gg", Color{}, true},
	}

	for _, tt := range tests {
		got, err := ColorFromHex(tt.hex)
		if (err != nil) != tt.wantErr {
			t.Errorf("ColorFromHex(%q) error = %v, wantErr %v", tt.hex, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ColorFromHex(%q) = %v, want %v", tt.hex, got, tt.want)
		}
	}
}

func TestColorString(t *testing.T) {
	tests := []struct {
		color Color
		want  string
	}{
		{NewRGB(255, 0, 0), "#ff0000"},
		{NewRGB(0, 255, 0), "#00ff00"},
		{NewRGB(0, 0, 255), "#0000ff"},
		{NewColor(255, 0, 0, 128), "#ff000080"},
	}

	for _, tt := range tests {
		if got := tt.color.String(); got != tt.want {
			t.Errorf("Color.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestPredefinedColors(t *testing.T) {
	if Black != NewRGB(0, 0, 0) {
		t.Error("Black should be #000000")
	}
	if White != NewRGB(255, 255, 255) {
		t.Error("White should be #ffffff")
	}
	if Red != NewRGB(255, 0, 0) {
		t.Error("Red should be #ff0000")
	}
	if Green != NewRGB(0, 128, 0) {
		t.Error("Green should be #008000")
	}
	if Blue != NewRGB(0, 0, 255) {
		t.Error("Blue should be #0000ff")
	}
}

func TestStroke(t *testing.T) {
	s := NewStroke(Red, inline.Abs(2)).
		WithCap(LineCapRound).
		WithJoin(LineJoinBevel)

	if s.Paint != Red {
		t.Error("Stroke paint should be red")
	}
	if s.Thickness != inline.Abs(2) {
		t.Errorf("Stroke thickness = %v, want 2", s.Thickness)
	}
	if s.Cap != LineCapRound {
		t.Errorf("Stroke cap = %v, want round", s.Cap)
	}
	if s.Join != LineJoinBevel {
		t.Errorf("Stroke join = %v, want bevel", s.Join)
	}
}

func TestStrokeDash(t *testing.T) {
	dash := NewDash(inline.Abs(4), inline.Abs(2))

	if len(dash.Array) != 2 {
		t.Errorf("Dash array length = %d, want 2", len(dash.Array))
	}
	if dash.Array[0] != inline.Abs(4) || dash.Array[1] != inline.Abs(2) {
		t.Errorf("Dash array = %v, want [4 2]", dash.Array)
	}
	if dash.Phase != 0 {
		t.Errorf("Dash phase = %v, want 0", dash.Phase)
	}
}

func TestStrokeToFixedStroke(t *testing.T) {
	s := NewStroke(Blue, inline.Abs(1.5)).
		WithCap(LineCapSquare).
		WithJoin(LineJoinRound).
		WithDash(NewDash(inline.Abs(3), inline.Abs(1)))

	fs := s.ToFixedStroke()

	if fs.Thickness != inline.Abs(1.5) {
		t.Errorf("FixedStroke thickness = %v, want 1.5", fs.Thickness)
	}
	if fs.LineCap != inline.LineCapSquare {
		t.Error("FixedStroke cap should be square")
	}
	if fs.LineJoin != inline.LineJoinRound {
		t.Error("FixedStroke join should be round")
	}
	if len(fs.DashArray) != 2 {
		t.Errorf("FixedStroke dash array length = %d, want 2", len(fs.DashArray))
	}
}

func TestLineCapStrings(t *testing.T) {
	tests := []struct {
		cap  LineCap
		want string
	}{
		{LineCapButt, "butt"},
		{LineCapRound, "round"},
		{LineCapSquare, "square"},
	}

	for _, tt := range tests {
		if got := tt.cap.String(); got != tt.want {
			t.Errorf("LineCap(%d).String() = %q, want %q", tt.cap, got, tt.want)
		}
	}
}

func TestLineJoinStrings(t *testing.T) {
	tests := []struct {
		join LineJoin
		want string
	}{
		{LineJoinMiter, "miter"},
		{LineJoinRound, "round"},
		{LineJoinBevel, "bevel"},
	}

	for _, tt := range tests {
		if got := tt.join.String(); got != tt.want {
			t.Errorf("LineJoin(%d).String() = %q, want %q", tt.join, got, tt.want)
		}
	}
}

func TestGradientKindStrings(t *testing.T) {
	tests := []struct {
		kind GradientKind
		want string
	}{
		{GradientLinear, "linear"},
		{GradientRadial, "radial"},
		{GradientConic, "conic"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("GradientKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestPaintInterface(t *testing.T) {
	// Verify all paint types implement Paint interface
	var _ Paint = Color{}
	var _ Paint = Gradient{}
	var _ Paint = Pattern{}
}
