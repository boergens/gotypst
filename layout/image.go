package layout

import (
	"github.com/boergens/gotypst/syntax"
)

// ImageElem represents the image element configuration.
type ImageElem struct {
	// Image is the image data.
	Image *Image
	// Fit determines how the image fits in its container.
	Fit ImageFit
	// Span is the source span for error reporting.
	Span syntax.Span
}

// LayoutImage lays out an image element.
//
// This function handles image sizing and positioning based on:
// - The image's natural size (pixel dimensions and DPI)
// - The available region
// - The image fit mode (contain, cover, stretch)
func LayoutImage(
	elem *ImageElem,
	region Region,
) (*Frame, error) {
	image := elem.Image

	// Determine the image's pixel dimensions.
	pxw := image.Width
	pxh := image.Height
	if pxh == 0 {
		pxh = 1 // Avoid division by zero
	}

	// Calculate aspect ratios.
	pxRatio := pxw / pxh
	if region.Size.Y == 0 {
		region.Size.Y = 1 // Avoid division by zero
	}
	regionRatio := float64(region.Size.X) / float64(region.Size.Y)

	// Find out whether the image is wider or taller than the region.
	wide := pxRatio > regionRatio

	// The space into which the image will be placed according to its fit.
	var target Size
	if region.Expand.X && region.Expand.Y {
		// If both width and height are forced, take them.
		target = region.Size
	} else if region.Expand.X {
		// If just width is forced, take it.
		constrainedHeight := region.Size.Y.Min(Abs(float64(region.Size.X) / pxRatio))
		target = NewSize(region.Size.X, constrainedHeight)
	} else if region.Expand.Y {
		// If just height is forced, take it.
		constrainedWidth := region.Size.X.Min(Abs(float64(region.Size.Y) * pxRatio))
		target = NewSize(constrainedWidth, region.Size.Y)
	} else {
		// If neither is forced, take the natural image size at the image's
		// DPI bounded by the available space.
		dpi := image.DPI
		if dpi == nil {
			defaultDPI := DefaultDPI
			dpi = &defaultDPI
		}

		// Calculate natural size in points.
		naturalX := Inches(pxw / *dpi)
		naturalY := Inches(pxh / *dpi)

		// Constrain to available space while maintaining aspect ratio.
		constrainedX := naturalX.Min(region.Size.X).Min(Abs(float64(region.Size.Y) * pxRatio))
		constrainedY := naturalY.Min(region.Size.Y).Min(Abs(float64(region.Size.X) / pxRatio))

		target = NewSize(constrainedX, constrainedY)
	}

	// Compute the actual size of the fitted image.
	fit := elem.Fit
	var fitted Size
	switch fit {
	case ImageFitCover, ImageFitContain:
		if wide == (fit == ImageFitContain) {
			// Width constrained, height follows ratio
			fitted = NewSize(target.X, Abs(float64(target.X)/pxRatio))
		} else {
			// Height constrained, width follows ratio
			fitted = NewSize(Abs(float64(target.Y)*pxRatio), target.Y)
		}
	case ImageFitStretch:
		fitted = target
	default:
		fitted = target
	}

	// First, place the image in a frame of exactly its size.
	frame := NewFrameSoft(fitted)

	// Push the image item to the frame.
	frame.Push(PointZero(), ImageItem{
		Image: *image,
		Size:  fitted,
		Span:  elem.Span,
	})

	// Resize the frame to the target size, center aligning the image in the process.
	frame.Resize(target, NewAxes(AlignCenter))

	// Create a clipping group if only part of the image should be visible.
	if fit == ImageFitCover && !target.Fits(fitted) {
		frame.Clip(CurveRect(frame.Size))
	}

	return frame, nil
}
