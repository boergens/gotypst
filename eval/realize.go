package eval

import (
	"regexp"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Realization System
// ----------------------------------------------------------------------------
//
// The realization system transforms content by applying show rules. It sits
// between evaluation and layout in the pipeline:
//
//   Source → Parse → Evaluate → REALIZE → Layout → Render
//
// The process involves:
// 1. Verdict determination - matching elements against selectors
// 2. Preparation - setting up transformations
// 3. Rule execution - applying transformations to matched elements

// ----------------------------------------------------------------------------
// Element Name Resolution
// ----------------------------------------------------------------------------

// ElementName returns the canonical name of a content element for selector matching.
// This maps each ContentElement type to its Typst element name.
func ElementName(elem ContentElement) string {
	switch elem.(type) {
	// Markup elements
	case *TextElement:
		return "text"
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
	case *LinebreakElement:
		return "linebreak"
	case *ParbreakElement:
		return "parbreak"
	case *SmartQuoteElement:
		return "smartquote"

	// Layout elements
	case *StackElement:
		return "stack"
	case *AlignElement:
		return "align"
	case *BoxElement:
		return "box"
	case *BlockElement:
		return "block"
	case *PadElement:
		return "pad"

	// Math elements
	case *EquationElement:
		return "equation"
	case *MathFracElement:
		return "math.frac"
	case *MathRootElement:
		return "math.root"
	case *MathAttachElement:
		return "math.attach"
	case *MathDelimitedElement:
		return "math.lr"
	case *MathAlignElement:
		return "math.align-point"
	case *MathSymbolElement:
		return "math.symbol"
	case *MathLimitsElement:
		return "math.limits"

	default:
		return ""
	}
}

// ----------------------------------------------------------------------------
// Selector Matching (Verdict Determination)
// ----------------------------------------------------------------------------

// MatchResult holds the result of selector matching.
type MatchResult struct {
	Matched bool
	// MatchedText is set when a TextSelector matches, containing the matched portion.
	MatchedText string
	// MatchStart and MatchEnd are indices for text matches.
	MatchStart int
	MatchEnd   int
}

// MatchSelector determines if a content element matches a selector.
// This is the "verdict determination" phase of show rule processing.
func MatchSelector(sel Selector, elem ContentElement, vm *Vm) (*MatchResult, error) {
	switch s := sel.(type) {
	case ElemSelector:
		return matchElemSelector(s, elem, vm)
	case LabelSelector:
		return matchLabelSelector(s, elem)
	case TextSelector:
		return matchTextSelector(s, elem)
	case FuncSelector:
		return matchFuncSelector(s, elem, vm)
	case OrSelector:
		return matchOrSelector(s, elem, vm)
	case AndSelector:
		return matchAndSelector(s, elem, vm)
	case BeforeSelector:
		// Before/After selectors require context about surrounding elements
		// which is handled at the Realize level, not here
		return &MatchResult{Matched: false}, nil
	case AfterSelector:
		return &MatchResult{Matched: false}, nil
	default:
		return &MatchResult{Matched: false}, nil
	}
}

// matchElemSelector matches elements by type.
func matchElemSelector(sel ElemSelector, elem ContentElement, vm *Vm) (*MatchResult, error) {
	elemName := ElementName(elem)
	if elemName == "" {
		return &MatchResult{Matched: false}, nil
	}

	if elemName != sel.Element.Name {
		return &MatchResult{Matched: false}, nil
	}

	// If there's a Where clause, evaluate it
	if sel.Where != nil {
		result, err := evalWhereClause(sel.Where, elem, vm)
		if err != nil {
			return nil, err
		}
		return &MatchResult{Matched: result}, nil
	}

	return &MatchResult{Matched: true}, nil
}

// matchLabelSelector matches elements with a specific label.
// In the current implementation, labels are attached via metadata which
// we don't track per-element. This is a simplified version.
func matchLabelSelector(_ LabelSelector, _ ContentElement) (*MatchResult, error) {
	// TODO: Implement label tracking on elements
	// For now, labels don't attach to elements in our model
	return &MatchResult{Matched: false}, nil
}

// matchTextSelector matches text content using a pattern.
func matchTextSelector(sel TextSelector, elem ContentElement) (*MatchResult, error) {
	textElem, ok := elem.(*TextElement)
	if !ok {
		return &MatchResult{Matched: false}, nil
	}

	if sel.IsRegex {
		re, err := regexp.Compile(sel.Text)
		if err != nil {
			return nil, &ShowRuleError{
				Message: "invalid regex in show rule: " + err.Error(),
			}
		}
		loc := re.FindStringIndex(textElem.Text)
		if loc == nil {
			return &MatchResult{Matched: false}, nil
		}
		return &MatchResult{
			Matched:     true,
			MatchedText: textElem.Text[loc[0]:loc[1]],
			MatchStart:  loc[0],
			MatchEnd:    loc[1],
		}, nil
	}

	// Plain string match - find substring
	idx := indexOf(textElem.Text, sel.Text)
	if idx < 0 {
		return &MatchResult{Matched: false}, nil
	}
	return &MatchResult{
		Matched:     true,
		MatchedText: sel.Text,
		MatchStart:  idx,
		MatchEnd:    idx + len(sel.Text),
	}, nil
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// matchFuncSelector matches using a predicate function.
func matchFuncSelector(sel FuncSelector, elem ContentElement, vm *Vm) (*MatchResult, error) {
	if sel.Func == nil {
		return &MatchResult{Matched: false}, nil
	}

	// Convert element to a value for the function call
	elemVal := elementToValue(elem)

	// Call the function with the element as argument
	args := &Args{
		Items: []Arg{{
			Value: syntax.Spanned[Value]{V: elemVal},
		}},
	}

	result, err := CallFunc(vm, sel.Func, args)
	if err != nil {
		return nil, err
	}

	// Result should be a boolean
	if b, ok := result.(BoolValue); ok {
		return &MatchResult{Matched: bool(b)}, nil
	}

	return &MatchResult{Matched: false}, nil
}

// matchOrSelector matches if any sub-selector matches.
func matchOrSelector(sel OrSelector, elem ContentElement, vm *Vm) (*MatchResult, error) {
	for _, sub := range sel.Selectors {
		result, err := MatchSelector(sub, elem, vm)
		if err != nil {
			return nil, err
		}
		if result.Matched {
			return result, nil
		}
	}
	return &MatchResult{Matched: false}, nil
}

// matchAndSelector matches only if all sub-selectors match.
func matchAndSelector(sel AndSelector, elem ContentElement, vm *Vm) (*MatchResult, error) {
	for _, sub := range sel.Selectors {
		result, err := MatchSelector(sub, elem, vm)
		if err != nil {
			return nil, err
		}
		if !result.Matched {
			return &MatchResult{Matched: false}, nil
		}
	}
	return &MatchResult{Matched: true}, nil
}

// evalWhereClause evaluates a where clause filter on an element.
func evalWhereClause(whereFunc *Func, elem ContentElement, vm *Vm) (bool, error) {
	// Convert element to a value
	elemVal := elementToValue(elem)

	// Call the where function with the element
	args := &Args{
		Items: []Arg{{
			Value: syntax.Spanned[Value]{V: elemVal},
		}},
	}

	result, err := CallFunc(vm, whereFunc, args)
	if err != nil {
		return false, err
	}

	if b, ok := result.(BoolValue); ok {
		return bool(b), nil
	}
	return false, nil
}

// elementToValue converts a ContentElement to a Value for function calls.
func elementToValue(elem ContentElement) Value {
	return ContentValue{Content: Content{Elements: []ContentElement{elem}}}
}

// ----------------------------------------------------------------------------
// Transformation Application (Rule Execution)
// ----------------------------------------------------------------------------

// ApplyTransformation applies a transformation to a content element.
// Returns the transformed content elements.
func ApplyTransformation(trans Transformation, elem ContentElement, vm *Vm) ([]ContentElement, error) {
	switch t := trans.(type) {
	case StyleTransformation:
		return applyStyleTransformation(t, elem)
	case FuncTransformation:
		return applyFuncTransformation(t, elem, vm)
	case ContentTransformation:
		return applyContentTransformation(t)
	case NoneTransformation:
		return applyNoneTransformation()
	default:
		// Unknown transformation, return element unchanged
		return []ContentElement{elem}, nil
	}
}

// applyStyleTransformation applies styles to an element.
// Styles don't change the element structure, they affect rendering.
func applyStyleTransformation(_ StyleTransformation, elem ContentElement) ([]ContentElement, error) {
	// Styles are accumulated and applied during layout, not here.
	// We return the element unchanged; the styles are tracked separately.
	return []ContentElement{elem}, nil
}

// applyFuncTransformation applies a function to transform content.
func applyFuncTransformation(trans FuncTransformation, elem ContentElement, vm *Vm) ([]ContentElement, error) {
	if trans.Func == nil {
		return []ContentElement{elem}, nil
	}

	// Convert element to content value for the function call
	elemVal := elementToValue(elem)

	// Call the transformation function with the element
	args := &Args{
		Items: []Arg{{
			Value: syntax.Spanned[Value]{V: elemVal},
		}},
	}

	result, err := CallFunc(vm, trans.Func, args)
	if err != nil {
		return nil, err
	}

	// Handle the result based on type
	switch v := result.(type) {
	case ContentValue:
		return v.Content.Elements, nil
	case NoneValue:
		// None means hide the content
		return nil, nil
	case StylesValue:
		// Styles transformation - return element unchanged
		return []ContentElement{elem}, nil
	default:
		// For other types, try to convert to content
		displayContent := v.Display()
		if len(displayContent.Elements) > 0 {
			return displayContent.Elements, nil
		}
		return []ContentElement{elem}, nil
	}
}

// applyContentTransformation replaces content directly.
func applyContentTransformation(trans ContentTransformation) ([]ContentElement, error) {
	return trans.Content.Elements, nil
}

// applyNoneTransformation hides the matched content.
func applyNoneTransformation() ([]ContentElement, error) {
	return nil, nil
}

// ----------------------------------------------------------------------------
// Content Realization
// ----------------------------------------------------------------------------

// RealizeOptions configures the realization process.
type RealizeOptions struct {
	// Kind specifies the realization context (document, fragment, math, etc.)
	Kind RealizationKind
}

// RealizationKind specifies the type of realization being performed.
type RealizationKind int

const (
	// RealizeDocument prepares content for full document layout.
	RealizeDocument RealizationKind = iota
	// RealizeFragment prepares content for fragment layout.
	RealizeFragment
	// RealizeMath prepares mathematical content.
	RealizeMath
)

// Realize transforms content by applying show rules.
// This is the main entry point for the realization system.
func Realize(content Content, styles *Styles, vm *Vm) (Content, error) {
	return RealizeWithOptions(content, styles, vm, RealizeOptions{Kind: RealizeDocument})
}

// RealizeWithOptions transforms content with specific options.
func RealizeWithOptions(content Content, styles *Styles, vm *Vm, _ RealizeOptions) (Content, error) {
	if styles == nil || len(styles.Recipes) == 0 {
		// No show rules to apply
		return content, nil
	}

	var result []ContentElement

	for _, elem := range content.Elements {
		transformed, err := realizeElement(elem, styles, vm)
		if err != nil {
			return Content{}, err
		}
		result = append(result, transformed...)
	}

	return Content{Elements: result}, nil
}

// realizeElement processes a single element through show rules.
func realizeElement(elem ContentElement, styles *Styles, vm *Vm) ([]ContentElement, error) {
	// First, recursively realize any child content
	elem, err := realizeChildren(elem, styles, vm)
	if err != nil {
		return nil, err
	}

	// Then, try to match and apply show rules
	for _, recipe := range styles.Recipes {
		matched, transformation, err := matchRecipe(recipe, elem, vm)
		if err != nil {
			return nil, err
		}
		if matched {
			// Apply the transformation
			transformed, err := ApplyTransformation(transformation, elem, vm)
			if err != nil {
				return nil, err
			}

			// Recursively realize the transformed content
			// (transformations may produce new content that needs processing)
			var result []ContentElement
			for _, te := range transformed {
				// Avoid infinite loops by not re-processing with the same recipe
				realized, err := realizeElementExcluding(te, styles, recipe, vm)
				if err != nil {
					return nil, err
				}
				result = append(result, realized...)
			}
			return result, nil
		}
	}

	// No rules matched, return element as-is
	return []ContentElement{elem}, nil
}

// realizeElementExcluding processes an element excluding a specific recipe to prevent loops.
func realizeElementExcluding(elem ContentElement, styles *Styles, exclude *Recipe, vm *Vm) ([]ContentElement, error) {
	// Create a filtered styles without the excluded recipe
	var filteredRecipes []*Recipe
	for _, r := range styles.Recipes {
		if r != exclude {
			filteredRecipes = append(filteredRecipes, r)
		}
	}

	if len(filteredRecipes) == 0 {
		return []ContentElement{elem}, nil
	}

	filteredStyles := &Styles{
		Rules:   styles.Rules,
		Recipes: filteredRecipes,
	}

	return realizeElement(elem, filteredStyles, vm)
}

// matchRecipe checks if a recipe matches an element and returns the transformation.
func matchRecipe(recipe *Recipe, elem ContentElement, vm *Vm) (bool, Transformation, error) {
	// If no selector, the rule matches everything
	if recipe.Selector == nil {
		return true, recipe.Transform, nil
	}

	result, err := MatchSelector(*recipe.Selector, elem, vm)
	if err != nil {
		return false, nil, err
	}

	return result.Matched, recipe.Transform, nil
}

// realizeChildren recursively realizes child content within an element.
func realizeChildren(elem ContentElement, styles *Styles, vm *Vm) (ContentElement, error) {
	switch e := elem.(type) {
	case *StrongElement:
		realized, err := RealizeWithOptions(e.Content, styles, vm, RealizeOptions{})
		if err != nil {
			return nil, err
		}
		return &StrongElement{Content: realized}, nil

	case *EmphElement:
		realized, err := RealizeWithOptions(e.Content, styles, vm, RealizeOptions{})
		if err != nil {
			return nil, err
		}
		return &EmphElement{Content: realized}, nil

	case *HeadingElement:
		realized, err := RealizeWithOptions(e.Content, styles, vm, RealizeOptions{})
		if err != nil {
			return nil, err
		}
		return &HeadingElement{Level: e.Level, Content: realized}, nil

	case *ListItemElement:
		realized, err := RealizeWithOptions(e.Content, styles, vm, RealizeOptions{})
		if err != nil {
			return nil, err
		}
		return &ListItemElement{Content: realized}, nil

	case *EnumItemElement:
		realized, err := RealizeWithOptions(e.Content, styles, vm, RealizeOptions{})
		if err != nil {
			return nil, err
		}
		return &EnumItemElement{Number: e.Number, Content: realized}, nil

	case *TermItemElement:
		termRealized, err := RealizeWithOptions(e.Term, styles, vm, RealizeOptions{})
		if err != nil {
			return nil, err
		}
		descRealized, err := RealizeWithOptions(e.Description, styles, vm, RealizeOptions{})
		if err != nil {
			return nil, err
		}
		return &TermItemElement{Term: termRealized, Description: descRealized}, nil

	case *ParagraphElement:
		realized, err := RealizeWithOptions(e.Body, styles, vm, RealizeOptions{})
		if err != nil {
			return nil, err
		}
		return &ParagraphElement{
			Body:            realized,
			Leading:         e.Leading,
			Justify:         e.Justify,
			Linebreaks:      e.Linebreaks,
			FirstLineIndent: e.FirstLineIndent,
			HangingIndent:   e.HangingIndent,
		}, nil

	case *EquationElement:
		realized, err := RealizeWithOptions(e.Body, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		return &EquationElement{Body: realized, Block: e.Block}, nil

	case *MathFracElement:
		numRealized, err := RealizeWithOptions(e.Num, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		denomRealized, err := RealizeWithOptions(e.Denom, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		return &MathFracElement{Num: numRealized, Denom: denomRealized}, nil

	case *MathRootElement:
		indexRealized, err := RealizeWithOptions(e.Index, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		radicandRealized, err := RealizeWithOptions(e.Radicand, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		return &MathRootElement{Index: indexRealized, Radicand: radicandRealized}, nil

	case *MathAttachElement:
		baseRealized, err := RealizeWithOptions(e.Base, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		subRealized, err := RealizeWithOptions(e.Subscript, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		superRealized, err := RealizeWithOptions(e.Superscript, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		return &MathAttachElement{
			Base:        baseRealized,
			Subscript:   subRealized,
			Superscript: superRealized,
			Primes:      e.Primes,
		}, nil

	case *MathDelimitedElement:
		realized, err := RealizeWithOptions(e.Body, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		return &MathDelimitedElement{Open: e.Open, Close: e.Close, Body: realized}, nil

	case *MathLimitsElement:
		nucleusRealized, err := RealizeWithOptions(e.Nucleus, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		upperRealized, err := RealizeWithOptions(e.Upper, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		lowerRealized, err := RealizeWithOptions(e.Lower, styles, vm, RealizeOptions{Kind: RealizeMath})
		if err != nil {
			return nil, err
		}
		return &MathLimitsElement{
			Nucleus: nucleusRealized,
			Upper:   upperRealized,
			Lower:   lowerRealized,
		}, nil

	default:
		// Element has no children to realize
		return elem, nil
	}
}

// ----------------------------------------------------------------------------
// Style Chain Operations
// ----------------------------------------------------------------------------

// MergeStyles merges two style collections, with s2 taking precedence.
func MergeStyles(s1, s2 *Styles) *Styles {
	if s1 == nil {
		return s2
	}
	if s2 == nil {
		return s1
	}

	return &Styles{
		Rules:   append(s1.Rules, s2.Rules...),
		Recipes: append(s1.Recipes, s2.Recipes...),
	}
}

// GetStyleProperty retrieves a property value from styles for a given element function.
func GetStyleProperty(styles *Styles, funcName string, propName string) Value {
	if styles == nil {
		return nil
	}

	// Search rules in reverse order (later rules take precedence)
	for i := len(styles.Rules) - 1; i >= 0; i-- {
		rule := styles.Rules[i]
		if rule.Func != nil && rule.Func.Name != nil && *rule.Func.Name == funcName {
			if rule.Args != nil {
				if val := rule.Args.GetNamed(propName); val != nil {
					return val.V
				}
			}
		}
	}

	return nil
}
