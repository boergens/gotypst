package flow

import (
	"github.com/boergens/gotypst/layout"
)

// Distribute distributes as many children as fit from composer.work into the
// first region and returns the resulting frame.
func Distribute(composer *Composer, regions Regions) (Frame, Stop) {
	d := &Distributor{
		composer:  composer,
		regions:   regions,
		items:     nil,
		sticky:    nil,
		stickable: nil,
	}
	init := d.snapshot()

	var forced bool
	if stop := d.run(); stop != nil {
		if finish, ok := stop.(StopFinish); ok {
			forced = finish.Forced
		} else {
			return Frame{}, stop
		}
	} else {
		forced = d.composer.Work.Done()
	}

	region := NewRegion(regions.Size, regions.Expand)
	return d.finalize(region, init, forced)
}

// Distributor holds state for distribution.
type Distributor struct {
	// composer handles insertions.
	composer *Composer
	// regions are continuously shrunk as new items are added.
	regions Regions
	// items are already laid out items, not yet aligned.
	items []Item
	// sticky is a snapshot that can be restored to migrate sticky blocks.
	sticky *distributionSnapshot
	// stickable tracks whether current sticky blocks can migrate.
	// nil means not processing sticky blocks.
	// true means migration is allowed.
	// false means stickiness is disabled for this group.
	stickable *bool
}

// distributionSnapshot captures the distribution state for restoration.
type distributionSnapshot struct {
	work  Work
	items int
}

// Item represents a laid out item in a distribution.
type Item interface {
	isItem()
	// Migratable returns true if this item should migrate to the next region
	// if the region consists solely of such items.
	Migratable() bool
}

// TagItem represents an introspection tag.
type TagItem struct {
	Tag *Tag
}

func (TagItem) isItem() {}

// Migratable returns true - tags are always migratable.
func (TagItem) Migratable() bool { return true }

// AbsItem represents absolute spacing and its weakness level.
type AbsItem struct {
	Amount   layout.Abs
	Weakness uint8
}

func (AbsItem) isItem() {}

// Migratable returns false - spacing is not migratable.
func (AbsItem) Migratable() bool { return false }

// FrItem represents fractional spacing or a fractional block.
type FrItem struct {
	Amount   layout.Fr
	Weakness uint8
	Single   *SingleChild // nil for pure spacing
}

func (FrItem) isItem() {}

// Migratable returns false - fractional items are not migratable.
func (FrItem) Migratable() bool { return false }

// FrameItem represents a frame for a laid out line or block.
type FlowFrameItem struct {
	Frame Frame
	Align Axes[FixedAlignment]
}

func (FlowFrameItem) isItem() {}

// Migratable returns true if the frame is empty with only links/tags.
func (f FlowFrameItem) Migratable() bool {
	if f.Frame.Size().Width != 0 || f.Frame.Size().Height != 0 {
		return false
	}
	for _, entry := range f.Frame.Items() {
		switch entry.Item.(type) {
		case FrameItemLink, FrameItemTag:
			// These are fine
		default:
			return false
		}
	}
	return true
}

// PlacedItem represents a frame for an absolutely placed child.
type PlacedItem struct {
	Frame  Frame
	Placed *PlacedChild
}

func (PlacedItem) isItem() {}

// Migratable returns true if the placed child is not floating.
func (p PlacedItem) Migratable() bool {
	return !p.Placed.Float
}

// run distributes content into the region.
func (d *Distributor) run() Stop {
	// First, handle spill of a breakable block.
	if spill := d.composer.Work.Spill; spill != nil {
		d.composer.Work.Spill = nil
		if stop := d.multiSpill(spill); stop != nil {
			return stop
		}
	}

	// Process children until no space or children are left.
	for {
		child := d.composer.Work.Head()
		if child == nil {
			break
		}
		if stop := d.child(child); stop != nil {
			return stop
		}
		d.composer.Work.Advance()
	}

	return nil
}

// child processes a single child.
func (d *Distributor) child(child Child) Stop {
	switch c := child.(type) {
	case TagChild:
		d.tag(c.Tag)
	case RelChild:
		d.rel(c.Amount, c.Weakness)
	case FrChild:
		d.fr(c.Amount, c.Weakness)
	case *LineChild:
		return d.line(c)
	case *SingleChild:
		return d.single(c)
	case *MultiChild:
		return d.multi(c)
	case *PlacedChild:
		return d.placed(c)
	case FlushChild:
		return d.flush()
	case BreakChild:
		return d.break_(c.Weak)
	}
	return nil
}

// tag processes a tag.
func (d *Distributor) tag(tag *Tag) {
	d.composer.Work.Tags = append(d.composer.Work.Tags, tag)
}

// flushTags generates items for pending tags.
func (d *Distributor) flushTags() {
	if len(d.composer.Work.Tags) == 0 {
		return
	}
	for _, tag := range d.composer.Work.Tags {
		d.items = append(d.items, TagItem{Tag: tag})
	}
	d.composer.Work.Tags = nil
}

// rel processes relative spacing.
func (d *Distributor) rel(amount Rel, weakness uint8) {
	resolved := amount.RelativeTo(d.regions.Base().Height)
	if weakness > 0 && !d.keepWeakRelSpacing(resolved, weakness) {
		return
	}

	d.regions.Size.Height -= resolved
	d.items = append(d.items, AbsItem{Amount: resolved, Weakness: weakness})
}

// fr processes fractional spacing.
func (d *Distributor) fr(fr layout.Fr, weakness uint8) {
	if weakness > 0 && !d.keepWeakFrSpacing(fr, weakness) {
		return
	}

	// If we decided to keep the fr spacing, trim previous spacing.
	d.trimSpacing()

	d.items = append(d.items, FrItem{Amount: fr, Weakness: weakness})
}

// keepWeakRelSpacing decides whether to keep weak spacing based on previous items.
func (d *Distributor) keepWeakRelSpacing(amount layout.Abs, weakness uint8) bool {
	for i := len(d.items) - 1; i >= 0; i-- {
		switch item := d.items[i].(type) {
		case AbsItem:
			if item.Weakness >= 1 {
				// Previous weak relative spacing exists.
				if weakness <= item.Weakness &&
					(weakness < item.Weakness || amount > item.Amount) {
					d.regions.Size.Height -= amount - item.Amount
					d.items[i] = AbsItem{Amount: amount, Weakness: weakness}
				}
				return false
			}
			// Strong abs spacing - peek beyond
		case TagItem, PlacedItem:
			// Peek beyond these
		case FrItem:
			if item.Single == nil {
				// Fractional spacing destructs weak relative spacing.
				return false
			}
			// Fractional block supports spacing.
			return true
		case FlowFrameItem:
			return true
		}
	}
	return false
}

// keepWeakFrSpacing decides whether to keep weak fractional spacing.
func (d *Distributor) keepWeakFrSpacing(fr layout.Fr, weakness uint8) bool {
	for i := len(d.items) - 1; i >= 0; i-- {
		switch item := d.items[i].(type) {
		case FrItem:
			if item.Weakness >= 1 && item.Single == nil {
				// Previous weak fr spacing exists.
				if weakness <= item.Weakness &&
					(weakness < item.Weakness || fr > item.Amount) {
					d.items[i] = FrItem{Amount: fr, Weakness: weakness}
				}
				return false
			}
			if item.Single == nil {
				// Strong fr spacing - keep both.
				return true
			}
			// Fractional block supports spacing.
			return true
		case TagItem, AbsItem, PlacedItem:
			// Peek beyond these
		case FlowFrameItem:
			return true
		}
	}
	return false
}

// trimSpacing trims trailing weak spacing from items.
func (d *Distributor) trimSpacing() {
	for i := len(d.items) - 1; i >= 0; i-- {
		switch item := d.items[i].(type) {
		case AbsItem:
			if item.Weakness >= 1 {
				d.regions.Size.Height += item.Amount
				d.items = append(d.items[:i], d.items[i+1:]...)
				return
			}
		case FrItem:
			if item.Weakness >= 1 && item.Single == nil {
				d.items = append(d.items[:i], d.items[i+1:]...)
				return
			}
		case TagItem, PlacedItem:
			// Continue searching
		case FlowFrameItem:
			return
		}
	}
}

// weakSpacing returns the amount of trailing weak spacing.
func (d *Distributor) weakSpacing() layout.Abs {
	for i := len(d.items) - 1; i >= 0; i-- {
		switch item := d.items[i].(type) {
		case AbsItem:
			if item.Weakness >= 1 {
				return item.Amount
			}
		case TagItem, PlacedItem:
			// Continue searching
		case FlowFrameItem, FrItem:
			return 0
		}
	}
	return 0
}

// line processes a line of a paragraph.
func (d *Distributor) line(line *LineChild) Stop {
	// If the line doesn't fit and a followup region may improve things,
	// finish the region.
	if !d.regions.Size.Height.Fits(line.Frame.Height()) && d.regions.MayProgress() {
		return StopFinish{Forced: false}
	}

	// If the line's need doesn't fit but does fit in the next region,
	// finish the region.
	if !d.regions.Size.Height.Fits(line.Need) {
		iter := d.regions.Iter()
		if len(iter) > 1 && iter[1].Fits(line.Need) {
			return StopFinish{Forced: false}
		}
	}

	return d.frame(line.Frame.Clone(), line.Align, false, false)
}

// single processes an unbreakable block.
func (d *Distributor) single(single *SingleChild) Stop {
	// Lay out the block.
	frame, err := single.Layout(d.composer.Engine, NewRegion(d.regions.Base(), d.regions.Expand))
	if err != nil {
		return StopError{Err: err}
	}

	// Handle fractionally sized blocks.
	if single.Fr != nil {
		if err := d.composer.Footnotes(&d.regions, &frame, 0, false, true); err != nil {
			return StopError{Err: err}
		}
		d.flushTags()
		d.items = append(d.items, FrItem{Amount: *single.Fr, Weakness: 0, Single: single})
		return nil
	}

	// If the block doesn't fit and a followup region may improve things,
	// finish the region.
	if !d.regions.Size.Height.Fits(frame.Height()) && d.regions.MayProgress() {
		return StopFinish{Forced: false}
	}

	return d.frame(frame, single.Align, single.Sticky, false)
}

// multi processes a breakable block.
func (d *Distributor) multi(multi *MultiChild) Stop {
	// Skip directly if the region is already (over)full.
	if d.regions.IsFull() {
		return StopFinish{Forced: false}
	}

	// Lay out the block.
	frame, spill, err := multi.Layout(d.composer.Engine, d.regions)
	if err != nil {
		return StopError{Err: err}
	}

	if frame.IsEmpty() && spill != nil && spill.ExistNonEmptyFrame && d.regions.MayProgress() {
		// If the first frame is empty but there are non-empty frames in spill,
		// move the whole child to the next region.
		return StopFinish{Forced: false}
	}

	if stop := d.frame(frame, multi.Align, multi.Sticky, true); stop != nil {
		return stop
	}

	// If the block didn't fully fit, save spill and finish region.
	if spill != nil {
		d.composer.Work.Spill = spill
		d.composer.Work.Advance()
		return StopFinish{Forced: false}
	}

	return nil
}

// multiSpill processes spillover from a breakable block.
func (d *Distributor) multiSpill(spill *MultiSpill) Stop {
	// Skip directly if the region is already (over)full.
	if d.regions.IsFull() {
		d.composer.Work.Spill = spill
		return StopFinish{Forced: false}
	}

	// Lay out the spilled remains.
	align := spill.Align()
	frame, nextSpill, err := spill.Layout(d.composer.Engine, d.regions)
	if err != nil {
		return StopError{Err: err}
	}

	if stop := d.frame(frame, align, false, true); stop != nil {
		return stop
	}

	// If there's still more, save it and finish region.
	if nextSpill != nil {
		d.composer.Work.Spill = nextSpill
		return StopFinish{Forced: false}
	}

	return nil
}

// frame processes an in-flow frame from a line or block.
func (d *Distributor) frame(
	frame Frame,
	align Axes[FixedAlignment],
	sticky bool,
	breakable bool,
) Stop {
	if sticky {
		// Make checkpoint for sticky blocks if needed.
		if d.sticky == nil {
			mayProgress := d.regions.MayProgress()
			if d.stickable == nil {
				d.stickable = &mayProgress
			}
			if *d.stickable {
				snapshot := d.snapshot()
				d.sticky = &snapshot
			}
		}
	} else if !frame.IsEmpty() {
		// Non-sticky, non-empty frame - forget previous snapshot.
		d.sticky = nil
		d.stickable = nil
	}

	// Handle footnotes.
	if err := d.composer.Footnotes(&d.regions, &frame, frame.Height(), breakable, true); err != nil {
		return StopError{Err: err}
	}

	// Push item for the frame.
	d.regions.Size.Height -= frame.Height()
	d.flushTags()
	d.items = append(d.items, FlowFrameItem{Frame: frame, Align: align})
	return nil
}

// placed processes an absolutely or floatingly placed child.
func (d *Distributor) placed(placed *PlacedChild) Stop {
	if placed.Float {
		// Let the composer handle floating elements.
		weakSpacing := d.weakSpacing()
		d.regions.Size.Height += weakSpacing

		hasFrames := false
		for _, item := range d.items {
			if _, ok := item.(FlowFrameItem); ok {
				hasFrames = true
				break
			}
		}

		if err := d.composer.Float(placed, &d.regions, hasFrames, true); err != nil {
			d.regions.Size.Height -= weakSpacing
			return StopError{Err: err}
		}
		d.regions.Size.Height -= weakSpacing
	} else {
		frame, err := placed.Layout(d.composer.Engine, d.regions.Base())
		if err != nil {
			return StopError{Err: err}
		}
		if err := d.composer.Footnotes(&d.regions, &frame, 0, true, true); err != nil {
			return StopError{Err: err}
		}
		d.flushTags()
		d.items = append(d.items, PlacedItem{Frame: frame, Placed: placed})
	}
	return nil
}

// flush processes a float flush.
func (d *Distributor) flush() Stop {
	// If there are still pending floats, finish the region.
	if len(d.composer.Work.Floats) > 0 {
		return StopFinish{Forced: false}
	}
	return nil
}

// break_ processes a column break.
func (d *Distributor) break_(weak bool) Stop {
	// If there is a region to break into, break into it.
	if (!weak || len(d.items) > 0) &&
		(len(d.regions.Backlog) > 0 || d.regions.Last != nil) {
		d.composer.Work.Advance()
		return StopFinish{Forced: true}
	}
	return nil
}

// finalize arranges the produced items into an output frame.
func (d *Distributor) finalize(
	region Region,
	init distributionSnapshot,
	forced bool,
) (Frame, Stop) {
	if forced {
		// End of flow - flush pending tags.
		d.flushTags()
	} else if len(d.items) > 0 && d.allMigratable() {
		// Restore initial state if all items are migratable.
		d.restore(init)
	} else if d.sticky != nil {
		// Restore sticky snapshot to move suffix to next region.
		d.restore(*d.sticky)
	}

	d.trimSpacing()

	var frs layout.Fr
	var used layout.Size
	hasFrChild := false

	// Determine used space and sum of fractionals.
	for _, item := range d.items {
		switch it := item.(type) {
		case AbsItem:
			used.Height += it.Amount
		case FrItem:
			frs += it.Amount
			hasFrChild = hasFrChild || it.Single != nil
		case FlowFrameItem:
			used.Height += it.Frame.Height()
			if it.Frame.Width() > used.Width {
				used.Width = it.Frame.Width()
			}
		case TagItem, PlacedItem:
			// No contribution to used space
		}
	}

	// When we have fractional spacing, occupy remaining space.
	var frSpace layout.Abs
	if frs > 0 && region.Size.Height > 0 {
		frSpace = region.Size.Height - used.Height
		used.Height = region.Size.Height
	}

	// Lay out fractionally sized blocks.
	var frFrames []Frame
	if hasFrChild {
		for _, item := range d.items {
			frItem, ok := item.(FrItem)
			if !ok || frItem.Single == nil {
				continue
			}
			length := share(frItem.Amount, frs, frSpace)
			pod := NewRegion(layout.Size{Width: region.Size.Width, Height: length}, region.Expand)
			frame, err := frItem.Single.Layout(d.composer.Engine, pod)
			if err != nil {
				return Frame{}, StopError{Err: err}
			}
			if frame.Width() > used.Width {
				used.Width = frame.Width()
			}
			frFrames = append(frFrames, frame)
		}
	}

	// Consider insertion width for alignment.
	if !region.Expand.X {
		insertionWidth := d.composer.InsertionWidth()
		if insertionWidth > used.Width {
			used.Width = insertionWidth
		}
	}

	// Determine region's size.
	size := selectSize(region.Expand, region.Size, minSize(used, region.Size))
	free := size.Height - used.Height

	output := Soft(size)
	ruler := FixedAlignStart
	var offset layout.Abs
	frFrameIdx := 0

	// Position all items.
	for _, item := range d.items {
		switch it := item.(type) {
		case TagItem:
			y := offset + ruler.Position(free)
			pos := layout.Point{X: 0, Y: y}
			output.Push(pos, FrameItemTag{Tag: it.Tag.Clone()})

		case AbsItem:
			offset += it.Amount

		case FrItem:
			length := share(it.Amount, frs, frSpace)
			if it.Single != nil {
				frame := frFrames[frFrameIdx]
				frFrameIdx++
				x := it.Single.Align.X.Position(size.Width - frame.Width())
				pos := layout.Point{X: x, Y: offset}
				output.PushFrame(pos, frame)
			}
			offset += length

		case FlowFrameItem:
			ruler = ruler.Max(it.Align.Y)
			x := it.Align.X.Position(size.Width - it.Frame.Width())
			y := offset + ruler.Position(free)
			pos := layout.Point{X: x, Y: y}
			offset += it.Frame.Height()
			output.PushFrame(pos, it.Frame)

		case PlacedItem:
			x := it.Placed.AlignX.Position(size.Width - it.Frame.Width())
			var y layout.Abs
			if it.Placed.AlignY != nil {
				y = it.Placed.AlignY.Position(size.Height - it.Frame.Height())
			} else {
				y = offset + ruler.Position(free)
			}
			delta := RelAxesToPoint(it.Placed.Delta, size)
			pos := layout.Point{X: x + delta.X, Y: y + delta.Y}
			output.PushFrame(pos, it.Frame)
		}
	}

	return output, nil
}

// snapshot creates a snapshot of the work and items.
func (d *Distributor) snapshot() distributionSnapshot {
	return distributionSnapshot{
		work:  d.composer.Work.Clone(),
		items: len(d.items),
	}
}

// restore restores a snapshot of the work and items.
func (d *Distributor) restore(snapshot distributionSnapshot) {
	*d.composer.Work = snapshot.work
	d.items = d.items[:snapshot.items]
}

// allMigratable returns true if all items are migratable.
func (d *Distributor) allMigratable() bool {
	for _, item := range d.items {
		if !item.Migratable() {
			return false
		}
	}
	return true
}

// share calculates the share of space for a fractional value.
func share(fr, total layout.Fr, space layout.Abs) layout.Abs {
	if total <= 0 {
		return 0
	}
	return layout.Abs(float64(fr) / float64(total) * float64(space))
}

// selectSize selects the size based on expansion settings.
func selectSize(expand Axes[bool], full, used layout.Size) layout.Size {
	result := used
	if expand.X {
		result.Width = full.Width
	}
	if expand.Y {
		result.Height = full.Height
	}
	return result
}

// minSize returns the component-wise minimum of two sizes.
func minSize(a, b layout.Size) layout.Size {
	result := a
	if b.Width < result.Width {
		result.Width = b.Width
	}
	if b.Height < result.Height {
		result.Height = b.Height
	}
	return result
}
