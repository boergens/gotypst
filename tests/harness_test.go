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
	// Try relative path first (when running from tests/ directory)
	if _, err := os.Stat("fixtures"); err == nil {
		return "fixtures"
	}
	// Try from repo root
	if _, err := os.Stat("tests/fixtures"); err == nil {
		return "tests/fixtures"
	}
	t.Fatal("cannot find fixtures directory")
	return ""
}

func TestLoadFixture(t *testing.T) {
	fixturesDir := getFixturesDir(t)

	tests := []struct {
		name          string
		file          string
		wantCategory  string
		wantTestCount int
		wantFirstTest string
	}{
		{
			name:          "syntax/comment.typ",
			file:          filepath.Join(fixturesDir, "syntax", "comment.typ"),
			wantCategory:  "syntax",
			wantTestCount: 4,
			wantFirstTest: "comments",
		},
		{
			name:          "syntax/numbers.typ",
			file:          filepath.Join(fixturesDir, "syntax", "numbers.typ"),
			wantCategory:  "syntax",
			wantTestCount: 1,
			wantFirstTest: "numbers",
		},
		{
			name:          "scripting/let.typ",
			file:          filepath.Join(fixturesDir, "scripting", "let.typ"),
			wantCategory:  "scripting",
			wantTestCount: 10, // Approximate, depends on fixture content
			wantFirstTest: "let-basic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture, err := LoadFixture(tt.file)
			if err != nil {
				t.Fatalf("LoadFixture() error = %v", err)
			}

			if fixture.Category != tt.wantCategory {
				t.Errorf("Category = %v, want %v", fixture.Category, tt.wantCategory)
			}

			if len(fixture.Tests) < 1 {
				t.Fatal("Expected at least one test case")
			}

			if fixture.Tests[0].Name != tt.wantFirstTest {
				t.Errorf("First test name = %v, want %v", fixture.Tests[0].Name, tt.wantFirstTest)
			}
		})
	}
}

func TestLoadFixture_ParsesErrorAnnotations(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	fixture, err := LoadFixture(filepath.Join(fixturesDir, "syntax", "comment.typ"))
	if err != nil {
		t.Fatalf("LoadFixture() error = %v", err)
	}

	// Find the test case with expected errors
	var errorTest *TestCase
	for i := range fixture.Tests {
		if fixture.Tests[i].Name == "comment-block-unclosed" {
			errorTest = &fixture.Tests[i]
			break
		}
	}

	if errorTest == nil {
		t.Fatal("Expected to find comment-block-unclosed test")
	}

	if len(errorTest.ExpectedErrors) == 0 {
		t.Error("Expected at least one error annotation")
	}

	// Check the error annotation content
	found := false
	for _, err := range errorTest.ExpectedErrors {
		if err.Span == "7-9" && strings.Contains(err.Message, "unexpected end of block comment") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find error annotation '7-9 unexpected end of block comment'")
	}

	// Check for hint annotation
	if len(errorTest.ExpectedHints) == 0 {
		t.Error("Expected at least one hint annotation")
	}
}

func TestLoadFixture_ParsesAttributes(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	fixture, err := LoadFixture(filepath.Join(fixturesDir, "syntax", "comment.typ"))
	if err != nil {
		t.Fatalf("LoadFixture() error = %v", err)
	}

	// The first test should have "paged" attribute
	if len(fixture.Tests) == 0 {
		t.Fatal("No tests found")
	}

	found := false
	for _, attr := range fixture.Tests[0].Attrs {
		if attr == "paged" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'paged' attribute, got %v", fixture.Tests[0].Attrs)
	}
}

func TestHarness_LoadAllFixtures(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	harness := NewHarness(fixturesDir)

	fixtures, err := harness.LoadAllFixtures()
	if err != nil {
		t.Fatalf("LoadAllFixtures() error = %v", err)
	}

	// Should load fixtures from all categories
	if len(fixtures) < 3 {
		t.Errorf("Expected at least 3 fixtures, got %d", len(fixtures))
	}

	// Verify categories are represented
	categories := make(map[string]bool)
	for _, f := range fixtures {
		categories[f.Category] = true
	}

	expectedCategories := []string{"syntax", "foundations", "scripting"}
	for _, cat := range expectedCategories {
		if !categories[cat] {
			t.Errorf("Expected category %s to be present", cat)
		}
	}
}

func TestHarness_LoadFixturesByCategory(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	harness := NewHarness(fixturesDir)

	tests := []struct {
		category      string
		wantMinCount  int
	}{
		{"syntax", 3},      // comment, numbers, escape, etc.
		{"foundations", 3}, // int, str, array
		{"scripting", 3},   // let, if, for, ops
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			fixtures, err := harness.LoadFixturesByCategory(tt.category)
			if err != nil {
				t.Fatalf("LoadFixturesByCategory(%q) error = %v", tt.category, err)
			}

			if len(fixtures) < tt.wantMinCount {
				t.Errorf("Got %d fixtures, want at least %d", len(fixtures), tt.wantMinCount)
			}

			// Verify all fixtures are from the expected category
			for _, f := range fixtures {
				if f.Category != tt.category {
					t.Errorf("Fixture %s has category %s, want %s", f.Path, f.Category, tt.category)
				}
			}
		})
	}
}

func TestHarness_Categories(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	harness := NewHarness(fixturesDir)

	categories, err := harness.Categories()
	if err != nil {
		t.Fatalf("Categories() error = %v", err)
	}

	expected := map[string]bool{
		"syntax":      false,
		"foundations": false,
		"scripting":   false,
	}

	for _, cat := range categories {
		if _, ok := expected[cat]; ok {
			expected[cat] = true
		}
	}

	for cat, found := range expected {
		if !found {
			t.Errorf("Expected category %s not found", cat)
		}
	}
}

// StubParser is a no-op parser for testing the harness infrastructure.
type StubParser struct{}

func (p *StubParser) Parse(source string) ParseResult {
	return ParseResult{}
}

func TestHarness_RunTest_WithoutParser(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	harness := NewHarness(fixturesDir)

	fixture, err := LoadFixture(filepath.Join(fixturesDir, "syntax", "numbers.typ"))
	if err != nil {
		t.Fatalf("LoadFixture() error = %v", err)
	}

	if len(fixture.Tests) == 0 {
		t.Fatal("No tests in fixture")
	}

	// Run without parser - should pass (no errors expected, no parser to generate them)
	result := harness.RunTest(fixture.Tests[0])
	if !result.Passed {
		t.Errorf("Test should pass when no parser is set")
	}
}

func TestHarness_RunTest_WithStubParser(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	harness := NewHarness(fixturesDir)
	harness.SetParser(&StubParser{})

	fixture, err := LoadFixture(filepath.Join(fixturesDir, "syntax", "numbers.typ"))
	if err != nil {
		t.Fatalf("LoadFixture() error = %v", err)
	}

	if len(fixture.Tests) == 0 {
		t.Fatal("No tests in fixture")
	}

	// Test without expected errors should pass with stub parser
	result := harness.RunTest(fixture.Tests[0])
	if !result.Passed {
		t.Errorf("Test without expected errors should pass with stub parser")
	}
}

func TestHarness_RunTest_ExpectedErrorMissing(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	harness := NewHarness(fixturesDir)
	harness.SetParser(&StubParser{})

	fixture, err := LoadFixture(filepath.Join(fixturesDir, "syntax", "comment.typ"))
	if err != nil {
		t.Fatalf("LoadFixture() error = %v", err)
	}

	// Find a test with expected errors
	var errorTest *TestCase
	for i := range fixture.Tests {
		if len(fixture.Tests[i].ExpectedErrors) > 0 {
			errorTest = &fixture.Tests[i]
			break
		}
	}

	if errorTest == nil {
		t.Skip("No test with expected errors found")
	}

	// With stub parser (no errors), tests expecting errors should fail
	result := harness.RunTest(*errorTest)
	if result.Passed {
		t.Errorf("Test expecting errors should fail when parser produces none")
	}

	if len(result.ErrorDiffs) == 0 {
		t.Errorf("Expected error diffs to be populated")
	}
}

func TestSummarize(t *testing.T) {
	results := []TestResult{
		{Passed: true},
		{Passed: true},
		{Passed: false},
		{Passed: true},
		{Passed: false},
	}

	summary := Summarize(results)

	if summary.Total != 5 {
		t.Errorf("Total = %d, want 5", summary.Total)
	}
	if summary.Passed != 3 {
		t.Errorf("Passed = %d, want 3", summary.Passed)
	}
	if summary.Failed != 2 {
		t.Errorf("Failed = %d, want 2", summary.Failed)
	}
}

func TestFormatResult(t *testing.T) {
	result := TestResult{
		Test: TestCase{
			Name:       "test-name",
			SourceFile: "test.typ",
			StartLine:  10,
		},
		Passed:     false,
		ErrorDiffs: []string{"expected error not found: 1-5: some error"},
	}

	formatted := FormatResult(result)

	if !strings.Contains(formatted, "FAIL") {
		t.Error("Expected FAIL in output")
	}
	if !strings.Contains(formatted, "test-name") {
		t.Error("Expected test name in output")
	}
	if !strings.Contains(formatted, "Error mismatch") {
		t.Error("Expected error mismatch in output")
	}
}

// TestAllFixturesLoad verifies that all fixtures can be loaded without error.
func TestAllFixturesLoad(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	harness := NewHarness(fixturesDir)

	fixtures, err := harness.LoadAllFixtures()
	if err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	totalTests := 0
	for _, f := range fixtures {
		t.Logf("Loaded %s: %d tests", f.Path, len(f.Tests))
		totalTests += len(f.Tests)

		// Verify each test has required fields
		for _, test := range f.Tests {
			if test.Name == "" {
				t.Errorf("Test in %s has empty name", f.Path)
			}
			if test.Code == "" {
				t.Errorf("Test %s in %s has empty code", test.Name, f.Path)
			}
		}
	}

	t.Logf("Total fixtures: %d, Total tests: %d", len(fixtures), totalTests)
}

// TestRunAllWithStubParser runs all fixtures with a stub parser to verify harness works end-to-end.
func TestRunAllWithStubParser(t *testing.T) {
	fixturesDir := getFixturesDir(t)
	harness := NewHarness(fixturesDir)
	harness.SetParser(&StubParser{})

	results, err := harness.RunAll()
	if err != nil {
		t.Fatalf("RunAll() error = %v", err)
	}

	summary := Summarize(results)
	t.Logf("Results: %d total, %d passed, %d failed", summary.Total, summary.Passed, summary.Failed)

	// With a stub parser, tests without expected errors should pass,
	// and tests with expected errors should fail
	if summary.Total == 0 {
		t.Error("Expected at least some test results")
	}
}

// BenchmarkLoadAllFixtures benchmarks fixture loading.
func BenchmarkLoadAllFixtures(b *testing.B) {
	// Find fixtures dir
	fixturesDir := "fixtures"
	if _, err := os.Stat(fixturesDir); err != nil {
		fixturesDir = "tests/fixtures"
	}

	harness := NewHarness(fixturesDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := harness.LoadAllFixtures()
		if err != nil {
			b.Fatal(err)
		}
	}
}
