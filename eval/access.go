// Mutable access to lvalues during evaluation.
// Translated from typst-eval/src/access.rs

package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Error Infrastructure (matches Rust's bail!/error!/.at() pattern)
// ----------------------------------------------------------------------------

// SpannedError wraps an error with a source span.
// Equivalent to Rust's .at(span) pattern.
type SpannedError struct {
	Err     error
	ErrSpan syntax.Span
}

func (e *SpannedError) Error() string {
	return e.Err.Error()
}

func (e *SpannedError) Span() syntax.Span {
	return e.ErrSpan
}

func (e *SpannedError) Unwrap() error {
	return e.Err
}

// atSpan wraps an error with span information.
// Equivalent to Rust's .at(span)
func atSpan(err error, span syntax.Span) error {
	if err == nil {
		return nil
	}
	return &SpannedError{Err: err, ErrSpan: span}
}

// HintedError wraps an error with a hint.
// Equivalent to Rust's .hint() pattern.
type HintedError struct {
	Err   error
	Hints []string
	Span  syntax.Span
}

func (e *HintedError) Error() string {
	return e.Err.Error()
}

func (e *HintedError) Unwrap() error {
	return e.Err
}

// WithHint adds a hint to an error.
func WithHint(err error, span syntax.Span, hint string) *HintedError {
	return &HintedError{Err: err, Hints: []string{hint}, Span: span}
}

// TracedError wraps an error with a tracepoint.
// Equivalent to Rust's .trace(world, point, span)
type TracedError struct {
	Err   error
	Point string
	Span  syntax.Span
}

func (e *TracedError) Error() string {
	return e.Err.Error()
}

func (e *TracedError) Unwrap() error {
	return e.Err
}

// ----------------------------------------------------------------------------
// Access Trait Implementation
// ----------------------------------------------------------------------------

// AccessExpr accesses an expression mutably.
// Matches Rust: impl Access for ast::Expr<'_>
func AccessExpr(vm *Vm, expr syntax.Expr) (*foundations.Value, error) {
	switch e := expr.(type) {
	case *syntax.IdentExpr:
		return accessIdent(vm, e)
	case *syntax.ParenthesizedExpr:
		return accessParenthesized(vm, e)
	case *syntax.FieldAccessExpr:
		return accessFieldAccess(vm, e)
	case *syntax.FuncCallExpr:
		return accessFuncCall(vm, e)
	default:
		// Matches: let _ = self.eval(vm)?; bail!(self.span(), "cannot mutate a temporary value");
		_, _ = evalExpr(vm, expr)
		return nil, atSpan(fmt.Errorf("cannot mutate a temporary value"), expr.ToUntyped().Span())
	}
}

// accessIdent accesses an identifier mutably.
// Matches Rust: impl Access for ast::Ident<'_>
func accessIdent(vm *Vm, ident *syntax.IdentExpr) (*foundations.Value, error) {
	span := ident.ToUntyped().Span()
	name := ident.Get()

	// Matches: if vm.inspected == Some(span) && let Ok(binding) = vm.scopes.get(&self)
	if vm.Inspected != nil && *vm.Inspected == span {
		if binding := vm.Scopes.Get(name); binding != nil {
			vm.Trace(binding.Read())
		}
	}

	// Matches: vm.scopes.get_mut(&self).and_then(|b| b.write().map_err(Into::into)).at(span)
	binding := vm.Scopes.GetMut(name)
	if binding == nil {
		return nil, atSpan(fmt.Errorf("unknown variable: %s", name), span)
	}

	slot, err := binding.Slot()
	if err != nil {
		// Convert CapturedVariableError to proper error message
		// Matches: .map_err(Into::into)
		if capErr, ok := err.(*foundations.CapturedVariableError); ok {
			return nil, atSpan(fmt.Errorf("variables from outside the %s are read-only and cannot be modified", capErr.Capturer), span)
		}
		return nil, atSpan(err, span)
	}

	return slot, nil
}

// accessParenthesized accesses a parenthesized expression mutably.
// Matches Rust: impl Access for ast::Parenthesized<'_>
func accessParenthesized(vm *Vm, paren *syntax.ParenthesizedExpr) (*foundations.Value, error) {
	return AccessExpr(vm, paren.Expr())
}

// accessFieldAccess accesses a field mutably.
// Matches Rust: impl Access for ast::FieldAccess<'_>
func accessFieldAccess(vm *Vm, access *syntax.FieldAccessExpr) (*foundations.Value, error) {
	// Matches: access_dict(vm, self)?.at_mut(self.field().get()).at(self.span())
	dict, err := accessDict(vm, access)
	if err != nil {
		return nil, err
	}

	field := access.Field()
	if field == nil {
		return nil, atSpan(fmt.Errorf("missing field name"), access.ToUntyped().Span())
	}

	fieldName := field.Get()
	ptr, err := dict.AtMut(fieldName)
	if err != nil {
		return nil, atSpan(fmt.Errorf("dictionary does not contain key %q", fieldName), access.ToUntyped().Span())
	}
	return ptr, nil
}

// accessFuncCall accesses via an accessor method call (e.g., array.at(i)).
// Matches Rust: impl Access for ast::FuncCall<'_>
func accessFuncCall(vm *Vm, call *syntax.FuncCallExpr) (*foundations.Value, error) {
	span := call.ToUntyped().Span()

	// Matches: if let ast::Expr::FieldAccess(access) = self.callee()
	callee := call.Callee()
	fieldAccess, ok := callee.(*syntax.FieldAccessExpr)
	if !ok {
		_, _ = evalExpr(vm, call)
		return nil, atSpan(fmt.Errorf("cannot mutate a temporary value"), span)
	}

	// Matches: let method = access.field(); if is_accessor_method(&method)
	method := fieldAccess.Field()
	if method == nil || !IsAccessorMethod(method.Get()) {
		_, _ = evalExpr(vm, call)
		return nil, atSpan(fmt.Errorf("cannot mutate a temporary value"), span)
	}

	// Matches: let args = self.args().eval(vm)?.spanned(span);
	args, err := evalArgs(vm, call.Args())
	if err != nil {
		return nil, err
	}
	args.Span = span

	// Matches: let value = access.target().access(vm)?;
	value, err := AccessExpr(vm, fieldAccess.Target())
	if err != nil {
		return nil, err
	}

	// Matches: let result = call_method_access(value, &method, args, span);
	// with tracing: result.trace(world, point, span)
	result, err := CallMethodAccess(value, method.Get(), args, span)
	if err != nil {
		return nil, &TracedError{
			Err:   err,
			Point: fmt.Sprintf("call to %s", method.Get()),
			Span:  span,
		}
	}
	return result, nil
}

// accessDict gets mutable access to a dictionary from a field access target.
// Matches Rust: pub(crate) fn access_dict
func accessDict(vm *Vm, access *syntax.FieldAccessExpr) (*foundations.Dict, error) {
	// Matches: match access.target().access(vm)?
	value, err := AccessExpr(vm, access.Target())
	if err != nil {
		return nil, err
	}

	span := access.Target().ToUntyped().Span()

	// Matches: Value::Dict(dict) => Ok(dict)
	if dict, ok := (*value).(*foundations.Dict); ok {
		return dict, nil
	}

	ty := (*value).Type()

	// Matches: if matches!(value, Value::Symbol(_) | Value::Content(_) | Value::Module(_) | Value::Func(_))
	switch (*value).(type) {
	case foundations.SymbolValue, foundations.ContentValue, foundations.ModuleValue, foundations.FuncValue:
		return nil, atSpan(fmt.Errorf("cannot mutate fields on %s", ty), span)
	}

	// Matches: else if typst_library::foundations::fields_on(ty).is_empty()
	if !foundations.HasFields(ty) {
		return nil, atSpan(fmt.Errorf("%s does not have accessible fields", ty), span)
	}

	// Matches: Err(...).hint(...)
	return nil, WithHint(
		fmt.Errorf("fields on %s are not yet mutable", ty),
		span,
		fmt.Sprintf("try creating a new %s with the updated field value instead", ty),
	)
}
