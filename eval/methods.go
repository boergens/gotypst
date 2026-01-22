// Package eval provides special built-in methods on values.
// Translated from typst-eval/src/methods.rs

package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// IsMutatingMethod returns true if the method is a mutating method.
// Matches Rust: is_mutating_method
func IsMutatingMethod(method string) bool {
	switch method {
	case "push", "pop", "insert", "remove":
		return true
	}
	return false
}

// IsAccessorMethod returns true if the method is an accessor method.
// Matches Rust: is_accessor_method
func IsAccessorMethod(method string) bool {
	switch method {
	case "first", "last", "at":
		return true
	}
	return false
}

// CallMethodMut calls a mutating method on a value.
// Matches Rust: call_method_mut
func CallMethodMut(
	value *foundations.Value,
	method string,
	args *foundations.Args,
	span syntax.Span,
) (foundations.Value, error) {
	ty := (*value).Type()
	missing := func() error {
		return atSpan(fmt.Errorf("type %s has no method `%s`", ty, method), span)
	}

	output := foundations.Value(foundations.NoneValue{})

	switch v := (*value).(type) {
	case *foundations.Array:
		switch method {
		case "push":
			arg, err := args.Expect("value")
			if err != nil {
				return nil, err
			}
			v.Push(arg.V)

		case "pop":
			val, err := v.Pop()
			if err != nil {
				return nil, atSpan(err, span)
			}
			output = val

		case "insert":
			indexArg, err := args.Expect("index")
			if err != nil {
				return nil, err
			}
			valueArg, err := args.Expect("value")
			if err != nil {
				return nil, err
			}
			index, ok := foundations.AsInt(indexArg.V)
			if !ok {
				return nil, atSpan(fmt.Errorf("expected integer, found %s", indexArg.V.Type()), indexArg.Span)
			}
			if err := v.Insert(index, valueArg.V); err != nil {
				return nil, atSpan(err, span)
			}

		case "remove":
			indexArg, err := args.Expect("index")
			if err != nil {
				return nil, err
			}
			index, ok := foundations.AsInt(indexArg.V)
			if !ok {
				return nil, atSpan(fmt.Errorf("expected integer, found %s", indexArg.V.Type()), indexArg.Span)
			}
			def := args.Named("default")
			val, err := v.Remove(index, def)
			if err != nil {
				return nil, atSpan(err, span)
			}
			output = val

		default:
			return nil, missing()
		}

	case *foundations.Dict:
		switch method {
		case "insert":
			keyArg, err := args.Expect("key")
			if err != nil {
				return nil, err
			}
			key, ok := foundations.AsStr(keyArg.V)
			if !ok {
				return nil, atSpan(fmt.Errorf("expected string, found %s", keyArg.V.Type()), keyArg.Span)
			}
			valueArg, err := args.Expect("value")
			if err != nil {
				return nil, err
			}
			v.Insert(key, valueArg.V)

		case "remove":
			keyArg, err := args.Expect("key")
			if err != nil {
				return nil, err
			}
			key, ok := foundations.AsStr(keyArg.V)
			if !ok {
				return nil, atSpan(fmt.Errorf("expected string, found %s", keyArg.V.Type()), keyArg.Span)
			}
			def := args.Named("default")
			val, err := v.Remove(key, def)
			if err != nil {
				return nil, atSpan(err, span)
			}
			output = val

		default:
			return nil, missing()
		}

	default:
		return nil, missing()
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}
	return output, nil
}

// CallMethodAccess calls an accessor method on a value, returning a mutable reference.
// Matches Rust: call_method_access
func CallMethodAccess(
	value *foundations.Value,
	method string,
	args *foundations.Args,
	span syntax.Span,
) (*foundations.Value, error) {
	ty := (*value).Type()
	missing := func() error {
		return atSpan(fmt.Errorf("type %s has no method `%s`", ty, method), span)
	}

	var slot *foundations.Value

	switch v := (*value).(type) {
	case *foundations.Array:
		switch method {
		case "first":
			ptr, err := v.FirstMut()
			if err != nil {
				return nil, atSpan(err, span)
			}
			slot = ptr

		case "last":
			ptr, err := v.LastMut()
			if err != nil {
				return nil, atSpan(err, span)
			}
			slot = ptr

		case "at":
			indexArg, err := args.Expect("index")
			if err != nil {
				return nil, err
			}
			index, ok := foundations.AsInt(indexArg.V)
			if !ok {
				return nil, atSpan(fmt.Errorf("expected integer, found %s", indexArg.V.Type()), indexArg.Span)
			}
			ptr, err := v.AtMut(index)
			if err != nil {
				return nil, atSpan(err, span)
			}
			slot = ptr

		default:
			return nil, missing()
		}

	case *foundations.Dict:
		switch method {
		case "at":
			keyArg, err := args.Expect("key")
			if err != nil {
				return nil, err
			}
			key, ok := foundations.AsStr(keyArg.V)
			if !ok {
				return nil, atSpan(fmt.Errorf("expected string, found %s", keyArg.V.Type()), keyArg.Span)
			}
			ptr, err := v.AtMut(key)
			if err != nil {
				return nil, atSpan(err, span)
			}
			slot = ptr

		default:
			return nil, missing()
		}

	default:
		return nil, missing()
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}
	return slot, nil
}
