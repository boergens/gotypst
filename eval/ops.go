package eval

import (
	"math"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Arithmetic Operators
// ----------------------------------------------------------------------------

// Add performs addition on two values.
func Add(lhs, rhs Value, span syntax.Span) (Value, error) {
	switch l := lhs.(type) {
	case IntValue:
		switch r := rhs.(type) {
		case IntValue:
			result, overflow := addInt64(int64(l), int64(r))
			if overflow {
				return nil, &OverflowError{Span: span}
			}
			return Int(result), nil
		case FloatValue:
			return Float(float64(l) + float64(r)), nil
		}

	case FloatValue:
		switch r := rhs.(type) {
		case IntValue:
			return Float(float64(l) + float64(r)), nil
		case FloatValue:
			return Float(float64(l) + float64(r)), nil
		}

	case LengthValue:
		switch r := rhs.(type) {
		case LengthValue:
			return LengthValue{Length: Length{Points: l.Length.Points + r.Length.Points}}, nil
		case RatioValue:
			return RelativeValue{Relative: Relative{Abs: l.Length, Rel: r.Ratio}}, nil
		case RelativeValue:
			return RelativeValue{Relative: Relative{
				Abs: Length{Points: l.Length.Points + r.Relative.Abs.Points},
				Rel: r.Relative.Rel,
			}}, nil
		}

	case RatioValue:
		switch r := rhs.(type) {
		case RatioValue:
			return RatioValue{Ratio: Ratio{Value: l.Ratio.Value + r.Ratio.Value}}, nil
		case LengthValue:
			return RelativeValue{Relative: Relative{Abs: r.Length, Rel: l.Ratio}}, nil
		}

	case RelativeValue:
		switch r := rhs.(type) {
		case LengthValue:
			return RelativeValue{Relative: Relative{
				Abs: Length{Points: l.Relative.Abs.Points + r.Length.Points},
				Rel: l.Relative.Rel,
			}}, nil
		case RatioValue:
			return RelativeValue{Relative: Relative{
				Abs: l.Relative.Abs,
				Rel: Ratio{Value: l.Relative.Rel.Value + r.Ratio.Value},
			}}, nil
		case RelativeValue:
			return RelativeValue{Relative: Relative{
				Abs: Length{Points: l.Relative.Abs.Points + r.Relative.Abs.Points},
				Rel: Ratio{Value: l.Relative.Rel.Value + r.Relative.Rel.Value},
			}}, nil
		}

	case AngleValue:
		if r, ok := rhs.(AngleValue); ok {
			return AngleValue{Angle: Angle{Radians: l.Angle.Radians + r.Angle.Radians}}, nil
		}

	case FractionValue:
		if r, ok := rhs.(FractionValue); ok {
			return FractionValue{Fraction: Fraction{Value: l.Fraction.Value + r.Fraction.Value}}, nil
		}

	case StrValue:
		if r, ok := rhs.(StrValue); ok {
			return Str(string(l) + string(r)), nil
		}

	case ArrayValue:
		if r, ok := rhs.(ArrayValue); ok {
			result := make(ArrayValue, 0, len(l)+len(r))
			result = append(result, l...)
			result = append(result, r...)
			return result, nil
		}

	case *DictValue:
		if r, ok := AsDict(rhs); ok {
			result := NewDict()
			for _, key := range l.Keys() {
				val, _ := l.Get(key)
				result.Set(key, val)
			}
			for _, key := range r.Keys() {
				val, _ := r.Get(key)
				result.Set(key, val)
			}
			return result, nil
		}

	case DictValue:
		if r, ok := AsDict(rhs); ok {
			result := NewDict()
			for _, key := range l.Keys() {
				val, _ := l.Get(key)
				result.Set(key, val)
			}
			for _, key := range r.Keys() {
				val, _ := r.Get(key)
				result.Set(key, val)
			}
			return result, nil
		}

	case ContentValue:
		return ContentValue{Content: appendToContent(l.Content, rhs)}, nil

	case DurationValue:
		if r, ok := rhs.(DurationValue); ok {
			return DurationValue{Nanoseconds: l.Nanoseconds + r.Nanoseconds}, nil
		}
	}

	return nil, &OperatorMismatchError{Op: "+", Lhs: lhs.Type(), Rhs: rhs.Type(), Span: span}
}

// Sub performs subtraction on two values.
func Sub(lhs, rhs Value, span syntax.Span) (Value, error) {
	switch l := lhs.(type) {
	case IntValue:
		switch r := rhs.(type) {
		case IntValue:
			return Int(int64(l) - int64(r)), nil
		case FloatValue:
			return Float(float64(l) - float64(r)), nil
		}

	case FloatValue:
		switch r := rhs.(type) {
		case IntValue:
			return Float(float64(l) - float64(r)), nil
		case FloatValue:
			return Float(float64(l) - float64(r)), nil
		}

	case LengthValue:
		switch r := rhs.(type) {
		case LengthValue:
			return LengthValue{Length: Length{Points: l.Length.Points - r.Length.Points}}, nil
		case RatioValue:
			return RelativeValue{Relative: Relative{Abs: l.Length, Rel: Ratio{Value: -r.Ratio.Value}}}, nil
		}

	case RatioValue:
		switch r := rhs.(type) {
		case RatioValue:
			return RatioValue{Ratio: Ratio{Value: l.Ratio.Value - r.Ratio.Value}}, nil
		case LengthValue:
			return RelativeValue{Relative: Relative{
				Abs: Length{Points: -r.Length.Points},
				Rel: l.Ratio,
			}}, nil
		}

	case RelativeValue:
		switch r := rhs.(type) {
		case LengthValue:
			return RelativeValue{Relative: Relative{
				Abs: Length{Points: l.Relative.Abs.Points - r.Length.Points},
				Rel: l.Relative.Rel,
			}}, nil
		case RatioValue:
			return RelativeValue{Relative: Relative{
				Abs: l.Relative.Abs,
				Rel: Ratio{Value: l.Relative.Rel.Value - r.Ratio.Value},
			}}, nil
		case RelativeValue:
			return RelativeValue{Relative: Relative{
				Abs: Length{Points: l.Relative.Abs.Points - r.Relative.Abs.Points},
				Rel: Ratio{Value: l.Relative.Rel.Value - r.Relative.Rel.Value},
			}}, nil
		}

	case AngleValue:
		if r, ok := rhs.(AngleValue); ok {
			return AngleValue{Angle: Angle{Radians: l.Angle.Radians - r.Angle.Radians}}, nil
		}

	case FractionValue:
		if r, ok := rhs.(FractionValue); ok {
			return FractionValue{Fraction: Fraction{Value: l.Fraction.Value - r.Fraction.Value}}, nil
		}

	case DurationValue:
		if r, ok := rhs.(DurationValue); ok {
			return DurationValue{Nanoseconds: l.Nanoseconds - r.Nanoseconds}, nil
		}
	}

	return nil, &OperatorMismatchError{Op: "-", Lhs: lhs.Type(), Rhs: rhs.Type(), Span: span}
}

// Mul performs multiplication on two values.
func Mul(lhs, rhs Value, span syntax.Span) (Value, error) {
	switch l := lhs.(type) {
	case IntValue:
		switch r := rhs.(type) {
		case IntValue:
			return Int(int64(l) * int64(r)), nil
		case FloatValue:
			return Float(float64(l) * float64(r)), nil
		case LengthValue:
			return LengthValue{Length: Length{Points: float64(l) * r.Length.Points}}, nil
		case AngleValue:
			return AngleValue{Angle: Angle{Radians: float64(l) * r.Angle.Radians}}, nil
		case RatioValue:
			return RatioValue{Ratio: Ratio{Value: float64(l) * r.Ratio.Value}}, nil
		case FractionValue:
			return FractionValue{Fraction: Fraction{Value: float64(l) * r.Fraction.Value}}, nil
		case StrValue:
			// String repetition
			result := ""
			for i := int64(0); i < int64(l); i++ {
				result += string(r)
			}
			return Str(result), nil
		case ArrayValue:
			// Array repetition
			result := make(ArrayValue, 0, len(r)*int(l))
			for i := int64(0); i < int64(l); i++ {
				result = append(result, r...)
			}
			return result, nil
		}

	case FloatValue:
		switch r := rhs.(type) {
		case IntValue:
			return Float(float64(l) * float64(r)), nil
		case FloatValue:
			return Float(float64(l) * float64(r)), nil
		case LengthValue:
			return LengthValue{Length: Length{Points: float64(l) * r.Length.Points}}, nil
		case AngleValue:
			return AngleValue{Angle: Angle{Radians: float64(l) * r.Angle.Radians}}, nil
		case RatioValue:
			return RatioValue{Ratio: Ratio{Value: float64(l) * r.Ratio.Value}}, nil
		case FractionValue:
			return FractionValue{Fraction: Fraction{Value: float64(l) * r.Fraction.Value}}, nil
		}

	case LengthValue:
		switch r := rhs.(type) {
		case IntValue:
			return LengthValue{Length: Length{Points: l.Length.Points * float64(r)}}, nil
		case FloatValue:
			return LengthValue{Length: Length{Points: l.Length.Points * float64(r)}}, nil
		case RatioValue:
			// Length * ratio = relative
			return RelativeValue{Relative: Relative{
				Abs: Length{},
				Rel: Ratio{Value: l.Length.Points * r.Ratio.Value},
			}}, nil
		}

	case RatioValue:
		switch r := rhs.(type) {
		case IntValue:
			return RatioValue{Ratio: Ratio{Value: l.Ratio.Value * float64(r)}}, nil
		case FloatValue:
			return RatioValue{Ratio: Ratio{Value: l.Ratio.Value * float64(r)}}, nil
		case LengthValue:
			return RelativeValue{Relative: Relative{
				Abs: Length{},
				Rel: Ratio{Value: l.Ratio.Value * r.Length.Points},
			}}, nil
		}

	case AngleValue:
		switch r := rhs.(type) {
		case IntValue:
			return AngleValue{Angle: Angle{Radians: l.Angle.Radians * float64(r)}}, nil
		case FloatValue:
			return AngleValue{Angle: Angle{Radians: l.Angle.Radians * float64(r)}}, nil
		}

	case FractionValue:
		switch r := rhs.(type) {
		case IntValue:
			return FractionValue{Fraction: Fraction{Value: l.Fraction.Value * float64(r)}}, nil
		case FloatValue:
			return FractionValue{Fraction: Fraction{Value: l.Fraction.Value * float64(r)}}, nil
		}

	case StrValue:
		if r, ok := rhs.(IntValue); ok {
			result := ""
			for i := int64(0); i < int64(r); i++ {
				result += string(l)
			}
			return Str(result), nil
		}

	case ArrayValue:
		if r, ok := rhs.(IntValue); ok {
			result := make(ArrayValue, 0, len(l)*int(r))
			for i := int64(0); i < int64(r); i++ {
				result = append(result, l...)
			}
			return result, nil
		}
	}

	return nil, &OperatorMismatchError{Op: "*", Lhs: lhs.Type(), Rhs: rhs.Type(), Span: span}
}

// Div performs division on two values.
func Div(lhs, rhs Value, span syntax.Span) (Value, error) {
	// Check for division by zero
	if isZero(rhs) {
		return nil, &DivisionByZeroError{Span: span}
	}

	switch l := lhs.(type) {
	case IntValue:
		switch r := rhs.(type) {
		case IntValue:
			return Float(float64(l) / float64(r)), nil
		case FloatValue:
			return Float(float64(l) / float64(r)), nil
		}

	case FloatValue:
		switch r := rhs.(type) {
		case IntValue:
			return Float(float64(l) / float64(r)), nil
		case FloatValue:
			return Float(float64(l) / float64(r)), nil
		}

	case LengthValue:
		switch r := rhs.(type) {
		case IntValue:
			return LengthValue{Length: Length{Points: l.Length.Points / float64(r)}}, nil
		case FloatValue:
			return LengthValue{Length: Length{Points: l.Length.Points / float64(r)}}, nil
		case LengthValue:
			// Length / length = ratio
			return Float(l.Length.Points / r.Length.Points), nil
		}

	case RatioValue:
		switch r := rhs.(type) {
		case IntValue:
			return RatioValue{Ratio: Ratio{Value: l.Ratio.Value / float64(r)}}, nil
		case FloatValue:
			return RatioValue{Ratio: Ratio{Value: l.Ratio.Value / float64(r)}}, nil
		}

	case AngleValue:
		switch r := rhs.(type) {
		case IntValue:
			return AngleValue{Angle: Angle{Radians: l.Angle.Radians / float64(r)}}, nil
		case FloatValue:
			return AngleValue{Angle: Angle{Radians: l.Angle.Radians / float64(r)}}, nil
		case AngleValue:
			// Angle / angle = ratio
			return Float(l.Angle.Radians / r.Angle.Radians), nil
		}

	case FractionValue:
		switch r := rhs.(type) {
		case IntValue:
			return FractionValue{Fraction: Fraction{Value: l.Fraction.Value / float64(r)}}, nil
		case FloatValue:
			return FractionValue{Fraction: Fraction{Value: l.Fraction.Value / float64(r)}}, nil
		}
	}

	return nil, &OperatorMismatchError{Op: "/", Lhs: lhs.Type(), Rhs: rhs.Type(), Span: span}
}

// isZero checks if a value is zero.
func isZero(v Value) bool {
	switch val := v.(type) {
	case IntValue:
		return val == 0
	case FloatValue:
		return val == 0
	default:
		return false
	}
}

// ----------------------------------------------------------------------------
// Comparison Operators
// ----------------------------------------------------------------------------

// Equal checks if two values are equal.
func Equal(lhs, rhs Value) bool {
	// Handle nil values
	if lhs == nil && rhs == nil {
		return true
	}
	if lhs == nil || rhs == nil {
		return false
	}

	// Type must match for equality
	if lhs.Type() != rhs.Type() {
		// Special case: int and float can be compared
		if lhsNum, ok := AsFloat(lhs); ok {
			if rhsNum, ok := AsFloat(rhs); ok {
				return lhsNum == rhsNum
			}
		}
		return false
	}

	switch l := lhs.(type) {
	case NoneValue:
		return true // All none values are equal
	case AutoValue:
		return true // All auto values are equal
	case BoolValue:
		return l == rhs.(BoolValue)
	case IntValue:
		return l == rhs.(IntValue)
	case FloatValue:
		return l == rhs.(FloatValue)
	case StrValue:
		return l == rhs.(StrValue)
	case LabelValue:
		return l == rhs.(LabelValue)
	case LengthValue:
		return l.Length.Points == rhs.(LengthValue).Length.Points
	case AngleValue:
		return l.Angle.Radians == rhs.(AngleValue).Angle.Radians
	case RatioValue:
		return l.Ratio.Value == rhs.(RatioValue).Ratio.Value
	case FractionValue:
		return l.Fraction.Value == rhs.(FractionValue).Fraction.Value

	case ArrayValue:
		r := rhs.(ArrayValue)
		if len(l) != len(r) {
			return false
		}
		for i := range l {
			if !Equal(l[i], r[i]) {
				return false
			}
		}
		return true

	case *DictValue:
		r, ok := rhs.(*DictValue)
		if !ok {
			if r2, ok := rhs.(DictValue); ok {
				r = &r2
			} else {
				return false
			}
		}
		if l.Len() != r.Len() {
			return false
		}
		for _, key := range l.Keys() {
			lVal, _ := l.Get(key)
			rVal, ok := r.Get(key)
			if !ok || !Equal(lVal, rVal) {
				return false
			}
		}
		return true

	case DictValue:
		return Equal(&l, rhs)

	case FuncValue:
		// Functions are equal only if they're the same function
		r := rhs.(FuncValue)
		return l.Func == r.Func

	default:
		// For other types, use pointer equality
		return false
	}
}

// Compare compares two values and returns whether the comparison predicate holds.
func Compare(lhs, rhs Value, pred func(int) bool, span syntax.Span) (Value, error) {
	cmp, err := compareValues(lhs, rhs)
	if err != nil {
		return nil, &OperatorMismatchError{Op: "comparison", Lhs: lhs.Type(), Rhs: rhs.Type(), Span: span}
	}
	return Bool(pred(cmp)), nil
}

// compareValues compares two values, returning -1, 0, or 1.
func compareValues(lhs, rhs Value) (int, error) {
	switch l := lhs.(type) {
	case IntValue:
		switch r := rhs.(type) {
		case IntValue:
			return compareInt(int64(l), int64(r)), nil
		case FloatValue:
			return compareFloat(float64(l), float64(r)), nil
		}

	case FloatValue:
		switch r := rhs.(type) {
		case IntValue:
			return compareFloat(float64(l), float64(r)), nil
		case FloatValue:
			return compareFloat(float64(l), float64(r)), nil
		}

	case StrValue:
		if r, ok := rhs.(StrValue); ok {
			return compareString(string(l), string(r)), nil
		}

	case LengthValue:
		if r, ok := rhs.(LengthValue); ok {
			return compareFloat(l.Length.Points, r.Length.Points), nil
		}

	case AngleValue:
		if r, ok := rhs.(AngleValue); ok {
			return compareFloat(l.Angle.Radians, r.Angle.Radians), nil
		}

	case RatioValue:
		if r, ok := rhs.(RatioValue); ok {
			return compareFloat(l.Ratio.Value, r.Ratio.Value), nil
		}

	case FractionValue:
		if r, ok := rhs.(FractionValue); ok {
			return compareFloat(l.Fraction.Value, r.Fraction.Value), nil
		}

	case DurationValue:
		if r, ok := rhs.(DurationValue); ok {
			return compareInt(l.Nanoseconds, r.Nanoseconds), nil
		}

	case VersionValue:
		if r, ok := rhs.(VersionValue); ok {
			if l.Major != r.Major {
				return compareInt(int64(l.Major), int64(r.Major)), nil
			}
			if l.Minor != r.Minor {
				return compareInt(int64(l.Minor), int64(r.Minor)), nil
			}
			return compareInt(int64(l.Patch), int64(r.Patch)), nil
		}
	}

	return 0, &OperatorMismatchError{Op: "comparison", Lhs: lhs.Type(), Rhs: rhs.Type()}
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

// ----------------------------------------------------------------------------
// Membership Operators
// ----------------------------------------------------------------------------

// Contains checks if a container contains a value.
func Contains(container, value Value, span syntax.Span) (Value, error) {
	switch c := container.(type) {
	case StrValue:
		if v, ok := value.(StrValue); ok {
			// Check if string contains substring
			for i := 0; i <= len(c)-len(v); i++ {
				if string(c[i:i+len(v)]) == string(v) {
					return True, nil
				}
			}
			return False, nil
		}

	case ArrayValue:
		for _, elem := range c {
			if Equal(elem, value) {
				return True, nil
			}
		}
		return False, nil

	case *DictValue:
		if key, ok := AsStr(value); ok {
			_, exists := c.Get(key)
			return Bool(exists), nil
		}

	case DictValue:
		if key, ok := AsStr(value); ok {
			_, exists := c.Get(key)
			return Bool(exists), nil
		}
	}

	return nil, &OperatorMismatchError{Op: "in", Lhs: value.Type(), Rhs: container.Type(), Span: span}
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// OperatorMismatchError is returned when operator types don't match.
type OperatorMismatchError struct {
	Op   string
	Lhs  Type
	Rhs  Type
	Span syntax.Span
}

func (e *OperatorMismatchError) Error() string {
	// Use verb form for arithmetic operators
	var verb string
	switch e.Op {
	case "+":
		verb = "add"
	case "-":
		verb = "subtract"
	case "*":
		verb = "multiply"
	case "/":
		verb = "divide"
	default:
		return "cannot apply operator " + e.Op + " to " + e.Lhs.FullName() + " and " + e.Rhs.FullName()
	}
	return "cannot " + verb + " " + e.Lhs.FullName() + " and " + e.Rhs.FullName()
}

// DivisionByZeroError is returned when dividing by zero.
type DivisionByZeroError struct {
	Span syntax.Span
}

func (e *DivisionByZeroError) Error() string {
	return "cannot divide by zero"
}

// OverflowError is returned when an arithmetic operation overflows.
type OverflowError struct {
	Span syntax.Span
}

func (e *OverflowError) Error() string {
	return "value is too large"
}

// addInt64 adds two int64 values and returns the result and whether it overflowed.
func addInt64(a, b int64) (int64, bool) {
	result := a + b
	// Check for overflow: if both inputs have the same sign and the result
	// has a different sign, then we overflowed
	if (a > 0 && b > 0 && result < 0) || (a < 0 && b < 0 && result > 0) {
		return 0, true
	}
	// Also check if the result is outside the valid int64 range
	// (this catches the edge case of MinInt64 + -1)
	if a > 0 && b > 0 && result > math.MaxInt64 {
		return 0, true
	}
	return result, false
}
