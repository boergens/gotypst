// Import evaluation for Typst.
// Translated from typst-eval/src/import.rs

package eval

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// evalModuleImport evaluates a module import expression.
//
// Matches Rust's impl Eval for ast::ModuleImport.
func evalModuleImport(vm *Vm, e *syntax.ModuleImportExpr) (Value, error) {
	sourceExpr := e.Source()
	if sourceExpr == nil {
		return nil, &ImportError{
			Message: "import requires a source",
			Span:    e.ToUntyped().Span(),
		}
	}

	sourceSpan := sourceExpr.ToUntyped().Span()

	// Evaluate the source expression.
	source, err := EvalExpr(vm, sourceExpr)
	if err != nil {
		return nil, err
	}

	replacedSource := false

	// Handle different source types.
	switch v := source.(type) {
	case FuncValue:
		// Check if function has a scope (only native functions do).
		if v.Func == nil || v.Func.Scope() == nil {
			return nil, &ImportError{
				Message: "cannot import from user-defined functions",
				Span:    sourceSpan,
			}
		}

	case TypeValue:
		// Types have scopes, nothing special to do.

	case ModuleValue:
		// Already a module, nothing to do.

	case StrValue:
		// String path - import file or package.
		module, err := Import(vm.Engine, string(v), sourceSpan)
		if err != nil {
			return nil, err
		}
		source = ModuleValue{Module: module}
		replacedSource = true

	default:
		return nil, &ImportError{
			Message: fmt.Sprintf("expected path, module, function, or type, found %s", source.Type()),
			Span:    sourceSpan,
		}
	}

	// If there is a rename, import the source itself under that name.
	newName := e.NewName()
	if newName != nil {
		// Warn on `import x as x` (same name).
		if ident, ok := sourceExpr.(*syntax.IdentExpr); ok {
			if ident.Get() == newName.Get() {
				// TODO: emit warning "unnecessary import rename to same name"
			}
		}

		// Define renamed module on the scope.
		vm.Define(newName, source)
	}

	// Get the scope from the source.
	scope := valueScope(source)
	if scope == nil {
		return nil, &ImportError{
			Message: fmt.Sprintf("cannot get scope from %s", source.Type()),
			Span:    sourceSpan,
		}
	}

	imports := e.Imports()
	switch imp := imports.(type) {
	case nil:
		// No imports clause - bare import.
		if newName == nil {
			// Derive name from source expression.
			name := deriveImportName(sourceExpr, replacedSource)
			if name == "" {
				return nil, &ImportError{
					Message: "dynamic import requires an explicit name",
					Hint:    "you can name the import with `as`",
					Span:    sourceSpan,
				}
			}
			vm.Scopes.Top().Bind(name, foundations.NewBinding(source, sourceSpan))
		}

	case *syntax.ImportsWildcard:
		// Wildcard: import "file.typ": * → bind all exports.
		scope.Iter(func(name string, binding foundations.Binding) {
			vm.Scopes.Top().Bind(name, binding.Clone())
		})

	case *syntax.ImportItemsNode:
		// Specific items: import "file.typ": a, b → bind specific items.
		var errors []error
		for _, item := range imp.Items() {
			path := item.Path()
			if len(path) == 0 {
				continue
			}

			currentScope := scope
			var binding *foundations.Binding

			for i, componentName := range path {
				componentSpan := sourceSpan // Use source span for string paths

				binding = currentScope.Get(componentName)
				if binding == nil {
					errors = append(errors, &ImportError{
						Message: fmt.Sprintf("unresolved import: %s", componentName),
						Span:    componentSpan,
					})
					break
				}

				if i < len(path)-1 {
					// Nested import - this must be a submodule.
					value := binding.Read()
					subScope := valueScope(value)
					if subScope == nil {
						var errMsg string
						if fv, ok := value.(FuncValue); ok && fv.Func != nil && fv.Func.Scope() == nil {
							errMsg = "cannot import from user-defined functions"
						} else {
							errMsg = fmt.Sprintf("expected module, function, or type, found %s", value.Type())
						}
						errors = append(errors, &ImportError{
							Message: errMsg,
							Span:    componentSpan,
						})
						binding = nil
						break
					}
					currentScope = subScope
				}
			}

			if binding != nil {
				// Bind the item using its bound name.
				boundName := item.BoundName()
				if boundName != nil {
					vm.Define(boundName, binding.Read())
				}
			}
		}

		if len(errors) > 0 {
			// Return first error (could collect all).
			return nil, errors[0]
		}
	}

	return None, nil
}

// deriveImportName derives a name for a bare import from the source expression.
func deriveImportName(sourceExpr syntax.Expr, replacedSource bool) string {
	// For string literals, derive from path.
	if str, ok := sourceExpr.(*syntax.StrExpr); ok {
		path := str.Get()
		return deriveNameFromPath(path)
	}

	// For identifiers, use the identifier name if not replaced.
	if ident, ok := sourceExpr.(*syntax.IdentExpr); ok && !replacedSource {
		return ident.Get()
	}

	// Dynamic imports need explicit names.
	return ""
}

// deriveNameFromPath derives a module name from a file path.
func deriveNameFromPath(path string) string {
	// Handle package imports.
	if strings.HasPrefix(path, "@") {
		// @namespace/name:version -> name
		path = strings.TrimPrefix(path, "@")
		if idx := strings.Index(path, "/"); idx >= 0 {
			path = path[idx+1:]
		}
		if idx := strings.Index(path, ":"); idx >= 0 {
			path = path[:idx]
		}
		return makeValidIdent(path)
	}

	// Regular file path.
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return makeValidIdent(name)
}

// makeValidIdent converts a string to a valid identifier.
func makeValidIdent(name string) string {
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	if name == "" {
		return ""
	}
	// Check if it's a valid identifier.
	// For now, just ensure it doesn't start with a digit.
	if name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}
	return name
}

// evalModuleInclude evaluates a module include expression.
//
// Matches Rust's impl Eval for ast::ModuleInclude.
func evalModuleInclude(vm *Vm, e *syntax.ModuleIncludeExpr) (Value, error) {
	sourceExpr := e.Source()
	if sourceExpr == nil {
		return nil, &ImportError{
			Message: "include requires a source",
			Span:    e.ToUntyped().Span(),
		}
	}

	span := sourceExpr.ToUntyped().Span()

	source, err := EvalExpr(vm, sourceExpr)
	if err != nil {
		return nil, err
	}

	var module *Module
	switch v := source.(type) {
	case StrValue:
		module, err = Import(vm.Engine, string(v), span)
		if err != nil {
			return nil, err
		}

	case ModuleValue:
		module = v.Module

	default:
		return nil, &ImportError{
			Message: fmt.Sprintf("expected path or module, found %s", source.Type()),
			Span:    span,
		}
	}

	return ContentValue{Content: module.Content}, nil
}

// Import imports a file or package by path.
//
// Matches Rust's import function.
func Import(engine *foundations.Engine, from string, span syntax.Span) (*Module, error) {
	if strings.HasPrefix(from, "@") {
		// Package import: @namespace/name:version
		spec, err := syntax.ParsePackageSpec(from)
		if err != nil {
			return nil, &ImportError{
				Message: fmt.Sprintf("invalid package specification: %v", err),
				Span:    span,
			}
		}
		return importPackage(engine, spec, span)
	}

	// File import - resolve relative to current file.
	id, err := resolvePathToFileId(engine, from, span)
	if err != nil {
		return nil, err
	}
	return importFile(engine, id, span)
}

// importFile imports a file by its ID.
//
// Matches Rust's import_file function.
func importFile(engine *foundations.Engine, id syntax.FileId, span syntax.Span) (*Module, error) {
	// Load the source file.
	source, err := engine.World.Source(id)
	if err != nil {
		return nil, &ImportError{
			Message: fmt.Sprintf("cannot read file: %v", err),
			Span:    span,
		}
	}

	// Prevent cyclic importing.
	if engine.Route != nil && engine.Route.Contains(source.Id()) {
		return nil, &ImportError{
			Message: "cyclic import",
			Span:    span,
		}
	}

	// Evaluate the file.
	return EvalSource(engine, source, span)
}

// importPackage imports an external package.
//
// Matches Rust's import_package function.
func importPackage(engine *foundations.Engine, spec *syntax.PackageSpec, span syntax.Span) (*Module, error) {
	name, id, err := resolvePackage(engine, spec, span)
	if err != nil {
		return nil, err
	}

	module, err := importFile(engine, id, span)
	if err != nil {
		return nil, err
	}

	// Override module name with package name.
	module.Name = name
	return module, nil
}

// resolvePackage resolves the name and entrypoint of a package.
//
// Matches Rust's resolve_package function.
func resolvePackage(engine *foundations.Engine, spec *syntax.PackageSpec, span syntax.Span) (string, syntax.FileId, error) {
	// Create file ID for the package manifest.
	manifestVPath, err := syntax.NewVirtualPath("typst.toml")
	if err != nil {
		return "", syntax.FileId{}, &ImportError{
			Message: fmt.Sprintf("invalid manifest path: %v", err),
			Span:    span,
		}
	}
	manifestPath := syntax.NewRootedPath(
		syntax.PackageRoot(*spec),
		*manifestVPath,
	)
	manifestId := manifestPath.Intern()

	// Load the manifest bytes.
	bytes, err := engine.World.File(manifestId)
	if err != nil {
		return "", syntax.FileId{}, &ImportError{
			Message: fmt.Sprintf("cannot read package manifest: %v", err),
			Span:    span,
		}
	}

	// Parse the manifest.
	manifest, err := parseManifest(string(bytes))
	if err != nil {
		return "", syntax.FileId{}, &ImportError{
			Message: fmt.Sprintf("package manifest is malformed: %v", err),
			Span:    span,
		}
	}

	// Validate against spec.
	if err := validateManifest(manifest, spec); err != nil {
		return "", syntax.FileId{}, &ImportError{
			Message: err.Error(),
			Span:    span,
		}
	}

	// Get the entry point.
	entrypoint := manifest.Entrypoint
	if entrypoint == "" {
		entrypoint = "lib.typ"
	}

	entryVPath, err := syntax.NewVirtualPath(entrypoint)
	if err != nil {
		return "", syntax.FileId{}, &ImportError{
			Message: fmt.Sprintf("invalid entrypoint path: %v", err),
			Span:    span,
		}
	}
	entryPath := syntax.NewRootedPath(
		syntax.PackageRoot(*spec),
		*entryVPath,
	)
	return manifest.Name, entryPath.Intern(), nil
}

// resolvePathToFileId resolves a path string to a FileId.
func resolvePathToFileId(engine *foundations.Engine, path string, span syntax.Span) (syntax.FileId, error) {
	// Get the current file's ID from the span.
	currentId := span.Id()

	var resolvedPath string
	if filepath.IsAbs(path) {
		resolvedPath = path
	} else if currentId != nil {
		// Resolve relative to current file.
		rootedPath := currentId.Get()
		if rootedPath != nil {
			vpath := rootedPath.VPath()
			if vpath != nil {
				currentVPath := vpath.String()
				dir := filepath.Dir(currentVPath)
				resolvedPath = filepath.Join(dir, path)
			} else {
				resolvedPath = path
			}
		} else {
			resolvedPath = path
		}
	} else {
		resolvedPath = path
	}

	// Clean the path.
	resolvedPath = filepath.Clean(resolvedPath)

	// Create a file ID for the resolved path (project-local file).
	vpath, err := syntax.NewVirtualPath(resolvedPath)
	if err != nil {
		return syntax.FileId{}, &ImportError{
			Message: fmt.Sprintf("invalid path: %v", err),
			Span:    span,
		}
	}
	rootedPath := syntax.NewRootedPath(
		syntax.ProjectRoot(),
		*vpath,
	)
	return rootedPath.Intern(), nil
}

// EvalSource evaluates a source file and returns a module.
//
// This is a simplified version that creates a module from source.
func EvalSource(engine *foundations.Engine, source *syntax.Source, span syntax.Span) (*Module, error) {
	root := source.Root()
	if root == nil {
		return nil, &ImportError{
			Message: "source has no root",
			Span:    span,
		}
	}

	// Check for parser errors.
	if errs := root.Errors(); len(errs) > 0 {
		return nil, &ImportError{
			Message: fmt.Sprintf("parse error: %v", errs[0]),
			Span:    span,
		}
	}

	// Create a new context for module evaluation.
	context := foundations.NewContext()

	// Create scopes with the standard library.
	scopes := foundations.NewScopes(engine.World.Library())

	// Create a new VM for module evaluation.
	moduleVm := NewVm(engine, context, scopes, root.Span())

	// Push this file onto the route if we have one.
	if engine.Route != nil {
		engine.Route.Push(source.Id())
		defer engine.Route.Pop()
	}

	// Evaluate the markup content.
	markup := syntax.MarkupNodeFromNode(root)
	if markup == nil {
		return nil, &ImportError{
			Message: "source root is not markup",
			Span:    span,
		}
	}

	content, err := evalMarkup(moduleVm, markup)
	if err != nil {
		return nil, err
	}

	// Check for forbidden flow events at top level.
	if moduleVm.HasFlow() {
		flow := moduleVm.Flow
		switch flow.(type) {
		case BreakEvent:
			return nil, &ImportError{
				Message: "break is not allowed at the top level",
				Span:    flow.Span(),
			}
		case ContinueEvent:
			return nil, &ImportError{
				Message: "continue is not allowed at the top level",
				Span:    flow.Span(),
			}
		case ReturnEvent:
			return nil, &ImportError{
				Message: "return is not allowed at the top level",
				Span:    flow.Span(),
			}
		}
	}

	// Extract content.
	var moduleContent Content
	if cv, ok := content.(ContentValue); ok {
		moduleContent = cv.Content
	}

	// Create the module with the exported scope.
	moduleName := deriveModuleName(source.Id())
	module := &Module{
		Name:    moduleName,
		Scope:   scopes.Top(),
		Content: moduleContent,
	}

	return module, nil
}

// deriveModuleName derives a module name from a file ID.
func deriveModuleName(id syntax.FileId) string {
	rootedPath := id.Get()
	if rootedPath == nil {
		return "module"
	}

	vpath := rootedPath.VPath()
	if vpath == nil {
		return "module"
	}

	pathStr := vpath.String()
	if pathStr == "" {
		return "module"
	}

	base := filepath.Base(pathStr)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Convert to valid identifier.
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")

	if name == "" {
		return "module"
	}

	return name
}

// valueScope returns the scope of a value, if it has one.
func valueScope(v Value) *Scope {
	switch val := v.(type) {
	case FuncValue:
		if val.Func != nil {
			return val.Func.Scope()
		}
	case TypeValue:
		return val.Inner.Scope()
	case ModuleValue:
		if val.Module != nil {
			return val.Module.Scope
		}
	}
	return nil
}

// Manifest represents a package manifest.
type Manifest struct {
	Name       string
	Version    string
	Entrypoint string
}

// parseManifest parses a TOML package manifest.
// This is a simplified parser - real implementation would use a TOML library.
func parseManifest(content string) (*Manifest, error) {
	manifest := &Manifest{
		Entrypoint: "lib.typ",
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")

		switch key {
		case "name":
			manifest.Name = value
		case "version":
			manifest.Version = value
		case "entrypoint":
			manifest.Entrypoint = value
		}
	}

	if manifest.Name == "" {
		return nil, fmt.Errorf("missing name")
	}

	return manifest, nil
}

// validateManifest validates that the manifest matches the package spec.
func validateManifest(manifest *Manifest, spec *syntax.PackageSpec) error {
	if manifest.Name != spec.Name {
		return fmt.Errorf("package name mismatch: expected %s, got %s", spec.Name, manifest.Name)
	}
	// Version validation could be added here.
	return nil
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// ImportError represents an error during import.
type ImportError struct {
	Message string
	Hint    string
	Span    syntax.Span
}

func (e *ImportError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s (hint: %s)", e.Message, e.Hint)
	}
	return e.Message
}
