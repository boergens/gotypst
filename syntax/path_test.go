package syntax

import (
	"path/filepath"
	"testing"
)

func mustPath(t *testing.T, p string) *VirtualPath {
	t.Helper()
	vp, err := NewVirtualPath(p)
	if err != nil {
		t.Fatalf("NewVirtualPath(%q) failed: %v", p, err)
	}
	return vp
}

func TestNewVirtualPath(t *testing.T) {
	tests := []struct {
		path    string
		want    string
		wantErr error
	}{
		{"", "/", nil},
		{"a/./file.txt", "/a/file.txt", nil},
		{"file.txt", "/file.txt", nil},
		{"/file.txt", "/file.txt", nil},
		{"hello/world", "/hello/world", nil},
		{"hello/world/", "/hello/world", nil},
		{"a///b", "/a/b", nil},
		{"/a///b", "/a/b", nil},
		{"./world.txt", "/world.txt", nil},
		{"./world.txt/", "/world.txt", nil},
		{"hello/.././/wor/ld.typ.extra", "/wor/ld.typ.extra", nil},
		{"hello/.../world", "/hello/.../world", nil},
		{"..", "", PathErrorEscapes},
		{"../world.txt", "", PathErrorEscapes},
		{"a\\world.txt", "", PathErrorBackslash},
	}

	for _, tt := range tests {
		vp, err := NewVirtualPath(tt.path)
		if tt.wantErr != nil {
			if err == nil {
				t.Errorf("NewVirtualPath(%q) expected error %v, got nil", tt.path, tt.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("NewVirtualPath(%q) unexpected error: %v", tt.path, err)
			continue
		}
		if got := vp.GetWithSlash(); got != tt.want {
			t.Errorf("NewVirtualPath(%q).GetWithSlash() = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestVirtualPathRealize(t *testing.T) {
	vp := mustPath(t, "src/text/main.typ")
	root := "/home/users/typst"
	want := filepath.Join(root, "src", "text", "main.typ")
	if got := vp.Realize(root); got != want {
		t.Errorf("Realize(%q) = %q, want %q", root, got, want)
	}
}

func TestVirtualPathFileOps(t *testing.T) {
	// Test file.typ
	p1 := mustPath(t, "src/text/file.typ")
	if got := p1.FileName(); got != "file.typ" {
		t.Errorf("FileName() = %q, want %q", got, "file.typ")
	}
	if got := p1.FileStem(); got != "file" {
		t.Errorf("FileStem() = %q, want %q", got, "file")
	}
	if got := p1.Extension(); got != "typ" {
		t.Errorf("Extension() = %q, want %q", got, "typ")
	}
	withExt := p1.WithExtension("txt")
	if got := withExt.GetWithSlash(); got != "/src/text/file.txt" {
		t.Errorf("WithExtension(\"txt\") = %q, want %q", got, "/src/text/file.txt")
	}
	parent := p1.Parent()
	if parent == nil {
		t.Error("Parent() returned nil")
	} else if got := parent.GetWithSlash(); got != "/src/text" {
		t.Errorf("Parent() = %q, want %q", got, "/src/text")
	}

	// Test src (no extension)
	p2 := mustPath(t, "src")
	if got := p2.FileName(); got != "src" {
		t.Errorf("FileName() = %q, want %q", got, "src")
	}
	if got := p2.FileStem(); got != "src" {
		t.Errorf("FileStem() = %q, want %q", got, "src")
	}
	if got := p2.Extension(); got != "" {
		t.Errorf("Extension() = %q, want %q", got, "")
	}
	withExt2 := p2.WithExtension("txt")
	if got := withExt2.GetWithSlash(); got != "/src.txt" {
		t.Errorf("WithExtension(\"txt\") = %q, want %q", got, "/src.txt")
	}
	parent2 := p2.Parent()
	if parent2 == nil {
		t.Error("Parent() returned nil")
	} else if got := parent2.GetWithSlash(); got != "/" {
		t.Errorf("Parent() = %q, want %q", got, "/")
	}

	// Test empty path (root)
	p3 := mustPath(t, "")
	if got := p3.FileName(); got != "" {
		t.Errorf("FileName() = %q, want %q", got, "")
	}
	if got := p3.FileStem(); got != "" {
		t.Errorf("FileStem() = %q, want %q", got, "")
	}
	if got := p3.Extension(); got != "" {
		t.Errorf("Extension() = %q, want %q", got, "")
	}
	parent3 := p3.Parent()
	if parent3 != nil {
		t.Errorf("Parent() = %v, want nil", parent3)
	}
}

func TestVirtualPathJoin(t *testing.T) {
	p1 := mustPath(t, "src")

	// Test backslash error
	_, err := p1.Join("a\\b")
	if err != PathErrorBackslash {
		t.Errorf("Join(\"a\\\\b\") error = %v, want PathErrorBackslash", err)
	}

	// Test normal join
	p2, err := p1.Join("text")
	if err != nil {
		t.Errorf("Join(\"text\") error = %v", err)
	}
	if got := p2.GetWithSlash(); got != "/src/text" {
		t.Errorf("Join(\"text\") = %q, want %q", got, "/src/text")
	}

	// Test join with parent
	p3, err := p2.Join("..")
	if err != nil {
		t.Errorf("Join(\"..\") error = %v", err)
	}
	if got := p3.GetWithSlash(); got != "/src" {
		t.Errorf("Join(\"..\") = %q, want %q", got, "/src")
	}

	// Test join parent from root
	p4 := mustPath(t, "/")
	_, err = p4.Join("..")
	if err != PathErrorEscapes {
		t.Errorf("root.Join(\"..\") error = %v, want PathErrorEscapes", err)
	}
}

func TestVirtualPathRelativeFrom(t *testing.T) {
	p1 := mustPath(t, "src/text/main.typ")

	tests := []struct {
		base string
		want string
	}{
		{"/src/text", "main.typ"},
		{"/src/data", "../text/main.typ"},
		{"src/", "text/main.typ"},
		{"/", "src/text/main.typ"},
	}

	for _, tt := range tests {
		base := mustPath(t, tt.base)
		if got := p1.RelativeFrom(base); got != tt.want {
			t.Errorf("RelativeFrom(%q) = %q, want %q", tt.base, got, tt.want)
		}
	}

	p2 := mustPath(t, "src")
	if got := p2.RelativeFrom(mustPath(t, "src")); got != "" {
		t.Errorf("RelativeFrom same path = %q, want %q", got, "")
	}
	if got := p2.RelativeFrom(mustPath(t, "src/data")); got != ".." {
		t.Errorf("RelativeFrom(\"src/data\") = %q, want %q", got, "..")
	}
}

func TestSegments(t *testing.T) {
	s := newSegments()
	if got := s.getWithSlash(); got != "/" {
		t.Errorf("getWithSlash() = %q, want %q", got, "/")
	}
	if got := s.getWithoutSlash(); got != "" {
		t.Errorf("getWithoutSlash() = %q, want %q", got, "")
	}

	seg1, _ := newSegment("to")
	s.push(seg1)
	if got := s.getWithSlash(); got != "/to" {
		t.Errorf("getWithSlash() = %q, want %q", got, "/to")
	}

	seg2, _ := newSegment("hi.txt")
	s.push(seg2)
	if got := s.getWithSlash(); got != "/to/hi.txt" {
		t.Errorf("getWithSlash() = %q, want %q", got, "/to/hi.txt")
	}
	if got := s.getWithoutSlash(); got != "to/hi.txt" {
		t.Errorf("getWithoutSlash() = %q, want %q", got, "to/hi.txt")
	}
	if got := s.last(); got != "hi.txt" {
		t.Errorf("last() = %q, want %q", got, "hi.txt")
	}

	if !s.pop() {
		t.Error("pop() returned false")
	}
	if got := s.getWithSlash(); got != "/to" {
		t.Errorf("after pop, getWithSlash() = %q, want %q", got, "/to")
	}

	if !s.pop() {
		t.Error("pop() returned false")
	}
	if got := s.getWithSlash(); got != "/" {
		t.Errorf("after second pop, getWithSlash() = %q, want %q", got, "/")
	}

	if s.pop() {
		t.Error("pop() on empty should return false")
	}
	if got := s.getWithSlash(); got != "/" {
		t.Errorf("after pop on empty, getWithSlash() = %q, want %q", got, "/")
	}
	if got := s.last(); got != "" {
		t.Errorf("last() on empty = %q, want %q", got, "")
	}
}

func TestNewSegment(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"valid", false},
		{"file.txt", false},
		{"", true},
		{".", true},
		{"..", true},
		{"a/b", true},
		{"a\\b", true},
	}

	for _, tt := range tests {
		_, err := newSegment(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("newSegment(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestRootedPath(t *testing.T) {
	vp := mustPath(t, "src/main.typ")
	root := ProjectRoot()
	rp := NewRootedPath(root, *vp)

	if got := rp.String(); got != "/src/main.typ" {
		t.Errorf("RootedPath.String() = %q, want %q", got, "/src/main.typ")
	}

	// Test with package root
	spec := PackageSpec{Namespace: "preview", Name: "example", Version: PackageVersion{0, 1, 0}}
	pkgRoot := PackageRoot(spec)
	rp2 := NewRootedPath(pkgRoot, *vp)

	want := "@preview/example:0.1.0/src/main.typ"
	if got := rp2.String(); got != want {
		t.Errorf("RootedPath.String() with package = %q, want %q", got, want)
	}
}

func TestFileId(t *testing.T) {
	vp := mustPath(t, "src/main.typ")
	root := ProjectRoot()
	rp := NewRootedPath(root, *vp)

	// Create FileId
	id1 := rp.Intern()
	id2 := NewFileId(*rp)

	// Same path should return same id
	if id1.IntoRaw() != id2.IntoRaw() {
		t.Error("Same path should return same FileId")
	}

	// Get should return equivalent path
	got := id1.Get()
	if got.String() != rp.String() {
		t.Errorf("FileId.Get().String() = %q, want %q", got.String(), rp.String())
	}

	// UniqueFileId should create different id
	id3 := UniqueFileId(*rp)
	if id3.IntoRaw() == id1.IntoRaw() {
		t.Error("UniqueFileId should create different id from regular FileId")
	}
}
