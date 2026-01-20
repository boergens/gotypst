package foundations

import (
	"math"
	"testing"
)

// --- Floor Tests ---

func TestFloor(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    Value
		wantErr bool
	}{
		{"int positive", Int(5), Int(5), false},
		{"int negative", Int(-5), Int(-5), false},
		{"int zero", Int(0), Int(0), false},
		{"float positive whole", Float(5.0), Int(5), false},
		{"float positive fractional", Float(5.7), Int(5), false},
		{"float negative fractional", Float(-5.7), Int(-6), false},
		{"float negative whole", Float(-5.0), Int(-5), false},
		{"float near zero positive", Float(0.3), Int(0), false},
		{"float near zero negative", Float(-0.3), Int(-1), false},
		{"invalid type string", Str("hello"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Floor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Floor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Floor() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Ceil Tests ---

func TestCeil(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    Value
		wantErr bool
	}{
		{"int positive", Int(5), Int(5), false},
		{"int negative", Int(-5), Int(-5), false},
		{"int zero", Int(0), Int(0), false},
		{"float positive whole", Float(5.0), Int(5), false},
		{"float positive fractional", Float(5.3), Int(6), false},
		{"float negative fractional", Float(-5.3), Int(-5), false},
		{"float negative whole", Float(-5.0), Int(-5), false},
		{"float near zero positive", Float(0.1), Int(1), false},
		{"float near zero negative", Float(-0.1), Int(0), false},
		{"invalid type string", Str("hello"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Ceil(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ceil() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Ceil() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Trunc Tests ---

func TestTrunc(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    Value
		wantErr bool
	}{
		{"int positive", Int(5), Int(5), false},
		{"int negative", Int(-5), Int(-5), false},
		{"int zero", Int(0), Int(0), false},
		{"float positive fractional", Float(5.7), Int(5), false},
		{"float negative fractional", Float(-5.7), Int(-5), false},
		{"float near zero positive", Float(0.9), Int(0), false},
		{"float near zero negative", Float(-0.9), Int(0), false},
		{"invalid type string", Str("hello"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Trunc(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Trunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("Trunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Fract Tests ---

func TestFract(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		want    Value
		wantErr bool
	}{
		{"int positive", Int(5), Float(0), false},
		{"int negative", Int(-5), Float(0), false},
		{"int zero", Int(0), Float(0), false},
		{"float positive fractional", Float(5.75), Float(0.75), false},
		{"float negative fractional", Float(-5.75), Float(-0.75), false},
		{"float whole positive", Float(5.0), Float(0), false},
		{"float whole negative", Float(-5.0), Float(0), false},
		{"invalid type string", Str("hello"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Fract(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotF, _ := toFloat64(got)
				wantF, _ := toFloat64(tt.want)
				if math.Abs(gotF-wantF) > 1e-10 {
					t.Errorf("Fract() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// --- Round Tests ---

func TestRound(t *testing.T) {
	tests := []struct {
		name    string
		input   Value
		digits  *Int
		want    Value
		wantErr bool
	}{
		{"int no digits", Int(5), nil, Int(5), false},
		{"float round down", Float(5.4), nil, Int(5), false},
		{"float round up", Float(5.5), nil, Int(6), false},
		{"float round down negative", Float(-5.4), nil, Int(-5), false},
		{"float round away from zero negative", Float(-5.5), nil, Int(-6), false},
		{"float with 2 decimal places", Float(5.456), intPtr(2), Float(5.46), false},
		{"float with 1 decimal place", Float(5.456), intPtr(1), Float(5.5), false},
		{"float with 0 decimal places", Float(5.456), intPtr(0), Float(5), false},
		{"float round integer digits", Float(1234.5), intPtr(-2), Float(1200), false},
		{"int round integer digits", Int(1234), intPtr(-2), Int(1200), false},
		{"invalid type string", Str("hello"), nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Round(tt.input, tt.digits)
			if (err != nil) != tt.wantErr {
				t.Errorf("Round() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotF, _ := toFloat64(got)
				wantF, _ := toFloat64(tt.want)
				if math.Abs(gotF-wantF) > 1e-10 {
					t.Errorf("Round() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func intPtr(i int64) *Int {
	v := Int(i)
	return &v
}

// --- Clamp Tests ---

func TestClamp(t *testing.T) {
	tests := []struct {
		name    string
		v       Value
		min     Value
		max     Value
		want    Value
		wantErr bool
	}{
		{"int in range", Int(5), Int(0), Int(10), Int(5), false},
		{"int below min", Int(-5), Int(0), Int(10), Int(0), false},
		{"int above max", Int(15), Int(0), Int(10), Int(10), false},
		{"int at min", Int(0), Int(0), Int(10), Int(0), false},
		{"int at max", Int(10), Int(0), Int(10), Int(10), false},
		{"float in range", Float(5.5), Float(0.0), Float(10.0), Float(5.5), false},
		{"float below min", Float(-5.5), Float(0.0), Float(10.0), Float(0.0), false},
		{"float above max", Float(15.5), Float(0.0), Float(10.0), Float(10.0), false},
		{"mixed int-float", Int(5), Float(0.0), Float(10.0), Float(5), false},
		{"min equals max", Int(5), Int(5), Int(5), Int(5), false},
		{"min greater than max", Int(5), Int(10), Int(0), nil, true},
		{"invalid type", Str("hello"), Int(0), Int(10), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Clamp(tt.v, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("Clamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotF, _ := toFloat64(got)
				wantF, _ := toFloat64(tt.want)
				if math.Abs(gotF-wantF) > 1e-10 {
					t.Errorf("Clamp() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// --- Min Tests ---

func TestMin(t *testing.T) {
	tests := []struct {
		name    string
		values  []Value
		want    Value
		wantErr bool
	}{
		{"single int", []Value{Int(5)}, Int(5), false},
		{"two ints", []Value{Int(5), Int(3)}, Int(3), false},
		{"three ints", []Value{Int(5), Int(3), Int(7)}, Int(3), false},
		{"with negatives", []Value{Int(5), Int(-3), Int(7)}, Int(-3), false},
		{"floats", []Value{Float(5.5), Float(3.3), Float(7.7)}, Float(3.3), false},
		{"mixed int-float", []Value{Int(5), Float(3.3), Int(7)}, Float(3.3), false},
		{"all same", []Value{Int(5), Int(5), Int(5)}, Int(5), false},
		{"empty", []Value{}, nil, true},
		{"invalid type", []Value{Str("hello"), Int(5)}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Min(tt.values...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Min() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotF, _ := toFloat64(got)
				wantF, _ := toFloat64(tt.want)
				if math.Abs(gotF-wantF) > 1e-10 {
					t.Errorf("Min() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// --- Max Tests ---

func TestMax(t *testing.T) {
	tests := []struct {
		name    string
		values  []Value
		want    Value
		wantErr bool
	}{
		{"single int", []Value{Int(5)}, Int(5), false},
		{"two ints", []Value{Int(5), Int(3)}, Int(5), false},
		{"three ints", []Value{Int(5), Int(3), Int(7)}, Int(7), false},
		{"with negatives", []Value{Int(-5), Int(-3), Int(-7)}, Int(-3), false},
		{"floats", []Value{Float(5.5), Float(3.3), Float(7.7)}, Float(7.7), false},
		{"mixed int-float", []Value{Int(5), Float(7.7), Int(3)}, Float(7.7), false},
		{"all same", []Value{Int(5), Int(5), Int(5)}, Int(5), false},
		{"empty", []Value{}, nil, true},
		{"invalid type", []Value{Str("hello"), Int(5)}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Max(tt.values...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Max() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotF, _ := toFloat64(got)
				wantF, _ := toFloat64(tt.want)
				if math.Abs(gotF-wantF) > 1e-10 {
					t.Errorf("Max() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// --- Edge Case Tests ---

func TestCalcEdgeCases(t *testing.T) {
	t.Run("Floor with infinity", func(t *testing.T) {
		got, err := Floor(Float(math.Inf(1)))
		if err != nil {
			t.Errorf("Floor(Inf) should not error: %v", err)
		}
		if got != nil {
			if v, ok := got.(Int); ok {
				// Infinity floored is MaxInt or MinInt depending on implementation
				_ = v
			}
		}
	})

	t.Run("Ceil with infinity", func(t *testing.T) {
		got, err := Ceil(Float(math.Inf(-1)))
		if err != nil {
			t.Errorf("Ceil(-Inf) should not error: %v", err)
		}
		_ = got
	})

	t.Run("Min with single float returns float", func(t *testing.T) {
		got, err := Min(Float(5.5))
		if err != nil {
			t.Errorf("Min(Float) error: %v", err)
		}
		if _, ok := got.(Float); !ok {
			t.Errorf("Min(Float) should return Float, got %T", got)
		}
	})

	t.Run("Max with single int returns int", func(t *testing.T) {
		got, err := Max(Int(5))
		if err != nil {
			t.Errorf("Max(Int) error: %v", err)
		}
		if _, ok := got.(Int); !ok {
			t.Errorf("Max(Int) should return Int, got %T", got)
		}
	})

	t.Run("Round with negative digits on int", func(t *testing.T) {
		digits := Int(-1)
		got, err := Round(Int(15), &digits)
		if err != nil {
			t.Errorf("Round error: %v", err)
		}
		if v, ok := got.(Int); !ok || v != Int(20) {
			t.Errorf("Round(15, -1) = %v, want 20", got)
		}
	})
}
