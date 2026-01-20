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

func TestRawElementStruct(t *testing.T) {
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
// Paragraph Tests
// ----------------------------------------------------------------------------

func TestParFunc(t *testing.T) {
	// Get the par function
	parFunc := ParFunc()

	if parFunc == nil {
		t.Fatal("ParFunc() returned nil")
	}

	if parFunc.Name == nil || *parFunc.Name != "par" {
		t.Errorf("expected function name 'par', got %v", parFunc.Name)
	}

	// Verify it's a native function
	_, ok := parFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestParNativeBasic(t *testing.T) {
	// Create a VM with minimal setup
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create body content
	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Hello, World!"}},
	}}

	// Create args with just the body
	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	// Call the par function
	result, err := parNative(vm, args)
	if err != nil {
		t.Fatalf("parNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one ParagraphElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	par, ok := content.Content.Elements[0].(*ParagraphElement)
	if !ok {
		t.Fatalf("expected *ParagraphElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties (defaults)
	if len(par.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(par.Body.Elements))
	}
	if par.Leading != nil {
		t.Errorf("Leading = %v, want nil (default)", par.Leading)
	}
	if par.Justify != nil {
		t.Errorf("Justify = %v, want nil (default)", par.Justify)
	}
	if par.Linebreaks != nil {
		t.Errorf("Linebreaks = %v, want nil (default)", par.Linebreaks)
	}
	if par.FirstLineIndent != nil {
		t.Errorf("FirstLineIndent = %v, want nil (default)", par.FirstLineIndent)
	}
	if par.HangingIndent != nil {
		t.Errorf("HangingIndent = %v, want nil (default)", par.HangingIndent)
	}
}

func TestParNativeWithJustify(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test paragraph"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("justify", True, syntax.Detached())

	result, err := parNative(vm, args)
	if err != nil {
		t.Fatalf("parNative() error: %v", err)
	}

	content := result.(ContentValue)
	par := content.Content.Elements[0].(*ParagraphElement)

	if par.Justify == nil || *par.Justify != true {
		t.Errorf("Justify = %v, want true", par.Justify)
	}
}

func TestParNativeWithLinebreaks(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test paragraph"}},
	}}

	// Test with "simple"
	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("linebreaks", Str("simple"), syntax.Detached())

	result, err := parNative(vm, args)
	if err != nil {
		t.Fatalf("parNative() error: %v", err)
	}

	content := result.(ContentValue)
	par := content.Content.Elements[0].(*ParagraphElement)

	if par.Linebreaks == nil || *par.Linebreaks != "simple" {
		t.Errorf("Linebreaks = %v, want 'simple'", par.Linebreaks)
	}

	// Test with "optimized"
	args = NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("linebreaks", Str("optimized"), syntax.Detached())

	result, err = parNative(vm, args)
	if err != nil {
		t.Fatalf("parNative() error: %v", err)
	}

	content = result.(ContentValue)
	par = content.Content.Elements[0].(*ParagraphElement)

	if par.Linebreaks == nil || *par.Linebreaks != "optimized" {
		t.Errorf("Linebreaks = %v, want 'optimized'", par.Linebreaks)
	}
}

func TestParNativeWithInvalidLinebreaks(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test paragraph"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("linebreaks", Str("invalid"), syntax.Detached())

	_, err := parNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid linebreaks value")
	}
}

func TestParNativeWithFirstLineIndent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test paragraph"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("first-line-indent", LengthValue{Length: Length{Points: 20}}, syntax.Detached())

	result, err := parNative(vm, args)
	if err != nil {
		t.Fatalf("parNative() error: %v", err)
	}

	content := result.(ContentValue)
	par := content.Content.Elements[0].(*ParagraphElement)

	if par.FirstLineIndent == nil || *par.FirstLineIndent != 20 {
		t.Errorf("FirstLineIndent = %v, want 20", par.FirstLineIndent)
	}
}

func TestParNativeWithLeading(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test paragraph"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("leading", LengthValue{Length: Length{Points: 14}}, syntax.Detached())

	result, err := parNative(vm, args)
	if err != nil {
		t.Fatalf("parNative() error: %v", err)
	}

	content := result.(ContentValue)
	par := content.Content.Elements[0].(*ParagraphElement)

	if par.Leading == nil || *par.Leading != 14 {
		t.Errorf("Leading = %v, want 14", par.Leading)
	}
}

func TestParNativeWithHangingIndent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test paragraph"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("hanging-indent", LengthValue{Length: Length{Points: 12}}, syntax.Detached())

	result, err := parNative(vm, args)
	if err != nil {
		t.Fatalf("parNative() error: %v", err)
	}

	content := result.(ContentValue)
	par := content.Content.Elements[0].(*ParagraphElement)

	if par.HangingIndent == nil || *par.HangingIndent != 12 {
		t.Errorf("HangingIndent = %v, want 12", par.HangingIndent)
	}
}

func TestParNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Full paragraph"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("leading", LengthValue{Length: Length{Points: 14}}, syntax.Detached())
	args.PushNamed("justify", True, syntax.Detached())
	args.PushNamed("linebreaks", Str("optimized"), syntax.Detached())
	args.PushNamed("first-line-indent", LengthValue{Length: Length{Points: 20}}, syntax.Detached())
	args.PushNamed("hanging-indent", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	result, err := parNative(vm, args)
	if err != nil {
		t.Fatalf("parNative() error: %v", err)
	}

	content := result.(ContentValue)
	par := content.Content.Elements[0].(*ParagraphElement)

	if par.Leading == nil || *par.Leading != 14 {
		t.Errorf("Leading = %v, want 14", par.Leading)
	}
	if par.Justify == nil || *par.Justify != true {
		t.Errorf("Justify = %v, want true", par.Justify)
	}
	if par.Linebreaks == nil || *par.Linebreaks != "optimized" {
		t.Errorf("Linebreaks = %v, want 'optimized'", par.Linebreaks)
	}
	if par.FirstLineIndent == nil || *par.FirstLineIndent != 20 {
		t.Errorf("FirstLineIndent = %v, want 20", par.FirstLineIndent)
	}
	if par.HangingIndent == nil || *par.HangingIndent != 10 {
		t.Errorf("HangingIndent = %v, want 10", par.HangingIndent)
	}
}

func TestParNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("justify", True, syntax.Detached())

	_, err := parNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
	}
}

func TestParNativeWrongBodyType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("not content"), syntax.Detached())

	_, err := parNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong body type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestParNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := parNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestParagraphElement(t *testing.T) {
	// Test ParagraphElement struct and ContentElement interface
	leading := 14.0
	justify := true
	linebreaks := "optimized"
	fli := 20.0
	hi := 10.0

	elem := &ParagraphElement{
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Hello, World!"}},
		},
		Leading:         &leading,
		Justify:         &justify,
		Linebreaks:      &linebreaks,
		FirstLineIndent: &fli,
		HangingIndent:   &hi,
	}

	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}
	if *elem.Leading != 14.0 {
		t.Errorf("Leading = %v, want 14.0", *elem.Leading)
	}
	if *elem.Justify != true {
		t.Errorf("Justify = %v, want true", *elem.Justify)
	}
	if *elem.Linebreaks != "optimized" {
		t.Errorf("Linebreaks = %v, want 'optimized'", *elem.Linebreaks)
	}
	if *elem.FirstLineIndent != 20.0 {
		t.Errorf("FirstLineIndent = %v, want 20.0", *elem.FirstLineIndent)
	}
	if *elem.HangingIndent != 10.0 {
		t.Errorf("HangingIndent = %v, want 10.0", *elem.HangingIndent)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Parbreak Tests
// ----------------------------------------------------------------------------

func TestParbreakFunc(t *testing.T) {
	// Get the parbreak function
	pbFunc := ParbreakFunc()

	if pbFunc == nil {
		t.Fatal("ParbreakFunc() returned nil")
	}

	if pbFunc.Name == nil || *pbFunc.Name != "parbreak" {
		t.Errorf("expected function name 'parbreak', got %v", pbFunc.Name)
	}

	// Verify it's a native function
	_, ok := pbFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestParbreakNative(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())

	result, err := parbreakNative(vm, args)
	if err != nil {
		t.Fatalf("parbreakNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one ParbreakElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	_, ok = content.Content.Elements[0].(*ParbreakElement)
	if !ok {
		t.Fatalf("expected *ParbreakElement, got %T", content.Content.Elements[0])
	}
}

func TestParbreakNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unexpected", Str("value"), syntax.Detached())

	_, err := parbreakNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestParbreakElement(t *testing.T) {
	// Test ParbreakElement struct and ContentElement interface
	elem := &ParbreakElement{}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Par and Parbreak
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesParAndParbreak(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify par function is registered
	parBinding := scope.Get("par")
	if parBinding == nil {
		t.Fatal("expected 'par' to be registered")
	}

	parFunc, ok := parBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for par, got %T", parBinding.Value)
	}
	if parFunc.Func.Name == nil || *parFunc.Func.Name != "par" {
		t.Errorf("expected function name 'par', got %v", parFunc.Func.Name)
	}

	// Verify parbreak function is registered
	pbBinding := scope.Get("parbreak")
	if pbBinding == nil {
		t.Fatal("expected 'parbreak' to be registered")
	}

	pbFunc, ok := pbBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for parbreak, got %T", pbBinding.Value)
	}
	if pbFunc.Func.Name == nil || *pbFunc.Func.Name != "parbreak" {
		t.Errorf("expected function name 'parbreak', got %v", pbFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesParAndParbreak(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["par"]; !ok {
		t.Error("expected 'par' in ElementFunctions()")
	}

	if _, ok := funcs["parbreak"]; !ok {
		t.Error("expected 'parbreak' in ElementFunctions()")
	}
}

// ----------------------------------------------------------------------------
// Grid Tests
// ----------------------------------------------------------------------------

func TestGridFunc(t *testing.T) {
	gridFunc := GridFunc()

	if gridFunc == nil {
		t.Fatal("GridFunc() returned nil")
	}

	if gridFunc.Name == nil || *gridFunc.Name != "grid" {
		t.Errorf("expected function name 'grid', got %v", gridFunc.Name)
	}

	_, ok := gridFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestGridNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with children
	args := NewArgs(syntax.Detached())
	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell 1"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell 2"}},
	}}
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

	result, err := gridNative(vm, args)
	if err != nil {
		t.Fatalf("gridNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	grid, ok := content.Content.Elements[0].(*GridElement)
	if !ok {
		t.Fatalf("expected *GridElement, got %T", content.Content.Elements[0])
	}

	// Verify children
	if len(grid.Children) != 2 {
		t.Errorf("Children count = %d, want 2", len(grid.Children))
	}

	// Verify defaults
	if grid.Columns != nil {
		t.Errorf("Columns = %v, want nil (default)", grid.Columns)
	}
	if grid.Rows != nil {
		t.Errorf("Rows = %v, want nil (default)", grid.Rows)
	}
	if grid.ColumnGutter != nil {
		t.Errorf("ColumnGutter = %v, want nil (default)", grid.ColumnGutter)
	}
	if grid.RowGutter != nil {
		t.Errorf("RowGutter = %v, want nil (default)", grid.RowGutter)
	}
}

func TestGridNativeWithColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with columns as array
	args := NewArgs(syntax.Detached())
	columns := ArrayValue{Auto, LengthValue{Length: Length{Points: 100}}}
	args.PushNamed("columns", columns, syntax.Detached())

	result, err := gridNative(vm, args)
	if err != nil {
		t.Fatalf("gridNative() error: %v", err)
	}

	content := result.(ContentValue)
	grid := content.Content.Elements[0].(*GridElement)

	if len(grid.Columns) != 2 {
		t.Fatalf("Columns count = %d, want 2", len(grid.Columns))
	}

	if !grid.Columns[0].Auto {
		t.Errorf("Columns[0].Auto = false, want true")
	}
	if grid.Columns[1].Length == nil || *grid.Columns[1].Length != 100 {
		t.Errorf("Columns[1].Length = %v, want 100", grid.Columns[1].Length)
	}
}

func TestGridNativeWithIntColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// columns: 3 should create 3 auto columns
	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Int(3), syntax.Detached())

	result, err := gridNative(vm, args)
	if err != nil {
		t.Fatalf("gridNative() error: %v", err)
	}

	content := result.(ContentValue)
	grid := content.Content.Elements[0].(*GridElement)

	if len(grid.Columns) != 3 {
		t.Fatalf("Columns count = %d, want 3", len(grid.Columns))
	}

	for i, col := range grid.Columns {
		if !col.Auto {
			t.Errorf("Columns[%d].Auto = false, want true", i)
		}
	}
}

func TestGridNativeWithGutter(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("gutter", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	result, err := gridNative(vm, args)
	if err != nil {
		t.Fatalf("gridNative() error: %v", err)
	}

	content := result.(ContentValue)
	grid := content.Content.Elements[0].(*GridElement)

	if grid.ColumnGutter == nil || *grid.ColumnGutter != 10 {
		t.Errorf("ColumnGutter = %v, want 10", grid.ColumnGutter)
	}
	if grid.RowGutter == nil || *grid.RowGutter != 10 {
		t.Errorf("RowGutter = %v, want 10", grid.RowGutter)
	}
}

func TestGridNativeWithSeparateGutters(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("column-gutter", LengthValue{Length: Length{Points: 15}}, syntax.Detached())
	args.PushNamed("row-gutter", LengthValue{Length: Length{Points: 20}}, syntax.Detached())

	result, err := gridNative(vm, args)
	if err != nil {
		t.Fatalf("gridNative() error: %v", err)
	}

	content := result.(ContentValue)
	grid := content.Content.Elements[0].(*GridElement)

	if grid.ColumnGutter == nil || *grid.ColumnGutter != 15 {
		t.Errorf("ColumnGutter = %v, want 15", grid.ColumnGutter)
	}
	if grid.RowGutter == nil || *grid.RowGutter != 20 {
		t.Errorf("RowGutter = %v, want 20", grid.RowGutter)
	}
}

func TestGridNativeWithAlign(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("align", Str("center"), syntax.Detached())

	result, err := gridNative(vm, args)
	if err != nil {
		t.Fatalf("gridNative() error: %v", err)
	}

	content := result.(ContentValue)
	grid := content.Content.Elements[0].(*GridElement)

	if grid.Align == nil || *grid.Align != "center" {
		t.Errorf("Align = %v, want 'center'", grid.Align)
	}
}

func TestGridNativeWithFrColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	columns := ArrayValue{
		FractionValue{Fraction: Fraction{Value: 1}},
		FractionValue{Fraction: Fraction{Value: 2}},
	}
	args.PushNamed("columns", columns, syntax.Detached())

	result, err := gridNative(vm, args)
	if err != nil {
		t.Fatalf("gridNative() error: %v", err)
	}

	content := result.(ContentValue)
	grid := content.Content.Elements[0].(*GridElement)

	if len(grid.Columns) != 2 {
		t.Fatalf("Columns count = %d, want 2", len(grid.Columns))
	}

	if grid.Columns[0].Fr == nil || *grid.Columns[0].Fr != 1 {
		t.Errorf("Columns[0].Fr = %v, want 1", grid.Columns[0].Fr)
	}
	if grid.Columns[1].Fr == nil || *grid.Columns[1].Fr != 2 {
		t.Errorf("Columns[1].Fr = %v, want 2", grid.Columns[1].Fr)
	}
}

func TestGridNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := gridNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestGridElement(t *testing.T) {
	colLen := 100.0
	fr := 1.0
	colGutter := 10.0
	rowGutter := 15.0
	align := "center"

	elem := &GridElement{
		Columns: []TrackSizing{
			{Auto: true},
			{Length: &colLen},
			{Fr: &fr},
		},
		Rows:         []TrackSizing{{Auto: true}},
		ColumnGutter: &colGutter,
		RowGutter:    &rowGutter,
		Align:        &align,
		Children: []Content{
			{Elements: []ContentElement{&TextElement{Text: "Cell"}}},
		},
	}

	if len(elem.Columns) != 3 {
		t.Errorf("Columns count = %d, want 3", len(elem.Columns))
	}
	if !elem.Columns[0].Auto {
		t.Errorf("Columns[0].Auto = false, want true")
	}
	if *elem.Columns[1].Length != 100 {
		t.Errorf("Columns[1].Length = %v, want 100", *elem.Columns[1].Length)
	}
	if *elem.Columns[2].Fr != 1 {
		t.Errorf("Columns[2].Fr = %v, want 1", *elem.Columns[2].Fr)
	}
	if *elem.ColumnGutter != 10 {
		t.Errorf("ColumnGutter = %v, want 10", *elem.ColumnGutter)
	}
	if *elem.RowGutter != 15 {
		t.Errorf("RowGutter = %v, want 15", *elem.RowGutter)
	}
	if *elem.Align != "center" {
		t.Errorf("Align = %v, want 'center'", *elem.Align)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Columns Tests
// ----------------------------------------------------------------------------

func TestColumnsFunc(t *testing.T) {
	colsFunc := ColumnsFunc()

	if colsFunc == nil {
		t.Fatal("ColumnsFunc() returned nil")
	}

	if colsFunc.Name == nil || *colsFunc.Name != "columns" {
		t.Errorf("expected function name 'columns', got %v", colsFunc.Name)
	}

	_, ok := colsFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestColumnsNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Column content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	cols, ok := content.Content.Elements[0].(*ColumnsElement)
	if !ok {
		t.Fatalf("expected *ColumnsElement, got %T", content.Content.Elements[0])
	}

	// Verify defaults
	if cols.Count != 2 {
		t.Errorf("Count = %d, want 2 (default)", cols.Count)
	}
	if cols.Gutter != nil {
		t.Errorf("Gutter = %v, want nil (default)", cols.Gutter)
	}
	if len(cols.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(cols.Body.Elements))
	}
}

func TestColumnsNativeWithCount(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Column content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("count", Int(3), syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	content := result.(ContentValue)
	cols := content.Content.Elements[0].(*ColumnsElement)

	if cols.Count != 3 {
		t.Errorf("Count = %d, want 3", cols.Count)
	}
}

func TestColumnsNativeWithGutter(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Column content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("gutter", LengthValue{Length: Length{Points: 12}}, syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	content := result.(ContentValue)
	cols := content.Content.Elements[0].(*ColumnsElement)

	if cols.Gutter == nil || *cols.Gutter != 12 {
		t.Errorf("Gutter = %v, want 12", cols.Gutter)
	}
}

func TestColumnsNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Full columns"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("count", Int(4), syntax.Detached())
	args.PushNamed("gutter", LengthValue{Length: Length{Points: 20}}, syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	content := result.(ContentValue)
	cols := content.Content.Elements[0].(*ColumnsElement)

	if cols.Count != 4 {
		t.Errorf("Count = %d, want 4", cols.Count)
	}
	if cols.Gutter == nil || *cols.Gutter != 20 {
		t.Errorf("Gutter = %v, want 20", cols.Gutter)
	}
}

func TestColumnsNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("count", Int(3), syntax.Detached())

	_, err := columnsNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
	}
}

func TestColumnsNativeWrongCountType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("count", Str("three"), syntax.Detached())
	args.Push(body, syntax.Detached())

	_, err := columnsNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong count type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestColumnsNativeWrongBodyType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("not content"), syntax.Detached())

	_, err := columnsNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong body type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestColumnsNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := columnsNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestColumnsElement(t *testing.T) {
	gutter := 15.0

	elem := &ColumnsElement{
		Count:  3,
		Gutter: &gutter,
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Column content"}},
		},
	}

	if elem.Count != 3 {
		t.Errorf("Count = %d, want 3", elem.Count)
	}
	if *elem.Gutter != 15 {
		t.Errorf("Gutter = %v, want 15", *elem.Gutter)
	}
	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Grid and Columns
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesGridAndColumns(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify grid function is registered
	gridBinding := scope.Get("grid")
	if gridBinding == nil {
		t.Fatal("expected 'grid' to be registered")
	}

	gridFunc, ok := gridBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for grid, got %T", gridBinding.Value)
	}
	if gridFunc.Func.Name == nil || *gridFunc.Func.Name != "grid" {
		t.Errorf("expected function name 'grid', got %v", gridFunc.Func.Name)
	}

	// Verify columns function is registered
	colsBinding := scope.Get("columns")
	if colsBinding == nil {
		t.Fatal("expected 'columns' to be registered")
	}

	colsFunc, ok := colsBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for columns, got %T", colsBinding.Value)
	}
	if colsFunc.Func.Name == nil || *colsFunc.Func.Name != "columns" {
		t.Errorf("expected function name 'columns', got %v", colsFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesGridAndColumns(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["grid"]; !ok {
		t.Error("expected 'grid' in ElementFunctions()")
	}

	if _, ok := funcs["columns"]; !ok {
		t.Error("expected 'columns' in ElementFunctions()")
	}
}
