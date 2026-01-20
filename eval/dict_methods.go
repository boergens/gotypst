package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// getDictMethod returns a built-in method for dictionaries, or nil if not found.
// All dict methods preserve insertion order as per Typst semantics.
func getDictMethod(d *DictValue, name string, span syntax.Span) Value {
	switch name {
	case "len":
		return makeDictMethod("len", func(vm *Vm, args *Args) (Value, error) {
			if err := args.Finish(); err != nil {
				return nil, err
			}
			return Int(int64(d.Len())), nil
		})

	case "is-empty":
		return makeDictMethod("is-empty", func(vm *Vm, args *Args) (Value, error) {
			if err := args.Finish(); err != nil {
				return nil, err
			}
			return Bool(d.Len() == 0), nil
		})

	case "at":
		return makeDictMethod("at", func(vm *Vm, args *Args) (Value, error) {
			// at(key) or at(key, default: value)
			keyArg, err := args.Expect("key")
			if err != nil {
				return nil, err
			}
			key, ok := AsStr(keyArg.V)
			if !ok {
				return nil, &TypeError{
					Expected: TypeStr,
					Got:      keyArg.V.Type(),
					Span:     keyArg.Span,
				}
			}

			// Check for default argument
			defaultArg := args.Find("default")

			if err := args.Finish(); err != nil {
				return nil, err
			}

			val, found := d.Get(key)
			if found {
				return val, nil
			}

			// Return default if provided, otherwise error
			if defaultArg != nil {
				return defaultArg.V, nil
			}
			return nil, &KeyNotFoundError{Key: key, Span: args.Span}
		})

	case "get":
		// get(key, default: none) - like at but returns none if not found by default
		return makeDictMethod("get", func(vm *Vm, args *Args) (Value, error) {
			keyArg, err := args.Expect("key")
			if err != nil {
				return nil, err
			}
			key, ok := AsStr(keyArg.V)
			if !ok {
				return nil, &TypeError{
					Expected: TypeStr,
					Got:      keyArg.V.Type(),
					Span:     keyArg.Span,
				}
			}

			// Check for default argument (defaults to none)
			defaultArg := args.Find("default")

			if err := args.Finish(); err != nil {
				return nil, err
			}

			val, found := d.Get(key)
			if found {
				return val, nil
			}

			if defaultArg != nil {
				return defaultArg.V, nil
			}
			return None, nil
		})

	case "contains":
		return makeDictMethod("contains", func(vm *Vm, args *Args) (Value, error) {
			keyArg, err := args.Expect("key")
			if err != nil {
				return nil, err
			}
			key, ok := AsStr(keyArg.V)
			if !ok {
				return nil, &TypeError{
					Expected: TypeStr,
					Got:      keyArg.V.Type(),
					Span:     keyArg.Span,
				}
			}

			if err := args.Finish(); err != nil {
				return nil, err
			}

			_, found := d.Get(key)
			return Bool(found), nil
		})

	case "insert":
		return makeDictMethod("insert", func(vm *Vm, args *Args) (Value, error) {
			keyArg, err := args.Expect("key")
			if err != nil {
				return nil, err
			}
			key, ok := AsStr(keyArg.V)
			if !ok {
				return nil, &TypeError{
					Expected: TypeStr,
					Got:      keyArg.V.Type(),
					Span:     keyArg.Span,
				}
			}

			valueArg, err := args.Expect("value")
			if err != nil {
				return nil, err
			}

			if err := args.Finish(); err != nil {
				return nil, err
			}

			d.Set(key, valueArg.V)
			return None, nil
		})

	case "remove":
		return makeDictMethod("remove", func(vm *Vm, args *Args) (Value, error) {
			keyArg, err := args.Expect("key")
			if err != nil {
				return nil, err
			}
			key, ok := AsStr(keyArg.V)
			if !ok {
				return nil, &TypeError{
					Expected: TypeStr,
					Got:      keyArg.V.Type(),
					Span:     keyArg.Span,
				}
			}

			// Check for default argument
			defaultArg := args.Find("default")

			if err := args.Finish(); err != nil {
				return nil, err
			}

			val, removed := d.Remove(key)
			if removed {
				return val, nil
			}

			if defaultArg != nil {
				return defaultArg.V, nil
			}
			return nil, &KeyNotFoundError{Key: key, Span: args.Span}
		})

	case "clear":
		return makeDictMethod("clear", func(vm *Vm, args *Args) (Value, error) {
			if err := args.Finish(); err != nil {
				return nil, err
			}
			d.Clear()
			return None, nil
		})

	case "keys":
		return makeDictMethod("keys", func(vm *Vm, args *Args) (Value, error) {
			if err := args.Finish(); err != nil {
				return nil, err
			}
			keys := d.Keys()
			arr := make(ArrayValue, len(keys))
			for i, k := range keys {
				arr[i] = Str(k)
			}
			return arr, nil
		})

	case "values":
		return makeDictMethod("values", func(vm *Vm, args *Args) (Value, error) {
			if err := args.Finish(); err != nil {
				return nil, err
			}
			values := d.Values()
			return ArrayValue(values), nil
		})

	case "pairs":
		return makeDictMethod("pairs", func(vm *Vm, args *Args) (Value, error) {
			if err := args.Finish(); err != nil {
				return nil, err
			}
			pairs := d.Pairs()
			arr := make(ArrayValue, len(pairs))
			for i, p := range pairs {
				arr[i] = ArrayValue{Str(p.Key), p.Value}
			}
			return arr, nil
		})

	case "filter":
		return makeDictMethod("filter", func(vm *Vm, args *Args) (Value, error) {
			testArg, err := args.Expect("test")
			if err != nil {
				return nil, err
			}
			testFn, ok := AsFunc(testArg.V)
			if !ok {
				return nil, &TypeError{
					Expected: TypeFunc,
					Got:      testArg.V.Type(),
					Span:     testArg.Span,
				}
			}

			if err := args.Finish(); err != nil {
				return nil, err
			}

			result := NewDict()
			for _, key := range d.Keys() {
				val, _ := d.Get(key)
				// Call test function with (key, value)
				callArgs := NewArgs(args.Span)
				callArgs.Push(Str(key), args.Span)
				callArgs.Push(val, args.Span)

				testResult, err := CallFunc(vm, testFn, callArgs)
				if err != nil {
					return nil, err
				}

				keep, ok := AsBool(testResult)
				if !ok {
					return nil, &TypeError{
						Expected: TypeBool,
						Got:      testResult.Type(),
						Span:     args.Span,
					}
				}

				if keep {
					result.Set(key, val)
				}
			}
			return result, nil
		})

	case "map":
		return makeDictMethod("map", func(vm *Vm, args *Args) (Value, error) {
			mapperArg, err := args.Expect("mapper")
			if err != nil {
				return nil, err
			}
			mapperFn, ok := AsFunc(mapperArg.V)
			if !ok {
				return nil, &TypeError{
					Expected: TypeFunc,
					Got:      mapperArg.V.Type(),
					Span:     mapperArg.Span,
				}
			}

			if err := args.Finish(); err != nil {
				return nil, err
			}

			result := NewDict()
			for _, key := range d.Keys() {
				val, _ := d.Get(key)
				// Call mapper function with (key, value)
				callArgs := NewArgs(args.Span)
				callArgs.Push(Str(key), args.Span)
				callArgs.Push(val, args.Span)

				mappedVal, err := CallFunc(vm, mapperFn, callArgs)
				if err != nil {
					return nil, err
				}

				result.Set(key, mappedVal)
			}
			return result, nil
		})
	}

	return nil
}

// makeDictMethod creates a FuncValue wrapping a native function for dict methods.
func makeDictMethod(name string, fn func(vm *Vm, args *Args) (Value, error)) Value {
	return FuncValue{Func: &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: fn,
			Info: &FuncInfo{Name: name},
		},
	}}
}
