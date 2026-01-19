# GoTypst

A Go implementation of the [Typst](https://typst.app) typesetting system.

## Package Structure

The module structure mirrors the Typst Rust crates:

| Package | Description |
|---------|-------------|
| `gotypst` | Core interfaces including World and Document types |
| `syntax` | Parser and syntax tree for Typst source |
| `library` | Standard library functions and elements |
| `pdf` | PDF export |
| `svg` | SVG export |
| `html` | HTML export |
| `render` | Image rendering |
| `kit` | Utilities for implementing World |

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
import "github.com/boergens/gotypst"
```

Implement the `World` interface to provide file system access and compile Typst documents.

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
