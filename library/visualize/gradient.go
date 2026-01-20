package visualize

import (
	"fmt"
	"math"
	"sort"
)

// Relative specifies how a gradient or tiling is positioned relative to content.
type Relative int

const (
	// RelativeAuto automatically determines placement.
	RelativeAuto Relative = iota
	// RelativeSelf positions relative to the element's own bounding box.
	RelativeSelf
	// RelativeParent positions relative to the parent's bounding box.
	RelativeParent
)

func (r Relative) String() string {
	switch r {
	case RelativeAuto:
		return "auto"
	case RelativeSelf:
		return "self"
	case RelativeParent:
		return "parent"
	default:
		return fmt.Sprintf("Relative(%d)", r)
	}
}

// GradientStop represents a single color stop in a gradient.
type GradientStop struct {
	// Color is the color at this stop.
	Color *Color
	// Offset is the position of this stop (0.0 to 1.0 for linear/radial, 0 to 360 for conic).
	// nil means the offset should be auto-calculated.
	Offset *float64
}

// NewGradientStop creates a gradient stop with an explicit offset.
func NewGradientStop(color *Color, offset float64) GradientStop {
	return GradientStop{Color: color, Offset: &offset}
}

// NewGradientStopAuto creates a gradient stop with auto-calculated offset.
func NewGradientStopAuto(color *Color) GradientStop {
	return GradientStop{Color: color, Offset: nil}
}

// GradientKind represents the type of gradient.
type GradientKind int

const (
	GradientKindLinear GradientKind = iota
	GradientKindRadial
	GradientKindConic
)

func (k GradientKind) String() string {
	switch k {
	case GradientKindLinear:
		return "linear"
	case GradientKindRadial:
		return "radial"
	case GradientKindConic:
		return "conic"
	default:
		return fmt.Sprintf("GradientKind(%d)", k)
	}
}

// Direction represents a gradient direction for linear gradients.
type Direction int

const (
	// DirectionLTR is left-to-right (default).
	DirectionLTR Direction = iota
	// DirectionRTL is right-to-left.
	DirectionRTL
	// DirectionTTB is top-to-bottom.
	DirectionTTB
	// DirectionBTT is bottom-to-top.
	DirectionBTT
)

func (d Direction) String() string {
	switch d {
	case DirectionLTR:
		return "ltr"
	case DirectionRTL:
		return "rtl"
	case DirectionTTB:
		return "ttb"
	case DirectionBTT:
		return "btt"
	default:
		return fmt.Sprintf("Direction(%d)", d)
	}
}

// ToAngle converts a direction to an angle in radians.
func (d Direction) ToAngle() float64 {
	switch d {
	case DirectionLTR:
		return 0
	case DirectionRTL:
		return math.Pi
	case DirectionTTB:
		return math.Pi / 2
	case DirectionBTT:
		return -math.Pi / 2
	default:
		return 0
	}
}

// Gradient represents a gradient fill.
// This is the base type that can represent linear, radial, or conic gradients.
type Gradient struct {
	// Kind specifies the type of gradient.
	Kind GradientKind

	// Stops contains the color stops.
	Stops []GradientStop

	// Space is the color space for interpolation.
	Space ColorSpace

	// Relative specifies placement relative to self or parent.
	Relative Relative

	// --- Linear gradient fields ---

	// Angle is the direction angle in radians (for linear and conic gradients).
	// For linear: 0 = left-to-right.
	// For conic: starting angle.
	Angle *float64

	// Direction is an alternative to Angle for linear gradients.
	Direction *Direction

	// --- Radial gradient fields ---

	// Center is the center point as (x, y) ratios (0.0 to 1.0).
	// Default: (0.5, 0.5) = center.
	Center *[2]float64

	// Radius is the outer radius as a ratio.
	// Default: 0.5 = 50%.
	Radius *float64

	// FocalCenter is the focal point for radial gradients.
	// Default: same as Center.
	FocalCenter *[2]float64

	// FocalRadius is the focal radius for radial gradients.
	// Default: 0.
	FocalRadius *float64
}

func (*Gradient) valueMarker() {}
func (*Gradient) Type() string { return "gradient" }
func (g *Gradient) String() string {
	if g == nil {
		return "gradient()"
	}

	switch g.Kind {
	case GradientKindLinear:
		return fmt.Sprintf("gradient.linear(%d stops)", len(g.Stops))
	case GradientKindRadial:
		return fmt.Sprintf("gradient.radial(%d stops)", len(g.Stops))
	case GradientKindConic:
		return fmt.Sprintf("gradient.conic(%d stops)", len(g.Stops))
	default:
		return fmt.Sprintf("gradient(%d stops)", len(g.Stops))
	}
}

// --- Linear Gradient Constructor ---

// NewLinearGradient creates a new linear gradient.
func NewLinearGradient(stops []GradientStop, opts ...GradientOption) *Gradient {
	g := &Gradient{
		Kind:     GradientKindLinear,
		Stops:    normalizeStops(stops),
		Space:    ColorSpaceOklab,
		Relative: RelativeAuto,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// NewLinearGradientFromColors creates a linear gradient from colors with evenly spaced stops.
func NewLinearGradientFromColors(colors []*Color, opts ...GradientOption) *Gradient {
	stops := make([]GradientStop, len(colors))
	for i, c := range colors {
		stops[i] = NewGradientStopAuto(c)
	}
	return NewLinearGradient(stops, opts...)
}

// --- Radial Gradient Constructor ---

// NewRadialGradient creates a new radial gradient.
func NewRadialGradient(stops []GradientStop, opts ...GradientOption) *Gradient {
	center := [2]float64{0.5, 0.5}
	radius := 0.5
	focalRadius := 0.0

	g := &Gradient{
		Kind:        GradientKindRadial,
		Stops:       normalizeStops(stops),
		Space:       ColorSpaceOklab,
		Relative:    RelativeAuto,
		Center:      &center,
		Radius:      &radius,
		FocalCenter: nil, // defaults to Center
		FocalRadius: &focalRadius,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// NewRadialGradientFromColors creates a radial gradient from colors with evenly spaced stops.
func NewRadialGradientFromColors(colors []*Color, opts ...GradientOption) *Gradient {
	stops := make([]GradientStop, len(colors))
	for i, c := range colors {
		stops[i] = NewGradientStopAuto(c)
	}
	return NewRadialGradient(stops, opts...)
}

// --- Conic Gradient Constructor ---

// NewConicGradient creates a new conic gradient.
func NewConicGradient(stops []GradientStop, opts ...GradientOption) *Gradient {
	center := [2]float64{0.5, 0.5}
	angle := 0.0

	g := &Gradient{
		Kind:     GradientKindConic,
		Stops:    normalizeStops(stops),
		Space:    ColorSpaceOklab,
		Relative: RelativeAuto,
		Center:   &center,
		Angle:    &angle,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// NewConicGradientFromColors creates a conic gradient from colors with evenly spaced stops.
func NewConicGradientFromColors(colors []*Color, opts ...GradientOption) *Gradient {
	stops := make([]GradientStop, len(colors))
	for i, c := range colors {
		stops[i] = NewGradientStopAuto(c)
	}
	return NewConicGradient(stops, opts...)
}

// --- Gradient Options ---

// GradientOption is a functional option for configuring gradients.
type GradientOption func(*Gradient)

// WithColorSpace sets the color space for interpolation.
func WithColorSpace(space ColorSpace) GradientOption {
	return func(g *Gradient) {
		g.Space = space
	}
}

// WithRelative sets the relative positioning mode.
func WithRelative(rel Relative) GradientOption {
	return func(g *Gradient) {
		g.Relative = rel
	}
}

// WithAngle sets the angle for linear or conic gradients.
func WithAngle(radians float64) GradientOption {
	return func(g *Gradient) {
		g.Angle = &radians
	}
}

// WithAngleDeg sets the angle in degrees for linear or conic gradients.
func WithAngleDeg(degrees float64) GradientOption {
	return func(g *Gradient) {
		radians := degrees * math.Pi / 180
		g.Angle = &radians
	}
}

// WithDirection sets the direction for linear gradients.
func WithDirection(dir Direction) GradientOption {
	return func(g *Gradient) {
		g.Direction = &dir
	}
}

// WithCenter sets the center point for radial or conic gradients.
func WithCenter(x, y float64) GradientOption {
	return func(g *Gradient) {
		center := [2]float64{x, y}
		g.Center = &center
	}
}

// WithRadius sets the radius for radial gradients.
func WithRadius(radius float64) GradientOption {
	return func(g *Gradient) {
		g.Radius = &radius
	}
}

// WithFocalCenter sets the focal center for radial gradients.
func WithFocalCenter(x, y float64) GradientOption {
	return func(g *Gradient) {
		center := [2]float64{x, y}
		g.FocalCenter = &center
	}
}

// WithFocalRadius sets the focal radius for radial gradients.
func WithFocalRadius(radius float64) GradientOption {
	return func(g *Gradient) {
		g.FocalRadius = &radius
	}
}

// --- Gradient Methods ---

// Sample returns the color at the given position.
// For linear/radial: t is a ratio from 0 to 1.
// For conic: t is an angle in radians from 0 to 2*pi.
func (g *Gradient) Sample(t float64) *Color {
	if g == nil || len(g.Stops) == 0 {
		return nil
	}

	// Get normalized stops with computed offsets
	stops := g.normalizedStops()
	if len(stops) == 0 {
		return nil
	}
	if len(stops) == 1 {
		return stops[0].Color
	}

	// For conic gradients, normalize t to 0-1 range
	if g.Kind == GradientKindConic {
		t = math.Mod(t/(2*math.Pi), 1)
		if t < 0 {
			t += 1
		}
	}

	// Clamp t to valid range
	if t <= *stops[0].Offset {
		return stops[0].Color
	}
	if t >= *stops[len(stops)-1].Offset {
		return stops[len(stops)-1].Color
	}

	// Find the two stops to interpolate between
	for i := 0; i < len(stops)-1; i++ {
		if t >= *stops[i].Offset && t <= *stops[i+1].Offset {
			// Compute interpolation factor
			range_ := *stops[i+1].Offset - *stops[i].Offset
			if range_ == 0 {
				return stops[i].Color
			}
			factor := (t - *stops[i].Offset) / range_
			return lerpColor(stops[i].Color, stops[i+1].Color, factor)
		}
	}

	return stops[len(stops)-1].Color
}

// Samples returns colors at multiple positions.
func (g *Gradient) Samples(positions ...float64) []*Color {
	colors := make([]*Color, len(positions))
	for i, t := range positions {
		colors[i] = g.Sample(t)
	}
	return colors
}

// GetStops returns the gradient's color stops.
func (g *Gradient) GetStops() []GradientStop {
	if g == nil {
		return nil
	}
	return g.Stops
}

// GetSpace returns the color space for interpolation.
func (g *Gradient) GetSpace() ColorSpace {
	if g == nil {
		return ColorSpaceOklab
	}
	return g.Space
}

// GetRelative returns the relative positioning mode.
func (g *Gradient) GetRelative() Relative {
	if g == nil {
		return RelativeAuto
	}
	return g.Relative
}

// GetAngle returns the gradient angle in radians.
func (g *Gradient) GetAngle() *float64 {
	if g == nil {
		return nil
	}
	if g.Angle != nil {
		return g.Angle
	}
	if g.Direction != nil {
		angle := g.Direction.ToAngle()
		return &angle
	}
	return nil
}

// GetCenter returns the center point for radial/conic gradients.
func (g *Gradient) GetCenter() *[2]float64 {
	if g == nil {
		return nil
	}
	return g.Center
}

// GetRadius returns the radius for radial gradients.
func (g *Gradient) GetRadius() *float64 {
	if g == nil {
		return nil
	}
	return g.Radius
}

// GetFocalCenter returns the focal center for radial gradients.
func (g *Gradient) GetFocalCenter() *[2]float64 {
	if g == nil {
		return nil
	}
	if g.FocalCenter != nil {
		return g.FocalCenter
	}
	return g.Center // Default to center
}

// GetFocalRadius returns the focal radius for radial gradients.
func (g *Gradient) GetFocalRadius() *float64 {
	if g == nil {
		return nil
	}
	return g.FocalRadius
}

// Sharp creates a version of the gradient with sharp color transitions.
// Steps specifies the number of discrete color bands.
// Smoothness (0-1) controls edge softness.
func (g *Gradient) Sharp(steps int, smoothness float64) *Gradient {
	if g == nil || steps <= 0 {
		return nil
	}

	smoothness = clamp(smoothness, 0, 1)
	sharpStops := make([]GradientStop, 0, steps*2)

	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		nextT := float64(i+1) / float64(steps)
		midT := (t + nextT) / 2

		color := g.Sample(midT)

		// Add two stops at nearly the same position for sharp edges
		offset1 := t
		offset2 := nextT
		if smoothness > 0 && i < steps-1 {
			// Slight overlap for smoothness
			overlap := (nextT - t) * smoothness * 0.1
			offset2 = nextT - overlap
		}

		sharpStops = append(sharpStops, NewGradientStop(color, offset1))
		if i < steps-1 {
			sharpStops = append(sharpStops, NewGradientStop(color, offset2))
		}
	}

	// Clone the gradient with new stops
	result := *g
	result.Stops = sharpStops
	return &result
}

// Repeat creates a repeating version of the gradient.
// Repetitions specifies how many times to repeat.
// Mirror toggles whether alternate repetitions are mirrored.
func (g *Gradient) Repeat(repetitions int, mirror bool) *Gradient {
	if g == nil || repetitions <= 0 {
		return nil
	}

	stops := g.normalizedStops()
	newStops := make([]GradientStop, 0, len(stops)*repetitions)

	for i := 0; i < repetitions; i++ {
		offset := float64(i) / float64(repetitions)
		scale := 1.0 / float64(repetitions)

		// Determine if we should mirror this repetition
		shouldMirror := mirror && i%2 == 1

		for j, stop := range stops {
			var newOffset float64
			if shouldMirror {
				// Reverse the offset within this repetition
				newOffset = offset + (1-*stop.Offset)*scale
			} else {
				newOffset = offset + *stop.Offset*scale
			}

			// Use the appropriate color for mirrored stops
			colorIdx := j
			if shouldMirror {
				colorIdx = len(stops) - 1 - j
			}

			newStops = append(newStops, NewGradientStop(stops[colorIdx].Color, newOffset))
		}
	}

	// Sort stops by offset
	sort.Slice(newStops, func(i, j int) bool {
		return *newStops[i].Offset < *newStops[j].Offset
	})

	result := *g
	result.Stops = newStops
	return &result
}

// --- Helper Functions ---

// normalizeStops returns stops with computed offsets.
func normalizeStops(stops []GradientStop) []GradientStop {
	if len(stops) == 0 {
		return stops
	}

	result := make([]GradientStop, len(stops))
	copy(result, stops)

	// Find stops without explicit offsets and compute them
	// First and last default to 0 and 1
	if result[0].Offset == nil {
		offset := 0.0
		result[0].Offset = &offset
	}
	if len(result) > 1 && result[len(result)-1].Offset == nil {
		offset := 1.0
		result[len(result)-1].Offset = &offset
	}

	// Distribute remaining stops evenly between defined ones
	i := 0
	for i < len(result) {
		if result[i].Offset != nil {
			i++
			continue
		}

		// Find the next defined stop
		j := i + 1
		for j < len(result) && result[j].Offset == nil {
			j++
		}

		// Distribute evenly between i-1 and j
		startOffset := *result[i-1].Offset
		endOffset := *result[j].Offset
		count := j - i + 1

		for k := i; k < j; k++ {
			offset := startOffset + (endOffset-startOffset)*float64(k-i+1)/float64(count)
			result[k].Offset = &offset
		}

		i = j + 1
	}

	return result
}

// normalizedStops returns a copy of stops with all offsets computed.
func (g *Gradient) normalizedStops() []GradientStop {
	return normalizeStops(g.Stops)
}
