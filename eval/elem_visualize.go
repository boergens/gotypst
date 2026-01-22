package eval

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Visualize Element Functions
// ----------------------------------------------------------------------------
// This file contains visualization element functions that correspond to Typst's
// visualize module:
// - image() - embedded images
//
// Reference: typst-reference/crates/typst-library/src/visualize/

// ----------------------------------------------------------------------------
// Image Element
// ----------------------------------------------------------------------------

// ImageFit represents how an image fits within its container.
type ImageFit string

const (
	// ImageFitContain scales the image to fit within the bounds while preserving aspect ratio.
	ImageFitContain ImageFit = "contain"
	// ImageFitCover scales the image to cover the bounds while preserving aspect ratio.
	ImageFitCover ImageFit = "cover"
	// ImageFitStretch stretches the image to fill the bounds exactly.
	ImageFitStretch ImageFit = "stretch"
)

// ImageElement represents an embedded image element.
type ImageElement struct {
	// Path is the source path of the image file.
	Path string
	// Data is the raw image bytes (loaded from Path).
	Data []byte
	// Width is the rendered width (in points). If nil, auto-sizes.
	Width *float64
	// Height is the rendered height (in points). If nil, auto-sizes.
	Height *float64
	// Fit specifies how the image fits within its bounds.
	Fit ImageFit
	// Alt is the alt text for accessibility.
	Alt *string
	// NaturalWidth is the natural width of the image in pixels.
	NaturalWidth int
	// NaturalHeight is the natural height of the image in pixels.
	NaturalHeight int
}

func (*ImageElement) IsContentElement() {}

// ImageFunc creates the image element function.
func ImageFunc() *Func {
	name := "image"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: imageNative,
			Info: &FuncInfo{
				Name: "image",
				Params: []ParamInfo{
					{Name: "path", Type: TypeStr, Named: false},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "fit", Type: TypeStr, Default: Str("contain"), Named: true},
					{Name: "alt", Type: TypeStr, Default: None, Named: true},
				},
			},
		},
	}
}

// imageNative implements the image() function.
// Loads an image from the given path and creates an ImageElement.
//
// Arguments:
//   - path (positional, str): The path to the image file
//   - width (named, length, default: auto): Rendered width
//   - height (named, length, default: auto): Rendered height
//   - fit (named, str, default: "contain"): How to fit within bounds ("contain", "cover", "stretch")
//   - alt (named, str, default: none): Alt text for accessibility
func imageNative(engine *Engine, context *Context, args *Args) (Value, error) {
	// Get required path argument
	pathArg, err := args.Expect("path")
	if err != nil {
		return nil, err
	}

	path, ok := AsStr(pathArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      pathArg.V.Type().String(),
			Span:     pathArg.Span,
		}
	}

	elem := &ImageElement{
		Path: path,
		Fit:  ImageFitContain,
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !IsAuto(widthArg.V) && !IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(LengthValue); ok {
				w := lv.Length.Points
				elem.Width = &w
			} else if rv, ok := widthArg.V.(RelativeValue); ok {
				// Handle relative values (percentages)
				w := rv.Relative.Rel.Value * 100 // Store as percentage
				elem.Width = &w
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      widthArg.V.Type().String(),
					Span:     widthArg.Span,
				}
			}
		}
	}

	// Get optional height argument
	if heightArg := args.Find("height"); heightArg != nil {
		if !IsAuto(heightArg.V) && !IsNone(heightArg.V) {
			if lv, ok := heightArg.V.(LengthValue); ok {
				h := lv.Length.Points
				elem.Height = &h
			} else if rv, ok := heightArg.V.(RelativeValue); ok {
				h := rv.Relative.Rel.Value * 100
				elem.Height = &h
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      heightArg.V.Type().String(),
					Span:     heightArg.Span,
				}
			}
		}
	}

	// Get optional fit argument (default: "contain")
	if fitArg := args.Find("fit"); fitArg != nil {
		if !IsNone(fitArg.V) && !IsAuto(fitArg.V) {
			fitStr, ok := AsStr(fitArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      fitArg.V.Type().String(),
					Span:     fitArg.Span,
				}
			}
			// Validate fit value
			switch fitStr {
			case "contain":
				elem.Fit = ImageFitContain
			case "cover":
				elem.Fit = ImageFitCover
			case "stretch":
				elem.Fit = ImageFitStretch
			default:
				return nil, &TypeMismatchError{
					Expected: "\"contain\", \"cover\", or \"stretch\"",
					Got:      "\"" + fitStr + "\"",
					Span:     fitArg.Span,
				}
			}
		}
	}

	// Get optional alt argument
	if altArg := args.Find("alt"); altArg != nil {
		if !IsNone(altArg.V) {
			altStr, ok := AsStr(altArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      altArg.V.Type().String(),
					Span:     altArg.Span,
				}
			}
			elem.Alt = &altStr
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Load the image file
	data, err := readFileFromWorld(engine, path)
	if err != nil {
		return nil, err
	}
	elem.Data = data

	// Decode image to get natural dimensions
	width, height, err := decodeImageDimensions(data)
	if err != nil {
		return nil, &FileParseError{
			Path:    path,
			Format:  "image",
			Message: err.Error(),
		}
	}
	elem.NaturalWidth = width
	elem.NaturalHeight = height

	// Create the ImageElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// decodeImageDimensions decodes an image to get its natural dimensions.
func decodeImageDimensions(data []byte) (width, height int, err error) {
	// Try JPEG first
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		return decodeJPEGDimensions(data)
	}

	// Try PNG
	if len(data) >= 8 &&
		data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return decodePNGDimensions(data)
	}

	// Fallback to generic image decode
	return decodeGenericDimensions(data)
}

// decodeJPEGDimensions extracts dimensions from JPEG data without full decode.
func decodeJPEGDimensions(data []byte) (width, height int, err error) {
	// Parse JPEG markers to find SOF
	i := 2 // Skip SOI marker
	for i < len(data)-1 {
		if data[i] != 0xFF {
			return 0, 0, fmt.Errorf("invalid JPEG marker")
		}
		marker := data[i+1]
		i += 2

		// Skip padding
		for marker == 0xFF && i < len(data) {
			marker = data[i]
			i++
		}

		// Check for SOF markers (Start of Frame)
		if (marker >= 0xC0 && marker <= 0xC3) || (marker >= 0xC5 && marker <= 0xC7) ||
			(marker >= 0xC9 && marker <= 0xCB) || (marker >= 0xCD && marker <= 0xCF) {
			if i+7 > len(data) {
				return 0, 0, fmt.Errorf("truncated JPEG")
			}
			height = int(data[i+1])<<8 | int(data[i+2])
			width = int(data[i+3])<<8 | int(data[i+4])
			return width, height, nil
		}

		// Skip other markers
		if marker == 0xD8 || marker == 0xD9 || (marker >= 0xD0 && marker <= 0xD7) {
			continue // No length field
		}

		if i+1 >= len(data) {
			return 0, 0, fmt.Errorf("truncated JPEG")
		}
		length := int(data[i])<<8 | int(data[i+1])
		i += length
	}

	return 0, 0, fmt.Errorf("no SOF marker found in JPEG")
}

// decodePNGDimensions extracts dimensions from PNG data without full decode.
func decodePNGDimensions(data []byte) (width, height int, err error) {
	// PNG dimensions are in the IHDR chunk, immediately after the signature
	if len(data) < 24 {
		return 0, 0, fmt.Errorf("truncated PNG")
	}

	// Skip signature (8 bytes), length (4 bytes), "IHDR" (4 bytes)
	// Then read width (4 bytes) and height (4 bytes)
	width = int(data[16])<<24 | int(data[17])<<16 | int(data[18])<<8 | int(data[19])
	height = int(data[20])<<24 | int(data[21])<<16 | int(data[22])<<8 | int(data[23])

	return width, height, nil
}

// decodeGenericDimensions uses Go's image package to decode dimensions.
func decodeGenericDimensions(data []byte) (width, height int, err error) {
	cfg, _, err := decodeImageConfig(data)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

// decodeImageConfig decodes image configuration (dimensions) from raw data.
func decodeImageConfig(data []byte) (image.Config, string, error) {
	return image.DecodeConfig(bytes.NewReader(data))
}

// ----------------------------------------------------------------------------
// Shape Element Functions (stubs)
// ----------------------------------------------------------------------------
// These are stub implementations for the visualize module shape elements.
// Full implementations should match typst-library/src/visualize/shape.rs

// RectFunc creates the rect element function.
func RectFunc() *Func {
	name := "rect"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: shapeStub("rect"),
			Info: &FuncInfo{Name: "rect"},
		},
	}
}

// CircleFunc creates the circle element function.
func CircleFunc() *Func {
	name := "circle"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: shapeStub("circle"),
			Info: &FuncInfo{Name: "circle"},
		},
	}
}

// EllipseFunc creates the ellipse element function.
func EllipseFunc() *Func {
	name := "ellipse"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: shapeStub("ellipse"),
			Info: &FuncInfo{Name: "ellipse"},
		},
	}
}

// LineFunc creates the line element function.
func LineFunc() *Func {
	name := "line"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: shapeStub("line"),
			Info: &FuncInfo{Name: "line"},
		},
	}
}

// PathFunc creates the path element function.
func PathFunc() *Func {
	name := "path"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: shapeStub("path"),
			Info: &FuncInfo{Name: "path"},
		},
	}
}

// PolygonFunc creates the polygon element function.
func PolygonFunc() *Func {
	name := "polygon"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: shapeStub("polygon"),
			Info: &FuncInfo{Name: "polygon"},
		},
	}
}

// SquareFunc creates the square element function.
func SquareFunc() *Func {
	name := "square"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: shapeStub("square"),
			Info: &FuncInfo{Name: "square"},
		},
	}
}

// shapeStub creates a stub native function for shape elements.
func shapeStub(name string) func(*Engine, *Context, *Args) (Value, error) {
	return func(engine *Engine, context *Context, args *Args) (Value, error) {
		// Consume all arguments silently
		args.Finish()
		// Return empty content
		return ContentValue{Content: Content{}}, nil
	}
}
