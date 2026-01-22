// Rule evaluation for Typst set and show rules.
// Translated from typst-eval/src/rules.rs

package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Set Rule Evaluation
// ----------------------------------------------------------------------------

// evalSetRuleToStyles evaluates a set rule and returns Styles.
// This is the internal function that returns the styles directly.
// Matches Rust: impl Eval for ast::SetRule (Output = Styles)
func evalSetRuleToStyles(vm *Vm, e *syntax.SetRuleExpr) (*foundations.Styles, error) {
	// Check optional condition
	if condExpr := e.Condition(); condExpr != nil {
		condVal, err := evalExpr(vm, condExpr)
		if err != nil {
			return nil, err
		}
		cond, ok := condVal.(foundations.Bool)
		if !ok {
			return nil, atSpan(fmt.Errorf("expected boolean, found %s", condVal.Type()), condExpr.ToUntyped().Span())
		}
		// If condition is false, return empty styles
		if !cond {
			return foundations.NewStyles(), nil
		}
	}

	// Evaluate the target expression (the function)
	targetExpr := e.Target()
	if targetExpr == nil {
		return nil, &SetRuleError{
			Message: "set rule missing target",
			Span:    e.ToUntyped().Span(),
		}
	}

	targetVal, err := evalExpr(vm, targetExpr)
	if err != nil {
		return nil, err
	}

	// Cast to function
	funcVal, ok := targetVal.(foundations.FuncValue)
	if !ok {
		err := atSpan(fmt.Errorf("expected function, found %s", targetVal.Type()), targetExpr.ToUntyped().Span())
		return nil, hintIfShadowedStd(vm, targetExpr, err)
	}

	// Get the element from the function
	elem := funcVal.Func.ToElement()
	if elem == nil {
		return nil, &SetRuleError{
			Message: "only element functions can be used in set rules",
			Span:    targetExpr.ToUntyped().Span(),
		}
	}

	// Evaluate arguments
	args, err := evalArgs(vm, e.Args())
	if err != nil {
		return nil, err
	}
	args = args.WithSpan(e.ToUntyped().Span())

	// Apply set rule to element
	styles, err := elem.Set(vm.Engine, args)
	if err != nil {
		return nil, err
	}

	// Mark styles as liftable and add span
	return styles.WithSpan(e.ToUntyped().Span()).Liftable(), nil
}

// evalSetRule evaluates a set rule expression and returns a StylesValue.
// This is the public function that returns a Value.
func evalSetRule(vm *Vm, e *syntax.SetRuleExpr) (foundations.Value, error) {
	styles, err := evalSetRuleToStyles(vm, e)
	if err != nil {
		return nil, err
	}
	return foundations.StylesValue{Styles: styles}, nil
}

// ----------------------------------------------------------------------------
// Show Rule Evaluation
// ----------------------------------------------------------------------------

// evalShowRuleToRecipe evaluates a show rule and returns a Recipe.
// This is the internal function that returns the recipe directly.
// Matches Rust: impl Eval for ast::ShowRule (Output = Recipe)
func evalShowRuleToRecipe(vm *Vm, e *syntax.ShowRuleExpr) (*foundations.Recipe, error) {
	// Evaluate optional selector
	var selector foundations.Selector
	if selExpr := e.Selector(); selExpr != nil {
		selVal, err := evalExpr(vm, selExpr)
		if err != nil {
			return nil, err
		}

		showableSel, err := castToShowableSelector(selVal, selExpr)
		if err != nil {
			return nil, hintIfShadowedStd(vm, selExpr, err)
		}
		selector = showableSel.Selector
	}

	// Evaluate transform expression
	transformExpr := e.Transform()
	if transformExpr == nil {
		return nil, &ShowRuleError{
			Message: "show rule missing transform",
			Span:    e.ToUntyped().Span(),
		}
	}

	var transform foundations.Transformation

	// Special case: if transform is a set rule, convert to StyleTransformation
	if setRule, ok := transformExpr.(*syntax.SetRuleExpr); ok {
		styles, err := evalSetRuleToStyles(vm, setRule)
		if err != nil {
			return nil, err
		}
		transform = foundations.StyleTransformation{Styles: styles}
	} else {
		transVal, err := evalExpr(vm, transformExpr)
		if err != nil {
			return nil, err
		}
		trans, err := castToTransformation(transVal, transformExpr.ToUntyped().Span())
		if err != nil {
			return nil, err
		}
		transform = trans
	}

	// Create recipe
	recipe := foundations.NewRecipe(selector, transform, e.ToUntyped().Span())

	// Validation warnings
	checkShowPageRule(vm, recipe)
	checkShowParSetBlock(vm, recipe)

	return recipe, nil
}

// evalShowRule evaluates a show rule expression and returns a StylesValue.
// This is the public function that returns a Value.
func evalShowRule(vm *Vm, e *syntax.ShowRuleExpr) (foundations.Value, error) {
	recipe, err := evalShowRuleToRecipe(vm, e)
	if err != nil {
		return nil, err
	}

	// Wrap recipe in styles
	styles := foundations.NewStyles()
	styles.AddRecipe(recipe)
	return foundations.StylesValue{Styles: styles}, nil
}

// ----------------------------------------------------------------------------
// Type Casting
// ----------------------------------------------------------------------------

// castToShowableSelector casts a value to a ShowableSelector.
func castToShowableSelector(val foundations.Value, expr syntax.Expr) (*foundations.ShowableSelector, error) {
	span := expr.ToUntyped().Span()

	switch v := val.(type) {
	case foundations.FuncValue:
		// Function selector - must be an element function
		if v.Func != nil {
			if elem := v.Func.ToElement(); elem != nil {
				return &foundations.ShowableSelector{
					Selector: foundations.ElemSelector{Element: *elem},
				}, nil
			}
		}
		return nil, atSpan(fmt.Errorf("expected element function, found function"), span)

	case foundations.LabelValue:
		return &foundations.ShowableSelector{
			Selector: foundations.LabelSelector{Label: string(v)},
		}, nil

	case foundations.Str:
		return &foundations.ShowableSelector{
			Selector: foundations.RegexSelector{Pattern: string(v)},
		}, nil

	case foundations.TypeValue:
		return &foundations.ShowableSelector{
			Selector: foundations.ElemSelector{Element: foundations.Element{Name: v.Inner.String()}},
		}, nil

	default:
		return nil, atSpan(fmt.Errorf("expected selector (function, label, string, or type), found %s", val.Type()), span)
	}
}

// castToTransformation casts a value to a Transformation.
func castToTransformation(val foundations.Value, span syntax.Span) (foundations.Transformation, error) {
	switch v := val.(type) {
	case foundations.FuncValue:
		return foundations.FuncTransformation{Func: v.Func}, nil

	case foundations.ContentValue:
		return foundations.ContentTransformation{Content: v.Content}, nil

	case foundations.NoneValue:
		return foundations.NoneTransformation{}, nil

	case foundations.StylesValue:
		return foundations.StyleTransformation{Styles: v.Styles}, nil

	default:
		return nil, atSpan(fmt.Errorf("expected transformation (function, content, none, or styles), found %s", val.Type()), span)
	}
}

// ----------------------------------------------------------------------------
// Validation Helpers
// ----------------------------------------------------------------------------

// checkShowPageRule warns if a show rule targets pages.
// Matches Rust: fn check_show_page_rule(vm: &mut Vm, recipe: &Recipe)
func checkShowPageRule(vm *Vm, recipe *foundations.Recipe) {
	if recipe.Selector == nil {
		return
	}
	elemSel, ok := recipe.Selector.(foundations.ElemSelector)
	if !ok {
		return
	}
	if elemSel.Element.Name == "page" {
		vm.Engine.Sink.Warn(foundations.SourceDiagnostic{
			Span:     recipe.Span,
			Severity: foundations.SeverityWarning,
			Message:  "`show page` is not supported and has no effect",
			Hints:    []string{"customize pages with `set page(..)` instead"},
		})
	}
}

// checkShowParSetBlock warns about deprecated show par: set block patterns.
// Matches Rust: fn check_show_par_set_block(vm: &mut Vm, recipe: &Recipe)
func checkShowParSetBlock(vm *Vm, recipe *foundations.Recipe) {
	if recipe.Selector == nil {
		return
	}
	elemSel, ok := recipe.Selector.(foundations.ElemSelector)
	if !ok {
		return
	}
	if elemSel.Element.Name != "par" {
		return
	}

	// Check if transform is a style transformation with block above/below
	styleTrans, ok := recipe.Transform.(foundations.StyleTransformation)
	if !ok {
		return
	}

	if styleTrans.Styles.HasBlockAbove() || styleTrans.Styles.HasBlockBelow() {
		vm.Engine.Sink.Warn(foundations.SourceDiagnostic{
			Span:     recipe.Span,
			Severity: foundations.SeverityWarning,
			Message:  "`show par: set block(spacing: ..)` has no effect anymore",
			Hints: []string{
				"write `set par(spacing: ..)` instead",
				"this is specific to paragraphs as they are not considered blocks anymore",
			},
		})
	}
}

// hintIfShadowedStd adds a hint if an identifier shadows a standard library function.
// Matches Rust: fn hint_if_shadowed_std(vm: &Vm, expr: &ast::Expr, err: EcoString) -> EcoString
func hintIfShadowedStd(vm *Vm, expr syntax.Expr, err error) error {
	ident, ok := expr.(*syntax.IdentExpr)
	if !ok {
		return err
	}

	name := ident.Get()

	// Check if there's a standard library function with this name
	if vm.Engine != nil && vm.Engine.World != nil {
		lib := vm.Engine.World.Library()
		if lib != nil {
			if binding := lib.Get(name); binding != nil {
				// The name exists in the standard library but was shadowed
				if spannedErr, ok := err.(*SpannedError); ok {
					return &HintedError{
						Err:   spannedErr.Err,
						Hints: []string{fmt.Sprintf("a variable named `%s` is shadowing a function from the standard library", name)},
						Span:  spannedErr.ErrSpan,
					}
				}
			}
		}
	}

	return err
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// SetRuleError is returned when a set rule is invalid.
type SetRuleError struct {
	Message string
	Span    syntax.Span
}

func (e *SetRuleError) Error() string {
	return e.Message
}

// ShowRuleError is returned when a show rule is invalid.
type ShowRuleError struct {
	Message string
	Span    syntax.Span
}

func (e *ShowRuleError) Error() string {
	return e.Message
}

