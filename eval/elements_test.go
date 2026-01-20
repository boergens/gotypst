package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

func TestRawFunc(t *testing.T) {
	// Get the raw function
	rawFunc := RawFunc()

	if rawFunc == nil {
		t.Fatal("RawFunc() returned nil")
	}

	if rawFunc.Name == nil || *rawFunc.Name != "raw" {
		t.Errorf("expected function name 'raw', got %v", rawFunc.Name)
	}

	// Verify it's a native function
	_, ok := rawFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestRawNativeBasic(t *testing.T) {
	// Create a VM with minimal setup
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with just the text
	args := NewArgs(syntax.Detached())
	args.Push(Str("print('hello')"), syntax.Detached())

	// Call the raw function
	result, err := rawNative(vm, args)
	if err != nil {
		t.Fatalf("rawNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one RawElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	raw, ok := content.Content.Elements[0].(*RawElement)
	if !ok {
		t.Fatalf("expected *RawElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties
	if raw.Text != "print('hello')" {
		t.Errorf("Text = %q, want %q", raw.Text, "print('hello')")
	}
	if raw.Block != false {
		t.Errorf("Block = %v, want false", raw.Block)
	}
	if raw.Lang != "" {
		t.Errorf("Lang = %q, want empty string", raw.Lang)
	}
}

func TestRawNativeWithLang(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with text and lang
	args := NewArgs(syntax.Detached())
	args.Push(Str("def foo(): pass"), syntax.Detached())
	args.PushNamed("lang", Str("python"), syntax.Detached())

	result, err := rawNative(vm, args)
	if err != nil {
		t.Fatalf("rawNative() error: %v", err)
	}

	content := result.(ContentValue)
	raw := content.Content.Elements[0].(*RawElement)

	if raw.Lang != "python" {
		t.Errorf("Lang = %q, want %q", raw.Lang, "python")
	}
}

func TestRawNativeWithBlock(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with text and block=true
	args := NewArgs(syntax.Detached())
	args.Push(Str("multi\nline\ncode"), syntax.Detached())
	args.PushNamed("block", True, syntax.Detached())

	result, err := rawNative(vm, args)
	if err != nil {
		t.Fatalf("rawNative() error: %v", err)
	}

	content := result.(ContentValue)
	raw := content.Content.Elements[0].(*RawElement)

	if raw.Block != true {
		t.Errorf("Block = %v, want true", raw.Block)
	}
}

func TestRawNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with all parameters
	args := NewArgs(syntax.Detached())
	args.Push(Str("fn main() {}"), syntax.Detached())
	args.PushNamed("lang", Str("rust"), syntax.Detached())
	args.PushNamed("block", True, syntax.Detached())

	result, err := rawNative(vm, args)
	if err != nil {
		t.Fatalf("rawNative() error: %v", err)
	}

	content := result.(ContentValue)
	raw := content.Content.Elements[0].(*RawElement)

	if raw.Text != "fn main() {}" {
		t.Errorf("Text = %q, want %q", raw.Text, "fn main() {}")
	}
	if raw.Lang != "rust" {
		t.Errorf("Lang = %q, want %q", raw.Lang, "rust")
	}
	if raw.Block != true {
		t.Errorf("Block = %v, want true", raw.Block)
	}
}

func TestRawNativeMissingText(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args without text
	args := NewArgs(syntax.Detached())
	args.PushNamed("lang", Str("python"), syntax.Detached())

	_, err := rawNative(vm, args)
	if err == nil {
		t.Error("expected error for missing text argument")
	}
}

func TestRawNativeWrongTextType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with wrong type for text
	args := NewArgs(syntax.Detached())
	args.Push(Int(42), syntax.Detached())

	_, err := rawNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong text type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestRawNativeWrongBlockType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with wrong type for block
	args := NewArgs(syntax.Detached())
	args.Push(Str("code"), syntax.Detached())
	args.PushNamed("block", Str("not a bool"), syntax.Detached())

	_, err := rawNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong block type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestRawNativeWrongLangType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with wrong type for lang
	args := NewArgs(syntax.Detached())
	args.Push(Str("code"), syntax.Detached())
	args.PushNamed("lang", Int(123), syntax.Detached())

	_, err := rawNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong lang type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestRawNativeLangNone(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with lang=none (explicit)
	args := NewArgs(syntax.Detached())
	args.Push(Str("code"), syntax.Detached())
	args.PushNamed("lang", None, syntax.Detached())

	result, err := rawNative(vm, args)
	if err != nil {
		t.Fatalf("rawNative() error: %v", err)
	}

	content := result.(ContentValue)
	raw := content.Content.Elements[0].(*RawElement)

	// lang=none should result in empty string
	if raw.Lang != "" {
		t.Errorf("Lang = %q, want empty string", raw.Lang)
	}
}

func TestRawNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with unexpected named argument
	args := NewArgs(syntax.Detached())
	args.Push(Str("code"), syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := rawNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestRegisterElementFunctions(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify raw function is registered
	binding := scope.Get("raw")
	if binding == nil {
		t.Fatal("expected 'raw' to be registered")
	}

	funcVal, ok := binding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue, got %T", binding.Value)
	}

	if funcVal.Func.Name == nil || *funcVal.Func.Name != "raw" {
		t.Errorf("expected function name 'raw', got %v", funcVal.Func.Name)
	}
}

func TestElementFunctions(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["raw"]; !ok {
		t.Error("expected 'raw' in ElementFunctions()")
	}
}

func TestRawElement(t *testing.T) {
	// Test RawElement struct and ContentElement interface
	elem := &RawElement{
		Text:  "print('hello')",
		Lang:  "python",
		Block: true,
	}

	if elem.Text != "print('hello')" {
		t.Errorf("Text = %q, want %q", elem.Text, "print('hello')")
	}
	if elem.Lang != "python" {
		t.Errorf("Lang = %q, want %q", elem.Lang, "python")
	}
	if elem.Block != true {
		t.Errorf("Block = %v, want true", elem.Block)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

func TestRawNativeWithNamedText(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with text as named argument
	args := NewArgs(syntax.Detached())
	args.PushNamed("text", Str("named text"), syntax.Detached())

	result, err := rawNative(vm, args)
	if err != nil {
		t.Fatalf("rawNative() error: %v", err)
	}

	content := result.(ContentValue)
	raw := content.Content.Elements[0].(*RawElement)

	if raw.Text != "named text" {
		t.Errorf("Text = %q, want %q", raw.Text, "named text")
	}
}

// ----------------------------------------------------------------------------
// Heading Element Tests
// ----------------------------------------------------------------------------

func TestHeadingFunc(t *testing.T) {
	headingFunc := HeadingFunc()

	if headingFunc == nil {
		t.Fatal("HeadingFunc() returned nil")
	}

	if headingFunc.Name == nil || *headingFunc.Name != "heading" {
		t.Errorf("expected function name 'heading', got %v", headingFunc.Name)
	}

	_, ok := headingFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestHeadingNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with just the body content
	bodyContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Introduction"}},
	}}
	args := NewArgs(syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := headingNative(vm, args)
	if err != nil {
		t.Fatalf("headingNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	heading, ok := content.Content.Elements[0].(*HeadingElement)
	if !ok {
		t.Fatalf("expected *HeadingElement, got %T", content.Content.Elements[0])
	}

	// Verify default values
	if heading.Level != 1 {
		t.Errorf("Level = %d, want 1", heading.Level)
	}
	if heading.Outlined != true {
		t.Errorf("Outlined = %v, want true", heading.Outlined)
	}
	if heading.Bookmarked != nil {
		t.Errorf("Bookmarked = %v, want nil (auto)", heading.Bookmarked)
	}
	if heading.Numbering != "" {
		t.Errorf("Numbering = %q, want empty string", heading.Numbering)
	}
}

func TestHeadingNativeWithLevel(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Subsection"}},
	}}
	args := NewArgs(syntax.Detached())
	args.Push(bodyContent, syntax.Detached())
	args.PushNamed("level", Int(2), syntax.Detached())

	result, err := headingNative(vm, args)
	if err != nil {
		t.Fatalf("headingNative() error: %v", err)
	}

	content := result.(ContentValue)
	heading := content.Content.Elements[0].(*HeadingElement)

	if heading.Level != 2 {
		t.Errorf("Level = %d, want 2", heading.Level)
	}
}

func TestHeadingNativeWithNumbering(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Chapter"}},
	}}
	args := NewArgs(syntax.Detached())
	args.Push(bodyContent, syntax.Detached())
	args.PushNamed("numbering", Str("1."), syntax.Detached())

	result, err := headingNative(vm, args)
	if err != nil {
		t.Fatalf("headingNative() error: %v", err)
	}

	content := result.(ContentValue)
	heading := content.Content.Elements[0].(*HeadingElement)

	if heading.Numbering != "1." {
		t.Errorf("Numbering = %q, want %q", heading.Numbering, "1.")
	}
}

func TestHeadingNativeWithOutlined(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Hidden"}},
	}}
	args := NewArgs(syntax.Detached())
	args.Push(bodyContent, syntax.Detached())
	args.PushNamed("outlined", False, syntax.Detached())

	result, err := headingNative(vm, args)
	if err != nil {
		t.Fatalf("headingNative() error: %v", err)
	}

	content := result.(ContentValue)
	heading := content.Content.Elements[0].(*HeadingElement)

	if heading.Outlined != false {
		t.Errorf("Outlined = %v, want false", heading.Outlined)
	}
}

func TestHeadingNativeWithBookmarked(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Bookmarked"}},
	}}
	args := NewArgs(syntax.Detached())
	args.Push(bodyContent, syntax.Detached())
	args.PushNamed("bookmarked", True, syntax.Detached())

	result, err := headingNative(vm, args)
	if err != nil {
		t.Fatalf("headingNative() error: %v", err)
	}

	content := result.(ContentValue)
	heading := content.Content.Elements[0].(*HeadingElement)

	if heading.Bookmarked == nil || *heading.Bookmarked != true {
		t.Errorf("Bookmarked = %v, want true", heading.Bookmarked)
	}
}

func TestHeadingNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Full Heading"}},
	}}
	supplementContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Supplement"}},
	}}
	args := NewArgs(syntax.Detached())
	args.Push(bodyContent, syntax.Detached())
	args.PushNamed("level", Int(3), syntax.Detached())
	args.PushNamed("numbering", Str("1.1.1"), syntax.Detached())
	args.PushNamed("outlined", False, syntax.Detached())
	args.PushNamed("bookmarked", False, syntax.Detached())
	args.PushNamed("supplement", supplementContent, syntax.Detached())

	result, err := headingNative(vm, args)
	if err != nil {
		t.Fatalf("headingNative() error: %v", err)
	}

	content := result.(ContentValue)
	heading := content.Content.Elements[0].(*HeadingElement)

	if heading.Level != 3 {
		t.Errorf("Level = %d, want 3", heading.Level)
	}
	if heading.Numbering != "1.1.1" {
		t.Errorf("Numbering = %q, want %q", heading.Numbering, "1.1.1")
	}
	if heading.Outlined != false {
		t.Errorf("Outlined = %v, want false", heading.Outlined)
	}
	if heading.Bookmarked == nil || *heading.Bookmarked != false {
		t.Errorf("Bookmarked = %v, want false", heading.Bookmarked)
	}
	if heading.Supplement == nil {
		t.Error("Supplement should not be nil")
	}
}

func TestHeadingNativeStringBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Test with string as body (should be converted to content)
	args := NewArgs(syntax.Detached())
	args.Push(Str("String Heading"), syntax.Detached())

	result, err := headingNative(vm, args)
	if err != nil {
		t.Fatalf("headingNative() error: %v", err)
	}

	content := result.(ContentValue)
	heading := content.Content.Elements[0].(*HeadingElement)

	if len(heading.Content.Elements) != 1 {
		t.Fatalf("expected 1 content element, got %d", len(heading.Content.Elements))
	}
	textElem, ok := heading.Content.Elements[0].(*TextElement)
	if !ok {
		t.Fatalf("expected *TextElement, got %T", heading.Content.Elements[0])
	}
	if textElem.Text != "String Heading" {
		t.Errorf("Text = %q, want %q", textElem.Text, "String Heading")
	}
}

func TestHeadingNativeLevelClamping(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	tests := []struct {
		name      string
		level     int64
		wantLevel int
	}{
		{"level 0 -> 1", 0, 1},
		{"level -1 -> 1", -1, 1},
		{"level 7 -> 6", 7, 6},
		{"level 100 -> 6", 100, 6},
		{"level 3 -> 3", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyContent := ContentValue{Content: Content{
				Elements: []ContentElement{&TextElement{Text: "Test"}},
			}}
			args := NewArgs(syntax.Detached())
			args.Push(bodyContent, syntax.Detached())
			args.PushNamed("level", Int(tt.level), syntax.Detached())

			result, err := headingNative(vm, args)
			if err != nil {
				t.Fatalf("headingNative() error: %v", err)
			}

			content := result.(ContentValue)
			heading := content.Content.Elements[0].(*HeadingElement)

			if heading.Level != tt.wantLevel {
				t.Errorf("Level = %d, want %d", heading.Level, tt.wantLevel)
			}
		})
	}
}

func TestHeadingNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("level", Int(2), syntax.Detached())

	_, err := headingNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
	}
}

func TestHeadingNativeWrongBodyType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Int(42), syntax.Detached())

	_, err := headingNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong body type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestHeadingElement(t *testing.T) {
	// Test HeadingElement struct and ContentElement interface
	bookmarked := true
	elem := &HeadingElement{
		Level:      2,
		Content:    Content{Elements: []ContentElement{&TextElement{Text: "Test"}}},
		Numbering:  "1.1",
		Outlined:   true,
		Bookmarked: &bookmarked,
	}

	if elem.Level != 2 {
		t.Errorf("Level = %d, want 2", elem.Level)
	}
	if elem.Numbering != "1.1" {
		t.Errorf("Numbering = %q, want %q", elem.Numbering, "1.1")
	}
	if elem.Outlined != true {
		t.Errorf("Outlined = %v, want true", elem.Outlined)
	}
	if !elem.IsBookmarked() {
		t.Error("IsBookmarked() = false, want true")
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

func TestHeadingElementIsBookmarked(t *testing.T) {
	tests := []struct {
		name       string
		bookmarked *bool
		outlined   bool
		want       bool
	}{
		{"bookmarked true", boolPtr(true), false, true},
		{"bookmarked false", boolPtr(false), true, false},
		{"bookmarked auto, outlined true", nil, true, true},
		{"bookmarked auto, outlined false", nil, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := &HeadingElement{
				Level:      1,
				Outlined:   tt.outlined,
				Bookmarked: tt.bookmarked,
			}
			if got := elem.IsBookmarked(); got != tt.want {
				t.Errorf("IsBookmarked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestRegisterElementFunctionsWithHeading(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify heading function is registered
	binding := scope.Get("heading")
	if binding == nil {
		t.Fatal("expected 'heading' to be registered")
	}

	funcVal, ok := binding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue, got %T", binding.Value)
	}

	if funcVal.Func.Name == nil || *funcVal.Func.Name != "heading" {
		t.Errorf("expected function name 'heading', got %v", funcVal.Func.Name)
	}
}

func TestElementFunctionsWithHeading(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["heading"]; !ok {
		t.Error("expected 'heading' in ElementFunctions()")
	}
}
