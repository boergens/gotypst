package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/boergens/gotypst/font"
	"github.com/boergens/gotypst/syntax"
)

// FileWorld is a concrete implementation of the World interface that provides
// filesystem access for Typst compilation.
//
// FileWorld handles:
//   - Reading source files from the filesystem
//   - Caching parsed sources and raw file bytes
//   - Resolving file paths relative to a project root
//   - Providing the current date for date functions
//   - Loading and managing fonts
type FileWorld struct {
	// root is the project root directory (absolute path).
	root string

	// mainFile is the main source file being compiled.
	mainFile FileID

	// library is the standard library scope.
	library *Scope

	// fontBook manages loaded fonts.
	fontBook *font.FontBook

	// sourceCache caches parsed sources by file path.
	sourceCache map[string]*syntax.Source

	// fileCache caches raw file bytes by file path.
	fileCache map[string][]byte

	// mu protects the caches.
	mu sync.RWMutex

	// packageResolver resolves package specifications to file system paths.
	// If nil, package imports are not supported.
	packageResolver PackageResolver
}

// PackageResolver resolves package specifications to file system paths.
type PackageResolver interface {
	// Resolve returns the root directory for a package specification.
	// Returns an error if the package cannot be found.
	Resolve(spec *PackageSpec) (string, error)
}

// FileWorldOption configures a FileWorld.
type FileWorldOption func(*FileWorld)

// WithLibrary sets the standard library scope.
func WithLibrary(lib *Scope) FileWorldOption {
	return func(w *FileWorld) {
		w.library = lib
	}
}

// WithPackageResolver sets the package resolver.
func WithPackageResolver(resolver PackageResolver) FileWorldOption {
	return func(w *FileWorld) {
		w.packageResolver = resolver
	}
}

// WithFontBook sets the font book for the world.
// If not set, system fonts will be loaded automatically.
func WithFontBook(book *font.FontBook) FileWorldOption {
	return func(w *FileWorld) {
		w.fontBook = book
	}
}

// WithFontDirs loads fonts from the specified directories.
func WithFontDirs(dirs ...string) FileWorldOption {
	return func(w *FileWorld) {
		fonts, _ := font.DiscoverFonts(dirs)
		if w.fontBook == nil {
			w.fontBook = font.NewFontBook()
		}
		w.fontBook.Add(fonts...)
	}
}

// NewFileWorld creates a new FileWorld for the given main file.
//
// The root parameter is the project root directory. All relative paths
// in the project are resolved relative to this directory.
//
// The mainPath parameter is the path to the main source file, relative
// to the root directory.
func NewFileWorld(root string, mainPath string, opts ...FileWorldOption) (*FileWorld, error) {
	// Ensure root is absolute
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve project root: %w", err)
	}

	// Verify root exists
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("project root does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project root is not a directory: %s", absRoot)
	}

	// Resolve main file path
	var absMainPath string
	if filepath.IsAbs(mainPath) {
		absMainPath = mainPath
	} else {
		absMainPath = filepath.Join(absRoot, mainPath)
	}

	// Verify main file exists
	if _, err := os.Stat(absMainPath); err != nil {
		return nil, fmt.Errorf("main file does not exist: %w", err)
	}

	w := &FileWorld{
		root: absRoot,
		mainFile: FileID{
			Package: nil,
			Path:    absMainPath,
		},
		library:     NewScope(),
		sourceCache: make(map[string]*syntax.Source),
		fileCache:   make(map[string][]byte),
	}

	for _, opt := range opts {
		opt(w)
	}

	// Load system fonts if no font book was provided
	if w.fontBook == nil {
		w.fontBook, _ = font.SystemFontBook()
		if w.fontBook == nil {
			// Fallback to empty font book
			w.fontBook = font.NewFontBook()
		}
	}

	return w, nil
}

// Library returns the standard library scope.
func (w *FileWorld) Library() *Scope {
	return w.library
}

// MainFile returns the main source file ID.
func (w *FileWorld) MainFile() FileID {
	return w.mainFile
}

// Source returns the parsed source content for a file.
//
// The source is cached after first access. If the file cannot be read
// or parsed, an error is returned.
func (w *FileWorld) Source(id FileID) (*syntax.Source, error) {
	path, err := w.resolvePath(id)
	if err != nil {
		return nil, err
	}

	// Check cache first
	w.mu.RLock()
	if src, ok := w.sourceCache[path]; ok {
		w.mu.RUnlock()
		return src, nil
	}
	w.mu.RUnlock()

	// Read the file
	content, err := w.readFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read source file %s: %w", path, err)
	}

	// Create file ID for the syntax package
	fileId := w.createSyntaxFileId(id, path)

	// Parse the source
	src := syntax.NewSource(fileId, string(content))

	// Cache it
	w.mu.Lock()
	w.sourceCache[path] = src
	w.mu.Unlock()

	return src, nil
}

// File returns the raw bytes of a file.
//
// The file content is cached after first access.
func (w *FileWorld) File(id FileID) ([]byte, error) {
	path, err := w.resolvePath(id)
	if err != nil {
		return nil, err
	}

	// Check cache first
	w.mu.RLock()
	if data, ok := w.fileCache[path]; ok {
		w.mu.RUnlock()
		return data, nil
	}
	w.mu.RUnlock()

	// Read the file
	data, err := w.readFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", path, err)
	}

	// Cache it
	w.mu.Lock()
	w.fileCache[path] = data
	w.mu.Unlock()

	return data, nil
}

// Font returns the font at the given index.
// Returns an error if the index is out of bounds.
func (w *FileWorld) Font(index int) (*font.Font, error) {
	f := w.fontBook.Font(index)
	if f == nil {
		return nil, fmt.Errorf("font index out of bounds: %d", index)
	}
	return f, nil
}

// FontCount returns the number of available fonts.
func (w *FileWorld) FontCount() int {
	return w.fontBook.Len()
}

// FontBook returns the font book for direct font access.
func (w *FileWorld) FontBook() *font.FontBook {
	return w.fontBook
}

// Today returns the current date, optionally adjusted by an offset.
//
// If offset is not nil, it adjusts the UTC offset used to determine
// the current date. For example, offset=8 would use UTC+8.
func (w *FileWorld) Today(offset *int) Date {
	now := time.Now()

	if offset != nil {
		// Apply UTC offset
		loc := time.FixedZone("", *offset*3600)
		now = now.In(loc)
	}

	return Date{
		Year:  now.Year(),
		Month: int(now.Month()),
		Day:   now.Day(),
	}
}

// resolvePath resolves a FileID to an absolute file system path.
func (w *FileWorld) resolvePath(id FileID) (string, error) {
	if id.Package != nil {
		// Package file - resolve through package resolver
		if w.packageResolver == nil {
			return "", fmt.Errorf("package imports not supported: no package resolver configured")
		}

		pkgRoot, err := w.packageResolver.Resolve(id.Package)
		if err != nil {
			return "", fmt.Errorf("cannot resolve package %s: %w", id.Package.Name, err)
		}

		return filepath.Join(pkgRoot, id.Path), nil
	}

	// Project file - resolve relative to root or use absolute path
	if filepath.IsAbs(id.Path) {
		return id.Path, nil
	}

	return filepath.Join(w.root, id.Path), nil
}

// readFile reads a file from the filesystem.
func (w *FileWorld) readFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &FileNotFoundError{Path: path}
		}
		return nil, err
	}
	return data, nil
}

// createSyntaxFileId creates a syntax.FileId from an eval.FileID.
func (w *FileWorld) createSyntaxFileId(id FileID, resolvedPath string) syntax.FileId {
	// Create a virtual path from the resolved path
	relPath, err := filepath.Rel(w.root, resolvedPath)
	if err != nil {
		// Fall back to absolute path
		relPath = resolvedPath
	}

	vpath, err := syntax.NewVirtualPath("/" + filepath.ToSlash(relPath))
	if err != nil {
		// Fall back to a simple path
		vpath, _ = syntax.NewVirtualPath("/unknown")
	}

	var root syntax.VirtualRoot
	if id.Package != nil {
		root = syntax.PackageRoot(syntax.PackageSpec{
			Namespace: id.Package.Namespace,
			Name:      id.Package.Name,
			Version: syntax.PackageVersion{
				Major: uint32(id.Package.Version.Major),
				Minor: uint32(id.Package.Version.Minor),
				Patch: uint32(id.Package.Version.Patch),
			},
		})
	} else {
		root = syntax.ProjectRoot()
	}

	rpath := syntax.NewRootedPath(root, *vpath)
	return rpath.Intern()
}

// ClearCache clears all cached sources and files.
func (w *FileWorld) ClearCache() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.sourceCache = make(map[string]*syntax.Source)
	w.fileCache = make(map[string][]byte)
}

// ClearSourceCache clears only the parsed source cache.
// This is useful when source files have been modified.
func (w *FileWorld) ClearSourceCache() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.sourceCache = make(map[string]*syntax.Source)
}

// InvalidateFile removes a specific file from the caches.
func (w *FileWorld) InvalidateFile(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.sourceCache, path)
	delete(w.fileCache, path)
}

// Root returns the project root directory.
func (w *FileWorld) Root() string {
	return w.root
}

// ----------------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------------

// FileNotFoundError is returned when a file cannot be found.
type FileNotFoundError struct {
	Path string
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("file not found: %s", e.Path)
}

// ----------------------------------------------------------------------------
// Simple Package Resolver
// ----------------------------------------------------------------------------

// LocalPackageResolver resolves packages from a local directory structure.
//
// Packages are expected to be organized as:
//
//	<root>/<namespace>/<name>/<version>/
//
// For example: ~/.cache/typst/packages/preview/example/1.0.0/
type LocalPackageResolver struct {
	// root is the root directory containing all packages.
	root string
}

// NewLocalPackageResolver creates a new local package resolver.
func NewLocalPackageResolver(root string) *LocalPackageResolver {
	return &LocalPackageResolver{root: root}
}

// Resolve returns the root directory for a package specification.
func (r *LocalPackageResolver) Resolve(spec *PackageSpec) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("nil package specification")
	}

	// Build the package path: <root>/<namespace>/<name>/<version>/
	versionStr := fmt.Sprintf("%d.%d.%d", spec.Version.Major, spec.Version.Minor, spec.Version.Patch)
	pkgPath := filepath.Join(r.root, spec.Namespace, spec.Name, versionStr)

	// Verify the package directory exists
	info, err := os.Stat(pkgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", &PackageNotFoundError{Spec: spec}
		}
		return "", fmt.Errorf("cannot access package directory: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("package path is not a directory: %s", pkgPath)
	}

	return pkgPath, nil
}

// PackageNotFoundError is returned when a package cannot be found.
type PackageNotFoundError struct {
	Spec *PackageSpec
}

func (e *PackageNotFoundError) Error() string {
	if e.Spec == nil {
		return "package not found: nil spec"
	}
	return fmt.Sprintf("package not found: @%s/%s:%d.%d.%d",
		e.Spec.Namespace, e.Spec.Name,
		e.Spec.Version.Major, e.Spec.Version.Minor, e.Spec.Version.Patch)
}
