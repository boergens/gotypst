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
				if err := destructure(vm, p.Pattern(), val.V); err != nil {
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

// evalSimpleExpr is a basic expression evaluator for testing.
func evalSimpleExpr(vm *Vm, expr syntax.Expr) (Value, error) {
	if expr == nil {
		return None, nil
	}

	switch e := expr.(type) {
	case *syntax.NoneExpr:
		return None, nil

	case *syntax.AutoExpr:
		return Auto, nil

	case *syntax.BoolExpr:
		return Bool(e.Get()), nil

	case *syntax.IntExpr:
		return Int(e.Get()), nil

	case *syntax.FloatExpr:
		return Float(e.Get()), nil

	case *syntax.StrExpr:
		return Str(e.Get()), nil

	case *syntax.IdentExpr:
		name := e.Get()
		binding := vm.Get(name)
		if binding == nil {
			return nil, &UndefinedVariableError{
				Name: name,
				Span: e.ToUntyped().Span(),
			}
		}
		return binding.Value, nil

	case *syntax.CodeBlockExpr:
		if body := e.Body(); body != nil {
			vm.EnterScope()
			defer vm.ExitScope()
			var result Value = None
			for _, child := range body.Exprs() {
				val, err := evalSimpleExpr(vm, child)
				if err != nil {
					return nil, err
				}
				result = val
				if vm.HasFlow() {
					break
				}
			}
			return result, nil
		}
		return None, nil

	case *syntax.ParenthesizedExpr:
		if inner := e.Expr(); inner != nil {
			return evalSimpleExpr(vm, inner)
		}
		return None, nil

	case *syntax.FuncCallExpr:
		return evalFuncCall(vm, e)

	case *syntax.FuncReturnExpr:
		var val Value = None
		if body := e.Body(); body != nil {
			v, err := evalSimpleExpr(vm, body)
			if err != nil {
				return nil, err
			}
			val = v
		}
		vm.SetFlow(NewReturnEventWithValue(e.ToUntyped().Span(), val))
		return None, nil

	case *syntax.ClosureExpr:
		return evalClosureExpr(vm, e)

	case *syntax.BinaryExpr:
		return evalBinaryExpr(vm, e)

	case *syntax.UnaryExpr:
		return evalUnaryExpr(vm, e)

	case *syntax.LetBindingExpr:
		return evalLetBinding(vm, e)

	default:
		return None, &UnsupportedExpressionError{
			Kind: expr.Kind(),
			Span: expr.ToUntyped().Span(),
		}
	}
}

// evalFuncCall evaluates a function call expression.
func evalFuncCall(vm *Vm, call *syntax.FuncCallExpr) (Value, error) {
	// Evaluate callee
	calleeExpr := call.Callee()
	if calleeExpr == nil {
		return nil, &InvalidCalleeError{Span: call.ToUntyped().Span()}
	}

	callee, err := evalSimpleExpr(vm, calleeExpr)
	if err != nil {
		return nil, err
	}

	// Get function from value
	f, ok := AsFunc(callee)
	if !ok {
		return nil, &InvalidCalleeError{
			Value: callee,
			Span:  call.ToUntyped().Span(),
		}
	}

	// Evaluate arguments
	args, err := evalArgs(vm, call.Args())
	if err != nil {
		return nil, err
	}

	// Call the function
	return CallFunc(vm, f, args)
}

// evalArgs evaluates a function call arguments node.
func evalArgs(vm *Vm, argsNode *syntax.ArgsNode) (*Args, error) {
	if argsNode == nil {
		return NewArgs(syntax.Span{}), nil
	}

	args := NewArgs(argsNode.ToUntyped().Span())

	for _, item := range argsNode.Items() {
		switch arg := item.(type) {
		case *syntax.PosArg:
			val, err := evalSimpleExpr(vm, arg.Expr())
			if err != nil {
				return nil, err
			}
			args.Push(val, arg.Expr().ToUntyped().Span())

		case *syntax.NamedArg:
			name := arg.Name().Get()
			val, err := evalSimpleExpr(vm, arg.Expr())
			if err != nil {
				return nil, err
			}
			args.PushNamed(name, val, arg.Expr().ToUntyped().Span())

		case *syntax.SpreadArg:
			val, err := evalSimpleExpr(vm, arg.Expr())
			if err != nil {
				return nil, err
			}
			// Spread array into positional args
			if arr, ok := val.(ArrayValue); ok {
				for _, v := range arr {
					args.Push(v, arg.Expr().ToUntyped().Span())
				}
			} else {
				return nil, &TypeMismatchError{
					Expected: "array",
					Got:      val.Type().String(),
					Span:     arg.Expr().ToUntyped().Span(),
				}
			}
		}
	}

	return args, nil
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

// evalBinaryExpr evaluates a binary expression.
func evalBinaryExpr(vm *Vm, expr *syntax.BinaryExpr) (Value, error) {
	op := expr.Op()

	// Handle short-circuit operators
	if op == syntax.BinOpAnd || op == syntax.BinOpOr {
		return evalShortCircuit(vm, expr)
	}

	// Handle assignment operators
	if isAssignOp(op) {
		return evalAssignment(vm, expr)
	}

	// Evaluate both operands
	lhs, err := evalSimpleExpr(vm, expr.Lhs())
	if err != nil {
		return nil, err
	}
	rhs, err := evalSimpleExpr(vm, expr.Rhs())
	if err != nil {
		return nil, err
	}

	// Apply operation
	return applyBinaryOp(op, lhs, rhs, expr.ToUntyped().Span())
}

// evalShortCircuit evaluates short-circuit boolean operators.
func evalShortCircuit(vm *Vm, expr *syntax.BinaryExpr) (Value, error) {
	lhs, err := evalSimpleExpr(vm, expr.Lhs())
	if err != nil {
		return nil, err
	}

	lhsBool, ok := lhs.(BoolValue)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "bool",
			Got:      lhs.Type().String(),
			Span:     expr.Lhs().ToUntyped().Span(),
		}
	}

	op := expr.Op()
	if op == syntax.BinOpAnd && !bool(lhsBool) {
		return False, nil
	}
	if op == syntax.BinOpOr && bool(lhsBool) {
		return True, nil
	}

	return evalSimpleExpr(vm, expr.Rhs())
}

// isAssignOp returns true if the operator is an assignment operator.
func isAssignOp(op syntax.BinOp) bool {
	switch op {
	case syntax.BinOpAssign, syntax.BinOpAddAssign, syntax.BinOpSubAssign,
		syntax.BinOpMulAssign, syntax.BinOpDivAssign:
		return true
	}
	return false
}

// evalAssignment evaluates an assignment expression.
func evalAssignment(vm *Vm, expr *syntax.BinaryExpr) (Value, error) {
	// Get the target identifier
	lhs := expr.Lhs()
	ident, ok := lhs.(*syntax.IdentExpr)
	if !ok {
		return nil, &InvalidAssignmentTargetError{Span: lhs.ToUntyped().Span()}
	}

	// Evaluate RHS
	rhs, err := evalSimpleExpr(vm, expr.Rhs())
	if err != nil {
		return nil, err
	}

	// Get the binding
	name := ident.Get()
	binding := vm.GetMut(name)
	if binding == nil {
		return nil, &UndefinedVariableError{Name: name, Span: ident.ToUntyped().Span()}
	}

	// For compound assignment, compute new value
	op := expr.Op()
	if op != syntax.BinOpAssign {
		newOp := compoundToSimple(op)
		rhs, err = applyBinaryOp(newOp, binding.Value, rhs, expr.ToUntyped().Span())
		if err != nil {
			return nil, err
		}
	}

	// Write the new value
	if err := binding.Write(rhs); err != nil {
		return nil, err
	}

	return rhs, nil
}

// compoundToSimple converts a compound assignment operator to its simple form.
func compoundToSimple(op syntax.BinOp) syntax.BinOp {
	switch op {
	case syntax.BinOpAddAssign:
		return syntax.BinOpAdd
	case syntax.BinOpSubAssign:
		return syntax.BinOpSub
	case syntax.BinOpMulAssign:
		return syntax.BinOpMul
	case syntax.BinOpDivAssign:
		return syntax.BinOpDiv
	default:
		return op
	}
}

// applyBinaryOp applies a binary operator to two values.
func applyBinaryOp(op syntax.BinOp, lhs, rhs Value, span syntax.Span) (Value, error) {
	switch op {
	case syntax.BinOpAdd:
		return addValues(lhs, rhs, span)
	case syntax.BinOpSub:
		return subValues(lhs, rhs, span)
	case syntax.BinOpMul:
		return mulValues(lhs, rhs, span)
	case syntax.BinOpDiv:
		return divValues(lhs, rhs, span)
	case syntax.BinOpEq:
		return Bool(valuesEqual(lhs, rhs)), nil
	case syntax.BinOpNeq:
		return Bool(!valuesEqual(lhs, rhs)), nil
	case syntax.BinOpLt:
		return compareValues(lhs, rhs, span, func(cmp int) bool { return cmp < 0 })
	case syntax.BinOpLeq:
		return compareValues(lhs, rhs, span, func(cmp int) bool { return cmp <= 0 })
	case syntax.BinOpGt:
		return compareValues(lhs, rhs, span, func(cmp int) bool { return cmp > 0 })
	case syntax.BinOpGeq:
		return compareValues(lhs, rhs, span, func(cmp int) bool { return cmp >= 0 })
	case syntax.BinOpAnd:
		return andValues(lhs, rhs, span)
	case syntax.BinOpOr:
		return orValues(lhs, rhs, span)
	default:
		return nil, &UnsupportedOperatorError{Op: op.String(), Span: span}
	}
}

// addValues implements the + operator.
func addValues(lhs, rhs Value, span syntax.Span) (Value, error) {
	switch l := lhs.(type) {
	case IntValue:
		if r, ok := rhs.(IntValue); ok {
			return Int(int64(l) + int64(r)), nil
		}
		if r, ok := rhs.(FloatValue); ok {
			return Float(float64(l) + float64(r)), nil
		}
	case FloatValue:
		if r, ok := rhs.(IntValue); ok {
			return Float(float64(l) + float64(r)), nil
		}
		if r, ok := rhs.(FloatValue); ok {
			return Float(float64(l) + float64(r)), nil
		}
	case StrValue:
		if r, ok := rhs.(StrValue); ok {
			return Str(string(l) + string(r)), nil
		}
	case ArrayValue:
		if r, ok := rhs.(ArrayValue); ok {
			result := make(ArrayValue, 0, len(l)+len(r))
			result = append(result, l...)
			result = append(result, r...)
			return result, nil
		}
	}
	return nil, &TypeMismatchError{
		Expected: fmt.Sprintf("%s or compatible type", lhs.Type()),
		Got:      rhs.Type().String(),
		Span:     span,
	}
}

// subValues implements the - operator.
func subValues(lhs, rhs Value, span syntax.Span) (Value, error) {
	switch l := lhs.(type) {
	case IntValue:
		if r, ok := rhs.(IntValue); ok {
			return Int(int64(l) - int64(r)), nil
		}
		if r, ok := rhs.(FloatValue); ok {
			return Float(float64(l) - float64(r)), nil
		}
	case FloatValue:
		if r, ok := rhs.(IntValue); ok {
			return Float(float64(l) - float64(r)), nil
		}
		if r, ok := rhs.(FloatValue); ok {
			return Float(float64(l) - float64(r)), nil
		}
	}
	return nil, &TypeMismatchError{
		Expected: "numeric type",
		Got:      lhs.Type().String() + " - " + rhs.Type().String(),
		Span:     span,
	}
}

// mulValues implements the * operator.
func mulValues(lhs, rhs Value, span syntax.Span) (Value, error) {
	switch l := lhs.(type) {
	case IntValue:
		if r, ok := rhs.(IntValue); ok {
			return Int(int64(l) * int64(r)), nil
		}
		if r, ok := rhs.(FloatValue); ok {
			return Float(float64(l) * float64(r)), nil
		}
	case FloatValue:
		if r, ok := rhs.(IntValue); ok {
			return Float(float64(l) * float64(r)), nil
		}
		if r, ok := rhs.(FloatValue); ok {
			return Float(float64(l) * float64(r)), nil
		}
	}
	return nil, &TypeMismatchError{
		Expected: "numeric type",
		Got:      lhs.Type().String() + " * " + rhs.Type().String(),
		Span:     span,
	}
}

// divValues implements the / operator.
func divValues(lhs, rhs Value, span syntax.Span) (Value, error) {
	switch l := lhs.(type) {
	case IntValue:
		if r, ok := rhs.(IntValue); ok {
			if r == 0 {
				return nil, &DivisionByZeroError{Span: span}
			}
			return Float(float64(l) / float64(r)), nil
		}
		if r, ok := rhs.(FloatValue); ok {
			if r == 0 {
				return nil, &DivisionByZeroError{Span: span}
			}
			return Float(float64(l) / float64(r)), nil
		}
	case FloatValue:
		if r, ok := rhs.(IntValue); ok {
			if r == 0 {
				return nil, &DivisionByZeroError{Span: span}
			}
			return Float(float64(l) / float64(r)), nil
		}
		if r, ok := rhs.(FloatValue); ok {
			if r == 0 {
				return nil, &DivisionByZeroError{Span: span}
			}
			return Float(float64(l) / float64(r)), nil
		}
	}
	return nil, &TypeMismatchError{
		Expected: "numeric type",
		Got:      lhs.Type().String() + " / " + rhs.Type().String(),
		Span:     span,
	}
}

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

// compareValues compares two values using the given comparison function.
func compareValues(lhs, rhs Value, span syntax.Span, cmp func(int) bool) (Value, error) {
	// Int comparison
	if l, ok := lhs.(IntValue); ok {
		if r, ok := rhs.(IntValue); ok {
			if l < r {
				return Bool(cmp(-1)), nil
			} else if l > r {
				return Bool(cmp(1)), nil
			}
			return Bool(cmp(0)), nil
		}
	}
	// Float comparison
	if l, lok := AsFloat(lhs); lok {
		if r, rok := AsFloat(rhs); rok {
			if l < r {
				return Bool(cmp(-1)), nil
			} else if l > r {
				return Bool(cmp(1)), nil
			}
			return Bool(cmp(0)), nil
		}
	}
	// String comparison
	if l, ok := lhs.(StrValue); ok {
		if r, ok := rhs.(StrValue); ok {
			if string(l) < string(r) {
				return Bool(cmp(-1)), nil
			} else if string(l) > string(r) {
				return Bool(cmp(1)), nil
			}
			return Bool(cmp(0)), nil
		}
	}
	return nil, &TypeMismatchError{
		Expected: "comparable types",
		Got:      lhs.Type().String() + " and " + rhs.Type().String(),
		Span:     span,
	}
}

// andValues implements the and operator.
func andValues(lhs, rhs Value, span syntax.Span) (Value, error) {
	l, ok := lhs.(BoolValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "bool", Got: lhs.Type().String(), Span: span}
	}
	r, ok := rhs.(BoolValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "bool", Got: rhs.Type().String(), Span: span}
	}
	return Bool(bool(l) && bool(r)), nil
}

// orValues implements the or operator.
func orValues(lhs, rhs Value, span syntax.Span) (Value, error) {
	l, ok := lhs.(BoolValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "bool", Got: lhs.Type().String(), Span: span}
	}
	r, ok := rhs.(BoolValue)
	if !ok {
		return nil, &TypeMismatchError{Expected: "bool", Got: rhs.Type().String(), Span: span}
	}
	return Bool(bool(l) || bool(r)), nil
}

// evalUnaryExpr evaluates a unary expression.
func evalUnaryExpr(vm *Vm, expr *syntax.UnaryExpr) (Value, error) {
	operand, err := evalSimpleExpr(vm, expr.Expr())
	if err != nil {
		return nil, err
	}

	op := expr.Op()
	switch op {
	case syntax.UnOpPos:
		// Unary plus - just return the value for numeric types
		switch operand.(type) {
		case IntValue, FloatValue:
			return operand, nil
		}
	case syntax.UnOpNeg:
		// Unary minus - negate numeric values
		switch v := operand.(type) {
		case IntValue:
			return Int(-int64(v)), nil
		case FloatValue:
			return Float(-float64(v)), nil
		}
	case syntax.UnOpNot:
		// Logical not
		if b, ok := operand.(BoolValue); ok {
			return Bool(!bool(b)), nil
		}
	}

	return nil, &TypeMismatchError{
		Expected: "numeric or bool",
		Got:      operand.Type().String(),
		Span:     expr.ToUntyped().Span(),
	}
}

// evalLetBinding evaluates a let binding expression.
func evalLetBinding(vm *Vm, expr *syntax.LetBindingExpr) (Value, error) {
	if expr.BindingKind() == syntax.LetBindingClosure {
		// Closure binding: let f(x) = ...
		init := expr.Init()
		if init == nil {
			return None, nil
		}
		val, err := evalSimpleExpr(vm, init)
		if err != nil {
			return nil, err
		}
		// Get the closure and bind it
		if closure, ok := init.(*syntax.ClosureExpr); ok {
			if name := closure.Name(); name != nil {
				vm.Define(name.Get(), val)
			}
		}
		return None, nil
	}

	// Plain binding: let x = ...
	pattern := expr.Pattern()
	init := expr.Init()

	if init == nil {
		return None, nil
	}

	val, err := evalSimpleExpr(vm, init)
	if err != nil {
		return nil, err
	}

	// For simple identifier pattern
	if ident, ok := pattern.(*syntax.NormalPattern); ok {
		vm.Define(ident.Name(), val)
		return None, nil
	}

	return None, nil
}

// destructure binds a value to a destructuring pattern.
func destructure(vm *Vm, pattern *syntax.DestructuringNode, value Value) error {
	// For now, only support simple array destructuring
	arr, ok := value.(ArrayValue)
	if !ok {
		return &TypeMismatchError{
			Expected: "array",
			Got:      value.Type().String(),
			Span:     pattern.ToUntyped().Span(),
		}
	}

	// Convert to DestructuringPattern to access items
	destPattern := &syntax.DestructuringPattern{}
	_ = destPattern // Use the pattern items from the node
	items := pattern.Items()
	if len(items) > len(arr) {
		return &DestructuringError{
			Message: "not enough values to unpack",
			Span:    pattern.ToUntyped().Span(),
		}
	}

	for i, item := range items {
		if err := destructureItem(vm, item, arr[i]); err != nil {
			return err
		}
	}

	return nil
}

// destructureItem binds a single value to a destructuring item.
func destructureItem(vm *Vm, item syntax.DestructuringItem, value Value) error {
	switch it := item.(type) {
	case *syntax.DestructuringBinding:
		return bindPattern(vm, it.Pattern(), value)
	case *syntax.DestructuringNamed:
		// For named destructuring like {a: x}
		return bindPattern(vm, it.Pattern(), value)
	case *syntax.DestructuringSpread:
		// Spread should collect remaining items
		if sink := it.Sink(); sink != nil {
			return bindPattern(vm, sink, value)
		}
	}
	return nil
}

// bindPattern binds a value to a pattern.
func bindPattern(vm *Vm, pattern syntax.Pattern, value Value) error {
	switch p := pattern.(type) {
	case *syntax.NormalPattern:
		vm.Define(p.Name(), value)
		return nil
	case *syntax.PlaceholderPattern:
		// Placeholder pattern discards the value
		return nil
	case *syntax.ParenthesizedPattern:
		return bindPattern(vm, p.Pattern(), value)
	case *syntax.DestructuringPattern:
		arr, ok := value.(ArrayValue)
		if !ok {
			return &TypeMismatchError{
				Expected: "array",
				Got:      value.Type().String(),
				Span:     p.ToUntyped().Span(),
			}
		}
		items := p.Items()
		if len(items) > len(arr) {
			return &DestructuringError{
				Message: "not enough values to unpack",
				Span:    p.ToUntyped().Span(),
			}
		}
		for i, item := range items {
			if err := destructureItem(vm, item, arr[i]); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
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
	return "invalid assignment target"
}

// DivisionByZeroError is returned when dividing by zero.
type DivisionByZeroError struct {
	Span syntax.Span
}

func (e *DivisionByZeroError) Error() string {
	return "division by zero"
}

// DestructuringError is returned when destructuring fails.
type DestructuringError struct {
	Message string
	Span    syntax.Span
}

func (e *DestructuringError) Error() string {
	return e.Message
}

