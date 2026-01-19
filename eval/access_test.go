package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

func TestAccessIdent(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	t.Run("access mutable binding", func(t *testing.T) {
		vm.EnterScope()
		defer vm.ExitScope()

		// Create a mutable binding
		binding := NewMutableBinding(Int(42), syntax.Detached())
		vm.Scopes.Insert("x", binding)

		// Get the binding
		b := vm.GetMut("x")
		if b == nil {
			t.Fatal("Expected to find binding")
		}

		// Modify through the binding
		b.Value = Int(100)

		// Verify the change persisted
		b2 := vm.Get("x")
		if v, _ := AsInt(b2.Value); v != 100 {
			t.Errorf("Expected 100, got %d", v)
		}
	})

	t.Run("access immutable binding", func(t *testing.T) {
		vm.EnterScope()
		defer vm.ExitScope()

		// Create an immutable binding
		binding := NewBinding(Int(42), syntax.Detached())
		vm.Scopes.Insert("y", binding)

		// Try to write - should fail
		b := vm.Get("y")
		if b == nil {
			t.Fatal("Expected to find binding")
		}

		err := b.Write(Int(100))
		if err == nil {
			t.Error("Expected error when writing to immutable binding")
		}
		if _, ok := err.(*ImmutableBindingError); !ok {
			t.Errorf("Expected ImmutableBindingError, got %T", err)
		}
	})
}

func TestAccessErrors(t *testing.T) {
	t.Run("not assignable error", func(t *testing.T) {
		err := &NotAssignableError{Span: syntax.Detached()}
		if err.Error() != "cannot mutate a temporary value" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("key not found error", func(t *testing.T) {
		err := &KeyNotFoundError{Key: "missing", Span: syntax.Detached()}
		if err.Error() != "key not found: missing" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("cannot mutate fields error", func(t *testing.T) {
		err := &CannotMutateFieldsError{Type: TypeFunc, Span: syntax.Detached()}
		if err.Error() != "cannot mutate fields on function" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("fields not mutable error", func(t *testing.T) {
		err := &FieldsNotMutableError{Type: TypeInt, Span: syntax.Detached()}
		if err.Error() != "fields on int are not yet mutable" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("missing argument error", func(t *testing.T) {
		err := &MissingArgumentError{What: "index", Span: syntax.Detached()}
		if err.Error() != "missing argument: index" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("invalid argument error", func(t *testing.T) {
		err := &InvalidArgumentError{Message: "expected number", Span: syntax.Detached()}
		if err.Error() != "expected number" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("not implemented error", func(t *testing.T) {
		err := &NotImplementedError{Feature: "array access", Span: syntax.Detached()}
		if err.Error() != "not implemented: array access" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("type mismatch error", func(t *testing.T) {
		err := &TypeMismatchError{Expected: "array", Got: "int", Span: syntax.Detached()}
		if err.Error() != "expected array, got int" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("index out of bounds error", func(t *testing.T) {
		err := &IndexOutOfBoundsError{Index: 5, Length: 3, Span: syntax.Detached()}
		if err.Error() != "index 5 is out of bounds for array of length 3" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("missing field error", func(t *testing.T) {
		err := &MissingFieldError{Span: syntax.Detached()}
		if err.Error() != "missing field name" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})
}

func TestIsAccessorMethod(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"at", true},
		{"first", true},
		{"last", true},
		{"push", false},
		{"pop", false},
		{"len", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAccessorMethod(tt.name); got != tt.expected {
				t.Errorf("isAccessorMethod(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestCallFirstAccess(t *testing.T) {
	t.Run("first element of array", func(t *testing.T) {
		arr := ArrayValue{Int(1), Int(2), Int(3)}
		var val Value = arr

		result, err := callFirstAccess(&val, syntax.Detached())
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if v, ok := AsInt(*result); !ok || v != 1 {
			t.Errorf("Expected 1, got %v", *result)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		arr := ArrayValue{}
		var val Value = arr

		_, err := callFirstAccess(&val, syntax.Detached())
		if err == nil {
			t.Error("Expected error for empty array")
		}
		if _, ok := err.(*IndexOutOfBoundsError); !ok {
			t.Errorf("Expected IndexOutOfBoundsError, got %T", err)
		}
	})

	t.Run("non-array type", func(t *testing.T) {
		var val Value = Int(42)

		_, err := callFirstAccess(&val, syntax.Detached())
		if err == nil {
			t.Error("Expected error for non-array type")
		}
		if _, ok := err.(*TypeMismatchError); !ok {
			t.Errorf("Expected TypeMismatchError, got %T", err)
		}
	})
}

func TestCallLastAccess(t *testing.T) {
	t.Run("last element of array", func(t *testing.T) {
		arr := ArrayValue{Int(1), Int(2), Int(3)}
		var val Value = arr

		result, err := callLastAccess(&val, syntax.Detached())
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if v, ok := AsInt(*result); !ok || v != 3 {
			t.Errorf("Expected 3, got %v", *result)
		}
	})

	t.Run("single element array", func(t *testing.T) {
		arr := ArrayValue{Int(42)}
		var val Value = arr

		result, err := callLastAccess(&val, syntax.Detached())
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if v, ok := AsInt(*result); !ok || v != 42 {
			t.Errorf("Expected 42, got %v", *result)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		arr := ArrayValue{}
		var val Value = arr

		_, err := callLastAccess(&val, syntax.Detached())
		if err == nil {
			t.Error("Expected error for empty array")
		}
		if _, ok := err.(*IndexOutOfBoundsError); !ok {
			t.Errorf("Expected IndexOutOfBoundsError, got %T", err)
		}
	})

	t.Run("non-array type", func(t *testing.T) {
		var val Value = Str("hello")

		_, err := callLastAccess(&val, syntax.Detached())
		if err == nil {
			t.Error("Expected error for non-array type")
		}
		if _, ok := err.(*TypeMismatchError); !ok {
			t.Errorf("Expected TypeMismatchError, got %T", err)
		}
	})
}

func TestAccessExprNil(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	_, err := AccessExpr(vm, nil)
	if err == nil {
		t.Error("Expected error for nil expression")
	}
	if _, ok := err.(*NotAssignableError); !ok {
		t.Errorf("Expected NotAssignableError, got %T", err)
	}
}
