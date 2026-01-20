package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

func TestDictLen(t *testing.T) {
	d := NewDict()
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})

	// Empty dict
	method := getDictMethod(&d, "len", syntax.Detached())
	if method == nil {
		t.Fatal("len method should not be nil")
	}
	fn, ok := AsFunc(method)
	if !ok {
		t.Fatal("len method should be a function")
	}

	result, err := CallFunc(vm, fn, NewArgs(syntax.Span{}))
	if err != nil {
		t.Fatalf("len() failed: %v", err)
	}
	if v, ok := result.(IntValue); !ok || v != 0 {
		t.Errorf("Expected 0, got %v", result)
	}

	// Add some entries
	d.Set("a", Int(1))
	d.Set("b", Int(2))

	result, err = CallFunc(vm, fn, NewArgs(syntax.Span{}))
	if err != nil {
		t.Fatalf("len() failed: %v", err)
	}
	if v, ok := result.(IntValue); !ok || v != 2 {
		t.Errorf("Expected 2, got %v", result)
	}
}

func TestDictIsEmpty(t *testing.T) {
	d := NewDict()
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})

	method := getDictMethod(&d, "is-empty", syntax.Detached())
	if method == nil {
		t.Fatal("is-empty method should not be nil")
	}
	fn, ok := AsFunc(method)
	if !ok {
		t.Fatal("is-empty method should be a function")
	}

	// Empty dict
	result, err := CallFunc(vm, fn, NewArgs(syntax.Span{}))
	if err != nil {
		t.Fatalf("is-empty() failed: %v", err)
	}
	if v, ok := result.(BoolValue); !ok || !bool(v) {
		t.Errorf("Expected true, got %v", result)
	}

	// Non-empty dict
	d.Set("a", Int(1))
	result, err = CallFunc(vm, fn, NewArgs(syntax.Span{}))
	if err != nil {
		t.Fatalf("is-empty() failed: %v", err)
	}
	if v, ok := result.(BoolValue); !ok || bool(v) {
		t.Errorf("Expected false, got %v", result)
	}
}

func TestDictAt(t *testing.T) {
	d := NewDict()
	d.Set("foo", Int(42))
	d.Set("bar", Str("hello"))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	span := syntax.Span{}

	method := getDictMethod(&d, "at", syntax.Detached())
	if method == nil {
		t.Fatal("at method should not be nil")
	}
	fn, ok := AsFunc(method)
	if !ok {
		t.Fatal("at method should be a function")
	}

	// Get existing key
	args := NewArgs(span)
	args.Push(Str("foo"), span)
	result, err := CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("at('foo') failed: %v", err)
	}
	if v, ok := result.(IntValue); !ok || v != 42 {
		t.Errorf("Expected 42, got %v", result)
	}

	// Get non-existing key without default - should error
	args = NewArgs(span)
	args.Push(Str("missing"), span)
	_, err = CallFunc(vm, fn, args)
	if err == nil {
		t.Error("at('missing') should return error")
	}
	if _, ok := err.(*KeyNotFoundError); !ok {
		t.Errorf("Expected KeyNotFoundError, got %T", err)
	}

	// Get non-existing key with default
	args = NewArgs(span)
	args.Push(Str("missing"), span)
	args.PushNamed("default", Int(99), span)
	result, err = CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("at('missing', default: 99) failed: %v", err)
	}
	if v, ok := result.(IntValue); !ok || v != 99 {
		t.Errorf("Expected 99, got %v", result)
	}
}

func TestDictGet(t *testing.T) {
	d := NewDict()
	d.Set("foo", Int(42))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	span := syntax.Span{}

	method := getDictMethod(&d, "get", syntax.Detached())
	fn, _ := AsFunc(method)

	// Get existing key
	args := NewArgs(span)
	args.Push(Str("foo"), span)
	result, err := CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("get('foo') failed: %v", err)
	}
	if v, ok := result.(IntValue); !ok || v != 42 {
		t.Errorf("Expected 42, got %v", result)
	}

	// Get non-existing key - should return none (no error)
	args = NewArgs(span)
	args.Push(Str("missing"), span)
	result, err = CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("get('missing') failed: %v", err)
	}
	if !IsNone(result) {
		t.Errorf("Expected none, got %v", result)
	}

	// Get non-existing key with custom default
	args = NewArgs(span)
	args.Push(Str("missing"), span)
	args.PushNamed("default", Int(77), span)
	result, err = CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("get('missing', default: 77) failed: %v", err)
	}
	if v, ok := result.(IntValue); !ok || v != 77 {
		t.Errorf("Expected 77, got %v", result)
	}
}

func TestDictContains(t *testing.T) {
	d := NewDict()
	d.Set("foo", Int(42))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	span := syntax.Span{}

	method := getDictMethod(&d, "contains", syntax.Detached())
	fn, _ := AsFunc(method)

	// Contains existing key
	args := NewArgs(span)
	args.Push(Str("foo"), span)
	result, err := CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("contains('foo') failed: %v", err)
	}
	if v, ok := result.(BoolValue); !ok || !bool(v) {
		t.Errorf("Expected true, got %v", result)
	}

	// Contains non-existing key
	args = NewArgs(span)
	args.Push(Str("bar"), span)
	result, err = CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("contains('bar') failed: %v", err)
	}
	if v, ok := result.(BoolValue); !ok || bool(v) {
		t.Errorf("Expected false, got %v", result)
	}
}

func TestDictInsert(t *testing.T) {
	d := NewDict()
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	span := syntax.Span{}

	method := getDictMethod(&d, "insert", syntax.Detached())
	fn, _ := AsFunc(method)

	// Insert new key
	args := NewArgs(span)
	args.Push(Str("foo"), span)
	args.Push(Int(42), span)
	result, err := CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("insert('foo', 42) failed: %v", err)
	}
	if !IsNone(result) {
		t.Errorf("Expected none, got %v", result)
	}

	// Verify insertion
	val, ok := d.Get("foo")
	if !ok {
		t.Error("Key 'foo' should exist")
	}
	if v, ok := val.(IntValue); !ok || v != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	// Update existing key
	args = NewArgs(span)
	args.Push(Str("foo"), span)
	args.Push(Int(99), span)
	_, err = CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("insert('foo', 99) failed: %v", err)
	}

	val, _ = d.Get("foo")
	if v, ok := val.(IntValue); !ok || v != 99 {
		t.Errorf("Expected 99, got %v", val)
	}
}

func TestDictRemove(t *testing.T) {
	d := NewDict()
	d.Set("foo", Int(42))
	d.Set("bar", Int(99))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	span := syntax.Span{}

	method := getDictMethod(&d, "remove", syntax.Detached())
	fn, _ := AsFunc(method)

	// Remove existing key
	args := NewArgs(span)
	args.Push(Str("foo"), span)
	result, err := CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("remove('foo') failed: %v", err)
	}
	if v, ok := result.(IntValue); !ok || v != 42 {
		t.Errorf("Expected 42, got %v", result)
	}

	// Verify removal
	_, ok := d.Get("foo")
	if ok {
		t.Error("Key 'foo' should be removed")
	}
	if d.Len() != 1 {
		t.Errorf("Dict should have 1 entry, got %d", d.Len())
	}

	// Remove non-existing key - should error
	args = NewArgs(span)
	args.Push(Str("missing"), span)
	_, err = CallFunc(vm, fn, args)
	if err == nil {
		t.Error("remove('missing') should return error")
	}

	// Remove non-existing key with default
	args = NewArgs(span)
	args.Push(Str("missing"), span)
	args.PushNamed("default", Int(0), span)
	result, err = CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("remove('missing', default: 0) failed: %v", err)
	}
	if v, ok := result.(IntValue); !ok || v != 0 {
		t.Errorf("Expected 0, got %v", result)
	}
}

func TestDictClear(t *testing.T) {
	d := NewDict()
	d.Set("a", Int(1))
	d.Set("b", Int(2))
	d.Set("c", Int(3))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})

	method := getDictMethod(&d, "clear", syntax.Detached())
	fn, _ := AsFunc(method)

	result, err := CallFunc(vm, fn, NewArgs(syntax.Span{}))
	if err != nil {
		t.Fatalf("clear() failed: %v", err)
	}
	if !IsNone(result) {
		t.Errorf("Expected none, got %v", result)
	}

	if d.Len() != 0 {
		t.Errorf("Dict should be empty, got %d entries", d.Len())
	}
}

func TestDictKeys(t *testing.T) {
	d := NewDict()
	d.Set("c", Int(3))
	d.Set("a", Int(1))
	d.Set("b", Int(2))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})

	method := getDictMethod(&d, "keys", syntax.Detached())
	fn, _ := AsFunc(method)

	result, err := CallFunc(vm, fn, NewArgs(syntax.Span{}))
	if err != nil {
		t.Fatalf("keys() failed: %v", err)
	}

	arr, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}
	if len(arr) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(arr))
	}

	// Verify insertion order is preserved
	expected := []string{"c", "a", "b"}
	for i, exp := range expected {
		if s, ok := arr[i].(StrValue); !ok || string(s) != exp {
			t.Errorf("Expected key[%d] = %q, got %v", i, exp, arr[i])
		}
	}
}

func TestDictValues(t *testing.T) {
	d := NewDict()
	d.Set("a", Int(1))
	d.Set("b", Int(2))
	d.Set("c", Int(3))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})

	method := getDictMethod(&d, "values", syntax.Detached())
	fn, _ := AsFunc(method)

	result, err := CallFunc(vm, fn, NewArgs(syntax.Span{}))
	if err != nil {
		t.Fatalf("values() failed: %v", err)
	}

	arr, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}
	if len(arr) != 3 {
		t.Errorf("Expected 3 values, got %d", len(arr))
	}

	// Verify insertion order is preserved
	expected := []int64{1, 2, 3}
	for i, exp := range expected {
		if v, ok := arr[i].(IntValue); !ok || int64(v) != exp {
			t.Errorf("Expected value[%d] = %d, got %v", i, exp, arr[i])
		}
	}
}

func TestDictPairs(t *testing.T) {
	d := NewDict()
	d.Set("x", Int(10))
	d.Set("y", Int(20))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})

	method := getDictMethod(&d, "pairs", syntax.Detached())
	fn, _ := AsFunc(method)

	result, err := CallFunc(vm, fn, NewArgs(syntax.Span{}))
	if err != nil {
		t.Fatalf("pairs() failed: %v", err)
	}

	arr, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}
	if len(arr) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(arr))
	}

	// Each pair should be an array of [key, value]
	pair0, ok := arr[0].(ArrayValue)
	if !ok || len(pair0) != 2 {
		t.Fatalf("Expected pair to be array of 2 elements, got %v", arr[0])
	}
	if s, ok := pair0[0].(StrValue); !ok || string(s) != "x" {
		t.Errorf("Expected pair[0][0] = 'x', got %v", pair0[0])
	}
	if v, ok := pair0[1].(IntValue); !ok || v != 10 {
		t.Errorf("Expected pair[0][1] = 10, got %v", pair0[1])
	}
}

func TestDictFilter(t *testing.T) {
	d := NewDict()
	d.Set("a", Int(1))
	d.Set("b", Int(2))
	d.Set("c", Int(3))
	d.Set("d", Int(4))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	span := syntax.Span{}

	// Create a filter function that keeps entries where value > 2
	filterFn := &Func{
		Name: strPtr("test"),
		Span: span,
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				args.Eat() // skip key
				valArg, _ := args.Expect("value")
				v, _ := AsInt(valArg.V)
				return Bool(v > 2), nil
			},
		},
	}

	method := getDictMethod(&d, "filter", syntax.Detached())
	fn, _ := AsFunc(method)

	args := NewArgs(span)
	args.Push(FuncValue{Func: filterFn}, span)
	result, err := CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("filter() failed: %v", err)
	}

	filtered, ok := AsDict(result)
	if !ok {
		t.Fatalf("Expected DictValue, got %T", result)
	}
	if filtered.Len() != 2 {
		t.Errorf("Expected 2 entries, got %d", filtered.Len())
	}

	// Verify correct entries remain
	if _, ok := filtered.Get("c"); !ok {
		t.Error("Key 'c' should exist in filtered dict")
	}
	if _, ok := filtered.Get("d"); !ok {
		t.Error("Key 'd' should exist in filtered dict")
	}
	if _, ok := filtered.Get("a"); ok {
		t.Error("Key 'a' should NOT exist in filtered dict")
	}
}

func TestDictMap(t *testing.T) {
	d := NewDict()
	d.Set("a", Int(1))
	d.Set("b", Int(2))
	vm := NewVm(nil, nil, NewScopes(nil), syntax.Span{})
	span := syntax.Span{}

	// Create a mapper function that doubles the value
	mapperFn := &Func{
		Name: strPtr("mapper"),
		Span: span,
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				args.Eat() // skip key
				valArg, _ := args.Expect("value")
				v, _ := AsInt(valArg.V)
				return Int(v * 2), nil
			},
		},
	}

	method := getDictMethod(&d, "map", syntax.Detached())
	fn, _ := AsFunc(method)

	args := NewArgs(span)
	args.Push(FuncValue{Func: mapperFn}, span)
	result, err := CallFunc(vm, fn, args)
	if err != nil {
		t.Fatalf("map() failed: %v", err)
	}

	mapped, ok := AsDict(result)
	if !ok {
		t.Fatalf("Expected DictValue, got %T", result)
	}
	if mapped.Len() != 2 {
		t.Errorf("Expected 2 entries, got %d", mapped.Len())
	}

	// Verify mapped values
	val, _ := mapped.Get("a")
	if v, ok := val.(IntValue); !ok || v != 2 {
		t.Errorf("Expected mapped['a'] = 2, got %v", val)
	}
	val, _ = mapped.Get("b")
	if v, ok := val.(IntValue); !ok || v != 4 {
		t.Errorf("Expected mapped['b'] = 4, got %v", val)
	}
}

func TestDictInsertionOrderPreserved(t *testing.T) {
	d := NewDict()
	keys := []string{"z", "m", "a", "q", "b"}
	for i, k := range keys {
		d.Set(k, Int(int64(i)))
	}

	// Verify keys() returns in insertion order
	gotKeys := d.Keys()
	if len(gotKeys) != len(keys) {
		t.Fatalf("Expected %d keys, got %d", len(keys), len(gotKeys))
	}
	for i, k := range keys {
		if gotKeys[i] != k {
			t.Errorf("Expected key[%d] = %q, got %q", i, k, gotKeys[i])
		}
	}

	// Verify updating a key doesn't change order
	d.Set("m", Int(100))
	gotKeys = d.Keys()
	for i, k := range keys {
		if gotKeys[i] != k {
			t.Errorf("After update: Expected key[%d] = %q, got %q", i, k, gotKeys[i])
		}
	}
}

func TestDictMethodUnknown(t *testing.T) {
	d := NewDict()
	method := getDictMethod(&d, "unknown_method", syntax.Detached())
	if method != nil {
		t.Error("Unknown method should return nil")
	}
}

func strPtr(s string) *string {
	return &s
}
