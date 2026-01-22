// Package text provides text-related functions for the Typst standard library.
package text

import (
	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/syntax"
)

// RawFunc creates a raw function that can be used to create raw text/code elements.
// Usage: #raw("code", lang: "python", block: false)
func RawFunc() eval.FuncValue {
	name := "raw"
	return eval.FuncValue{
		Func: &eval.Func{
			Name: &name,
			Span: syntax.Detached(),
			Repr: eval.NativeFunc{
				Func: rawImpl,
				Info: &eval.FuncInfo{
					Name: "raw",
					Params: []eval.ParamInfo{
						{Name: "text", Type: eval.TypeStr, Named: false},
						{Name: "lang", Type: eval.TypeStr, Default: eval.None, Named: true},
						{Name: "block", Type: eval.TypeBool, Default: eval.False, Named: true},
						{Name: "align", Type: eval.TypeAuto, Default: eval.Auto, Named: true},
						{Name: "syntaxes", Type: eval.TypeAuto, Default: eval.Auto, Named: true},
						{Name: "theme", Type: eval.TypeAuto, Default: eval.Auto, Named: true},
					},
				},
			},
		},
	}
}

// rawImpl is the native implementation of the raw function.
func rawImpl(engine *eval.Engine, context *eval.Context, args *eval.Args) (eval.Value, error) {
	// Get the required text argument (first positional)
	textArg, err := args.Expect("text")
	if err != nil {
		return nil, err
	}

	text := ""
	if s, ok := textArg.V.(eval.StrValue); ok {
		text = string(s)
	} else {
		return nil, &eval.TypeMismatchError{
			Expected: "str",
			Got:      textArg.V.Type().String(),
			Span:     textArg.Span,
		}
	}

	// Get optional lang argument
	lang := ""
	if langArg := args.Find("lang"); langArg != nil {
		if s, ok := langArg.V.(eval.StrValue); ok {
			lang = string(s)
		}
		// none is also valid, leave lang as empty string
	}

	// Get optional block argument
	block := false
	if blockArg := args.Find("block"); blockArg != nil {
		if b, ok := blockArg.V.(eval.BoolValue); ok {
			block = bool(b)
		}
	}

	// Consume remaining arguments to check for unexpected ones
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create and return the raw element as content
	element := &eval.RawElement{
		Text:  text,
		Lang:  lang,
		Block: block,
	}

	return eval.ContentValue{
		Content: eval.Content{
			Elements: []eval.ContentElement{element},
		},
	}, nil
}

// RawLang returns the language of a raw element.
// This is a method that can be called on raw content: raw-content.lang
func RawLang(element *eval.RawElement) eval.Value {
	if element.Lang == "" {
		return eval.None
	}
	return eval.StrValue(element.Lang)
}

// RawText returns the text content of a raw element.
func RawText(element *eval.RawElement) eval.Value {
	return eval.StrValue(element.Text)
}

// RawBlock returns whether the raw element is a block.
func RawBlock(element *eval.RawElement) eval.Value {
	return eval.BoolValue(element.Block)
}

// RawLines returns the lines of a raw element as an array.
func RawLines(element *eval.RawElement) eval.Value {
	var lines []eval.Value
	start := 0
	for i, c := range element.Text {
		if c == '\n' {
			lines = append(lines, eval.StrValue(element.Text[start:i]))
			start = i + 1
		}
	}
	// Add last line if not empty or if there was content
	if start <= len(element.Text) {
		lines = append(lines, eval.StrValue(element.Text[start:]))
	}
	return eval.ArrayValue(lines)
}
