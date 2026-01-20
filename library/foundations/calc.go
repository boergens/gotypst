// Calc module provides mathematical calculation functions for Typst.
// This includes rounding functions and comparison/clamping operations.
package foundations

import (
	"math"
)

// --- Rounding Functions ---

// Floor rounds a number down to the nearest integer.
// For integers, returns the value unchanged.
// For floats, returns the largest integer less than or equal to the value.
func Floor(v Value) (Value, error) {
	switch x := v.(type) {
	case Int:
		return x, nil
	case Float:
		return Int(int64(math.Floor(float64(x)))), nil
	default:
		return nil, &OpError{Message: "expected integer or float for floor"}
	}
}

// Ceil rounds a number up to the nearest integer.
// For integers, returns the value unchanged.
// For floats, returns the smallest integer greater than or equal to the value.
func Ceil(v Value) (Value, error) {
	switch x := v.(type) {
	case Int:
		return x, nil
	case Float:
		return Int(int64(math.Ceil(float64(x)))), nil
	default:
		return nil, &OpError{Message: "expected integer or float for ceil"}
	}
}

// Trunc truncates a number towards zero.
// For integers, returns the value unchanged.
// For floats, removes the fractional part.
func Trunc(v Value) (Value, error) {
	switch x := v.(type) {
	case Int:
		return x, nil
	case Float:
		return Int(int64(math.Trunc(float64(x)))), nil
	default:
		return nil, &OpError{Message: "expected integer or float for trunc"}
	}
}

// Fract returns the fractional part of a number.
// For integers, returns 0.0.
// For floats, returns the part after the decimal point (value - trunc(value)).
func Fract(v Value) (Value, error) {
	switch x := v.(type) {
	case Int:
		return Float(0), nil
	case Float:
		_, frac := math.Modf(float64(x))
		return Float(frac), nil
	default:
		return nil, &OpError{Message: "expected integer or float for fract"}
	}
}

// Round rounds a number to the nearest integer.
// Rounds half away from zero (0.5 rounds to 1, -0.5 rounds to -1).
// For integers, returns the value unchanged.
// An optional digits parameter specifies decimal places (positive) or
// significant integer digits (negative).
func Round(v Value, digits *Int) (Value, error) {
	switch x := v.(type) {
	case Int:
		if digits == nil || *digits >= 0 {
			return x, nil
		}
		// Round to significant integer digits
		d := int64(*digits)
		shift := math.Pow(10, float64(-d))
		rounded := math.Round(float64(x)/shift) * shift
		return Int(int64(rounded)), nil
	case Float:
		if digits == nil {
			// Round to nearest integer
			return Int(int64(roundHalfAwayFromZero(float64(x)))), nil
		}
		d := int64(*digits)
		if d >= 0 {
			// Round to specified decimal places
			shift := math.Pow(10, float64(d))
			return Float(roundHalfAwayFromZero(float64(x)*shift) / shift), nil
		}
		// Round to significant integer digits
		shift := math.Pow(10, float64(-d))
		return Float(roundHalfAwayFromZero(float64(x)/shift) * shift), nil
	default:
		return nil, &OpError{Message: "expected integer or float for round"}
	}
}

// roundHalfAwayFromZero rounds to nearest integer, with 0.5 going away from zero.
func roundHalfAwayFromZero(x float64) float64 {
	if x >= 0 {
		return math.Floor(x + 0.5)
	}
	return math.Ceil(x - 0.5)
}

// --- Comparison Functions ---

// Clamp restricts a value to be within a range.
// Returns min if value < min, max if value > max, otherwise value.
func Clamp(v, minVal, maxVal Value) (Value, error) {
	// Get numeric values
	vNum, vOk := toFloat64(v)
	minNum, minOk := toFloat64(minVal)
	maxNum, maxOk := toFloat64(maxVal)

	if !vOk || !minOk || !maxOk {
		return nil, &OpError{Message: "expected numeric values for clamp"}
	}

	if minNum > maxNum {
		return nil, &OpError{Message: "min must be less than or equal to max"}
	}

	result := vNum
	if result < minNum {
		result = minNum
	}
	if result > maxNum {
		result = maxNum
	}

	// Return the same type as the input value
	if _, isInt := v.(Int); isInt {
		if _, minIsInt := minVal.(Int); minIsInt {
			if _, maxIsInt := maxVal.(Int); maxIsInt {
				return Int(int64(result)), nil
			}
		}
	}
	return Float(result), nil
}

// Min returns the minimum of the given values.
// Requires at least one argument.
func Min(values ...Value) (Value, error) {
	if len(values) == 0 {
		return nil, &OpError{Message: "min requires at least one argument"}
	}

	minVal := values[0]
	minNum, ok := toFloat64(minVal)
	if !ok {
		return nil, &OpError{Message: "expected numeric value for min"}
	}

	allInt := isInt(minVal)
	for _, v := range values[1:] {
		num, ok := toFloat64(v)
		if !ok {
			return nil, &OpError{Message: "expected numeric value for min"}
		}
		if num < minNum {
			minNum = num
			minVal = v
		}
		if !isInt(v) {
			allInt = false
		}
	}

	if allInt {
		return Int(int64(minNum)), nil
	}
	return Float(minNum), nil
}

// Max returns the maximum of the given values.
// Requires at least one argument.
func Max(values ...Value) (Value, error) {
	if len(values) == 0 {
		return nil, &OpError{Message: "max requires at least one argument"}
	}

	maxVal := values[0]
	maxNum, ok := toFloat64(maxVal)
	if !ok {
		return nil, &OpError{Message: "expected numeric value for max"}
	}

	allInt := isInt(maxVal)
	for _, v := range values[1:] {
		num, ok := toFloat64(v)
		if !ok {
			return nil, &OpError{Message: "expected numeric value for max"}
		}
		if num > maxNum {
			maxNum = num
			maxVal = v
		}
		if !isInt(v) {
			allInt = false
		}
	}

	if allInt {
		return Int(int64(maxNum)), nil
	}
	return Float(maxNum), nil
}

// --- Helper Functions ---

// toFloat64 converts a numeric value to float64.
func toFloat64(v Value) (float64, bool) {
	switch x := v.(type) {
	case Int:
		return float64(x), true
	case Float:
		return float64(x), true
	default:
		return 0, false
	}
}

// isInt returns true if the value is an integer.
func isInt(v Value) bool {
	_, ok := v.(Int)
	return ok
}
