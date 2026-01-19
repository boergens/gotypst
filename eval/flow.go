package eval

import "github.com/boergens/gotypst/syntax"

// MaxIterations is the maximum number of loop iterations allowed.
const MaxIterations = 10_000

// FlowEvent represents a control flow event during evaluation.
//
// When a break, continue, or return statement is executed, a FlowEvent
// is stored in the VM's flow field. This event propagates up the call stack
// until it is handled by the appropriate construct (loop or function).
type FlowEvent interface {
	// Span returns the source location where this flow event was triggered.
	Span() syntax.Span

	// isFlowEvent is a marker method to seal the interface.
	isFlowEvent()
}

// BreakEvent represents a loop break statement.
type BreakEvent struct {
	span syntax.Span
}

func (e BreakEvent) Span() syntax.Span { return e.span }
func (e BreakEvent) isFlowEvent()      {}

// NewBreakEvent creates a new break event.
func NewBreakEvent(span syntax.Span) BreakEvent {
	return BreakEvent{span: span}
}

// ContinueEvent represents a loop continue statement.
type ContinueEvent struct {
	span syntax.Span
}

func (e ContinueEvent) Span() syntax.Span { return e.span }
func (e ContinueEvent) isFlowEvent()      {}

// NewContinueEvent creates a new continue event.
func NewContinueEvent(span syntax.Span) ContinueEvent {
	return ContinueEvent{span: span}
}

// ReturnEvent represents a function return statement.
type ReturnEvent struct {
	span syntax.Span

	// Value is the optional return value (nil for bare `return`).
	Value Value

	// Conditional indicates whether this was a conditional return.
	// Conditional returns don't produce warnings for missing returns.
	Conditional bool
}

func (e ReturnEvent) Span() syntax.Span { return e.span }
func (e ReturnEvent) isFlowEvent()      {}

// NewReturnEvent creates a new return event without a value.
func NewReturnEvent(span syntax.Span) ReturnEvent {
	return ReturnEvent{span: span, Value: nil, Conditional: false}
}

// NewReturnEventWithValue creates a new return event with a value.
func NewReturnEventWithValue(span syntax.Span, value Value) ReturnEvent {
	return ReturnEvent{span: span, Value: value, Conditional: false}
}

// NewConditionalReturnEvent creates a new conditional return event.
func NewConditionalReturnEvent(span syntax.Span, value Value) ReturnEvent {
	return ReturnEvent{span: span, Value: value, Conditional: true}
}

// IsBreak returns true if the flow event is a break.
func IsBreak(e FlowEvent) bool {
	_, ok := e.(BreakEvent)
	return ok
}

// IsContinue returns true if the flow event is a continue.
func IsContinue(e FlowEvent) bool {
	_, ok := e.(ContinueEvent)
	return ok
}

// IsReturn returns true if the flow event is a return.
func IsReturn(e FlowEvent) bool {
	_, ok := e.(ReturnEvent)
	return ok
}

// IsLoopFlow returns true if the flow event is break or continue.
func IsLoopFlow(e FlowEvent) bool {
	return IsBreak(e) || IsContinue(e)
}

// ForbiddenFlowError is returned when a flow event is used in a forbidden context.
type ForbiddenFlowError struct {
	Event FlowEvent
	Span  syntax.Span
}

func (e *ForbiddenFlowError) Error() string {
	switch e.Event.(type) {
	case BreakEvent:
		return "break is not allowed here"
	case ContinueEvent:
		return "continue is not allowed here"
	case ReturnEvent:
		return "return is not allowed here"
	default:
		return "flow event is not allowed here"
	}
}

// CheckForbiddenFlow checks if a flow event is forbidden in the current context.
// Returns an error if the flow event is not allowed.
func CheckForbiddenFlow(e FlowEvent, allowBreak, allowContinue, allowReturn bool) error {
	if e == nil {
		return nil
	}

	switch e.(type) {
	case BreakEvent:
		if !allowBreak {
			return &ForbiddenFlowError{Event: e, Span: e.Span()}
		}
	case ContinueEvent:
		if !allowContinue {
			return &ForbiddenFlowError{Event: e, Span: e.Span()}
		}
	case ReturnEvent:
		if !allowReturn {
			return &ForbiddenFlowError{Event: e, Span: e.Span()}
		}
	}
	return nil
}

// MarkReturnAsConditional marks a return event as conditional if present.
// This is called after evaluating conditional/loop bodies to indicate
// that the return may not always execute.
func MarkReturnAsConditional(vm *Vm) {
	if ret, ok := vm.Flow.(ReturnEvent); ok {
		ret.Conditional = true
		vm.Flow = ret
	}
}

// InfiniteLoopError is returned when an infinite loop is detected.
type InfiniteLoopError struct {
	Span    syntax.Span
	Message string
}

func (e *InfiniteLoopError) Error() string {
	return e.Message
}

// isInvariant checks if an expression always evaluates to the same value.
// This is used to detect infinite loops with constant conditions.
func isInvariant(node *syntax.SyntaxNode) bool {
	if node == nil {
		return true
	}

	kind := node.Kind()

	// Identifiers and math identifiers can change (variables)
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

// canDiverge checks if an expression contains a break or return that could
// exit the loop early. This is used alongside isInvariant to detect infinite loops.
func canDiverge(node *syntax.SyntaxNode) bool {
	if node == nil {
		return false
	}

	kind := node.Kind()

	// Break and return can exit early
	if kind == syntax.Break || kind == syntax.Return ||
		kind == syntax.LoopBreak || kind == syntax.FuncReturn {
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
