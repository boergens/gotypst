package gotypst

import (
	"testing"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/layout/pages"
	"github.com/boergens/gotypst/syntax"
)

// mockWorld is a simple World implementation for testing.
type mockWorld struct {
	mainFile eval.FileID
	sources  map[string]*syntax.Source
	library  *eval.Scope
}

func newMockWorld(mainPath string, mainText string) *mockWorld {
	lib := CreateStandardLibrary()
	mainFile := eval.FileID{Path: mainPath}
	sources := make(map[string]*syntax.Source)
	sources[mainPath] = syntax.NewDetachedSource(mainText)

	return &mockWorld{
		mainFile: mainFile,
		sources:  sources,
		library:  lib,
	}
}

func (w *mockWorld) Library() *eval.Scope {
	return w.library
}

func (w *mockWorld) MainFile() eval.FileID {
	return w.mainFile
}

func (w *mockWorld) Source(id eval.FileID) (*syntax.Source, error) {
	src, ok := w.sources[id.Path]
	if !ok {
		return nil, &fileNotFoundError{path: id.Path}
	}
	return src, nil
}

func (w *mockWorld) File(id eval.FileID) ([]byte, error) {
	return nil, &fileNotFoundError{path: id.Path}
}

func (w *mockWorld) Today(offset *int) eval.Date {
	return eval.Date{Year: 2026, Month: 1, Day: 19}
}

type fileNotFoundError struct {
	path string
}

func (e *fileNotFoundError) Error() string {
	return "file not found: " + e.path
}

func TestCompileHelloWorld(t *testing.T) {
	world := newMockWorld("main.typ", `Hello World`)

	result := Compile(world)

	if !result.Success() {
		for _, err := range result.Errors {
			t.Errorf("Compile error: %s", err.Message)
		}
		t.Fatal("Compilation failed")
	}

	if result.Document == nil {
		t.Fatal("No document produced")
	}

	if len(result.Document.Pages) == 0 {
		t.Error("No pages in document")
	}
}

func TestCompileWithVariable(t *testing.T) {
	world := newMockWorld("main.typ", `#let x = "World"
Hello #x`)

	result := Compile(world)

	if !result.Success() {
		for _, err := range result.Errors {
			t.Errorf("Compile error: %s", err.Message)
		}
		t.Fatal("Compilation failed")
	}

	if result.Document == nil {
		t.Fatal("No document produced")
	}
}

func TestCompileWithFunction(t *testing.T) {
	world := newMockWorld("main.typ", `#let greet(name) = [Hello #name!]
#greet("World")`)

	result := Compile(world)

	if !result.Success() {
		for _, err := range result.Errors {
			t.Errorf("Compile error: %s", err.Message)
		}
		t.Fatal("Compilation failed")
	}
}

func TestCompileParseError(t *testing.T) {
	// Unclosed bracket should cause parse error
	world := newMockWorld("main.typ", `#let x = [unclosed`)

	result := Compile(world)

	if result.Success() {
		t.Error("Expected compilation to fail with parse error")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected at least one error")
	}
}

func TestCompileFileNotFound(t *testing.T) {
	// Create a world that returns file not found
	world := &mockWorld{
		mainFile: eval.FileID{Path: "nonexistent.typ"},
		sources:  make(map[string]*syntax.Source),
		library:  CreateStandardLibrary(),
	}

	result := Compile(world)

	if result.Success() {
		t.Error("Expected compilation to fail with file not found")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected at least one error")
	}
}

func TestCompileResultSuccess(t *testing.T) {
	result := &CompileResult{}
	if result.Success() {
		t.Error("Empty result should not be successful")
	}

	result.Document = &pages.PagedDocument{}
	if !result.Success() {
		t.Error("Result with document and no errors should be successful")
	}

	result.Errors = append(result.Errors, SourceDiagnostic{
		Severity: SeverityError,
		Message:  "test error",
	})
	if result.Success() {
		t.Error("Result with errors should not be successful")
	}
}

func TestCreateStandardLibrary(t *testing.T) {
	lib := CreateStandardLibrary()

	if lib == nil {
		t.Fatal("CreateStandardLibrary returned nil")
	}

	// Check that element functions are registered
	funcs := []string{"raw", "par", "parbreak", "box", "block"}
	for _, name := range funcs {
		binding := lib.Get(name)
		if binding == nil {
			t.Errorf("Standard library should contain %q function", name)
		}
	}
}

func TestCompileEmptyContent(t *testing.T) {
	world := newMockWorld("main.typ", ``)

	result := Compile(world)

	// Empty content should still compile successfully
	if !result.Success() {
		for _, err := range result.Errors {
			t.Errorf("Compile error: %s", err.Message)
		}
		t.Fatal("Empty content should compile successfully")
	}
}

func TestCompileWithBasicExpression(t *testing.T) {
	// Test simple math expression that the evaluator supports
	world := newMockWorld("main.typ", `#let x = 42
#x`)

	result := Compile(world)

	if !result.Success() {
		for _, err := range result.Errors {
			t.Errorf("Compile error: %s", err.Message)
		}
		t.Fatal("Compilation failed")
	}
}

func TestCompileWithStyledContent(t *testing.T) {
	// Test strong and emphasis markup
	world := newMockWorld("main.typ", `*bold* and _italic_ text`)

	result := Compile(world)

	if !result.Success() {
		for _, err := range result.Errors {
			t.Errorf("Compile error: %s", err.Message)
		}
		t.Fatal("Compilation failed")
	}
}

func TestCompileWithRaw(t *testing.T) {
	// Test raw code block
	world := newMockWorld("main.typ", "```python\nprint('hello')\n```")

	result := Compile(world)

	if !result.Success() {
		for _, err := range result.Errors {
			t.Errorf("Compile error: %s", err.Message)
		}
		t.Fatal("Compilation failed")
	}
}
