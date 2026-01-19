package flow

import (
	"github.com/boergens/gotypst/layout"
)

// Stop represents a control flow signal during composition.
type Stop int

const (
	// StopFinish indicates layout is complete.
	StopFinish Stop = iota
	// StopRelayoutColumn indicates column needs to be relaid out.
	StopRelayoutColumn
	// StopRelayoutParent indicates parent (page) needs to be relaid out.
	StopRelayoutParent
)

// FlowResult represents the result of a flow layout operation.
type FlowResult struct {
	Frame *Frame
	Stop  Stop
	Err   error
}

// Ok creates a successful FlowResult.
func Ok(frame *Frame) FlowResult {
	return FlowResult{Frame: frame}
}

// Stopped creates a FlowResult with a stop signal.
func Stopped(stop Stop) FlowResult {
	return FlowResult{Stop: stop}
}

// Error creates a FlowResult with an error.
func Error(err error) FlowResult {
	return FlowResult{Err: err}
}

// Locator tracks element positions for introspection.
type Locator struct {
	// Internal state for generating unique locations.
	counter uint64
}

// NewLocator creates a new root locator.
func NewLocator() Locator {
	return Locator{counter: 0}
}

// Relayout creates a locator for relayout operations.
func (l Locator) Relayout() Locator {
	return l
}

// Split creates a splittable locator for sequential elements.
func (l Locator) Split() *SplitLocator {
	return &SplitLocator{base: l}
}

// Next generates the next location.
func (l *Locator) Next(hint interface{}) Locator {
	l.counter++
	return *l
}

// SplitLocator generates sequential locations.
type SplitLocator struct {
	base    Locator
	counter uint64
}

// Next generates the next location in sequence.
func (s *SplitLocator) Next(hint interface{}) Locator {
	s.counter++
	return Locator{counter: s.base.counter + s.counter}
}

// Synthesize creates a locator from an existing location.
func Synthesize(loc Location) Locator {
	return Locator{counter: loc.hash}
}

// Composer holds the state for page/column composition.
type Composer struct {
	// engine is the layout engine.
	engine interface{}
	// work holds the content being laid out.
	work *Work
	// config holds layout configuration.
	config *Config
	// column is the current column index.
	column int
	// pageBase is the base size for the page.
	pageBase layout.Size
	// pageInsertions collects page-level floats.
	pageInsertions *Insertions
	// columnInsertions collects column-level floats and footnotes.
	columnInsertions *Insertions
	// footnoteSpill holds footnote frames spilling to next column.
	footnoteSpill []*Frame
	// footnoteQueue holds footnotes waiting to be processed.
	footnoteQueue []interface{} // Packed<FootnoteElem>
}

// Compose composes contents of a single page/region with multiple columns.
// It handles layout of out-of-flow insertions (floats and footnotes) through
// per-page and per-column loops that rerun when floats are added.
func Compose(
	engine interface{},
	work *Work,
	config *Config,
	locator Locator,
	regions Regions,
) (*Frame, error) {
	c := &Composer{
		engine:           engine,
		config:           config,
		pageBase:         regions.Base(),
		column:           0,
		pageInsertions:   NewInsertions(),
		columnInsertions: NewInsertions(),
		work:             work,
		footnoteSpill:    nil,
		footnoteQueue:    nil,
	}
	return c.page(locator, regions)
}

// page lays out a container/page region including insertions.
func (c *Composer) page(locator Locator, regions Regions) (*Frame, error) {
	checkpoint := c.work.Clone()
	var output *Frame

	for {
		// Create working regions with space for page insertions
		pod := regions.Clone()
		pod.Size.Height -= c.pageInsertions.Height()

		result := c.pageContents(locator.Relayout(), pod)
		if result.Err != nil {
			return nil, result.Err
		}

		switch result.Stop {
		case StopRelayoutColumn:
			// Shouldn't happen at page level
			continue
		case StopRelayoutParent:
			// Reset and try again with updated insertions
			*c.work = checkpoint.Clone()
			continue
		default:
			output = result.Frame
		}
		break
	}

	return c.pageInsertions.Finalize(c.work, c.config, output), nil
}

// pageContents lays out inner contents of a container/page.
func (c *Composer) pageContents(locator Locator, regions Regions) FlowResult {
	// Single column case
	if c.config.Columns.Count == 1 {
		return c.layoutColumn(locator, regions)
	}

	// Multi-column layout
	columnHeight := regions.Size.Height
	backlog := make([]layout.Abs, 0, len(regions.Backlog)*c.config.Columns.Count)

	// Build backlog for all columns
	heights := append([]layout.Abs{columnHeight}, regions.Backlog...)
	for _, h := range heights {
		for i := 0; i < c.config.Columns.Count; i++ {
			if i > 0 || &h != &heights[0] {
				backlog = append(backlog, h)
			}
		}
	}

	inner := Regions{
		Size:    layout.Size{Width: c.config.Columns.Width, Height: columnHeight},
		Backlog: backlog,
		Expand:  Axes[bool]{X: true, Y: regions.Expand.Y},
		Full:    regions.Full,
		Last:    regions.Last,
	}

	// Determine output size
	outputHeight := layout.Abs(0)
	if regions.Expand.Y {
		outputHeight = regions.Size.Height
	}
	size := layout.Size{Width: regions.Size.Width, Height: outputHeight}

	output := NewHardFrame(size)
	offset := layout.Abs(0)
	splitLocator := locator.Split()

	for i := 0; i < c.config.Columns.Count; i++ {
		c.column = i
		result := c.layoutColumn(splitLocator.Next(nil), inner)
		if result.Err != nil {
			return result
		}
		if result.Stop != 0 {
			return result
		}

		frame := result.Frame
		if !regions.Expand.Y {
			if frame.Height() > output.Height() {
				output.SizeMut().Height = frame.Height()
			}
		}

		width := frame.Width()
		var x layout.Abs
		if c.config.Columns.Dir == layout.DirLTR {
			x = offset
		} else {
			x = regions.Size.Width - offset - width
		}
		offset += width + c.config.Columns.Gutter

		output.PushFrame(layout.Point{X: x, Y: 0}, frame)
		inner.Next()
	}

	return Ok(output)
}

// layoutColumn lays out a column including column insertions.
func (c *Composer) layoutColumn(locator Locator, regions Regions) FlowResult {
	c.columnInsertions.Reset()

	// Handle footnote spillover from previous column
	if len(c.work.FootnoteSpill) > 0 {
		spill := c.work.FootnoteSpill
		c.work.FootnoteSpill = nil
		if err := c.footnoteSpillover(spill, regions.Base()); err != nil {
			return Error(err)
		}
	}

	checkpoint := c.work.Clone()
	var inner *Frame

	for {
		// Create working regions with space for column insertions
		pod := regions.Clone()
		pod.Size.Height -= c.columnInsertions.Height()

		result := c.columnContents(pod)
		if result.Err != nil {
			return result
		}

		switch result.Stop {
		case StopRelayoutColumn:
			*c.work = checkpoint.Clone()
			continue
		case StopRelayoutParent:
			return result
		default:
			inner = result.Frame
		}
		break
	}

	// Transfer footnote queue back to work
	c.work.Footnotes = append(c.work.Footnotes, c.footnoteQueue...)
	c.footnoteQueue = c.footnoteQueue[:0]
	if len(c.footnoteSpill) > 0 {
		c.work.FootnoteSpill = c.footnoteSpill
		c.footnoteSpill = nil
	}

	// Finalize column with insertions
	output := c.columnInsertions.Finalize(c.work, c.config, inner)

	// Add line numbers if configured
	if c.config.LineNumbers != nil {
		if err := c.layoutLineNumbers(locator, output); err != nil {
			return Error(err)
		}
	}

	return Ok(output)
}

// columnContents lays out inner contents of a column.
func (c *Composer) columnContents(regions Regions) FlowResult {
	// Process pending footnotes
	for _, note := range c.work.Footnotes {
		regionsCopy := regions.Clone()
		result := c.processFootnote(note, &regionsCopy, 0, false)
		if result.Err != nil {
			return result
		}
	}
	c.work.Footnotes = c.work.Footnotes[:0]

	// Process pending floats
	for _, placed := range c.work.Floats {
		result := c.processFloat(placed, &regions, false, false)
		if result.Err != nil {
			return result
		}
	}
	c.work.Floats = c.work.Floats[:0]

	// Distribute content
	return c.distribute(regions)
}

// distribute distributes content into the available regions.
// This is a placeholder - the actual implementation would be in distribute.go
func (c *Composer) distribute(regions Regions) FlowResult {
	// Create an empty frame with the region size
	frame := NewSoftFrame(regions.Size)
	return Ok(frame)
}

// processFloat handles a floating element.
func (c *Composer) processFloat(
	placed *PlacedChild,
	regions *Regions,
	clearance bool,
	migratable bool,
) FlowResult {
	loc := placed.Location()
	if c.skipped(loc) {
		return Ok(nil)
	}

	// If there are queued floats, queue this one too
	if len(c.work.Floats) > 0 {
		c.work.Floats = append(c.work.Floats, placed)
		return Ok(nil)
	}

	// Determine base size for layout
	var base layout.Size
	switch placed.Scope {
	case PlacementScopeColumn:
		base = regions.Base()
	case PlacementScopeParent:
		base = c.pageBase
	}

	// Layout the float content
	frame, err := placed.Layout(c.engine, base)
	if err != nil {
		return Error(err)
	}

	// Calculate remaining space
	var remaining layout.Abs
	switch placed.Scope {
	case PlacementScopeColumn:
		remaining = regions.Size.Height
	case PlacementScopeParent:
		var total layout.Abs
		for _, size := range regions.Iter() {
			total += size.Height
		}
		columnsRemaining := c.config.Columns.Count - c.column
		if columnsRemaining > 0 {
			remaining = total / layout.Abs(columnsRemaining)
		}
	}

	// Check if float fits
	clearanceAmt := layout.Abs(0)
	if clearance {
		clearanceAmt = placed.Clearance
	}
	need := frame.Height() + clearanceAmt

	if !remaining.Fits(need) && regions.MayProgress() {
		// Float doesn't fit, queue for later
		c.work.Floats = append(c.work.Floats, placed)
		return Ok(nil)
	}

	// Determine vertical alignment
	alignY := FixedAlignmentStart
	if placed.AlignY != nil {
		alignY = *placed.AlignY
	} else {
		// Auto-determine based on position
		used := base.Height - remaining
		half := need / 2.0
		ratio := float64(used+half) / float64(base.Height)
		if ratio > 0.5 {
			alignY = FixedAlignmentEnd
		}
	}

	// Add to appropriate insertions
	var area *Insertions
	switch placed.Scope {
	case PlacementScopeColumn:
		area = c.columnInsertions
	case PlacementScopeParent:
		area = c.pageInsertions
	}

	area.PushFloat(placed, frame, alignY)
	area.AddSkip(loc)

	// Signal relayout needed
	switch placed.Scope {
	case PlacementScopeColumn:
		return Stopped(StopRelayoutColumn)
	default:
		return Stopped(StopRelayoutParent)
	}
}

// processFootnote handles a footnote element.
func (c *Composer) processFootnote(
	elem interface{},
	regions *Regions,
	flowNeed layout.Abs,
	migratable bool,
) FlowResult {
	// Placeholder implementation
	return Ok(nil)
}

// footnoteSpillover handles footnote frames that spill to the next column.
func (c *Composer) footnoteSpillover(frames []*Frame, base layout.Size) error {
	if len(frames) == 0 {
		return nil
	}

	// Add separator
	separator, err := c.layoutFootnoteSeparator(base)
	if err != nil {
		return err
	}
	c.columnInsertions.PushFootnoteSeparator(c.config, separator)

	// Add first spill frame
	frame := frames[0]
	c.columnInsertions.PushFootnote(c.config, frame)

	// Queue remaining frames for next column
	if len(frames) > 1 {
		c.footnoteSpill = frames[1:]
	}

	return nil
}

// layoutFootnoteSeparator creates the footnote separator frame.
func (c *Composer) layoutFootnoteSeparator(base layout.Size) (*Frame, error) {
	// Placeholder - would layout the separator content
	return NewSoftFrame(layout.Size{Width: base.Width, Height: 1}), nil
}

// layoutLineNumbers adds line numbers to the output frame.
func (c *Composer) layoutLineNumbers(locator Locator, output *Frame) error {
	// Placeholder implementation
	return nil
}

// skipped checks if a location has already been processed.
func (c *Composer) skipped(loc Location) bool {
	if _, ok := c.work.Skips[loc]; ok {
		return true
	}
	for _, skip := range c.pageInsertions.Skips() {
		if skip == loc {
			return true
		}
	}
	for _, skip := range c.columnInsertions.Skips() {
		if skip == loc {
			return true
		}
	}
	return false
}

// InsertionWidth returns the width needed by insertions.
func (c *Composer) InsertionWidth() layout.Abs {
	return maxAbs(c.columnInsertions.Width(), c.pageInsertions.Width())
}
