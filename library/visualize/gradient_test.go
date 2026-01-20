package visualize

import (
	"math"
	"testing"
)

func TestNewLinearGradient(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	stops := []GradientStop{
		NewGradientStopAuto(red),
		NewGradientStopAuto(blue),
	}

	g := NewLinearGradient(stops)

	if g.Kind != GradientKindLinear {
		t.Errorf("expected linear gradient, got %v", g.Kind)
	}
	if len(g.Stops) != 2 {
		t.Errorf("expected 2 stops, got %d", len(g.Stops))
	}
	if g.Space != ColorSpaceOklab {
		t.Errorf("expected oklab color space, got %v", g.Space)
	}
	if g.Relative != RelativeAuto {
		t.Errorf("expected auto relative, got %v", g.Relative)
	}
}

func TestLinearGradientFromColors(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	green := NewColorRGB(0, 255, 0)
	blue := NewColorRGB(0, 0, 255)

	g := NewLinearGradientFromColors([]*Color{red, green, blue})

	if len(g.Stops) != 3 {
		t.Errorf("expected 3 stops, got %d", len(g.Stops))
	}
}

func TestGradientWithOptions(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	g := NewLinearGradientFromColors([]*Color{red, blue},
		WithColorSpace(ColorSpaceSRGB),
		WithRelative(RelativeSelf),
		WithAngleDeg(45),
	)

	if g.Space != ColorSpaceSRGB {
		t.Errorf("expected srgb color space, got %v", g.Space)
	}
	if g.Relative != RelativeSelf {
		t.Errorf("expected self relative, got %v", g.Relative)
	}
	if g.Angle == nil {
		t.Error("expected angle to be set")
	} else if math.Abs(*g.Angle-math.Pi/4) > 0.001 {
		t.Errorf("expected 45deg (%.4f rad), got %.4f", math.Pi/4, *g.Angle)
	}
}

func TestRadialGradient(t *testing.T) {
	white := NewColorRGB(255, 255, 255)
	black := NewColorRGB(0, 0, 0)

	g := NewRadialGradientFromColors([]*Color{white, black},
		WithCenter(0.25, 0.75),
		WithRadius(0.8),
	)

	if g.Kind != GradientKindRadial {
		t.Errorf("expected radial gradient, got %v", g.Kind)
	}
	if g.Center == nil {
		t.Error("expected center to be set")
	} else {
		if g.Center[0] != 0.25 || g.Center[1] != 0.75 {
			t.Errorf("expected center (0.25, 0.75), got (%.2f, %.2f)", g.Center[0], g.Center[1])
		}
	}
	if g.Radius == nil || *g.Radius != 0.8 {
		t.Errorf("expected radius 0.8, got %v", g.Radius)
	}
}

func TestConicGradient(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	yellow := NewColorRGB(255, 255, 0)
	green := NewColorRGB(0, 255, 0)
	cyan := NewColorRGB(0, 255, 255)
	blue := NewColorRGB(0, 0, 255)
	magenta := NewColorRGB(255, 0, 255)

	g := NewConicGradientFromColors([]*Color{red, yellow, green, cyan, blue, magenta},
		WithAngleDeg(90),
	)

	if g.Kind != GradientKindConic {
		t.Errorf("expected conic gradient, got %v", g.Kind)
	}
	if len(g.Stops) != 6 {
		t.Errorf("expected 6 stops, got %d", len(g.Stops))
	}
}

func TestGradientSample(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	g := NewLinearGradientFromColors([]*Color{red, blue})

	// Sample at the start
	c := g.Sample(0)
	if c.R != 255 || c.G != 0 || c.B != 0 {
		t.Errorf("expected red at t=0, got rgb(%d, %d, %d)", c.R, c.G, c.B)
	}

	// Sample at the end
	c = g.Sample(1)
	if c.R != 0 || c.G != 0 || c.B != 255 {
		t.Errorf("expected blue at t=1, got rgb(%d, %d, %d)", c.R, c.G, c.B)
	}

	// Sample in the middle
	c = g.Sample(0.5)
	// Should be purple-ish (interpolated)
	if c.R < 100 || c.R > 150 || c.B < 100 || c.B > 150 {
		t.Errorf("expected purple-ish at t=0.5, got rgb(%d, %d, %d)", c.R, c.G, c.B)
	}
}

func TestGradientSamples(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	g := NewLinearGradientFromColors([]*Color{red, blue})

	colors := g.Samples(0, 0.25, 0.5, 0.75, 1.0)
	if len(colors) != 5 {
		t.Errorf("expected 5 colors, got %d", len(colors))
	}

	// First should be red
	if colors[0].R != 255 {
		t.Errorf("expected first color to be red, got rgb(%d, %d, %d)", colors[0].R, colors[0].G, colors[0].B)
	}
	// Last should be blue
	if colors[4].B != 255 {
		t.Errorf("expected last color to be blue, got rgb(%d, %d, %d)", colors[4].R, colors[4].G, colors[4].B)
	}
}

func TestGradientRepeat(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	g := NewLinearGradientFromColors([]*Color{red, blue})
	repeated := g.Repeat(2, false)

	if repeated == nil {
		t.Fatal("expected repeated gradient")
	}
	if len(repeated.Stops) != 4 {
		t.Errorf("expected 4 stops after 2x repeat, got %d", len(repeated.Stops))
	}
}

func TestGradientSharp(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	g := NewLinearGradientFromColors([]*Color{red, blue})
	sharp := g.Sharp(3, 0)

	if sharp == nil {
		t.Fatal("expected sharp gradient")
	}
	// 3 steps should create approximately 5 stops (2 per step minus overlap)
	if len(sharp.Stops) < 3 {
		t.Errorf("expected at least 3 stops, got %d", len(sharp.Stops))
	}
}

func TestGradientStopNormalization(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	green := NewColorRGB(0, 255, 0)
	blue := NewColorRGB(0, 0, 255)

	// Create stops without explicit offsets
	stops := []GradientStop{
		NewGradientStopAuto(red),
		NewGradientStopAuto(green),
		NewGradientStopAuto(blue),
	}

	normalized := normalizeStops(stops)

	if *normalized[0].Offset != 0 {
		t.Errorf("expected first offset 0, got %f", *normalized[0].Offset)
	}
	if *normalized[2].Offset != 1 {
		t.Errorf("expected last offset 1, got %f", *normalized[2].Offset)
	}
	if *normalized[1].Offset != 0.5 {
		t.Errorf("expected middle offset 0.5, got %f", *normalized[1].Offset)
	}
}

func TestGradientMixedStops(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	green := NewColorRGB(0, 255, 0)
	blue := NewColorRGB(0, 0, 255)

	// Mix explicit and auto offsets
	stops := []GradientStop{
		NewGradientStop(red, 0),
		NewGradientStopAuto(green), // Should be distributed
		NewGradientStop(blue, 1),
	}

	normalized := normalizeStops(stops)

	if *normalized[1].Offset != 0.5 {
		t.Errorf("expected middle offset 0.5, got %f", *normalized[1].Offset)
	}
}

func TestDirectionToAngle(t *testing.T) {
	tests := []struct {
		dir      Direction
		expected float64
	}{
		{DirectionLTR, 0},
		{DirectionRTL, math.Pi},
		{DirectionTTB, math.Pi / 2},
		{DirectionBTT, -math.Pi / 2},
	}

	for _, tt := range tests {
		got := tt.dir.ToAngle()
		if math.Abs(got-tt.expected) > 0.001 {
			t.Errorf("Direction %v: expected angle %.4f, got %.4f", tt.dir, tt.expected, got)
		}
	}
}

func TestGradientString(t *testing.T) {
	red := NewColorRGB(255, 0, 0)
	blue := NewColorRGB(0, 0, 255)

	linear := NewLinearGradientFromColors([]*Color{red, blue})
	if linear.String() != "gradient.linear(2 stops)" {
		t.Errorf("unexpected string: %s", linear.String())
	}

	radial := NewRadialGradientFromColors([]*Color{red, blue})
	if radial.String() != "gradient.radial(2 stops)" {
		t.Errorf("unexpected string: %s", radial.String())
	}

	conic := NewConicGradientFromColors([]*Color{red, blue})
	if conic.String() != "gradient.conic(2 stops)" {
		t.Errorf("unexpected string: %s", conic.String())
	}
}
