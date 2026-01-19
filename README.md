# GoTypst

A Go implementation of Typst's core syntax types.

## Packages

### syntax

The `syntax` package provides:

- **Package Types** (`package.go`): Package manifest parsing types including:
  - `PackageManifest` - Parsed package manifest
  - `PackageInfo` - Package metadata
  - `PackageSpec` - Package identifier with namespace/name/version
  - `PackageVersion` - Semantic versioning
  - `VersionBound` - Version compatibility bounds
  - `TemplateInfo` - Template configuration
  - `ToolInfo` - Third-party tool configuration

- **Path Types** (`path.go`): Virtual, cross-platform reproducible path handling:
  - `VirtualPath` - Path in a virtual file system
  - `RootedPath` - Path with virtual root (project or package)
  - `FileId` - Interned file specification
  - `VirtualRoot` - Root of virtual file system (Project or Package)

## Usage

```go
import "github.com/boergens/gotypst/syntax"

// Parse a package spec
spec, err := syntax.ParsePackageSpec("@preview/example:0.1.0")

// Create a virtual path
vp, err := syntax.NewVirtualPath("src/main.typ")

// Join paths
joined, err := vp.Join("../lib/utils.typ")

// Realize to filesystem path
realPath := vp.Realize("/home/project")
```
