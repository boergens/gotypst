package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Package Specification Parsing Tests
// ----------------------------------------------------------------------------

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      *PackageSpec
		wantError bool
	}{
		{
			name:  "simple package",
			input: "@preview/example:1.0.0",
			want: &PackageSpec{
				Namespace: "preview",
				Name:      "example",
				Version:   Version{Major: 1, Minor: 0, Patch: 0},
			},
		},
		{
			name:  "package without version",
			input: "@preview/example",
			want: &PackageSpec{
				Namespace: "preview",
				Name:      "example",
				Version:   Version{},
			},
		},
		{
			name:  "nested package name",
			input: "@local/my/nested/pkg:2.1.3",
			want: &PackageSpec{
				Namespace: "local",
				Name:      "my/nested/pkg",
				Version:   Version{Major: 2, Minor: 1, Patch: 3},
			},
		},
		{
			name:  "version with minor only",
			input: "@preview/test:1.2",
			want: &PackageSpec{
				Namespace: "preview",
				Name:      "test",
				Version:   Version{Major: 1, Minor: 2, Patch: 0},
			},
		},
		{
			name:      "missing @ prefix",
			input:     "preview/example:1.0.0",
			wantError: true,
		},
		{
			name:      "missing namespace/name",
			input:     "@example",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePackageSpec(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("parsePackageSpec(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parsePackageSpec(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Namespace != tt.want.Namespace {
				t.Errorf("Namespace = %q, want %q", got.Namespace, tt.want.Namespace)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.want.Version)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Version Parsing Tests
// ----------------------------------------------------------------------------

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Version
		wantError bool
	}{
		{
			name:  "full version",
			input: "1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "major.minor only",
			input: "1.2",
			want:  Version{Major: 1, Minor: 2, Patch: 0},
		},
		{
			name:  "major only",
			input: "1",
			want:  Version{Major: 1, Minor: 0, Patch: 0},
		},
		{
			name:  "zero version",
			input: "0.0.0",
			want:  Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:  "large version numbers",
			input: "100.200.300",
			want:  Version{Major: 100, Minor: 200, Patch: 300},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVersion(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("parseVersion(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseVersion(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Module Name Derivation Tests
// ----------------------------------------------------------------------------

func TestDeriveModuleName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/path/to/file.typ", "file"},
		{"/path/to/my-file.typ", "my_file"},
		{"my file.typ", "my_file"},
		{"simple.typ", "simple"},
		{"/path/to/lib.typ", "lib"},
		{"utils.typ", "utils"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := deriveModuleName(tt.input)
			if got != tt.want {
				t.Errorf("deriveModuleName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Route (Cycle Detection) Tests
// ----------------------------------------------------------------------------

func TestRoute(t *testing.T) {
	t.Run("empty route contains nothing", func(t *testing.T) {
		r := NewRoute()
		if r.Contains(FileID{Path: "test.typ"}) {
			t.Error("empty route should not contain any file")
		}
	})

	t.Run("route contains pushed file", func(t *testing.T) {
		r := NewRoute()
		file := FileID{Path: "test.typ"}
		r.Push(file)
		if !r.Contains(file) {
			t.Error("route should contain pushed file")
		}
	})

	t.Run("route does not contain after pop", func(t *testing.T) {
		r := NewRoute()
		file := FileID{Path: "test.typ"}
		r.Push(file)
		r.Pop()
		if r.Contains(file) {
			t.Error("route should not contain file after pop")
		}
	})

	t.Run("route tracks multiple files", func(t *testing.T) {
		r := NewRoute()
		file1 := FileID{Path: "a.typ"}
		file2 := FileID{Path: "b.typ"}
		file3 := FileID{Path: "c.typ"}

		r.Push(file1)
		r.Push(file2)
		r.Push(file3)

		if !r.Contains(file1) || !r.Contains(file2) || !r.Contains(file3) {
			t.Error("route should contain all pushed files")
		}

		r.Pop()
		if r.Contains(file3) {
			t.Error("route should not contain file3 after pop")
		}
		if !r.Contains(file1) || !r.Contains(file2) {
			t.Error("route should still contain file1 and file2")
		}
	})

	t.Run("current file returns last pushed", func(t *testing.T) {
		r := NewRoute()

		if r.CurrentFile() != nil {
			t.Error("empty route should have nil current file")
		}

		file1 := FileID{Path: "a.typ"}
		file2 := FileID{Path: "b.typ"}

		r.Push(file1)
		if r.CurrentFile() == nil || r.CurrentFile().Path != "a.typ" {
			t.Error("current file should be a.typ")
		}

		r.Push(file2)
		if r.CurrentFile() == nil || r.CurrentFile().Path != "b.typ" {
			t.Error("current file should be b.typ")
		}
	})

	t.Run("clone creates independent copy", func(t *testing.T) {
		r := NewRoute()
		r.Push(FileID{Path: "test.typ"})

		clone := r.Clone()
		clone.Push(FileID{Path: "other.typ"})

		if r.Contains(FileID{Path: "other.typ"}) {
			t.Error("original route should not be affected by clone modifications")
		}
	})
}

// ----------------------------------------------------------------------------
// Import Error Tests
// ----------------------------------------------------------------------------

func TestImportError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &ImportError{
			Message: "cannot import module",
			Span:    syntax.Span{},
		}
		if err.Error() != "cannot import module" {
			t.Errorf("Error() = %q, want %q", err.Error(), "cannot import module")
		}
	})
}

func TestCyclicImportError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := &CyclicImportError{
			File: FileID{Path: "cyclic.typ"},
			Span: syntax.Span{},
		}
		msg := err.Error()
		if msg != "cyclic import detected: cyclic.typ" {
			t.Errorf("Error() = %q, want %q", msg, "cyclic import detected: cyclic.typ")
		}
	})
}

// ----------------------------------------------------------------------------
// Module Resolution Tests
// ----------------------------------------------------------------------------

func TestFunctionToModule(t *testing.T) {
	t.Run("nil function returns error", func(t *testing.T) {
		_, err := functionToModule(nil, syntax.Span{})
		if err == nil {
			t.Error("functionToModule(nil) should return error")
		}
	})

	t.Run("named function creates module with name", func(t *testing.T) {
		name := "myFunc"
		fn := &Func{
			Name: &name,
			Span: syntax.Span{},
			Repr: ClosureFunc{
				Closure: &Closure{
					Captured: NewScope(),
				},
			},
		}

		module, err := functionToModule(fn, syntax.Span{})
		if err != nil {
			t.Fatalf("functionToModule() unexpected error: %v", err)
		}
		if module.Name != "myFunc" {
			t.Errorf("module.Name = %q, want %q", module.Name, "myFunc")
		}
	})

	t.Run("anonymous function creates module named 'function'", func(t *testing.T) {
		fn := &Func{
			Name: nil,
			Span: syntax.Span{},
			Repr: ClosureFunc{
				Closure: &Closure{
					Captured: NewScope(),
				},
			},
		}

		module, err := functionToModule(fn, syntax.Span{})
		if err != nil {
			t.Fatalf("functionToModule() unexpected error: %v", err)
		}
		if module.Name != "function" {
			t.Errorf("module.Name = %q, want %q", module.Name, "function")
		}
	})
}

func TestTypeToModule(t *testing.T) {
	t.Run("creates module from type", func(t *testing.T) {
		module, err := typeToModule(TypeInt, syntax.Span{})
		if err != nil {
			t.Fatalf("typeToModule() unexpected error: %v", err)
		}
		if module.Name != "int" {
			t.Errorf("module.Name = %q, want %q", module.Name, "int")
		}
		if module.Scope == nil {
			t.Error("module.Scope should not be nil")
		}
	})
}

// ----------------------------------------------------------------------------
// Import All Exports Tests
// ----------------------------------------------------------------------------

func TestImportAllExports(t *testing.T) {
	t.Run("imports public bindings", func(t *testing.T) {
		// Create a module with some exports
		moduleScope := NewScope()
		moduleScope.Define("publicFunc", Int(1), syntax.Span{})
		moduleScope.Define("publicVar", Int(2), syntax.Span{})
		moduleScope.Define("_privateVar", Int(3), syntax.Span{})

		module := &Module{
			Name:  "test",
			Scope: moduleScope,
		}

		// Create a VM
		engine := NewEngine(nil)
		scopes := NewScopes(nil)
		vm := NewVm(engine, NewContext(), scopes, syntax.Span{})

		// Import all exports
		err := importAllExports(vm, module, syntax.Span{})
		if err != nil {
			t.Fatalf("importAllExports() unexpected error: %v", err)
		}

		// Check that public bindings were imported
		if b := vm.Get("publicFunc"); b == nil {
			t.Error("publicFunc should be imported")
		}
		if b := vm.Get("publicVar"); b == nil {
			t.Error("publicVar should be imported")
		}

		// Check that private binding was not imported
		if b := vm.Get("_privateVar"); b != nil {
			t.Error("_privateVar should not be imported")
		}
	})

	t.Run("handles nil scope", func(t *testing.T) {
		module := &Module{
			Name:  "test",
			Scope: nil,
		}

		engine := NewEngine(nil)
		scopes := NewScopes(nil)
		vm := NewVm(engine, NewContext(), scopes, syntax.Span{})

		err := importAllExports(vm, module, syntax.Span{})
		if err != nil {
			t.Errorf("importAllExports(nil scope) unexpected error: %v", err)
		}
	})
}

// ----------------------------------------------------------------------------
// Resolve Import Path Tests
// ----------------------------------------------------------------------------

func TestResolveImportPath(t *testing.T) {
	t.Run("simple path", func(t *testing.T) {
		moduleScope := NewScope()
		moduleScope.Define("value", Int(42), syntax.Span{})

		module := &Module{
			Name:  "test",
			Scope: moduleScope,
		}

		value, err := resolveImportPath(module, []string{"value"}, syntax.Span{})
		if err != nil {
			t.Fatalf("resolveImportPath() unexpected error: %v", err)
		}

		if i, ok := value.(IntValue); !ok || int64(i) != 42 {
			t.Errorf("value = %v, want IntValue(42)", value)
		}
	})

	t.Run("nested path through module", func(t *testing.T) {
		innerScope := NewScope()
		innerScope.Define("inner", Int(99), syntax.Span{})

		innerModule := &Module{
			Name:  "inner",
			Scope: innerScope,
		}

		outerScope := NewScope()
		outerScope.Define("nested", ModuleValue{Module: innerModule}, syntax.Span{})

		module := &Module{
			Name:  "outer",
			Scope: outerScope,
		}

		value, err := resolveImportPath(module, []string{"nested", "inner"}, syntax.Span{})
		if err != nil {
			t.Fatalf("resolveImportPath() unexpected error: %v", err)
		}

		if i, ok := value.(IntValue); !ok || int64(i) != 99 {
			t.Errorf("value = %v, want IntValue(99)", value)
		}
	})

	t.Run("not exported error", func(t *testing.T) {
		moduleScope := NewScope()

		module := &Module{
			Name:  "test",
			Scope: moduleScope,
		}

		_, err := resolveImportPath(module, []string{"nonexistent"}, syntax.Span{})
		if err == nil {
			t.Error("resolveImportPath(nonexistent) should return error")
		}
	})

	t.Run("empty path error", func(t *testing.T) {
		module := &Module{
			Name:  "test",
			Scope: NewScope(),
		}

		_, err := resolveImportPath(module, []string{}, syntax.Span{})
		if err == nil {
			t.Error("resolveImportPath(empty) should return error")
		}
	})
}

// ----------------------------------------------------------------------------
// Package Manifest Parsing Tests
// ----------------------------------------------------------------------------

func TestParsePackageManifest(t *testing.T) {
	t.Run("full manifest", func(t *testing.T) {
		data := []byte(`
[package]
name = "my-package"
version = "1.2.3"
entrypoint = "main.typ"
description = "A test package"
license = "MIT"
`)

		manifest, err := parsePackageManifest(data)
		if err != nil {
			t.Fatalf("parsePackageManifest() unexpected error: %v", err)
		}

		if manifest.Name != "my-package" {
			t.Errorf("Name = %q, want %q", manifest.Name, "my-package")
		}
		if manifest.Version != (Version{Major: 1, Minor: 2, Patch: 3}) {
			t.Errorf("Version = %v, want {1, 2, 3}", manifest.Version)
		}
		if manifest.EntryPoint != "main.typ" {
			t.Errorf("EntryPoint = %q, want %q", manifest.EntryPoint, "main.typ")
		}
	})

	t.Run("minimal manifest", func(t *testing.T) {
		data := []byte(`
name = "simple"
version = "0.1.0"
`)

		manifest, err := parsePackageManifest(data)
		if err != nil {
			t.Fatalf("parsePackageManifest() unexpected error: %v", err)
		}

		if manifest.Name != "simple" {
			t.Errorf("Name = %q, want %q", manifest.Name, "simple")
		}
		// Default entrypoint
		if manifest.EntryPoint != "lib.typ" {
			t.Errorf("EntryPoint = %q, want %q", manifest.EntryPoint, "lib.typ")
		}
	})

	t.Run("missing name error", func(t *testing.T) {
		data := []byte(`version = "1.0.0"`)

		_, err := parsePackageManifest(data)
		if err == nil {
			t.Error("parsePackageManifest(missing name) should return error")
		}
	})
}

// ----------------------------------------------------------------------------
// Package Manifest Validation Tests
// ----------------------------------------------------------------------------

func TestValidatePackageManifest(t *testing.T) {
	t.Run("matching spec", func(t *testing.T) {
		manifest := &PackageManifest{
			Name:    "mypackage",
			Version: Version{Major: 1, Minor: 2, Patch: 3},
		}
		spec := &PackageSpec{
			Name:    "mypackage",
			Version: Version{Major: 1, Minor: 2, Patch: 0},
		}

		err := validatePackageManifest(manifest, spec)
		if err != nil {
			t.Errorf("validatePackageManifest() unexpected error: %v", err)
		}
	})

	t.Run("name mismatch", func(t *testing.T) {
		manifest := &PackageManifest{
			Name:    "package-a",
			Version: Version{Major: 1, Minor: 0, Patch: 0},
		}
		spec := &PackageSpec{
			Name:    "package-b",
			Version: Version{Major: 1, Minor: 0, Patch: 0},
		}

		err := validatePackageManifest(manifest, spec)
		if err == nil {
			t.Error("validatePackageManifest(name mismatch) should return error")
		}
	})

	t.Run("major version mismatch", func(t *testing.T) {
		manifest := &PackageManifest{
			Name:    "mypackage",
			Version: Version{Major: 2, Minor: 0, Patch: 0},
		}
		spec := &PackageSpec{
			Name:    "mypackage",
			Version: Version{Major: 1, Minor: 0, Patch: 0},
		}

		err := validatePackageManifest(manifest, spec)
		if err == nil {
			t.Error("validatePackageManifest(major version mismatch) should return error")
		}
	})

	t.Run("minor version too old", func(t *testing.T) {
		manifest := &PackageManifest{
			Name:    "mypackage",
			Version: Version{Major: 1, Minor: 1, Patch: 0},
		}
		spec := &PackageSpec{
			Name:    "mypackage",
			Version: Version{Major: 1, Minor: 2, Patch: 0},
		}

		err := validatePackageManifest(manifest, spec)
		if err == nil {
			t.Error("validatePackageManifest(minor version too old) should return error")
		}
	})
}

// ----------------------------------------------------------------------------
// Scope.All() Tests
// ----------------------------------------------------------------------------

func TestScopeAll(t *testing.T) {
	t.Run("returns all bindings", func(t *testing.T) {
		scope := NewScope()
		scope.Define("a", Int(1), syntax.Span{})
		scope.Define("b", Int(2), syntax.Span{})
		scope.Define("c", Int(3), syntax.Span{})

		all := scope.All()
		if len(all) != 3 {
			t.Errorf("All() returned %d bindings, want 3", len(all))
		}

		for _, name := range []string{"a", "b", "c"} {
			if _, ok := all[name]; !ok {
				t.Errorf("All() missing binding %q", name)
			}
		}
	})

	t.Run("empty scope", func(t *testing.T) {
		scope := NewScope()
		all := scope.All()
		if len(all) != 0 {
			t.Errorf("All() on empty scope returned %d bindings, want 0", len(all))
		}
	})
}
