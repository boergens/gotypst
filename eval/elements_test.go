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
// Pad Tests
// ----------------------------------------------------------------------------

func TestPadFunc(t *testing.T) {
	// Get the pad function
	padFunc := PadFunc()

	if padFunc == nil {
		t.Fatal("PadFunc() returned nil")
	}

	if padFunc.Name == nil || *padFunc.Name != "pad" {
		t.Errorf("expected function name 'pad', got %v", padFunc.Name)
	}

	// Verify it's a native function
	_, ok := padFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestPadNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	pad, ok := content.Content.Elements[0].(*PadElement)
	if !ok {
		t.Fatalf("expected *PadElement, got %T", content.Content.Elements[0])
	}

	// Verify zero padding defaults
	if pad.Left.Abs.Points != 0 || pad.Left.Rel.Value != 0 {
		t.Errorf("Left = %+v, want zero", pad.Left)
	}
	if pad.Top.Abs.Points != 0 || pad.Top.Rel.Value != 0 {
		t.Errorf("Top = %+v, want zero", pad.Top)
	}
	if pad.Right.Abs.Points != 0 || pad.Right.Rel.Value != 0 {
		t.Errorf("Right = %+v, want zero", pad.Right)
	}
	if pad.Bottom.Abs.Points != 0 || pad.Bottom.Rel.Value != 0 {
		t.Errorf("Bottom = %+v, want zero", pad.Bottom)
	}
}

func TestPadNativeWithLength(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("left", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.PushNamed("top", LengthValue{Length: Length{Points: 20}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	if pad.Left.Abs.Points != 10 {
		t.Errorf("Left.Abs.Points = %v, want 10", pad.Left.Abs.Points)
	}
	if pad.Top.Abs.Points != 20 {
		t.Errorf("Top.Abs.Points = %v, want 20", pad.Top.Abs.Points)
	}
}

func TestPadNativeWithRatio(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("right", RatioValue{Ratio: Ratio{Value: 0.1}}, syntax.Detached()) // 10%

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	if pad.Right.Rel.Value != 0.1 {
		t.Errorf("Right.Rel.Value = %v, want 0.1", pad.Right.Rel.Value)
	}
}

func TestPadNativeWithRelative(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	// 10pt + 5%
	args.PushNamed("bottom", RelativeValue{Relative: Relative{
		Abs: Length{Points: 10},
		Rel: Ratio{Value: 0.05},
	}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	if pad.Bottom.Abs.Points != 10 {
		t.Errorf("Bottom.Abs.Points = %v, want 10", pad.Bottom.Abs.Points)
	}
	if pad.Bottom.Rel.Value != 0.05 {
		t.Errorf("Bottom.Rel.Value = %v, want 0.05", pad.Bottom.Rel.Value)
	}
}

func TestPadNativeWithRest(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("rest", LengthValue{Length: Length{Points: 15}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	// rest should set all four sides
	if pad.Left.Abs.Points != 15 {
		t.Errorf("Left.Abs.Points = %v, want 15", pad.Left.Abs.Points)
	}
	if pad.Top.Abs.Points != 15 {
		t.Errorf("Top.Abs.Points = %v, want 15", pad.Top.Abs.Points)
	}
	if pad.Right.Abs.Points != 15 {
		t.Errorf("Right.Abs.Points = %v, want 15", pad.Right.Abs.Points)
	}
	if pad.Bottom.Abs.Points != 15 {
		t.Errorf("Bottom.Abs.Points = %v, want 15", pad.Bottom.Abs.Points)
	}
}

func TestPadNativeWithX(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("x", LengthValue{Length: Length{Points: 8}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	// x should set left and right
	if pad.Left.Abs.Points != 8 {
		t.Errorf("Left.Abs.Points = %v, want 8", pad.Left.Abs.Points)
	}
	if pad.Right.Abs.Points != 8 {
		t.Errorf("Right.Abs.Points = %v, want 8", pad.Right.Abs.Points)
	}
	// top and bottom should be zero
	if pad.Top.Abs.Points != 0 {
		t.Errorf("Top.Abs.Points = %v, want 0", pad.Top.Abs.Points)
	}
	if pad.Bottom.Abs.Points != 0 {
		t.Errorf("Bottom.Abs.Points = %v, want 0", pad.Bottom.Abs.Points)
	}
}

func TestPadNativeWithY(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("y", LengthValue{Length: Length{Points: 12}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	// y should set top and bottom
	if pad.Top.Abs.Points != 12 {
		t.Errorf("Top.Abs.Points = %v, want 12", pad.Top.Abs.Points)
	}
	if pad.Bottom.Abs.Points != 12 {
		t.Errorf("Bottom.Abs.Points = %v, want 12", pad.Bottom.Abs.Points)
	}
	// left and right should be zero
	if pad.Left.Abs.Points != 0 {
		t.Errorf("Left.Abs.Points = %v, want 0", pad.Left.Abs.Points)
	}
	if pad.Right.Abs.Points != 0 {
		t.Errorf("Right.Abs.Points = %v, want 0", pad.Right.Abs.Points)
	}
}

func TestPadNativeOverrideShorthand(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("rest", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.PushNamed("left", LengthValue{Length: Length{Points: 20}}, syntax.Detached()) // Override left

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	// left should be overridden
	if pad.Left.Abs.Points != 20 {
		t.Errorf("Left.Abs.Points = %v, want 20", pad.Left.Abs.Points)
	}
	// other sides should be from rest
	if pad.Top.Abs.Points != 10 {
		t.Errorf("Top.Abs.Points = %v, want 10", pad.Top.Abs.Points)
	}
	if pad.Right.Abs.Points != 10 {
		t.Errorf("Right.Abs.Points = %v, want 10", pad.Right.Abs.Points)
	}
	if pad.Bottom.Abs.Points != 10 {
		t.Errorf("Bottom.Abs.Points = %v, want 10", pad.Bottom.Abs.Points)
	}
}

func TestPadNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("left", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	_, err := padNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
	}
}

func TestPadNativeWrongBodyType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("not content"), syntax.Detached())

	_, err := padNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong body type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestPadNativeWrongPaddingType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("left", Str("not a length"), syntax.Detached())

	_, err := padNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong padding type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestPadElement(t *testing.T) {
	elem := &PadElement{
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Content"}},
		},
		Left:   Relative{Abs: Length{Points: 10}},
		Top:    Relative{Abs: Length{Points: 20}},
		Right:  Relative{Rel: Ratio{Value: 0.1}},
		Bottom: Relative{Abs: Length{Points: 5}, Rel: Ratio{Value: 0.05}},
	}

	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}
	if elem.Left.Abs.Points != 10 {
		t.Errorf("Left.Abs.Points = %v, want 10", elem.Left.Abs.Points)
	}
	if elem.Top.Abs.Points != 20 {
		t.Errorf("Top.Abs.Points = %v, want 20", elem.Top.Abs.Points)
	}
	if elem.Right.Rel.Value != 0.1 {
		t.Errorf("Right.Rel.Value = %v, want 0.1", elem.Right.Rel.Value)
	}
	if elem.Bottom.Abs.Points != 5 || elem.Bottom.Rel.Value != 0.05 {
		t.Errorf("Bottom = %+v, want Abs=5, Rel=0.05", elem.Bottom)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Place Tests
// ----------------------------------------------------------------------------

func TestPlaceFunc(t *testing.T) {
	placeFunc := PlaceFunc()

	if placeFunc == nil {
		t.Fatal("PlaceFunc() returned nil")
	}

	if placeFunc.Name == nil || *placeFunc.Name != "place" {
		t.Errorf("expected function name 'place', got %v", placeFunc.Name)
	}

	_, ok := placeFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestPlaceNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Placed content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := placeNative(vm, args)
	if err != nil {
		t.Fatalf("placeNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	place, ok := content.Content.Elements[0].(*PlaceElement)
	if !ok {
		t.Fatalf("expected *PlaceElement, got %T", content.Content.Elements[0])
	}

	// Verify defaults
	if place.Scope != "column" {
		t.Errorf("Scope = %q, want 'column'", place.Scope)
	}
	if place.Float != false {
		t.Errorf("Float = %v, want false", place.Float)
	}
	if place.Clearance != 18.0 {
		t.Errorf("Clearance = %v, want 18.0", place.Clearance)
	}
}

func TestPlaceNativeWithScope(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Placed content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("scope", Str("parent"), syntax.Detached())

	result, err := placeNative(vm, args)
	if err != nil {
		t.Fatalf("placeNative() error: %v", err)
	}

	content := result.(ContentValue)
	place := content.Content.Elements[0].(*PlaceElement)

	if place.Scope != "parent" {
		t.Errorf("Scope = %q, want 'parent'", place.Scope)
	}
}

func TestPlaceNativeWithInvalidScope(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Placed content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("scope", Str("invalid"), syntax.Detached())

	_, err := placeNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid scope")
	}
}

func TestPlaceNativeWithFloat(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Floating content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("float", True, syntax.Detached())

	result, err := placeNative(vm, args)
	if err != nil {
		t.Fatalf("placeNative() error: %v", err)
	}

	content := result.(ContentValue)
	place := content.Content.Elements[0].(*PlaceElement)

	if place.Float != true {
		t.Errorf("Float = %v, want true", place.Float)
	}
}

func TestPlaceNativeWithClearance(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Floating content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("clearance", LengthValue{Length: Length{Points: 24}}, syntax.Detached())

	result, err := placeNative(vm, args)
	if err != nil {
		t.Fatalf("placeNative() error: %v", err)
	}

	content := result.(ContentValue)
	place := content.Content.Elements[0].(*PlaceElement)

	if place.Clearance != 24 {
		t.Errorf("Clearance = %v, want 24", place.Clearance)
	}
}

func TestPlaceNativeWithOffset(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Offset content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("dx", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.PushNamed("dy", RatioValue{Ratio: Ratio{Value: 0.2}}, syntax.Detached())

	result, err := placeNative(vm, args)
	if err != nil {
		t.Fatalf("placeNative() error: %v", err)
	}

	content := result.(ContentValue)
	place := content.Content.Elements[0].(*PlaceElement)

	if place.Dx.Abs.Points != 10 {
		t.Errorf("Dx.Abs.Points = %v, want 10", place.Dx.Abs.Points)
	}
	if place.Dy.Rel.Value != 0.2 {
		t.Errorf("Dy.Rel.Value = %v, want 0.2", place.Dy.Rel.Value)
	}
}

func TestPlaceNativeWithAlignment(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Aligned content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("alignment", Str("center"), syntax.Detached())

	result, err := placeNative(vm, args)
	if err != nil {
		t.Fatalf("placeNative() error: %v", err)
	}

	content := result.(ContentValue)
	place := content.Content.Elements[0].(*PlaceElement)

	if place.AlignmentX != "center" {
		t.Errorf("AlignmentX = %q, want 'center'", place.AlignmentX)
	}
}

func TestPlaceNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("scope", Str("column"), syntax.Detached())

	_, err := placeNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
	}
}

func TestPlaceNativeWrongBodyType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("not content"), syntax.Detached())

	_, err := placeNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong body type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestPlaceElement(t *testing.T) {
	elem := &PlaceElement{
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Content"}},
		},
		AlignmentX: "center",
		AlignmentY: "top",
		Scope:      "parent",
		Float:      true,
		Clearance:  24.0,
		Dx:         Relative{Abs: Length{Points: 10}},
		Dy:         Relative{Rel: Ratio{Value: 0.1}},
	}

	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}
	if elem.AlignmentX != "center" {
		t.Errorf("AlignmentX = %q, want 'center'", elem.AlignmentX)
	}
	if elem.AlignmentY != "top" {
		t.Errorf("AlignmentY = %q, want 'top'", elem.AlignmentY)
	}
	if elem.Scope != "parent" {
		t.Errorf("Scope = %q, want 'parent'", elem.Scope)
	}
	if elem.Float != true {
		t.Errorf("Float = %v, want true", elem.Float)
	}
	if elem.Clearance != 24.0 {
		t.Errorf("Clearance = %v, want 24.0", elem.Clearance)
	}
	if elem.Dx.Abs.Points != 10 {
		t.Errorf("Dx.Abs.Points = %v, want 10", elem.Dx.Abs.Points)
	}
	if elem.Dy.Rel.Value != 0.1 {
		t.Errorf("Dy.Rel.Value = %v, want 0.1", elem.Dy.Rel.Value)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// PlaceFlush Tests
// ----------------------------------------------------------------------------

func TestPlaceFlushFunc(t *testing.T) {
	flushFunc := PlaceFlushFunc()

	if flushFunc == nil {
		t.Fatal("PlaceFlushFunc() returned nil")
	}

	if flushFunc.Name == nil || *flushFunc.Name != "flush" {
		t.Errorf("expected function name 'flush', got %v", flushFunc.Name)
	}

	_, ok := flushFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestPlaceFlushNative(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())

	result, err := placeFlushNative(vm, args)
	if err != nil {
		t.Fatalf("placeFlushNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	_, ok = content.Content.Elements[0].(*PlaceFlushElement)
	if !ok {
		t.Fatalf("expected *PlaceFlushElement, got %T", content.Content.Elements[0])
	}
}

func TestPlaceFlushNativeUnexpectedArg(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("unexpected", Str("value"), syntax.Detached())

	_, err := placeFlushNative(vm, args)
	if err == nil {
		t.Error("expected error for unexpected argument")
	}
	if _, ok := err.(*UnexpectedArgumentError); !ok {
		t.Errorf("expected UnexpectedArgumentError, got %T", err)
	}
}

func TestPlaceFlushElement(t *testing.T) {
	elem := &PlaceFlushElement{}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Pad and Place
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesPadAndPlace(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify pad function is registered
	padBinding := scope.Get("pad")
	if padBinding == nil {
		t.Fatal("expected 'pad' to be registered")
	}

	padFunc, ok := padBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for pad, got %T", padBinding.Value)
	}
	if padFunc.Func.Name == nil || *padFunc.Func.Name != "pad" {
		t.Errorf("expected function name 'pad', got %v", padFunc.Func.Name)
	}

	// Verify place function is registered
	placeBinding := scope.Get("place")
	if placeBinding == nil {
		t.Fatal("expected 'place' to be registered")
	}

	placeFunc, ok := placeBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for place, got %T", placeBinding.Value)
	}
	if placeFunc.Func.Name == nil || *placeFunc.Func.Name != "place" {
		t.Errorf("expected function name 'place', got %v", placeFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesPadAndPlace(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["pad"]; !ok {
		t.Error("expected 'pad' in ElementFunctions()")
	}

	if _, ok := funcs["place"]; !ok {
		t.Error("expected 'place' in ElementFunctions()")
	}
}
