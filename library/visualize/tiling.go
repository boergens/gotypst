package visualize

import (
	"fmt"
)

// Length represents a length value in points.
type Length float64

// Tiling represents a repeating pattern fill.
// Tilings consist of content that is repeated to fill a shape.
type Tiling struct {
	// Size is the size of each cell in the tiling pattern.
	// nil means auto (use the content's natural size).
	Size *[2]Length

	// Spacing is the gap between cells.
	// Default: (0, 0).
	Spacing [2]Length

	// Relative specifies placement relative to self or parent.
	Relative Relative

	// Body is the content of each cell.
	// This is stored as an interface to allow different content types.
	Body interface{}
}

func (*Tiling) valueMarker() {}
func (*Tiling) Type() string { return "tiling" }
func (t *Tiling) String() string {
	if t == nil {
		return "tiling()"
	}
	if t.Size != nil {
		return fmt.Sprintf("tiling(size: (%.1fpt, %.1fpt))", t.Size[0], t.Size[1])
	}
	return "tiling(size: auto)"
}

// NewTiling creates a new tiling pattern.
func NewTiling(body interface{}, opts ...TilingOption) *Tiling {
	t := &Tiling{
		Size:     nil, // auto
		Spacing:  [2]Length{0, 0},
		Relative: RelativeAuto,
		Body:     body,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// TilingOption is a functional option for configuring tilings.
type TilingOption func(*Tiling)

// WithSize sets the cell size for the tiling.
func WithSize(width, height Length) TilingOption {
	return func(t *Tiling) {
		size := [2]Length{width, height}
		t.Size = &size
	}
}

// WithSpacing sets the spacing between cells.
func WithSpacing(x, y Length) TilingOption {
	return func(t *Tiling) {
		t.Spacing = [2]Length{x, y}
	}
}

// WithTilingRelative sets the relative positioning mode.
func WithTilingRelative(rel Relative) TilingOption {
	return func(t *Tiling) {
		t.Relative = rel
	}
}

// --- Tiling Methods ---

// GetSize returns the cell size, or nil for auto.
func (t *Tiling) GetSize() *[2]Length {
	if t == nil {
		return nil
	}
	return t.Size
}

// GetSpacing returns the spacing between cells.
func (t *Tiling) GetSpacing() [2]Length {
	if t == nil {
		return [2]Length{0, 0}
	}
	return t.Spacing
}

// GetRelative returns the relative positioning mode.
func (t *Tiling) GetRelative() Relative {
	if t == nil {
		return RelativeAuto
	}
	return t.Relative
}

// GetBody returns the content body of each cell.
func (t *Tiling) GetBody() interface{} {
	if t == nil {
		return nil
	}
	return t.Body
}

// IsAuto returns true if the size is auto-determined.
func (t *Tiling) IsAuto() bool {
	return t == nil || t.Size == nil
}

// CellWidth returns the cell width in points, or 0 if auto.
func (t *Tiling) CellWidth() Length {
	if t == nil || t.Size == nil {
		return 0
	}
	return t.Size[0]
}

// CellHeight returns the cell height in points, or 0 if auto.
func (t *Tiling) CellHeight() Length {
	if t == nil || t.Size == nil {
		return 0
	}
	return t.Size[1]
}

// HorizontalSpacing returns the horizontal spacing between cells.
func (t *Tiling) HorizontalSpacing() Length {
	if t == nil {
		return 0
	}
	return t.Spacing[0]
}

// VerticalSpacing returns the vertical spacing between cells.
func (t *Tiling) VerticalSpacing() Length {
	if t == nil {
		return 0
	}
	return t.Spacing[1]
}

// EffectiveCellSize returns the effective cell size including spacing.
// Returns (0, 0) if size is auto.
func (t *Tiling) EffectiveCellSize() (width, height Length) {
	if t == nil || t.Size == nil {
		return 0, 0
	}
	return t.Size[0] + t.Spacing[0], t.Size[1] + t.Spacing[1]
}

// Clone creates a copy of the tiling.
func (t *Tiling) Clone() *Tiling {
	if t == nil {
		return nil
	}
	clone := *t
	if t.Size != nil {
		size := *t.Size
		clone.Size = &size
	}
	return &clone
}
