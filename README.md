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

## Usage

```go
import "github.com/gotypst/gotypst"
```

Implement the `World` interface to provide file system access and compile Typst documents.
