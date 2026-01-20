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

	// Create content children
	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "First item"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Second item"}},
	}}

	// Create args with children
	args := NewArgs(syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

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

	// Verify element properties (defaults)
	if list.Tight != true {
		t.Errorf("Tight = %v, want true", list.Tight)
	}
	if list.Marker != nil {
		t.Errorf("Marker = %v, want nil", list.Marker)
	}
	if list.Indent != nil {
		t.Errorf("Indent = %v, want nil", list.Indent)
	}
	if list.BodyIndent != nil {
		t.Errorf("BodyIndent = %v, want nil", list.BodyIndent)
	}
	if list.Spacing != nil {
		t.Errorf("Spacing = %v, want nil", list.Spacing)
	}
	if len(list.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(list.Items))
	}
}

func TestListNativeWithTight(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("tight", False, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Tight != false {
		t.Errorf("Tight = %v, want false", list.Tight)
	}
}

func TestListNativeWithMarker(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	marker := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "→"}},
	}}
	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("marker", marker, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Marker == nil {
		t.Error("Marker = nil, want non-nil")
	} else if len(list.Marker.Elements) != 1 {
		t.Errorf("Marker elements = %d, want 1", len(list.Marker.Elements))
	}
}

func TestListNativeWithIndent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("indent", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Indent == nil || *list.Indent != 10 {
		t.Errorf("Indent = %v, want 10", list.Indent)
	}
}

func TestListNativeWithBodyIndent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("body-indent", LengthValue{Length: Length{Points: 8}}, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.BodyIndent == nil || *list.BodyIndent != 8 {
		t.Errorf("BodyIndent = %v, want 8", list.BodyIndent)
	}
}

func TestListNativeWithSpacing(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("spacing", LengthValue{Length: Length{Points: 12}}, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Spacing == nil || *list.Spacing != 12 {
		t.Errorf("Spacing = %v, want 12", list.Spacing)
	}
}

func TestListNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	marker := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "•"}},
	}}
	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "First"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Second"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("tight", False, syntax.Detached())
	args.PushNamed("marker", marker, syntax.Detached())
	args.PushNamed("indent", LengthValue{Length: Length{Points: 5}}, syntax.Detached())
	args.PushNamed("body-indent", LengthValue{Length: Length{Points: 7}}, syntax.Detached())
	args.PushNamed("spacing", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Tight != false {
		t.Errorf("Tight = %v, want false", list.Tight)
	}
	if list.Marker == nil {
		t.Error("Marker = nil, want non-nil")
	}
	if list.Indent == nil || *list.Indent != 5 {
		t.Errorf("Indent = %v, want 5", list.Indent)
	}
	if list.BodyIndent == nil || *list.BodyIndent != 7 {
		t.Errorf("BodyIndent = %v, want 7", list.BodyIndent)
	}
	if list.Spacing == nil || *list.Spacing != 10 {
		t.Errorf("Spacing = %v, want 10", list.Spacing)
	}
	if len(list.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(list.Items))
	}
}

func TestListNativeWrongChildType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("not content"), syntax.Detached())

	_, err := listNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong child type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
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
	indent := 10.0
	bodyIndent := 8.0
	spacing := 12.0
	marker := Content{Elements: []ContentElement{&TextElement{Text: "•"}}}

	elem := &ListElement{
		Items: []*ListItemElement{
			{Content: Content{Elements: []ContentElement{&TextElement{Text: "A"}}}},
			{Content: Content{Elements: []ContentElement{&TextElement{Text: "B"}}}},
		},
		Tight:      true,
		Marker:     &marker,
		Indent:     &indent,
		BodyIndent: &bodyIndent,
		Spacing:    &spacing,
	}

	if len(elem.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(elem.Items))
	}
	if elem.Tight != true {
		t.Errorf("Tight = %v, want true", elem.Tight)
	}
	if *elem.Indent != 10.0 {
		t.Errorf("Indent = %v, want 10.0", *elem.Indent)
	}
	if *elem.BodyIndent != 8.0 {
		t.Errorf("BodyIndent = %v, want 8.0", *elem.BodyIndent)
	}
	if *elem.Spacing != 12.0 {
		t.Errorf("Spacing = %v, want 12.0", *elem.Spacing)
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

	// Create content children
	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "First item"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Second item"}},
	}}

	// Create args with children
	args := NewArgs(syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

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

	// Verify element properties (defaults)
	if enum.Tight != true {
		t.Errorf("Tight = %v, want true", enum.Tight)
	}
	if enum.Numbering != "1." {
		t.Errorf("Numbering = %q, want %q", enum.Numbering, "1.")
	}
	if enum.Start != nil {
		t.Errorf("Start = %v, want nil", enum.Start)
	}
	if enum.Full != false {
		t.Errorf("Full = %v, want false", enum.Full)
	}
	if enum.Reversed != false {
		t.Errorf("Reversed = %v, want false", enum.Reversed)
	}
	if len(enum.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(enum.Items))
	}

	// Verify item numbers
	if enum.Items[0].Number != 1 {
		t.Errorf("Items[0].Number = %d, want 1", enum.Items[0].Number)
	}
	if enum.Items[1].Number != 2 {
		t.Errorf("Items[1].Number = %d, want 2", enum.Items[1].Number)
	}
}

func TestEnumNativeWithNumbering(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("numbering", Str("a)"), syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Numbering != "a)" {
		t.Errorf("Numbering = %q, want %q", enum.Numbering, "a)")
	}
}

func TestEnumNativeWithStart(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "First"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Second"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("start", Int(5), syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Start == nil || *enum.Start != 5 {
		t.Errorf("Start = %v, want 5", enum.Start)
	}
	// Verify item numbers start from 5
	if enum.Items[0].Number != 5 {
		t.Errorf("Items[0].Number = %d, want 5", enum.Items[0].Number)
	}
	if enum.Items[1].Number != 6 {
		t.Errorf("Items[1].Number = %d, want 6", enum.Items[1].Number)
	}
}

func TestEnumNativeWithFull(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("full", True, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Full != true {
		t.Errorf("Full = %v, want true", enum.Full)
	}
}

func TestEnumNativeWithReversed(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("reversed", True, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Reversed != true {
		t.Errorf("Reversed = %v, want true", enum.Reversed)
	}
}

func TestEnumNativeWithIndent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("indent", LengthValue{Length: Length{Points: 15}}, syntax.Detached())
	args.Push(child, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Indent == nil || *enum.Indent != 15 {
		t.Errorf("Indent = %v, want 15", enum.Indent)
	}
}

func TestEnumNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "First"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Second"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("tight", False, syntax.Detached())
	args.PushNamed("numbering", Str("I."), syntax.Detached())
	args.PushNamed("start", Int(3), syntax.Detached())
	args.PushNamed("full", True, syntax.Detached())
	args.PushNamed("reversed", True, syntax.Detached())
	args.PushNamed("indent", LengthValue{Length: Length{Points: 5}}, syntax.Detached())
	args.PushNamed("body-indent", LengthValue{Length: Length{Points: 7}}, syntax.Detached())
	args.PushNamed("spacing", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Tight != false {
		t.Errorf("Tight = %v, want false", enum.Tight)
	}
	if enum.Numbering != "I." {
		t.Errorf("Numbering = %q, want %q", enum.Numbering, "I.")
	}
	if enum.Start == nil || *enum.Start != 3 {
		t.Errorf("Start = %v, want 3", enum.Start)
	}
	if enum.Full != true {
		t.Errorf("Full = %v, want true", enum.Full)
	}
	if enum.Reversed != true {
		t.Errorf("Reversed = %v, want true", enum.Reversed)
	}
	if enum.Indent == nil || *enum.Indent != 5 {
		t.Errorf("Indent = %v, want 5", enum.Indent)
	}
	if enum.BodyIndent == nil || *enum.BodyIndent != 7 {
		t.Errorf("BodyIndent = %v, want 7", enum.BodyIndent)
	}
	if enum.Spacing == nil || *enum.Spacing != 10 {
		t.Errorf("Spacing = %v, want 10", enum.Spacing)
	}
	if len(enum.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(enum.Items))
	}
}

func TestEnumNativeWrongChildType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.Push(Str("not content"), syntax.Detached())

	_, err := enumNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong child type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
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
	start := 1
	indent := 10.0
	bodyIndent := 8.0
	spacing := 12.0

	elem := &EnumElement{
		Items: []*EnumItemElement{
			{Number: 1, Content: Content{Elements: []ContentElement{&TextElement{Text: "A"}}}},
			{Number: 2, Content: Content{Elements: []ContentElement{&TextElement{Text: "B"}}}},
		},
		Tight:       true,
		Numbering:   "1.",
		Start:       &start,
		Full:        false,
		Reversed:    false,
		Indent:      &indent,
		BodyIndent:  &bodyIndent,
		Spacing:     &spacing,
		NumberAlign: "end",
	}

	if len(elem.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(elem.Items))
	}
	if elem.Tight != true {
		t.Errorf("Tight = %v, want true", elem.Tight)
	}
	if elem.Numbering != "1." {
		t.Errorf("Numbering = %q, want %q", elem.Numbering, "1.")
	}
	if *elem.Start != 1 {
		t.Errorf("Start = %v, want 1", *elem.Start)
	}
	if *elem.Indent != 10.0 {
		t.Errorf("Indent = %v, want 10.0", *elem.Indent)
	}
	if *elem.BodyIndent != 8.0 {
		t.Errorf("BodyIndent = %v, want 8.0", *elem.BodyIndent)
	}
	if *elem.Spacing != 12.0 {
		t.Errorf("Spacing = %v, want 12.0", *elem.Spacing)
	}
	if elem.NumberAlign != "end" {
		t.Errorf("NumberAlign = %q, want %q", elem.NumberAlign, "end")
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
// Nesting Tests for List and Enum
// ----------------------------------------------------------------------------

func TestListNativeWithNestedList(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create nested list content
	nestedChild := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Nested item"}},
	}}
	nestedArgs := NewArgs(syntax.Detached())
	nestedArgs.Push(nestedChild, syntax.Detached())
	nestedResult, _ := listNative(vm, nestedArgs)
	nestedContent := nestedResult.(ContentValue)

	// Create parent list with nested list inside
	parentChild := ContentValue{Content: Content{
		Elements: append(
			[]ContentElement{&TextElement{Text: "Parent item: "}},
			nestedContent.Content.Elements...,
		),
	}}

	args := NewArgs(syntax.Detached())
	args.Push(parentChild, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() with nested list error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if len(list.Items) != 1 {
		t.Errorf("Items length = %d, want 1", len(list.Items))
	}

	// Verify the item content contains both text and nested list
	item := list.Items[0]
	if len(item.Content.Elements) < 2 {
		t.Errorf("Item content elements = %d, want at least 2", len(item.Content.Elements))
	}
}

func TestEnumNativeWithNestedEnum(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create nested enum content
	nestedChild := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Nested item"}},
	}}
	nestedArgs := NewArgs(syntax.Detached())
	nestedArgs.Push(nestedChild, syntax.Detached())
	nestedResult, _ := enumNative(vm, nestedArgs)
	nestedContent := nestedResult.(ContentValue)

	// Create parent enum with nested enum inside
	parentChild := ContentValue{Content: Content{
		Elements: append(
			[]ContentElement{&TextElement{Text: "Parent item: "}},
			nestedContent.Content.Elements...,
		),
	}}

	args := NewArgs(syntax.Detached())
	args.Push(parentChild, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() with nested enum error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if len(enum.Items) != 1 {
		t.Errorf("Items length = %d, want 1", len(enum.Items))
	}

	// Verify the item content contains both text and nested enum
	item := enum.Items[0]
	if len(item.Content.Elements) < 2 {
		t.Errorf("Item content elements = %d, want at least 2", len(item.Content.Elements))
	}
}
