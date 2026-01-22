package foundations

import (
	"math"
	"testing"
)

// --- Abs Tests ---

func TestAbs(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    Value
		wantErr bool
	}{
		{"int positive", Int(5), Int(5), false},
		{"int negative", Int(-5), Int(5), false},
		{"int zero", Int(0), Int(0), false},
		{"float positive", Float(3.14), Float(3.14), false},
		{"float negative", Float(-3.14), Float(3.14), false},
		{"float zero", Float(0), Float(0), false},
		{"int min overflow", Int(math.MinInt64), nil, true},
		{"string error", Str("hello"), nil, true},
		{"bool error", Bool(true), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Abs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Abs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !Equal(got, tt.want) {
				t.Errorf("Abs() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Pow Tests ---

func TestPow(t *testing.T) {
	tests := []struct {
		name     string
		base     Value
		exponent Value
		want     Value
		wantErr  bool
	}{
		{"int 2^3", Int(2), Int(3), Int(8), false},
		{"int 2^0", Int(2), Int(0), Int(1), false},
		{"int 0^5", Int(0), Int(5), Int(0), false},
		{"int 1^100", Int(1), Int(100), Int(1), false},
		{"int (-1)^2", Int(-1), Int(2), Int(1), false},
		{"int (-1)^3", Int(-1), Int(3), Int(-1), false},
		{"int 10^18", Int(10), Int(18), Int(1000000000000000000), false},
		{"float 2.0^3.0", Float(2.0), Float(3.0), Float(8.0), false},
		{"float 4.0^0.5", Float(4.0), Float(0.5), Float(2.0), false},
		{"mixed int^float", Int(4), Float(0.5), Float(2.0), false},
		{"mixed float^int", Float(2.0), Int(3), Float(8.0), false},
		{"string base error", Str("2"), Int(3), nil, true},
		{"string exp error", Int(2), Str("3"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Pow(tt.base, tt.exponent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !valueApproxEqual(got, tt.want) {
					t.Errorf("Pow() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// --- Exp Tests ---

func TestExp(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    float64
		wantErr bool
	}{
		{"exp(0)", Int(0), 1.0, false},
		{"exp(1)", Int(1), math.E, false},
		{"exp(-1)", Int(-1), 1 / math.E, false},
		{"exp(2)", Int(2), math.E * math.E, false},
		{"exp(0.0)", Float(0.0), 1.0, false},
		{"exp(1.0)", Float(1.0), math.E, false},
		{"string error", Str("1"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Exp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotF, ok := got.(Float)
				if !ok {
					t.Errorf("Exp() returned %T, want Float", got)
					return
				}
				if !approxEqual(float64(gotF), tt.want) {
					t.Errorf("Exp() = %v, want %v", gotF, tt.want)
				}
			}
		})
	}
}

// --- Sqrt Tests ---

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    float64
		wantErr bool
	}{
		{"sqrt(4)", Int(4), 2.0, false},
		{"sqrt(0)", Int(0), 0.0, false},
		{"sqrt(1)", Int(1), 1.0, false},
		{"sqrt(2)", Int(2), math.Sqrt(2), false},
		{"sqrt(4.0)", Float(4.0), 2.0, false},
		{"sqrt(0.25)", Float(0.25), 0.5, false},
		{"sqrt(-1) error", Int(-1), 0, true},
		{"sqrt(-4.0) error", Float(-4.0), 0, true},
		{"string error", Str("4"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sqrt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sqrt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotF, ok := got.(Float)
				if !ok {
					t.Errorf("Sqrt() returned %T, want Float", got)
					return
				}
				if !approxEqual(float64(gotF), tt.want) {
					t.Errorf("Sqrt() = %v, want %v", gotF, tt.want)
				}
			}
		})
	}
}

// --- Root Tests ---

func TestRoot(t *testing.T) {
	tests := []struct {
		name     string
		radicand Value
		index    Value
		want     float64
		wantErr  bool
	}{
		{"root(8, 3)", Int(8), Int(3), 2.0, false},
		{"root(16, 4)", Int(16), Int(4), 2.0, false},
		{"root(27, 3)", Int(27), Int(3), 3.0, false},
		{"root(4, 2)", Int(4), Int(2), 2.0, false},
		{"root(8.0, 3.0)", Float(8.0), Float(3.0), 2.0, false},
		{"root(-8, 3) odd root", Int(-8), Int(3), -2.0, false},
		{"root(-27, 3) odd root", Int(-27), Int(3), -3.0, false},
		{"root(-4, 2) even error", Int(-4), Int(2), 0, true},
		{"root(8, 0) zero error", Int(8), Int(0), 0, true},
		{"string radicand error", Str("8"), Int(3), 0, true},
		{"string index error", Int(8), Str("3"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Root(tt.radicand, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("Root() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotF, ok := got.(Float)
				if !ok {
					t.Errorf("Root() returned %T, want Float", got)
					return
				}
				if !approxEqual(float64(gotF), tt.want) {
					t.Errorf("Root() = %v, want %v", gotF, tt.want)
				}
			}
		})
	}
}

// --- Helper Functions ---

// approxEqual checks if two floats are approximately Equal.
func approxEqual(a, b float64) bool {
	const epsilon = 1e-10
	if a == b {
		return true
	}
	diff := math.Abs(a - b)
	if a == 0 || b == 0 {
		return diff < epsilon
	}
	return diff/math.Max(math.Abs(a), math.Abs(b)) < epsilon
}

// valueApproxEqual checks if two Values are approximately Equal.
// For floats, uses approximate comparison; for ints, exact comparison.
func valueApproxEqual(a, b Value) bool {
	switch av := a.(type) {
	case Int:
		switch bv := b.(type) {
		case Int:
			return av == bv
		case Float:
			return approxEqual(float64(av), float64(bv))
		}
	case Float:
		switch bv := b.(type) {
		case Int:
			return approxEqual(float64(av), float64(bv))
		case Float:
			return approxEqual(float64(av), float64(bv))
		}
	}
	return Equal(a, b)
}
