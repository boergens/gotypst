package foundations

import (
	"math"
)

// Calc module: basic math functions
// Translated from foundations/calc.rs

// Abs returns the absolute value of a number.
// For integers, returns an integer. For floats, returns a float.
func Abs(v Value) (Value, error) {
	switch x := v.(type) {
	case Int:
		if x < 0 {
			// Check for overflow (most negative int64)
			if x == math.MinInt64 {
				return nil, &OpError{Message: "integer overflow"}
			}
			return Int(-x), nil
		}
		return x, nil
	case Float:
		return Float(math.Abs(float64(x))), nil
	default:
		return nil, &OpError{
			Message: "expected integer or float, found " + v.Type(),
		}
	}
}

// Pow raises a base to an exponent.
// Returns a float unless both base and exponent are non-negative integers
// and the exponent is small enough.
func Pow(base, exponent Value) (Value, error) {
	// Convert to float64 for calculation
	var baseF, expF float64
	baseIsInt := false
	expIsInt := false
	var expInt int64

	switch b := base.(type) {
	case Int:
		baseF = float64(b)
		baseIsInt = true
	case Float:
		baseF = float64(b)
	default:
		return nil, &OpError{
			Message: "expected integer or float for base, found " + base.Type(),
		}
	}

	switch e := exponent.(type) {
	case Int:
		expF = float64(e)
		expInt = int64(e)
		expIsInt = true
	case Float:
		expF = float64(e)
	default:
		return nil, &OpError{
			Message: "expected integer or float for exponent, found " + exponent.Type(),
		}
	}

	// Check for special cases
	if baseF < 0 && expIsInt && expInt != int64(expF) {
		// Non-integer exponent with negative base
		return nil, &OpError{
			Message: "cannot raise negative number to non-integer power",
		}
	}

	result := math.Pow(baseF, expF)

	// Check for overflow/underflow
	if math.IsInf(result, 0) {
		return Float(result), nil // Return infinity as Float
	}

	// Return integer if both operands were integers, exponent was non-negative,
	// and result fits in int64
	if baseIsInt && expIsInt && expInt >= 0 && expInt <= 62 {
		// Try integer exponentiation for small exponents
		if intResult, ok := intPow(int64(base.(Int)), expInt); ok {
			return Int(intResult), nil
		}
	}

	return Float(result), nil
}

// intPow computes base^exp for integers, checking for overflow.
func intPow(base, exp int64) (int64, bool) {
	if exp == 0 {
		return 1, true
	}
	if exp == 1 {
		return base, true
	}
	if base == 0 {
		return 0, true
	}
	if base == 1 {
		return 1, true
	}
	if base == -1 {
		if exp%2 == 0 {
			return 1, true
		}
		return -1, true
	}

	result := int64(1)
	for exp > 0 {
		if exp%2 == 1 {
			// Check overflow before multiplication
			if willOverflow(result, base) {
				return 0, false
			}
			result *= base
		}
		exp /= 2
		if exp > 0 {
			if willOverflow(base, base) {
				return 0, false
			}
			base *= base
		}
	}
	return result, true
}

// willOverflow checks if a * b would overflow int64.
func willOverflow(a, b int64) bool {
	if a == 0 || b == 0 {
		return false
	}
	result := a * b
	return result/a != b
}

// Exp computes e raised to the power of v.
func Exp(v Value) (Value, error) {
	var x float64
	switch n := v.(type) {
	case Int:
		x = float64(n)
	case Float:
		x = float64(n)
	default:
		return nil, &OpError{
			Message: "expected integer or float, found " + v.Type(),
		}
	}

	return Float(math.Exp(x)), nil
}

// Sqrt computes the square root of a number.
// Returns an error for negative numbers.
func Sqrt(v Value) (Value, error) {
	var x float64
	switch n := v.(type) {
	case Int:
		x = float64(n)
	case Float:
		x = float64(n)
	default:
		return nil, &OpError{
			Message: "expected integer or float, found " + v.Type(),
		}
	}

	if x < 0 {
		return nil, &OpError{
			Message: "cannot take square root of negative number",
		}
	}

	return Float(math.Sqrt(x)), nil
}

// Root computes the n-th root of a number.
// root(x, n) = x^(1/n)
// Returns an error for negative base with even root.
func Root(radicand, index Value) (Value, error) {
	var x, n float64

	switch r := radicand.(type) {
	case Int:
		x = float64(r)
	case Float:
		x = float64(r)
	default:
		return nil, &OpError{
			Message: "expected integer or float for radicand, found " + radicand.Type(),
		}
	}

	switch i := index.(type) {
	case Int:
		n = float64(i)
	case Float:
		n = float64(i)
	default:
		return nil, &OpError{
			Message: "expected integer or float for index, found " + index.Type(),
		}
	}

	if n == 0 {
		return nil, &OpError{
			Message: "cannot compute zeroth root",
		}
	}

	// Handle negative radicand
	if x < 0 {
		// Check if index is an odd integer (which allows negative radicand)
		indexInt, isInt := index.(Int)
		if isInt && indexInt%2 != 0 {
			// Odd root of negative number is negative
			return Float(-math.Pow(-x, 1/n)), nil
		}
		return nil, &OpError{
			Message: "cannot take even root of negative number",
		}
	}

	return Float(math.Pow(x, 1/n)), nil
}
