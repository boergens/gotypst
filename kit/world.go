// World implementations for Typst.
// Provides concrete implementations of the foundations.World interface.
//
// Translated from typst-kit/src/files.rs

package kit

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/boergens/gotypst/font"
	"github.com/boergens/gotypst/library/foundations"
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
	mainFile syntax.FileId

	// library is the standard library scope.
	library *foundations.Scope

	// fontBook manages loaded fonts.
	fontBook *font.FontBook

	// sourceCache caches parsed sources by file ID.
	sourceCache map[syntax.FileId]*syntax.Source

	// fileCache caches raw file bytes by file ID.
	fileCache map[syntax.FileId][]byte

	// pathCache maps file IDs to resolved absolute paths.
	pathCache map[syntax.FileId]string

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
	Resolve(spec *syntax.PackageSpec) (string, error)
}

// FileWorldOption configures a FileWorld.
type FileWorldOption func(*FileWorld)

// WithLibrary sets the standard library scope.
func WithLibrary(lib *foundations.Scope) FileWorldOption {
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

	// Create the main file ID
	relPath, err := filepath.Rel(absRoot, absMainPath)
	if err != nil {
		relPath = absMainPath
	}
	vpath, err := syntax.NewVirtualPath("/" + filepath.ToSlash(relPath))
	if err != nil {
		vpath, _ = syntax.NewVirtualPath("/main.typ")
	}
	rpath := syntax.NewRootedPath(syntax.ProjectRoot(), *vpath)
	mainFileId := rpath.Intern()

	w := &FileWorld{
		root:        absRoot,
		mainFile:    mainFileId,
		library:     foundations.NewScope(),
		sourceCache: make(map[syntax.FileId]*syntax.Source),
		fileCache:   make(map[syntax.FileId][]byte),
		pathCache:   make(map[syntax.FileId]string),
	}

	// Store the path mapping
	w.pathCache[mainFileId] = absMainPath

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
func (w *FileWorld) Library() *foundations.Scope {
	return w.library
}

// MainFile returns the main source file ID.
func (w *FileWorld) MainFile() syntax.FileId {
	return w.mainFile
}

// Source returns the parsed source content for a file.
//
// The source is cached after first access. If the file cannot be read
// or parsed, an error is returned.
func (w *FileWorld) Source(id syntax.FileId) (*syntax.Source, error) {
	// Check cache first
	w.mu.RLock()
	if src, ok := w.sourceCache[id]; ok {
		w.mu.RUnlock()
		return src, nil
	}
	w.mu.RUnlock()

	// Resolve the path
	path, err := w.resolvePath(id)
	if err != nil {
		return nil, err
	}

	// Read the file
	content, err := w.readFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read source file %s: %w", path, err)
	}

	// Parse the source
	src := syntax.NewSource(id, string(content))

	// Cache it
	w.mu.Lock()
	w.sourceCache[id] = src
	w.mu.Unlock()

	return src, nil
}

// File returns the raw bytes of a file.
//
// The file content is cached after first access.
func (w *FileWorld) File(id syntax.FileId) ([]byte, error) {
	// Check cache first
	w.mu.RLock()
	if data, ok := w.fileCache[id]; ok {
		w.mu.RUnlock()
		return data, nil
	}
	w.mu.RUnlock()

	// Resolve the path
	path, err := w.resolvePath(id)
	if err != nil {
		return nil, err
	}

	// Read the file
	data, err := w.readFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", path, err)
	}

	// Cache it
	w.mu.Lock()
	w.fileCache[id] = data
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
func (w *FileWorld) Today(offset *int) *foundations.Datetime {
	return foundations.Today(offset)
}

// resolvePath resolves a FileId to an absolute file system path.
func (w *FileWorld) resolvePath(id syntax.FileId) (string, error) {
	// Check path cache first
	w.mu.RLock()
	if path, ok := w.pathCache[id]; ok {
		w.mu.RUnlock()
		return path, nil
	}
	w.mu.RUnlock()

	// Get the rooted path from the file ID
	rpath := id.Get()
	if rpath == nil {
		return "", fmt.Errorf("cannot resolve file ID: no rooted path")
	}

	var absPath string
	root := rpath.Root()

	// Check if it's a package root
	if spec := rpath.Package(); spec != nil {
		// Package file - resolve through package resolver
		if w.packageResolver == nil {
			return "", fmt.Errorf("package imports not supported: no package resolver configured")
		}

		pkgRoot, err := w.packageResolver.Resolve(spec)
		if err != nil {
			return "", fmt.Errorf("cannot resolve package %s: %w", spec.Name, err)
		}

		vpath := rpath.VPath()
		absPath = vpath.Realize(pkgRoot)
	} else if root == syntax.ProjectRoot() {
		// Project file - resolve relative to root
		vpath := rpath.VPath()
		absPath = vpath.Realize(w.root)
	} else {
		return "", fmt.Errorf("unknown virtual root type")
	}

	// Cache the resolved path
	w.mu.Lock()
	w.pathCache[id] = absPath
	w.mu.Unlock()

	return absPath, nil
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

// ClearCache clears all cached sources and files.
func (w *FileWorld) ClearCache() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.sourceCache = make(map[syntax.FileId]*syntax.Source)
	w.fileCache = make(map[syntax.FileId][]byte)
}

// ClearSourceCache clears only the parsed source cache.
// This is useful when source files have been modified.
func (w *FileWorld) ClearSourceCache() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.sourceCache = make(map[syntax.FileId]*syntax.Source)
}

// InvalidateFile removes a specific file from the caches.
func (w *FileWorld) InvalidateFile(id syntax.FileId) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.sourceCache, id)
	delete(w.fileCache, id)
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
func (r *LocalPackageResolver) Resolve(spec *syntax.PackageSpec) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("nil package specification")
	}

	// Build the package path: <root>/<namespace>/<name>/<version>/
	versionStr := spec.Version.String()
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
	Spec *syntax.PackageSpec
}

func (e *PackageNotFoundError) Error() string {
	if e.Spec == nil {
		return "package not found: nil spec"
	}
	return fmt.Sprintf("package not found: @%s/%s:%s",
		e.Spec.Namespace, e.Spec.Name, e.Spec.Version.String())
}
