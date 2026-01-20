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
// Table Tests
// ----------------------------------------------------------------------------

func TestTableFunc(t *testing.T) {
	// Get the table function
	tableFunc := TableFunc()

	if tableFunc == nil {
		t.Fatal("TableFunc() returned nil")
	}

	if tableFunc.Name == nil || *tableFunc.Name != "table" {
		t.Errorf("expected function name 'table', got %v", tableFunc.Name)
	}

	// Verify it's a native function
	_, ok := tableFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestTableNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create table cells as content
	cell1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell 1"}},
	}}
	cell2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell 2"}},
	}}
	cell3 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell 3"}},
	}}
	cell4 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell 4"}},
	}}

	// Create args with children
	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Int(2), syntax.Detached())
	args.Push(cell1, syntax.Detached())
	args.Push(cell2, syntax.Detached())
	args.Push(cell3, syntax.Detached())
	args.Push(cell4, syntax.Detached())

	result, err := tableNative(vm, args)
	if err != nil {
		t.Fatalf("tableNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one TableElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	table, ok := content.Content.Elements[0].(*TableElement)
	if !ok {
		t.Fatalf("expected *TableElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties
	if len(table.Columns) != 2 {
		t.Errorf("Columns length = %d, want 2", len(table.Columns))
	}
	if len(table.Cells) != 4 {
		t.Errorf("Cells length = %d, want 4", len(table.Cells))
	}
	if table.Align != nil {
		t.Errorf("Align = %v, want nil (default)", table.Align)
	}
}

func TestTableNativeWithAlign(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	cell := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Int(1), syntax.Detached())
	args.PushNamed("align", Str("center"), syntax.Detached())
	args.Push(cell, syntax.Detached())

	result, err := tableNative(vm, args)
	if err != nil {
		t.Fatalf("tableNative() error: %v", err)
	}

	content := result.(ContentValue)
	table := content.Content.Elements[0].(*TableElement)

	if table.Align == nil {
		t.Fatal("Align = nil, want non-nil")
	}
	if table.Align.Horizontal == nil || *table.Align.Horizontal != "center" {
		t.Errorf("Align.Horizontal = %v, want 'center'", table.Align.Horizontal)
	}
}

func TestTableNativeWithGutter(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	cell := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Int(2), syntax.Detached())
	args.PushNamed("gutter", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.Push(cell, syntax.Detached())

	result, err := tableNative(vm, args)
	if err != nil {
		t.Fatalf("tableNative() error: %v", err)
	}

	content := result.(ContentValue)
	table := content.Content.Elements[0].(*TableElement)

	// gutter should set both column-gutter and row-gutter
	if table.ColumnGutter == nil || *table.ColumnGutter != 10 {
		t.Errorf("ColumnGutter = %v, want 10", table.ColumnGutter)
	}
	if table.RowGutter == nil || *table.RowGutter != 10 {
		t.Errorf("RowGutter = %v, want 10", table.RowGutter)
	}
}

func TestTableNativeWithInset(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	cell := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Int(1), syntax.Detached())
	args.PushNamed("inset", LengthValue{Length: Length{Points: 5}}, syntax.Detached())
	args.Push(cell, syntax.Detached())

	result, err := tableNative(vm, args)
	if err != nil {
		t.Fatalf("tableNative() error: %v", err)
	}

	content := result.(ContentValue)
	table := content.Content.Elements[0].(*TableElement)

	if table.Inset == nil || *table.Inset != 5 {
		t.Errorf("Inset = %v, want 5", table.Inset)
	}
}

func TestTableNativeWithArrayColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	cell := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	// Create array of column sizes: [auto, 100pt]
	cols := ArrayValue{
		Auto,
		LengthValue{Length: Length{Points: 100}},
	}

	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", cols, syntax.Detached())
	args.Push(cell, syntax.Detached())

	result, err := tableNative(vm, args)
	if err != nil {
		t.Fatalf("tableNative() error: %v", err)
	}

	content := result.(ContentValue)
	table := content.Content.Elements[0].(*TableElement)

	if len(table.Columns) != 2 {
		t.Fatalf("Columns length = %d, want 2", len(table.Columns))
	}
	if !table.Columns[0].Auto {
		t.Error("Columns[0].Auto = false, want true")
	}
	if table.Columns[1].Points != 100 {
		t.Errorf("Columns[1].Points = %v, want 100", table.Columns[1].Points)
	}
}

func TestTableNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := tableNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestTableElement(t *testing.T) {
	// Test TableElement struct and ContentElement interface
	inset := 5.0
	elem := &TableElement{
		Columns: []TableSizing{
			{Auto: true},
			{Points: 100},
		},
		Cells: []*TableCellElement{
			{Content: Content{Elements: []ContentElement{&TextElement{Text: "A"}}}, X: -1, Y: -1},
			{Content: Content{Elements: []ContentElement{&TextElement{Text: "B"}}}, X: -1, Y: -1},
		},
		Inset: &inset,
	}

	if len(elem.Columns) != 2 {
		t.Errorf("Columns length = %d, want 2", len(elem.Columns))
	}
	if len(elem.Cells) != 2 {
		t.Errorf("Cells length = %d, want 2", len(elem.Cells))
	}
	if *elem.Inset != 5.0 {
		t.Errorf("Inset = %v, want 5.0", *elem.Inset)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Table Cell Tests
// ----------------------------------------------------------------------------

func TestTableCellFunc(t *testing.T) {
	// Get the table.cell function
	cellFunc := TableCellFunc()

	if cellFunc == nil {
		t.Fatal("TableCellFunc() returned nil")
	}

	if cellFunc.Name == nil || *cellFunc.Name != "table.cell" {
		t.Errorf("expected function name 'table.cell', got %v", cellFunc.Name)
	}

	// Verify it's a native function
	_, ok := cellFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestTableCellNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := tableCellNative(vm, args)
	if err != nil {
		t.Fatalf("tableCellNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one TableCellElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	cell, ok := content.Content.Elements[0].(*TableCellElement)
	if !ok {
		t.Fatalf("expected *TableCellElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties (defaults)
	if len(cell.Content.Elements) != 1 {
		t.Errorf("Content elements = %d, want 1", len(cell.Content.Elements))
	}
	if cell.X != -1 {
		t.Errorf("X = %d, want -1 (auto)", cell.X)
	}
	if cell.Y != -1 {
		t.Errorf("Y = %d, want -1 (auto)", cell.Y)
	}
	if cell.Colspan != 1 {
		t.Errorf("Colspan = %d, want 1", cell.Colspan)
	}
	if cell.Rowspan != 1 {
		t.Errorf("Rowspan = %d, want 1", cell.Rowspan)
	}
}

func TestTableCellNativeWithPosition(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("x", Int(2), syntax.Detached())
	args.PushNamed("y", Int(3), syntax.Detached())

	result, err := tableCellNative(vm, args)
	if err != nil {
		t.Fatalf("tableCellNative() error: %v", err)
	}

	content := result.(ContentValue)
	cell := content.Content.Elements[0].(*TableCellElement)

	if cell.X != 2 {
		t.Errorf("X = %d, want 2", cell.X)
	}
	if cell.Y != 3 {
		t.Errorf("Y = %d, want 3", cell.Y)
	}
}

func TestTableCellNativeWithSpan(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("colspan", Int(2), syntax.Detached())
	args.PushNamed("rowspan", Int(3), syntax.Detached())

	result, err := tableCellNative(vm, args)
	if err != nil {
		t.Fatalf("tableCellNative() error: %v", err)
	}

	content := result.(ContentValue)
	cell := content.Content.Elements[0].(*TableCellElement)

	if cell.Colspan != 2 {
		t.Errorf("Colspan = %d, want 2", cell.Colspan)
	}
	if cell.Rowspan != 3 {
		t.Errorf("Rowspan = %d, want 3", cell.Rowspan)
	}
}

func TestTableCellNativeWithAlign(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("align", Str("right"), syntax.Detached())

	result, err := tableCellNative(vm, args)
	if err != nil {
		t.Fatalf("tableCellNative() error: %v", err)
	}

	content := result.(ContentValue)
	cell := content.Content.Elements[0].(*TableCellElement)

	if cell.Align == nil {
		t.Fatal("Align = nil, want non-nil")
	}
	if cell.Align.Horizontal == nil || *cell.Align.Horizontal != "right" {
		t.Errorf("Align.Horizontal = %v, want 'right'", cell.Align.Horizontal)
	}
}

func TestTableCellNativeInvalidColspan(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("colspan", Int(0), syntax.Detached())

	_, err := tableCellNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid colspan (0)")
	}
}

func TestTableCellNativeInvalidRowspan(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("rowspan", Int(-1), syntax.Detached())

	_, err := tableCellNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid rowspan (-1)")
	}
}

func TestTableCellNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Cell"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("unknown", Str("value"), syntax.Detached())

	_, err := tableCellNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestTableCellElement(t *testing.T) {
	// Test TableCellElement struct and ContentElement interface
	inset := 8.0
	elem := &TableCellElement{
		Content: Content{Elements: []ContentElement{&TextElement{Text: "Cell"}}},
		X:       1,
		Y:       2,
		Colspan: 2,
		Rowspan: 3,
		Inset:   &inset,
	}

	if len(elem.Content.Elements) != 1 {
		t.Errorf("Content elements = %d, want 1", len(elem.Content.Elements))
	}
	if elem.X != 1 {
		t.Errorf("X = %d, want 1", elem.X)
	}
	if elem.Y != 2 {
		t.Errorf("Y = %d, want 2", elem.Y)
	}
	if elem.Colspan != 2 {
		t.Errorf("Colspan = %d, want 2", elem.Colspan)
	}
	if elem.Rowspan != 3 {
		t.Errorf("Rowspan = %d, want 3", elem.Rowspan)
	}
	if *elem.Inset != 8.0 {
		t.Errorf("Inset = %v, want 8.0", *elem.Inset)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Table
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesTable(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify table function is registered
	tableBinding := scope.Get("table")
	if tableBinding == nil {
		t.Fatal("expected 'table' to be registered")
	}

	tableFunc, ok := tableBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for table, got %T", tableBinding.Value)
	}
	if tableFunc.Func.Name == nil || *tableFunc.Func.Name != "table" {
		t.Errorf("expected function name 'table', got %v", tableFunc.Func.Name)
	}

	// Verify table.cell function is registered
	cellBinding := scope.Get("table.cell")
	if cellBinding == nil {
		t.Fatal("expected 'table.cell' to be registered")
	}

	cellFunc, ok := cellBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for table.cell, got %T", cellBinding.Value)
	}
	if cellFunc.Func.Name == nil || *cellFunc.Func.Name != "table.cell" {
		t.Errorf("expected function name 'table.cell', got %v", cellFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesTable(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["table"]; !ok {
		t.Error("expected 'table' in ElementFunctions()")
	}

	if _, ok := funcs["table.cell"]; !ok {
		t.Error("expected 'table.cell' in ElementFunctions()")
	}
}
