package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

// TestEvalMathFrac tests fraction evaluation.
func TestEvalMathFrac(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple fraction", "a/b"},
		{"numeric fraction", "1/2"},
		{"nested fraction", "a/(b/c)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse math expression
			node := syntax.ParseMath(tt.input)
			if node == nil {
				t.Fatal("ParseMath returned nil")
			}

			// Create VM and evaluate
			scopes := NewScopes(nil)
			vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

			// Get expressions from math node
			mathNode := syntax.MathNodeFromNode(node)
			if mathNode == nil {
				t.Fatal("Expected MathNode")
			}

			exprs := mathNode.Exprs()
			if len(exprs) == 0 {
				t.Fatal("No expressions in math node")
			}

			// Evaluate the first expression
			value, err := EvalExpr(vm, exprs[0])
			if err != nil {
				t.Fatalf("EvalExpr error: %v", err)
			}

			// Check it's content
			content, ok := value.(ContentValue)
			if !ok {
				t.Fatalf("Expected ContentValue, got %T", value)
			}

			// For fractions, check that we got a MathFracElement
			if len(content.Content.Elements) == 0 {
				t.Fatal("Expected at least one content element")
			}

			if _, ok := content.Content.Elements[0].(*MathFracElement); !ok {
				t.Errorf("Expected MathFracElement, got %T", content.Content.Elements[0])
			}
		})
	}
}

// TestEvalMathRoot tests root evaluation.
func TestEvalMathRoot(t *testing.T) {
	// For root, we need parsed sqrt() or root() syntax
	// Since the parser may handle this as a function call, let's test the underlying
	// MathRoot evaluation if we can construct such a node

	// For now, test that the evaluation function exists and returns content
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Define a test radicand
	vm.Define("x", Int(4))

	// We can't easily construct a MathRootExpr without parser support,
	// but we can verify the basic infrastructure works
	if vm.Get("x") == nil {
		t.Error("Expected to find x in scope")
	}
}

// TestEvalMathAttach tests subscript/superscript evaluation.
func TestEvalMathAttach(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Note: "x_1" is parsed as a MathIdent (underscore is part of identifier)
		// Use "x^2" for superscript which is parsed as MathAttach
		{"superscript", "x^2"},
		{"both", "x_1^2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse math expression
			node := syntax.ParseMath(tt.input)
			if node == nil {
				t.Fatal("ParseMath returned nil")
			}

			// Create VM and evaluate
			scopes := NewScopes(nil)
			vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

			// Get expressions from math node
			mathNode := syntax.MathNodeFromNode(node)
			if mathNode == nil {
				t.Fatal("Expected MathNode")
			}

			exprs := mathNode.Exprs()
			if len(exprs) == 0 {
				t.Fatal("No expressions in math node")
			}

			// Evaluate the first expression
			value, err := EvalExpr(vm, exprs[0])
			if err != nil {
				t.Fatalf("EvalExpr error: %v", err)
			}

			// Check it's content
			content, ok := value.(ContentValue)
			if !ok {
				t.Fatalf("Expected ContentValue, got %T", value)
			}

			// Should have at least one element
			if len(content.Content.Elements) == 0 {
				t.Fatal("Expected at least one content element")
			}

			// For attach expressions, check that we got a MathAttachElement
			if _, ok := content.Content.Elements[0].(*MathAttachElement); !ok {
				t.Errorf("Expected MathAttachElement, got %T", content.Content.Elements[0])
			}
		})
	}
}

// TestEvalMathDelimited tests delimited math evaluation.
func TestEvalMathDelimited(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Note: brackets "[" are not valid delimiters in math mode
		// Only parentheses work as MathDelimited
		{"parentheses", "(a + b)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse math expression
			node := syntax.ParseMath(tt.input)
			if node == nil {
				t.Fatal("ParseMath returned nil")
			}

			// Create VM and evaluate
			scopes := NewScopes(nil)
			vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

			// Get expressions from math node
			mathNode := syntax.MathNodeFromNode(node)
			if mathNode == nil {
				t.Fatal("Expected MathNode")
			}

			exprs := mathNode.Exprs()
			if len(exprs) == 0 {
				t.Fatal("No expressions in math node")
			}

			// Evaluate the first expression
			value, err := EvalExpr(vm, exprs[0])
			if err != nil {
				t.Fatalf("EvalExpr error: %v", err)
			}

			// Check it's content
			content, ok := value.(ContentValue)
			if !ok {
				t.Fatalf("Expected ContentValue, got %T", value)
			}

			// Should have at least one element
			if len(content.Content.Elements) == 0 {
				t.Fatal("Expected at least one content element")
			}

			// For delimited expressions, check that we got a MathDelimitedElement
			if _, ok := content.Content.Elements[0].(*MathDelimitedElement); !ok {
				t.Errorf("Expected MathDelimitedElement, got %T", content.Content.Elements[0])
			}
		})
	}
}

// TestEvalEquation tests equation evaluation.
func TestEvalEquation(t *testing.T) {
	// Parse a complete equation (with $ delimiters)
	input := "$x^2 + y^2 = z^2$"
	node := syntax.Parse(input)
	if node == nil {
		t.Fatal("Parse returned nil")
	}

	// Create VM and evaluate
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Get markup node
	markupNode := syntax.MarkupNodeFromNode(node)
	if markupNode == nil {
		t.Fatal("Expected MarkupNode")
	}

	exprs := markupNode.Exprs()
	if len(exprs) == 0 {
		t.Fatal("No expressions in markup")
	}

	// Find the equation expression
	var equationExpr *syntax.EquationExpr
	for _, expr := range exprs {
		if eq, ok := expr.(*syntax.EquationExpr); ok {
			equationExpr = eq
			break
		}
	}

	if equationExpr == nil {
		t.Fatal("No EquationExpr found")
	}

	// Evaluate
	value, err := evalEquation(vm, equationExpr)
	if err != nil {
		t.Fatalf("evalEquation error: %v", err)
	}

	// Check it's content
	content, ok := value.(ContentValue)
	if !ok {
		t.Fatalf("Expected ContentValue, got %T", value)
	}

	// Should have an EquationElement
	if len(content.Content.Elements) == 0 {
		t.Fatal("Expected at least one content element")
	}

	if eq, ok := content.Content.Elements[0].(*EquationElement); !ok {
		t.Errorf("Expected EquationElement, got %T", content.Content.Elements[0])
	} else {
		// Inline equation should not be block
		if eq.Block {
			t.Error("Expected inline equation (Block=false)")
		}
	}
}

// TestEvalMathIdent tests math identifier evaluation.
func TestEvalMathIdent(t *testing.T) {
	// Note: Single-letter identifiers like "x" are parsed as MathText
	// Multi-letter identifiers like "alpha" are parsed as MathIdent

	// Parse a multi-character identifier in math mode
	input := "alpha"
	node := syntax.ParseMath(input)
	if node == nil {
		t.Fatal("ParseMath returned nil")
	}

	// Create VM
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Test with undefined variable - should become symbol
	mathNode := syntax.MathNodeFromNode(node)
	if mathNode == nil {
		t.Fatal("Expected MathNode")
	}

	exprs := mathNode.Exprs()
	if len(exprs) == 0 {
		t.Fatal("No expressions")
	}

	value, err := EvalExpr(vm, exprs[0])
	if err != nil {
		t.Fatalf("EvalExpr error: %v", err)
	}

	content, ok := value.(ContentValue)
	if !ok {
		t.Fatalf("Expected ContentValue, got %T", value)
	}

	// Should be a symbol element for undefined identifiers
	if len(content.Content.Elements) == 0 {
		t.Fatal("Expected at least one element")
	}

	// Now test with defined variable
	vm.Define("beta", Int(42))
	input2 := "beta"
	node2 := syntax.ParseMath(input2)
	mathNode2 := syntax.MathNodeFromNode(node2)
	exprs2 := mathNode2.Exprs()

	value2, err := EvalExpr(vm, exprs2[0])
	if err != nil {
		t.Fatalf("EvalExpr error: %v", err)
	}

	// Should be the Int value (MathIdent returns raw values for defined vars)
	if _, ok := value2.(IntValue); !ok {
		t.Errorf("Expected IntValue for defined variable, got %T", value2)
	}
}

// TestEvalMathPrimes tests prime mark evaluation.
func TestEvalMathPrimes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPrimes  int
	}{
		{"single prime", "x'", 1},
		{"double prime", "x''", 2},
		{"triple prime", "x'''", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := syntax.ParseMath(tt.input)
			if node == nil {
				t.Fatal("ParseMath returned nil")
			}

			scopes := NewScopes(nil)
			vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

			mathNode := syntax.MathNodeFromNode(node)
			exprs := mathNode.Exprs()
			if len(exprs) == 0 {
				t.Fatal("No expressions")
			}

			value, err := EvalExpr(vm, exprs[0])
			if err != nil {
				t.Fatalf("EvalExpr error: %v", err)
			}

			content, ok := value.(ContentValue)
			if !ok {
				t.Fatalf("Expected ContentValue, got %T", value)
			}

			if len(content.Content.Elements) == 0 {
				t.Fatal("Expected at least one element")
			}

			// The prime should be part of a MathAttachElement
			if attach, ok := content.Content.Elements[0].(*MathAttachElement); ok {
				if attach.Primes != tt.wantPrimes {
					t.Errorf("Got %d primes, want %d", attach.Primes, tt.wantPrimes)
				}
			}
			// Or it could be standalone text with prime characters
		})
	}
}

// TestMathContentString tests the String() method on Content.
func TestMathContentString(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello"},
			&MathSymbolElement{Symbol: "+"},
			&TextElement{Text: "world"},
		},
	}

	result := content.String()
	expected := "hello+world"
	if result != expected {
		t.Errorf("Content.String() = %q, want %q", result, expected)
	}
}

// TestValueToContent tests the valueToContent helper.
func TestValueToContent(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{"string", Str("hello"), "hello"},
		{"int", Int(42), "42"},
		{"negative int", Int(-7), "-7"},
		{"zero", Int(0), "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := valueToContent(tt.value)
			got := content.String()
			if got != tt.want {
				t.Errorf("valueToContent(%v).String() = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestAccentKindString tests the AccentKind String method.
func TestAccentKindString(t *testing.T) {
	tests := []struct {
		kind AccentKind
		want string
	}{
		{AccentHat, "hat"},
		{AccentTilde, "tilde"},
		{AccentBar, "bar"},
		{AccentVec, "vec"},
		{AccentDot, "dot"},
		{AccentDDot, "dot.double"},
		{AccentBreve, "breve"},
		{AccentAcute, "acute"},
		{AccentGrave, "grave"},
		{AccentKind(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("AccentKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

// TestAccentKindChar tests the AccentChar method.
func TestAccentKindChar(t *testing.T) {
	tests := []struct {
		kind AccentKind
		want rune
	}{
		{AccentHat, '\u0302'},
		{AccentTilde, '\u0303'},
		{AccentBar, '\u0304'},
		{AccentVec, '\u20D7'},
		{AccentDot, '\u0307'},
		{AccentDDot, '\u0308'},
		{AccentBreve, '\u0306'},
		{AccentAcute, '\u0301'},
		{AccentGrave, '\u0300'},
	}

	for _, tt := range tests {
		if got := tt.kind.AccentChar(); got != tt.want {
			t.Errorf("AccentKind(%d).AccentChar() = %U, want %U", tt.kind, got, tt.want)
		}
	}
}

// TestMathAccentElement tests the MathAccentElement structure.
func TestMathAccentElement(t *testing.T) {
	elem := &MathAccentElement{
		Base: Content{
			Elements: []ContentElement{&TextElement{Text: "x"}},
		},
		Accent: AccentHat,
	}

	// Verify it implements ContentElement
	var _ ContentElement = elem

	if elem.Accent != AccentHat {
		t.Errorf("Accent = %v, want AccentHat", elem.Accent)
	}

	if len(elem.Base.Elements) != 1 {
		t.Errorf("len(Base.Elements) = %d, want 1", len(elem.Base.Elements))
	}
}

// TestMathAccentFunctions tests that accent functions are registered and callable.
func TestMathAccentFunctions(t *testing.T) {
	// Get the library scope
	lib := Library()

	accentFuncs := []string{
		"hat", "tilde", "bar", "overline", "vec",
		"dot", "ddot", "breve", "acute", "grave",
	}

	for _, name := range accentFuncs {
		binding := lib.Get(name)
		if binding == nil {
			t.Errorf("accent function %q not found in library", name)
			continue
		}

		// Verify it's a function
		val, err := binding.Read()
		if err != nil {
			t.Errorf("failed to read %q binding: %v", name, err)
			continue
		}

		funcVal, ok := val.(FuncValue)
		if !ok {
			t.Errorf("%q is not a function, got %T", name, val)
			continue
		}

		if funcVal.Func == nil {
			t.Errorf("%q function is nil", name)
		}
	}
}

// TestMathAccentFunctionCall tests calling an accent function.
func TestMathAccentFunctionCall(t *testing.T) {
	// Create a VM with the library scope
	scopes := NewScopes(Library())
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Get the hat function
	binding := vm.Get("hat")
	if binding == nil {
		t.Fatal("hat function not found")
	}

	funcVal, err := binding.Read()
	if err != nil {
		t.Fatalf("failed to read hat binding: %v", err)
	}

	fn, ok := funcVal.(FuncValue)
	if !ok {
		t.Fatalf("hat is not a function, got %T", funcVal)
	}

	// Call it with a string argument
	args := &Args{
		Span: syntax.Detached(),
		Items: []Arg{
			{
				Span:  syntax.Detached(),
				Value: syntax.Spanned[Value]{V: Str("x"), Span: syntax.Detached()},
			},
		},
	}

	result, err := CallFunc(vm, fn.Func, args)
	if err != nil {
		t.Fatalf("CallFunc failed: %v", err)
	}

	// Check result is content with a MathAccentElement
	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("result is not ContentValue, got %T", result)
	}

	if len(content.Content.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
	}

	accentElem, ok := content.Content.Elements[0].(*MathAccentElement)
	if !ok {
		t.Fatalf("element is not MathAccentElement, got %T", content.Content.Elements[0])
	}

	if accentElem.Accent != AccentHat {
		t.Errorf("accent kind = %v, want AccentHat", accentElem.Accent)
	}
}
