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
// Box Tests
// ----------------------------------------------------------------------------

func TestBoxFunc(t *testing.T) {
	// Get the box function
	boxFunc := BoxFunc()

	if boxFunc == nil {
		t.Fatal("BoxFunc() returned nil")
	}

	if boxFunc.Name == nil || *boxFunc.Name != "box" {
		t.Errorf("expected function name 'box', got %v", boxFunc.Name)
	}

	// Verify it's a native function
	_, ok := boxFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestBoxNativeEmpty(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with no body (empty box)
	args := NewArgs(syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one BoxElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	box, ok := content.Content.Elements[0].(*BoxElement)
	if !ok {
		t.Fatalf("expected *BoxElement, got %T", content.Content.Elements[0])
	}

	// Verify defaults
	if len(box.Body.Elements) != 0 {
		t.Errorf("Body elements = %d, want 0", len(box.Body.Elements))
	}
	if box.Width != nil {
		t.Errorf("Width = %v, want nil (default)", box.Width)
	}
	if box.Height != nil {
		t.Errorf("Height = %v, want nil (default)", box.Height)
	}
	if box.Baseline != nil {
		t.Errorf("Baseline = %v, want nil (default)", box.Baseline)
	}
	if box.Fill != nil {
		t.Errorf("Fill = %v, want nil (default)", box.Fill)
	}
	if box.Inset != nil {
		t.Errorf("Inset = %v, want nil (default)", box.Inset)
	}
	if box.Outset != nil {
		t.Errorf("Outset = %v, want nil (default)", box.Outset)
	}
	if box.Radius != nil {
		t.Errorf("Radius = %v, want nil (default)", box.Radius)
	}
	if box.Clip != nil {
		t.Errorf("Clip = %v, want nil (default)", box.Clip)
	}
}

func TestBoxNativeWithBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Hello, Box!"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if len(box.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(box.Body.Elements))
	}
}

func TestBoxNativeWithWidth(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("width", LengthValue{Length: Length{Points: 100}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Width == nil || *box.Width != 100 {
		t.Errorf("Width = %v, want 100", box.Width)
	}
}

func TestBoxNativeWithHeight(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("height", LengthValue{Length: Length{Points: 50}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Height == nil || *box.Height != 50 {
		t.Errorf("Height = %v, want 50", box.Height)
	}
}

func TestBoxNativeWithBaseline(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("baseline", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Baseline == nil || *box.Baseline != 10 {
		t.Errorf("Baseline = %v, want 10", box.Baseline)
	}
}

func TestBoxNativeWithFill(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("fill", ColorValue{Color: Color{R: 255, G: 0, B: 0, A: 255}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Fill == nil {
		t.Fatal("Fill = nil, want non-nil")
	}
	if box.Fill.R != 255 || box.Fill.G != 0 || box.Fill.B != 0 {
		t.Errorf("Fill = %v, want red", box.Fill)
	}
}

func TestBoxNativeWithInset(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("inset", LengthValue{Length: Length{Points: 5}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Inset == nil || *box.Inset != 5 {
		t.Errorf("Inset = %v, want 5", box.Inset)
	}
}

func TestBoxNativeWithRadius(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("radius", LengthValue{Length: Length{Points: 4}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Radius == nil || *box.Radius != 4 {
		t.Errorf("Radius = %v, want 4", box.Radius)
	}
}

func TestBoxNativeWithClip(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("clip", False, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Clip == nil || *box.Clip != false {
		t.Errorf("Clip = %v, want false", box.Clip)
	}
}

func TestBoxNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Full box"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("width", LengthValue{Length: Length{Points: 100}}, syntax.Detached())
	args.PushNamed("height", LengthValue{Length: Length{Points: 50}}, syntax.Detached())
	args.PushNamed("baseline", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.PushNamed("fill", ColorValue{Color: Color{R: 0, G: 255, B: 0, A: 255}}, syntax.Detached())
	args.PushNamed("inset", LengthValue{Length: Length{Points: 5}}, syntax.Detached())
	args.PushNamed("outset", LengthValue{Length: Length{Points: 2}}, syntax.Detached())
	args.PushNamed("radius", LengthValue{Length: Length{Points: 4}}, syntax.Detached())
	args.PushNamed("clip", True, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if len(box.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(box.Body.Elements))
	}
	if box.Width == nil || *box.Width != 100 {
		t.Errorf("Width = %v, want 100", box.Width)
	}
	if box.Height == nil || *box.Height != 50 {
		t.Errorf("Height = %v, want 50", box.Height)
	}
	if box.Baseline == nil || *box.Baseline != 10 {
		t.Errorf("Baseline = %v, want 10", box.Baseline)
	}
	if box.Fill == nil || box.Fill.G != 255 {
		t.Errorf("Fill = %v, want green", box.Fill)
	}
	if box.Inset == nil || *box.Inset != 5 {
		t.Errorf("Inset = %v, want 5", box.Inset)
	}
	if box.Outset == nil || *box.Outset != 2 {
		t.Errorf("Outset = %v, want 2", box.Outset)
	}
	if box.Radius == nil || *box.Radius != 4 {
		t.Errorf("Radius = %v, want 4", box.Radius)
	}
	if box.Clip == nil || *box.Clip != true {
		t.Errorf("Clip = %v, want true", box.Clip)
	}
}

func TestBoxNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := boxNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestBoxNativeWrongFillType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("fill", Str("not a color"), syntax.Detached())

	_, err := boxNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong fill type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestBoxElement(t *testing.T) {
	// Test BoxElement struct and ContentElement interface
	width := 100.0
	height := 50.0
	baseline := 10.0
	fill := Color{R: 255, G: 0, B: 0, A: 255}
	inset := 5.0
	outset := 2.0
	radius := 4.0
	clip := true

	elem := &BoxElement{
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Hello, Box!"}},
		},
		Width:    &width,
		Height:   &height,
		Baseline: &baseline,
		Fill:     &fill,
		Inset:    &inset,
		Outset:   &outset,
		Radius:   &radius,
		Clip:     &clip,
	}

	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}
	if *elem.Width != 100.0 {
		t.Errorf("Width = %v, want 100.0", *elem.Width)
	}
	if *elem.Height != 50.0 {
		t.Errorf("Height = %v, want 50.0", *elem.Height)
	}
	if *elem.Baseline != 10.0 {
		t.Errorf("Baseline = %v, want 10.0", *elem.Baseline)
	}
	if elem.Fill.R != 255 {
		t.Errorf("Fill.R = %v, want 255", elem.Fill.R)
	}
	if *elem.Inset != 5.0 {
		t.Errorf("Inset = %v, want 5.0", *elem.Inset)
	}
	if *elem.Outset != 2.0 {
		t.Errorf("Outset = %v, want 2.0", *elem.Outset)
	}
	if *elem.Radius != 4.0 {
		t.Errorf("Radius = %v, want 4.0", *elem.Radius)
	}
	if *elem.Clip != true {
		t.Errorf("Clip = %v, want true", *elem.Clip)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Block Tests
// ----------------------------------------------------------------------------

func TestBlockFunc(t *testing.T) {
	// Get the block function
	blockFunc := BlockFunc()

	if blockFunc == nil {
		t.Fatal("BlockFunc() returned nil")
	}

	if blockFunc.Name == nil || *blockFunc.Name != "block" {
		t.Errorf("expected function name 'block', got %v", blockFunc.Name)
	}

	// Verify it's a native function
	_, ok := blockFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestBlockNativeEmpty(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with no body (empty block)
	args := NewArgs(syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one BlockElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	block, ok := content.Content.Elements[0].(*BlockElement)
	if !ok {
		t.Fatalf("expected *BlockElement, got %T", content.Content.Elements[0])
	}

	// Verify defaults
	if len(block.Body.Elements) != 0 {
		t.Errorf("Body elements = %d, want 0", len(block.Body.Elements))
	}
	if block.Width != nil {
		t.Errorf("Width = %v, want nil (default)", block.Width)
	}
	if block.Height != nil {
		t.Errorf("Height = %v, want nil (default)", block.Height)
	}
	if block.Fill != nil {
		t.Errorf("Fill = %v, want nil (default)", block.Fill)
	}
	if block.Inset != nil {
		t.Errorf("Inset = %v, want nil (default)", block.Inset)
	}
	if block.Outset != nil {
		t.Errorf("Outset = %v, want nil (default)", block.Outset)
	}
	if block.Radius != nil {
		t.Errorf("Radius = %v, want nil (default)", block.Radius)
	}
	if block.Clip != nil {
		t.Errorf("Clip = %v, want nil (default)", block.Clip)
	}
	if block.Breakable != nil {
		t.Errorf("Breakable = %v, want nil (default)", block.Breakable)
	}
	if block.Above != nil {
		t.Errorf("Above = %v, want nil (default)", block.Above)
	}
	if block.Below != nil {
		t.Errorf("Below = %v, want nil (default)", block.Below)
	}
}

func TestBlockNativeWithBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Hello, Block!"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if len(block.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(block.Body.Elements))
	}
}

func TestBlockNativeWithWidth(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("width", LengthValue{Length: Length{Points: 200}}, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if block.Width == nil || *block.Width != 200 {
		t.Errorf("Width = %v, want 200", block.Width)
	}
}

func TestBlockNativeWithBreakable(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("breakable", False, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if block.Breakable == nil || *block.Breakable != false {
		t.Errorf("Breakable = %v, want false", block.Breakable)
	}
}

func TestBlockNativeWithAboveBelow(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("above", LengthValue{Length: Length{Points: 12}}, syntax.Detached())
	args.PushNamed("below", LengthValue{Length: Length{Points: 8}}, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if block.Above == nil || *block.Above != 12 {
		t.Errorf("Above = %v, want 12", block.Above)
	}
	if block.Below == nil || *block.Below != 8 {
		t.Errorf("Below = %v, want 8", block.Below)
	}
}

func TestBlockNativeWithFill(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("fill", ColorValue{Color: Color{R: 0, G: 0, B: 255, A: 255}}, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if block.Fill == nil {
		t.Fatal("Fill = nil, want non-nil")
	}
	if block.Fill.B != 255 {
		t.Errorf("Fill.B = %v, want 255", block.Fill.B)
	}
}

func TestBlockNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Full block"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("width", LengthValue{Length: Length{Points: 200}}, syntax.Detached())
	args.PushNamed("height", LengthValue{Length: Length{Points: 100}}, syntax.Detached())
	args.PushNamed("fill", ColorValue{Color: Color{R: 128, G: 128, B: 128, A: 255}}, syntax.Detached())
	args.PushNamed("inset", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.PushNamed("outset", LengthValue{Length: Length{Points: 3}}, syntax.Detached())
	args.PushNamed("radius", LengthValue{Length: Length{Points: 6}}, syntax.Detached())
	args.PushNamed("clip", False, syntax.Detached())
	args.PushNamed("breakable", True, syntax.Detached())
	args.PushNamed("above", LengthValue{Length: Length{Points: 12}}, syntax.Detached())
	args.PushNamed("below", LengthValue{Length: Length{Points: 8}}, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if len(block.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(block.Body.Elements))
	}
	if block.Width == nil || *block.Width != 200 {
		t.Errorf("Width = %v, want 200", block.Width)
	}
	if block.Height == nil || *block.Height != 100 {
		t.Errorf("Height = %v, want 100", block.Height)
	}
	if block.Fill == nil || block.Fill.R != 128 {
		t.Errorf("Fill = %v, want gray", block.Fill)
	}
	if block.Inset == nil || *block.Inset != 10 {
		t.Errorf("Inset = %v, want 10", block.Inset)
	}
	if block.Outset == nil || *block.Outset != 3 {
		t.Errorf("Outset = %v, want 3", block.Outset)
	}
	if block.Radius == nil || *block.Radius != 6 {
		t.Errorf("Radius = %v, want 6", block.Radius)
	}
	if block.Clip == nil || *block.Clip != false {
		t.Errorf("Clip = %v, want false", block.Clip)
	}
	if block.Breakable == nil || *block.Breakable != true {
		t.Errorf("Breakable = %v, want true", block.Breakable)
	}
	if block.Above == nil || *block.Above != 12 {
		t.Errorf("Above = %v, want 12", block.Above)
	}
	if block.Below == nil || *block.Below != 8 {
		t.Errorf("Below = %v, want 8", block.Below)
	}
}

func TestBlockNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := blockNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestBlockNativeWrongFillType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("fill", Int(42), syntax.Detached())

	_, err := blockNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong fill type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestBlockElement(t *testing.T) {
	// Test BlockElement struct and ContentElement interface
	width := 200.0
	height := 100.0
	fill := Color{R: 128, G: 128, B: 128, A: 255}
	inset := 10.0
	outset := 3.0
	radius := 6.0
	clip := false
	breakable := true
	above := 12.0
	below := 8.0

	elem := &BlockElement{
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Hello, Block!"}},
		},
		Width:     &width,
		Height:    &height,
		Fill:      &fill,
		Inset:     &inset,
		Outset:    &outset,
		Radius:    &radius,
		Clip:      &clip,
		Breakable: &breakable,
		Above:     &above,
		Below:     &below,
	}

	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}
	if *elem.Width != 200.0 {
		t.Errorf("Width = %v, want 200.0", *elem.Width)
	}
	if *elem.Height != 100.0 {
		t.Errorf("Height = %v, want 100.0", *elem.Height)
	}
	if elem.Fill.R != 128 {
		t.Errorf("Fill.R = %v, want 128", elem.Fill.R)
	}
	if *elem.Inset != 10.0 {
		t.Errorf("Inset = %v, want 10.0", *elem.Inset)
	}
	if *elem.Outset != 3.0 {
		t.Errorf("Outset = %v, want 3.0", *elem.Outset)
	}
	if *elem.Radius != 6.0 {
		t.Errorf("Radius = %v, want 6.0", *elem.Radius)
	}
	if *elem.Clip != false {
		t.Errorf("Clip = %v, want false", *elem.Clip)
	}
	if *elem.Breakable != true {
		t.Errorf("Breakable = %v, want true", *elem.Breakable)
	}
	if *elem.Above != 12.0 {
		t.Errorf("Above = %v, want 12.0", *elem.Above)
	}
	if *elem.Below != 8.0 {
		t.Errorf("Below = %v, want 8.0", *elem.Below)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Box and Block
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesBoxAndBlock(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify box function is registered
	boxBinding := scope.Get("box")
	if boxBinding == nil {
		t.Fatal("expected 'box' to be registered")
	}

	boxFunc, ok := boxBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for box, got %T", boxBinding.Value)
	}
	if boxFunc.Func.Name == nil || *boxFunc.Func.Name != "box" {
		t.Errorf("expected function name 'box', got %v", boxFunc.Func.Name)
	}

	// Verify block function is registered
	blockBinding := scope.Get("block")
	if blockBinding == nil {
		t.Fatal("expected 'block' to be registered")
	}

	blockFunc, ok := blockBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for block, got %T", blockBinding.Value)
	}
	if blockFunc.Func.Name == nil || *blockFunc.Func.Name != "block" {
		t.Errorf("expected function name 'block', got %v", blockFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesBoxAndBlock(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["box"]; !ok {
		t.Error("expected 'box' in ElementFunctions()")
	}

	if _, ok := funcs["block"]; !ok {
		t.Error("expected 'block' in ElementFunctions()")
	}
}
