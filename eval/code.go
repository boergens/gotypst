// Code evaluation for Typst.
// Translated from typst-eval/src/code.rs

package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Code Evaluation
// ----------------------------------------------------------------------------

// evalCode evaluates a stream of expressions.
// Matches Rust: fn eval_code(vm: &mut Vm, exprs: &mut impl Iterator<Item = ast::Expr>) -> SourceResult<Value>
func evalCode(vm *Vm, exprs []syntax.Expr) (foundations.Value, error) {
	flow := vm.TakeFlow()
	var output foundations.Value = foundations.None

	i := 0
loop:
	for i < len(exprs) {
		expr := exprs[i]
		span := expr.ToUntyped().Span()
		i++

		var value foundations.Value
		var err error

		switch e := expr.(type) {
		case *syntax.SetRuleExpr:
			// Evaluate set rule to get styles
			styles, err := evalSetRuleToStyles(vm, e)
			if err != nil {
				return nil, err
			}
			if vm.HasFlow() {
				break loop
			}

			// Evaluate the tail (remaining expressions)
			tail, err := evalCode(vm, exprs[i:])
			if err != nil {
				return nil, err
			}
			tailContent := Display(tail)
			value = foundations.ContentValue{Content: tailContent.StyledWithMap(styles)}
			i = len(exprs) // Mark all remaining exprs consumed

		case *syntax.ShowRuleExpr:
			// Evaluate show rule to get recipe
			recipe, err := evalShowRuleToRecipe(vm, e)
			if err != nil {
				return nil, err
			}
			if vm.HasFlow() {
				break loop
			}

			// Evaluate the tail
			tail, err := evalCode(vm, exprs[i:])
			if err != nil {
				return nil, err
			}
			tailContent := Display(tail)
			styledContent, err := tailContent.StyledWithRecipe(vm.Engine, vm.Context, recipe)
			if err != nil {
				return nil, err
			}
			value = foundations.ContentValue{Content: styledContent}
			i = len(exprs)

		default:
			value, err = evalExpr(vm, expr)
		}

		if err != nil {
			return nil, err
		}

		// Join values
		output, err = Join(output, value)
		if err != nil {
			return nil, atSpan(err, span)
		}

		if vm.Flow != nil {
			warnForDiscardedContent(vm.Engine, vm.Flow, output)
			break loop
		}
	}

	if flow != nil {
		vm.SetFlow(flow)
	}

	return output, nil
}

// ----------------------------------------------------------------------------
// Expression Evaluation
// ----------------------------------------------------------------------------

// evalExpr evaluates an expression.
// Matches Rust: impl Eval for ast::Expr
func evalExpr(vm *Vm, expr syntax.Expr) (foundations.Value, error) {
	span := expr.ToUntyped().Span()
	forbidden := func(name string) error {
		return fmt.Errorf("%s is only allowed directly in code and content blocks", name)
	}

	var value foundations.Value
	var err error

	switch e := expr.(type) {
	// Markup expressions
	case *syntax.TextExpr:
		value, err = evalText(vm, e)
	case *syntax.SpaceExpr:
		value, err = evalSpace(vm, e)
	case *syntax.LinebreakExpr:
		value, err = evalLinebreak(vm, e)
	case *syntax.ParbreakExpr:
		value, err = evalParbreak(vm, e)
	case *syntax.EscapeExpr:
		value, err = evalEscape(vm, e)
	case *syntax.ShorthandExpr:
		value, err = evalShorthand(vm, e)
	case *syntax.SmartQuoteExpr:
		value, err = evalSmartQuote(vm, e)
	case *syntax.StrongExpr:
		value, err = evalStrong(vm, e)
	case *syntax.EmphExpr:
		value, err = evalEmph(vm, e)
	case *syntax.RawExpr:
		value, err = evalRaw(vm, e)
	case *syntax.LinkExpr:
		value, err = evalLink(vm, e)
	case *syntax.LabelExpr:
		value, err = evalLabel(vm, e)
	case *syntax.RefExpr:
		value, err = evalRef(vm, e)
	case *syntax.HeadingExpr:
		value, err = evalHeading(vm, e)
	case *syntax.ListItemExpr:
		value, err = evalListItem(vm, e)
	case *syntax.EnumItemExpr:
		value, err = evalEnumItem(vm, e)
	case *syntax.TermItemExpr:
		value, err = evalTermItem(vm, e)

	// Math expressions
	case *syntax.EquationExpr:
		value, err = evalEquation(vm, e)
	case *syntax.MathExpr:
		value, err = evalMath(vm, e)
	case *syntax.MathTextExpr:
		value, err = evalMathText(vm, e)
	case *syntax.MathIdentExpr:
		value, err = evalMathIdent(vm, e)
	case *syntax.MathShorthandExpr:
		value, err = evalMathShorthand(vm, e)
	case *syntax.MathAlignPointExpr:
		value, err = evalMathAlignPoint(vm, e)
	case *syntax.MathDelimitedExpr:
		value, err = evalMathDelimited(vm, e)
	case *syntax.MathAttachExpr:
		value, err = evalMathAttach(vm, e)
	case *syntax.MathPrimesExpr:
		value, err = evalMathPrimes(vm, e)
	case *syntax.MathFracExpr:
		value, err = evalMathFrac(vm, e)
	case *syntax.MathRootExpr:
		value, err = evalMathRoot(vm, e)

	// Literals
	case *syntax.IdentExpr:
		value, err = evalIdent(vm, e)
	case *syntax.NoneExpr:
		value, err = evalNone(vm, e)
	case *syntax.AutoExpr:
		value, err = evalAuto(vm, e)
	case *syntax.BoolExpr:
		value, err = evalBool(vm, e)
	case *syntax.IntExpr:
		value, err = evalInt(vm, e)
	case *syntax.FloatExpr:
		value, err = evalFloat(vm, e)
	case *syntax.NumericExpr:
		value, err = evalNumeric(vm, e)
	case *syntax.StrExpr:
		value, err = evalStr(vm, e)

	// Blocks and grouping
	case *syntax.CodeBlockExpr:
		value, err = evalCodeBlock(vm, e)
	case *syntax.ContentBlockExpr:
		value, err = evalContentBlock(vm, e)
	case *syntax.ArrayExpr:
		value, err = evalArray(vm, e)
	case *syntax.DictExpr:
		value, err = evalDict(vm, e)
	case *syntax.ParenthesizedExpr:
		value, err = evalParenthesized(vm, e)

	// Access and calls
	case *syntax.FieldAccessExpr:
		value, err = evalFieldAccess(vm, e)
	case *syntax.FuncCallExpr:
		value, err = evalFuncCall(vm, e)

	// Closures
	case *syntax.ClosureExpr:
		value, err = evalClosure(vm, e)

	// Operators
	case *syntax.UnaryExpr:
		value, err = evalUnary(vm, e)
	case *syntax.BinaryExpr:
		value, err = evalBinary(vm, e)

	// Bindings
	case *syntax.LetBindingExpr:
		value, err = evalLetBinding(vm, e)
	case *syntax.DestructAssignmentExpr:
		value, err = evalDestructAssignment(vm, e)

	// Set/Show rules - forbidden outside code blocks
	case *syntax.SetRuleExpr:
		return nil, forbidden("set")
	case *syntax.ShowRuleExpr:
		return nil, forbidden("show")

	// Contextual
	case *syntax.ContextualExpr:
		value, err = evalContextual(vm, e)

	// Control flow
	case *syntax.ConditionalExpr:
		value, err = evalConditional(vm, e)
	case *syntax.WhileLoopExpr:
		value, err = evalWhileLoop(vm, e)
	case *syntax.ForLoopExpr:
		value, err = evalForLoop(vm, e)

	// Module system
	case *syntax.ModuleImportExpr:
		value, err = evalModuleImport(vm, e)
	case *syntax.ModuleIncludeExpr:
		value, err = evalModuleInclude(vm, e)

	// Flow control
	case *syntax.LoopBreakExpr:
		value, err = evalLoopBreak(vm, e)
	case *syntax.LoopContinueExpr:
		value, err = evalLoopContinue(vm, e)
	case *syntax.FuncReturnExpr:
		value, err = evalFuncReturn(vm, e)

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}

	if err != nil {
		return nil, err
	}

	// Add span to the value
	value = foundations.Spanned(value, span)

	// Trace for IDE inspection
	if vm.Inspected != nil && *vm.Inspected == span {
		vm.Trace(value)
	}

	return value, nil
}

// ----------------------------------------------------------------------------
// Identifier Evaluation
// ----------------------------------------------------------------------------

// evalIdent evaluates an identifier expression.
// Matches Rust: impl Eval for ast::Ident
func evalIdent(vm *Vm, e *syntax.IdentExpr) (foundations.Value, error) {
	span := e.ToUntyped().Span()
	binding := vm.Scopes.Get(e.Get())
	if binding == nil {
		return nil, atSpan(fmt.Errorf("unknown variable: %s", e.Get()), span)
	}
	return binding.ReadChecked(vm.Engine, span), nil
}

// ----------------------------------------------------------------------------
// Literal Evaluators
// ----------------------------------------------------------------------------

// evalNone evaluates a none literal.
// Matches Rust: impl Eval for ast::None
func evalNone(_ *Vm, _ *syntax.NoneExpr) (foundations.Value, error) {
	return foundations.None, nil
}

// evalAuto evaluates an auto literal.
// Matches Rust: impl Eval for ast::Auto
func evalAuto(_ *Vm, _ *syntax.AutoExpr) (foundations.Value, error) {
	return foundations.Auto, nil
}

// evalBool evaluates a boolean literal.
// Matches Rust: impl Eval for ast::Bool
func evalBool(_ *Vm, e *syntax.BoolExpr) (foundations.Value, error) {
	return foundations.Bool(e.Get()), nil
}

// evalInt evaluates an integer literal.
// Matches Rust: impl Eval for ast::Int
func evalInt(_ *Vm, e *syntax.IntExpr) (foundations.Value, error) {
	return foundations.Int(e.Get()), nil
}

// evalFloat evaluates a float literal.
// Matches Rust: impl Eval for ast::Float
func evalFloat(_ *Vm, e *syntax.FloatExpr) (foundations.Value, error) {
	return foundations.Float(e.Get()), nil
}

// evalNumeric evaluates a numeric literal with unit.
// Matches Rust: impl Eval for ast::Numeric
func evalNumeric(_ *Vm, e *syntax.NumericExpr) (foundations.Value, error) {
	return foundations.Numeric(e.Value(), e.Unit()), nil
}

// evalStr evaluates a string literal.
// Matches Rust: impl Eval for ast::Str
func evalStr(_ *Vm, e *syntax.StrExpr) (foundations.Value, error) {
	return foundations.Str(e.Get()), nil
}

// ----------------------------------------------------------------------------
// Collection Evaluators
// ----------------------------------------------------------------------------

// evalArray evaluates an array expression.
// Matches Rust: impl Eval for ast::Array
func evalArray(vm *Vm, e *syntax.ArrayExpr) (foundations.Value, error) {
	items := e.Items()
	arr := foundations.NewArrayWithCapacity(len(items))

	for _, item := range items {
		switch i := item.(type) {
		case *syntax.ArrayPosItem:
			value, err := evalExpr(vm, i.Expr())
			if err != nil {
				return nil, err
			}
			arr.Push(value)

		case *syntax.ArraySpreadItem:
			spreadExpr := i.Expr()
			if spreadExpr == nil {
				continue
			}
			value, err := evalExpr(vm, spreadExpr)
			if err != nil {
				return nil, err
			}
			// Handle spread
			switch v := value.(type) {
			case foundations.NoneValue:
				// None spreads as nothing
			case *foundations.Array:
				arr.Extend(v)
			default:
				return nil, fmt.Errorf("cannot spread %s into array", value.Type())
			}
		}
	}

	return arr, nil
}

// evalDict evaluates a dictionary expression.
// Matches Rust: impl Eval for ast::Dict
func evalDict(vm *Vm, e *syntax.DictExpr) (foundations.Value, error) {
	dict := foundations.NewDict()

	for _, item := range e.Items() {
		switch i := item.(type) {
		case *syntax.DictNamedItem:
			ident := i.Name()
			if ident == nil {
				continue
			}
			value, err := evalExpr(vm, i.Expr())
			if err != nil {
				return nil, err
			}
			dict.Set(ident.Get(), value)

		case *syntax.DictKeyedItem:
			keyExpr := i.Key()
			keyValue, err := evalExpr(vm, keyExpr)
			if err != nil {
				return nil, err
			}
			keyStr, ok := keyValue.(foundations.Str)
			if !ok {
				return nil, atSpan(fmt.Errorf("expected string, found %s", keyValue.Type()), keyExpr.ToUntyped().Span())
			}
			value, err := evalExpr(vm, i.Expr())
			if err != nil {
				return nil, err
			}
			dict.Set(string(keyStr), value)

		case *syntax.DictSpreadItem:
			spreadExpr := i.Expr()
			if spreadExpr == nil {
				continue
			}
			value, err := evalExpr(vm, spreadExpr)
			if err != nil {
				return nil, err
			}
			// Handle spread
			switch v := value.(type) {
			case foundations.NoneValue:
				// None spreads as nothing
			case *foundations.Dict:
				dict.Extend(v)
			default:
				return nil, fmt.Errorf("cannot spread %s into dictionary", value.Type())
			}
		}
	}

	return dict, nil
}

// ----------------------------------------------------------------------------
// Block Evaluators
// ----------------------------------------------------------------------------

// evalCodeBlock evaluates a code block.
// Matches Rust: impl Eval for ast::CodeBlock
func evalCodeBlock(vm *Vm, e *syntax.CodeBlockExpr) (foundations.Value, error) {
	vm.Scopes.Enter()
	defer vm.Scopes.Exit()

	body := e.Body()
	if body == nil {
		return foundations.None, nil
	}
	return evalCode(vm, body.Exprs())
}

// evalContentBlock evaluates a content block.
// Matches Rust: impl Eval for ast::ContentBlock
func evalContentBlock(vm *Vm, e *syntax.ContentBlockExpr) (foundations.Value, error) {
	vm.Scopes.Enter()
	defer vm.Scopes.Exit()

	body := e.Body()
	if body == nil {
		return foundations.ContentValue{}, nil
	}
	return evalMarkup(vm, body)
}

// evalParenthesized evaluates a parenthesized expression.
// Matches Rust: impl Eval for ast::Parenthesized
func evalParenthesized(vm *Vm, e *syntax.ParenthesizedExpr) (foundations.Value, error) {
	inner := e.Expr()
	if inner == nil {
		return foundations.None, nil
	}
	return evalExpr(vm, inner)
}

// ----------------------------------------------------------------------------
// Field Access
// ----------------------------------------------------------------------------

// evalFieldAccess evaluates a field access expression.
// Matches Rust: impl Eval for ast::FieldAccess
func evalFieldAccess(vm *Vm, e *syntax.FieldAccessExpr) (foundations.Value, error) {
	target, err := evalExpr(vm, e.Target())
	if err != nil {
		return nil, err
	}

	field := e.Field()
	if field == nil {
		return nil, fmt.Errorf("missing field name")
	}
	fieldName := field.Get()
	fieldSpan := field.ToUntyped().Span()

	// Try normal field access
	value, fieldErr := target.Field(fieldName, vm.Engine, fieldSpan)
	if fieldErr == nil {
		return value, nil
	}

	// Check for get rule field access (accessing element fields from style chain)
	if funcVal, ok := target.(foundations.FuncValue); ok {
		if funcVal.Func != nil {
			if elem := funcVal.Func.ToElement(); elem != nil {
				if fieldID := elem.FieldID(fieldName); fieldID != nil {
					styles, err := vm.Context.Styles.At(fieldSpan)
					if err != nil {
						return nil, err
					}
					if value, err := elem.FieldFromStyles(*fieldID, styles); err == nil {
						return value, nil
					}
				}
			}
		}
	}

	return nil, fieldErr
}

// ----------------------------------------------------------------------------
// Contextual Evaluation
// ----------------------------------------------------------------------------

// evalContextual evaluates a contextual expression.
// Matches Rust: impl Eval for ast::Contextual
func evalContextual(vm *Vm, e *syntax.ContextualExpr) (foundations.Value, error) {
	body := e.Body()
	if body == nil {
		return foundations.ContentValue{}, nil
	}

	// Collect captured variables
	captured := captureScope(vm, e)

	// Define the closure
	closure := &foundations.Closure{
		Node:         foundations.ContextNode{Node: body.ToUntyped()},
		Defaults:     nil,
		Captured:     captured,
		NumPosParams: 0,
	}

	fn := foundations.NewFunc(closure).WithSpan(body.ToUntyped().Span())

	// Wrap in ContextElem and return as content
	contextElem := foundations.NewContextElem(fn)
	return foundations.ContentValue{Content: contextElem.Pack().WithSpan(body.ToUntyped().Span())}, nil
}

// ----------------------------------------------------------------------------
// Warning Helper
// ----------------------------------------------------------------------------

// warnForDiscardedContent emits a warning when content is discarded by an unconditional return.
// Matches Rust: fn warn_for_discarded_content(engine: &mut Engine, event: &FlowEvent, joined: &Value)
func warnForDiscardedContent(engine *foundations.Engine, event FlowEvent, joined foundations.Value) {
	// Check if this is an unconditional return with a value
	ret, ok := event.(ReturnEvent)
	if !ok {
		return
	}
	if ret.Value == nil || ret.Conditional {
		return
	}

	// Check if joined value is content
	content, ok := joined.(foundations.ContentValue)
	if !ok {
		return
	}

	// Build the warning
	warning := foundations.SourceDiagnostic{
		Span:     ret.Span(),
		Severity: foundations.SeverityWarning,
		Message:  "this return unconditionally discards the content before it",
		Hints:    []string{"try omitting the `return` to automatically join all values"},
	}

	// Check if content contains state/counter updates
	if content.Content.ContainsStateOrCounter() {
		warning.Hints = append(warning.Hints,
			"state/counter updates are content that must end up in the document to have an effect")
	}

	engine.Sink.Warn(warning)
}

// ----------------------------------------------------------------------------
// Helper Functions
// ----------------------------------------------------------------------------

// Display converts a value to displayable content.
// Matches Rust: Value::display()
func Display(v foundations.Value) foundations.Content {
	switch val := v.(type) {
	case foundations.ContentValue:
		return val.Content
	case foundations.NoneValue:
		return foundations.Content{}
	case foundations.Str:
		return foundations.TextContent(string(val))
	case foundations.SymbolValue:
		return foundations.TextContent(string(val.Char))
	default:
		// Display other values as their repr
		return foundations.TextContent(fmt.Sprintf("%v", v))
	}
}

