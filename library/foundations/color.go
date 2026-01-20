// Color space types and conversions for Typst.
// Translated from visualize/color.rs

package foundations

import (
	"fmt"
	"math"
)

// Color represents a color value in any supported color space.
// This interface is implemented by all color space types.
type Color interface {
	Value
	// colorMarker is an unexported method to seal the interface.
	colorMarker()
	// Space returns the name of the color space.
	Space() string
	// Alpha returns the alpha (opacity) component.
	Alpha() float64
	// ToRgba converts this color to RGBA.
	ToRgba() Rgba
}

// Ensure all color types implement Color.
var (
	_ Color = Luma{}
	_ Color = Rgba{}
	_ Color = LinearRgba{}
	_ Color = Oklab{}
	_ Color = Oklch{}
	_ Color = Hsl{}
	_ Color = Hsv{}
	_ Color = Cmyk{}
)

// ----------------------------------------------------------------------------
// Luma (Grayscale)
// ----------------------------------------------------------------------------

// Luma represents a grayscale color.
type Luma struct {
	// L is the lightness component [0, 1].
	L float64
	// A is the alpha component [0, 1].
	A float64
}

func (Luma) valueMarker() {}
func (Luma) colorMarker() {}
func (Luma) Type() string  { return "color" }
func (Luma) Space() string { return "luma" }
func (c Luma) Alpha() float64 { return c.A }

func (c Luma) String() string {
	if c.A == 1.0 {
		return fmt.Sprintf("luma(%d%%)", int(c.L*100))
	}
	return fmt.Sprintf("luma(%d%%, %d%%)", int(c.L*100), int(c.A*100))
}

// ToRgba converts Luma to RGBA.
func (c Luma) ToRgba() Rgba {
	// Luma is sRGB grayscale
	return Rgba{R: c.L, G: c.L, B: c.L, A: c.A}
}

// NewLuma creates a new Luma color.
func NewLuma(lightness, alpha float64) Luma {
	return Luma{L: clamp01(lightness), A: clamp01(alpha)}
}

// ----------------------------------------------------------------------------
// Rgba (sRGB)
// ----------------------------------------------------------------------------

// Rgba represents a color in the sRGB color space.
type Rgba struct {
	// R, G, B are the red, green, blue components [0, 1].
	R, G, B float64
	// A is the alpha component [0, 1].
	A float64
}

func (Rgba) valueMarker() {}
func (Rgba) colorMarker() {}
func (Rgba) Type() string  { return "color" }
func (Rgba) Space() string { return "rgb" }
func (c Rgba) Alpha() float64 { return c.A }

func (c Rgba) String() string {
	if c.A == 1.0 {
		return fmt.Sprintf("rgb(%d%%, %d%%, %d%%)", int(c.R*100), int(c.G*100), int(c.B*100))
	}
	return fmt.Sprintf("rgb(%d%%, %d%%, %d%%, %d%%)", int(c.R*100), int(c.G*100), int(c.B*100), int(c.A*100))
}

// ToRgba returns itself (identity conversion).
func (c Rgba) ToRgba() Rgba { return c }

// NewRgba creates a new Rgba color.
func NewRgba(r, g, b, a float64) Rgba {
	return Rgba{R: clamp01(r), G: clamp01(g), B: clamp01(b), A: clamp01(a)}
}

// NewRgbaFromBytes creates an Rgba from 0-255 byte values.
func NewRgbaFromBytes(r, g, b, a uint8) Rgba {
	return Rgba{
		R: float64(r) / 255.0,
		G: float64(g) / 255.0,
		B: float64(b) / 255.0,
		A: float64(a) / 255.0,
	}
}

// ToBytes converts to 0-255 byte values.
func (c Rgba) ToBytes() (r, g, b, a uint8) {
	return uint8(c.R*255 + 0.5), uint8(c.G*255 + 0.5), uint8(c.B*255 + 0.5), uint8(c.A*255 + 0.5)
}

// ToHex returns the color as a hex string.
func (c Rgba) ToHex() string {
	r, g, b, a := c.ToBytes()
	if a == 255 {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
}

// ----------------------------------------------------------------------------
// LinearRgba (Linear RGB)
// ----------------------------------------------------------------------------

// LinearRgba represents a color in linear RGB color space.
// This is used for physically correct color operations.
type LinearRgba struct {
	// R, G, B are the red, green, blue components [0, 1].
	R, G, B float64
	// A is the alpha component [0, 1].
	A float64
}

func (LinearRgba) valueMarker() {}
func (LinearRgba) colorMarker() {}
func (LinearRgba) Type() string  { return "color" }
func (LinearRgba) Space() string { return "linear-rgb" }
func (c LinearRgba) Alpha() float64 { return c.A }

func (c LinearRgba) String() string {
	if c.A == 1.0 {
		return fmt.Sprintf("color.linear-rgb(%d%%, %d%%, %d%%)", int(c.R*100), int(c.G*100), int(c.B*100))
	}
	return fmt.Sprintf("color.linear-rgb(%d%%, %d%%, %d%%, %d%%)", int(c.R*100), int(c.G*100), int(c.B*100), int(c.A*100))
}

// ToRgba converts linear RGB to sRGB.
func (c LinearRgba) ToRgba() Rgba {
	return Rgba{
		R: linearToSrgb(c.R),
		G: linearToSrgb(c.G),
		B: linearToSrgb(c.B),
		A: c.A,
	}
}

// NewLinearRgba creates a new LinearRgba color.
func NewLinearRgba(r, g, b, a float64) LinearRgba {
	return LinearRgba{R: clamp01(r), G: clamp01(g), B: clamp01(b), A: clamp01(a)}
}

// ----------------------------------------------------------------------------
// Oklab
// ----------------------------------------------------------------------------

// Oklab represents a color in the Oklab perceptually uniform color space.
type Oklab struct {
	// L is the lightness component [0, 1].
	L float64
	// A is the green-red component (approximately [-0.4, 0.4]).
	Ab float64
	// B is the blue-yellow component (approximately [-0.4, 0.4]).
	Bb float64
	// Alpha is the alpha component [0, 1].
	Alpha_ float64
}

func (Oklab) valueMarker() {}
func (Oklab) colorMarker() {}
func (Oklab) Type() string  { return "color" }
func (Oklab) Space() string { return "oklab" }
func (c Oklab) Alpha() float64 { return c.Alpha_ }

func (c Oklab) String() string {
	if c.Alpha_ == 1.0 {
		return fmt.Sprintf("oklab(%d%%, %.3f, %.3f)", int(c.L*100), c.Ab, c.Bb)
	}
	return fmt.Sprintf("oklab(%d%%, %.3f, %.3f, %d%%)", int(c.L*100), c.Ab, c.Bb, int(c.Alpha_*100))
}

// ToRgba converts Oklab to sRGB.
func (c Oklab) ToRgba() Rgba {
	linear := c.ToLinearRgba()
	return linear.ToRgba()
}

// ToLinearRgba converts Oklab to linear RGB.
func (c Oklab) ToLinearRgba() LinearRgba {
	// Oklab to linear sRGB conversion
	// Using the standard Oklab to linear RGB matrix
	l_ := c.L + 0.3963377774*c.Ab + 0.2158037573*c.Bb
	m_ := c.L - 0.1055613458*c.Ab - 0.0638541728*c.Bb
	s_ := c.L - 0.0894841775*c.Ab - 1.2914855480*c.Bb

	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	return LinearRgba{
		R: clamp01(+4.0767416621*l - 3.3077115913*m + 0.2309699292*s),
		G: clamp01(-1.2684380046*l + 2.6097574011*m - 0.3413193965*s),
		B: clamp01(-0.0041960863*l - 0.7034186147*m + 1.7076147010*s),
		A: c.Alpha_,
	}
}

// NewOklab creates a new Oklab color.
func NewOklab(l, a, b, alpha float64) Oklab {
	return Oklab{L: clamp01(l), Ab: a, Bb: b, Alpha_: clamp01(alpha)}
}

// ----------------------------------------------------------------------------
// Oklch
// ----------------------------------------------------------------------------

// Oklch represents a color in the Oklch color space (cylindrical form of Oklab).
type Oklch struct {
	// L is the lightness component [0, 1].
	L float64
	// C is the chroma component [0, ~0.4].
	C float64
	// H is the hue in degrees [0, 360).
	H float64
	// Alpha is the alpha component [0, 1].
	Alpha_ float64
}

func (Oklch) valueMarker() {}
func (Oklch) colorMarker() {}
func (Oklch) Type() string  { return "color" }
func (Oklch) Space() string { return "oklch" }
func (c Oklch) Alpha() float64 { return c.Alpha_ }

func (c Oklch) String() string {
	if c.Alpha_ == 1.0 {
		return fmt.Sprintf("oklch(%d%%, %.3f, %.1fdeg)", int(c.L*100), c.C, c.H)
	}
	return fmt.Sprintf("oklch(%d%%, %.3f, %.1fdeg, %d%%)", int(c.L*100), c.C, c.H, int(c.Alpha_*100))
}

// ToRgba converts Oklch to sRGB.
func (c Oklch) ToRgba() Rgba {
	return c.ToOklab().ToRgba()
}

// ToOklab converts Oklch to Oklab.
func (c Oklch) ToOklab() Oklab {
	hRad := c.H * math.Pi / 180.0
	return Oklab{
		L:      c.L,
		Ab:     c.C * math.Cos(hRad),
		Bb:     c.C * math.Sin(hRad),
		Alpha_: c.Alpha_,
	}
}

// NewOklch creates a new Oklch color.
func NewOklch(l, c, h, alpha float64) Oklch {
	// Normalize hue to [0, 360)
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return Oklch{L: clamp01(l), C: math.Max(0, c), H: h, Alpha_: clamp01(alpha)}
}

// ----------------------------------------------------------------------------
// Hsl
// ----------------------------------------------------------------------------

// Hsl represents a color in the HSL (Hue, Saturation, Lightness) color space.
type Hsl struct {
	// H is the hue in degrees [0, 360).
	H float64
	// S is the saturation [0, 1].
	S float64
	// L is the lightness [0, 1].
	L float64
	// A is the alpha component [0, 1].
	A float64
}

func (Hsl) valueMarker() {}
func (Hsl) colorMarker() {}
func (Hsl) Type() string  { return "color" }
func (Hsl) Space() string { return "hsl" }
func (c Hsl) Alpha() float64 { return c.A }

func (c Hsl) String() string {
	if c.A == 1.0 {
		return fmt.Sprintf("color.hsl(%.1fdeg, %d%%, %d%%)", c.H, int(c.S*100), int(c.L*100))
	}
	return fmt.Sprintf("color.hsl(%.1fdeg, %d%%, %d%%, %d%%)", c.H, int(c.S*100), int(c.L*100), int(c.A*100))
}

// ToRgba converts HSL to sRGB.
func (c Hsl) ToRgba() Rgba {
	h := c.H / 360.0
	s := c.S
	l := c.L

	if s == 0 {
		// Achromatic (gray)
		return Rgba{R: l, G: l, B: l, A: c.A}
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r := hueToRgb(p, q, h+1.0/3.0)
	g := hueToRgb(p, q, h)
	b := hueToRgb(p, q, h-1.0/3.0)

	return Rgba{R: r, G: g, B: b, A: c.A}
}

// NewHsl creates a new Hsl color.
func NewHsl(h, s, l, a float64) Hsl {
	// Normalize hue to [0, 360)
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return Hsl{H: h, S: clamp01(s), L: clamp01(l), A: clamp01(a)}
}

// ----------------------------------------------------------------------------
// Hsv
// ----------------------------------------------------------------------------

// Hsv represents a color in the HSV (Hue, Saturation, Value) color space.
type Hsv struct {
	// H is the hue in degrees [0, 360).
	H float64
	// S is the saturation [0, 1].
	S float64
	// V is the value [0, 1].
	V float64
	// A is the alpha component [0, 1].
	A float64
}

func (Hsv) valueMarker() {}
func (Hsv) colorMarker() {}
func (Hsv) Type() string  { return "color" }
func (Hsv) Space() string { return "hsv" }
func (c Hsv) Alpha() float64 { return c.A }

func (c Hsv) String() string {
	if c.A == 1.0 {
		return fmt.Sprintf("color.hsv(%.1fdeg, %d%%, %d%%)", c.H, int(c.S*100), int(c.V*100))
	}
	return fmt.Sprintf("color.hsv(%.1fdeg, %d%%, %d%%, %d%%)", c.H, int(c.S*100), int(c.V*100), int(c.A*100))
}

// ToRgba converts HSV to sRGB.
func (c Hsv) ToRgba() Rgba {
	h := c.H / 60.0
	s := c.S
	v := c.V

	if s == 0 {
		// Achromatic (gray)
		return Rgba{R: v, G: v, B: v, A: c.A}
	}

	i := int(h) % 6
	f := h - float64(int(h))
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	var r, g, b float64
	switch i {
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

	return Rgba{R: r, G: g, B: b, A: c.A}
}

// NewHsv creates a new Hsv color.
func NewHsv(h, s, v, a float64) Hsv {
	// Normalize hue to [0, 360)
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return Hsv{H: h, S: clamp01(s), V: clamp01(v), A: clamp01(a)}
}

// ----------------------------------------------------------------------------
// Cmyk
// ----------------------------------------------------------------------------

// Cmyk represents a color in the CMYK (Cyan, Magenta, Yellow, Key/Black) color space.
// This is primarily used for print production.
type Cmyk struct {
	// C, M, Y, K are the cyan, magenta, yellow, key (black) components [0, 1].
	C, M, Y, K float64
}

func (Cmyk) valueMarker() {}
func (Cmyk) colorMarker() {}
func (Cmyk) Type() string  { return "color" }
func (Cmyk) Space() string { return "cmyk" }
func (Cmyk) Alpha() float64 { return 1.0 } // CMYK doesn't have alpha

func (c Cmyk) String() string {
	return fmt.Sprintf("cmyk(%d%%, %d%%, %d%%, %d%%)", int(c.C*100), int(c.M*100), int(c.Y*100), int(c.K*100))
}

// ToRgba converts CMYK to sRGB.
// Note: This is a simple conversion and may not match print output exactly.
func (c Cmyk) ToRgba() Rgba {
	r := (1 - c.C) * (1 - c.K)
	g := (1 - c.M) * (1 - c.K)
	b := (1 - c.Y) * (1 - c.K)
	return Rgba{R: r, G: g, B: b, A: 1.0}
}

// NewCmyk creates a new Cmyk color.
func NewCmyk(c, m, y, k float64) Cmyk {
	return Cmyk{C: clamp01(c), M: clamp01(m), Y: clamp01(y), K: clamp01(k)}
}

// ----------------------------------------------------------------------------
// Conversion Functions: To Linear RGB
// ----------------------------------------------------------------------------

// RgbaToLinear converts sRGB to linear RGB.
func RgbaToLinear(c Rgba) LinearRgba {
	return LinearRgba{
		R: srgbToLinear(c.R),
		G: srgbToLinear(c.G),
		B: srgbToLinear(c.B),
		A: c.A,
	}
}

// LinearRgbaToOklab converts linear RGB to Oklab.
func LinearRgbaToOklab(c LinearRgba) Oklab {
	// Linear sRGB to Oklab conversion
	l := 0.4122214708*c.R + 0.5363325363*c.G + 0.0514459929*c.B
	m := 0.2119034982*c.R + 0.6806995451*c.G + 0.1073969566*c.B
	s := 0.0883024619*c.R + 0.2817188376*c.G + 0.6299787005*c.B

	l_ := math.Cbrt(l)
	m_ := math.Cbrt(m)
	s_ := math.Cbrt(s)

	return Oklab{
		L:      0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_,
		Ab:     1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_,
		Bb:     0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_,
		Alpha_: c.A,
	}
}

// OklabToOklch converts Oklab to Oklch.
func OklabToOklch(c Oklab) Oklch {
	chroma := math.Sqrt(c.Ab*c.Ab + c.Bb*c.Bb)
	hue := math.Atan2(c.Bb, c.Ab) * 180.0 / math.Pi
	if hue < 0 {
		hue += 360
	}
	return Oklch{L: c.L, C: chroma, H: hue, Alpha_: c.Alpha_}
}

// ----------------------------------------------------------------------------
// Conversion Functions: From sRGB
// ----------------------------------------------------------------------------

// RgbaToHsl converts sRGB to HSL.
func RgbaToHsl(c Rgba) Hsl {
	r, g, b := c.R, c.G, c.B

	maxC := math.Max(math.Max(r, g), b)
	minC := math.Min(math.Min(r, g), b)
	l := (maxC + minC) / 2

	if maxC == minC {
		// Achromatic
		return Hsl{H: 0, S: 0, L: l, A: c.A}
	}

	d := maxC - minC
	var s float64
	if l > 0.5 {
		s = d / (2 - maxC - minC)
	} else {
		s = d / (maxC + minC)
	}

	var h float64
	switch maxC {
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

	return Hsl{H: h, S: s, L: l, A: c.A}
}

// RgbaToHsv converts sRGB to HSV.
func RgbaToHsv(c Rgba) Hsv {
	r, g, b := c.R, c.G, c.B

	maxC := math.Max(math.Max(r, g), b)
	minC := math.Min(math.Min(r, g), b)
	d := maxC - minC

	v := maxC

	if maxC == 0 {
		return Hsv{H: 0, S: 0, V: v, A: c.A}
	}

	s := d / maxC

	if d == 0 {
		// Achromatic
		return Hsv{H: 0, S: 0, V: v, A: c.A}
	}

	var h float64
	switch maxC {
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

	return Hsv{H: h, S: s, V: v, A: c.A}
}

// RgbaToOklab converts sRGB to Oklab.
func RgbaToOklab(c Rgba) Oklab {
	linear := RgbaToLinear(c)
	return LinearRgbaToOklab(linear)
}

// RgbaToOklch converts sRGB to Oklch.
func RgbaToOklch(c Rgba) Oklch {
	oklab := RgbaToOklab(c)
	return OklabToOklch(oklab)
}

// RgbaToCmyk converts sRGB to CMYK.
func RgbaToCmyk(c Rgba) Cmyk {
	k := 1 - math.Max(math.Max(c.R, c.G), c.B)
	if k == 1 {
		// Pure black
		return Cmyk{C: 0, M: 0, Y: 0, K: 1}
	}
	return Cmyk{
		C: (1 - c.R - k) / (1 - k),
		M: (1 - c.G - k) / (1 - k),
		Y: (1 - c.B - k) / (1 - k),
		K: k,
	}
}

// RgbaToLuma converts sRGB to Luma (grayscale).
// Uses the standard luminance formula for sRGB.
func RgbaToLuma(c Rgba) Luma {
	// Standard luminance coefficients for sRGB
	l := 0.2126*c.R + 0.7152*c.G + 0.0722*c.B
	return Luma{L: l, A: c.A}
}

// ----------------------------------------------------------------------------
// Generic Conversion
// ----------------------------------------------------------------------------

// ConvertColor converts a color to the specified color space.
func ConvertColor(c Color, space string) (Color, error) {
	switch space {
	case "luma":
		return RgbaToLuma(c.ToRgba()), nil
	case "rgb":
		return c.ToRgba(), nil
	case "linear-rgb":
		return RgbaToLinear(c.ToRgba()), nil
	case "oklab":
		return RgbaToOklab(c.ToRgba()), nil
	case "oklch":
		return RgbaToOklch(c.ToRgba()), nil
	case "hsl":
		return RgbaToHsl(c.ToRgba()), nil
	case "hsv":
		return RgbaToHsv(c.ToRgba()), nil
	case "cmyk":
		return RgbaToCmyk(c.ToRgba()), nil
	default:
		return nil, &OpError{Message: "unknown color space: " + space}
	}
}

// ----------------------------------------------------------------------------
// Helper Functions
// ----------------------------------------------------------------------------

// clamp01 clamps a value to the range [0, 1].
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// srgbToLinear converts an sRGB component to linear.
func srgbToLinear(x float64) float64 {
	if x <= 0.04045 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}

// linearToSrgb converts a linear component to sRGB.
func linearToSrgb(x float64) float64 {
	if x <= 0.0031308 {
		return 12.92 * x
	}
	return 1.055*math.Pow(x, 1/2.4) - 0.055
}

// hueToRgb is a helper function for HSL to RGB conversion.
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
