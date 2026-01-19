package eval

import "github.com/boergens/gotypst/syntax"

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
