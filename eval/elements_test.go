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
// Columns Tests
// ----------------------------------------------------------------------------

func TestColumnsFunc(t *testing.T) {
	// Get the columns function
	columnsFunc := ColumnsFunc()

	if columnsFunc == nil {
		t.Fatal("ColumnsFunc() returned nil")
	}

	if columnsFunc.Name == nil || *columnsFunc.Name != "columns" {
		t.Errorf("expected function name 'columns', got %v", columnsFunc.Name)
	}

	// Verify it's a native function
	_, ok := columnsFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestColumnsNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create body content
	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Column content"}},
	}}

	// Create args with just the body (default count: 2)
	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	// Verify result is ContentValue
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	// Verify it contains one ColumnsElement
	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	columns, ok := content.Content.Elements[0].(*ColumnsElement)
	if !ok {
		t.Fatalf("expected *ColumnsElement, got %T", content.Content.Elements[0])
	}

	// Verify element properties (defaults)
	if columns.Count != nil {
		t.Errorf("Count = %v, want nil (default 2)", columns.Count)
	}
	if columns.Gutter != nil {
		t.Errorf("Gutter = %v, want nil (default)", columns.Gutter)
	}
	if len(columns.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(columns.Body.Elements))
	}
}

func TestColumnsNativeWithCount(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Column content"}},
	}}

	// Create args with count and body
	args := NewArgs(syntax.Detached())
	args.Push(Int(3), syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	content := result.(ContentValue)
	columns := content.Content.Elements[0].(*ColumnsElement)

	if columns.Count == nil || *columns.Count != 3 {
		t.Errorf("Count = %v, want 3", columns.Count)
	}
}

func TestColumnsNativeWithNamedCount(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Column content"}},
	}}

	// Create args with named count and body
	args := NewArgs(syntax.Detached())
	args.PushNamed("count", Int(4), syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	content := result.(ContentValue)
	columns := content.Content.Elements[0].(*ColumnsElement)

	if columns.Count == nil || *columns.Count != 4 {
		t.Errorf("Count = %v, want 4", columns.Count)
	}
}

func TestColumnsNativeWithGutter(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Column content"}},
	}}

	// Create args with gutter and body
	args := NewArgs(syntax.Detached())
	args.PushNamed("gutter", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	content := result.(ContentValue)
	columns := content.Content.Elements[0].(*ColumnsElement)

	if columns.Gutter == nil || *columns.Gutter != 10 {
		t.Errorf("Gutter = %v, want 10", columns.Gutter)
	}
}

func TestColumnsNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Full column layout"}},
	}}

	// Create args with all parameters
	args := NewArgs(syntax.Detached())
	args.Push(Int(5), syntax.Detached())
	args.PushNamed("gutter", LengthValue{Length: Length{Points: 15}}, syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := columnsNative(vm, args)
	if err != nil {
		t.Fatalf("columnsNative() error: %v", err)
	}

	content := result.(ContentValue)
	columns := content.Content.Elements[0].(*ColumnsElement)

	if columns.Count == nil || *columns.Count != 5 {
		t.Errorf("Count = %v, want 5", columns.Count)
	}
	if columns.Gutter == nil || *columns.Gutter != 15 {
		t.Errorf("Gutter = %v, want 15", columns.Gutter)
	}
	if len(columns.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(columns.Body.Elements))
	}
}

func TestColumnsNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("gutter", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	_, err := columnsNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
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

func TestColumnsNativeInvalidCount(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Content"}},
	}}

	// Test with count < 1
	args := NewArgs(syntax.Detached())
	args.Push(Int(0), syntax.Detached())
	args.Push(body, syntax.Detached())

	_, err := columnsNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid count (0)")
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
	// Test ColumnsElement struct and ContentElement interface
	count := 3
	gutter := 12.0
	elem := &ColumnsElement{
		Count:  &count,
		Gutter: &gutter,
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Column text"}},
		},
	}

	if *elem.Count != 3 {
		t.Errorf("Count = %v, want 3", *elem.Count)
	}
	if *elem.Gutter != 12.0 {
		t.Errorf("Gutter = %v, want 12.0", *elem.Gutter)
	}
	if len(elem.Body.Elements) != 1 {
		t.Errorf("Body elements = %d, want 1", len(elem.Body.Elements))
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Columns
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesColumns(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify columns function is registered
	columnsBinding := scope.Get("columns")
	if columnsBinding == nil {
		t.Fatal("expected 'columns' to be registered")
	}

	columnsFunc, ok := columnsBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for columns, got %T", columnsBinding.Value)
	}
	if columnsFunc.Func.Name == nil || *columnsFunc.Func.Name != "columns" {
		t.Errorf("expected function name 'columns', got %v", columnsFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesColumns(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["columns"]; !ok {
		t.Error("expected 'columns' in ElementFunctions()")
	}
}

// ----------------------------------------------------------------------------
// Box Tests
// ----------------------------------------------------------------------------

func TestBoxFunc(t *testing.T) {
	boxFunc := BoxFunc()

	if boxFunc == nil {
		t.Fatal("BoxFunc() returned nil")
	}

	if boxFunc.Name == nil || *boxFunc.Name != "box" {
		t.Errorf("expected function name 'box', got %v", boxFunc.Name)
	}

	_, ok := boxFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestBoxNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Box content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := boxNative(vm, args)
	if err != nil {
		t.Fatalf("boxNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	box, ok := content.Content.Elements[0].(*BoxElement)
	if !ok {
		t.Fatalf("expected *BoxElement, got %T", content.Content.Elements[0])
	}

	if box.Width != nil {
		t.Errorf("Width = %v, want nil (default)", box.Width)
	}
	if box.Height != nil {
		t.Errorf("Height = %v, want nil (default)", box.Height)
	}
	if box.Clip != false {
		t.Errorf("Clip = %v, want false", box.Clip)
	}
}

func TestBoxNativeWithWidth(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Box content"}},
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

	if box.Width == nil || *box.Width != 100 {
		t.Errorf("Width = %v, want 100", box.Width)
	}
}

func TestBoxNativeWithClip(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Box content"}},
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

	if box.Clip != true {
		t.Errorf("Clip = %v, want true", box.Clip)
	}
}

func TestBoxNativeEmpty(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

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

func TestBoxElement(t *testing.T) {
	width := 50.0
	height := 30.0
	elem := &BoxElement{
		Width:  &width,
		Height: &height,
		Clip:   true,
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Content"}},
		},
	}

	if *elem.Width != 50.0 {
		t.Errorf("Width = %v, want 50.0", *elem.Width)
	}
	if *elem.Height != 30.0 {
		t.Errorf("Height = %v, want 30.0", *elem.Height)
	}
	if elem.Clip != true {
		t.Errorf("Clip = %v, want true", elem.Clip)
	}

	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Block Tests
// ----------------------------------------------------------------------------

func TestBlockFunc(t *testing.T) {
	blockFunc := BlockFunc()

	if blockFunc == nil {
		t.Fatal("BlockFunc() returned nil")
	}

	if blockFunc.Name == nil || *blockFunc.Name != "block" {
		t.Errorf("expected function name 'block', got %v", blockFunc.Name)
	}

	_, ok := blockFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestBlockNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Block content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	block, ok := content.Content.Elements[0].(*BlockElement)
	if !ok {
		t.Fatalf("expected *BlockElement, got %T", content.Content.Elements[0])
	}

	if block.Width != nil {
		t.Errorf("Width = %v, want nil (default)", block.Width)
	}
	if block.Breakable != nil {
		t.Errorf("Breakable = %v, want nil (default)", block.Breakable)
	}
}

func TestBlockNativeWithBreakable(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Block content"}},
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
		Elements: []ContentElement{&TextElement{Text: "Block content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("above", LengthValue{Length: Length{Points: 20}}, syntax.Detached())
	args.PushNamed("below", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	result, err := blockNative(vm, args)
	if err != nil {
		t.Fatalf("blockNative() error: %v", err)
	}

	content := result.(ContentValue)
	block := content.Content.Elements[0].(*BlockElement)

	if block.Above == nil || *block.Above != 20 {
		t.Errorf("Above = %v, want 20", block.Above)
	}
	if block.Below == nil || *block.Below != 10 {
		t.Errorf("Below = %v, want 10", block.Below)
	}
}

func TestBlockElement(t *testing.T) {
	width := 100.0
	breakable := false
	sticky := true
	elem := &BlockElement{
		Width:     &width,
		Breakable: &breakable,
		Sticky:    sticky,
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Content"}},
		},
	}

	if *elem.Width != 100.0 {
		t.Errorf("Width = %v, want 100.0", *elem.Width)
	}
	if *elem.Breakable != false {
		t.Errorf("Breakable = %v, want false", *elem.Breakable)
	}
	if elem.Sticky != true {
		t.Errorf("Sticky = %v, want true", elem.Sticky)
	}

	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Pad Tests
// ----------------------------------------------------------------------------

func TestPadFunc(t *testing.T) {
	padFunc := PadFunc()

	if padFunc == nil {
		t.Fatal("PadFunc() returned nil")
	}

	if padFunc.Name == nil || *padFunc.Name != "pad" {
		t.Errorf("expected function name 'pad', got %v", padFunc.Name)
	}

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

	if pad.Left != nil {
		t.Errorf("Left = %v, want nil (default)", pad.Left)
	}
	if pad.Top != nil {
		t.Errorf("Top = %v, want nil (default)", pad.Top)
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
	args.PushNamed("x", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	if pad.Left == nil || *pad.Left != 10 {
		t.Errorf("Left = %v, want 10", pad.Left)
	}
	if pad.Right == nil || *pad.Right != 10 {
		t.Errorf("Right = %v, want 10", pad.Right)
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
	args.PushNamed("y", LengthValue{Length: Length{Points: 20}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	if pad.Top == nil || *pad.Top != 20 {
		t.Errorf("Top = %v, want 20", pad.Top)
	}
	if pad.Bottom == nil || *pad.Bottom != 20 {
		t.Errorf("Bottom = %v, want 20", pad.Bottom)
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
	args.PushNamed("rest", LengthValue{Length: Length{Points: 5}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	if pad.Left == nil || *pad.Left != 5 {
		t.Errorf("Left = %v, want 5", pad.Left)
	}
	if pad.Top == nil || *pad.Top != 5 {
		t.Errorf("Top = %v, want 5", pad.Top)
	}
	if pad.Right == nil || *pad.Right != 5 {
		t.Errorf("Right = %v, want 5", pad.Right)
	}
	if pad.Bottom == nil || *pad.Bottom != 5 {
		t.Errorf("Bottom = %v, want 5", pad.Bottom)
	}
}

func TestPadNativeWithIndividualSides(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	body := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Padded content"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(body, syntax.Detached())
	args.PushNamed("left", LengthValue{Length: Length{Points: 1}}, syntax.Detached())
	args.PushNamed("top", LengthValue{Length: Length{Points: 2}}, syntax.Detached())
	args.PushNamed("right", LengthValue{Length: Length{Points: 3}}, syntax.Detached())
	args.PushNamed("bottom", LengthValue{Length: Length{Points: 4}}, syntax.Detached())

	result, err := padNative(vm, args)
	if err != nil {
		t.Fatalf("padNative() error: %v", err)
	}

	content := result.(ContentValue)
	pad := content.Content.Elements[0].(*PadElement)

	if pad.Left == nil || *pad.Left != 1 {
		t.Errorf("Left = %v, want 1", pad.Left)
	}
	if pad.Top == nil || *pad.Top != 2 {
		t.Errorf("Top = %v, want 2", pad.Top)
	}
	if pad.Right == nil || *pad.Right != 3 {
		t.Errorf("Right = %v, want 3", pad.Right)
	}
	if pad.Bottom == nil || *pad.Bottom != 4 {
		t.Errorf("Bottom = %v, want 4", pad.Bottom)
	}
}

func TestPadNativeMissingBody(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("x", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	_, err := padNative(vm, args)
	if err == nil {
		t.Error("expected error for missing body argument")
	}
}

func TestPadElement(t *testing.T) {
	left := 10.0
	top := 20.0
	right := 30.0
	bottom := 40.0
	elem := &PadElement{
		Left:   &left,
		Top:    &top,
		Right:  &right,
		Bottom: &bottom,
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "Content"}},
		},
	}

	if *elem.Left != 10.0 {
		t.Errorf("Left = %v, want 10.0", *elem.Left)
	}
	if *elem.Top != 20.0 {
		t.Errorf("Top = %v, want 20.0", *elem.Top)
	}
	if *elem.Right != 30.0 {
		t.Errorf("Right = %v, want 30.0", *elem.Right)
	}
	if *elem.Bottom != 40.0 {
		t.Errorf("Bottom = %v, want 40.0", *elem.Bottom)
	}

	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for Box, Block, and Pad
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesBoxBlockPad(t *testing.T) {
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
}

func TestElementFunctionsIncludesBoxBlockPad(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["box"]; !ok {
		t.Error("expected 'box' in ElementFunctions()")
	}

	if _, ok := funcs["block"]; !ok {
		t.Error("expected 'block' in ElementFunctions()")
	}

	if _, ok := funcs["pad"]; !ok {
		t.Error("expected 'pad' in ElementFunctions()")
	}
}

// ----------------------------------------------------------------------------
// Table Element Tests
// ----------------------------------------------------------------------------

func TestTableFunc(t *testing.T) {
	// Get the table function
	tableFunc := TableFunc()

	if tableFunc == nil {
		t.Fatal("TableFunc() returned nil")
	}

	if tableFunc.Name == nil || *tableFunc.Name != "table" {
		t.Error("TableFunc name should be 'table'")
	}
}

func TestTableNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create table cells as content
	cell1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Name"}},
	}}
	cell2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Value"}},
	}}
	cell3 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Alpha"}},
	}}
	cell4 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "1"}},
	}}

	// Create args with columns and cells
	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Int(2), syntax.Detached())
	args.Push(cell1, syntax.Detached())
	args.Push(cell2, syntax.Detached())
	args.Push(cell3, syntax.Detached())
	args.Push(cell4, syntax.Detached())

	// Call table native function
	tableFunc := TableFunc()
	result, err := tableFunc.Repr.(NativeFunc).Func(vm, args)

	if err != nil {
		t.Fatalf("tableNative failed: %v", err)
	}

	cv, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(cv.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(cv.Content.Elements))
	}

	tableElem, ok := cv.Content.Elements[0].(*TableElement)
	if !ok {
		t.Fatalf("expected TableElement, got %T", cv.Content.Elements[0])
	}

	if tableElem.Columns != 2 {
		t.Errorf("expected 2 columns, got %d", tableElem.Columns)
	}

	if len(tableElem.Cells) != 4 {
		t.Errorf("expected 4 cells, got %d", len(tableElem.Cells))
	}
}

func TestTableNativeMissingColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args without columns
	args := NewArgs(syntax.Detached())
	args.Push(ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "cell"}},
	}}, syntax.Detached())

	// Call table native function
	tableFunc := TableFunc()
	_, err := tableFunc.Repr.(NativeFunc).Func(vm, args)

	if err == nil {
		t.Fatal("expected error for missing columns argument")
	}

	if _, ok := err.(*MissingArgumentError); !ok {
		t.Errorf("expected MissingArgumentError, got %T", err)
	}
}

func TestTableNativeInvalidColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create args with invalid columns (0)
	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Int(0), syntax.Detached())

	// Call table native function
	tableFunc := TableFunc()
	_, err := tableFunc.Repr.(NativeFunc).Func(vm, args)

	if err == nil {
		t.Fatal("expected error for columns < 1")
	}

	if _, ok := err.(*InvalidArgumentError); !ok {
		t.Errorf("expected InvalidArgumentError, got %T", err)
	}
}

func TestTableElement(t *testing.T) {
	// Test TableElement struct and ContentElement interface
	elem := &TableElement{
		Columns: 3,
		Cells: []Content{
			{Elements: []ContentElement{&TextElement{Text: "A"}}},
			{Elements: []ContentElement{&TextElement{Text: "B"}}},
			{Elements: []ContentElement{&TextElement{Text: "C"}}},
		},
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem

	if elem.Columns != 3 {
		t.Errorf("expected 3 columns, got %d", elem.Columns)
	}

	if len(elem.Cells) != 3 {
		t.Errorf("expected 3 cells, got %d", len(elem.Cells))
	}
}

func TestRegisterElementFunctionsIncludesTable(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify table function is registered
	tableBinding := scope.Get("table")
	if tableBinding == nil {
		t.Error("table function not registered")
	}

	if tableBinding != nil {
		fv, ok := tableBinding.Value.(FuncValue)
		if !ok {
			t.Error("table binding should be a function")
		}
		if fv.Func.Name == nil || *fv.Func.Name != "table" {
			t.Error("table function name mismatch")
		}
	}
}

func TestElementFunctionsIncludesTable(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["table"]; !ok {
		t.Error("expected 'table' in ElementFunctions()")
	}
}
