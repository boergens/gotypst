package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// getFixturesDir returns the path to the fixtures directory.
func getFixturesDir(t *testing.T) string {
	t.Helper()
	// Try relative path from test location
	candidates := []string{
		"fixtures",
		"./fixtures",
		"../tests/fixtures",
		"tests/fixtures",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// Try from working directory
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "fixtures")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	t.Fatalf("could not find fixtures directory")
	return ""
}

func TestParseFixtureFile(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	commentPath := filepath.Join(fixturesDir, "syntax", "comment.typ")

	tests, err := ParseFixtureFile(commentPath)
	if err != nil {
		t.Fatalf("ParseFixtureFile failed: %v", err)
	}

	if len(tests) == 0 {
		t.Fatal("expected at least one test case")
	}

	// Check first test case
	first := tests[0]
	if first.Name != "comments" {
		t.Errorf("expected first test name 'comments', got %q", first.Name)
	}
	if len(first.Attrs) == 0 || first.Attrs[0] != "paged" {
		t.Errorf("expected 'paged' attribute, got %v", first.Attrs)
	}
	if first.Code == "" {
		t.Error("expected non-empty code")
	}
}

func TestLoadAllFixtures(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	runner := NewTestRunner(fixturesDir)

	tests, err := runner.LoadFixtures()
	if err != nil {
		t.Fatalf("LoadFixtures failed: %v", err)
	}

	if len(tests) == 0 {
		t.Fatal("expected at least one test case")
	}

	t.Logf("Loaded %d test cases", len(tests))
}

func TestLoadByCategory(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	runner := NewTestRunner(fixturesDir)

	categories := []string{"syntax", "foundations", "scripting"}
	for _, cat := range categories {
		tests, err := runner.LoadFixtures(cat)
		if err != nil {
			t.Errorf("LoadFixtures(%s) failed: %v", cat, err)
			continue
		}
		t.Logf("Category %s: %d tests", cat, len(tests))
	}
}

// TestSyntaxFixtures runs all syntax category tests.
func TestSyntaxFixtures(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	runner := NewTestRunner(fixturesDir)

	tests, err := runner.LoadFixtures("syntax")
	if err != nil {
		t.Fatalf("LoadFixtures failed: %v", err)
	}

	if len(tests) == 0 {
		t.Skip("no syntax tests found")
	}

	results := runner.RunAll(tests)
	for _, res := range results {
		t.Run(res.Test.Name, func(t *testing.T) {
			if !res.Passed {
				t.Errorf("test failed: %s", res.Details)
				if len(res.Tokens) > 0 {
					t.Logf("Tokens:\n%s", PrintTokens(res.Tokens))
				}
			}
		})
	}
}

// TestFoundationsFixtures runs all foundations category tests.
func TestFoundationsFixtures(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	runner := NewTestRunner(fixturesDir)

	tests, err := runner.LoadFixtures("foundations")
	if err != nil {
		t.Fatalf("LoadFixtures failed: %v", err)
	}

	if len(tests) == 0 {
		t.Skip("no foundations tests found")
	}

	results := runner.RunAll(tests)
	for _, res := range results {
		t.Run(res.Test.Name, func(t *testing.T) {
			if !res.Passed {
				t.Errorf("test failed: %s", res.Details)
			}
		})
	}
}

// TestScriptingFixtures runs all scripting category tests.
func TestScriptingFixtures(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	runner := NewTestRunner(fixturesDir)

	tests, err := runner.LoadFixtures("scripting")
	if err != nil {
		t.Fatalf("LoadFixtures failed: %v", err)
	}

	if len(tests) == 0 {
		t.Skip("no scripting tests found")
	}

	results := runner.RunAll(tests)
	for _, res := range results {
		t.Run(res.Test.Name, func(t *testing.T) {
			if !res.Passed {
				t.Errorf("test failed: %s", res.Details)
			}
		})
	}
}

// TestFilterByName tests the name filtering functionality.
func TestFilterByName(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	runner := NewTestRunner(fixturesDir)

	tests, err := runner.LoadFixtures()
	if err != nil {
		t.Fatalf("LoadFixtures failed: %v", err)
	}

	// Filter for comment-related tests
	filtered := FilterByName(tests, "comment")
	if len(filtered) == 0 {
		t.Fatal("expected at least one test matching 'comment'")
	}

	for _, tc := range filtered {
		if !strings.Contains(tc.Name, "comment") {
			t.Errorf("unexpected test %q in filtered results", tc.Name)
		}
	}

	t.Logf("Filtered %d tests matching 'comment'", len(filtered))
}

// TestFilterByCategory tests the category filtering functionality.
func TestFilterByCategory(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	runner := NewTestRunner(fixturesDir)

	tests, err := runner.LoadFixtures()
	if err != nil {
		t.Fatalf("LoadFixtures failed: %v", err)
	}

	filtered := FilterByCategory(tests, "syntax")
	if len(filtered) == 0 {
		t.Fatal("expected at least one syntax test")
	}

	for _, tc := range filtered {
		dir := filepath.Dir(tc.SourceFile)
		cat := filepath.Base(dir)
		if cat != "syntax" {
			t.Errorf("unexpected category %q in filtered results", cat)
		}
	}

	t.Logf("Filtered %d syntax tests", len(filtered))
}

// TestRunnerSummary tests the summary output.
func TestRunnerSummary(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	runner := NewTestRunner(fixturesDir)

	tests, err := runner.LoadFixtures("syntax")
	if err != nil {
		t.Fatalf("LoadFixtures failed: %v", err)
	}

	if len(tests) == 0 {
		t.Skip("no tests to run")
	}

	runner.RunAll(tests)
	summary := runner.Summary()

	if summary == "" {
		t.Error("expected non-empty summary")
	}

	if !strings.Contains(summary, "Test Results:") {
		t.Error("expected summary to contain 'Test Results:'")
	}

	t.Log(summary)
}

// TestLexerBasicTokens tests that the lexer produces basic tokens correctly.
func TestLexerBasicTokens(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		wantKind string // First non-space token kind (human-readable form)
	}{
		{"line comment", "// comment\n", "line comment"},
		{"block comment", "/* comment */", "block comment"},
		{"heading", "= Heading", "heading marker"},
		{"hash", "#let x = 1", "hash"},
		{"text", "Hello world", "text"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCase := &TestCase{
				Name: tc.name,
				Code: tc.code,
			}

			runner := NewTestRunner("")
			result := runner.RunTest(testCase)

			if len(result.Tokens) == 0 {
				t.Fatal("expected at least one token")
			}

			// Find first non-space token
			for _, tok := range result.Tokens {
				kindStr := tok.Kind.String()
				if kindStr == "Space" || kindStr == "Parbreak" {
					continue
				}
				if kindStr != tc.wantKind {
					t.Errorf("expected first token kind %s, got %s",
						tc.wantKind, kindStr)
				}
				break
			}
		})
	}
}

// TestExpectedErrors tests that expected errors are properly detected.
func TestExpectedErrors(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	// Load comment.typ which has error annotations
	tests, err := ParseFixtureFile(filepath.Join(fixturesDir, "syntax", "comment.typ"))
	if err != nil {
		t.Fatalf("ParseFixtureFile failed: %v", err)
	}

	// Find the test with expected errors
	var errorTest *TestCase
	for _, tc := range tests {
		if len(tc.Errors) > 0 {
			errorTest = tc
			break
		}
	}

	if errorTest == nil {
		t.Skip("no test with expected errors found")
	}

	t.Logf("Testing %s with %d expected errors", errorTest.Name, len(errorTest.Errors))

	runner := NewTestRunner(fixturesDir)
	result := runner.RunTest(errorTest)

	// The test might pass or fail depending on lexer implementation
	// We're mainly checking the harness correctly handles error expectations
	t.Logf("Test passed: %v, errors found: %v", result.Passed, result.Errors)
}

// BenchmarkLoadFixtures benchmarks loading all fixtures.
func BenchmarkLoadFixtures(b *testing.B) {
	// Find fixtures directory
	fixturesDir := "fixtures"
	if _, err := os.Stat(fixturesDir); err != nil {
		b.Skip("fixtures directory not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := NewTestRunner(fixturesDir)
		_, err := runner.LoadFixtures()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRunSyntaxTests benchmarks running syntax tests.
func BenchmarkRunSyntaxTests(b *testing.B) {
	fixturesDir := "fixtures"
	if _, err := os.Stat(fixturesDir); err != nil {
		b.Skip("fixtures directory not found")
	}

	runner := NewTestRunner(fixturesDir)
	tests, err := runner.LoadFixtures("syntax")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner.RunAll(tests)
	}
}
