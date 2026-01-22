// Measurement value types for Typst.
// Length, Angle, Ratio, Relative, Fraction types.

package foundations

// Length represents a physical length value.
type Length struct {
	// Points is the length in typographic points (1/72 inch).
	Points float64
}

// LengthValue represents a length as a Value.
type LengthValue struct {
	Length Length
}

func (LengthValue) Type() Type         { return TypeLength }
func (v LengthValue) Display() Content { return Content{} }
func (v LengthValue) Clone() Value     { return v }
func (LengthValue) isValue()           {}

// Angle represents an angle value.
type Angle struct {
	// Radians is the angle in radians.
	Radians float64
}

// AngleValue represents an angle as a Value.
type AngleValue struct {
	Angle Angle
}

func (AngleValue) Type() Type         { return TypeAngle }
func (v AngleValue) Display() Content { return Content{} }
func (v AngleValue) Clone() Value     { return v }
func (AngleValue) isValue()           {}

// Ratio represents a ratio (percentage) value.
type Ratio struct {
	// Value is the ratio as a fraction (0.5 = 50%).
	Value float64
}

// RatioValue represents a ratio as a Value.
type RatioValue struct {
	Ratio Ratio
}

func (RatioValue) Type() Type         { return TypeRatio }
func (v RatioValue) Display() Content { return Content{} }
func (v RatioValue) Clone() Value     { return v }
func (RatioValue) isValue()           {}

// Relative represents a combination of absolute length and ratio.
type Relative struct {
	// Abs is the absolute component.
	Abs Length
	// Rel is the relative component.
	Rel Ratio
}

// RelativeValue represents a relative length as a Value.
type RelativeValue struct {
	Relative Relative
}

func (RelativeValue) Type() Type         { return TypeRelative }
func (v RelativeValue) Display() Content { return Content{} }
func (v RelativeValue) Clone() Value     { return v }
func (RelativeValue) isValue()           {}

// Fraction represents a fraction of remaining space.
type Fraction struct {
	// Value is the number of fractions (1fr = 1.0).
	Value float64
}

// FractionValue represents a fraction as a Value.
type FractionValue struct {
	Fraction Fraction
}

func (FractionValue) Type() Type         { return TypeFraction }
func (v FractionValue) Display() Content { return Content{} }
func (v FractionValue) Clone() Value     { return v }
func (FractionValue) isValue()           {}
