// Destructuring support for Typst bindings.
// Translated from typst-eval/src/binding.rs
//
// The Binding and Scope types are in library/foundations/scope.go

package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// destructure destructures a value into a pattern, creating new bindings.
// This is used for let bindings: let (a, b) = expr
func destructure(vm *Vm, pattern syntax.Pattern, value foundations.Value) error {
	return destructureImpl(vm, pattern, value, func(vm *Vm, expr syntax.Expr, value foundations.Value) error {
		ident, ok := expr.(*syntax.IdentExpr)
		if !ok {
			return &CannotAssignError{Span: expr.ToUntyped().Span()}
		}
		vm.Define(ident, value)
		return nil
	})
}

// destructureAssign destructures a value into a pattern, assigning to existing bindings.
// This is used for destructuring assignments: (a, b) = expr
func destructureAssign(vm *Vm, pattern syntax.Pattern, value foundations.Value) error {
	return destructureImpl(vm, pattern, value, func(vm *Vm, expr syntax.Expr, value foundations.Value) error {
		location, err := AccessExpr(vm, expr)
		if err != nil {
			return err
		}
		*location = value
		return nil
	})
}

// bindingFunc is a callback function used during destructuring.
type bindingFunc func(vm *Vm, expr syntax.Expr, value foundations.Value) error

// destructureImpl is the core recursive destructuring function.
func destructureImpl(vm *Vm, pattern syntax.Pattern, value foundations.Value, f bindingFunc) error {
	if pattern == nil {
		return nil
	}

	switch p := pattern.(type) {
	case *syntax.NormalPattern:
		return f(vm, syntax.ExprFromNode(p.ToUntyped()), value)

	case *syntax.PlaceholderPattern:
		return nil

	case *syntax.ParenthesizedPattern:
		return destructureImpl(vm, p.Pattern(), value, f)

	case *syntax.DestructuringPattern:
		switch v := value.(type) {
		case *foundations.Array:
			return destructureArray(vm, p, v, f)
		case *foundations.Dict:
			return destructureDict(vm, p, v, f)
		default:
			return &CannotDestructureError{
				Type: value.Type(),
				Span: p.ToUntyped().Span(),
			}
		}

	default:
		return &UnknownPatternError{Span: pattern.ToUntyped().Span()}
	}
}

func destructureArray(vm *Vm, destruct *syntax.DestructuringPattern, arr *foundations.Array, f bindingFunc) error {
	items := destruct.Items()
	length := arr.Len()
	var idx int

	for _, item := range items {
		switch it := item.(type) {
		case *syntax.DestructuringBinding:
			if idx >= length {
				return wrongNumberOfElements(destruct, length)
			}
			if err := destructureImpl(vm, it.Pattern(), arr.At(idx), f); err != nil {
				return err
			}
			idx++

		case *syntax.DestructuringSpread:
			patternCount := countPatterns(items)
			sinkSize := length + 1 - patternCount
			if sinkSize < 0 || idx+sinkSize > length {
				return wrongNumberOfElements(destruct, length)
			}

			if sink := it.Sink(); sink != nil {
				sinkArray := foundations.NewArray()
				for i := idx; i < idx+sinkSize; i++ {
					sinkArray.Push(arr.At(i))
				}
				if err := destructureImpl(vm, sink, sinkArray, f); err != nil {
					return err
				}
			}
			idx += sinkSize

		case *syntax.DestructuringNamed:
			return &CannotDestructureNamedFromArrayError{Span: destruct.ToUntyped().Span()}
		}
	}

	if idx < length {
		return wrongNumberOfElements(destruct, length)
	}

	return nil
}

func destructureDict(vm *Vm, destruct *syntax.DestructuringPattern, dict *foundations.Dict, f bindingFunc) error {
	items := destruct.Items()
	var sink syntax.Pattern
	used := make(map[string]bool)

	for _, item := range items {
		switch it := item.(type) {
		case *syntax.DestructuringBinding:
			pattern := it.Pattern()
			if normalPat, ok := pattern.(*syntax.NormalPattern); ok {
				name := normalPat.Name()
				val, ok := dict.Get(name)
				if !ok {
					return atSpan(fmt.Errorf("dictionary does not contain key %q", name), normalPat.ToUntyped().Span())
				}
				expr := syntax.ExprFromNode(normalPat.ToUntyped())
				if err := f(vm, expr, val); err != nil {
					return err
				}
				used[name] = true
			} else {
				return &CannotDestructureUnnamedFromDictError{Span: pattern.ToUntyped().Span()}
			}

		case *syntax.DestructuringNamed:
			nameIdent := it.Name()
			if nameIdent == nil {
				continue
			}
			name := nameIdent.Get()
			val, ok := dict.Get(name)
			if !ok {
				return atSpan(fmt.Errorf("dictionary does not contain key %q", name), nameIdent.ToUntyped().Span())
			}
			if err := destructureImpl(vm, it.Pattern(), val, f); err != nil {
				return err
			}
			used[name] = true

		case *syntax.DestructuringSpread:
			sink = it.Sink()
		}
	}

	if sink != nil {
		sinkDict := foundations.NewDict()
		for _, key := range dict.Keys() {
			if !used[key] {
				val, _ := dict.Get(key)
				sinkDict.Set(key, val)
			}
		}
		expr := syntax.ExprFromNode(sink.ToUntyped())
		if err := f(vm, expr, sinkDict); err != nil {
			return err
		}
	}

	return nil
}

func countPatterns(items []syntax.DestructuringItem) int {
	count := 0
	for _, item := range items {
		if _, ok := item.(*syntax.DestructuringBinding); ok {
			count++
		}
	}
	return count
}

func wrongNumberOfElements(destruct *syntax.DestructuringPattern, length int) error {
	items := destruct.Items()
	count := 0
	hasSpread := false

	for _, item := range items {
		switch item.(type) {
		case *syntax.DestructuringBinding:
			count++
		case *syntax.DestructuringSpread:
			hasSpread = true
		}
	}

	var quantifier string
	if length > count {
		quantifier = "too many"
	} else {
		quantifier = "not enough"
	}

	var expected string
	if hasSpread {
		expected = fmt.Sprintf("at least %d elements", count)
	} else {
		expected = fmt.Sprintf("%d elements", count)
	}

	return &WrongNumberOfElementsError{
		Quantifier: quantifier,
		Expected:   expected,
		Got:        length,
		Span:       destruct.ToUntyped().Span(),
	}
}

// Error types

type CannotAssignError struct {
	Span syntax.Span
}

func (e *CannotAssignError) Error() string {
	return "cannot assign to this expression"
}

type CannotDestructureError struct {
	Type foundations.Type
	Span syntax.Span
}

func (e *CannotDestructureError) Error() string {
	return fmt.Sprintf("cannot destructure %s", e.Type)
}

type UnknownPatternError struct {
	Span syntax.Span
}

func (e *UnknownPatternError) Error() string {
	return "unknown pattern type"
}

type CannotDestructureNamedFromArrayError struct {
	Span syntax.Span
}

func (e *CannotDestructureNamedFromArrayError) Error() string {
	return "cannot destructure named pattern from an array"
}

type CannotDestructureUnnamedFromDictError struct {
	Span syntax.Span
}

func (e *CannotDestructureUnnamedFromDictError) Error() string {
	return "cannot destructure unnamed pattern from dictionary"
}

type WrongNumberOfElementsError struct {
	Quantifier string
	Expected   string
	Got        int
	Span       syntax.Span
}

func (e *WrongNumberOfElementsError) Error() string {
	return fmt.Sprintf("%s elements to destructure; the provided array has a length of %d, but the pattern expects %s",
		e.Quantifier, e.Got, e.Expected)
}

