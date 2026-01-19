// Package tests provides a test harness for running Typst fixture tests.
//
// This package loads .typ fixture files extracted from the Typst test suite
// and runs them through the gotypst parser to validate correctness.
package tests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TestCase represents a single test case extracted from a fixture file.
type TestCase struct {
	// Name is the test identifier (e.g., "let-basic", "comment-block-unclosed")
	Name string

	// Attrs are test attributes (e.g., "paged", "html")
	Attrs []string

	// Code is the Typst source code for this test
	Code string

	// ExpectedErrors are error annotations in the format "line-column message"
	ExpectedErrors []ExpectedError

	// ExpectedHints are hint annotations in the format "line-column message"
	ExpectedHints []ExpectedHint

	// SourceFile is the path to the fixture file this test came from
	SourceFile string

	// StartLine is the line number where this test starts in the source file
	StartLine int
}

// ExpectedError represents an expected error annotation.
type ExpectedError struct {
	// Span is the error location (e.g., "2-7" or "1:5-2:3")
	Span string

	// Message is the expected error message
	Message string

	// Line is the line number of the annotation in the source
	Line int
}

// ExpectedHint represents an expected hint annotation.
type ExpectedHint struct {
	// Span is the hint location
	Span string

	// Message is the expected hint message
	Message string

	// Line is the line number of the annotation in the source
	Line int
}

// Fixture represents a collection of test cases from a single .typ file.
type Fixture struct {
	// Path is the file path of the fixture
	Path string

	// Category is the test category (syntax, foundations, scripting)
	Category string

	// Tests are the individual test cases
	Tests []TestCase
}

// ParseResult represents the result of parsing a test case.
type ParseResult struct {
	// Errors are the actual errors from parsing
	Errors []ParseError

	// Hints are the actual hints from parsing
	Hints []ParseHint

	// AST is a string representation of the parsed AST (for debugging)
	AST string
}

// ParseError represents an actual error from the parser.
type ParseError struct {
	Span    string
	Message string
}

// ParseHint represents an actual hint from the parser.
type ParseHint struct {
	Span    string
	Message string
}

// TestResult represents the result of running a single test case.
type TestResult struct {
	// Test is the test case that was run
	Test TestCase

	// Passed indicates if the test passed
	Passed bool

	// ErrorDiffs are mismatches between expected and actual errors
	ErrorDiffs []string

	// HintDiffs are mismatches between expected and actual hints
	HintDiffs []string

	// ParseResult is the result from parsing
	ParseResult ParseResult
}

// Parser is the interface that the syntax package must implement.
type Parser interface {
	// Parse parses Typst source code and returns the result.
	Parse(source string) ParseResult
}

// Harness is the test harness for running fixture tests.
type Harness struct {
	// FixturesDir is the path to the fixtures directory
	FixturesDir string

	// Parser is the parser implementation to use
	Parser Parser

	// Verbose enables verbose output
	Verbose bool
}

// NewHarness creates a new test harness.
func NewHarness(fixturesDir string) *Harness {
	return &Harness{
		FixturesDir: fixturesDir,
	}
}

// SetParser sets the parser implementation.
func (h *Harness) SetParser(p Parser) {
	h.Parser = p
}

// delimiterRegex matches the test delimiter lines: "--- name attr1 attr2 ---"
var delimiterRegex = regexp.MustCompile(`^---\s+([a-zA-Z0-9_-]+)(?:\s+(.+?))?\s+---\s*$`)

// errorAnnotationRegex matches error annotations: "// Error: span message"
var errorAnnotationRegex = regexp.MustCompile(`^//\s*Error:\s*(\S+)\s+(.+)$`)

// hintAnnotationRegex matches hint annotations: "// Hint: span message"
var hintAnnotationRegex = regexp.MustCompile(`^//\s*Hint:\s*(\S+)\s+(.+)$`)

// LoadFixture loads and parses a single fixture file.
func LoadFixture(path string) (*Fixture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open fixture: %w", err)
	}
	defer file.Close()

	// Determine category from path
	dir := filepath.Dir(path)
	category := filepath.Base(dir)

	fixture := &Fixture{
		Path:     path,
		Category: category,
		Tests:    []TestCase{},
	}

	scanner := bufio.NewScanner(file)
	var currentTest *TestCase
	var codeBuilder strings.Builder
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for delimiter
		if matches := delimiterRegex.FindStringSubmatch(line); matches != nil {
			// Save previous test if any
			if currentTest != nil {
				currentTest.Code = strings.TrimSpace(codeBuilder.String())
				fixture.Tests = append(fixture.Tests, *currentTest)
			}

			// Start new test
			name := matches[1]
			var attrs []string
			if matches[2] != "" {
				attrs = strings.Fields(matches[2])
			}

			currentTest = &TestCase{
				Name:       name,
				Attrs:      attrs,
				SourceFile: path,
				StartLine:  lineNum,
			}
			codeBuilder.Reset()
			continue
		}

		// If we're in a test, process the line
		if currentTest != nil {
			// Check for error annotation
			if matches := errorAnnotationRegex.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
				currentTest.ExpectedErrors = append(currentTest.ExpectedErrors, ExpectedError{
					Span:    matches[1],
					Message: matches[2],
					Line:    lineNum - currentTest.StartLine,
				})
			}

			// Check for hint annotation
			if matches := hintAnnotationRegex.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
				currentTest.ExpectedHints = append(currentTest.ExpectedHints, ExpectedHint{
					Span:    matches[1],
					Message: matches[2],
					Line:    lineNum - currentTest.StartLine,
				})
			}

			codeBuilder.WriteString(line)
			codeBuilder.WriteString("\n")
		}
	}

	// Save last test
	if currentTest != nil {
		currentTest.Code = strings.TrimSpace(codeBuilder.String())
		fixture.Tests = append(fixture.Tests, *currentTest)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read fixture: %w", err)
	}

	return fixture, nil
}

// LoadAllFixtures loads all fixtures from the fixtures directory.
func (h *Harness) LoadAllFixtures() ([]*Fixture, error) {
	var fixtures []*Fixture

	err := filepath.Walk(h.FixturesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".typ") {
			fixture, err := LoadFixture(path)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", path, err)
			}
			fixtures = append(fixtures, fixture)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fixtures, nil
}

// LoadFixturesByCategory loads fixtures from a specific category.
func (h *Harness) LoadFixturesByCategory(category string) ([]*Fixture, error) {
	categoryDir := filepath.Join(h.FixturesDir, category)

	var fixtures []*Fixture

	err := filepath.Walk(categoryDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".typ") {
			fixture, err := LoadFixture(path)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", path, err)
			}
			fixtures = append(fixtures, fixture)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fixtures, nil
}

// RunTest runs a single test case and returns the result.
func (h *Harness) RunTest(test TestCase) TestResult {
	result := TestResult{
		Test:   test,
		Passed: true,
	}

	if h.Parser == nil {
		// No parser set - just validate fixture loading
		result.ParseResult = ParseResult{}
		return result
	}

	// Parse the test code
	parseResult := h.Parser.Parse(test.Code)
	result.ParseResult = parseResult

	// Compare expected vs actual errors
	result.ErrorDiffs = compareErrors(test.ExpectedErrors, parseResult.Errors)
	if len(result.ErrorDiffs) > 0 {
		result.Passed = false
	}

	// Compare expected vs actual hints
	result.HintDiffs = compareHints(test.ExpectedHints, parseResult.Hints)
	if len(result.HintDiffs) > 0 {
		result.Passed = false
	}

	return result
}

// RunFixture runs all tests in a fixture and returns results.
func (h *Harness) RunFixture(fixture *Fixture) []TestResult {
	var results []TestResult
	for _, test := range fixture.Tests {
		results = append(results, h.RunTest(test))
	}
	return results
}

// RunAll runs all loaded fixtures and returns results.
func (h *Harness) RunAll() ([]TestResult, error) {
	fixtures, err := h.LoadAllFixtures()
	if err != nil {
		return nil, err
	}

	var results []TestResult
	for _, fixture := range fixtures {
		results = append(results, h.RunFixture(fixture)...)
	}
	return results, nil
}

// Summary returns a summary of test results.
type Summary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

// Summarize generates a summary from test results.
func Summarize(results []TestResult) Summary {
	summary := Summary{Total: len(results)}
	for _, r := range results {
		if r.Passed {
			summary.Passed++
		} else {
			summary.Failed++
		}
	}
	return summary
}

// FormatResult formats a test result for display.
func FormatResult(r TestResult) string {
	var sb strings.Builder

	status := "PASS"
	if !r.Passed {
		status = "FAIL"
	}

	sb.WriteString(fmt.Sprintf("[%s] %s (%s:%d)\n", status, r.Test.Name, r.Test.SourceFile, r.Test.StartLine))

	if !r.Passed {
		for _, diff := range r.ErrorDiffs {
			sb.WriteString(fmt.Sprintf("  Error mismatch: %s\n", diff))
		}
		for _, diff := range r.HintDiffs {
			sb.WriteString(fmt.Sprintf("  Hint mismatch: %s\n", diff))
		}
	}

	return sb.String()
}

// compareErrors compares expected and actual errors.
func compareErrors(expected []ExpectedError, actual []ParseError) []string {
	var diffs []string

	// Build maps for comparison
	expectedSet := make(map[string]bool)
	for _, e := range expected {
		key := fmt.Sprintf("%s: %s", e.Span, e.Message)
		expectedSet[key] = true
	}

	actualSet := make(map[string]bool)
	for _, a := range actual {
		key := fmt.Sprintf("%s: %s", a.Span, a.Message)
		actualSet[key] = true
	}

	// Find missing expected errors
	for key := range expectedSet {
		if !actualSet[key] {
			diffs = append(diffs, fmt.Sprintf("expected error not found: %s", key))
		}
	}

	// Find unexpected errors
	for key := range actualSet {
		if !expectedSet[key] {
			diffs = append(diffs, fmt.Sprintf("unexpected error: %s", key))
		}
	}

	return diffs
}

// compareHints compares expected and actual hints.
func compareHints(expected []ExpectedHint, actual []ParseHint) []string {
	var diffs []string

	// Build maps for comparison
	expectedSet := make(map[string]bool)
	for _, e := range expected {
		key := fmt.Sprintf("%s: %s", e.Span, e.Message)
		expectedSet[key] = true
	}

	actualSet := make(map[string]bool)
	for _, a := range actual {
		key := fmt.Sprintf("%s: %s", a.Span, a.Message)
		actualSet[key] = true
	}

	// Find missing expected hints
	for key := range expectedSet {
		if !actualSet[key] {
			diffs = append(diffs, fmt.Sprintf("expected hint not found: %s", key))
		}
	}

	// Find unexpected hints
	for key := range actualSet {
		if !expectedSet[key] {
			diffs = append(diffs, fmt.Sprintf("unexpected hint: %s", key))
		}
	}

	return diffs
}

// Categories returns the list of fixture categories.
func (h *Harness) Categories() ([]string, error) {
	entries, err := os.ReadDir(h.FixturesDir)
	if err != nil {
		return nil, err
	}

	var categories []string
	for _, entry := range entries {
		if entry.IsDir() {
			categories = append(categories, entry.Name())
		}
	}
	return categories, nil
}

// TestsInCategory returns the number of tests in a category.
func (h *Harness) TestsInCategory(category string) (int, error) {
	fixtures, err := h.LoadFixturesByCategory(category)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, f := range fixtures {
		count += len(f.Tests)
	}
	return count, nil
}
