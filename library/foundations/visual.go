// Visual value types for Typst.
// Gradient, Tiling, Symbol types.
// Note: Color types are in color.go

package foundations

// GradientValue represents a gradient.
type GradientValue struct {
	// Stops contains the color stops.
	Stops []GradientStop
}

// GradientStop represents a single stop in a gradient.
type GradientStop struct {
	// Color is the color at this stop (implements Color interface from color.go).
	Color  Value
	Offset float64
}

func (GradientValue) Type() Type         { return TypeGradient }
func (v GradientValue) Display() Content { return Content{} }
func (v GradientValue) Clone() Value {
	if v.Stops == nil {
		return GradientValue{}
	}
	stops := make([]GradientStop, len(v.Stops))
	copy(stops, v.Stops)
	return GradientValue{Stops: stops}
}
func (GradientValue) isValue() {}

// TilingValue represents a tiling pattern.
type TilingValue struct {
	// Content is the pattern content.
	Content Content
}

func (TilingValue) Type() Type         { return TypeTiling }
func (v TilingValue) Display() Content { return Content{} }
func (v TilingValue) Clone() Value     { return v }
func (TilingValue) isValue()           {}

// SymbolValue represents a symbol character.
type SymbolValue struct {
	// Char is the symbol character.
	Char rune
}

func (SymbolValue) Type() Type         { return TypeSymbol }
func (v SymbolValue) Display() Content { return Content{} }
func (v SymbolValue) Clone() Value     { return v }
func (SymbolValue) isValue()           {}

// DynValue represents a dynamically-typed value.
type DynValue struct {
	// Inner is the underlying dynamic value.
	Inner interface{}
	// TypeName is the name of the dynamic type.
	TypeName string
}

func (DynValue) Type() Type         { return TypeDyn }
func (v DynValue) Display() Content { return Content{} }
func (v DynValue) Clone() Value     { return v }
func (DynValue) isValue()           {}
