// Float type and constructor for Typst.
// Translated from foundations/float.rs

package foundations

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/boergens/gotypst/syntax"
)

// FloatConstruct converts a value to a float.
// Supports: bool, int, float, decimal, ratio, string.
//
// This matches Rust's float::construct function.
func FloatConstruct(args *Args) (Value, error) {
	spanned, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	value := spanned.V

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case Float:
		return v, nil

	case Int:
		return Float(float64(v)), nil

	case Bool:
		if v {
			return Float(1.0), nil
		}
		return Float(0.0), nil

	case DecimalValue:
		if v.Value == nil {
			return Float(0), nil
		}
		f, _ := v.Value.Float64()
		return Float(f), nil

	case RatioValue:
		return Float(v.Ratio.Value), nil

	case Str:
		return parseFloatFromString(string(v), spanned.Span)

	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("expected integer, boolean, float, decimal, ratio, or string, found %s", value.Type().String()),
			Span:    spanned.Span,
		}
	}
}

// parseFloatFromString parses a float from a string.
func parseFloatFromString(s string, span syntax.Span) (Value, error) {
	if s == "" {
		return nil, &ConstructorError{
			Message: "string must not be empty",
			Span:    span,
		}
	}

	// Handle special values
	switch strings.ToLower(s) {
	case "inf", "+inf":
		return Float(math.Inf(1)), nil
	case "-inf":
		return Float(math.Inf(-1)), nil
	case "nan":
		return Float(math.NaN()), nil
	}

	// Handle Unicode minus sign (U+2212)
	s = strings.ReplaceAll(s, "\u2212", "-")

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, &ConstructorError{
			Message: fmt.Sprintf("invalid float: %s", s),
			Span:    span,
		}
	}
	return Float(f), nil
}

// Float methods

// FloatIsNaN checks if a float is NaN.
func FloatIsNaN(f Float) Bool {
	return Bool(math.IsNaN(float64(f)))
}

// FloatIsInfinite checks if a float is infinite.
func FloatIsInfinite(f Float) Bool {
	return Bool(math.IsInf(float64(f), 0))
}

// FloatSignum returns the sign of a float.
func FloatSignum(f Float) Float {
	if math.IsNaN(float64(f)) {
		return Float(math.NaN())
	}
	return Float(math.Copysign(1, float64(f)))
}
