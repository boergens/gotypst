package eval

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// GetArrayMethod returns a bound method for an array value.
func GetArrayMethod(target ArrayValue, methodName string, span syntax.Span) Value {
	return createArrayMethod(target, methodName, span)
}

// createArrayMethod creates a method bound to an array value.
func createArrayMethod(target ArrayValue, methodName string, span syntax.Span) Value {
	var fn func(vm *Vm, args *Args) (Value, error)

	switch methodName {
	case "len":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayLen(target, args)
		}
	case "first":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayFirst(target, args, span)
		}
	case "last":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayLast(target, args, span)
		}
	case "at":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayAt(target, args, span)
		}
	case "push":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayPush(&target, args)
		}
	case "pop":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayPop(&target, args, span)
		}
	case "insert":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayInsert(&target, args, span)
		}
	case "remove":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayRemove(&target, args, span)
		}
	case "slice":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArraySlice(target, args, span)
		}
	case "contains":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayContains(target, args)
		}
	case "find":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayFind(vm, target, args)
		}
	case "position":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayPosition(vm, target, args)
		}
	case "filter":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayFilter(vm, target, args)
		}
	case "map":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayMap(vm, target, args)
		}
	case "fold":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayFold(vm, target, args)
		}
	case "sum":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArraySum(target, args, span)
		}
	case "product":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayProduct(target, args, span)
		}
	case "rev":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayRev(target, args)
		}
	case "join":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayJoin(target, args)
		}
	case "sorted":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArraySorted(vm, target, args)
		}
	case "zip":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayZip(target, args)
		}
	case "enumerate":
		fn = func(vm *Vm, args *Args) (Value, error) {
			return ArrayEnumerate(target, args)
		}
	case "dedup":
		fn = func(vm *Vm, args *Args) (Value, error) {
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

// ArrayLen returns the length of an array.
func ArrayLen(arr ArrayValue, args *Args) (Value, error) {
	return IntValue(len(arr)), nil
}

// ArrayFirst returns the first element of an array.
func ArrayFirst(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	if len(arr) == 0 {
		return nil, &ArrayEmptyError{Method: "first", Span: span}
	}
	return arr[0], nil
}

// ArrayLast returns the last element of an array.
func ArrayLast(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	if len(arr) == 0 {
		return nil, &ArrayEmptyError{Method: "last", Span: span}
	}
	return arr[len(arr)-1], nil
}

// ArrayAt returns the element at the given index.
func ArrayAt(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	indexSpanned := args.Eat()
	if indexSpanned == nil {
		return nil, &MissingArgumentError{What: "index", Span: span}
	}

	index, ok := indexSpanned.V.(IntValue)
	if !ok {
		return nil, &TypeError{Expected: TypeInt, Got: indexSpanned.V.Type(), Span: indexSpanned.Span}
	}

	// Handle negative indices
	idx := int(index)
	if idx < 0 {
		idx = len(arr) + idx
	}

	// Check for default parameter
	defaultVal := args.Find("default")

	if idx < 0 || idx >= len(arr) {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, &ArrayIndexError{Index: int(index), Length: len(arr), Span: span}
	}

	return arr[idx], nil
}

// ArrayPush adds an element to the end of an array.
func ArrayPush(arr *ArrayValue, args *Args) (Value, error) {
	val := args.Eat()
	if val == nil {
		return None, nil
	}
	*arr = append(*arr, val.V)
	return None, nil
}

// ArrayPop removes and returns the last element of an array.
func ArrayPop(arr *ArrayValue, args *Args, span syntax.Span) (Value, error) {
	if len(*arr) == 0 {
		return nil, &ArrayEmptyError{Method: "pop", Span: span}
	}
	last := (*arr)[len(*arr)-1]
	*arr = (*arr)[:len(*arr)-1]
	return last, nil
}

// ArrayInsert inserts an element at the given index.
func ArrayInsert(arr *ArrayValue, args *Args, span syntax.Span) (Value, error) {
	indexSpanned := args.Eat()
	if indexSpanned == nil {
		return nil, &MissingArgumentError{What: "index", Span: span}
	}
	valSpanned := args.Eat()
	if valSpanned == nil {
		return nil, &MissingArgumentError{What: "value", Span: span}
	}

	index, ok := indexSpanned.V.(IntValue)
	if !ok {
		return nil, &TypeError{Expected: TypeInt, Got: indexSpanned.V.Type(), Span: indexSpanned.Span}
	}

	idx := int(index)
	if idx < 0 || idx > len(*arr) {
		return nil, &ArrayIndexError{Index: idx, Length: len(*arr), Span: span}
	}

	*arr = append((*arr)[:idx], append(ArrayValue{valSpanned.V}, (*arr)[idx:]...)...)
	return None, nil
}

// ArrayRemove removes and returns the element at the given index.
func ArrayRemove(arr *ArrayValue, args *Args, span syntax.Span) (Value, error) {
	indexSpanned := args.Eat()
	if indexSpanned == nil {
		return nil, &MissingArgumentError{What: "index", Span: span}
	}

	index, ok := indexSpanned.V.(IntValue)
	if !ok {
		return nil, &TypeError{Expected: TypeInt, Got: indexSpanned.V.Type(), Span: indexSpanned.Span}
	}

	idx := int(index)
	if idx < 0 || idx >= len(*arr) {
		return nil, &ArrayIndexError{Index: idx, Length: len(*arr), Span: span}
	}

	removed := (*arr)[idx]
	*arr = append((*arr)[:idx], (*arr)[idx+1:]...)
	return removed, nil
}

// ArraySlice returns a slice of the array.
func ArraySlice(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	startSpanned := args.Eat()
	if startSpanned == nil {
		return nil, &MissingArgumentError{What: "start", Span: span}
	}

	startVal, ok := startSpanned.V.(IntValue)
	if !ok {
		return nil, &TypeError{Expected: TypeInt, Got: startSpanned.V.Type(), Span: startSpanned.Span}
	}

	start := int(startVal)
	if start < 0 {
		start = len(arr) + start
	}
	if start < 0 {
		start = 0
	}

	// Optional end parameter
	end := len(arr)
	if endSpanned := args.Eat(); endSpanned != nil {
		endVal, ok := endSpanned.V.(IntValue)
		if !ok {
			return nil, &TypeError{Expected: TypeInt, Got: endSpanned.V.Type(), Span: endSpanned.Span}
		}
		end = int(endVal)
		if end < 0 {
			end = len(arr) + end
		}
	}

	if start > len(arr) {
		start = len(arr)
	}
	if end > len(arr) {
		end = len(arr)
	}
	if start > end {
		return ArrayValue{}, nil
	}

	result := make(ArrayValue, end-start)
	copy(result, arr[start:end])
	return result, nil
}

// ArrayContains checks if the array contains a value.
func ArrayContains(arr ArrayValue, args *Args) (Value, error) {
	valSpanned := args.Eat()
	if valSpanned == nil {
		return Bool(false), nil
	}

	for _, elem := range arr {
		if Equal(elem, valSpanned.V) {
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

// ArrayFind finds the first element matching a predicate.
func ArrayFind(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnSpanned := args.Eat()
	if fnSpanned == nil {
		return None, nil
	}

	fn, ok := fnSpanned.V.(FuncValue)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnSpanned.V.Type(), Span: fnSpanned.Span}
	}

	for _, elem := range arr {
		result, err := callArrayFunc(vm, fn, elem)
		if err != nil {
			return nil, err
		}
		if b, ok := result.(BoolValue); ok && bool(b) {
			return elem, nil
		}
	}
	return None, nil
}

// ArrayPosition finds the index of the first element matching a predicate.
func ArrayPosition(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnSpanned := args.Eat()
	if fnSpanned == nil {
		return None, nil
	}

	fn, ok := fnSpanned.V.(FuncValue)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnSpanned.V.Type(), Span: fnSpanned.Span}
	}

	for i, elem := range arr {
		result, err := callArrayFunc(vm, fn, elem)
		if err != nil {
			return nil, err
		}
		if b, ok := result.(BoolValue); ok && bool(b) {
			return IntValue(i), nil
		}
	}
	return None, nil
}

// ArrayFilter filters elements matching a predicate.
func ArrayFilter(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnSpanned := args.Eat()
	if fnSpanned == nil {
		return ArrayValue{}, nil
	}

	fn, ok := fnSpanned.V.(FuncValue)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnSpanned.V.Type(), Span: fnSpanned.Span}
	}

	var result ArrayValue
	for _, elem := range arr {
		predResult, err := callArrayFunc(vm, fn, elem)
		if err != nil {
			return nil, err
		}
		if b, ok := predResult.(BoolValue); ok && bool(b) {
			result = append(result, elem)
		}
	}
	return result, nil
}

// ArrayMap applies a function to each element.
func ArrayMap(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	fnSpanned := args.Eat()
	if fnSpanned == nil {
		return ArrayValue{}, nil
	}

	fn, ok := fnSpanned.V.(FuncValue)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnSpanned.V.Type(), Span: fnSpanned.Span}
	}

	result := make(ArrayValue, len(arr))
	for i, elem := range arr {
		mapped, err := callArrayFunc(vm, fn, elem)
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

// ArrayFold reduces the array to a single value.
func ArrayFold(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	initSpanned := args.Eat()
	if initSpanned == nil {
		return None, nil
	}

	fnSpanned := args.Eat()
	if fnSpanned == nil {
		return None, nil
	}

	fn, ok := fnSpanned.V.(FuncValue)
	if !ok {
		return nil, &TypeError{Expected: TypeFunc, Got: fnSpanned.V.Type(), Span: fnSpanned.Span}
	}

	acc := initSpanned.V
	for _, elem := range arr {
		result, err := callArrayFuncWithArgs(vm, fn, acc, elem)
		if err != nil {
			return nil, err
		}
		acc = result
	}
	return acc, nil
}

// ArraySum returns the sum of all elements.
func ArraySum(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	// Check for default parameter
	defaultVal := args.Find("default")

	if len(arr) == 0 {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return IntValue(0), nil
	}

	// Sum all elements
	var sum Value = IntValue(0)
	for _, elem := range arr {
		result, err := Add(sum, elem, span)
		if err != nil {
			return nil, err
		}
		sum = result
	}
	return sum, nil
}

// ArrayProduct returns the product of all elements.
func ArrayProduct(arr ArrayValue, args *Args, span syntax.Span) (Value, error) {
	// Check for default parameter
	defaultVal := args.Find("default")

	if len(arr) == 0 {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return IntValue(1), nil
	}

	// Product of all elements
	var product Value = IntValue(1)
	for _, elem := range arr {
		result, err := Mul(product, elem, span)
		if err != nil {
			return nil, err
		}
		product = result
	}
	return product, nil
}

// ArrayRev returns the array in reverse order.
func ArrayRev(arr ArrayValue, args *Args) (Value, error) {
	result := make(ArrayValue, len(arr))
	for i, elem := range arr {
		result[len(arr)-1-i] = elem
	}
	return result, nil
}

// ArrayJoin joins array elements with a separator.
func ArrayJoin(arr ArrayValue, args *Args) (Value, error) {
	sep := StrValue("")
	if sepSpanned := args.Eat(); sepSpanned != nil {
		if s, ok := sepSpanned.V.(StrValue); ok {
			sep = s
		}
	}

	var result string
	for i, elem := range arr {
		if i > 0 {
			result += string(sep)
		}
		if s, ok := elem.(StrValue); ok {
			result += string(s)
		} else {
			// Convert to string representation
			result += fmt.Sprintf("%v", elem)
		}
	}
	return StrValue(result), nil
}

// ArraySorted returns a sorted copy of the array.
func ArraySorted(vm *Vm, arr ArrayValue, args *Args) (Value, error) {
	result := make(ArrayValue, len(arr))
	copy(result, arr)

	// Simple bubble sort for now (can be optimized)
	var sortErr error
	for i := 0; i < len(result) && sortErr == nil; i++ {
		for j := i + 1; j < len(result) && sortErr == nil; j++ {
			// Compare elements
			cmp, err := compareValues(result[i], result[j])
			if err != nil {
				sortErr = err
				break
			}
			if cmp > 0 {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	if sortErr != nil {
		return nil, sortErr
	}
	return result, nil
}

// ArrayZip zips this array with another array.
func ArrayZip(arr ArrayValue, args *Args) (Value, error) {
	otherSpanned := args.Eat()
	if otherSpanned == nil {
		return ArrayValue{}, nil
	}

	other, ok := otherSpanned.V.(ArrayValue)
	if !ok {
		return nil, &TypeError{Expected: TypeArray, Got: otherSpanned.V.Type(), Span: otherSpanned.Span}
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

// ArrayEnumerate returns an array of (index, value) pairs.
func ArrayEnumerate(arr ArrayValue, args *Args) (Value, error) {
	result := make(ArrayValue, len(arr))
	for i, elem := range arr {
		result[i] = ArrayValue{IntValue(i), elem}
	}
	return result, nil
}

// ArrayDedup removes duplicate elements.
func ArrayDedup(arr ArrayValue, args *Args) (Value, error) {
	var result ArrayValue
	seen := make(map[string]bool)

	for _, elem := range arr {
		// Use string representation as key (simple approach)
		key := fmt.Sprintf("%v", elem)
		if !seen[key] {
			seen[key] = true
			result = append(result, elem)
		}
	}
	return result, nil
}

// callArrayFunc calls a function with one argument.
func callArrayFunc(vm *Vm, fn FuncValue, arg Value) (Value, error) {
	args := &Args{
		Items: []Arg{{Value: syntax.Spanned[Value]{V: arg}}},
	}
	return callFunc(vm, fn, args, syntax.Detached())
}

// callArrayFuncWithArgs calls a function with two arguments.
func callArrayFuncWithArgs(vm *Vm, fn FuncValue, arg1, arg2 Value) (Value, error) {
	args := &Args{
		Items: []Arg{
			{Value: syntax.Spanned[Value]{V: arg1}},
			{Value: syntax.Spanned[Value]{V: arg2}},
		},
	}
	return callFunc(vm, fn, args, syntax.Detached())
}

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
	Index  int
	Length int
	Span   syntax.Span
}

func (e *ArrayIndexError) Error() string {
	return fmt.Sprintf("array index out of bounds (index: %d, len: %d) and no default value was specified", e.Index, e.Length)
}
