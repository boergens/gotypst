// Package syntax provides the foundational types for Typst's syntax tree.
// It defines SyntaxKind (token and node types) and SyntaxSet (bitset for kinds).
package syntax

// SyntaxKind represents the type of a syntax node or token.
// This is the foundation type for the Typst syntax tree.
type SyntaxKind uint8

// All syntax kinds in Typst.
const (
	End SyntaxKind = iota
	Error

	// Comments
	Shebang
	LineComment
	BlockComment

	// Markup
	Markup
	Text
	Space
	Linebreak
	Parbreak

	// Escape sequences
	Escape
	Shorthand
	SmartQuote

	// Text formatting
	Strong
	Emph

	// Raw blocks
	Raw
	RawLang
	RawDelim
	RawTrimmed

	// References and labels
	Link
	Label
	Ref
	RefMarker

	// Headings and lists
	Heading
	HeadingMarker
	ListItem
	ListMarker
	EnumItem
	EnumMarker
	TermItem
	TermMarker

	// Math
	Equation
	Math
	MathText
	MathIdent
	MathShorthand
	MathAlignPoint
	MathDelimited
	MathAttach
	MathPrimes
	MathFrac
	MathRoot

	// Operators and delimiters
	Hash
	LeftBrace
	RightBrace
	LeftBracket
	RightBracket
	LeftParen
	RightParen
	Comma
	Semicolon
	Colon
	Star
	Underscore
	Dollar
	Plus
	Minus
	Slash
	Hat
	Dot
	Eq
	EqEq
	ExclEq
	Lt
	LtEq
	Gt
	GtEq
	PlusEq
	HyphEq
	StarEq
	SlashEq
	Dots
	Arrow
	Root
	Bang

	// Keyword operators
	Not
	And
	Or

	// Keyword literals
	None
	Auto

	// Keywords
	Let
	Set
	Show
	Context
	If
	Else
	For
	In
	While
	Break
	Continue
	Return
	Import
	Include
	As

	// Code elements
	Code
	Ident
	Bool
	Int
	Float
	Numeric
	Str

	// Expressions and blocks
	CodeBlock
	ContentBlock
	Parenthesized
	Array
	Dict
	Named
	Keyed
	Unary
	Binary
	FieldAccess
	FuncCall
	Args
	Spread
	Closure
	Params

	// Statements and control flow
	LetBinding
	SetRule
	ShowRule
	Contextual
	Conditional
	WhileLoop
	ForLoop
	ModuleImport
	ImportItems
	ImportItemPath
	RenamedImportItem
	ModuleInclude
	LoopBreak
	LoopContinue
	FuncReturn
	Destructuring
	DestructAssignment
)

// IsGrouping returns true if this kind is a bracket, brace, or parenthesis.
func (k SyntaxKind) IsGrouping() bool {
	switch k {
	case LeftBrace, RightBrace, LeftBracket, RightBracket, LeftParen, RightParen:
		return true
	}
	return false
}

// IsTerminator returns true if this kind terminates an expression.
func (k SyntaxKind) IsTerminator() bool {
	switch k {
	case End, Semicolon, RightBrace, RightParen, RightBracket:
		return true
	}
	return false
}

// IsBlock returns true if this kind is a code or content block.
func (k SyntaxKind) IsBlock() bool {
	switch k {
	case CodeBlock, ContentBlock:
		return true
	}
	return false
}

// IsStmt returns true if this kind is a statement-level construct.
func (k SyntaxKind) IsStmt() bool {
	switch k {
	case LetBinding, SetRule, ShowRule, ModuleImport, ModuleInclude:
		return true
	}
	return false
}

// IsTrivia returns true if this kind is automatically skipped in code/math mode.
func (k SyntaxKind) IsTrivia() bool {
	switch k {
	case Shebang, LineComment, BlockComment, Space, Parbreak:
		return true
	}
	return false
}

// IsKeyword returns true if this kind is a language keyword.
func (k SyntaxKind) IsKeyword() bool {
	switch k {
	case Not, And, Or, None, Auto,
		Let, Set, Show, Context,
		If, Else, For, In, While,
		Break, Continue, Return,
		Import, Include, As:
		return true
	}
	return false
}

// IsError returns true if this kind is an error node.
func (k SyntaxKind) IsError() bool {
	return k == Error
}

// Name returns a human-readable name for the syntax kind.
func (k SyntaxKind) Name() string {
	switch k {
	case End:
		return "end of tokens"
	case Error:
		return "syntax error"
	case Shebang:
		return "shebang"
	case LineComment:
		return "line comment"
	case BlockComment:
		return "block comment"
	case Markup:
		return "markup"
	case Text:
		return "text"
	case Space:
		return "space"
	case Linebreak:
		return "line break"
	case Parbreak:
		return "paragraph break"
	case Escape:
		return "escape sequence"
	case Shorthand:
		return "shorthand"
	case SmartQuote:
		return "smart quote"
	case Strong:
		return "strong content"
	case Emph:
		return "emphasized content"
	case Raw:
		return "raw block"
	case RawLang:
		return "raw language tag"
	case RawDelim:
		return "raw delimiter"
	case RawTrimmed:
		return "raw trimmed"
	case Link:
		return "link"
	case Label:
		return "label"
	case Ref:
		return "reference"
	case RefMarker:
		return "reference marker"
	case Heading:
		return "heading"
	case HeadingMarker:
		return "heading marker"
	case ListItem:
		return "list item"
	case ListMarker:
		return "list marker"
	case EnumItem:
		return "enum item"
	case EnumMarker:
		return "enum marker"
	case TermItem:
		return "term list item"
	case TermMarker:
		return "term marker"
	case Equation:
		return "equation"
	case Math:
		return "math"
	case MathText:
		return "math text"
	case MathIdent:
		return "math identifier"
	case MathShorthand:
		return "math shorthand"
	case MathAlignPoint:
		return "math alignment point"
	case MathDelimited:
		return "delimited math"
	case MathAttach:
		return "math attachments"
	case MathPrimes:
		return "math primes"
	case MathFrac:
		return "math fraction"
	case MathRoot:
		return "math root"
	case Hash:
		return "hash"
	case LeftBrace:
		return "opening brace"
	case RightBrace:
		return "closing brace"
	case LeftBracket:
		return "opening bracket"
	case RightBracket:
		return "closing bracket"
	case LeftParen:
		return "opening paren"
	case RightParen:
		return "closing paren"
	case Comma:
		return "comma"
	case Semicolon:
		return "semicolon"
	case Colon:
		return "colon"
	case Star:
		return "star"
	case Underscore:
		return "underscore"
	case Dollar:
		return "dollar sign"
	case Plus:
		return "plus"
	case Minus:
		return "minus"
	case Slash:
		return "slash"
	case Hat:
		return "hat"
	case Dot:
		return "dot"
	case Eq:
		return "equals sign"
	case EqEq:
		return "equality operator"
	case ExclEq:
		return "inequality operator"
	case Lt:
		return "less-than operator"
	case LtEq:
		return "less-than or equal operator"
	case Gt:
		return "greater-than operator"
	case GtEq:
		return "greater-than or equal operator"
	case PlusEq:
		return "add-assign operator"
	case HyphEq:
		return "subtract-assign operator"
	case StarEq:
		return "multiply-assign operator"
	case SlashEq:
		return "divide-assign operator"
	case Dots:
		return "dots"
	case Arrow:
		return "arrow"
	case Root:
		return "root"
	case Bang:
		return "exclamation mark"
	case Not:
		return "operator `not`"
	case And:
		return "operator `and`"
	case Or:
		return "operator `or`"
	case None:
		return "`none`"
	case Auto:
		return "`auto`"
	case Let:
		return "keyword `let`"
	case Set:
		return "keyword `set`"
	case Show:
		return "keyword `show`"
	case Context:
		return "keyword `context`"
	case If:
		return "keyword `if`"
	case Else:
		return "keyword `else`"
	case For:
		return "keyword `for`"
	case In:
		return "keyword `in`"
	case While:
		return "keyword `while`"
	case Break:
		return "keyword `break`"
	case Continue:
		return "keyword `continue`"
	case Return:
		return "keyword `return`"
	case Import:
		return "keyword `import`"
	case Include:
		return "keyword `include`"
	case As:
		return "keyword `as`"
	case Code:
		return "code"
	case Ident:
		return "identifier"
	case Bool:
		return "boolean"
	case Int:
		return "integer"
	case Float:
		return "float"
	case Numeric:
		return "numeric value"
	case Str:
		return "string"
	case CodeBlock:
		return "code block"
	case ContentBlock:
		return "content block"
	case Parenthesized:
		return "group"
	case Array:
		return "array"
	case Dict:
		return "dictionary"
	case Named:
		return "named pair"
	case Keyed:
		return "keyed pair"
	case Unary:
		return "unary expression"
	case Binary:
		return "binary expression"
	case FieldAccess:
		return "field access"
	case FuncCall:
		return "function call"
	case Args:
		return "call arguments"
	case Spread:
		return "spread"
	case Closure:
		return "closure"
	case Params:
		return "closure parameters"
	case LetBinding:
		return "`let` expression"
	case SetRule:
		return "`set` expression"
	case ShowRule:
		return "`show` expression"
	case Contextual:
		return "`context` expression"
	case Conditional:
		return "`if` expression"
	case WhileLoop:
		return "while-loop expression"
	case ForLoop:
		return "for-loop expression"
	case ModuleImport:
		return "`import` expression"
	case ImportItems:
		return "import items"
	case ImportItemPath:
		return "imported item path"
	case RenamedImportItem:
		return "renamed import item"
	case ModuleInclude:
		return "`include` expression"
	case LoopBreak:
		return "`break` expression"
	case LoopContinue:
		return "`continue` expression"
	case FuncReturn:
		return "`return` expression"
	case Destructuring:
		return "destructuring pattern"
	case DestructAssignment:
		return "destructuring assignment expression"
	default:
		return "unknown"
	}
}

// String returns the name of the syntax kind (same as Name).
func (k SyntaxKind) String() string {
	return k.Name()
}
