package visualize

import (
	"testing"
)

func TestNewColor(t *testing.T) {
	c := NewColor(255, 128, 64, 200)

	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 200 {
		t.Errorf("expected (255, 128, 64, 200), got (%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
	}
}

func TestNewColorRGB(t *testing.T) {
	c := NewColorRGB(255, 128, 64)

	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 255 {
		t.Errorf("expected (255, 128, 64, 255), got (%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
	}
}

func TestNewColorFromHex(t *testing.T) {
	tests := []struct {
		hex      string
		expected *Color
		hasError bool
	}{
		{"#ff0000", &Color{255, 0, 0, 255}, false},
		{"ff0000", &Color{255, 0, 0, 255}, false},
		{"#00ff00", &Color{0, 255, 0, 255}, false},
		{"#0000ff", &Color{0, 0, 255, 255}, false},
		{"#ffffff80", &Color{255, 255, 255, 128}, false},
		{"#f00", &Color{255, 0, 0, 255}, false},
		{"#f008", &Color{255, 0, 0, 136}, false},
		{"invalid", nil, true},
		{"#gg0000", nil, true},
	}

	for _, tt := range tests {
		c, err := NewColorFromHex(tt.hex)
		if tt.hasError {
			if err == nil {
				t.Errorf("hex %q: expected error, got none", tt.hex)
			}
			continue
		}
		if err != nil {
			t.Errorf("hex %q: unexpected error: %v", tt.hex, err)
			continue
		}
		if c.R != tt.expected.R || c.G != tt.expected.G || c.B != tt.expected.B || c.A != tt.expected.A {
			t.Errorf("hex %q: expected %v, got %v", tt.hex, tt.expected, c)
		}
	}
}

func TestColorToHex(t *testing.T) {
	tests := []struct {
		color    *Color
		expected string
	}{
		{NewColorRGB(255, 0, 0), "#ff0000"},
		{NewColorRGB(0, 255, 0), "#00ff00"},
		{NewColorRGB(0, 0, 255), "#0000ff"},
		{NewColor(255, 255, 255, 128), "#ffffff80"},
		{nil, "#000000"},
	}

	for _, tt := range tests {
		got := tt.color.ToHex()
		if got != tt.expected {
			t.Errorf("color %v: expected %q, got %q", tt.color, tt.expected, got)
		}
	}
}

func TestColorString(t *testing.T) {
	opaque := NewColorRGB(255, 128, 64)
	if opaque.String() != "rgb(255, 128, 64)" {
		t.Errorf("unexpected string: %s", opaque.String())
	}

	transparent := NewColor(255, 128, 64, 128)
	expected := "rgba(255, 128, 64, 0.50)"
	if transparent.String() != expected {
		t.Errorf("expected %q, got %q", expected, transparent.String())
	}

	var nilColor *Color = nil
	if nilColor.String() != "rgb(0, 0, 0)" {
		t.Errorf("unexpected nil color string: %s", nilColor.String())
	}
}

func TestColorLighten(t *testing.T) {
	red := NewColorRGB(255, 0, 0)

	// No change
	same := red.Lighten(0)
	if same.R != 255 || same.G != 0 || same.B != 0 {
		t.Errorf("lighten(0) should not change color: got %v", same)
	}

	// Full lighten = white
	white := red.Lighten(1)
	if white.R != 255 || white.G != 255 || white.B != 255 {
		t.Errorf("lighten(1) should be white: got %v", white)
	}

	// Partial lighten
	partial := red.Lighten(0.5)
	if partial.R != 255 || partial.G < 100 || partial.G > 130 {
		t.Errorf("lighten(0.5) unexpected result: got %v", partial)
	}
}

func TestColorDarken(t *testing.T) {
	red := NewColorRGB(255, 0, 0)

	// No change
	same := red.Darken(0)
	if same.R != 255 || same.G != 0 || same.B != 0 {
		t.Errorf("darken(0) should not change color: got %v", same)
	}

	// Full darken = black
	black := red.Darken(1)
	if black.R != 0 || black.G != 0 || black.B != 0 {
		t.Errorf("darken(1) should be black: got %v", black)
	}

	// Partial darken
	partial := red.Darken(0.5)
	if partial.R < 120 || partial.R > 135 {
		t.Errorf("darken(0.5) unexpected result: got %v", partial)
	}
}

func TestColorNegate(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	cyan := red.Negate()

	if cyan.R != 0 || cyan.G != 255 || cyan.B != 255 {
		t.Errorf("negate of red should be cyan: got %v", cyan)
	}
}

func TestColorTransparentize(t *testing.T) {
	opaque := NewColorRGB(255, 0, 0)

	// Full transparency
	transparent := opaque.Transparentize(1)
	if transparent.A != 0 {
		t.Errorf("transparentize(1) should be fully transparent: got alpha %d", transparent.A)
	}

	// Partial
	partial := opaque.Transparentize(0.5)
	if partial.A < 120 || partial.A > 135 {
		t.Errorf("transparentize(0.5) unexpected alpha: got %d", partial.A)
	}
}

func TestColorOpacify(t *testing.T) {
	semiTransparent := NewColor(255, 0, 0, 128)

	// Full opacify
	opaque := semiTransparent.Opacify(1)
	if opaque.A != 255 {
		t.Errorf("opacify(1) should be fully opaque: got alpha %d", opaque.A)
	}

	// Partial
	partial := semiTransparent.Opacify(0.5)
	if partial.A < 180 || partial.A > 200 {
		t.Errorf("opacify(0.5) unexpected alpha: got %d", partial.A)
	}
}

func TestColorMix(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	// 0% mix = red
	atRed := red.Mix(blue, 0, ColorSpaceSRGB)
	if atRed.R != 255 || atRed.G != 0 || atRed.B != 0 {
		t.Errorf("mix at 0 should be red: got %v", atRed)
	}

	// 100% mix = blue
	atBlue := red.Mix(blue, 1, ColorSpaceSRGB)
	if atBlue.R != 0 || atBlue.G != 0 || atBlue.B != 255 {
		t.Errorf("mix at 1 should be blue: got %v", atBlue)
	}

	// 50% mix = purple
	purple := red.Mix(blue, 0.5, ColorSpaceSRGB)
	if purple.R < 100 || purple.R > 135 || purple.B < 100 || purple.B > 135 {
		t.Errorf("mix at 0.5 should be purple: got %v", purple)
	}
}

func TestColorComponents(t *testing.T) {
	c := NewColor(128, 64, 255, 128)
	r, g, b, a := c.Components()

	if r < 0.49 || r > 0.51 {
		t.Errorf("expected r ~0.5, got %f", r)
	}
	if g < 0.24 || g > 0.26 {
		t.Errorf("expected g ~0.25, got %f", g)
	}
	if b != 1.0 {
		t.Errorf("expected b 1.0, got %f", b)
	}
	if a < 0.49 || a > 0.51 {
		t.Errorf("expected a ~0.5, got %f", a)
	}
}

func TestColorType(t *testing.T) {
	c := NewColorRGB(0, 0, 0)
	if c.Type() != "color" {
		t.Errorf("expected type 'color', got %q", c.Type())
	}
}

func TestColorSpaceString(t *testing.T) {
	tests := []struct {
		space    ColorSpace
		expected string
	}{
		{ColorSpaceOklab, "oklab"},
		{ColorSpaceSRGB, "srgb"},
		{ColorSpaceLinearRGB, "linear-rgb"},
		{ColorSpaceHSL, "hsl"},
		{ColorSpaceHSV, "hsv"},
		{ColorSpaceOklch, "oklch"},
		{ColorSpaceLuma, "luma"},
		{ColorSpaceCMYK, "cmyk"},
	}

	for _, tt := range tests {
		if tt.space.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.space.String())
		}
	}
}

func TestClamp(t *testing.T) {
	if clamp(-1, 0, 1) != 0 {
		t.Error("clamp should limit to min")
	}
	if clamp(2, 0, 1) != 1 {
		t.Error("clamp should limit to max")
	}
	if clamp(0.5, 0, 1) != 0.5 {
		t.Error("clamp should not change value in range")
	}
}

func TestLerpColor(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	// At t=0, should be red
	c := lerpColor(red, blue, 0)
	if c.R != 255 || c.B != 0 {
		t.Errorf("lerpColor at 0 should be red: got %v", c)
	}

	// At t=1, should be blue
	c = lerpColor(red, blue, 1)
	if c.R != 0 || c.B != 255 {
		t.Errorf("lerpColor at 1 should be blue: got %v", c)
	}

	// At t=0.5, should be in between
	c = lerpColor(red, blue, 0.5)
	if c.R < 120 || c.R > 135 || c.B < 120 || c.B > 135 {
		t.Errorf("lerpColor at 0.5 unexpected: got %v", c)
	}
}
