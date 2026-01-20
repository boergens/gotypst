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

// MathMatrixElement represents a matrix or array in math mode.
type MathMatrixElement struct {
	// Rows contains the matrix rows, each row is a slice of cell contents.
	Rows [][]Content
	// Delim specifies the delimiter style: "(", "[", "{", "|", "||", or empty for none.
	Delim string
	// Augment specifies the column index for an augmented matrix line (optional).
	Augment *int
	// RowGap is the gap between rows (in points, optional).
	RowGap *float64
	// ColumnGap is the gap between columns (in points, optional).
	ColumnGap *float64
}

func (*MathMatrixElement) IsContentElement() {}

// MathVecElement represents a column vector in math mode.
// This is a convenience type that's equivalent to a single-column matrix.
type MathVecElement struct {
	// Elements contains the vector elements.
	Elements []Content
	// Delim specifies the delimiter style: "(", "[", "{", "|", "||", or empty for none.
	Delim string
	// Gap is the gap between elements (in points, optional).
	Gap *float64
}

func (*MathVecElement) IsContentElement() {}

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

// ----------------------------------------------------------------------------
// Math Element Functions
// ----------------------------------------------------------------------------

// MatFunc creates the mat (matrix) function for math mode.
func MatFunc() *Func {
	name := "mat"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: matNative,
			Info: &FuncInfo{
				Name: "mat",
				Params: []ParamInfo{
					{Name: "delim", Type: TypeStr, Default: Str("("), Named: true},
					{Name: "augment", Type: TypeInt, Default: None, Named: true},
					{Name: "row-gap", Type: TypeLength, Default: None, Named: true},
					{Name: "column-gap", Type: TypeLength, Default: None, Named: true},
					{Name: "rows", Type: TypeArray, Named: false, Variadic: true},
				},
			},
		},
	}
}

// matNative implements the mat() function.
// Creates a MathMatrixElement from rows of values.
//
// Arguments:
//   - delim (named, str, default: "("): The delimiter style
//   - augment (named, int, default: none): Column index for augmented matrix line
//   - row-gap (named, length, default: none): Gap between rows
//   - column-gap (named, length, default: none): Gap between columns
//   - rows (positional, variadic): The matrix rows (arrays separated by semicolons)
func matNative(vm *Vm, args *Args) (Value, error) {
	// Get optional delim argument (default: "(")
	delim := "("
	if delimArg := args.Find("delim"); delimArg != nil {
		if !IsNone(delimArg.V) {
			delimStr, ok := AsStr(delimArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      delimArg.V.Type().String(),
					Span:     delimArg.Span,
				}
			}
			// Validate delimiter
			switch delimStr {
			case "(", "[", "{", "|", "||", "":
				delim = delimStr
			default:
				return nil, &TypeMismatchError{
					Expected: "\"(\", \"[\", \"{\", \"|\", \"||\", or empty string",
					Got:      "\"" + delimStr + "\"",
					Span:     delimArg.Span,
				}
			}
		}
	}

	// Get optional augment argument
	var augment *int
	if augmentArg := args.Find("augment"); augmentArg != nil {
		if !IsNone(augmentArg.V) {
			augmentVal, ok := augmentArg.V.(IntValue)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      augmentArg.V.Type().String(),
					Span:     augmentArg.Span,
				}
			}
			augInt := int(augmentVal)
			augment = &augInt
		}
	}

	// Get optional row-gap argument
	var rowGap *float64
	if rowGapArg := args.Find("row-gap"); rowGapArg != nil {
		if !IsNone(rowGapArg.V) && !IsAuto(rowGapArg.V) {
			if lv, ok := rowGapArg.V.(LengthValue); ok {
				rg := lv.Length.Points
				rowGap = &rg
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      rowGapArg.V.Type().String(),
					Span:     rowGapArg.Span,
				}
			}
		}
	}

	// Get optional column-gap argument
	var columnGap *float64
	if columnGapArg := args.Find("column-gap"); columnGapArg != nil {
		if !IsNone(columnGapArg.V) && !IsAuto(columnGapArg.V) {
			if lv, ok := columnGapArg.V.(LengthValue); ok {
				cg := lv.Length.Points
				columnGap = &cg
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      columnGapArg.V.Type().String(),
					Span:     columnGapArg.Span,
				}
			}
		}
	}

	// Collect rows from remaining positional arguments
	// In math mode, semicolons create Array nodes which become ArrayValue
	var rows [][]Content
	for {
		rowArg := args.Eat()
		if rowArg == nil {
			break
		}

		// Each positional argument is either:
		// - An ArrayValue (row with multiple cells from semicolon separation)
		// - A ContentValue (single cell row)
		// - Other value types
		row := parseMatrixRow(rowArg.V)
		rows = append(rows, row)
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the MathMatrixElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathMatrixElement{
			Rows:      rows,
			Delim:     delim,
			Augment:   augment,
			RowGap:    rowGap,
			ColumnGap: columnGap,
		}},
	}}, nil
}

// parseMatrixRow parses a value into a row of content cells.
func parseMatrixRow(v Value) []Content {
	switch val := v.(type) {
	case ArrayValue:
		// Array value - each element is a cell
		row := make([]Content, len(val))
		for i, elem := range val {
			row[i] = valueToContent(elem)
		}
		return row
	case ContentValue:
		// Single content cell
		return []Content{val.Content}
	default:
		// Convert other values to content
		return []Content{valueToContent(v)}
	}
}

// VecFunc creates the vec (column vector) function for math mode.
func VecFunc() *Func {
	name := "vec"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: vecNative,
			Info: &FuncInfo{
				Name: "vec",
				Params: []ParamInfo{
					{Name: "delim", Type: TypeStr, Default: Str("("), Named: true},
					{Name: "gap", Type: TypeLength, Default: None, Named: true},
					{Name: "elements", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// vecNative implements the vec() function.
// Creates a MathVecElement (column vector) from elements.
//
// Arguments:
//   - delim (named, str, default: "("): The delimiter style
//   - gap (named, length, default: none): Gap between elements
//   - elements (positional, variadic): The vector elements
func vecNative(vm *Vm, args *Args) (Value, error) {
	// Get optional delim argument (default: "(")
	delim := "("
	if delimArg := args.Find("delim"); delimArg != nil {
		if !IsNone(delimArg.V) {
			delimStr, ok := AsStr(delimArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      delimArg.V.Type().String(),
					Span:     delimArg.Span,
				}
			}
			// Validate delimiter
			switch delimStr {
			case "(", "[", "{", "|", "||", "":
				delim = delimStr
			default:
				return nil, &TypeMismatchError{
					Expected: "\"(\", \"[\", \"{\", \"|\", \"||\", or empty string",
					Got:      "\"" + delimStr + "\"",
					Span:     delimArg.Span,
				}
			}
		}
	}

	// Get optional gap argument
	var gap *float64
	if gapArg := args.Find("gap"); gapArg != nil {
		if !IsNone(gapArg.V) && !IsAuto(gapArg.V) {
			if lv, ok := gapArg.V.(LengthValue); ok {
				g := lv.Length.Points
				gap = &g
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      gapArg.V.Type().String(),
					Span:     gapArg.Span,
				}
			}
		}
	}

	// Collect elements from remaining positional arguments
	var elements []Content
	for {
		elemArg := args.Eat()
		if elemArg == nil {
			break
		}
		elements = append(elements, valueToContent(elemArg.V))
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the MathVecElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathVecElement{
			Elements: elements,
			Delim:    delim,
			Gap:      gap,
		}},
	}}, nil
}

// CasesFunc creates the cases function for math mode (piecewise functions).
func CasesFunc() *Func {
	name := "cases"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: casesNative,
			Info: &FuncInfo{
				Name: "cases",
				Params: []ParamInfo{
					{Name: "delim", Type: TypeStr, Default: Str("{"), Named: true},
					{Name: "reverse", Type: TypeBool, Default: False, Named: true},
					{Name: "gap", Type: TypeLength, Default: None, Named: true},
					{Name: "branches", Type: TypeArray, Named: false, Variadic: true},
				},
			},
		},
	}
}

// casesNative implements the cases() function.
// Creates a piecewise function definition using matrix layout.
func casesNative(vm *Vm, args *Args) (Value, error) {
	// Get optional delim argument (default: "{")
	delim := "{"
	if delimArg := args.Find("delim"); delimArg != nil {
		if !IsNone(delimArg.V) {
			delimStr, ok := AsStr(delimArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      delimArg.V.Type().String(),
					Span:     delimArg.Span,
				}
			}
			delim = delimStr
		}
	}

	// Get optional reverse argument (default: false)
	reverse := false
	if reverseArg := args.Find("reverse"); reverseArg != nil {
		if reverseVal, ok := AsBool(reverseArg.V); ok {
			reverse = reverseVal
		}
	}

	// Get optional gap argument
	var gap *float64
	if gapArg := args.Find("gap"); gapArg != nil {
		if !IsNone(gapArg.V) && !IsAuto(gapArg.V) {
			if lv, ok := gapArg.V.(LengthValue); ok {
				g := lv.Length.Points
				gap = &g
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      gapArg.V.Type().String(),
					Span:     gapArg.Span,
				}
			}
		}
	}

	// Collect branches from remaining positional arguments
	var rows [][]Content
	for {
		branchArg := args.Eat()
		if branchArg == nil {
			break
		}
		row := parseMatrixRow(branchArg.V)
		rows = append(rows, row)
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// For cases, we use a matrix with special delimiter handling
	// If reverse is true, the delimiter goes on the right side
	// We represent this as a matrix element for now
	_ = reverse // TODO: Handle reverse in layout

	return ContentValue{Content: Content{
		Elements: []ContentElement{&MathMatrixElement{
			Rows:   rows,
			Delim:  delim,
			RowGap: gap,
		}},
	}}, nil
}
