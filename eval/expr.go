package eval

import (
	"fmt"
	"math"
	"strconv"

	"github.com/boergens/gotypst/syntax"
)

// EvalExpr evaluates an expression and returns its value.
//
// This is the main entry point for expression evaluation. It dispatches
// to the appropriate handler based on the expression type.
func EvalExpr(vm *Vm, expr syntax.Expr) (Value, error) {
	if vm.HasFlow() {
		return None, nil
	}

	switch e := expr.(type) {
	// Literals
	case *syntax.NoneExpr:
		return evalNone(vm, e)
	case *syntax.AutoExpr:
		return evalAuto(vm, e)
	case *syntax.BoolExpr:
		return evalBool(vm, e)
	case *syntax.IntExpr:
		return evalInt(vm, e)
	case *syntax.FloatExpr:
		return evalFloat(vm, e)
	case *syntax.NumericExpr:
		return evalNumeric(vm, e)
	case *syntax.StrExpr:
		return evalStr(vm, e)

	// Identifiers
	case *syntax.IdentExpr:
		return evalIdent(vm, e)

	// Collections
	case *syntax.ArrayExpr:
		return evalArray(vm, e)
	case *syntax.DictExpr:
		return evalDict(vm, e)

	// Grouping
	case *syntax.ParenthesizedExpr:
		return evalParenthesized(vm, e)
	case *syntax.CodeBlockExpr:
		return evalCodeBlock(vm, e)
	case *syntax.ContentBlockExpr:
		return evalContentBlock(vm, e)

	// Operations
	case *syntax.UnaryExpr:
		return evalUnary(vm, e)
	case *syntax.BinaryExpr:
		return evalBinary(vm, e)
	case *syntax.FieldAccessExpr:
		return evalFieldAccess(vm, e)
	case *syntax.FuncCallExpr:
		return evalFuncCall(vm, e)

	// Closures and bindings
	case *syntax.ClosureExpr:
		return evalClosure(vm, e)
	case *syntax.LetBindingExpr:
		return evalLetBinding(vm, e)
	case *syntax.DestructAssignmentExpr:
		return evalDestructAssignment(vm, e)

	// Control flow (deferred to control_flow.go)
	case *syntax.ConditionalExpr:
		return evalConditional(vm, e)
	case *syntax.WhileLoopExpr:
		return evalWhileLoop(vm, e)
	case *syntax.ForLoopExpr:
		return evalForLoop(vm, e)
	case *syntax.LoopBreakExpr:
		return evalLoopBreak(vm, e)
	case *syntax.LoopContinueExpr:
		return evalLoopContinue(vm, e)
	case *syntax.FuncReturnExpr:
		return evalFuncReturn(vm, e)

	// Contextual
	case *syntax.ContextualExpr:
		return evalContextual(vm, e)

	// Set/Show rules (deferred to rules.go)
	case *syntax.SetRuleExpr:
		return evalSetRule(vm, e)
	case *syntax.ShowRuleExpr:
		return evalShowRule(vm, e)

	// Module imports (deferred to import.go)
	case *syntax.ModuleImportExpr:
		return evalModuleImport(vm, e)
	case *syntax.ModuleIncludeExpr:
		return evalModuleInclude(vm, e)

	// Markup expressions (evaluate to Content)
	case *syntax.TextExpr:
		return evalText(vm, e)
	case *syntax.SpaceExpr:
		return evalSpace(vm, e)
	case *syntax.LinebreakExpr:
		return evalLinebreak(vm, e)
	case *syntax.ParbreakExpr:
		return evalParbreak(vm, e)
	case *syntax.StrongExpr:
		return evalStrong(vm, e)
	case *syntax.EmphExpr:
		return evalEmph(vm, e)
	case *syntax.RawExpr:
		return evalRaw(vm, e)
	case *syntax.LinkExpr:
		return evalLink(vm, e)
	case *syntax.LabelExpr:
		return evalLabel(vm, e)
	case *syntax.RefExpr:
		return evalRef(vm, e)
	case *syntax.HeadingExpr:
		return evalHeading(vm, e)
	case *syntax.ListItemExpr:
		return evalListItem(vm, e)
	case *syntax.EnumItemExpr:
		return evalEnumItem(vm, e)
	case *syntax.TermItemExpr:
		return evalTermItem(vm, e)
	case *syntax.EscapeExpr:
		return evalEscape(vm, e)
	case *syntax.ShorthandExpr:
		return evalShorthand(vm, e)
	case *syntax.SmartQuoteExpr:
		return evalSmartQuote(vm, e)

	// Math expressions (deferred to math.go)
	case *syntax.EquationExpr:
		return evalEquation(vm, e)
	case *syntax.MathTextExpr:
		return evalMathText(vm, e)
	case *syntax.MathIdentExpr:
		return evalMathIdent(vm, e)
	case *syntax.MathShorthandExpr:
		return evalMathShorthand(vm, e)
	case *syntax.MathAlignPointExpr:
		return evalMathAlignPoint(vm, e)
	case *syntax.MathDelimitedExpr:
		return evalMathDelimited(vm, e)
	case *syntax.MathAttachExpr:
		return evalMathAttach(vm, e)
	case *syntax.MathPrimesExpr:
		return evalMathPrimes(vm, e)
	case *syntax.MathFracExpr:
		return evalMathFrac(vm, e)
	case *syntax.MathRootExpr:
		return evalMathRoot(vm, e)

	default:
		return nil, &UnsupportedExprError{Expr: expr}
	}
}

// ----------------------------------------------------------------------------
// Literal Evaluators
// ----------------------------------------------------------------------------

func evalNone(_ *Vm, _ *syntax.NoneExpr) (Value, error) {
	return None, nil
}

func evalAuto(_ *Vm, _ *syntax.AutoExpr) (Value, error) {
	return Auto, nil
}

func evalBool(_ *Vm, e *syntax.BoolExpr) (Value, error) {
	return Bool(e.Get()), nil
}

func evalInt(_ *Vm, e *syntax.IntExpr) (Value, error) {
	return Int(e.Get()), nil
}

func evalFloat(_ *Vm, e *syntax.FloatExpr) (Value, error) {
	return Float(e.Get()), nil
}

func evalNumeric(_ *Vm, e *syntax.NumericExpr) (Value, error) {
	value := e.Value()
	unit := e.Unit()

	switch unit {
	// Length units
	case syntax.UnitPt:
		return LengthValue{Length: Length{Points: value}}, nil
	case syntax.UnitMm:
		return LengthValue{Length: Length{Points: value * 2.83465}}, nil
	case syntax.UnitCm:
		return LengthValue{Length: Length{Points: value * 28.3465}}, nil
	case syntax.UnitIn:
		return LengthValue{Length: Length{Points: value * 72}}, nil
	case syntax.UnitEm:
		// Em is relative to font size, store as a special relative length
		return RelativeValue{Relative: Relative{Abs: Length{}, Rel: Ratio{Value: value}}}, nil

	// Angle units
	case syntax.UnitDeg:
		return AngleValue{Angle: Angle{Radians: value * math.Pi / 180}}, nil
	case syntax.UnitRad:
		return AngleValue{Angle: Angle{Radians: value}}, nil

	// Ratio (percentage)
	case syntax.UnitPercent:
		return RatioValue{Ratio: Ratio{Value: value / 100}}, nil

	// Fraction (fr units)
	case syntax.UnitFr:
		return FractionValue{Fraction: Fraction{Value: value}}, nil

	default:
		// Unknown unit, treat as plain number
		return Float(value), nil
	}
}

func evalStr(_ *Vm, e *syntax.StrExpr) (Value, error) {
	// Get the string content (with escape processing)
	text := e.Get()
	return Str(processStringEscapes(text)), nil
}

// processStringEscapes handles escape sequences in strings.
func processStringEscapes(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result = append(result, '\n')
				i += 2
			case 'r':
				result = append(result, '\r')
				i += 2
			case 't':
				result = append(result, '\t')
				i += 2
			case '\\':
				result = append(result, '\\')
				i += 2
			case '"':
				result = append(result, '"')
				i += 2
			case 'u':
				// Unicode escape: \u{XXXX}
				if i+3 < len(s) && s[i+2] == '{' {
					end := i + 3
					for end < len(s) && s[end] != '}' {
						end++
					}
					if end < len(s) {
						hex := s[i+3 : end]
						if codepoint, err := strconv.ParseInt(hex, 16, 32); err == nil {
							result = append(result, string(rune(codepoint))...)
						}
						i = end + 1
						continue
					}
				}
				result = append(result, s[i])
				i++
			default:
				result = append(result, s[i])
				i++
			}
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}

// ----------------------------------------------------------------------------
// Identifier Evaluation
// ----------------------------------------------------------------------------

func evalIdent(vm *Vm, e *syntax.IdentExpr) (Value, error) {
	name := e.Get()
	binding := vm.Get(name)
	if binding == nil {
		return nil, &UndefinedVariableError{Name: name, Span: e.ToUntyped().Span()}
	}

	value, err := binding.ReadChecked(e.ToUntyped().Span())
	if err != nil {
		return nil, err
	}

	vm.Trace(value, e.ToUntyped().Span())
	return value, nil
}

// ----------------------------------------------------------------------------
// Collection Evaluators
// ----------------------------------------------------------------------------

func evalArray(vm *Vm, e *syntax.ArrayExpr) (Value, error) {
	var result ArrayValue
	items := e.Items()

	for _, item := range items {
		if vm.HasFlow() {
			return None, nil
		}

		switch i := item.(type) {
		case *syntax.ArrayPosItem:
			value, err := EvalExpr(vm, i.Expr())
			if err != nil {
				return nil, err
			}
			result = append(result, value)

		case *syntax.ArraySpreadItem:
			spreadExpr := i.Expr()
			if spreadExpr == nil {
				continue
			}
			value, err := EvalExpr(vm, spreadExpr)
			if err != nil {
				return nil, err
			}
			// Spread the value into the array
			if arr, ok := value.(ArrayValue); ok {
				result = append(result, arr...)
			} else {
				return nil, &TypeError{
					Expected: TypeArray,
					Got:      value.Type(),
					Span:     spreadExpr.ToUntyped().Span(),
				}
			}
		}
	}

	return result, nil
}

func evalDict(vm *Vm, e *syntax.DictExpr) (Value, error) {
	result := NewDict()
	items := e.Items()

	for _, item := range items {
		if vm.HasFlow() {
			return None, nil
		}

		switch i := item.(type) {
		case *syntax.DictNamedItem:
			ident := i.Name()
			if ident == nil {
				continue
			}
			key := ident.Get()
			value, err := EvalExpr(vm, i.Expr())
			if err != nil {
				return nil, err
			}
			result.Set(key, value)

		case *syntax.DictKeyedItem:
			keyExpr := i.Key()
			keyValue, err := EvalExpr(vm, keyExpr)
			if err != nil {
				return nil, err
			}
			keyStr, ok := AsStr(keyValue)
			if !ok {
				return nil, &TypeError{
					Expected: TypeStr,
					Got:      keyValue.Type(),
					Span:     keyExpr.ToUntyped().Span(),
				}
			}
			value, err := EvalExpr(vm, i.Expr())
			if err != nil {
				return nil, err
			}
			result.Set(keyStr, value)

		case *syntax.DictSpreadItem:
			spreadExpr := i.Expr()
			if spreadExpr == nil {
				continue
			}
			value, err := EvalExpr(vm, spreadExpr)
			if err != nil {
				return nil, err
			}
			// Spread the dictionary
			if dict, ok := AsDict(value); ok {
				for _, key := range dict.Keys() {
					val, _ := dict.Get(key)
					result.Set(key, val)
				}
			} else {
				return nil, &TypeError{
					Expected: TypeDict,
					Got:      value.Type(),
					Span:     spreadExpr.ToUntyped().Span(),
				}
			}
		}
	}

	return result, nil
}

// ----------------------------------------------------------------------------
// Grouping Evaluators
// ----------------------------------------------------------------------------

func evalParenthesized(vm *Vm, e *syntax.ParenthesizedExpr) (Value, error) {
	inner := e.Expr()
	if inner == nil {
		return None, nil
	}
	return EvalExpr(vm, inner)
}

func evalCodeBlock(vm *Vm, e *syntax.CodeBlockExpr) (Value, error) {
	body := e.Body()
	if body == nil {
		return None, nil
	}

	vm.EnterScope()
	defer vm.ExitScope()

	return evalCode(vm, body)
}

func evalContentBlock(vm *Vm, e *syntax.ContentBlockExpr) (Value, error) {
	body := e.Body()
	if body == nil {
		return ContentValue{Content: Content{}}, nil
	}

	vm.EnterScope()
	defer vm.ExitScope()

	return evalMarkup(vm, body)
}

// evalCode evaluates a code block and returns its result.
func evalCode(vm *Vm, code *syntax.CodeNode) (Value, error) {
	var result Value = None
	exprs := code.Exprs()

	for _, expr := range exprs {
		if vm.HasFlow() {
			break
		}

		value, err := EvalExpr(vm, expr)
		if err != nil {
			return nil, err
		}

		// Join values
		result, err = joinValues(result, value)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// evalMarkup evaluates markup content and returns Content.
func evalMarkup(vm *Vm, markup *syntax.MarkupNode) (Value, error) {
	var content Content
	exprs := markup.Exprs()

	for _, expr := range exprs {
		if vm.HasFlow() {
			break
		}

		value, err := EvalExpr(vm, expr)
		if err != nil {
			return nil, err
		}

		// Append to content
		content = appendToContent(content, value)
	}

	return ContentValue{Content: content}, nil
}

// joinValues joins two values together.
func joinValues(a, b Value) (Value, error) {
	if IsNone(a) {
		return b, nil
	}
	if IsNone(b) {
		return a, nil
	}

	// Join content values
	if ac, ok := a.(ContentValue); ok {
		return ContentValue{Content: appendToContent(ac.Content, b)}, nil
	}

	// Return the second value (last expression wins in code blocks)
	return b, nil
}

// appendToContent appends a value to content.
func appendToContent(content Content, value Value) Content {
	switch v := value.(type) {
	case ContentValue:
		content.Elements = append(content.Elements, v.Content.Elements...)
	case StrValue:
		content.Elements = append(content.Elements, &TextElement{Text: string(v)})
	case NoneValue:
		// None doesn't contribute to content
	default:
		// Display other values as text
		content.Elements = append(content.Elements, &TextElement{Text: fmt.Sprintf("%v", value)})
	}
	return content
}

// TextElement is a simple text content element.
type TextElement struct {
	Text string
}

func (*TextElement) IsContentElement() {}

// ----------------------------------------------------------------------------
// Operator Evaluators
// ----------------------------------------------------------------------------

func evalUnary(vm *Vm, e *syntax.UnaryExpr) (Value, error) {
	operand, err := EvalExpr(vm, e.Expr())
	if err != nil {
		return nil, err
	}

	op := e.Op()
	span := e.ToUntyped().Span()

	switch op {
	case syntax.UnOpPos:
		return applyPos(operand, span)
	case syntax.UnOpNeg:
		return applyNeg(operand, span)
	case syntax.UnOpNot:
		return applyNot(operand, span)
	default:
		return nil, &UnsupportedOperatorError{Op: op.String(), Span: span}
	}
}

func applyPos(v Value, span syntax.Span) (Value, error) {
	switch val := v.(type) {
	case IntValue:
		return val, nil
	case FloatValue:
		return val, nil
	case LengthValue:
		return val, nil
	case AngleValue:
		return val, nil
	case RatioValue:
		return val, nil
	case FractionValue:
		return val, nil
	default:
		return nil, &TypeError{Expected: TypeInt, Got: v.Type(), Span: span}
	}
}

func applyNeg(v Value, span syntax.Span) (Value, error) {
	switch val := v.(type) {
	case IntValue:
		return Int(-int64(val)), nil
	case FloatValue:
		return Float(-float64(val)), nil
	case LengthValue:
		return LengthValue{Length: Length{Points: -val.Length.Points}}, nil
	case AngleValue:
		return AngleValue{Angle: Angle{Radians: -val.Angle.Radians}}, nil
	case RatioValue:
		return RatioValue{Ratio: Ratio{Value: -val.Ratio.Value}}, nil
	case FractionValue:
		return FractionValue{Fraction: Fraction{Value: -val.Fraction.Value}}, nil
	default:
		return nil, &TypeError{Expected: TypeInt, Got: v.Type(), Span: span}
	}
}

func applyNot(v Value, span syntax.Span) (Value, error) {
	b, ok := AsBool(v)
	if !ok {
		return nil, &TypeError{Expected: TypeBool, Got: v.Type(), Span: span}
	}
	return Bool(!b), nil
}

func evalBinary(vm *Vm, e *syntax.BinaryExpr) (Value, error) {
	op := e.Op()
	span := e.ToUntyped().Span()

	// Handle short-circuit operators
	if op == syntax.BinOpAnd || op == syntax.BinOpOr {
		return evalShortCircuit(vm, e)
	}

	// Handle assignment operators
	if op.IsAssignment() {
		return evalAssignment(vm, e)
	}

	// Evaluate both operands
	lhs, err := EvalExpr(vm, e.Lhs())
	if err != nil {
		return nil, err
	}

	rhs, err := EvalExpr(vm, e.Rhs())
	if err != nil {
		return nil, err
	}

	// Apply the operator
	return applyBinaryOp(op, lhs, rhs, span)
}

func evalShortCircuit(vm *Vm, e *syntax.BinaryExpr) (Value, error) {
	op := e.Op()

	lhs, err := EvalExpr(vm, e.Lhs())
	if err != nil {
		return nil, err
	}

	lhsBool, ok := AsBool(lhs)
	if !ok {
		return nil, &TypeError{Expected: TypeBool, Got: lhs.Type(), Span: e.Lhs().ToUntyped().Span()}
	}

	// Short-circuit evaluation
	if op == syntax.BinOpAnd && !lhsBool {
		return False, nil
	}
	if op == syntax.BinOpOr && lhsBool {
		return True, nil
	}

	// Evaluate right side
	rhs, err := EvalExpr(vm, e.Rhs())
	if err != nil {
		return nil, err
	}

	rhsBool, ok := AsBool(rhs)
	if !ok {
		return nil, &TypeError{Expected: TypeBool, Got: rhs.Type(), Span: e.Rhs().ToUntyped().Span()}
	}

	return Bool(rhsBool), nil
}

func evalAssignment(vm *Vm, e *syntax.BinaryExpr) (Value, error) {
	op := e.Op()
	lhsExpr := e.Lhs()
	rhsExpr := e.Rhs()

	// Evaluate the right-hand side
	rhs, err := EvalExpr(vm, rhsExpr)
	if err != nil {
		return nil, err
	}

	// For compound assignments, we need to compute the new value
	if op != syntax.BinOpAssign {
		lhs, err := EvalExpr(vm, lhsExpr)
		if err != nil {
			return nil, err
		}

		var opResult Value
		span := e.ToUntyped().Span()

		switch op {
		case syntax.BinOpAddAssign:
			opResult, err = Add(lhs, rhs, span)
		case syntax.BinOpSubAssign:
			opResult, err = Sub(lhs, rhs, span)
		case syntax.BinOpMulAssign:
			opResult, err = Mul(lhs, rhs, span)
		case syntax.BinOpDivAssign:
			opResult, err = Div(lhs, rhs, span)
		default:
			return nil, &UnsupportedOperatorError{Op: op.String(), Span: span}
		}
		if err != nil {
			return nil, err
		}
		rhs = opResult
	}

	// Perform the assignment
	return assignToExpr(vm, lhsExpr, rhs)
}

// assignToExpr assigns a value to an lvalue expression.
func assignToExpr(vm *Vm, expr syntax.Expr, value Value) (Value, error) {
	switch e := expr.(type) {
	case *syntax.IdentExpr:
		name := e.Get()
		binding := vm.GetMut(name)
		if binding == nil {
			return nil, &UndefinedVariableError{Name: name, Span: e.ToUntyped().Span()}
		}
		if err := binding.Write(value); err != nil {
			return nil, err
		}
		return None, nil

	case *syntax.FieldAccessExpr:
		// Get the target object
		target, err := EvalExpr(vm, e.Target())
		if err != nil {
			return nil, err
		}

		field := e.Field()
		if field == nil {
			return nil, &TypeError{Expected: TypeDict, Got: target.Type(), Span: e.ToUntyped().Span()}
		}
		fieldName := field.Get()

		// For dictionary field access
		if dict, ok := AsDict(target); ok {
			dict.Set(fieldName, value)
			return None, nil
		}

		return nil, &TypeError{Expected: TypeDict, Got: target.Type(), Span: e.ToUntyped().Span()}

	default:
		return nil, &InvalidAssignmentTargetError{Span: expr.ToUntyped().Span()}
	}
}

func applyBinaryOp(op syntax.BinOp, lhs, rhs Value, span syntax.Span) (Value, error) {
	switch op {
	case syntax.BinOpAdd:
		return Add(lhs, rhs, span)
	case syntax.BinOpSub:
		return Sub(lhs, rhs, span)
	case syntax.BinOpMul:
		return Mul(lhs, rhs, span)
	case syntax.BinOpDiv:
		return Div(lhs, rhs, span)
	case syntax.BinOpEq:
		return Bool(Equal(lhs, rhs)), nil
	case syntax.BinOpNeq:
		return Bool(!Equal(lhs, rhs)), nil
	case syntax.BinOpLt:
		return Compare(lhs, rhs, func(c int) bool { return c < 0 }, span)
	case syntax.BinOpLeq:
		return Compare(lhs, rhs, func(c int) bool { return c <= 0 }, span)
	case syntax.BinOpGt:
		return Compare(lhs, rhs, func(c int) bool { return c > 0 }, span)
	case syntax.BinOpGeq:
		return Compare(lhs, rhs, func(c int) bool { return c >= 0 }, span)
	case syntax.BinOpIn:
		return Contains(rhs, lhs, span)
	default:
		return nil, &UnsupportedOperatorError{Op: op.String(), Span: span}
	}
}

// ----------------------------------------------------------------------------
// Field Access and Function Call
// ----------------------------------------------------------------------------

func evalFieldAccess(vm *Vm, e *syntax.FieldAccessExpr) (Value, error) {
	target, err := EvalExpr(vm, e.Target())
	if err != nil {
		return nil, err
	}

	field := e.Field()
	if field == nil {
		return nil, &TypeError{Expected: TypeDict, Got: target.Type(), Span: e.ToUntyped().Span()}
	}
	fieldName := field.Get()

	// Handle different target types
	switch t := target.(type) {
	case *DictValue:
		if val, ok := t.Get(fieldName); ok {
			return val, nil
		}
		return nil, &FieldNotFoundError{Field: fieldName, Type: target.Type(), Span: e.ToUntyped().Span()}

	case DictValue:
		if val, ok := t.Get(fieldName); ok {
			return val, nil
		}
		return nil, &FieldNotFoundError{Field: fieldName, Type: target.Type(), Span: e.ToUntyped().Span()}

	case ModuleValue:
		if t.Module != nil && t.Module.Scope != nil {
			if binding := t.Module.Scope.Get(fieldName); binding != nil {
				return binding.Value, nil
			}
		}
		return nil, &FieldNotFoundError{Field: fieldName, Type: target.Type(), Span: e.ToUntyped().Span()}

	case FuncValue:
		// Functions can have fields (their scope)
		if t.Func != nil && t.Func.Repr != nil {
			if nf, ok := t.Func.Repr.(NativeFunc); ok && nf.Info != nil {
				// Native functions might have metadata fields
			}
		}
		return nil, &FieldNotFoundError{Field: fieldName, Type: target.Type(), Span: e.ToUntyped().Span()}

	case TypeValue:
		// Type values have static methods (e.g., str.from-unicode)
		method := GetTypeMethod(t.Inner, fieldName, e.ToUntyped().Span())
		if method != nil {
			return method, nil
		}
		return nil, &FieldNotFoundError{Field: fieldName, Type: target.Type(), Span: e.ToUntyped().Span()}

	default:
		// Check for built-in methods
		method := getBuiltinMethod(target, fieldName, e.ToUntyped().Span())
		if method != nil {
			return method, nil
		}
		return nil, &FieldNotFoundError{Field: fieldName, Type: target.Type(), Span: e.ToUntyped().Span()}
	}
}

// getBuiltinMethod returns a built-in method for a value, or nil if not found.
func getBuiltinMethod(target Value, name string, span syntax.Span) Value {
	switch t := target.(type) {
	case StrValue:
		return GetStrMethod(t, name, span)
	case ArrayValue:
		return GetArrayMethod(t, name, span)
	}
	return nil
}

func evalFuncCall(vm *Vm, e *syntax.FuncCallExpr) (Value, error) {
	// Check call depth
	if err := vm.CheckCallDepth(); err != nil {
		return nil, err
	}

	// Evaluate the callee
	calleeExpr := e.Callee()
	callee, err := EvalExpr(vm, calleeExpr)
	if err != nil {
		return nil, err
	}

	// Build arguments
	args, err := evalArgs(vm, e.Args())
	if err != nil {
		return nil, err
	}

	// Call the function
	return callFunc(vm, callee, args, e.ToUntyped().Span())
}

// evalArgs evaluates function arguments.
func evalArgs(vm *Vm, argsNode *syntax.ArgsNode) (*Args, error) {
	if argsNode == nil {
		return &Args{}, nil
	}

	args := &Args{Span: argsNode.ToUntyped().Span()}
	items := argsNode.Items()

	for _, item := range items {
		if vm.HasFlow() {
			break
		}

		switch a := item.(type) {
		case *syntax.PosArg:
			value, err := EvalExpr(vm, a.Expr())
			if err != nil {
				return nil, err
			}
			args.Items = append(args.Items, Arg{
				Span:  argsNode.ToUntyped().Span(),
				Value: syntax.Spanned[Value]{V: value, Span: argsNode.ToUntyped().Span()},
			})

		case *syntax.NamedArg:
			ident := a.Name()
			if ident == nil {
				continue
			}
			name := ident.Get()
			value, err := EvalExpr(vm, a.Expr())
			if err != nil {
				return nil, err
			}
			args.Items = append(args.Items, Arg{
				Span:  argsNode.ToUntyped().Span(),
				Name:  &name,
				Value: syntax.Spanned[Value]{V: value, Span: argsNode.ToUntyped().Span()},
			})

		case *syntax.SpreadArg:
			value, err := EvalExpr(vm, a.Expr())
			if err != nil {
				return nil, err
			}
			// Spread into arguments
			switch v := value.(type) {
			case ArrayValue:
				for _, elem := range v {
					args.Items = append(args.Items, Arg{
						Span:  argsNode.ToUntyped().Span(),
						Value: syntax.Spanned[Value]{V: elem, Span: argsNode.ToUntyped().Span()},
					})
				}
			case *ArgsValue:
				if v.Args != nil {
					args.Items = append(args.Items, v.Args.Items...)
				}
			default:
				return nil, &TypeError{Expected: TypeArray, Got: value.Type(), Span: argsNode.ToUntyped().Span()}
			}
		}
	}

	return args, nil
}

// callFunc calls a function with the given arguments.
func callFunc(vm *Vm, callee Value, args *Args, span syntax.Span) (Value, error) {
	fn, ok := AsFunc(callee)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: callee.Type(), Span: span}
	}

	vm.EnterCall()
	defer vm.ExitCall()

	switch repr := fn.Repr.(type) {
	case NativeFunc:
		return repr.Func(vm, args)

	case ClosureFunc:
		return evalClosureCall(vm, fn, repr.Closure, args)

	case WithFunc:
		// Merge pre-applied args with new args
		merged := mergeArgs(repr.Args, args)
		return callFunc(vm, FuncValue{Func: repr.Func}, merged, span)

	default:
		return nil, &TypeError{Expected: TypeFunc, Got: callee.Type(), Span: span}
	}
}

// mergeArgs merges pre-applied arguments with new arguments.
func mergeArgs(pre, new *Args) *Args {
	if pre == nil {
		return new
	}
	if new == nil {
		return pre
	}
	result := &Args{Span: new.Span}
	result.Items = append(result.Items, pre.Items...)
	result.Items = append(result.Items, new.Items...)
	return result
}

// evalClosureCall evaluates a closure call.
func evalClosureCall(vm *Vm, fn *Func, closure *Closure, args *Args) (Value, error) {
	if closure == nil {
		return nil, &TypeError{Expected: TypeFunc, Got: TypeNone, Span: fn.Span}
	}

	// Create new scopes with captured variables
	scopes := NewScopes(nil)
	if closure.Captured != nil {
		scopes.SetTop(closure.Captured.Clone())
	}

	// Create new VM for closure evaluation
	closureVm := NewVm(vm.Engine, vm.Context, scopes, fn.Span)

	// Bind function name for recursion
	if fn.Name != nil {
		closureVm.Define(*fn.Name, FuncValue{Func: fn})
	}

	// Bind parameters from arguments
	// This is simplified - full implementation would handle all param types
	if closure.Node != nil {
		if closureAst, ok := closure.Node.(ClosureAstNode); ok {
			closureExpr := syntax.ClosureExprFromNode(closureAst.Node)
			if closureExpr != nil {
				if err := bindParams(closureVm, closureExpr.Params(), args, closure.Defaults); err != nil {
					return nil, err
				}

				// Evaluate body
				body := closureExpr.Body()
				if body != nil {
					result, err := EvalExpr(closureVm, body)
					if err != nil {
						return nil, err
					}

					// Handle return flow
					if closureVm.HasFlow() {
						if ret, ok := closureVm.Flow.(ReturnEvent); ok {
							if ret.Value != nil {
								return ret.Value, nil
							}
							return result, nil
						}
					}
					return result, nil
				}
			}
		}
	}

	return None, nil
}

// bindParams binds function parameters from arguments.
func bindParams(vm *Vm, params *syntax.ParamsNode, args *Args, defaults []Value) error {
	if params == nil {
		return nil
	}

	paramList := params.Children()
	argIndex := 0
	defaultIndex := 0

	for _, param := range paramList {
		switch p := param.(type) {
		case *syntax.PosParam:
			// Positional parameter
			ident := p.Name()
			if ident == nil {
				continue
			}
			name := ident.Get()

			if argIndex < len(args.Items) && args.Items[argIndex].Name == nil {
				vm.Define(name, args.Items[argIndex].Value.V)
				argIndex++
			} else {
				// No argument provided
				return &MissingArgumentError{What: name, Span: ident.ToUntyped().Span()}
			}

		case *syntax.NamedParam:
			// Named parameter with default
			ident := p.Name()
			if ident == nil {
				continue
			}
			name := ident.Get()

			// Look for named argument
			found := false
			for _, arg := range args.Items {
				if arg.Name != nil && *arg.Name == name {
					vm.Define(name, arg.Value.V)
					found = true
					break
				}
			}

			if !found {
				// Use default value
				if defaultIndex < len(defaults) {
					vm.Define(name, defaults[defaultIndex])
				} else {
					vm.Define(name, None)
				}
			}
			defaultIndex++

		case *syntax.SinkParam:
			// Rest parameter - collect remaining positional args
			ident := p.Name()
			if ident == nil {
				continue
			}
			name := ident.Get()

			var rest ArrayValue
			for argIndex < len(args.Items) {
				if args.Items[argIndex].Name == nil {
					rest = append(rest, args.Items[argIndex].Value.V)
				}
				argIndex++
			}
			vm.Define(name, rest)
		}
	}

	return nil
}

// ----------------------------------------------------------------------------
// Closure Evaluation
// ----------------------------------------------------------------------------

func evalClosure(vm *Vm, e *syntax.ClosureExpr) (Value, error) {
	// Evaluate default values for named parameters
	var defaults []Value
	params := e.Params()
	if params != nil {
		for _, param := range params.Children() {
			if np, ok := param.(*syntax.NamedParam); ok {
				if defExpr := np.Default(); defExpr != nil {
					defVal, err := EvalExpr(vm, defExpr)
					if err != nil {
						return nil, err
					}
					defaults = append(defaults, defVal)
				}
			}
		}
	}

	// Capture variables
	captured := captureVariables(vm, e)

	// Count positional parameters
	numPosParams := countPosParams(e)

	// Get optional name
	var name *string
	if nameExpr := e.Name(); nameExpr != nil {
		n := nameExpr.Get()
		name = &n
	}

	// Create closure
	closure := &Closure{
		Node:         ClosureAstNode{Node: e.ToUntyped()},
		Defaults:     defaults,
		Captured:     captured,
		NumPosParams: numPosParams,
	}

	fn := &Func{
		Name: name,
		Span: e.ToUntyped().Span(),
		Repr: ClosureFunc{Closure: closure},
	}

	return FuncValue{Func: fn}, nil
}

// captureVariables captures variables referenced by a closure.
func captureVariables(vm *Vm, e *syntax.ClosureExpr) *Scope {
	// For now, capture all accessible variables
	// A proper implementation would use static analysis to capture only referenced variables
	return vm.Scopes.FlattenToScope()
}

// countPosParams counts positional parameters in a closure.
func countPosParams(e *syntax.ClosureExpr) int {
	params := e.Params()
	if params == nil {
		return 0
	}

	count := 0
	for _, param := range params.Children() {
		if _, ok := param.(*syntax.PosParam); ok {
			count++
		}
	}
	return count
}

// ----------------------------------------------------------------------------
// Let Binding and Destructuring
// ----------------------------------------------------------------------------

func evalLetBinding(vm *Vm, e *syntax.LetBindingExpr) (Value, error) {
	if e.BindingKind() == syntax.LetBindingClosure {
		// Closure binding: let f(x) = ...
		init := e.Init()
		if init == nil {
			return None, nil
		}

		value, err := EvalExpr(vm, init)
		if err != nil {
			return nil, err
		}
		if vm.HasFlow() {
			return None, nil
		}

		// For closure bindings, the pattern contains the function name
		if closure, ok := init.(*syntax.ClosureExpr); ok {
			if name := closure.Name(); name != nil {
				vm.Define(name.Get(), value)
			}
		}
		return None, nil
	}

	// Plain binding: let x = ...
	var value Value = None
	if init := e.Init(); init != nil {
		var err error
		value, err = EvalExpr(vm, init)
		if err != nil {
			return nil, err
		}
		if vm.HasFlow() {
			return None, nil
		}
	}

	// Destructure the pattern using the complete binding.go implementation
	pattern := e.Pattern()
	if err := Destructure(vm, pattern, value); err != nil {
		return nil, err
	}
	return None, nil
}

func evalDestructAssignment(vm *Vm, e *syntax.DestructAssignmentExpr) (Value, error) {
	// Evaluate the value
	valueExpr := e.Value()
	if valueExpr == nil {
		return None, nil
	}

	value, err := EvalExpr(vm, valueExpr)
	if err != nil {
		return nil, err
	}

	// Destructure into the pattern (reassignment) using the complete binding.go implementation
	destructNode := e.Pattern()
	if destructNode == nil {
		return None, nil
	}

	// Convert DestructuringNode to DestructuringPattern for DestructureAssign
	pattern := syntax.DestructuringPatternFromNode(destructNode.ToUntyped())
	if err := DestructureAssign(vm, pattern, value); err != nil {
		return nil, err
	}
	return None, nil
}

// ----------------------------------------------------------------------------
// Control Flow Stubs
// ----------------------------------------------------------------------------

func evalConditional(vm *Vm, e *syntax.ConditionalExpr) (Value, error) {
	// Evaluate condition
	condition := e.Condition()
	cond, err := EvalExpr(vm, condition)
	if err != nil {
		return nil, err
	}

	condBool, ok := AsBool(cond)
	if !ok {
		return nil, &TypeError{Expected: TypeBool, Got: cond.Type(), Span: condition.ToUntyped().Span()}
	}

	var output Value
	if condBool {
		if body := e.IfBody(); body != nil {
			output, err = EvalExpr(vm, body)
			if err != nil {
				return nil, err
			}
		} else {
			output = None
		}
	} else {
		if body := e.ElseBody(); body != nil {
			output, err = EvalExpr(vm, body)
			if err != nil {
				return nil, err
			}
		} else {
			output = None
		}
	}

	// Mark the return as conditional (it occurred inside an if/else branch)
	MarkReturnAsConditional(vm)

	return output, nil
}

func evalWhileLoop(vm *Vm, e *syntax.WhileLoopExpr) (Value, error) {
	// Save any existing flow event to restore after the loop
	savedFlow := vm.TakeFlow()

	var output Value = None
	var i int

	condition := e.Condition()
	body := e.Body()

	for {
		// Check condition
		cond, err := EvalExpr(vm, condition)
		if err != nil {
			return nil, err
		}

		condBool, ok := AsBool(cond)
		if !ok {
			return nil, &TypeError{Expected: TypeBool, Got: cond.Type(), Span: condition.ToUntyped().Span()}
		}

		if !condBool {
			break
		}

		// Check for infinite loop on first iteration
		if i == 0 && body != nil {
			condNode := condition.ToUntyped()
			bodyNode := body.ToUntyped()
			if isInvariant(condNode) && !canDiverge(bodyNode) {
				return nil, &InfiniteLoopError{
					Span:    condNode.Span(),
					Message: "condition is always true",
				}
			}
		} else if i >= MaxIterations {
			return nil, &InfiniteLoopError{
				Span:    e.ToUntyped().Span(),
				Message: "loop seems to be infinite",
			}
		}

		// Execute body
		if body != nil {
			value, err := EvalExpr(vm, body)
			if err != nil {
				return nil, err
			}
			output, _ = joinValues(output, value)
		}

		// Handle flow events
		switch flow := vm.Flow.(type) {
		case BreakEvent:
			vm.ClearFlow()
			goto done
		case ContinueEvent:
			vm.ClearFlow()
		case ReturnEvent:
			_ = flow // Return propagates up, exit loop
			goto done
		}

		i++
	}

done:
	// Restore saved flow if there was one
	if savedFlow != nil {
		vm.SetFlow(savedFlow)
	}

	// Mark the return as conditional (it occurred inside a loop)
	MarkReturnAsConditional(vm)

	return output, nil
}

func evalForLoop(vm *Vm, e *syntax.ForLoopExpr) (Value, error) {
	// Save any existing flow event to restore after the loop
	savedFlow := vm.TakeFlow()

	var output Value = None

	// Evaluate iterable
	iterExpr := e.Iter()
	if iterExpr == nil {
		return None, nil
	}

	iterable, err := EvalExpr(vm, iterExpr)
	if err != nil {
		return nil, err
	}

	pattern := e.Pattern()
	iterableType := iterable.Type()

	// Helper to run the loop body and handle flow
	runBody := func(vm *Vm, body syntax.Expr) (shouldBreak bool, err error) {
		if body == nil {
			return false, nil
		}
		value, err := EvalExpr(vm, body)
		if err != nil {
			return false, err
		}
		output, _ = joinValues(output, value)

		// Handle flow events
		switch flow := vm.Flow.(type) {
		case BreakEvent:
			vm.ClearFlow()
			return true, nil
		case ContinueEvent:
			vm.ClearFlow()
			return false, nil
		case ReturnEvent:
			_ = flow // Return propagates up
			return true, nil
		}
		return false, nil
	}

	// Enter scope once for the entire loop
	vm.EnterScope()

	switch v := iterable.(type) {
	case ArrayValue:
		// Iterate over values of array
		for _, elem := range v {
			if err := Destructure(vm, pattern, elem); err != nil {
				vm.ExitScope()
				return nil, err
			}

			shouldBreak, err := runBody(vm, e.Body())
			if err != nil {
				vm.ExitScope()
				return nil, err
			}
			if shouldBreak {
				break
			}
		}

	case *DictValue, DictValue:
		// Iterate over key-value pairs of dict
		dict, _ := AsDict(iterable)
		for _, key := range dict.Keys() {
			val, _ := dict.Get(key)
			pair := ArrayValue{Str(key), val}
			if err := Destructure(vm, pattern, pair); err != nil {
				vm.ExitScope()
				return nil, err
			}

			shouldBreak, err := runBody(vm, e.Body())
			if err != nil {
				vm.ExitScope()
				return nil, err
			}
			if shouldBreak {
				break
			}
		}

	case StrValue:
		// Check for destructuring pattern on string
		if _, isDestructure := pattern.(*syntax.DestructuringPattern); isDestructure {
			vm.ExitScope()
			return nil, &IterationError{
				Span:    pattern.ToUntyped().Span(),
				Message: "cannot destructure values of " + iterableType.String(),
			}
		}
		// Iterate over graphemes of string (using runes for now)
		for _, ch := range string(v) {
			if err := Destructure(vm, pattern, Str(string(ch))); err != nil {
				vm.ExitScope()
				return nil, err
			}

			shouldBreak, err := runBody(vm, e.Body())
			if err != nil {
				vm.ExitScope()
				return nil, err
			}
			if shouldBreak {
				break
			}
		}

	case BytesValue:
		// Check for destructuring pattern on bytes
		if _, isDestructure := pattern.(*syntax.DestructuringPattern); isDestructure {
			vm.ExitScope()
			return nil, &IterationError{
				Span:    pattern.ToUntyped().Span(),
				Message: "cannot destructure values of " + iterableType.String(),
			}
		}
		// Iterate over the integers of bytes
		for _, b := range v {
			if err := Destructure(vm, pattern, Int(int64(b))); err != nil {
				vm.ExitScope()
				return nil, err
			}

			shouldBreak, err := runBody(vm, e.Body())
			if err != nil {
				vm.ExitScope()
				return nil, err
			}
			if shouldBreak {
				break
			}
		}

	default:
		vm.ExitScope()
		return nil, &IterationError{
			Span:    iterExpr.ToUntyped().Span(),
			Message: "cannot loop over " + iterableType.String(),
		}
	}

	vm.ExitScope()

	// Restore saved flow if there was one
	if savedFlow != nil {
		vm.SetFlow(savedFlow)
	}

	// Mark the return as conditional (it occurred inside a loop)
	MarkReturnAsConditional(vm)

	return output, nil
}

func evalLoopBreak(vm *Vm, e *syntax.LoopBreakExpr) (Value, error) {
	// Only set break if no flow event is already pending
	if !vm.HasFlow() {
		vm.SetFlow(NewBreakEvent(e.ToUntyped().Span()))
	}
	return None, nil
}

func evalLoopContinue(vm *Vm, e *syntax.LoopContinueExpr) (Value, error) {
	// Only set continue if no flow event is already pending
	if !vm.HasFlow() {
		vm.SetFlow(NewContinueEvent(e.ToUntyped().Span()))
	}
	return None, nil
}

func evalFuncReturn(vm *Vm, e *syntax.FuncReturnExpr) (Value, error) {
	// Evaluate return value first (even if flow is already set)
	var value Value = None
	if body := e.Body(); body != nil {
		var err error
		value, err = EvalExpr(vm, body)
		if err != nil {
			return nil, err
		}
	}

	// Only set return if no flow event is already pending
	if !vm.HasFlow() {
		vm.SetFlow(NewReturnEventWithValue(e.ToUntyped().Span(), value))
	}
	return None, nil
}

func evalContextual(vm *Vm, e *syntax.ContextualExpr) (Value, error) {
	// Context expressions create a closure that captures context
	body := e.Body()
	if body == nil {
		return None, nil
	}

	// Create a contextual closure
	captured := vm.Scopes.FlattenToScope()

	closure := &Closure{
		Node:         ContextAstNode{Node: e.ToUntyped()},
		Defaults:     nil,
		Captured:     captured,
		NumPosParams: 0,
	}

	fn := &Func{
		Name: nil,
		Span: e.ToUntyped().Span(),
		Repr: ClosureFunc{Closure: closure},
	}

	return FuncValue{Func: fn}, nil
}

// Note: evalSetRule and evalShowRule are implemented in rules.go
// Note: evalModuleImport and evalModuleInclude are implemented in import.go

// ----------------------------------------------------------------------------
// Markup Evaluators
// ----------------------------------------------------------------------------

func evalText(_ *Vm, e *syntax.TextExpr) (Value, error) {
	return ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: e.Get()}},
	}}, nil
}

func evalSpace(_ *Vm, _ *syntax.SpaceExpr) (Value, error) {
	return ContentValue{Content: Content{
		Elements: []ContentElement{&SpaceElement{}},
	}}, nil
}

func evalLinebreak(_ *Vm, _ *syntax.LinebreakExpr) (Value, error) {
	return ContentValue{Content: Content{
		Elements: []ContentElement{&LinebreakElement{}},
	}}, nil
}

type LinebreakElement struct{}

func (*LinebreakElement) IsContentElement() {}

func evalParbreak(_ *Vm, _ *syntax.ParbreakExpr) (Value, error) {
	return ContentValue{Content: Content{
		Elements: []ContentElement{&ParbreakElement{}},
	}}, nil
}

type ParbreakElement struct{}

func (*ParbreakElement) IsContentElement() {}

// ParagraphElement represents a paragraph with styling properties.
// This wraps content in paragraph-level formatting.
type ParagraphElement struct {
	// Body is the content of the paragraph.
	Body Content
	// Leading is the spacing between lines (in points).
	// If nil, uses default leading (0.65em).
	Leading *float64
	// Justify indicates whether to justify the paragraph text.
	// If nil, uses default (false).
	Justify *bool
	// Linebreaks specifies the line breaking algorithm.
	// Values: "simple", "optimized", or nil for auto.
	Linebreaks *string
	// FirstLineIndent is the indent for the first line (in points).
	// If nil, uses default (0pt).
	FirstLineIndent *float64
	// HangingIndent is the indent for subsequent lines (in points).
	// If nil, uses default (0pt).
	HangingIndent *float64
}

func (*ParagraphElement) IsContentElement() {}

func evalStrong(vm *Vm, e *syntax.StrongExpr) (Value, error) {
	body := e.Body()
	if body == nil {
		return ContentValue{}, nil
	}
	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}
	if c, ok := content.(ContentValue); ok {
		return ContentValue{Content: Content{
			Elements: []ContentElement{&StrongElement{Content: c.Content}},
		}}, nil
	}
	return content, nil
}

type StrongElement struct {
	Content Content
}

func (*StrongElement) IsContentElement() {}

func evalEmph(vm *Vm, e *syntax.EmphExpr) (Value, error) {
	body := e.Body()
	if body == nil {
		return ContentValue{}, nil
	}
	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}
	if c, ok := content.(ContentValue); ok {
		return ContentValue{Content: Content{
			Elements: []ContentElement{&EmphElement{Content: c.Content}},
		}}, nil
	}
	return content, nil
}

type EmphElement struct {
	Content Content
}

func (*EmphElement) IsContentElement() {}

func evalRaw(_ *Vm, e *syntax.RawExpr) (Value, error) {
	// Join lines into a single string
	lines := e.Lines()
	text := ""
	for i, line := range lines {
		if i > 0 {
			text += "\n"
		}
		text += line
	}
	return ContentValue{Content: Content{
		Elements: []ContentElement{&RawElement{Text: text, Lang: e.Lang(), Block: e.Block()}},
	}}, nil
}

type RawElement struct {
	Text  string
	Lang  string
	Block bool
}

func (*RawElement) IsContentElement() {}

func evalLink(_ *Vm, e *syntax.LinkExpr) (Value, error) {
	return ContentValue{Content: Content{
		Elements: []ContentElement{&LinkElement{URL: e.Get()}},
	}}, nil
}

type LinkElement struct {
	URL string
}

func (*LinkElement) IsContentElement() {}

func evalLabel(_ *Vm, e *syntax.LabelExpr) (Value, error) {
	return LabelValue(e.Get()), nil
}

func evalRef(vm *Vm, e *syntax.RefExpr) (Value, error) {
	// Evaluate optional supplement content
	var supplement *Content
	if supp := e.Supplement(); supp != nil {
		suppValue, err := EvalExpr(vm, supp)
		if err != nil {
			return nil, err
		}
		if c, ok := suppValue.(ContentValue); ok {
			supplement = &c.Content
		}
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&RefElement{Target: e.Target(), Supplement: supplement}},
	}}, nil
}

// RefElement represents a reference to a labeled element.
type RefElement struct {
	Target     string   // The label being referenced
	Supplement *Content // Optional supplement content (e.g., @label[supplement])
}

func (*RefElement) IsContentElement() {}

func evalHeading(vm *Vm, e *syntax.HeadingExpr) (Value, error) {
	body := e.Body()
	if body == nil {
		return ContentValue{}, nil
	}
	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}
	if c, ok := content.(ContentValue); ok {
		return ContentValue{Content: Content{
			Elements: []ContentElement{&HeadingElement{
				Level:    e.Level(),
				Content:  c.Content,
				Outlined: true, // Default: show in outline
			}},
		}}, nil
	}
	return content, nil
}

type HeadingElement struct {
	Level      int
	Content    Content
	Numbering  *string // Optional numbering pattern (e.g., "1.", "1.1", "I.")
	Supplement *Content
	Outlined   bool
	Bookmarked *bool
}

func (*HeadingElement) IsContentElement() {}

func evalListItem(vm *Vm, e *syntax.ListItemExpr) (Value, error) {
	body := e.Body()
	if body == nil {
		return ContentValue{}, nil
	}
	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}
	if c, ok := content.(ContentValue); ok {
		return ContentValue{Content: Content{
			Elements: []ContentElement{&ListItemElement{Content: c.Content}},
		}}, nil
	}
	return content, nil
}

type ListItemElement struct {
	Content Content
}

func (*ListItemElement) IsContentElement() {}

func evalEnumItem(vm *Vm, e *syntax.EnumItemExpr) (Value, error) {
	body := e.Body()
	if body == nil {
		return ContentValue{}, nil
	}
	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}
	if c, ok := content.(ContentValue); ok {
		return ContentValue{Content: Content{
			Elements: []ContentElement{&EnumItemElement{Number: e.Number(), Content: c.Content}},
		}}, nil
	}
	return content, nil
}

type EnumItemElement struct {
	Number  int
	Content Content
}

func (*EnumItemElement) IsContentElement() {}

func evalTermItem(vm *Vm, e *syntax.TermItemExpr) (Value, error) {
	term := e.Term()
	desc := e.Description()

	var termContent, descContent Content
	if term != nil {
		if v, err := evalMarkup(vm, term); err == nil {
			if c, ok := v.(ContentValue); ok {
				termContent = c.Content
			}
		}
	}
	if desc != nil {
		if v, err := evalMarkup(vm, desc); err == nil {
			if c, ok := v.(ContentValue); ok {
				descContent = c.Content
			}
		}
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&TermItemElement{Term: termContent, Description: descContent}},
	}}, nil
}

type TermItemElement struct {
	Term        Content
	Description Content
}

func (*TermItemElement) IsContentElement() {}

func evalEscape(_ *Vm, e *syntax.EscapeExpr) (Value, error) {
	// Get the full escape text to handle Unicode escapes
	text := e.ToUntyped().Text()
	char := parseEscapeSequence(text)
	return ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: char}},
	}}, nil
}

// parseEscapeSequence parses an escape sequence and returns the resulting character(s).
func parseEscapeSequence(text string) string {
	if len(text) < 2 || text[0] != '\\' {
		return text
	}

	// Handle Unicode escape: \u{XXXX}
	if len(text) >= 4 && text[1] == 'u' && text[2] == '{' {
		// Find closing brace
		end := 3
		for end < len(text) && text[end] != '}' {
			end++
		}
		if end < len(text) {
			hex := text[3:end]
			if codepoint, err := strconv.ParseUint(hex, 16, 32); err == nil {
				return string(rune(codepoint))
			}
		}
	}

	// Simple escape: \X returns X
	return string(text[1])
}

func evalShorthand(_ *Vm, e *syntax.ShorthandExpr) (Value, error) {
	// Convert shorthand to its Unicode symbol
	text := e.Get()
	symbol := shorthandToSymbol(text)
	return ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: symbol}},
	}}, nil
}

// shorthandToSymbol converts a shorthand text to its Unicode symbol.
func shorthandToSymbol(text string) string {
	switch text {
	case "~":
		return "\u00A0" // Non-breaking space
	case "---":
		return "\u2014" // Em dash
	case "--":
		return "\u2013" // En dash
	case "-?":
		return "\u00AD" // Soft hyphen
	case "...":
		return "\u2026" // Horizontal ellipsis
	default:
		// Check for minus sign before numbers (e.g., "-1")
		if len(text) >= 2 && text[0] == '-' && text[1] >= '0' && text[1] <= '9' {
			return "\u2212" + text[1:] // Minus sign + number
		}
		return text
	}
}

func evalSmartQuote(_ *Vm, e *syntax.SmartQuoteExpr) (Value, error) {
	// Create a SmartQuoteElement that tracks the quote type
	// The actual opening/closing determination happens during layout/rendering
	return ContentValue{Content: Content{
		Elements: []ContentElement{&SmartQuoteElement{Double: e.Double()}},
	}}, nil
}

// SmartQuoteElement represents a smart quote in content.
// The actual quote character is determined during layout based on context.
type SmartQuoteElement struct {
	Double bool // true for double quotes, false for single quotes
}

func (*SmartQuoteElement) IsContentElement() {}

// PageElement represents a page configuration element.
// It can be used to set page properties and optionally wrap content.
// When used as `#page()[content]`, it creates a page break and applies
// the properties to that specific page.
type PageElement struct {
	// Body is the optional content for this page.
	// If nil, this element only applies set-rule style configuration.
	Body *Content

	// Width is the page width in points.
	// If nil, uses default (A4 width: 595.276pt).
	Width *float64

	// Height is the page height in points.
	// If nil, uses default (A4 height: 841.89pt).
	Height *float64

	// Margin specifies page margins.
	// Individual margins (top, bottom, left, right) can be set independently.
	Margin *PageMargin

	// Flipped indicates whether width and height should be swapped.
	// If nil, uses default (false).
	Flipped *bool

	// Fill is the page background fill (color or gradient).
	// If nil, uses default (none/transparent).
	Fill *Color

	// Numbering is the page numbering pattern (e.g., "1", "i", "a").
	// If nil, uses default (no numbering).
	Numbering *string

	// NumberAlign specifies where page numbers are placed.
	// Values: "center", "left", "right", or combined like "center + bottom".
	// If nil, uses default ("center + bottom").
	NumberAlign *Alignment2D

	// Header is the header content.
	// Can be content or a function receiving page context.
	// If nil, uses default (none).
	Header *Content

	// Footer is the footer content.
	// Can be content or a function receiving page context.
	// If nil, uses default (none).
	Footer *Content

	// HeaderAscent is the space between header baseline and main content.
	// If nil, uses default (30% of top margin).
	HeaderAscent *float64

	// FooterDescent is the space between footer baseline and main content.
	// If nil, uses default (30% of bottom margin).
	FooterDescent *float64

	// Background is content placed behind the page content.
	// If nil, uses default (none).
	Background *Content

	// Foreground is content placed in front of the page content.
	// If nil, uses default (none).
	Foreground *Content

	// Columns is the number of columns for the page.
	// If nil, uses default (1).
	Columns *int

	// Binding specifies which side the page is bound.
	// Values: "left" or "right".
	// If nil, uses default based on text direction (left for LTR).
	Binding *string
}

func (*PageElement) IsContentElement() {}

// PageMargin represents page margin configuration.
type PageMargin struct {
	// Top margin in points. If nil, uses default.
	Top *float64
	// Bottom margin in points. If nil, uses default.
	Bottom *float64
	// Left margin in points. If nil, uses default.
	Left *float64
	// Right margin in points. If nil, uses default.
	Right *float64
	// Inside margin for two-sided documents. If nil, uses Left.
	Inside *float64
	// Outside margin for two-sided documents. If nil, uses Right.
	Outside *float64
	// X sets both left and right margins. If nil, uses individual values.
	X *float64
	// Y sets both top and bottom margins. If nil, uses individual values.
	Y *float64
	// Rest sets all unspecified margins. If nil, uses default.
	Rest *float64
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// UnsupportedExprError is returned when evaluating an unsupported expression type.
type UnsupportedExprError struct {
	Expr syntax.Expr
}

func (e *UnsupportedExprError) Error() string {
	if e.Expr != nil {
		return fmt.Sprintf("unsupported expression type: %s", e.Expr.Kind())
	}
	return "unsupported expression type"
}

// TypeError is returned when a value has an unexpected type.
type TypeError struct {
	Expected Type
	Got      Type
	Span     syntax.Span
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("expected %s, got %s", e.Expected, e.Got)
}

// IterationError is returned when a loop iteration fails.
type IterationError struct {
	Message string
	Span    syntax.Span
}

func (e *IterationError) Error() string {
	return e.Message
}

// FieldNotFoundError is returned when accessing a non-existent field.
type FieldNotFoundError struct {
	Field string
	Type  Type
	Span  syntax.Span
}

func (e *FieldNotFoundError) Error() string {
	return fmt.Sprintf("field %q not found on %s", e.Field, e.Type)
}
