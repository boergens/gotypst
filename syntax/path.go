// Package syntax provides virtual, cross-platform reproducible path handling.
package syntax

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// RootedPath represents a path in a specific virtual file system root.
// This identifies a location in a project or package.
type RootedPath struct {
	root  VirtualRoot
	vpath VirtualPath
}

// NewRootedPath creates a rooted path from a root and a virtual path within the root.
func NewRootedPath(root VirtualRoot, vpath VirtualPath) *RootedPath {
	return &RootedPath{root: root, vpath: vpath}
}

// Intern turns this path into a FileId.
func (p *RootedPath) Intern() FileId {
	return NewFileId(*p)
}

// Root returns the root this path resides in.
func (p *RootedPath) Root() VirtualRoot {
	return p.root
}

// Package returns the package the path resides in, if any.
// Deprecated: use Root instead.
func (p *RootedPath) Package() *PackageSpec {
	if pkg, ok := p.root.(*VirtualRootPackage); ok {
		return &pkg.Spec
	}
	return nil
}

// VPath returns the absolute and normalized path to the file
// within the project or package.
func (p *RootedPath) VPath() *VirtualPath {
	return &p.vpath
}

// Map maps the virtual path while retaining the root.
func (p *RootedPath) Map(f func(*VirtualPath) VirtualPath) *RootedPath {
	newVPath := f(&p.vpath)
	return NewRootedPath(p.root, newVPath)
}

// String returns a debug string representation of the RootedPath.
func (p *RootedPath) String() string {
	vpath := p.VPath()
	switch r := p.root.(type) {
	case *VirtualRootProject:
		return vpath.String()
	case *VirtualRootPackage:
		return fmt.Sprintf("%s%s", r.Spec.String(), vpath.String())
	default:
		return vpath.String()
	}
}

// VirtualRoot represents the root of a virtual file system.
type VirtualRoot interface {
	virtualRoot()
}

// VirtualRootProject represents the canonical root of the Typst project.
// This is what TYPST_ROOT defines.
type VirtualRootProject struct{}

func (*VirtualRootProject) virtualRoot() {}

// VirtualRootPackage represents a root in a package.
type VirtualRootPackage struct {
	Spec PackageSpec
}

func (*VirtualRootPackage) virtualRoot() {}

// ProjectRoot returns a VirtualRoot representing the project root.
func ProjectRoot() VirtualRoot {
	return &VirtualRootProject{}
}

// PackageRoot returns a VirtualRoot representing a package root.
func PackageRoot(spec PackageSpec) VirtualRoot {
	return &VirtualRootPackage{Spec: spec}
}

// Global interner for rooted paths.
var interner = struct {
	sync.RWMutex
	toId   map[string]FileId // Use string key for map comparison
	fromId []*RootedPath
}{
	toId:   make(map[string]FileId),
	fromId: make([]*RootedPath, 0),
}

// FileId is an interned version of RootedPath.
// This type is globally interned and thus cheap to copy, compare, and hash.
type FileId struct {
	id uint16 // Non-zero ID
}

// NewFileId creates a new interned file specification.
// This is the same as RootedPath.Intern().
func NewFileId(path RootedPath) FileId {
	interner.Lock()
	defer interner.Unlock()

	key := pathKey(&path)
	if id, ok := interner.toId[key]; ok {
		return id
	}

	// Create new entry
	num := len(interner.fromId) + 1
	if num > 65535 {
		panic("out of file ids")
	}

	id := FileId{id: uint16(num)}
	pathCopy := path // Make a copy
	interner.toId[key] = id
	interner.fromId = append(interner.fromId, &pathCopy)
	return id
}

// UniqueFileId creates a new unique ("fake") file specification,
// which is not accessible by path.
//
// Caution: the ID returned by this method is the *only* identifier of the
// file, constructing a file ID with a path will *not* reuse the ID even if
// the path is the same. This method should only be used for generating
// "virtual" file ids such as content read from stdin.
func UniqueFileId(path RootedPath) FileId {
	interner.Lock()
	defer interner.Unlock()

	num := len(interner.fromId) + 1
	if num > 65535 {
		panic("out of file ids")
	}

	id := FileId{id: uint16(num)}
	pathCopy := path
	interner.fromId = append(interner.fromId, &pathCopy)
	return id
}

// FromRaw constructs a FileId from a raw number.
// Should only be used with numbers retrieved via IntoRaw.
// Misuse may result in panics, but no unsafety.
func FileIdFromRaw(v uint16) FileId {
	if v == 0 {
		panic("FileId cannot be zero")
	}
	return FileId{id: v}
}

// IntoRaw extracts the raw underlying number.
func (f FileId) IntoRaw() uint16 {
	return f.id
}

// Get returns the static, interned rooted path.
func (f FileId) Get() *RootedPath {
	interner.RLock()
	defer interner.RUnlock()
	return interner.fromId[f.id-1]
}

// String returns a debug string representation of the FileId.
func (f FileId) String() string {
	return f.Get().String()
}

// pathKey generates a unique string key for a RootedPath for map lookup.
func pathKey(p *RootedPath) string {
	var rootStr string
	switch r := p.root.(type) {
	case *VirtualRootProject:
		rootStr = "project:"
	case *VirtualRootPackage:
		rootStr = fmt.Sprintf("package:%s:", r.Spec.String())
	}
	return rootStr + p.vpath.GetWithSlash()
}

// PathError represents an error that can occur on construction or modification of a VirtualPath.
type PathError int

const (
	// PathErrorEscapes indicates the constructed or modified path would escape the root.
	// This would happen, for instance, when trying to join ".." to the path "/".
	// Note that a path might still escape through symlinks.
	PathErrorEscapes PathError = iota
	// PathErrorBackslash indicates the path contains a backslash.
	// This is not allowed as it leads to cross-platform compatibility hazards
	// (since Windows uses backslashes as a path separator).
	PathErrorBackslash
)

func (e PathError) Error() string {
	switch e {
	case PathErrorEscapes:
		return "path would escape root"
	case PathErrorBackslash:
		return "path contains backslash"
	default:
		return "unknown path error"
	}
}

// VirtualizeError represents an error that can occur in VirtualPath.Virtualize.
type VirtualizeError struct {
	PathErr *PathError
	Invalid string
	Utf8    bool
}

func (e *VirtualizeError) Error() string {
	if e.PathErr != nil {
		return e.PathErr.Error()
	}
	if e.Invalid != "" {
		return fmt.Sprintf("invalid path component: %s", e.Invalid)
	}
	if e.Utf8 {
		return "path contains non-UTF-8 bytes"
	}
	return "unknown virtualize error"
}

// VirtualPath represents a path in a virtual file system.
type VirtualPath struct {
	segments segments
}

// NewVirtualPath creates a new virtual path.
func NewVirtualPath(path string) (*VirtualPath, error) {
	segs, err := normalizeSegments(components(path))
	if err != nil {
		return nil, err
	}
	return &VirtualPath{segments: segs}, nil
}

// Virtualize creates a virtual path from a real path and a real root.
//
// Returns an error if the file path is not contained in the root (i.e. if
// rootPath is not a lexical prefix of path). No file system operations are performed.
//
// This is the single function that translates from a real path to a virtual path.
// Its counterpart is VirtualPath.Realize.
func Virtualize(rootPath, path string) (*VirtualPath, error) {
	// Check if path has rootPath as prefix
	relPath, err := filepath.Rel(rootPath, path)
	if err != nil {
		pe := PathErrorEscapes
		return nil, &VirtualizeError{PathErr: &pe}
	}

	// Check for escape
	if strings.HasPrefix(relPath, "..") {
		pe := PathErrorEscapes
		return nil, &VirtualizeError{PathErr: &pe}
	}

	var segs segments
	segs.data = "/"

	// Split the relative path and add each component
	if relPath != "." && relPath != "" {
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		for _, part := range parts {
			if part == "" || part == "." {
				continue
			}
			if part == ".." {
				if !segs.pop() {
					pe := PathErrorEscapes
					return nil, &VirtualizeError{PathErr: &pe}
				}
				continue
			}
			// Check for backslash
			if strings.Contains(part, "\\") {
				pe := PathErrorBackslash
				return nil, &VirtualizeError{PathErr: &pe}
			}
			seg, err := newSegment(part)
			if err != nil {
				return nil, &VirtualizeError{Invalid: part}
			}
			segs.push(seg)
		}
	}

	return &VirtualPath{segments: segs}, nil
}

// Realize turns the virtual path into an actual file system path
// (where the project or package resides). You need to provide the appropriate
// root path, relative to which this path will be resolved.
//
// This can be used in the implementations of World.Source and World.File.
//
// This is the single function that translates from a virtual path to a real path.
// Its counterpart is Virtualize.
func (v *VirtualPath) Realize(root string) string {
	out := root
	for _, seg := range v.segments.iter() {
		out = filepath.Join(out, seg)
	}
	return out
}

// GetWithSlash returns the path with a leading slash.
func (v *VirtualPath) GetWithSlash() string {
	return v.segments.getWithSlash()
}

// GetWithoutSlash returns the path without a leading slash.
func (v *VirtualPath) GetWithoutSlash() string {
	return v.segments.getWithoutSlash()
}

// FileName returns the file name portion of the path.
func (v *VirtualPath) FileName() string {
	last := v.segments.last()
	if last == "" {
		return ""
	}
	return last
}

// FileStem returns the file name portion of the path without the extension.
func (v *VirtualPath) FileStem() string {
	last := v.segments.last()
	if last == "" {
		return ""
	}
	before, after := splitDot(last)
	if before != "" {
		return before
	}
	return after
}

// Extension returns the file extension of the path.
func (v *VirtualPath) Extension() string {
	last := v.segments.last()
	if last == "" {
		return ""
	}
	before, after := splitDot(last)
	if before != "" && after != "" {
		return after
	}
	return ""
}

// WithExtension returns a modified path with an adjusted extension.
// Panics if the resulting path segment would be invalid, e.g. because the
// extension contains a forward or backslash.
func (v *VirtualPath) WithExtension(ext string) *VirtualPath {
	stem := v.FileStem()
	if stem == "" {
		return v
	}

	buf := fmt.Sprintf("%s.%s", stem, ext)
	seg, err := newSegment(buf)
	if err != nil {
		panic("extension is invalid")
	}

	newSegs := v.segments.clone()
	newSegs.pop()
	newSegs.push(seg)
	return &VirtualPath{segments: newSegs}
}

// Parent returns the path with its final component removed.
// Returns nil if the path is already at the root.
func (v *VirtualPath) Parent() *VirtualPath {
	newSegs := v.segments.clone()
	if !newSegs.pop() {
		return nil
	}
	return &VirtualPath{segments: newSegs}
}

// Join joins the given path to this path.
func (v *VirtualPath) Join(path string) (*VirtualPath, error) {
	// Start with current segments as component iterator
	var comps []component
	for _, seg := range v.segments.iter() {
		comps = append(comps, component{typ: componentNormal, value: seg})
	}

	// Add new path components
	for comp := range components(path) {
		comps = append(comps, comp)
	}

	segs, err := normalizeComponents(comps)
	if err != nil {
		return nil, err
	}

	return &VirtualPath{segments: segs}, nil
}

// RelativeFrom tries to express this path as a relative path from the given base path.
func (v *VirtualPath) RelativeFrom(base *VirtualPath) string {
	// Adapted from rustc's path_relative_from function (MIT).
	// Copyright 2012-2015 The Rust Project Developers.
	iterA := v.segments.iter()
	iterB := base.segments.iter()

	var buf []string
	idxA := 0
	idxB := 0

	for {
		var a, b string
		hasA := idxA < len(iterA)
		hasB := idxB < len(iterB)

		if hasA {
			a = iterA[idxA]
		}
		if hasB {
			b = iterB[idxB]
		}

		if !hasA && !hasB {
			break
		}

		if hasA && !hasB {
			buf = append(buf, a)
			idxA++
			for idxA < len(iterA) {
				buf = append(buf, iterA[idxA])
				idxA++
			}
			break
		}

		if !hasA && hasB {
			buf = append(buf, "..")
			idxB++
			continue
		}

		if len(buf) == 0 && a == b {
			idxA++
			idxB++
			continue
		}

		// Count remaining items in iterB
		remaining := len(iterB) - idxB
		for i := 0; i < remaining; i++ {
			buf = append(buf, "..")
		}
		buf = append(buf, a)
		idxA++
		for idxA < len(iterA) {
			buf = append(buf, iterA[idxA])
			idxA++
		}
		break
	}

	return strings.Join(buf, "/")
}

// String returns a debug string representation of the VirtualPath.
func (v *VirtualPath) String() string {
	return v.GetWithSlash()
}

// component represents a component in a virtual path.
type component struct {
	typ   componentType
	value string
}

type componentType int

const (
	componentRoot componentType = iota
	componentCurrent
	componentParent
	componentNormal
)

const separator = '/'
const currentDir = "."
const parentDir = ".."

// components splits a user-supplied path into its constituent parts.
func components(path string) func(yield func(component) bool) {
	return func(yield func(component) bool) {
		parts := strings.Split(path, string(separator))
		for i, s := range parts {
			var comp component
			switch {
			case s == "" && i == 0 && path != "":
				comp = component{typ: componentRoot}
			case s == "" || s == currentDir:
				comp = component{typ: componentCurrent}
			case s == parentDir:
				comp = component{typ: componentParent}
			default:
				if strings.Contains(s, "\\") {
					// Will be handled as error during normalization
					comp = component{typ: componentNormal, value: s}
				} else {
					comp = component{typ: componentNormal, value: s}
				}
			}
			if !yield(comp) {
				return
			}
		}
	}
}

// segment represents a segment in a normalized path.
// A segment is never empty, ".", or ".." and it never contains back- or forward slashes.
type segment string

func newSegment(s string) (segment, error) {
	if s == "" || s == currentDir || s == parentDir {
		return "", errors.New("invalid segment")
	}
	if strings.ContainsAny(s, "/\\") {
		return "", errors.New("segment contains slash")
	}
	return segment(s), nil
}

// splitDot splits a segment at the last dot, returning (before, after).
// If there's no dot or the segment starts with a dot, returns appropriately.
func splitDot(s string) (before, after string) {
	idx := strings.LastIndex(s, ".")
	if idx == -1 {
		return "", s
	}
	if idx == 0 {
		return s, ""
	}
	return s[:idx], s[idx+1:]
}

// segments stores a sequence of path segments as a string.
// The underlying string always represents a normalized absolute path and is
// guaranteed to start with a slash.
type segments struct {
	data string
}

func newSegments() segments {
	return segments{data: "/"}
}

func normalizeSegments(comps func(yield func(component) bool)) (segments, error) {
	var compList []component
	comps(func(c component) bool {
		compList = append(compList, c)
		return true
	})
	return normalizeComponents(compList)
}

func normalizeComponents(comps []component) (segments, error) {
	out := newSegments()
	for _, comp := range comps {
		if err := out.pushComponent(comp); err != nil {
			return segments{}, err
		}
	}
	return out, nil
}

func (s *segments) isEmpty() bool {
	return len(s.data) == 1
}

func (s *segments) getWithSlash() string {
	return s.data
}

func (s *segments) getWithoutSlash() string {
	if len(s.data) > 1 {
		return s.data[1:]
	}
	return ""
}

func (s *segments) clear() {
	s.data = "/"
}

func (s *segments) pushComponent(comp component) error {
	switch comp.typ {
	case componentRoot:
		s.clear()
	case componentCurrent:
		// No effect
	case componentParent:
		if !s.pop() {
			return PathErrorEscapes
		}
	case componentNormal:
		// Check for backslash
		if strings.Contains(comp.value, "\\") {
			return PathErrorBackslash
		}
		seg, err := newSegment(comp.value)
		if err != nil {
			return err
		}
		s.push(seg)
	}
	return nil
}

func (s *segments) push(seg segment) {
	if !s.isEmpty() {
		s.data += string(separator)
	}
	s.data += string(seg)
}

func (s *segments) pop() bool {
	if s.isEmpty() {
		return false
	}
	idx := strings.LastIndex(s.data, string(separator))
	if idx <= 0 {
		s.data = "/"
	} else {
		s.data = s.data[:idx]
	}
	return true
}

func (s *segments) last() string {
	if s.isEmpty() {
		return ""
	}
	idx := strings.LastIndex(s.data, string(separator))
	return s.data[idx+1:]
}

func (s *segments) iter() []string {
	if s.isEmpty() {
		return nil
	}
	// Skip leading slash
	str := s.data[1:]
	if str == "" {
		return nil
	}
	return strings.Split(str, string(separator))
}

func (s *segments) clone() segments {
	return segments{data: s.data}
}
