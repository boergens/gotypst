package syntax

import "testing"

func TestSyntaxSetNew(t *testing.T) {
	s := NewSyntaxSet()
	if !s.IsEmpty() {
		t.Error("NewSyntaxSet() should create an empty set")
	}
}

func TestSyntaxSetAdd(t *testing.T) {
	s := NewSyntaxSet().Add(And).Add(Or)
	if !s.Contains(And) {
		t.Error("set should contain And")
	}
	if !s.Contains(Or) {
		t.Error("set should contain Or")
	}
	if s.Contains(Not) {
		t.Error("set should not contain Not")
	}
}

func TestSyntaxSetOf(t *testing.T) {
	s := SyntaxSetOf(And, Or, Not)
	if !s.Contains(And) {
		t.Error("set should contain And")
	}
	if !s.Contains(Or) {
		t.Error("set should contain Or")
	}
	if !s.Contains(Not) {
		t.Error("set should contain Not")
	}
	if s.Contains(Plus) {
		t.Error("set should not contain Plus")
	}
}

func TestSyntaxSetRemove(t *testing.T) {
	s := SyntaxSetOf(And, Or, Not)
	s = s.Remove(Or)
	if !s.Contains(And) {
		t.Error("set should still contain And")
	}
	if s.Contains(Or) {
		t.Error("set should not contain Or after removal")
	}
	if !s.Contains(Not) {
		t.Error("set should still contain Not")
	}
}

func TestSyntaxSetUnion(t *testing.T) {
	s1 := SyntaxSetOf(And, Or)
	s2 := SyntaxSetOf(Not, Plus)
	s := s1.Union(s2)
	for _, k := range []SyntaxKind{And, Or, Not, Plus} {
		if !s.Contains(k) {
			t.Errorf("union should contain %s", k.Name())
		}
	}
}

func TestSyntaxSetContainsHighBits(t *testing.T) {
	// Test kinds with discriminator >= 64 (use the hi bits)
	// Let's use some kinds that we know are >= 64
	// Code = 95, Ident = 96, etc.
	s := SyntaxSetOf(Code, Ident, Bool)
	if !s.Contains(Code) {
		t.Error("set should contain Code")
	}
	if !s.Contains(Ident) {
		t.Error("set should contain Ident")
	}
	if !s.Contains(Bool) {
		t.Error("set should contain Bool")
	}
	if s.Contains(And) {
		t.Error("set should not contain And")
	}
}

func TestSyntaxSetContainsOutOfRange(t *testing.T) {
	s := SyntaxSetOf(And, Or)
	// Kinds >= 128 should always return false (cannot be in set)
	if s.Contains(DestructAssignment) {
		t.Error("set.Contains should return false for kinds >= 128")
	}
}

func TestSyntaxSetAddPanicsForHighKinds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Add should panic for kinds >= 128")
		}
	}()
	_ = NewSyntaxSet().Add(DestructAssignment)
}

func TestSyntaxSetRemovePanicsForHighKinds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Remove should panic for kinds >= 128")
		}
	}()
	_ = NewSyntaxSet().Remove(DestructAssignment)
}

func TestSyntaxSetIsEmpty(t *testing.T) {
	s := NewSyntaxSet()
	if !s.IsEmpty() {
		t.Error("new set should be empty")
	}
	s = s.Add(And)
	if s.IsEmpty() {
		t.Error("set with And should not be empty")
	}
	s = s.Remove(And)
	if !s.IsEmpty() {
		t.Error("set after removing And should be empty")
	}
}

func TestPredefinedSets(t *testing.T) {
	// Test that predefined sets contain expected kinds
	if !StmtSet.Contains(Let) {
		t.Error("StmtSet should contain Let")
	}
	if !UnaryOpSet.Contains(Not) {
		t.Error("UnaryOpSet should contain Not")
	}
	if !BinaryOpSet.Contains(Plus) {
		t.Error("BinaryOpSet should contain Plus")
	}
	if !CodeExprSet.Contains(Ident) {
		t.Error("CodeExprSet should contain Ident")
	}
}
