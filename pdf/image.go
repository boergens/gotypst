package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/boergens/gotypst/layout/pages"
)

// ImageXObject represents a PDF image XObject.
type ImageXObject struct {
	// Ref is the indirect object reference for this image.
	Ref Ref
	// Width is the image width in pixels.
	Width int
	// Height is the image height in pixels.
	Height int
	// BitsPerComponent is typically 8.
	BitsPerComponent int
	// ColorSpace is the PDF color space name.
	ColorSpace string
	// Filter is the compression filter (e.g., DCTDecode, FlateDecode).
	Filter string
	// Data is the compressed image data.
	Data []byte
	// SMask is an optional soft mask for transparency.
	SMask *ImageXObject
}

// ToIndirectObject converts the image XObject to a PDF indirect object.
func (img *ImageXObject) ToIndirectObject() IndirectObject {
	dict := Dict{
		Name("Type"):             Name("XObject"),
		Name("Subtype"):          Name("Image"),
		Name("Width"):            Int(img.Width),
		Name("Height"):           Int(img.Height),
		Name("BitsPerComponent"): Int(img.BitsPerComponent),
		Name("ColorSpace"):       Name(img.ColorSpace),
	}

	if img.Filter != "" {
		dict[Name("Filter")] = Name(img.Filter)
	}

	if img.SMask != nil {
		dict[Name("SMask")] = img.SMask.Ref
	}

	return IndirectObject{
		Ref: img.Ref,
		Object: Stream{
			Dict: dict,
			Data: img.Data,
		},
	}
}

// EncodeImage encodes a pages.Image into an ImageXObject for PDF embedding.
func EncodeImage(img *pages.Image, ref Ref) (*ImageXObject, error) {
	switch img.Format {
	case pages.ImageFormatJPEG:
		return encodeJPEGImage(img, ref)
	case pages.ImageFormatPNG:
		return encodePNGImage(img, ref)
	case pages.ImageFormatRaw:
		return encodeRawImage(img, ref)
	default:
		return nil, errors.New("unsupported image format")
	}
}

// encodeJPEGImage creates an ImageXObject from JPEG data.
// JPEG data can be embedded directly using DCTDecode filter.
func encodeJPEGImage(img *pages.Image, ref Ref) (*ImageXObject, error) {
	return &ImageXObject{
		Ref:              ref,
		Width:            img.Width,
		Height:           img.Height,
		BitsPerComponent: img.BitsPerComponent,
		ColorSpace:       img.ColorSpace.String(),
		Filter:           "DCTDecode",
		Data:             img.Data,
	}, nil
}

// encodePNGImage creates an ImageXObject from PNG data.
// PNG data needs to be decoded and re-encoded for PDF.
func encodePNGImage(img *pages.Image, ref Ref) (*ImageXObject, error) {
	// Decode PNG to get raw pixel data
	pngImg, err := png.Decode(bytes.NewReader(img.Data))
	if err != nil {
		return nil, err
	}

	return encodeGoImage(pngImg, ref)
}

// encodeRawImage creates an ImageXObject from raw pixel data.
func encodeRawImage(img *pages.Image, ref Ref) (*ImageXObject, error) {
	// Compress raw data with zlib/FlateDecode
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(img.Data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	xobj := &ImageXObject{
		Ref:              ref,
		Width:            img.Width,
		Height:           img.Height,
		BitsPerComponent: img.BitsPerComponent,
		ColorSpace:       img.ColorSpace.String(),
		Filter:           "FlateDecode",
		Data:             buf.Bytes(),
	}

	// Handle alpha channel if present
	if len(img.Alpha) > 0 {
		smask, err := encodeAlphaMask(img.Alpha, img.Width, img.Height, Ref{ID: ref.ID + 1})
		if err != nil {
			return nil, err
		}
		xobj.SMask = smask
	}

	return xobj, nil
}

// encodeGoImage encodes a Go image.Image to an ImageXObject.
func encodeGoImage(img image.Image, ref Ref) (*ImageXObject, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Extract RGB and alpha data
	rgb := make([]byte, width*height*3)
	var alpha []byte
	hasAlpha := false

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			idx := (y*width + x) * 3
			rgb[idx] = uint8(r >> 8)
			rgb[idx+1] = uint8(g >> 8)
			rgb[idx+2] = uint8(b >> 8)

			if a != 0xFFFF {
				if alpha == nil {
					alpha = make([]byte, width*height)
					// Fill previous pixels with full opacity
					for i := 0; i < y*width+x; i++ {
						alpha[i] = 255
					}
				}
				hasAlpha = true
			}
			if alpha != nil {
				alpha[y*width+x] = uint8(a >> 8)
			}
		}
	}

	// Compress RGB data
	var rgbBuf bytes.Buffer
	zw := zlib.NewWriter(&rgbBuf)
	if _, err := zw.Write(rgb); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	xobj := &ImageXObject{
		Ref:              ref,
		Width:            width,
		Height:           height,
		BitsPerComponent: 8,
		ColorSpace:       "DeviceRGB",
		Filter:           "FlateDecode",
		Data:             rgbBuf.Bytes(),
	}

	// Create soft mask for alpha if present
	if hasAlpha && alpha != nil {
		smask, err := encodeAlphaMask(alpha, width, height, Ref{ID: ref.ID + 1})
		if err != nil {
			return nil, err
		}
		xobj.SMask = smask
	}

	return xobj, nil
}

// encodeAlphaMask creates a soft mask XObject for alpha transparency.
func encodeAlphaMask(alpha []byte, width, height int, ref Ref) (*ImageXObject, error) {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(alpha); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	return &ImageXObject{
		Ref:              ref,
		Width:            width,
		Height:           height,
		BitsPerComponent: 8,
		ColorSpace:       "DeviceGray",
		Filter:           "FlateDecode",
		Data:             buf.Bytes(),
	}, nil
}

// DecodeImageFile decodes an image from raw file data.
// It auto-detects the format (JPEG, PNG) and returns a pages.Image.
func DecodeImageFile(data []byte) (*pages.Image, error) {
	// Try JPEG first
	if isJPEG(data) {
		return decodeJPEGFile(data)
	}

	// Try PNG
	if isPNG(data) {
		return decodePNGFile(data)
	}

	// Try generic image.Decode
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return imageToPages(img, format)
}

// isJPEG checks if data starts with JPEG magic bytes.
func isJPEG(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8
}

// isPNG checks if data starts with PNG magic bytes.
func isPNG(data []byte) bool {
	return len(data) >= 8 &&
		data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A
}

// decodeJPEGFile decodes JPEG data and returns a pages.Image.
// For JPEG, we can embed the original data directly with DCTDecode.
func decodeJPEGFile(data []byte) (*pages.Image, error) {
	// Decode to get dimensions
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	colorSpace := pages.ColorSpaceDeviceRGB
	if cfg.ColorModel == color.GrayModel {
		colorSpace = pages.ColorSpaceDeviceGray
	} else if cfg.ColorModel == color.CMYKModel {
		colorSpace = pages.ColorSpaceDeviceCMYK
	}

	return &pages.Image{
		Data:             data,
		Format:           pages.ImageFormatJPEG,
		Width:            cfg.Width,
		Height:           cfg.Height,
		BitsPerComponent: 8,
		ColorSpace:       colorSpace,
	}, nil
}

// decodePNGFile decodes PNG data and returns a pages.Image.
func decodePNGFile(data []byte) (*pages.Image, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Extract RGB and alpha data
	rgb := make([]byte, width*height*3)
	var alpha []byte

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			idx := (y*width + x) * 3
			rgb[idx] = uint8(r >> 8)
			rgb[idx+1] = uint8(g >> 8)
			rgb[idx+2] = uint8(b >> 8)

			if a != 0xFFFF {
				if alpha == nil {
					alpha = make([]byte, width*height)
					// Fill previous pixels with full opacity
					for i := 0; i < y*width+x; i++ {
						alpha[i] = 255
					}
				}
			}
			if alpha != nil {
				alpha[y*width+x] = uint8(a >> 8)
			}
		}
	}

	return &pages.Image{
		Data:             rgb,
		Format:           pages.ImageFormatRaw,
		Width:            width,
		Height:           height,
		BitsPerComponent: 8,
		ColorSpace:       pages.ColorSpaceDeviceRGB,
		Alpha:            alpha,
	}, nil
}

// imageToPages converts a Go image.Image to a pages.Image.
func imageToPages(img image.Image, format string) (*pages.Image, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Extract RGB and alpha data
	rgb := make([]byte, width*height*3)
	var alpha []byte

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			idx := (y*width + x) * 3
			rgb[idx] = uint8(r >> 8)
			rgb[idx+1] = uint8(g >> 8)
			rgb[idx+2] = uint8(b >> 8)

			if a != 0xFFFF {
				if alpha == nil {
					alpha = make([]byte, width*height)
					for i := 0; i < y*width+x; i++ {
						alpha[i] = 255
					}
				}
			}
			if alpha != nil {
				alpha[y*width+x] = uint8(a >> 8)
			}
		}
	}

	return &pages.Image{
		Data:             rgb,
		Format:           pages.ImageFormatRaw,
		Width:            width,
		Height:           height,
		BitsPerComponent: 8,
		ColorSpace:       pages.ColorSpaceDeviceRGB,
		Alpha:            alpha,
	}, nil
}

// EncodeJPEG encodes an image to JPEG format for efficient PDF embedding.
func EncodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ReadPNGInfo reads PNG header information without fully decoding.
func ReadPNGInfo(r io.Reader) (width, height int, err error) {
	// Read PNG signature
	sig := make([]byte, 8)
	if _, err := io.ReadFull(r, sig); err != nil {
		return 0, 0, err
	}
	if !isPNG(sig) {
		return 0, 0, errors.New("not a PNG file")
	}

	// Read IHDR chunk
	var chunkLen uint32
	if err := binary.Read(r, binary.BigEndian, &chunkLen); err != nil {
		return 0, 0, err
	}

	chunkType := make([]byte, 4)
	if _, err := io.ReadFull(r, chunkType); err != nil {
		return 0, 0, err
	}

	if string(chunkType) != "IHDR" {
		return 0, 0, errors.New("expected IHDR chunk")
	}

	var w, h uint32
	if err := binary.Read(r, binary.BigEndian, &w); err != nil {
		return 0, 0, err
	}
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return 0, 0, err
	}

	return int(w), int(h), nil
}
