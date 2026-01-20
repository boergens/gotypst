package eval

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Args Methods - Argument Processing
// ----------------------------------------------------------------------------

// NewArgs creates a new empty Args.
func NewArgs(span syntax.Span) *Args {
	return &Args{
		Span:  span,
		Items: nil,
	}
}

// NewArgsFrom creates Args from a slice of argument items.
func NewArgsFrom(span syntax.Span, items []Arg) *Args {
	return &Args{
		Span:  span,
		Items: items,
	}
}

// Push adds a positional argument to the end.
func (a *Args) Push(value Value, span syntax.Span) {
	a.Items = append(a.Items, Arg{
		Span:  span,
		Name:  nil,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	})
}

// PushNamed adds a named argument.
func (a *Args) PushNamed(name string, value Value, span syntax.Span) {
	a.Items = append(a.Items, Arg{
		Span:  span,
		Name:  &name,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	})
}

// InsertAt inserts a positional argument at the given index.
func (a *Args) InsertAt(index int, value Value, span syntax.Span) {
	arg := Arg{
		Span:  span,
		Name:  nil,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	}
	if index >= len(a.Items) {
		a.Items = append(a.Items, arg)
	} else {
		a.Items = append(a.Items[:index+1], a.Items[index:]...)
		a.Items[index] = arg
	}
}

// Expect takes and returns the first positional argument.
// Returns an error if no positional argument is available.
func (a *Args) Expect(what string) (syntax.Spanned[Value], error) {
	for i, arg := range a.Items {
		if arg.Name == nil {
			result := arg.Value
			// Remove the argument
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return result, nil
		}
	}
	return syntax.Spanned[Value]{}, &MissingArgumentError{
		What: what,
		Span: a.Span,
	}
}

// Find takes and returns a named argument by name.
// Returns nil if not found.
func (a *Args) Find(name string) *syntax.Spanned[Value] {
	for i, arg := range a.Items {
		if arg.Name != nil && *arg.Name == name {
			result := arg.Value
			// Remove the argument
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return &result
		}
	}
	return nil
}

// Named takes and returns a named argument by name.
// If not found, returns the default value.
func (a *Args) Named(name string) *syntax.Spanned[Value] {
	return a.Find(name)
}

// NamedOrDefault takes a named argument or returns the default.
func (a *Args) NamedOrDefault(name string, def Value) syntax.Spanned[Value] {
	if result := a.Find(name); result != nil {
		return *result
	}
	return syntax.Spanned[Value]{V: def, Span: a.Span}
}

// Eat attempts to take and return the first positional argument.
// Returns nil if no positional argument is available (no error).
func (a *Args) Eat() *syntax.Spanned[Value] {
	for i, arg := range a.Items {
		if arg.Name == nil {
			result := arg.Value
			// Remove the argument
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return &result
		}
	}
	return nil
}

// All takes and returns all remaining positional arguments.
func (a *Args) All() []syntax.Spanned[Value] {
	var result []syntax.Spanned[Value]
	var remaining []Arg
	for _, arg := range a.Items {
		if arg.Name == nil {
			result = append(result, arg.Value)
		} else {
			remaining = append(remaining, arg)
		}
	}
	a.Items = remaining
	return result
}

// Take returns the first positional argument without removing it.
func (a *Args) Take() *syntax.Spanned[Value] {
	for _, arg := range a.Items {
		if arg.Name == nil {
			return &arg.Value
		}
	}
	return nil
}

// HasPositional returns true if there are positional arguments remaining.
func (a *Args) HasPositional() bool {
	for _, arg := range a.Items {
		if arg.Name == nil {
			return true
		}
	}
	return false
}

// HasNamed returns true if a named argument with the given name exists.
func (a *Args) HasNamed(name string) bool {
	for _, arg := range a.Items {
		if arg.Name != nil && *arg.Name == name {
			return true
		}
	}
	return false
}

// GetNamed returns a named argument by name without removing it.
// Returns nil if not found.
func (a *Args) GetNamed(name string) *syntax.Spanned[Value] {
	for _, arg := range a.Items {
		if arg.Name != nil && *arg.Name == name {
			return &arg.Value
		}
	}
	return nil
}

// Remaining returns the number of remaining arguments.
func (a *Args) Remaining() int {
	return len(a.Items)
}

// IsEmpty returns true if no arguments remain.
func (a *Args) IsEmpty() bool {
	return len(a.Items) == 0
}

// Finish checks that all arguments have been consumed.
// Returns an error if unexpected arguments remain.
func (a *Args) Finish() error {
	if len(a.Items) > 0 {
		arg := a.Items[0]
		if arg.Name != nil {
			return &UnexpectedArgumentError{
				Name: arg.Name,
				Span: arg.Span,
			}
		}
		return &UnexpectedArgumentError{
			Name: nil,
			Span: arg.Span,
		}
	}
	return nil
}

// Clone creates a deep copy of the Args.
func (a *Args) Clone() *Args {
	if a == nil {
		return nil
	}
	items := make([]Arg, len(a.Items))
	for i, arg := range a.Items {
		items[i] = Arg{
			Span:  arg.Span,
			Name:  arg.Name,
			Value: syntax.Spanned[Value]{V: arg.Value.V.Clone(), Span: arg.Value.Span},
		}
	}
	return &Args{
		Span:  a.Span,
		Items: items,
	}
}

// ----------------------------------------------------------------------------
// Function Call System
// ----------------------------------------------------------------------------

// CallFunc calls a function with the given arguments.
// This is the main entry point for calling functions in the evaluator.
func CallFunc(vm *Vm, f *Func, args *Args) (Value, error) {
	// Check call depth
	if err := vm.CheckCallDepth(); err != nil {
		return nil, err
	}

	// Dispatch based on function representation
	switch repr := f.Repr.(type) {
	case NativeFunc:
		return callNative(vm, repr, args)
	case ClosureFunc:
		return callClosure(vm, f, repr.Closure, args)
	case WithFunc:
		return callWith(vm, repr, args)
	default:
		return nil, &InvalidCalleeError{
			Value: FuncValue{Func: f},
			Span:  f.Span,
		}
	}
}

// callNative calls a native (built-in) function.
func callNative(vm *Vm, native NativeFunc, args *Args) (Value, error) {
	vm.EnterCall()
	defer vm.ExitCall()

	return native.Func(vm, args)
}

// callClosure calls a user-defined closure.
func callClosure(vm *Vm, f *Func, closure *Closure, args *Args) (Value, error) {
	vm.EnterCall()
	defer vm.ExitCall()

	// Create new scopes starting with captured scope
	savedScopes := vm.Scopes
	vm.Scopes = NewScopes(nil)
	vm.Scopes.SetTop(closure.Captured.Clone())
	defer func() { vm.Scopes = savedScopes }()

	// Bind function name for recursion (if named)
	if f.Name != nil {
		vm.Define(*f.Name, FuncValue{Func: f})
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
				val, err := args.Expect(p.Name().Get())
				if err != nil {
					return nil, err
				}
				vm.Define(p.Name().Get(), val.V)

			case *syntax.NamedParam:
				// Named parameter with default
				name := p.Name().Get()
				val := args.Find(name)
				if val != nil {
					vm.Define(name, val.V)
				} else if defaultIdx < len(closure.Defaults) {
					vm.Define(name, closure.Defaults[defaultIdx])
				} else {
					return nil, &MissingArgumentError{What: name, Span: args.Span}
				}
				defaultIdx++

			case *syntax.SinkParam:
				// Sink parameter (..rest)
				if p.Name() != nil {
					// Collect remaining positional args
					remaining := args.All()
					arr := make(ArrayValue, len(remaining))
					for i, v := range remaining {
						arr[i] = v.V
					}
					vm.Define(p.Name().Get(), arr)
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
				// Convert DestructuringNode to DestructuringPattern for Destructure
				pattern := syntax.DestructuringPatternFromNode(p.Pattern().ToUntyped())
				if err := Destructure(vm, pattern, val.V); err != nil {
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
			return None, nil
		}
		// Propagate break/continue as error (forbidden in functions)
		return nil, CheckForbiddenFlow(flow, false, false, false)
	}

	_ = bodySpan // For future error reporting
	return output, nil
}

// callWith calls a function with pre-applied arguments.
func callWith(vm *Vm, with WithFunc, args *Args) (Value, error) {
	// Merge pre-applied args with new args
	merged := with.Args.Clone()
	for _, arg := range args.Items {
		merged.Items = append(merged.Items, arg)
	}
	return CallFunc(vm, with.Func, merged)
}

// evalExpr evaluates an expression and returns its value.
// This is a placeholder that should dispatch to the actual evaluator.
func evalExpr(vm *Vm, expr syntax.Expr) (Value, error) {
	if expr == nil {
		return None, nil
	}
	// TODO: Implement full expression evaluation
	// For now, delegate to a simple evaluator
	return evalSimpleExpr(vm, expr)
}

// evalSimpleExpr delegates to the main expression evaluator.
// This exists for backwards compatibility - new code should use EvalExpr directly.
func evalSimpleExpr(vm *Vm, expr syntax.Expr) (Value, error) {
	return EvalExpr(vm, expr)
}

// evalClosureExpr evaluates a closure expression and creates a closure value.
func evalClosureExpr(vm *Vm, expr *syntax.ClosureExpr) (Value, error) {
	// Evaluate default values for named parameters
	var defaults []Value
	if params := expr.Params(); params != nil {
		for _, param := range params.Children() {
			if named, ok := param.(*syntax.NamedParam); ok {
				if defExpr := named.Default(); defExpr != nil {
					def, err := evalSimpleExpr(vm, defExpr)
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
	closure := &Closure{
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
	f := &Func{
		Name: name,
		Span: expr.ToUntyped().Span(),
		Repr: ClosureFunc{Closure: closure},
	}

	return FuncValue{Func: f}, nil
}

// captureScope captures variables from the current scope for a closure.
func captureScope(vm *Vm, expr *syntax.ClosureExpr) *Scope {
	// For simplicity, capture the entire flattened scope
	// A more sophisticated implementation would use CapturesVisitor
	// to only capture variables that are actually used
	return vm.Scopes.FlattenToScope()
}

// ----------------------------------------------------------------------------
// Value Comparison
// ----------------------------------------------------------------------------

// valuesEqual checks if two values are equal.
func valuesEqual(lhs, rhs Value) bool {
	// Type must match
	if lhs.Type() != rhs.Type() {
		// Special case: int/float comparison
		if lf, lok := AsFloat(lhs); lok {
			if rf, rok := AsFloat(rhs); rok {
				return lf == rf
			}
		}
		return false
	}

	switch l := lhs.(type) {
	case NoneValue:
		return true
	case AutoValue:
		return true
	case BoolValue:
		if r, ok := rhs.(BoolValue); ok {
			return l == r
		}
	case IntValue:
		if r, ok := rhs.(IntValue); ok {
			return l == r
		}
	case FloatValue:
		if r, ok := rhs.(FloatValue); ok {
			return l == r
		}
	case StrValue:
		if r, ok := rhs.(StrValue); ok {
			return l == r
		}
	}
	return false
}

// ----------------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------------

// MissingArgumentError is returned when a required argument is missing.
type MissingArgumentError struct {
	What string
	Span syntax.Span
}

func (e *MissingArgumentError) Error() string {
	return fmt.Sprintf("missing argument: %s", e.What)
}

// UnexpectedArgumentError is returned when an unexpected argument is provided.
type UnexpectedArgumentError struct {
	Name *string
	Span syntax.Span
}

func (e *UnexpectedArgumentError) Error() string {
	if e.Name != nil {
		return fmt.Sprintf("unexpected argument: %s", *e.Name)
	}
	return "unexpected positional argument"
}

// InvalidCalleeError is returned when trying to call a non-function value.
type InvalidCalleeError struct {
	Value Value
	Span  syntax.Span
}

func (e *InvalidCalleeError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("cannot call value of type %s", e.Value.Type())
	}
	return "cannot call nil value"
}

// TypeMismatchError is returned when a value has an unexpected type.
type TypeMismatchError struct {
	Expected string
	Got      string
	Span     syntax.Span
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("expected %s, got %s", e.Expected, e.Got)
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

