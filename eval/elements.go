package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Element Functions
// ----------------------------------------------------------------------------

// This file contains element constructor functions that can be called from
// Typst code to create content elements programmatically.
// For example: raw("print('hello')", lang: "python")

// RawFunc creates the raw element function.
func RawFunc() *Func {
	name := "raw"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: rawNative,
			Info: &FuncInfo{
				Name: "raw",
				Params: []ParamInfo{
					{Name: "text", Type: TypeStr, Named: false},
					{Name: "block", Type: TypeBool, Default: False, Named: true},
					{Name: "lang", Type: TypeStr, Default: None, Named: true},
				},
			},
		},
	}
}

// rawNative implements the raw() function.
// Creates a RawElement from the given text, with optional language and block parameters.
//
// Arguments:
//   - text (positional, str): The raw text content
//   - block (named, bool, default: false): Whether this is a block-level element
//   - lang (named, str or none, default: none): The syntax highlighting language
func rawNative(vm *Vm, args *Args) (Value, error) {
	// Get required text argument (can be positional or named)
	textArg := args.Find("text")
	if textArg == nil {
		textArgSpanned, err := args.Expect("text")
		if err != nil {
			return nil, err
		}
		textArg = &textArgSpanned
	}

	text, ok := AsStr(textArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      textArg.V.Type().String(),
			Span:     textArg.Span,
		}
	}

	// Get optional block argument (default: false)
	block := false
	if blockArg := args.Find("block"); blockArg != nil {
		blockVal, ok := AsBool(blockArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      blockArg.V.Type().String(),
				Span:     blockArg.Span,
			}
		}
		block = blockVal
	}

	// Get optional lang argument (default: none/empty string)
	lang := ""
	if langArg := args.Find("lang"); langArg != nil {
		if !IsNone(langArg.V) {
			langStr, ok := AsStr(langArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      langArg.V.Type().String(),
					Span:     langArg.Span,
				}
			}
			lang = langStr
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the RawElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&RawElement{
			Text:  text,
			Lang:  lang,
			Block: block,
		}},
	}}, nil
}

// HeadingFunc creates the heading element function.
// Usage: #heading(body, level: 1, numbering: none, outlined: true, bookmarked: auto)
func HeadingFunc() *Func {
	name := "heading"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: headingNative,
			Info: &FuncInfo{
				Name: "heading",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "level", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "outlined", Type: TypeBool, Default: True, Named: true},
					{Name: "bookmarked", Type: TypeAuto, Default: Auto, Named: true},
					{Name: "supplement", Type: TypeContent, Default: None, Named: true},
				},
			},
		},
	}
}

// headingNative implements the heading() function.
// Creates a HeadingElement from the given body content with optional level and style parameters.
//
// Arguments:
//   - body (positional, content): The heading content
//   - level (named, int, default: 1): The heading level (1-6)
//   - numbering (named, str or none, default: none): Numbering pattern (e.g., "1.", "1.1", "I.A")
//   - outlined (named, bool, default: true): Whether to include in document outline
//   - bookmarked (named, bool or auto, default: auto): Whether to include in PDF bookmarks
//   - supplement (named, content or none, default: none): Supplement for references
func headingNative(vm *Vm, args *Args) (Value, error) {
	// Get required body argument (can be positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArgSpanned, err := args.Expect("body")
		if err != nil {
			return nil, err
		}
		bodyArg = &bodyArgSpanned
	}

	var bodyContent Content
	switch v := bodyArg.V.(type) {
	case ContentValue:
		bodyContent = v.Content
	case StrValue:
		// Allow string as body, convert to text content
		bodyContent = Content{Elements: []ContentElement{&TextElement{Text: string(v)}}}
	default:
		return nil, &TypeMismatchError{
			Expected: "content or string",
			Got:      bodyArg.V.Type().String(),
			Span:     bodyArg.Span,
		}
	}

	// Get optional level argument (default: 1)
	level := int64(1)
	if levelArg := args.Find("level"); levelArg != nil {
		levelVal, ok := AsInt(levelArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "int",
				Got:      levelArg.V.Type().String(),
				Span:     levelArg.Span,
			}
		}
		level = levelVal
		// Clamp level to valid range
		if level < 1 {
			level = 1
		} else if level > 6 {
			level = 6
		}
	}

	// Get optional numbering argument (default: none/empty string)
	numbering := ""
	if numberingArg := args.Find("numbering"); numberingArg != nil {
		if !IsNone(numberingArg.V) {
			numberingStr, ok := AsStr(numberingArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      numberingArg.V.Type().String(),
					Span:     numberingArg.Span,
				}
			}
			numbering = numberingStr
		}
	}

	// Get optional outlined argument (default: true)
	outlined := true
	if outlinedArg := args.Find("outlined"); outlinedArg != nil {
		outlinedVal, ok := AsBool(outlinedArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      outlinedArg.V.Type().String(),
				Span:     outlinedArg.Span,
			}
		}
		outlined = outlinedVal
	}

	// Get optional bookmarked argument (default: auto/nil)
	var bookmarked *bool
	if bookmarkedArg := args.Find("bookmarked"); bookmarkedArg != nil {
		if !IsAuto(bookmarkedArg.V) {
			bookmarkedVal, ok := AsBool(bookmarkedArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool or auto",
					Got:      bookmarkedArg.V.Type().String(),
					Span:     bookmarkedArg.Span,
				}
			}
			bookmarked = &bookmarkedVal
		}
	}

	// Get optional supplement argument (default: none/nil)
	var supplement *Content
	if supplementArg := args.Find("supplement"); supplementArg != nil {
		if !IsNone(supplementArg.V) {
			if c, ok := supplementArg.V.(ContentValue); ok {
				supplement = &c.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      supplementArg.V.Type().String(),
					Span:     supplementArg.Span,
				}
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the HeadingElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&HeadingElement{
			Level:      int(level),
			Content:    bodyContent,
			Numbering:  numbering,
			Outlined:   outlined,
			Bookmarked: bookmarked,
			Supplement: supplement,
		}},
	}}, nil
}

// ----------------------------------------------------------------------------
// Library Registration
// ----------------------------------------------------------------------------

// RegisterElementFunctions registers all element functions in the given scope.
// Call this when setting up the standard library scope.
func RegisterElementFunctions(scope *Scope) {
	// Register raw element function
	scope.DefineFunc("raw", RawFunc())
	// Register heading element function
	scope.DefineFunc("heading", HeadingFunc())
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		"raw":     RawFunc(),
		"heading": HeadingFunc(),
	}
}
