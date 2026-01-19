package layout

import "math"

// Region represents layout constraints for a single region.
// A region defines the available space and expansion behavior.
type Region struct {
	// Size is the available space in the region.
	Size Size
	// Expand indicates whether content should expand to fill each axis.
	Expand Axes[bool]
}

// NewRegion creates a new region with the given size.
func NewRegion(size Size) Region {
	return Region{Size: size}
}

// NewExpandedRegion creates a region that expands on both axes.
func NewExpandedRegion(size Size) Region {
	return Region{Size: size, Expand: Axes[bool]{X: true, Y: true}}
}

// Width returns the region's available width.
func (r Region) Width() Abs {
	return r.Size.Width
}

// Height returns the region's available height.
func (r Region) Height() Abs {
	return r.Size.Height
}

// IsFinite returns true if both dimensions are finite.
func (r Region) IsFinite() bool {
	return r.Size.Width.IsFinite() && r.Size.Height.IsFinite()
}

// WithExpand returns a new region with the given expansion flags.
func (r Region) WithExpand(expand Axes[bool]) Region {
	return Region{Size: r.Size, Expand: expand}
}

// WithSize returns a new region with the given size.
func (r Region) WithSize(size Size) Region {
	return Region{Size: size, Expand: r.Expand}
}

// Shrink reduces the region size by the given inset.
func (r Region) Shrink(inset Sides[Abs]) Region {
	return Region{
		Size: Size{
			Width:  (r.Size.Width - SumHorizontal(inset)).Max(0),
			Height: (r.Size.Height - SumVertical(inset)).Max(0),
		},
		Expand: r.Expand,
	}
}

// Regions represents layout constraints for multiple regions (pages/columns).
// It tracks the current region, remaining regions, and expansion behavior.
type Regions struct {
	// Size is the available space in the current region.
	Size Size
	// Full is the full height of the current region (before any usage).
	Full Abs
	// Backlog contains heights of additional regions.
	Backlog []Abs
	// Last is the height of the infinitely repeatable last region, if any.
	Last *Abs
	// Expand indicates whether content should expand to fill each axis.
	Expand Axes[bool]
}

// NewRegions creates new regions from a base size.
func NewRegions(size Size) *Regions {
	return &Regions{
		Size: size,
		Full: size.Height,
	}
}

// Width returns the current region's width.
func (r *Regions) Width() Abs {
	return r.Size.Width
}

// Height returns the current region's available height.
func (r *Regions) Height() Abs {
	return r.Size.Height
}

// CanBreak returns true if the layout can break to the next region.
func (r *Regions) CanBreak() bool {
	return len(r.Backlog) > 0 || r.Last != nil
}

// InLast returns true if we're in the infinitely repeatable last region.
func (r *Regions) InLast() bool {
	return len(r.Backlog) == 0 && r.Last != nil
}

// First returns the first region as a single Region.
func (r *Regions) First() Region {
	return Region{
		Size:   r.Size,
		Expand: r.Expand,
	}
}

// Iter returns an iterator over the regions.
func (r *Regions) Iter() *RegionsIter {
	return &RegionsIter{
		regions: r,
		index:   -1, // Start before first region
	}
}

// Next advances to the next region and returns whether there is one.
func (r *Regions) Next() bool {
	if len(r.Backlog) > 0 {
		// Move to next backlog region
		r.Size.Height = r.Backlog[0]
		r.Full = r.Backlog[0]
		r.Backlog = r.Backlog[1:]
		return true
	}
	if r.Last != nil {
		// Enter the repeatable last region
		r.Size.Height = *r.Last
		r.Full = *r.Last
		return true
	}
	return false
}

// Clone creates a copy of the regions.
func (r *Regions) Clone() *Regions {
	clone := &Regions{
		Size:    r.Size,
		Full:    r.Full,
		Expand:  r.Expand,
	}
	if len(r.Backlog) > 0 {
		clone.Backlog = make([]Abs, len(r.Backlog))
		copy(clone.Backlog, r.Backlog)
	}
	if r.Last != nil {
		last := *r.Last
		clone.Last = &last
	}
	return clone
}

// WithSize returns regions with a new current size.
func (r *Regions) WithSize(size Size) *Regions {
	clone := r.Clone()
	clone.Size = size
	return clone
}

// WithExpand returns regions with new expansion flags.
func (r *Regions) WithExpand(expand Axes[bool]) *Regions {
	clone := r.Clone()
	clone.Expand = expand
	return clone
}

// Shrink reduces all region sizes by the given inset.
func (r *Regions) Shrink(inset Sides[Abs]) *Regions {
	clone := &Regions{
		Size: Size{
			Width:  (r.Size.Width - SumHorizontal(inset)).Max(0),
			Height: (r.Size.Height - SumVertical(inset)).Max(0),
		},
		Full:   (r.Full - SumVertical(inset)).Max(0),
		Expand: r.Expand,
	}
	if len(r.Backlog) > 0 {
		clone.Backlog = make([]Abs, len(r.Backlog))
		for i, h := range r.Backlog {
			clone.Backlog[i] = (h - SumVertical(inset)).Max(0)
		}
	}
	if r.Last != nil {
		last := (*r.Last - SumVertical(inset)).Max(0)
		clone.Last = &last
	}
	return clone
}

// ShrinkMultiple applies insets to current size, full height, backlog, and last.
// This is used for breakable containers where insets apply to all regions.
func (r *Regions) ShrinkMultiple(inset Sides[Abs]) *Regions {
	return r.Shrink(inset)
}

// RegionsIter iterates over regions.
type RegionsIter struct {
	regions *Regions
	index   int
}

// Next advances to the next region.
func (it *RegionsIter) Next() (*Region, bool) {
	it.index++
	if it.index == 0 {
		// First region
		r := Region{
			Size:   it.regions.Size,
			Expand: it.regions.Expand,
		}
		return &r, true
	}

	backlogIdx := it.index - 1
	if backlogIdx < len(it.regions.Backlog) {
		r := Region{
			Size: Size{
				Width:  it.regions.Size.Width,
				Height: it.regions.Backlog[backlogIdx],
			},
			Expand: it.regions.Expand,
		}
		return &r, true
	}

	// Beyond backlog - check for repeatable last
	if it.regions.Last != nil {
		r := Region{
			Size: Size{
				Width:  it.regions.Size.Width,
				Height: *it.regions.Last,
			},
			Expand: it.regions.Expand,
		}
		return &r, true
	}

	return nil, false
}

// Infinite returns a height representing "infinite" space.
func Infinite() Abs {
	return Abs(math.Inf(1))
}
