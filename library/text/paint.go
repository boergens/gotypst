package text

import (
	"fmt"

	"github.com/boergens/gotypst/layout/inline"
)

// Paint represents a fill or stroke paint (color, gradient, or pattern).
type Paint interface {
	isPaint()
	// String returns a string representation of the paint.
	String() string
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

func (Color) isPaint() {}

// String returns the color as a hex string.
func (c Color) String() string {
	if c.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}

// RGBA returns the color components.
func (c Color) RGBA() (r, g, b, a uint8) {
	return c.R, c.G, c.B, c.A
}

// NewColor creates a new color from RGBA values.
func NewColor(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// NewRGB creates a new opaque color from RGB values.
func NewRGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, A: 255}
}

// ColorFromHex parses a hex color string (e.g., "#ff0000" or "ff0000").
func ColorFromHex(hex string) (Color, error) {
	if len(hex) == 0 {
		return Color{}, fmt.Errorf("empty color string")
	}

	// Strip leading #
	if hex[0] == '#' {
		hex = hex[1:]
	}

	switch len(hex) {
	case 3: // Short form RGB
		r := hexDigit(hex[0]) * 17
		g := hexDigit(hex[1]) * 17
		b := hexDigit(hex[2]) * 17
		return NewRGB(r, g, b), nil
	case 4: // Short form RGBA
		r := hexDigit(hex[0]) * 17
		g := hexDigit(hex[1]) * 17
		b := hexDigit(hex[2]) * 17
		a := hexDigit(hex[3]) * 17
		return NewColor(r, g, b, a), nil
	case 6: // Long form RGB
		r := hexByte(hex[0:2])
		g := hexByte(hex[2:4])
		b := hexByte(hex[4:6])
		return NewRGB(r, g, b), nil
	case 8: // Long form RGBA
		r := hexByte(hex[0:2])
		g := hexByte(hex[2:4])
		b := hexByte(hex[4:6])
		a := hexByte(hex[6:8])
		return NewColor(r, g, b, a), nil
	default:
		return Color{}, fmt.Errorf("invalid hex color: %s", hex)
	}
}

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

func hexByte(s string) uint8 {
	return hexDigit(s[0])*16 + hexDigit(s[1])
}

// Predefined colors.
var (
	Black   = NewRGB(0, 0, 0)
	White   = NewRGB(255, 255, 255)
	Red     = NewRGB(255, 0, 0)
	Green   = NewRGB(0, 128, 0)
	Blue    = NewRGB(0, 0, 255)
	Yellow  = NewRGB(255, 255, 0)
	Cyan    = NewRGB(0, 255, 255)
	Magenta = NewRGB(255, 0, 255)
	Gray    = NewRGB(128, 128, 128)
)

// Gradient represents a color gradient.
type Gradient struct {
	// Kind is the gradient type.
	Kind GradientKind
	// Stops are the color stops.
	Stops []GradientStop
	// Angle is the gradient angle (for linear gradients).
	Angle float64
	// Center is the center point (for radial gradients).
	Center [2]float64
	// Radius is the radius (for radial gradients).
	Radius float64
}

func (Gradient) isPaint() {}

// String returns a string representation.
func (g Gradient) String() string {
	return fmt.Sprintf("gradient.%s(...)", g.Kind)
}

// GradientKind represents the type of gradient.
type GradientKind int

const (
	// GradientLinear is a linear gradient.
	GradientLinear GradientKind = iota
	// GradientRadial is a radial gradient.
	GradientRadial
	// GradientConic is a conic gradient.
	GradientConic
)

// String returns the gradient kind as a string.
func (k GradientKind) String() string {
	switch k {
	case GradientRadial:
		return "radial"
	case GradientConic:
		return "conic"
	default:
		return "linear"
	}
}

// GradientStop represents a color stop in a gradient.
type GradientStop struct {
	// Color is the color at this stop.
	Color Color
	// Offset is the position (0.0 to 1.0).
	Offset float64
}

// Pattern represents a tiling pattern.
type Pattern struct {
	// Size is the pattern tile size.
	Size [2]float64
	// Spacing is the spacing between tiles.
	Spacing [2]float64
}

func (Pattern) isPaint() {}

// String returns a string representation.
func (p Pattern) String() string {
	return "pattern(...)"
}

// Stroke represents stroke styling.
type Stroke struct {
	// Paint is the stroke paint (color, gradient, etc.).
	Paint Paint
	// Thickness is the stroke width.
	Thickness inline.Abs
	// Cap is the line cap style.
	Cap LineCap
	// Join is the line join style.
	Join LineJoin
	// Dash is the dash pattern (nil for solid).
	Dash *StrokeDash
	// MiterLimit is the miter limit for miter joins.
	MiterLimit float64
}

// NewStroke creates a new stroke with defaults.
func NewStroke(paint Paint, thickness inline.Abs) *Stroke {
	return &Stroke{
		Paint:      paint,
		Thickness:  thickness,
		Cap:        LineCapButt,
		Join:       LineJoinMiter,
		MiterLimit: 4.0,
	}
}

// WithCap sets the line cap.
func (s *Stroke) WithCap(cap LineCap) *Stroke {
	s.Cap = cap
	return s
}

// WithJoin sets the line join.
func (s *Stroke) WithJoin(join LineJoin) *Stroke {
	s.Join = join
	return s
}

// WithDash sets the dash pattern.
func (s *Stroke) WithDash(dash *StrokeDash) *Stroke {
	s.Dash = dash
	return s
}

// ToFixedStroke converts to the inline package FixedStroke type.
func (s *Stroke) ToFixedStroke() inline.FixedStroke {
	fs := inline.FixedStroke{
		Paint:     s.Paint,
		Thickness: s.Thickness,
		LineCap:   s.Cap.ToInline(),
		LineJoin:  s.Join.ToInline(),
	}
	if s.Dash != nil {
		fs.DashArray = s.Dash.Array
		fs.DashPhase = s.Dash.Phase
	}
	return fs
}

// StrokeDash represents a dash pattern.
type StrokeDash struct {
	// Array contains the dash lengths.
	Array []inline.Abs
	// Phase is the starting offset into the pattern.
	Phase inline.Abs
}

// NewDash creates a simple dash pattern.
func NewDash(on, off inline.Abs) *StrokeDash {
	return &StrokeDash{
		Array: []inline.Abs{on, off},
		Phase: 0,
	}
}

// LineCap represents line cap styles.
type LineCap int

const (
	// LineCapButt ends lines at their endpoints with no extension.
	LineCapButt LineCap = iota
	// LineCapRound ends lines with a semicircular extension.
	LineCapRound
	// LineCapSquare ends lines with a rectangular extension.
	LineCapSquare
)

// String returns the line cap as a string.
func (c LineCap) String() string {
	switch c {
	case LineCapRound:
		return "round"
	case LineCapSquare:
		return "square"
	default:
		return "butt"
	}
}

// ToInline converts to the inline package LineCap type.
func (c LineCap) ToInline() inline.LineCap {
	switch c {
	case LineCapRound:
		return inline.LineCapRound
	case LineCapSquare:
		return inline.LineCapSquare
	default:
		return inline.LineCapButt
	}
}

// LineJoin represents line join styles.
type LineJoin int

const (
	// LineJoinMiter joins lines with a sharp corner.
	LineJoinMiter LineJoin = iota
	// LineJoinRound joins lines with a rounded corner.
	LineJoinRound
	// LineJoinBevel joins lines with a beveled corner.
	LineJoinBevel
)

// String returns the line join as a string.
func (j LineJoin) String() string {
	switch j {
	case LineJoinRound:
		return "round"
	case LineJoinBevel:
		return "bevel"
	default:
		return "miter"
	}
}

// ToInline converts to the inline package LineJoin type.
func (j LineJoin) ToInline() inline.LineJoin {
	switch j {
	case LineJoinRound:
		return inline.LineJoinRound
	case LineJoinBevel:
		return inline.LineJoinBevel
	default:
		return inline.LineJoinMiter
	}
}
