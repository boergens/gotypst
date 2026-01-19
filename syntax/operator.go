package syntax

// UnOp represents a unary operator.
type UnOp int

const (
	// UnOpPos is the positive operator (+).
	UnOpPos UnOp = iota
	// UnOpNeg is the negative operator (-).
	UnOpNeg
	// UnOpNot is the logical not operator (not).
	UnOpNot
)

// String returns the string representation of the operator.
func (op UnOp) String() string {
	switch op {
	case UnOpPos:
		return "+"
	case UnOpNeg:
		return "-"
	case UnOpNot:
		return "not"
	default:
		return "unknown"
	}
}

// Name returns a human-readable name for the operator.
func (op UnOp) Name() string {
	switch op {
	case UnOpPos:
		return "positive"
	case UnOpNeg:
		return "negation"
	case UnOpNot:
		return "logical not"
	default:
		return "unknown"
	}
}

// BinOp represents a binary operator.
type BinOp int

const (
	// Arithmetic operators
	BinOpAdd BinOp = iota // +
	BinOpSub              // -
	BinOpMul              // *
	BinOpDiv              // /

	// Logical operators
	BinOpAnd // and
	BinOpOr  // or

	// Comparison operators
	BinOpEq  // ==
	BinOpNeq // !=
	BinOpLt  // <
	BinOpLeq // <=
	BinOpGt  // >
	BinOpGeq // >=

	// Assignment operators
	BinOpAssign    // =
	BinOpAddAssign // +=
	BinOpSubAssign // -=
	BinOpMulAssign // *=
	BinOpDivAssign // /=

	// Membership operators
	BinOpIn    // in
	BinOpNotIn // not in
)

// String returns the string representation of the operator.
func (op BinOp) String() string {
	switch op {
	case BinOpAdd:
		return "+"
	case BinOpSub:
		return "-"
	case BinOpMul:
		return "*"
	case BinOpDiv:
		return "/"
	case BinOpAnd:
		return "and"
	case BinOpOr:
		return "or"
	case BinOpEq:
		return "=="
	case BinOpNeq:
		return "!="
	case BinOpLt:
		return "<"
	case BinOpLeq:
		return "<="
	case BinOpGt:
		return ">"
	case BinOpGeq:
		return ">="
	case BinOpAssign:
		return "="
	case BinOpAddAssign:
		return "+="
	case BinOpSubAssign:
		return "-="
	case BinOpMulAssign:
		return "*="
	case BinOpDivAssign:
		return "/="
	case BinOpIn:
		return "in"
	case BinOpNotIn:
		return "not in"
	default:
		return "unknown"
	}
}

// Name returns a human-readable name for the operator.
func (op BinOp) Name() string {
	switch op {
	case BinOpAdd:
		return "addition"
	case BinOpSub:
		return "subtraction"
	case BinOpMul:
		return "multiplication"
	case BinOpDiv:
		return "division"
	case BinOpAnd:
		return "logical and"
	case BinOpOr:
		return "logical or"
	case BinOpEq:
		return "equality"
	case BinOpNeq:
		return "inequality"
	case BinOpLt:
		return "less than"
	case BinOpLeq:
		return "less than or equal"
	case BinOpGt:
		return "greater than"
	case BinOpGeq:
		return "greater than or equal"
	case BinOpAssign:
		return "assignment"
	case BinOpAddAssign:
		return "add-assign"
	case BinOpSubAssign:
		return "subtract-assign"
	case BinOpMulAssign:
		return "multiply-assign"
	case BinOpDivAssign:
		return "divide-assign"
	case BinOpIn:
		return "membership"
	case BinOpNotIn:
		return "non-membership"
	default:
		return "unknown"
	}
}

// Assoc returns the associativity of the operator.
func (op BinOp) Assoc() Assoc {
	switch op {
	case BinOpAssign, BinOpAddAssign, BinOpSubAssign, BinOpMulAssign, BinOpDivAssign:
		return AssocRight
	default:
		return AssocLeft
	}
}

// Precedence returns the precedence of the operator.
// Higher values bind more tightly.
func (op BinOp) Precedence() int {
	switch op {
	case BinOpOr:
		return 1
	case BinOpAnd:
		return 2
	case BinOpEq, BinOpNeq, BinOpLt, BinOpLeq, BinOpGt, BinOpGeq, BinOpIn, BinOpNotIn:
		return 3
	case BinOpAdd, BinOpSub:
		return 4
	case BinOpMul, BinOpDiv:
		return 5
	case BinOpAssign, BinOpAddAssign, BinOpSubAssign, BinOpMulAssign, BinOpDivAssign:
		return 0
	default:
		return 0
	}
}

// IsComparison returns true if this is a comparison operator.
func (op BinOp) IsComparison() bool {
	switch op {
	case BinOpEq, BinOpNeq, BinOpLt, BinOpLeq, BinOpGt, BinOpGeq:
		return true
	}
	return false
}

// IsAssignment returns true if this is an assignment operator.
func (op BinOp) IsAssignment() bool {
	switch op {
	case BinOpAssign, BinOpAddAssign, BinOpSubAssign, BinOpMulAssign, BinOpDivAssign:
		return true
	}
	return false
}

// BinOpFromSyntaxKind converts a SyntaxKind to a BinOp.
// Returns -1 if the kind is not a binary operator.
func BinOpFromSyntaxKind(kind SyntaxKind) BinOp {
	switch kind {
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
	case PlusEq:
		return BinOpAddAssign
	case HyphEq:
		return BinOpSubAssign
	case StarEq:
		return BinOpMulAssign
	case SlashEq:
		return BinOpDivAssign
	case In:
		return BinOpIn
	default:
		return -1
	}
}

// Assoc represents operator associativity.
type Assoc int

const (
	// AssocLeft means operators of the same precedence are evaluated left-to-right.
	AssocLeft Assoc = iota
	// AssocRight means operators of the same precedence are evaluated right-to-left.
	AssocRight
)

// String returns the string representation of the associativity.
func (a Assoc) String() string {
	switch a {
	case AssocLeft:
		return "left"
	case AssocRight:
		return "right"
	default:
		return "unknown"
	}
}
