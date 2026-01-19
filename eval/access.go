package eval

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// Access provides mutable access to lvalues during evaluation.
//
// This interface is implemented by expression types that can appear on the
// left side of an assignment. It allows the evaluator to modify values in place.
type Access interface {
	// Access returns a pointer to the value that can be mutated.
	Access(vm *Vm) (*Value, error)
}

// AccessExpr attempts to get mutable access to an expression's value.
// Returns an error if the expression is not an assignable lvalue.
func AccessExpr(vm *Vm, expr syntax.Expr) (*Value, error) {
	if expr == nil {
		return nil, &NotAssignableError{Span: syntax.Detached()}
	}

	switch e := expr.(type) {
	case *syntax.IdentExpr:
		return accessIdent(vm, e)
	case *syntax.ParenthesizedExpr:
		return AccessExpr(vm, e.Expr())
	case *syntax.FieldAccessExpr:
		return accessFieldAccess(vm, e)
	case *syntax.FuncCallExpr:
		return accessFuncCall(vm, e)
	default:
		// Try to evaluate as a temporary value - this will produce an error
		return nil, &NotAssignableError{Span: expr.ToUntyped().Span()}
	}
}

// accessIdent provides mutable access to an identifier binding.
func accessIdent(vm *Vm, ident *syntax.IdentExpr) (*Value, error) {
	name := ident.Get()
	span := ident.ToUntyped().Span()

	// Check if inspected
	if vm.Inspected != nil && *vm.Inspected == span {
		if binding := vm.Scopes.Get(name); binding != nil {
			v, _ := binding.Read()
			vm.Trace(v, span)
		}
	}

	// Get mutable binding
	binding := vm.Scopes.GetMut(name)
	if binding == nil {
		return nil, &UndefinedVariableError{Name: name, Span: span}
	}

	// Check if mutable
	if !binding.Mutable {
		return nil, &ImmutableBindingError{}
	}

	return &binding.Value, nil
}

// accessFieldAccess provides mutable access to a dictionary field.
func accessFieldAccess(vm *Vm, access *syntax.FieldAccessExpr) (*Value, error) {
	dict, err := accessDict(vm, access)
	if err != nil {
		return nil, err
	}

	field := access.Field()
	if field == nil {
		return nil, &MissingFieldError{
			Span: access.ToUntyped().Span(),
		}
	}

	fieldName := field.Get()
	span := access.ToUntyped().Span()

	// Get or create the field
	val, ok := dict.Get(fieldName)
	if !ok {
		return nil, &KeyNotFoundError{Key: fieldName, Span: span}
	}

	// To allow mutation, we need to return a pointer to the value.
	// We set it back immediately to get a pointer, then return.
	dict.Set(fieldName, val)
	for i := range dict.entries {
		if dict.entries[i].Key == fieldName {
			return &dict.entries[i].Value, nil
		}
	}

	return nil, &KeyNotFoundError{Key: fieldName, Span: span}
}

// accessFuncCall handles accessor method calls like array.at(i).
func accessFuncCall(vm *Vm, call *syntax.FuncCallExpr) (*Value, error) {
	callee := call.Callee()
	if callee == nil {
		return nil, &NotAssignableError{Span: call.ToUntyped().Span()}
	}

	// Check if this is a method call on a field access
	fieldAccess, ok := callee.(*syntax.FieldAccessExpr)
	if !ok {
		return nil, &NotAssignableError{Span: call.ToUntyped().Span()}
	}

	methodName := ""
	if field := fieldAccess.Field(); field != nil {
		methodName = field.Get()
	}

	// Check if this is an accessor method
	if !isAccessorMethod(methodName) {
		return nil, &NotAssignableError{Span: call.ToUntyped().Span()}
	}

	// Get the target value mutably
	targetPtr, err := AccessExpr(vm, fieldAccess.Target())
	if err != nil {
		return nil, err
	}

	// Handle accessor methods
	switch methodName {
	case "at":
		return callAtAccess(vm, call, targetPtr)
	case "first":
		return callFirstAccess(targetPtr, call.ToUntyped().Span())
	case "last":
		return callLastAccess(targetPtr, call.ToUntyped().Span())
	}

	return nil, &NotAssignableError{Span: call.ToUntyped().Span()}
}

// accessDict gets mutable access to a dictionary value.
func accessDict(vm *Vm, access *syntax.FieldAccessExpr) (*DictValue, error) {
	targetPtr, err := AccessExpr(vm, access.Target())
	if err != nil {
		return nil, err
	}

	target := *targetPtr
	span := access.Target().ToUntyped().Span()

	switch v := target.(type) {
	case DictValue:
		// We need to update the binding to use a pointer so we can mutate it
		dictPtr := &v
		*targetPtr = dictPtr
		return dictPtr, nil
	case *DictValue:
		return v, nil
	default:
		ty := target.Type()

		// Check for types that cannot have mutable fields
		switch target.(type) {
		case SymbolValue, ContentValue, ModuleValue, FuncValue:
			return nil, &CannotMutateFieldsError{Type: ty, Span: span}
		}

		return nil, &FieldsNotMutableError{Type: ty, Span: span}
	}
}

// isAccessorMethod returns true if the method name is a known accessor method.
func isAccessorMethod(name string) bool {
	switch name {
	case "at", "first", "last":
		return true
	}
	return false
}

// callAtAccess handles the .at() accessor method for arrays and dicts.
func callAtAccess(vm *Vm, call *syntax.FuncCallExpr, targetPtr *Value) (*Value, error) {
	span := call.ToUntyped().Span()
	args := call.Args()
	if args == nil {
		return nil, &MissingArgumentError{What: "index", Span: span}
	}

	items := args.Items()
	if len(items) == 0 {
		return nil, &MissingArgumentError{What: "index", Span: span}
	}

	// Get the index argument
	firstArg, ok := items[0].(*syntax.PosArg)
	if !ok {
		return nil, &InvalidArgumentError{Message: "expected positional argument", Span: span}
	}

	indexExpr := firstArg.Expr()
	if indexExpr == nil {
		return nil, &MissingArgumentError{What: "index", Span: span}
	}

	// Evaluate the index - we need an eval function here
	// For now, we'll just return an error indicating this needs implementation
	// TODO: Implement eval for expressions
	_ = indexExpr

	target := *targetPtr

	switch v := target.(type) {
	case ArrayValue:
		// Would need to evaluate indexExpr to get the index
		// For now, return an error
		return nil, &NotImplementedError{Feature: "array .at() access", Span: span}
	case DictValue:
		// Would need to evaluate indexExpr to get the key
		return nil, &NotImplementedError{Feature: "dict .at() access", Span: span}
	case *DictValue:
		return nil, &NotImplementedError{Feature: "dict .at() access", Span: span}
	default:
		return nil, &TypeMismatchError{
			Expected: "array or dictionary",
			Got:      v.Type().String(),
			Span:     span,
		}
	}
}

// callFirstAccess handles the .first() accessor method for arrays.
func callFirstAccess(targetPtr *Value, span syntax.Span) (*Value, error) {
	target := *targetPtr

	arr, ok := target.(ArrayValue)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "array",
			Got:      target.Type().String(),
			Span:     span,
		}
	}

	if len(arr) == 0 {
		return nil, &IndexOutOfBoundsError{Index: 0, Length: 0, Span: span}
	}

	// Return pointer to first element
	return &arr[0], nil
}

// callLastAccess handles the .last() accessor method for arrays.
func callLastAccess(targetPtr *Value, span syntax.Span) (*Value, error) {
	target := *targetPtr

	arr, ok := target.(ArrayValue)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "array",
			Got:      target.Type().String(),
			Span:     span,
		}
	}

	if len(arr) == 0 {
		return nil, &IndexOutOfBoundsError{Index: 0, Length: 0, Span: span}
	}

	// Return pointer to last element
	return &arr[len(arr)-1], nil
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// NotAssignableError is returned when attempting to assign to a non-lvalue.
type NotAssignableError struct {
	Span syntax.Span
}

func (e *NotAssignableError) Error() string {
	return "cannot mutate a temporary value"
}

// MissingFieldError is returned when a field access has no field name.
type MissingFieldError struct {
	Span syntax.Span
}

func (e *MissingFieldError) Error() string {
	return "missing field name"
}

// KeyNotFoundError is returned when a dictionary key is not found.
type KeyNotFoundError struct {
	Key  string
	Span syntax.Span
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("key not found: %s", e.Key)
}

// CannotMutateFieldsError is returned when trying to mutate fields on certain types.
type CannotMutateFieldsError struct {
	Type Type
	Span syntax.Span
}

func (e *CannotMutateFieldsError) Error() string {
	return fmt.Sprintf("cannot mutate fields on %s", e.Type)
}

// FieldsNotMutableError is returned when field mutation is not yet supported.
type FieldsNotMutableError struct {
	Type Type
	Span syntax.Span
}

func (e *FieldsNotMutableError) Error() string {
	return fmt.Sprintf("fields on %s are not yet mutable", e.Type)
}

// InvalidArgumentError is returned when an argument is invalid.
type InvalidArgumentError struct {
	Message string
	Span    syntax.Span
}

func (e *InvalidArgumentError) Error() string {
	return e.Message
}

// NotImplementedError is returned for features not yet implemented.
type NotImplementedError struct {
	Feature string
	Span    syntax.Span
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("not implemented: %s", e.Feature)
}

// IndexOutOfBoundsError is returned when an array index is out of bounds.
type IndexOutOfBoundsError struct {
	Index  int
	Length int
	Span   syntax.Span
}

func (e *IndexOutOfBoundsError) Error() string {
	return fmt.Sprintf("index %d is out of bounds for array of length %d", e.Index, e.Length)
}
