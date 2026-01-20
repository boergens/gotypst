package eval

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Rule Types
// ----------------------------------------------------------------------------

// Recipe represents a show rule recipe that defines how to transform content.
type Recipe struct {
	// Selector optionally restricts which content the recipe applies to.
	Selector *Selector

	// Transform defines how to transform matching content.
	Transform Transformation

	// Span is the source location of the recipe.
	Span syntax.Span
}

// NewRecipe creates a new recipe with the given components.
func NewRecipe(selector *Selector, transform Transformation, span syntax.Span) *Recipe {
	return &Recipe{
		Selector:  selector,
		Transform: transform,
		Span:      span,
	}
}

// Selector represents a selector for matching content in show rules.
type Selector interface {
	isSelector()
}

// ElemSelector matches content of a specific element type.
type ElemSelector struct {
	// Element is the element type to match.
	Element Element

	// Where is an optional filter function.
	Where *Func
}

func (ElemSelector) isSelector() {}

// LabelSelector matches content with a specific label.
type LabelSelector struct {
	Label string
}

func (LabelSelector) isSelector() {}

// TextSelector matches text content using a pattern.
type TextSelector struct {
	// Text is the text pattern to match (string or regex).
	Text string
	// IsRegex indicates if Text is a regular expression.
	IsRegex bool
}

func (TextSelector) isSelector() {}

// FuncSelector matches content based on a predicate function.
type FuncSelector struct {
	Func *Func
}

func (FuncSelector) isSelector() {}

// OrSelector matches content that matches any of the selectors.
type OrSelector struct {
	Selectors []Selector
}

func (OrSelector) isSelector() {}

// AndSelector matches content that matches all selectors.
type AndSelector struct {
	Selectors []Selector
}

func (AndSelector) isSelector() {}

// BeforeSelector matches content before another selector matches.
type BeforeSelector struct {
	Selector Selector
	End      Selector
}

func (BeforeSelector) isSelector() {}

// AfterSelector matches content after another selector matches.
type AfterSelector struct {
	Selector Selector
	Start    Selector
}

func (AfterSelector) isSelector() {}

// Element represents an element type (e.g., text, heading, par).
type Element struct {
	// Name is the element name.
	Name string
}

// ShowableSelector wraps a Selector for show rule usage.
// This type exists for type-safe casting from values.
type ShowableSelector struct {
	Selector Selector
}

// Transformation represents how content should be transformed in a show rule.
type Transformation interface {
	isTransformation()
}

// StyleTransformation applies styles to content.
type StyleTransformation struct {
	Styles *Styles
}

func (StyleTransformation) isTransformation() {}

// FuncTransformation applies a function to transform content.
type FuncTransformation struct {
	Func *Func
}

func (FuncTransformation) isTransformation() {}

// ContentTransformation replaces content directly.
type ContentTransformation struct {
	Content Content
}

func (ContentTransformation) isTransformation() {}

// NoneTransformation hides the matched content.
type NoneTransformation struct{}

func (NoneTransformation) isTransformation() {}

// ----------------------------------------------------------------------------
// Set Rule Evaluation
// ----------------------------------------------------------------------------

// evalSetRule evaluates a set rule expression.
// Set rules configure element defaults: `set text(fill: red)`.
func evalSetRule(vm *Vm, e *syntax.SetRuleExpr) (Value, error) {
	// Check optional condition
	if condExpr := e.Condition(); condExpr != nil {
		condVal, err := EvalExpr(vm, condExpr)
		if err != nil {
			return nil, err
		}
		// Cast to bool
		cond, ok := condVal.(BoolValue)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      condVal.Type().String(),
				Span:     condExpr.ToUntyped().Span(),
			}
		}
		// If condition is false, return empty styles
		if !cond {
			return StylesValue{Styles: &Styles{}}, nil
		}
	}

	// Get the target expression.
	// In Typst syntax, `set text(fill: red)` has a FuncCall as the target.
	targetExpr := e.Target()
	if targetExpr == nil {
		return nil, &SetRuleError{
			Message: "set rule missing target",
			Span:    e.ToUntyped().Span(),
		}
	}

	// The target should be a function call expression
	funcCallExpr, ok := targetExpr.(*syntax.FuncCallExpr)
	if !ok {
		return nil, &SetRuleError{
			Message: "set rule target must be a function call",
			Span:    targetExpr.ToUntyped().Span(),
		}
	}

	// Evaluate the callee to get the function
	calleeExpr := funcCallExpr.Callee()
	if calleeExpr == nil {
		return nil, &SetRuleError{
			Message: "set rule missing function",
			Span:    funcCallExpr.ToUntyped().Span(),
		}
	}

	calleeVal, err := EvalExpr(vm, calleeExpr)
	if err != nil {
		return nil, err
	}

	// Cast to function, with hint if shadowed
	funcVal, ok := calleeVal.(FuncValue)
	if !ok {
		err := &TypeMismatchError{
			Expected: "function",
			Got:      calleeVal.Type().String(),
			Span:     calleeExpr.ToUntyped().Span(),
		}
		return nil, hintIfShadowedStd(vm, calleeExpr, err)
	}

	// Verify it's an element function (can be used in set rules)
	elem := funcToElement(funcVal.Func)
	if elem == nil {
		return nil, &SetRuleError{
			Message: "only element functions can be used in set rules",
			Span:    calleeExpr.ToUntyped().Span(),
		}
	}

	// Evaluate arguments from the function call
	argsNode := funcCallExpr.Args()
	args, err := evalArgs(vm, argsNode)
	if err != nil {
		return nil, err
	}
	args.Span = e.ToUntyped().Span()

	// Apply set rule to get styles
	styles, err := applySetRule(vm.Engine, elem, args)
	if err != nil {
		return nil, err
	}

	// Return styles scoped to the rule's span
	styles = scopeStyles(styles, e.ToUntyped().Span())
	return StylesValue{Styles: styles}, nil
}

// funcToElement extracts the Element from a function if it's an element function.
// Returns nil if the function is not an element function.
func funcToElement(f *Func) *Element {
	if f == nil {
		return nil
	}
	// Element functions are identified by having IsElement = true
	// For now, we check by name against known elements
	if f.Name != nil {
		name := *f.Name
		// Known element functions
		elementNames := map[string]bool{
			"text": true, "par": true, "heading": true,
			"list": true, "enum": true, "terms": true,
			"raw": true, "strong": true, "emph": true,
			"link": true, "ref": true, "footnote": true,
			"cite": true, "quote": true, "figure": true,
			"table": true, "image": true, "rect": true,
			"circle": true, "ellipse": true, "line": true,
			"polygon": true, "path": true, "block": true,
			"box": true, "stack": true, "grid": true,
			"columns": true, "place": true, "align": true,
			"pad": true, "repeat": true, "move": true,
			"rotate": true, "scale": true, "hide": true,
			"page": true, "pagebreak": true, "colbreak": true,
			"linebreak": true, "parbreak": true, "v": true,
			"h": true, "equation": true, "math.frac": true,
			"math.root": true, "math.attach": true, "math.lr": true,
			"outline": true, "bibliography": true, "numbering": true,
		}
		if elementNames[name] {
			return &Element{Name: name}
		}
	}
	return nil
}

// applySetRule applies a set rule to an element, returning the resulting styles.
func applySetRule(engine *Engine, elem *Element, args *Args) (*Styles, error) {
	// Create a style rule from the element and arguments
	rule := StyleRule{
		Func: &Func{
			Name: &elem.Name,
			Span: args.Span,
		},
		Args: args,
	}

	return &Styles{
		Rules: []StyleRule{rule},
	}, nil
}

// scopeStyles wraps styles with span information for error reporting.
func scopeStyles(styles *Styles, span syntax.Span) *Styles {
	// In the full implementation, this would mark the styles as "liftable"
	// and associate them with the span for proper scoping.
	return styles
}

// ----------------------------------------------------------------------------
// Show Rule Evaluation
// ----------------------------------------------------------------------------

// evalShowRule evaluates a show rule expression.
// Show rules transform content: `show heading: it => emph(it.body)`.
func evalShowRule(vm *Vm, e *syntax.ShowRuleExpr) (Value, error) {
	// Evaluate optional selector
	var selector *Selector
	if selExpr := e.Selector(); selExpr != nil {
		selVal, err := EvalExpr(vm, selExpr)
		if err != nil {
			return nil, err
		}

		// Cast to ShowableSelector
		sel, err := castToShowableSelector(selVal, selExpr)
		if err != nil {
			return nil, hintIfShadowedStd(vm, selExpr, err)
		}
		selector = &sel.Selector
	}

	// Evaluate transform expression
	transformExpr := e.Transform()
	if transformExpr == nil {
		return nil, &ShowRuleError{
			Message: "show rule missing transform",
			Span:    e.ToUntyped().Span(),
		}
	}

	var transform Transformation

	// Special case: if transform is a set rule, convert to StyleTransformation
	if setRule, ok := transformExpr.(*syntax.SetRuleExpr); ok {
		stylesVal, err := evalSetRule(vm, setRule)
		if err != nil {
			return nil, err
		}
		styles := stylesVal.(StylesValue).Styles
		transform = StyleTransformation{Styles: styles}
	} else {
		// Evaluate normally and cast to Transformation
		transVal, err := EvalExpr(vm, transformExpr)
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
	recipe := NewRecipe(selector, transform, e.ToUntyped().Span())

	// Validation warnings
	checkShowPageRule(vm, recipe)
	checkShowParSetBlock(vm, recipe)

	// Return recipe wrapped in styles
	// (Show rules produce styles containing recipes)
	return StylesValue{
		Styles: &Styles{
			Recipes: []*Recipe{recipe},
		},
	}, nil
}

// castToShowableSelector casts a value to a ShowableSelector.
func castToShowableSelector(val Value, expr syntax.Expr) (*ShowableSelector, error) {
	span := expr.ToUntyped().Span()

	switch v := val.(type) {
	case FuncValue:
		// Function selector
		elem := funcToElement(v.Func)
		if elem != nil {
			return &ShowableSelector{
				Selector: ElemSelector{Element: *elem},
			}, nil
		}
		// Function as predicate
		return &ShowableSelector{
			Selector: FuncSelector{Func: v.Func},
		}, nil

	case LabelValue:
		return &ShowableSelector{
			Selector: LabelSelector{Label: string(v)},
		}, nil

	case StrValue:
		return &ShowableSelector{
			Selector: TextSelector{Text: string(v)},
		}, nil

	case RegexValue:
		return &ShowableSelector{
			Selector: TextSelector{Text: v.Pattern, IsRegex: true},
		}, nil

	case TypeValue:
		// Type as element selector
		return &ShowableSelector{
			Selector: ElemSelector{Element: Element{Name: v.Inner.String()}},
		}, nil

	default:
		return nil, &TypeMismatchError{
			Expected: "selector (function, label, string, regex, or type)",
			Got:      val.Type().String(),
			Span:     span,
		}
	}
}

// castToTransformation casts a value to a Transformation.
func castToTransformation(val Value, span syntax.Span) (Transformation, error) {
	switch v := val.(type) {
	case FuncValue:
		return FuncTransformation{Func: v.Func}, nil

	case ContentValue:
		return ContentTransformation{Content: v.Content}, nil

	case NoneValue:
		return NoneTransformation{}, nil

	case StylesValue:
		return StyleTransformation{Styles: v.Styles}, nil

	default:
		return nil, &TypeMismatchError{
			Expected: "transformation (function, content, none, or styles)",
			Got:      val.Type().String(),
			Span:     span,
		}
	}
}

// ----------------------------------------------------------------------------
// Validation Helpers
// ----------------------------------------------------------------------------

// checkShowPageRule warns if a show rule targets pages.
func checkShowPageRule(vm *Vm, recipe *Recipe) {
	if recipe.Selector == nil {
		return
	}
	elemSel, ok := (*recipe.Selector).(ElemSelector)
	if !ok {
		return
	}
	if elemSel.Element.Name == "page" {
		vm.Engine.Sink.Warn(SourceDiagnostic{
			Span:     recipe.Span,
			Severity: SeverityWarning,
			Message:  "`show page` is not supported and has no effect",
			Hints:    []string{"customize pages with `set page(..)` instead"},
		})
	}
}

// checkShowParSetBlock warns about deprecated show par: set block patterns.
func checkShowParSetBlock(vm *Vm, recipe *Recipe) {
	if recipe.Selector == nil {
		return
	}
	elemSel, ok := (*recipe.Selector).(ElemSelector)
	if !ok {
		return
	}
	if elemSel.Element.Name != "par" {
		return
	}

	// Check if transform is a style transformation with block spacing
	styleTrans, ok := recipe.Transform.(StyleTransformation)
	if !ok {
		return
	}

	// Check if styles contain block.above or block.below
	for _, rule := range styleTrans.Styles.Rules {
		if rule.Func != nil && rule.Func.Name != nil && *rule.Func.Name == "block" {
			// Check for above/below args
			if rule.Args != nil {
				if rule.Args.HasNamed("above") || rule.Args.HasNamed("below") || rule.Args.HasNamed("spacing") {
					vm.Engine.Sink.Warn(SourceDiagnostic{
						Span:     recipe.Span,
						Severity: SeverityWarning,
						Message:  "`show par: set block(spacing: ..)` has no effect anymore",
						Hints: []string{
							"write `set par(spacing: ..)` instead",
							"this is specific to paragraphs as they are not considered blocks anymore",
						},
					})
					return
				}
			}
		}
	}
}

// hintIfShadowedStd adds a hint to an error if the identifier shadows a standard library function.
func hintIfShadowedStd(vm *Vm, expr syntax.Expr, err error) error {
	// Check if the expression is an identifier
	ident, ok := expr.(*syntax.IdentExpr)
	if !ok {
		return err
	}

	name := ident.Get()

	// Check if there's a standard library function with this name
	if vm.World() != nil {
		lib := vm.World().Library()
		if lib != nil {
			if binding := lib.Get(name); binding != nil {
				// The name exists in the standard library but was shadowed
				// Add a hint to the error
				if tmErr, ok := err.(*TypeMismatchError); ok {
					return &TypeMismatchErrorWithHint{
						TypeMismatchError: *tmErr,
						Hint:              fmt.Sprintf("a variable named `%s` is shadowing a function from the standard library", name),
					}
				}
			}
		}
	}

	return err
}

// ----------------------------------------------------------------------------
// Styles Extensions
// ----------------------------------------------------------------------------

// WithRecipes returns a copy of styles with recipes added.
func (s *Styles) WithRecipes(recipes []*Recipe) *Styles {
	return &Styles{
		Rules:   s.Rules,
		Recipes: recipes,
	}
}

// HasRule checks if styles have a rule for a specific function/property.
func (s *Styles) HasRule(funcName string, propName string) bool {
	for _, rule := range s.Rules {
		if rule.Func != nil && rule.Func.Name != nil && *rule.Func.Name == funcName {
			if rule.Args != nil && rule.Args.HasNamed(propName) {
				return true
			}
		}
	}
	return false
}

// ----------------------------------------------------------------------------
// Additional Error Types
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

// TypeMismatchErrorWithHint extends TypeMismatchError with a hint.
type TypeMismatchErrorWithHint struct {
	TypeMismatchError
	Hint string
}

func (e *TypeMismatchErrorWithHint) Error() string {
	return fmt.Sprintf("%s; hint: %s", e.TypeMismatchError.Error(), e.Hint)
}
