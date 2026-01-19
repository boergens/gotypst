package syntax

import "testing"

func TestUnOpString(t *testing.T) {
	tests := []struct {
		op   UnOp
		want string
	}{
		{UnOpPos, "+"},
		{UnOpNeg, "-"},
		{UnOpNot, "not"},
	}

	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestUnOpName(t *testing.T) {
	tests := []struct {
		op   UnOp
		want string
	}{
		{UnOpPos, "positive"},
		{UnOpNeg, "negation"},
		{UnOpNot, "logical not"},
	}

	for _, tt := range tests {
		if got := tt.op.Name(); got != tt.want {
			t.Errorf("%v.Name() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestBinOpString(t *testing.T) {
	tests := []struct {
		op   BinOp
		want string
	}{
		{BinOpAdd, "+"},
		{BinOpSub, "-"},
		{BinOpMul, "*"},
		{BinOpDiv, "/"},
		{BinOpAnd, "and"},
		{BinOpOr, "or"},
		{BinOpEq, "=="},
		{BinOpNeq, "!="},
		{BinOpLt, "<"},
		{BinOpLeq, "<="},
		{BinOpGt, ">"},
		{BinOpGeq, ">="},
		{BinOpAssign, "="},
		{BinOpAddAssign, "+="},
		{BinOpIn, "in"},
		{BinOpNotIn, "not in"},
	}

	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestBinOpAssoc(t *testing.T) {
	// Assignment operators are right-associative
	rightAssoc := []BinOp{BinOpAssign, BinOpAddAssign, BinOpSubAssign, BinOpMulAssign, BinOpDivAssign}
	for _, op := range rightAssoc {
		if op.Assoc() != AssocRight {
			t.Errorf("%v should be right-associative", op)
		}
	}

	// Most operators are left-associative
	leftAssoc := []BinOp{BinOpAdd, BinOpSub, BinOpMul, BinOpDiv, BinOpAnd, BinOpOr}
	for _, op := range leftAssoc {
		if op.Assoc() != AssocLeft {
			t.Errorf("%v should be left-associative", op)
		}
	}
}

func TestBinOpPrecedence(t *testing.T) {
	// Verify multiplication has higher precedence than addition
	if BinOpMul.Precedence() <= BinOpAdd.Precedence() {
		t.Error("* should have higher precedence than +")
	}

	// Verify comparison has lower precedence than arithmetic
	if BinOpLt.Precedence() >= BinOpAdd.Precedence() {
		t.Error("< should have lower precedence than +")
	}

	// Verify and has lower precedence than comparison
	if BinOpAnd.Precedence() >= BinOpLt.Precedence() {
		t.Error("and should have lower precedence than <")
	}

	// Verify or has lowest precedence among non-assignment
	if BinOpOr.Precedence() >= BinOpAnd.Precedence() {
		t.Error("or should have lower precedence than and")
	}
}

func TestBinOpIsComparison(t *testing.T) {
	compOps := []BinOp{BinOpEq, BinOpNeq, BinOpLt, BinOpLeq, BinOpGt, BinOpGeq}
	for _, op := range compOps {
		if !op.IsComparison() {
			t.Errorf("%v should be a comparison operator", op)
		}
	}

	nonCompOps := []BinOp{BinOpAdd, BinOpMul, BinOpAnd, BinOpAssign, BinOpIn}
	for _, op := range nonCompOps {
		if op.IsComparison() {
			t.Errorf("%v should not be a comparison operator", op)
		}
	}
}

func TestBinOpIsAssignment(t *testing.T) {
	assignOps := []BinOp{BinOpAssign, BinOpAddAssign, BinOpSubAssign, BinOpMulAssign, BinOpDivAssign}
	for _, op := range assignOps {
		if !op.IsAssignment() {
			t.Errorf("%v should be an assignment operator", op)
		}
	}

	nonAssignOps := []BinOp{BinOpAdd, BinOpEq, BinOpAnd, BinOpIn}
	for _, op := range nonAssignOps {
		if op.IsAssignment() {
			t.Errorf("%v should not be an assignment operator", op)
		}
	}
}

func TestBinOpFromSyntaxKind(t *testing.T) {
	tests := []struct {
		kind SyntaxKind
		want BinOp
	}{
		{Plus, BinOpAdd},
		{Minus, BinOpSub},
		{Star, BinOpMul},
		{Slash, BinOpDiv},
		{And, BinOpAnd},
		{Or, BinOpOr},
		{EqEq, BinOpEq},
		{ExclEq, BinOpNeq},
		{Lt, BinOpLt},
		{LtEq, BinOpLeq},
		{Gt, BinOpGt},
		{GtEq, BinOpGeq},
		{Eq, BinOpAssign},
		{PlusEq, BinOpAddAssign},
		{In, BinOpIn},
	}

	for _, tt := range tests {
		got := BinOpFromSyntaxKind(tt.kind)
		if got != tt.want {
			t.Errorf("BinOpFromSyntaxKind(%v) = %v, want %v", tt.kind, got, tt.want)
		}
	}

	// Test invalid kind
	if BinOpFromSyntaxKind(Ident) != -1 {
		t.Error("BinOpFromSyntaxKind(Ident) should return -1")
	}
}

func TestAssocString(t *testing.T) {
	if AssocLeft.String() != "left" {
		t.Errorf("AssocLeft.String() = %q, want %q", AssocLeft.String(), "left")
	}
	if AssocRight.String() != "right" {
		t.Errorf("AssocRight.String() = %q, want %q", AssocRight.String(), "right")
	}
}
