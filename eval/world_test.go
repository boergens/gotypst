package eval

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boergens/gotypst/syntax"
)

func TestNewFileWorld(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create a main file
	mainFile := filepath.Join(tmpDir, "main.typ")
	if err := os.WriteFile(mainFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create main file: %v", err)
	}

	// Test successful creation
	world, err := NewFileWorld(tmpDir, "main.typ")
	if err != nil {
		t.Fatalf("NewFileWorld failed: %v", err)
	}

	if world.Root() != tmpDir {
		t.Errorf("expected root %s, got %s", tmpDir, world.Root())
	}

	mainID := world.MainFile()
	if mainID.Path != mainFile {
		t.Errorf("expected main file path %s, got %s", mainFile, mainID.Path)
	}

	if mainID.Package != nil {
		t.Error("expected main file package to be nil")
	}
}

func TestNewFileWorld_Errors(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		mainPath string
		wantErr  string
	}{
		{
			name:     "non-existent root",
			root:     "/nonexistent/path/that/does/not/exist",
			mainPath: "main.typ",
			wantErr:  "project root does not exist",
		},
		{
			name:     "non-existent main file",
			root:     os.TempDir(),
			mainPath: "nonexistent_main_file_12345.typ",
			wantErr:  "main file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFileWorld(tt.root, tt.mainPath)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestFileWorld_Library(t *testing.T) {
	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte(""), 0644)

	// Test default (empty) library
	world, _ := NewFileWorld(tmpDir, "main.typ")
	lib := world.Library()
	if lib == nil {
		t.Fatal("Library() returned nil")
	}
	if lib.Len() != 0 {
		t.Errorf("expected empty library, got %d bindings", lib.Len())
	}

	// Test with custom library
	customLib := NewScope()
	customLib.Define("test", IntValue(42), syntax.Span{})
	world2, _ := NewFileWorld(tmpDir, "main.typ", WithLibrary(customLib))
	lib2 := world2.Library()
	if lib2.Len() != 1 {
		t.Errorf("expected 1 binding in custom library, got %d", lib2.Len())
	}
}

func TestFileWorld_Source(t *testing.T) {
	tmpDir := t.TempDir()

	// Create main file with Typst content
	mainFile := filepath.Join(tmpDir, "main.typ")
	content := "= Hello World\n\nThis is a test."
	os.WriteFile(mainFile, []byte(content), 0644)

	world, _ := NewFileWorld(tmpDir, "main.typ")

	// Test loading main file source
	mainID := world.MainFile()
	src, err := world.Source(mainID)
	if err != nil {
		t.Fatalf("Source() failed: %v", err)
	}

	if src.Text() != content {
		t.Errorf("expected text %q, got %q", content, src.Text())
	}

	// Test caching - second call should return same instance
	src2, _ := world.Source(mainID)
	if src != src2 {
		t.Error("expected cached source to be same instance")
	}
}

func TestFileWorld_Source_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte(""), 0644)

	world, _ := NewFileWorld(tmpDir, "main.typ")

	// Try to load non-existent file
	_, err := world.Source(FileID{Path: filepath.Join(tmpDir, "nonexistent.typ")})
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestFileWorld_File(t *testing.T) {
	tmpDir := t.TempDir()

	// Create main file
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte("main"), 0644)

	// Create a data file
	dataFile := filepath.Join(tmpDir, "data.json")
	dataContent := []byte(`{"key": "value"}`)
	os.WriteFile(dataFile, dataContent, 0644)

	world, _ := NewFileWorld(tmpDir, "main.typ")

	// Test loading data file
	data, err := world.File(FileID{Path: dataFile})
	if err != nil {
		t.Fatalf("File() failed: %v", err)
	}

	if string(data) != string(dataContent) {
		t.Errorf("expected content %q, got %q", dataContent, data)
	}

	// Test caching
	data2, _ := world.File(FileID{Path: dataFile})
	if string(data) != string(data2) {
		t.Error("cached file content mismatch")
	}
}

func TestFileWorld_File_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte(""), 0644)

	world, _ := NewFileWorld(tmpDir, "main.typ")

	_, err := world.File(FileID{Path: filepath.Join(tmpDir, "nonexistent.json")})
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// Should be FileNotFoundError (possibly wrapped)
	var fnfErr *FileNotFoundError
	if !errors.As(err, &fnfErr) {
		t.Errorf("expected FileNotFoundError, got %T: %v", err, err)
	}
}

func TestFileWorld_Today(t *testing.T) {
	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte(""), 0644)

	world, _ := NewFileWorld(tmpDir, "main.typ")

	// Test today without offset
	date := world.Today(nil)
	now := time.Now()

	if date.Year != now.Year() {
		t.Errorf("expected year %d, got %d", now.Year(), date.Year)
	}
	if date.Month != int(now.Month()) {
		t.Errorf("expected month %d, got %d", now.Month(), date.Month)
	}
	if date.Day != now.Day() {
		t.Errorf("expected day %d, got %d", now.Day(), date.Day)
	}

	// Test with offset (just verify it doesn't error)
	offset := 0
	dateWithOffset := world.Today(&offset)
	if dateWithOffset.Year == 0 {
		t.Error("Today with offset returned invalid date")
	}
}

func TestFileWorld_ClearCache(t *testing.T) {
	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte("content"), 0644)

	world, _ := NewFileWorld(tmpDir, "main.typ")

	// Load and cache the file
	mainID := world.MainFile()
	src1, _ := world.Source(mainID)

	// Clear cache
	world.ClearCache()

	// Load again - should parse fresh
	src2, _ := world.Source(mainID)

	// Should be different instances after cache clear
	if src1 == src2 {
		t.Error("expected different source instances after cache clear")
	}
}

func TestFileWorld_InvalidateFile(t *testing.T) {
	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte("original"), 0644)

	world, _ := NewFileWorld(tmpDir, "main.typ")
	mainID := world.MainFile()

	// Load and cache
	src1, _ := world.Source(mainID)
	file1, _ := world.File(mainID)

	// Invalidate
	world.InvalidateFile(mainFile)

	// Update file content
	os.WriteFile(mainFile, []byte("updated"), 0644)

	// Load again
	src2, _ := world.Source(mainID)
	file2, _ := world.File(mainID)

	if src1 == src2 {
		t.Error("expected different source after invalidation")
	}
	if string(file1) == string(file2) {
		t.Error("expected different file content after invalidation")
	}
}

func TestFileWorld_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory structure
	subDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(subDir, 0755)

	mainFile := filepath.Join(subDir, "main.typ")
	os.WriteFile(mainFile, []byte("main"), 0644)

	// Create world with subdirectory main file using relative path
	world, err := NewFileWorld(tmpDir, "src/main.typ")
	if err != nil {
		t.Fatalf("NewFileWorld failed: %v", err)
	}

	mainID := world.MainFile()
	if mainID.Path != mainFile {
		t.Errorf("expected main path %s, got %s", mainFile, mainID.Path)
	}

	// Test loading relative file
	otherFile := filepath.Join(subDir, "other.typ")
	os.WriteFile(otherFile, []byte("other"), 0644)

	// Load using relative path from root
	data, err := world.File(FileID{Path: "src/other.typ"})
	if err != nil {
		t.Fatalf("File() with relative path failed: %v", err)
	}
	if string(data) != "other" {
		t.Errorf("expected 'other', got %q", data)
	}
}

func TestLocalPackageResolver(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package directory structure
	pkgDir := filepath.Join(tmpDir, "preview", "example", "1.0.0")
	os.MkdirAll(pkgDir, 0755)

	// Create a file in the package
	os.WriteFile(filepath.Join(pkgDir, "lib.typ"), []byte("// package"), 0644)

	resolver := NewLocalPackageResolver(tmpDir)

	// Test resolving existing package
	spec := &PackageSpec{
		Namespace: "preview",
		Name:      "example",
		Version:   Version{Major: 1, Minor: 0, Patch: 0},
	}

	path, err := resolver.Resolve(spec)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if path != pkgDir {
		t.Errorf("expected path %s, got %s", pkgDir, path)
	}

	// Test resolving non-existent package
	spec2 := &PackageSpec{
		Namespace: "preview",
		Name:      "nonexistent",
		Version:   Version{Major: 1, Minor: 0, Patch: 0},
	}

	_, err = resolver.Resolve(spec2)
	if err == nil {
		t.Fatal("expected error for non-existent package")
	}

	if _, ok := err.(*PackageNotFoundError); !ok {
		t.Errorf("expected PackageNotFoundError, got %T", err)
	}
}

func TestFileWorld_WithPackageResolver(t *testing.T) {
	tmpDir := t.TempDir()

	// Create main project
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte("main"), 0644)

	// Create package structure
	pkgRoot := filepath.Join(tmpDir, "packages")
	pkgDir := filepath.Join(pkgRoot, "preview", "test", "1.0.0")
	os.MkdirAll(pkgDir, 0755)
	os.WriteFile(filepath.Join(pkgDir, "lib.typ"), []byte("package code"), 0644)

	resolver := NewLocalPackageResolver(pkgRoot)
	world, _ := NewFileWorld(tmpDir, "main.typ", WithPackageResolver(resolver))

	// Test loading package file
	pkgFileID := FileID{
		Package: &PackageSpec{
			Namespace: "preview",
			Name:      "test",
			Version:   Version{Major: 1, Minor: 0, Patch: 0},
		},
		Path: "lib.typ",
	}

	data, err := world.File(pkgFileID)
	if err != nil {
		t.Fatalf("File() for package failed: %v", err)
	}

	if string(data) != "package code" {
		t.Errorf("expected 'package code', got %q", data)
	}
}

func TestFileWorld_PackageWithoutResolver(t *testing.T) {
	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.typ")
	os.WriteFile(mainFile, []byte(""), 0644)

	world, _ := NewFileWorld(tmpDir, "main.typ")

	// Try to load package file without resolver
	pkgFileID := FileID{
		Package: &PackageSpec{
			Namespace: "preview",
			Name:      "test",
			Version:   Version{Major: 1, Minor: 0, Patch: 0},
		},
		Path: "lib.typ",
	}

	_, err := world.File(pkgFileID)
	if err == nil {
		t.Fatal("expected error when loading package without resolver")
	}

	if !contains(err.Error(), "package imports not supported") {
		t.Errorf("expected 'package imports not supported' error, got: %v", err)
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
