// Type conversion utilities for element parsing.
// This file provides the conversion layer between Typst Values and Go types.

package foundations

import (
	"fmt"
	"reflect"

	"github.com/boergens/gotypst/syntax"
)

// TypeMismatchError is returned when a value doesn't match the expected type.
type TypeMismatchError struct {
	Expected string
	Got      string
	Field    string      // optional field name for better error messages
	Span     syntax.Span // source location
}

func (e *TypeMismatchError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("expected %s for field %q, got %s", e.Expected, e.Field, e.Got)
	}
	return fmt.Sprintf("expected %s, got %s", e.Expected, e.Got)
}

// ConstructorError is returned when a type constructor fails.
type ConstructorError struct {
	Message string
	Span    syntax.Span
	Hints   []string
}

func (e *ConstructorError) Error() string {
	return e.Message
}

// ConvertValue converts a Typst Value to a Go value based on the target type.
// Returns the converted value and any error.
//
// For optional fields (pointer types), returns nil if the value is None.
// For required fields, None is an error.
func ConvertValue(v Value, targetType Type, goType reflect.Type) (any, error) {
	// Handle none - returns nil for pointer types
	if IsNone(v) {
		if goType.Kind() == reflect.Ptr {
			return reflect.Zero(goType).Interface(), nil
		}
		return nil, &TypeMismatchError{Expected: targetType.String(), Got: "none"}
	}

	// Handle auto - passes through as AutoValue for types that accept it
	if IsAuto(v) {
		// Auto is valid for certain types - let the caller handle it
		// We return the AutoValue itself
		return v, nil
	}

	switch targetType {
	case TypeLength:
		return convertToLength(v, goType)
	case TypeContent:
		return convertToContent(v, goType)
	case TypeInt:
		return convertToInt(v, goType)
	case TypeFloat:
		return convertToFloat(v, goType)
	case TypeStr:
		return convertToStr(v, goType)
	case TypeBool:
		return convertToBool(v, goType)
	case TypeColor:
		return convertToColor(v, goType)
	case TypeRatio:
		return convertToRatio(v, goType)
	case TypeRelative:
		return convertToRelative(v, goType)
	case TypeAngle:
		return convertToAngle(v, goType)
	case TypeFraction:
		return convertToFraction(v, goType)
	case TypeArray:
		return convertToArray(v, goType)
	case TypeDict:
		return convertToDict(v, goType)
	case TypeFunc:
		return convertToFunc(v, goType)
	case TypeLabel:
		return convertToLabel(v, goType)
	default:
		// For unknown types, try direct type assertion
		return v, nil
	}
}

// convertToLength converts a Value to a Length.
func convertToLength(v Value, goType reflect.Type) (any, error) {
	lv, ok := v.(LengthValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "length", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &lv.Length, nil
	}
	return lv.Length, nil
}

// convertToContent converts a Value to Content.
func convertToContent(v Value, goType reflect.Type) (any, error) {
	cv, ok := v.(ContentValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "content", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &cv.Content, nil
	}
	return cv.Content, nil
}

// convertToInt converts a Value to int64.
func convertToInt(v Value, goType reflect.Type) (any, error) {
	i, ok := AsInt(v)
	if !ok {
		return nil, &TypeMismatchError{Expected: "integer", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &i, nil
	}
	return i, nil
}

// convertToFloat converts a Value to float64.
func convertToFloat(v Value, goType reflect.Type) (any, error) {
	f, ok := AsFloat(v)
	if !ok {
		return nil, &TypeMismatchError{Expected: "float", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &f, nil
	}
	return f, nil
}

// convertToStr converts a Value to string.
func convertToStr(v Value, goType reflect.Type) (any, error) {
	s, ok := AsStr(v)
	if !ok {
		return nil, &TypeMismatchError{Expected: "string", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &s, nil
	}
	return s, nil
}

// convertToBool converts a Value to bool.
func convertToBool(v Value, goType reflect.Type) (any, error) {
	b, ok := AsBool(v)
	if !ok {
		return nil, &TypeMismatchError{Expected: "boolean", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &b, nil
	}
	return b, nil
}

// convertToColor converts a Value to a Color.
func convertToColor(v Value, goType reflect.Type) (any, error) {
	c, ok := v.(Color)
	if !ok {
		return nil, &TypeMismatchError{Expected: "color", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &c, nil
	}
	return c, nil
}

// convertToRatio converts a Value to a Ratio.
func convertToRatio(v Value, goType reflect.Type) (any, error) {
	rv, ok := v.(RatioValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "ratio", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &rv.Ratio, nil
	}
	return rv.Ratio, nil
}

// convertToRelative converts a Value to a Relative.
func convertToRelative(v Value, goType reflect.Type) (any, error) {
	switch val := v.(type) {
	case RelativeValue:
		if goType.Kind() == reflect.Ptr {
			return &val.Relative, nil
		}
		return val.Relative, nil
	case LengthValue:
		// Length is also valid for relative - it's just abs with zero rel
		rel := Relative{Abs: val.Length, Rel: Ratio{}}
		if goType.Kind() == reflect.Ptr {
			return &rel, nil
		}
		return rel, nil
	case RatioValue:
		// Ratio is also valid for relative - it's just rel with zero abs
		rel := Relative{Abs: Length{}, Rel: val.Ratio}
		if goType.Kind() == reflect.Ptr {
			return &rel, nil
		}
		return rel, nil
	default:
		return nil, &TypeMismatchError{Expected: "relative", Got: v.Type().String()}
	}
}

// convertToAngle converts a Value to an Angle.
func convertToAngle(v Value, goType reflect.Type) (any, error) {
	av, ok := v.(AngleValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "angle", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &av.Angle, nil
	}
	return av.Angle, nil
}

// convertToFraction converts a Value to a Fraction.
func convertToFraction(v Value, goType reflect.Type) (any, error) {
	fv, ok := v.(FractionValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "fraction", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return &fv.Fraction, nil
	}
	return fv.Fraction, nil
}

// convertToArray converts a Value to an Array.
func convertToArray(v Value, goType reflect.Type) (any, error) {
	arr, ok := AsArray(v)
	if !ok {
		return nil, &TypeMismatchError{Expected: "array", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return arr, nil
	}
	return *arr, nil
}

// convertToDict converts a Value to a Dict.
func convertToDict(v Value, goType reflect.Type) (any, error) {
	dict, ok := AsDict(v)
	if !ok {
		return nil, &TypeMismatchError{Expected: "dictionary", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return dict, nil
	}
	return *dict, nil
}

// convertToFunc converts a Value to a Func.
func convertToFunc(v Value, goType reflect.Type) (any, error) {
	fv, ok := v.(FuncValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "function", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		return fv.Func, nil
	}
	return *fv.Func, nil
}

// convertToLabel converts a Value to a LabelValue.
func convertToLabel(v Value, goType reflect.Type) (any, error) {
	lv, ok := v.(LabelValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "label", Got: v.Type().String()}
	}

	// Handle pointer type
	if goType.Kind() == reflect.Ptr {
		s := string(lv)
		return &s, nil
	}
	return string(lv), nil
}

// ParseTypstType converts a string type name to a Type constant.
func ParseTypstType(name string) Type {
	switch name {
	case "none":
		return TypeNone
	case "auto":
		return TypeAuto
	case "bool", "boolean":
		return TypeBool
	case "int", "integer":
		return TypeInt
	case "float":
		return TypeFloat
	case "length":
		return TypeLength
	case "angle":
		return TypeAngle
	case "ratio":
		return TypeRatio
	case "relative":
		return TypeRelative
	case "fraction", "fr":
		return TypeFraction
	case "str", "string":
		return TypeStr
	case "bytes":
		return TypeBytes
	case "label":
		return TypeLabel
	case "datetime":
		return TypeDatetime
	case "duration":
		return TypeDuration
	case "decimal":
		return TypeDecimal
	case "color":
		return TypeColor
	case "gradient":
		return TypeGradient
	case "tiling":
		return TypeTiling
	case "symbol":
		return TypeSymbol
	case "content":
		return TypeContent
	case "array":
		return TypeArray
	case "dict", "dictionary":
		return TypeDict
	case "func", "function":
		return TypeFunc
	case "args", "arguments":
		return TypeArgs
	case "type":
		return TypeType
	case "module":
		return TypeModule
	case "styles":
		return TypeStyles
	case "version":
		return TypeVersion
	default:
		return TypeDyn // Unknown type - treat as dynamic
	}
}
