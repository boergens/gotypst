// Math expression evaluation for Typst.
// Translated from typst-eval/src/math.rs

package eval

import (
	"fmt"
	"unicode"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/library/math"
	"github.com/boergens/gotypst/syntax"
)

// evalEquation evaluates a math equation expression.
// Matches Rust: impl Eval for ast::Equation
func evalEquation(vm *Vm, e *syntax.EquationExpr) (foundations.Value, error) {
	body, err := evalMath(vm, e.Body())
	if err != nil {
		return nil, err
	}

	bodyContent, ok := body.(foundations.ContentValue)
	if !ok {
		bodyContent = foundations.ContentValue{Content: foundations.Content{}}
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{&math.EquationElem{
			Body:  bodyContent.Content,
			Block: e.Block(),
		}},
	}}, nil
}

// evalMath evaluates math content and returns Content.
// Matches Rust: impl Eval for ast::Math
func evalMath(vm *Vm, m *syntax.MathNode) (foundations.Value, error) {
	if m == nil {
		return foundations.ContentValue{Content: foundations.Content{}}, nil
	}

	var contents []foundations.Content
	for _, expr := range m.Exprs() {
		if vm.HasFlow() {
			break
		}

		content, err := evalDisplay(vm, expr)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return foundations.ContentValue{Content: sequenceContent(contents)}, nil
}

// evalMathText evaluates text within math mode.
// Matches Rust: impl Eval for ast::MathText
func evalMathText(_ *Vm, e *syntax.MathTextExpr) (foundations.Value, error) {
	text := e.Get()
	if len(text) == 0 {
		return foundations.ContentValue{Content: foundations.Content{}}, nil
	}

	// Check if it's a number (first rune is numeric)
	firstRune := []rune(text)[0]
	if unicode.IsDigit(firstRune) {
		// Numbers use TextElem
		return foundations.ContentValue{Content: foundations.Content{
			Elements: []foundations.ContentElement{textElemPacked(text)},
		}}, nil
	}

	// Graphemes use SymbolElem
	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{&foundations.SymbolElem{Text: text}},
	}}, nil
}

// evalMathIdent evaluates an identifier in math mode.
// Matches Rust: impl Eval for ast::MathIdent
func evalMathIdent(vm *Vm, e *syntax.MathIdentExpr) (foundations.Value, error) {
	name := e.Get()
	span := e.ToUntyped().Span()

	// Use GetInMath for math-specific lookup
	binding := vm.GetInMath(name)
	if binding == nil {
		return nil, &UnknownVariableError{Name: name, Span: span}
	}

	value := binding.ReadChecked(span)
	vm.Trace(value)
	return value, nil
}

// evalMathShorthand evaluates a math shorthand (like -> for arrow).
// Matches Rust: impl Eval for ast::MathShorthand
// Returns Value::Symbol, not Content.
func evalMathShorthand(_ *Vm, e *syntax.MathShorthandExpr) (foundations.Value, error) {
	// Get the shorthand character
	shorthand := e.Get()
	if len(shorthand) == 0 {
		return foundations.SymbolValue{}, nil
	}

	// Return as Symbol value (runtime_char in Rust)
	firstRune := []rune(shorthand)[0]
	return foundations.SymbolValue{Char: firstRune}, nil
}

// evalMathAlignPoint evaluates an alignment point in equations.
// Matches Rust: impl Eval for ast::MathAlignPoint
func evalMathAlignPoint(_ *Vm, _ *syntax.MathAlignPointExpr) (foundations.Value, error) {
	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{math.SharedAlignPoint()},
	}}, nil
}

// evalMathDelimited evaluates delimited math content like (a + b).
// Matches Rust: impl Eval for ast::MathDelimited
func evalMathDelimited(vm *Vm, e *syntax.MathDelimitedExpr) (foundations.Value, error) {
	// In Rust, open() and close() are Exprs that get eval_display'd.
	// In Go, they're strings. For now, convert strings to SymbolElem.
	openStr := e.Open()
	open := symbolElemContent(openStr)

	// Evaluate body
	body := e.Body()
	var bodyContent foundations.Content
	if body != nil {
		content, err := evalMath(vm, body)
		if err != nil {
			return nil, err
		}
		if c, ok := content.(foundations.ContentValue); ok {
			bodyContent = c.Content
		}
	}

	closeStr := e.Close()
	close := symbolElemContent(closeStr)

	// Combine: open + body + close
	combined := foundations.Content{}
	combined.Elements = append(combined.Elements, open.Elements...)
	combined.Elements = append(combined.Elements, bodyContent.Elements...)
	combined.Elements = append(combined.Elements, close.Elements...)

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{&math.LrElem{Body: combined}},
	}}, nil
}

// evalMathAttach evaluates subscripts and superscripts (x^2_i).
// Matches Rust: impl Eval for ast::MathAttach
// Order: base, top, primes, bottom (matching Rust)
func evalMathAttach(vm *Vm, e *syntax.MathAttachExpr) (foundations.Value, error) {
	// Evaluate the base expression
	base, err := evalDisplay(vm, e.Base())
	if err != nil {
		return nil, err
	}

	elem := &math.AttachElem{Base: base}

	// Evaluate top (superscript)
	if top := e.Top(); top != nil {
		content, err := evalDisplay(vm, top)
		if err != nil {
			return nil, err
		}
		elem.T = &content
	}

	// Handle primes (always top-right, scripts style)
	if primes := e.Primes(); primes != nil {
		primesContent, err := evalMathPrimes(vm, primes)
		if err != nil {
			return nil, err
		}
		if pc, ok := primesContent.(foundations.ContentValue); ok {
			elem.TR = &pc.Content
		}
	}

	// Evaluate bottom (subscript)
	if bottom := e.Bottom(); bottom != nil {
		content, err := evalDisplay(vm, bottom)
		if err != nil {
			return nil, err
		}
		elem.B = &content
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}

// evalMathPrimes evaluates prime marks (x', x'').
// Matches Rust: impl Eval for ast::MathPrimes
func evalMathPrimes(_ *Vm, e *syntax.MathPrimesExpr) (foundations.Value, error) {
	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{&math.PrimesElem{Count: e.Count()}},
	}}, nil
}

// evalMathFrac evaluates a fraction (a/b).
// Matches Rust: impl Eval for ast::MathFrac
func evalMathFrac(vm *Vm, e *syntax.MathFracExpr) (foundations.Value, error) {
	// Evaluate numerator
	numExpr := e.Num()
	num, err := evalDisplay(vm, numExpr)
	if err != nil {
		return nil, err
	}

	// Evaluate denominator
	denomExpr := e.Denom()
	denom, err := evalDisplay(vm, denomExpr)
	if err != nil {
		return nil, err
	}

	// Note: Rust tracks num_deparenthesized and denom_deparenthesized
	// via matches!(num_expr, ast::Expr::Math(math) if math.was_deparenthesized())
	// Go doesn't have was_deparenthesized() yet, so we skip that.

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{&math.FracElem{
			Num:   num,
			Denom: denom,
		}},
	}}, nil
}

// evalMathRoot evaluates a root expression (sqrt(x) or root(n, x)).
// Matches Rust: impl Eval for ast::MathRoot
// In Rust, index() returns Option<u8> based on the root symbol (√, ∛, ∜).
func evalMathRoot(vm *Vm, e *syntax.MathRootExpr) (foundations.Value, error) {
	// Get the index from the root symbol
	// In Rust: √ → None (square root), ∛ → Some(3), ∜ → Some(4)
	index := e.Index()
	var indexContent foundations.Content
	if index != nil {
		// The Go AST returns an expression, but in Rust this is actually
		// derived from the root symbol itself. For compatibility, we
		// evaluate the expression and convert to text.
		indexVal, err := EvalExpr(vm, index)
		if err != nil {
			return nil, err
		}
		// Convert to TextElem (matching Rust's TextElem::packed for numbers)
		indexContent = valueToTextContent(indexVal)
	}

	// Evaluate radicand
	radicand, err := evalDisplay(vm, e.Radicand())
	if err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{&math.RootElem{
			Index:    indexContent,
			Radicand: radicand,
		}},
	}}, nil
}

// evalDisplay evaluates an expression and converts the result to Content.
// Matches Rust: trait ExprExt { fn eval_display(&self, vm) -> Content }
// This is: self.eval(vm)?.display().spanned(self.span())
func evalDisplay(vm *Vm, expr syntax.Expr) (foundations.Content, error) {
	if expr == nil {
		return foundations.Content{}, nil
	}

	value, err := EvalExpr(vm, expr)
	if err != nil {
		return foundations.Content{}, err
	}

	return valueToContent(value), nil
}

// valueToContent converts a Value to Content for display.
// Matches Rust's Value::display() trait implementation.
func valueToContent(v foundations.Value) foundations.Content {
	switch val := v.(type) {
	case foundations.ContentValue:
		return val.Content
	case foundations.Str:
		return foundations.Content{Elements: []foundations.ContentElement{textElemPacked(string(val))}}
	case foundations.Int:
		return foundations.Content{Elements: []foundations.ContentElement{textElemPacked(fmt.Sprintf("%d", val))}}
	case foundations.Float:
		return foundations.Content{Elements: []foundations.ContentElement{textElemPacked(fmt.Sprintf("%g", val))}}
	case foundations.NoneValue:
		return foundations.Content{}
	case foundations.SymbolValue:
		// Symbols display as their character
		return foundations.Content{Elements: []foundations.ContentElement{&foundations.SymbolElem{Text: string(val.Char)}}}
	default:
		if v == nil {
			return foundations.Content{}
		}
		// Fallback: format as string
		return foundations.Content{Elements: []foundations.ContentElement{textElemPacked(fmt.Sprintf("%v", v))}}
	}
}

// valueToTextContent converts a value to TextElem content (for numbers).
func valueToTextContent(v foundations.Value) foundations.Content {
	switch val := v.(type) {
	case foundations.Int:
		return foundations.Content{Elements: []foundations.ContentElement{textElemPacked(fmt.Sprintf("%d", val))}}
	case foundations.Float:
		return foundations.Content{Elements: []foundations.ContentElement{textElemPacked(fmt.Sprintf("%g", val))}}
	case foundations.Str:
		return foundations.Content{Elements: []foundations.ContentElement{textElemPacked(string(val))}}
	default:
		if v == nil {
			return foundations.Content{}
		}
		return foundations.Content{Elements: []foundations.ContentElement{textElemPacked(fmt.Sprintf("%v", v))}}
	}
}

// textElemPacked creates a TextElem as a content element.
// Matches Rust: TextElem::packed(text)
// Note: Uses TextElement from code.go to avoid import cycle with library/text.
func textElemPacked(s string) *TextElement {
	return &TextElement{Text: s}
}

// symbolElemContent creates Content containing a SymbolElem.
func symbolElemContent(s string) foundations.Content {
	if s == "" {
		return foundations.Content{}
	}
	return foundations.Content{Elements: []foundations.ContentElement{&foundations.SymbolElem{Text: s}}}
}

// sequenceContent combines multiple contents into one.
func sequenceContent(contents []foundations.Content) foundations.Content {
	var elements []foundations.ContentElement
	for _, c := range contents {
		elements = append(elements, c.Elements...)
	}
	return foundations.Content{Elements: elements}
}
