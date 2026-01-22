// Control flow evaluation for Typst.
// Translated from typst-eval/src/flow.rs

package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// MaxIterations is the maximum number of loop iterations.
const MaxIterations = 10_000

// ----------------------------------------------------------------------------
// Flow Events
// ----------------------------------------------------------------------------

// FlowEvent represents a control flow event that occurred during evaluation.
// Matches Rust: pub enum FlowEvent
type FlowEvent interface {
	// Span returns the source location where this flow event was triggered.
	Span() syntax.Span

	// Forbidden returns an error stating that this control flow is forbidden.
	Forbidden() error

	isFlowEvent()
}

// BreakEvent represents stopping iteration in a loop.
// Matches Rust: FlowEvent::Break(Span)
type BreakEvent struct {
	span syntax.Span
}

func (e BreakEvent) Span() syntax.Span { return e.span }
func (e BreakEvent) isFlowEvent()      {}
func (e BreakEvent) Forbidden() error {
	return fmt.Errorf("cannot break outside of loop")
}

// ContinueEvent represents skipping the remainder of the current iteration.
// Matches Rust: FlowEvent::Continue(Span)
type ContinueEvent struct {
	span syntax.Span
}

func (e ContinueEvent) Span() syntax.Span { return e.span }
func (e ContinueEvent) isFlowEvent()      {}
func (e ContinueEvent) Forbidden() error {
	return fmt.Errorf("cannot continue outside of loop")
}

// ReturnEvent represents stopping execution of a function early.
// Matches Rust: FlowEvent::Return(Span, Option<Value>, bool)
type ReturnEvent struct {
	span syntax.Span

	// Value is the optional return value (nil for bare `return`).
	Value foundations.Value

	// Conditional indicates whether the return was conditional.
	// Conditional returns don't produce warnings for discarding content.
	Conditional bool
}

func (e ReturnEvent) Span() syntax.Span { return e.span }
func (e ReturnEvent) isFlowEvent()      {}
func (e ReturnEvent) Forbidden() error {
	return fmt.Errorf("cannot return outside of function")
}

// NewBreakEvent creates a new break event.
func NewBreakEvent(span syntax.Span) FlowEvent {
	return BreakEvent{span: span}
}

// NewContinueEvent creates a new continue event.
func NewContinueEvent(span syntax.Span) FlowEvent {
	return ContinueEvent{span: span}
}

// NewReturnEvent creates a new return event with no value.
func NewReturnEvent(span syntax.Span) FlowEvent {
	return ReturnEvent{span: span}
}

// CheckForbiddenFlow checks if a flow event is forbidden in the current context.
// Returns nil if the flow is allowed, otherwise returns the forbidden error.
func CheckForbiddenFlow(flow FlowEvent, allowBreak, allowContinue, allowReturn bool) error {
	if flow == nil {
		return nil
	}
	switch flow.(type) {
	case BreakEvent:
		if allowBreak {
			return nil
		}
	case ContinueEvent:
		if allowContinue {
			return nil
		}
	case ReturnEvent:
		if allowReturn {
			return nil
		}
	}
	return flow.Forbidden()
}

// ----------------------------------------------------------------------------
// Conditional Evaluation
// ----------------------------------------------------------------------------

// evalConditional evaluates a conditional expression (if/else).
// Matches Rust: impl Eval for ast::Conditional
func evalConditional(vm *Vm, e *syntax.ConditionalExpr) (foundations.Value, error) {
	condition := e.Condition()
	condValue, err := evalExpr(vm, condition)
	if err != nil {
		return nil, err
	}

	condBool, ok := condValue.(foundations.Bool)
	if !ok {
		return nil, &TypeError{
			Expected: foundations.TypeBool,
			Got:      condValue.Type(),
			Span:     condition.ToUntyped().Span(),
		}
	}

	var output foundations.Value
	if condBool {
		ifBody := e.IfBody()
		if ifBody != nil {
			output, err = evalExpr(vm, ifBody)
			if err != nil {
				return nil, err
			}
		} else {
			output = foundations.None
		}
	} else if elseBody := e.ElseBody(); elseBody != nil {
		output, err = evalExpr(vm, elseBody)
		if err != nil {
			return nil, err
		}
	} else {
		output = foundations.None
	}

	// Mark the return as conditional
	if ret, ok := vm.Flow.(ReturnEvent); ok {
		ret.Conditional = true
		vm.Flow = ret
	}

	return output, nil
}

// ----------------------------------------------------------------------------
// While Loop Evaluation
// ----------------------------------------------------------------------------

// evalWhileLoop evaluates a while loop.
// Matches Rust: impl Eval for ast::WhileLoop
func evalWhileLoop(vm *Vm, e *syntax.WhileLoopExpr) (foundations.Value, error) {
	flow := vm.TakeFlow()
	var output foundations.Value = foundations.None
	i := 0

	condition := e.Condition()
	body := e.Body()

	for {
		// Evaluate condition
		condValue, err := evalExpr(vm, condition)
		if err != nil {
			return nil, err
		}

		condBool, ok := condValue.(foundations.Bool)
		if !ok {
			return nil, &TypeError{
				Expected: foundations.TypeBool,
				Got:      condValue.Type(),
				Span:     condition.ToUntyped().Span(),
			}
		}

		if !condBool {
			break
		}

		// Check for infinite loop
		if i == 0 && isInvariant(condition.ToUntyped()) && !canDiverge(body.ToUntyped()) {
			return nil, fmt.Errorf("condition is always true")
		} else if i >= MaxIterations {
			return nil, fmt.Errorf("loop seems to be infinite")
		}

		// Evaluate body
		value, err := evalExpr(vm, body)
		if err != nil {
			return nil, err
		}

		output, err = Join(output, value)
		if err != nil {
			return nil, wrapErrorAt(err, body.ToUntyped().Span())
		}

		// Handle flow events
		switch vm.Flow.(type) {
		case BreakEvent:
			vm.Flow = nil
			goto done
		case ContinueEvent:
			vm.Flow = nil
		case ReturnEvent:
			goto done
		}

		i++
	}

done:
	if flow != nil {
		vm.Flow = flow
	}

	// Mark the return as conditional
	if ret, ok := vm.Flow.(ReturnEvent); ok {
		ret.Conditional = true
		vm.Flow = ret
	}

	return output, nil
}

// ----------------------------------------------------------------------------
// For Loop Evaluation
// ----------------------------------------------------------------------------

// evalForLoop evaluates a for loop.
// Matches Rust: impl Eval for ast::ForLoop
func evalForLoop(vm *Vm, e *syntax.ForLoopExpr) (foundations.Value, error) {
	flow := vm.TakeFlow()
	var output foundations.Value = foundations.None

	pattern := e.Pattern()
	iterableExpr := e.Iter()
	if iterableExpr == nil {
		return foundations.None, nil
	}

	iterable, err := evalExpr(vm, iterableExpr)
	if err != nil {
		return nil, err
	}
	iterableType := iterable.Type()

	// Helper to run the loop body
	runBody := func() (shouldBreak bool, err error) {
		body := e.Body()
		if body == nil {
			return false, nil
		}

		value, err := evalExpr(vm, body)
		if err != nil {
			return false, err
		}

		output, err = Join(output, value)
		if err != nil {
			return false, wrapErrorAt(err, body.ToUntyped().Span())
		}

		// Handle flow events
		switch vm.Flow.(type) {
		case BreakEvent:
			vm.Flow = nil
			return true, nil
		case ContinueEvent:
			vm.Flow = nil
			return false, nil
		case ReturnEvent:
			return true, nil
		}

		return false, nil
	}

	vm.Scopes.Enter()

	switch v := iterable.(type) {
	case *foundations.Array:
		// Iterate over values of array
		for i := 0; i < v.Len(); i++ {
			if err := destructure(vm, pattern, v.At(i)); err != nil {
				vm.Scopes.Exit()
				return nil, err
			}
			if shouldBreak, err := runBody(); err != nil {
				vm.Scopes.Exit()
				return nil, err
			} else if shouldBreak {
				break
			}
		}

	case *foundations.Dict:
		// Iterate over key-value pairs of dict
		for _, key := range v.Keys() {
			val, _ := v.Get(key)
			pair := foundations.NewArray(foundations.Str(key), val)
			if err := destructure(vm, pattern, pair); err != nil {
				vm.Scopes.Exit()
				return nil, err
			}
			if shouldBreak, err := runBody(); err != nil {
				vm.Scopes.Exit()
				return nil, err
			} else if shouldBreak {
				break
			}
		}

	case foundations.Str:
		// Check for destructuring pattern on string
		if _, isDestructure := pattern.(*syntax.DestructuringPattern); isDestructure {
			vm.Scopes.Exit()
			return nil, fmt.Errorf("cannot destructure values of %s", iterableType)
		}
		// Iterate over graphemes of string
		for _, grapheme := range graphemes(string(v)) {
			if err := destructure(vm, pattern, foundations.Str(grapheme)); err != nil {
				vm.Scopes.Exit()
				return nil, err
			}
			if shouldBreak, err := runBody(); err != nil {
				vm.Scopes.Exit()
				return nil, err
			} else if shouldBreak {
				break
			}
		}

	case foundations.BytesValue:
		// Check for destructuring pattern on bytes
		if _, isDestructure := pattern.(*syntax.DestructuringPattern); isDestructure {
			vm.Scopes.Exit()
			return nil, fmt.Errorf("cannot destructure values of %s", iterableType)
		}
		// Iterate over the integers of bytes
		for _, b := range v {
			if err := destructure(vm, pattern, foundations.Int(int64(b))); err != nil {
				vm.Scopes.Exit()
				return nil, err
			}
			if shouldBreak, err := runBody(); err != nil {
				vm.Scopes.Exit()
				return nil, err
			} else if shouldBreak {
				break
			}
		}

	default:
		vm.Scopes.Exit()
		return nil, fmt.Errorf("cannot loop over %s", iterableType)
	}

	vm.Scopes.Exit()

	if flow != nil {
		vm.Flow = flow
	}

	// Mark the return as conditional
	if ret, ok := vm.Flow.(ReturnEvent); ok {
		ret.Conditional = true
		vm.Flow = ret
	}

	return output, nil
}

// ----------------------------------------------------------------------------
// Break, Continue, Return
// ----------------------------------------------------------------------------

// evalLoopBreak evaluates a break statement.
// Matches Rust: impl Eval for ast::LoopBreak
func evalLoopBreak(vm *Vm, e *syntax.LoopBreakExpr) (foundations.Value, error) {
	if vm.Flow == nil {
		vm.Flow = BreakEvent{span: e.ToUntyped().Span()}
	}
	return foundations.None, nil
}

// evalLoopContinue evaluates a continue statement.
// Matches Rust: impl Eval for ast::LoopContinue
func evalLoopContinue(vm *Vm, e *syntax.LoopContinueExpr) (foundations.Value, error) {
	if vm.Flow == nil {
		vm.Flow = ContinueEvent{span: e.ToUntyped().Span()}
	}
	return foundations.None, nil
}

// evalFuncReturn evaluates a return statement.
// Matches Rust: impl Eval for ast::FuncReturn
func evalFuncReturn(vm *Vm, e *syntax.FuncReturnExpr) (foundations.Value, error) {
	var value foundations.Value
	if body := e.Body(); body != nil {
		var err error
		value, err = evalExpr(vm, body)
		if err != nil {
			return nil, err
		}
	}

	if vm.Flow == nil {
		vm.Flow = ReturnEvent{
			span:        e.ToUntyped().Span(),
			Value:       value,
			Conditional: false,
		}
	}
	return foundations.None, nil
}

// ----------------------------------------------------------------------------
// Helper Functions
// ----------------------------------------------------------------------------

// isInvariant checks whether the expression always evaluates to the same value.
// Matches Rust: fn is_invariant(expr: &SyntaxNode) -> bool
func isInvariant(node *syntax.SyntaxNode) bool {
	if node == nil {
		return true
	}

	kind := node.Kind()

	// Identifiers can change (variables)
	if kind == syntax.Ident || kind == syntax.MathIdent {
		return false
	}

	// For field access, check if the target is invariant
	if kind == syntax.FieldAccess {
		children := node.Children()
		if len(children) > 0 {
			return isInvariant(children[0])
		}
		return true
	}

	// For function calls, both callee and args must be invariant
	if kind == syntax.FuncCall {
		children := node.Children()
		for _, child := range children {
			if !isInvariant(child) {
				return false
			}
		}
		return true
	}

	// For all other nodes, check all children
	for _, child := range node.Children() {
		if !isInvariant(child) {
			return false
		}
	}
	return true
}

// canDiverge checks whether the expression contains a break or return.
// Matches Rust: fn can_diverge(expr: &SyntaxNode) -> bool
func canDiverge(node *syntax.SyntaxNode) bool {
	if node == nil {
		return false
	}

	kind := node.Kind()

	// Break and return can exit early
	if kind == syntax.Break || kind == syntax.Return {
		return true
	}

	// Recursively check children
	for _, child := range node.Children() {
		if canDiverge(child) {
			return true
		}
	}
	return false
}

// graphemes splits a string into grapheme clusters.
// This is a simplified implementation - a full implementation would use
// the unicode/norm package or a dedicated grapheme library.
func graphemes(s string) []string {
	// For now, use runes as a simple approximation
	// TODO: Use proper grapheme segmentation
	var result []string
	for _, r := range s {
		result = append(result, string(r))
	}
	return result
}
