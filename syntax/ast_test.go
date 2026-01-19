package syntax

import "testing"

// TestSyntaxNodeBasics tests the basic SyntaxNode operations.
func TestSyntaxNodeBasics(t *testing.T) {
	// Test nil node
	var nilNode *SyntaxNode
	if nilNode.Kind() != Error {
		t.Errorf("nil node Kind() should return Error, got %v", nilNode.Kind())
	}
	if nilNode.Text() != "" {
		t.Errorf("nil node Text() should return empty string")
	}
	if nilNode.Children() != nil {
		t.Errorf("nil node Children() should return nil")
	}

	// Test basic node
	node := NewSyntaxNode(Ident, "foo", nil, Span{0, 3})
	if node.Kind() != Ident {
		t.Errorf("Expected Ident, got %v", node.Kind())
	}
	if node.Text() != "foo" {
		t.Errorf("Expected 'foo', got %q", node.Text())
	}
	if node.Span().Start != 0 || node.Span().End != 3 {
		t.Errorf("Unexpected span: %v", node.Span())
	}
}

// TestSyntaxNodeCast tests the cast operations on SyntaxNode.
func TestSyntaxNodeCast(t *testing.T) {
	child1 := NewSyntaxNode(Ident, "x", nil, Span{0, 1})
	child2 := NewSyntaxNode(Plus, "+", nil, Span{2, 3})
	child3 := NewSyntaxNode(Ident, "y", nil, Span{4, 5})

	parent := NewSyntaxNode(Binary, "", []*SyntaxNode{child1, child2, child3}, Span{0, 5})

	// Test Cast
	if parent.Cast(Binary) == nil {
		t.Error("Cast to Binary should succeed")
	}
	if parent.Cast(Unary) != nil {
		t.Error("Cast to Unary should fail")
	}

	// Test CastFirst
	first := parent.CastFirst(Ident)
	if first == nil || first.Text() != "x" {
		t.Error("CastFirst(Ident) should return first Ident child")
	}
	if parent.CastFirst(Str) != nil {
		t.Error("CastFirst(Str) should return nil")
	}

	// Test CastAll
	idents := parent.CastAll(Ident)
	if len(idents) != 2 {
		t.Errorf("Expected 2 Ident children, got %d", len(idents))
	}
}

// TestExprFromNode tests expression creation from nodes.
func TestExprFromNode(t *testing.T) {
	tests := []struct {
		kind     SyntaxKind
		wantKind SyntaxKind
		wantNil  bool
	}{
		{Text, Text, false},
		{Ident, Ident, false},
		{Int, Int, false},
		{Binary, Binary, false},
		{FuncCall, FuncCall, false},
		{Conditional, Conditional, false},
		{End, Error, true}, // Not an expression kind
	}

	for _, tt := range tests {
		node := NewSyntaxNode(tt.kind, "", nil, Span{})
		expr := ExprFromNode(node)

		if tt.wantNil {
			if expr != nil {
				t.Errorf("ExprFromNode(%v): expected nil, got %v", tt.kind, expr)
			}
		} else {
			if expr == nil {
				t.Errorf("ExprFromNode(%v): expected non-nil", tt.kind)
			} else if expr.Kind() != tt.wantKind {
				t.Errorf("ExprFromNode(%v): expected kind %v, got %v", tt.kind, tt.wantKind, expr.Kind())
			}
		}
	}

	// Test nil node
	if ExprFromNode(nil) != nil {
		t.Error("ExprFromNode(nil) should return nil")
	}
}

// TestTextExpr tests TextExpr functionality.
func TestTextExpr(t *testing.T) {
	node := NewSyntaxNode(Text, "hello world", nil, Span{0, 11})
	expr := TextExprFromNode(node)

	if expr == nil {
		t.Fatal("TextExprFromNode should not return nil")
	}
	if expr.Get() != "hello world" {
		t.Errorf("Get() = %q, want %q", expr.Get(), "hello world")
	}
	if expr.Kind() != Text {
		t.Errorf("Kind() = %v, want %v", expr.Kind(), Text)
	}

	// Test wrong kind
	wrongNode := NewSyntaxNode(Ident, "x", nil, Span{})
	if TextExprFromNode(wrongNode) != nil {
		t.Error("TextExprFromNode should return nil for non-Text node")
	}
}

// TestBoolExpr tests BoolExpr functionality.
func TestBoolExpr(t *testing.T) {
	trueNode := NewSyntaxNode(Bool, "true", nil, Span{})
	falseNode := NewSyntaxNode(Bool, "false", nil, Span{})

	trueExpr := BoolExprFromNode(trueNode)
	if trueExpr == nil || !trueExpr.Get() {
		t.Error("BoolExpr for 'true' should return true")
	}

	falseExpr := BoolExprFromNode(falseNode)
	if falseExpr == nil || falseExpr.Get() {
		t.Error("BoolExpr for 'false' should return false")
	}
}

// TestIntExpr tests IntExpr functionality.
func TestIntExpr(t *testing.T) {
	node := NewSyntaxNode(Int, "42", nil, Span{})
	expr := IntExprFromNode(node)

	if expr == nil {
		t.Fatal("IntExprFromNode should not return nil")
	}
	if expr.Get() != 42 {
		t.Errorf("Get() = %d, want 42", expr.Get())
	}
}

// TestFloatExpr tests FloatExpr functionality.
func TestFloatExpr(t *testing.T) {
	node := NewSyntaxNode(Float, "3.14", nil, Span{})
	expr := FloatExprFromNode(node)

	if expr == nil {
		t.Fatal("FloatExprFromNode should not return nil")
	}
	// Allow for floating point imprecision
	got := expr.Get()
	if got < 3.13 || got > 3.15 {
		t.Errorf("Get() = %f, want approximately 3.14", got)
	}
}

// TestStrExpr tests StrExpr functionality.
func TestStrExpr(t *testing.T) {
	node := NewSyntaxNode(Str, "\"hello\"", nil, Span{})
	expr := StrExprFromNode(node)

	if expr == nil {
		t.Fatal("StrExprFromNode should not return nil")
	}
	if expr.Get() != "hello" {
		t.Errorf("Get() = %q, want %q", expr.Get(), "hello")
	}
}

// TestIdentExpr tests IdentExpr functionality.
func TestIdentExpr(t *testing.T) {
	node := NewSyntaxNode(Ident, "myVar", nil, Span{})
	expr := IdentExprFromNode(node)

	if expr == nil {
		t.Fatal("IdentExprFromNode should not return nil")
	}
	if expr.Get() != "myVar" {
		t.Errorf("Get() = %q, want %q", expr.Get(), "myVar")
	}
}

// TestLabelExpr tests LabelExpr functionality.
func TestLabelExpr(t *testing.T) {
	node := NewSyntaxNode(Label, "<fig:chart>", nil, Span{})
	expr := LabelExprFromNode(node)

	if expr == nil {
		t.Fatal("LabelExprFromNode should not return nil")
	}
	if expr.Get() != "fig:chart" {
		t.Errorf("Get() = %q, want %q", expr.Get(), "fig:chart")
	}
}

// TestSmartQuoteExpr tests SmartQuoteExpr functionality.
func TestSmartQuoteExpr(t *testing.T) {
	singleNode := NewSyntaxNode(SmartQuote, "'", nil, Span{})
	doubleNode := NewSyntaxNode(SmartQuote, "\"", nil, Span{})

	singleExpr := SmartQuoteExprFromNode(singleNode)
	if singleExpr.Double() {
		t.Error("Single quote should not be Double()")
	}

	doubleExpr := SmartQuoteExprFromNode(doubleNode)
	if !doubleExpr.Double() {
		t.Error("Double quote should be Double()")
	}
}

// TestUnaryExpr tests UnaryExpr functionality.
func TestUnaryExpr(t *testing.T) {
	minusNode := NewSyntaxNode(Minus, "-", nil, Span{0, 1})
	numNode := NewSyntaxNode(Int, "5", nil, Span{1, 2})
	unaryNode := NewSyntaxNode(Unary, "", []*SyntaxNode{minusNode, numNode}, Span{0, 2})

	expr := UnaryExprFromNode(unaryNode)
	if expr == nil {
		t.Fatal("UnaryExprFromNode should not return nil")
	}

	if expr.Op() != UnOpNeg {
		t.Errorf("Op() = %v, want UnOpNeg", expr.Op())
	}

	operand := expr.Expr()
	if operand == nil || operand.Kind() != Int {
		t.Error("Expr() should return the operand")
	}
}

// TestBinaryExpr tests BinaryExpr functionality.
func TestBinaryExpr(t *testing.T) {
	lhsNode := NewSyntaxNode(Int, "1", nil, Span{0, 1})
	opNode := NewSyntaxNode(Plus, "+", nil, Span{2, 3})
	rhsNode := NewSyntaxNode(Int, "2", nil, Span{4, 5})
	binaryNode := NewSyntaxNode(Binary, "", []*SyntaxNode{lhsNode, opNode, rhsNode}, Span{0, 5})

	expr := BinaryExprFromNode(binaryNode)
	if expr == nil {
		t.Fatal("BinaryExprFromNode should not return nil")
	}

	if expr.Op() != BinOpAdd {
		t.Errorf("Op() = %v, want BinOpAdd", expr.Op())
	}

	lhs := expr.Lhs()
	if lhs == nil || lhs.Kind() != Int {
		t.Error("Lhs() should return Int")
	}

	rhs := expr.Rhs()
	if rhs == nil || rhs.Kind() != Int {
		t.Error("Rhs() should return Int")
	}
}

// TestHeadingExpr tests HeadingExpr functionality.
func TestHeadingExpr(t *testing.T) {
	markerNode := NewSyntaxNode(HeadingMarker, "==", nil, Span{0, 2})
	markupNode := NewSyntaxNode(Markup, "", nil, Span{3, 10})
	headingNode := NewSyntaxNode(Heading, "", []*SyntaxNode{markerNode, markupNode}, Span{0, 10})

	expr := HeadingExprFromNode(headingNode)
	if expr == nil {
		t.Fatal("HeadingExprFromNode should not return nil")
	}

	if expr.Level() != 2 {
		t.Errorf("Level() = %d, want 2", expr.Level())
	}

	if expr.Body() == nil {
		t.Error("Body() should not return nil")
	}
}

// TestRawExpr tests RawExpr functionality.
func TestRawExpr(t *testing.T) {
	delimNode := NewSyntaxNode(RawDelim, "```", nil, Span{0, 3})
	langNode := NewSyntaxNode(RawLang, "go", nil, Span{3, 5})
	textNode := NewSyntaxNode(Text, "code here", nil, Span{6, 15})
	rawNode := NewSyntaxNode(Raw, "", []*SyntaxNode{delimNode, langNode, textNode}, Span{0, 18})

	expr := RawExprFromNode(rawNode)
	if expr == nil {
		t.Fatal("RawExprFromNode should not return nil")
	}

	if expr.Lang() != "go" {
		t.Errorf("Lang() = %q, want %q", expr.Lang(), "go")
	}

	if !expr.Block() {
		t.Error("Block() should be true for triple backticks")
	}

	lines := expr.Lines()
	if len(lines) != 1 || lines[0] != "code here" {
		t.Errorf("Lines() = %v, want [code here]", lines)
	}
}

// TestNumericExpr tests NumericExpr functionality.
func TestNumericExpr(t *testing.T) {
	tests := []struct {
		text      string
		wantValue float64
		wantUnit  Unit
	}{
		{"12pt", 12, UnitPt},
		{"1.5em", 1.5, UnitEm},
		{"100%", 100, UnitPercent},
		{"3.14rad", 3.14, UnitRad},
	}

	for _, tt := range tests {
		node := NewSyntaxNode(Numeric, tt.text, nil, Span{})
		expr := NumericExprFromNode(node)

		if expr == nil {
			t.Errorf("NumericExprFromNode(%q): should not return nil", tt.text)
			continue
		}

		// Allow for floating point imprecision
		got := expr.Value()
		if got < tt.wantValue-0.01 || got > tt.wantValue+0.01 {
			t.Errorf("NumericExpr(%q).Value() = %f, want %f", tt.text, got, tt.wantValue)
		}

		if expr.Unit() != tt.wantUnit {
			t.Errorf("NumericExpr(%q).Unit() = %v, want %v", tt.text, expr.Unit(), tt.wantUnit)
		}
	}
}

// TestAstNodeInterface tests that all expression types implement AstNode.
func TestAstNodeInterface(t *testing.T) {
	// Create various expressions
	exprs := []Expr{
		&TextExpr{node: NewSyntaxNode(Text, "", nil, Span{})},
		&IdentExpr{node: NewSyntaxNode(Ident, "", nil, Span{})},
		&IntExpr{node: NewSyntaxNode(Int, "", nil, Span{})},
		&BinaryExpr{node: NewSyntaxNode(Binary, "", nil, Span{})},
		&FuncCallExpr{node: NewSyntaxNode(FuncCall, "", nil, Span{})},
	}

	for _, expr := range exprs {
		// All Expr types should be AstNode
		var _ AstNode = expr

		// ToUntyped should return the underlying node
		if expr.ToUntyped() == nil {
			t.Errorf("%T.ToUntyped() returned nil", expr)
		}
	}
}
