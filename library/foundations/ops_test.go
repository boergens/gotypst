package foundations

import (
	"math"
	"testing"
)

// --- Unary Operator Tests ---

func TestPos(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    Value
		wantErr bool
	}{
		{"int positive", Int(5), Int(5), false},
		{"int negative", Int(-5), Int(-5), false},
		{"int zero", Int(0), Int(0), false},
		{"float positive", Float(3.14), Float(3.14), false},
		{"float negative", Float(-3.14), Float(-3.14), false},
		{"string error", Str("hello"), nil, true},
		{"bool error", Bool(true), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Pos(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Pos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    Value
		wantErr bool
	}{
		{"int positive", Int(5), Int(-5), false},
		{"int negative", Int(-5), Int(5), false},
		{"int zero", Int(0), Int(0), false},
		{"float positive", Float(3.14), Float(-3.14), false},
		{"float negative", Float(-3.14), Float(3.14), false},
		{"int min overflow", Int(math.MinInt64), nil, true},
		{"string error", Str("hello"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Neg(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Neg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Neg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNot(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    Value
		wantErr bool
	}{
		{"true", Bool(true), Bool(false), false},
		{"false", Bool(false), Bool(true), false},
		{"int error", Int(0), nil, true},
		{"string error", Str(""), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Not(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Not() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Not() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Binary Arithmetic Tests ---

func TestAdd(t *testing.T) {
	tests := []struct {
		name    string
		lhs     Value
		rhs     Value
		want    Value
		wantErr bool
	}{
		{"int + int", Int(2), Int(3), Int(5), false},
		{"int + float", Int(2), Float(3.5), Float(5.5), false},
		{"float + int", Float(2.5), Int(3), Float(5.5), false},
		{"float + float", Float(2.5), Float(3.5), Float(6.0), false},
		{"string + string", Str("hello"), Str(" world"), Str("hello world"), false},
		{"array + array", NewArray(Int(1)), NewArray(Int(2)), NewArray(Int(1), Int(2)), false},
		{"int overflow", Int(math.MaxInt64), Int(1), nil, true},
		{"type mismatch", Int(1), Str("a"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Add(tt.lhs, tt.rhs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		name    string
		lhs     Value
		rhs     Value
		want    Value
		wantErr bool
	}{
		{"int - int", Int(5), Int(3), Int(2), false},
		{"int - float", Int(5), Float(2.5), Float(2.5), false},
		{"float - int", Float(5.5), Int(3), Float(2.5), false},
		{"float - float", Float(5.5), Float(2.5), Float(3.0), false},
		{"int underflow", Int(math.MinInt64), Int(1), nil, true},
		{"type mismatch", Int(1), Str("a"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sub(tt.lhs, tt.rhs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMul(t *testing.T) {
	tests := []struct {
		name    string
		lhs     Value
		rhs     Value
		want    Value
		wantErr bool
	}{
		{"int * int", Int(3), Int(4), Int(12), false},
		{"int * float", Int(3), Float(2.5), Float(7.5), false},
		{"float * int", Float(2.5), Int(4), Float(10.0), false},
		{"float * float", Float(2.5), Float(4.0), Float(10.0), false},
		{"string * int", Str("ab"), Int(3), Str("ababab"), false},
		{"int * string", Int(2), Str("xy"), Str("xyxy"), false},
		{"array * int", NewArray(Int(1), Int(2)), Int(2), NewArray(Int(1), Int(2), Int(1), Int(2)), false},
		{"string * 0", Str("hello"), Int(0), Str(""), false},
		{"string * negative", Str("a"), Int(-1), nil, true},
		{"int overflow", Int(math.MaxInt64), Int(2), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mul(tt.lhs, tt.rhs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mul() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Mul() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiv(t *testing.T) {
	tests := []struct {
		name    string
		lhs     Value
		rhs     Value
		want    Value
		wantErr bool
	}{
		{"int / int", Int(10), Int(4), Float(2.5), false},
		{"int / float", Int(10), Float(4.0), Float(2.5), false},
		{"float / int", Float(10.0), Int(4), Float(2.5), false},
		{"float / float", Float(10.0), Float(4.0), Float(2.5), false},
		{"div by zero int", Int(10), Int(0), nil, true},
		{"div by zero float", Float(10.0), Float(0.0), nil, true},
		{"type mismatch", Int(10), Str("2"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Div(tt.lhs, tt.rhs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Div() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Div() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Logical Operator Tests ---

func TestAnd(t *testing.T) {
	tests := []struct {
		name    string
		lhs     Value
		rhs     Value
		want    Value
		wantErr bool
	}{
		{"true and true", Bool(true), Bool(true), Bool(true), false},
		{"true and false", Bool(true), Bool(false), Bool(false), false},
		{"false and true", Bool(false), Bool(true), Bool(false), false},
		{"false and false", Bool(false), Bool(false), Bool(false), false},
		{"int and bool", Int(1), Bool(true), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := And(tt.lhs, tt.rhs)
			if (err != nil) != tt.wantErr {
				t.Errorf("And() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("And() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOr(t *testing.T) {
	tests := []struct {
		name    string
		lhs     Value
		rhs     Value
		want    Value
		wantErr bool
	}{
		{"true or true", Bool(true), Bool(true), Bool(true), false},
		{"true or false", Bool(true), Bool(false), Bool(true), false},
		{"false or true", Bool(false), Bool(true), Bool(true), false},
		{"false or false", Bool(false), Bool(false), Bool(false), false},
		{"int or bool", Int(1), Bool(true), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Or(tt.lhs, tt.rhs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Or() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Or() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Comparison Operator Tests ---

func TestEq(t *testing.T) {
	tests := []struct {
		name string
		lhs  Value
		rhs  Value
		want Value
	}{
		{"int == int true", Int(5), Int(5), Bool(true)},
		{"int == int false", Int(5), Int(6), Bool(false)},
		{"int == float true", Int(5), Float(5.0), Bool(true)},
		{"int == float false", Int(5), Float(5.5), Bool(false)},
		{"float == float true", Float(3.14), Float(3.14), Bool(true)},
		{"string == string true", Str("hello"), Str("hello"), Bool(true)},
		{"string == string false", Str("hello"), Str("world"), Bool(false)},
		{"bool == bool true", Bool(true), Bool(true), Bool(true)},
		{"bool == bool false", Bool(true), Bool(false), Bool(false)},
		{"none == none", None, None, Bool(true)},
		{"auto == auto", Auto, Auto, Bool(true)},
		{"int == string", Int(1), Str("1"), Bool(false)},
		{"array == array true", NewArray(Int(1), Int(2)), NewArray(Int(1), Int(2)), Bool(true)},
		{"array == array false", NewArray(Int(1)), NewArray(Int(2)), Bool(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Eq(tt.lhs, tt.rhs)
			if err != nil {
				t.Errorf("Eq() unexpected error = %v", err)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("Eq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNeq(t *testing.T) {
	tests := []struct {
		name string
		lhs  Value
		rhs  Value
		want Value
	}{
		{"int != int false", Int(5), Int(5), Bool(false)},
		{"int != int true", Int(5), Int(6), Bool(true)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Neq(tt.lhs, tt.rhs)
			if err != nil {
				t.Errorf("Neq() unexpected error = %v", err)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("Neq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLt(t *testing.T) {
	tests := []struct {
		name    string
		lhs     Value
		rhs     Value
		want    Value
		wantErr bool
	}{
		{"int < int true", Int(3), Int(5), Bool(true), false},
		{"int < int false", Int(5), Int(3), Bool(false), false},
		{"int < int equal", Int(5), Int(5), Bool(false), false},
		{"int < float", Int(3), Float(3.5), Bool(true), false},
		{"float < int", Float(2.5), Int(3), Bool(true), false},
		{"string < string", Str("abc"), Str("abd"), Bool(true), false},
		{"incomparable", Int(1), Str("1"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Lt(tt.lhs, tt.rhs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Lt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLeq(t *testing.T) {
	tests := []struct {
		name string
		lhs  Value
		rhs  Value
		want Value
	}{
		{"int <= int true less", Int(3), Int(5), Bool(true)},
		{"int <= int true equal", Int(5), Int(5), Bool(true)},
		{"int <= int false", Int(6), Int(5), Bool(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Leq(tt.lhs, tt.rhs)
			if err != nil {
				t.Errorf("Leq() unexpected error = %v", err)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("Leq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGt(t *testing.T) {
	tests := []struct {
		name string
		lhs  Value
		rhs  Value
		want Value
	}{
		{"int > int true", Int(5), Int(3), Bool(true)},
		{"int > int false", Int(3), Int(5), Bool(false)},
		{"int > int equal", Int(5), Int(5), Bool(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Gt(tt.lhs, tt.rhs)
			if err != nil {
				t.Errorf("Gt() unexpected error = %v", err)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("Gt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeq(t *testing.T) {
	tests := []struct {
		name string
		lhs  Value
		rhs  Value
		want Value
	}{
		{"int >= int true greater", Int(5), Int(3), Bool(true)},
		{"int >= int true equal", Int(5), Int(5), Bool(true)},
		{"int >= int false", Int(3), Int(5), Bool(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Geq(tt.lhs, tt.rhs)
			if err != nil {
				t.Errorf("Geq() unexpected error = %v", err)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("Geq() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Membership Operator Tests ---

func TestIn(t *testing.T) {
	tests := []struct {
		name    string
		element Value
		container Value
		want    Value
		wantErr bool
	}{
		{"string in string true", Str("ell"), Str("hello"), Bool(true), false},
		{"string in string false", Str("xyz"), Str("hello"), Bool(false), false},
		{"int in array true", Int(2), NewArray(Int(1), Int(2), Int(3)), Bool(true), false},
		{"int in array false", Int(5), NewArray(Int(1), Int(2), Int(3)), Bool(false), false},
		{"string in dict true", Str("key"), dictWithKeys("key", "other"), Bool(true), false},
		{"string in dict false", Str("missing"), dictWithKeys("key"), Bool(false), false},
		{"int in string error", Int(1), Str("hello"), nil, true},
		{"string in int error", Str("a"), Int(1), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := In(tt.element, tt.container)
			if (err != nil) != tt.wantErr {
				t.Errorf("In() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotIn(t *testing.T) {
	tests := []struct {
		name      string
		element   Value
		container Value
		want      Value
	}{
		{"string not in string true", Str("xyz"), Str("hello"), Bool(true)},
		{"string not in string false", Str("ell"), Str("hello"), Bool(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NotIn(tt.element, tt.container)
			if err != nil {
				t.Errorf("NotIn() unexpected error = %v", err)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("NotIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Join Operation Tests ---

func TestJoin(t *testing.T) {
	tests := []struct {
		name string
		lhs  Value
		rhs  Value
		want Value
	}{
		{"none + value", None, Int(5), Int(5)},
		{"value + none", Int(5), None, Int(5)},
		{"none + none", None, None, None},
		{"string + string", Str("hello"), Str(" world"), Str("hello world")},
		{"array + array", NewArray(Int(1)), NewArray(Int(2)), NewArray(Int(1), Int(2))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Join(tt.lhs, tt.rhs)
			if err != nil {
				t.Errorf("Join() unexpected error = %v", err)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("Join() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Helper Tests ---

func TestIsZero(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  bool
	}{
		{"int zero", Int(0), true},
		{"int nonzero", Int(5), false},
		{"float zero", Float(0), true},
		{"float nonzero", Float(0.1), false},
		{"string", Str(""), false},
		{"none", None, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsZero(tt.input); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Dict Merge Tests ---

func TestDictMerge(t *testing.T) {
	d1 := NewDict()
	d1.Set("a", Int(1))
	d1.Set("b", Int(2))

	d2 := NewDict()
	d2.Set("b", Int(3))
	d2.Set("c", Int(4))

	result, err := Add(d1, d2)
	if err != nil {
		t.Fatalf("Add() dict merge failed: %v", err)
	}

	merged := result.(*Dict)
	if v, _ := merged.Get("a"); !equal(v, Int(1)) {
		t.Errorf("expected a=1, got %v", v)
	}
	if v, _ := merged.Get("b"); !equal(v, Int(3)) {
		t.Errorf("expected b=3 (overwritten), got %v", v)
	}
	if v, _ := merged.Get("c"); !equal(v, Int(4)) {
		t.Errorf("expected c=4, got %v", v)
	}
}

// --- Helper function ---

func dictWithKeys(keys ...string) *Dict {
	d := NewDict()
	for _, k := range keys {
		d.Set(k, None)
	}
	return d
}
