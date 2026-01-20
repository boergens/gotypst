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
// Stack Tests
// ----------------------------------------------------------------------------

func TestStackFunc(t *testing.T) {
	// Get the stack function
	stackFunc := StackFunc()

	if stackFunc == nil {
		t.Fatal("StackFunc() returned nil")
	}

	if stackFunc.Name == nil || *stackFunc.Name != "stack" {
		t.Errorf("expected function name 'stack', got %v", stackFunc.Name)
	}

	// Verify it's a native function
	_, ok := stackFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestStackNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create content children
	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "First"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Second"}},
	}}

	// Create args with children
	args := NewArgs(syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

	result, err := stackNative(vm, args)
	if err != nil {
		t.Fatalf("stackNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one StackElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	stack, ok := content.Content.Elements[0].(*StackElement)
	if !ok {
		t.Fatalf("expected *StackElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties (defaults)
	if stack.Dir != StackTTB {
		t.Errorf("Dir = %q, want %q", stack.Dir, StackTTB)
	}
	if stack.Spacing != nil {
		t.Errorf("Spacing = %v, want nil", stack.Spacing)
	}
	if len(stack.Children) != 2 {
		t.Errorf("Children length = %d, want 2", len(stack.Children))
	}
}

func TestStackNativeWithDir(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	tests := []struct {
		dir      string
		expected StackDirection
	}{
		{"ltr", StackLTR},
		{"rtl", StackRTL},
		{"ttb", StackTTB},
		{"btt", StackBTT},
	}

	for _, tt := range tests {
		args := NewArgs(syntax.Detached())
		args.PushNamed("dir", Str(tt.dir), syntax.Detached())

		result, err := stackNative(vm, args)
		if err != nil {
			t.Fatalf("stackNative() with dir=%q error: %v", tt.dir, err)
		}

		content := result.(ContentValue)
		stack := content.Content.Elements[0].(*StackElement)

		if stack.Dir != tt.expected {
			t.Errorf("dir=%q: got Dir=%q, want %q", tt.dir, stack.Dir, tt.expected)
		}
	}
}

func TestStackNativeWithInvalidDir(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("dir", Str("invalid"), syntax.Detached())

	_, err := stackNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid dir value")
	}
}

func TestStackNativeWithSpacing(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("spacing", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	result, err := stackNative(vm, args)
	if err != nil {
		t.Fatalf("stackNative() error: %v", err)
	}

	content := result.(ContentValue)
	stack := content.Content.Elements[0].(*StackElement)

	if stack.Spacing == nil || *stack.Spacing != 10 {
		t.Errorf("Spacing = %v, want 10", stack.Spacing)
	}
}

func TestStackNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("dir", Str("ltr"), syntax.Detached())
	args.PushNamed("spacing", LengthValue{Length: Length{Points: 5}}, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := stackNative(vm, args)
	if err != nil {
		t.Fatalf("stackNative() error: %v", err)
	}

	content := result.(ContentValue)
	stack := content.Content.Elements[0].(*StackElement)

	if stack.Dir != StackLTR {
		t.Errorf("Dir = %q, want %q", stack.Dir, StackLTR)
	}
	if stack.Spacing == nil || *stack.Spacing != 5 {
		t.Errorf("Spacing = %v, want 5", stack.Spacing)
	}
	if len(stack.Children) != 1 {
		t.Errorf("Children length = %d, want 1", len(stack.Children))
	}
}

func TestStackNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := stackNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestStackElement(t *testing.T) {
	// Test StackElement struct and ContentElement interface
	spacing := 10.0
	elem := &StackElement{
		Dir:     StackLTR,
		Spacing: &spacing,
		Children: []Content{
			{Elements: []ContentElement{&TextElement{Text: "A"}}},
			{Elements: []ContentElement{&TextElement{Text: "B"}}},
		},
	}

	if elem.Dir != StackLTR {
		t.Errorf("Dir = %q, want %q", elem.Dir, StackLTR)
	}
	if *elem.Spacing != 10.0 {
		t.Errorf("Spacing = %v, want 10.0", *elem.Spacing)
	}
	if len(elem.Children) != 2 {
		t.Errorf("Children length = %d, want 2", len(elem.Children))
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Align Tests
// ----------------------------------------------------------------------------

func TestAlignFunc(t *testing.T) {
	// Get the align function
	alignFunc := AlignFunc()

	if alignFunc == nil {
		t.Fatal("AlignFunc() returned nil")
	}

	if alignFunc.Name == nil || *alignFunc.Name != "align" {
		t.Errorf("expected function name 'align', got %v", alignFunc.Name)
	}

	// Verify it's a native function
	_, ok := alignFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestAlignNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Centered text"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(Str("center"), syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := alignNative(vm, args)
	if err != nil {
		t.Fatalf("alignNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one AlignElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	align, ok := content.Content.Elements[0].(*AlignElement)
	if !ok {
		t.Fatalf("expected *AlignElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties
	if align.Alignment.Horizontal == nil || *align.Alignment.Horizontal != "center" {
		t.Errorf("Horizontal = %v, want 'center'", align.Alignment.Horizontal)
	}
	if len(align.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(align.Body.Elements))
	}
}

func TestAlignNativeWithDifferentAlignments(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test"}},
	}}

	tests := []struct {
		alignment    string
		expectHoriz  *string
		expectVert   *string
	}{
		{"left", stringPtr("left"), nil},
		{"center", stringPtr("center"), nil},
		{"right", stringPtr("right"), nil},
		{"top", nil, stringPtr("top")},
		{"horizon", nil, stringPtr("horizon")},
		{"bottom", nil, stringPtr("bottom")},
		{"start", stringPtr("start"), nil},
		{"end", stringPtr("end"), nil},
	}

	for _, tt := range tests {
		args := NewArgs(syntax.Detached())
		args.Push(Str(tt.alignment), syntax.Detached())
		args.Push(body, syntax.Detached())

		result, err := alignNative(vm, args)
		if err != nil {
			t.Fatalf("alignNative() with alignment=%q error: %v", tt.alignment, err)
		}

		content := result.(ContentValue)
		align := content.Content.Elements[0].(*AlignElement)

		if tt.expectHoriz != nil {
			if align.Alignment.Horizontal == nil || *align.Alignment.Horizontal != *tt.expectHoriz {
				t.Errorf("alignment=%q: Horizontal = %v, want %q", tt.alignment, align.Alignment.Horizontal, *tt.expectHoriz)
			}
		}
		if tt.expectVert != nil {
			if align.Alignment.Vertical == nil || *align.Alignment.Vertical != *tt.expectVert {
				t.Errorf("alignment=%q: Vertical = %v, want %q", tt.alignment, align.Alignment.Vertical, *tt.expectVert)
			}
		}
	}
}

// stringPtr is a helper to get a pointer to a string.
func stringPtr(s string) *string {
	return &s
}

func TestAlignNativeWithInvalidAlignment(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(Str("invalid"), syntax.Detached())
	args.Push(body, syntax.Detached())

	_, err := alignNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid alignment value")
	}
}

func TestAlignNativeMissingAlignment(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached()) // Only body, no alignment

	_, err := alignNative(vm, args)
	// This should pass since body is content but first arg is interpreted as alignment
	// Actually, it will fail because ContentValue is not a string for alignment
	if err == nil {
		t.Error("expected error when alignment is missing/wrong type")
	}
}

func TestAlignNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("center"), syntax.Detached())

	_, err := alignNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
	}
}

func TestAlignNativeWrongBodyType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("center"), syntax.Detached())
	args.Push(Str("not content"), syntax.Detached())

	_, err := alignNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong body type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestAlignNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Test"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(Str("center"), syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := alignNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestAlignElement(t *testing.T) {
	// Test AlignElement struct and ContentElement interface
	horiz := "center"
	elem := &AlignElement{
		Alignment: Alignment2D{
			Horizontal: &horiz,
			Vertical:   nil,
		},
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Centered"}},
		},
	}

	if *elem.Alignment.Horizontal != "center" {
		t.Errorf("Horizontal = %v, want 'center'", *elem.Alignment.Horizontal)
	}
	if elem.Alignment.Vertical != nil {
		t.Errorf("Vertical = %v, want nil", elem.Alignment.Vertical)
	}
	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Stack and Align
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesStackAndAlign(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify stack function is registered
	stackBinding := scope.Get("stack")
	if stackBinding == nil {
		t.Fatal("expected 'stack' to be registered")
	}

	stackFunc, ok := stackBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for stack, got %T", stackBinding.Value)
	}
	if stackFunc.Func.Name == nil || *stackFunc.Func.Name != "stack" {
		t.Errorf("expected function name 'stack', got %v", stackFunc.Func.Name)
	}

	// Verify align function is registered
	alignBinding := scope.Get("align")
	if alignBinding == nil {
		t.Fatal("expected 'align' to be registered")
	}

	alignFunc, ok := alignBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for align, got %T", alignBinding.Value)
	}
	if alignFunc.Func.Name == nil || *alignFunc.Func.Name != "align" {
		t.Errorf("expected function name 'align', got %v", alignFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesStackAndAlign(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["stack"]; !ok {
		t.Error("expected 'stack' in ElementFunctions()")
	}

	if _, ok := funcs["align"]; !ok {
		t.Error("expected 'align' in ElementFunctions()")
	}
}

// ----------------------------------------------------------------------------
// List Tests
// ----------------------------------------------------------------------------

func TestListFunc(t *testing.T) {
	// Get the list function
	listFunc := ListFunc()

	if listFunc == nil {
		t.Fatal("ListFunc() returned nil")
	}

	if listFunc.Name == nil || *listFunc.Name != "list" {
		t.Errorf("expected function name 'list', got %v", listFunc.Name)
	}

	// Verify it's a native function
	_, ok := listFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestListNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create list items as content
	item1 := ContentValue{Content: Content{
		Elements: []ContentElement{&ListItemElement{Content: Content{
			Elements: []ContentElement{&TextElement{Text: "First item"}},
		}}},
	}}
	item2 := ContentValue{Content: Content{
		Elements: []ContentElement{&ListItemElement{Content: Content{
			Elements: []ContentElement{&TextElement{Text: "Second item"}},
		}}},
	}}

	// Create args with children
	args := NewArgs(syntax.Detached())
	args.Push(item1, syntax.Detached())
	args.Push(item2, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one ListElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	list, ok := content.Content.Elements[0].(*ListElement)
	if !ok {
		t.Fatalf("expected *ListElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties
	if len(list.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(list.Items))
	}
	if list.Tight != nil {
		t.Errorf("Tight = %v, want nil (default)", list.Tight)
	}
	if list.Marker != nil {
		t.Errorf("Marker = %v, want nil (default)", list.Marker)
	}
}

func TestListNativeWithTight(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	item := ContentValue{Content: Content{
		Elements: []ContentElement{&ListItemElement{Content: Content{
			Elements: []ContentElement{&TextElement{Text: "Item"}},
		}}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("tight", False, syntax.Detached())
	args.Push(item, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Tight == nil || *list.Tight != false {
		t.Errorf("Tight = %v, want false", list.Tight)
	}
}

func TestListNativeWithMarker(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	item := ContentValue{Content: Content{
		Elements: []ContentElement{&ListItemElement{Content: Content{
			Elements: []ContentElement{&TextElement{Text: "Item"}},
		}}},
	}}

	// Test with string marker
	args := NewArgs(syntax.Detached())
	args.PushNamed("marker", Str("*"), syntax.Detached())
	args.Push(item, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Marker == nil {
		t.Fatal("Marker = nil, want non-nil")
	}
	if len(list.Marker.Elements) != 1 {
		t.Fatalf("Marker elements = %d, want 1", len(list.Marker.Elements))
	}
	text, ok := list.Marker.Elements[0].(*TextElement)
	if !ok {
		t.Fatalf("expected *TextElement, got %T", list.Marker.Elements[0])
	}
	if text.Text != "*" {
		t.Errorf("Marker text = %q, want '*'", text.Text)
	}
}

func TestListNativeWithContentAsItem(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Pass plain content (not ListItemElement), should be wrapped
	item := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Plain text item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(item, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if len(list.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(list.Items))
	}
	if len(list.Items[0].Content.Elements) != 1 {
		t.Errorf("Item content elements = %d, want 1", len(list.Items[0].Content.Elements))
	}
}

func TestListNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := listNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestListElement(t *testing.T) {
	// Test ListElement struct and ContentElement interface
	tight := true
	marker := Content{Elements: []ContentElement{&TextElement{Text: "•"}}}
	elem := &ListElement{
		Items: []*ListItemElement{
			{Content: Content{Elements: []ContentElement{&TextElement{Text: "Item 1"}}}},
			{Content: Content{Elements: []ContentElement{&TextElement{Text: "Item 2"}}}},
		},
		Tight:  &tight,
		Marker: &marker,
	}

	if len(elem.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(elem.Items))
	}
	if *elem.Tight != true {
		t.Errorf("Tight = %v, want true", *elem.Tight)
	}
	if elem.Marker == nil {
		t.Error("Marker = nil, want non-nil")
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Enum Tests
// ----------------------------------------------------------------------------

func TestEnumFunc(t *testing.T) {
	// Get the enum function
	enumFunc := EnumFunc()

	if enumFunc == nil {
		t.Fatal("EnumFunc() returned nil")
	}

	if enumFunc.Name == nil || *enumFunc.Name != "enum" {
		t.Errorf("expected function name 'enum', got %v", enumFunc.Name)
	}

	// Verify it's a native function
	_, ok := enumFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestEnumNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create enum items as content
	item1 := ContentValue{Content: Content{
		Elements: []ContentElement{&EnumItemElement{Number: 1, Content: Content{
			Elements: []ContentElement{&TextElement{Text: "First item"}},
		}}},
	}}
	item2 := ContentValue{Content: Content{
		Elements: []ContentElement{&EnumItemElement{Number: 2, Content: Content{
			Elements: []ContentElement{&TextElement{Text: "Second item"}},
		}}},
	}}

	// Create args with children
	args := NewArgs(syntax.Detached())
	args.Push(item1, syntax.Detached())
	args.Push(item2, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one EnumElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	enum, ok := content.Content.Elements[0].(*EnumElement)
	if !ok {
		t.Fatalf("expected *EnumElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties
	if len(enum.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(enum.Items))
	}
	if enum.Tight != nil {
		t.Errorf("Tight = %v, want nil (default)", enum.Tight)
	}
	if enum.Numbering != nil {
		t.Errorf("Numbering = %v, want nil (default)", enum.Numbering)
	}
	if enum.Start != nil {
		t.Errorf("Start = %v, want nil (default)", enum.Start)
	}
	if enum.Full != nil {
		t.Errorf("Full = %v, want nil (default)", enum.Full)
	}
}

func TestEnumNativeWithNumbering(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	item := ContentValue{Content: Content{
		Elements: []ContentElement{&EnumItemElement{Number: 1, Content: Content{
			Elements: []ContentElement{&TextElement{Text: "Item"}},
		}}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("numbering", Str("a)"), syntax.Detached())
	args.Push(item, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Numbering == nil || *enum.Numbering != "a)" {
		t.Errorf("Numbering = %v, want 'a)'", enum.Numbering)
	}
}

func TestEnumNativeWithStart(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	item := ContentValue{Content: Content{
		Elements: []ContentElement{&EnumItemElement{Number: 0, Content: Content{
			Elements: []ContentElement{&TextElement{Text: "Item"}},
		}}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("start", Int(5), syntax.Detached())
	args.Push(item, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Start == nil || *enum.Start != 5 {
		t.Errorf("Start = %v, want 5", enum.Start)
	}
	// Items with Number=0 should be auto-numbered starting from Start
	if len(enum.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(enum.Items))
	}
	if enum.Items[0].Number != 5 {
		t.Errorf("Item number = %d, want 5", enum.Items[0].Number)
	}
}

func TestEnumNativeWithFull(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	item := ContentValue{Content: Content{
		Elements: []ContentElement{&EnumItemElement{Number: 1, Content: Content{
			Elements: []ContentElement{&TextElement{Text: "Item"}},
		}}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("full", True, syntax.Detached())
	args.Push(item, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Full == nil || *enum.Full != true {
		t.Errorf("Full = %v, want true", enum.Full)
	}
}

func TestEnumNativeWithTight(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	item := ContentValue{Content: Content{
		Elements: []ContentElement{&EnumItemElement{Number: 1, Content: Content{
			Elements: []ContentElement{&TextElement{Text: "Item"}},
		}}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("tight", False, syntax.Detached())
	args.Push(item, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Tight == nil || *enum.Tight != false {
		t.Errorf("Tight = %v, want false", enum.Tight)
	}
}

func TestEnumNativeWithContentAsItem(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Pass plain content (not EnumItemElement), should be wrapped
	item := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Plain text item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(item, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if len(enum.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(enum.Items))
	}
	if enum.Items[0].Number != 1 {
		t.Errorf("Item number = %d, want 1", enum.Items[0].Number)
	}
	if len(enum.Items[0].Content.Elements) != 1 {
		t.Errorf("Item content elements = %d, want 1", len(enum.Items[0].Content.Elements))
	}
}

func TestEnumNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := enumNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestEnumElement(t *testing.T) {
	// Test EnumElement struct and ContentElement interface
	tight := true
	numbering := "1."
	start := 1
	full := false
	elem := &EnumElement{
		Items: []*EnumItemElement{
			{Number: 1, Content: Content{Elements: []ContentElement{&TextElement{Text: "Item 1"}}}},
			{Number: 2, Content: Content{Elements: []ContentElement{&TextElement{Text: "Item 2"}}}},
		},
		Tight:     &tight,
		Numbering: &numbering,
		Start:     &start,
		Full:      &full,
	}

	if len(elem.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(elem.Items))
	}
	if *elem.Tight != true {
		t.Errorf("Tight = %v, want true", *elem.Tight)
	}
	if *elem.Numbering != "1." {
		t.Errorf("Numbering = %v, want '1.'", *elem.Numbering)
	}
	if *elem.Start != 1 {
		t.Errorf("Start = %v, want 1", *elem.Start)
	}
	if *elem.Full != false {
		t.Errorf("Full = %v, want false", *elem.Full)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for List and Enum
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesListAndEnum(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify list function is registered
	listBinding := scope.Get("list")
	if listBinding == nil {
		t.Fatal("expected 'list' to be registered")
	}

	listFunc, ok := listBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for list, got %T", listBinding.Value)
	}
	if listFunc.Func.Name == nil || *listFunc.Func.Name != "list" {
		t.Errorf("expected function name 'list', got %v", listFunc.Func.Name)
	}

	// Verify enum function is registered
	enumBinding := scope.Get("enum")
	if enumBinding == nil {
		t.Fatal("expected 'enum' to be registered")
	}

	enumFunc, ok := enumBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for enum, got %T", enumBinding.Value)
	}
	if enumFunc.Func.Name == nil || *enumFunc.Func.Name != "enum" {
		t.Errorf("expected function name 'enum', got %v", enumFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesListAndEnum(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["list"]; !ok {
		t.Error("expected 'list' in ElementFunctions()")
	}

	if _, ok := funcs["enum"]; !ok {
		t.Error("expected 'enum' in ElementFunctions()")
	}
}

// ----------------------------------------------------------------------------
// Page Tests
// ----------------------------------------------------------------------------

func TestPageFunc(t *testing.T) {
	// Get the page function
	pageFunc := PageFunc()

	if pageFunc == nil {
		t.Fatal("PageFunc() returned nil")
	}

	if pageFunc.Name == nil || *pageFunc.Name != "page" {
		t.Errorf("expected function name 'page', got %v", pageFunc.Name)
	}

	// Verify it's a native function
	_, ok := pageFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestPageNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create body content
	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Page content"}},
	}}

	// Create args with just the body
	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one PageElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	page, ok := content.Content.Elements[0].(*PageElement)
	if !ok {
		t.Fatalf("expected *PageElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties (defaults)
	if len(page.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(page.Body.Elements))
	}
	if page.Paper != nil {
		t.Errorf("Paper = %v, want nil (default)", page.Paper)
	}
	if page.Width != nil {
		t.Errorf("Width = %v, want nil (default)", page.Width)
	}
	if page.Height != nil {
		t.Errorf("Height = %v, want nil (default)", page.Height)
	}
	if page.Margin != nil {
		t.Errorf("Margin = %v, want nil (default)", page.Margin)
	}
	if page.Header != nil {
		t.Errorf("Header = %v, want nil (default)", page.Header)
	}
	if page.Footer != nil {
		t.Errorf("Footer = %v, want nil (default)", page.Footer)
	}
}

func TestPageNativeWithPaper(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("paper", Str("a4"), syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Paper == nil || *page.Paper != "a4" {
		t.Errorf("Paper = %v, want 'a4'", page.Paper)
	}
}

func TestPageNativeWithInvalidPaper(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("paper", Str("invalid-paper-size"), syntax.Detached())

	_, err := pageNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid paper size")
	}
}

func TestPageNativeWithDimensions(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("width", LengthValue{Length: Length{Points: 612}}, syntax.Detached())
	args.PushNamed("height", LengthValue{Length: Length{Points: 792}}, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Width == nil || *page.Width != 612 {
		t.Errorf("Width = %v, want 612", page.Width)
	}
	if page.Height == nil || *page.Height != 792 {
		t.Errorf("Height = %v, want 792", page.Height)
	}
}

func TestPageNativeWithUniformMargin(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("margin", LengthValue{Length: Length{Points: 72}}, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Margin == nil || *page.Margin != 72 {
		t.Errorf("Margin = %v, want 72", page.Margin)
	}
}

func TestPageNativeWithMarginDict(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	// Create margin dictionary
	marginDict := NewDict()
	marginDict.Set("top", LengthValue{Length: Length{Points: 50}})
	marginDict.Set("bottom", LengthValue{Length: Length{Points: 50}})
	marginDict.Set("left", LengthValue{Length: Length{Points: 72}})
	marginDict.Set("right", LengthValue{Length: Length{Points: 72}})

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("margin", marginDict, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.MarginTop == nil || *page.MarginTop != 50 {
		t.Errorf("MarginTop = %v, want 50", page.MarginTop)
	}
	if page.MarginBottom == nil || *page.MarginBottom != 50 {
		t.Errorf("MarginBottom = %v, want 50", page.MarginBottom)
	}
	if page.MarginLeft == nil || *page.MarginLeft != 72 {
		t.Errorf("MarginLeft = %v, want 72", page.MarginLeft)
	}
	if page.MarginRight == nil || *page.MarginRight != 72 {
		t.Errorf("MarginRight = %v, want 72", page.MarginRight)
	}
}

func TestPageNativeWithMarginXY(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	// Create margin dictionary with x and y
	marginDict := NewDict()
	marginDict.Set("x", LengthValue{Length: Length{Points: 72}})
	marginDict.Set("y", LengthValue{Length: Length{Points: 50}})

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("margin", marginDict, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.MarginLeft == nil || *page.MarginLeft != 72 {
		t.Errorf("MarginLeft = %v, want 72", page.MarginLeft)
	}
	if page.MarginRight == nil || *page.MarginRight != 72 {
		t.Errorf("MarginRight = %v, want 72", page.MarginRight)
	}
	if page.MarginTop == nil || *page.MarginTop != 50 {
		t.Errorf("MarginTop = %v, want 50", page.MarginTop)
	}
	if page.MarginBottom == nil || *page.MarginBottom != 50 {
		t.Errorf("MarginBottom = %v, want 50", page.MarginBottom)
	}
}

func TestPageNativeWithHeader(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}
	header := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "My Header"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("header", header, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Header == nil {
		t.Fatal("Header = nil, want non-nil")
	}
	if len(page.Header.Elements) != 1 {
		t.Errorf("Header elements = %d, want 1", len(page.Header.Elements))
	}
}

func TestPageNativeWithFooter(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}
	footer := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Page 1"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("footer", footer, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Footer == nil {
		t.Fatal("Footer = nil, want non-nil")
	}
	if len(page.Footer.Elements) != 1 {
		t.Errorf("Footer elements = %d, want 1", len(page.Footer.Elements))
	}
}

func TestPageNativeWithNumbering(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("numbering", Str("1"), syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Numbering == nil || *page.Numbering != "1" {
		t.Errorf("Numbering = %v, want '1'", page.Numbering)
	}
}

func TestPageNativeWithBinding(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	// Test with "left"
	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("binding", Str("left"), syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Binding == nil || *page.Binding != "left" {
		t.Errorf("Binding = %v, want 'left'", page.Binding)
	}

	// Test with "right"
	args = NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("binding", Str("right"), syntax.Detached())

	result, err = pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content = result.(ContentValue)
	page = content.Content.Elements[0].(*PageElement)

	if page.Binding == nil || *page.Binding != "right" {
		t.Errorf("Binding = %v, want 'right'", page.Binding)
	}
}

func TestPageNativeWithInvalidBinding(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("binding", Str("invalid"), syntax.Detached())

	_, err := pageNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid binding value")
	}
}

func TestPageNativeWithColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("columns", Int(2), syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Columns == nil || *page.Columns != 2 {
		t.Errorf("Columns = %v, want 2", page.Columns)
	}
}

func TestPageNativeWithInvalidColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("columns", Int(0), syntax.Detached())

	_, err := pageNative(vm, args)
	if err == nil {
		t.Error("expected error for columns < 1")
	}
}

func TestPageNativeWithFlipped(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("flipped", True, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Flipped == nil || *page.Flipped != true {
		t.Errorf("Flipped = %v, want true", page.Flipped)
	}
}

func TestPageNativeWithBackground(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}
	background := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Background"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("background", background, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Background == nil {
		t.Fatal("Background = nil, want non-nil")
	}
}

func TestPageNativeWithForeground(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}
	foreground := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Foreground"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("foreground", foreground, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Foreground == nil {
		t.Fatal("Foreground = nil, want non-nil")
	}
}

func TestPageNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("paper", Str("a4"), syntax.Detached())

	_, err := pageNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
	}
}

func TestPageNativeWrongBodyType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("not content"), syntax.Detached())

	_, err := pageNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong body type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestPageNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := pageNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestPageElement(t *testing.T) {
	// Test PageElement struct and ContentElement interface
	paper := "a4"
	width := 595.0
	height := 842.0
	margin := 72.0
	numbering := "1"
	columns := 1

	elem := &PageElement{
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Page content"}},
		},
		Paper:     &paper,
		Width:     &width,
		Height:    &height,
		Margin:    &margin,
		Numbering: &numbering,
		Columns:   &columns,
	}

	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}
	if *elem.Paper != "a4" {
		t.Errorf("Paper = %v, want 'a4'", *elem.Paper)
	}
	if *elem.Width != 595.0 {
		t.Errorf("Width = %v, want 595.0", *elem.Width)
	}
	if *elem.Height != 842.0 {
		t.Errorf("Height = %v, want 842.0", *elem.Height)
	}
	if *elem.Margin != 72.0 {
		t.Errorf("Margin = %v, want 72.0", *elem.Margin)
	}
	if *elem.Numbering != "1" {
		t.Errorf("Numbering = %v, want '1'", *elem.Numbering)
	}
	if *elem.Columns != 1 {
		t.Errorf("Columns = %v, want 1", *elem.Columns)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Page
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesPage(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify page function is registered
	pageBinding := scope.Get("page")
	if pageBinding == nil {
		t.Fatal("expected 'page' to be registered")
	}

	pageFunc, ok := pageBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for page, got %T", pageBinding.Value)
	}
	if pageFunc.Func.Name == nil || *pageFunc.Func.Name != "page" {
		t.Errorf("expected function name 'page', got %v", pageFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesPage(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["page"]; !ok {
		t.Error("expected 'page' in ElementFunctions()")
	}
}

// ----------------------------------------------------------------------------
// Helper Tests for Page
// ----------------------------------------------------------------------------

func TestIsValidPaperSize(t *testing.T) {
	validSizes := []string{
		"a4", "a3", "a5", "us-letter", "us-legal",
		"jis-b5", "presentation-16-9",
	}
	for _, size := range validSizes {
		if !isValidPaperSize(size) {
			t.Errorf("isValidPaperSize(%q) = false, want true", size)
		}
	}

	invalidSizes := []string{
		"invalid", "A4", "letter", "custom",
	}
	for _, size := range invalidSizes {
		if isValidPaperSize(size) {
			t.Errorf("isValidPaperSize(%q) = true, want false", size)
		}
	}
}

func TestParseColorName(t *testing.T) {
	tests := []struct {
		name     string
		expected *Color
	}{
		{"black", &Color{R: 0, G: 0, B: 0, A: 255}},
		{"white", &Color{R: 255, G: 255, B: 255, A: 255}},
		{"red", &Color{R: 255, G: 0, B: 0, A: 255}},
		{"green", &Color{R: 0, G: 255, B: 0, A: 255}},
		{"blue", &Color{R: 0, G: 0, B: 255, A: 255}},
		{"unknown", nil},
	}

	for _, tt := range tests {
		result := parseColorName(tt.name)
		if tt.expected == nil {
			if result != nil {
				t.Errorf("parseColorName(%q) = %v, want nil", tt.name, result)
			}
		} else {
			if result == nil {
				t.Errorf("parseColorName(%q) = nil, want %v", tt.name, tt.expected)
			} else if *result != *tt.expected {
				t.Errorf("parseColorName(%q) = %v, want %v", tt.name, *result, *tt.expected)
			}
		}
	}
}
