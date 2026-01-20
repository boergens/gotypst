package eval

import (
	"math"
	"testing"
)

func TestLibrary(t *testing.T) {
	lib := Library()
	if lib == nil {
		t.Fatal("Library() returned nil")
	}
}

func TestLibrary_Colors(t *testing.T) {
	lib := Library()

	colors := []string{"black", "white", "red", "green", "blue", "gray", "silver",
		"navy", "aqua", "teal", "eastern", "purple", "fuchsia", "maroon",
		"orange", "yellow", "olive", "lime"}

	for _, name := range colors {
		binding := lib.Get(name)
		if binding == nil {
			t.Errorf("Expected color %q to be defined", name)
			continue
		}
		if _, ok := binding.Value.(ColorValue); !ok {
			t.Errorf("Expected %q to be a ColorValue, got %T", name, binding.Value)
		}
	}
}

func TestLibrary_Alignments(t *testing.T) {
	lib := Library()

	alignments := []string{"start", "center", "end", "left", "right", "top", "bottom", "horizon"}
	for _, name := range alignments {
		binding := lib.Get(name)
		if binding == nil {
			t.Errorf("Expected alignment %q to be defined", name)
			continue
		}
		if _, ok := binding.Value.(AlignmentValue); !ok {
			t.Errorf("Expected %q to be an AlignmentValue, got %T", name, binding.Value)
		}
	}

	directions := []string{"ltr", "rtl", "ttb", "btt"}
	for _, name := range directions {
		binding := lib.Get(name)
		if binding == nil {
			t.Errorf("Expected direction %q to be defined", name)
			continue
		}
		if _, ok := binding.Value.(DirectionValue); !ok {
			t.Errorf("Expected %q to be a DirectionValue, got %T", name, binding.Value)
		}
	}
}

func TestLibrary_Types(t *testing.T) {
	lib := Library()

	types := []string{"none", "auto", "bool", "int", "float", "str", "array", "dictionary", "content"}
	for _, name := range types {
		binding := lib.Get(name)
		if binding == nil {
			t.Errorf("Expected type %q to be defined", name)
			continue
		}
		if _, ok := binding.Value.(TypeValue); !ok {
			t.Errorf("Expected %q to be a TypeValue, got %T", name, binding.Value)
		}
	}
}

func TestLibrary_CalcModule(t *testing.T) {
	lib := Library()

	calcBinding := lib.Get("calc")
	if calcBinding == nil {
		t.Fatal("Expected calc module to be defined")
	}

	calcModule, ok := calcBinding.Value.(ModuleValue)
	if !ok {
		t.Fatalf("Expected calc to be a ModuleValue, got %T", calcBinding.Value)
	}

	// Check constants
	constants := []string{"pi", "e", "inf", "nan"}
	for _, name := range constants {
		binding := calcModule.Module.Scope.Get(name)
		if binding == nil {
			t.Errorf("Expected calc.%s to be defined", name)
		}
	}

	// Check functions
	funcs := []string{"abs", "pow", "exp", "sqrt", "root", "log", "ln",
		"sin", "cos", "tan", "asin", "acos", "atan", "atan2",
		"sinh", "cosh", "tanh",
		"floor", "ceil", "round", "trunc",
		"min", "max", "clamp", "rem", "quo"}

	for _, name := range funcs {
		binding := calcModule.Module.Scope.Get(name)
		if binding == nil {
			t.Errorf("Expected calc.%s to be defined", name)
			continue
		}
		if _, ok := binding.Value.(FuncValue); !ok {
			t.Errorf("Expected calc.%s to be a FuncValue, got %T", name, binding.Value)
		}
	}
}

func TestLibrary_ElementFunctions(t *testing.T) {
	lib := Library()

	elements := []string{"raw", "par", "parbreak"}
	for _, name := range elements {
		binding := lib.Get(name)
		if binding == nil {
			t.Errorf("Expected element function %q to be defined", name)
			continue
		}
		if _, ok := binding.Value.(FuncValue); !ok {
			t.Errorf("Expected %q to be a FuncValue, got %T", name, binding.Value)
		}
	}
}

func TestLibrary_CalcConstants(t *testing.T) {
	lib := Library()
	calcBinding := lib.Get("calc")
	calcModule := calcBinding.Value.(ModuleValue)
	calcScope := calcModule.Module.Scope

	// Test pi
	piBinding := calcScope.Get("pi")
	if piBinding == nil {
		t.Fatal("calc.pi not found")
	}
	if pi, ok := piBinding.Value.(FloatValue); ok {
		if math.Abs(float64(pi)-math.Pi) > 1e-10 {
			t.Errorf("calc.pi = %v, expected %v", pi, math.Pi)
		}
	} else {
		t.Errorf("calc.pi is not a FloatValue: %T", piBinding.Value)
	}

	// Test e
	eBinding := calcScope.Get("e")
	if eBinding == nil {
		t.Fatal("calc.e not found")
	}
	if e, ok := eBinding.Value.(FloatValue); ok {
		if math.Abs(float64(e)-math.E) > 1e-10 {
			t.Errorf("calc.e = %v, expected %v", e, math.E)
		}
	} else {
		t.Errorf("calc.e is not a FloatValue: %T", eBinding.Value)
	}

	// Test inf
	infBinding := calcScope.Get("inf")
	if infBinding == nil {
		t.Fatal("calc.inf not found")
	}
	if inf, ok := infBinding.Value.(FloatValue); ok {
		if !math.IsInf(float64(inf), 1) {
			t.Errorf("calc.inf = %v, expected +Inf", inf)
		}
	} else {
		t.Errorf("calc.inf is not a FloatValue: %T", infBinding.Value)
	}

	// Test nan
	nanBinding := calcScope.Get("nan")
	if nanBinding == nil {
		t.Fatal("calc.nan not found")
	}
	if nan, ok := nanBinding.Value.(FloatValue); ok {
		if !math.IsNaN(float64(nan)) {
			t.Errorf("calc.nan = %v, expected NaN", nan)
		}
	} else {
		t.Errorf("calc.nan is not a FloatValue: %T", nanBinding.Value)
	}
}

func TestLibrary_ColorValues(t *testing.T) {
	lib := Library()

	// Test specific color values
	tests := []struct {
		name string
		r, g, b, a uint8
	}{
		{"black", 0, 0, 0, 255},
		{"white", 255, 255, 255, 255},
		{"red", 255, 67, 67, 255},
		{"green", 31, 172, 79, 255},
		{"blue", 0, 101, 189, 255},
	}

	for _, tc := range tests {
		binding := lib.Get(tc.name)
		if binding == nil {
			t.Errorf("color %q not found", tc.name)
			continue
		}
		cv, ok := binding.Value.(ColorValue)
		if !ok {
			t.Errorf("%q is not a ColorValue", tc.name)
			continue
		}
		if cv.Color.R != tc.r || cv.Color.G != tc.g || cv.Color.B != tc.b || cv.Color.A != tc.a {
			t.Errorf("%q = rgba(%d,%d,%d,%d), expected rgba(%d,%d,%d,%d)",
				tc.name, cv.Color.R, cv.Color.G, cv.Color.B, cv.Color.A,
				tc.r, tc.g, tc.b, tc.a)
		}
	}
}
