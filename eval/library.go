package eval

import (
	"math"
	"strings"

	"github.com/boergens/gotypst/syntax"
)

// Library returns the standard library scope containing all built-in
// functions, types, and prelude values.
//
// The returned scope includes:
//   - Prelude colors (black, white, red, blue, etc.)
//   - Alignments and directions (left, right, center, top, bottom, start, end)
//   - Type values (int, str, float, bool, etc.)
//   - The calc module with mathematical functions
//   - Element functions (raw, par, parbreak)
//   - Math accent functions (hat, tilde, bar, vec, etc.)
func Library() *Scope {
	scope := NewScope()

	// Register prelude colors
	registerColors(scope)

	// Register alignments and directions
	registerAlignments(scope)

	// Register type values
	registerTypes(scope)

	// Register calc module
	registerCalcModule(scope)

	// Register element functions
	RegisterElementFunctions(scope)

	// Register math accent functions
	registerMathAccentFunctions(scope)

	// Register text utility functions
	registerTextFunctions(scope)

	return scope
}

// ----------------------------------------------------------------------------
// Prelude Colors
// ----------------------------------------------------------------------------

// registerColors adds all standard colors to the scope.
func registerColors(scope *Scope) {
	// Standard web colors
	colors := map[string]Color{
		// Grayscale
		"black":  {R: 0, G: 0, B: 0, A: 255},
		"gray":   {R: 170, G: 170, B: 170, A: 255},
		"silver": {R: 221, G: 222, B: 223, A: 255},
		"white":  {R: 255, G: 255, B: 255, A: 255},

		// Blues
		"navy":    {R: 0, G: 31, B: 98, A: 255},
		"blue":    {R: 0, G: 101, B: 189, A: 255},
		"aqua":    {R: 36, G: 138, B: 168, A: 255},
		"teal":    {R: 0, G: 150, B: 137, A: 255},
		"eastern": {R: 33, G: 145, B: 192, A: 255},

		// Purples/Pinks
		"purple":  {R: 150, G: 55, B: 145, A: 255},
		"fuchsia": {R: 206, G: 49, B: 137, A: 255},

		// Reds/Oranges
		"maroon": {R: 144, G: 12, B: 63, A: 255},
		"red":    {R: 255, G: 67, B: 67, A: 255},
		"orange": {R: 255, G: 130, B: 67, A: 255},

		// Yellows/Greens
		"yellow": {R: 255, G: 220, B: 0, A: 255},
		"olive":  {R: 121, G: 148, B: 41, A: 255},
		"green":  {R: 31, G: 172, B: 79, A: 255},
		"lime":   {R: 0, G: 166, B: 82, A: 255},
	}

	for name, color := range colors {
		scope.Define(name, ColorValue{Color: color}, syntax.Detached())
	}
}

// ----------------------------------------------------------------------------
// Alignments and Directions
// ----------------------------------------------------------------------------

// Alignment represents a horizontal or vertical alignment value.
type Alignment struct {
	// Horizontal alignment (-1 = start, 0 = center, 1 = end)
	H *float64
	// Vertical alignment (-1 = top, 0 = horizon, 1 = bottom)
	V *float64
}

// AlignmentValue wraps an Alignment as a Value.
type AlignmentValue struct {
	Alignment Alignment
}

func (AlignmentValue) Type() Type          { return TypeDyn }
func (v AlignmentValue) Display() Content  { return Content{} }
func (v AlignmentValue) Clone() Value      { return v }
func (AlignmentValue) isValue()            {}

// Direction represents a text direction value.
type Direction int

const (
	DirectionLTR Direction = iota
	DirectionRTL
	DirectionTTB
	DirectionBTT
)

// DirectionValue wraps a Direction as a Value.
type DirectionValue struct {
	Direction Direction
}

func (DirectionValue) Type() Type          { return TypeDyn }
func (v DirectionValue) Display() Content  { return Content{} }
func (v DirectionValue) Clone() Value      { return v }
func (DirectionValue) isValue()            {}

// registerAlignments adds alignment and direction values to the scope.
func registerAlignments(scope *Scope) {
	// Horizontal alignments
	start := -1.0
	center := 0.0
	end := 1.0

	scope.Define("start", AlignmentValue{Alignment: Alignment{H: &start}}, syntax.Detached())
	scope.Define("center", AlignmentValue{Alignment: Alignment{H: &center}}, syntax.Detached())
	scope.Define("end", AlignmentValue{Alignment: Alignment{H: &end}}, syntax.Detached())
	scope.Define("left", AlignmentValue{Alignment: Alignment{H: &start}}, syntax.Detached())
	scope.Define("right", AlignmentValue{Alignment: Alignment{H: &end}}, syntax.Detached())

	// Vertical alignments
	top := -1.0
	horizon := 0.0
	bottom := 1.0

	scope.Define("top", AlignmentValue{Alignment: Alignment{V: &top}}, syntax.Detached())
	scope.Define("horizon", AlignmentValue{Alignment: Alignment{V: &horizon}}, syntax.Detached())
	scope.Define("bottom", AlignmentValue{Alignment: Alignment{V: &bottom}}, syntax.Detached())

	// Text directions
	scope.Define("ltr", DirectionValue{Direction: DirectionLTR}, syntax.Detached())
	scope.Define("rtl", DirectionValue{Direction: DirectionRTL}, syntax.Detached())
	scope.Define("ttb", DirectionValue{Direction: DirectionTTB}, syntax.Detached())
	scope.Define("btt", DirectionValue{Direction: DirectionBTT}, syntax.Detached())
}

// ----------------------------------------------------------------------------
// Type Values
// ----------------------------------------------------------------------------

// registerTypes adds type constructor values to the scope.
func registerTypes(scope *Scope) {
	types := []Type{
		TypeNone,
		TypeAuto,
		TypeBool,
		TypeInt,
		TypeFloat,
		TypeLength,
		TypeAngle,
		TypeRatio,
		TypeRelative,
		TypeFraction,
		TypeStr,
		TypeBytes,
		TypeLabel,
		TypeDatetime,
		TypeDuration,
		TypeDecimal,
		TypeColor,
		TypeGradient,
		TypeTiling,
		TypeSymbol,
		TypeContent,
		TypeArray,
		TypeDict,
		TypeFunc,
		TypeArgs,
		TypeType,
		TypeModule,
		TypeStyles,
		TypeVersion,
	}

	for _, t := range types {
		scope.Define(t.Ident(), TypeValue{Inner: t}, syntax.Detached())
	}
}

// ----------------------------------------------------------------------------
// Calc Module
// ----------------------------------------------------------------------------

// registerCalcModule adds the calc module to the scope.
func registerCalcModule(scope *Scope) {
	calcScope := NewScope()

	// Constants
	calcScope.Define("pi", FloatValue(math.Pi), syntax.Detached())
	calcScope.Define("e", FloatValue(math.E), syntax.Detached())
	calcScope.Define("inf", FloatValue(math.Inf(1)), syntax.Detached())
	calcScope.Define("nan", FloatValue(math.NaN()), syntax.Detached())

	// Basic math functions
	calcScope.DefineFunc("abs", calcAbsFunc())
	calcScope.DefineFunc("pow", calcPowFunc())
	calcScope.DefineFunc("exp", calcExpFunc())
	calcScope.DefineFunc("sqrt", calcSqrtFunc())
	calcScope.DefineFunc("root", calcRootFunc())
	calcScope.DefineFunc("log", calcLogFunc())
	calcScope.DefineFunc("ln", calcLnFunc())

	// Trigonometric functions
	calcScope.DefineFunc("sin", calcSinFunc())
	calcScope.DefineFunc("cos", calcCosFunc())
	calcScope.DefineFunc("tan", calcTanFunc())
	calcScope.DefineFunc("asin", calcAsinFunc())
	calcScope.DefineFunc("acos", calcAcosFunc())
	calcScope.DefineFunc("atan", calcAtanFunc())
	calcScope.DefineFunc("atan2", calcAtan2Func())
	calcScope.DefineFunc("sinh", calcSinhFunc())
	calcScope.DefineFunc("cosh", calcCoshFunc())
	calcScope.DefineFunc("tanh", calcTanhFunc())

	// Rounding functions
	calcScope.DefineFunc("floor", calcFloorFunc())
	calcScope.DefineFunc("ceil", calcCeilFunc())
	calcScope.DefineFunc("round", calcRoundFunc())
	calcScope.DefineFunc("trunc", calcTruncFunc())

	// Comparison functions
	calcScope.DefineFunc("min", calcMinFunc())
	calcScope.DefineFunc("max", calcMaxFunc())
	calcScope.DefineFunc("clamp", calcClampFunc())

	// Other functions
	calcScope.DefineFunc("rem", calcRemFunc())
	calcScope.DefineFunc("quo", calcQuoFunc())

	// Create calc module
	calcModule := &Module{
		Name:  "calc",
		Scope: calcScope,
	}

	scope.Define("calc", ModuleValue{Module: calcModule}, syntax.Detached())
}

// Helper to convert Value to float64 for calc functions.
func toFloat64(v Value) (float64, bool) {
	switch x := v.(type) {
	case IntValue:
		return float64(x), true
	case FloatValue:
		return float64(x), true
	default:
		return 0, false
	}
}

// calcAbsFunc creates the calc.abs function.
func calcAbsFunc() *Func {
	name := "abs"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				switch x := arg.V.(type) {
				case IntValue:
					if x < 0 {
						return IntValue(-x), nil
					}
					return x, nil
				case FloatValue:
					return FloatValue(math.Abs(float64(x))), nil
				default:
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
			},
			Info: &FuncInfo{
				Name:   "abs",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcPowFunc creates the calc.pow function.
func calcPowFunc() *Func {
	name := "pow"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				base, err := args.Expect("base")
				if err != nil {
					return nil, err
				}
				exp, err := args.Expect("exponent")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				baseF, ok := toFloat64(base.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      base.V.Type().String(),
						Span:     base.Span,
					}
				}
				expF, ok := toFloat64(exp.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      exp.V.Type().String(),
						Span:     exp.Span,
					}
				}
				return FloatValue(math.Pow(baseF, expF)), nil
			},
			Info: &FuncInfo{
				Name: "pow",
				Params: []ParamInfo{
					{Name: "base", Type: TypeFloat},
					{Name: "exponent", Type: TypeFloat},
				},
			},
		},
	}
}

// calcExpFunc creates the calc.exp function.
func calcExpFunc() *Func {
	name := "exp"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("exponent")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Exp(x)), nil
			},
			Info: &FuncInfo{
				Name:   "exp",
				Params: []ParamInfo{{Name: "exponent", Type: TypeFloat}},
			},
		},
	}
}

// calcSqrtFunc creates the calc.sqrt function.
func calcSqrtFunc() *Func {
	name := "sqrt"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Sqrt(x)), nil
			},
			Info: &FuncInfo{
				Name:   "sqrt",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcRootFunc creates the calc.root function.
func calcRootFunc() *Func {
	name := "root"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				radicand, err := args.Expect("radicand")
				if err != nil {
					return nil, err
				}
				index, err := args.Expect("index")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				x, ok := toFloat64(radicand.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      radicand.V.Type().String(),
						Span:     radicand.Span,
					}
				}
				n, ok := toFloat64(index.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      index.V.Type().String(),
						Span:     index.Span,
					}
				}

				return FloatValue(math.Pow(x, 1/n)), nil
			},
			Info: &FuncInfo{
				Name: "root",
				Params: []ParamInfo{
					{Name: "radicand", Type: TypeFloat},
					{Name: "index", Type: TypeFloat},
				},
			},
		},
	}
}

// calcLogFunc creates the calc.log function.
func calcLogFunc() *Func {
	name := "log"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				// Optional base argument (default: 10)
				baseArg := args.Find("base")
				if err := args.Finish(); err != nil {
					return nil, err
				}

				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}

				base := 10.0
				if baseArg != nil {
					b, ok := toFloat64(baseArg.V)
					if !ok {
						return nil, &TypeMismatchError{
							Expected: "number",
							Got:      baseArg.V.Type().String(),
							Span:     baseArg.Span,
						}
					}
					base = b
				}

				return FloatValue(math.Log(x) / math.Log(base)), nil
			},
			Info: &FuncInfo{
				Name: "log",
				Params: []ParamInfo{
					{Name: "value", Type: TypeFloat},
					{Name: "base", Type: TypeFloat, Default: FloatValue(10), Named: true},
				},
			},
		},
	}
}

// calcLnFunc creates the calc.ln function (natural logarithm).
func calcLnFunc() *Func {
	name := "ln"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Log(x)), nil
			},
			Info: &FuncInfo{
				Name:   "ln",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcSinFunc creates the calc.sin function.
func calcSinFunc() *Func {
	name := "sin"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("angle")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				// Handle both raw numbers (radians) and Angle values
				var radians float64
				switch v := arg.V.(type) {
				case AngleValue:
					radians = v.Angle.Radians
				case IntValue:
					radians = float64(v)
				case FloatValue:
					radians = float64(v)
				default:
					return nil, &TypeMismatchError{
						Expected: "angle or number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Sin(radians)), nil
			},
			Info: &FuncInfo{
				Name:   "sin",
				Params: []ParamInfo{{Name: "angle", Type: TypeAngle}},
			},
		},
	}
}

// calcCosFunc creates the calc.cos function.
func calcCosFunc() *Func {
	name := "cos"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("angle")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				var radians float64
				switch v := arg.V.(type) {
				case AngleValue:
					radians = v.Angle.Radians
				case IntValue:
					radians = float64(v)
				case FloatValue:
					radians = float64(v)
				default:
					return nil, &TypeMismatchError{
						Expected: "angle or number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Cos(radians)), nil
			},
			Info: &FuncInfo{
				Name:   "cos",
				Params: []ParamInfo{{Name: "angle", Type: TypeAngle}},
			},
		},
	}
}

// calcTanFunc creates the calc.tan function.
func calcTanFunc() *Func {
	name := "tan"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("angle")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				var radians float64
				switch v := arg.V.(type) {
				case AngleValue:
					radians = v.Angle.Radians
				case IntValue:
					radians = float64(v)
				case FloatValue:
					radians = float64(v)
				default:
					return nil, &TypeMismatchError{
						Expected: "angle or number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Tan(radians)), nil
			},
			Info: &FuncInfo{
				Name:   "tan",
				Params: []ParamInfo{{Name: "angle", Type: TypeAngle}},
			},
		},
	}
}

// calcAsinFunc creates the calc.asin function.
func calcAsinFunc() *Func {
	name := "asin"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return AngleValue{Angle: Angle{Radians: math.Asin(x)}}, nil
			},
			Info: &FuncInfo{
				Name:   "asin",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcAcosFunc creates the calc.acos function.
func calcAcosFunc() *Func {
	name := "acos"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return AngleValue{Angle: Angle{Radians: math.Acos(x)}}, nil
			},
			Info: &FuncInfo{
				Name:   "acos",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcAtanFunc creates the calc.atan function.
func calcAtanFunc() *Func {
	name := "atan"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return AngleValue{Angle: Angle{Radians: math.Atan(x)}}, nil
			},
			Info: &FuncInfo{
				Name:   "atan",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcAtan2Func creates the calc.atan2 function.
func calcAtan2Func() *Func {
	name := "atan2"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				y, err := args.Expect("y")
				if err != nil {
					return nil, err
				}
				x, err := args.Expect("x")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				yF, ok := toFloat64(y.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      y.V.Type().String(),
						Span:     y.Span,
					}
				}
				xF, ok := toFloat64(x.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      x.V.Type().String(),
						Span:     x.Span,
					}
				}
				return AngleValue{Angle: Angle{Radians: math.Atan2(yF, xF)}}, nil
			},
			Info: &FuncInfo{
				Name: "atan2",
				Params: []ParamInfo{
					{Name: "y", Type: TypeFloat},
					{Name: "x", Type: TypeFloat},
				},
			},
		},
	}
}

// calcSinhFunc creates the calc.sinh function.
func calcSinhFunc() *Func {
	name := "sinh"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Sinh(x)), nil
			},
			Info: &FuncInfo{
				Name:   "sinh",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcCoshFunc creates the calc.cosh function.
func calcCoshFunc() *Func {
	name := "cosh"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Cosh(x)), nil
			},
			Info: &FuncInfo{
				Name:   "cosh",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcTanhFunc creates the calc.tanh function.
func calcTanhFunc() *Func {
	name := "tanh"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return FloatValue(math.Tanh(x)), nil
			},
			Info: &FuncInfo{
				Name:   "tanh",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcFloorFunc creates the calc.floor function.
func calcFloorFunc() *Func {
	name := "floor"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return IntValue(int64(math.Floor(x))), nil
			},
			Info: &FuncInfo{
				Name:   "floor",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcCeilFunc creates the calc.ceil function.
func calcCeilFunc() *Func {
	name := "ceil"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return IntValue(int64(math.Ceil(x))), nil
			},
			Info: &FuncInfo{
				Name:   "ceil",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcRoundFunc creates the calc.round function.
func calcRoundFunc() *Func {
	name := "round"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				// Optional digits argument
				digitsArg := args.Find("digits")
				if err := args.Finish(); err != nil {
					return nil, err
				}

				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}

				if digitsArg != nil {
					d, ok := digitsArg.V.(IntValue)
					if !ok {
						return nil, &TypeMismatchError{
							Expected: "integer",
							Got:      digitsArg.V.Type().String(),
							Span:     digitsArg.Span,
						}
					}
					multiplier := math.Pow(10, float64(d))
					return FloatValue(math.Round(x*multiplier) / multiplier), nil
				}

				return IntValue(int64(math.Round(x))), nil
			},
			Info: &FuncInfo{
				Name: "round",
				Params: []ParamInfo{
					{Name: "value", Type: TypeFloat},
					{Name: "digits", Type: TypeInt, Default: None, Named: true},
				},
			},
		},
	}
}

// calcTruncFunc creates the calc.trunc function.
func calcTruncFunc() *Func {
	name := "trunc"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				arg, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}
				x, ok := toFloat64(arg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      arg.V.Type().String(),
						Span:     arg.Span,
					}
				}
				return IntValue(int64(math.Trunc(x))), nil
			},
			Info: &FuncInfo{
				Name:   "trunc",
				Params: []ParamInfo{{Name: "value", Type: TypeFloat}},
			},
		},
	}
}

// calcMinFunc creates the calc.min function.
func calcMinFunc() *Func {
	name := "min"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				first, err := args.Expect("values")
				if err != nil {
					return nil, err
				}

				minVal := first.V
				minF, ok := toFloat64(minVal)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      first.V.Type().String(),
						Span:     first.Span,
					}
				}

				// Consume remaining positional arguments
				for {
					next := args.Eat()
					if next == nil {
						break
					}
					nextF, ok := toFloat64(next.V)
					if !ok {
						return nil, &TypeMismatchError{
							Expected: "number",
							Got:      next.V.Type().String(),
							Span:     next.Span,
						}
					}
					if nextF < minF {
						minF = nextF
						minVal = next.V
					}
				}

				if err := args.Finish(); err != nil {
					return nil, err
				}

				return minVal, nil
			},
			Info: &FuncInfo{
				Name:   "min",
				Params: []ParamInfo{{Name: "values", Type: TypeFloat, Variadic: true}},
			},
		},
	}
}

// calcMaxFunc creates the calc.max function.
func calcMaxFunc() *Func {
	name := "max"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				first, err := args.Expect("values")
				if err != nil {
					return nil, err
				}

				maxVal := first.V
				maxF, ok := toFloat64(maxVal)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      first.V.Type().String(),
						Span:     first.Span,
					}
				}

				// Consume remaining positional arguments
				for {
					next := args.Eat()
					if next == nil {
						break
					}
					nextF, ok := toFloat64(next.V)
					if !ok {
						return nil, &TypeMismatchError{
							Expected: "number",
							Got:      next.V.Type().String(),
							Span:     next.Span,
						}
					}
					if nextF > maxF {
						maxF = nextF
						maxVal = next.V
					}
				}

				if err := args.Finish(); err != nil {
					return nil, err
				}

				return maxVal, nil
			},
			Info: &FuncInfo{
				Name:   "max",
				Params: []ParamInfo{{Name: "values", Type: TypeFloat, Variadic: true}},
			},
		},
	}
}

// calcClampFunc creates the calc.clamp function.
func calcClampFunc() *Func {
	name := "clamp"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				val, err := args.Expect("value")
				if err != nil {
					return nil, err
				}
				minArg, err := args.Expect("min")
				if err != nil {
					return nil, err
				}
				maxArg, err := args.Expect("max")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				v, ok := toFloat64(val.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      val.V.Type().String(),
						Span:     val.Span,
					}
				}
				minV, ok := toFloat64(minArg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      minArg.V.Type().String(),
						Span:     minArg.Span,
					}
				}
				maxV, ok := toFloat64(maxArg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      maxArg.V.Type().String(),
						Span:     maxArg.Span,
					}
				}

				// Clamp while preserving type
				clamped := math.Max(minV, math.Min(maxV, v))
				if _, isInt := val.V.(IntValue); isInt {
					return IntValue(int64(clamped)), nil
				}
				return FloatValue(clamped), nil
			},
			Info: &FuncInfo{
				Name: "clamp",
				Params: []ParamInfo{
					{Name: "value", Type: TypeFloat},
					{Name: "min", Type: TypeFloat},
					{Name: "max", Type: TypeFloat},
				},
			},
		},
	}
}

// calcRemFunc creates the calc.rem function (remainder).
func calcRemFunc() *Func {
	name := "rem"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				dividend, err := args.Expect("dividend")
				if err != nil {
					return nil, err
				}
				divisor, err := args.Expect("divisor")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				a, ok := toFloat64(dividend.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      dividend.V.Type().String(),
						Span:     dividend.Span,
					}
				}
				b, ok := toFloat64(divisor.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      divisor.V.Type().String(),
						Span:     divisor.Span,
					}
				}

				// Integer rem for integer operands
				if ai, aIsInt := dividend.V.(IntValue); aIsInt {
					if bi, bIsInt := divisor.V.(IntValue); bIsInt {
						if bi == 0 {
							return nil, &DivisionByZeroError{Span: divisor.Span}
						}
						return IntValue(ai % bi), nil
					}
				}

				return FloatValue(math.Remainder(a, b)), nil
			},
			Info: &FuncInfo{
				Name: "rem",
				Params: []ParamInfo{
					{Name: "dividend", Type: TypeFloat},
					{Name: "divisor", Type: TypeFloat},
				},
			},
		},
	}
}

// calcQuoFunc creates the calc.quo function (integer quotient).
func calcQuoFunc() *Func {
	name := "quo"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				dividend, err := args.Expect("dividend")
				if err != nil {
					return nil, err
				}
				divisor, err := args.Expect("divisor")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				a, ok := toFloat64(dividend.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      dividend.V.Type().String(),
						Span:     dividend.Span,
					}
				}
				b, ok := toFloat64(divisor.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "number",
						Got:      divisor.V.Type().String(),
						Span:     divisor.Span,
					}
				}

				if b == 0 {
					return nil, &DivisionByZeroError{Span: divisor.Span}
				}

				return IntValue(int64(a / b)), nil
			},
			Info: &FuncInfo{
				Name: "quo",
				Params: []ParamInfo{
					{Name: "dividend", Type: TypeFloat},
					{Name: "divisor", Type: TypeFloat},
				},
			},
		},
	}
}

// ----------------------------------------------------------------------------
// Math Accent Functions
// ----------------------------------------------------------------------------

// registerMathAccentFunctions adds math accent functions to the scope.
// These functions create accented math content (hat, tilde, bar, vec, etc.).
func registerMathAccentFunctions(scope *Scope) {
	// Primary accent functions
	scope.DefineFunc("hat", mathAccentFunc("hat", AccentHat))
	scope.DefineFunc("tilde", mathAccentFunc("tilde", AccentTilde))
	scope.DefineFunc("bar", mathAccentFunc("bar", AccentBar))
	scope.DefineFunc("overline", mathAccentFunc("overline", AccentBar)) // Alias for bar
	scope.DefineFunc("vec", mathAccentFunc("vec", AccentVec))

	// Additional accent functions
	scope.DefineFunc("dot", mathAccentFunc("dot", AccentDot))
	scope.DefineFunc("ddot", mathAccentFunc("ddot", AccentDDot))
	scope.DefineFunc("breve", mathAccentFunc("breve", AccentBreve))
	scope.DefineFunc("acute", mathAccentFunc("acute", AccentAcute))
	scope.DefineFunc("grave", mathAccentFunc("grave", AccentGrave))
}

// mathAccentFunc creates a math accent function for the given accent kind.
func mathAccentFunc(name string, kind AccentKind) *Func {
	funcName := name
	return &Func{
		Name: &funcName,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				// Get required base argument (content)
				baseArg, err := args.Expect("base")
				if err != nil {
					return nil, err
				}

				if err := args.Finish(); err != nil {
					return nil, err
				}

				// Convert argument to content
				var baseContent Content
				switch v := baseArg.V.(type) {
				case ContentValue:
					baseContent = v.Content
				case StrValue:
					baseContent = Content{Elements: []ContentElement{&TextElement{Text: string(v)}}}
				default:
					baseContent = Content{Elements: []ContentElement{&TextElement{Text: v.Display().String()}}}
				}

				// Create the accent element
				return ContentValue{Content: Content{
					Elements: []ContentElement{&MathAccentElement{
						Base:   baseContent,
						Accent: kind,
					}},
				}}, nil
			},
			Info: &FuncInfo{
				Name:   name,
				Params: []ParamInfo{{Name: "base", Type: TypeContent}},
			},
		},
	}
}

// ----------------------------------------------------------------------------
// Text Utility Functions
// ----------------------------------------------------------------------------

// registerTextFunctions adds text utility functions to the scope.
func registerTextFunctions(scope *Scope) {
	scope.DefineFunc("lorem", loremFunc())
}

// loremWords contains the Lorem Ipsum text generated by the lipsum crate with seed 97.
// This matches Typst's deterministic output exactly.
var loremWords = []string{
	"Lorem", "ipsum", "dolor", "sit", "amet,", "consectetur", "adipiscing", "elit,", "sed", "do",
	"eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore", "magnam", "aliquam", "quaerat",
	"voluptatem.", "Ut", "enim", "aeque", "doleamus", "animo,", "cum", "corpore", "dolemus,", "fieri",
	"tamen", "permagna", "accessio", "potest,", "si", "aliquod", "aeternum", "et", "infinitum", "impendere",
	"malum", "nobis", "opinemur.", "Quod", "idem", "licet", "transferre", "in", "voluptatem,", "ut",
	"postea", "variari", "voluptas", "distinguique", "possit,", "augeri", "amplificarique", "non", "possit.", "At",
	"etiam", "Athenis,", "ut", "e", "patre", "audiebam", "facete", "et", "urbane", "Stoicos",
	"irridente,", "statua", "est", "in", "quo", "a", "nobis", "philosophia", "defensa", "et",
	"collaudata", "est,", "cum", "id,", "quod", "maxime", "placeat,", "facere", "possimus,", "omnis",
	"voluptas", "assumenda", "est,", "omnis", "dolor", "repellendus.", "Temporibus", "autem", "quibusdam", "et",
	"aut", "officiis", "debitis", "aut", "rerum", "necessitatibus", "saepe", "eveniet,", "ut", "et",
	"voluptates", "repudiandae", "sint", "et", "molestiae", "non", "recusandae.", "Itaque", "earum", "rerum",
	"defuturum,", "quas", "natura", "non", "depravata", "desiderat.", "Et", "quem", "ad", "me",
	"accedis,", "saluto:", "'chaere,'", "inquam,", "'Tite!'", "lictores,", "turma", "omnis", "chorusque:", "'chaere,",
	"Tite!'", "hinc", "hostis", "mi", "Albucius,", "hinc", "inimicus.", "Sed", "iure", "Mucius.",
	"Ego", "autem", "mirari", "satis", "non", "queo", "unde", "hoc", "sit", "tam",
	"insolens", "domesticarum", "rerum", "fastidium.", "Non", "est", "omnino", "hic", "docendi", "locus;",
	"sed", "ita", "prorsus", "existimo,", "neque", "eum", "Torquatum,", "qui", "hoc", "primus",
	"cognomen", "invenerit,", "aut", "torquem", "illum", "hosti", "detraxisse,", "ut", "aliquam", "ex",
	"eo", "est", "consecutus?", "–", "Laudem", "et", "caritatem,", "quae", "sunt", "vitae",
	"sine", "metu", "degendae", "praesidia", "firmissima.", "–", "Filium", "morte", "multavit.", "–",
	"Si", "sine", "causa,", "nollem", "me", "ab", "eo", "delectari,", "quod", "ista",
	"Platonis,", "Aristoteli,", "Theophrasti", "orationis", "ornamenta", "neglexerit.", "Nam", "illud", "quidem", "physici,",
	"credere", "aliquid", "esse", "minimum,", "quod", "profecto", "numquam", "putavisset,", "si", "a",
	"Polyaeno,", "familiari", "suo,", "geometrica", "discere", "maluisset", "quam", "illum", "etiam", "ipsum",
	"dedocere.", "Sol", "Democrito", "magnus", "videtur,", "quippe", "homini", "erudito", "in", "geometriaque",
	"perfecto,", "huic", "pedalis", "fortasse;", "tantum", "enim", "esse", "omnino", "in", "nostris",
	"poetis", "aut", "inertissimae", "segnitiae", "est", "aut", "fastidii", "delicatissimi.", "Mihi", "quidem",
	"videtur,", "inermis", "ac", "nudus", "est.", "Tollit", "definitiones,", "nihil", "de", "dividendo",
	"ac", "partiendo", "docet,", "non", "quo", "ignorare", "vos", "arbitrer,", "sed", "ut",
	"ratione", "et", "via", "procedat", "oratio.", "Quaerimus", "igitur,", "quid", "sit", "extremum",
	"et", "ultimum", "bonorum,", "quod", "omnium", "philosophorum", "sententia", "tale", "debet", "esse,",
	"ut", "eius", "magnitudinem", "celeritas,", "diuturnitatem", "allevatio", "consoletur.", "Ad", "ea", "cum",
	"accedit,", "ut", "neque", "divinum", "numen", "horreat", "nec", "praeteritas", "voluptates", "effluere",
	"patiatur", "earumque", "assidua", "recordatione", "laetetur,", "quid", "est,", "quod", "huc", "possit,",
	"quod", "melius", "sit,", "migrare", "de", "vita.", "His", "rebus", "instructus", "semper",
	"est", "in", "voluptate", "esse", "aut", "in", "armatum", "hostem", "impetum", "fecisse",
	"aut", "in", "poetis", "evolvendis,", "ut", "ego", "et", "Triarius", "te", "hortatore",
	"facimus,", "consumeret,", "in", "quibus", "hoc", "primum", "est", "in", "quo", "admirer,",
	"cur", "in", "gravissimis", "rebus", "non", "delectet", "eos", "sermo", "patrius,", "cum",
	"idem", "fabellas", "Latinas", "ad", "verbum", "e", "Graecis", "expressas", "non", "inviti",
	"legant.", "Quis", "enim", "tam", "inimicus", "paene", "nomini", "Romano", "est,", "qui",
	"Ennii", "Medeam", "aut", "Antiopam", "Pacuvii", "spernat", "aut", "reiciat,", "quod", "se",
	"isdem", "Euripidis", "fabulis", "delectari", "dicat,", "Latinas", "litteras", "oderit?", "Synephebos", "ego,",
	"inquit,", "potius", "Caecilii", "aut", "Andriam", "Terentii", "quam", "utramque", "Menandri", "legam?",
	"A", "quibus", "tantum", "dissentio,", "ut,", "cum", "Sophocles", "vel", "optime", "scripserit",
	"Electram,", "tamen", "male", "conversam", "Atilii", "mihi", "legendam", "putem,", "de", "quo",
	"Lucilius:", "'ferreum", "scriptorem',", "verum,", "opinor,", "scriptorem", "tamen,", "ut", "legendus", "sit.",
	"Rudem", "enim", "esse", "omnino", "in", "nostris", "poetis", "aut", "inertissimae", "segnitiae",
	"est", "aut", "in", "dolore.", "Omnis", "autem", "privatione", "doloris", "putat", "Epicurus",
	"terminari", "summam", "voluptatem,", "ut", "postea", "variari", "voluptas", "distinguique", "possit,", "augeri",
	"amplificarique", "non", "possit.", "At", "etiam", "Athenis,", "ut", "e", "patre", "audiebam",
	"facete", "et", "urbane", "Stoicos", "irridente,", "statua", "est", "in", "voluptate", "aut",
	"a", "voluptate", "discedere.", "Nam", "cum", "ignoratione", "rerum", "bonarum", "et", "malarum",
	"maxime", "hominum", "vita", "vexetur,", "ob", "eumque", "errorem", "et", "voluptatibus", "maximis",
	"saepe", "priventur", "et", "durissimis", "animi", "doloribus", "torqueantur,", "sapientia", "est", "adhibenda,",
	"quae", "et", "terroribus", "cupiditatibusque", "detractis", "et", "omnium", "falsarum", "opinionum", "temeritate",
	"derepta", "certissimam", "se", "nobis", "ducem", "praebeat", "ad", "voluptatem.", "Sapientia", "enim",
	"est", "una,", "quae", "maestitiam", "pellat", "ex", "animis,", "quae", "nos", "exhorrescere",
	"metu", "non", "sinat.", "Qua", "praeceptrice", "in", "tranquillitate", "vivi", "potest", "omnium",
	"cupiditatum", "ardore", "restincto.", "Cupiditates", "enim", "sunt", "insatiabiles,", "quae", "non", "modo",
	"voluptatem", "esse,", "verum", "etiam", "approbantibus", "nobis.", "Sic", "enim", "ab", "Epicuro",
	"reprehensa", "et", "correcta", "permulta.", "Nunc", "dicam", "de", "voluptate,", "nihil", "scilicet",
	"novi,", "ea", "tamen,", "quae", "te", "ipsum", "probaturum", "esse", "confidam.", "Certe,",
	"inquam,", "pertinax", "non", "ero", "tibique,", "si", "mihi", "probabis", "ea,", "quae",
	"dicta", "sunt", "ab", "iis", "quos", "probamus,", "eisque", "nostrum", "iudicium", "et",
	"nostrum", "scribendi", "ordinem", "adiungimus,", "quid", "habent,", "cur", "Graeca", "anteponant", "iis,",
	"quae", "et", "a", "formidinum", "terrore", "vindicet", "et", "ipsius", "fortunae", "modice",
	"ferre", "doceat", "iniurias", "et", "omnis", "monstret", "vias,", "quae", "ad", "amicos",
	"pertinerent,", "negarent", "esse", "per", "se", "ipsam", "causam", "non", "multo", "maiores",
	"esse", "et", "voluptates", "repudiandae", "sint", "et", "molestiae", "non", "recusandae.", "Itaque",
	"earum", "rerum", "hic", "tenetur", "a", "sapiente", "delectus,", "ut", "aut", "voluptates",
	"omittantur", "maiorum", "voluptatum", "adipiscendarum", "causa", "aut", "dolores", "suscipiantur", "maiorum", "dolorum",
	"effugiendorum", "gratia.", "Sed", "de", "clarorum", "hominum", "factis", "illustribus", "et", "gloriosis",
	"satis", "hoc", "loco", "dictum", "sit.", "Erit", "enim", "iam", "de", "omnium",
	"virtutum", "cursu", "ad", "voluptatem", "proprius", "disserendi", "locus.", "Nunc", "autem", "explicabo,",
	"voluptas", "ipsa", "quae", "qualisque", "sit,", "ut", "tollatur", "error", "omnis", "imperitorum",
	"intellegaturque", "ea,", "quae", "voluptaria,", "delicata,", "mollis", "habeatur", "disciplina,", "quam", "gravis,",
	"quam", "continens,", "quam", "severa", "sit.", "Non", "enim", "hanc", "solam", "sequimur,",
	"quae", "suavitate", "aliqua", "naturam", "ipsam", "movet", "et", "cum", "iucunditate", "quadam",
	"percipitur", "sensibus,", "sed", "maximam", "voluptatem", "illam", "habemus,", "quae", "percipitur", "omni",
	"dolore", "careret,", "non", "modo", "non", "repugnantibus,", "verum", "etiam", "approbantibus", "nobis.",
	"Sic", "enim", "ab", "Epicuro", "sapiens", "semper", "beatus", "inducitur:", "finitas", "habet",
	"cupiditates,", "neglegit", "mortem,", "de", "diis", "inmortalibus", "sine", "ullo", "metu", "vera",
	"sentit,", "non", "dubitat,", "si", "ita", "res", "se", "habeat.", "Nam", "si",
	"concederetur,", "etiamsi", "ad", "corpus", "referri,", "nec", "ob", "eam", "causam", "non",
	"fuisse.", "–", "Torquem", "detraxit", "hosti.", "–", "Et", "quidem", "se", "texit,",
	"ne", "interiret.", "–", "At", "magnum", "periculum", "adiit.", "–", "In", "oculis",
	"quidem", "exercitus.", "–", "Quid", "ex", "eo", "est", "consecutus?", "–", "Laudem",
	"et", "caritatem,", "quae", "sunt", "vitae", "sine", "metu", "degendae", "praesidia", "firmissima.",
	"–", "Filium", "morte", "multavit.", "–", "Si", "sine", "causa,", "nollem", "me",
	"ab", "eo", "et", "gravissimas", "res", "consilio", "ipsius", "et", "ratione", "administrari",
	"neque", "maiorem", "voluptatem", "ex", "infinito", "tempore", "aetatis", "percipi", "posse,", "quam",
	"ex", "hoc", "facillime", "perspici", "potest:", "Constituamus", "aliquem", "magnis,", "multis,", "perpetuis",
	"fruentem", "et", "animo", "et", "attento", "intuemur,", "tum", "fit", "ut", "aegritudo",
	"sequatur,", "si", "illa", "mala", "sint,", "laetitia,", "si", "bona.", "O", "praeclaram",
	"beate", "vivendi", "et", "apertam", "et", "simplicem", "et", "directam", "viam!", "Cum",
	"enim", "certe", "nihil", "homini", "possit", "melius", "esse", "quam", "Graecam.", "Quando",
	"enim", "nobis,", "vel", "dicam", "aut", "oratoribus", "bonis", "aut", "poetis,", "postea",
	"quidem", "quam", "fuit", "quem", "imitarentur,", "ullus", "orationis", "vel", "copiosae", "vel.",
}

// loremFunc creates the lorem function for generating placeholder text.
func loremFunc() *Func {
	name := "lorem"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				// Get required words argument
				wordsArg, err := args.Expect("words")
				if err != nil {
					return nil, err
				}
				if err := args.Finish(); err != nil {
					return nil, err
				}

				// Convert to integer
				n, ok := AsInt(wordsArg.V)
				if !ok {
					return nil, &TypeMismatchError{
						Expected: "integer",
						Got:      wordsArg.V.Type().String(),
						Span:     wordsArg.Span,
					}
				}

				if n <= 0 {
					return StrValue(""), nil
				}

				// Generate n words of Lorem Ipsum text
				// This matches Typst's behavior via the lipsum crate
				wordCount := int(n)
				result := make([]string, wordCount)
				loremLen := len(loremWords)

				for i := 0; i < wordCount; i++ {
					result[i] = loremWords[i%loremLen]
				}

				// Join words and apply sentence-ending logic (matching lipsum crate)
				text := strings.Join(result, " ")

				// Strip trailing punctuation and add period if needed
				// This matches lipsum's join_words() behavior
				lastRune := []rune(text)[len([]rune(text))-1]
				if lastRune != '.' && lastRune != '!' && lastRune != '?' {
					// Trim trailing punctuation before adding period
					text = strings.TrimRight(text, ",:;")
					text += "."
				}

				return StrValue(text), nil
			},
			Info: &FuncInfo{
				Name:   "lorem",
				Params: []ParamInfo{{Name: "words", Type: TypeInt}},
			},
		},
	}
}

