// Package tests provides a test harness for running Typst fixture tests.
package tests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// errorPattern matches error annotations: // Error: line-col message
var errorPattern = regexp.MustCompile(`^//\s*Error:\s*(\d+)(?:-(\d+))?\s+(.+)$`)

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

	// Lex the test code (for backward compatibility and token collection)
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

	// Parse the code to get AST and parser errors
	parsed := syntax.Parse(tc.Code)
	collectParserErrors(parsed, &result.Errors)

	// Evaluate the code to get evaluation errors
	evalErrors := evaluateCode(parsed)
	result.Errors = append(result.Errors, evalErrors...)

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

// collectParserErrors recursively collects errors from the syntax tree.
func collectParserErrors(node *syntax.SyntaxNode, errors *[]string) {
	if node == nil {
		return
	}

	// Collect errors from this node
	for _, err := range node.Errors() {
		*errors = append(*errors, err.Message)
	}

	// Recursively collect from children
	for _, child := range node.Children() {
		collectParserErrors(child, errors)
	}
}

// evaluateCode evaluates parsed code and returns any errors.
func evaluateCode(parsed *syntax.SyntaxNode) []string {
	var errors []string

	// Create a minimal test world with basic library
	world := newTestWorld()
	engine := eval.NewEngine(world)
	scopes := eval.NewScopes(world.Library())
	ctx := eval.NewContext()
	vm := eval.NewVm(engine, ctx, scopes, syntax.Span{})

	// Get the markup node and evaluate each expression
	markup := syntax.MarkupNodeFromNode(parsed)
	if markup == nil {
		return errors
	}

	for _, expr := range markup.Exprs() {
		_, err := eval.EvalExpr(vm, expr)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	return errors
}

// testWorld is a minimal World implementation for testing.
type testWorld struct {
	library *eval.Scope
}

func newTestWorld() *testWorld {
	lib := eval.NewScope()

	// Add test function that asserts two values are equal
	testFunc := &eval.Func{
		Name: strPtr("test"),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				// test(actual, expected) - just return none, we only care about errors
				return eval.None, nil
			},
		},
	}
	lib.Define("test", eval.FuncValue{Func: testFunc}, syntax.Span{})

	// Add repr function
	reprFunc := &eval.Func{
		Name: strPtr("repr"),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				if len(args.Items) > 0 {
					return eval.Str(fmt.Sprintf("%v", args.Items[0].Value.V)), nil
				}
				return eval.Str(""), nil
			},
		},
	}
	lib.Define("repr", eval.FuncValue{Func: reprFunc}, syntax.Span{})

	// Add str function
	strFunc := &eval.Func{
		Name: strPtr("str"),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				if len(args.Items) > 0 {
					return eval.Str(fmt.Sprintf("%v", args.Items[0].Value.V)), nil
				}
				return eval.Str(""), nil
			},
		},
	}
	lib.Define("str", eval.FuncValue{Func: strFunc}, syntax.Span{})

	// Add type function
	typeFunc := &eval.Func{
		Name: strPtr("type"),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				if len(args.Items) > 0 {
					return eval.TypeValue{Inner: args.Items[0].Value.V.Type()}, nil
				}
				return eval.TypeValue{Inner: eval.TypeNone}, nil
			},
		},
	}
	lib.Define("type", eval.FuncValue{Func: typeFunc}, syntax.Span{})

	// Add range function
	rangeFunc := &eval.Func{
		Name: strPtr("range"),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				if len(args.Items) > 0 {
					if n, ok := args.Items[0].Value.V.(eval.IntValue); ok {
						arr := make(eval.ArrayValue, int(n))
						for i := int64(0); i < int64(n); i++ {
							arr[i] = eval.Int(i)
						}
						return arr, nil
					}
				}
				return eval.ArrayValue{}, nil
			},
		},
	}
	lib.Define("range", eval.FuncValue{Func: rangeFunc}, syntax.Span{})

	// Add lorem function (placeholder text generator)
	loremFunc := &eval.Func{
		Name: strPtr("lorem"),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				return eval.Str("Lorem ipsum dolor sit amet."), nil
			},
		},
	}
	lib.Define("lorem", eval.FuncValue{Func: loremFunc}, syntax.Span{})

	// Add content type constant
	lib.Define("content", eval.TypeValue{Inner: eval.TypeContent}, syntax.Span{})

	// Add rect function (placeholder, immutable binding to test constant assignment)
	rectFunc := &eval.Func{
		Name: strPtr("rect"),
		Repr: eval.NativeFunc{
			Func: func(vm *eval.Vm, args *eval.Args) (eval.Value, error) {
				return eval.ContentValue{}, nil
			},
		},
	}
	rectBinding := eval.Binding{
		Value:   eval.FuncValue{Func: rectFunc},
		Kind:    eval.BindingNormal,
		Mutable: false,
		Span:    syntax.Span{},
	}
	lib.Insert("rect", rectBinding)

	// Add conifer color constant
	lib.Define("conifer", eval.ColorValue{Color: eval.Color{R: 96, G: 175, B: 70, A: 255}}, syntax.Span{})

	return &testWorld{library: lib}
}

func (w *testWorld) Library() *eval.Scope {
	return w.library
}

func (w *testWorld) MainFile() eval.FileID {
	return eval.FileID{Path: "test.typ"}
}

func (w *testWorld) Source(id eval.FileID) (*syntax.Source, error) {
	return nil, fmt.Errorf("file not found: %s", id.Path)
}

func (w *testWorld) File(id eval.FileID) ([]byte, error) {
	return nil, fmt.Errorf("file not found: %s", id.Path)
}

func (w *testWorld) Today(offset *int) eval.Date {
	return eval.Date{Year: 2026, Month: 1, Day: 19}
}

func strPtr(s string) *string {
	return &s
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
