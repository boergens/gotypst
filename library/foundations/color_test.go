package foundations

import (
	"math"
	"testing"
)

const epsilon = 0.001

// floatApprox checks if two floats are approximately equal.
func floatApprox(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

// colorApprox checks if two colors are approximately equal.
func rgbaApprox(a, b Rgba, eps float64) bool {
	return floatApprox(a.R, b.R, eps) &&
		floatApprox(a.G, b.G, eps) &&
		floatApprox(a.B, b.B, eps) &&
		floatApprox(a.A, b.A, eps)
}

// --- Luma Tests ---

func TestLumaToRgba(t *testing.T) {
	tests := []struct {
		name string
		luma Luma
		want Rgba
	}{
		{"black", Luma{L: 0, A: 1}, Rgba{R: 0, G: 0, B: 0, A: 1}},
		{"white", Luma{L: 1, A: 1}, Rgba{R: 1, G: 1, B: 1, A: 1}},
		{"gray50", Luma{L: 0.5, A: 1}, Rgba{R: 0.5, G: 0.5, B: 0.5, A: 1}},
		{"transparent", Luma{L: 0.5, A: 0.5}, Rgba{R: 0.5, G: 0.5, B: 0.5, A: 0.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.luma.ToRgba()
			if !rgbaApprox(got, tt.want, epsilon) {
				t.Errorf("Luma.ToRgba() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- RGB <-> Linear RGB Tests ---

func TestRgbaToLinear(t *testing.T) {
	tests := []struct {
		name string
		rgba Rgba
		want LinearRgba
	}{
		{"black", Rgba{R: 0, G: 0, B: 0, A: 1}, LinearRgba{R: 0, G: 0, B: 0, A: 1}},
		{"white", Rgba{R: 1, G: 1, B: 1, A: 1}, LinearRgba{R: 1, G: 1, B: 1, A: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RgbaToLinear(tt.rgba)
			if !floatApprox(got.R, tt.want.R, epsilon) ||
				!floatApprox(got.G, tt.want.G, epsilon) ||
				!floatApprox(got.B, tt.want.B, epsilon) ||
				!floatApprox(got.A, tt.want.A, epsilon) {
				t.Errorf("RgbaToLinear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinearRoundTrip(t *testing.T) {
	// Test that RGB -> Linear -> RGB is identity (approximately)
	colors := []Rgba{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 0.5, G: 0.5, B: 0.5, A: 1},
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 0.2, G: 0.4, B: 0.8, A: 0.9},
	}

	for _, c := range colors {
		linear := RgbaToLinear(c)
		back := linear.ToRgba()
		if !rgbaApprox(c, back, epsilon) {
			t.Errorf("Round trip failed: %v -> %v -> %v", c, linear, back)
		}
	}
}

// --- RGB <-> HSL Tests ---

func TestRgbaToHsl(t *testing.T) {
	tests := []struct {
		name string
		rgba Rgba
		want Hsl
	}{
		{"black", Rgba{R: 0, G: 0, B: 0, A: 1}, Hsl{H: 0, S: 0, L: 0, A: 1}},
		{"white", Rgba{R: 1, G: 1, B: 1, A: 1}, Hsl{H: 0, S: 0, L: 1, A: 1}},
		{"red", Rgba{R: 1, G: 0, B: 0, A: 1}, Hsl{H: 0, S: 1, L: 0.5, A: 1}},
		{"green", Rgba{R: 0, G: 1, B: 0, A: 1}, Hsl{H: 120, S: 1, L: 0.5, A: 1}},
		{"blue", Rgba{R: 0, G: 0, B: 1, A: 1}, Hsl{H: 240, S: 1, L: 0.5, A: 1}},
		{"gray50", Rgba{R: 0.5, G: 0.5, B: 0.5, A: 1}, Hsl{H: 0, S: 0, L: 0.5, A: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RgbaToHsl(tt.rgba)
			if !floatApprox(got.H, tt.want.H, 1) || // Hue has ~1 degree tolerance
				!floatApprox(got.S, tt.want.S, epsilon) ||
				!floatApprox(got.L, tt.want.L, epsilon) ||
				!floatApprox(got.A, tt.want.A, epsilon) {
				t.Errorf("RgbaToHsl(%v) = %v, want %v", tt.rgba, got, tt.want)
			}
		})
	}
}

func TestHslRoundTrip(t *testing.T) {
	// Test that RGB -> HSL -> RGB is identity (approximately)
	colors := []Rgba{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 0.5, G: 0.5, B: 0.5, A: 1},
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 0.2, G: 0.4, B: 0.8, A: 0.9},
		{R: 0.8, G: 0.2, B: 0.6, A: 1},
	}

	for _, c := range colors {
		hsl := RgbaToHsl(c)
		back := hsl.ToRgba()
		if !rgbaApprox(c, back, epsilon) {
			t.Errorf("HSL round trip failed: %v -> %v -> %v", c, hsl, back)
		}
	}
}

// --- RGB <-> HSV Tests ---

func TestRgbaToHsv(t *testing.T) {
	tests := []struct {
		name string
		rgba Rgba
		want Hsv
	}{
		{"black", Rgba{R: 0, G: 0, B: 0, A: 1}, Hsv{H: 0, S: 0, V: 0, A: 1}},
		{"white", Rgba{R: 1, G: 1, B: 1, A: 1}, Hsv{H: 0, S: 0, V: 1, A: 1}},
		{"red", Rgba{R: 1, G: 0, B: 0, A: 1}, Hsv{H: 0, S: 1, V: 1, A: 1}},
		{"green", Rgba{R: 0, G: 1, B: 0, A: 1}, Hsv{H: 120, S: 1, V: 1, A: 1}},
		{"blue", Rgba{R: 0, G: 0, B: 1, A: 1}, Hsv{H: 240, S: 1, V: 1, A: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RgbaToHsv(tt.rgba)
			if !floatApprox(got.H, tt.want.H, 1) || // Hue has ~1 degree tolerance
				!floatApprox(got.S, tt.want.S, epsilon) ||
				!floatApprox(got.V, tt.want.V, epsilon) ||
				!floatApprox(got.A, tt.want.A, epsilon) {
				t.Errorf("RgbaToHsv(%v) = %v, want %v", tt.rgba, got, tt.want)
			}
		})
	}
}

func TestHsvRoundTrip(t *testing.T) {
	// Test that RGB -> HSV -> RGB is identity (approximately)
	colors := []Rgba{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 0.5, G: 0.5, B: 0.5, A: 1},
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 0.2, G: 0.4, B: 0.8, A: 0.9},
		{R: 0.8, G: 0.2, B: 0.6, A: 1},
	}

	for _, c := range colors {
		hsv := RgbaToHsv(c)
		back := hsv.ToRgba()
		if !rgbaApprox(c, back, epsilon) {
			t.Errorf("HSV round trip failed: %v -> %v -> %v", c, hsv, back)
		}
	}
}

// --- RGB <-> Oklab Tests ---

func TestOklabRoundTrip(t *testing.T) {
	// Test that RGB -> Oklab -> RGB is identity (approximately)
	colors := []Rgba{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 0.5, G: 0.5, B: 0.5, A: 1},
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 0.2, G: 0.4, B: 0.8, A: 0.9},
	}

	for _, c := range colors {
		oklab := RgbaToOklab(c)
		back := oklab.ToRgba()
		if !rgbaApprox(c, back, 0.01) { // Oklab conversion has slightly more error
			t.Errorf("Oklab round trip failed: %v -> %v -> %v", c, oklab, back)
		}
	}
}

// --- Oklab <-> Oklch Tests ---

func TestOklabToOklch(t *testing.T) {
	tests := []struct {
		name  string
		oklab Oklab
		wantL float64
		wantC float64
	}{
		{"origin", Oklab{L: 0.5, Ab: 0, Bb: 0, Alpha_: 1}, 0.5, 0},
		{"positive a", Oklab{L: 0.5, Ab: 0.1, Bb: 0, Alpha_: 1}, 0.5, 0.1},
		{"positive b", Oklab{L: 0.5, Ab: 0, Bb: 0.1, Alpha_: 1}, 0.5, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OklabToOklch(tt.oklab)
			if !floatApprox(got.L, tt.wantL, epsilon) ||
				!floatApprox(got.C, tt.wantC, epsilon) {
				t.Errorf("OklabToOklch(%v) = L:%v C:%v, want L:%v C:%v",
					tt.oklab, got.L, got.C, tt.wantL, tt.wantC)
			}
		})
	}
}

func TestOklchRoundTrip(t *testing.T) {
	// Test that RGB -> Oklch -> RGB is identity (approximately)
	colors := []Rgba{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 0.2, G: 0.4, B: 0.8, A: 0.9},
	}

	for _, c := range colors {
		oklch := RgbaToOklch(c)
		back := oklch.ToRgba()
		if !rgbaApprox(c, back, 0.01) { // Oklch conversion has slightly more error
			t.Errorf("Oklch round trip failed: %v -> %v -> %v", c, oklch, back)
		}
	}
}

// --- RGB <-> CMYK Tests ---

func TestRgbaToCmyk(t *testing.T) {
	tests := []struct {
		name string
		rgba Rgba
		want Cmyk
	}{
		{"black", Rgba{R: 0, G: 0, B: 0, A: 1}, Cmyk{C: 0, M: 0, Y: 0, K: 1}},
		{"white", Rgba{R: 1, G: 1, B: 1, A: 1}, Cmyk{C: 0, M: 0, Y: 0, K: 0}},
		{"red", Rgba{R: 1, G: 0, B: 0, A: 1}, Cmyk{C: 0, M: 1, Y: 1, K: 0}},
		{"green", Rgba{R: 0, G: 1, B: 0, A: 1}, Cmyk{C: 1, M: 0, Y: 1, K: 0}},
		{"blue", Rgba{R: 0, G: 0, B: 1, A: 1}, Cmyk{C: 1, M: 1, Y: 0, K: 0}},
		{"cyan", Rgba{R: 0, G: 1, B: 1, A: 1}, Cmyk{C: 1, M: 0, Y: 0, K: 0}},
		{"magenta", Rgba{R: 1, G: 0, B: 1, A: 1}, Cmyk{C: 0, M: 1, Y: 0, K: 0}},
		{"yellow", Rgba{R: 1, G: 1, B: 0, A: 1}, Cmyk{C: 0, M: 0, Y: 1, K: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RgbaToCmyk(tt.rgba)
			if !floatApprox(got.C, tt.want.C, epsilon) ||
				!floatApprox(got.M, tt.want.M, epsilon) ||
				!floatApprox(got.Y, tt.want.Y, epsilon) ||
				!floatApprox(got.K, tt.want.K, epsilon) {
				t.Errorf("RgbaToCmyk(%v) = %v, want %v", tt.rgba, got, tt.want)
			}
		})
	}
}

func TestCmykRoundTrip(t *testing.T) {
	// Test that RGB -> CMYK -> RGB is identity (approximately)
	colors := []Rgba{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 0.2, G: 0.4, B: 0.8, A: 1},
	}

	for _, c := range colors {
		cmyk := RgbaToCmyk(c)
		back := cmyk.ToRgba()
		// CMYK loses alpha information
		expected := Rgba{R: c.R, G: c.G, B: c.B, A: 1}
		if !rgbaApprox(expected, back, epsilon) {
			t.Errorf("CMYK round trip failed: %v -> %v -> %v", c, cmyk, back)
		}
	}
}

// --- RGB <-> Luma Tests ---

func TestRgbaToLuma(t *testing.T) {
	tests := []struct {
		name string
		rgba Rgba
		want Luma
	}{
		{"black", Rgba{R: 0, G: 0, B: 0, A: 1}, Luma{L: 0, A: 1}},
		{"white", Rgba{R: 1, G: 1, B: 1, A: 1}, Luma{L: 1, A: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RgbaToLuma(tt.rgba)
			if !floatApprox(got.L, tt.want.L, epsilon) ||
				!floatApprox(got.A, tt.want.A, epsilon) {
				t.Errorf("RgbaToLuma(%v) = %v, want %v", tt.rgba, got, tt.want)
			}
		})
	}
}

// --- ConvertColor Tests ---

func TestConvertColor(t *testing.T) {
	red := Rgba{R: 1, G: 0, B: 0, A: 1}

	tests := []struct {
		name  string
		color Color
		space string
		check func(Color) bool
	}{
		{"to rgb", red, "rgb", func(c Color) bool { _, ok := c.(Rgba); return ok }},
		{"to hsl", red, "hsl", func(c Color) bool { _, ok := c.(Hsl); return ok }},
		{"to hsv", red, "hsv", func(c Color) bool { _, ok := c.(Hsv); return ok }},
		{"to oklab", red, "oklab", func(c Color) bool { _, ok := c.(Oklab); return ok }},
		{"to oklch", red, "oklch", func(c Color) bool { _, ok := c.(Oklch); return ok }},
		{"to linear-rgb", red, "linear-rgb", func(c Color) bool { _, ok := c.(LinearRgba); return ok }},
		{"to cmyk", red, "cmyk", func(c Color) bool { _, ok := c.(Cmyk); return ok }},
		{"to luma", red, "luma", func(c Color) bool { _, ok := c.(Luma); return ok }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertColor(tt.color, tt.space)
			if err != nil {
				t.Errorf("ConvertColor() error = %v", err)
				return
			}
			if !tt.check(got) {
				t.Errorf("ConvertColor() returned wrong type: %T", got)
			}
		})
	}
}

func TestConvertColorUnknownSpace(t *testing.T) {
	red := Rgba{R: 1, G: 0, B: 0, A: 1}
	_, err := ConvertColor(red, "unknown")
	if err == nil {
		t.Error("ConvertColor() expected error for unknown space")
	}
}

// --- Helper Function Tests ---

func TestClamp01(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0.5, 0.5},
		{0, 0},
		{1, 1},
		{-0.5, 0},
		{1.5, 1},
	}

	for _, tt := range tests {
		got := clamp01(tt.input)
		if got != tt.want {
			t.Errorf("clamp01(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// --- String Representation Tests ---

func TestColorString(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		want  string
	}{
		{"luma no alpha", Luma{L: 0.5, A: 1}, "luma(50%)"},
		{"luma with alpha", Luma{L: 0.5, A: 0.8}, "luma(50%, 80%)"},
		{"rgba no alpha", Rgba{R: 1, G: 0.5, B: 0, A: 1}, "rgb(100%, 50%, 0%)"},
		{"rgba with alpha", Rgba{R: 1, G: 0.5, B: 0, A: 0.8}, "rgb(100%, 50%, 0%, 80%)"},
		{"cmyk", Cmyk{C: 1, M: 0.5, Y: 0, K: 0.2}, "cmyk(100%, 50%, 0%, 20%)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.color.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- ToHex Tests ---

func TestRgbaToHex(t *testing.T) {
	tests := []struct {
		name string
		rgba Rgba
		want string
	}{
		{"black", Rgba{R: 0, G: 0, B: 0, A: 1}, "#000000"},
		{"white", Rgba{R: 1, G: 1, B: 1, A: 1}, "#ffffff"},
		{"red", Rgba{R: 1, G: 0, B: 0, A: 1}, "#ff0000"},
		{"green", Rgba{R: 0, G: 1, B: 0, A: 1}, "#00ff00"},
		{"blue", Rgba{R: 0, G: 0, B: 1, A: 1}, "#0000ff"},
		{"with alpha", Rgba{R: 1, G: 0, B: 0, A: 0.5}, "#ff000080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rgba.ToHex()
			if got != tt.want {
				t.Errorf("ToHex() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- NewRgbaFromBytes Tests ---

func TestNewRgbaFromBytes(t *testing.T) {
	tests := []struct {
		name       string
		r, g, b, a uint8
		want       Rgba
	}{
		{"black", 0, 0, 0, 255, Rgba{R: 0, G: 0, B: 0, A: 1}},
		{"white", 255, 255, 255, 255, Rgba{R: 1, G: 1, B: 1, A: 1}},
		{"red", 255, 0, 0, 255, Rgba{R: 1, G: 0, B: 0, A: 1}},
		{"half alpha", 255, 0, 0, 128, Rgba{R: 1, G: 0, B: 0, A: 128.0 / 255.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRgbaFromBytes(tt.r, tt.g, tt.b, tt.a)
			if !rgbaApprox(got, tt.want, 0.01) {
				t.Errorf("NewRgbaFromBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- ToBytes Round Trip ---

func TestToBytesRoundTrip(t *testing.T) {
	colors := []Rgba{
		{R: 0, G: 0, B: 0, A: 1},
		{R: 1, G: 1, B: 1, A: 1},
		{R: 0.5, G: 0.5, B: 0.5, A: 1},
		{R: 1, G: 0, B: 0, A: 0.5},
	}

	for _, c := range colors {
		r, g, b, a := c.ToBytes()
		back := NewRgbaFromBytes(r, g, b, a)
		if !rgbaApprox(c, back, 0.01) { // Byte conversion has ~1/255 error
			t.Errorf("ToBytes round trip failed: %v -> (%d,%d,%d,%d) -> %v", c, r, g, b, a, back)
		}
	}
}
