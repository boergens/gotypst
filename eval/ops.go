// Operations on values.
// Translated from typst-library/src/foundations/ops.rs

package eval

import (
	"fmt"
	"strings"

	"github.com/boergens/gotypst/library/foundations"
)

// mismatch returns a type mismatch error.
func mismatch(format string, values ...foundations.Value) error {
	types := make([]interface{}, len(values))
	for i, v := range values {
		types[i] = v.Type()
	}
	return fmt.Errorf(format, types...)
}

// tooLarge returns the overflow error message.
func tooLarge() error {
	return fmt.Errorf("value is too large")
}

// ----------------------------------------------------------------------------
// Join
// ----------------------------------------------------------------------------

// Join joins a value with another value.
// Matches Rust: pub fn join(lhs: Value, rhs: Value) -> StrResult<Value>
func Join(lhs, rhs foundations.Value) (foundations.Value, error) {
	// Handle None
	if _, ok := lhs.(foundations.NoneValue); ok {
		return rhs, nil
	}
	if _, ok := rhs.(foundations.NoneValue); ok {
		return lhs, nil
	}

	switch l := lhs.(type) {
	case foundations.Str:
		if r, ok := rhs.(foundations.Str); ok {
			return foundations.Str(string(l) + string(r)), nil
		}
		if r, ok := rhs.(foundations.SymbolValue); ok {
			return foundations.Str(string(l) + string(r.Char)), nil
		}
		if r, ok := rhs.(foundations.ContentValue); ok {
			return foundations.ContentValue{Content: joinContent(strToContent(l), r.Content)}, nil
		}

	case foundations.SymbolValue:
		if r, ok := rhs.(foundations.Str); ok {
			return foundations.Str(string(l.Char) + string(r)), nil
		}
		if r, ok := rhs.(foundations.SymbolValue); ok {
			return foundations.Str(string(l.Char) + string(r.Char)), nil
		}
		if r, ok := rhs.(foundations.ContentValue); ok {
			return foundations.ContentValue{Content: joinContent(symbolToContent(l), r.Content)}, nil
		}

	case foundations.ContentValue:
		if r, ok := rhs.(foundations.ContentValue); ok {
			return foundations.ContentValue{Content: joinContent(l.Content, r.Content)}, nil
		}
		if r, ok := rhs.(foundations.SymbolValue); ok {
			return foundations.ContentValue{Content: joinContent(l.Content, symbolToContent(r))}, nil
		}
		if r, ok := rhs.(foundations.Str); ok {
			return foundations.ContentValue{Content: joinContent(l.Content, strToContent(r))}, nil
		}

	case *foundations.Array:
		if r, ok := rhs.(*foundations.Array); ok {
			return arrayAdd(l, r), nil
		}

	case *foundations.Dict:
		if r, ok := rhs.(*foundations.Dict); ok {
			return dictAdd(l, r), nil
		}
	}

	return nil, mismatch("cannot join %s with %s", lhs, rhs)
}

// ----------------------------------------------------------------------------
// Unary Operators
// ----------------------------------------------------------------------------

// Pos applies the unary plus operator to a value.
// Matches Rust: pub fn pos(value: Value) -> HintedStrResult<Value>
func Pos(value foundations.Value) (foundations.Value, error) {
	switch v := value.(type) {
	case foundations.Int:
		return v, nil
	case foundations.Float:
		return v, nil
	case foundations.LengthValue:
		return v, nil
	case foundations.AngleValue:
		return v, nil
	case foundations.RatioValue:
		return v, nil
	case foundations.RelativeValue:
		return v, nil
	case foundations.FractionValue:
		return v, nil
	}
	return nil, mismatch("cannot apply unary '+' to %s", value)
}

// Neg computes the negation of a value.
// Matches Rust: pub fn neg(value: Value) -> HintedStrResult<Value>
func Neg(value foundations.Value) (foundations.Value, error) {
	switch v := value.(type) {
	case foundations.Int:
		if v == foundations.Int(^uint64(0)>>1+1) { // MinInt64
			return nil, tooLarge()
		}
		return -v, nil
	case foundations.Float:
		return -v, nil
	case foundations.LengthValue:
		return foundations.LengthValue{Length: foundations.Length{Points: -v.Length.Points}}, nil
	case foundations.AngleValue:
		return foundations.AngleValue{Angle: foundations.Angle{Radians: -v.Angle.Radians}}, nil
	case foundations.RatioValue:
		return foundations.RatioValue{Ratio: foundations.Ratio{Value: -v.Ratio.Value}}, nil
	case foundations.RelativeValue:
		return foundations.RelativeValue{Relative: foundations.Relative{
			Abs: foundations.Length{Points: -v.Relative.Abs.Points},
			Rel: foundations.Ratio{Value: -v.Relative.Rel.Value},
		}}, nil
	case foundations.FractionValue:
		return foundations.FractionValue{Fraction: foundations.Fraction{Value: -v.Fraction.Value}}, nil
	case foundations.Duration:
		return -v, nil
	}
	return nil, mismatch("cannot apply '-' to %s", value)
}

// ----------------------------------------------------------------------------
// Arithmetic Operators
// ----------------------------------------------------------------------------

// Add computes the sum of two values.
// Matches Rust: pub fn add(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func Add(lhs, rhs foundations.Value) (foundations.Value, error) {
	// Handle None
	if _, ok := lhs.(foundations.NoneValue); ok {
		return rhs, nil
	}
	if _, ok := rhs.(foundations.NoneValue); ok {
		return lhs, nil
	}

	switch l := lhs.(type) {
	case foundations.Int:
		switch r := rhs.(type) {
		case foundations.Int:
			result := int64(l) + int64(r)
			// Check overflow
			if (int64(r) > 0 && result < int64(l)) || (int64(r) < 0 && result > int64(l)) {
				return nil, tooLarge()
			}
			return foundations.Int(result), nil
		case foundations.Float:
			return foundations.Float(float64(l) + float64(r)), nil
		}

	case foundations.Float:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.Float(float64(l) + float64(r)), nil
		case foundations.Float:
			return foundations.Float(float64(l) + float64(r)), nil
		}

	case foundations.AngleValue:
		if r, ok := rhs.(foundations.AngleValue); ok {
			return foundations.AngleValue{Angle: foundations.Angle{Radians: l.Angle.Radians + r.Angle.Radians}}, nil
		}

	case foundations.LengthValue:
		switch r := rhs.(type) {
		case foundations.LengthValue:
			return foundations.LengthValue{Length: foundations.Length{Points: l.Length.Points + r.Length.Points}}, nil
		case foundations.RatioValue:
			return foundations.RelativeValue{Relative: foundations.Relative{Abs: l.Length, Rel: r.Ratio}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Length.Points + r.Relative.Abs.Points},
				Rel: r.Relative.Rel,
			}}, nil
		}

	case foundations.RatioValue:
		switch r := rhs.(type) {
		case foundations.LengthValue:
			return foundations.RelativeValue{Relative: foundations.Relative{Abs: r.Length, Rel: l.Ratio}}, nil
		case foundations.RatioValue:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: l.Ratio.Value + r.Ratio.Value}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: r.Relative.Abs,
				Rel: foundations.Ratio{Value: l.Ratio.Value + r.Relative.Rel.Value},
			}}, nil
		}

	case foundations.RelativeValue:
		switch r := rhs.(type) {
		case foundations.LengthValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points + r.Length.Points},
				Rel: l.Relative.Rel,
			}}, nil
		case foundations.RatioValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: l.Relative.Abs,
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value + r.Ratio.Value},
			}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points + r.Relative.Abs.Points},
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value + r.Relative.Rel.Value},
			}}, nil
		}

	case foundations.FractionValue:
		if r, ok := rhs.(foundations.FractionValue); ok {
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: l.Fraction.Value + r.Fraction.Value}}, nil
		}

	case foundations.Str:
		switch r := rhs.(type) {
		case foundations.Str:
			return foundations.Str(string(l) + string(r)), nil
		case foundations.SymbolValue:
			return foundations.Str(string(l) + string(r.Char)), nil
		case foundations.ContentValue:
			return foundations.ContentValue{Content: joinContent(strToContent(l), r.Content)}, nil
		}

	case foundations.SymbolValue:
		switch r := rhs.(type) {
		case foundations.SymbolValue:
			return foundations.Str(string(l.Char) + string(r.Char)), nil
		case foundations.Str:
			return foundations.Str(string(l.Char) + string(r)), nil
		case foundations.ContentValue:
			return foundations.ContentValue{Content: joinContent(symbolToContent(l), r.Content)}, nil
		}

	case foundations.ContentValue:
		switch r := rhs.(type) {
		case foundations.ContentValue:
			return foundations.ContentValue{Content: joinContent(l.Content, r.Content)}, nil
		case foundations.SymbolValue:
			return foundations.ContentValue{Content: joinContent(l.Content, symbolToContent(r))}, nil
		case foundations.Str:
			return foundations.ContentValue{Content: joinContent(l.Content, strToContent(r))}, nil
		}

	case *foundations.Array:
		if r, ok := rhs.(*foundations.Array); ok {
			return arrayAdd(l, r), nil
		}

	case *foundations.Dict:
		if r, ok := rhs.(*foundations.Dict); ok {
			return dictAdd(l, r), nil
		}

	case foundations.Duration:
		if r, ok := rhs.(foundations.Duration); ok {
			return foundations.Duration(int64(l) + int64(r)), nil
		}
	}

	return nil, mismatch("cannot add %s and %s", lhs, rhs)
}

// Sub computes the difference of two values.
// Matches Rust: pub fn sub(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func Sub(lhs, rhs foundations.Value) (foundations.Value, error) {
	switch l := lhs.(type) {
	case foundations.Int:
		switch r := rhs.(type) {
		case foundations.Int:
			result := int64(l) - int64(r)
			// Check overflow
			if (int64(r) > 0 && result > int64(l)) || (int64(r) < 0 && result < int64(l)) {
				return nil, tooLarge()
			}
			return foundations.Int(result), nil
		case foundations.Float:
			return foundations.Float(float64(l) - float64(r)), nil
		}

	case foundations.Float:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.Float(float64(l) - float64(r)), nil
		case foundations.Float:
			return foundations.Float(float64(l) - float64(r)), nil
		}

	case foundations.AngleValue:
		if r, ok := rhs.(foundations.AngleValue); ok {
			return foundations.AngleValue{Angle: foundations.Angle{Radians: l.Angle.Radians - r.Angle.Radians}}, nil
		}

	case foundations.LengthValue:
		switch r := rhs.(type) {
		case foundations.LengthValue:
			return foundations.LengthValue{Length: foundations.Length{Points: l.Length.Points - r.Length.Points}}, nil
		case foundations.RatioValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: l.Length,
				Rel: foundations.Ratio{Value: -r.Ratio.Value},
			}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Length.Points - r.Relative.Abs.Points},
				Rel: foundations.Ratio{Value: -r.Relative.Rel.Value},
			}}, nil
		}

	case foundations.RatioValue:
		switch r := rhs.(type) {
		case foundations.LengthValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: -r.Length.Points},
				Rel: l.Ratio,
			}}, nil
		case foundations.RatioValue:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: l.Ratio.Value - r.Ratio.Value}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: -r.Relative.Abs.Points},
				Rel: foundations.Ratio{Value: l.Ratio.Value - r.Relative.Rel.Value},
			}}, nil
		}

	case foundations.RelativeValue:
		switch r := rhs.(type) {
		case foundations.LengthValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points - r.Length.Points},
				Rel: l.Relative.Rel,
			}}, nil
		case foundations.RatioValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: l.Relative.Abs,
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value - r.Ratio.Value},
			}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points - r.Relative.Abs.Points},
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value - r.Relative.Rel.Value},
			}}, nil
		}

	case foundations.FractionValue:
		if r, ok := rhs.(foundations.FractionValue); ok {
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: l.Fraction.Value - r.Fraction.Value}}, nil
		}

	case foundations.Duration:
		if r, ok := rhs.(foundations.Duration); ok {
			return foundations.Duration(int64(l) - int64(r)), nil
		}
	}

	return nil, mismatch("cannot subtract %s from %s", rhs, lhs)
}

// Mul computes the product of two values.
// Matches Rust: pub fn mul(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func Mul(lhs, rhs foundations.Value) (foundations.Value, error) {
	switch l := lhs.(type) {
	case foundations.Int:
		switch r := rhs.(type) {
		case foundations.Int:
			result := int64(l) * int64(r)
			// Simple overflow check
			if int64(l) != 0 && result/int64(l) != int64(r) {
				return nil, tooLarge()
			}
			return foundations.Int(result), nil
		case foundations.Float:
			return foundations.Float(float64(l) * float64(r)), nil
		case foundations.LengthValue:
			return foundations.LengthValue{Length: foundations.Length{Points: float64(l) * r.Length.Points}}, nil
		case foundations.AngleValue:
			return foundations.AngleValue{Angle: foundations.Angle{Radians: float64(l) * r.Angle.Radians}}, nil
		case foundations.RatioValue:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: float64(l) * r.Ratio.Value}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: float64(l) * r.Relative.Abs.Points},
				Rel: foundations.Ratio{Value: float64(l) * r.Relative.Rel.Value},
			}}, nil
		case foundations.FractionValue:
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: float64(l) * r.Fraction.Value}}, nil
		case foundations.Str:
			return strRepeat(r, int(l))
		case *foundations.Array:
			return arrayRepeat(r, int(l))
		case foundations.Duration:
			return foundations.Duration(float64(r) * float64(l)), nil
		}

	case foundations.Float:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.Float(float64(l) * float64(r)), nil
		case foundations.Float:
			return foundations.Float(float64(l) * float64(r)), nil
		case foundations.LengthValue:
			return foundations.LengthValue{Length: foundations.Length{Points: float64(l) * r.Length.Points}}, nil
		case foundations.AngleValue:
			return foundations.AngleValue{Angle: foundations.Angle{Radians: float64(l) * r.Angle.Radians}}, nil
		case foundations.RatioValue:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: float64(l) * r.Ratio.Value}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: float64(l) * r.Relative.Abs.Points},
				Rel: foundations.Ratio{Value: float64(l) * r.Relative.Rel.Value},
			}}, nil
		case foundations.FractionValue:
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: float64(l) * r.Fraction.Value}}, nil
		case foundations.Duration:
			return foundations.Duration(float64(r) * float64(l)), nil
		}

	case foundations.LengthValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.LengthValue{Length: foundations.Length{Points: l.Length.Points * float64(r)}}, nil
		case foundations.Float:
			return foundations.LengthValue{Length: foundations.Length{Points: l.Length.Points * float64(r)}}, nil
		case foundations.RatioValue:
			return foundations.LengthValue{Length: foundations.Length{Points: l.Length.Points * r.Ratio.Value}}, nil
		}

	case foundations.AngleValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.AngleValue{Angle: foundations.Angle{Radians: l.Angle.Radians * float64(r)}}, nil
		case foundations.Float:
			return foundations.AngleValue{Angle: foundations.Angle{Radians: l.Angle.Radians * float64(r)}}, nil
		case foundations.RatioValue:
			return foundations.AngleValue{Angle: foundations.Angle{Radians: l.Angle.Radians * r.Ratio.Value}}, nil
		}

	case foundations.RatioValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: l.Ratio.Value * float64(r)}}, nil
		case foundations.Float:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: l.Ratio.Value * float64(r)}}, nil
		case foundations.RatioValue:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: l.Ratio.Value * r.Ratio.Value}}, nil
		case foundations.LengthValue:
			return foundations.LengthValue{Length: foundations.Length{Points: l.Ratio.Value * r.Length.Points}}, nil
		case foundations.AngleValue:
			return foundations.AngleValue{Angle: foundations.Angle{Radians: l.Ratio.Value * r.Angle.Radians}}, nil
		case foundations.RelativeValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Ratio.Value * r.Relative.Abs.Points},
				Rel: foundations.Ratio{Value: l.Ratio.Value * r.Relative.Rel.Value},
			}}, nil
		case foundations.FractionValue:
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: l.Ratio.Value * r.Fraction.Value}}, nil
		}

	case foundations.RelativeValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points * float64(r)},
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value * float64(r)},
			}}, nil
		case foundations.Float:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points * float64(r)},
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value * float64(r)},
			}}, nil
		case foundations.RatioValue:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points * r.Ratio.Value},
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value * r.Ratio.Value},
			}}, nil
		}

	case foundations.FractionValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: l.Fraction.Value * float64(r)}}, nil
		case foundations.Float:
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: l.Fraction.Value * float64(r)}}, nil
		case foundations.RatioValue:
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: l.Fraction.Value * r.Ratio.Value}}, nil
		}

	case foundations.Str:
		if r, ok := rhs.(foundations.Int); ok {
			return strRepeat(l, int(r))
		}

	case *foundations.Array:
		if r, ok := rhs.(foundations.Int); ok {
			return arrayRepeat(l, int(r))
		}

	case foundations.Duration:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.Duration(float64(l) * float64(r)), nil
		case foundations.Float:
			return foundations.Duration(float64(l) * float64(r)), nil
		}
	}

	return nil, mismatch("cannot multiply %s with %s", lhs, rhs)
}

// Div computes the quotient of two values.
// Matches Rust: pub fn div(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func Div(lhs, rhs foundations.Value) (foundations.Value, error) {
	if isZero(rhs) {
		return nil, fmt.Errorf("cannot divide by zero")
	}

	switch l := lhs.(type) {
	case foundations.Int:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.Float(float64(l) / float64(r)), nil
		case foundations.Float:
			return foundations.Float(float64(l) / float64(r)), nil
		}

	case foundations.Float:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.Float(float64(l) / float64(r)), nil
		case foundations.Float:
			return foundations.Float(float64(l) / float64(r)), nil
		}

	case foundations.LengthValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.LengthValue{Length: foundations.Length{Points: l.Length.Points / float64(r)}}, nil
		case foundations.Float:
			return foundations.LengthValue{Length: foundations.Length{Points: l.Length.Points / float64(r)}}, nil
		case foundations.LengthValue:
			return foundations.Float(l.Length.Points / r.Length.Points), nil
		}

	case foundations.AngleValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.AngleValue{Angle: foundations.Angle{Radians: l.Angle.Radians / float64(r)}}, nil
		case foundations.Float:
			return foundations.AngleValue{Angle: foundations.Angle{Radians: l.Angle.Radians / float64(r)}}, nil
		case foundations.AngleValue:
			return foundations.Float(l.Angle.Radians / r.Angle.Radians), nil
		}

	case foundations.RatioValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: l.Ratio.Value / float64(r)}}, nil
		case foundations.Float:
			return foundations.RatioValue{Ratio: foundations.Ratio{Value: l.Ratio.Value / float64(r)}}, nil
		case foundations.RatioValue:
			return foundations.Float(l.Ratio.Value / r.Ratio.Value), nil
		}

	case foundations.RelativeValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points / float64(r)},
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value / float64(r)},
			}}, nil
		case foundations.Float:
			return foundations.RelativeValue{Relative: foundations.Relative{
				Abs: foundations.Length{Points: l.Relative.Abs.Points / float64(r)},
				Rel: foundations.Ratio{Value: l.Relative.Rel.Value / float64(r)},
			}}, nil
		}

	case foundations.FractionValue:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: l.Fraction.Value / float64(r)}}, nil
		case foundations.Float:
			return foundations.FractionValue{Fraction: foundations.Fraction{Value: l.Fraction.Value / float64(r)}}, nil
		case foundations.FractionValue:
			return foundations.Float(l.Fraction.Value / r.Fraction.Value), nil
		}

	case foundations.Duration:
		switch r := rhs.(type) {
		case foundations.Int:
			return foundations.Duration(float64(l) / float64(r)), nil
		case foundations.Float:
			return foundations.Duration(float64(l) / float64(r)), nil
		case foundations.Duration:
			return foundations.Float(float64(l) / float64(r)), nil
		}
	}

	return nil, mismatch("cannot divide %s by %s", lhs, rhs)
}

// isZero checks if a value is zero.
// Matches Rust: fn is_zero(v: &Value) -> bool
func isZero(v foundations.Value) bool {
	switch val := v.(type) {
	case foundations.Int:
		return val == 0
	case foundations.Float:
		return val == 0.0
	case foundations.LengthValue:
		return val.Length.Points == 0
	case foundations.AngleValue:
		return val.Angle.Radians == 0
	case foundations.RatioValue:
		return val.Ratio.Value == 0
	case foundations.RelativeValue:
		return val.Relative.Abs.Points == 0 && val.Relative.Rel.Value == 0
	case foundations.FractionValue:
		return val.Fraction.Value == 0
	case foundations.Duration:
		return val == 0
	}
	return false
}

// ----------------------------------------------------------------------------
// Logical Operators
// ----------------------------------------------------------------------------

// Not computes the logical "not" of a value.
// Matches Rust: pub fn not(value: Value) -> HintedStrResult<Value>
func Not(value foundations.Value) (foundations.Value, error) {
	if b, ok := value.(foundations.Bool); ok {
		return foundations.Bool(!b), nil
	}
	return nil, mismatch("cannot apply 'not' to %s", value)
}

// And computes the logical "and" of two values.
// Matches Rust: pub fn and(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func And(lhs, rhs foundations.Value) (foundations.Value, error) {
	if a, ok := lhs.(foundations.Bool); ok {
		if b, ok := rhs.(foundations.Bool); ok {
			return foundations.Bool(a && b), nil
		}
	}
	return nil, mismatch("cannot apply 'and' to %s and %s", lhs, rhs)
}

// Or computes the logical "or" of two values.
// Matches Rust: pub fn or(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func Or(lhs, rhs foundations.Value) (foundations.Value, error) {
	if a, ok := lhs.(foundations.Bool); ok {
		if b, ok := rhs.(foundations.Bool); ok {
			return foundations.Bool(a || b), nil
		}
	}
	return nil, mismatch("cannot apply 'or' to %s and %s", lhs, rhs)
}

// ----------------------------------------------------------------------------
// Comparison Operators
// ----------------------------------------------------------------------------

// Eq computes whether two values are equal.
// Matches Rust: pub fn eq(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func Eq(lhs, rhs foundations.Value) (foundations.Value, error) {
	return foundations.Bool(Equal(lhs, rhs)), nil
}

// Neq computes whether two values are unequal.
// Matches Rust: pub fn neq(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func Neq(lhs, rhs foundations.Value) (foundations.Value, error) {
	return foundations.Bool(!Equal(lhs, rhs)), nil
}

// Lt computes whether lhs < rhs.
// Matches Rust: comparison!(lt, "<", Ordering::Less)
func Lt(lhs, rhs foundations.Value) (foundations.Value, error) {
	cmp, err := Compare(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return foundations.Bool(cmp < 0), nil
}

// Leq computes whether lhs <= rhs.
// Matches Rust: comparison!(leq, "<=", Ordering::Less | Ordering::Equal)
func Leq(lhs, rhs foundations.Value) (foundations.Value, error) {
	cmp, err := Compare(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return foundations.Bool(cmp <= 0), nil
}

// Gt computes whether lhs > rhs.
// Matches Rust: comparison!(gt, ">", Ordering::Greater)
func Gt(lhs, rhs foundations.Value) (foundations.Value, error) {
	cmp, err := Compare(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return foundations.Bool(cmp > 0), nil
}

// Geq computes whether lhs >= rhs.
// Matches Rust: comparison!(geq, ">=", Ordering::Greater | Ordering::Equal)
func Geq(lhs, rhs foundations.Value) (foundations.Value, error) {
	cmp, err := Compare(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return foundations.Bool(cmp >= 0), nil
}

// Equal determines whether two values are equal.
// Matches Rust: pub fn equal(lhs: &Value, rhs: &Value) -> bool
func Equal(lhs, rhs foundations.Value) bool {
	if lhs == nil && rhs == nil {
		return true
	}
	if lhs == nil || rhs == nil {
		return false
	}

	switch l := lhs.(type) {
	case foundations.NoneValue:
		_, ok := rhs.(foundations.NoneValue)
		return ok
	case foundations.AutoValue:
		_, ok := rhs.(foundations.AutoValue)
		return ok
	case foundations.Bool:
		if r, ok := rhs.(foundations.Bool); ok {
			return l == r
		}
	case foundations.Int:
		switch r := rhs.(type) {
		case foundations.Int:
			return l == r
		case foundations.Float:
			return float64(l) == float64(r)
		}
	case foundations.Float:
		switch r := rhs.(type) {
		case foundations.Float:
			return l == r
		case foundations.Int:
			return float64(l) == float64(r)
		}
	case foundations.Str:
		if r, ok := rhs.(foundations.Str); ok {
			return l == r
		}
	case foundations.LengthValue:
		if r, ok := rhs.(foundations.LengthValue); ok {
			return l.Length.Points == r.Length.Points
		}
		if r, ok := rhs.(foundations.RelativeValue); ok {
			return l.Length.Points == r.Relative.Abs.Points && r.Relative.Rel.Value == 0
		}
	case foundations.AngleValue:
		if r, ok := rhs.(foundations.AngleValue); ok {
			return l.Angle.Radians == r.Angle.Radians
		}
	case foundations.RatioValue:
		if r, ok := rhs.(foundations.RatioValue); ok {
			return l.Ratio.Value == r.Ratio.Value
		}
		if r, ok := rhs.(foundations.RelativeValue); ok {
			return l.Ratio.Value == r.Relative.Rel.Value && r.Relative.Abs.Points == 0
		}
	case foundations.RelativeValue:
		switch r := rhs.(type) {
		case foundations.RelativeValue:
			return l.Relative.Abs.Points == r.Relative.Abs.Points && l.Relative.Rel.Value == r.Relative.Rel.Value
		case foundations.LengthValue:
			return l.Relative.Abs.Points == r.Length.Points && l.Relative.Rel.Value == 0
		case foundations.RatioValue:
			return l.Relative.Rel.Value == r.Ratio.Value && l.Relative.Abs.Points == 0
		}
	case foundations.FractionValue:
		if r, ok := rhs.(foundations.FractionValue); ok {
			return l.Fraction.Value == r.Fraction.Value
		}
	case foundations.LabelValue:
		if r, ok := rhs.(foundations.LabelValue); ok {
			return l == r
		}
	case foundations.SymbolValue:
		if r, ok := rhs.(foundations.SymbolValue); ok {
			return l.Char == r.Char
		}
	case foundations.VersionValue:
		if r, ok := rhs.(foundations.VersionValue); ok {
			return l == r
		}
	case foundations.Duration:
		if r, ok := rhs.(foundations.Duration); ok {
			return l == r
		}
	case *foundations.Array:
		if r, ok := rhs.(*foundations.Array); ok {
			return arrayEqual(l, r)
		}
	case *foundations.Dict:
		if r, ok := rhs.(*foundations.Dict); ok {
			return dictEqual(l, r)
		}
	case foundations.FuncValue:
		if r, ok := rhs.(foundations.FuncValue); ok {
			return l.Func == r.Func
		}
	case foundations.ContentValue:
		if r, ok := rhs.(foundations.ContentValue); ok {
			return contentEqual(l.Content, r.Content)
		}
	}

	return false
}

// Compare compares two values.
// Matches Rust: pub fn compare(lhs: &Value, rhs: &Value) -> StrResult<Ordering>
// Returns -1, 0, or 1 for less, equal, greater.
func Compare(lhs, rhs foundations.Value) (int, error) {
	switch l := lhs.(type) {
	case foundations.Bool:
		if r, ok := rhs.(foundations.Bool); ok {
			return compareBool(bool(l), bool(r)), nil
		}
	case foundations.Int:
		switch r := rhs.(type) {
		case foundations.Int:
			return compareInt(int64(l), int64(r)), nil
		case foundations.Float:
			return compareFloat(float64(l), float64(r)), nil
		}
	case foundations.Float:
		switch r := rhs.(type) {
		case foundations.Int:
			return compareFloat(float64(l), float64(r)), nil
		case foundations.Float:
			return compareFloat(float64(l), float64(r)), nil
		}
	case foundations.Str:
		if r, ok := rhs.(foundations.Str); ok {
			return compareString(string(l), string(r)), nil
		}
	case foundations.LengthValue:
		if r, ok := rhs.(foundations.LengthValue); ok {
			return compareFloat(l.Length.Points, r.Length.Points), nil
		}
	case foundations.AngleValue:
		if r, ok := rhs.(foundations.AngleValue); ok {
			return compareFloat(l.Angle.Radians, r.Angle.Radians), nil
		}
	case foundations.RatioValue:
		if r, ok := rhs.(foundations.RatioValue); ok {
			return compareFloat(l.Ratio.Value, r.Ratio.Value), nil
		}
	case foundations.FractionValue:
		if r, ok := rhs.(foundations.FractionValue); ok {
			return compareFloat(l.Fraction.Value, r.Fraction.Value), nil
		}
	case foundations.VersionValue:
		if r, ok := rhs.(foundations.VersionValue); ok {
			if l.Major != r.Major {
				return compareInt(int64(l.Major), int64(r.Major)), nil
			}
			if l.Minor != r.Minor {
				return compareInt(int64(l.Minor), int64(r.Minor)), nil
			}
			return compareInt(int64(l.Patch), int64(r.Patch)), nil
		}
	case foundations.Duration:
		if r, ok := rhs.(foundations.Duration); ok {
			return compareInt(int64(l), int64(r)), nil
		}
	case *foundations.Array:
		if r, ok := rhs.(*foundations.Array); ok {
			return compareArrays(l, r)
		}
	}

	return 0, mismatch("cannot compare %s and %s", lhs, rhs)
}

func compareBool(a, b bool) int {
	if a == b {
		return 0
	}
	if !a && b {
		return -1
	}
	return 1
}

func compareInt(a, b int64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareFloat(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareString(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareArrays(a, b *foundations.Array) (int, error) {
	aItems := a.Items()
	bItems := b.Items()
	minLen := len(aItems)
	if len(bItems) < minLen {
		minLen = len(bItems)
	}

	for i := 0; i < minLen; i++ {
		cmp, err := Compare(aItems[i], bItems[i])
		if err != nil {
			return 0, err
		}
		if cmp != 0 {
			return cmp, nil
		}
	}

	return compareInt(int64(len(aItems)), int64(len(bItems))), nil
}

// ----------------------------------------------------------------------------
// Containment Operators
// ----------------------------------------------------------------------------

// In tests whether one value is "in" another one.
// Matches Rust: pub fn in_(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func In(lhs, rhs foundations.Value) (foundations.Value, error) {
	if result := Contains(lhs, rhs); result != nil {
		return foundations.Bool(*result), nil
	}
	return nil, mismatch("cannot apply 'in' to %s and %s", lhs, rhs)
}

// NotIn tests whether one value is "not in" another one.
// Matches Rust: pub fn not_in(lhs: Value, rhs: Value) -> HintedStrResult<Value>
func NotIn(lhs, rhs foundations.Value) (foundations.Value, error) {
	if result := Contains(lhs, rhs); result != nil {
		return foundations.Bool(!*result), nil
	}
	return nil, mismatch("cannot apply 'not in' to %s and %s", lhs, rhs)
}

// Contains tests for containment.
// Matches Rust: pub fn contains(lhs: &Value, rhs: &Value) -> Option<bool>
func Contains(lhs, rhs foundations.Value) *bool {
	result := func(b bool) *bool { return &b }

	switch l := lhs.(type) {
	case foundations.Str:
		if r, ok := rhs.(foundations.Str); ok {
			return result(strings.Contains(string(r), string(l)))
		}
		if r, ok := rhs.(*foundations.Dict); ok {
			return result(r.Contains(string(l)))
		}
	}

	if r, ok := rhs.(*foundations.Array); ok {
		return result(r.Contains(lhs))
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helper Functions
// ----------------------------------------------------------------------------

func strRepeat(s foundations.Str, n int) (foundations.Value, error) {
	if n < 0 {
		return nil, fmt.Errorf("cannot repeat string a negative number of times")
	}
	return foundations.Str(strings.Repeat(string(s), n)), nil
}

func arrayRepeat(arr *foundations.Array, n int) (foundations.Value, error) {
	if n < 0 {
		return nil, fmt.Errorf("cannot repeat array a negative number of times")
	}
	result := foundations.NewArray()
	for i := 0; i < n; i++ {
		for _, item := range arr.Items() {
			result.Push(item)
		}
	}
	return result, nil
}

func arrayAdd(a, b *foundations.Array) *foundations.Array {
	result := foundations.NewArray()
	for _, item := range a.Items() {
		result.Push(item)
	}
	for _, item := range b.Items() {
		result.Push(item)
	}
	return result
}

func dictAdd(a, b *foundations.Dict) *foundations.Dict {
	result := foundations.NewDict()
	for _, key := range a.Keys() {
		val, _ := a.Get(key)
		result.Set(key, val)
	}
	for _, key := range b.Keys() {
		val, _ := b.Get(key)
		result.Set(key, val)
	}
	return result
}

func arrayEqual(a, b *foundations.Array) bool {
	if a.Len() != b.Len() {
		return false
	}
	aItems := a.Items()
	bItems := b.Items()
	for i := range aItems {
		if !Equal(aItems[i], bItems[i]) {
			return false
		}
	}
	return true
}

func dictEqual(a, b *foundations.Dict) bool {
	if a.Len() != b.Len() {
		return false
	}
	for _, key := range a.Keys() {
		aVal, _ := a.Get(key)
		bVal, ok := b.Get(key)
		if !ok || !Equal(aVal, bVal) {
			return false
		}
	}
	return true
}

func contentEqual(a, b foundations.Content) bool {
	if len(a.Elements) != len(b.Elements) {
		return false
	}
	// For now, just compare by reference
	// A proper implementation would deep-compare elements
	for i := range a.Elements {
		if a.Elements[i] != b.Elements[i] {
			return false
		}
	}
	return true
}

func joinContent(a, b foundations.Content) foundations.Content {
	elements := make([]foundations.ContentElement, 0, len(a.Elements)+len(b.Elements))
	elements = append(elements, a.Elements...)
	elements = append(elements, b.Elements...)
	return foundations.Content{Elements: elements}
}

func strToContent(s foundations.Str) foundations.Content {
	return foundations.Content{Elements: []foundations.ContentElement{&TextElement{Text: string(s)}}}
}

func symbolToContent(s foundations.SymbolValue) foundations.Content {
	return foundations.Content{Elements: []foundations.ContentElement{&foundations.SymbolElem{Text: string(s.Char)}}}
}
