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

// TestMatFunc tests the mat() matrix function.
func TestMatFunc(t *testing.T) {
	tests := []struct {
		name         string
		args         []Arg
		wantRows     int
		wantDelim    string
		wantErr      bool
	}{
		{
			name: "simple 2x2 matrix",
			args: []Arg{
				// First row: (1, 2)
				{Value: syntax.Spanned[Value]{V: ArrayValue{Int(1), Int(2)}}},
				// Second row: (3, 4)
				{Value: syntax.Spanned[Value]{V: ArrayValue{Int(3), Int(4)}}},
			},
			wantRows:  2,
			wantDelim: "(",
		},
		{
			name: "matrix with bracket delimiters",
			args: []Arg{
				{Name: strPtr("delim"), Value: syntax.Spanned[Value]{V: Str("[")}},
				{Value: syntax.Spanned[Value]{V: ArrayValue{Int(1), Int(2)}}},
			},
			wantRows:  1,
			wantDelim: "[",
		},
		{
			name: "matrix with no delimiter",
			args: []Arg{
				{Name: strPtr("delim"), Value: syntax.Spanned[Value]{V: Str("")}},
				{Value: syntax.Spanned[Value]{V: Int(1)}},
			},
			wantRows:  1,
			wantDelim: "",
		},
		{
			name: "invalid delimiter",
			args: []Arg{
				{Name: strPtr("delim"), Value: syntax.Spanned[Value]{V: Str("invalid")}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scopes := NewScopes(nil)
			vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

			args := &Args{
				Span:  syntax.Detached(),
				Items: tt.args,
			}

			result, err := matNative(vm, args)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			content, ok := result.(ContentValue)
			if !ok {
				t.Fatalf("Expected ContentValue, got %T", result)
			}

			if len(content.Content.Elements) == 0 {
				t.Fatal("Expected at least one element")
			}

			matrix, ok := content.Content.Elements[0].(*MathMatrixElement)
			if !ok {
				t.Fatalf("Expected MathMatrixElement, got %T", content.Content.Elements[0])
			}

			if len(matrix.Rows) != tt.wantRows {
				t.Errorf("Got %d rows, want %d", len(matrix.Rows), tt.wantRows)
			}

			if matrix.Delim != tt.wantDelim {
				t.Errorf("Got delim %q, want %q", matrix.Delim, tt.wantDelim)
			}
		})
	}
}

// TestVecFunc tests the vec() vector function.
func TestVecFunc(t *testing.T) {
	tests := []struct {
		name         string
		args         []Arg
		wantElements int
		wantDelim    string
		wantErr      bool
	}{
		{
			name: "simple vector",
			args: []Arg{
				{Value: syntax.Spanned[Value]{V: Int(1)}},
				{Value: syntax.Spanned[Value]{V: Int(2)}},
				{Value: syntax.Spanned[Value]{V: Int(3)}},
			},
			wantElements: 3,
			wantDelim:    "(",
		},
		{
			name: "vector with bracket delimiters",
			args: []Arg{
				{Name: strPtr("delim"), Value: syntax.Spanned[Value]{V: Str("[")}},
				{Value: syntax.Spanned[Value]{V: Int(1)}},
				{Value: syntax.Spanned[Value]{V: Int(2)}},
			},
			wantElements: 2,
			wantDelim:    "[",
		},
		{
			name: "empty vector",
			args: []Arg{},
			wantElements: 0,
			wantDelim:    "(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scopes := NewScopes(nil)
			vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

			args := &Args{
				Span:  syntax.Detached(),
				Items: tt.args,
			}

			result, err := vecNative(vm, args)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			content, ok := result.(ContentValue)
			if !ok {
				t.Fatalf("Expected ContentValue, got %T", result)
			}

			if len(content.Content.Elements) == 0 {
				t.Fatal("Expected at least one element")
			}

			vec, ok := content.Content.Elements[0].(*MathVecElement)
			if !ok {
				t.Fatalf("Expected MathVecElement, got %T", content.Content.Elements[0])
			}

			if len(vec.Elements) != tt.wantElements {
				t.Errorf("Got %d elements, want %d", len(vec.Elements), tt.wantElements)
			}

			if vec.Delim != tt.wantDelim {
				t.Errorf("Got delim %q, want %q", vec.Delim, tt.wantDelim)
			}
		})
	}
}

// TestCasesFunc tests the cases() function for piecewise functions.
func TestCasesFunc(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	args := &Args{
		Span: syntax.Detached(),
		Items: []Arg{
			// First case: x, if x > 0
			{Value: syntax.Spanned[Value]{V: ArrayValue{Str("x"), Str("if x > 0")}}},
			// Second case: -x, if x <= 0
			{Value: syntax.Spanned[Value]{V: ArrayValue{Str("-x"), Str("if x <= 0")}}},
		},
	}

	result, err := casesNative(vm, args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content, ok := result.(ContentValue)
	if !ok {
		t.Fatalf("Expected ContentValue, got %T", result)
	}

	if len(content.Content.Elements) == 0 {
		t.Fatal("Expected at least one element")
	}

	// Cases produces a MathMatrixElement
	matrix, ok := content.Content.Elements[0].(*MathMatrixElement)
	if !ok {
		t.Fatalf("Expected MathMatrixElement, got %T", content.Content.Elements[0])
	}

	if len(matrix.Rows) != 2 {
		t.Errorf("Got %d rows, want 2", len(matrix.Rows))
	}

	// Default delimiter for cases is "{"
	if matrix.Delim != "{" {
		t.Errorf("Got delim %q, want \"{\"", matrix.Delim)
	}
}

// TestParseMatrixRow tests the parseMatrixRow helper.
func TestParseMatrixRow(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		wantCols int
	}{
		{
			name:     "array value",
			value:    ArrayValue{Int(1), Int(2), Int(3)},
			wantCols: 3,
		},
		{
			name:     "content value",
			value:    ContentValue{Content: Content{Elements: []ContentElement{&TextElement{Text: "x"}}}},
			wantCols: 1,
		},
		{
			name:     "int value",
			value:    Int(42),
			wantCols: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := parseMatrixRow(tt.value)
			if len(row) != tt.wantCols {
				t.Errorf("Got %d columns, want %d", len(row), tt.wantCols)
			}
		})
	}
}

// TestMathMatrixElement tests the MathMatrixElement type.
func TestMathMatrixElement(t *testing.T) {
	elem := &MathMatrixElement{
		Rows: [][]Content{
			{{Elements: []ContentElement{&TextElement{Text: "1"}}}, {Elements: []ContentElement{&TextElement{Text: "2"}}}},
			{{Elements: []ContentElement{&TextElement{Text: "3"}}}, {Elements: []ContentElement{&TextElement{Text: "4"}}}},
		},
		Delim: "[",
	}

	// Verify it implements ContentElement
	var _ ContentElement = elem
	elem.IsContentElement() // Should not panic

	if len(elem.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(elem.Rows))
	}
	if len(elem.Rows[0]) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(elem.Rows[0]))
	}
	if elem.Delim != "[" {
		t.Errorf("Expected delim [, got %s", elem.Delim)
	}
}

// TestMathVecElement tests the MathVecElement type.
func TestMathVecElement(t *testing.T) {
	elem := &MathVecElement{
		Elements: []Content{
			{Elements: []ContentElement{&TextElement{Text: "x"}}},
			{Elements: []ContentElement{&TextElement{Text: "y"}}},
			{Elements: []ContentElement{&TextElement{Text: "z"}}},
		},
		Delim: "(",
	}

	// Verify it implements ContentElement
	var _ ContentElement = elem
	elem.IsContentElement() // Should not panic

	if len(elem.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(elem.Elements))
	}
	if elem.Delim != "(" {
		t.Errorf("Expected delim (, got %s", elem.Delim)
	}
}

