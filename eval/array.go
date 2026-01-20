package eval

import (
	"slices"
	"strconv"

	"github.com/boergens/gotypst/syntax"
)

// Array method implementations for Typst arrays.
// These implement the standard library array methods.

// getArrayMethod returns a bound method for an array, or nil if not found.
// Note: Mutable methods (push, pop, insert, remove) are handled specially in evalFuncCall.
func getArrayMethod(arr ArrayValue, name string, span syntax.Span) Value {
	switch name {
	case "len":
		return makeArrayMethod(arr, span, arrayLen)
	case "first":
		return makeArrayMethod(arr, span, arrayFirst)
	case "last":
		return makeArrayMethod(arr, span, arrayLast)
	case "at":
		return makeArrayMethod(arr, span, arrayAt)
	case "slice":
		return makeArrayMethod(arr, span, arraySlice)
	case "contains":
		return makeArrayMethod(arr, span, arrayContains)
	case "find":
		return makeArrayMethod(arr, span, arrayFind)
	case "position":
		return makeArrayMethod(arr, span, arrayPosition)
	case "filter":
		return makeArrayMethod(arr, span, arrayFilter)
	case "map":
		return makeArrayMethod(arr, span, arrayMap)
	case "enumerate":
		return makeArrayMethod(arr, span, arrayEnumerate)
	case "flatten":
		return makeArrayMethod(arr, span, arrayFlatten)
	case "rev":
		return makeArrayMethod(arr, span, arrayRev)
	case "sorted":
		return makeArrayMethod(arr, span, arraySorted)
	case "dedup":
		return makeArrayMethod(arr, span, arrayDedup)
	case "zip":
		return makeArrayMethod(arr, span, arrayZip)
	case "join":
		return makeArrayMethod(arr, span, arrayJoin)
	case "fold":
		return makeArrayMethod(arr, span, arrayFold)
	case "reduce":
		return makeArrayMethod(arr, span, arrayReduce)
	case "sum":
		return makeArrayMethod(arr, span, arraySum)
	case "product":
		return makeArrayMethod(arr, span, arrayProduct)
	default:
		return nil
	}
}

// makeArrayMethod creates a bound method for an array.
func makeArrayMethod(arr ArrayValue, span syntax.Span, fn func(vm *Vm, arr ArrayValue, args *Args) (Value, error)) Value {
	arrCopy := arr // Capture a copy
	return FuncValue{Func: &Func{
		Span: span,
		Repr: NativeFunc{
			Func: func(vm *Vm, args *Args) (Value, error) {
				return fn(vm, arrCopy, args)
			},
		},
	}}
}


// ----------------------------------------------------------------------------
// Array Methods - Basic Access
// ----------------------------------------------------------------------------

// arrayLen returns the length of an array.
func arrayLen(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}
	return Int(int64(len(arr))), nil
}

// arrayFirst returns the first element of an array.
func arrayFirst(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, &ArrayEmptyError{Method: "first", Span: args.Span}
	}
	return arr[0], nil
}

// arrayLast returns the last element of an array.
func arrayLast(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, &ArrayEmptyError{Method: "last", Span: args.Span}
	}
	return arr[len(arr)-1], nil
}

// arrayAt returns the element at the given index.
func arrayAt(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	// Get the index argument
	indexArg, err := args.Expect("index")
	if err != nil {
		return nil, err
	}

	index, ok := AsInt(indexArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeInt, Got: indexArg.V.Type(), Span: indexArg.Span}
	}

	// Check for default argument
	defaultVal := args.Find("default")

	// Normalize negative index
	idx := int(index)
	if idx < 0 {
		idx = len(arr) + idx
	}

	// Check bounds
	if idx < 0 || idx >= len(arr) {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, &ArrayIndexError{Index: int(index), Len: len(arr), Span: args.Span}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return arr[idx], nil
}

// ----------------------------------------------------------------------------
// Array Methods - Mutation (called from evalMutableArrayMethodCall in expr.go)
// ----------------------------------------------------------------------------

// arrayPushMutable adds an element to the end of the array.
func arrayPushMutable(arr *ArrayValue, args *Args) (Value, error) {
	val, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	if err := args.Finish(); err != nil {
		return nil, err
	}
	*arr = append(*arr, val.V)
	return None, nil
}

// arrayPopMutable removes and returns the last element.
func arrayPopMutable(arr *ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}
	if len(*arr) == 0 {
		return nil, &ArrayEmptyError{Method: "pop", Span: args.Span}
	}
	last := (*arr)[len(*arr)-1]
	*arr = (*arr)[:len(*arr)-1]
	return last, nil
}

// arrayInsertMutable inserts an element at the given index.
func arrayInsertMutable(arr *ArrayValue, args *Args) (Value, error) {
	indexArg, err := args.Expect("index")
	if err != nil {
		return nil, err
	}
	index, ok := AsInt(indexArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeInt, Got: indexArg.V.Type(), Span: indexArg.Span}
	}

	valArg, err := args.Expect("value")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Normalize negative index
	idx := int(index)
	if idx < 0 {
		idx = len(*arr) + idx + 1
	}

	// Bounds check
	if idx < 0 || idx > len(*arr) {
		return nil, &ArrayIndexError{Index: int(index), Len: len(*arr), Span: args.Span}
	}

	// Insert at index
	*arr = slices.Insert(*arr, idx, valArg.V)
	return None, nil
}

// arrayRemoveMutable removes and returns the element at the given index.
func arrayRemoveMutable(arr *ArrayValue, args *Args) (Value, error) {
	indexArg, err := args.Expect("index")
	if err != nil {
		return nil, err
	}
	index, ok := AsInt(indexArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeInt, Got: indexArg.V.Type(), Span: indexArg.Span}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Normalize negative index
	idx := int(index)
	if idx < 0 {
		idx = len(*arr) + idx
	}

	// Bounds check
	if idx < 0 || idx >= len(*arr) {
		return nil, &ArrayIndexError{Index: int(index), Len: len(*arr), Span: args.Span}
	}

	// Remove and return
	removed := (*arr)[idx]
	*arr = slices.Delete(*arr, idx, idx+1)
	return removed, nil
}

// arraySlice returns a slice of the array.
func arraySlice(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	startArg, err := args.Expect("start")
	if err != nil {
		return nil, err
	}
	start, ok := AsInt(startArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeInt, Got: startArg.V.Type(), Span: startArg.Span}
	}

	// Optional end argument
	var end int64
	if endArg := args.Eat(); endArg != nil {
		var ok bool
		end, ok = AsInt(endArg.V)
		if !ok {
			return nil, &TypeError{Expected: TypeInt, Got: endArg.V.Type(), Span: endArg.Span}
		}
	} else {
		end = int64(len(arr))
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Normalize indices
	startIdx := int(start)
	endIdx := int(end)
	if startIdx < 0 {
		startIdx = len(arr) + startIdx
	}
	if endIdx < 0 {
		endIdx = len(arr) + endIdx
	}

	// Clamp to bounds
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(arr) {
		endIdx = len(arr)
	}
	if startIdx > endIdx {
		startIdx = endIdx
	}

	// Create slice copy
	result := make(ArrayValue, endIdx-startIdx)
	copy(result, arr[startIdx:endIdx])
	return result, nil
}

// ----------------------------------------------------------------------------
// Array Methods - Search
// ----------------------------------------------------------------------------

// arrayContains checks if the array contains a value.
func arrayContains(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	val, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	if err := args.Finish(); err != nil {
		return nil, err
	}

	for _, elem := range arr {
		if Equal(elem, val.V) {
			return True, nil
		}
	}
	return False, nil
}

// arrayFind finds the first element matching a predicate.
func arrayFind(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnArg, err := args.Expect("function")
	if err != nil {
		return nil, err
	}
	fn, ok := AsFunc(fnArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnArg.V.Type(), Span: fnArg.Span}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	for _, elem := range arr {
		callArgs := NewArgs(args.Span)
		callArgs.Push(elem, args.Span)
		result, err := CallFunc(vm, fn, callArgs)
		if err != nil {
			return nil, err
		}
		if b, ok := AsBool(result); ok && b {
			return elem, nil
		}
	}
	return None, nil
}

// arrayPosition finds the index of the first element matching a predicate.
func arrayPosition(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnArg, err := args.Expect("function")
	if err != nil {
		return nil, err
	}
	fn, ok := AsFunc(fnArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnArg.V.Type(), Span: fnArg.Span}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	for i, elem := range arr {
		callArgs := NewArgs(args.Span)
		callArgs.Push(elem, args.Span)
		result, err := CallFunc(vm, fn, callArgs)
		if err != nil {
			return nil, err
		}
		if b, ok := AsBool(result); ok && b {
			return Int(int64(i)), nil
		}
	}
	return None, nil
}

// ----------------------------------------------------------------------------
// Array Methods - Transformation
// ----------------------------------------------------------------------------

// arrayFilter filters the array using a predicate.
func arrayFilter(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnArg, err := args.Expect("function")
	if err != nil {
		return nil, err
	}
	fn, ok := AsFunc(fnArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnArg.V.Type(), Span: fnArg.Span}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	var result ArrayValue
	for _, elem := range arr {
		callArgs := NewArgs(args.Span)
		callArgs.Push(elem, args.Span)
		match, err := CallFunc(vm, fn, callArgs)
		if err != nil {
			return nil, err
		}
		if b, ok := AsBool(match); ok && b {
			result = append(result, elem)
		}
	}
	return result, nil
}

// arrayMap maps a function over the array.
func arrayMap(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnArg, err := args.Expect("function")
	if err != nil {
		return nil, err
	}
	fn, ok := AsFunc(fnArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnArg.V.Type(), Span: fnArg.Span}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	result := make(ArrayValue, len(arr))
	for i, elem := range arr {
		callArgs := NewArgs(args.Span)
		callArgs.Push(elem, args.Span)
		mapped, err := CallFunc(vm, fn, callArgs)
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

// arrayEnumerate returns (index, value) pairs.
func arrayEnumerate(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	// Optional start argument
	start := int64(0)
	if startArg := args.Find("start"); startArg != nil {
		var ok bool
		start, ok = AsInt(startArg.V)
		if !ok {
			return nil, &TypeError{Expected: TypeInt, Got: startArg.V.Type(), Span: startArg.Span}
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	result := make(ArrayValue, len(arr))
	for i, elem := range arr {
		result[i] = ArrayValue{Int(start + int64(i)), elem}
	}
	return result, nil
}

// arrayFlatten flattens nested arrays by one level.
func arrayFlatten(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}

	var result ArrayValue
	for _, elem := range arr {
		if nested, ok := elem.(ArrayValue); ok {
			result = append(result, nested...)
		} else {
			result = append(result, elem)
		}
	}
	return result, nil
}

// ----------------------------------------------------------------------------
// Array Methods - Reordering
// ----------------------------------------------------------------------------

// arrayRev returns a reversed copy of the array.
func arrayRev(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}

	result := make(ArrayValue, len(arr))
	for i, elem := range arr {
		result[len(arr)-1-i] = elem
	}
	return result, nil
}

// arraySorted returns a sorted copy of the array.
func arraySorted(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	// Optional key function
	keyFn := args.Find("key")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create a copy
	result := make(ArrayValue, len(arr))
	copy(result, arr)

	// Sort with optional key function
	var sortErr error
	slices.SortStableFunc(result, func(a, b Value) int {
		if sortErr != nil {
			return 0
		}

		var aKey, bKey Value
		if keyFn != nil {
			fn, ok := AsFunc(keyFn.V)
			if !ok {
				sortErr = &TypeError{Expected: TypeFunc, Got: keyFn.V.Type(), Span: keyFn.Span}
				return 0
			}

			aArgs := NewArgs(args.Span)
			aArgs.Push(a, args.Span)
			aKey, sortErr = CallFunc(vm, fn, aArgs)
			if sortErr != nil {
				return 0
			}

			bArgs := NewArgs(args.Span)
			bArgs.Push(b, args.Span)
			bKey, sortErr = CallFunc(vm, fn, bArgs)
			if sortErr != nil {
				return 0
			}
		} else {
			aKey, bKey = a, b
		}

		cmpResult, err := compareValues(aKey, bKey)
		if err != nil {
			sortErr = err
			return 0
		}
		return cmpResult
	})

	if sortErr != nil {
		return nil, sortErr
	}
	return result, nil
}


// arrayDedup removes consecutive duplicates.
func arrayDedup(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	// Optional key function
	keyFn := args.Find("key")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return ArrayValue{}, nil
	}

	var result ArrayValue
	result = append(result, arr[0])

	for i := 1; i < len(arr); i++ {
		var prev, curr Value
		if keyFn != nil {
			// With key function - would need to call the function
			// For now, use the values directly
			prev, curr = arr[i-1], arr[i]
		} else {
			prev, curr = arr[i-1], arr[i]
		}

		if !Equal(prev, curr) {
			result = append(result, arr[i])
		}
	}
	return result, nil
}

// ----------------------------------------------------------------------------
// Array Methods - Aggregation
// ----------------------------------------------------------------------------

// arrayZip zips this array with another array.
func arrayZip(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	otherArg, err := args.Expect("other")
	if err != nil {
		return nil, err
	}
	other, ok := AsArray(otherArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeArray, Got: otherArg.V.Type(), Span: otherArg.Span}
	}

	// Optional exact argument
	exact := false
	if exactArg := args.Find("exact"); exactArg != nil {
		exact, _ = AsBool(exactArg.V)
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	if exact && len(arr) != len(other) {
		return nil, &ArrayLengthMismatchError{
			Len1: len(arr),
			Len2: len(other),
			Span: args.Span,
		}
	}

	// Use minimum length
	length := len(arr)
	if len(other) < length {
		length = len(other)
	}

	result := make(ArrayValue, length)
	for i := 0; i < length; i++ {
		result[i] = ArrayValue{arr[i], other[i]}
	}
	return result, nil
}

// arrayJoin joins array elements with a separator.
func arrayJoin(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	sepArg, err := args.Expect("separator")
	if err != nil {
		return nil, err
	}

	// Optional last separator
	lastSep := args.Find("last")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return None, nil
	}

	if len(arr) == 1 {
		return arr[0], nil
	}

	// Determine if we're joining strings or content
	sep := sepArg.V
	var result Value

	// Check if separator is a string
	if sepStr, ok := AsStr(sep); ok {
		// Join as strings
		var sb []byte
		for i, elem := range arr {
			if i > 0 {
				if i == len(arr)-1 && lastSep != nil {
					if lastStr, ok := AsStr(lastSep.V); ok {
						sb = append(sb, lastStr...)
					}
				} else {
					sb = append(sb, sepStr...)
				}
			}
			if elemStr, ok := AsStr(elem); ok {
				sb = append(sb, elemStr...)
			} else {
				// Non-string element, convert to string representation
				sb = append(sb, formatValue(vm, elem)...)
			}
		}
		result = Str(string(sb))
	} else {
		// Join as content
		var elements []ContentElement
		for i, elem := range arr {
			if i > 0 {
				if i == len(arr)-1 && lastSep != nil {
					elements = appendValueToContent(elements, lastSep.V)
				} else {
					elements = appendValueToContent(elements, sep)
				}
			}
			elements = appendValueToContent(elements, elem)
		}
		result = ContentValue{Content: Content{Elements: elements}}
	}

	return result, nil
}

// formatValue formats a value as a string.
func formatValue(_ *Vm, v Value) string {
	switch val := v.(type) {
	case StrValue:
		return string(val)
	case IntValue:
		return strconv.FormatInt(int64(val), 10)
	case FloatValue:
		return strconv.FormatFloat(float64(val), 'g', -1, 64)
	default:
		return v.Type().String()
	}
}

// appendValueToContent appends a value to content elements.
func appendValueToContent(elements []ContentElement, v Value) []ContentElement {
	switch val := v.(type) {
	case ContentValue:
		return append(elements, val.Content.Elements...)
	case StrValue:
		return append(elements, &TextElement{Text: string(val)})
	default:
		// Display other values as text
		return append(elements, &TextElement{Text: formatValue(nil, val)})
	}
}

// arrayFold folds the array with an initial value.
func arrayFold(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	initArg, err := args.Expect("initial")
	if err != nil {
		return nil, err
	}

	fnArg, err := args.Expect("function")
	if err != nil {
		return nil, err
	}
	fn, ok := AsFunc(fnArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnArg.V.Type(), Span: fnArg.Span}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	acc := initArg.V
	for _, elem := range arr {
		callArgs := NewArgs(args.Span)
		callArgs.Push(acc, args.Span)
		callArgs.Push(elem, args.Span)
		var err error
		acc, err = CallFunc(vm, fn, callArgs)
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

// arrayReduce reduces the array (like fold but uses first element as initial).
func arrayReduce(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnArg, err := args.Expect("function")
	if err != nil {
		return nil, err
	}
	fn, ok := AsFunc(fnArg.V)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnArg.V.Type(), Span: fnArg.Span}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return nil, &ArrayEmptyError{Method: "reduce", Span: args.Span}
	}

	acc := arr[0]
	for _, elem := range arr[1:] {
		callArgs := NewArgs(args.Span)
		callArgs.Push(acc, args.Span)
		callArgs.Push(elem, args.Span)
		var err error
		acc, err = CallFunc(vm, fn, callArgs)
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

// arraySum sums all elements in the array.
func arraySum(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	defaultVal := args.Find("default")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, &ArrayEmptyError{Method: "sum", Span: args.Span}
	}

	// Start with first element
	var sum Value = arr[0]
	for _, elem := range arr[1:] {
		var err error
		sum, err = Add(sum, elem, args.Span)
		if err != nil {
			return nil, err
		}
	}
	return sum, nil
}

// arrayProduct multiplies all elements in the array.
func arrayProduct(_ *Vm, arr ArrayValue, args *Args) (Value, error) {
	defaultVal := args.Find("default")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, &ArrayEmptyError{Method: "product", Span: args.Span}
	}

	// Start with first element
	var product Value = arr[0]
	for _, elem := range arr[1:] {
		var err error
		product, err = Mul(product, elem, args.Span)
		if err != nil {
			return nil, err
		}
	}
	return product, nil
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// ArrayEmptyError is returned when an operation requires a non-empty array.
type ArrayEmptyError struct {
	Method string
	Span   syntax.Span
}

func (e *ArrayEmptyError) Error() string {
	return "array is empty"
}

// ArrayIndexError is returned when an array index is out of bounds.
type ArrayIndexError struct {
	Index int
	Len   int
	Span  syntax.Span
}

func (e *ArrayIndexError) Error() string {
	return "array index out of bounds (index: " + Int(int64(e.Index)).Display().String() +
		", len: " + Int(int64(e.Len)).Display().String() + ") and no default value was specified"
}

// ArrayLengthMismatchError is returned when array lengths don't match.
type ArrayLengthMismatchError struct {
	Len1 int
	Len2 int
	Span syntax.Span
}

func (e *ArrayLengthMismatchError) Error() string {
	return "array lengths do not match"
}
