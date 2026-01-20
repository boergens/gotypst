package library

import (
	"math"
	"testing"

	"github.com/boergens/gotypst/eval"
)

func TestLibrary(t *testing.T) {
	lib := Library()

	t.Run("returns non-nil scope", func(t *testing.T) {
		if lib == nil {
			t.Fatal("Library() returned nil")
		}
	})

	t.Run("has prelude values", func(t *testing.T) {
		prelude := []string{"none", "true", "false", "auto"}
		for _, name := range prelude {
			binding := lib.Get(name)
			if binding == nil {
				t.Errorf("missing prelude value: %s", name)
			}
		}
	})

	t.Run("has standard colors", func(t *testing.T) {
		colors := []string{"black", "white", "red", "green", "blue", "yellow"}
		for _, name := range colors {
			binding := lib.Get(name)
			if binding == nil {
				t.Errorf("missing color: %s", name)
			}
			if binding != nil {
				if _, ok := binding.Value.(eval.ColorValue); !ok {
					t.Errorf("color %s is not a ColorValue, got %T", name, binding.Value)
				}
			}
		}
	})

	t.Run("has alignment values", func(t *testing.T) {
		alignments := []string{"left", "right", "center", "top", "bottom", "start", "end"}
		for _, name := range alignments {
			binding := lib.Get(name)
			if binding == nil {
				t.Errorf("missing alignment: %s", name)
			}
			if binding != nil {
				if dyn, ok := binding.Value.(eval.DynValue); ok {
					if dyn.TypeName != "alignment" {
						t.Errorf("alignment %s has wrong type name: %s", name, dyn.TypeName)
					}
				} else {
					t.Errorf("alignment %s is not a DynValue, got %T", name, binding.Value)
				}
			}
		}
	})

	t.Run("has direction values", func(t *testing.T) {
		directions := []string{"ltr", "rtl", "ttb", "btt"}
		for _, name := range directions {
			binding := lib.Get(name)
			if binding == nil {
				t.Errorf("missing direction: %s", name)
			}
			if binding != nil {
				if dyn, ok := binding.Value.(eval.DynValue); ok {
					if dyn.TypeName != "direction" {
						t.Errorf("direction %s has wrong type name: %s", name, dyn.TypeName)
					}
				} else {
					t.Errorf("direction %s is not a DynValue, got %T", name, binding.Value)
				}
			}
		}
	})

	t.Run("has calc module", func(t *testing.T) {
		calcBinding := lib.Get("calc")
		if calcBinding == nil {
			t.Fatal("missing calc module")
		}
		calcMod, ok := calcBinding.Value.(eval.ModuleValue)
		if !ok {
			t.Fatalf("calc is not a ModuleValue, got %T", calcBinding.Value)
		}

		// Check calc constants
		t.Run("calc.pi", func(t *testing.T) {
			piBinding := calcMod.Module.Scope.Get("pi")
			if piBinding == nil {
				t.Fatal("missing calc.pi")
			}
			piVal, ok := piBinding.Value.(eval.FloatValue)
			if !ok {
				t.Fatalf("pi is not a FloatValue, got %T", piBinding.Value)
			}
			if float64(piVal) != math.Pi {
				t.Errorf("pi = %v, want %v", piVal, math.Pi)
			}
		})

		t.Run("calc.e", func(t *testing.T) {
			eBinding := calcMod.Module.Scope.Get("e")
			if eBinding == nil {
				t.Fatal("missing calc.e")
			}
			eVal, ok := eBinding.Value.(eval.FloatValue)
			if !ok {
				t.Fatalf("e is not a FloatValue, got %T", eBinding.Value)
			}
			if float64(eVal) != math.E {
				t.Errorf("e = %v, want %v", eVal, math.E)
			}
		})

		t.Run("calc.inf", func(t *testing.T) {
			infBinding := calcMod.Module.Scope.Get("inf")
			if infBinding == nil {
				t.Fatal("missing calc.inf")
			}
			infVal, ok := infBinding.Value.(eval.FloatValue)
			if !ok {
				t.Fatalf("inf is not a FloatValue, got %T", infBinding.Value)
			}
			if !math.IsInf(float64(infVal), 1) {
				t.Errorf("inf = %v, want +Inf", infVal)
			}
		})

		// Check calc functions
		calcFuncs := []string{"sin", "cos", "tan", "asin", "acos", "atan", "abs", "sqrt", "pow", "exp", "ln", "log", "floor", "ceil", "round", "min", "max"}
		for _, name := range calcFuncs {
			binding := calcMod.Module.Scope.Get(name)
			if binding == nil {
				t.Errorf("missing calc function: %s", name)
			}
			if binding != nil {
				if _, ok := binding.Value.(eval.FuncValue); !ok {
					t.Errorf("calc.%s is not a FuncValue, got %T", name, binding.Value)
				}
			}
		}
	})

	t.Run("has element functions", func(t *testing.T) {
		elements := []string{"raw", "par", "parbreak"}
		for _, name := range elements {
			binding := lib.Get(name)
			if binding == nil {
				t.Errorf("missing element function: %s", name)
			}
			if binding != nil {
				if _, ok := binding.Value.(eval.FuncValue); !ok {
					t.Errorf("element %s is not a FuncValue, got %T", name, binding.Value)
				}
			}
		}
	})
}

func TestColorValues(t *testing.T) {
	lib := Library()

	testCases := []struct {
		name     string
		expected eval.Color
	}{
		{"black", eval.Color{R: 0, G: 0, B: 0, A: 255}},
		{"white", eval.Color{R: 255, G: 255, B: 255, A: 255}},
		{"red", eval.Color{R: 255, G: 0, B: 0, A: 255}},
		{"green", eval.Color{R: 0, G: 128, B: 0, A: 255}},
		{"blue", eval.Color{R: 0, G: 0, B: 255, A: 255}},
		{"yellow", eval.Color{R: 255, G: 255, B: 0, A: 255}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			binding := lib.Get(tc.name)
			if binding == nil {
				t.Fatalf("color %s not found", tc.name)
			}
			colorVal, ok := binding.Value.(eval.ColorValue)
			if !ok {
				t.Fatalf("expected ColorValue, got %T", binding.Value)
			}
			if colorVal.Color != tc.expected {
				t.Errorf("color %s = %+v, want %+v", tc.name, colorVal.Color, tc.expected)
			}
		})
	}
}
