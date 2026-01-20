package eval

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Math Content Element Types
// ----------------------------------------------------------------------------

// EquationElement represents a mathematical equation.
type EquationElement struct {
	// Body is the equation content.
	Body Content
	// Block indicates if this is a block (display) equation.
	Block bool
}

func (*EquationElement) IsContentElement() {}

// MathFracElement represents a fraction.
type MathFracElement struct {
	// Num is the numerator content.
	Num Content
	// Denom is the denominator content.
	Denom Content
}

func (*MathFracElement) IsContentElement() {}

// MathRootElement represents a root (square root, nth root).
type MathRootElement struct {
	// Index is the optional root index (nil for square root).
	Index Content
	// Radicand is the content under the root sign.
	Radicand Content
}

func (*MathRootElement) IsContentElement() {}

// MathAttachElement represents subscripts and superscripts.
type MathAttachElement struct {
	// Base is the base expression.
	Base Content
	// Subscript is the subscript content (may be empty).
	Subscript Content
	// Superscript is the superscript content (may be empty).
	Superscript Content
	// Primes is the number of prime marks.
	Primes int
}

func (*MathAttachElement) IsContentElement() {}

// MathDelimitedElement represents delimited math content.
type MathDelimitedElement struct {
	// Open is the opening delimiter.
	Open string
	// Close is the closing delimiter.
	Close string
	// Body is the content between delimiters.
	Body Content
}

func (*MathDelimitedElement) IsContentElement() {}

// MathAlignElement represents an alignment point in equations.
type MathAlignElement struct{}

func (*MathAlignElement) IsContentElement() {}

// MathSymbolElement represents a math symbol.
type MathSymbolElement struct {
	// Symbol is the symbol text.
	Symbol string
}

func (*MathSymbolElement) IsContentElement() {}

// ----------------------------------------------------------------------------
// Math Expression Evaluators
// ----------------------------------------------------------------------------

// evalEquation evaluates a math equation expression.
func evalEquation(vm *Vm, e *syntax.EquationExpr) (Value, error) {
	body := e.Body()
	if body == nil {
		return ContentValue{Content: Content{
			Elements: []ContentElement{&EquationElement{Block: e.Block()}},
		}}, nil
	}

	// Evaluate the math body
	content, err := evalMath(vm, body)
	if err != nil {
		return nil, err
	}

	mathContent, ok := content.(ContentValue)
	if !ok {
		mathContent = ContentValue{Content: Content{}}
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&EquationElement{
			Body:  mathContent.Content,
			Block: e.Block(),
		}},
	}}, nil
}

// evalMath evaluates math content and returns Content.
func evalMath(vm *Vm, math *syntax.MathNode) (Value, error) {
	var content Content
	exprs := math.Exprs()

	for _, expr := range exprs {
		if vm.HasFlow() {
			break
		}

		value, err := EvalExpr(vm, expr)
		if err != nil {
			return nil, err
		}

		// Append to content
		content = appendToContent(content, value)
	}

	return ContentValue{Content: content}, nil
}

// evalMathText evaluates text within math mode.
func evalMathText(_ *Vm, e *syntax.MathTextExpr) (Value, error) {
	return ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: e.Get()}},
	}}, nil
}

// evalMathIdent evaluates an identifier in math mode.
// Math identifiers can refer to variables or become symbol text.
func evalMathIdent(vm *Vm, e *syntax.MathIdentExpr) (Value, error) {
	name := e.Get()

	// First, try to look up as a variable
	binding := vm.Get(name)
	if binding != nil {
		value, err := binding.ReadChecked(e.ToUntyped().Span())
		if err != nil {
			return nil, err
		}
		vm.Trace(value, e.ToUntyped().Span())
		return value, nil
	}

	// Unknown identifiers in math mode become symbol text
	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathSymbolElement{Symbol: name}},
	}}, nil
}

// evalMathShorthand evaluates a math shorthand (like -> for arrow).
func evalMathShorthand(_ *Vm, e *syntax.MathShorthandExpr) (Value, error) {
	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathSymbolElement{Symbol: e.Get()}},
	}}, nil
}

// evalMathAlignPoint evaluates an alignment point in equations.
func evalMathAlignPoint(_ *Vm, _ *syntax.MathAlignPointExpr) (Value, error) {
	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathAlignElement{}},
	}}, nil
}

// evalMathDelimited evaluates delimited math content like (a + b).
func evalMathDelimited(vm *Vm, e *syntax.MathDelimitedExpr) (Value, error) {
	body := e.Body()

	var bodyContent Content
	if body != nil {
		content, err := evalMath(vm, body)
		if err != nil {
			return nil, err
		}
		if c, ok := content.(ContentValue); ok {
			bodyContent = c.Content
		}
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathDelimitedElement{
			Open:  e.Open(),
			Close: e.Close(),
			Body:  bodyContent,
		}},
	}}, nil
}

// evalMathAttach evaluates subscripts and superscripts (x^2_i).
func evalMathAttach(vm *Vm, e *syntax.MathAttachExpr) (Value, error) {
	// Evaluate the base expression
	base := e.Base()
	var baseContent Content
	if base != nil {
		value, err := EvalExpr(vm, base)
		if err != nil {
			return nil, err
		}
		if c, ok := value.(ContentValue); ok {
			baseContent = c.Content
		} else {
			baseContent = Content{Elements: []ContentElement{&TextElement{Text: value.Display().String()}}}
		}
	}

	// Parse children to find subscript, superscript, and primes
	var subscript, superscript Content
	primes := 0

	children := e.ToUntyped().Children()
	foundBase := false

	for _, child := range children {
		kind := child.Kind()

		// Skip the base (first non-operator child)
		if !foundBase && kind != syntax.Hat && kind != syntax.Underscore && kind != syntax.MathPrimes {
			foundBase = true
			continue
		}

		switch kind {
		case syntax.Underscore:
			// Next child is subscript
			continue
		case syntax.Hat:
			// Next child is superscript
			continue
		case syntax.MathPrimes:
			primes = len(child.Text())
		default:
			// This is either subscript or superscript content
			expr := syntax.ExprFromNode(child)
			if expr != nil {
				value, err := EvalExpr(vm, expr)
				if err != nil {
					return nil, err
				}
				if c, ok := value.(ContentValue); ok {
					// Determine if this is subscript or superscript based on preceding operator
					// We need to track the previous token
					prevIndex := findChildIndex(children, child) - 1
					if prevIndex >= 0 {
						prevKind := children[prevIndex].Kind()
						if prevKind == syntax.Underscore {
							subscript = c.Content
						} else if prevKind == syntax.Hat {
							superscript = c.Content
						}
					}
				}
			}
		}
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathAttachElement{
			Base:        baseContent,
			Subscript:   subscript,
			Superscript: superscript,
			Primes:      primes,
		}},
	}}, nil
}

// findChildIndex finds the index of a child in a slice.
func findChildIndex(children []*syntax.SyntaxNode, target *syntax.SyntaxNode) int {
	for i, child := range children {
		if child == target {
			return i
		}
	}
	return -1
}

// evalMathPrimes evaluates prime marks (x', x'').
func evalMathPrimes(_ *Vm, e *syntax.MathPrimesExpr) (Value, error) {
	count := e.Count()
	primes := ""
	for i := 0; i < count; i++ {
		primes += "′" // Unicode prime character
	}
	return ContentValue{Content: Content{
		Elements: []ContentElement{&TextElement{Text: primes}},
	}}, nil
}

// evalMathFrac evaluates a fraction (a/b).
func evalMathFrac(vm *Vm, e *syntax.MathFracExpr) (Value, error) {
	// Evaluate numerator
	num := e.Num()
	var numContent Content
	if num != nil {
		value, err := EvalExpr(vm, num)
		if err != nil {
			return nil, err
		}
		if c, ok := value.(ContentValue); ok {
			numContent = c.Content
		} else {
			numContent = valueToContent(value)
		}
	}

	// Evaluate denominator
	denom := e.Denom()
	var denomContent Content
	if denom != nil {
		value, err := EvalExpr(vm, denom)
		if err != nil {
			return nil, err
		}
		if c, ok := value.(ContentValue); ok {
			denomContent = c.Content
		} else {
			denomContent = valueToContent(value)
		}
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathFracElement{
			Num:   numContent,
			Denom: denomContent,
		}},
	}}, nil
}

// evalMathRoot evaluates a root expression (sqrt(x) or root(n, x)).
func evalMathRoot(vm *Vm, e *syntax.MathRootExpr) (Value, error) {
	// Evaluate optional index
	index := e.Index()
	var indexContent Content
	if index != nil {
		value, err := EvalExpr(vm, index)
		if err != nil {
			return nil, err
		}
		if c, ok := value.(ContentValue); ok {
			indexContent = c.Content
		} else {
			indexContent = valueToContent(value)
		}
	}

	// Evaluate radicand
	radicand := e.Radicand()
	var radicandContent Content
	if radicand != nil {
		value, err := EvalExpr(vm, radicand)
		if err != nil {
			return nil, err
		}
		if c, ok := value.(ContentValue); ok {
			radicandContent = c.Content
		} else {
			radicandContent = valueToContent(value)
		}
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathRootElement{
			Index:    indexContent,
			Radicand: radicandContent,
		}},
	}}, nil
}

// valueToContent converts a Value to Content for display.
func valueToContent(v Value) Content {
	switch val := v.(type) {
	case ContentValue:
		return val.Content
	case StrValue:
		return Content{Elements: []ContentElement{&TextElement{Text: string(val)}}}
	case IntValue:
		return Content{Elements: []ContentElement{&TextElement{Text: intToString(int64(val))}}}
	case FloatValue:
		return Content{Elements: []ContentElement{&TextElement{Text: floatToString(float64(val))}}}
	case NoneValue:
		return Content{}
	default:
		return Content{Elements: []ContentElement{&TextElement{Text: v.Display().String()}}}
	}
}

// intToString converts an int64 to a string.
func intToString(i int64) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// floatToString converts a float64 to a string.
func floatToString(f float64) string {
	// Simple conversion - for more precision, use strconv.FormatFloat
	if f == float64(int64(f)) {
		return intToString(int64(f))
	}
	// For now, just use a simple format
	// In production, use strconv.FormatFloat
	return fmt.Sprintf("%g", f)
}

// String returns a string representation of the Content.
func (c Content) String() string {
	var result string
	for _, elem := range c.Elements {
		if text, ok := elem.(*TextElement); ok {
			result += text.Text
		} else if symbol, ok := elem.(*MathSymbolElement); ok {
			result += symbol.Symbol
		}
	}
	return result
}
