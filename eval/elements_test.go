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
// Page Element Tests
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
	native, ok := pageFunc.Repr.(NativeFunc)
	if !ok {
		t.Error("expected NativeFunc representation")
	}

	// Verify function info
	if native.Info == nil {
		t.Fatal("expected FuncInfo to be set")
	}
	if native.Info.Name != "page" {
		t.Errorf("FuncInfo.Name = %q, want %q", native.Info.Name, "page")
	}
}

func TestPageNativeBasic(t *testing.T) {
	// Create a VM with minimal setup
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create body content
	bodyContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Hello, World!"}},
	}}

	// Create args with just body
	args := NewArgs(syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	// Call the page function
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

	// Verify defaults
	if page.Paper != "" {
		t.Errorf("Paper = %q, want empty", page.Paper)
	}
	if page.Columns != 1 {
		t.Errorf("Columns = %d, want 1", page.Columns)
	}
	if page.Flipped != false {
		t.Errorf("Flipped = %v, want false", page.Flipped)
	}

	// Verify body content is set
	if len(page.Body.Elements) != 1 {
		t.Fatalf("expected 1 body element, got %d", len(page.Body.Elements))
	}
}

func TestPageNativeWithPaper(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("paper", Str("a4"), syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Paper != "a4" {
		t.Errorf("Paper = %q, want %q", page.Paper, "a4")
	}
}

func TestPageNativeWithDimensions(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("width", LengthValue{Length: Length{Points: 612}}, syntax.Detached())
	args.PushNamed("height", LengthValue{Length: Length{Points: 792}}, syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Width == nil {
		t.Fatal("Width should not be nil")
	}
	if page.Width.Abs.Points != 612 {
		t.Errorf("Width = %v, want 612pt", page.Width.Abs.Points)
	}

	if page.Height == nil {
		t.Fatal("Height should not be nil")
	}
	if page.Height.Abs.Points != 792 {
		t.Errorf("Height = %v, want 792pt", page.Height.Abs.Points)
	}
}

func TestPageNativeWithUniformMargin(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("margin", LengthValue{Length: Length{Points: 72}}, syntax.Detached()) // 1 inch
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Margin == nil {
		t.Fatal("Margin should not be nil")
	}
	if page.Margin.Uniform == nil {
		t.Fatal("Margin.Uniform should not be nil")
	}
	if page.Margin.Uniform.Abs.Points != 72 {
		t.Errorf("Margin = %v, want 72pt", page.Margin.Uniform.Abs.Points)
	}
}

func TestPageNativeWithDictMargin(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	// Create dictionary margin
	marginDict := NewDict()
	marginDict.Set("left", LengthValue{Length: Length{Points: 36}})
	marginDict.Set("right", LengthValue{Length: Length{Points: 36}})
	marginDict.Set("top", LengthValue{Length: Length{Points: 72}})
	marginDict.Set("bottom", LengthValue{Length: Length{Points: 72}})

	args := NewArgs(syntax.Detached())
	args.PushNamed("margin", marginDict, syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Margin == nil {
		t.Fatal("Margin should not be nil")
	}
	if page.Margin.Left == nil || page.Margin.Left.Abs.Points != 36 {
		t.Errorf("Margin.Left = %v, want 36pt", page.Margin.Left)
	}
	if page.Margin.Top == nil || page.Margin.Top.Abs.Points != 72 {
		t.Errorf("Margin.Top = %v, want 72pt", page.Margin.Top)
	}
}

func TestPageNativeWithColumns(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Int(2), syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Columns != 2 {
		t.Errorf("Columns = %d, want 2", page.Columns)
	}
}

func TestPageNativeWithFill(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	// Create a white color
	fillColor := ColorValue{Color: Color{R: 255, G: 255, B: 255, A: 255}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("fill", fillColor, syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Fill == nil {
		t.Fatal("Fill should not be nil")
	}
	fill, ok := page.Fill.(ColorValue)
	if !ok {
		t.Fatalf("Fill = %T, want ColorValue", page.Fill)
	}
	if fill.Color.R != 255 || fill.Color.G != 255 || fill.Color.B != 255 {
		t.Errorf("Fill color = %v, want white", fill.Color)
	}
}

func TestPageNativeWithNumbering(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("numbering", Str("1"), syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Numbering == nil {
		t.Fatal("Numbering should not be nil")
	}
	if *page.Numbering != "1" {
		t.Errorf("Numbering = %q, want %q", *page.Numbering, "1")
	}
}

func TestPageNativeWithHeader(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}
	headerContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Header"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("header", headerContent, syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Header == nil {
		t.Fatal("Header should not be nil")
	}
	if len(page.Header.Elements) != 1 {
		t.Errorf("Header elements = %d, want 1", len(page.Header.Elements))
	}
}

func TestPageNativeWithFooter(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}
	footerContent := ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: "Footer"}},
	}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("footer", footerContent, syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Footer == nil {
		t.Fatal("Footer should not be nil")
	}
	if len(page.Footer.Elements) != 1 {
		t.Errorf("Footer elements = %d, want 1", len(page.Footer.Elements))
	}
}

func TestPageNativeWithFlipped(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("paper", Str("a4"), syntax.Detached())
	args.PushNamed("flipped", True, syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Flipped != true {
		t.Errorf("Flipped = %v, want true", page.Flipped)
	}
}

func TestPageNativeWithBinding(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("binding", Str("left"), syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	result, err := pageNative(vm, args)
	if err != nil {
		t.Fatalf("pageNative() error: %v", err)
	}

	content := result.(ContentValue)
	page := content.Content.Elements[0].(*PageElement)

	if page.Binding == nil {
		t.Fatal("Binding should not be nil")
	}
	if *page.Binding != PageBindingLeft {
		t.Errorf("Binding = %v, want PageBindingLeft", *page.Binding)
	}
}

func TestPageElement(t *testing.T) {
	// Test PageElement struct and ContentElement interface
	numbering := "1"
	elem := &PageElement{
		Paper:     "a4",
		Columns:   2,
		Numbering: &numbering,
		Body: Content{
			Elements: []ContentElement{&TextElement{Text: "content"}},
		},
	}

	if elem.Paper != "a4" {
		t.Errorf("Paper = %q, want %q", elem.Paper, "a4")
	}
	if elem.Columns != 2 {
		t.Errorf("Columns = %d, want 2", elem.Columns)
	}
	if elem.Numbering == nil || *elem.Numbering != "1" {
		t.Errorf("Numbering = %v, want %q", elem.Numbering, "1")
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = elem
}

func TestPaperSizes(t *testing.T) {
	// Test standard paper sizes are defined
	testCases := []struct {
		name          string
		expectedWidth float64
		expectedHeight float64
	}{
		{"a4", 595.28, 841.89},
		{"us-letter", 612.0, 792.0},
		{"a3", 841.89, 1190.55},
		{"us-legal", 612.0, 1008.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			size := GetPaperSize(tc.name)
			if size == nil {
				t.Fatalf("Paper size %q not found", tc.name)
			}
			// Allow small floating point tolerance
			if diff := size.Width - tc.expectedWidth; diff < -0.01 || diff > 0.01 {
				t.Errorf("%s width = %v, want %v", tc.name, size.Width, tc.expectedWidth)
			}
			if diff := size.Height - tc.expectedHeight; diff < -0.01 || diff > 0.01 {
				t.Errorf("%s height = %v, want %v", tc.name, size.Height, tc.expectedHeight)
			}
		})
	}
}

func TestPaperSizeNotFound(t *testing.T) {
	size := GetPaperSize("nonexistent")
	if size != nil {
		t.Errorf("expected nil for nonexistent paper size, got %v", size)
	}
}

func TestPageNativeWrongColumnType(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("columns", Str("two"), syntax.Detached())
	args.Push(bodyContent, syntax.Detached())

	_, err := pageNative(vm, args)
	if err == nil {
		t.Error("expected error for wrong columns type")
	}
	if _, ok := err.(*TypeMismatchError); !ok {
		t.Errorf("expected TypeMismatchError, got %T", err)
	}
}

func TestPageNativeInvalidBinding(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	bodyContent := ContentValue{Content: Content{}}

	args := NewArgs(syntax.Detached())
	args.PushNamed("binding", Str("center"), syntax.Detached()) // invalid value
	args.Push(bodyContent, syntax.Detached())

	_, err := pageNative(vm, args)
	if err == nil {
		t.Error("expected error for invalid binding value")
	}
}

func TestPageFunctionRegistration(t *testing.T) {
	scope := NewScope()
	RegisterElementFunctions(scope)

	// Verify page function is registered
	binding := scope.Get("page")
	if binding == nil {
		t.Fatal("expected 'page' to be registered")
	}

	funcVal, ok := binding.Value.(FuncValue)
	if !ok {
		t.Fatalf("expected FuncValue, got %T", binding.Value)
	}

	if funcVal.Func.Name == nil || *funcVal.Func.Name != "page" {
		t.Errorf("expected function name 'page', got %v", funcVal.Func.Name)
	}
}

func TestElementFunctionsIncludesPage(t *testing.T) {
	funcs := ElementFunctions()

	if _, ok := funcs["page"]; !ok {
		t.Error("expected 'page' in ElementFunctions()")
	}
}
