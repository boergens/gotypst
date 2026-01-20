package syntax

// ----------------------------------------------------------------------------
// Argument Types
// ----------------------------------------------------------------------------

// Arg represents an argument in a function call.
type Arg interface {
	isArg()
}

// ArgFromNode creates an Arg from a syntax node.
func ArgFromNode(node *SyntaxNode) Arg {
	if node == nil {
		return nil
	}
	// Skip trivia and delimiters - they're not arguments
	switch node.Kind() {
	case Space, Linebreak, Parbreak, Comma, LeftParen, RightParen:
		return nil
	case Spread:
		return &SpreadArg{node: node}
	case Named:
		return &NamedArg{node: node}
	default:
		expr := ExprFromNode(node)
		if expr != nil {
			return &PosArg{expr: expr}
		}
	}
	return nil
}

// PosArg represents a positional argument: f(x).
type PosArg struct {
	expr Expr
}

func (a *PosArg) isArg() {}

// Expr returns the argument expression.
func (a *PosArg) Expr() Expr {
	return a.expr
}

// NamedArg represents a named argument: f(name: value).
type NamedArg struct {
	node *SyntaxNode
}

func (a *NamedArg) isArg() {}

// Name returns the argument name.
func (a *NamedArg) Name() *IdentExpr {
	child := a.node.CastFirst(Ident)
	if child != nil {
		return &IdentExpr{node: child}
	}
	return nil
}

// Expr returns the argument value.
func (a *NamedArg) Expr() Expr {
	children := a.node.Children()
	foundColon := false
	for _, child := range children {
		if child.Kind() == Colon {
			foundColon = true
			continue
		}
		if foundColon {
			// Skip trivia - find the actual value
			switch child.Kind() {
			case Space, Linebreak, Parbreak:
				continue
			}
			return ExprFromNode(child)
		}
	}
	return nil
}

// SpreadArg represents a spread argument: f(..args).
type SpreadArg struct {
	node *SyntaxNode
}

func (a *SpreadArg) isArg() {}

// Expr returns the spread expression.
func (a *SpreadArg) Expr() Expr {
	for _, child := range a.node.Children() {
		if child.Kind() != Dots {
			return ExprFromNode(child)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Array Item Types
// ----------------------------------------------------------------------------

// ArrayItem represents an item in an array literal.
type ArrayItem interface {
	isArrayItem()
}

// ArrayItemFromNode creates an ArrayItem from a syntax node.
func ArrayItemFromNode(node *SyntaxNode) ArrayItem {
	if node == nil {
		return nil
	}
	// Skip trivia and delimiters - they're not array items
	switch node.Kind() {
	case Space, Linebreak, Parbreak, Comma, LeftParen, RightParen:
		return nil
	case Spread:
		return &ArraySpreadItem{node: node}
	default:
		expr := ExprFromNode(node)
		if expr != nil {
			return &ArrayPosItem{expr: expr}
		}
	}
	return nil
}

// ArrayPosItem represents a positional array item: (1, 2, 3).
type ArrayPosItem struct {
	expr Expr
}

func (i *ArrayPosItem) isArrayItem() {}

// Expr returns the item expression.
func (i *ArrayPosItem) Expr() Expr {
	return i.expr
}

// ArraySpreadItem represents a spread array item: (..items).
type ArraySpreadItem struct {
	node *SyntaxNode
}

func (i *ArraySpreadItem) isArrayItem() {}

// Expr returns the spread expression.
func (i *ArraySpreadItem) Expr() Expr {
	for _, child := range i.node.Children() {
		if child.Kind() != Dots {
			return ExprFromNode(child)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Dict Item Types
// ----------------------------------------------------------------------------

// DictItem represents an item in a dictionary literal.
type DictItem interface {
	isDictItem()
}

// DictItemFromNode creates a DictItem from a syntax node.
func DictItemFromNode(node *SyntaxNode) DictItem {
	if node == nil {
		return nil
	}
	switch node.Kind() {
	case Spread:
		return &DictSpreadItem{node: node}
	case Named:
		return &DictNamedItem{node: node}
	case Keyed:
		return &DictKeyedItem{node: node}
	}
	return nil
}

// DictNamedItem represents a named dict item: (a: 1).
type DictNamedItem struct {
	node *SyntaxNode
}

func (i *DictNamedItem) isDictItem() {}

// Name returns the key name.
func (i *DictNamedItem) Name() *IdentExpr {
	child := i.node.CastFirst(Ident)
	if child != nil {
		return &IdentExpr{node: child}
	}
	return nil
}

// Expr returns the value expression.
func (i *DictNamedItem) Expr() Expr {
	children := i.node.Children()
	foundColon := false
	for _, child := range children {
		if child.Kind() == Colon {
			foundColon = true
			continue
		}
		if foundColon {
			return ExprFromNode(child)
		}
	}
	return nil
}

// DictKeyedItem represents a keyed dict item: ("key": value).
type DictKeyedItem struct {
	node *SyntaxNode
}

func (i *DictKeyedItem) isDictItem() {}

// Key returns the key expression.
func (i *DictKeyedItem) Key() Expr {
	children := i.node.Children()
	if len(children) > 0 {
		return ExprFromNode(children[0])
	}
	return nil
}

// Expr returns the value expression.
func (i *DictKeyedItem) Expr() Expr {
	children := i.node.Children()
	foundColon := false
	for _, child := range children {
		if child.Kind() == Colon {
			foundColon = true
			continue
		}
		if foundColon {
			return ExprFromNode(child)
		}
	}
	return nil
}

// DictSpreadItem represents a spread dict item: (..other).
type DictSpreadItem struct {
	node *SyntaxNode
}

func (i *DictSpreadItem) isDictItem() {}

// Expr returns the spread expression.
func (i *DictSpreadItem) Expr() Expr {
	for _, child := range i.node.Children() {
		if child.Kind() != Dots {
			return ExprFromNode(child)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Parameter Types
// ----------------------------------------------------------------------------

// Param represents a parameter in a function definition.
type Param interface {
	isParam()
}

// ParamFromNode creates a Param from a syntax node.
func ParamFromNode(node *SyntaxNode) Param {
	if node == nil {
		return nil
	}
	switch node.Kind() {
	case Spread:
		return &SinkParam{node: node}
	case Named:
		return &NamedParam{node: node}
	case Destructuring:
		return &DestructuringParam{node: node}
	case Ident:
		return &PosParam{node: node}
	case Underscore:
		return &PlaceholderParam{node: node}
	}
	return nil
}

// PosParam represents a positional parameter: (x).
type PosParam struct {
	node *SyntaxNode
}

func (p *PosParam) isParam() {}

// Name returns the parameter name.
func (p *PosParam) Name() *IdentExpr {
	return &IdentExpr{node: p.node}
}

// PlaceholderParam represents a placeholder parameter: (_).
type PlaceholderParam struct {
	node *SyntaxNode
}

func (p *PlaceholderParam) isParam() {}

// NamedParam represents a named parameter with default: (x: 1).
type NamedParam struct {
	node *SyntaxNode
}

func (p *NamedParam) isParam() {}

// Name returns the parameter name.
func (p *NamedParam) Name() *IdentExpr {
	child := p.node.CastFirst(Ident)
	if child != nil {
		return &IdentExpr{node: child}
	}
	return nil
}

// Default returns the default value expression.
func (p *NamedParam) Default() Expr {
	children := p.node.Children()
	foundColon := false
	for _, child := range children {
		if child.Kind() == Colon {
			foundColon = true
			continue
		}
		if foundColon {
			return ExprFromNode(child)
		}
	}
	return nil
}

// SinkParam represents a sink parameter: (..rest).
type SinkParam struct {
	node *SyntaxNode
}

func (p *SinkParam) isParam() {}

// Name returns the optional sink name.
func (p *SinkParam) Name() *IdentExpr {
	child := p.node.CastFirst(Ident)
	if child != nil {
		return &IdentExpr{node: child}
	}
	return nil
}

// DestructuringParam represents a destructuring parameter: ((a, b)).
type DestructuringParam struct {
	node *SyntaxNode
}

func (p *DestructuringParam) isParam() {}

// Pattern returns the destructuring pattern.
func (p *DestructuringParam) Pattern() *DestructuringNode {
	return &DestructuringNode{node: p.node}
}

// ----------------------------------------------------------------------------
// Import Types
// ----------------------------------------------------------------------------

// Imports represents the items being imported.
type Imports interface {
	isImports()
}

// ImportsWildcard represents a wildcard import: import "x": *.
type ImportsWildcard struct{}

func (i *ImportsWildcard) isImports() {}

// ImportItemsNode represents explicit import items.
type ImportItemsNode struct {
	node *SyntaxNode
}

func (i *ImportItemsNode) isImports() {}

// Items returns the import items.
func (i *ImportItemsNode) Items() []*ImportItem {
	var items []*ImportItem
	for _, child := range i.node.Children() {
		if child.Kind() == ImportItemPath || child.Kind() == RenamedImportItem {
			items = append(items, &ImportItem{node: child})
		}
	}
	return items
}

// ImportItem represents a single import item.
type ImportItem struct {
	node *SyntaxNode
}

// Path returns the import path.
func (i *ImportItem) Path() []string {
	var path []string
	for _, child := range i.node.Children() {
		if child.Kind() == Ident {
			path = append(path, child.Text())
		}
	}
	return path
}

// NewName returns the rename, if any.
func (i *ImportItem) NewName() *IdentExpr {
	if i.node.Kind() == RenamedImportItem {
		children := i.node.Children()
		for j, child := range children {
			if child.Kind() == As && j+1 < len(children) {
				nextChild := children[j+1]
				if nextChild.Kind() == Ident {
					return &IdentExpr{node: nextChild}
				}
			}
		}
	}
	return nil
}
