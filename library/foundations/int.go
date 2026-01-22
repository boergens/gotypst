// Int type and constructor for Typst.
// Translated from foundations/int.rs

package foundations

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/boergens/gotypst/syntax"
)

// IntConstruct converts a value to an integer.
// Supports: bool, int, float, decimal, string (with optional base).
//
// This matches Rust's int::construct function.
func IntConstruct(args *Args) (Value, error) {
	// Get the value to convert
	spanned, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	value := spanned.V

	// Check for base argument (only valid for strings)
	base := 10
	if baseArg := args.Find("base"); baseArg != nil {
		// Base is only valid for strings
		if _, ok := value.(Str); !ok {
			return nil, &ConstructorError{
				Message: "base is only supported for strings",
				Span:    baseArg.Span,
			}
		}
		baseVal, ok := baseArg.V.(Int)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "integer",
				Got:      baseArg.V.Type().String(),
				Span:     baseArg.Span,
			}
		}
		base = int(baseVal)
		if base < 2 || base > 36 {
			return nil, &ConstructorError{
				Message: "base must be between 2 and 36",
				Span:    baseArg.Span,
			}
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case Int:
		return v, nil

	case Bool:
		if v {
			return Int(1), nil
		}
		return Int(0), nil

	case Float:
		return Int(int64(v)), nil

	case DecimalValue:
		// Truncate decimal to integer
		if v.Value == nil {
			return Int(0), nil
		}
		f, _ := v.Value.Float64()
		return Int(int64(f)), nil

	case Str:
		return parseIntFromString(string(v), base, spanned.Span)

	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("expected integer, boolean, float, decimal, or string, found %s", value.Type().String()),
			Span:    spanned.Span,
		}
	}
}

// parseIntFromString parses an integer from a string with the given base.
func parseIntFromString(s string, base int, span syntax.Span) (Value, error) {
	if s == "" {
		return nil, &ConstructorError{
			Message: "string must not be empty",
			Span:    span,
		}
	}

	// Handle Unicode minus sign (U+2212)
	s = strings.ReplaceAll(s, "\u2212", "-")

	// Parse with the given base
	result, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		numErr, ok := err.(*strconv.NumError)
		if ok {
			switch numErr.Err {
			case strconv.ErrRange:
				if strings.HasPrefix(s, "-") {
					return nil, &ConstructorError{
						Message: "integer value is too small",
						Span:    span,
						Hints:   []string{"value does not fit into a signed 64-bit integer", "try using a floating point number"},
					}
				}
				return nil, &ConstructorError{
					Message: "integer value is too large",
					Span:    span,
					Hints:   []string{"value does not fit into a signed 64-bit integer", "try using a floating point number"},
				}
			case strconv.ErrSyntax:
				if base != 10 {
					return nil, &ConstructorError{
						Message: fmt.Sprintf("string contains invalid digits for a base %d integer", base),
						Span:    span,
					}
				}
				return nil, &ConstructorError{
					Message: "string contains invalid digits",
					Span:    span,
				}
			}
		}
		return nil, &ConstructorError{
			Message: "string contains invalid digits",
			Span:    span,
		}
	}
	return Int(result), nil
}

// Int methods

// IntSignum returns the sign of an integer: 1 for positive, -1 for negative, 0 for zero.
func IntSignum(n Int) Int {
	if n > 0 {
		return 1
	} else if n < 0 {
		return -1
	}
	return 0
}

// IntBitNot returns the bitwise NOT of an integer.
func IntBitNot(n Int) Int {
	return ^n
}

// IntBitAnd returns the bitwise AND of two integers.
func IntBitAnd(a, b Int) Int {
	return a & b
}

// IntBitOr returns the bitwise OR of two integers.
func IntBitOr(a, b Int) Int {
	return a | b
}

// IntBitXor returns the bitwise XOR of two integers.
func IntBitXor(a, b Int) Int {
	return a ^ b
}

// IntBitLshift shifts the bits left by the given amount.
func IntBitLshift(n Int, shift uint) (Int, error) {
	if shift >= 64 {
		return 0, &OpError{Message: "the result is too large"}
	}
	return n << shift, nil
}

// IntBitRshift shifts the bits right by the given amount.
// If logical is true, performs a logical (unsigned) shift.
func IntBitRshift(n Int, shift uint, logical bool) Int {
	if logical {
		if shift >= 64 {
			return 0
		}
		return Int(uint64(n) >> shift)
	}
	// Arithmetic shift - saturate at 63 bits
	if shift >= 64 {
		shift = 63
	}
	return n >> shift
}
