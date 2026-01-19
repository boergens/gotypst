package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

func TestValueTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		wantType Type
	}{
		{"None", None, TypeNone},
		{"Auto", Auto, TypeAuto},
		{"True", True, TypeBool},
		{"False", False, TypeBool},
		{"Int", Int(42), TypeInt},
		{"Float", Float(3.14), TypeFloat},
		{"Str", Str("hello"), TypeStr},
		{"Array", ArrayValue{Int(1), Int(2)}, TypeArray},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.Type(); got != tt.wantType {
				t.Errorf("Type() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

func TestValueClone(t *testing.T) {
	// Test that Clone creates independent copies
	arr := ArrayValue{Int(1), Int(2), Int(3)}
	clone := arr.Clone().(ArrayValue)

	if len(clone) != len(arr) {
		t.Errorf("Clone length = %d, want %d", len(clone), len(arr))
	}

	// Verify they're independent copies
	clone[0] = Int(999)
	if arr[0] == clone[0] {
		t.Error("Clone should create independent copy")
	}
}

func TestScope(t *testing.T) {
	scope := NewScope()

	// Test Define and Get
	scope.Define("x", Int(42), syntax.Detached())
	binding := scope.Get("x")
	if binding == nil {
		t.Fatal("Expected binding, got nil")
	}

	if v, ok := binding.Value.(IntValue); !ok || v != 42 {
		t.Errorf("Got %v, want Int(42)", binding.Value)
	}

	// Test Contains
	if !scope.Contains("x") {
		t.Error("Expected scope to contain 'x'")
	}
	if scope.Contains("y") {
		t.Error("Expected scope to not contain 'y'")
	}

	// Test Names
	scope.Define("y", Str("hello"), syntax.Detached())
	names := scope.Names()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
}

func TestScopes(t *testing.T) {
	base := NewScope()
	base.Define("stdlib", Str("standard library"), syntax.Detached())

	scopes := NewScopes(base)

	// Can access base scope
	if binding := scopes.Get("stdlib"); binding == nil {
		t.Error("Expected to find 'stdlib' from base scope")
	}

	// Define in top scope
	scopes.Define("x", Int(1), syntax.Detached())

	// Enter new scope and shadow
	scopes.Enter()
	scopes.Define("x", Int(2), syntax.Detached())

	binding := scopes.Get("x")
	if binding == nil {
		t.Fatal("Expected binding, got nil")
	}
	if v, ok := binding.Value.(IntValue); !ok || v != 2 {
		t.Errorf("Got %v, want Int(2)", binding.Value)
	}

	// Exit scope and see original
	scopes.Exit()
	binding = scopes.Get("x")
	if binding == nil {
		t.Fatal("Expected binding, got nil")
	}
	if v, ok := binding.Value.(IntValue); !ok || v != 1 {
		t.Errorf("Got %v, want Int(1)", binding.Value)
	}
}

func TestFlowEvents(t *testing.T) {
	span := syntax.Detached()

	breakEvent := NewBreakEvent(span)
	if !IsBreak(breakEvent) {
		t.Error("Expected break event")
	}
	if !IsLoopFlow(breakEvent) {
		t.Error("Expected loop flow for break")
	}

	continueEvent := NewContinueEvent(span)
	if !IsContinue(continueEvent) {
		t.Error("Expected continue event")
	}
	if !IsLoopFlow(continueEvent) {
		t.Error("Expected loop flow for continue")
	}

	returnEvent := NewReturnEventWithValue(span, Int(42))
	if !IsReturn(returnEvent) {
		t.Error("Expected return event")
	}
	if IsLoopFlow(returnEvent) {
		t.Error("Return should not be loop flow")
	}
}

func TestDict(t *testing.T) {
	dict := NewDict()

	// Test Set and Get
	dict.Set("name", Str("Alice"))
	dict.Set("age", Int(30))

	if v, ok := dict.Get("name"); !ok {
		t.Error("Expected to find 'name'")
	} else if s, ok := v.(StrValue); !ok || s != "Alice" {
		t.Errorf("Got %v, want Str('Alice')", v)
	}

	// Test update existing key
	dict.Set("name", Str("Bob"))
	if v, _ := dict.Get("name"); v.(StrValue) != "Bob" {
		t.Error("Expected name to be updated to 'Bob'")
	}

	// Test Len
	if dict.Len() != 2 {
		t.Errorf("Expected length 2, got %d", dict.Len())
	}

	// Test Keys order
	keys := dict.Keys()
	if len(keys) != 2 || keys[0] != "name" || keys[1] != "age" {
		t.Errorf("Keys = %v, want [name, age]", keys)
	}
}

func TestVm(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Test Define and Get
	vm.Define("x", Int(42))
	if binding := vm.Get("x"); binding == nil {
		t.Fatal("Expected binding, got nil")
	}

	// Test call depth
	if err := vm.CheckCallDepth(); err != nil {
		t.Error("Expected no error for initial call depth")
	}

	// Test scope operations
	vm.EnterScope()
	vm.Define("y", Int(100))
	if vm.Get("y") == nil {
		t.Error("Expected to find 'y' in new scope")
	}
	vm.ExitScope()
	if vm.Get("y") != nil {
		t.Error("Expected 'y' to not be visible after exit")
	}

	// Test flow events
	if vm.HasFlow() {
		t.Error("Expected no flow initially")
	}
	vm.SetFlow(NewBreakEvent(syntax.Detached()))
	if !vm.HasFlow() {
		t.Error("Expected flow after SetFlow")
	}
	flow := vm.TakeFlow()
	if !IsBreak(flow) {
		t.Error("Expected break flow")
	}
	if vm.HasFlow() {
		t.Error("Expected no flow after TakeFlow")
	}
}

func TestBinding(t *testing.T) {
	span := syntax.Detached()

	// Test immutable binding
	immutable := NewBinding(Int(42), span)
	if err := immutable.Write(Int(100)); err == nil {
		t.Error("Expected error writing to immutable binding")
	}

	// Test mutable binding
	mutable := NewMutableBinding(Int(42), span)
	if err := mutable.Write(Int(100)); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	v, _ := mutable.Read()
	if v.(IntValue) != 100 {
		t.Errorf("Got %v, want Int(100)", v)
	}
}
