package syntax

import (
	"math"
	"testing"
)

func TestUnitString(t *testing.T) {
	tests := []struct {
		u    Unit
		want string
	}{
		{UnitNone, ""},
		{UnitPt, "pt"},
		{UnitMm, "mm"},
		{UnitCm, "cm"},
		{UnitIn, "in"},
		{UnitRad, "rad"},
		{UnitDeg, "deg"},
		{UnitEm, "em"},
		{UnitFr, "fr"},
		{UnitPercent, "%"},
	}

	for _, tt := range tests {
		if got := tt.u.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.u, got, tt.want)
		}
	}
}

func TestUnitName(t *testing.T) {
	tests := []struct {
		u    Unit
		want string
	}{
		{UnitNone, "none"},
		{UnitPt, "points"},
		{UnitMm, "millimeters"},
		{UnitCm, "centimeters"},
		{UnitIn, "inches"},
		{UnitRad, "radians"},
		{UnitDeg, "degrees"},
		{UnitEm, "em"},
		{UnitFr, "fraction"},
		{UnitPercent, "percent"},
	}

	for _, tt := range tests {
		if got := tt.u.Name(); got != tt.want {
			t.Errorf("%v.Name() = %q, want %q", tt.u, got, tt.want)
		}
	}
}

func TestUnitIsLength(t *testing.T) {
	lengthUnits := []Unit{UnitPt, UnitMm, UnitCm, UnitIn}
	for _, u := range lengthUnits {
		if !u.IsLength() {
			t.Errorf("%v should be a length unit", u)
		}
	}

	nonLengthUnits := []Unit{UnitNone, UnitRad, UnitDeg, UnitEm, UnitFr, UnitPercent}
	for _, u := range nonLengthUnits {
		if u.IsLength() {
			t.Errorf("%v should not be a length unit", u)
		}
	}
}

func TestUnitIsAngle(t *testing.T) {
	angleUnits := []Unit{UnitRad, UnitDeg}
	for _, u := range angleUnits {
		if !u.IsAngle() {
			t.Errorf("%v should be an angle unit", u)
		}
	}

	nonAngleUnits := []Unit{UnitNone, UnitPt, UnitMm, UnitEm, UnitPercent}
	for _, u := range nonAngleUnits {
		if u.IsAngle() {
			t.Errorf("%v should not be an angle unit", u)
		}
	}
}

func TestUnitIsRelative(t *testing.T) {
	relativeUnits := []Unit{UnitEm, UnitFr, UnitPercent}
	for _, u := range relativeUnits {
		if !u.IsRelative() {
			t.Errorf("%v should be a relative unit", u)
		}
	}

	nonRelativeUnits := []Unit{UnitNone, UnitPt, UnitMm, UnitRad}
	for _, u := range nonRelativeUnits {
		if u.IsRelative() {
			t.Errorf("%v should not be a relative unit", u)
		}
	}
}

func TestUnitFromString(t *testing.T) {
	tests := []struct {
		s    string
		want Unit
	}{
		{"pt", UnitPt},
		{"PT", UnitPt}, // case insensitive
		{"mm", UnitMm},
		{"cm", UnitCm},
		{"in", UnitIn},
		{"rad", UnitRad},
		{"deg", UnitDeg},
		{"em", UnitEm},
		{"fr", UnitFr},
		{"%", UnitPercent},
		{"", UnitNone},
		{"invalid", UnitNone},
	}

	for _, tt := range tests {
		if got := UnitFromString(tt.s); got != tt.want {
			t.Errorf("UnitFromString(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestUnitConvertToSameUnit(t *testing.T) {
	// Converting to same unit should return same value
	value := 10.0
	units := []Unit{UnitPt, UnitMm, UnitCm, UnitIn, UnitRad, UnitDeg}

	for _, u := range units {
		got, ok := u.ConvertTo(value, u)
		if !ok {
			t.Errorf("ConvertTo same unit %v should succeed", u)
		}
		if got != value {
			t.Errorf("ConvertTo same unit %v: got %f, want %f", u, got, value)
		}
	}
}

func TestUnitConvertToLength(t *testing.T) {
	// Test some known conversions
	tests := []struct {
		value  float64
		from   Unit
		to     Unit
		want   float64
		margin float64
	}{
		{72, UnitPt, UnitIn, 1, 0.001},     // 72pt = 1in
		{1, UnitIn, UnitPt, 72, 0.001},     // 1in = 72pt
		{10, UnitMm, UnitCm, 1, 0.001},     // 10mm = 1cm
		{1, UnitCm, UnitMm, 10, 0.001},     // 1cm = 10mm
		{25.4, UnitMm, UnitIn, 1, 0.01},    // 25.4mm ≈ 1in
		{1, UnitIn, UnitMm, 25.4, 0.01},    // 1in ≈ 25.4mm
		{2.54, UnitCm, UnitIn, 1, 0.001},   // 2.54cm = 1in
	}

	for _, tt := range tests {
		got, ok := tt.from.ConvertTo(tt.value, tt.to)
		if !ok {
			t.Errorf("ConvertTo(%f %v to %v) should succeed", tt.value, tt.from, tt.to)
			continue
		}
		if math.Abs(got-tt.want) > tt.margin {
			t.Errorf("ConvertTo(%f %v to %v) = %f, want %f (±%f)", tt.value, tt.from, tt.to, got, tt.want, tt.margin)
		}
	}
}

func TestUnitConvertToAngle(t *testing.T) {
	// Test angle conversions
	tests := []struct {
		value  float64
		from   Unit
		to     Unit
		want   float64
		margin float64
	}{
		{180, UnitDeg, UnitRad, math.Pi, 0.0001},    // 180° = π rad
		{math.Pi, UnitRad, UnitDeg, 180, 0.0001},    // π rad = 180°
		{90, UnitDeg, UnitRad, math.Pi / 2, 0.0001}, // 90° = π/2 rad
		{360, UnitDeg, UnitRad, 2 * math.Pi, 0.001}, // 360° = 2π rad
	}

	for _, tt := range tests {
		got, ok := tt.from.ConvertTo(tt.value, tt.to)
		if !ok {
			t.Errorf("ConvertTo(%f %v to %v) should succeed", tt.value, tt.from, tt.to)
			continue
		}
		if math.Abs(got-tt.want) > tt.margin {
			t.Errorf("ConvertTo(%f %v to %v) = %f, want %f (±%f)", tt.value, tt.from, tt.to, got, tt.want, tt.margin)
		}
	}
}

func TestUnitConvertToIncompatible(t *testing.T) {
	// Converting between incompatible units should fail
	tests := []struct {
		from Unit
		to   Unit
	}{
		{UnitPt, UnitRad},  // length to angle
		{UnitDeg, UnitMm},  // angle to length
		{UnitEm, UnitPt},   // relative to length
		{UnitFr, UnitDeg},  // fraction to angle
		{UnitNone, UnitPt}, // none to anything
	}

	for _, tt := range tests {
		_, ok := tt.from.ConvertTo(10, tt.to)
		if ok {
			t.Errorf("ConvertTo(%v to %v) should fail for incompatible units", tt.from, tt.to)
		}
	}
}
