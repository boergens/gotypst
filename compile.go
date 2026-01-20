// Package gotypst provides a Go implementation of the Typst typesetting system.
//
// This file implements the compile pipeline that wires together:
// Parse -> Evaluate -> Realize -> Layout -> Render

package gotypst

import (
	"fmt"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/layout/pages"
	"github.com/boergens/gotypst/syntax"
)

// CompileResult holds the result of a compilation.
type CompileResult struct {
	// Document is the laid out document, nil if compilation failed.
	Document *pages.PagedDocument

	// Warnings contains non-fatal warnings generated during compilation.
	Warnings []SourceDiagnostic

	// Errors contains fatal errors that prevented compilation.
	Errors []SourceDiagnostic
}

// Success returns true if compilation completed without errors.
func (r *CompileResult) Success() bool {
	return len(r.Errors) == 0 && r.Document != nil
}

// SourceDiagnostic represents a diagnostic message with source location.
type SourceDiagnostic struct {
	// Span is the source location of the diagnostic.
	Span syntax.Span

	// Severity indicates error or warning.
	Severity DiagnosticSeverity

	// Message is the diagnostic message.
	Message string

	// Hints are optional suggestions for resolving the issue.
	Hints []string
}

// DiagnosticSeverity indicates the severity of a diagnostic.
type DiagnosticSeverity int

const (
	SeverityError DiagnosticSeverity = iota
	SeverityWarning
)

// Compile compiles a Typst document from the given world.
//
// The compilation pipeline consists of:
//  1. Parse: Read and parse the main source file
//  2. Evaluate: Execute the source to produce content
//  3. Realize: Transform content by applying show rules (currently pass-through)
//  4. Layout: Arrange content into pages
//
// The World interface provides access to source files, the standard library,
// and other resources needed during compilation.
func Compile(world eval.World) *CompileResult {
	result := &CompileResult{}

	// Step 1: Get the main source file
	mainFile := world.MainFile()
	source, err := world.Source(mainFile)
	if err != nil {
		result.Errors = append(result.Errors, SourceDiagnostic{
			Severity: SeverityError,
			Message:  fmt.Sprintf("cannot read main file: %v", err),
		})
		return result
	}

	// Step 2: Parse and evaluate the source
	content, warnings, err := evaluate(world, source, mainFile)
	result.Warnings = append(result.Warnings, warnings...)
	if err != nil {
		result.Errors = append(result.Errors, diagnosticFromError(err))
		return result
	}

	// Step 3: Layout the content into pages
	// Create a layout engine
	layoutEngine := &pages.Engine{
		World: world,
	}

	// Create an empty style chain for now
	styles := pages.StyleChain{}

	// Convert eval.Content to pages.Content using wrapper elements
	var elemInterfaces []interface{}
	for _, elem := range content.Elements {
		elemInterfaces = append(elemInterfaces, elem)
	}
	pagesContent := pages.ContentFromEval(elemInterfaces)

	// Layout the document
	doc, err := pages.LayoutDocument(layoutEngine, pagesContent, styles)
	if err != nil {
		result.Errors = append(result.Errors, diagnosticFromError(err))
		return result
	}

	result.Document = doc
	return result
}

// evaluate parses and evaluates a source file.
func evaluate(world eval.World, source *syntax.Source, fileID eval.FileID) (*eval.Content, []SourceDiagnostic, error) {
	var warnings []SourceDiagnostic

	// Check for parser errors
	root := source.Root()
	if root == nil {
		return nil, warnings, fmt.Errorf("source has no root")
	}

	if errs := root.Errors(); len(errs) > 0 {
		return nil, warnings, fmt.Errorf("parse error: %v", errs[0])
	}

	// Create the evaluation engine
	engine := eval.NewEngine(world)

	// Create scopes with the standard library
	scopes := eval.NewScopes(world.Library())

	// Create the VM for evaluation
	vm := eval.NewVm(engine, eval.NewContext(), scopes, root.Span())

	// Get markup from root
	markup := syntax.MarkupNodeFromNode(root)
	if markup == nil {
		return nil, warnings, fmt.Errorf("source root is not markup")
	}

	// Evaluate the markup content
	value, err := eval.EvalMarkup(vm, markup)
	if err != nil {
		return nil, warnings, err
	}

	// Check for forbidden flow events at top level
	if vm.HasFlow() {
		flow := vm.Flow
		switch flow.(type) {
		case eval.BreakEvent:
			return nil, warnings, fmt.Errorf("break is not allowed at the top level")
		case eval.ContinueEvent:
			return nil, warnings, fmt.Errorf("continue is not allowed at the top level")
		case eval.ReturnEvent:
			return nil, warnings, fmt.Errorf("return is not allowed at the top level")
		}
	}

	// Extract content value
	if cv, ok := value.(eval.ContentValue); ok {
		// Collect warnings from the engine sink
		for _, w := range engine.Sink.Warnings {
			warnings = append(warnings, SourceDiagnostic{
				Span:     w.Span,
				Severity: SeverityWarning,
				Message:  w.Message,
				Hints:    w.Hints,
			})
		}
		return &cv.Content, warnings, nil
	}

	return nil, warnings, fmt.Errorf("evaluation did not produce content")
}

// diagnosticFromError creates a SourceDiagnostic from an error.
func diagnosticFromError(err error) SourceDiagnostic {
	// Check for typed errors with span information
	if spanErr, ok := err.(interface{ Span() syntax.Span }); ok {
		return SourceDiagnostic{
			Span:     spanErr.Span(),
			Severity: SeverityError,
			Message:  err.Error(),
		}
	}

	return SourceDiagnostic{
		Severity: SeverityError,
		Message:  err.Error(),
	}
}

// CompileOptions configures the compilation process.
type CompileOptions struct {
	// TraceSpans enables tracing for the given spans (for IDE support).
	TraceSpans []syntax.Span
}

// CompileWithOptions compiles a Typst document with the given options.
func CompileWithOptions(world eval.World, opts CompileOptions) *CompileResult {
	// For now, ignore options and use the basic compile
	// TODO: Support tracing and other options
	return Compile(world)
}

// CreateStandardLibrary creates a standard library scope with all built-in functions.
//
// This should be called once and passed to NewFileWorld via WithLibrary option.
func CreateStandardLibrary() *eval.Scope {
	lib := eval.NewScope()

	// Register element functions (raw, par, parbreak, box, block, etc.)
	eval.RegisterElementFunctions(lib)

	// TODO: Register more standard library functions as they are implemented
	// - text functions (text, emph, strong, etc.)
	// - layout functions (page, grid, table, etc.)
	// - math functions
	// - data functions (str, int, float, etc.)

	return lib
}
