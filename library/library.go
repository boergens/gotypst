// Package library provides the Typst standard library.
//
// This package contains all built-in functions, types, and elements
// that are available in Typst documents by default.
package library

import (
	"math"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/syntax"
)

// Library returns the standard library scope containing all built-in
// functions, types, and prelude values available in Typst documents.
//
// The returned scope includes:
//   - Primitive values: none, true, false
//   - Math constants and functions via the calc module
//   - Standard colors: black, white, red, green, blue, etc.
//   - Alignment values: left, center, right, top, bottom, start, end
//   - Direction values: ltr, rtl, ttb, btt
//   - Element functions: raw, par, parbreak, etc.
//
// This scope is typically passed to NewFileWorld via WithLibrary().
func Library() *eval.Scope {
	scope := eval.NewScope()

	// Register primitive prelude values
	registerPrelude(scope)

	// Register colors
	registerColors(scope)

	// Register alignment and direction values
	registerAlignments(scope)
	registerDirections(scope)

	// Register the calc module with math functions and constants
	registerCalcModule(scope)

	// Register element functions
	eval.RegisterElementFunctions(scope)

	return scope
}

// registerPrelude adds primitive prelude values to the scope.
func registerPrelude(scope *eval.Scope) {
	// none, true, false are handled specially by the evaluator
	// since they're keywords. But we register them here for completeness
	// in case they're accessed via reflection/introspection.
	scope.Define("none", eval.None, syntax.Detached())
	scope.Define("true", eval.True, syntax.Detached())
	scope.Define("false", eval.False, syntax.Detached())
	scope.Define("auto", eval.Auto, syntax.Detached())
}

// registerColors adds standard color values to the scope.
func registerColors(scope *eval.Scope) {
	// Standard web colors
	colors := map[string]eval.Color{
		// Basic colors
		"black":   {R: 0, G: 0, B: 0, A: 255},
		"gray":    {R: 128, G: 128, B: 128, A: 255},
		"silver":  {R: 192, G: 192, B: 192, A: 255},
		"white":   {R: 255, G: 255, B: 255, A: 255},
		"red":     {R: 255, G: 0, B: 0, A: 255},
		"green":   {R: 0, G: 128, B: 0, A: 255},
		"blue":    {R: 0, G: 0, B: 255, A: 255},
		"yellow":  {R: 255, G: 255, B: 0, A: 255},
		"cyan":    {R: 0, G: 255, B: 255, A: 255},
		"magenta": {R: 255, G: 0, B: 255, A: 255},

		// Extended colors
		"maroon":  {R: 128, G: 0, B: 0, A: 255},
		"olive":   {R: 128, G: 128, B: 0, A: 255},
		"lime":    {R: 0, G: 255, B: 0, A: 255},
		"aqua":    {R: 0, G: 255, B: 255, A: 255},
		"teal":    {R: 0, G: 128, B: 128, A: 255},
		"navy":    {R: 0, G: 0, B: 128, A: 255},
		"fuchsia": {R: 255, G: 0, B: 255, A: 255},
		"purple":  {R: 128, G: 0, B: 128, A: 255},
		"orange":  {R: 255, G: 165, B: 0, A: 255},
	}

	for name, color := range colors {
		scope.Define(name, eval.ColorValue{Color: color}, syntax.Detached())
	}
}

// Alignment represents an alignment value for the layout system.
type Alignment int

const (
	AlignStart Alignment = iota
	AlignEnd
	AlignLeft
	AlignRight
	AlignCenter
	AlignTop
	AlignBottom
	AlignHorizon
)

// registerAlignments adds alignment values to the scope.
func registerAlignments(scope *eval.Scope) {
	alignments := map[string]Alignment{
		"start":   AlignStart,
		"end":     AlignEnd,
		"left":    AlignLeft,
		"right":   AlignRight,
		"center":  AlignCenter,
		"top":     AlignTop,
		"bottom":  AlignBottom,
		"horizon": AlignHorizon,
	}

	for name, alignment := range alignments {
		scope.Define(name, eval.DynValue{Inner: alignment, TypeName: "alignment"}, syntax.Detached())
	}
}

// Direction represents a text direction value.
type Direction int

const (
	DirLTR Direction = iota // Left to right
	DirRTL                  // Right to left
	DirTTB                  // Top to bottom
	DirBTT                  // Bottom to top
)

// registerDirections adds direction values to the scope.
func registerDirections(scope *eval.Scope) {
	directions := map[string]Direction{
		"ltr": DirLTR,
		"rtl": DirRTL,
		"ttb": DirTTB,
		"btt": DirBTT,
	}

	for name, dir := range directions {
		scope.Define(name, eval.DynValue{Inner: dir, TypeName: "direction"}, syntax.Detached())
	}
}

// registerCalcModule adds the calc module with math functions and constants.
func registerCalcModule(scope *eval.Scope) {
	// Create the calc module scope
	calcScope := eval.NewScopeWithCategory(&eval.Category{Name: "calc"})

	// Math constants
	calcScope.Define("pi", eval.FloatValue(math.Pi), syntax.Detached())
	calcScope.Define("e", eval.FloatValue(math.E), syntax.Detached())
	calcScope.Define("inf", eval.FloatValue(math.Inf(1)), syntax.Detached())
	calcScope.Define("nan", eval.FloatValue(math.NaN()), syntax.Detached())
	calcScope.Define("tau", eval.FloatValue(2*math.Pi), syntax.Detached())

	// Math functions
	calcScope.DefineFunc("sin", sinFunc())
	calcScope.DefineFunc("cos", cosFunc())
	calcScope.DefineFunc("tan", tanFunc())
	calcScope.DefineFunc("asin", asinFunc())
	calcScope.DefineFunc("acos", acosFunc())
	calcScope.DefineFunc("atan", atanFunc())
	calcScope.DefineFunc("atan2", atan2Func())
	calcScope.DefineFunc("sinh", sinhFunc())
	calcScope.DefineFunc("cosh", coshFunc())
	calcScope.DefineFunc("tanh", tanhFunc())
	calcScope.DefineFunc("abs", absFunc())
	calcScope.DefineFunc("pow", powFunc())
	calcScope.DefineFunc("exp", expFunc())
	calcScope.DefineFunc("sqrt", sqrtFunc())
	calcScope.DefineFunc("ln", lnFunc())
	calcScope.DefineFunc("log", logFunc())
	calcScope.DefineFunc("floor", floorFunc())
	calcScope.DefineFunc("ceil", ceilFunc())
	calcScope.DefineFunc("round", roundFunc())
	calcScope.DefineFunc("min", minFunc())
	calcScope.DefineFunc("max", maxFunc())

	// Register calc as a module
	calcModule := &eval.Module{
		Name:  "calc",
		Scope: calcScope,
	}
	scope.Define("calc", eval.ModuleValue{Module: calcModule}, syntax.Detached())
}

// Helper function to extract a float from an argument value.
func toFloat64(v eval.Value) (float64, bool) {
	switch x := v.(type) {
	case eval.IntValue:
		return float64(x), true
	case eval.FloatValue:
		return float64(x), true
	default:
		return 0, false
	}
}

// numericArg helper for getting a numeric argument.
func numericArg(args *eval.Args, name string) (float64, error) {
	arg, err := args.Expect(name)
	if err != nil {
		return 0, err
	}
	f, ok := toFloat64(arg.V)
	if !ok {
		return 0, &eval.TypeMismatchError{
			Expected: "number",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
	return f, nil
}

// sinFunc returns the sin function.
func sinFunc() *eval.Func {
	name := "sin"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "angle")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Sin(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "sin",
				Params: []eval.ParamInfo{{Name: "angle", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// cosFunc returns the cos function.
func cosFunc() *eval.Func {
	name := "cos"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "angle")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Cos(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "cos",
				Params: []eval.ParamInfo{{Name: "angle", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// tanFunc returns the tan function.
func tanFunc() *eval.Func {
	name := "tan"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "angle")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Tan(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "tan",
				Params: []eval.ParamInfo{{Name: "angle", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// asinFunc returns the asin function.
func asinFunc() *eval.Func {
	name := "asin"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Asin(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "asin",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// acosFunc returns the acos function.
func acosFunc() *eval.Func {
	name := "acos"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Acos(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "acos",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// atanFunc returns the atan function.
func atanFunc() *eval.Func {
	name := "atan"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Atan(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "atan",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// atan2Func returns the atan2 function.
func atan2Func() *eval.Func {
	name := "atan2"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				y, err := numericArg(args, "y")
				if err != nil {
					return nil, err
				}
				x, err := numericArg(args, "x")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Atan2(y, x)), nil
			},
			Info: &eval.FuncInfo{
				Name: "atan2",
				Params: []eval.ParamInfo{
					{Name: "y", Type: eval.TypeFloat, Named: false},
					{Name: "x", Type: eval.TypeFloat, Named: false},
				},
			},
		},
	}
}

// sinhFunc returns the sinh function.
func sinhFunc() *eval.Func {
	name := "sinh"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Sinh(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "sinh",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// coshFunc returns the cosh function.
func coshFunc() *eval.Func {
	name := "cosh"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Cosh(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "cosh",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// tanhFunc returns the tanh function.
func tanhFunc() *eval.Func {
	name := "tanh"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Tanh(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "tanh",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// absFunc returns the abs function.
func absFunc() *eval.Func {
	name := "abs"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				switch v := arg.V.(type) {
				case eval.IntValue:
					if v < 0 {
						return eval.IntValue(-v), nil
					}
					return v, nil
				case eval.FloatValue:
					return eval.FloatValue(math.Abs(float64(v))), nil
				default:
					return nil, &eval.TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
			},
			Info: &eval.FuncInfo{
				Name:   "abs",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// powFunc returns the pow function.
func powFunc() *eval.Func {
	name := "pow"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				base, err := numericArg(args, "base")
				if err != nil {
					return nil, err
				}
				exp, err := numericArg(args, "exponent")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Pow(base, exp)), nil
			},
			Info: &eval.FuncInfo{
				Name: "pow",
				Params: []eval.ParamInfo{
					{Name: "base", Type: eval.TypeFloat, Named: false},
					{Name: "exponent", Type: eval.TypeFloat, Named: false},
				},
			},
		},
	}
}

// expFunc returns the exp function (e^x).
func expFunc() *eval.Func {
	name := "exp"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "exponent")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Exp(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "exp",
				Params: []eval.ParamInfo{{Name: "exponent", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// sqrtFunc returns the sqrt function.
func sqrtFunc() *eval.Func {
	name := "sqrt"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Sqrt(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "sqrt",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// lnFunc returns the ln (natural log) function.
func lnFunc() *eval.Func {
	name := "ln"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.FloatValue(math.Log(x)), nil
			},
			Info: &eval.FuncInfo{
				Name:   "ln",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// logFunc returns the log function (base 10 by default, or custom base).
func logFunc() *eval.Func {
	name := "log"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}

				// Check for optional base argument
				base := 10.0
				if baseArg := args.Find("base"); baseArg != nil {
					b, ok := toFloat64(baseArg.V)
					if !ok {
						return nil, &eval.TypeMismatchError{
							Expected: "number",
							Got:      baseArg.V.Type().String(),
							Span:     baseArg.Span,
						}
					}
					base = b
				}

				if err := args.Finish(); err != nil {
					return nil, err
				}

				// log_b(x) = ln(x) / ln(b)
				result := math.Log(x) / math.Log(base)
				return eval.FloatValue(result), nil
			},
			Info: &eval.FuncInfo{
				Name: "log",
				Params: []eval.ParamInfo{
					{Name: "value", Type: eval.TypeFloat, Named: false},
					{Name: "base", Type: eval.TypeFloat, Default: eval.FloatValue(10), Named: true},
				},
			},
		},
	}
}

// floorFunc returns the floor function.
func floorFunc() *eval.Func {
	name := "floor"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.IntValue(int64(math.Floor(x))), nil
			},
			Info: &eval.FuncInfo{
				Name:   "floor",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// ceilFunc returns the ceil function.
func ceilFunc() *eval.Func {
	name := "ceil"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.IntValue(int64(math.Ceil(x))), nil
			},
			Info: &eval.FuncInfo{
				Name:   "ceil",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// roundFunc returns the round function.
func roundFunc() *eval.Func {
	name := "round"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				x, err := numericArg(args, "value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return eval.IntValue(int64(math.Round(x))), nil
			},
			Info: &eval.FuncInfo{
				Name:   "round",
				Params: []eval.ParamInfo{{Name: "value", Type: eval.TypeFloat, Named: false}},
			},
		},
	}
}

// minFunc returns the min function (variadic).
func minFunc() *eval.Func {
	name := "min"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				// Get all positional arguments
				all := args.All()
				if len(all) == 0 {
					return nil, &eval.MissingArgumentError{What: "values", Span: args.Span}
				}

				if err := args.Finish(); err != nil {
					return nil, err
				}

				// Find minimum
				minVal, ok := toFloat64(all[0].V)
				if !ok {
					return nil, &eval.TypeMismatchError{
						Expected: "number",
						Got:      all[0].V.Type().String(),
						Span:     all[0].Span,
					}
				}
				minIsInt := true
				if _, ok := all[0].V.(eval.IntValue); !ok {
					minIsInt = false
				}

				for _, arg := range all[1:] {
					v, ok := toFloat64(arg.V)
					if !ok {
						return nil, &eval.TypeMismatchError{
							Expected: "number",
							Got:      arg.V.Type().String(),
							Span:     arg.Span,
						}
					}
					if v < minVal {
						minVal = v
						if _, ok := arg.V.(eval.IntValue); !ok {
							minIsInt = false
						}
					}
					if _, ok := arg.V.(eval.IntValue); !ok {
						minIsInt = false
					}
				}

				// Return int if all values were ints
				if minIsInt {
					return eval.IntValue(int64(minVal)), nil
				}
				return eval.FloatValue(minVal), nil
			},
			Info: &eval.FuncInfo{
				Name:   "min",
				Params: []eval.ParamInfo{{Name: "values", Type: eval.TypeFloat, Variadic: true}},
			},
		},
	}
}

// maxFunc returns the max function (variadic).
func maxFunc() *eval.Func {
	name := "max"
	return &eval.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				// Get all positional arguments
				all := args.All()
				if len(all) == 0 {
					return nil, &eval.MissingArgumentError{What: "values", Span: args.Span}
				}

				if err := args.Finish(); err != nil {
					return nil, err
				}

				// Find maximum
				maxVal, ok := toFloat64(all[0].V)
				if !ok {
					return nil, &eval.TypeMismatchError{
						Expected: "number",
						Got:      all[0].V.Type().String(),
						Span:     all[0].Span,
					}
				}
				maxIsInt := true
				if _, ok := all[0].V.(eval.IntValue); !ok {
					maxIsInt = false
				}

				for _, arg := range all[1:] {
					v, ok := toFloat64(arg.V)
					if !ok {
						return nil, &eval.TypeMismatchError{
							Expected: "number",
							Got:      arg.V.Type().String(),
							Span:     arg.Span,
						}
					}
					if v > maxVal {
						maxVal = v
						if _, ok := arg.V.(eval.IntValue); !ok {
							maxIsInt = false
						}
					}
					if _, ok := arg.V.(eval.IntValue); !ok {
						maxIsInt = false
					}
				}

				// Return int if all values were ints
				if maxIsInt {
					return eval.IntValue(int64(maxVal)), nil
				}
				return eval.FloatValue(maxVal), nil
			},
			Info: &eval.FuncInfo{
				Name:   "max",
				Params: []eval.ParamInfo{{Name: "values", Type: eval.TypeFloat, Variadic: true}},
			},
		},
	}
}
