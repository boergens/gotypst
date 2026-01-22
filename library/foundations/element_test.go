package foundations

import (
	"reflect"
	"testing"

	"github.com/boergens/gotypst/syntax"
)

// TestElement is a test element with various field types.
type TestElement struct {
	Name     string   `typst:"name,positional,required"`
	Count    *int64   `typst:"count,type=int"`
	Width    *Length  `typst:"width,type=length"`
	Content  Content  `typst:"body,positional"`
}

func (TestElement) IsContentElement() {}

// TestElement2 has shorthands.
type TestElement2 struct {
	Left   *Length `typst:"left,type=length"`
	Right  *Length `typst:"right,type=length"`
	Top    *Length `typst:"top,type=length"`
	Bottom *Length `typst:"bottom,type=length"`
}

func (TestElement2) IsContentElement() {}

func TestRegisterElement(t *testing.T) {
	def := RegisterElement[TestElement]("test", nil)

	if def.Name != "test" {
		t.Errorf("expected name 'test', got %q", def.Name)
	}

	if len(def.Fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(def.Fields))
	}

	// Check name field
	nameField := def.FieldByName("name")
	if nameField == nil {
		t.Fatal("expected 'name' field to exist")
	}
	if !nameField.Positional {
		t.Error("expected 'name' field to be positional")
	}
	if !nameField.Required {
		t.Error("expected 'name' field to be required")
	}
	if nameField.Type != TypeStr {
		t.Errorf("expected 'name' field type to be string, got %v", nameField.Type)
	}

	// Check count field (named, optional)
	countField := def.FieldByName("count")
	if countField == nil {
		t.Fatal("expected 'count' field to exist")
	}
	if countField.Positional {
		t.Error("expected 'count' field to be named")
	}
	if countField.Required {
		t.Error("expected 'count' field to be optional")
	}
	if countField.Type != TypeInt {
		t.Errorf("expected 'count' field type to be int, got %v", countField.Type)
	}

	// Check width field
	widthField := def.FieldByName("width")
	if widthField == nil {
		t.Fatal("expected 'width' field to exist")
	}
	if widthField.Type != TypeLength {
		t.Errorf("expected 'width' field type to be length, got %v", widthField.Type)
	}
}

func TestRegisterElementWithShorthands(t *testing.T) {
	shorthands := map[string][]string{
		"rest": {"left", "right", "top", "bottom"},
		"x":    {"left", "right"},
		"y":    {"top", "bottom"},
	}
	def := RegisterElement[TestElement2]("test2", shorthands)

	if len(def.Shorthands) != 3 {
		t.Errorf("expected 3 shorthands, got %d", len(def.Shorthands))
	}

	if targets := def.Shorthands["rest"]; len(targets) != 4 {
		t.Errorf("expected 'rest' to have 4 targets, got %d", len(targets))
	}
}

func TestParseElementRequired(t *testing.T) {
	def := GetElement("test")
	if def == nil {
		t.Fatal("test element not registered")
	}

	// Create args without required field
	args := NewArgs(syntax.Detached())

	_, err := ParseElement[TestElement](def, args)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}

	missingErr, ok := err.(*MissingArgumentError)
	if !ok {
		t.Fatalf("expected MissingArgumentError, got %T: %v", err, err)
	}
	if missingErr.Name != "name" {
		t.Errorf("expected missing field 'name', got %q", missingErr.Name)
	}
}

func TestParseElementPositional(t *testing.T) {
	def := GetElement("test")
	if def == nil {
		t.Fatal("test element not registered")
	}

	// Create args with required positional field
	args := NewArgs(syntax.Detached())
	args.Push(Str("hello"), syntax.Detached())

	elem, err := ParseElement[TestElement](def, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elem.Name != "hello" {
		t.Errorf("expected name 'hello', got %q", elem.Name)
	}
}

func TestParseElementNamed(t *testing.T) {
	def := GetElement("test")
	if def == nil {
		t.Fatal("test element not registered")
	}

	// Create args with required positional and optional named field
	args := NewArgs(syntax.Detached())
	args.Push(Str("hello"), syntax.Detached())
	args.PushNamed("count", Int(42), syntax.Detached())

	elem, err := ParseElement[TestElement](def, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elem.Name != "hello" {
		t.Errorf("expected name 'hello', got %q", elem.Name)
	}
	if elem.Count == nil {
		t.Fatal("expected count to be set")
	}
	if *elem.Count != 42 {
		t.Errorf("expected count 42, got %d", *elem.Count)
	}
}

func TestParseElementWithLength(t *testing.T) {
	def := GetElement("test")
	if def == nil {
		t.Fatal("test element not registered")
	}

	// Create args with required positional and optional named length field
	args := NewArgs(syntax.Detached())
	args.Push(Str("hello"), syntax.Detached())
	args.PushNamed("width", LengthValue{Length: Length{Points: 10.5}}, syntax.Detached())

	elem, err := ParseElement[TestElement](def, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elem.Width == nil {
		t.Fatal("expected width to be set")
	}
	if elem.Width.Points != 10.5 {
		t.Errorf("expected width 10.5pt, got %f", elem.Width.Points)
	}
}

func TestParseElementShorthands(t *testing.T) {
	def := GetElement("test2")
	if def == nil {
		t.Fatal("test2 element not registered")
	}

	// Create args with 'rest' shorthand
	args := NewArgs(syntax.Detached())
	args.PushNamed("rest", LengthValue{Length: Length{Points: 5.0}}, syntax.Detached())

	elem, err := ParseElement[TestElement2](def, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All four sides should be set
	if elem.Left == nil || elem.Left.Points != 5.0 {
		t.Errorf("expected left 5pt, got %v", elem.Left)
	}
	if elem.Right == nil || elem.Right.Points != 5.0 {
		t.Errorf("expected right 5pt, got %v", elem.Right)
	}
	if elem.Top == nil || elem.Top.Points != 5.0 {
		t.Errorf("expected top 5pt, got %v", elem.Top)
	}
	if elem.Bottom == nil || elem.Bottom.Points != 5.0 {
		t.Errorf("expected bottom 5pt, got %v", elem.Bottom)
	}
}

func TestParseElementShorthandOverride(t *testing.T) {
	def := RegisterElementWithOrder[TestElement2](
		"test3",
		map[string][]string{
			"rest": {"left", "right", "top", "bottom"},
			"x":    {"left", "right"},
		},
		[]string{"rest", "x"},
	)

	// Create args with 'rest' and individual override
	args := NewArgs(syntax.Detached())
	args.PushNamed("rest", LengthValue{Length: Length{Points: 5.0}}, syntax.Detached())
	args.PushNamed("left", LengthValue{Length: Length{Points: 10.0}}, syntax.Detached())

	elem, err := ParseElement[TestElement2](def, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Left should be overridden to 10pt
	if elem.Left == nil || elem.Left.Points != 10.0 {
		t.Errorf("expected left 10pt (override), got %v", elem.Left)
	}
	// Others should be 5pt from rest
	if elem.Right == nil || elem.Right.Points != 5.0 {
		t.Errorf("expected right 5pt, got %v", elem.Right)
	}
}

func TestParseElementTypeMismatch(t *testing.T) {
	def := GetElement("test")
	if def == nil {
		t.Fatal("test element not registered")
	}

	// Create args with wrong type for count (string instead of int)
	args := NewArgs(syntax.Detached())
	args.Push(Str("hello"), syntax.Detached())
	args.PushNamed("count", Str("not an int"), syntax.Detached())

	_, err := ParseElement[TestElement](def, args)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}

	typeErr, ok := err.(*TypeMismatchError)
	if !ok {
		t.Fatalf("expected TypeMismatchError, got %T: %v", err, err)
	}
	if typeErr.Expected != "integer" {
		t.Errorf("expected 'integer', got %q", typeErr.Expected)
	}
	if typeErr.Field != "count" {
		t.Errorf("expected field 'count', got %q", typeErr.Field)
	}
}

func TestParseElementUnexpected(t *testing.T) {
	def := GetElement("test")
	if def == nil {
		t.Fatal("test element not registered")
	}

	// Create args with an unexpected named argument
	args := NewArgs(syntax.Detached())
	args.Push(Str("hello"), syntax.Detached())
	args.PushNamed("unknown", Int(123), syntax.Detached())

	_, err := ParseElement[TestElement](def, args)
	if err == nil {
		t.Fatal("expected error for unexpected argument")
	}

	unexpectedErr, ok := err.(*UnexpectedArgumentError)
	if !ok {
		t.Fatalf("expected UnexpectedArgumentError, got %T: %v", err, err)
	}
	if unexpectedErr.Arg.Name == nil || *unexpectedErr.Arg.Name != "unknown" {
		t.Errorf("expected unexpected arg 'unknown', got %v", unexpectedErr.Arg.Name)
	}
}

func TestToFuncInfo(t *testing.T) {
	def := GetElement("test")
	if def == nil {
		t.Fatal("test element not registered")
	}

	info := def.ToFuncInfo()

	if info.Name != "test" {
		t.Errorf("expected name 'test', got %q", info.Name)
	}

	if len(info.Params) != 4 {
		t.Errorf("expected 4 params, got %d", len(info.Params))
	}

	// Find the 'name' param
	var nameParam *ParamInfo
	for i := range info.Params {
		if info.Params[i].Name == "name" {
			nameParam = &info.Params[i]
			break
		}
	}

	if nameParam == nil {
		t.Fatal("expected 'name' param")
	}
	if nameParam.Named {
		t.Error("expected 'name' param to be positional (Named=false)")
	}
}

func TestConvertValueLength(t *testing.T) {
	v := LengthValue{Length: Length{Points: 12.0}}

	// Convert to *Length
	result, err := ConvertValue(v, TypeLength, ptrType[Length]())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	length, ok := result.(*Length)
	if !ok {
		t.Fatalf("expected *Length, got %T", result)
	}
	if length.Points != 12.0 {
		t.Errorf("expected 12pt, got %f", length.Points)
	}
}

func TestConvertValueNone(t *testing.T) {
	// Converting none to a pointer type should return nil
	result, err := ConvertValue(None, TypeLength, ptrType[Length]())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != (*Length)(nil) {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestConvertValueInt(t *testing.T) {
	v := Int(42)

	result, err := ConvertValue(v, TypeInt, ptrType[int64]())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	i, ok := result.(*int64)
	if !ok {
		t.Fatalf("expected *int64, got %T", result)
	}
	if *i != 42 {
		t.Errorf("expected 42, got %d", *i)
	}
}

func TestParseTypstType(t *testing.T) {
	tests := []struct {
		input    string
		expected Type
	}{
		{"length", TypeLength},
		{"content", TypeContent},
		{"int", TypeInt},
		{"integer", TypeInt},
		{"float", TypeFloat},
		{"str", TypeStr},
		{"string", TypeStr},
		{"bool", TypeBool},
		{"boolean", TypeBool},
		{"color", TypeColor},
		{"none", TypeNone},
		{"auto", TypeAuto},
		{"array", TypeArray},
		{"dict", TypeDict},
		{"dictionary", TypeDict},
		{"function", TypeFunc},
		{"func", TypeFunc},
	}

	for _, tt := range tests {
		result := ParseTypstType(tt.input)
		if result != tt.expected {
			t.Errorf("ParseTypstType(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

// Helper to get reflect.Type for a pointer type
func ptrType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil))
}
