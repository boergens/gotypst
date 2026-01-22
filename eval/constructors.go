package eval

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/boergens/gotypst/syntax"
)

// ConstructorError is returned when a type constructor fails.
type ConstructorError struct {
	Message string
	Span    syntax.Span
	Hints   []string
}

func (e *ConstructorError) Error() string {
	return e.Message
}

// typeConstructorFunc returns a constructor function for the given type.
func typeConstructorFunc(t Type) *Func {
	name := t.String()
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(engine *Engine, context *Context, args *Args) (Value, error) {
				return callTypeConstructor(t, args)
			},
		},
	}
}

// callTypeConstructor dispatches to the appropriate type constructor.
func callTypeConstructor(t Type, args *Args) (Value, error) {
	switch t {
	case TypeInt:
		return intConstructor(args)
	case TypeStr:
		return strConstructor(args)
	case TypeFloat:
		return floatConstructor(args)
	case TypeBool:
		return boolConstructor(args)
	case TypeArray:
		return arrayConstructor(args)
	case TypeDict:
		return dictConstructor(args)
	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("type %s is not callable as a constructor", t.String()),
			Span:    args.Span,
		}
	}
}

// ----------------------------------------------------------------------------
// Int Constructor
// ----------------------------------------------------------------------------

// intConstructor converts a value to an integer.
// Supports: bool, int, float, decimal, string (with optional base)
func intConstructor(args *Args) (Value, error) {
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
		if _, ok := value.(StrValue); !ok {
			return nil, &ConstructorError{
				Message: "base is only supported for strings",
				Span:    baseArg.Span,
			}
		}
		baseVal, ok := baseArg.V.(IntValue)
		if !ok {
			return nil, &TypeError{
				Expected: TypeInt,
				Got:      baseArg.V.Type(),
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

	switch v := value.(type) {
	case IntValue:
		return v, nil

	case BoolValue:
		if v {
			return IntValue(1), nil
		}
		return IntValue(0), nil

	case FloatValue:
		return IntValue(int64(v)), nil

	case DecimalValue:
		// Truncate decimal to integer
		if v.Value == nil {
			return IntValue(0), nil
		}
		f, _ := v.Value.Float64()
		return IntValue(int64(f)), nil

	case StrValue:
		s := string(v)
		if s == "" {
			return nil, &ConstructorError{
				Message: "string must not be empty",
				Span:    spanned.Span,
			}
		}

		// Handle Unicode minus sign
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
							Span:    spanned.Span,
							Hints:   []string{"value does not fit into a signed 64-bit integer", "try using a floating point number"},
						}
					}
					return nil, &ConstructorError{
						Message: "integer value is too large",
						Span:    spanned.Span,
						Hints:   []string{"value does not fit into a signed 64-bit integer", "try using a floating point number"},
					}
				case strconv.ErrSyntax:
					if base != 10 {
						return nil, &ConstructorError{
							Message: fmt.Sprintf("string contains invalid digits for a base %d integer", base),
							Span:    spanned.Span,
						}
					}
					return nil, &ConstructorError{
						Message: "string contains invalid digits",
						Span:    spanned.Span,
					}
				}
			}
			return nil, &ConstructorError{
				Message: "string contains invalid digits",
				Span:    spanned.Span,
			}
		}
		return IntValue(result), nil

	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("expected integer, boolean, float, decimal, or string, found %s", value.Type().String()),
			Span:    spanned.Span,
		}
	}
}

// ----------------------------------------------------------------------------
// Str Constructor
// ----------------------------------------------------------------------------

// strConstructor converts a value to a string.
func strConstructor(args *Args) (Value, error) {
	// Check for from-unicode call: str.from-unicode(n)
	// This is handled separately via method access

	// Get the value to convert
	spanned, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	value := spanned.V

	// Check for base argument (only valid for integers)
	base := 10
	if baseArg := args.Find("base"); baseArg != nil {
		baseVal, ok := baseArg.V.(IntValue)
		if !ok {
			return nil, &TypeError{
				Expected: TypeInt,
				Got:      baseArg.V.Type(),
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

	switch v := value.(type) {
	case StrValue:
		return v, nil

	case IntValue:
		if base != 10 {
			return StrValue(strconv.FormatInt(int64(v), base)), nil
		}
		return StrValue(strconv.FormatInt(int64(v), 10)), nil

	case FloatValue:
		return StrValue(strconv.FormatFloat(float64(v), 'g', -1, 64)), nil

	case DecimalValue:
		if v.Value == nil {
			return StrValue("0"), nil
		}
		return StrValue(v.Value.FloatString(10)), nil

	case VersionValue:
		return StrValue(fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)), nil

	case BytesValue:
		return StrValue(string(v)), nil

	case LabelValue:
		return StrValue(string(v)), nil

	case TypeValue:
		return StrValue(v.Inner.String()), nil

	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("expected integer, float, decimal, version, bytes, label, type, or string, found %s", value.Type().String()),
			Span:    spanned.Span,
		}
	}
}

// ----------------------------------------------------------------------------
// Float Constructor
// ----------------------------------------------------------------------------

// floatConstructor converts a value to a float.
func floatConstructor(args *Args) (Value, error) {
	spanned, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	value := spanned.V

	switch v := value.(type) {
	case FloatValue:
		return v, nil

	case IntValue:
		return FloatValue(float64(v)), nil

	case BoolValue:
		if v {
			return FloatValue(1.0), nil
		}
		return FloatValue(0.0), nil

	case DecimalValue:
		if v.Value == nil {
			return FloatValue(0), nil
		}
		f, _ := v.Value.Float64()
		return FloatValue(f), nil

	case StrValue:
		s := string(v)
		if s == "" {
			return nil, &ConstructorError{
				Message: "string must not be empty",
				Span:    spanned.Span,
			}
		}
		// Handle special values
		switch strings.ToLower(s) {
		case "inf", "+inf":
			return FloatValue(math.Inf(1)), nil
		case "-inf":
			return FloatValue(math.Inf(-1)), nil
		case "nan":
			return FloatValue(math.NaN()), nil
		}
		// Handle Unicode minus sign
		s = strings.ReplaceAll(s, "\u2212", "-")
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, &ConstructorError{
				Message: "string contains invalid float",
				Span:    spanned.Span,
			}
		}
		return FloatValue(f), nil

	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("expected integer, boolean, float, decimal, or string, found %s", value.Type().String()),
			Span:    spanned.Span,
		}
	}
}

// ----------------------------------------------------------------------------
// Bool Constructor
// ----------------------------------------------------------------------------

// boolConstructor converts a value to a boolean.
func boolConstructor(args *Args) (Value, error) {
	spanned, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	value := spanned.V

	switch v := value.(type) {
	case BoolValue:
		return v, nil

	case IntValue:
		return BoolValue(v != 0), nil

	case StrValue:
		s := strings.ToLower(string(v))
		switch s {
		case "true", "1", "yes":
			return BoolValue(true), nil
		case "false", "0", "no":
			return BoolValue(false), nil
		default:
			return nil, &ConstructorError{
				Message: fmt.Sprintf("cannot convert string %q to boolean", s),
				Span:    spanned.Span,
			}
		}

	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("expected boolean, integer, or string, found %s", value.Type().String()),
			Span:    spanned.Span,
		}
	}
}

// ----------------------------------------------------------------------------
// Array Constructor
// ----------------------------------------------------------------------------

// arrayConstructor creates an array from arguments.
func arrayConstructor(args *Args) (Value, error) {
	var result ArrayValue
	for _, arg := range args.Items {
		if arg.Name != nil {
			return nil, &ConstructorError{
				Message: "array constructor does not accept named arguments",
				Span:    arg.Span,
			}
		}
		result = append(result, arg.Value.V)
	}
	return result, nil
}

// ----------------------------------------------------------------------------
// Dict Constructor
// ----------------------------------------------------------------------------

// dictConstructor creates a dictionary from arguments.
func dictConstructor(args *Args) (Value, error) {
	result := NewDict()
	for _, arg := range args.Items {
		if arg.Name == nil {
			return nil, &ConstructorError{
				Message: "dictionary constructor only accepts named arguments",
				Span:    arg.Span,
			}
		}
		result.Set(*arg.Name, arg.Value.V)
	}
	return &result, nil
}

// ----------------------------------------------------------------------------
// Type Methods (static methods on types)
// ----------------------------------------------------------------------------

// GetTypeMethod returns a static method for a type.
func GetTypeMethod(t Type, name string, span syntax.Span) Value {
	switch t {
	case TypeStr:
		return getStrTypeMethod(name, span)
	default:
		return nil
	}
}

// getStrTypeMethod returns a static method for the str type.
func getStrTypeMethod(name string, span syntax.Span) Value {
	switch name {
	case "from-unicode":
		return FuncValue{Func: strFromUnicodeFunc()}
	case "to-unicode":
		return FuncValue{Func: strToUnicodeFunc()}
	default:
		return nil
	}
}

// strFromUnicodeFunc returns a function that creates a string from a Unicode codepoint.
func strFromUnicodeFunc() *Func {
	name := "from-unicode"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(engine *Engine, context *Context, args *Args) (Value, error) {
				spanned, err := args.Expect("codepoint")
				if err != nil {
					return nil, err
				}
				value := spanned.V

				intVal, ok := value.(IntValue)
				if !ok {
					return nil, &ConstructorError{
						Message: fmt.Sprintf("expected integer, found %s", value.Type().String()),
						Span:    spanned.Span,
					}
				}

				n := int64(intVal)
				if n < 0 {
					return nil, &ConstructorError{
						Message: "number must be at least zero",
						Span:    spanned.Span,
					}
				}
				if n > 0x10FFFF {
					return nil, &ConstructorError{
						Message: fmt.Sprintf("0x%x is not a valid codepoint", n),
						Span:    spanned.Span,
					}
				}

				return StrValue(string(rune(n))), nil
			},
		},
	}
}

// strToUnicodeFunc returns a function that gets the Unicode codepoint of a character.
func strToUnicodeFunc() *Func {
	name := "to-unicode"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(engine *Engine, context *Context, args *Args) (Value, error) {
				spanned, err := args.Expect("character")
				if err != nil {
					return nil, err
				}
				value := spanned.V

				strVal, ok := value.(StrValue)
				if !ok {
					return nil, &ConstructorError{
						Message: fmt.Sprintf("expected string, found %s", value.Type().String()),
						Span:    spanned.Span,
					}
				}

				s := string(strVal)
				if utf8.RuneCountInString(s) != 1 {
					return nil, &ConstructorError{
						Message: "expected exactly one character",
						Span:    spanned.Span,
					}
				}

				r, _ := utf8.DecodeRuneInString(s)
				return IntValue(int64(r)), nil
			},
		},
	}
}
