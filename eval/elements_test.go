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

func TestBoxNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create body content
	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Hello, Box!"}},
	}}

	// Create args with just the body
	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	// Call the box function
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

	// Verify element properties (defaults)
	if len(box.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(box.Body.Elements))
	}
	if box.Width != nil {
		t.Errorf("Width = %v, want nil (default)", box.Width)
	}
	if box.Height != nil {
		t.Errorf("Height = %v, want nil (default)", box.Height)
	}
	if box.Fill != nil {
		t.Errorf("Fill = %v, want nil (default)", box.Fill)
	}
	if box.Stroke != nil {
		t.Errorf("Stroke = %v, want nil (default)", box.Stroke)
	}
	if box.Radius != nil {
		t.Errorf("Radius = %v, want nil (default)", box.Radius)
	}
	if box.Inset != nil {
		t.Errorf("Inset = %v, want nil (default)", box.Inset)
	}
	if box.Outset != nil {
		t.Errorf("Outset = %v, want nil (default)", box.Outset)
	}
	if box.Clip != nil {
		t.Errorf("Clip = %v, want nil (default)", box.Clip)
	}
}

func TestBoxNativeWithWidth(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Sized box"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("width", LengthValue{Length: Length{Points: 100}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Width == nil {
		t.Fatal("Width = nil, want LengthValue")
	}
	if lv, ok := box.Width.(LengthValue); !ok || lv.Length.Points != 100 {
		t.Errorf("Width = %v, want 100pt", box.Width)
	}
}

func TestBoxNativeWithFill(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Colored box"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("fill", ColorValue{Color: Color{R: 255, G: 0, B: 0, A: 255}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Fill == nil {
		t.Fatal("Fill = nil, want ColorValue")
	}
	if cv, ok := box.Fill.(ColorValue); !ok || cv.Color.R != 255 {
		t.Errorf("Fill = %v, want red color", box.Fill)
	}
}

func TestBoxNativeWithClip(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Clipped box"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("clip", True, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Clip == nil || *box.Clip != true {
		t.Errorf("Clip = %v, want true", box.Clip)
	}
}

func TestBoxNativeWithInset(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded box"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("inset", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if box.Inset == nil {
		t.Fatal("Inset = nil, want LengthValue")
	}
	if lv, ok := box.Inset.(LengthValue); !ok || lv.Length.Points != 10 {
		t.Errorf("Inset = %v, want 10pt", box.Inset)
	}
}

func TestBoxNativeEmpty(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Box with no body
	args := NewArgs(syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content := result.(ContentValue)
	box := content.Content.Elements[0].(*BoxElement)

	if len(box.Body.Elements) != 0 {
		t.Errorf("Body elements = %d, want 0", len(box.Body.Elements))
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

func TestBoxElement(t *testing.T) {
	// Test BoxElement struct and ContentElement interface
	clip := true
	baseline := 0.5

	elem := &BoxElement{
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Box content"}},
		},
		Width:    LengthValue{Length: Length{Points: 100}},
		Height:   LengthValue{Length: Length{Points: 50}},
		Baseline: &baseline,
		Fill:     ColorValue{Color: Color{R: 255, G: 0, B: 0, A: 255}},
		Clip:     &clip,
	}

	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}
	if elem.Width == nil {
		t.Error("Width = nil, want value")
	}
	if elem.Height == nil {
		t.Error("Height = nil, want value")
	}
	if elem.Baseline == nil || *elem.Baseline != 0.5 {
		t.Errorf("Baseline = %v, want 0.5", elem.Baseline)
	}
	if elem.Clip == nil || *elem.Clip != true {
		t.Errorf("Clip = %v, want true", elem.Clip)
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

func TestBlockNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create body content
	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Hello, Block!"}},
	}}

	// Create args with just the body
	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	// Call the block function
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

	// Verify element properties (defaults)
	if len(block.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(block.Body.Elements))
	}
	if block.Width != nil {
		t.Errorf("Width = %v, want nil (default)", block.Width)
	}
	if block.Height != nil {
		t.Errorf("Height = %v, want nil (default)", block.Height)
	}
	if block.Breakable != nil {
		t.Errorf("Breakable = %v, want nil (default)", block.Breakable)
	}
	if block.Fill != nil {
		t.Errorf("Fill = %v, want nil (default)", block.Fill)
	}
	if block.Stroke != nil {
		t.Errorf("Stroke = %v, want nil (default)", block.Stroke)
	}
	if block.Above != nil {
		t.Errorf("Above = %v, want nil (default)", block.Above)
	}
	if block.Below != nil {
		t.Errorf("Below = %v, want nil (default)", block.Below)
	}
	if block.Sticky != nil {
		t.Errorf("Sticky = %v, want nil (default)", block.Sticky)
	}
}

func TestBlockNativeWithBreakable(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Unbreakable block"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
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

func TestBlockNativeWithSpacing(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Spaced block"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("above", LengthValue{Length: Length{Points: 20}}, syntax.Detached())
	args.PushNamed("below", LengthValue{Length: Length{Points: 15}}, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if block.Above == nil {
		t.Fatal("Above = nil, want LengthValue")
	}
	if lv, ok := block.Above.(LengthValue); !ok || lv.Length.Points != 20 {
		t.Errorf("Above = %v, want 20pt", block.Above)
	}

	if block.Below == nil {
		t.Fatal("Below = nil, want LengthValue")
	}
	if lv, ok := block.Below.(LengthValue); !ok || lv.Length.Points != 15 {
		t.Errorf("Below = %v, want 15pt", block.Below)
	}
}

func TestBlockNativeWithSticky(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Sticky block"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("sticky", True, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if block.Sticky == nil || *block.Sticky != true {
		t.Errorf("Sticky = %v, want true", block.Sticky)
	}
}

func TestBlockNativeWithFillAndInset(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Styled block"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("fill", ColorValue{Color: Color{R: 200, G: 200, B: 200, A: 255}}, syntax.Detached())
	args.PushNamed("inset", LengthValue{Length: Length{Points: 8}}, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if block.Fill == nil {
		t.Fatal("Fill = nil, want ColorValue")
	}
	if cv, ok := block.Fill.(ColorValue); !ok || cv.Color.R != 200 {
		t.Errorf("Fill = %v, want gray color", block.Fill)
	}

	if block.Inset == nil {
		t.Fatal("Inset = nil, want LengthValue")
	}
	if lv, ok := block.Inset.(LengthValue); !ok || lv.Length.Points != 8 {
		t.Errorf("Inset = %v, want 8pt", block.Inset)
	}
}

func TestBlockNativeEmpty(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Block with no body
	args := NewArgs(syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if len(block.Body.Elements) != 0 {
		t.Errorf("Body elements = %d, want 0", len(block.Body.Elements))
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

func TestBlockElement(t *testing.T) {
	// Test BlockElement struct and ContentElement interface
	breakable := false
	clip := true
	sticky := true

	elem := &BlockElement{
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Block content"}},
		},
		Width:     LengthValue{Length: Length{Points: 200}},
		Height:    LengthValue{Length: Length{Points: 100}},
		Breakable: &breakable,
		Fill:      ColorValue{Color: Color{R: 0, G: 0, B: 255, A: 255}},
		Above:     LengthValue{Length: Length{Points: 20}},
		Below:     LengthValue{Length: Length{Points: 15}},
		Sticky:    &sticky,
		Clip:      &clip,
	}

	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}
	if elem.Width == nil {
		t.Error("Width = nil, want value")
	}
	if elem.Height == nil {
		t.Error("Height = nil, want value")
	}
	if elem.Breakable == nil || *elem.Breakable != false {
		t.Errorf("Breakable = %v, want false", elem.Breakable)
	}
	if elem.Sticky == nil || *elem.Sticky != true {
		t.Errorf("Sticky = %v, want true", elem.Sticky)
	}
	if elem.Clip == nil || *elem.Clip != true {
		t.Errorf("Clip = %v, want true", elem.Clip)
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

	blockFn, ok := blockBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for block, got %T", blockBinding.Value)
	}
	if blockFn.Func.Name == nil || *blockFn.Func.Name != "block" {
		t.Errorf("expected function name 'block', got %v", blockFn.Func.Name)
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
