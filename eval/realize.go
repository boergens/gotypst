package eval

import (
	"regexp"
	"strings"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Show Rule Realization
// ----------------------------------------------------------------------------

// This file implements show rule application: matching elements against
// selectors (verdict determination), preparing transformations, and
// executing the transformation to produce new content.

// Verdict represents the result of matching a selector against content.
type Verdict int

const (
	// VerdictNone means the selector does not match.
	VerdictNone Verdict = iota
	// VerdictTransparent means pass through to inner content (for containers).
	VerdictTransparent
	// VerdictAccept means the selector matches and the recipe should apply.
	VerdictAccept
)

// RealizeContext holds state needed during content realization.
type RealizeContext struct {
	// VM is the evaluation context.
	VM *Vm
	// Recipes are the active show rule recipes, in reverse order of definition
	// (later recipes should be checked first for proper shadowing).
	Recipes []*Recipe
	// Depth tracks nesting depth for recursion limits.
	Depth int
	// MaxDepth is the maximum allowed nesting depth (default: 64).
	MaxDepth int
}

// NewRealizeContext creates a new realization context.
func NewRealizeContext(vm *Vm, recipes []*Recipe) *RealizeContext {
	return &RealizeContext{
		VM:       vm,
		Recipes:  recipes,
		Depth:    0,
		MaxDepth: 64,
	}
}

// ----------------------------------------------------------------------------
// Verdict Determination
// ----------------------------------------------------------------------------

// DetermineVerdict checks if a recipe's selector matches an element.
// Returns the verdict and, for Accept verdicts, the matched content.
func DetermineVerdict(recipe *Recipe, elem ContentElement, ctx *RealizeContext) (Verdict, Value) {
	if recipe.Selector == nil {
		// No selector means match everything ("show: ...")
		return VerdictAccept, elementToValue(elem)
	}
	return matchSelector(*recipe.Selector, elem, ctx)
}

// matchSelector checks if a selector matches an element.
func matchSelector(sel Selector, elem ContentElement, ctx *RealizeContext) (Verdict, Value) {
	switch s := sel.(type) {
	case ElemSelector:
		return matchElemSelector(s, elem, ctx)
	case LabelSelector:
		return matchLabelSelector(s, elem)
	case TextSelector:
		return matchTextSelector(s, elem)
	case FuncSelector:
		return matchFuncSelector(s, elem, ctx)
	case OrSelector:
		return matchOrSelector(s, elem, ctx)
	case AndSelector:
		return matchAndSelector(s, elem, ctx)
	case BeforeSelector:
		return matchBeforeSelector(s, elem, ctx)
	case AfterSelector:
		return matchAfterSelector(s, elem, ctx)
	default:
		return VerdictNone, nil
	}
}

// matchElemSelector checks if an element matches an element type selector.
func matchElemSelector(sel ElemSelector, elem ContentElement, ctx *RealizeContext) (Verdict, Value) {
	elemName := getElementName(elem)
	if elemName != sel.Element.Name {
		return VerdictNone, nil
	}

	elemVal := elementToValue(elem)

	// Check optional where clause
	if sel.Where != nil {
		matched, err := checkWhereClause(sel.Where, elemVal, ctx)
		if err != nil || !matched {
			return VerdictNone, nil
		}
	}

	return VerdictAccept, elemVal
}

// matchLabelSelector checks if an element has a matching label.
func matchLabelSelector(sel LabelSelector, elem ContentElement) (Verdict, Value) {
	label := getElementLabel(elem)
	if label == "" || label != sel.Label {
		return VerdictNone, nil
	}
	return VerdictAccept, elementToValue(elem)
}

// matchTextSelector checks if a text element matches a text pattern.
func matchTextSelector(sel TextSelector, elem ContentElement) (Verdict, Value) {
	textElem, ok := elem.(*TextElement)
	if !ok {
		return VerdictNone, nil
	}

	if sel.IsRegex {
		re, err := regexp.Compile(sel.Text)
		if err != nil {
			return VerdictNone, nil
		}
		if !re.MatchString(textElem.Text) {
			return VerdictNone, nil
		}
		// For regex, return the matched portion
		match := re.FindString(textElem.Text)
		return VerdictAccept, Str(match)
	}

	// Literal text match
	if !strings.Contains(textElem.Text, sel.Text) {
		return VerdictNone, nil
	}
	return VerdictAccept, Str(sel.Text)
}

// matchFuncSelector checks if an element matches a predicate function.
func matchFuncSelector(sel FuncSelector, elem ContentElement, ctx *RealizeContext) (Verdict, Value) {
	elemVal := elementToValue(elem)

	// Call the predicate function with the element
	args := NewArgs(syntax.Detached())
	args.Push(elemVal, syntax.Detached())

	result, err := CallFunc(ctx.VM, sel.Func, args)
	if err != nil {
		return VerdictNone, nil
	}

	// Check if result is truthy
	if b, ok := result.(BoolValue); ok && bool(b) {
		return VerdictAccept, elemVal
	}

	return VerdictNone, nil
}

// matchOrSelector checks if an element matches any of the selectors.
func matchOrSelector(sel OrSelector, elem ContentElement, ctx *RealizeContext) (Verdict, Value) {
	for _, sub := range sel.Selectors {
		verdict, val := matchSelector(sub, elem, ctx)
		if verdict == VerdictAccept {
			return VerdictAccept, val
		}
	}
	return VerdictNone, nil
}

// matchAndSelector checks if an element matches all selectors.
func matchAndSelector(sel AndSelector, elem ContentElement, ctx *RealizeContext) (Verdict, Value) {
	var lastVal Value
	for _, sub := range sel.Selectors {
		verdict, val := matchSelector(sub, elem, ctx)
		if verdict != VerdictAccept {
			return VerdictNone, nil
		}
		lastVal = val
	}
	return VerdictAccept, lastVal
}

// matchBeforeSelector is a placeholder for positional matching.
// Full implementation requires document-level context.
func matchBeforeSelector(sel BeforeSelector, elem ContentElement, ctx *RealizeContext) (Verdict, Value) {
	// For now, just match the base selector
	return matchSelector(sel.Selector, elem, ctx)
}

// matchAfterSelector is a placeholder for positional matching.
// Full implementation requires document-level context.
func matchAfterSelector(sel AfterSelector, elem ContentElement, ctx *RealizeContext) (Verdict, Value) {
	// For now, just match the base selector
	return matchSelector(sel.Selector, elem, ctx)
}

// checkWhereClause evaluates a where predicate for element filtering.
func checkWhereClause(f *Func, elemVal Value, ctx *RealizeContext) (bool, error) {
	args := NewArgs(syntax.Detached())
	args.Push(elemVal, syntax.Detached())

	result, err := CallFunc(ctx.VM, f, args)
	if err != nil {
		return false, err
	}

	if b, ok := result.(BoolValue); ok {
		return bool(b), nil
	}
	return false, nil
}

// ----------------------------------------------------------------------------
// Helper Functions
// ----------------------------------------------------------------------------

// getElementName returns the Typst name of a content element.
func getElementName(elem ContentElement) string {
	switch elem.(type) {
	case *TextElement:
		return "text"
	case *LinebreakElement:
		return "linebreak"
	case *ParbreakElement:
		return "parbreak"
	case *ParagraphElement:
		return "par"
	case *StrongElement:
		return "strong"
	case *EmphElement:
		return "emph"
	case *RawElement:
		return "raw"
	case *LinkElement:
		return "link"
	case *RefElement:
		return "ref"
	case *HeadingElement:
		return "heading"
	case *ListItemElement:
		return "list.item"
	case *EnumItemElement:
		return "enum.item"
	case *TermItemElement:
		return "terms.item"
	case *ListElement:
		return "list"
	case *EnumElement:
		return "enum"
	case *TermsElement:
		return "terms"
	case *SmartQuoteElement:
		return "smartquote"
	case *EquationElement:
		return "math.equation"
	case *MathFracElement:
		return "math.frac"
	case *MathRootElement:
		return "math.root"
	case *MathAttachElement:
		return "math.attach"
	case *MathDelimitedElement:
		return "math.lr"
	case *MathAlignElement:
		return "math.align"
	case *MathSymbolElement:
		return "math.symbol"
	default:
		return ""
	}
}

// getElementLabel returns the label of an element if it has one.
func getElementLabel(elem ContentElement) string {
	// Check for label field on elements that support it
	switch e := elem.(type) {
	case *RefElement:
		return e.Target
	}
	return ""
}

// elementToValue wraps a ContentElement in a Value for use in transformations.
func elementToValue(elem ContentElement) Value {
	return ContentValue{Content: Content{Elements: []ContentElement{elem}}}
}

// ----------------------------------------------------------------------------
// Transformation Preparation
// ----------------------------------------------------------------------------

// PreparedTransform holds the prepared state for applying a transformation.
type PreparedTransform struct {
	// Recipe is the source recipe.
	Recipe *Recipe
	// MatchedValue is the value that matched the selector.
	MatchedValue Value
	// Transform is the transformation to apply.
	Transform Transformation
}

// PrepareTransform prepares a transformation for execution.
// This sets up any necessary context like closure environments.
func PrepareTransform(recipe *Recipe, matchedVal Value) *PreparedTransform {
	return &PreparedTransform{
		Recipe:       recipe,
		MatchedValue: matchedVal,
		Transform:    recipe.Transform,
	}
}

// ----------------------------------------------------------------------------
// Transformation Execution
// ----------------------------------------------------------------------------

// ApplyTransform executes a prepared transformation to produce new content.
func ApplyTransform(prep *PreparedTransform, ctx *RealizeContext) (Content, error) {
	switch t := prep.Transform.(type) {
	case StyleTransformation:
		return applyStyleTransform(t, prep.MatchedValue, ctx)
	case FuncTransformation:
		return applyFuncTransform(t, prep.MatchedValue, ctx)
	case ContentTransformation:
		return applyContentTransform(t, prep.MatchedValue, ctx)
	case NoneTransformation:
		return applyNoneTransform()
	default:
		// Unknown transformation type - return original content
		if cv, ok := prep.MatchedValue.(ContentValue); ok {
			return cv.Content, nil
		}
		return Content{}, nil
	}
}

// applyStyleTransform applies style rules to content.
func applyStyleTransform(t StyleTransformation, matchedVal Value, ctx *RealizeContext) (Content, error) {
	// Style transformations wrap the content with styles
	// The styles will be applied during layout
	cv, ok := matchedVal.(ContentValue)
	if !ok {
		return Content{}, nil
	}

	// Create a styled content element
	styled := &StyledElement{
		Content: cv.Content,
		Styles:  t.Styles,
	}

	return Content{Elements: []ContentElement{styled}}, nil
}

// applyFuncTransform calls a function to transform content.
func applyFuncTransform(t FuncTransformation, matchedVal Value, ctx *RealizeContext) (Content, error) {
	// Check recursion depth
	if ctx.Depth >= ctx.MaxDepth {
		return Content{}, &RecursionLimitError{
			Message: "maximum show rule recursion depth exceeded",
			Depth:   ctx.Depth,
		}
	}

	// Call the transformation function with the matched content
	args := NewArgs(syntax.Detached())
	args.Push(matchedVal, syntax.Detached())

	result, err := CallFunc(ctx.VM, t.Func, args)
	if err != nil {
		return Content{}, err
	}

	// Convert result to content
	switch v := result.(type) {
	case ContentValue:
		return v.Content, nil
	case StrValue:
		return Content{Elements: []ContentElement{&TextElement{Text: string(v)}}}, nil
	case NoneValue:
		return Content{}, nil
	default:
		// Try to display other values
		return v.Display(), nil
	}
}

// applyContentTransform replaces with literal content.
func applyContentTransform(t ContentTransformation, matchedVal Value, ctx *RealizeContext) (Content, error) {
	return t.Content, nil
}

// applyNoneTransform removes the content (returns empty).
func applyNoneTransform() (Content, error) {
	return Content{}, nil
}

// ----------------------------------------------------------------------------
// Content Realization
// ----------------------------------------------------------------------------

// RealizeContent applies all active show rules to content, producing
// transformed content ready for layout.
func RealizeContent(content Content, ctx *RealizeContext) (Content, error) {
	var result []ContentElement

	for _, elem := range content.Elements {
		realized, err := realizeElement(elem, ctx)
		if err != nil {
			return Content{}, err
		}
		result = append(result, realized.Elements...)
	}

	return Content{Elements: result}, nil
}

// realizeElement applies show rules to a single element.
func realizeElement(elem ContentElement, ctx *RealizeContext) (Content, error) {
	// Check each recipe in reverse order (most recent first)
	for i := len(ctx.Recipes) - 1; i >= 0; i-- {
		recipe := ctx.Recipes[i]

		verdict, matchedVal := DetermineVerdict(recipe, elem, ctx)

		switch verdict {
		case VerdictAccept:
			// Recipe matches - prepare and apply transformation
			prep := PrepareTransform(recipe, matchedVal)

			// Create child context with incremented depth
			childCtx := &RealizeContext{
				VM:       ctx.VM,
				Recipes:  ctx.Recipes[:i], // Exclude this recipe to prevent self-recursion
				Depth:    ctx.Depth + 1,
				MaxDepth: ctx.MaxDepth,
			}

			transformed, err := ApplyTransform(prep, childCtx)
			if err != nil {
				return Content{}, err
			}

			// Recursively realize the transformed content
			return RealizeContent(transformed, childCtx)

		case VerdictTransparent:
			// Pass through to children
			continue

		case VerdictNone:
			// No match, try next recipe
			continue
		}
	}

	// No recipe matched - realize children if this is a container element
	return realizeChildren(elem, ctx)
}

// realizeChildren recursively realizes child content of container elements.
func realizeChildren(elem ContentElement, ctx *RealizeContext) (Content, error) {
	switch e := elem.(type) {
	case *StrongElement:
		realized, err := RealizeContent(e.Content, ctx)
		if err != nil {
			return Content{}, err
		}
		return Content{Elements: []ContentElement{&StrongElement{Content: realized}}}, nil

	case *EmphElement:
		realized, err := RealizeContent(e.Content, ctx)
		if err != nil {
			return Content{}, err
		}
		return Content{Elements: []ContentElement{&EmphElement{Content: realized}}}, nil

	case *HeadingElement:
		realized, err := RealizeContent(e.Content, ctx)
		if err != nil {
			return Content{}, err
		}
		return Content{Elements: []ContentElement{&HeadingElement{
			Content: realized,
			Level:   e.Level,
		}}}, nil

	case *LinkElement:
		// LinkElement only has URL, no body to realize
		return Content{Elements: []ContentElement{elem}}, nil

	case *ParagraphElement:
		realized, err := RealizeContent(e.Body, ctx)
		if err != nil {
			return Content{}, err
		}
		return Content{Elements: []ContentElement{&ParagraphElement{
			Body:            realized,
			Leading:         e.Leading,
			Justify:         e.Justify,
			Linebreaks:      e.Linebreaks,
			FirstLineIndent: e.FirstLineIndent,
			HangingIndent:   e.HangingIndent,
		}}}, nil

	case *ListElement:
		var children []Content
		for _, child := range e.Children {
			realized, err := RealizeContent(child, ctx)
			if err != nil {
				return Content{}, err
			}
			children = append(children, realized)
		}
		return Content{Elements: []ContentElement{&ListElement{
			Children:   children,
			Marker:     e.Marker,
			Indent:     e.Indent,
			BodyIndent: e.BodyIndent,
			Spacing:    e.Spacing,
			Tight:      e.Tight,
		}}}, nil

	case *EnumElement:
		var children []Content
		for _, child := range e.Children {
			realized, err := RealizeContent(child, ctx)
			if err != nil {
				return Content{}, err
			}
			children = append(children, realized)
		}
		return Content{Elements: []ContentElement{&EnumElement{
			Children:   children,
			Numbering:  e.Numbering,
			Start:      e.Start,
			Full:       e.Full,
			Indent:     e.Indent,
			BodyIndent: e.BodyIndent,
			Spacing:    e.Spacing,
			Tight:      e.Tight,
		}}}, nil

	case *TermsElement:
		var children []Content
		for _, child := range e.Children {
			realized, err := RealizeContent(child, ctx)
			if err != nil {
				return Content{}, err
			}
			children = append(children, realized)
		}
		return Content{Elements: []ContentElement{&TermsElement{
			Children:      children,
			Separator:     e.Separator,
			Indent:        e.Indent,
			HangingIndent: e.HangingIndent,
			Spacing:       e.Spacing,
			Tight:         e.Tight,
		}}}, nil

	case *ListItemElement:
		realized, err := RealizeContent(e.Content, ctx)
		if err != nil {
			return Content{}, err
		}
		return Content{Elements: []ContentElement{&ListItemElement{Content: realized}}}, nil

	case *EnumItemElement:
		realized, err := RealizeContent(e.Content, ctx)
		if err != nil {
			return Content{}, err
		}
		return Content{Elements: []ContentElement{&EnumItemElement{
			Content: realized,
			Number:  e.Number,
		}}}, nil

	case *TermItemElement:
		termRealized, err := RealizeContent(e.Term, ctx)
		if err != nil {
			return Content{}, err
		}
		descRealized, err := RealizeContent(e.Description, ctx)
		if err != nil {
			return Content{}, err
		}
		return Content{Elements: []ContentElement{&TermItemElement{
			Term:        termRealized,
			Description: descRealized,
		}}}, nil

	case *EquationElement:
		realized, err := RealizeContent(e.Body, ctx)
		if err != nil {
			return Content{}, err
		}
		return Content{Elements: []ContentElement{&EquationElement{
			Body:  realized,
			Block: e.Block,
		}}}, nil

	default:
		// Leaf element - return as-is
		return Content{Elements: []ContentElement{elem}}, nil
	}
}

// ----------------------------------------------------------------------------
// StyledElement
// ----------------------------------------------------------------------------

// StyledElement wraps content with applied styles from show rules.
type StyledElement struct {
	Content Content
	Styles  *Styles
}

func (*StyledElement) isContentElement() {}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// RecursionLimitError is returned when show rule recursion exceeds the limit.
type RecursionLimitError struct {
	Message string
	Depth   int
}

func (e *RecursionLimitError) Error() string {
	return e.Message
}
