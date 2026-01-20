package eval

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/boergens/gotypst/syntax"
)

// evalModuleImport evaluates a module import expression.
//
// Import syntax:
//   - import "file.typ"              → bind module as filename
//   - import "file.typ" as x         → bind module as x
//   - import "file.typ": *           → bind all exports
//   - import "file.typ": a, b        → bind specific items
//   - import func                    → import from function scope
//   - import type                    → import from type scope
func evalModuleImport(vm *Vm, e *syntax.ModuleImportExpr) (Value, error) {
	sourceExpr := e.Source()
	if sourceExpr == nil {
		return nil, &ImportError{
			Message: "import requires a source",
			Span:    e.ToUntyped().Span(),
		}
	}

	// Evaluate the source expression
	source, err := EvalExpr(vm, sourceExpr)
	if err != nil {
		return nil, err
	}

	span := e.ToUntyped().Span()

	// Resolve the source to a module
	module, err := resolveModuleSource(vm, source, span)
	if err != nil {
		return nil, err
	}

	// Handle import bindings based on the import pattern
	imports := e.Imports()
	newName := e.NewName()

	if imports == nil && newName == nil {
		// Bare import: import "file.typ" → bind as module name
		vm.Define(module.Name, ModuleValue{Module: module})
	} else if newName != nil {
		// Renamed import: import "file.typ" as x → bind as x
		vm.Define(newName.Get(), ModuleValue{Module: module})
	} else {
		// Import with items
		switch imp := imports.(type) {
		case *syntax.ImportsWildcard:
			// Wildcard: import "file.typ": * → bind all exports
			if err := importAllExports(vm, module, span); err != nil {
				return nil, err
			}
		case *syntax.ImportItemsNode:
			// Specific items: import "file.typ": a, b → bind specific items
			if err := importItems(vm, module, imp.Items(), span); err != nil {
				return nil, err
			}
		}
	}

	return None, nil
}

// evalModuleInclude evaluates a module include expression.
//
// Include returns the content of the included file, not the module.
// Syntax: include "file.typ"
func evalModuleInclude(vm *Vm, e *syntax.ModuleIncludeExpr) (Value, error) {
	sourceExpr := e.Source()
	if sourceExpr == nil {
		return nil, &ImportError{
			Message: "include requires a source",
			Span:    e.ToUntyped().Span(),
		}
	}

	// Evaluate the source expression
	source, err := EvalExpr(vm, sourceExpr)
	if err != nil {
		return nil, err
	}

	span := e.ToUntyped().Span()

	// Source must be a string path
	path, ok := AsStr(source)
	if !ok {
		return nil, &ImportError{
			Message: fmt.Sprintf("include source must be a string, got %s", source.Type()),
			Span:    span,
		}
	}

	// Import the file
	module, err := importPath(vm, path, span)
	if err != nil {
		return nil, err
	}

	// Return the module's content
	return ContentValue{Content: module.Content}, nil
}

// resolveModuleSource resolves an import source to a module.
//
// The source can be:
//   - A string path (file or package)
//   - A function (imports its scope)
//   - A type (imports its scope)
//   - A module (used as-is)
func resolveModuleSource(vm *Vm, source Value, span syntax.Span) (*Module, error) {
	switch v := source.(type) {
	case StrValue:
		// String path - import file or package
		path := string(v)
		return importPath(vm, path, span)

	case FuncValue:
		// Import from function scope
		return functionToModule(v.Func, span)

	case TypeValue:
		// Import from type scope
		return typeToModule(v.Inner, span)

	case ModuleValue:
		// Already a module
		return v.Module, nil

	default:
		return nil, &ImportError{
			Message: fmt.Sprintf("cannot import from %s", source.Type()),
			Span:    span,
		}
	}
}

// importPath imports a file or package by path.
func importPath(vm *Vm, path string, span syntax.Span) (*Module, error) {
	if strings.HasPrefix(path, "@") {
		// Package import: @namespace/name:version
		return importPackage(vm, path, span)
	}

	// File import
	return importFile(vm, path, span)
}

// importFile imports a file by its path.
func importFile(vm *Vm, path string, span syntax.Span) (*Module, error) {
	// Resolve the file ID
	fileID, err := resolveFilePath(vm, path, span)
	if err != nil {
		return nil, err
	}

	// Check for cyclic imports
	if vm.Engine.Route.Contains(fileID) {
		return nil, &CyclicImportError{
			File: fileID,
			Span: span,
		}
	}

	// Push this file onto the route
	vm.Engine.Route.Push(fileID)
	defer vm.Engine.Route.Pop()

	// Load the source
	source, err := vm.World().Source(fileID)
	if err != nil {
		return nil, &ImportError{
			Message: fmt.Sprintf("cannot read file: %v", err),
			Span:    span,
		}
	}

	// Evaluate the file
	return evalModule(vm, source, fileID)
}

// importPackage imports a package by its specification.
func importPackage(vm *Vm, spec string, span syntax.Span) (*Module, error) {
	// Parse the package specification: @namespace/name:version
	pkgSpec, err := parsePackageSpec(spec)
	if err != nil {
		return nil, &ImportError{
			Message: err.Error(),
			Span:    span,
		}
	}

	// Resolve the package to a file ID
	fileID, pkgName, err := resolvePackage(vm, pkgSpec, span)
	if err != nil {
		return nil, err
	}

	// Check for cyclic imports
	if vm.Engine.Route.Contains(fileID) {
		return nil, &CyclicImportError{
			File: fileID,
			Span: span,
		}
	}

	// Push this file onto the route
	vm.Engine.Route.Push(fileID)
	defer vm.Engine.Route.Pop()

	// Load the source
	source, err := vm.World().Source(fileID)
	if err != nil {
		return nil, &ImportError{
			Message: fmt.Sprintf("cannot read package: %v", err),
			Span:    span,
		}
	}

	// Evaluate the file
	module, err := evalModule(vm, source, fileID)
	if err != nil {
		return nil, err
	}

	// Override module name with package name
	module.Name = pkgName

	return module, nil
}

// resolveFilePath resolves a file path to a FileID.
func resolveFilePath(vm *Vm, path string, span syntax.Span) (FileID, error) {
	// Get the current file's directory for relative paths
	currentFile := vm.Engine.Route.CurrentFile()

	var resolvedPath string
	if filepath.IsAbs(path) {
		resolvedPath = path
	} else if currentFile != nil {
		// Resolve relative to current file
		dir := filepath.Dir(currentFile.Path)
		resolvedPath = filepath.Join(dir, path)
	} else {
		// Resolve relative to main file
		mainFile := vm.World().MainFile()
		dir := filepath.Dir(mainFile.Path)
		resolvedPath = filepath.Join(dir, path)
	}

	// Clean the path
	resolvedPath = filepath.Clean(resolvedPath)

	return FileID{
		Package: nil,
		Path:    resolvedPath,
	}, nil
}

// parsePackageSpec parses a package specification string.
// Format: @namespace/name:version
func parsePackageSpec(spec string) (*PackageSpec, error) {
	if !strings.HasPrefix(spec, "@") {
		return nil, fmt.Errorf("package specification must start with @")
	}

	spec = spec[1:] // Remove @

	// Split by colon for version
	parts := strings.SplitN(spec, ":", 2)
	pathPart := parts[0]
	var version Version

	if len(parts) == 2 {
		var err error
		version, err = parseVersion(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid version: %v", err)
		}
	}

	// Split namespace and name
	pathParts := strings.Split(pathPart, "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("package specification must have namespace/name format")
	}

	namespace := pathParts[0]
	name := strings.Join(pathParts[1:], "/")

	return &PackageSpec{
		Namespace: namespace,
		Name:      name,
		Version:   version,
	}, nil
}

// parseVersion parses a semantic version string.
func parseVersion(s string) (Version, error) {
	var v Version
	parts := strings.Split(s, ".")

	if len(parts) >= 1 {
		if _, err := fmt.Sscanf(parts[0], "%d", &v.Major); err != nil {
			return v, fmt.Errorf("invalid major version: %s", parts[0])
		}
	}
	if len(parts) >= 2 {
		if _, err := fmt.Sscanf(parts[1], "%d", &v.Minor); err != nil {
			return v, fmt.Errorf("invalid minor version: %s", parts[1])
		}
	}
	if len(parts) >= 3 {
		if _, err := fmt.Sscanf(parts[2], "%d", &v.Patch); err != nil {
			return v, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return v, nil
}

// resolvePackage resolves a package specification to a file ID and package name.
func resolvePackage(vm *Vm, spec *PackageSpec, span syntax.Span) (FileID, string, error) {
	// Create file ID for the package manifest
	manifestID := FileID{
		Package: spec,
		Path:    "typst.toml",
	}

	// Load the manifest
	manifestBytes, err := vm.World().File(manifestID)
	if err != nil {
		return FileID{}, "", &ImportError{
			Message: fmt.Sprintf("cannot read package manifest: %v", err),
			Span:    span,
		}
	}

	// Parse the manifest (simplified - real implementation would use TOML parser)
	manifest, err := parsePackageManifest(manifestBytes)
	if err != nil {
		return FileID{}, "", &ImportError{
			Message: fmt.Sprintf("invalid package manifest: %v", err),
			Span:    span,
		}
	}

	// Validate the package spec
	if err := validatePackageManifest(manifest, spec); err != nil {
		return FileID{}, "", &ImportError{
			Message: err.Error(),
			Span:    span,
		}
	}

	// Return the entry point file ID
	entryPoint := manifest.EntryPoint
	if entryPoint == "" {
		entryPoint = "lib.typ"
	}

	return FileID{
		Package: spec,
		Path:    entryPoint,
	}, manifest.Name, nil
}

// PackageManifest represents the typst.toml manifest file.
type PackageManifest struct {
	Name        string
	Version     Version
	EntryPoint  string
	Description string
	Authors     []string
	License     string
}

// parsePackageManifest parses a package manifest from bytes.
// This is a simplified implementation - real version would use proper TOML parsing.
func parsePackageManifest(data []byte) (*PackageManifest, error) {
	manifest := &PackageManifest{
		EntryPoint: "lib.typ",
	}

	content := string(data)
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
			v, err := parseVersion(value)
			if err != nil {
				return nil, err
			}
			manifest.Version = v
		case "entrypoint":
			manifest.EntryPoint = value
		case "description":
			manifest.Description = value
		case "license":
			manifest.License = value
		}
	}

	if manifest.Name == "" {
		return nil, fmt.Errorf("package manifest missing name")
	}

	return manifest, nil
}

// validatePackageManifest validates that the manifest matches the package spec.
func validatePackageManifest(manifest *PackageManifest, spec *PackageSpec) error {
	if manifest.Name != spec.Name {
		return fmt.Errorf("package name mismatch: expected %s, got %s", spec.Name, manifest.Name)
	}

	// Version checking (if specified)
	if spec.Version.Major > 0 || spec.Version.Minor > 0 || spec.Version.Patch > 0 {
		if manifest.Version.Major != spec.Version.Major {
			return fmt.Errorf("package major version mismatch: expected %d, got %d",
				spec.Version.Major, manifest.Version.Major)
		}
		// Minor version must be >= requested
		if manifest.Version.Minor < spec.Version.Minor {
			return fmt.Errorf("package minor version too old: expected >= %d.%d, got %d.%d",
				spec.Version.Major, spec.Version.Minor,
				manifest.Version.Major, manifest.Version.Minor)
		}
	}

	return nil
}

// evalModule evaluates a source file and returns a module.
func evalModule(vm *Vm, source *syntax.Source, fileID FileID) (*Module, error) {
	// Parse the source if needed
	root := source.Root()
	if root == nil {
		return nil, &ImportError{
			Message: "source has no root",
			Span:    syntax.Span{},
		}
	}

	// Check for parser errors
	if errs := root.Errors(); len(errs) > 0 {
		// Report the first error
		return nil, &ImportError{
			Message: fmt.Sprintf("parse error: %v", errs[0]),
			Span:    syntax.Span{},
		}
	}

	// Create a new VM for module evaluation
	scopes := NewScopes(vm.World().Library())
	moduleVm := NewVm(vm.Engine, NewContext(), scopes, root.Span())

	// Evaluate the markup content
	markup := syntax.MarkupNodeFromNode(root)
	if markup == nil {
		return nil, &ImportError{
			Message: "source root is not markup",
			Span:    syntax.Span{},
		}
	}

	content, err := EvalMarkup(moduleVm, markup)
	if err != nil {
		return nil, err
	}

	// Check for forbidden flow events at top level
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

	// Extract content
	var moduleContent Content
	if cv, ok := content.(ContentValue); ok {
		moduleContent = cv.Content
	}

	// Create the module with the exported scope
	moduleName := deriveModuleName(fileID.Path)
	module := &Module{
		Name:    moduleName,
		Scope:   scopes.Top(),
		Content: moduleContent,
	}

	return module, nil
}

// deriveModuleName derives a module name from a file path.
func deriveModuleName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Convert to valid identifier
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")

	return name
}

// functionToModule converts a function's scope to a module.
func functionToModule(fn *Func, span syntax.Span) (*Module, error) {
	if fn == nil {
		return nil, &ImportError{
			Message: "cannot import from nil function",
			Span:    span,
		}
	}

	// Get the function's scope
	scope := NewScope()

	// For closure functions, we can expose the captured scope
	if fn.Repr != nil {
		if cf, ok := fn.Repr.(ClosureFunc); ok && cf.Closure != nil {
			if cf.Closure.Captured != nil {
				scope = cf.Closure.Captured.Clone()
			}
		}
	}

	// Create a module with the function name
	name := "function"
	if fn.Name != nil {
		name = *fn.Name
	}

	return &Module{
		Name:    name,
		Scope:   scope,
		Content: Content{},
	}, nil
}

// typeToModule converts a type's scope to a module.
func typeToModule(t Type, span syntax.Span) (*Module, error) {
	// Types have associated methods/constructors that form their "scope"
	scope := NewScope()

	// For now, return an empty module - type scopes would be populated
	// by the standard library
	return &Module{
		Name:    t.String(),
		Scope:   scope,
		Content: Content{},
	}, nil
}

// importAllExports binds all exports from a module to the current scope.
func importAllExports(vm *Vm, module *Module, span syntax.Span) error {
	if module.Scope == nil {
		return nil
	}

	// Iterate over all bindings in the module's scope
	for name, binding := range module.Scope.All() {
		// Skip private bindings (those starting with underscore)
		if strings.HasPrefix(name, "_") {
			continue
		}

		// Define in the current scope
		vm.Bind(name, binding)
	}

	return nil
}

// importItems binds specific items from a module to the current scope.
func importItems(vm *Vm, module *Module, items []*syntax.ImportItem, span syntax.Span) error {
	if module.Scope == nil {
		return &ImportError{
			Message: "module has no exports",
			Span:    span,
		}
	}

	for _, item := range items {
		// Get the path (can be dotted for nested access)
		path := item.Path()
		if len(path) == 0 {
			continue
		}

		// Resolve the value through the path
		value, err := resolveImportPath(module, path, span)
		if err != nil {
			return err
		}

		// Determine the binding name
		bindName := path[len(path)-1] // Default: last element of path
		if newName := item.NewName(); newName != nil {
			bindName = newName.Get()
		}

		// Define in the current scope
		vm.Define(bindName, value)
	}

	return nil
}

// resolveImportPath resolves a dotted path in a module's scope.
func resolveImportPath(module *Module, path []string, span syntax.Span) (Value, error) {
	if len(path) == 0 {
		return nil, &ImportError{
			Message: "empty import path",
			Span:    span,
		}
	}

	// Start with the first element
	name := path[0]
	binding := module.Scope.Get(name)
	if binding == nil {
		return nil, &ImportError{
			Message: fmt.Sprintf("'%s' is not exported from module '%s'", name, module.Name),
			Span:    span,
		}
	}

	value := binding.Value

	// Traverse the rest of the path
	for i := 1; i < len(path); i++ {
		fieldName := path[i]

		switch v := value.(type) {
		case ModuleValue:
			if v.Module == nil || v.Module.Scope == nil {
				return nil, &ImportError{
					Message: fmt.Sprintf("cannot access '%s' in module", fieldName),
					Span:    span,
				}
			}
			nextBinding := v.Module.Scope.Get(fieldName)
			if nextBinding == nil {
				return nil, &ImportError{
					Message: fmt.Sprintf("'%s' is not exported from module", fieldName),
					Span:    span,
				}
			}
			value = nextBinding.Value

		case *DictValue, DictValue:
			dict, _ := AsDict(value)
			if val, ok := dict.Get(fieldName); ok {
				value = val
			} else {
				return nil, &ImportError{
					Message: fmt.Sprintf("field '%s' not found", fieldName),
					Span:    span,
				}
			}

		default:
			return nil, &ImportError{
				Message: fmt.Sprintf("cannot access '%s' on %s", fieldName, value.Type()),
				Span:    span,
			}
		}
	}

	return value, nil
}

// ----------------------------------------------------------------------------
// Route Extensions
// ----------------------------------------------------------------------------

// CurrentFile returns the current file being evaluated, or nil if none.
func (r *Route) CurrentFile() *FileID {
	if len(r.files) == 0 {
		return nil
	}
	return &r.files[len(r.files)-1]
}

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// ImportError represents an error during import.
type ImportError struct {
	Message string
	Span    syntax.Span
}

func (e *ImportError) Error() string {
	return e.Message
}
