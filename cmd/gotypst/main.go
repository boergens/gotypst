// Package main provides the CLI entry point for gotypst.
//
// Usage:
//
//	gotypst compile input.typ -o output.pdf
//	gotypst compile input.typ                   # outputs to input.pdf
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/layout/pages"
	"github.com/boergens/gotypst/pdf"
	"github.com/boergens/gotypst/realize"
	"github.com/boergens/gotypst/syntax"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compile", "c":
		if err := runCompile(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	case "version", "-v", "--version":
		printVersion()
	default:
		// Assume single argument is input file for compile
		if err := runCompile(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println(`gotypst - A Go implementation of Typst

Usage:
  gotypst compile <input.typ> [-o <output.pdf>]
  gotypst <input.typ> [-o <output.pdf>]
  gotypst help
  gotypst version

Commands:
  compile, c    Compile a Typst document to PDF
  help          Show this help message
  version       Show version information

Options:
  -o, --output  Output file path (default: input file with .pdf extension)
  --root        Project root directory (default: input file directory)
  --font-path   Additional font directories (can be specified multiple times)`)
}

func printVersion() {
	fmt.Println("gotypst version 0.1.0")
}

func runCompile(args []string) error {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)
	output := fs.String("o", "", "Output file path")
	outputLong := fs.String("output", "", "Output file path (long form)")
	root := fs.String("root", "", "Project root directory")
	var fontPaths []string
	fs.Func("font-path", "Additional font directory", func(s string) error {
		fontPaths = append(fontPaths, s)
		return nil
	})

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("missing input file")
	}

	input := fs.Arg(0)

	// Determine output path
	outPath := *output
	if outPath == "" {
		outPath = *outputLong
	}
	if outPath == "" {
		// Default to input file with .pdf extension
		ext := filepath.Ext(input)
		outPath = strings.TrimSuffix(input, ext) + ".pdf"
	}

	// Determine project root
	projectRoot := *root
	if projectRoot == "" {
		projectRoot = filepath.Dir(input)
	}

	return compile(input, outPath, projectRoot, fontPaths)
}

// compile performs the full compilation pipeline:
// Parse -> Evaluate -> Layout -> Render
func compile(inputPath, outputPath, projectRoot string, fontPaths []string) error {
	// Get absolute paths
	absInput, err := filepath.Abs(inputPath)
	if err != nil {
		return fmt.Errorf("cannot resolve input path: %w", err)
	}

	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("cannot resolve project root: %w", err)
	}

	// Create the FileWorld
	opts := []eval.FileWorldOption{}
	if len(fontPaths) > 0 {
		opts = append(opts, eval.WithFontDirs(fontPaths...))
	}

	// Get relative path from root
	mainPath, err := filepath.Rel(absRoot, absInput)
	if err != nil {
		mainPath = absInput
	}

	world, err := eval.NewFileWorld(absRoot, mainPath, opts...)
	if err != nil {
		return fmt.Errorf("cannot create world: %w", err)
	}

	// Set up standard library
	stdlib := buildStandardLibrary()
	world = mustRebuildWorldWithLibrary(world, stdlib)

	// Get and parse the main source
	source, err := world.Source(world.MainFile())
	if err != nil {
		return fmt.Errorf("cannot read source: %w", err)
	}

	// Check for parse errors
	if errs := source.Root().Errors(); len(errs) > 0 {
		return formatParseErrors(source, errs)
	}

	// Evaluate the source
	content, err := evaluate(world, source)
	if err != nil {
		return fmt.Errorf("evaluation failed: %w", err)
	}

	// Layout the document
	doc, err := layout(world, content)
	if err != nil {
		return fmt.Errorf("layout failed: %w", err)
	}

	// Render to PDF
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer outFile.Close()

	if err := pdf.Export(doc, outFile); err != nil {
		return fmt.Errorf("PDF export failed: %w", err)
	}

	fmt.Printf("Compiled %s -> %s\n", inputPath, outputPath)
	return nil
}

// buildStandardLibrary constructs the standard library scope.
func buildStandardLibrary() *eval.Scope {
	return eval.Library()
}

// mustRebuildWorldWithLibrary creates a new FileWorld with the given library.
// This is a workaround since FileWorld doesn't allow changing library after creation.
func mustRebuildWorldWithLibrary(old *eval.FileWorld, lib *eval.Scope) *eval.FileWorld {
	mainFile := old.MainFile()
	root := old.Root()
	fontBook := old.FontBook()

	world, err := eval.NewFileWorld(root, mainFile.Path,
		eval.WithLibrary(lib),
		eval.WithFontBook(fontBook),
	)
	if err != nil {
		// Should not happen if old world was valid
		panic(fmt.Sprintf("failed to rebuild world: %v", err))
	}
	return world
}

// evaluate evaluates the source and returns content.
func evaluate(world *eval.FileWorld, source *syntax.Source) (*eval.Content, error) {
	// Create the evaluation engine
	engine := eval.NewEngine(world)

	// Create VM for evaluation
	scopes := eval.NewScopes(world.Library())
	ctx := eval.NewContext()
	vm := eval.NewVm(engine, ctx, scopes, source.Root().Span())

	// Get the markup node
	markup := syntax.MarkupNodeFromNode(source.Root())
	if markup == nil {
		return nil, fmt.Errorf("source root is not markup")
	}

	// Evaluate the markup to get content
	value, err := evalMarkup(vm, markup)
	if err != nil {
		return nil, err
	}

	// Extract content from the value
	if cv, ok := value.(eval.ContentValue); ok {
		return &cv.Content, nil
	}

	return &eval.Content{}, nil
}

// evalMarkup evaluates markup and returns content.
// This wraps the internal eval function.
func evalMarkup(vm *eval.Vm, markup *syntax.MarkupNode) (eval.Value, error) {
	// Evaluate all expressions in the markup
	var elements []eval.ContentElement

	for _, expr := range markup.Exprs() {
		val, err := eval.EvalExpr(vm, expr)
		if err != nil {
			return nil, err
		}
		if vm.HasFlow() {
			break
		}

		// Convert value to content elements
		if cv, ok := val.(eval.ContentValue); ok {
			elements = append(elements, cv.Content.Elements...)
		}
	}

	return eval.ContentValue{Content: eval.Content{Elements: elements}}, nil
}

// layout converts evaluated content to a paged document.
// This is the main entry point that wires up realization and page collection.
func layout(world *eval.FileWorld, content *eval.Content) (*pages.PagedDocument, error) {
	// Create evaluation engine for realization
	evalEngine := eval.NewEngine(world)

	// Create empty styles for initial realization
	realizeStyles := realize.EmptyStyleChain()

	// Realize the content - apply show rules, group elements, collapse spaces
	realizedPairs, err := realize.Realize(
		realize.LayoutDocument{},
		evalEngine,
		content,
		realizeStyles,
	)
	if err != nil {
		return nil, fmt.Errorf("realization failed: %w", err)
	}

	// Convert realized pairs to pages.Content
	pageContent := convertRealizedContent(realizedPairs)

	// Create layout engine
	layoutEngine := &pages.Engine{
		World: world,
	}

	// Create default style chain for layout
	layoutStyles := pages.StyleChain{
		Styles: map[string]interface{}{},
	}

	// Layout the document
	return pages.LayoutDocument(layoutEngine, pageContent, layoutStyles)
}

// convertRealizedContent converts realized pairs to pages.Content.
// This bridges the realize package output to the pages package input.
func convertRealizedContent(pairs []realize.Pair) *pages.Content {
	if len(pairs) == 0 {
		return &pages.Content{
			Elements: make([]eval.ContentElement, 0),
		}
	}

	elements := make([]eval.ContentElement, 0, len(pairs))
	for _, pair := range pairs {
		if pair.Element != nil {
			elements = append(elements, pair.Element)
		}
	}

	return &pages.Content{
		Elements: elements,
	}
}

// formatParseErrors formats parse errors for display.
func formatParseErrors(source *syntax.Source, errs []*syntax.SyntaxError) error {
	if len(errs) == 0 {
		return nil
	}

	var msgs []string
	for _, err := range errs {
		msgs = append(msgs, err.Message)
	}

	return fmt.Errorf("parse errors:\n  %s", strings.Join(msgs, "\n  "))
}
