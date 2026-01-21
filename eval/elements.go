package eval

// ----------------------------------------------------------------------------
// Element Function Registration
// ----------------------------------------------------------------------------
// This file contains the registration functions that hook up all element
// functions to the standard library scope.
//
// Element implementations are organized by Typst module:
// - elem_model.go:     par, parbreak, heading, list, enum, link, table
// - elem_layout.go:    stack, align, columns, box, block, pad
// - elem_visualize.go: image
// - text_funcs.go:     text, strong, emph, raw

// RegisterElementFunctions registers all element functions in the given scope.
// Call this when setting up the standard library scope.
func RegisterElementFunctions(scope *Scope) {
	// Text module functions (from text_funcs.go)
	scope.DefineFunc("text", TextFunc())
	scope.DefineFunc("strong", StrongFunc())
	scope.DefineFunc("emph", EmphFunc())
	scope.DefineFunc("raw", RawFunc())

	// Model module functions (from elem_model.go)
	scope.DefineFunc("par", ParFunc())
	scope.DefineFunc("parbreak", ParbreakFunc())
	scope.DefineFunc("heading", HeadingFunc())
	scope.DefineFunc("list", ListFunc())
	scope.DefineFunc("enum", EnumFunc())
	scope.DefineFunc("link", LinkFunc())
	scope.DefineFunc("table", TableFunc())

	// Layout module functions (from elem_layout.go)
	scope.DefineFunc("stack", StackFunc())
	scope.DefineFunc("align", AlignFunc())
	scope.DefineFunc("columns", ColumnsFunc())
	scope.DefineFunc("box", BoxFunc())
	scope.DefineFunc("block", BlockFunc())
	scope.DefineFunc("pad", PadFunc())

	// Visualize module functions (from elem_visualize.go)
	scope.DefineFunc("image", ImageFunc())
	scope.DefineFunc("rect", RectFunc())
	scope.DefineFunc("circle", CircleFunc())
	scope.DefineFunc("ellipse", EllipseFunc())
	scope.DefineFunc("line", LineFunc())
	scope.DefineFunc("path", PathFunc())
	scope.DefineFunc("polygon", PolygonFunc())
	scope.DefineFunc("square", SquareFunc())
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		// Text module
		"text":   TextFunc(),
		"strong": StrongFunc(),
		"emph":   EmphFunc(),
		"raw":    RawFunc(),

		// Model module
		"par":      ParFunc(),
		"parbreak": ParbreakFunc(),
		"heading":  HeadingFunc(),
		"list":     ListFunc(),
		"enum":     EnumFunc(),
		"link":     LinkFunc(),
		"table":    TableFunc(),

		// Layout module
		"stack":   StackFunc(),
		"align":   AlignFunc(),
		"columns": ColumnsFunc(),
		"box":     BoxFunc(),
		"block":   BlockFunc(),
		"pad":     PadFunc(),

		// Visualize module
		"image": ImageFunc(),
	}
}
