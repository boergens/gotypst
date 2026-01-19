package syntax

import "testing"

func TestPatternFromNode(t *testing.T) {
	// Test normal pattern (identifier)
	identNode := NewSyntaxNode(Ident, "x", nil, Span{})
	pattern := PatternFromNode(identNode)
	if pattern == nil {
		t.Error("PatternFromNode(Ident) should not return nil")
	}
	if _, ok := pattern.(*NormalPattern); !ok {
		t.Errorf("Expected NormalPattern, got %T", pattern)
	}

	// Test placeholder pattern
	underscoreNode := NewSyntaxNode(Underscore, "_", nil, Span{})
	pattern = PatternFromNode(underscoreNode)
	if pattern == nil {
		t.Error("PatternFromNode(Underscore) should not return nil")
	}
	if _, ok := pattern.(*PlaceholderPattern); !ok {
		t.Errorf("Expected PlaceholderPattern, got %T", pattern)
	}

	// Test nil node
	if PatternFromNode(nil) != nil {
		t.Error("PatternFromNode(nil) should return nil")
	}
}

func TestNormalPattern(t *testing.T) {
	node := NewSyntaxNode(Ident, "myVar", nil, Span{})
	pattern := NormalPatternFromNode(node)

	if pattern == nil {
		t.Fatal("NormalPatternFromNode should not return nil")
	}

	if pattern.Name() != "myVar" {
		t.Errorf("Name() = %q, want %q", pattern.Name(), "myVar")
	}

	if pattern.Kind() != Ident {
		t.Errorf("Kind() = %v, want Ident", pattern.Kind())
	}

	// Test wrong kind
	wrongNode := NewSyntaxNode(Underscore, "_", nil, Span{})
	if NormalPatternFromNode(wrongNode) != nil {
		t.Error("NormalPatternFromNode should return nil for non-Ident node")
	}
}

func TestPlaceholderPattern(t *testing.T) {
	node := NewSyntaxNode(Underscore, "_", nil, Span{})
	pattern := PlaceholderPatternFromNode(node)

	if pattern == nil {
		t.Fatal("PlaceholderPatternFromNode should not return nil")
	}

	if pattern.Kind() != Underscore {
		t.Errorf("Kind() = %v, want Underscore", pattern.Kind())
	}

	// Test wrong kind
	wrongNode := NewSyntaxNode(Ident, "x", nil, Span{})
	if PlaceholderPatternFromNode(wrongNode) != nil {
		t.Error("PlaceholderPatternFromNode should return nil for non-Underscore node")
	}
}

func TestParenthesizedPattern(t *testing.T) {
	innerNode := NewSyntaxNode(Ident, "x", nil, Span{1, 2})
	leftParen := NewSyntaxNode(LeftParen, "(", nil, Span{0, 1})
	rightParen := NewSyntaxNode(RightParen, ")", nil, Span{2, 3})
	parenNode := NewSyntaxNode(Parenthesized, "", []*SyntaxNode{leftParen, innerNode, rightParen}, Span{0, 3})

	pattern := ParenthesizedPatternFromNode(parenNode)
	if pattern == nil {
		t.Fatal("ParenthesizedPatternFromNode should not return nil")
	}

	inner := pattern.Pattern()
	if inner == nil {
		t.Fatal("Pattern() should return inner pattern")
	}

	if normal, ok := inner.(*NormalPattern); !ok || normal.Name() != "x" {
		t.Error("Inner pattern should be NormalPattern for 'x'")
	}

	// Test wrong kind
	wrongNode := NewSyntaxNode(Ident, "x", nil, Span{})
	if ParenthesizedPatternFromNode(wrongNode) != nil {
		t.Error("ParenthesizedPatternFromNode should return nil for non-Parenthesized node")
	}
}

func TestDestructuringPattern(t *testing.T) {
	aNode := NewSyntaxNode(Ident, "a", nil, Span{1, 2})
	bNode := NewSyntaxNode(Ident, "b", nil, Span{4, 5})
	destructNode := NewSyntaxNode(Destructuring, "", []*SyntaxNode{aNode, bNode}, Span{0, 6})

	pattern := DestructuringPatternFromNode(destructNode)
	if pattern == nil {
		t.Fatal("DestructuringPatternFromNode should not return nil")
	}

	items := pattern.Items()
	if len(items) != 2 {
		t.Errorf("Items() should return 2 items, got %d", len(items))
	}

	// Test wrong kind
	wrongNode := NewSyntaxNode(Ident, "x", nil, Span{})
	if DestructuringPatternFromNode(wrongNode) != nil {
		t.Error("DestructuringPatternFromNode should return nil for non-Destructuring node")
	}
}

func TestDestructuringItemFromNode(t *testing.T) {
	// Test binding item (identifier)
	identNode := NewSyntaxNode(Ident, "x", nil, Span{})
	item := DestructuringItemFromNode(identNode)
	if item == nil {
		t.Error("DestructuringItemFromNode(Ident) should not return nil")
	}
	if _, ok := item.(*DestructuringBinding); !ok {
		t.Errorf("Expected DestructuringBinding, got %T", item)
	}

	// Test spread item
	dotsNode := NewSyntaxNode(Dots, "..", nil, Span{0, 2})
	restNode := NewSyntaxNode(Ident, "rest", nil, Span{2, 6})
	spreadNode := NewSyntaxNode(Spread, "", []*SyntaxNode{dotsNode, restNode}, Span{0, 6})
	item = DestructuringItemFromNode(spreadNode)
	if item == nil {
		t.Error("DestructuringItemFromNode(Spread) should not return nil")
	}
	if _, ok := item.(*DestructuringSpread); !ok {
		t.Errorf("Expected DestructuringSpread, got %T", item)
	}

	// Test named item
	nameNode := NewSyntaxNode(Ident, "field", nil, Span{0, 5})
	colonNode := NewSyntaxNode(Colon, ":", nil, Span{5, 6})
	valueNode := NewSyntaxNode(Ident, "x", nil, Span{7, 8})
	namedNode := NewSyntaxNode(Named, "", []*SyntaxNode{nameNode, colonNode, valueNode}, Span{0, 8})
	item = DestructuringItemFromNode(namedNode)
	if item == nil {
		t.Error("DestructuringItemFromNode(Named) should not return nil")
	}
	if _, ok := item.(*DestructuringNamed); !ok {
		t.Errorf("Expected DestructuringNamed, got %T", item)
	}

	// Test nil
	if DestructuringItemFromNode(nil) != nil {
		t.Error("DestructuringItemFromNode(nil) should return nil")
	}
}

func TestDestructuringBinding(t *testing.T) {
	identNode := NewSyntaxNode(Ident, "x", nil, Span{})
	pattern := NormalPatternFromNode(identNode)
	binding := &DestructuringBinding{pattern: pattern}

	if binding.Pattern() != pattern {
		t.Error("Pattern() should return the bound pattern")
	}
}

func TestDestructuringNamed(t *testing.T) {
	nameNode := NewSyntaxNode(Ident, "field", nil, Span{0, 5})
	colonNode := NewSyntaxNode(Colon, ":", nil, Span{5, 6})
	valueNode := NewSyntaxNode(Ident, "x", nil, Span{7, 8})
	namedNode := NewSyntaxNode(Named, "", []*SyntaxNode{nameNode, colonNode, valueNode}, Span{0, 8})

	named := &DestructuringNamed{node: namedNode}

	name := named.Name()
	if name == nil || name.Get() != "field" {
		t.Error("Name() should return 'field'")
	}

	pattern := named.Pattern()
	if pattern == nil {
		t.Error("Pattern() should not return nil")
	}
}

func TestDestructuringSpread(t *testing.T) {
	dotsNode := NewSyntaxNode(Dots, "..", nil, Span{0, 2})
	restNode := NewSyntaxNode(Ident, "rest", nil, Span{2, 6})
	spreadNode := NewSyntaxNode(Spread, "", []*SyntaxNode{dotsNode, restNode}, Span{0, 6})

	spread := &DestructuringSpread{node: spreadNode}

	sink := spread.Sink()
	if sink == nil {
		t.Error("Sink() should not return nil when there's a sink pattern")
	}
	if normal, ok := sink.(*NormalPattern); !ok || normal.Name() != "rest" {
		t.Error("Sink() should return NormalPattern for 'rest'")
	}

	// Test spread without sink
	justDots := NewSyntaxNode(Spread, "", []*SyntaxNode{dotsNode}, Span{0, 2})
	spreadNoSink := &DestructuringSpread{node: justDots}
	if spreadNoSink.Sink() != nil {
		t.Error("Sink() should return nil when there's no sink pattern")
	}
}

func TestPatternInterface(t *testing.T) {
	// All pattern types should implement Pattern interface
	patterns := []Pattern{
		&NormalPattern{node: NewSyntaxNode(Ident, "", nil, Span{})},
		&PlaceholderPattern{node: NewSyntaxNode(Underscore, "", nil, Span{})},
		&ParenthesizedPattern{node: NewSyntaxNode(Parenthesized, "", nil, Span{})},
		&DestructuringPattern{node: NewSyntaxNode(Destructuring, "", nil, Span{})},
	}

	for _, p := range patterns {
		// All Pattern types should be AstNode
		var _ AstNode = p

		if p.ToUntyped() == nil {
			t.Errorf("%T.ToUntyped() returned nil", p)
		}
	}
}
