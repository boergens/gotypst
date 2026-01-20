// Package inline provides inline/paragraph layout including text shaping.
package inline

import (
	"math"

	"github.com/boergens/gotypst/layout"
)

// Fragment represents a collection of frames from layout.
type Fragment struct {
	Frames []*FinalFrame
}

// NewFragment creates a Fragment from frames.
func NewFragment(frames []*FinalFrame) Fragment {
	return Fragment{Frames: frames}
}

// FinalSize represents 2D dimensions for finalization.
type FinalSize struct {
	Width, Height Abs
}

// FinalPoint represents a 2D coordinate for finalization.
type FinalPoint struct {
	X, Y Abs
}

// FinalFrame represents a layout frame.
type FinalFrame struct {
	Size     FinalSize
	Baseline Abs
	Items    []FinalFrameEntry
}

// FinalFrameEntry is a positioned item in a frame.
type FinalFrameEntry struct {
	Pos  FinalPoint
	Item FinalFrameItem
}

// FinalFrameItem represents an item that can be placed in a frame.
type FinalFrameItem interface {
	isFinalFrameItem()
}

// FinalTextItem represents text in a frame.
type FinalTextItem struct {
	Text *ShapedText
}

func (FinalTextItem) isFinalFrameItem() {}

// Push adds an item to the frame.
func (f *FinalFrame) Push(pos FinalPoint, item FinalFrameItem) {
	f.Items = append(f.Items, FinalFrameEntry{Pos: pos, Item: item})
}

// PushFrame adds a child frame at the given position.
func (f *FinalFrame) PushFrame(pos FinalPoint, child *FinalFrame) {
	// Merge the child's items into the parent with adjusted positions
	for _, entry := range child.Items {
		f.Items = append(f.Items, FinalFrameEntry{
			Pos:  FinalPoint{X: pos.X + entry.Pos.X, Y: pos.Y + entry.Pos.Y},
			Item: entry.Item,
		})
	}
}

// Width returns the frame's width.
func (f *FinalFrame) Width() Abs {
	return f.Size.Width
}

// Height returns the frame's height.
func (f *FinalFrame) Height() Abs {
	return f.Size.Height
}

// Finalize turns the selected lines into frames.
func Finalize(
	p *Preparation,
	lines []*Line,
	region FinalSize,
	expand bool,
) (Fragment, error) {
	// Determine the resulting width: Full width of the region if we should
	// expand or there's fractional spacing, fit-to-width otherwise.
	width := region.Width

	if !math.IsInf(float64(region.Width), 0) {
		allZeroFr := true
		maxLineWidth := Abs(0)

		for _, line := range lines {
			if line.Fr() != 0 {
				allZeroFr = false
			}
			if line.Width > maxLineWidth {
				maxLineWidth = line.Width
			}
		}

		if !expand && allZeroFr {
			fitWidth := p.Config.HangingIndent + maxLineWidth
			if fitWidth < region.Width {
				width = fitWidth
			}
		}
	}

	// Stack the lines into frames.
	frames := make([]*FinalFrame, 0, len(lines))
	for _, line := range lines {
		frame, err := Commit(p, line, width, region.Height)
		if err != nil {
			return Fragment{}, err
		}
		frames = append(frames, frame)
	}

	return NewFragment(frames), nil
}

// Commit builds a frame for a single line.
func Commit(
	p *Preparation,
	line *Line,
	width Abs,
	fullHeight Abs,
) (*FinalFrame, error) {
	remaining := width - line.Width - p.Config.HangingIndent
	offset := Abs(0)

	// In LTR, add hanging indent to offset. In RTL, it arises naturally.
	if p.Config.Dir == DirLTR {
		offset += p.Config.HangingIndent
	}

	// Handle hanging punctuation to the left
	if leadingText := line.LeadingText(); leadingText != nil {
		if glyphs := leadingText.Glyphs.Kept(); len(glyphs) > 0 {
			glyph := &glyphs[0]
			if !leadingText.Dir.IsPositive() && (len(line.Items) > 1 || len(glyphs) > 1) {
				amount := overhang(glyph.Char) * glyph.XAdvance.At(glyph.Size)
				offset -= amount
				remaining += amount
			}
		}
	}

	// Handle hanging punctuation to the right
	if trailingText := line.TrailingText(); trailingText != nil {
		if glyphs := trailingText.Glyphs.Kept(); len(glyphs) > 0 {
			glyph := &glyphs[len(glyphs)-1]
			if trailingText.Dir.IsPositive() && (len(line.Items) > 1 || len(glyphs) > 1) {
				amount := overhang(glyph.Char) * glyph.XAdvance.At(glyph.Size)
				remaining += amount
			}
		}
	}

	// Calculate justification parameters
	fr := line.Fr()
	justificationRatio := 0.0
	extraJustification := Abs(0)

	shrinkability := line.Shrinkability()
	stretchability := line.Stretchability()

	if remaining < 0 && shrinkability > 0 {
		// Attempt to reduce line length using shrinkability
		ratio := float64(remaining / shrinkability)
		if ratio < -1.0 {
			ratio = -1.0
		}
		justificationRatio = ratio
		adjusted := remaining + shrinkability
		if adjusted > 0 {
			adjusted = 0
		}
		remaining = adjusted
	} else if line.Justify && fr == 0 {
		// Attempt to increase line length using stretchability
		if stretchability > 0 {
			ratio := float64(remaining / stretchability)
			if ratio > 1.0 {
				ratio = 1.0
			}
			justificationRatio = ratio
			adjusted := remaining - stretchability
			if adjusted < 0 {
				adjusted = 0
			}
			remaining = adjusted
		}

		justifiables := line.Justifiables()
		if justifiables > 0 && remaining > 0 {
			// Underfull line: distribute extra space
			extraJustification = remaining / Abs(justifiables)
			remaining = 0
		}
	}

	// Build frames and determine height/baseline
	var top, bottom Abs
	type positionedFrame struct {
		offset Abs
		frame  *FinalFrame
		idx    int // For stable sorting
	}
	var posFrames []positionedFrame

	for idx, item := range line.Items {
		switch it := item.(type) {
		case *AbsoluteItem:
			offset += it.Amount

		case *FractionalItem:
			amount := frShare(it.Amount, fr, remaining)
			offset += amount

		case *TextItem:
			if it.shaped == nil {
				continue
			}
			frame := buildTextFrame(it.shaped, justificationRatio, extraJustification)
			if frame.Baseline > top {
				top = frame.Baseline
			}
			if frame.Size.Height-frame.Baseline > bottom {
				bottom = frame.Size.Height - frame.Baseline
			}
			posFrames = append(posFrames, positionedFrame{offset, frame, idx})
			offset += frame.Size.Width

		case *InlineFrameItem:
			frame := &FinalFrame{Size: FinalSize{Width: it.width, Height: 0}}
			posFrames = append(posFrames, positionedFrame{offset, frame, idx})
			offset += it.width

		case *TagItem:
			// Tags are zero-width
			frame := &FinalFrame{Size: FinalSize{}}
			posFrames = append(posFrames, positionedFrame{offset, frame, idx})

		case *SkipItem:
			// Skip items are invisible

		case *MathRootItem:
			// Build frame for math root
			ctx := &MathContext{
				Config:   DefaultMathLayoutConfig(p.Config.FontSize),
				FontSize: p.Config.FontSize,
			}
			frame := BuildMathRootFrame(it, ctx)
			if frame.Baseline > top {
				top = frame.Baseline
			}
			if frame.Size.Height-frame.Baseline > bottom {
				bottom = frame.Size.Height - frame.Baseline
			}
			posFrames = append(posFrames, positionedFrame{offset, frame, idx})
			offset += frame.Size.Width

		case *MathFracItem:
			// Build frame for math fraction
			ctx := &MathContext{
				Config:   DefaultMathLayoutConfig(p.Config.FontSize),
				FontSize: p.Config.FontSize,
			}
			frame := BuildMathFracFrame(it, ctx)
			if frame.Baseline > top {
				top = frame.Baseline
			}
			if frame.Size.Height-frame.Baseline > bottom {
				bottom = frame.Size.Height - frame.Baseline
			}
			posFrames = append(posFrames, positionedFrame{offset, frame, idx})
			offset += frame.Size.Width

		case *MathAttachItem:
			// Build frame for math attach (subscripts/superscripts)
			ctx := &MathContext{
				Config:   DefaultMathLayoutConfig(p.Config.FontSize),
				FontSize: p.Config.FontSize,
			}
			frame := BuildMathAttachFrame(it, ctx)
			if frame.Baseline > top {
				top = frame.Baseline
			}
			if frame.Size.Height-frame.Baseline > bottom {
				bottom = frame.Size.Height - frame.Baseline
			}
			posFrames = append(posFrames, positionedFrame{offset, frame, idx})
			offset += frame.Size.Width

		case *MathDelimitedItem:
			// Build frame for delimited math
			ctx := &MathContext{
				Config:   DefaultMathLayoutConfig(p.Config.FontSize),
				FontSize: p.Config.FontSize,
			}
			frame := BuildMathDelimitedFrame(it, ctx)
			if frame.Baseline > top {
				top = frame.Baseline
			}
			if frame.Size.Height-frame.Baseline > bottom {
				bottom = frame.Size.Height - frame.Baseline
			}
			posFrames = append(posFrames, positionedFrame{offset, frame, idx})
			offset += frame.Size.Width
		}
	}

	// Handle remaining fractional space
	if fr != 0 {
		remaining = 0
	}

	// Create output frame
	size := FinalSize{Width: width, Height: top + bottom}
	output := &FinalFrame{
		Size:     size,
		Baseline: top,
	}

	// Calculate alignment offset
	alignOffset := alignPosition(p.Config.Align, remaining)

	// Add positioned frames to output
	for _, pf := range posFrames {
		x := pf.offset + alignOffset
		y := top - pf.frame.Baseline
		output.PushFrame(FinalPoint{X: x, Y: y}, pf.frame)
	}

	return output, nil
}

// overhang returns how much a character should hang into the margin.
func overhang(c rune) Abs {
	switch c {
	// Dashes
	case '–', '—':
		return 0.2
	case '-', '\u00ad':
		return 0.55
	// Punctuation
	case '.', ',':
		return 0.8
	case ':', ';':
		return 0.3
	// Arabic
	case '\u060C', '\u06D4':
		return 0.4
	default:
		return 0
	}
}

// frShare calculates the share of remaining space for a fractional item.
func frShare(amount layout.Fr, total layout.Fr, remaining Abs) Abs {
	if total == 0 {
		return 0
	}
	return Abs(float64(amount) / float64(total) * float64(remaining))
}

// alignPosition calculates offset based on alignment.
func alignPosition(align layout.Alignment, remaining Abs) Abs {
	switch align {
	case layout.AlignStart:
		return 0
	case layout.AlignCenter:
		return remaining / 2
	case layout.AlignEnd:
		return remaining
	default:
		return 0
	}
}

// buildTextFrame builds a frame from shaped text with justification.
func buildTextFrame(
	shaped *ShapedText,
	justificationRatio float64,
	extraJustification Abs,
) *FinalFrame {
	var width, height Abs
	var baseline Abs

	// Calculate dimensions from glyphs
	for _, g := range shaped.Glyphs.Kept() {
		advance := g.XAdvance.At(g.Size)

		// Apply justification adjustments
		if justificationRatio > 0 {
			// Stretching
			stretch := g.Stretchability()
			advance += (stretch[0] + stretch[1]).At(g.Size) * Abs(justificationRatio)
		} else if justificationRatio < 0 {
			// Shrinking
			shrink := g.Shrinkability()
			advance += (shrink[0] + shrink[1]).At(g.Size) * Abs(justificationRatio)
		}

		// Add extra justification for justifiable glyphs
		if g.IsJustifiable && extraJustification > 0 {
			advance += extraJustification
		}

		width += advance

		// Track height (simplified - would use actual glyph metrics)
		glyphHeight := g.Size * 1.2 // Approximation
		if glyphHeight > height {
			height = glyphHeight
		}
	}

	// Baseline is typically at 80% from top (simplified)
	baseline = height * 0.8

	frame := &FinalFrame{
		Size:     FinalSize{Width: width, Height: height},
		Baseline: baseline,
	}

	// Add text item to frame
	frame.Push(FinalPoint{X: 0, Y: 0}, FinalTextItem{Text: shaped})

	return frame
}

// LeadingText returns the first text item in the line, skipping tags.
func (l *Line) LeadingText() *ShapedText {
	for _, item := range l.Items {
		switch it := item.(type) {
		case *TextItem:
			if it.shaped != nil {
				return it.shaped
			}
		case *TagItem:
			continue
		default:
			return nil
		}
	}
	return nil
}

// TrailingText returns the last text item in the line, skipping tags.
func (l *Line) TrailingText() *ShapedText {
	for i := len(l.Items) - 1; i >= 0; i-- {
		switch it := l.Items[i].(type) {
		case *TextItem:
			if it.shaped != nil {
				return it.shaped
			}
		case *TagItem:
			continue
		default:
			return nil
		}
	}
	return nil
}
