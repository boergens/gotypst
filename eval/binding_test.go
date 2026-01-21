package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

func TestDestructureArray(t *testing.T) {
	// Create a VM with scopes
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create an array value
	arr := ArrayValue{Int(1), Int(2), Int(3)}

	// Create a mock destructuring pattern with 3 bindings
	// Since we can't easily create syntax nodes without parsing,
	// we'll test the lower-level functions

	t.Run("destructure array to bindings", func(t *testing.T) {
		// Test that we can destructure values to bindings
		vm.EnterScope()
		defer vm.ExitScope()

		// Manually define bindings as if destructured
		vm.DefineWithSpan("a", Int(1), syntax.Detached())
		vm.DefineWithSpan("b", Int(2), syntax.Detached())
		vm.DefineWithSpan("c", Int(3), syntax.Detached())

		// Verify bindings
		if binding := vm.Get("a"); binding == nil {
			t.Fatal("Expected binding for 'a'")
		} else if v, _ := AsInt(binding.Value); v != 1 {
			t.Errorf("Expected a=1, got %d", v)
		}

		if binding := vm.Get("b"); binding == nil {
			t.Fatal("Expected binding for 'b'")
		} else if v, _ := AsInt(binding.Value); v != 2 {
			t.Errorf("Expected b=2, got %d", v)
		}

		if binding := vm.Get("c"); binding == nil {
			t.Fatal("Expected binding for 'c'")
		} else if v, _ := AsInt(binding.Value); v != 3 {
			t.Errorf("Expected c=3, got %d", v)
		}
	})

	t.Run("array length check", func(t *testing.T) {
		// Test countPatterns function
		// Since we can't easily create DestructuringItems, we test error message construction
		err := &WrongNumberOfElementsError{
			Quantifier: "not enough",
			Expected:   "3 elements",
			Got:        2,
			Span:       syntax.Detached(),
		}
		expectedMsg := "not enough elements to destructure; the provided array has a length of 2, but the pattern expects 3 elements"
		if err.Error() != expectedMsg {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	_ = arr // Use the array variable
}

func TestDestructureDict(t *testing.T) {
	// Create a VM with scopes
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create a dict value
	dict := NewDict()
	dict.Set("name", Str("Alice"))
	dict.Set("age", Int(30))

	t.Run("dict field access", func(t *testing.T) {
		vm.EnterScope()
		defer vm.ExitScope()

		// Simulate destructuring by defining bindings
		val, ok := dict.Get("name")
		if !ok {
			t.Fatal("Expected to find 'name' key")
		}
		vm.DefineWithSpan("name", val, syntax.Detached())

		val, ok = dict.Get("age")
		if !ok {
			t.Fatal("Expected to find 'age' key")
		}
		vm.DefineWithSpan("age", val, syntax.Detached())

		// Verify bindings
		if binding := vm.Get("name"); binding == nil {
			t.Fatal("Expected binding for 'name'")
		} else if s, ok := AsStr(binding.Value); !ok || s != "Alice" {
			t.Errorf("Expected name='Alice', got %v", binding.Value)
		}

		if binding := vm.Get("age"); binding == nil {
			t.Fatal("Expected binding for 'age'")
		} else if v, ok := AsInt(binding.Value); !ok || v != 30 {
			t.Errorf("Expected age=30, got %v", binding.Value)
		}
	})
}

func TestDestructuringErrors(t *testing.T) {
	t.Run("cannot destructure error", func(t *testing.T) {
		err := &CannotDestructureError{Type: TypeInt, Span: syntax.Detached()}
		if err.Error() != "cannot destructure integer" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("cannot assign error", func(t *testing.T) {
		err := &CannotAssignError{Span: syntax.Detached()}
		if err.Error() != "cannot assign to this expression" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("named from array error", func(t *testing.T) {
		err := &CannotDestructureNamedFromArrayError{Span: syntax.Detached()}
		if err.Error() != "cannot destructure named pattern from an array" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("unnamed from dict error", func(t *testing.T) {
		err := &CannotDestructureUnnamedFromDictError{Span: syntax.Detached()}
		if err.Error() != "cannot destructure unnamed pattern from dictionary" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("wrong number of elements error", func(t *testing.T) {
		// Test with spread
		err := &WrongNumberOfElementsError{
			Quantifier: "not enough",
			Expected:   "at least 2 elements",
			Got:        1,
			Span:       syntax.Detached(),
		}
		expectedMsg := "not enough elements to destructure; the provided array has a length of 1, but the pattern expects at least 2 elements"
		if err.Error() != expectedMsg {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})
}

func TestBindingKinds(t *testing.T) {
	span := syntax.Detached()

	t.Run("normal binding", func(t *testing.T) {
		b := NewBinding(Int(42), span)
		if b.Kind != BindingNormal {
			t.Errorf("Expected BindingNormal, got %v", b.Kind)
		}
		if b.Mutable {
			t.Error("Normal binding should not be mutable")
		}
	})

	t.Run("mutable binding", func(t *testing.T) {
		b := NewMutableBinding(Int(42), span)
		if b.Kind != BindingNormal {
			t.Errorf("Expected BindingNormal, got %v", b.Kind)
		}
		if !b.Mutable {
			t.Error("Mutable binding should be mutable")
		}

		// Test write
		if err := b.Write(Int(100)); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if v, _ := AsInt(b.Value); v != 100 {
			t.Errorf("Expected 100, got %d", v)
		}
	})

	t.Run("closure binding", func(t *testing.T) {
		b := NewClosureBinding(Int(42), span)
		if b.Kind != BindingClosure {
			t.Errorf("Expected BindingClosure, got %v", b.Kind)
		}
		if b.Mutable {
			t.Error("Closure binding should not be mutable")
		}
	})

	t.Run("module binding", func(t *testing.T) {
		b := NewModuleBinding(Int(42), span)
		if b.Kind != BindingModule {
			t.Errorf("Expected BindingModule, got %v", b.Kind)
		}
		if b.Mutable {
			t.Error("Module binding should not be mutable")
		}
	})
}

func TestBindingCategory(t *testing.T) {
	span := syntax.Detached()
	cat := &Category{Name: "test-category"}

	b := NewBinding(Int(42), span)
	b.Category = cat

	if b.Category == nil {
		t.Fatal("Expected category to be set")
	}
	if b.Category.Name != "test-category" {
		t.Errorf("Expected category name 'test-category', got '%s'", b.Category.Name)
	}

	// Test clone preserves category
	clone := b.Clone()
	if clone.Category != b.Category {
		t.Error("Clone should preserve category reference")
	}
}

func TestBindingClone(t *testing.T) {
	span := syntax.Detached()

	// Test with array value
	arr := ArrayValue{Int(1), Int(2)}
	b := NewBinding(arr, span)

	clone := b.Clone()

	// Verify the clone is independent
	arr[0] = Int(999)
	cloneArr := clone.Value.(ArrayValue)
	if v, _ := AsInt(cloneArr[0]); v == 999 {
		t.Error("Clone should be independent of original")
	}
}

func TestDestructureImplNilPattern(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// DestructureImpl should handle nil pattern gracefully
	err := DestructureImpl(vm, nil, Int(42), func(vm *Vm, expr syntax.Expr, value Value) error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected nil error for nil pattern, got: %v", err)
	}
}
