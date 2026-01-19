package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

func TestMaxIterations(t *testing.T) {
	if MaxIterations != 10_000 {
		t.Errorf("MaxIterations = %d, want 10000", MaxIterations)
	}
}

func TestMarkReturnAsConditional(t *testing.T) {
	span := syntax.Detached()
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, span)

	// Test with no flow
	MarkReturnAsConditional(vm)
	if vm.HasFlow() {
		t.Error("Expected no flow after marking with no existing flow")
	}

	// Test with break event (should not modify)
	vm.SetFlow(NewBreakEvent(span))
	MarkReturnAsConditional(vm)
	if _, ok := vm.Flow.(BreakEvent); !ok {
		t.Error("Expected break event to remain unchanged")
	}
	vm.ClearFlow()

	// Test with non-conditional return
	vm.SetFlow(NewReturnEventWithValue(span, Int(42)))
	ret := vm.Flow.(ReturnEvent)
	if ret.Conditional {
		t.Error("Expected return to not be conditional initially")
	}

	MarkReturnAsConditional(vm)
	ret = vm.Flow.(ReturnEvent)
	if !ret.Conditional {
		t.Error("Expected return to be marked as conditional")
	}
}

func TestInfiniteLoopError(t *testing.T) {
	err := &InfiniteLoopError{
		Span:    syntax.Detached(),
		Message: "condition is always true",
	}

	if err.Error() != "condition is always true" {
		t.Errorf("Error() = %q, want %q", err.Error(), "condition is always true")
	}
}

func TestIsInvariant(t *testing.T) {
	// Test nil node
	if !isInvariant(nil) {
		t.Error("nil node should be invariant")
	}

	// Test leaf nodes (literals are invariant)
	// We can't easily create AST nodes here without parsing,
	// so we'll test through the Ident kind check
}

func TestCanDiverge(t *testing.T) {
	// Test nil node
	if canDiverge(nil) {
		t.Error("nil node should not diverge")
	}
}

func TestForbiddenFlowError(t *testing.T) {
	span := syntax.Detached()

	tests := []struct {
		name    string
		event   FlowEvent
		wantMsg string
	}{
		{
			name:    "break",
			event:   NewBreakEvent(span),
			wantMsg: "break is not allowed here",
		},
		{
			name:    "continue",
			event:   NewContinueEvent(span),
			wantMsg: "continue is not allowed here",
		},
		{
			name:    "return",
			event:   NewReturnEvent(span),
			wantMsg: "return is not allowed here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ForbiddenFlowError{Event: tt.event, Span: span}
			if err.Error() != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestCheckForbiddenFlow(t *testing.T) {
	span := syntax.Detached()

	// Test nil event (always allowed)
	if err := CheckForbiddenFlow(nil, false, false, false); err != nil {
		t.Errorf("Expected no error for nil event, got %v", err)
	}

	// Test break when allowed
	if err := CheckForbiddenFlow(NewBreakEvent(span), true, false, false); err != nil {
		t.Errorf("Expected no error for allowed break, got %v", err)
	}

	// Test break when not allowed
	if err := CheckForbiddenFlow(NewBreakEvent(span), false, false, false); err == nil {
		t.Error("Expected error for disallowed break")
	}

	// Test continue when allowed
	if err := CheckForbiddenFlow(NewContinueEvent(span), false, true, false); err != nil {
		t.Errorf("Expected no error for allowed continue, got %v", err)
	}

	// Test continue when not allowed
	if err := CheckForbiddenFlow(NewContinueEvent(span), false, false, false); err == nil {
		t.Error("Expected error for disallowed continue")
	}

	// Test return when allowed
	if err := CheckForbiddenFlow(NewReturnEvent(span), false, false, true); err != nil {
		t.Errorf("Expected no error for allowed return, got %v", err)
	}

	// Test return when not allowed
	if err := CheckForbiddenFlow(NewReturnEvent(span), false, false, false); err == nil {
		t.Error("Expected error for disallowed return")
	}
}

func TestConditionalReturnEvent(t *testing.T) {
	span := syntax.Detached()

	// Test NewConditionalReturnEvent
	event := NewConditionalReturnEvent(span, Int(42))
	if !event.Conditional {
		t.Error("Expected conditional flag to be true")
	}
	if event.Value != Int(42) {
		t.Errorf("Expected value Int(42), got %v", event.Value)
	}

	// Test NewReturnEvent
	event2 := NewReturnEvent(span)
	if event2.Conditional {
		t.Error("Expected conditional flag to be false for NewReturnEvent")
	}
	if event2.Value != nil {
		t.Errorf("Expected nil value, got %v", event2.Value)
	}

	// Test NewReturnEventWithValue
	event3 := NewReturnEventWithValue(span, Str("hello"))
	if event3.Conditional {
		t.Error("Expected conditional flag to be false for NewReturnEventWithValue")
	}
	if event3.Value != Str("hello") {
		t.Errorf("Expected value Str(hello), got %v", event3.Value)
	}
}

func TestIterationError(t *testing.T) {
	err := &IterationError{
		Span:    syntax.Detached(),
		Message: "cannot loop over int",
	}

	if err.Error() != "cannot loop over int" {
		t.Errorf("Error() = %q, want %q", err.Error(), "cannot loop over int")
	}
}
