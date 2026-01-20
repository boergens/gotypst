package foundations

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// ColorSpace identifies which color space a color is stored in.
type ColorSpace int

const (
	ColorSpaceLuma ColorSpace = iota
	ColorSpaceRgb
	ColorSpaceLinearRgb
	ColorSpaceOklab
	ColorSpaceOklch
	ColorSpaceHsl
	ColorSpaceHsv
	ColorSpaceCmyk
)

// String returns the name of the color space.
func (cs ColorSpace) String() string {
	switch cs {
	case ColorSpaceLuma:
		return "luma"
	case ColorSpaceRgb:
		return "rgb"
	case ColorSpaceLinearRgb:
		return "linear-rgb"
	case ColorSpaceOklab:
		return "oklab"
	case ColorSpaceOklch:
		return "oklch"
	case ColorSpaceHsl:
		return "hsl"
	case ColorSpaceHsv:
		return "hsv"
	case ColorSpaceCmyk:
		return "cmyk"
	default:
		return fmt.Sprintf("ColorSpace(%d)", cs)
	}
}

// Color represents a color value in Typst.
// Colors are stored in their native color space with components as float64
// values typically in the range [0, 1] (except for Oklab/Oklch which use
// different ranges).
type Color struct {
	space ColorSpace
	// Components are stored in a fixed array.
	// The interpretation depends on the color space:
	// - Luma: [lightness, _, _, alpha]
	// - Rgb/LinearRgb: [red, green, blue, alpha]
	// - Oklab: [L, a, b, alpha]
	// - Oklch: [L, chroma, hue, alpha]
	// - Hsl: [hue, saturation, lightness, alpha]
	// - Hsv: [hue, saturation, value, alpha]
	// - Cmyk: [cyan, magenta, yellow, key] (no alpha)
	c [4]float64
}

// Ensure Color implements Value.
var _ Value = (*Color)(nil)

func (*Color) valueMarker() {}
func (*Color) Type() string { return "color" }

// String returns a string representation of the color.
func (c *Color) String() string {
	if c == nil {
		return "rgb(0, 0, 0)"
	}
	switch c.space {
	case ColorSpaceLuma:
		if c.c[3] == 1 {
			return fmt.Sprintf("luma(%d%%)", int(c.c[0]*100))
		}
		return fmt.Sprintf("luma(%d%%, %d%%)", int(c.c[0]*100), int(c.c[3]*100))
	case ColorSpaceRgb:
		r, g, b := int(c.c[0]*255+0.5), int(c.c[1]*255+0.5), int(c.c[2]*255+0.5)
		if c.c[3] == 1 {
			return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
		}
		return fmt.Sprintf("rgb(%d, %d, %d, %d%%)", r, g, b, int(c.c[3]*100))
	case ColorSpaceLinearRgb:
		if c.c[3] == 1 {
			return fmt.Sprintf("color.linear-rgb(%d%%, %d%%, %d%%)",
				int(c.c[0]*100), int(c.c[1]*100), int(c.c[2]*100))
		}
		return fmt.Sprintf("color.linear-rgb(%d%%, %d%%, %d%%, %d%%)",
			int(c.c[0]*100), int(c.c[1]*100), int(c.c[2]*100), int(c.c[3]*100))
	case ColorSpaceOklab:
		if c.c[3] == 1 {
			return fmt.Sprintf("oklab(%.1f%%, %.3f, %.3f)",
				c.c[0]*100, c.c[1], c.c[2])
		}
		return fmt.Sprintf("oklab(%.1f%%, %.3f, %.3f, %d%%)",
			c.c[0]*100, c.c[1], c.c[2], int(c.c[3]*100))
	case ColorSpaceOklch:
		if c.c[3] == 1 {
			return fmt.Sprintf("oklch(%.1f%%, %.3f, %.1fdeg)",
				c.c[0]*100, c.c[1], c.c[2])
		}
		return fmt.Sprintf("oklch(%.1f%%, %.3f, %.1fdeg, %d%%)",
			c.c[0]*100, c.c[1], c.c[2], int(c.c[3]*100))
	case ColorSpaceHsl:
		if c.c[3] == 1 {
			return fmt.Sprintf("color.hsl(%.1fdeg, %d%%, %d%%)",
				c.c[0], int(c.c[1]*100), int(c.c[2]*100))
		}
		return fmt.Sprintf("color.hsl(%.1fdeg, %d%%, %d%%, %d%%)",
			c.c[0], int(c.c[1]*100), int(c.c[2]*100), int(c.c[3]*100))
	case ColorSpaceHsv:
		if c.c[3] == 1 {
			return fmt.Sprintf("color.hsv(%.1fdeg, %d%%, %d%%)",
				c.c[0], int(c.c[1]*100), int(c.c[2]*100))
		}
		return fmt.Sprintf("color.hsv(%.1fdeg, %d%%, %d%%, %d%%)",
			c.c[0], int(c.c[1]*100), int(c.c[2]*100), int(c.c[3]*100))
	case ColorSpaceCmyk:
		return fmt.Sprintf("cmyk(%d%%, %d%%, %d%%, %d%%)",
			int(c.c[0]*100), int(c.c[1]*100), int(c.c[2]*100), int(c.c[3]*100))
	default:
		return fmt.Sprintf("color(%v)", c.c)
	}
}

// Space returns the color space of this color.
func (c *Color) Space() ColorSpace {
	if c == nil {
		return ColorSpaceRgb
	}
	return c.space
}

// Alpha returns the alpha component (0-1). CMYK colors return 1.
func (c *Color) Alpha() float64 {
	if c == nil {
		return 1
	}
	if c.space == ColorSpaceCmyk {
		return 1 // CMYK has no alpha
	}
	return c.c[3]
}

// Components returns the color components.
// If alpha is true, includes alpha as last component (except CMYK).
func (c *Color) Components(alpha bool) []float64 {
	if c == nil {
		return []float64{0, 0, 0, 1}
	}
	switch c.space {
	case ColorSpaceLuma:
		if alpha {
			return []float64{c.c[0], c.c[3]}
		}
		return []float64{c.c[0]}
	case ColorSpaceCmyk:
		// CMYK has no alpha
		return []float64{c.c[0], c.c[1], c.c[2], c.c[3]}
	default:
		if alpha {
			return []float64{c.c[0], c.c[1], c.c[2], c.c[3]}
		}
		return []float64{c.c[0], c.c[1], c.c[2]}
	}
}

// --- Color Constructors ---

// Luma creates a grayscale color.
// lightness and alpha are in [0, 1].
func Luma(lightness float64, alpha ...float64) *Color {
	a := 1.0
	if len(alpha) > 0 {
		a = clamp01(alpha[0])
	}
	return &Color{
		space: ColorSpaceLuma,
		c:     [4]float64{clamp01(lightness), 0, 0, a},
	}
}

// Rgb creates an sRGB color from components in [0, 1].
func Rgb(r, g, b float64, alpha ...float64) *Color {
	a := 1.0
	if len(alpha) > 0 {
		a = clamp01(alpha[0])
	}
	return &Color{
		space: ColorSpaceRgb,
		c:     [4]float64{clamp01(r), clamp01(g), clamp01(b), a},
	}
}

// Rgb8 creates an sRGB color from 8-bit components (0-255).
func Rgb8(r, g, b uint8, alpha ...float64) *Color {
	return Rgb(float64(r)/255, float64(g)/255, float64(b)/255, alpha...)
}

// hexPattern matches CSS hex color strings.
var hexPattern = regexp.MustCompile(`^#?([0-9a-fA-F]{3,8})$`)

// RgbHex creates an sRGB color from a hex string.
// Supports formats: #RGB, #RGBA, #RRGGBB, #RRGGBBAA
func RgbHex(hex string) (*Color, error) {
	matches := hexPattern.FindStringSubmatch(hex)
	if matches == nil {
		return nil, &OpError{Message: fmt.Sprintf("invalid hex color: %s", hex)}
	}

	hex = matches[1]
	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(hex) {
	case 3: // RGB
		r = hexDigit(hex[0]) * 17
		g = hexDigit(hex[1]) * 17
		b = hexDigit(hex[2]) * 17
	case 4: // RGBA
		r = hexDigit(hex[0]) * 17
		g = hexDigit(hex[1]) * 17
		b = hexDigit(hex[2]) * 17
		a = hexDigit(hex[3]) * 17
	case 6: // RRGGBB
		r = hexDigit(hex[0])*16 + hexDigit(hex[1])
		g = hexDigit(hex[2])*16 + hexDigit(hex[3])
		b = hexDigit(hex[4])*16 + hexDigit(hex[5])
	case 8: // RRGGBBAA
		r = hexDigit(hex[0])*16 + hexDigit(hex[1])
		g = hexDigit(hex[2])*16 + hexDigit(hex[3])
		b = hexDigit(hex[4])*16 + hexDigit(hex[5])
		a = hexDigit(hex[6])*16 + hexDigit(hex[7])
	default:
		return nil, &OpError{Message: fmt.Sprintf("invalid hex color length: %s", hex)}
	}

	return &Color{
		space: ColorSpaceRgb,
		c:     [4]float64{float64(r) / 255, float64(g) / 255, float64(b) / 255, float64(a) / 255},
	}, nil
}

// hexDigit converts a hex character to its value.
func hexDigit(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	default:
		return 0
	}
}

// LinearRgb creates a linear RGB color.
func LinearRgb(r, g, b float64, alpha ...float64) *Color {
	a := 1.0
	if len(alpha) > 0 {
		a = clamp01(alpha[0])
	}
	return &Color{
		space: ColorSpaceLinearRgb,
		c:     [4]float64{clamp01(r), clamp01(g), clamp01(b), a},
	}
}

// Oklab creates an Oklab color.
// L is in [0, 1], a and b are in approximately [-0.4, 0.4].
func Oklab(l, a, b float64, alpha ...float64) *Color {
	al := 1.0
	if len(alpha) > 0 {
		al = clamp01(alpha[0])
	}
	return &Color{
		space: ColorSpaceOklab,
		c:     [4]float64{clamp01(l), a, b, al},
	}
}

// Oklch creates an Oklch color.
// L is in [0, 1], chroma >= 0, hue is in degrees [0, 360).
func Oklch(l, chroma, hue float64, alpha ...float64) *Color {
	a := 1.0
	if len(alpha) > 0 {
		a = clamp01(alpha[0])
	}
	return &Color{
		space: ColorSpaceOklch,
		c:     [4]float64{clamp01(l), math.Max(0, chroma), normalizeDegrees(hue), a},
	}
}

// Hsl creates an HSL color.
// Hue is in degrees [0, 360), saturation and lightness are in [0, 1].
func Hsl(h, s, l float64, alpha ...float64) *Color {
	a := 1.0
	if len(alpha) > 0 {
		a = clamp01(alpha[0])
	}
	return &Color{
		space: ColorSpaceHsl,
		c:     [4]float64{normalizeDegrees(h), clamp01(s), clamp01(l), a},
	}
}

// Hsv creates an HSV color.
// Hue is in degrees [0, 360), saturation and value are in [0, 1].
func Hsv(h, s, v float64, alpha ...float64) *Color {
	a := 1.0
	if len(alpha) > 0 {
		a = clamp01(alpha[0])
	}
	return &Color{
		space: ColorSpaceHsv,
		c:     [4]float64{normalizeDegrees(h), clamp01(s), clamp01(v), a},
	}
}

// Cmyk creates a CMYK color.
// All components are in [0, 1].
func Cmyk(c, m, y, k float64) *Color {
	return &Color{
		space: ColorSpaceCmyk,
		c:     [4]float64{clamp01(c), clamp01(m), clamp01(y), clamp01(k)},
	}
}

// --- Color Manipulation Methods ---

// Lighten increases the lightness of the color by the given factor.
// Factor is in [0, 1], where 1 means fully white.
func (c *Color) Lighten(factor float64) *Color {
	if c == nil {
		return nil
	}
	factor = clamp01(factor)

	// Convert to Oklch for perceptually uniform lightening
	oklch := c.ToOklch()
	// Increase L towards 1
	oklch.c[0] = oklch.c[0] + (1-oklch.c[0])*factor

	// Convert back to original space
	return oklch.ToSpace(c.space)
}

// Darken decreases the lightness of the color by the given factor.
// Factor is in [0, 1], where 1 means fully black.
func (c *Color) Darken(factor float64) *Color {
	if c == nil {
		return nil
	}
	factor = clamp01(factor)

	// Convert to Oklch for perceptually uniform darkening
	oklch := c.ToOklch()
	// Decrease L towards 0
	oklch.c[0] = oklch.c[0] * (1 - factor)

	return oklch.ToSpace(c.space)
}

// Saturate increases the saturation of the color by the given factor.
// Factor is in [0, 1], where 1 doubles the saturation.
func (c *Color) Saturate(factor float64) *Color {
	if c == nil {
		return nil
	}
	factor = clamp01(factor)

	// Convert to Oklch where chroma represents saturation
	oklch := c.ToOklch()
	// Increase chroma
	oklch.c[1] = oklch.c[1] * (1 + factor)

	return oklch.ToSpace(c.space)
}

// Desaturate decreases the saturation of the color by the given factor.
// Factor is in [0, 1], where 1 means fully desaturated (gray).
func (c *Color) Desaturate(factor float64) *Color {
	if c == nil {
		return nil
	}
	factor = clamp01(factor)

	// Convert to Oklch
	oklch := c.ToOklch()
	// Decrease chroma towards 0
	oklch.c[1] = oklch.c[1] * (1 - factor)

	return oklch.ToSpace(c.space)
}

// Negate returns the complementary color.
// If space is nil, uses Oklch for perceptually accurate negation.
func (c *Color) Negate(space *ColorSpace) *Color {
	if c == nil {
		return nil
	}

	targetSpace := ColorSpaceOklch
	if space != nil {
		targetSpace = *space
	}

	converted := c.ToSpace(targetSpace)

	switch targetSpace {
	case ColorSpaceRgb, ColorSpaceLinearRgb:
		converted.c[0] = 1 - converted.c[0]
		converted.c[1] = 1 - converted.c[1]
		converted.c[2] = 1 - converted.c[2]
	case ColorSpaceOklab:
		converted.c[1] = -converted.c[1]
		converted.c[2] = -converted.c[2]
	case ColorSpaceOklch:
		// In Oklch, hue is in index 2
		converted.c[2] = normalizeDegrees(converted.c[2] + 180)
	case ColorSpaceHsl, ColorSpaceHsv:
		// In HSL/HSV, hue is in index 0
		converted.c[0] = normalizeDegrees(converted.c[0] + 180)
	case ColorSpaceLuma:
		converted.c[0] = 1 - converted.c[0]
	case ColorSpaceCmyk:
		converted.c[0] = 1 - converted.c[0]
		converted.c[1] = 1 - converted.c[1]
		converted.c[2] = 1 - converted.c[2]
		converted.c[3] = 1 - converted.c[3]
	}

	return converted.ToSpace(c.space)
}

// Rotate rotates the hue of the color by the given angle in degrees.
// If space is nil, uses Oklch.
func (c *Color) Rotate(angle float64, space *ColorSpace) *Color {
	if c == nil {
		return nil
	}

	targetSpace := ColorSpaceOklch
	if space != nil {
		targetSpace = *space
	}

	converted := c.ToSpace(targetSpace)

	switch targetSpace {
	case ColorSpaceOklch:
		converted.c[2] = normalizeDegrees(converted.c[2] + angle)
	case ColorSpaceHsl, ColorSpaceHsv:
		converted.c[0] = normalizeDegrees(converted.c[0] + angle)
	default:
		// For non-cylindrical spaces, convert to Oklch, rotate, convert back
		oklch := converted.ToOklch()
		oklch.c[2] = normalizeDegrees(oklch.c[2] + angle)
		converted = oklch.ToSpace(targetSpace)
	}

	return converted.ToSpace(c.space)
}

// Mix blends this color with another color.
// Ratio is in [0, 1], where 0 is fully this color and 1 is fully the other.
// If space is nil, uses Oklch for perceptually uniform mixing.
func (c *Color) Mix(other *Color, ratio float64, space *ColorSpace) *Color {
	if c == nil || other == nil {
		return c
	}
	ratio = clamp01(ratio)

	targetSpace := ColorSpaceOklch
	if space != nil {
		targetSpace = *space
	}

	c1 := c.ToSpace(targetSpace)
	c2 := other.ToSpace(targetSpace)

	// Linear interpolation in the target space
	result := &Color{
		space: targetSpace,
		c: [4]float64{
			lerp(c1.c[0], c2.c[0], ratio),
			lerp(c1.c[1], c2.c[1], ratio),
			lerp(c1.c[2], c2.c[2], ratio),
			lerp(c1.c[3], c2.c[3], ratio),
		},
	}

	// Special handling for hue interpolation in cylindrical spaces
	if targetSpace == ColorSpaceOklch || targetSpace == ColorSpaceHsl || targetSpace == ColorSpaceHsv {
		hueIdx := 0
		if targetSpace == ColorSpaceOklch {
			hueIdx = 2
		}
		result.c[hueIdx] = lerpAngle(c1.c[hueIdx], c2.c[hueIdx], ratio)
	}

	return result.ToSpace(c.space)
}

// Transparentize reduces the alpha by a relative amount.
// Scale is in [0, 1], where 1 makes the color fully transparent.
func (c *Color) Transparentize(scale float64) *Color {
	if c == nil {
		return nil
	}
	scale = clamp01(scale)

	result := *c
	if c.space != ColorSpaceCmyk {
		result.c[3] = c.c[3] * (1 - scale)
	}
	return &result
}

// Opacify increases the alpha by a relative amount.
// Scale is in [0, 1], where 1 makes the color fully opaque.
func (c *Color) Opacify(scale float64) *Color {
	if c == nil {
		return nil
	}
	scale = clamp01(scale)

	result := *c
	if c.space != ColorSpaceCmyk {
		result.c[3] = c.c[3] + (1-c.c[3])*scale
	}
	return &result
}

// ToHex returns the color as an RGB hex string.
func (c *Color) ToHex() string {
	if c == nil {
		return "#000000"
	}

	rgb := c.ToRgb()
	r := uint8(math.Round(rgb.c[0] * 255))
	g := uint8(math.Round(rgb.c[1] * 255))
	b := uint8(math.Round(rgb.c[2] * 255))

	if rgb.c[3] < 1 {
		a := uint8(math.Round(rgb.c[3] * 255))
		return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// --- Color Space Conversions ---

// ToSpace converts the color to the specified color space.
func (c *Color) ToSpace(space ColorSpace) *Color {
	if c == nil {
		return nil
	}
	if c.space == space {
		// Return a copy
		result := *c
		return &result
	}

	switch space {
	case ColorSpaceLuma:
		return c.ToLuma()
	case ColorSpaceRgb:
		return c.ToRgb()
	case ColorSpaceLinearRgb:
		return c.ToLinearRgb()
	case ColorSpaceOklab:
		return c.ToOklab()
	case ColorSpaceOklch:
		return c.ToOklch()
	case ColorSpaceHsl:
		return c.ToHsl()
	case ColorSpaceHsv:
		return c.ToHsv()
	case ColorSpaceCmyk:
		return c.ToCmyk()
	default:
		return c.ToRgb()
	}
}

// ToLuma converts the color to grayscale (Luma).
func (c *Color) ToLuma() *Color {
	if c == nil {
		return nil
	}
	if c.space == ColorSpaceLuma {
		result := *c
		return &result
	}

	// Convert to linear RGB first for accurate luminance calculation
	linear := c.ToLinearRgb()
	// ITU-R BT.709 luminance coefficients
	l := 0.2126*linear.c[0] + 0.7152*linear.c[1] + 0.0722*linear.c[2]

	return &Color{
		space: ColorSpaceLuma,
		c:     [4]float64{l, 0, 0, c.Alpha()},
	}
}

// ToRgb converts the color to sRGB.
func (c *Color) ToRgb() *Color {
	if c == nil {
		return nil
	}
	if c.space == ColorSpaceRgb {
		result := *c
		return &result
	}

	switch c.space {
	case ColorSpaceLuma:
		// Grayscale to RGB
		v := c.c[0]
		return &Color{
			space: ColorSpaceRgb,
			c:     [4]float64{v, v, v, c.c[3]},
		}

	case ColorSpaceLinearRgb:
		return &Color{
			space: ColorSpaceRgb,
			c: [4]float64{
				linearToSrgb(c.c[0]),
				linearToSrgb(c.c[1]),
				linearToSrgb(c.c[2]),
				c.c[3],
			},
		}

	case ColorSpaceOklab:
		return c.oklabToLinearRgb().ToRgb()

	case ColorSpaceOklch:
		return c.ToOklab().ToRgb()

	case ColorSpaceHsl:
		return hslToRgb(c.c[0], c.c[1], c.c[2], c.c[3])

	case ColorSpaceHsv:
		return hsvToRgb(c.c[0], c.c[1], c.c[2], c.c[3])

	case ColorSpaceCmyk:
		return cmykToRgb(c.c[0], c.c[1], c.c[2], c.c[3])

	default:
		return &Color{space: ColorSpaceRgb, c: [4]float64{0, 0, 0, 1}}
	}
}

// ToLinearRgb converts the color to linear RGB.
func (c *Color) ToLinearRgb() *Color {
	if c == nil {
		return nil
	}
	if c.space == ColorSpaceLinearRgb {
		result := *c
		return &result
	}

	// Convert via sRGB
	rgb := c.ToRgb()
	return &Color{
		space: ColorSpaceLinearRgb,
		c: [4]float64{
			srgbToLinear(rgb.c[0]),
			srgbToLinear(rgb.c[1]),
			srgbToLinear(rgb.c[2]),
			rgb.c[3],
		},
	}
}

// ToOklab converts the color to Oklab.
func (c *Color) ToOklab() *Color {
	if c == nil {
		return nil
	}
	if c.space == ColorSpaceOklab {
		result := *c
		return &result
	}

	if c.space == ColorSpaceOklch {
		// Convert from Oklch to Oklab
		L, C, h := c.c[0], c.c[1], c.c[2]*math.Pi/180 // Convert hue to radians
		a := C * math.Cos(h)
		b := C * math.Sin(h)
		return &Color{
			space: ColorSpaceOklab,
			c:     [4]float64{L, a, b, c.c[3]},
		}
	}

	// Convert via linear RGB
	linear := c.ToLinearRgb()
	return linear.linearRgbToOklab()
}

// ToOklch converts the color to Oklch.
func (c *Color) ToOklch() *Color {
	if c == nil {
		return nil
	}
	if c.space == ColorSpaceOklch {
		result := *c
		return &result
	}

	oklab := c.ToOklab()
	L, a, b := oklab.c[0], oklab.c[1], oklab.c[2]

	C := math.Sqrt(a*a + b*b)
	h := math.Atan2(b, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}

	return &Color{
		space: ColorSpaceOklch,
		c:     [4]float64{L, C, h, oklab.c[3]},
	}
}

// ToHsl converts the color to HSL.
func (c *Color) ToHsl() *Color {
	if c == nil {
		return nil
	}
	if c.space == ColorSpaceHsl {
		result := *c
		return &result
	}

	rgb := c.ToRgb()
	return rgbToHsl(rgb.c[0], rgb.c[1], rgb.c[2], rgb.c[3])
}

// ToHsv converts the color to HSV.
func (c *Color) ToHsv() *Color {
	if c == nil {
		return nil
	}
	if c.space == ColorSpaceHsv {
		result := *c
		return &result
	}

	rgb := c.ToRgb()
	return rgbToHsv(rgb.c[0], rgb.c[1], rgb.c[2], rgb.c[3])
}

// ToCmyk converts the color to CMYK.
func (c *Color) ToCmyk() *Color {
	if c == nil {
		return nil
	}
	if c.space == ColorSpaceCmyk {
		result := *c
		return &result
	}

	rgb := c.ToRgb()
	return rgbToCmyk(rgb.c[0], rgb.c[1], rgb.c[2])
}

// --- Internal Conversion Functions ---

// srgbToLinear converts an sRGB component to linear RGB.
func srgbToLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// linearToSrgb converts a linear RGB component to sRGB.
func linearToSrgb(v float64) float64 {
	if v <= 0.0031308 {
		return v * 12.92
	}
	return 1.055*math.Pow(v, 1/2.4) - 0.055
}

// linearRgbToOklab converts linear RGB to Oklab.
func (c *Color) linearRgbToOklab() *Color {
	r, g, b := c.c[0], c.c[1], c.c[2]

	// Linear RGB to LMS
	l := 0.4122214708*r + 0.5363325363*g + 0.0514459929*b
	m := 0.2119034982*r + 0.6806995451*g + 0.1073969566*b
	s := 0.0883024619*r + 0.2817188376*g + 0.6299787005*b

	// Cube root
	l_ := math.Cbrt(l)
	m_ := math.Cbrt(m)
	s_ := math.Cbrt(s)

	// LMS to Oklab
	L := 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_
	a := 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_
	bb := 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_

	return &Color{
		space: ColorSpaceOklab,
		c:     [4]float64{L, a, bb, c.c[3]},
	}
}

// oklabToLinearRgb converts Oklab to linear RGB.
func (c *Color) oklabToLinearRgb() *Color {
	L, a, b := c.c[0], c.c[1], c.c[2]

	// Oklab to LMS
	l_ := L + 0.3963377774*a + 0.2158037573*b
	m_ := L - 0.1055613458*a - 0.0638541728*b
	s_ := L - 0.0894841775*a - 1.2914855480*b

	// Cube
	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	// LMS to linear RGB
	r := 4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	g := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	bb := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s

	return &Color{
		space: ColorSpaceLinearRgb,
		c:     [4]float64{clamp01(r), clamp01(g), clamp01(bb), c.c[3]},
	}
}

// hslToRgb converts HSL to RGB.
func hslToRgb(h, s, l, a float64) *Color {
	if s == 0 {
		return &Color{
			space: ColorSpaceRgb,
			c:     [4]float64{l, l, l, a},
		}
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r := hueToRgb(p, q, h/360+1.0/3.0)
	g := hueToRgb(p, q, h/360)
	b := hueToRgb(p, q, h/360-1.0/3.0)

	return &Color{
		space: ColorSpaceRgb,
		c:     [4]float64{r, g, b, a},
	}
}

func hueToRgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 0.5 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// rgbToHsl converts RGB to HSL.
func rgbToHsl(r, g, b, a float64) *Color {
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	l := (max + min) / 2

	if max == min {
		return &Color{
			space: ColorSpaceHsl,
			c:     [4]float64{0, 0, l, a},
		}
	}

	d := max - min
	var s float64
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	var h float64
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h *= 60

	return &Color{
		space: ColorSpaceHsl,
		c:     [4]float64{h, s, l, a},
	}
}

// hsvToRgb converts HSV to RGB.
func hsvToRgb(h, s, v, a float64) *Color {
	if s == 0 {
		return &Color{
			space: ColorSpaceRgb,
			c:     [4]float64{v, v, v, a},
		}
	}

	h = h / 60
	i := math.Floor(h)
	f := h - i
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	var r, g, b float64
	switch int(i) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	case 5:
		r, g, b = v, p, q
	}

	return &Color{
		space: ColorSpaceRgb,
		c:     [4]float64{r, g, b, a},
	}
}

// rgbToHsv converts RGB to HSV.
func rgbToHsv(r, g, b, a float64) *Color {
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	v := max
	d := max - min

	var s float64
	if max != 0 {
		s = d / max
	}

	if max == min {
		return &Color{
			space: ColorSpaceHsv,
			c:     [4]float64{0, s, v, a},
		}
	}

	var h float64
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h *= 60

	return &Color{
		space: ColorSpaceHsv,
		c:     [4]float64{h, s, v, a},
	}
}

// cmykToRgb converts CMYK to RGB.
func cmykToRgb(c, m, y, k float64) *Color {
	r := (1 - c) * (1 - k)
	g := (1 - m) * (1 - k)
	b := (1 - y) * (1 - k)
	return &Color{
		space: ColorSpaceRgb,
		c:     [4]float64{r, g, b, 1},
	}
}

// rgbToCmyk converts RGB to CMYK.
func rgbToCmyk(r, g, b float64) *Color {
	k := 1 - math.Max(math.Max(r, g), b)
	if k == 1 {
		return &Color{
			space: ColorSpaceCmyk,
			c:     [4]float64{0, 0, 0, 1},
		}
	}
	c := (1 - r - k) / (1 - k)
	m := (1 - g - k) / (1 - k)
	y := (1 - b - k) / (1 - k)
	return &Color{
		space: ColorSpaceCmyk,
		c:     [4]float64{c, m, y, k},
	}
}

// --- Helper Functions ---

// clamp01 clamps a value to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// normalizeDegrees normalizes an angle to [0, 360).
func normalizeDegrees(deg float64) float64 {
	deg = math.Mod(deg, 360)
	if deg < 0 {
		deg += 360
	}
	return deg
}

// lerp performs linear interpolation.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// lerpAngle performs linear interpolation on angles in degrees,
// taking the shortest path around the circle.
func lerpAngle(a, b, t float64) float64 {
	diff := b - a
	// Take the shortest path
	if diff > 180 {
		diff -= 360
	} else if diff < -180 {
		diff += 360
	}
	return normalizeDegrees(a + diff*t)
}

// --- Named Colors (CSS Colors) ---

// namedColors maps CSS color names to their hex values.
var namedColors = map[string]string{
	"black":   "#000000",
	"white":   "#ffffff",
	"red":     "#ff0000",
	"green":   "#00ff00",
	"blue":    "#0000ff",
	"yellow":  "#ffff00",
	"cyan":    "#00ffff",
	"magenta": "#ff00ff",
	"gray":    "#808080",
	"grey":    "#808080",
	"silver":  "#c0c0c0",
	"maroon":  "#800000",
	"olive":   "#808000",
	"lime":    "#00ff00",
	"aqua":    "#00ffff",
	"teal":    "#008080",
	"navy":    "#000080",
	"fuchsia": "#ff00ff",
	"purple":  "#800080",
	"orange":  "#ffa500",
}

// NamedColor returns a color by its CSS name, or nil if not found.
func NamedColor(name string) *Color {
	hex, ok := namedColors[strings.ToLower(name)]
	if !ok {
		return nil
	}
	c, _ := RgbHex(hex)
	return c
}

// MixColors blends multiple colors together.
// Space is optional; if nil, uses Oklch.
func MixColors(colors []*Color, space *ColorSpace) *Color {
	if len(colors) == 0 {
		return nil
	}
	if len(colors) == 1 {
		result := *colors[0]
		return &result
	}

	targetSpace := ColorSpaceOklch
	if space != nil {
		targetSpace = *space
	}

	// Convert all colors to the target space
	converted := make([]*Color, len(colors))
	for i, c := range colors {
		converted[i] = c.ToSpace(targetSpace)
	}

	// Average the components
	var sum [4]float64
	for _, c := range converted {
		sum[0] += c.c[0]
		sum[1] += c.c[1]
		sum[2] += c.c[2]
		sum[3] += c.c[3]
	}

	n := float64(len(colors))
	result := &Color{
		space: targetSpace,
		c:     [4]float64{sum[0] / n, sum[1] / n, sum[2] / n, sum[3] / n},
	}

	// For cylindrical spaces, use proper angle averaging for hue
	if targetSpace == ColorSpaceOklch || targetSpace == ColorSpaceHsl || targetSpace == ColorSpaceHsv {
		hueIdx := 0
		if targetSpace == ColorSpaceOklch {
			hueIdx = 2
		}

		// Average hue using circular mean
		var sinSum, cosSum float64
		for _, c := range converted {
			rad := c.c[hueIdx] * math.Pi / 180
			sinSum += math.Sin(rad)
			cosSum += math.Cos(rad)
		}
		avgHue := math.Atan2(sinSum, cosSum) * 180 / math.Pi
		if avgHue < 0 {
			avgHue += 360
		}
		result.c[hueIdx] = avgHue
	}

	return result
}
