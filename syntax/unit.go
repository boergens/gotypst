package syntax

import "strings"

// Unit represents a unit for numeric values.
type Unit int

const (
	// UnitNone represents no unit.
	UnitNone Unit = iota
	// Length units
	UnitPt  // pt (points)
	UnitMm  // mm (millimeters)
	UnitCm  // cm (centimeters)
	UnitIn  // in (inches)
	// Angle units
	UnitRad // rad (radians)
	UnitDeg // deg (degrees)
	// Relative units
	UnitEm // em (relative to font size)
	UnitFr // fr (flex fraction)
	// Ratio
	UnitPercent // % (percent)
)

// String returns the string representation of the unit.
func (u Unit) String() string {
	switch u {
	case UnitNone:
		return ""
	case UnitPt:
		return "pt"
	case UnitMm:
		return "mm"
	case UnitCm:
		return "cm"
	case UnitIn:
		return "in"
	case UnitRad:
		return "rad"
	case UnitDeg:
		return "deg"
	case UnitEm:
		return "em"
	case UnitFr:
		return "fr"
	case UnitPercent:
		return "%"
	default:
		return "unknown"
	}
}

// Name returns a human-readable name for the unit.
func (u Unit) Name() string {
	switch u {
	case UnitNone:
		return "none"
	case UnitPt:
		return "points"
	case UnitMm:
		return "millimeters"
	case UnitCm:
		return "centimeters"
	case UnitIn:
		return "inches"
	case UnitRad:
		return "radians"
	case UnitDeg:
		return "degrees"
	case UnitEm:
		return "em"
	case UnitFr:
		return "fraction"
	case UnitPercent:
		return "percent"
	default:
		return "unknown"
	}
}

// IsLength returns true if this is a length unit.
func (u Unit) IsLength() bool {
	switch u {
	case UnitPt, UnitMm, UnitCm, UnitIn:
		return true
	}
	return false
}

// IsAngle returns true if this is an angle unit.
func (u Unit) IsAngle() bool {
	switch u {
	case UnitRad, UnitDeg:
		return true
	}
	return false
}

// IsRelative returns true if this is a relative unit.
func (u Unit) IsRelative() bool {
	switch u {
	case UnitEm, UnitFr, UnitPercent:
		return true
	}
	return false
}

// UnitFromString parses a unit from its string representation.
func UnitFromString(s string) Unit {
	switch strings.ToLower(s) {
	case "pt":
		return UnitPt
	case "mm":
		return UnitMm
	case "cm":
		return UnitCm
	case "in":
		return UnitIn
	case "rad":
		return UnitRad
	case "deg":
		return UnitDeg
	case "em":
		return UnitEm
	case "fr":
		return UnitFr
	case "%":
		return UnitPercent
	default:
		return UnitNone
	}
}

// ConvertTo converts a value from this unit to the target unit.
// Returns the converted value and true if conversion is possible,
// or 0 and false if conversion between these units is not supported.
func (u Unit) ConvertTo(value float64, target Unit) (float64, bool) {
	if u == target {
		return value, true
	}

	// Length conversions (base unit: pt)
	if u.IsLength() && target.IsLength() {
		// Convert to points first
		var pts float64
		switch u {
		case UnitPt:
			pts = value
		case UnitMm:
			pts = value * 2.83465 // 1mm = 2.83465pt
		case UnitCm:
			pts = value * 28.3465 // 1cm = 28.3465pt
		case UnitIn:
			pts = value * 72 // 1in = 72pt
		}

		// Convert from points to target
		switch target {
		case UnitPt:
			return pts, true
		case UnitMm:
			return pts / 2.83465, true
		case UnitCm:
			return pts / 28.3465, true
		case UnitIn:
			return pts / 72, true
		}
	}

	// Angle conversions (base unit: rad)
	if u.IsAngle() && target.IsAngle() {
		var rads float64
		switch u {
		case UnitRad:
			rads = value
		case UnitDeg:
			rads = value * 3.14159265358979323846 / 180 // deg to rad
		}

		switch target {
		case UnitRad:
			return rads, true
		case UnitDeg:
			return rads * 180 / 3.14159265358979323846, true
		}
	}

	return 0, false
}
