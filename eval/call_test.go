package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

func TestArgsBasic(t *testing.T) {
	span := syntax.Span{}
	args := NewArgs(span)

	if !args.IsEmpty() {
		t.Error("New args should be empty")
	}

	// Test Push and Eat
	args.Push(Int(42), span)
	if args.IsEmpty() {
		t.Error("Args should not be empty after push")
	}

	eaten := args.Eat()
	if eaten == nil {
		t.Error("Eat should return a value")
	}
	if val, ok := eaten.V.(IntValue); !ok || val != 42 {
		t.Errorf("Expected Int(42), got %v", eaten.V)
	}

	if !args.IsEmpty() {
		t.Error("Args should be empty after eating the only argument")
	}
}

func TestArgsNamed(t *testing.T) {
	span := syntax.Span{}
	args := NewArgs(span)

	args.PushNamed("foo", Int(1), span)
	args.PushNamed("bar", Int(2), span)

	// Find named argument
	found := args.Find("foo")
	if found == nil {
		t.Error("Should find named argument 'foo'")
	}
	if val, ok := found.V.(IntValue); !ok || val != 1 {
		t.Errorf("Expected Int(1), got %v", found.V)
	}

	// Should not find it again (removed)
	found = args.Find("foo")
	if found != nil {
		t.Error("Should not find 'foo' after removal")
	}

	// bar should still be there
	found = args.Find("bar")
	if found == nil {
		t.Error("Should find named argument 'bar'")
	}
}

func TestArgsExpect(t *testing.T) {
	span := syntax.Span{}
	args := NewArgs(span)

	// Expect on empty args should error
	_, err := args.Expect("value")
	if err == nil {
		t.Error("Expect on empty args should return error")
	}

	// Add positional argument
	args.Push(Str("hello"), span)
	val, err := args.Expect("value")
	if err != nil {
		t.Errorf("Expect should succeed: %v", err)
	}
	if s, ok := val.V.(StrValue); !ok || string(s) != "hello" {
		t.Errorf("Expected Str(hello), got %v", val.V)
	}
}

func TestArgsAll(t *testing.T) {
	span := syntax.Span{}
	args := NewArgs(span)

	args.Push(Int(1), span)
	args.Push(Int(2), span)
	args.Push(Int(3), span)
	args.PushNamed("named", Int(4), span)

	// All should return only positional
	all := args.All()
	if len(all) != 3 {
		t.Errorf("Expected 3 positional args, got %d", len(all))
	}

	// Named should still be there
	if args.Remaining() != 1 {
		t.Errorf("Expected 1 remaining arg, got %d", args.Remaining())
	}
}

func TestArgsFinish(t *testing.T) {
	span := syntax.Span{}
	args := NewArgs(span)

	// Empty args should finish without error
	if err := args.Finish(); err != nil {
		t.Errorf("Empty args should finish cleanly: %v", err)
	}

	// Args with remaining should error
	args.Push(Int(1), span)
	if err := args.Finish(); err == nil {
		t.Error("Args with remaining positional should error")
	}

	args = NewArgs(span)
	args.PushNamed("extra", Int(1), span)
	if err := args.Finish(); err == nil {
		t.Error("Args with remaining named should error")
	}
}

func TestCallNativeFunc(t *testing.T) {
	// Create a simple native function that adds two numbers
	addFunc := &Func{
		Name: ptr("add"),
		Span: syntax.Span{},
		Repr: NativeFunc{
			Func: func(engine *Engine, context *Context, args *Args) (Value, error) {
				a, err := args.Expect("a")
				if err != nil {
					return nil, err
				}
				b, err := args.Expect("b")
				if err != nil {
					return nil, err
				}
				aInt, ok := a.V.(IntValue)
				if !ok {
					return nil, &TypeMismatchError{Expected: "int", Got: a.V.Type().String()}
				}
				bInt, ok := b.V.(IntValue)
				if !ok {
					return nil, &TypeMismatchError{Expected: "int", Got: b.V.Type().String()}
				}
				return Int(int64(aInt) + int64(bInt)), nil
			},
		},
	}

	// Create VM and call the function
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	args := NewArgs(syntax.Span{})
	args.Push(Int(10), syntax.Span{})
	args.Push(Int(20), syntax.Span{})

	result, err := CallFunc(vm, addFunc, args)
	if err != nil {
		t.Fatalf("CallFunc failed: %v", err)
	}

	intResult, ok := result.(IntValue)
	if !ok {
		t.Fatalf("Expected IntValue, got %T", result)
	}
	if intResult != 30 {
		t.Errorf("Expected 30, got %d", intResult)
	}
}

func TestCallNativeFuncMissingArg(t *testing.T) {
	addFunc := &Func{
		Name: ptr("add"),
		Span: syntax.Span{},
		Repr: NativeFunc{
			Func: func(engine *Engine, context *Context, args *Args) (Value, error) {
				_, err := args.Expect("a")
				if err != nil {
					return nil, err
				}
				_, err = args.Expect("b")
				if err != nil {
					return nil, err
				}
				return Int(0), nil
			},
		},
	}

	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	args := NewArgs(syntax.Span{})
	args.Push(Int(10), syntax.Span{}) // Only one arg

	_, err := CallFunc(vm, addFunc, args)
	if err == nil {
		t.Error("Expected error for missing argument")
	}
	if _, ok := err.(*MissingArgumentError); !ok {
		t.Errorf("Expected MissingArgumentError, got %T", err)
	}
}

func TestCallDepthLimit(t *testing.T) {
	// Create a recursive function that exceeds call depth
	var recursiveFunc *Func
	recursiveFunc = &Func{
		Name: ptr("recurse"),
		Span: syntax.Span{},
		Repr: NativeFunc{
			Func: func(engine *Engine, context *Context, args *Args) (Value, error) {
				// Use engine.CallFunc for recursive calls
				return engine.CallFunc(context, FuncValue{Func: recursiveFunc}, NewArgs(syntax.Span{}), syntax.Span{})
			},
		},
	}

	engine := NewEngine(nil)
	_, err := engine.CallFunc(nil, FuncValue{Func: recursiveFunc}, NewArgs(syntax.Span{}), syntax.Span{})
	if err == nil {
		t.Error("Expected error for exceeding call depth")
	}
	if _, ok := err.(*CallDepthExceededError); !ok {
		t.Errorf("Expected CallDepthExceededError, got %T: %v", err, err)
	}
}

func TestBinaryOperators(t *testing.T) {
	span := syntax.Span{}

	tests := []struct {
		name string
		op   syntax.BinOp
		lhs  Value
		rhs  Value
		want Value
	}{
		{"int+int", syntax.BinOpAdd, Int(1), Int(2), Int(3)},
		{"int-int", syntax.BinOpSub, Int(5), Int(3), Int(2)},
		{"int*int", syntax.BinOpMul, Int(3), Int(4), Int(12)},
		{"float+float", syntax.BinOpAdd, Float(1.5), Float(2.5), Float(4.0)},
		{"int+float", syntax.BinOpAdd, Int(1), Float(2.5), Float(3.5)},
		{"str+str", syntax.BinOpAdd, Str("hello"), Str(" world"), Str("hello world")},
		// Note: and/or are short-circuit operators and handled separately, not via applyBinaryOp
		{"int==int true", syntax.BinOpEq, Int(5), Int(5), True},
		{"int==int false", syntax.BinOpEq, Int(5), Int(6), False},
		{"int<int true", syntax.BinOpLt, Int(3), Int(5), True},
		{"int<int false", syntax.BinOpLt, Int(5), Int(3), False},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyBinaryOp(tt.op, tt.lhs, tt.rhs, span)
			if err != nil {
				t.Fatalf("applyBinaryOp failed: %v", err)
			}
			if !valuesEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDivisionByZero(t *testing.T) {
	span := syntax.Span{}

	_, err := applyBinaryOp(syntax.BinOpDiv, Int(10), Int(0), span)
	if err == nil {
		t.Error("Expected error for division by zero")
	}
	if _, ok := err.(*DivisionByZeroError); !ok {
		t.Errorf("Expected DivisionByZeroError, got %T", err)
	}
}

func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name string
		lhs  Value
		rhs  Value
		want bool
	}{
		{"none==none", None, None, true},
		{"auto==auto", Auto, Auto, true},
		{"int==int same", Int(5), Int(5), true},
		{"int==int diff", Int(5), Int(6), false},
		{"float==float same", Float(3.14), Float(3.14), true},
		{"str==str same", Str("hello"), Str("hello"), true},
		{"str==str diff", Str("hello"), Str("world"), false},
		{"bool==bool same", True, True, true},
		{"bool==bool diff", True, False, false},
		// Different types
		{"int==str", Int(5), Str("5"), false},
		{"none==auto", None, Auto, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesEqual(tt.lhs, tt.rhs)
			if got != tt.want {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.lhs, tt.rhs, got, tt.want)
			}
		})
	}
}

func TestArgsClone(t *testing.T) {
	span := syntax.Span{}
	args := NewArgs(span)
	args.Push(Int(1), span)
	args.PushNamed("foo", Int(2), span)

	clone := args.Clone()

	// Modify original
	args.Push(Int(3), span)

	// Clone should not be affected
	if clone.Remaining() != 2 {
		t.Errorf("Clone should have 2 args, got %d", clone.Remaining())
	}
}

func ptr(s string) *string {
	return &s
}
