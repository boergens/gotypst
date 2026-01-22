package eval

import (
	"regexp"
	"strings"

	"github.com/boergens/gotypst/syntax"
)

// StrContains checks if a string contains a pattern.
// Pattern can be a string (literal match) or a regex.
func StrContains(target StrValue, args *Args) (Value, error) {
	pattern, err := args.Expect("pattern")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch p := pattern.V.(type) {
	case StrValue:
		return Bool(strings.Contains(string(target), string(p))), nil
	default:
		return nil, &StrMethodError{
			Method:  "contains",
			Message: "expected string pattern, got " + pattern.V.Type().String(),
			Span:    pattern.Span,
		}
	}
}

// StrStartsWith checks if a string starts with a pattern.
func StrStartsWith(target StrValue, args *Args) (Value, error) {
	pattern, err := args.Expect("pattern")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch p := pattern.V.(type) {
	case StrValue:
		return Bool(strings.HasPrefix(string(target), string(p))), nil
	default:
		return nil, &StrMethodError{
			Method:  "starts-with",
			Message: "expected string pattern, got " + pattern.V.Type().String(),
			Span:    pattern.Span,
		}
	}
}

// StrEndsWith checks if a string ends with a pattern.
func StrEndsWith(target StrValue, args *Args) (Value, error) {
	pattern, err := args.Expect("pattern")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch p := pattern.V.(type) {
	case StrValue:
		return Bool(strings.HasSuffix(string(target), string(p))), nil
	default:
		return nil, &StrMethodError{
			Method:  "ends-with",
			Message: "expected string pattern, got " + pattern.V.Type().String(),
			Span:    pattern.Span,
		}
	}
}

// StrFind finds the first occurrence of a pattern in the string.
// Returns the matched substring or none if not found.
func StrFind(target StrValue, args *Args) (Value, error) {
	pattern, err := args.Expect("pattern")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch p := pattern.V.(type) {
	case StrValue:
		idx := strings.Index(string(target), string(p))
		if idx == -1 {
			return None, nil
		}
		return Str(string(p)), nil
	default:
		return nil, &StrMethodError{
			Method:  "find",
			Message: "expected string pattern, got " + pattern.V.Type().String(),
			Span:    pattern.Span,
		}
	}
}

// StrPosition finds the position of the first occurrence of a pattern.
// Returns the byte position or none if not found.
func StrPosition(target StrValue, args *Args) (Value, error) {
	pattern, err := args.Expect("pattern")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch p := pattern.V.(type) {
	case StrValue:
		idx := strings.Index(string(target), string(p))
		if idx == -1 {
			return None, nil
		}
		return Int(int64(idx)), nil
	default:
		return nil, &StrMethodError{
			Method:  "position",
			Message: "expected string pattern, got " + pattern.V.Type().String(),
			Span:    pattern.Span,
		}
	}
}

// StrMatch finds the first match of a pattern and returns match details.
// Returns a dictionary with start, end, text, and captures, or none if not found.
func StrMatch(target StrValue, args *Args) (Value, error) {
	pattern, err := args.Expect("pattern")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch p := pattern.V.(type) {
	case StrValue:
		// For plain strings, return a simple match
		idx := strings.Index(string(target), string(p))
		if idx == -1 {
			return None, nil
		}
		result := NewDict()
		result.Set("start", Int(int64(idx)))
		result.Set("end", Int(int64(idx+len(p))))
		result.Set("text", Str(string(p)))
		result.Set("captures", ArrayValue{})
		return result, nil
	default:
		return nil, &StrMethodError{
			Method:  "match",
			Message: "expected string pattern, got " + pattern.V.Type().String(),
			Span:    pattern.Span,
		}
	}
}

// StrMatches finds all matches of a pattern in the string.
// Returns an array of match dictionaries.
func StrMatches(target StrValue, args *Args) (Value, error) {
	pattern, err := args.Expect("pattern")
	if err != nil {
		return nil, err
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch p := pattern.V.(type) {
	case StrValue:
		// For plain strings, find all non-overlapping occurrences
		targetStr := string(target)
		patternStr := string(p)
		var matches ArrayValue

		if len(patternStr) == 0 {
			// Empty pattern matches at every position including start and end
			for i := 0; i <= len(targetStr); i++ {
				match := NewDict()
				match.Set("start", Int(int64(i)))
				match.Set("end", Int(int64(i)))
				match.Set("text", Str(""))
				match.Set("captures", ArrayValue{})
				matches = append(matches, match)
			}
			return matches, nil
		}

		idx := 0
		for {
			pos := strings.Index(targetStr[idx:], patternStr)
			if pos == -1 {
				break
			}
			absPos := idx + pos
			match := NewDict()
			match.Set("start", Int(int64(absPos)))
			match.Set("end", Int(int64(absPos+len(patternStr))))
			match.Set("text", Str(patternStr))
			match.Set("captures", ArrayValue{})
			matches = append(matches, match)
			idx = absPos + len(patternStr)
		}
		return matches, nil
	default:
		return nil, &StrMethodError{
			Method:  "matches",
			Message: "expected string pattern, got " + pattern.V.Type().String(),
			Span:    pattern.Span,
		}
	}
}

// StrLen returns the length of the string in graphemes/characters.
func StrLen(target StrValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}
	// Count runes (Unicode code points) as a simple approximation of graphemes
	count := 0
	for range string(target) {
		count++
	}
	return Int(int64(count)), nil
}

// StrFirst returns the first character of the string.
func StrFirst(target StrValue, args *Args, span syntax.Span) (Value, error) {
	// Check for 'default' named argument
	defaultVal := args.Find("default")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	s := string(target)
	if len(s) == 0 {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, &StrMethodError{
			Method:  "first",
			Message: "string is empty",
			Span:    span,
		}
	}

	// Return first rune as string
	for _, r := range s {
		return Str(string(r)), nil
	}
	return None, nil
}

// StrLast returns the last character of the string.
func StrLast(target StrValue, args *Args, span syntax.Span) (Value, error) {
	// Check for 'default' named argument
	defaultVal := args.Find("default")

	if err := args.Finish(); err != nil {
		return nil, err
	}

	s := string(target)
	if len(s) == 0 {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, &StrMethodError{
			Method:  "last",
			Message: "string is empty",
			Span:    span,
		}
	}

	// Get last rune
	var lastRune rune
	for _, r := range s {
		lastRune = r
	}
	return Str(string(lastRune)), nil
}

// StrAt returns the character at a given index.
func StrAt(target StrValue, args *Args, span syntax.Span) (Value, error) {
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
		return nil, &StrMethodError{
			Method:  "at",
			Message: "expected integer index, got " + indexArg.V.Type().String(),
			Span:    indexArg.Span,
		}
	}

	s := string(target)
	runes := []rune(s)
	length := len(runes)

	// Handle negative indices
	if idx < 0 {
		idx = int64(length) + idx
	}

	if idx < 0 || idx >= int64(length) {
		if defaultVal != nil {
			return defaultVal.V, nil
		}
		return nil, &StrMethodError{
			Method:  "at",
			Message: "no default value was specified and string index out of bounds (index: " + string(rune('0'+idx)) + ", len: " + string(rune('0'+int64(length))) + ")",
			Span:    span,
		}
	}

	return Str(string(runes[idx])), nil
}

// StrSlice returns a substring from start to end index.
func StrSlice(target StrValue, args *Args, span syntax.Span) (Value, error) {
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
		return nil, &StrMethodError{
			Method:  "slice",
			Message: "expected integer start index, got " + startArg.V.Type().String(),
			Span:    startArg.Span,
		}
	}

	s := string(target)
	runes := []rune(s)
	length := int64(len(runes))

	// Handle negative start index
	if startIdx < 0 {
		startIdx = length + startIdx
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > length {
		startIdx = length
	}

	endIdx := length
	if endArg != nil {
		end, ok := AsInt(endArg.V)
		if !ok {
			return nil, &StrMethodError{
				Method:  "slice",
				Message: "expected integer end index, got " + endArg.V.Type().String(),
				Span:    endArg.Span,
			}
		}
		// Handle negative end index
		if end < 0 {
			end = length + end
		}
		endIdx = end
	}

	if endIdx < 0 {
		endIdx = 0
	}
	if endIdx > length {
		endIdx = length
	}

	if startIdx > endIdx {
		return Str(""), nil
	}

	return Str(string(runes[startIdx:endIdx])), nil
}

// StrSplit splits the string by a pattern.
func StrSplit(target StrValue, args *Args) (Value, error) {
	patternArg := args.Eat()

	if err := args.Finish(); err != nil {
		return nil, err
	}

	s := string(target)

	if patternArg == nil {
		// No pattern - split on whitespace
		parts := strings.Fields(s)
		result := make(ArrayValue, len(parts))
		for i, part := range parts {
			result[i] = Str(part)
		}
		return result, nil
	}

	switch p := patternArg.V.(type) {
	case StrValue:
		patternStr := string(p)
		if patternStr == "" {
			// Split into individual characters with empty strings at boundaries
			// "abc".split("") returns ("", "a", "b", "c", "")
			runes := []rune(s)
			result := make(ArrayValue, len(runes)+2)
			result[0] = Str("")
			for i, r := range runes {
				result[i+1] = Str(string(r))
			}
			result[len(runes)+1] = Str("")
			return result, nil
		}
		parts := strings.Split(s, patternStr)
		result := make(ArrayValue, len(parts))
		for i, part := range parts {
			result[i] = Str(part)
		}
		return result, nil
	default:
		return nil, &StrMethodError{
			Method:  "split",
			Message: "expected string pattern, got " + patternArg.V.Type().String(),
			Span:    patternArg.Span,
		}
	}
}

// StrRev reverses the string.
func StrRev(target StrValue, args *Args) (Value, error) {
	if err := args.Finish(); err != nil {
		return nil, err
	}

	runes := []rune(string(target))
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return Str(string(runes)), nil
}

// RegexValue represents a compiled regular expression.
type RegexValue struct {
	Pattern string
	Regex   *regexp.Regexp
}

func (RegexValue) Type() Type       { return TypeRegex }
func (v RegexValue) Display() Content { return Content{} }
func (v RegexValue) Clone() Value   { return v }
func (RegexValue) isValue()         {}

// TypeRegex is the type for regex values.
const TypeRegex Type = 100

// StrMethodError is returned when a string method fails.
type StrMethodError struct {
	Method  string
	Message string
	Span    syntax.Span
}

func (e *StrMethodError) Error() string {
	return e.Method + ": " + e.Message
}

// createStrMethod creates a bound method for a string.
func createStrMethod(target StrValue, methodName string, span syntax.Span) Value {
	var fn func(engine *Engine, context *Context, args *Args) (Value, error)

	switch methodName {
	case "contains":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrContains(target, args)
		}
	case "starts-with":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrStartsWith(target, args)
		}
	case "ends-with":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrEndsWith(target, args)
		}
	case "find":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrFind(target, args)
		}
	case "position":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrPosition(target, args)
		}
	case "match":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrMatch(target, args)
		}
	case "matches":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrMatches(target, args)
		}
	case "len":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrLen(target, args)
		}
	case "first":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrFirst(target, args, span)
		}
	case "last":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrLast(target, args, span)
		}
	case "at":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrAt(target, args, span)
		}
	case "slice":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrSlice(target, args, span)
		}
	case "split":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrSplit(target, args)
		}
	case "rev":
		fn = func(engine *Engine, context *Context, args *Args) (Value, error) {
			return StrRev(target, args)
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

// GetStrMethod returns a bound method for a string value.
func GetStrMethod(target StrValue, methodName string, span syntax.Span) Value {
	return createStrMethod(target, methodName, span)
}
