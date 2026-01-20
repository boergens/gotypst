package calc

import (
	"math"

	"github.com/boergens/gotypst/library/foundations"
)

// toFloat64 converts a numeric Value to float64.
func toFloat64(v foundations.Value) (float64, bool) {
	switch x := v.(type) {
	case foundations.Int:
		return float64(x), true
	case foundations.Float:
		return float64(x), true
	default:
		return 0, false
	}
}

// numericTypeError returns an error for non-numeric input.
func numericTypeError(fn string, v foundations.Value) error {
	return &foundations.OpError{
		Message: fn + " expected numeric value, got " + v.Type(),
	}
}

// Sin computes the sine of an angle in radians.
func Sin(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("sin", v)
	}
	return foundations.Float(math.Sin(x)), nil
}

// Cos computes the cosine of an angle in radians.
func Cos(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("cos", v)
	}
	return foundations.Float(math.Cos(x)), nil
}

// Tan computes the tangent of an angle in radians.
func Tan(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("tan", v)
	}
	return foundations.Float(math.Tan(x)), nil
}

// Asin computes the arc sine (inverse sine) of a value.
// The result is in radians, in the range [-pi/2, pi/2].
// Returns NaN if the input is outside [-1, 1].
func Asin(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("asin", v)
	}
	return foundations.Float(math.Asin(x)), nil
}

// Acos computes the arc cosine (inverse cosine) of a value.
// The result is in radians, in the range [0, pi].
// Returns NaN if the input is outside [-1, 1].
func Acos(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("acos", v)
	}
	return foundations.Float(math.Acos(x)), nil
}

// Atan computes the arc tangent (inverse tangent) of a value.
// The result is in radians, in the range [-pi/2, pi/2].
func Atan(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("atan", v)
	}
	return foundations.Float(math.Atan(x)), nil
}

// Atan2 computes the arc tangent of y/x, using the signs of both
// arguments to determine the quadrant of the return value.
// The result is in radians, in the range [-pi, pi].
func Atan2(y, x foundations.Value) (foundations.Value, error) {
	yf, ok := toFloat64(y)
	if !ok {
		return nil, numericTypeError("atan2", y)
	}
	xf, ok := toFloat64(x)
	if !ok {
		return nil, numericTypeError("atan2", x)
	}
	return foundations.Float(math.Atan2(yf, xf)), nil
}

// Sinh computes the hyperbolic sine of a value.
func Sinh(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("sinh", v)
	}
	return foundations.Float(math.Sinh(x)), nil
}

// Cosh computes the hyperbolic cosine of a value.
func Cosh(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("cosh", v)
	}
	return foundations.Float(math.Cosh(x)), nil
}

// Tanh computes the hyperbolic tangent of a value.
func Tanh(v foundations.Value) (foundations.Value, error) {
	x, ok := toFloat64(v)
	if !ok {
		return nil, numericTypeError("tanh", v)
	}
	return foundations.Float(math.Tanh(x)), nil
}
