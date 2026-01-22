package syntax

// Pattern represents a binding pattern in the AST.
// Patterns are used in let bindings, for loops, and closures.
type Pattern interface {
	AstNode
	isPattern()
	// Bindings returns all identifiers bound by this pattern.
	Bindings() []*IdentExpr
}

// PatternFromNode attempts to create a Pattern from an untyped syntax node.
func PatternFromNode(node *SyntaxNode) Pattern {
	if node == nil {
		return nil
	}

	// Check for destructuring pattern
	if child := node.CastFirst(Destructuring); child != nil {
		return &DestructuringPattern{node: child}
	}

	// Check for parenthesized pattern
	if child := node.CastFirst(Parenthesized); child != nil {
		return &ParenthesizedPattern{node: child}
	}

	// Check for placeholder (underscore)
	if child := node.CastFirst(Underscore); child != nil {
		return &PlaceholderPattern{node: child}
	}

	// Check for normal identifier pattern
	if child := node.CastFirst(Ident); child != nil {
		return &NormalPattern{node: child}
	}

	// The node itself might be the pattern
	switch node.Kind() {
	case Destructuring:
		return &DestructuringPattern{node: node}
	case Parenthesized:
		return &ParenthesizedPattern{node: node}
	case Underscore:
		return &PlaceholderPattern{node: node}
	case Ident:
		return &NormalPattern{node: node}
	}

	return nil
}

// NormalPattern represents a simple identifier pattern: x.
type NormalPattern struct {
	node *SyntaxNode
}

func (p *NormalPattern) Kind() SyntaxKind      { return Ident }
func (p *NormalPattern) ToUntyped() *SyntaxNode { return p.node }
func (p *NormalPattern) isAstNode()            {}
func (p *NormalPattern) isPattern()            {}

// Name returns the identifier name.
func (p *NormalPattern) Name() string {
	return p.node.Text()
}

// Bindings returns the single identifier bound by this pattern.
func (p *NormalPattern) Bindings() []*IdentExpr {
	return []*IdentExpr{{node: p.node}}
}

// NormalPatternFromNode casts a syntax node to a NormalPattern.
func NormalPatternFromNode(node *SyntaxNode) *NormalPattern {
	if node == nil || node.Kind() != Ident {
		return nil
	}
	return &NormalPattern{node: node}
}

// PlaceholderPattern represents a placeholder pattern: _.
type PlaceholderPattern struct {
	node *SyntaxNode
}

func (p *PlaceholderPattern) Kind() SyntaxKind      { return Underscore }
func (p *PlaceholderPattern) ToUntyped() *SyntaxNode { return p.node }
func (p *PlaceholderPattern) isAstNode()            {}
func (p *PlaceholderPattern) isPattern()            {}

// PlaceholderPatternFromNode casts a syntax node to a PlaceholderPattern.
func PlaceholderPatternFromNode(node *SyntaxNode) *PlaceholderPattern {
	if node == nil || node.Kind() != Underscore {
		return nil
	}
	return &PlaceholderPattern{node: node}
}

// Bindings returns no identifiers (placeholder binds nothing).
func (p *PlaceholderPattern) Bindings() []*IdentExpr {
	return nil
}

// ParenthesizedPattern represents a parenthesized pattern: (x).
type ParenthesizedPattern struct {
	node *SyntaxNode
}

func (p *ParenthesizedPattern) Kind() SyntaxKind      { return Parenthesized }
func (p *ParenthesizedPattern) ToUntyped() *SyntaxNode { return p.node }
func (p *ParenthesizedPattern) isAstNode()            {}
func (p *ParenthesizedPattern) isPattern()            {}

// Pattern returns the inner pattern.
func (p *ParenthesizedPattern) Pattern() Pattern {
	for _, child := range p.node.Children() {
		if child.Kind() != LeftParen && child.Kind() != RightParen {
			return PatternFromNode(child)
		}
	}
	return nil
}

// ParenthesizedPatternFromNode casts a syntax node to a ParenthesizedPattern.
func ParenthesizedPatternFromNode(node *SyntaxNode) *ParenthesizedPattern {
	if node == nil || node.Kind() != Parenthesized {
		return nil
	}
	return &ParenthesizedPattern{node: node}
}

// Bindings returns the identifiers bound by the inner pattern.
func (p *ParenthesizedPattern) Bindings() []*IdentExpr {
	inner := p.Pattern()
	if inner == nil {
		return nil
	}
	return inner.Bindings()
}

// DestructuringPattern represents a destructuring pattern: (a, b).
type DestructuringPattern struct {
	node *SyntaxNode
}

func (p *DestructuringPattern) Kind() SyntaxKind      { return Destructuring }
func (p *DestructuringPattern) ToUntyped() *SyntaxNode { return p.node }
func (p *DestructuringPattern) isAstNode()            {}
func (p *DestructuringPattern) isPattern()            {}

// Items returns the destructuring items.
func (p *DestructuringPattern) Items() []DestructuringItem {
	var items []DestructuringItem
	for _, child := range p.node.Children() {
		item := DestructuringItemFromNode(child)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

// DestructuringPatternFromNode casts a syntax node to a DestructuringPattern.
func DestructuringPatternFromNode(node *SyntaxNode) *DestructuringPattern {
	if node == nil || node.Kind() != Destructuring {
		return nil
	}
	return &DestructuringPattern{node: node}
}

// Bindings returns all identifiers bound by this destructuring pattern.
func (p *DestructuringPattern) Bindings() []*IdentExpr {
	var result []*IdentExpr
	for _, item := range p.Items() {
		switch i := item.(type) {
		case *DestructuringBinding:
			if pat := i.Pattern(); pat != nil {
				result = append(result, pat.Bindings()...)
			}
		case *DestructuringNamed:
			if pat := i.Pattern(); pat != nil {
				result = append(result, pat.Bindings()...)
			}
		case *DestructuringSpread:
			if sink := i.Sink(); sink != nil {
				result = append(result, sink.Bindings()...)
			}
		}
	}
	return result
}

// DestructuringItem represents an item in a destructuring pattern.
type DestructuringItem interface {
	isDestructuringItem()
}

// DestructuringItemFromNode creates a DestructuringItem from a syntax node.
func DestructuringItemFromNode(node *SyntaxNode) DestructuringItem {
	if node == nil {
		return nil
	}
	switch node.Kind() {
	case Spread:
		return &DestructuringSpread{node: node}
	case Named:
		return &DestructuringNamed{node: node}
	case Ident, Underscore, Parenthesized, Destructuring:
		pattern := PatternFromNode(node)
		if pattern != nil {
			return &DestructuringBinding{pattern: pattern}
		}
	}
	return nil
}

// DestructuringBinding represents a simple binding item: x.
type DestructuringBinding struct {
	pattern Pattern
}

func (d *DestructuringBinding) isDestructuringItem() {}

// Pattern returns the binding pattern.
func (d *DestructuringBinding) Pattern() Pattern {
	return d.pattern
}

// DestructuringNamed represents a named binding item: name: pattern.
type DestructuringNamed struct {
	node *SyntaxNode
}

func (d *DestructuringNamed) isDestructuringItem() {}

// Name returns the field name.
func (d *DestructuringNamed) Name() *IdentExpr {
	child := d.node.CastFirst(Ident)
	if child != nil {
		return &IdentExpr{node: child}
	}
	return nil
}

// Pattern returns the pattern to bind to.
func (d *DestructuringNamed) Pattern() Pattern {
	children := d.node.Children()
	foundColon := false
	for _, child := range children {
		if child.Kind() == Colon {
			foundColon = true
			continue
		}
		if foundColon {
			return PatternFromNode(child)
		}
	}
	return nil
}

// DestructuringSpread represents a spread item: ..rest.
type DestructuringSpread struct {
	node *SyntaxNode
}

func (d *DestructuringSpread) isDestructuringItem() {}

// Sink returns the optional sink pattern.
func (d *DestructuringSpread) Sink() Pattern {
	for _, child := range d.node.Children() {
		if child.Kind() != Dots {
			return PatternFromNode(child)
		}
	}
	return nil
}
