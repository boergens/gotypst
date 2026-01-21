package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Text Element Functions
// ----------------------------------------------------------------------------
// These functions mirror Typst's text module:
// - text() - styles text content (applies styles, doesn't create new element)
// - strong() - bold emphasis (creates StrongElement)
// - emph() - italic emphasis (creates EmphElement)
//
// Reference: typst-reference/crates/typst-library/src/text/mod.rs
//            typst-reference/crates/typst-library/src/model/strong.rs
//            typst-reference/crates/typst-library/src/model/emph.rs

// ----------------------------------------------------------------------------
// StyledElement - Content with attached styles
// ----------------------------------------------------------------------------

// StyledElement wraps content with attached styles.
// This matches Rust's StyledElem in typst-library/src/foundations/content.
// When the layout/realization system encounters this element, it should
// apply the styles to a style chain before processing the child content.
type StyledElement struct {
	// Child is the content being styled.
	Child Content
	// Styles are the styles to apply to the child content.
	Styles *Styles
}

func (*StyledElement) IsContentElement() {}

// StyledWithMap wraps content with a style map.
// This is the Go equivalent of Rust's Content::styled_with_map.
func StyledWithMap(content Content, styles *Styles) Content {
	if styles == nil || (len(styles.Rules) == 0 && len(styles.Recipes) == 0) {
		return content
	}
	return Content{
		Elements: []ContentElement{&StyledElement{
			Child:  content,
			Styles: styles,
		}},
	}
}

// ----------------------------------------------------------------------------
// text() function
// ----------------------------------------------------------------------------

// TextFunc creates the text element function.
// Matches Typst's text() which applies styling to text content.
//
// IMPORTANT: In Typst, text() doesn't create a new element type. It collects
// style properties and wraps the body content with those styles using
// styled_with_map. This allows styles to properly cascade through the
// style chain system.
//
// Reference: typst-reference/crates/typst-library/src/text/mod.rs
func TextFunc() *Func {
	name := "text"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: textNative,
			Info: &FuncInfo{
				Name: "text",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "size", Type: TypeLength, Default: None, Named: true},
					{Name: "font", Type: TypeStr, Default: None, Named: true},
					{Name: "weight", Type: TypeInt, Default: None, Named: true},
					{Name: "style", Type: TypeStr, Default: None, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
				},
			},
		},
	}
}

// textNative implements the text() function.
// Collects style properties and wraps the body content with those styles.
//
// This matches Rust's TextElem::construct which does:
//   let styles = Self::set(engine, args)?;
//   let body = args.expect::<Content>("body")?;
//   Ok(body.styled_with_map(styles))
//
// Arguments:
//   - body (positional, content): The content to style
//   - size (named, length, default: none): Font size
//   - font (named, str, default: none): Font family
//   - weight (named, int, default: none): Font weight (100-900)
//   - style (named, str, default: none): Font style ("normal", "italic", "oblique")
//   - fill (named, color, default: none): Text color
//
// Reference: typst-reference/crates/typst-library/src/text/mod.rs
func textNative(vm *Vm, args *Args) (Value, error) {
	// Get required body argument
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArgSpanned, err := args.Expect("body")
		if err != nil {
			return nil, err
		}
		bodyArg = &bodyArgSpanned
	}

	var body Content
	if cv, ok := bodyArg.V.(ContentValue); ok {
		body = cv.Content
	} else if sv, ok := bodyArg.V.(StrValue); ok {
		// text() can take a string directly - convert to TextElement
		body = Content{Elements: []ContentElement{&TextElement{Text: string(sv)}}}
	} else {
		return nil, &TypeMismatchError{
			Expected: "content or string",
			Got:      bodyArg.V.Type().String(),
			Span:     bodyArg.Span,
		}
	}

	// Collect style arguments into a slice of Arg
	// This mimics Rust's Self::set(engine, args)
	var styleItems []Arg

	// Get optional size argument
	if sizeArg := args.Find("size"); sizeArg != nil {
		if !IsNone(sizeArg.V) && !IsAuto(sizeArg.V) {
			name := "size"
			styleItems = append(styleItems, Arg{
				Span:  sizeArg.Span,
				Name:  &name,
				Value: *sizeArg,
			})
		}
	}

	// Get optional font argument
	if fontArg := args.Find("font"); fontArg != nil {
		if !IsNone(fontArg.V) && !IsAuto(fontArg.V) {
			name := "font"
			styleItems = append(styleItems, Arg{
				Span:  fontArg.Span,
				Name:  &name,
				Value: *fontArg,
			})
		}
	}

	// Get optional weight argument
	if weightArg := args.Find("weight"); weightArg != nil {
		if !IsNone(weightArg.V) && !IsAuto(weightArg.V) {
			name := "weight"
			// Handle named weights like "bold", "normal"
			var weightValue Value
			if ws, ok := AsStr(weightArg.V); ok {
				weightValue = IntValue(parseNamedWeight(ws))
			} else {
				weightValue = weightArg.V
			}
			styleItems = append(styleItems, Arg{
				Span: weightArg.Span,
				Name: &name,
				Value: syntax.Spanned[Value]{
					V:    weightValue,
					Span: weightArg.Span,
				},
			})
		}
	}

	// Get optional style argument
	if styleArg := args.Find("style"); styleArg != nil {
		if !IsNone(styleArg.V) && !IsAuto(styleArg.V) {
			if sv, ok := AsStr(styleArg.V); ok {
				if sv != "normal" && sv != "italic" && sv != "oblique" {
					return nil, &TypeMismatchError{
						Expected: "\"normal\", \"italic\", or \"oblique\"",
						Got:      "\"" + sv + "\"",
						Span:     styleArg.Span,
					}
				}
			}
			name := "style"
			styleItems = append(styleItems, Arg{
				Span:  styleArg.Span,
				Name:  &name,
				Value: *styleArg,
			})
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) && !IsAuto(fillArg.V) {
			name := "fill"
			styleItems = append(styleItems, Arg{
				Span:  fillArg.Span,
				Name:  &name,
				Value: *fillArg,
			})
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// If no style arguments, just return the body unchanged
	if len(styleItems) == 0 {
		return ContentValue{Content: body}, nil
	}

	// Create a style rule for the text element
	textFuncName := "text"
	styleArgs := NewArgsFrom(args.Span, styleItems)
	styles := &Styles{
		Rules: []StyleRule{{
			Func: &Func{
				Name: &textFuncName,
				Span: args.Span,
			},
			Args:     styleArgs,
			Span:     args.Span,
			Liftable: false, // Constructor-applied styles don't lift to page level
		}},
	}

	// Wrap body with styles (matches Rust's body.styled_with_map(styles))
	styledContent := StyledWithMap(body, styles)

	return ContentValue{Content: styledContent}, nil
}

// ----------------------------------------------------------------------------
// strong() function
// ----------------------------------------------------------------------------

// StrongFunc creates the strong (bold) element function.
// Matches Typst's strong() which increases font weight by delta (default 300).
//
// Unlike text(), strong() DOES create a new element (StrongElement) because
// it's a semantic element that can be targeted by show rules.
//
// Reference: typst-reference/crates/typst-library/src/model/strong.rs
func StrongFunc() *Func {
	name := "strong"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: strongNative,
			Info: &FuncInfo{
				Name: "strong",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "delta", Type: TypeInt, Default: IntValue(300), Named: true},
				},
			},
		},
	}
}

// strongNative implements the strong() function.
// Creates a StrongElement that increases font weight.
//
// Arguments:
//   - body (positional, content): The content to emphasize
//   - delta (named, int, default: 300): The font weight increase
//
// Reference: typst-reference/crates/typst-library/src/model/strong.rs
func strongNative(vm *Vm, args *Args) (Value, error) {
	// Get required body argument
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArgSpanned, err := args.Expect("body")
		if err != nil {
			return nil, err
		}
		bodyArg = &bodyArgSpanned
	}

	var body Content
	if cv, ok := bodyArg.V.(ContentValue); ok {
		body = cv.Content
	} else {
		return nil, &TypeMismatchError{
			Expected: "content",
			Got:      bodyArg.V.Type().String(),
			Span:     bodyArg.Span,
		}
	}

	// Get optional delta argument (default: 300)
	delta := int64(300)
	if deltaArg := args.Find("delta"); deltaArg != nil {
		if d, ok := AsInt(deltaArg.V); ok {
			delta = d
		} else {
			return nil, &TypeMismatchError{
				Expected: "integer",
				Got:      deltaArg.V.Type().String(),
				Span:     deltaArg.Span,
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the StrongElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&StrongElement{
			Content: body,
			Delta:   delta,
		}},
	}}, nil
}

// ----------------------------------------------------------------------------
// emph() function
// ----------------------------------------------------------------------------

// EmphFunc creates the emph (italic) element function.
// Matches Typst's emph() which toggles font style to italic.
//
// Unlike text(), emph() DOES create a new element (EmphElement) because
// it's a semantic element that can be targeted by show rules.
//
// Reference: typst-reference/crates/typst-library/src/model/emph.rs
func EmphFunc() *Func {
	name := "emph"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: emphNative,
			Info: &FuncInfo{
				Name: "emph",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// emphNative implements the emph() function.
// Creates an EmphElement that toggles italic style.
//
// Arguments:
//   - body (positional, content): The content to emphasize
//
// Reference: typst-reference/crates/typst-library/src/model/emph.rs
func emphNative(vm *Vm, args *Args) (Value, error) {
	// Get required body argument
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArgSpanned, err := args.Expect("body")
		if err != nil {
			return nil, err
		}
		bodyArg = &bodyArgSpanned
	}

	var body Content
	if cv, ok := bodyArg.V.(ContentValue); ok {
		body = cv.Content
	} else {
		return nil, &TypeMismatchError{
			Expected: "content",
			Got:      bodyArg.V.Type().String(),
			Span:     bodyArg.Span,
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the EmphElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&EmphElement{Content: body}},
	}}, nil
}

// ----------------------------------------------------------------------------
// raw() function
// ----------------------------------------------------------------------------

// RawFunc creates the raw element function.
// Matches Typst's raw() which creates a code block with optional syntax highlighting.
//
// Reference: typst-reference/crates/typst-library/src/text/raw.rs
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

// ----------------------------------------------------------------------------
// Helper functions
// ----------------------------------------------------------------------------

// parseNamedWeight converts a named weight string to its numeric value.
// Matches Typst's weight naming convention.
func parseNamedWeight(name string) int64 {
	switch name {
	case "thin":
		return 100
	case "extralight":
		return 200
	case "light":
		return 300
	case "regular", "normal":
		return 400
	case "medium":
		return 500
	case "semibold":
		return 600
	case "bold":
		return 700
	case "extrabold":
		return 800
	case "black":
		return 900
	default:
		return 400 // Default to normal
	}
}
