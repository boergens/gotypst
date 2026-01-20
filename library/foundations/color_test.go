package foundations

import (
	"math"
	"testing"
)

func TestColorSpaceString(t *testing.T) {
	tests := []struct {
		space ColorSpace
		want  string
	}{
		{ColorSpaceLuma, "luma"},
		{ColorSpaceRgb, "rgb"},
		{ColorSpaceLinearRgb, "linear-rgb"},
		{ColorSpaceOklab, "oklab"},
		{ColorSpaceOklch, "oklch"},
		{ColorSpaceHsl, "hsl"},
		{ColorSpaceHsv, "hsv"},
		{ColorSpaceCmyk, "cmyk"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.space.String(); got != tt.want {
				t.Errorf("ColorSpace.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLumaConstructor(t *testing.T) {
	tests := []struct {
		name      string
		lightness float64
		alpha     []float64
		wantL     float64
		wantA     float64
	}{
		{"black", 0, nil, 0, 1},
		{"white", 1, nil, 1, 1},
		{"gray", 0.5, nil, 0.5, 1},
		{"with alpha", 0.5, []float64{0.5}, 0.5, 0.5},
		{"clamp low", -0.5, nil, 0, 1},
		{"clamp high", 1.5, nil, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Luma(tt.lightness, tt.alpha...)
			if c.space != ColorSpaceLuma {
				t.Errorf("space = %v, want %v", c.space, ColorSpaceLuma)
			}
			if c.c[0] != tt.wantL {
				t.Errorf("lightness = %v, want %v", c.c[0], tt.wantL)
			}
			if c.c[3] != tt.wantA {
				t.Errorf("alpha = %v, want %v", c.c[3], tt.wantA)
			}
		})
	}
}

func TestRgbConstructor(t *testing.T) {
	tests := []struct {
		name string
		r, g, b float64
		alpha   []float64
		wantR, wantG, wantB, wantA float64
	}{
		{"black", 0, 0, 0, nil, 0, 0, 0, 1},
		{"white", 1, 1, 1, nil, 1, 1, 1, 1},
		{"red", 1, 0, 0, nil, 1, 0, 0, 1},
		{"with alpha", 1, 0, 0, []float64{0.5}, 1, 0, 0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Rgb(tt.r, tt.g, tt.b, tt.alpha...)
			if c.space != ColorSpaceRgb {
				t.Errorf("space = %v, want %v", c.space, ColorSpaceRgb)
			}
			if c.c[0] != tt.wantR || c.c[1] != tt.wantG || c.c[2] != tt.wantB {
				t.Errorf("rgb = (%v, %v, %v), want (%v, %v, %v)",
					c.c[0], c.c[1], c.c[2], tt.wantR, tt.wantG, tt.wantB)
			}
			if c.c[3] != tt.wantA {
				t.Errorf("alpha = %v, want %v", c.c[3], tt.wantA)
			}
		})
	}
}

func TestRgb8Constructor(t *testing.T) {
	c := Rgb8(255, 128, 0)
	if c.space != ColorSpaceRgb {
		t.Errorf("space = %v, want %v", c.space, ColorSpaceRgb)
	}
	if c.c[0] != 1 {
		t.Errorf("red = %v, want 1", c.c[0])
	}
	if math.Abs(c.c[1]-128.0/255) > 0.001 {
		t.Errorf("green = %v, want ~0.502", c.c[1])
	}
	if c.c[2] != 0 {
		t.Errorf("blue = %v, want 0", c.c[2])
	}
}

func TestRgbHex(t *testing.T) {
	tests := []struct {
		hex     string
		wantR   float64
		wantG   float64
		wantB   float64
		wantA   float64
		wantErr bool
	}{
		{"#ff0000", 1, 0, 0, 1, false},
		{"ff0000", 1, 0, 0, 1, false},
		{"#f00", 1, 0, 0, 1, false},
		{"#00ff00", 0, 1, 0, 1, false},
		{"#0000ff", 0, 0, 1, 1, false},
		{"#ffffff", 1, 1, 1, 1, false},
		{"#000000", 0, 0, 0, 1, false},
		{"#ff000080", 1, 0, 0, 128.0 / 255, false},
		{"#f008", 1, 0, 0, 136.0 / 255, false},
		{"invalid", 0, 0, 0, 0, true},
		{"#gg0000", 0, 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			c, err := RgbHex(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("RgbHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if c.c[0] != tt.wantR {
				t.Errorf("red = %v, want %v", c.c[0], tt.wantR)
			}
			if c.c[1] != tt.wantG {
				t.Errorf("green = %v, want %v", c.c[1], tt.wantG)
			}
			if c.c[2] != tt.wantB {
				t.Errorf("blue = %v, want %v", c.c[2], tt.wantB)
			}
			if math.Abs(c.c[3]-tt.wantA) > 0.01 {
				t.Errorf("alpha = %v, want %v", c.c[3], tt.wantA)
			}
		})
	}
}

func TestHslConstructor(t *testing.T) {
	c := Hsl(0, 1, 0.5) // Red in HSL
	if c.space != ColorSpaceHsl {
		t.Errorf("space = %v, want %v", c.space, ColorSpaceHsl)
	}
	if c.c[0] != 0 {
		t.Errorf("hue = %v, want 0", c.c[0])
	}
	if c.c[1] != 1 {
		t.Errorf("saturation = %v, want 1", c.c[1])
	}
	if c.c[2] != 0.5 {
		t.Errorf("lightness = %v, want 0.5", c.c[2])
	}
}

func TestHsvConstructor(t *testing.T) {
	c := Hsv(0, 1, 1) // Red in HSV
	if c.space != ColorSpaceHsv {
		t.Errorf("space = %v, want %v", c.space, ColorSpaceHsv)
	}
	if c.c[0] != 0 {
		t.Errorf("hue = %v, want 0", c.c[0])
	}
	if c.c[1] != 1 {
		t.Errorf("saturation = %v, want 1", c.c[1])
	}
	if c.c[2] != 1 {
		t.Errorf("value = %v, want 1", c.c[2])
	}
}

func TestCmykConstructor(t *testing.T) {
	c := Cmyk(0, 1, 1, 0) // Red in CMYK
	if c.space != ColorSpaceCmyk {
		t.Errorf("space = %v, want %v", c.space, ColorSpaceCmyk)
	}
	if c.c[0] != 0 {
		t.Errorf("cyan = %v, want 0", c.c[0])
	}
	if c.c[1] != 1 {
		t.Errorf("magenta = %v, want 1", c.c[1])
	}
	if c.c[2] != 1 {
		t.Errorf("yellow = %v, want 1", c.c[2])
	}
	if c.c[3] != 0 {
		t.Errorf("key = %v, want 0", c.c[3])
	}
}

func TestOklabConstructor(t *testing.T) {
	c := Oklab(0.5, 0.1, -0.1)
	if c.space != ColorSpaceOklab {
		t.Errorf("space = %v, want %v", c.space, ColorSpaceOklab)
	}
	if c.c[0] != 0.5 {
		t.Errorf("L = %v, want 0.5", c.c[0])
	}
	if c.c[1] != 0.1 {
		t.Errorf("a = %v, want 0.1", c.c[1])
	}
	if c.c[2] != -0.1 {
		t.Errorf("b = %v, want -0.1", c.c[2])
	}
}

func TestOklchConstructor(t *testing.T) {
	c := Oklch(0.5, 0.2, 30)
	if c.space != ColorSpaceOklch {
		t.Errorf("space = %v, want %v", c.space, ColorSpaceOklch)
	}
	if c.c[0] != 0.5 {
		t.Errorf("L = %v, want 0.5", c.c[0])
	}
	if c.c[1] != 0.2 {
		t.Errorf("chroma = %v, want 0.2", c.c[1])
	}
	if c.c[2] != 30 {
		t.Errorf("hue = %v, want 30", c.c[2])
	}
}

func TestColorToRgb(t *testing.T) {
	tests := []struct {
		name string
		c    *Color
		wantR, wantG, wantB float64
		tolerance float64
	}{
		{"rgb red", Rgb(1, 0, 0), 1, 0, 0, 0.001},
		{"rgb green", Rgb(0, 1, 0), 0, 1, 0, 0.001},
		{"rgb blue", Rgb(0, 0, 1), 0, 0, 1, 0.001},
		{"luma black", Luma(0), 0, 0, 0, 0.001},
		{"luma white", Luma(1), 1, 1, 1, 0.001},
		{"luma gray", Luma(0.5), 0.5, 0.5, 0.5, 0.001},
		{"hsl red", Hsl(0, 1, 0.5), 1, 0, 0, 0.01},
		{"hsl green", Hsl(120, 1, 0.5), 0, 1, 0, 0.01},
		{"hsl blue", Hsl(240, 1, 0.5), 0, 0, 1, 0.01},
		{"hsv red", Hsv(0, 1, 1), 1, 0, 0, 0.01},
		{"hsv green", Hsv(120, 1, 1), 0, 1, 0, 0.01},
		{"hsv blue", Hsv(240, 1, 1), 0, 0, 1, 0.01},
		{"cmyk red", Cmyk(0, 1, 1, 0), 1, 0, 0, 0.01},
		{"cmyk green", Cmyk(1, 0, 1, 0), 0, 1, 0, 0.01},
		{"cmyk blue", Cmyk(1, 1, 0, 0), 0, 0, 1, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rgb := tt.c.ToRgb()
			if math.Abs(rgb.c[0]-tt.wantR) > tt.tolerance {
				t.Errorf("red = %v, want %v", rgb.c[0], tt.wantR)
			}
			if math.Abs(rgb.c[1]-tt.wantG) > tt.tolerance {
				t.Errorf("green = %v, want %v", rgb.c[1], tt.wantG)
			}
			if math.Abs(rgb.c[2]-tt.wantB) > tt.tolerance {
				t.Errorf("blue = %v, want %v", rgb.c[2], tt.wantB)
			}
		})
	}
}

func TestColorRoundTrip(t *testing.T) {
	// Test that converting to a space and back gives the same result
	original := Rgb(0.7, 0.3, 0.5)

	spaces := []ColorSpace{
		ColorSpaceLinearRgb,
		ColorSpaceOklab,
		ColorSpaceOklch,
		ColorSpaceHsl,
		ColorSpaceHsv,
	}

	for _, space := range spaces {
		t.Run(space.String(), func(t *testing.T) {
			converted := original.ToSpace(space)
			back := converted.ToRgb()

			if math.Abs(back.c[0]-original.c[0]) > 0.01 {
				t.Errorf("red: got %v, want %v", back.c[0], original.c[0])
			}
			if math.Abs(back.c[1]-original.c[1]) > 0.01 {
				t.Errorf("green: got %v, want %v", back.c[1], original.c[1])
			}
			if math.Abs(back.c[2]-original.c[2]) > 0.01 {
				t.Errorf("blue: got %v, want %v", back.c[2], original.c[2])
			}
		})
	}
}

func TestColorLighten(t *testing.T) {
	c := Rgb(0.5, 0.5, 0.5)

	// Lighten by 50%
	lighter := c.Lighten(0.5)
	rgb := lighter.ToRgb()

	// Should be lighter than original
	if rgb.c[0] <= c.c[0] {
		t.Errorf("lightened red = %v should be > %v", rgb.c[0], c.c[0])
	}

	// Fully lighten should be near white
	white := c.Lighten(1)
	rgbWhite := white.ToRgb()
	if rgbWhite.c[0] < 0.9 || rgbWhite.c[1] < 0.9 || rgbWhite.c[2] < 0.9 {
		t.Errorf("fully lightened should be near white, got (%v, %v, %v)",
			rgbWhite.c[0], rgbWhite.c[1], rgbWhite.c[2])
	}
}

func TestColorDarken(t *testing.T) {
	c := Rgb(0.5, 0.5, 0.5)

	// Darken by 50%
	darker := c.Darken(0.5)
	rgb := darker.ToRgb()

	// Should be darker than original
	if rgb.c[0] >= c.c[0] {
		t.Errorf("darkened red = %v should be < %v", rgb.c[0], c.c[0])
	}

	// Fully darken should be near black
	black := c.Darken(1)
	rgbBlack := black.ToRgb()
	if rgbBlack.c[0] > 0.1 || rgbBlack.c[1] > 0.1 || rgbBlack.c[2] > 0.1 {
		t.Errorf("fully darkened should be near black, got (%v, %v, %v)",
			rgbBlack.c[0], rgbBlack.c[1], rgbBlack.c[2])
	}
}

func TestColorSaturate(t *testing.T) {
	// Start with a desaturated color
	c := Hsl(0, 0.3, 0.5)

	// Saturate
	saturated := c.Saturate(0.5)
	hsl := saturated.ToHsl()

	// Should have higher saturation
	if hsl.c[1] <= c.c[1] {
		t.Errorf("saturated saturation = %v should be > %v", hsl.c[1], c.c[1])
	}
}

func TestColorDesaturate(t *testing.T) {
	// Start with a saturated color
	c := Hsl(0, 1, 0.5)

	// Desaturate
	desaturated := c.Desaturate(0.5)
	hsl := desaturated.ToHsl()

	// Should have lower saturation
	if hsl.c[1] >= c.c[1] {
		t.Errorf("desaturated saturation = %v should be < %v", hsl.c[1], c.c[1])
	}

	// Fully desaturate should be gray (saturation near 0)
	gray := c.Desaturate(1)
	hslGray := gray.ToHsl()
	if hslGray.c[1] > 0.1 {
		t.Errorf("fully desaturated should have low saturation, got %v", hslGray.c[1])
	}
}

func TestColorRotate(t *testing.T) {
	// Start with red (hue 0) - use HSL space for rotation to avoid Oklch conversion
	c := Hsl(0, 1, 0.5)
	hslSpace := ColorSpaceHsl

	// Rotate 120 degrees to green in HSL space
	rotated := c.Rotate(120, &hslSpace)
	hsl := rotated.ToHsl()

	// Should have hue around 120
	if math.Abs(hsl.c[0]-120) > 1 {
		t.Errorf("rotated hue = %v, want ~120", hsl.c[0])
	}

	// Rotate 240 degrees to blue
	rotated = c.Rotate(240, &hslSpace)
	hsl = rotated.ToHsl()

	// Should have hue around 240
	if math.Abs(hsl.c[0]-240) > 1 {
		t.Errorf("rotated hue = %v, want ~240", hsl.c[0])
	}

	// Test rotation through Oklch (perceptual space - wider tolerance)
	rotatedOklch := c.Rotate(180, nil)
	hslOklch := rotatedOklch.ToHsl()
	// In Oklch, rotation by 180 gives complementary - hue should be ~180
	if math.Abs(hslOklch.c[0]-180) > 30 {
		t.Errorf("oklch rotated hue = %v, want ~180", hslOklch.c[0])
	}
}

func TestColorNegate(t *testing.T) {
	// Test RGB negation directly
	rgbSpace := ColorSpaceRgb
	red := Rgb(1, 0, 0)
	negatedRgb := red.Negate(&rgbSpace)
	rgb := negatedRgb.ToRgb()

	// In RGB negate of red (1,0,0) should be cyan (0,1,1)
	if math.Abs(rgb.c[0]-0) > 0.01 || math.Abs(rgb.c[1]-1) > 0.01 || math.Abs(rgb.c[2]-1) > 0.01 {
		t.Errorf("RGB negated red = (%v, %v, %v), want (0, 1, 1)", rgb.c[0], rgb.c[1], rgb.c[2])
	}

	// Test Oklch negation (default)
	negatedOklch := red.Negate(nil)
	hslOklch := negatedOklch.ToHsl()
	redHsl := red.ToHsl()

	// Negated hue should be ~180 degrees from original in perceptual space
	// This uses Oklch which may not map exactly to HSL hue rotation
	hueDiff := math.Abs(hslOklch.c[0] - redHsl.c[0])
	if hueDiff > 180 {
		hueDiff = 360 - hueDiff
	}
	// Allow wider tolerance for perceptual color space transformations
	if hueDiff < 90 {
		t.Errorf("negated hue diff = %v, want at least 90", hueDiff)
	}
}

func TestColorMix(t *testing.T) {
	red := Rgb(1, 0, 0)
	blue := Rgb(0, 0, 1)

	// Mix 50/50 should give purple-ish
	mixed := red.Mix(blue, 0.5, nil)
	rgb := mixed.ToRgb()

	// Should have both red and blue components
	if rgb.c[0] < 0.2 {
		t.Errorf("mixed should have red component, got %v", rgb.c[0])
	}
	if rgb.c[2] < 0.2 {
		t.Errorf("mixed should have blue component, got %v", rgb.c[2])
	}

	// Mix 0% should be original
	mixed = red.Mix(blue, 0, nil)
	rgb = mixed.ToRgb()
	if math.Abs(rgb.c[0]-1) > 0.1 {
		t.Errorf("0%% mix should be red, got %v", rgb.c[0])
	}

	// Mix 100% should be other
	mixed = red.Mix(blue, 1, nil)
	rgb = mixed.ToRgb()
	if math.Abs(rgb.c[2]-1) > 0.1 {
		t.Errorf("100%% mix should be blue, got %v", rgb.c[2])
	}
}

func TestColorTransparentize(t *testing.T) {
	c := Rgb(1, 0, 0, 1)

	// 50% transparentize
	trans := c.Transparentize(0.5)
	if trans.Alpha() != 0.5 {
		t.Errorf("alpha = %v, want 0.5", trans.Alpha())
	}

	// 100% transparentize
	trans = c.Transparentize(1)
	if trans.Alpha() != 0 {
		t.Errorf("alpha = %v, want 0", trans.Alpha())
	}

	// 0% transparentize
	trans = c.Transparentize(0)
	if trans.Alpha() != 1 {
		t.Errorf("alpha = %v, want 1", trans.Alpha())
	}
}

func TestColorOpacify(t *testing.T) {
	c := Rgb(1, 0, 0, 0.5)

	// 50% opacify
	opaque := c.Opacify(0.5)
	if opaque.Alpha() != 0.75 {
		t.Errorf("alpha = %v, want 0.75", opaque.Alpha())
	}

	// 100% opacify
	opaque = c.Opacify(1)
	if opaque.Alpha() != 1 {
		t.Errorf("alpha = %v, want 1", opaque.Alpha())
	}

	// 0% opacify
	opaque = c.Opacify(0)
	if opaque.Alpha() != 0.5 {
		t.Errorf("alpha = %v, want 0.5", opaque.Alpha())
	}
}

func TestColorToHex(t *testing.T) {
	tests := []struct {
		name string
		c    *Color
		want string
	}{
		{"black", Rgb(0, 0, 0), "#000000"},
		{"white", Rgb(1, 1, 1), "#ffffff"},
		{"red", Rgb(1, 0, 0), "#ff0000"},
		{"green", Rgb(0, 1, 0), "#00ff00"},
		{"blue", Rgb(0, 0, 1), "#0000ff"},
		{"with alpha", Rgb(1, 0, 0, 0.5), "#ff000080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.ToHex()
			if got != tt.want {
				t.Errorf("ToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorComponents(t *testing.T) {
	c := Rgb(0.5, 0.3, 0.7, 0.8)

	// Without alpha
	comp := c.Components(false)
	if len(comp) != 3 {
		t.Errorf("components length = %v, want 3", len(comp))
	}
	if comp[0] != 0.5 || comp[1] != 0.3 || comp[2] != 0.7 {
		t.Errorf("components = %v, want [0.5, 0.3, 0.7]", comp)
	}

	// With alpha
	comp = c.Components(true)
	if len(comp) != 4 {
		t.Errorf("components length = %v, want 4", len(comp))
	}
	if comp[3] != 0.8 {
		t.Errorf("alpha component = %v, want 0.8", comp[3])
	}
}

func TestColorString(t *testing.T) {
	tests := []struct {
		name string
		c    *Color
	}{
		{"rgb", Rgb(1, 0, 0)},
		{"rgba", Rgb(1, 0, 0, 0.5)},
		{"luma", Luma(0.5)},
		{"hsl", Hsl(120, 0.5, 0.5)},
		{"hsv", Hsv(120, 0.5, 0.5)},
		{"cmyk", Cmyk(0, 1, 1, 0)},
		{"oklab", Oklab(0.5, 0.1, 0.1)},
		{"oklch", Oklch(0.5, 0.1, 120)},
		{"linear-rgb", LinearRgb(0.5, 0.5, 0.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.c.String()
			if s == "" {
				t.Error("String() returned empty string")
			}
		})
	}
}

func TestNamedColor(t *testing.T) {
	tests := []struct {
		name  string
		wantR float64
		wantG float64
		wantB float64
	}{
		{"red", 1, 0, 0},
		{"green", 0, 1, 0},
		{"blue", 0, 0, 1},
		{"white", 1, 1, 1},
		{"black", 0, 0, 0},
		{"RED", 1, 0, 0}, // Case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NamedColor(tt.name)
			if c == nil {
				t.Fatal("NamedColor returned nil")
			}
			if c.c[0] != tt.wantR || c.c[1] != tt.wantG || c.c[2] != tt.wantB {
				t.Errorf("got (%v, %v, %v), want (%v, %v, %v)",
					c.c[0], c.c[1], c.c[2], tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}

	// Unknown color
	c := NamedColor("notacolor")
	if c != nil {
		t.Error("NamedColor should return nil for unknown color")
	}
}

func TestMixColors(t *testing.T) {
	// Test mixing in RGB space (direct averaging)
	red := Rgb(1, 0, 0)
	green := Rgb(0, 1, 0)
	blue := Rgb(0, 0, 1)

	rgbSpace := ColorSpaceRgb
	mixedRgb := MixColors([]*Color{red, green, blue}, &rgbSpace)
	if mixedRgb == nil {
		t.Fatal("MixColors returned nil")
	}

	rgb := mixedRgb.ToRgb()
	// Should average to ~(0.33, 0.33, 0.33)
	expected := 1.0 / 3.0
	if math.Abs(rgb.c[0]-expected) > 0.02 || math.Abs(rgb.c[1]-expected) > 0.02 || math.Abs(rgb.c[2]-expected) > 0.02 {
		t.Errorf("RGB mixed should be (~0.33, ~0.33, ~0.33), got (%v, %v, %v)",
			rgb.c[0], rgb.c[1], rgb.c[2])
	}

	// Test mixing two colors in Oklch
	mixedOklch := MixColors([]*Color{red, blue}, nil)
	if mixedOklch == nil {
		t.Fatal("MixColors (Oklch) returned nil")
	}
	// Just verify it produces a result - Oklch mixing produces perceptually uniform results
	rgb = mixedOklch.ToRgb()
	if math.IsNaN(rgb.c[0]) || math.IsNaN(rgb.c[1]) || math.IsNaN(rgb.c[2]) {
		t.Error("MixColors produced NaN values")
	}
}

func TestNilColorMethods(t *testing.T) {
	var c *Color

	// All methods should handle nil gracefully
	if c.ToRgb() != nil {
		t.Error("ToRgb on nil should return nil")
	}
	if c.ToHsl() != nil {
		t.Error("ToHsl on nil should return nil")
	}
	if c.ToHsv() != nil {
		t.Error("ToHsv on nil should return nil")
	}
	if c.Lighten(0.5) != nil {
		t.Error("Lighten on nil should return nil")
	}
	if c.Darken(0.5) != nil {
		t.Error("Darken on nil should return nil")
	}
}

func TestLinearRgbConversion(t *testing.T) {
	// Test sRGB to Linear conversion
	rgb := Rgb(0.5, 0.5, 0.5)
	linear := rgb.ToLinearRgb()

	// sRGB 0.5 should be about 0.214 in linear
	expected := 0.214
	if math.Abs(linear.c[0]-expected) > 0.01 {
		t.Errorf("linear red = %v, want ~%v", linear.c[0], expected)
	}

	// Convert back
	back := linear.ToRgb()
	if math.Abs(back.c[0]-0.5) > 0.01 {
		t.Errorf("back to sRGB = %v, want ~0.5", back.c[0])
	}
}

func TestOklabOklchConversion(t *testing.T) {
	// Create an Oklab color
	oklab := Oklab(0.5, 0.1, 0.1)

	// Convert to Oklch
	oklch := oklab.ToOklch()

	// L should be preserved
	if math.Abs(oklch.c[0]-0.5) > 0.001 {
		t.Errorf("L = %v, want 0.5", oklch.c[0])
	}

	// Chroma should be sqrt(a^2 + b^2)
	expectedC := math.Sqrt(0.1*0.1 + 0.1*0.1)
	if math.Abs(oklch.c[1]-expectedC) > 0.001 {
		t.Errorf("chroma = %v, want %v", oklch.c[1], expectedC)
	}

	// Convert back
	back := oklch.ToOklab()
	if math.Abs(back.c[0]-0.5) > 0.001 {
		t.Errorf("back L = %v, want 0.5", back.c[0])
	}
	if math.Abs(back.c[1]-0.1) > 0.001 {
		t.Errorf("back a = %v, want 0.1", back.c[1])
	}
	if math.Abs(back.c[2]-0.1) > 0.001 {
		t.Errorf("back b = %v, want 0.1", back.c[2])
	}
}

func TestColorType(t *testing.T) {
	c := Rgb(1, 0, 0)
	if c.Type() != "color" {
		t.Errorf("Type() = %v, want 'color'", c.Type())
	}
}

func TestColorSpace(t *testing.T) {
	tests := []struct {
		name  string
		c     *Color
		space ColorSpace
	}{
		{"rgb", Rgb(1, 0, 0), ColorSpaceRgb},
		{"luma", Luma(0.5), ColorSpaceLuma},
		{"hsl", Hsl(0, 1, 0.5), ColorSpaceHsl},
		{"hsv", Hsv(0, 1, 1), ColorSpaceHsv},
		{"cmyk", Cmyk(0, 1, 1, 0), ColorSpaceCmyk},
		{"oklab", Oklab(0.5, 0.1, 0.1), ColorSpaceOklab},
		{"oklch", Oklch(0.5, 0.1, 30), ColorSpaceOklch},
		{"linear-rgb", LinearRgb(0.5, 0.5, 0.5), ColorSpaceLinearRgb},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c.Space() != tt.space {
				t.Errorf("Space() = %v, want %v", tt.c.Space(), tt.space)
			}
		})
	}
}
