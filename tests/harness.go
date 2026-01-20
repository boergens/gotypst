// Package tests provides a test harness for running Typst fixture tests.
package tests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/syntax"
)

// TestCase represents a single test case extracted from a fixture file.
type TestCase struct {
	Name       string            // Test name from delimiter
	Attrs      []string          // Attributes (paged, html, etc.)
	Code       string            // Test code
	Errors     []ExpectedError   // Expected error annotations
	Hints      []ExpectedHint    // Expected hint annotations
	SourceFile string            // Path to source fixture file
	LineNumber int               // Line number where test starts
}

// ExpectedError represents an expected error annotation.
type ExpectedError struct {
	Line    int    // Line number (1-indexed from test start)
	ColFrom int    // Start column
	ColTo   int    // End column (0 if not specified)
	Message string // Expected error message
}

// ExpectedHint represents an expected hint annotation.
type ExpectedHint struct {
	Line    int    // Line number
	ColFrom int    // Start column
	ColTo   int    // End column
	Message string // Expected hint message
}

// TestResult represents the result of running a single test.
type TestResult struct {
	Test    *TestCase
	Passed  bool
	Tokens  []*TokenInfo // Tokens produced by lexer
	Errors  []string     // Actual errors encountered
	Details string       // Detailed failure information
}

// TokenInfo holds information about a lexed token.
type TokenInfo struct {
	Kind syntax.SyntaxKind
	Text string
	Span Span
}

// Span represents a position range in source code.
type Span struct {
	Start int
	End   int
}

// TestRunner runs fixture tests and collects results.
type TestRunner struct {
	fixturesDir string
	results     []*TestResult
	verbose     bool
}

// NewTestRunner creates a new test runner for the given fixtures directory.
func NewTestRunner(fixturesDir string) *TestRunner {
	return &TestRunner{
		fixturesDir: fixturesDir,
		results:     make([]*TestResult, 0),
		verbose:     false,
	}
}

// SetVerbose enables verbose output.
func (r *TestRunner) SetVerbose(v bool) {
	r.verbose = v
}

// LoadFixtures loads all fixture files from the fixtures directory.
func (r *TestRunner) LoadFixtures(categories ...string) ([]*TestCase, error) {
	var tests []*TestCase

	if len(categories) == 0 {
		// Load all categories
		entries, err := os.ReadDir(r.fixturesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read fixtures dir: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() {
				categories = append(categories, e.Name())
			}
		}
	}

	for _, cat := range categories {
		catDir := filepath.Join(r.fixturesDir, cat)
		entries, err := os.ReadDir(catDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to read category %s: %w", cat, err)
		}

		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".typ") {
				continue
			}

			path := filepath.Join(catDir, e.Name())
			cases, err := ParseFixtureFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", path, err)
			}
			tests = append(tests, cases...)
		}
	}

	return tests, nil
}

// delimiterPattern matches test case delimiters: --- name attr1 attr2 ---
var delimiterPattern = regexp.MustCompile(`^---\s+([a-zA-Z0-9_-]+(?:\s+[a-zA-Z0-9_-]+)*)\s*---\s*$`)

// errorPattern matches error annotations: // Error: col message or // Error: col-col message or // Error: col-line:col message
var errorPattern = regexp.MustCompile(`^//\s*Error:\s*(\d+)(?:-(\d+(?::\d+)?))?\s+(.+)$`)

// hintPattern matches hint annotations: // Hint: line-col message
var hintPattern = regexp.MustCompile(`^//\s*Hint:\s*(\d+)(?:-(\d+))?\s+(.+)$`)

// ParseFixtureFile parses a fixture file and extracts test cases.
func ParseFixtureFile(path string) ([]*TestCase, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tests []*TestCase
	var current *TestCase
	var codeLines []string
	lineNum := 0

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := delimiterPattern.FindStringSubmatch(line); matches != nil {
			// Save previous test case
			if current != nil {
				current.Code = strings.Join(codeLines, "\n")
				tests = append(tests, current)
			}

			// Start new test case
			parts := strings.Fields(matches[1])
			current = &TestCase{
				Name:       parts[0],
				Attrs:      parts[1:],
				SourceFile: path,
				LineNumber: lineNum,
			}
			codeLines = nil
			continue
		}

		if current == nil {
			// Content before first delimiter - skip or treat as preamble
			continue
		}

		// Check for error annotations
		if matches := errorPattern.FindStringSubmatch(line); matches != nil {
			lineNo := parseInt(matches[1])
			colTo := 0
			if matches[2] != "" {
				colTo = parseInt(matches[2])
			}
			current.Errors = append(current.Errors, ExpectedError{
				Line:    lineNo,
				ColFrom: lineNo, // First number is often col start
				ColTo:   colTo,
				Message: strings.TrimSpace(matches[3]),
			})
		}

		// Check for hint annotations
		if matches := hintPattern.FindStringSubmatch(line); matches != nil {
			lineNo := parseInt(matches[1])
			colTo := 0
			if matches[2] != "" {
				colTo = parseInt(matches[2])
			}
			current.Hints = append(current.Hints, ExpectedHint{
				Line:    lineNo,
				ColFrom: lineNo,
				ColTo:   colTo,
				Message: strings.TrimSpace(matches[3]),
			})
		}

		codeLines = append(codeLines, line)
	}

	// Save last test case
	if current != nil {
		current.Code = strings.Join(codeLines, "\n")
		tests = append(tests, current)
	}

	return tests, scanner.Err()
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// RunTest runs a single test case and returns the result.
func (r *TestRunner) RunTest(tc *TestCase) *TestResult {
	result := &TestResult{
		Test:   tc,
		Passed: true,
	}

	// Determine the lexer mode based on the test code
	mode := syntax.ModeMarkup
	if strings.HasPrefix(strings.TrimSpace(tc.Code), "#") {
		// If code starts with #, it's likely code mode
		mode = syntax.ModeMarkup // Still start in markup, # switches to code
	}

	// Lex the test code
	lexer := syntax.NewLexer(tc.Code, mode)
	var tokens []*TokenInfo

	for {
		kind, node := lexer.Next()
		if kind == syntax.End {
			break
		}

		span := Span{
			Start: lexer.Cursor() - node.Len(),
			End:   lexer.Cursor(),
		}

		tokens = append(tokens, &TokenInfo{
			Kind: kind,
			Text: node.Text(),
			Span: span,
		})

		if kind == syntax.Error && getFirstError(node) != nil {
			result.Errors = append(result.Errors, getFirstError(node).Message)
		}
	}

	result.Tokens = tokens

	// If no lexer errors, try to parse and evaluate the code
	if len(result.Errors) == 0 {
		evalErrors := r.evaluateCode(tc.Code)
		result.Errors = append(result.Errors, evalErrors...)
	}

	// Validate expected errors
	if len(tc.Errors) > 0 {
		// Check if we got the expected errors
		if len(result.Errors) < len(tc.Errors) {
			result.Passed = false
			result.Details = fmt.Sprintf("expected %d errors, got %d",
				len(tc.Errors), len(result.Errors))
		} else {
			// Check error messages match
			for i, expected := range tc.Errors {
				if i >= len(result.Errors) {
					break
				}
				if !strings.Contains(result.Errors[i], expected.Message) {
					result.Passed = false
					result.Details = fmt.Sprintf("error mismatch: expected %q, got %q",
						expected.Message, result.Errors[i])
					break
				}
			}
		}
	} else {
		// No expected errors - any error is a failure
		if len(result.Errors) > 0 {
			result.Passed = false
			result.Details = fmt.Sprintf("unexpected errors: %v", result.Errors)
		}
	}

	return result
}

// evaluateCode parses and evaluates the given Typst code, returning any errors.
func (r *TestRunner) evaluateCode(code string) []string {
	var errors []string

	// Parse the code
	ast := syntax.Parse(code)
	if ast == nil {
		return errors
	}

	// Check for parse errors
	for _, err := range ast.Errors() {
		errors = append(errors, err.Message)
	}
	if len(errors) > 0 {
		return errors
	}

	// Create a minimal world for evaluation
	world := &testWorld{}
	engine := eval.NewEngine(world)
	scopes := eval.NewScopes(world.Library())
	ctx := eval.NewContext()
	vm := eval.NewVm(engine, ctx, scopes, syntax.Detached())

	// Register the test function and other standard functions
	registerStdlib(vm)

	// Convert AST to expression and evaluate
	// The AST is a markup node at the top level
	markup := syntax.MarkupNodeFromNode(ast)
	if markup != nil {
		_, err := evalMarkupNode(vm, markup)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	return errors
}

// testWorld is a minimal World implementation for testing.
type testWorld struct{}

func (w *testWorld) Library() *eval.Scope {
	return eval.NewScope()
}

func (w *testWorld) MainFile() eval.FileID {
	return eval.FileID{Path: "test.typ"}
}

func (w *testWorld) Source(id eval.FileID) (*syntax.Source, error) {
	return nil, fmt.Errorf("source not available")
}

func (w *testWorld) File(id eval.FileID) ([]byte, error) {
	return nil, fmt.Errorf("file not available")
}

func (w *testWorld) Today(offset *int) eval.Date {
	return eval.Date{Year: 2026, Month: 1, Day: 19}
}

// registerStdlib registers standard library functions in the VM.
func registerStdlib(vm *eval.Vm) {
	// Register the test function
	vm.Define("test", eval.FuncValue{
		Func: &eval.Func{
			Name: strPtr("test"),
			Span: syntax.Detached(),
			Repr: eval.NativeFunc{
				Func: testFunc,
				Info: &eval.FuncInfo{Name: "test"},
			},
		},
	})

	// Register calc module with rem function
	calcModule := eval.NewDict()
	calcModule.Set("rem", eval.FuncValue{
		Func: &eval.Func{
			Name: strPtr("rem"),
			Span: syntax.Detached(),
			Repr: eval.NativeFunc{
				Func: calcRemFunc,
				Info: &eval.FuncInfo{Name: "rem"},
			},
		},
	})
	vm.Define("calc", calcModule)

	// Register int function
	vm.Define("int", eval.FuncValue{
		Func: &eval.Func{
			Name: strPtr("int"),
			Span: syntax.Detached(),
			Repr: eval.NativeFunc{
				Func: intFunc,
				Info: &eval.FuncInfo{Name: "int"},
			},
		},
	})

	// Register str function
	vm.Define("str", eval.FuncValue{
		Func: &eval.Func{
			Name: strPtr("str"),
			Span: syntax.Detached(),
			Repr: eval.NativeFunc{
				Func: strFunc,
				Info: &eval.FuncInfo{Name: "str"},
			},
		},
	})
}

func strPtr(s string) *string {
	return &s
}

// testFunc implements the test function that compares two values.
func testFunc(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
	expected, err := args.Expect("expected")
	if err != nil {
		return nil, err
	}
	actual, err := args.Expect("actual")
	if err != nil {
		return nil, err
	}
	if err := args.Finish(); err != nil {
		return nil, err
	}

	if !eval.Equal(expected.V, actual.V) {
		return nil, fmt.Errorf("assertion failed: expected %v, got %v", expected.V, actual.V)
	}
	return eval.None, nil
}

// calcRemFunc implements calc.rem for modulo operations.
func calcRemFunc(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
	dividend, err := args.Expect("dividend")
	if err != nil {
		return nil, err
	}
	divisor, err := args.Expect("divisor")
	if err != nil {
		return nil, err
	}
	if err := args.Finish(); err != nil {
		return nil, err
	}

	a, ok := eval.AsInt(dividend.V)
	if !ok {
		return nil, fmt.Errorf("expected integer for dividend")
	}
	b, ok := eval.AsInt(divisor.V)
	if !ok {
		return nil, fmt.Errorf("expected integer for divisor")
	}
	if b == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return eval.Int(a % b), nil
}

// intFunc implements the int constructor.
func intFunc(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
	value, err := args.Expect("value")
	if err != nil {
		return nil, err
	}

	// Check for base parameter
	baseArg := args.Find("base")
	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch v := value.V.(type) {
	case eval.BoolValue:
		if bool(v) {
			return eval.Int(1), nil
		}
		return eval.Int(0), nil
	case eval.IntValue:
		if baseArg != nil {
			return nil, fmt.Errorf("base is only supported for strings")
		}
		return v, nil
	case eval.FloatValue:
		if baseArg != nil {
			return nil, fmt.Errorf("base is only supported for strings")
		}
		return eval.Int(int64(v)), nil
	case eval.StrValue:
		s := string(v)
		if s == "" {
			return nil, fmt.Errorf("string must not be empty")
		}

		base := 10
		if baseArg != nil {
			b, ok := eval.AsInt(baseArg.V)
			if !ok {
				return nil, fmt.Errorf("base must be an integer")
			}
			if b < 2 || b > 36 {
				return nil, fmt.Errorf("base must be between 2 and 36")
			}
			base = int(b)
		}

		// Handle minus sign variants
		s = strings.ReplaceAll(s, "\u2212", "-")
		s = strings.TrimPrefix(s, "+")

		// Parse the integer using strconv for overflow detection
		result, parseErr := strconv.ParseInt(s, base, 64)
		if parseErr != nil {
			// Check for overflow/underflow
			if numErr, ok := parseErr.(*strconv.NumError); ok {
				if numErr.Err == strconv.ErrRange {
					if strings.HasPrefix(s, "-") {
						return nil, fmt.Errorf("integer value is too small")
					}
					return nil, fmt.Errorf("integer value is too large")
				}
			}
			// Invalid digits
			if base == 10 {
				return nil, fmt.Errorf("string contains invalid digits")
			}
			return nil, fmt.Errorf("string contains invalid digits for a base %d integer", base)
		}

		return eval.Int(result), nil
	default:
		return nil, fmt.Errorf("expected integer, boolean, float, decimal, or string, found %s", value.V.Type())
	}
}

// strFunc implements the str constructor.
func strFunc(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
	value, err := args.Expect("value")
	if err != nil {
		return nil, err
	}

	// Check for base parameter
	baseArg := args.Find("base")
	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch v := value.V.(type) {
	case eval.IntValue:
		base := 10
		if baseArg != nil {
			b, ok := eval.AsInt(baseArg.V)
			if !ok {
				return nil, fmt.Errorf("base must be an integer")
			}
			if b < 2 || b > 36 {
				return nil, fmt.Errorf("base must be between 2 and 36")
			}
			base = int(b)
		}
		if base == 10 {
			return eval.Str(fmt.Sprintf("%d", int64(v))), nil
		}
		// Convert to base
		n := int64(v)
		negative := n < 0
		if negative {
			n = -n
		}
		if n == 0 {
			return eval.Str("0"), nil
		}
		digits := ""
		for n > 0 {
			d := n % int64(base)
			if d < 10 {
				digits = string('0'+byte(d)) + digits
			} else {
				digits = string('a'+byte(d-10)) + digits
			}
			n /= int64(base)
		}
		if negative {
			digits = "-" + digits
		}
		return eval.Str(digits), nil
	case eval.FloatValue:
		if baseArg != nil {
			return nil, fmt.Errorf("base is only supported for integers")
		}
		return eval.Str(fmt.Sprintf("%g", float64(v))), nil
	case eval.StrValue:
		return v, nil
	default:
		return nil, fmt.Errorf("expected integer, float, decimal, version, bytes, label, type, or string, found %s", value.V.Type())
	}
}

// evalMarkupNode evaluates a markup node.
func evalMarkupNode(vm *eval.Vm, markup *syntax.MarkupNode) (eval.Value, error) {
	var result eval.Value = eval.None

	for _, expr := range markup.Exprs() {
		val, err := eval.EvalExpr(vm, expr)
		if err != nil {
			return nil, err
		}
		result = val
	}

	return result, nil
}

// RunAll runs all loaded test cases.
func (r *TestRunner) RunAll(tests []*TestCase) []*TestResult {
	r.results = make([]*TestResult, 0, len(tests))
	for _, tc := range tests {
		result := r.RunTest(tc)
		r.results = append(r.results, result)
	}
	return r.results
}

// Summary returns a summary of test results.
func (r *TestRunner) Summary() string {
	passed := 0
	failed := 0
	for _, res := range r.results {
		if res.Passed {
			passed++
		} else {
			failed++
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\nTest Results: %d passed, %d failed, %d total\n",
		passed, failed, passed+failed))

	if failed > 0 {
		sb.WriteString("\nFailed tests:\n")
		for _, res := range r.results {
			if !res.Passed {
				sb.WriteString(fmt.Sprintf("  - %s (%s:%d)\n",
					res.Test.Name, res.Test.SourceFile, res.Test.LineNumber))
				if res.Details != "" {
					sb.WriteString(fmt.Sprintf("    %s\n", res.Details))
				}
			}
		}
	}

	return sb.String()
}

// FilterByCategory returns tests matching the given categories.
func FilterByCategory(tests []*TestCase, categories ...string) []*TestCase {
	if len(categories) == 0 {
		return tests
	}

	catSet := make(map[string]bool)
	for _, c := range categories {
		catSet[c] = true
	}

	var filtered []*TestCase
	for _, tc := range tests {
		// Extract category from source file path
		dir := filepath.Dir(tc.SourceFile)
		cat := filepath.Base(dir)
		if catSet[cat] {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// FilterByName returns tests matching the given name pattern.
func FilterByName(tests []*TestCase, pattern string) []*TestCase {
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fall back to substring match
		var filtered []*TestCase
		for _, tc := range tests {
			if strings.Contains(tc.Name, pattern) {
				filtered = append(filtered, tc)
			}
		}
		return filtered
	}

	var filtered []*TestCase
	for _, tc := range tests {
		if re.MatchString(tc.Name) {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// PrintTokens prints the tokens from a test result in a readable format.
func PrintTokens(tokens []*TokenInfo) string {
	var sb strings.Builder
	for _, t := range tokens {
		text := t.Text
		if len(text) > 40 {
			text = text[:37] + "..."
		}
		text = strings.ReplaceAll(text, "\n", "\\n")
		text = strings.ReplaceAll(text, "\t", "\\t")
		sb.WriteString(fmt.Sprintf("  %s %q (%d-%d)\n",
			t.Kind, text, t.Span.Start, t.Span.End))
	}
	return sb.String()
}

// getFirstError returns the first error from a node, or nil if none.
func getFirstError(node *syntax.SyntaxNode) *syntax.SyntaxError {
	errors := node.Errors()
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}
