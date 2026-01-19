package foundations

import (
	"cmp"
	"fmt"
	"math"
	"strings"
)

// OpError represents an error that occurred during an operation.
type OpError struct {
	Message string
	Hint    string
}

func (e *OpError) Error() string {
	if e.Hint != "" {
		return e.Message + " (hint: " + e.Hint + ")"
	}
	return e.Message
}

// mismatch creates an error for type mismatches in operations.
func mismatch(op string, lhs, rhs Value) *OpError {
	if rhs == nil {
		return &OpError{Message: fmt.Sprintf("cannot apply %s to %s", op, lhs.Type())}
	}
	return &OpError{Message: fmt.Sprintf("cannot apply %s to %s and %s", op, lhs.Type(), rhs.Type())}
}

// --- Unary Operators ---

// Pos applies the unary positive operator.
// Returns the value unchanged for numeric types.
func Pos(v Value) (Value, error) {
	switch x := v.(type) {
	case Int:
		return x, nil
	case Float:
		return x, nil
	default:
		return nil, mismatch("unary '+'", v, nil)
	}
}

// Neg applies the unary negation operator.
func Neg(v Value) (Value, error) {
	switch x := v.(type) {
	case Int:
		// Check for overflow (most negative int64 cannot be negated)
		if x == math.MinInt64 {
			return nil, &OpError{Message: "integer overflow"}
		}
		return Int(-x), nil
	case Float:
		return Float(-x), nil
	default:
		return nil, mismatch("unary '-'", v, nil)
	}
}

// Not applies the logical not operator.
func Not(v Value) (Value, error) {
	switch x := v.(type) {
	case Bool:
		return Bool(!x), nil
	default:
		return nil, mismatch("'not'", v, nil)
	}
}

// --- Binary Arithmetic Operators ---

// Add adds two values.
func Add(lhs, rhs Value) (Value, error) {
	switch a := lhs.(type) {
	case Int:
		switch b := rhs.(type) {
		case Int:
			result, ok := checkedAddInt(int64(a), int64(b))
			if !ok {
				return nil, &OpError{Message: "integer overflow"}
			}
			return Int(result), nil
		case Float:
			return Float(float64(a) + float64(b)), nil
		}
	case Float:
		switch b := rhs.(type) {
		case Int:
			return Float(float64(a) + float64(b)), nil
		case Float:
			return Float(a + b), nil
		}
	case Str:
		switch b := rhs.(type) {
		case Str:
			return Str(string(a) + string(b)), nil
		}
	case *Array:
		switch b := rhs.(type) {
		case *Array:
			// Concatenate arrays
			items := make([]Value, 0, a.Len()+b.Len())
			items = append(items, a.Items()...)
			items = append(items, b.Items()...)
			return &Array{items: items}, nil
		}
	case *Dict:
		switch b := rhs.(type) {
		case *Dict:
			// Merge dictionaries (b's values override a's)
			result := NewDict()
			for i, k := range a.keys {
				result.Set(k, a.values[i])
			}
			for i, k := range b.keys {
				result.Set(k, b.values[i])
			}
			return result, nil
		}
	}
	return nil, mismatch("'+'", lhs, rhs)
}

// Sub subtracts two values.
func Sub(lhs, rhs Value) (Value, error) {
	switch a := lhs.(type) {
	case Int:
		switch b := rhs.(type) {
		case Int:
			result, ok := checkedSubInt(int64(a), int64(b))
			if !ok {
				return nil, &OpError{Message: "integer overflow"}
			}
			return Int(result), nil
		case Float:
			return Float(float64(a) - float64(b)), nil
		}
	case Float:
		switch b := rhs.(type) {
		case Int:
			return Float(float64(a) - float64(b)), nil
		case Float:
			return Float(a - b), nil
		}
	}
	return nil, mismatch("'-'", lhs, rhs)
}

// Mul multiplies two values.
func Mul(lhs, rhs Value) (Value, error) {
	switch a := lhs.(type) {
	case Int:
		switch b := rhs.(type) {
		case Int:
			result, ok := checkedMulInt(int64(a), int64(b))
			if !ok {
				return nil, &OpError{Message: "integer overflow"}
			}
			return Int(result), nil
		case Float:
			return Float(float64(a) * float64(b)), nil
		case Str:
			return repeatStr(b, int64(a))
		case *Array:
			return repeatArray(b, int64(a))
		}
	case Float:
		switch b := rhs.(type) {
		case Int:
			return Float(float64(a) * float64(b)), nil
		case Float:
			return Float(a * b), nil
		}
	case Str:
		switch b := rhs.(type) {
		case Int:
			return repeatStr(a, int64(b))
		}
	case *Array:
		switch b := rhs.(type) {
		case Int:
			return repeatArray(a, int64(b))
		}
	}
	return nil, mismatch("'*'", lhs, rhs)
}

// Div divides two values.
func Div(lhs, rhs Value) (Value, error) {
	switch a := lhs.(type) {
	case Int:
		switch b := rhs.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			// Integer division returns float in Typst
			return Float(float64(a) / float64(b)), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			return Float(float64(a) / float64(b)), nil
		}
	case Float:
		switch b := rhs.(type) {
		case Int:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			return Float(float64(a) / float64(b)), nil
		case Float:
			if b == 0 {
				return nil, &OpError{Message: "cannot divide by zero"}
			}
			return Float(a / b), nil
		}
	}
	return nil, mismatch("'/'", lhs, rhs)
}

// --- Logical Operators ---

// And performs logical and.
func And(lhs, rhs Value) (Value, error) {
	a, ok := lhs.(Bool)
	if !ok {
		return nil, mismatch("'and'", lhs, rhs)
	}
	b, ok := rhs.(Bool)
	if !ok {
		return nil, mismatch("'and'", lhs, rhs)
	}
	return Bool(a && b), nil
}

// Or performs logical or.
func Or(lhs, rhs Value) (Value, error) {
	a, ok := lhs.(Bool)
	if !ok {
		return nil, mismatch("'or'", lhs, rhs)
	}
	b, ok := rhs.(Bool)
	if !ok {
		return nil, mismatch("'or'", lhs, rhs)
	}
	return Bool(a || b), nil
}

// --- Comparison Operators ---

// Eq checks if two values are equal.
func Eq(lhs, rhs Value) (Value, error) {
	return Bool(equal(lhs, rhs)), nil
}

// Neq checks if two values are not equal.
func Neq(lhs, rhs Value) (Value, error) {
	return Bool(!equal(lhs, rhs)), nil
}

// Lt checks if lhs < rhs.
func Lt(lhs, rhs Value) (Value, error) {
	result, err := compare(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return Bool(result < 0), nil
}

// Leq checks if lhs <= rhs.
func Leq(lhs, rhs Value) (Value, error) {
	result, err := compare(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return Bool(result <= 0), nil
}

// Gt checks if lhs > rhs.
func Gt(lhs, rhs Value) (Value, error) {
	result, err := compare(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return Bool(result > 0), nil
}

// Geq checks if lhs >= rhs.
func Geq(lhs, rhs Value) (Value, error) {
	result, err := compare(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return Bool(result >= 0), nil
}

// --- Membership Operators ---

// In checks if lhs is contained in rhs.
func In(lhs, rhs Value) (Value, error) {
	result, err := contains(rhs, lhs)
	if err != nil {
		return nil, err
	}
	return Bool(result), nil
}

// NotIn checks if lhs is not contained in rhs.
func NotIn(lhs, rhs Value) (Value, error) {
	result, err := contains(rhs, lhs)
	if err != nil {
		return nil, err
	}
	return Bool(!result), nil
}

// --- Join Operation ---

// Join concatenates compatible values.
// Used for joining content in markup mode.
func Join(lhs, rhs Value) (Value, error) {
	// None values are absorbed
	if _, ok := lhs.(NoneValue); ok {
		return rhs, nil
	}
	if _, ok := rhs.(NoneValue); ok {
		return lhs, nil
	}

	switch a := lhs.(type) {
	case Str:
		switch b := rhs.(type) {
		case Str:
			return Str(string(a) + string(b)), nil
		}
	case *Array:
		switch b := rhs.(type) {
		case *Array:
			items := make([]Value, 0, a.Len()+b.Len())
			items = append(items, a.Items()...)
			items = append(items, b.Items()...)
			return &Array{items: items}, nil
		}
	case *Dict:
		switch b := rhs.(type) {
		case *Dict:
			result := NewDict()
			for i, k := range a.keys {
				result.Set(k, a.values[i])
			}
			for i, k := range b.keys {
				result.Set(k, b.values[i])
			}
			return result, nil
		}
	}

	// Default: try Add
	return Add(lhs, rhs)
}

// --- Helper Functions ---

// equal checks deep equality between values.
func equal(lhs, rhs Value) bool {
	switch a := lhs.(type) {
	case NoneValue:
		_, ok := rhs.(NoneValue)
		return ok
	case AutoValue:
		_, ok := rhs.(AutoValue)
		return ok
	case Bool:
		b, ok := rhs.(Bool)
		return ok && a == b
	case Int:
		switch b := rhs.(type) {
		case Int:
			return a == b
		case Float:
			return float64(a) == float64(b)
		}
	case Float:
		switch b := rhs.(type) {
		case Int:
			return float64(a) == float64(b)
		case Float:
			return a == b
		}
	case Str:
		b, ok := rhs.(Str)
		return ok && a == b
	case *Array:
		b, ok := rhs.(*Array)
		if !ok || a.Len() != b.Len() {
			return false
		}
		for i := range a.items {
			if !equal(a.items[i], b.items[i]) {
				return false
			}
		}
		return true
	case *Dict:
		b, ok := rhs.(*Dict)
		if !ok || a.Len() != b.Len() {
			return false
		}
		for i, k := range a.keys {
			v, exists := b.Get(k)
			if !exists || !equal(a.values[i], v) {
				return false
			}
		}
		return true
	}
	return false
}

// compare returns -1, 0, or 1 comparing lhs to rhs.
func compare(lhs, rhs Value) (int, error) {
	switch a := lhs.(type) {
	case Int:
		switch b := rhs.(type) {
		case Int:
			return cmp.Compare(int64(a), int64(b)), nil
		case Float:
			return cmp.Compare(float64(a), float64(b)), nil
		}
	case Float:
		switch b := rhs.(type) {
		case Int:
			return cmp.Compare(float64(a), float64(b)), nil
		case Float:
			return cmp.Compare(float64(a), float64(b)), nil
		}
	case Str:
		b, ok := rhs.(Str)
		if ok {
			return strings.Compare(string(a), string(b)), nil
		}
	}
	return 0, &OpError{
		Message: fmt.Sprintf("cannot compare %s with %s", lhs.Type(), rhs.Type()),
	}
}

// contains checks if container contains element.
func contains(container, element Value) (bool, error) {
	switch c := container.(type) {
	case Str:
		e, ok := element.(Str)
		if !ok {
			return false, &OpError{
				Message: fmt.Sprintf("string cannot contain %s", element.Type()),
			}
		}
		return strings.Contains(string(c), string(e)), nil
	case *Array:
		for _, item := range c.Items() {
			if equal(item, element) {
				return true, nil
			}
		}
		return false, nil
	case *Dict:
		e, ok := element.(Str)
		if !ok {
			return false, &OpError{
				Message: fmt.Sprintf("dictionary keys must be strings, got %s", element.Type()),
			}
		}
		return c.Contains(string(e)), nil
	default:
		return false, &OpError{
			Message: fmt.Sprintf("cannot check membership in %s", container.Type()),
		}
	}
}

// repeatStr repeats a string n times.
func repeatStr(s Str, n int64) (Value, error) {
	if n < 0 {
		return nil, &OpError{Message: "cannot repeat a negative number of times"}
	}
	if n == 0 {
		return Str(""), nil
	}
	return Str(strings.Repeat(string(s), int(n))), nil
}

// repeatArray repeats an array n times.
func repeatArray(a *Array, n int64) (Value, error) {
	if n < 0 {
		return nil, &OpError{Message: "cannot repeat a negative number of times"}
	}
	if n == 0 {
		return NewArray(), nil
	}
	items := make([]Value, 0, a.Len()*int(n))
	for i := int64(0); i < n; i++ {
		items = append(items, a.Items()...)
	}
	return &Array{items: items}, nil
}

// Checked integer arithmetic to detect overflow.

func checkedAddInt(a, b int64) (int64, bool) {
	result := a + b
	// Overflow if signs are same and result sign differs
	if (a > 0 && b > 0 && result < 0) || (a < 0 && b < 0 && result > 0) {
		return 0, false
	}
	return result, true
}

func checkedSubInt(a, b int64) (int64, bool) {
	result := a - b
	// Overflow if signs differ and result sign differs from a
	if (a > 0 && b < 0 && result < 0) || (a < 0 && b > 0 && result > 0) {
		return 0, false
	}
	return result, true
}

func checkedMulInt(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	result := a * b
	if result/a != b {
		return 0, false
	}
	return result, true
}

// IsZero checks if a value is a zero value.
func IsZero(v Value) bool {
	switch x := v.(type) {
	case Int:
		return x == 0
	case Float:
		return x == 0
	default:
		return false
	}
}
