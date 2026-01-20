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

	// Create args with no children
	args := NewArgs(syntax.Detached())

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
	if len(list.Children) != 0 {
		t.Errorf("Children = %d, want 0", len(list.Children))
	}
	if list.Marker != nil {
		t.Errorf("Marker = %v, want nil (default)", list.Marker)
	}
	if list.Indent != nil {
		t.Errorf("Indent = %v, want nil (default)", list.Indent)
	}
	if list.BodyIndent != nil {
		t.Errorf("BodyIndent = %v, want nil (default)", list.BodyIndent)
	}
	if list.Spacing != nil {
		t.Errorf("Spacing = %v, want nil (default)", list.Spacing)
	}
	if list.Tight != nil {
		t.Errorf("Tight = %v, want nil (default)", list.Tight)
	}
}

func TestListNativeWithChildren(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create two child items
	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "First item"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Second item"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if len(list.Children) != 2 {
		t.Errorf("Children = %d, want 2", len(list.Children))
	}
}

func TestListNativeWithMarker(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	marker := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "→"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("marker", marker, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Marker == nil {
		t.Fatal("Marker = nil, want non-nil")
	}
}

func TestListNativeWithIndent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("indent", LengthValue{Length: Length{Points: 20}}, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Indent == nil || *list.Indent != 20 {
		t.Errorf("Indent = %v, want 20", list.Indent)
	}
}

func TestListNativeWithBodyIndent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("body-indent", LengthValue{Length: Length{Points: 15}}, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.BodyIndent == nil || *list.BodyIndent != 15 {
		t.Errorf("BodyIndent = %v, want 15", list.BodyIndent)
	}
}

func TestListNativeWithSpacing(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("spacing", LengthValue{Length: Length{Points: 10}}, syntax.Detached())

	result, err := listNative(vm, args)
	if err != nil {
		t.Fatalf("listNative() error: %v", err)
	}

	content := result.(ContentValue)
	list := content.Content.Elements[0].(*ListElement)

	if list.Spacing == nil || *list.Spacing != 10 {
		t.Errorf("Spacing = %v, want 10", list.Spacing)
	}
}

func TestListNativeWithTight(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("tight", False, syntax.Detached())

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
	indent := 20.0
	bodyIndent := 15.0
	spacing := 10.0
	tight := true

	elem := &ListElement{
		Children: []Content{
			{Elements: []ContentElement{&TextElement{Text: "Item 1"}}},
			{Elements: []ContentElement{&TextElement{Text: "Item 2"}}},
		},
		Indent:     &indent,
		BodyIndent: &bodyIndent,
		Spacing:    &spacing,
		Tight:      &tight,
	}

	if len(elem.Children) != 2 {
		t.Errorf("Children = %d, want 2", len(elem.Children))
	}
	if *elem.Indent != 20.0 {
		t.Errorf("Indent = %v, want 20.0", *elem.Indent)
	}
	if *elem.BodyIndent != 15.0 {
		t.Errorf("BodyIndent = %v, want 15.0", *elem.BodyIndent)
	}
	if *elem.Spacing != 10.0 {
		t.Errorf("Spacing = %v, want 10.0", *elem.Spacing)
	}
	if *elem.Tight != true {
		t.Errorf("Tight = %v, want true", *elem.Tight)
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

	args := NewArgs(syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	enum, ok := content.Content.Elements[0].(*EnumElement)
	if !ok {
		t.Fatalf("expected *EnumElement, got %T", content.Content.Elements[0])
	}

	// Verify defaults
	if enum.Numbering != nil {
		t.Errorf("Numbering = %v, want nil", enum.Numbering)
	}
	if enum.Start != nil {
		t.Errorf("Start = %v, want nil", enum.Start)
	}
	if enum.Full != nil {
		t.Errorf("Full = %v, want nil", enum.Full)
	}
}

func TestEnumNativeWithNumbering(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("numbering", Str("a)"), syntax.Detached())

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

	args := NewArgs(syntax.Detached())
	args.PushNamed("start", Int(5), syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Start == nil || *enum.Start != 5 {
		t.Errorf("Start = %v, want 5", enum.Start)
	}
}

func TestEnumNativeWithFull(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("full", True, syntax.Detached())

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

func TestEnumNativeWithChildren(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "First"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Second"}},
	}}
	child3 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Third"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())
	args.Push(child3, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if len(enum.Children) != 3 {
		t.Errorf("Children = %d, want 3", len(enum.Children))
	}
}

func TestEnumNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("numbering", Str("(i)"), syntax.Detached())
	args.PushNamed("start", Int(3), syntax.Detached())
	args.PushNamed("full", True, syntax.Detached())
	args.PushNamed("indent", LengthValue{Length: Length{Points: 10}}, syntax.Detached())
	args.PushNamed("body-indent", LengthValue{Length: Length{Points: 20}}, syntax.Detached())
	args.PushNamed("spacing", LengthValue{Length: Length{Points: 5}}, syntax.Detached())
	args.PushNamed("tight", False, syntax.Detached())

	result, err := enumNative(vm, args)
	if err != nil {
		t.Fatalf("enumNative() error: %v", err)
	}

	content := result.(ContentValue)
	enum := content.Content.Elements[0].(*EnumElement)

	if enum.Numbering == nil || *enum.Numbering != "(i)" {
		t.Errorf("Numbering = %v, want '(i)'", enum.Numbering)
	}
	if enum.Start == nil || *enum.Start != 3 {
		t.Errorf("Start = %v, want 3", enum.Start)
	}
	if enum.Full == nil || *enum.Full != true {
		t.Errorf("Full = %v, want true", enum.Full)
	}
	if enum.Indent == nil || *enum.Indent != 10 {
		t.Errorf("Indent = %v, want 10", enum.Indent)
	}
	if enum.BodyIndent == nil || *enum.BodyIndent != 20 {
		t.Errorf("BodyIndent = %v, want 20", enum.BodyIndent)
	}
	if enum.Spacing == nil || *enum.Spacing != 5 {
		t.Errorf("Spacing = %v, want 5", enum.Spacing)
	}
	if enum.Tight == nil || *enum.Tight != false {
		t.Errorf("Tight = %v, want false", enum.Tight)
	}
}

func TestEnumElement(t *testing.T) {
	numbering := "a)"
	start := 1
	full := true

	elem := &EnumElement{
		Children: []Content{
			{Elements: []ContentElement{&TextElement{Text: "Item 1"}}},
		},
		Numbering: &numbering,
		Start:     &start,
		Full:      &full,
	}

	if len(elem.Children) != 1 {
		t.Errorf("Children = %d, want 1", len(elem.Children))
	}
	if *elem.Numbering != "a)" {
		t.Errorf("Numbering = %v, want 'a)'", *elem.Numbering)
	}
	if *elem.Start != 1 {
		t.Errorf("Start = %v, want 1", *elem.Start)
	}
	if *elem.Full != true {
		t.Errorf("Full = %v, want true", *elem.Full)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Terms Tests
// ----------------------------------------------------------------------------

func TestTermsFunc(t *testing.T) {
	// Get the terms function
	termsFunc := TermsFunc()

	if termsFunc == nil {
		t.Fatal("TermsFunc() returned nil")
	}

	if termsFunc.Name == nil || *termsFunc.Name != "terms" {
		t.Errorf("expected function name 'terms', got %v", termsFunc.Name)
	}

	// Verify it's a native function
	_, ok := termsFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}
}

func TestTermsNativeBasic(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())

	result, err := termsNative(vm, args)
	if err != nil {
		t.Fatalf("termsNative() error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	terms, ok := content.Content.Elements[0].(*TermsElement)
	if !ok {
		t.Fatalf("expected *TermsElement, got %T", content.Content.Elements[0])
	}

	// Verify defaults
	if terms.Separator != nil {
		t.Errorf("Separator = %v, want nil", terms.Separator)
	}
	if terms.Indent != nil {
		t.Errorf("Indent = %v, want nil", terms.Indent)
	}
	if terms.HangingIndent != nil {
		t.Errorf("HangingIndent = %v, want nil", terms.HangingIndent)
	}
}

func TestTermsNativeWithSeparator(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	sep := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: " — "}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("separator", sep, syntax.Detached())

	result, err := termsNative(vm, args)
	if err != nil {
		t.Fatalf("termsNative() error: %v", err)
	}

	content := result.(ContentValue)
	terms := content.Content.Elements[0].(*TermsElement)

	if terms.Separator == nil {
		t.Fatal("Separator = nil, want non-nil")
	}
}

func TestTermsNativeWithHangingIndent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := NewArgs(syntax.Detached())
	args.PushNamed("hanging-indent", LengthValue{Length: Length{Points: 24}}, syntax.Detached())

	result, err := termsNative(vm, args)
	if err != nil {
		t.Fatalf("termsNative() error: %v", err)
	}

	content := result.(ContentValue)
	terms := content.Content.Elements[0].(*TermsElement)

	if terms.HangingIndent == nil || *terms.HangingIndent != 24 {
		t.Errorf("HangingIndent = %v, want 24", terms.HangingIndent)
	}
}

func TestTermsNativeWithChildren(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	child1 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Term: Desc"}},
	}}
	child2 := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Another: Term"}},
	}}

	args := NewArgs(syntax.Detached())
	args.Push(child1, syntax.Detached())
	args.Push(child2, syntax.Detached())

	result, err := termsNative(vm, args)
	if err != nil {
		t.Fatalf("termsNative() error: %v", err)
	}

	content := result.(ContentValue)
	terms := content.Content.Elements[0].(*TermsElement)

	if len(terms.Children) != 2 {
		t.Errorf("Children = %d, want 2", len(terms.Children))
	}
}

func TestTermsNativeWithAllParams(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	sep := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: ": "}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("separator", sep, syntax.Detached())
	args.PushNamed("indent", LengthValue{Length: Length{Points: 5}}, syntax.Detached())
	args.PushNamed("hanging-indent", LengthValue{Length: Length{Points: 20}}, syntax.Detached())
	args.PushNamed("spacing", LengthValue{Length: Length{Points: 8}}, syntax.Detached())
	args.PushNamed("tight", True, syntax.Detached())

	result, err := termsNative(vm, args)
	if err != nil {
		t.Fatalf("termsNative() error: %v", err)
	}

	content := result.(ContentValue)
	terms := content.Content.Elements[0].(*TermsElement)

	if terms.Separator == nil {
		t.Fatal("Separator = nil, want non-nil")
	}
	if terms.Indent == nil || *terms.Indent != 5 {
		t.Errorf("Indent = %v, want 5", terms.Indent)
	}
	if terms.HangingIndent == nil || *terms.HangingIndent != 20 {
		t.Errorf("HangingIndent = %v, want 20", terms.HangingIndent)
	}
	if terms.Spacing == nil || *terms.Spacing != 8 {
		t.Errorf("Spacing = %v, want 8", terms.Spacing)
	}
	if terms.Tight == nil || *terms.Tight != true {
		t.Errorf("Tight = %v, want true", terms.Tight)
	}
}

func TestTermsElement(t *testing.T) {
	indent := 5.0
	hangingIndent := 20.0

	elem := &TermsElement{
		Children: []Content{
			{Elements: []ContentElement{&TextElement{Text: "Term"}}},
		},
		Indent:        &indent,
		HangingIndent: &hangingIndent,
	}

	if len(elem.Children) != 1 {
		t.Errorf("Children = %d, want 1", len(elem.Children))
	}
	if *elem.Indent != 5.0 {
		t.Errorf("Indent = %v, want 5.0", *elem.Indent)
	}
	if *elem.HangingIndent != 20.0 {
		t.Errorf("HangingIndent = %v, want 20.0", *elem.HangingIndent)
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

// ----------------------------------------------------------------------------
// Registration Tests for List, Enum, and Terms
// ----------------------------------------------------------------------------

func TestRegisterElementFunctionsIncludesListEnumTerms(t *testing.T) {
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

	// Verify terms function is registered
	termsBinding := scope.Get("terms")
	if termsBinding == nil {
		t.Fatal("expected 'terms' to be registered")
	}
	termsFunc, ok := termsBinding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue for terms, got %T", termsBinding.Value)
	}
	if termsFunc.Func.Name == nil || *termsFunc.Func.Name != "terms" {
		t.Errorf("expected function name 'terms', got %v", termsFunc.Func.Name)
	}
}

func TestElementFunctionsIncludesListEnumTerms(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["list"]; !ok {
		t.Error("expected 'list' in ElementFunctions()")
	}

	if _, ok := funcs["enum"]; !ok {
		t.Error("expected 'enum' in ElementFunctions()")
	}

	if _, ok := funcs["terms"]; !ok {
		t.Error("expected 'terms' in ElementFunctions()")
	}
}
