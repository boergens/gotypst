package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Args Constructors - Re-exported for convenience
// ----------------------------------------------------------------------------

// NewArgs creates a new empty Args.
// Re-exported from foundations for convenience.
// Type aliases for foundations types used in the eval package.
type Value = foundations.Value
type Scope = foundations.Scope
type Scopes = foundations.Scopes
type Binding = foundations.Binding
type BindingKind = foundations.BindingKind
type Category = foundations.Category
type Func = foundations.Func
type Closure = foundations.Closure
type Args = foundations.Args
type NativeFunc = foundations.NativeFunc
type ClosureFunc = foundations.ClosureFunc
type WithFunc = foundations.WithFunc
type ClosureAstNode = foundations.ClosureAstNode
type ContextAstNode = foundations.ContextAstNode

// Value type aliases - using actual foundations type names
type NoneValue = foundations.NoneValue
type AutoValue = foundations.AutoValue
type IntValue = foundations.Int
type FloatValue = foundations.Float
type LengthValue = foundations.LengthValue
type AngleValue = foundations.AngleValue
type RatioValue = foundations.RatioValue
type RelativeValue = foundations.RelativeValue
type FractionValue = foundations.FractionValue
type ArrayValue = foundations.Array
type DictValue = foundations.Dict
type FuncValue = foundations.FuncValue
type ContentValue = foundations.ContentValue
type TypeValue = foundations.TypeValue
type ModuleValue = foundations.ModuleValue
type SymbolValue = foundations.SymbolValue
type LabelValue = foundations.LabelValue
type BytesValue = foundations.BytesValue
type VersionValue = foundations.VersionValue

// Helper type aliases
type Type = foundations.Type
type Length = foundations.Length
type Angle = foundations.Angle
type Ratio = foundations.Ratio
type Relative = foundations.Relative
type Fraction = foundations.Fraction
type Content = foundations.Content
type ContentElement = foundations.ContentElement
type Module = foundations.Module
type FileID = syntax.FileId

// Type constants
const (
	TypeNone     = foundations.TypeNone
	TypeAuto     = foundations.TypeAuto
	TypeBool     = foundations.TypeBool
	TypeInt      = foundations.TypeInt
	TypeFloat    = foundations.TypeFloat
	TypeStr      = foundations.TypeStr
	TypeLength   = foundations.TypeLength
	TypeAngle    = foundations.TypeAngle
	TypeRatio    = foundations.TypeRatio
	TypeRelative = foundations.TypeRelative
	TypeFraction = foundations.TypeFraction
	TypeColor    = foundations.TypeColor
	TypeArray    = foundations.TypeArray
	TypeDict     = foundations.TypeDict
	TypeFunc     = foundations.TypeFunc
	TypeContent  = foundations.TypeContent
	TypeType     = foundations.TypeType
	TypeModule   = foundations.TypeModule
	TypeSymbol   = foundations.TypeSymbol
	TypeLabel    = foundations.TypeLabel
	TypeBytes    = foundations.TypeBytes
	TypeVersion  = foundations.TypeVersion
)

// Value constructors and singletons
var (
	None = foundations.NoneValue{}
	Auto = foundations.AutoValue{}
)

// Constructor functions
func Bool(b bool) foundations.Bool   { return foundations.Bool(b) }
func Int(i int64) foundations.Int    { return foundations.Int(i) }
func Float(f float64) foundations.Float { return foundations.Float(f) }
func Str(s string) foundations.Str   { return foundations.Str(s) }

// Scope constructors
var NewScope = foundations.NewScope
var NewScopes = foundations.NewScopes

// Args constructors
var NewArgs = foundations.NewArgs

// Arg type alias
type Arg = foundations.Arg

// ArgsValue type alias
type ArgsValue = foundations.ArgsValue

// AsFunc attempts to cast a Value to a Func.
// Returns the Func and true if the value is a function, nil and false otherwise.
func AsFunc(v foundations.Value) (*foundations.Func, bool) {
	if fv, ok := v.(foundations.FuncValue); ok {
		return fv.Func, true
	}
	return nil, false
}

// AsInt attempts to cast a Value to int64.
func AsInt(v foundations.Value) (int64, bool) {
	if i, ok := v.(foundations.Int); ok {
		return int64(i), true
	}
	return 0, false
}

// AsStr attempts to cast a Value to string.
func AsStr(v foundations.Value) (string, bool) {
	if s, ok := v.(foundations.Str); ok {
		return string(s), true
	}
	return "", false
}

// AsBool attempts to cast a Value to bool.
func AsBool(v foundations.Value) (bool, bool) {
	if b, ok := v.(foundations.Bool); ok {
		return bool(b), true
	}
	return false, false
}

// AsDict attempts to cast a Value to *Dict.
func AsDict(v foundations.Value) (*foundations.Dict, bool) {
	if d, ok := v.(*foundations.Dict); ok {
		return d, true
	}
	return nil, false
}

// IsNone returns true if the value is None.
func IsNone(v foundations.Value) bool {
	_, ok := v.(foundations.NoneValue)
	return ok
}

// Bool singletons
var (
	True  = foundations.Bool(true)
	False = foundations.Bool(false)
)

// StrValue type alias for use in type switches
type StrValue = foundations.Str

// NewDict creates a new empty dictionary.
var NewDict = foundations.NewDict

// NewArgsFrom creates Args from a slice of argument items.
// Re-exported from foundations for convenience.
var NewArgsFrom = foundations.NewArgsFrom

// ----------------------------------------------------------------------------
// Args Evaluation
// ----------------------------------------------------------------------------

// evalArgs evaluates function call arguments.
// Matches Rust: impl Eval for ast::Args<'_>
func evalArgs(vm *Vm, argsNode *syntax.ArgsNode) (*foundations.Args, error) {
	if argsNode == nil {
		return foundations.NewArgs(syntax.Detached()), nil
	}

	items := argsNode.Items()
	// We use Span::detached() initially since the callsite span is set later
	args := foundations.NewArgs(syntax.Detached())

	for _, item := range items {
		span := item.ToUntyped().Span()

		switch arg := item.(type) {
		case *syntax.PosArg:
			// Positional argument
			expr := arg.Expr()
			value, err := evalExpr(vm, expr)
			if err != nil {
				return nil, err
			}
			args.Items = append(args.Items, foundations.Arg{
				Span:  span,
				Name:  nil,
				Value: syntax.Spanned[foundations.Value]{V: value, Span: expr.ToUntyped().Span()},
			})

		case *syntax.NamedArg:
			// Named argument
			expr := arg.Expr()
			value, err := evalExpr(vm, expr)
			if err != nil {
				return nil, err
			}
			var name *foundations.Str
			if nameIdent := arg.Name(); nameIdent != nil {
				n := foundations.Str(nameIdent.Get())
				name = &n
			}
			args.Items = append(args.Items, foundations.Arg{
				Span:  span,
				Name:  name,
				Value: syntax.Spanned[foundations.Value]{V: value, Span: expr.ToUntyped().Span()},
			})

		case *syntax.SpreadArg:
			// Spread argument
			expr := arg.Expr()
			if expr == nil {
				continue
			}
			value, err := evalExpr(vm, expr)
			if err != nil {
				return nil, err
			}

			switch v := value.(type) {
			case foundations.NoneValue:
				// None spreads as nothing
			case *foundations.Array:
				// Spread array values as positional args
				for _, elem := range v.AsSlice() {
					args.Items = append(args.Items, foundations.Arg{
						Span:  span,
						Name:  nil,
						Value: syntax.Spanned[foundations.Value]{V: elem, Span: span},
					})
				}
			case *foundations.Dict:
				// Spread dict entries as named args
				for _, key := range v.Keys() {
					val, _ := v.Get(key)
					k := foundations.Str(key)
					args.Items = append(args.Items, foundations.Arg{
						Span:  span,
						Name:  &k,
						Value: syntax.Spanned[foundations.Value]{V: val, Span: span},
					})
				}
			case foundations.ArgsValue:
				// Spread args directly
				args.Items = append(args.Items, v.Args.Items...)
			default:
				return nil, atSpan(fmt.Errorf("cannot spread %s", value.Type()), arg.ToUntyped().Span())
			}
		}
	}

	return args, nil
}

// ----------------------------------------------------------------------------
// Function Call System
// ----------------------------------------------------------------------------

// EvalFuncCallExpr evaluates a function call expression.
// This is the main entry point matching Rust's `impl Eval for ast::FuncCall`.
func EvalFuncCallExpr(vm *Vm, expr *syntax.FuncCallExpr) (foundations.Value, error) {
	span := expr.ToUntyped().Span()
	calleeExpr := expr.Callee()
	calleeSpan := calleeExpr.ToUntyped().Span()
	argsNode := expr.Args()

	// Check call depth.
	if err := vm.CheckCallDepth(); err != nil {
		return nil, err
	}

	var calleeValue foundations.Value
	var argsValue *foundations.Args

	// Try to evaluate as a call to an associated function or field.
	if fieldAccess, ok := calleeExpr.(*syntax.FieldAccessExpr); ok {
		targetExpr := fieldAccess.Target()
		field := fieldAccess.Field()

		result, err := evalFieldCall(vm, targetExpr, field, argsNode, span)
		if err != nil {
			return nil, err
		}

		if result.Kind == FieldCallResolved {
			return result.Value, nil
		}

		// Trace for IDE inspection.
		if vm.Inspected != nil && *vm.Inspected == calleeSpan {
			vm.Trace(result.Callee)
		}
		calleeValue = result.Callee
		argsValue = result.Args
	} else {
		// Normal function call: evaluate callee before arguments.
		var err error
		calleeValue, err = evalExpr(vm, calleeExpr)
		if err != nil {
			return nil, err
		}

		argsValue, err = evalArgs(vm, argsNode)
		if err != nil {
			return nil, err
		}
		argsValue.Span = span
	}

	// Try to cast to a function.
	fn, isFunc := AsFunc(calleeValue)

	// If not a function and in math mode, wrap as math content.
	if !isFunc && inMath(calleeExpr) {
		trailingComma := false
		if argsNode != nil {
			trailingComma = argsNode.TrailingComma()
		}
		return wrapArgsInMath(calleeValue, calleeSpan, argsValue, trailingComma)
	}

	if !isFunc {
		return nil, &TypeError{Expected: TypeFunc, Got: calleeValue.Type(), Span: calleeSpan}
	}

	// Call the function with tracing.
	return CallFunc(vm, fn, argsValue)
}

// CallFunc calls a function with the given arguments.
// This is the main entry point for calling functions in the evaluator.
func CallFunc(vm *Vm, f *foundations.Func, args *foundations.Args) (foundations.Value, error) {
	// Check call depth
	if err := vm.CheckCallDepth(); err != nil {
		return nil, err
	}

	// Dispatch based on function representation
	switch repr := f.Repr.(type) {
	case foundations.NativeFunc:
		return callNative(vm, repr, args)
	case foundations.ClosureFunc:
		return callClosure(vm, f, repr.Closure, args)
	case foundations.WithFunc:
		return callWith(vm, repr, args)
	default:
		return nil, &InvalidCalleeError{
			Value: foundations.FuncValue{Func: f},
			Span:  f.Span,
		}
	}
}

// callNative calls a native (built-in) function.
// Passes Engine and Context explicitly, matching Rust's pattern.
func callNative(vm *Vm, native foundations.NativeFunc, args *foundations.Args) (foundations.Value, error) {
	vm.EnterCall()
	defer vm.ExitCall()

	return native.Func(*vm.Engine, *vm.Context, args)
}

// callClosure calls a user-defined closure.
func callClosure(vm *Vm, f *foundations.Func, closure *foundations.Closure, args *foundations.Args) (foundations.Value, error) {
	vm.EnterCall()
	defer vm.ExitCall()

	// Create new scopes starting with captured scope
	savedScopes := vm.Scopes
	vm.Scopes = foundations.NewScopes(nil)
	vm.Scopes.SetTop(closure.Captured)
	defer func() { vm.Scopes = savedScopes }()

	// Bind function name for recursion (if named)
	if f.Name != nil {
		vm.DefineSimple(*f.Name, foundations.FuncValue{Func: f})
	}

	// Get closure node info
	var params *syntax.ParamsNode
	var body syntax.Expr
	var bodySpan syntax.Span

	switch node := closure.Node.(type) {
	case ClosureAstNode:
		closureExpr := syntax.ClosureExprFromNode(node.Node)
		if closureExpr != nil {
			params = closureExpr.Params()
			body = closureExpr.Body()
			if body != nil {
				bodySpan = body.ToUntyped().Span()
			}
		}
	case ContextAstNode:
		contextExpr := syntax.ContextualExprFromNode(node.Node)
		if contextExpr != nil {
			body = contextExpr.Body()
			if body != nil {
				bodySpan = body.ToUntyped().Span()
			}
		}
	}

	// Bind parameters
	if params != nil {
		defaultIdx := 0
		for _, param := range params.Children() {
			switch p := param.(type) {
			case *syntax.PosParam:
				// Positional parameter
				name := p.Name().Get()
				val, err := args.Expect(name)
				if err != nil {
					return nil, err
				}
				vm.DefineSimple(name, val.V)

			case *syntax.NamedParam:
				// Named parameter with default
				name := p.Name().Get()
				val := args.Find(name)
				if val != nil {
					vm.DefineSimple(name, val.V)
				} else if defaultIdx < len(closure.Defaults) {
					vm.DefineSimple(name, closure.Defaults[defaultIdx])
				} else {
					return nil, &MissingArgumentError{Name: name, Span: args.Span}
				}
				defaultIdx++

			case *syntax.SinkParam:
				// Sink parameter (..rest)
				if p.Name() != nil {
					name := p.Name().Get()
					// Collect remaining positional args
					remaining := args.All()
					arr := foundations.NewArray()
					for _, v := range remaining {
						arr.Push(v.V)
					}
					vm.DefineSimple(name, arr)
				}

			case *syntax.PlaceholderParam:
				// Placeholder parameter (_) - consume but don't bind
				args.Eat()

			case *syntax.DestructuringParam:
				// Destructuring parameter
				val, err := args.Expect("argument")
				if err != nil {
					return nil, err
				}
				// Convert DestructuringNode to DestructuringPattern for destructure
				pattern := syntax.DestructuringPatternFromNode(p.Pattern().ToUntyped())
				if err := destructure(vm, pattern, val.V); err != nil {
					return nil, err
				}
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Evaluate body
	output, err := evalExpr(vm, body)
	if err != nil {
		return nil, err
	}

	// Handle return flow event
	if vm.HasFlow() {
		flow := vm.TakeFlow()
		if ret, ok := flow.(ReturnEvent); ok {
			if ret.Value != nil {
				return ret.Value, nil
			}
			return foundations.None, nil
		}
		// Propagate break/continue as error (forbidden in functions)
		return nil, CheckForbiddenFlow(flow, false, false, false)
	}

	_ = bodySpan // For future error reporting
	return output, nil
}

// callWith calls a function with pre-applied arguments.
func callWith(vm *Vm, with foundations.WithFunc, args *foundations.Args) (foundations.Value, error) {
	// Merge pre-applied args with new args
	merged := with.Args.Clone()
	for _, arg := range args.Items {
		merged.Items = append(merged.Items, arg)
	}
	return CallFunc(vm, with.Func, merged)
}


// evalClosureExpr evaluates a closure expression and creates a closure value.
func evalClosureExpr(vm *Vm, expr *syntax.ClosureExpr) (foundations.Value, error) {
	// Evaluate default values for named parameters
	var defaults []foundations.Value
	if params := expr.Params(); params != nil {
		for _, param := range params.Children() {
			if named, ok := param.(*syntax.NamedParam); ok {
				if defExpr := named.Default(); defExpr != nil {
					def, err := evalExpr(vm, defExpr)
					if err != nil {
						return nil, err
					}
					defaults = append(defaults, def)
				}
			}
		}
	}

	// Capture current scope
	captured := captureScope(vm, expr)

	// Count positional parameters
	numPosParams := 0
	if params := expr.Params(); params != nil {
		for _, param := range params.Children() {
			switch param.(type) {
			case *syntax.PosParam, *syntax.DestructuringParam, *syntax.PlaceholderParam:
				numPosParams++
			}
		}
	}

	// Create closure
	closure := &foundations.Closure{
		Node:         ClosureAstNode{Node: expr.ToUntyped()},
		Defaults:     defaults,
		Captured:     captured,
		NumPosParams: numPosParams,
	}

	// Get function name if present
	var name *string
	if nameExpr := expr.Name(); nameExpr != nil {
		n := nameExpr.Get()
		name = &n
	}

	// Create function
	f := &foundations.Func{
		Name: name,
		Span: expr.ToUntyped().Span(),
		Repr: foundations.ClosureFunc{Closure: closure},
	}

	return foundations.FuncValue{Func: f}, nil
}

// captureScope captures variables from the current scope for a closure.
// Returns a foundations.Scope for storage in Closure.
// Uses CapturesVisitor to only capture variables that are actually used.
func captureScope(vm *Vm, expr *syntax.ClosureExpr) *foundations.Scope {
	visitor := NewCapturesVisitor(vm.Scopes, foundations.CapturerFunction)
	visitor.Visit(expr.ToUntyped())
	return visitor.Finish()
}

// ----------------------------------------------------------------------------
// Value Comparison
// ----------------------------------------------------------------------------

// valuesEqual checks if two values are equal.
func valuesEqual(lhs, rhs foundations.Value) bool {
	// Type must match
	if lhs.Type() != rhs.Type() {
		// Special case: int/float comparison
		if lf, lok := foundations.AsFloat(lhs); lok {
			if rf, rok := foundations.AsFloat(rhs); rok {
				return lf == rf
			}
		}
		return false
	}

	switch l := lhs.(type) {
	case foundations.NoneValue:
		return true
	case foundations.AutoValue:
		return true
	case foundations.Bool:
		if r, ok := rhs.(foundations.Bool); ok {
			return l == r
		}
	case foundations.Int:
		if r, ok := rhs.(foundations.Int); ok {
			return l == r
		}
	case foundations.Float:
		if r, ok := rhs.(foundations.Float); ok {
			return l == r
		}
	case foundations.Str:
		if r, ok := rhs.(foundations.Str); ok {
			return l == r
		}
	}
	return false
}

// ----------------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------------

// MissingArgumentError and UnexpectedArgumentError are defined in foundations/args.go
// Type aliases for convenience
type MissingArgumentError = foundations.MissingArgumentError
type UnexpectedArgumentError = foundations.UnexpectedArgumentError

// InvalidCalleeError is returned when trying to call a non-function value.
type InvalidCalleeError struct {
	Value foundations.Value
	Span  syntax.Span
}

func (e *InvalidCalleeError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("cannot call value of type %s", e.Value.Type())
	}
	return "cannot call nil value"
}


// UnsupportedExpressionError is returned for unimplemented expression types.
type UnsupportedExpressionError struct {
	Kind syntax.SyntaxKind
	Span syntax.Span
}

func (e *UnsupportedExpressionError) Error() string {
	return fmt.Sprintf("unsupported expression kind: %s", e.Kind)
}

// UnsupportedOperatorError is returned for unimplemented operators.
type UnsupportedOperatorError struct {
	Op   string
	Span syntax.Span
}

func (e *UnsupportedOperatorError) Error() string {
	return fmt.Sprintf("unsupported operator: %s", e.Op)
}

// InvalidAssignmentTargetError is returned when assigning to an invalid target.
type InvalidAssignmentTargetError struct {
	Span syntax.Span
}

func (e *InvalidAssignmentTargetError) Error() string {
	return "cannot mutate a temporary value"
}

// DestructuringError is returned when destructuring fails.
type DestructuringError struct {
	Message string
	Span    syntax.Span
}

func (e *DestructuringError) Error() string {
	return e.Message
}

// ----------------------------------------------------------------------------
// Field Call Dispatch
// ----------------------------------------------------------------------------

// FieldCallKind indicates the result of evaluating a field call.
type FieldCallKind int

const (
	// FieldCallNormal means we have a callee and args to call.
	FieldCallNormal FieldCallKind = iota
	// FieldCallResolved means the call has already been resolved.
	FieldCallResolved
)

// FieldCallResult holds the result of evaluating a field call.
type FieldCallResult struct {
	Kind   FieldCallKind
	Callee foundations.Value
	Args   *foundations.Args
	Value  foundations.Value // Only set if Kind == FieldCallResolved
}

// evalFieldCall evaluates a field call's callee and arguments.
//
// This follows the normal function call order: we evaluate the callee before the
// arguments. Prioritizes associated functions on the value's type (e.g., methods)
// over its fields.
//
// Matches Rust's eval_field_call in call.rs.
func evalFieldCall(vm *Vm, targetExpr syntax.Expr, field *syntax.IdentExpr, argsNode *syntax.ArgsNode, span syntax.Span) (*FieldCallResult, error) {
	fieldName := field.Get()
	fieldSpan := field.ToUntyped().Span()

	// Evaluate the field-call's target and overall arguments.
	var target foundations.Value
	var args *foundations.Args
	var err error

	if IsMutatingMethod(fieldName) {
		// If field looks like a mutating method, evaluate arguments first,
		// because accessing the target mutably borrows the vm.
		args, err = evalArgs(vm, argsNode)
		if err != nil {
			return nil, err
		}
		args.Span = span

		// Access the target for mutation.
		targetValue, accessErr := evalExpr(vm, targetExpr)
		if accessErr != nil {
			return nil, accessErr
		}

		// Only arrays and dictionaries have mutable methods.
		switch targetValue.(type) {
		case *foundations.Array:
			value, callErr := CallMethodMut(&targetValue, fieldName, args, span)
			if callErr != nil {
				return nil, callErr
			}
			return &FieldCallResult{Kind: FieldCallResolved, Value: value}, nil
		case *foundations.Dict:
			value, callErr := CallMethodMut(&targetValue, fieldName, args, span)
			if callErr != nil {
				return nil, callErr
			}
			return &FieldCallResult{Kind: FieldCallResolved, Value: value}, nil
		default:
			target = targetValue
		}
	} else {
		// Normal order: evaluate target first, then arguments.
		target, err = evalExpr(vm, targetExpr)
		if err != nil {
			return nil, err
		}
		args, err = evalArgs(vm, argsNode)
		if err != nil {
			return nil, err
		}
		args.Span = span
	}

	// TODO: Look up method in target's type scope.
	// This requires implementing Type.Scope() which maps types to their method scopes.
	// For now, we only support direct field access on specific types.

	// Certain value types have their own ways to access method fields.
	switch target.(type) {
	case SymbolValue, FuncValue, TypeValue, ModuleValue:
		value, fieldErr := getField(target, fieldName, fieldSpan)
		if fieldErr != nil {
			return nil, fieldErr
		}
		return &FieldCallResult{Kind: FieldCallNormal, Callee: value, Args: args}, nil
	}

	// Otherwise we cannot call this field.
	return nil, missingFieldCallError(target, field)
}

// GetTypeMethod retrieves a method from a type's scope.
// This is called for method lookup on type values like `int.is-odd`.
func GetTypeMethod(t foundations.Type, method string, span syntax.Span) foundations.Value {
	scope := t.Scope()
	if scope == nil {
		return nil
	}
	binding := scope.Get(method)
	if binding == nil {
		return nil
	}
	return binding.Value()
}

// getField retrieves a field from a value.
func getField(target foundations.Value, field string, span syntax.Span) (foundations.Value, error) {
	switch v := target.(type) {
	case FuncValue:
		if v.Func != nil && v.Func.Repr != nil {
			if nf, ok := v.Func.Repr.(NativeFunc); ok && nf.Scope != nil {
				if binding := nf.Scope.Get(field); binding != nil {
					return binding.Value(), nil
				}
			}
		}
	case TypeValue:
		method := GetTypeMethod(v.Inner, field, span)
		if method != nil {
			return method, nil
		}
	case ModuleValue:
		if v.Module != nil && v.Module.Scope != nil {
			if binding := v.Module.Scope.Get(field); binding != nil {
				return binding.Value(), nil
			}
		}
	case SymbolValue:
		// Symbols can have modifiers accessed as fields.
		// For now, return an error - full implementation would look up symbol modifiers.
	}
	return nil, &FieldNotFoundError{Field: field, Type: target.Type(), Span: span}
}

// missingFieldCallError produces an error when we cannot call the field.
func missingFieldCallError(target foundations.Value, field *syntax.IdentExpr) error {
	fieldName := field.Get()
	fieldSpan := field.ToUntyped().Span()

	// TODO: When Content has Elem support, check for element-specific methods
	// if content, ok := target.(ContentValue); ok {
	//     if elem := content.Content.Elem(); elem != nil {
	//         return &MissingMethodError{Type: elem.Type(), Method: fieldName, Span: fieldSpan}
	//     }
	// }

	return &MissingMethodError{
		Type:   target.Type(),
		Method: fieldName,
		Span:   fieldSpan,
	}
}

// ----------------------------------------------------------------------------
// Math Mode Support
// ----------------------------------------------------------------------------

// inMath checks if the expression is in a math context.
// MathIdent expressions and field accesses on math expressions are in math mode.
func inMath(expr syntax.Expr) bool {
	switch e := expr.(type) {
	case *syntax.MathIdentExpr:
		return true
	case *syntax.FieldAccessExpr:
		return inMath(e.Target())
	default:
		return false
	}
}

// displayString converts a value to its string representation for display.
func displayString(v foundations.Value) string {
	// For now, use a simple string representation
	// TODO: implement proper Display trait support
	switch val := v.(type) {
	case foundations.Str:
		return string(val)
	case foundations.Int:
		return fmt.Sprintf("%d", int64(val))
	case foundations.Float:
		return fmt.Sprintf("%g", float64(val))
	case foundations.Bool:
		if val {
			return "true"
		}
		return "false"
	case foundations.NoneValue:
		return "none"
	case foundations.AutoValue:
		return "auto"
	default:
		return fmt.Sprintf("<%s>", v.Type())
	}
}

// wrapArgsInMath wraps arguments in math mode for non-function calls.
// When calling a non-function in math mode, wrap as callee(arg1, arg2, ...) content.
func wrapArgsInMath(callee foundations.Value, calleeSpan syntax.Span, args *foundations.Args, trailingComma bool) (foundations.Value, error) {
	// Build the body content: arg1, arg2, ...
	var body foundations.Content

	allArgs := args.All()

	for i, arg := range allArgs {
		if i > 0 {
			// Add comma separator
			body.Elements = append(body.Elements, &SymbolElement{Char: ','})
		}
		// Add the argument as content
		if content, ok := arg.V.(ContentValue); ok {
			body.Elements = append(body.Elements, content.Content.Elements...)
		} else {
			// Convert to content using string representation
			body.Elements = append(body.Elements, &TextElement{Text: displayString(arg.V)})
		}
	}

	if trailingComma {
		body.Elements = append(body.Elements, &SymbolElement{Char: ','})
	}

	// Build: callee + LrElem(( + body + ))
	var result foundations.Content

	// Add callee as content using string representation
	result.Elements = append(result.Elements, &TextElement{Text: displayString(callee)})

	// Add parenthesized content
	result.Elements = append(result.Elements, &LrElement{
		Left:  '(',
		Right: ')',
		Body:  body,
	})

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: result}, nil
}

// SymbolElement represents a math symbol character.
type SymbolElement struct {
	Char rune
}

func (*SymbolElement) IsContentElement() {}

// LrElement represents left-right delimited content in math.
type LrElement struct {
	Left  rune
	Right rune
	Body  foundations.Content
}

func (*LrElement) IsContentElement() {}
