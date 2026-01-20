package visualize

import (
	"fmt"
	"math"
)

// ColorSpace represents a color space for color representation and interpolation.
type ColorSpace int

const (
	// ColorSpaceOklab is the perceptually uniform Oklab color space (default).
	ColorSpaceOklab ColorSpace = iota
	// ColorSpaceSRGB is the standard sRGB color space.
	ColorSpaceSRGB
	// ColorSpaceLinearRGB is the linear RGB color space.
	ColorSpaceLinearRGB
	// ColorSpaceHSL is the hue-saturation-lightness color space.
	ColorSpaceHSL
	// ColorSpaceHSV is the hue-saturation-value color space.
	ColorSpaceHSV
	// ColorSpaceOklch is the Oklch polar color space.
	ColorSpaceOklch
	// ColorSpaceLuma is the grayscale color space.
	ColorSpaceLuma
	// ColorSpaceCMYK is the cyan-magenta-yellow-key color space.
	ColorSpaceCMYK
)

func (cs ColorSpace) String() string {
	switch cs {
	case ColorSpaceOklab:
		return "oklab"
	case ColorSpaceSRGB:
		return "srgb"
	case ColorSpaceLinearRGB:
		return "linear-rgb"
	case ColorSpaceHSL:
		return "hsl"
	case ColorSpaceHSV:
		return "hsv"
	case ColorSpaceOklch:
		return "oklch"
	case ColorSpaceLuma:
		return "luma"
	case ColorSpaceCMYK:
		return "cmyk"
	default:
		return fmt.Sprintf("ColorSpace(%d)", cs)
	}
}

// Color represents a color value in RGBA format.
// Components are stored as 8-bit values (0-255).
type Color struct {
	R, G, B, A uint8
}

func (*Color) valueMarker() {}
func (*Color) Type() string { return "color" }
func (c *Color) String() string {
	if c == nil {
		return "rgb(0, 0, 0)"
	}
	if c.A == 255 {
		return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", c.R, c.G, c.B, float64(c.A)/255)
}

// NewColor creates a new color from RGBA components (0-255).
func NewColor(r, g, b, a uint8) *Color {
	return &Color{R: r, G: g, B: b, A: a}
}

// NewColorRGB creates a new opaque color from RGB components (0-255).
func NewColorRGB(r, g, b uint8) *Color {
	return &Color{R: r, G: g, B: b, A: 255}
}

// NewColorFromHex creates a color from a hex string (e.g., "#ff0000" or "ff0000").
func NewColorFromHex(hex string) (*Color, error) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(hex) {
	case 3: // RGB shorthand
		_, err := fmt.Sscanf(hex, "%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", hex)
		}
		r, g, b = r*17, g*17, b*17
	case 4: // RGBA shorthand
		_, err := fmt.Sscanf(hex, "%1x%1x%1x%1x", &r, &g, &b, &a)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", hex)
		}
		r, g, b, a = r*17, g*17, b*17, a*17
	case 6: // RGB
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", hex)
		}
	case 8: // RGBA
		_, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", hex)
		}
	default:
		return nil, fmt.Errorf("invalid hex color length: %s", hex)
	}

	return &Color{R: r, G: g, B: b, A: a}, nil
}

// ToHex returns the color as a hex string (e.g., "#ff0000" or "#ff000080" with alpha).
func (c *Color) ToHex() string {
	if c == nil {
		return "#000000"
	}
	if c.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}

// Lighten returns a lightened version of the color.
// Amount is a ratio from 0 (no change) to 1 (fully white).
func (c *Color) Lighten(amount float64) *Color {
	if c == nil {
		return nil
	}
	amount = clamp(amount, 0, 1)
	return &Color{
		R: uint8(float64(c.R) + (255-float64(c.R))*amount),
		G: uint8(float64(c.G) + (255-float64(c.G))*amount),
		B: uint8(float64(c.B) + (255-float64(c.B))*amount),
		A: c.A,
	}
}

// Darken returns a darkened version of the color.
// Amount is a ratio from 0 (no change) to 1 (fully black).
func (c *Color) Darken(amount float64) *Color {
	if c == nil {
		return nil
	}
	amount = clamp(amount, 0, 1)
	return &Color{
		R: uint8(float64(c.R) * (1 - amount)),
		G: uint8(float64(c.G) * (1 - amount)),
		B: uint8(float64(c.B) * (1 - amount)),
		A: c.A,
	}
}

// Negate returns the color negative.
func (c *Color) Negate() *Color {
	if c == nil {
		return nil
	}
	return &Color{
		R: 255 - c.R,
		G: 255 - c.G,
		B: 255 - c.B,
		A: c.A,
	}
}

// Transparentize returns a more transparent version of the color.
// Amount is a ratio from 0 (no change) to 1 (fully transparent).
func (c *Color) Transparentize(amount float64) *Color {
	if c == nil {
		return nil
	}
	amount = clamp(amount, 0, 1)
	return &Color{
		R: c.R,
		G: c.G,
		B: c.B,
		A: uint8(float64(c.A) * (1 - amount)),
	}
}

// Opacify returns a more opaque version of the color.
// Amount is a ratio from 0 (no change) to 1 (fully opaque).
func (c *Color) Opacify(amount float64) *Color {
	if c == nil {
		return nil
	}
	amount = clamp(amount, 0, 1)
	return &Color{
		R: c.R,
		G: c.G,
		B: c.B,
		A: uint8(float64(c.A) + (255-float64(c.A))*amount),
	}
}

// Mix blends this color with another color.
// Ratio determines the mix: 0 = this color, 1 = other color.
func (c *Color) Mix(other *Color, ratio float64, space ColorSpace) *Color {
	if c == nil || other == nil {
		return nil
	}
	ratio = clamp(ratio, 0, 1)

	// Simple RGB mixing (TODO: implement proper color space interpolation)
	return &Color{
		R: uint8(float64(c.R)*(1-ratio) + float64(other.R)*ratio),
		G: uint8(float64(c.G)*(1-ratio) + float64(other.G)*ratio),
		B: uint8(float64(c.B)*(1-ratio) + float64(other.B)*ratio),
		A: uint8(float64(c.A)*(1-ratio) + float64(other.A)*ratio),
	}
}

// Components returns the color components as floats (0-1 range).
func (c *Color) Components() (r, g, b, a float64) {
	if c == nil {
		return 0, 0, 0, 0
	}
	return float64(c.R) / 255, float64(c.G) / 255, float64(c.B) / 255, float64(c.A) / 255
}

// clamp restricts a value to a range.
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// lerpColor linearly interpolates between two colors.
func lerpColor(a, b *Color, t float64) *Color {
	if a == nil || b == nil {
		return nil
	}
	t = clamp(t, 0, 1)
	return &Color{
		R: uint8(math.Round(float64(a.R)*(1-t) + float64(b.R)*t)),
		G: uint8(math.Round(float64(a.G)*(1-t) + float64(b.G)*t)),
		B: uint8(math.Round(float64(a.B)*(1-t) + float64(b.B)*t)),
		A: uint8(math.Round(float64(a.A)*(1-t) + float64(b.A)*t)),
	}
}
