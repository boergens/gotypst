package calc

import (
	"math"
	"testing"

	"github.com/boergens/gotypst/library/foundations"
)

// floatClose checks if two floats are close within epsilon.
func floatClose(a, b, epsilon float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsInf(a, 1) && math.IsInf(b, 1) {
		return true
	}
	if math.IsInf(a, -1) && math.IsInf(b, -1) {
		return true
	}
	return math.Abs(a-b) < epsilon
}

// assertFloatResult checks that the result is a Float close to expected.
func assertFloatResult(t *testing.T, name string, got foundations.Value, err error, expected float64, wantErr bool) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("%s() error = %v, wantErr %v", name, err, wantErr)
		return
	}
	if wantErr {
		return
	}
	f, ok := got.(foundations.Float)
	if !ok {
		t.Errorf("%s() returned %T, want Float", name, got)
		return
	}
	if !floatClose(float64(f), expected, 1e-10) {
		t.Errorf("%s() = %v, want %v", name, float64(f), expected)
	}
}

func TestSin(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), 0, false},
		{"pi/2", foundations.Float(math.Pi / 2), 1, false},
		{"pi", foundations.Float(math.Pi), 0, false},
		{"negative", foundations.Float(-math.Pi / 2), -1, false},
		{"int input", foundations.Int(0), 0, false},
		{"string error", foundations.Str("hello"), 0, true},
		{"bool error", foundations.Bool(true), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sin(tt.input)
			assertFloatResult(t, "Sin", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestCos(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), 1, false},
		{"pi/2", foundations.Float(math.Pi / 2), 0, false},
		{"pi", foundations.Float(math.Pi), -1, false},
		{"int input", foundations.Int(0), 1, false},
		{"string error", foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Cos(tt.input)
			assertFloatResult(t, "Cos", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestTan(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), 0, false},
		{"pi/4", foundations.Float(math.Pi / 4), 1, false},
		{"negative pi/4", foundations.Float(-math.Pi / 4), -1, false},
		{"int input", foundations.Int(0), 0, false},
		{"string error", foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Tan(tt.input)
			assertFloatResult(t, "Tan", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestAsin(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), 0, false},
		{"one", foundations.Float(1), math.Pi / 2, false},
		{"negative one", foundations.Float(-1), -math.Pi / 2, false},
		{"half", foundations.Float(0.5), math.Asin(0.5), false},
		{"out of range returns nan", foundations.Float(2), math.NaN(), false},
		{"int input", foundations.Int(0), 0, false},
		{"int one", foundations.Int(1), math.Pi / 2, false},
		{"string error", foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Asin(tt.input)
			assertFloatResult(t, "Asin", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestAcos(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), math.Pi / 2, false},
		{"one", foundations.Float(1), 0, false},
		{"negative one", foundations.Float(-1), math.Pi, false},
		{"half", foundations.Float(0.5), math.Acos(0.5), false},
		{"out of range returns nan", foundations.Float(2), math.NaN(), false},
		{"int input", foundations.Int(1), 0, false},
		{"string error", foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Acos(tt.input)
			assertFloatResult(t, "Acos", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestAtan(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), 0, false},
		{"one", foundations.Float(1), math.Pi / 4, false},
		{"negative one", foundations.Float(-1), -math.Pi / 4, false},
		{"large positive", foundations.Float(1000), math.Atan(1000), false},
		{"int input", foundations.Int(0), 0, false},
		{"string error", foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Atan(tt.input)
			assertFloatResult(t, "Atan", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestAtan2(t *testing.T) {
	tests := []struct {
		name    string
		y       foundations.Value
		x       foundations.Value
		want    float64
		wantErr bool
	}{
		{"first quadrant", foundations.Float(1), foundations.Float(1), math.Pi / 4, false},
		{"second quadrant", foundations.Float(1), foundations.Float(-1), 3 * math.Pi / 4, false},
		{"third quadrant", foundations.Float(-1), foundations.Float(-1), -3 * math.Pi / 4, false},
		{"fourth quadrant", foundations.Float(-1), foundations.Float(1), -math.Pi / 4, false},
		{"positive y axis", foundations.Float(1), foundations.Float(0), math.Pi / 2, false},
		{"negative y axis", foundations.Float(-1), foundations.Float(0), -math.Pi / 2, false},
		{"positive x axis", foundations.Float(0), foundations.Float(1), 0, false},
		{"negative x axis", foundations.Float(0), foundations.Float(-1), math.Pi, false},
		{"int inputs", foundations.Int(1), foundations.Int(1), math.Pi / 4, false},
		{"mixed int float", foundations.Int(1), foundations.Float(1), math.Pi / 4, false},
		{"y string error", foundations.Str("hello"), foundations.Float(1), 0, true},
		{"x string error", foundations.Float(1), foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Atan2(tt.y, tt.x)
			assertFloatResult(t, "Atan2", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestSinh(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), 0, false},
		{"one", foundations.Float(1), math.Sinh(1), false},
		{"negative", foundations.Float(-1), math.Sinh(-1), false},
		{"large", foundations.Float(10), math.Sinh(10), false},
		{"int input", foundations.Int(0), 0, false},
		{"string error", foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sinh(tt.input)
			assertFloatResult(t, "Sinh", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestCosh(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), 1, false},
		{"one", foundations.Float(1), math.Cosh(1), false},
		{"negative", foundations.Float(-1), math.Cosh(-1), false},
		{"large", foundations.Float(10), math.Cosh(10), false},
		{"int input", foundations.Int(0), 1, false},
		{"string error", foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Cosh(tt.input)
			assertFloatResult(t, "Cosh", got, err, tt.want, tt.wantErr)
		})
	}
}

func TestTanh(t *testing.T) {
	tests := []struct {
		name    string
		input   foundations.Value
		want    float64
		wantErr bool
	}{
		{"zero", foundations.Float(0), 0, false},
		{"one", foundations.Float(1), math.Tanh(1), false},
		{"negative", foundations.Float(-1), math.Tanh(-1), false},
		{"large approaches 1", foundations.Float(100), 1, false},
		{"large negative approaches -1", foundations.Float(-100), -1, false},
		{"int input", foundations.Int(0), 0, false},
		{"string error", foundations.Str("hello"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Tanh(tt.input)
			assertFloatResult(t, "Tanh", got, err, tt.want, tt.wantErr)
		})
	}
}
