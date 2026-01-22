package eval

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// GetArrayMethod returns a bound method for an array value.
func GetArrayMethod(target ArrayValue, methodName string, span syntax.Span) Value {
	return createArrayMethod(target, methodName, span)
}

// createArrayMethod creates a bound method function for array methods.
func createArrayMethod(target ArrayValue, methodName string, span syntax.Span) Value {
	var fn func(engine *Engine, context *Context, args *Args) (Value, error)

	switch methodName {
	case "len":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayLen(target, args)
		}
	case "first":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayFirst(target, args, span)
		}
	case "last":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayLast(target, args, span)
		}
	case "at":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayAt(target, args, span)
		}
	case "push":
		// Note: push is a mutating method, needs special handling
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return None, fmt.Errorf("push requires mutable access")
		}
	case "pop":
		// Note: pop is a mutating method, needs special handling
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return nil, &ArrayEmptyError{Span: span}
		}
	case "insert":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return None, fmt.Errorf("insert requires mutable access")
		}
	case "remove":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return None, fmt.Errorf("remove requires mutable access")
		}
	case "slice":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArraySlice(target, args)
		}
	case "contains":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayContains(target, args)
		}
	case "find":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayFind(engine, context, target, args)
		}
	case "position":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayPosition(engine, context, target, args)
		}
	case "filter":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayFilter(engine, context, target, args)
		}
	case "map":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayMap(engine, context, target, args)
		}
	case "fold":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayFold(engine, context, target, args)
		}
	case "sum":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArraySum(target, args)
		}
	case "product":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayProduct(target, args)
		}
	case "rev":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayRev(target, args)
		}
	case "join":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayJoin(target, args)
		}
	case "sorted":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArraySorted(target, args)
		}
	case "zip":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayZip(target, args)
		}
	case "enumerate":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayEnumerate(target, args)
		}
	case "dedup":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return ArrayDedup(target, args)
		}
	default:
		return nil
	}

	name := methodName
	return FuncValue{Func: &Func{
		Name: &name,
		Span: span,
		Repr: NativeFunc{
			Func: fn,
			Info: &FuncInfo{Name: methodName},
		},
	}}
}

// ArrayEmptyError is returned when operating on an empty array.
type ArrayEmptyError struct {
	Span syntax.Span
}

func (e *ArrayEmptyError) Error() string {
	return "array is empty"
}

// ArrayIndexError is returned when an array index is out of bounds.
type ArrayIndexError struct {
	Index   int64
	Length  int
	Default bool
	Span    syntax.Span
}

func (e *ArrayIndexError) Error() string {
	if e.Default {
		return fmt.Sprintf("array index out of bounds (index: %d, len: %d)", e.Index, e.Length)
	}
	return fmt.Sprintf("array index out of bounds (index: %d, len: %d) and no default value was specified", e.Index, e.Length)
}

// ArrayLen returns the length of an array.
func ArrayLen(arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}
	return Int(int64(len(arr))), nil
}

// ArrayFirst returns the first element of an array.
func ArrayFirst(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, &ArrayEmptyError{Span: span}
	}
	return arr[0], nil
}

// ArrayLast returns the last element of an array.
func ArrayLast(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, &ArrayEmptyError{Span: span}
	}
	return arr[len(arr)-1], nil
}

// ArrayAt returns an element at a given index with optional default.
func ArrayAt(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	indexArg, err := args.Expect("index")
	if err != nil {
		return nil, err
	}

	defaultVal := args.Find("default")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	idx, ok := AsInt(indexArg.V)
	if !ok {
		return nil, fmt.Errorf("expected integer index, got %s", indexArg.V.Type())
	}

	// Handle negative indices
	index := idx
	if index < 0 {
		index = int64(len(arr)) + index
	}

	// Check bounds
	if index < 0 || index >= int64(len(arr)) {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, &ArrayIndexError{Index: idx, Length: len(arr), Default: false, Span: span}
	}

	return arr[index], nil
}

// ArraySlice returns a slice of the array.
func ArraySlice(arr ArrayValue, args *Args) (Value, error) {
	startArg, err := args.Expect("start")
	if err != nil {
		return nil, err
	}

	endArg := args.Eat()

	if err := args.Finish(); err != nil {
		return nil, err
	}

	startIdx, ok := AsInt(startArg.V)
	if !ok {
		return nil, fmt.Errorf("expected integer start, got %s", startArg.V.Type())
	}

	// Handle negative start
	if startIdx < 0 {
		startIdx = int64(len(arr)) + startIdx
	}
	if startIdx < 0 {
		startIdx = 0
	}

	// Optional end
	endIdx := int64(len(arr))
	if endArg != nil {
		end, ok := AsInt(endArg.V)
		if !ok {
			return nil, fmt.Errorf("expected integer end, got %s", endArg.V.Type())
		}
		endIdx = end
		if endIdx < 0 {
			endIdx = int64(len(arr)) + endIdx
		}
	}

	if startIdx > int64(len(arr)) {
		startIdx = int64(len(arr))
	}
	if endIdx > int64(len(arr)) {
		endIdx = int64(len(arr))
	}
	if startIdx > endIdx {
		return ArrayValue{}, nil
	}

	return ArrayValue(arr[startIdx:endIdx]), nil
}

// ArrayContains checks if an array contains a value.
func ArrayContains(arr ArrayValue, args *Args) (Value, error) {
	valArg, err := args.Expect("value")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	for _, item := range arr {
		if valuesEqual(item, valArg.V) {
			return True, nil
		}
	}
	return False, nil
}

// ArrayFind finds the first element matching a predicate.
func ArrayFind(engine *Engine, context *Context, arr ArrayValue, args *Args) (Value, error) {
	predArg, err := args.Expect("predicate")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	for _, item := range arr {
		predArgs := NewArgs(args.Span)
		predArgs.Push(item, args.Span)
		result, err := engine.callFuncInternal(context, predArg.V, predArgs, args.Span)
		if err != nil {
			return nil, err
		}
		if b, ok := AsBool(result); ok && b {
			return item, nil
		}
	}
	return None, nil
}

// ArrayPosition finds the index of the first matching element.
func ArrayPosition(engine *Engine, context *Context, arr ArrayValue, args *Args) (Value, error) {
	predArg, err := args.Expect("predicate")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	for i, item := range arr {
		predArgs := NewArgs(args.Span)
		predArgs.Push(item, args.Span)
		result, err := engine.callFuncInternal(context, predArg.V, predArgs, args.Span)
		if err != nil {
			return nil, err
		}
		if b, ok := AsBool(result); ok && b {
			return Int(int64(i)), nil
		}
	}
	return None, nil
}

// ArrayFilter filters array elements by a predicate.
func ArrayFilter(engine *Engine, context *Context, arr ArrayValue, args *Args) (Value, error) {
	predArg, err := args.Expect("predicate")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	result := make(ArrayValue, 0)
	for _, item := range arr {
		predArgs := NewArgs(args.Span)
		predArgs.Push(item, args.Span)
		res, err := engine.callFuncInternal(context, predArg.V, predArgs, args.Span)
		if err != nil {
			return nil, err
		}
		if b, ok := AsBool(res); ok && b {
			result = append(result, item)
		}
	}
	return result, nil
}

// ArrayMap maps a function over array elements.
func ArrayMap(engine *Engine, context *Context, arr ArrayValue, args *Args) (Value, error) {
	fnArg, err := args.Expect("function")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	result := make(ArrayValue, len(arr))
	for i, item := range arr {
		fnArgs := NewArgs(args.Span)
		fnArgs.Push(item, args.Span)
		res, err := engine.callFuncInternal(context, fnArg.V, fnArgs, args.Span)
		if err != nil {
			return nil, err
		}
		result[i] = res
	}
	return result, nil
}

// ArrayFold folds an array using a function and initial value.
func ArrayFold(engine *Engine, context *Context, arr ArrayValue, args *Args) (Value, error) {
	initArg, err := args.Expect("initial")
	if err != nil {
		return nil, err
	}
	fnArg, err := args.Expect("function")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	acc := initArg.V
	for _, item := range arr {
		fnArgs := NewArgs(args.Span)
		fnArgs.Push(acc, args.Span)
		fnArgs.Push(item, args.Span)
		acc, err = engine.callFuncInternal(context, fnArg.V, fnArgs, args.Span)
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

// ArraySum sums all elements of an array.
func ArraySum(arr ArrayValue, args *Args) (Value, error) {
	defaultVal := args.Find("default")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, fmt.Errorf("cannot compute sum of empty array without default")
	}

	var sum int64
	for _, item := range arr {
		n, ok := AsInt(item)
		if !ok {
			return nil, fmt.Errorf("cannot sum non-integer values")
		}
		sum += n
	}
	return Int(sum), nil
}

// ArrayProduct computes the product of all elements.
func ArrayProduct(arr ArrayValue, args *Args) (Value, error) {
	defaultVal := args.Find("default")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, fmt.Errorf("cannot compute product of empty array without default")
	}

	var product int64 = 1
	for _, item := range arr {
		n, ok := AsInt(item)
		if !ok {
			return nil, fmt.Errorf("cannot multiply non-integer values")
		}
		product *= n
	}
	return Int(product), nil
}

// ArrayRev reverses an array.
func ArrayRev(arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}

	result := make(ArrayValue, len(arr))
	for i, item := range arr {
		result[len(arr)-1-i] = item
	}
	return result, nil
}

// ArrayJoin joins array elements with a separator.
func ArrayJoin(arr ArrayValue, args *Args) (Value, error) {
	sepArg, err := args.Expect("separator")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	// For now, simple string join
	sepStr, ok := sepArg.V.(StrValue)
	if ok {
		var result string
		for i, item := range arr {
			if i > 0 {
				result += string(sepStr)
			}
			if s, ok := item.(StrValue); ok {
				result += string(s)
			} else {
				result += fmt.Sprintf("%v", item)
			}
		}
		return Str(result), nil
	}

	// For content join, return content
	return ContentValue{}, nil
}

// ArraySorted returns a sorted copy of the array.
func ArraySorted(arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}

	result := make(ArrayValue, len(arr))
	copy(result, arr)

	// Simple bubble sort for now (strings and ints)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if arrayCompareValues(result[j], result[i]) < 0 {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result, nil
}

// ArrayZip zips two arrays together.
func ArrayZip(arr ArrayValue, args *Args) (Value, error) {
	otherArg, err := args.Expect("other")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	other, ok := AsArray(otherArg.V)
	if !ok {
		return nil, fmt.Errorf("expected array, got %s", otherArg.V.Type())
	}

	minLen := len(arr)
	if len(other) < minLen {
		minLen = len(other)
	}

	result := make(ArrayValue, minLen)
	for i := 0; i < minLen; i++ {
		result[i] = ArrayValue{arr[i], other[i]}
	}
	return result, nil
}

// ArrayEnumerate returns array with indices.
func ArrayEnumerate(arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}

	result := make(ArrayValue, len(arr))
	for i, item := range arr {
		result[i] = ArrayValue{Int(int64(i)), item}
	}
	return result, nil
}

// ArrayDedup removes consecutive duplicates.
func ArrayDedup(arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return ArrayValue{}, nil
	}

	result := ArrayValue{arr[0]}
	for i := 1; i < len(arr); i++ {
		if !valuesEqual(arr[i], arr[i-1]) {
			result = append(result, arr[i])
		}
	}
	return result, nil
}

// arrayCompareValues compares two values for sorting.
// Named differently to avoid conflict with ops.go's compareValues.
func arrayCompareValues(a, b Value) int {
	// Compare integers
	aInt, aIsInt := AsInt(a)
	bInt, bIsInt := AsInt(b)
	if aIsInt && bIsInt {
		if aInt < bInt {
			return -1
		} else if aInt > bInt {
			return 1
		}
		return 0
	}

	// Compare strings
	aStr, aIsStr := a.(StrValue)
	bStr, bIsStr := b.(StrValue)
	if aIsStr && bIsStr {
		if string(aStr) < string(bStr) {
			return -1
		} else if string(aStr) > string(bStr) {
			return 1
		}
		return 0
	}

	return 0
}
