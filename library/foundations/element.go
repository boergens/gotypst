// Declarative element definition system.
// This file provides the infrastructure for defining Typst elements using
// struct tags, matching Rust's #[elem] macro pattern.
//
// Example usage:
//
//	type PadElement struct {
//	    Left   *Length `typst:"left,type=length"`
//	    Top    *Length `typst:"top,type=length"`
//	    Body   Content `typst:"body,positional,required"`
//	}
//
//	var PadShorthands = map[string][]string{
//	    "rest": {"left", "top", "right", "bottom"},
//	}
//
//	func init() {
//	    RegisterElement[PadElement]("pad", PadShorthands)
//	}

package foundations

import (
	"reflect"
	"strings"
	"sync"

	"github.com/boergens/gotypst/syntax"
)

// ElementDef describes an element's structure for parsing.
type ElementDef struct {
	// Name is the element function name (e.g., "pad", "text", "grid").
	Name string

	// Type is the Go struct type for this element.
	Type reflect.Type

	// Fields describes each field's parsing behavior.
	Fields []FieldDef

	// Shorthands maps shorthand argument names to the fields they expand to.
	// For example, "rest" -> {"left", "top", "right", "bottom"} for pad.
	// Shorthands are processed before individual fields, so individual fields
	// can override shorthand values.
	Shorthands map[string][]string

	// ShorthandOrder defines the order in which shorthands should be processed.
	// Shorthands processed earlier can be overridden by those processed later.
	// If empty, shorthands are processed in map iteration order (undefined).
	ShorthandOrder []string
}

// FieldDef describes a single field's parsing behavior.
type FieldDef struct {
	// Name is the Typst argument name (from the tag).
	Name string

	// GoField is the Go struct field name.
	GoField string

	// GoFieldIndex is the struct field index for reflection.
	GoFieldIndex int

	// Type is the expected Typst type.
	Type Type

	// GoType is the Go type for the field.
	GoType reflect.Type

	// Default is the default value (nil means no default).
	Default Value

	// Positional indicates this field is parsed from positional args.
	Positional bool

	// Required indicates this field must be provided.
	Required bool

	// Variadic indicates this field collects remaining positional args.
	// The field must be a slice type.
	Variadic bool

	// Settable indicates this field can be set via set rules.
	// Required fields are typically not settable.
	Settable bool
}

// FieldByName returns the field definition with the given Typst name.
func (d *ElementDef) FieldByName(name string) *FieldDef {
	for i := range d.Fields {
		if d.Fields[i].Name == name {
			return &d.Fields[i]
		}
	}
	return nil
}

// ToFuncInfo converts the element definition to FuncInfo for function metadata.
func (d *ElementDef) ToFuncInfo() *FuncInfo {
	params := make([]ParamInfo, 0, len(d.Fields)+len(d.Shorthands))

	// Add shorthand parameters first (they appear as named params)
	for shorthand := range d.Shorthands {
		// Determine the type from the first target field
		targets := d.Shorthands[shorthand]
		var typ Type
		if len(targets) > 0 {
			if f := d.FieldByName(targets[0]); f != nil {
				typ = f.Type
			}
		}
		params = append(params, ParamInfo{
			Name:    shorthand,
			Type:    typ,
			Default: None,
			Named:   true,
		})
	}

	// Add field parameters
	for _, field := range d.Fields {
		params = append(params, ParamInfo{
			Name:     field.Name,
			Type:     field.Type,
			Default:  field.Default,
			Variadic: field.Variadic,
			Named:    !field.Positional,
		})
	}

	return &FuncInfo{
		Name:   d.Name,
		Params: params,
	}
}

// elements is the global registry of element definitions.
var (
	elements   = make(map[string]*ElementDef)
	elementsMu sync.RWMutex
)

// RegisterElement registers an element type with the given name and shorthands.
// The type T must be a struct with typst tags on its fields.
//
// Example:
//
//	RegisterElement[PadElement]("pad", map[string][]string{
//	    "rest": {"left", "top", "right", "bottom"},
//	    "x":    {"left", "right"},
//	    "y":    {"top", "bottom"},
//	})
func RegisterElement[T any](name string, shorthands map[string][]string) *ElementDef {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	def := parseElementDef(typ, name, shorthands)

	elementsMu.Lock()
	elements[name] = def
	elementsMu.Unlock()

	return def
}

// RegisterElementWithOrder registers an element with explicit shorthand processing order.
func RegisterElementWithOrder[T any](name string, shorthands map[string][]string, shorthandOrder []string) *ElementDef {
	def := RegisterElement[T](name, shorthands)
	def.ShorthandOrder = shorthandOrder
	return def
}

// GetElement retrieves an element definition by name.
func GetElement(name string) *ElementDef {
	elementsMu.RLock()
	defer elementsMu.RUnlock()
	return elements[name]
}

// parseElementDef parses struct tags to create an ElementDef.
func parseElementDef(typ reflect.Type, name string, shorthands map[string][]string) *ElementDef {
	if typ.Kind() != reflect.Struct {
		panic("RegisterElement requires a struct type")
	}

	def := &ElementDef{
		Name:       name,
		Type:       typ,
		Fields:     make([]FieldDef, 0, typ.NumField()),
		Shorthands: shorthands,
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("typst")
		if tag == "" {
			continue // Skip fields without typst tag
		}

		fieldDef := parseFieldTag(field, i, tag)
		def.Fields = append(def.Fields, fieldDef)
	}

	return def
}

// parseFieldTag parses a typst struct tag into a FieldDef.
//
// Tag format: `typst:"name[,option...]"`
//
// Options:
//   - positional: Field is parsed from positional args (default: named)
//   - required: Field must be provided (error if missing)
//   - type=TYPE: Expected Typst type (length, content, int, etc.)
//   - default=VALUE: Default value (none, auto, 0, true, "text")
//   - variadic: Collects remaining positional args (field must be slice)
//   - settable: Field can be set via set rules (default: true for non-required)
func parseFieldTag(field reflect.StructField, index int, tag string) FieldDef {
	parts := strings.Split(tag, ",")
	name := strings.TrimSpace(parts[0])

	def := FieldDef{
		Name:         name,
		GoField:      field.Name,
		GoFieldIndex: index,
		GoType:       field.Type,
		Settable:     true, // Default: settable
	}

	// Parse options
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, "type=") {
			typeName := strings.TrimPrefix(part, "type=")
			def.Type = ParseTypstType(typeName)
		} else if strings.HasPrefix(part, "default=") {
			defaultStr := strings.TrimPrefix(part, "default=")
			def.Default = parseDefaultValue(defaultStr)
		} else {
			switch part {
			case "positional":
				def.Positional = true
			case "required":
				def.Required = true
				def.Settable = false // Required fields typically aren't settable via set rules
			case "variadic":
				def.Variadic = true
				def.Positional = true // Variadic implies positional
			case "settable":
				def.Settable = true
			case "nosettable":
				def.Settable = false
			}
		}
	}

	// Infer type from Go type if not specified
	if def.Type == 0 {
		def.Type = inferTypstType(field.Type)
	}

	return def
}

// parseDefaultValue parses a default value string into a Value.
func parseDefaultValue(s string) Value {
	switch s {
	case "none":
		return None
	case "auto":
		return Auto
	case "true":
		return True
	case "false":
		return False
	case "0":
		return Int(0)
	case "0.0":
		return Float(0)
	case "":
		return None
	default:
		// String default (remove quotes if present)
		s = strings.Trim(s, "\"'")
		return Str(s)
	}
}

// inferTypstType infers the Typst type from a Go type.
func inferTypstType(t reflect.Type) Type {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check for known types
	switch t {
	case reflect.TypeOf(Length{}):
		return TypeLength
	case reflect.TypeOf(Content{}):
		return TypeContent
	case reflect.TypeOf(Angle{}):
		return TypeAngle
	case reflect.TypeOf(Ratio{}):
		return TypeRatio
	case reflect.TypeOf(Relative{}):
		return TypeRelative
	case reflect.TypeOf(Fraction{}):
		return TypeFraction
	}

	// Check by kind
	switch t.Kind() {
	case reflect.Bool:
		return TypeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return TypeInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return TypeInt
	case reflect.Float32, reflect.Float64:
		return TypeFloat
	case reflect.String:
		return TypeStr
	case reflect.Slice:
		// Check if it's a slice of Values (for variadic)
		if t.Elem().Implements(reflect.TypeOf((*Value)(nil)).Elem()) {
			return TypeArray
		}
	}

	return TypeDyn
}

// StyleProperty identifies a property that can be styled.
type StyleProperty struct {
	Element string
	Field   string
}

// SetProperty sets a style property value.
func (s *Styles) SetProperty(prop StyleProperty, value any) {
	// For now, we store as style rules with a synthetic function
	// This will be refined when we implement full style system integration
	name := prop.Element
	f := &Func{
		Name: &name,
		Span: syntax.Detached(),
	}

	args := NewArgs(syntax.Detached())
	if v, ok := value.(Value); ok {
		args.PushNamed(prop.Field, v, syntax.Detached())
	}

	s.Rules = append(s.Rules, StyleRule{
		Func: f,
		Args: args,
	})
}
