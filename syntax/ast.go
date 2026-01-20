package syntax

// AstNode is the interface implemented by all typed AST nodes.
// It provides methods to convert between typed and untyped representations.
type AstNode interface {
	// Kind returns the syntax kind of this node.
	Kind() SyntaxKind

	// ToUntyped returns the underlying untyped syntax node.
	ToUntyped() *SyntaxNode

	// isAstNode is a marker method to ensure only AST nodes implement this interface.
	isAstNode()
}

// ----------------------------------------------------------------------------
// Expr is the interface for all expression types in the AST.
// It extends AstNode with expression-specific behavior.
// ----------------------------------------------------------------------------

// Expr represents any expression in the AST.
// This is a sum type interface - each variant is a separate struct type.
type Expr interface {
	AstNode
	isExpr()
}

// ExprFromNode attempts to create an Expr from an untyped syntax node.
// Returns nil if the node's kind is not a valid expression kind.
func ExprFromNode(node *SyntaxNode) Expr {
	if node == nil {
		return nil
	}
	switch node.Kind() {
	// Markup elements
	case Text:
		return &TextExpr{node: node}
	case Space:
		return &SpaceExpr{node: node}
	case Linebreak:
		return &LinebreakExpr{node: node}
	case Parbreak:
		return &ParbreakExpr{node: node}
	case Escape:
		return &EscapeExpr{node: node}
	case Shorthand:
		return &ShorthandExpr{node: node}
	case SmartQuote:
		return &SmartQuoteExpr{node: node}
	case Strong:
		return &StrongExpr{node: node}
	case Emph:
		return &EmphExpr{node: node}
	case Raw:
		return &RawExpr{node: node}
	case Link:
		return &LinkExpr{node: node}
	case Label:
		return &LabelExpr{node: node}
	case Ref:
		return &RefExpr{node: node}
	case Heading:
		return &HeadingExpr{node: node}
	case ListItem:
		return &ListItemExpr{node: node}
	case EnumItem:
		return &EnumItemExpr{node: node}
	case TermItem:
		return &TermItemExpr{node: node}

	// Math expressions
	case Equation:
		return &EquationExpr{node: node}
	case MathText:
		return &MathTextExpr{node: node}
	case MathIdent:
		return &MathIdentExpr{node: node}
	case MathShorthand:
		return &MathShorthandExpr{node: node}
	case MathAlignPoint:
		return &MathAlignPointExpr{node: node}
	case MathDelimited:
		return &MathDelimitedExpr{node: node}
	case MathAttach:
		return &MathAttachExpr{node: node}
	case MathPrimes:
		return &MathPrimesExpr{node: node}
	case MathFrac:
		return &MathFracExpr{node: node}
	case MathRoot:
		return &MathRootExpr{node: node}

	// Literals
	case None:
		return &NoneExpr{node: node}
	case Auto:
		return &AutoExpr{node: node}
	case Bool:
		return &BoolExpr{node: node}
	case Int:
		return &IntExpr{node: node}
	case Float:
		return &FloatExpr{node: node}
	case Numeric:
		return &NumericExpr{node: node}
	case Str:
		return &StrExpr{node: node}
	case Ident:
		return &IdentExpr{node: node}

	// Collections and grouping
	case Array:
		return &ArrayExpr{node: node}
	case Dict:
		return &DictExpr{node: node}
	case CodeBlock:
		return &CodeBlockExpr{node: node}
	case ContentBlock:
		return &ContentBlockExpr{node: node}
	case Parenthesized:
		return &ParenthesizedExpr{node: node}

	// Operations
	case Unary:
		return &UnaryExpr{node: node}
	case Binary:
		return &BinaryExpr{node: node}
	case FieldAccess:
		return &FieldAccessExpr{node: node}
	case FuncCall:
		return &FuncCallExpr{node: node}

	// Control flow
	case Closure:
		return &ClosureExpr{node: node}
	case LetBinding:
		return &LetBindingExpr{node: node}
	case DestructAssignment:
		return &DestructAssignmentExpr{node: node}
	case SetRule:
		return &SetRuleExpr{node: node}
	case ShowRule:
		return &ShowRuleExpr{node: node}
	case Contextual:
		return &ContextualExpr{node: node}
	case Conditional:
		return &ConditionalExpr{node: node}
	case WhileLoop:
		return &WhileLoopExpr{node: node}
	case ForLoop:
		return &ForLoopExpr{node: node}

	// Modules
	case ModuleImport:
		return &ModuleImportExpr{node: node}
	case ModuleInclude:
		return &ModuleIncludeExpr{node: node}

	// Loop control
	case LoopBreak:
		return &LoopBreakExpr{node: node}
	case LoopContinue:
		return &LoopContinueExpr{node: node}
	case FuncReturn:
		return &FuncReturnExpr{node: node}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Markup Expression Types
// ----------------------------------------------------------------------------

// TextExpr represents plain text content.
type TextExpr struct {
	node *SyntaxNode
}

func (e *TextExpr) Kind() SyntaxKind      { return Text }
func (e *TextExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *TextExpr) isAstNode()            {}
func (e *TextExpr) isExpr()               {}

// Get returns the text content.
func (e *TextExpr) Get() string {
	return e.node.Text()
}

// TextExprFromNode casts a syntax node to a TextExpr.
func TextExprFromNode(node *SyntaxNode) *TextExpr {
	if node == nil || node.Kind() != Text {
		return nil
	}
	return &TextExpr{node: node}
}

// SpaceExpr represents whitespace.
type SpaceExpr struct {
	node *SyntaxNode
}

func (e *SpaceExpr) Kind() SyntaxKind      { return Space }
func (e *SpaceExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *SpaceExpr) isAstNode()            {}
func (e *SpaceExpr) isExpr()               {}

// SpaceExprFromNode casts a syntax node to a SpaceExpr.
func SpaceExprFromNode(node *SyntaxNode) *SpaceExpr {
	if node == nil || node.Kind() != Space {
		return nil
	}
	return &SpaceExpr{node: node}
}

// LinebreakExpr represents a line break (\\).
type LinebreakExpr struct {
	node *SyntaxNode
}

func (e *LinebreakExpr) Kind() SyntaxKind      { return Linebreak }
func (e *LinebreakExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *LinebreakExpr) isAstNode()            {}
func (e *LinebreakExpr) isExpr()               {}

// LinebreakExprFromNode casts a syntax node to a LinebreakExpr.
func LinebreakExprFromNode(node *SyntaxNode) *LinebreakExpr {
	if node == nil || node.Kind() != Linebreak {
		return nil
	}
	return &LinebreakExpr{node: node}
}

// ParbreakExpr represents a paragraph break (blank line).
type ParbreakExpr struct {
	node *SyntaxNode
}

func (e *ParbreakExpr) Kind() SyntaxKind      { return Parbreak }
func (e *ParbreakExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ParbreakExpr) isAstNode()            {}
func (e *ParbreakExpr) isExpr()               {}

// ParbreakExprFromNode casts a syntax node to a ParbreakExpr.
func ParbreakExprFromNode(node *SyntaxNode) *ParbreakExpr {
	if node == nil || node.Kind() != Parbreak {
		return nil
	}
	return &ParbreakExpr{node: node}
}

// EscapeExpr represents an escape sequence like \n.
type EscapeExpr struct {
	node *SyntaxNode
}

func (e *EscapeExpr) Kind() SyntaxKind      { return Escape }
func (e *EscapeExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *EscapeExpr) isAstNode()            {}
func (e *EscapeExpr) isExpr()               {}

// Get returns the escaped character.
func (e *EscapeExpr) Get() rune {
	text := e.node.Text()
	if len(text) >= 2 {
		return rune(text[1])
	}
	return 0
}

// EscapeExprFromNode casts a syntax node to an EscapeExpr.
func EscapeExprFromNode(node *SyntaxNode) *EscapeExpr {
	if node == nil || node.Kind() != Escape {
		return nil
	}
	return &EscapeExpr{node: node}
}

// ShorthandExpr represents a shorthand like ~, ---, or ...
type ShorthandExpr struct {
	node *SyntaxNode
}

func (e *ShorthandExpr) Kind() SyntaxKind      { return Shorthand }
func (e *ShorthandExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ShorthandExpr) isAstNode()            {}
func (e *ShorthandExpr) isExpr()               {}

// Get returns the shorthand character(s).
func (e *ShorthandExpr) Get() string {
	return e.node.Text()
}

// ShorthandExprFromNode casts a syntax node to a ShorthandExpr.
func ShorthandExprFromNode(node *SyntaxNode) *ShorthandExpr {
	if node == nil || node.Kind() != Shorthand {
		return nil
	}
	return &ShorthandExpr{node: node}
}

// SmartQuoteExpr represents a smart quote (' or ").
type SmartQuoteExpr struct {
	node *SyntaxNode
}

func (e *SmartQuoteExpr) Kind() SyntaxKind      { return SmartQuote }
func (e *SmartQuoteExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *SmartQuoteExpr) isAstNode()            {}
func (e *SmartQuoteExpr) isExpr()               {}

// Double returns true if this is a double quote.
func (e *SmartQuoteExpr) Double() bool {
	return e.node.Text() == "\""
}

// SmartQuoteExprFromNode casts a syntax node to a SmartQuoteExpr.
func SmartQuoteExprFromNode(node *SyntaxNode) *SmartQuoteExpr {
	if node == nil || node.Kind() != SmartQuote {
		return nil
	}
	return &SmartQuoteExpr{node: node}
}

// StrongExpr represents strong (bold) content: *text*.
type StrongExpr struct {
	node *SyntaxNode
}

func (e *StrongExpr) Kind() SyntaxKind      { return Strong }
func (e *StrongExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *StrongExpr) isAstNode()            {}
func (e *StrongExpr) isExpr()               {}

// Body returns the content inside the strong markers.
func (e *StrongExpr) Body() *MarkupNode {
	child := e.node.CastFirst(Markup)
	if child != nil {
		return &MarkupNode{node: child}
	}
	return nil
}

// StrongExprFromNode casts a syntax node to a StrongExpr.
func StrongExprFromNode(node *SyntaxNode) *StrongExpr {
	if node == nil || node.Kind() != Strong {
		return nil
	}
	return &StrongExpr{node: node}
}

// EmphExpr represents emphasized (italic) content: _text_.
type EmphExpr struct {
	node *SyntaxNode
}

func (e *EmphExpr) Kind() SyntaxKind      { return Emph }
func (e *EmphExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *EmphExpr) isAstNode()            {}
func (e *EmphExpr) isExpr()               {}

// Body returns the content inside the emphasis markers.
func (e *EmphExpr) Body() *MarkupNode {
	child := e.node.CastFirst(Markup)
	if child != nil {
		return &MarkupNode{node: child}
	}
	return nil
}

// EmphExprFromNode casts a syntax node to an EmphExpr.
func EmphExprFromNode(node *SyntaxNode) *EmphExpr {
	if node == nil || node.Kind() != Emph {
		return nil
	}
	return &EmphExpr{node: node}
}

// RawExpr represents raw/verbatim text: `code` or ```code```.
type RawExpr struct {
	node *SyntaxNode
}

func (e *RawExpr) Kind() SyntaxKind      { return Raw }
func (e *RawExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *RawExpr) isAstNode()            {}
func (e *RawExpr) isExpr()               {}

// Lang returns the optional language tag.
func (e *RawExpr) Lang() string {
	child := e.node.CastFirst(RawLang)
	if child != nil {
		return child.Text()
	}
	return ""
}

// Block returns true if this is a block raw (triple backticks).
func (e *RawExpr) Block() bool {
	// A block raw has at least 3 backticks
	delims := e.node.CastAll(RawDelim)
	if len(delims) > 0 {
		return len(delims[0].Text()) >= 3
	}
	return false
}

// Lines returns the lines of raw text.
func (e *RawExpr) Lines() []string {
	var lines []string
	for _, child := range e.node.Children() {
		if child.Kind() == Text || child.Kind() == RawTrimmed {
			lines = append(lines, child.Text())
		}
	}
	return lines
}

// RawExprFromNode casts a syntax node to a RawExpr.
func RawExprFromNode(node *SyntaxNode) *RawExpr {
	if node == nil || node.Kind() != Raw {
		return nil
	}
	return &RawExpr{node: node}
}

// LinkExpr represents a hyperlink: https://example.com
type LinkExpr struct {
	node *SyntaxNode
}

func (e *LinkExpr) Kind() SyntaxKind      { return Link }
func (e *LinkExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *LinkExpr) isAstNode()            {}
func (e *LinkExpr) isExpr()               {}

// Get returns the link URL.
func (e *LinkExpr) Get() string {
	return e.node.Text()
}

// LinkExprFromNode casts a syntax node to a LinkExpr.
func LinkExprFromNode(node *SyntaxNode) *LinkExpr {
	if node == nil || node.Kind() != Link {
		return nil
	}
	return &LinkExpr{node: node}
}

// LabelExpr represents a label: <label>.
type LabelExpr struct {
	node *SyntaxNode
}

func (e *LabelExpr) Kind() SyntaxKind      { return Label }
func (e *LabelExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *LabelExpr) isAstNode()            {}
func (e *LabelExpr) isExpr()               {}

// Get returns the label name (without angle brackets).
func (e *LabelExpr) Get() string {
	text := e.node.Text()
	if len(text) >= 2 && text[0] == '<' && text[len(text)-1] == '>' {
		return text[1 : len(text)-1]
	}
	return text
}

// LabelExprFromNode casts a syntax node to a LabelExpr.
func LabelExprFromNode(node *SyntaxNode) *LabelExpr {
	if node == nil || node.Kind() != Label {
		return nil
	}
	return &LabelExpr{node: node}
}

// RefExpr represents a reference: @label.
type RefExpr struct {
	node *SyntaxNode
}

func (e *RefExpr) Kind() SyntaxKind      { return Ref }
func (e *RefExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *RefExpr) isAstNode()            {}
func (e *RefExpr) isExpr()               {}

// Target returns the referenced label.
func (e *RefExpr) Target() string {
	child := e.node.CastFirst(RefMarker)
	if child != nil {
		text := child.Text()
		if len(text) > 0 && text[0] == '@' {
			return text[1:]
		}
		return text
	}
	return ""
}

// Supplement returns the optional supplement content.
func (e *RefExpr) Supplement() *ContentBlockExpr {
	child := e.node.CastFirst(ContentBlock)
	if child != nil {
		return &ContentBlockExpr{node: child}
	}
	return nil
}

// RefExprFromNode casts a syntax node to a RefExpr.
func RefExprFromNode(node *SyntaxNode) *RefExpr {
	if node == nil || node.Kind() != Ref {
		return nil
	}
	return &RefExpr{node: node}
}

// HeadingExpr represents a heading: = Title.
type HeadingExpr struct {
	node *SyntaxNode
}

func (e *HeadingExpr) Kind() SyntaxKind      { return Heading }
func (e *HeadingExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *HeadingExpr) isAstNode()            {}
func (e *HeadingExpr) isExpr()               {}

// Level returns the heading level (number of = signs).
func (e *HeadingExpr) Level() int {
	child := e.node.CastFirst(HeadingMarker)
	if child != nil {
		return len(child.Text())
	}
	return 1
}

// Body returns the heading content.
func (e *HeadingExpr) Body() *MarkupNode {
	child := e.node.CastFirst(Markup)
	if child != nil {
		return &MarkupNode{node: child}
	}
	return nil
}

// HeadingExprFromNode casts a syntax node to a HeadingExpr.
func HeadingExprFromNode(node *SyntaxNode) *HeadingExpr {
	if node == nil || node.Kind() != Heading {
		return nil
	}
	return &HeadingExpr{node: node}
}

// ListItemExpr represents a list item: - item.
type ListItemExpr struct {
	node *SyntaxNode
}

func (e *ListItemExpr) Kind() SyntaxKind      { return ListItem }
func (e *ListItemExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ListItemExpr) isAstNode()            {}
func (e *ListItemExpr) isExpr()               {}

// Body returns the list item content.
func (e *ListItemExpr) Body() *MarkupNode {
	child := e.node.CastFirst(Markup)
	if child != nil {
		return &MarkupNode{node: child}
	}
	return nil
}

// ListItemExprFromNode casts a syntax node to a ListItemExpr.
func ListItemExprFromNode(node *SyntaxNode) *ListItemExpr {
	if node == nil || node.Kind() != ListItem {
		return nil
	}
	return &ListItemExpr{node: node}
}

// EnumItemExpr represents an enum item: + item or 1. item.
type EnumItemExpr struct {
	node *SyntaxNode
}

func (e *EnumItemExpr) Kind() SyntaxKind      { return EnumItem }
func (e *EnumItemExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *EnumItemExpr) isAstNode()            {}
func (e *EnumItemExpr) isExpr()               {}

// Number returns the explicit number, or 0 if auto-numbered.
func (e *EnumItemExpr) Number() int {
	child := e.node.CastFirst(EnumMarker)
	if child != nil {
		text := child.Text()
		// Parse number from "1." format
		var num int
		for _, c := range text {
			if c >= '0' && c <= '9' {
				num = num*10 + int(c-'0')
			} else {
				break
			}
		}
		return num
	}
	return 0
}

// Body returns the enum item content.
func (e *EnumItemExpr) Body() *MarkupNode {
	child := e.node.CastFirst(Markup)
	if child != nil {
		return &MarkupNode{node: child}
	}
	return nil
}

// EnumItemExprFromNode casts a syntax node to an EnumItemExpr.
func EnumItemExprFromNode(node *SyntaxNode) *EnumItemExpr {
	if node == nil || node.Kind() != EnumItem {
		return nil
	}
	return &EnumItemExpr{node: node}
}

// TermItemExpr represents a term list item: / Term: description.
type TermItemExpr struct {
	node *SyntaxNode
}

func (e *TermItemExpr) Kind() SyntaxKind      { return TermItem }
func (e *TermItemExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *TermItemExpr) isAstNode()            {}
func (e *TermItemExpr) isExpr()               {}

// Term returns the term.
func (e *TermItemExpr) Term() *MarkupNode {
	children := e.node.CastAll(Markup)
	if len(children) > 0 {
		return &MarkupNode{node: children[0]}
	}
	return nil
}

// Description returns the description.
func (e *TermItemExpr) Description() *MarkupNode {
	children := e.node.CastAll(Markup)
	if len(children) > 1 {
		return &MarkupNode{node: children[1]}
	}
	return nil
}

// TermItemExprFromNode casts a syntax node to a TermItemExpr.
func TermItemExprFromNode(node *SyntaxNode) *TermItemExpr {
	if node == nil || node.Kind() != TermItem {
		return nil
	}
	return &TermItemExpr{node: node}
}

// ----------------------------------------------------------------------------
// Math Expression Types
// ----------------------------------------------------------------------------

// EquationExpr represents a math equation: $...$
type EquationExpr struct {
	node *SyntaxNode
}

func (e *EquationExpr) Kind() SyntaxKind      { return Equation }
func (e *EquationExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *EquationExpr) isAstNode()            {}
func (e *EquationExpr) isExpr()               {}

// Body returns the math content.
func (e *EquationExpr) Body() *MathNode {
	child := e.node.CastFirst(Math)
	if child != nil {
		return &MathNode{node: child}
	}
	return nil
}

// Block returns true if this is a block equation ($$...$$).
func (e *EquationExpr) Block() bool {
	// Block equations start with double dollar
	return len(e.node.Text()) > 0 && len(e.node.Children()) > 0
}

// EquationExprFromNode casts a syntax node to an EquationExpr.
func EquationExprFromNode(node *SyntaxNode) *EquationExpr {
	if node == nil || node.Kind() != Equation {
		return nil
	}
	return &EquationExpr{node: node}
}

// MathTextExpr represents text within math mode.
type MathTextExpr struct {
	node *SyntaxNode
}

func (e *MathTextExpr) Kind() SyntaxKind      { return MathText }
func (e *MathTextExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathTextExpr) isAstNode()            {}
func (e *MathTextExpr) isExpr()               {}

// Get returns the text content.
func (e *MathTextExpr) Get() string {
	return e.node.Text()
}

// MathTextExprFromNode casts a syntax node to a MathTextExpr.
func MathTextExprFromNode(node *SyntaxNode) *MathTextExpr {
	if node == nil || node.Kind() != MathText {
		return nil
	}
	return &MathTextExpr{node: node}
}

// MathIdentExpr represents an identifier in math mode.
type MathIdentExpr struct {
	node *SyntaxNode
}

func (e *MathIdentExpr) Kind() SyntaxKind      { return MathIdent }
func (e *MathIdentExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathIdentExpr) isAstNode()            {}
func (e *MathIdentExpr) isExpr()               {}

// Get returns the identifier name.
func (e *MathIdentExpr) Get() string {
	return e.node.Text()
}

// MathIdentExprFromNode casts a syntax node to a MathIdentExpr.
func MathIdentExprFromNode(node *SyntaxNode) *MathIdentExpr {
	if node == nil || node.Kind() != MathIdent {
		return nil
	}
	return &MathIdentExpr{node: node}
}

// MathShorthandExpr represents a math shorthand like ->.
type MathShorthandExpr struct {
	node *SyntaxNode
}

func (e *MathShorthandExpr) Kind() SyntaxKind      { return MathShorthand }
func (e *MathShorthandExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathShorthandExpr) isAstNode()            {}
func (e *MathShorthandExpr) isExpr()               {}

// Get returns the shorthand symbol.
func (e *MathShorthandExpr) Get() string {
	return e.node.Text()
}

// MathShorthandExprFromNode casts a syntax node to a MathShorthandExpr.
func MathShorthandExprFromNode(node *SyntaxNode) *MathShorthandExpr {
	if node == nil || node.Kind() != MathShorthand {
		return nil
	}
	return &MathShorthandExpr{node: node}
}

// MathAlignPointExpr represents an alignment point (&) in math.
type MathAlignPointExpr struct {
	node *SyntaxNode
}

func (e *MathAlignPointExpr) Kind() SyntaxKind      { return MathAlignPoint }
func (e *MathAlignPointExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathAlignPointExpr) isAstNode()            {}
func (e *MathAlignPointExpr) isExpr()               {}

// MathAlignPointExprFromNode casts a syntax node to a MathAlignPointExpr.
func MathAlignPointExprFromNode(node *SyntaxNode) *MathAlignPointExpr {
	if node == nil || node.Kind() != MathAlignPoint {
		return nil
	}
	return &MathAlignPointExpr{node: node}
}

// MathDelimitedExpr represents delimited math: (a + b).
type MathDelimitedExpr struct {
	node *SyntaxNode
}

func (e *MathDelimitedExpr) Kind() SyntaxKind      { return MathDelimited }
func (e *MathDelimitedExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathDelimitedExpr) isAstNode()            {}
func (e *MathDelimitedExpr) isExpr()               {}

// Open returns the opening delimiter.
func (e *MathDelimitedExpr) Open() string {
	children := e.node.Children()
	if len(children) > 0 {
		return children[0].Text()
	}
	return ""
}

// Close returns the closing delimiter.
func (e *MathDelimitedExpr) Close() string {
	children := e.node.Children()
	if len(children) > 0 {
		return children[len(children)-1].Text()
	}
	return ""
}

// Body returns the content between delimiters.
func (e *MathDelimitedExpr) Body() *MathNode {
	child := e.node.CastFirst(Math)
	if child != nil {
		return &MathNode{node: child}
	}
	return nil
}

// MathDelimitedExprFromNode casts a syntax node to a MathDelimitedExpr.
func MathDelimitedExprFromNode(node *SyntaxNode) *MathDelimitedExpr {
	if node == nil || node.Kind() != MathDelimited {
		return nil
	}
	return &MathDelimitedExpr{node: node}
}

// MathAttachExpr represents attachments (subscripts/superscripts): x^2_i.
type MathAttachExpr struct {
	node *SyntaxNode
}

func (e *MathAttachExpr) Kind() SyntaxKind      { return MathAttach }
func (e *MathAttachExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathAttachExpr) isAstNode()            {}
func (e *MathAttachExpr) isExpr()               {}

// Base returns the base expression.
func (e *MathAttachExpr) Base() Expr {
	for _, child := range e.node.Children() {
		if child.Kind() != Hat && child.Kind() != Underscore && child.Kind() != MathPrimes {
			return ExprFromNode(child)
		}
	}
	return nil
}

// MathAttachExprFromNode casts a syntax node to a MathAttachExpr.
func MathAttachExprFromNode(node *SyntaxNode) *MathAttachExpr {
	if node == nil || node.Kind() != MathAttach {
		return nil
	}
	return &MathAttachExpr{node: node}
}

// MathPrimesExpr represents prime marks: x', x''.
type MathPrimesExpr struct {
	node *SyntaxNode
}

func (e *MathPrimesExpr) Kind() SyntaxKind      { return MathPrimes }
func (e *MathPrimesExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathPrimesExpr) isAstNode()            {}
func (e *MathPrimesExpr) isExpr()               {}

// Count returns the number of primes.
func (e *MathPrimesExpr) Count() int {
	return len(e.node.Text())
}

// MathPrimesExprFromNode casts a syntax node to a MathPrimesExpr.
func MathPrimesExprFromNode(node *SyntaxNode) *MathPrimesExpr {
	if node == nil || node.Kind() != MathPrimes {
		return nil
	}
	return &MathPrimesExpr{node: node}
}

// MathFracExpr represents a fraction: a/b.
type MathFracExpr struct {
	node *SyntaxNode
}

func (e *MathFracExpr) Kind() SyntaxKind      { return MathFrac }
func (e *MathFracExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathFracExpr) isAstNode()            {}
func (e *MathFracExpr) isExpr()               {}

// Num returns the numerator.
func (e *MathFracExpr) Num() Expr {
	children := e.node.Children()
	if len(children) > 0 {
		return ExprFromNode(children[0])
	}
	return nil
}

// Denom returns the denominator.
func (e *MathFracExpr) Denom() Expr {
	children := e.node.Children()
	for i := 1; i < len(children); i++ {
		if children[i].Kind() != Slash {
			return ExprFromNode(children[i])
		}
	}
	return nil
}

// MathFracExprFromNode casts a syntax node to a MathFracExpr.
func MathFracExprFromNode(node *SyntaxNode) *MathFracExpr {
	if node == nil || node.Kind() != MathFrac {
		return nil
	}
	return &MathFracExpr{node: node}
}

// MathRootExpr represents a root: sqrt(x) or root(n, x).
type MathRootExpr struct {
	node *SyntaxNode
}

func (e *MathRootExpr) Kind() SyntaxKind      { return MathRoot }
func (e *MathRootExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *MathRootExpr) isAstNode()            {}
func (e *MathRootExpr) isExpr()               {}

// Index returns the optional root index (nil for square root).
func (e *MathRootExpr) Index() Expr {
	children := e.node.Children()
	if len(children) > 1 {
		return ExprFromNode(children[0])
	}
	return nil
}

// Radicand returns the expression under the root.
func (e *MathRootExpr) Radicand() Expr {
	children := e.node.Children()
	if len(children) > 0 {
		return ExprFromNode(children[len(children)-1])
	}
	return nil
}

// MathRootExprFromNode casts a syntax node to a MathRootExpr.
func MathRootExprFromNode(node *SyntaxNode) *MathRootExpr {
	if node == nil || node.Kind() != MathRoot {
		return nil
	}
	return &MathRootExpr{node: node}
}

// ----------------------------------------------------------------------------
// Literal Expression Types
// ----------------------------------------------------------------------------

// NoneExpr represents the none literal.
type NoneExpr struct {
	node *SyntaxNode
}

func (e *NoneExpr) Kind() SyntaxKind      { return None }
func (e *NoneExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *NoneExpr) isAstNode()            {}
func (e *NoneExpr) isExpr()               {}

// NoneExprFromNode casts a syntax node to a NoneExpr.
func NoneExprFromNode(node *SyntaxNode) *NoneExpr {
	if node == nil || node.Kind() != None {
		return nil
	}
	return &NoneExpr{node: node}
}

// AutoExpr represents the auto literal.
type AutoExpr struct {
	node *SyntaxNode
}

func (e *AutoExpr) Kind() SyntaxKind      { return Auto }
func (e *AutoExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *AutoExpr) isAstNode()            {}
func (e *AutoExpr) isExpr()               {}

// AutoExprFromNode casts a syntax node to an AutoExpr.
func AutoExprFromNode(node *SyntaxNode) *AutoExpr {
	if node == nil || node.Kind() != Auto {
		return nil
	}
	return &AutoExpr{node: node}
}

// BoolExpr represents a boolean literal: true or false.
type BoolExpr struct {
	node *SyntaxNode
}

func (e *BoolExpr) Kind() SyntaxKind      { return Bool }
func (e *BoolExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *BoolExpr) isAstNode()            {}
func (e *BoolExpr) isExpr()               {}

// Get returns the boolean value.
func (e *BoolExpr) Get() bool {
	return e.node.Text() == "true"
}

// BoolExprFromNode casts a syntax node to a BoolExpr.
func BoolExprFromNode(node *SyntaxNode) *BoolExpr {
	if node == nil || node.Kind() != Bool {
		return nil
	}
	return &BoolExpr{node: node}
}

// IntExpr represents an integer literal: 42.
type IntExpr struct {
	node *SyntaxNode
}

func (e *IntExpr) Kind() SyntaxKind      { return Int }
func (e *IntExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *IntExpr) isAstNode()            {}
func (e *IntExpr) isExpr()               {}

// Get returns the integer value.
func (e *IntExpr) Get() int64 {
	text := e.node.Text()

	// Handle different bases
	base := 10
	digits := text
	if len(text) >= 2 {
		switch text[0:2] {
		case "0x", "0X":
			base = 16
			digits = text[2:]
		case "0b", "0B":
			base = 2
			digits = text[2:]
		case "0o", "0O":
			base = 8
			digits = text[2:]
		}
	}

	var result int64
	for _, c := range digits {
		var digit int64
		switch {
		case c >= '0' && c <= '9':
			digit = int64(c - '0')
		case c >= 'a' && c <= 'f':
			digit = int64(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			digit = int64(c - 'A' + 10)
		default:
			continue // Skip non-digit characters
		}
		if digit >= int64(base) {
			continue // Invalid digit for this base
		}
		result = result*int64(base) + digit
	}
	return result
}

// IntExprFromNode casts a syntax node to an IntExpr.
func IntExprFromNode(node *SyntaxNode) *IntExpr {
	if node == nil || node.Kind() != Int {
		return nil
	}
	return &IntExpr{node: node}
}

// FloatExpr represents a float literal: 3.14.
type FloatExpr struct {
	node *SyntaxNode
}

func (e *FloatExpr) Kind() SyntaxKind      { return Float }
func (e *FloatExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *FloatExpr) isAstNode()            {}
func (e *FloatExpr) isExpr()               {}

// Get returns the float value.
func (e *FloatExpr) Get() float64 {
	// Simple parse, full implementation would handle scientific notation
	text := e.node.Text()
	var result float64
	var decimal float64 = 1
	afterDot := false
	for _, c := range text {
		if c == '.' {
			afterDot = true
			continue
		}
		if c >= '0' && c <= '9' {
			if afterDot {
				decimal /= 10
				result += float64(c-'0') * decimal
			} else {
				result = result*10 + float64(c-'0')
			}
		}
	}
	return result
}

// FloatExprFromNode casts a syntax node to a FloatExpr.
func FloatExprFromNode(node *SyntaxNode) *FloatExpr {
	if node == nil || node.Kind() != Float {
		return nil
	}
	return &FloatExpr{node: node}
}

// NumericExpr represents a numeric literal with a unit: 12pt, 1em.
type NumericExpr struct {
	node *SyntaxNode
}

func (e *NumericExpr) Kind() SyntaxKind      { return Numeric }
func (e *NumericExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *NumericExpr) isAstNode()            {}
func (e *NumericExpr) isExpr()               {}

// Value returns the numeric value.
func (e *NumericExpr) Value() float64 {
	text := e.node.Text()
	var result float64
	var decimal float64 = 1
	afterDot := false
	for _, c := range text {
		if c == '.' {
			afterDot = true
			continue
		}
		if c >= '0' && c <= '9' {
			if afterDot {
				decimal /= 10
				result += float64(c-'0') * decimal
			} else {
				result = result*10 + float64(c-'0')
			}
		} else {
			break // Hit unit
		}
	}
	return result
}

// Unit returns the unit type.
func (e *NumericExpr) Unit() Unit {
	text := e.node.Text()
	// Find where the unit starts
	for i := len(text) - 1; i >= 0; i-- {
		c := text[i]
		if c >= '0' && c <= '9' || c == '.' {
			unitStr := text[i+1:]
			return UnitFromString(unitStr)
		}
	}
	return UnitNone
}

// NumericExprFromNode casts a syntax node to a NumericExpr.
func NumericExprFromNode(node *SyntaxNode) *NumericExpr {
	if node == nil || node.Kind() != Numeric {
		return nil
	}
	return &NumericExpr{node: node}
}

// StrExpr represents a string literal: "hello".
type StrExpr struct {
	node *SyntaxNode
}

func (e *StrExpr) Kind() SyntaxKind      { return Str }
func (e *StrExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *StrExpr) isAstNode()            {}
func (e *StrExpr) isExpr()               {}

// Get returns the string value (without quotes).
func (e *StrExpr) Get() string {
	text := e.node.Text()
	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		return text[1 : len(text)-1]
	}
	return text
}

// StrExprFromNode casts a syntax node to a StrExpr.
func StrExprFromNode(node *SyntaxNode) *StrExpr {
	if node == nil || node.Kind() != Str {
		return nil
	}
	return &StrExpr{node: node}
}

// IdentExpr represents an identifier.
type IdentExpr struct {
	node *SyntaxNode
}

func (e *IdentExpr) Kind() SyntaxKind      { return Ident }
func (e *IdentExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *IdentExpr) isAstNode()            {}
func (e *IdentExpr) isExpr()               {}

// Get returns the identifier name.
func (e *IdentExpr) Get() string {
	return e.node.Text()
}

// IdentExprFromNode casts a syntax node to an IdentExpr.
func IdentExprFromNode(node *SyntaxNode) *IdentExpr {
	if node == nil || node.Kind() != Ident {
		return nil
	}
	return &IdentExpr{node: node}
}

// ----------------------------------------------------------------------------
// Collection and Grouping Expression Types
// ----------------------------------------------------------------------------

// ArrayExpr represents an array literal: (1, 2, 3).
type ArrayExpr struct {
	node *SyntaxNode
}

func (e *ArrayExpr) Kind() SyntaxKind      { return Array }
func (e *ArrayExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ArrayExpr) isAstNode()            {}
func (e *ArrayExpr) isExpr()               {}

// Items returns the array items.
func (e *ArrayExpr) Items() []ArrayItem {
	var items []ArrayItem
	for _, child := range e.node.Children() {
		item := ArrayItemFromNode(child)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

// ArrayExprFromNode casts a syntax node to an ArrayExpr.
func ArrayExprFromNode(node *SyntaxNode) *ArrayExpr {
	if node == nil || node.Kind() != Array {
		return nil
	}
	return &ArrayExpr{node: node}
}

// DictExpr represents a dictionary literal: (a: 1, b: 2).
type DictExpr struct {
	node *SyntaxNode
}

func (e *DictExpr) Kind() SyntaxKind      { return Dict }
func (e *DictExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *DictExpr) isAstNode()            {}
func (e *DictExpr) isExpr()               {}

// Items returns the dictionary items.
func (e *DictExpr) Items() []DictItem {
	var items []DictItem
	for _, child := range e.node.Children() {
		item := DictItemFromNode(child)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

// DictExprFromNode casts a syntax node to a DictExpr.
func DictExprFromNode(node *SyntaxNode) *DictExpr {
	if node == nil || node.Kind() != Dict {
		return nil
	}
	return &DictExpr{node: node}
}

// CodeBlockExpr represents a code block: { ... }.
type CodeBlockExpr struct {
	node *SyntaxNode
}

func (e *CodeBlockExpr) Kind() SyntaxKind      { return CodeBlock }
func (e *CodeBlockExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *CodeBlockExpr) isAstNode()            {}
func (e *CodeBlockExpr) isExpr()               {}

// Body returns the code inside the block.
func (e *CodeBlockExpr) Body() *CodeNode {
	child := e.node.CastFirst(Code)
	if child != nil {
		return &CodeNode{node: child}
	}
	return nil
}

// CodeBlockExprFromNode casts a syntax node to a CodeBlockExpr.
func CodeBlockExprFromNode(node *SyntaxNode) *CodeBlockExpr {
	if node == nil || node.Kind() != CodeBlock {
		return nil
	}
	return &CodeBlockExpr{node: node}
}

// ContentBlockExpr represents a content block: [ ... ].
type ContentBlockExpr struct {
	node *SyntaxNode
}

func (e *ContentBlockExpr) Kind() SyntaxKind      { return ContentBlock }
func (e *ContentBlockExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ContentBlockExpr) isAstNode()            {}
func (e *ContentBlockExpr) isExpr()               {}

// Body returns the markup inside the block.
func (e *ContentBlockExpr) Body() *MarkupNode {
	child := e.node.CastFirst(Markup)
	if child != nil {
		return &MarkupNode{node: child}
	}
	return nil
}

// ContentBlockExprFromNode casts a syntax node to a ContentBlockExpr.
func ContentBlockExprFromNode(node *SyntaxNode) *ContentBlockExpr {
	if node == nil || node.Kind() != ContentBlock {
		return nil
	}
	return &ContentBlockExpr{node: node}
}

// ParenthesizedExpr represents a parenthesized expression: (expr).
type ParenthesizedExpr struct {
	node *SyntaxNode
}

func (e *ParenthesizedExpr) Kind() SyntaxKind      { return Parenthesized }
func (e *ParenthesizedExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ParenthesizedExpr) isAstNode()            {}
func (e *ParenthesizedExpr) isExpr()               {}

// Expr returns the inner expression.
func (e *ParenthesizedExpr) Expr() Expr {
	for _, child := range e.node.Children() {
		if child.Kind() != LeftParen && child.Kind() != RightParen {
			return ExprFromNode(child)
		}
	}
	return nil
}

// ParenthesizedExprFromNode casts a syntax node to a ParenthesizedExpr.
func ParenthesizedExprFromNode(node *SyntaxNode) *ParenthesizedExpr {
	if node == nil || node.Kind() != Parenthesized {
		return nil
	}
	return &ParenthesizedExpr{node: node}
}

// ----------------------------------------------------------------------------
// Operation Expression Types
// ----------------------------------------------------------------------------

// UnaryExpr represents a unary operation: -x, +x, not x.
type UnaryExpr struct {
	node *SyntaxNode
}

func (e *UnaryExpr) Kind() SyntaxKind      { return Unary }
func (e *UnaryExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *UnaryExpr) isAstNode()            {}
func (e *UnaryExpr) isExpr()               {}

// Op returns the unary operator.
func (e *UnaryExpr) Op() UnOp {
	for _, child := range e.node.Children() {
		switch child.Kind() {
		case Plus:
			return UnOpPos
		case Minus:
			return UnOpNeg
		case Not:
			return UnOpNot
		}
	}
	return UnOpPos
}

// Expr returns the operand.
func (e *UnaryExpr) Expr() Expr {
	for _, child := range e.node.Children() {
		if child.Kind() != Plus && child.Kind() != Minus && child.Kind() != Not {
			return ExprFromNode(child)
		}
	}
	return nil
}

// UnaryExprFromNode casts a syntax node to a UnaryExpr.
func UnaryExprFromNode(node *SyntaxNode) *UnaryExpr {
	if node == nil || node.Kind() != Unary {
		return nil
	}
	return &UnaryExpr{node: node}
}

// BinaryExpr represents a binary operation: a + b, a and b.
type BinaryExpr struct {
	node *SyntaxNode
}

func (e *BinaryExpr) Kind() SyntaxKind      { return Binary }
func (e *BinaryExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *BinaryExpr) isAstNode()            {}
func (e *BinaryExpr) isExpr()               {}

// Lhs returns the left-hand side.
func (e *BinaryExpr) Lhs() Expr {
	children := e.node.Children()
	if len(children) > 0 {
		return ExprFromNode(children[0])
	}
	return nil
}

// Op returns the binary operator.
func (e *BinaryExpr) Op() BinOp {
	for _, child := range e.node.Children() {
		switch child.Kind() {
		case Plus:
			return BinOpAdd
		case Minus:
			return BinOpSub
		case Star:
			return BinOpMul
		case Slash:
			return BinOpDiv
		case And:
			return BinOpAnd
		case Or:
			return BinOpOr
		case EqEq:
			return BinOpEq
		case ExclEq:
			return BinOpNeq
		case Lt:
			return BinOpLt
		case LtEq:
			return BinOpLeq
		case Gt:
			return BinOpGt
		case GtEq:
			return BinOpGeq
		case Eq:
			return BinOpAssign
		case In:
			return BinOpIn
		case PlusEq:
			return BinOpAddAssign
		case HyphEq:
			return BinOpSubAssign
		case StarEq:
			return BinOpMulAssign
		case SlashEq:
			return BinOpDivAssign
		}
	}
	return BinOpAdd
}

// Rhs returns the right-hand side.
func (e *BinaryExpr) Rhs() Expr {
	children := e.node.Children()
	if len(children) >= 3 {
		return ExprFromNode(children[2])
	}
	return nil
}

// BinaryExprFromNode casts a syntax node to a BinaryExpr.
func BinaryExprFromNode(node *SyntaxNode) *BinaryExpr {
	if node == nil || node.Kind() != Binary {
		return nil
	}
	return &BinaryExpr{node: node}
}

// FieldAccessExpr represents field access: x.y.
type FieldAccessExpr struct {
	node *SyntaxNode
}

func (e *FieldAccessExpr) Kind() SyntaxKind      { return FieldAccess }
func (e *FieldAccessExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *FieldAccessExpr) isAstNode()            {}
func (e *FieldAccessExpr) isExpr()               {}

// Target returns the object being accessed.
func (e *FieldAccessExpr) Target() Expr {
	children := e.node.Children()
	if len(children) > 0 {
		return ExprFromNode(children[0])
	}
	return nil
}

// Field returns the field name.
func (e *FieldAccessExpr) Field() *IdentExpr {
	// The field is the last identifier child (after the dot)
	// e.g., in "x.y.z", z is the field for the outer access
	children := e.node.Children()
	for i := len(children) - 1; i >= 0; i-- {
		if children[i].Kind() == Ident {
			return &IdentExpr{node: children[i]}
		}
	}
	return nil
}

// FieldAccessExprFromNode casts a syntax node to a FieldAccessExpr.
func FieldAccessExprFromNode(node *SyntaxNode) *FieldAccessExpr {
	if node == nil || node.Kind() != FieldAccess {
		return nil
	}
	return &FieldAccessExpr{node: node}
}

// FuncCallExpr represents a function call: f(x, y).
type FuncCallExpr struct {
	node *SyntaxNode
}

func (e *FuncCallExpr) Kind() SyntaxKind      { return FuncCall }
func (e *FuncCallExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *FuncCallExpr) isAstNode()            {}
func (e *FuncCallExpr) isExpr()               {}

// Callee returns the function being called.
func (e *FuncCallExpr) Callee() Expr {
	children := e.node.Children()
	if len(children) > 0 {
		return ExprFromNode(children[0])
	}
	return nil
}

// Args returns the function arguments.
func (e *FuncCallExpr) Args() *ArgsNode {
	child := e.node.CastFirst(Args)
	if child != nil {
		return &ArgsNode{node: child}
	}
	return nil
}

// FuncCallExprFromNode casts a syntax node to a FuncCallExpr.
func FuncCallExprFromNode(node *SyntaxNode) *FuncCallExpr {
	if node == nil || node.Kind() != FuncCall {
		return nil
	}
	return &FuncCallExpr{node: node}
}

// ----------------------------------------------------------------------------
// Control Flow Expression Types
// ----------------------------------------------------------------------------

// ClosureExpr represents a closure: (x, y) => x + y.
type ClosureExpr struct {
	node *SyntaxNode
}

func (e *ClosureExpr) Kind() SyntaxKind      { return Closure }
func (e *ClosureExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ClosureExpr) isAstNode()            {}
func (e *ClosureExpr) isExpr()               {}

// Name returns the optional name (for named functions).
func (e *ClosureExpr) Name() *IdentExpr {
	// Named closures have the name as first child before params
	for _, child := range e.node.Children() {
		if child.Kind() == Ident {
			return &IdentExpr{node: child}
		}
		if child.Kind() == Params {
			break
		}
	}
	return nil
}

// Params returns the parameters.
func (e *ClosureExpr) Params() *ParamsNode {
	child := e.node.CastFirst(Params)
	if child != nil {
		return &ParamsNode{node: child}
	}
	return nil
}

// Body returns the closure body.
func (e *ClosureExpr) Body() Expr {
	children := e.node.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if child.Kind() != Arrow && child.Kind() != Params && child.Kind() != Ident {
			return ExprFromNode(child)
		}
	}
	return nil
}

// ClosureExprFromNode casts a syntax node to a ClosureExpr.
func ClosureExprFromNode(node *SyntaxNode) *ClosureExpr {
	if node == nil || node.Kind() != Closure {
		return nil
	}
	return &ClosureExpr{node: node}
}

// LetBindingExpr represents a let binding: let x = 1.
type LetBindingExpr struct {
	node *SyntaxNode
}

func (e *LetBindingExpr) Kind() SyntaxKind      { return LetBinding }
func (e *LetBindingExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *LetBindingExpr) isAstNode()            {}
func (e *LetBindingExpr) isExpr()               {}

// Kind returns LetBindingPlain for simple bindings, LetBindingClosure for closures.
func (e *LetBindingExpr) BindingKind() LetBindingKind {
	// If there's a Closure child, it's a closure binding
	if e.node.CastFirst(Closure) != nil {
		return LetBindingClosure
	}
	return LetBindingPlain
}

// Pattern returns the binding pattern (for plain bindings).
func (e *LetBindingExpr) Pattern() Pattern {
	return PatternFromNode(e.node)
}

// Init returns the initializer expression.
func (e *LetBindingExpr) Init() Expr {
	// Find expression after '='
	children := e.node.Children()
	foundEq := false
	for _, child := range children {
		if child.Kind() == Eq {
			foundEq = true
			continue
		}
		if foundEq {
			return ExprFromNode(child)
		}
	}
	// For closure bindings
	child := e.node.CastFirst(Closure)
	if child != nil {
		return &ClosureExpr{node: child}
	}
	return nil
}

// LetBindingExprFromNode casts a syntax node to a LetBindingExpr.
func LetBindingExprFromNode(node *SyntaxNode) *LetBindingExpr {
	if node == nil || node.Kind() != LetBinding {
		return nil
	}
	return &LetBindingExpr{node: node}
}

// LetBindingKind indicates the kind of let binding.
type LetBindingKind int

const (
	LetBindingPlain LetBindingKind = iota
	LetBindingClosure
)

// DestructAssignmentExpr represents a destructuring assignment: (a, b) = expr.
type DestructAssignmentExpr struct {
	node *SyntaxNode
}

func (e *DestructAssignmentExpr) Kind() SyntaxKind      { return DestructAssignment }
func (e *DestructAssignmentExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *DestructAssignmentExpr) isAstNode()            {}
func (e *DestructAssignmentExpr) isExpr()               {}

// Pattern returns the destructuring pattern.
func (e *DestructAssignmentExpr) Pattern() *DestructuringNode {
	child := e.node.CastFirst(Destructuring)
	if child != nil {
		return &DestructuringNode{node: child}
	}
	return nil
}

// Value returns the value being destructured.
func (e *DestructAssignmentExpr) Value() Expr {
	children := e.node.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if child.Kind() != Destructuring && child.Kind() != Eq {
			return ExprFromNode(child)
		}
	}
	return nil
}

// DestructAssignmentExprFromNode casts a syntax node to a DestructAssignmentExpr.
func DestructAssignmentExprFromNode(node *SyntaxNode) *DestructAssignmentExpr {
	if node == nil || node.Kind() != DestructAssignment {
		return nil
	}
	return &DestructAssignmentExpr{node: node}
}

// SetRuleExpr represents a set rule: set text(fill: red).
type SetRuleExpr struct {
	node *SyntaxNode
}

func (e *SetRuleExpr) Kind() SyntaxKind      { return SetRule }
func (e *SetRuleExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *SetRuleExpr) isAstNode()            {}
func (e *SetRuleExpr) isExpr()               {}

// Target returns the function to configure.
func (e *SetRuleExpr) Target() Expr {
	// Skip the 'set' keyword
	for _, child := range e.node.Children() {
		if child.Kind() != Set {
			return ExprFromNode(child)
		}
	}
	return nil
}

// Condition returns the optional if condition.
func (e *SetRuleExpr) Condition() Expr {
	// Look for 'if' keyword and following expression
	children := e.node.Children()
	for i, child := range children {
		if child.Kind() == If && i+1 < len(children) {
			return ExprFromNode(children[i+1])
		}
	}
	return nil
}

// SetRuleExprFromNode casts a syntax node to a SetRuleExpr.
func SetRuleExprFromNode(node *SyntaxNode) *SetRuleExpr {
	if node == nil || node.Kind() != SetRule {
		return nil
	}
	return &SetRuleExpr{node: node}
}

// ShowRuleExpr represents a show rule: show heading: it => ...
type ShowRuleExpr struct {
	node *SyntaxNode
}

func (e *ShowRuleExpr) Kind() SyntaxKind      { return ShowRule }
func (e *ShowRuleExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ShowRuleExpr) isAstNode()            {}
func (e *ShowRuleExpr) isExpr()               {}

// Selector returns the optional selector.
func (e *ShowRuleExpr) Selector() Expr {
	// First expression after 'show' keyword (before colon)
	children := e.node.Children()
	for i, child := range children {
		if child.Kind() == Show {
			continue
		}
		if child.Kind() == Colon {
			break
		}
		if i > 0 {
			return ExprFromNode(child)
		}
	}
	return nil
}

// Transform returns the transform expression.
func (e *ShowRuleExpr) Transform() Expr {
	// Expression after colon
	children := e.node.Children()
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

// ShowRuleExprFromNode casts a syntax node to a ShowRuleExpr.
func ShowRuleExprFromNode(node *SyntaxNode) *ShowRuleExpr {
	if node == nil || node.Kind() != ShowRule {
		return nil
	}
	return &ShowRuleExpr{node: node}
}

// ContextualExpr represents a context expression: context text.lang.
type ContextualExpr struct {
	node *SyntaxNode
}

func (e *ContextualExpr) Kind() SyntaxKind      { return Contextual }
func (e *ContextualExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ContextualExpr) isAstNode()            {}
func (e *ContextualExpr) isExpr()               {}

// Body returns the body expression.
func (e *ContextualExpr) Body() Expr {
	for _, child := range e.node.Children() {
		if child.Kind() != Context {
			return ExprFromNode(child)
		}
	}
	return nil
}

// ContextualExprFromNode casts a syntax node to a ContextualExpr.
func ContextualExprFromNode(node *SyntaxNode) *ContextualExpr {
	if node == nil || node.Kind() != Contextual {
		return nil
	}
	return &ContextualExpr{node: node}
}

// ConditionalExpr represents a conditional: if cond { ... } else { ... }.
type ConditionalExpr struct {
	node *SyntaxNode
}

func (e *ConditionalExpr) Kind() SyntaxKind      { return Conditional }
func (e *ConditionalExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ConditionalExpr) isAstNode()            {}
func (e *ConditionalExpr) isExpr()               {}

// Condition returns the condition expression.
func (e *ConditionalExpr) Condition() Expr {
	children := e.node.Children()
	for i, child := range children {
		if child.Kind() == If && i+1 < len(children) {
			return ExprFromNode(children[i+1])
		}
	}
	return nil
}

// IfBody returns the if branch body.
func (e *ConditionalExpr) IfBody() Expr {
	children := e.node.Children()
	count := 0
	for _, child := range children {
		if child.Kind() == CodeBlock || child.Kind() == ContentBlock {
			if count == 0 {
				return ExprFromNode(child)
			}
			count++
		}
	}
	return nil
}

// ElseBody returns the optional else branch body.
func (e *ConditionalExpr) ElseBody() Expr {
	children := e.node.Children()
	foundElse := false
	for _, child := range children {
		if child.Kind() == Else {
			foundElse = true
			continue
		}
		if foundElse {
			return ExprFromNode(child)
		}
	}
	return nil
}

// ConditionalExprFromNode casts a syntax node to a ConditionalExpr.
func ConditionalExprFromNode(node *SyntaxNode) *ConditionalExpr {
	if node == nil || node.Kind() != Conditional {
		return nil
	}
	return &ConditionalExpr{node: node}
}

// WhileLoopExpr represents a while loop: while cond { ... }.
type WhileLoopExpr struct {
	node *SyntaxNode
}

func (e *WhileLoopExpr) Kind() SyntaxKind      { return WhileLoop }
func (e *WhileLoopExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *WhileLoopExpr) isAstNode()            {}
func (e *WhileLoopExpr) isExpr()               {}

// Condition returns the loop condition.
func (e *WhileLoopExpr) Condition() Expr {
	children := e.node.Children()
	for i, child := range children {
		if child.Kind() == While && i+1 < len(children) {
			return ExprFromNode(children[i+1])
		}
	}
	return nil
}

// Body returns the loop body.
func (e *WhileLoopExpr) Body() Expr {
	child := e.node.CastFirst(CodeBlock)
	if child != nil {
		return &CodeBlockExpr{node: child}
	}
	child = e.node.CastFirst(ContentBlock)
	if child != nil {
		return &ContentBlockExpr{node: child}
	}
	return nil
}

// WhileLoopExprFromNode casts a syntax node to a WhileLoopExpr.
func WhileLoopExprFromNode(node *SyntaxNode) *WhileLoopExpr {
	if node == nil || node.Kind() != WhileLoop {
		return nil
	}
	return &WhileLoopExpr{node: node}
}

// ForLoopExpr represents a for loop: for x in items { ... }.
type ForLoopExpr struct {
	node *SyntaxNode
}

func (e *ForLoopExpr) Kind() SyntaxKind      { return ForLoop }
func (e *ForLoopExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ForLoopExpr) isAstNode()            {}
func (e *ForLoopExpr) isExpr()               {}

// Pattern returns the binding pattern.
func (e *ForLoopExpr) Pattern() Pattern {
	return PatternFromNode(e.node)
}

// Iter returns the iterable expression.
func (e *ForLoopExpr) Iter() Expr {
	children := e.node.Children()
	foundIn := false
	for _, child := range children {
		if child.Kind() == In {
			foundIn = true
			continue
		}
		if foundIn && child.Kind() != CodeBlock && child.Kind() != ContentBlock {
			return ExprFromNode(child)
		}
	}
	return nil
}

// Body returns the loop body.
func (e *ForLoopExpr) Body() Expr {
	child := e.node.CastFirst(CodeBlock)
	if child != nil {
		return &CodeBlockExpr{node: child}
	}
	child = e.node.CastFirst(ContentBlock)
	if child != nil {
		return &ContentBlockExpr{node: child}
	}
	return nil
}

// ForLoopExprFromNode casts a syntax node to a ForLoopExpr.
func ForLoopExprFromNode(node *SyntaxNode) *ForLoopExpr {
	if node == nil || node.Kind() != ForLoop {
		return nil
	}
	return &ForLoopExpr{node: node}
}

// ----------------------------------------------------------------------------
// Module Expression Types
// ----------------------------------------------------------------------------

// ModuleImportExpr represents an import: import "file.typ".
type ModuleImportExpr struct {
	node *SyntaxNode
}

func (e *ModuleImportExpr) Kind() SyntaxKind      { return ModuleImport }
func (e *ModuleImportExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ModuleImportExpr) isAstNode()            {}
func (e *ModuleImportExpr) isExpr()               {}

// Source returns the import source expression.
func (e *ModuleImportExpr) Source() Expr {
	for _, child := range e.node.Children() {
		if child.Kind() != Import && child.Kind() != Colon && child.Kind() != As &&
		   child.Kind() != Ident && child.Kind() != ImportItems && child.Kind() != Star {
			return ExprFromNode(child)
		}
	}
	return nil
}

// NewName returns the optional rename (import "x" as y).
func (e *ModuleImportExpr) NewName() *IdentExpr {
	children := e.node.Children()
	for i, child := range children {
		if child.Kind() == As && i+1 < len(children) {
			nextChild := children[i+1]
			if nextChild.Kind() == Ident {
				return &IdentExpr{node: nextChild}
			}
		}
	}
	return nil
}

// Imports returns the import items.
func (e *ModuleImportExpr) Imports() Imports {
	// Check for wildcard
	if e.node.CastFirst(Star) != nil {
		return &ImportsWildcard{}
	}
	// Check for explicit items
	child := e.node.CastFirst(ImportItems)
	if child != nil {
		return &ImportItemsNode{node: child}
	}
	return nil
}

// ModuleImportExprFromNode casts a syntax node to a ModuleImportExpr.
func ModuleImportExprFromNode(node *SyntaxNode) *ModuleImportExpr {
	if node == nil || node.Kind() != ModuleImport {
		return nil
	}
	return &ModuleImportExpr{node: node}
}

// ModuleIncludeExpr represents an include: include "file.typ".
type ModuleIncludeExpr struct {
	node *SyntaxNode
}

func (e *ModuleIncludeExpr) Kind() SyntaxKind      { return ModuleInclude }
func (e *ModuleIncludeExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *ModuleIncludeExpr) isAstNode()            {}
func (e *ModuleIncludeExpr) isExpr()               {}

// Source returns the include source expression.
func (e *ModuleIncludeExpr) Source() Expr {
	for _, child := range e.node.Children() {
		if child.Kind() != Include {
			return ExprFromNode(child)
		}
	}
	return nil
}

// ModuleIncludeExprFromNode casts a syntax node to a ModuleIncludeExpr.
func ModuleIncludeExprFromNode(node *SyntaxNode) *ModuleIncludeExpr {
	if node == nil || node.Kind() != ModuleInclude {
		return nil
	}
	return &ModuleIncludeExpr{node: node}
}

// ----------------------------------------------------------------------------
// Loop Control Expression Types
// ----------------------------------------------------------------------------

// LoopBreakExpr represents a break statement.
type LoopBreakExpr struct {
	node *SyntaxNode
}

func (e *LoopBreakExpr) Kind() SyntaxKind      { return LoopBreak }
func (e *LoopBreakExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *LoopBreakExpr) isAstNode()            {}
func (e *LoopBreakExpr) isExpr()               {}

// LoopBreakExprFromNode casts a syntax node to a LoopBreakExpr.
func LoopBreakExprFromNode(node *SyntaxNode) *LoopBreakExpr {
	if node == nil || node.Kind() != LoopBreak {
		return nil
	}
	return &LoopBreakExpr{node: node}
}

// LoopContinueExpr represents a continue statement.
type LoopContinueExpr struct {
	node *SyntaxNode
}

func (e *LoopContinueExpr) Kind() SyntaxKind      { return LoopContinue }
func (e *LoopContinueExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *LoopContinueExpr) isAstNode()            {}
func (e *LoopContinueExpr) isExpr()               {}

// LoopContinueExprFromNode casts a syntax node to a LoopContinueExpr.
func LoopContinueExprFromNode(node *SyntaxNode) *LoopContinueExpr {
	if node == nil || node.Kind() != LoopContinue {
		return nil
	}
	return &LoopContinueExpr{node: node}
}

// FuncReturnExpr represents a return statement.
type FuncReturnExpr struct {
	node *SyntaxNode
}

func (e *FuncReturnExpr) Kind() SyntaxKind      { return FuncReturn }
func (e *FuncReturnExpr) ToUntyped() *SyntaxNode { return e.node }
func (e *FuncReturnExpr) isAstNode()            {}
func (e *FuncReturnExpr) isExpr()               {}

// Body returns the optional return value.
func (e *FuncReturnExpr) Body() Expr {
	for _, child := range e.node.Children() {
		if child.Kind() != Return {
			return ExprFromNode(child)
		}
	}
	return nil
}

// FuncReturnExprFromNode casts a syntax node to a FuncReturnExpr.
func FuncReturnExprFromNode(node *SyntaxNode) *FuncReturnExpr {
	if node == nil || node.Kind() != FuncReturn {
		return nil
	}
	return &FuncReturnExpr{node: node}
}

// ----------------------------------------------------------------------------
// Container Node Types (non-expression wrapper nodes)
// ----------------------------------------------------------------------------

// MarkupNode represents a markup content container.
type MarkupNode struct {
	node *SyntaxNode
}

func (n *MarkupNode) Kind() SyntaxKind      { return Markup }
func (n *MarkupNode) ToUntyped() *SyntaxNode { return n.node }
func (n *MarkupNode) isAstNode()            {}

// Exprs returns the expressions in this markup.
func (n *MarkupNode) Exprs() []Expr {
	var exprs []Expr
	for _, child := range n.node.Children() {
		expr := ExprFromNode(child)
		if expr != nil {
			exprs = append(exprs, expr)
		}
	}
	return exprs
}

// MarkupNodeFromNode casts a syntax node to a MarkupNode.
func MarkupNodeFromNode(node *SyntaxNode) *MarkupNode {
	if node == nil || node.Kind() != Markup {
		return nil
	}
	return &MarkupNode{node: node}
}

// CodeNode represents a code content container.
type CodeNode struct {
	node *SyntaxNode
}

func (n *CodeNode) Kind() SyntaxKind      { return Code }
func (n *CodeNode) ToUntyped() *SyntaxNode { return n.node }
func (n *CodeNode) isAstNode()            {}

// Exprs returns the expressions in this code block.
func (n *CodeNode) Exprs() []Expr {
	var exprs []Expr
	for _, child := range n.node.Children() {
		expr := ExprFromNode(child)
		if expr != nil {
			exprs = append(exprs, expr)
		}
	}
	return exprs
}

// CodeNodeFromNode casts a syntax node to a CodeNode.
func CodeNodeFromNode(node *SyntaxNode) *CodeNode {
	if node == nil || node.Kind() != Code {
		return nil
	}
	return &CodeNode{node: node}
}

// MathNode represents a math content container.
type MathNode struct {
	node *SyntaxNode
}

func (n *MathNode) Kind() SyntaxKind      { return Math }
func (n *MathNode) ToUntyped() *SyntaxNode { return n.node }
func (n *MathNode) isAstNode()            {}

// Exprs returns the expressions in this math block.
func (n *MathNode) Exprs() []Expr {
	var exprs []Expr
	for _, child := range n.node.Children() {
		expr := ExprFromNode(child)
		if expr != nil {
			exprs = append(exprs, expr)
		}
	}
	return exprs
}

// MathNodeFromNode casts a syntax node to a MathNode.
func MathNodeFromNode(node *SyntaxNode) *MathNode {
	if node == nil || node.Kind() != Math {
		return nil
	}
	return &MathNode{node: node}
}

// ArgsNode represents function call arguments.
type ArgsNode struct {
	node *SyntaxNode
}

func (n *ArgsNode) Kind() SyntaxKind      { return Args }
func (n *ArgsNode) ToUntyped() *SyntaxNode { return n.node }
func (n *ArgsNode) isAstNode()            {}

// Items returns the argument items.
func (n *ArgsNode) Items() []Arg {
	var items []Arg
	for _, child := range n.node.Children() {
		item := ArgFromNode(child)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

// TrailingComma returns true if there's a trailing comma.
func (n *ArgsNode) TrailingComma() bool {
	children := n.node.Children()
	if len(children) > 0 {
		return children[len(children)-1].Kind() == Comma
	}
	return false
}

// ArgsNodeFromNode casts a syntax node to an ArgsNode.
func ArgsNodeFromNode(node *SyntaxNode) *ArgsNode {
	if node == nil || node.Kind() != Args {
		return nil
	}
	return &ArgsNode{node: node}
}

// ParamsNode represents function parameters.
type ParamsNode struct {
	node *SyntaxNode
}

func (n *ParamsNode) Kind() SyntaxKind      { return Params }
func (n *ParamsNode) ToUntyped() *SyntaxNode { return n.node }
func (n *ParamsNode) isAstNode()            {}

// Children returns the parameter items.
func (n *ParamsNode) Children() []Param {
	var params []Param
	for _, child := range n.node.Children() {
		param := ParamFromNode(child)
		if param != nil {
			params = append(params, param)
		}
	}
	return params
}

// ParamsNodeFromNode casts a syntax node to a ParamsNode.
func ParamsNodeFromNode(node *SyntaxNode) *ParamsNode {
	if node == nil || node.Kind() != Params {
		return nil
	}
	return &ParamsNode{node: node}
}

// DestructuringNode represents a destructuring pattern.
type DestructuringNode struct {
	node *SyntaxNode
}

func (n *DestructuringNode) Kind() SyntaxKind      { return Destructuring }
func (n *DestructuringNode) ToUntyped() *SyntaxNode { return n.node }
func (n *DestructuringNode) isAstNode()            {}

// Items returns the destructuring items.
func (n *DestructuringNode) Items() []DestructuringItem {
	var items []DestructuringItem
	for _, child := range n.node.Children() {
		item := DestructuringItemFromNode(child)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

// DestructuringNodeFromNode casts a syntax node to a DestructuringNode.
func DestructuringNodeFromNode(node *SyntaxNode) *DestructuringNode {
	if node == nil || node.Kind() != Destructuring {
		return nil
	}
	return &DestructuringNode{node: node}
}
