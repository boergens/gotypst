package syntax

import "testing"

// TestParseBasicMarkup tests parsing basic markup content.
func TestParseBasicMarkup(t *testing.T) {
	tests := []struct {
		name  string
		input string
		kind  SyntaxKind
	}{
		{"empty", "", Markup},
		{"text only", "Hello world", Markup},
		{"with emphasis", "_emphasis_", Markup},
		{"with strong", "*strong*", Markup},
		{"with heading", "= Heading", Markup},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := Parse(tt.input)
			if node == nil {
				t.Fatal("Parse returned nil")
			}
			if node.Kind() != tt.kind {
				t.Errorf("Parse(%q).Kind() = %v, want %v", tt.input, node.Kind(), tt.kind)
			}
		})
	}
}

// TestParseBasicCode tests parsing basic code expressions.
func TestParseBasicCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		kind  SyntaxKind
	}{
		{"empty", "", Code},
		{"integer", "42", Code},
		{"float", "3.14", Code},
		{"string", `"hello"`, Code},
		{"identifier", "foo", Code},
		{"let binding", "let x = 1", Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ParseCode(tt.input)
			if node == nil {
				t.Fatal("ParseCode returned nil")
			}
			if node.Kind() != tt.kind {
				t.Errorf("ParseCode(%q).Kind() = %v, want %v", tt.input, node.Kind(), tt.kind)
			}
		})
	}
}

// TestParseBasicMath tests parsing basic math expressions.
func TestParseBasicMath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		kind  SyntaxKind
	}{
		{"empty", "", Math},
		{"variable", "x", Math},
		{"fraction", "a/b", Math},
		{"subscript", "x_1", Math},
		{"superscript", "x^2", Math},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := ParseMath(tt.input)
			if node == nil {
				t.Fatal("ParseMath returned nil")
			}
			if node.Kind() != tt.kind {
				t.Errorf("ParseMath(%q).Kind() = %v, want %v", tt.input, node.Kind(), tt.kind)
			}
		})
	}
}

// TestParseBinaryExpression tests parsing binary expressions.
func TestParseBinaryExpression(t *testing.T) {
	input := "1 + 2"
	node := ParseCode(input)
	if node == nil {
		t.Fatal("ParseCode returned nil")
	}

	// Check that we have a Code node
	if node.Kind() != Code {
		t.Fatalf("Expected Code, got %v", node.Kind())
	}

	// Find the Binary expression
	var foundBinary bool
	for _, child := range node.Children() {
		if child.Kind() == Binary {
			foundBinary = true
			break
		}
	}

	if !foundBinary {
		t.Error("Expected to find a Binary expression")
	}
}

// TestParseConditional tests parsing if-else expressions.
func TestParseConditional(t *testing.T) {
	input := "if true { 1 } else { 2 }"
	node := ParseCode(input)
	if node == nil {
		t.Fatal("ParseCode returned nil")
	}

	// Find the Conditional expression
	var foundConditional bool
	for _, child := range node.Children() {
		if child.Kind() == Conditional {
			foundConditional = true
			break
		}
	}

	if !foundConditional {
		t.Error("Expected to find a Conditional expression")
	}
}

// TestParseEmbeddedCode tests parsing code embedded in markup.
func TestParseEmbeddedCode(t *testing.T) {
	input := "Hello #name!"
	node := Parse(input)
	if node == nil {
		t.Fatal("Parse returned nil")
	}

	// Check that we have a Markup node with embedded code
	if node.Kind() != Markup {
		t.Fatalf("Expected Markup, got %v", node.Kind())
	}

	// Should have children (Text, Hash, Ident, etc.)
	if len(node.Children()) == 0 {
		t.Error("Expected children in markup node")
	}
}

// TestParseEquation tests parsing inline math equations.
func TestParseEquation(t *testing.T) {
	input := "The equation $x^2$ is simple."
	node := Parse(input)
	if node == nil {
		t.Fatal("Parse returned nil")
	}

	// Find the Equation
	var foundEquation bool
	for _, child := range node.Children() {
		if child.Kind() == Equation {
			foundEquation = true
			break
		}
	}

	if !foundEquation {
		t.Error("Expected to find an Equation")
	}
}

// TestParseList tests parsing bullet lists.
func TestParseList(t *testing.T) {
	input := "- Item 1\n- Item 2"
	node := Parse(input)
	if node == nil {
		t.Fatal("Parse returned nil")
	}

	// Find list items
	var foundListItem bool
	for _, child := range node.Children() {
		if child.Kind() == ListItem {
			foundListItem = true
			break
		}
	}

	if !foundListItem {
		t.Error("Expected to find a ListItem")
	}
}

// TestParseHeading tests parsing headings.
func TestParseHeading(t *testing.T) {
	input := "= Heading"
	node := Parse(input)
	if node == nil {
		t.Fatal("Parse returned nil")
	}

	// Find heading
	var foundHeading bool
	for _, child := range node.Children() {
		if child.Kind() == Heading {
			foundHeading = true
			break
		}
	}

	if !foundHeading {
		t.Error("Expected to find a Heading")
	}
}

// TestParseArray tests parsing arrays.
func TestParseArray(t *testing.T) {
	input := "(1, 2, 3)"
	node := ParseCode(input)
	if node == nil {
		t.Fatal("ParseCode returned nil")
	}

	// Find the Array
	var foundArray bool
	for _, child := range node.Children() {
		if child.Kind() == Array {
			foundArray = true
			break
		}
	}

	if !foundArray {
		t.Error("Expected to find an Array")
	}
}

// TestParseDict tests parsing dictionaries.
func TestParseDict(t *testing.T) {
	input := "(a: 1, b: 2)"
	node := ParseCode(input)
	if node == nil {
		t.Fatal("ParseCode returned nil")
	}

	// Find the Dict
	var foundDict bool
	for _, child := range node.Children() {
		if child.Kind() == Dict {
			foundDict = true
			break
		}
	}

	if !foundDict {
		t.Error("Expected to find a Dict")
	}
}

// TestParseFunctionCall tests parsing function calls.
func TestParseFunctionCall(t *testing.T) {
	input := "foo(1, 2)"
	node := ParseCode(input)
	if node == nil {
		t.Fatal("ParseCode returned nil")
	}

	// Find the FuncCall
	var foundFuncCall bool
	for _, child := range node.Children() {
		if child.Kind() == FuncCall {
			foundFuncCall = true
			break
		}
	}

	if !foundFuncCall {
		t.Error("Expected to find a FuncCall")
	}
}

// TestParseClosure tests parsing closures.
func TestParseClosure(t *testing.T) {
	input := "x => x + 1"
	node := ParseCode(input)
	if node == nil {
		t.Fatal("ParseCode returned nil")
	}

	// Find the Closure
	var foundClosure bool
	for _, child := range node.Children() {
		if child.Kind() == Closure {
			foundClosure = true
			break
		}
	}

	if !foundClosure {
		t.Error("Expected to find a Closure")
	}
}

// TestParseForLoop tests parsing for loops.
func TestParseForLoop(t *testing.T) {
	input := "for x in (1, 2, 3) { x }"
	node := ParseCode(input)
	if node == nil {
		t.Fatal("ParseCode returned nil")
	}

	// Find the ForLoop
	var foundForLoop bool
	for _, child := range node.Children() {
		if child.Kind() == ForLoop {
			foundForLoop = true
			break
		}
	}

	if !foundForLoop {
		t.Error("Expected to find a ForLoop")
	}
}

// TestParseWhileLoop tests parsing while loops.
func TestParseWhileLoop(t *testing.T) {
	input := "while true { break }"
	node := ParseCode(input)
	if node == nil {
		t.Fatal("ParseCode returned nil")
	}

	// Find the WhileLoop
	var foundWhileLoop bool
	for _, child := range node.Children() {
		if child.Kind() == WhileLoop {
			foundWhileLoop = true
			break
		}
	}

	if !foundWhileLoop {
		t.Error("Expected to find a WhileLoop")
	}
}
