package syntax

// SyntaxSet is a set of syntax kinds implemented as a bitset.
// It can hold kinds with discriminator values less than 128.
//
// Based on rust-analyzer's TokenSet:
// https://github.com/rust-lang/rust-analyzer/blob/master/crates/parser/src/token_set.rs
type SyntaxSet struct {
	lo uint64 // bits 0-63
	hi uint64 // bits 64-127
}

const maxSetBit = 128

// NewSyntaxSet creates a new empty set.
func NewSyntaxSet() SyntaxSet {
	return SyntaxSet{}
}

// SyntaxSetOf creates a set containing the given kinds.
func SyntaxSetOf(kinds ...SyntaxKind) SyntaxSet {
	s := SyntaxSet{}
	for _, k := range kinds {
		s = s.Add(k)
	}
	return s
}

// Add inserts a syntax kind into the set and returns the new set.
// Panics if the kind's discriminator is >= 128.
func (s SyntaxSet) Add(kind SyntaxKind) SyntaxSet {
	if kind >= maxSetBit {
		panic("SyntaxSet.Add: kind discriminator must be < 128")
	}
	if kind < 64 {
		s.lo |= 1 << kind
	} else {
		s.hi |= 1 << (kind - 64)
	}
	return s
}

// Remove removes a syntax kind from the set and returns the new set.
// Does nothing if the kind is not present.
// Panics if the kind's discriminator is >= 128.
func (s SyntaxSet) Remove(kind SyntaxKind) SyntaxSet {
	if kind >= maxSetBit {
		panic("SyntaxSet.Remove: kind discriminator must be < 128")
	}
	if kind < 64 {
		s.lo &^= 1 << kind
	} else {
		s.hi &^= 1 << (kind - 64)
	}
	return s
}

// Union combines two syntax sets.
func (s SyntaxSet) Union(other SyntaxSet) SyntaxSet {
	return SyntaxSet{
		lo: s.lo | other.lo,
		hi: s.hi | other.hi,
	}
}

// Contains returns true if the set contains the given syntax kind.
func (s SyntaxSet) Contains(kind SyntaxKind) bool {
	if kind >= maxSetBit {
		return false
	}
	if kind < 64 {
		return (s.lo & (1 << kind)) != 0
	}
	return (s.hi & (1 << (kind - 64))) != 0
}

// IsEmpty returns true if the set contains no kinds.
func (s SyntaxSet) IsEmpty() bool {
	return s.lo == 0 && s.hi == 0
}

// Predefined syntax sets for common use cases.

// StmtSet contains syntax kinds that can start a statement.
var StmtSet = SyntaxSetOf(Let, Set, Show, Import, Include, Return)

// MathExprSet contains syntax kinds that can start a math expression.
var MathExprSet = SyntaxSetOf(
	Hash,
	MathIdent,
	FieldAccess,
	Dot,
	Comma,
	Semicolon,
	LeftBrace,
	RightBrace,
	LeftParen,
	RightParen,
	MathText,
	MathShorthand,
	Linebreak,
	MathAlignPoint,
	MathPrimes,
	Escape,
	Str,
	Root,
	Bang,
)

// UnaryOpSet contains syntax kinds that are unary operators.
var UnaryOpSet = SyntaxSetOf(Plus, Minus, Not)

// BinaryOpSet contains syntax kinds that are binary operators.
var BinaryOpSet = SyntaxSetOf(
	Plus, Minus, Star, Slash, And, Or, EqEq, ExclEq,
	Lt, LtEq, Gt, GtEq, Eq, In, PlusEq, HyphEq, StarEq, SlashEq,
)

// AtomicCodePrimarySet contains syntax kinds that can start an atomic code primary.
var AtomicCodePrimarySet = SyntaxSetOf(
	Ident,
	LeftBrace,
	LeftBracket,
	LeftParen,
	Dollar,
	Let,
	Set,
	Show,
	Context,
	If,
	While,
	For,
	Import,
	Include,
	Break,
	Continue,
	Return,
	None,
	Auto,
	Int,
	Float,
	Bool,
	Numeric,
	Str,
	Label,
	Raw,
)

// CodePrimarySet contains syntax kinds that can start a code primary.
var CodePrimarySet = AtomicCodePrimarySet.Add(Underscore)

// CodeExprSet contains syntax kinds that can start a code expression.
var CodeExprSet = CodePrimarySet.Union(UnaryOpSet)

// AtomicCodeExprSet contains syntax kinds that can start an atomic code expression.
var AtomicCodeExprSet = AtomicCodePrimarySet

// ArrayOrDictItemSet contains syntax kinds that can start an array or dict item.
var ArrayOrDictItemSet = CodeExprSet.Add(Dots)

// ArgSet contains syntax kinds that can start an argument in a function call.
var ArgSet = CodeExprSet.Add(Dots)

// PatternLeafSet contains syntax kinds that can start a pattern leaf.
var PatternLeafSet = AtomicCodeExprSet

// PatternSet contains syntax kinds that can start a pattern.
var PatternSet = PatternLeafSet.Add(LeftParen).Add(Underscore)

// ParamSet contains syntax kinds that can start a parameter in a parameter list.
var ParamSet = PatternSet.Add(Dots)

// DestructuringItemSet contains syntax kinds that can start a destructuring item.
var DestructuringItemSet = PatternSet.Add(Dots)
