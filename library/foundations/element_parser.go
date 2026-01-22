// Generic element parser for declaratively-defined elements.
// This file provides the parsing logic that converts Args into element structs.

package foundations

import (
	"reflect"

	"github.com/boergens/gotypst/syntax"
)

// ParseElement parses arguments into an element struct using the element definition.
// This is the generic parser that handles construct (element creation).
//
// The function:
// 1. Processes shorthands in order (they expand to multiple fields)
// 2. Processes named arguments (override shorthand values)
// 3. Processes positional arguments in field order
// 4. Validates required fields are present
// 5. Checks for unexpected arguments
func ParseElement[T any](def *ElementDef, args *Args) (*T, error) {
	elem := new(T)
	v := reflect.ValueOf(elem).Elem()

	// Process shorthands first (in specified order if available)
	if err := processShorthands(def, args, v); err != nil {
		return nil, err
	}

	// Process named arguments (these override shorthand values)
	if err := processNamedArgs(def, args, v); err != nil {
		return nil, err
	}

	// Process positional arguments
	if err := processPositionalArgs(def, args, v); err != nil {
		return nil, err
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	return elem, nil
}

// ParseSetRule parses arguments for a set rule, returning styles.
// Set rules only affect optional/settable fields - required fields cannot be set.
func ParseSetRule[T any](def *ElementDef, args *Args) (*Styles, error) {
	styles := NewStyles()

	// Process shorthands for set rules
	if err := processShorthandsForSet(def, args, styles); err != nil {
		return nil, err
	}

	// Process named arguments
	for _, field := range def.Fields {
		if !field.Settable || field.Positional {
			continue
		}

		if arg := args.Find(field.Name); arg != nil {
			converted, err := ConvertValue(arg.V, field.Type, field.GoType)
			if err != nil {
				if tme, ok := err.(*TypeMismatchError); ok {
					tme.Field = field.Name
				}
				return nil, err
			}

			// Store in styles
			if v, ok := converted.(Value); ok {
				styles.SetProperty(StyleProperty{Element: def.Name, Field: field.Name}, v)
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	return styles, nil
}

// processShorthands handles shorthand arguments like "rest", "x", "y".
func processShorthands(def *ElementDef, args *Args, v reflect.Value) error {
	// Determine order of shorthand processing
	order := def.ShorthandOrder
	if len(order) == 0 {
		// Use map keys (undefined order, but consistent within single run)
		order = make([]string, 0, len(def.Shorthands))
		for name := range def.Shorthands {
			order = append(order, name)
		}
	}

	for _, shorthand := range order {
		targets, ok := def.Shorthands[shorthand]
		if !ok {
			continue
		}

		arg := args.Find(shorthand)
		if arg == nil {
			continue
		}

		// Skip none and auto for shorthands
		if IsNone(arg.V) || IsAuto(arg.V) {
			continue
		}

		// Apply to all target fields
		for _, targetName := range targets {
			field := def.FieldByName(targetName)
			if field == nil {
				continue
			}

			if err := setFieldValue(v, field, arg.V); err != nil {
				return err
			}
		}
	}

	return nil
}

// processShorthandsForSet handles shorthands for set rules.
func processShorthandsForSet(def *ElementDef, args *Args, styles *Styles) error {
	order := def.ShorthandOrder
	if len(order) == 0 {
		order = make([]string, 0, len(def.Shorthands))
		for name := range def.Shorthands {
			order = append(order, name)
		}
	}

	for _, shorthand := range order {
		targets, ok := def.Shorthands[shorthand]
		if !ok {
			continue
		}

		arg := args.Find(shorthand)
		if arg == nil {
			continue
		}

		// Skip none and auto for shorthands
		if IsNone(arg.V) || IsAuto(arg.V) {
			continue
		}

		// Apply to all target fields
		for _, targetName := range targets {
			field := def.FieldByName(targetName)
			if field == nil || !field.Settable {
				continue
			}

			converted, err := ConvertValue(arg.V, field.Type, field.GoType)
			if err != nil {
				return err
			}

			if v, ok := converted.(Value); ok {
				styles.SetProperty(StyleProperty{Element: def.Name, Field: targetName}, v)
			}
		}
	}

	return nil
}

// processNamedArgs handles named (non-positional) arguments.
func processNamedArgs(def *ElementDef, args *Args, v reflect.Value) error {
	for _, field := range def.Fields {
		if field.Positional {
			continue
		}

		// Try to find the named argument
		arg := args.Find(field.Name)
		if arg == nil {
			// Field not provided - check if required
			if field.Required {
				return &MissingArgumentError{Name: field.Name, Span: args.Span}
			}
			// Apply default if available
			if field.Default != nil && !IsNone(field.Default) {
				if err := setFieldValue(v, &field, field.Default); err != nil {
					return err
				}
			}
			continue
		}

		if err := setFieldValue(v, &field, arg.V); err != nil {
			return err
		}
	}

	return nil
}

// processPositionalArgs handles positional arguments.
func processPositionalArgs(def *ElementDef, args *Args, v reflect.Value) error {
	for _, field := range def.Fields {
		if !field.Positional {
			continue
		}

		if field.Variadic {
			// Collect all remaining positional arguments
			remaining := args.All()
			if err := setVariadicField(v, &field, remaining); err != nil {
				return err
			}
		} else {
			// Single positional argument
			arg := args.Eat()
			if arg == nil {
				// Field not provided - check if required
				if field.Required {
					return &MissingArgumentError{Name: field.Name, Span: args.Span}
				}
				// Apply default if available
				if field.Default != nil && !IsNone(field.Default) {
					if err := setFieldValue(v, &field, field.Default); err != nil {
						return err
					}
				}
				continue
			}

			if err := setFieldValue(v, &field, arg.V); err != nil {
				return err
			}
		}
	}

	return nil
}

// setFieldValue sets a struct field from a Typst value.
func setFieldValue(v reflect.Value, field *FieldDef, val Value) error {
	converted, err := ConvertValue(val, field.Type, field.GoType)
	if err != nil {
		if tme, ok := err.(*TypeMismatchError); ok {
			tme.Field = field.Name
		}
		return err
	}

	fieldValue := v.Field(field.GoFieldIndex)
	return setReflectValue(fieldValue, converted, field)
}

// setVariadicField sets a variadic field from multiple values.
func setVariadicField(v reflect.Value, field *FieldDef, values []syntax.Spanned[Value]) error {
	fieldValue := v.Field(field.GoFieldIndex)

	// Create a slice of the appropriate type
	elemType := field.GoType.Elem()
	slice := reflect.MakeSlice(field.GoType, 0, len(values))

	for _, spanned := range values {
		converted, err := ConvertValue(spanned.V, field.Type, elemType)
		if err != nil {
			if tme, ok := err.(*TypeMismatchError); ok {
				tme.Field = field.Name
			}
			return err
		}

		elemVal := reflect.ValueOf(converted)
		if elemVal.Type().ConvertibleTo(elemType) {
			slice = reflect.Append(slice, elemVal.Convert(elemType))
		} else {
			slice = reflect.Append(slice, elemVal)
		}
	}

	fieldValue.Set(slice)
	return nil
}

// setReflectValue sets a reflect.Value from a Go interface{} value.
func setReflectValue(fieldValue reflect.Value, converted any, field *FieldDef) error {
	if converted == nil {
		// nil means the value was none - leave the field as zero value
		return nil
	}

	convertedVal := reflect.ValueOf(converted)

	// Handle pointer types
	if fieldValue.Kind() == reflect.Ptr {
		if convertedVal.Kind() == reflect.Ptr {
			fieldValue.Set(convertedVal)
		} else {
			// Create a pointer to the value
			ptr := reflect.New(fieldValue.Type().Elem())
			ptr.Elem().Set(convertedVal)
			fieldValue.Set(ptr)
		}
		return nil
	}

	// Handle non-pointer types
	if convertedVal.Kind() == reflect.Ptr {
		if !convertedVal.IsNil() {
			fieldValue.Set(convertedVal.Elem())
		}
	} else {
		if convertedVal.Type().ConvertibleTo(fieldValue.Type()) {
			fieldValue.Set(convertedVal.Convert(fieldValue.Type()))
		} else {
			fieldValue.Set(convertedVal)
		}
	}

	return nil
}

// ElementFunc creates a NativeFunc wrapper for an element constructor.
// This generates the function that creates element instances from arguments.
func ElementFunc[T ContentElement](name string) *Func {
	def := GetElement(name)
	if def == nil {
		panic("element not registered: " + name)
	}

	funcName := name
	return &Func{
		Name: &funcName,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(engine Engine, context Context, args *Args) (Value, error) {
				elem, err := ParseElement[T](def, args)
				if err != nil {
					return nil, err
				}
				return ContentValue{Content: Content{Elements: []ContentElement{*elem}}}, nil
			},
			Info: def.ToFuncInfo(),
		},
	}
}

// ElementSetFunc creates a set rule function for an element.
// This generates the function that creates style rules from arguments.
func ElementSetFunc[T any](name string) func(args *Args) (*Styles, error) {
	def := GetElement(name)
	if def == nil {
		panic("element not registered: " + name)
	}

	return func(args *Args) (*Styles, error) {
		return ParseSetRule[T](def, args)
	}
}

// MakeElementFunc creates both constructor and set functions for an element.
// Returns (constructFunc, setFunc).
func MakeElementFunc[T ContentElement](name string) (*Func, func(args *Args) (*Styles, error)) {
	return ElementFunc[T](name), ElementSetFunc[T](name)
}
