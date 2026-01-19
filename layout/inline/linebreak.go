package inline

import (
	"math"
	"strings"
	"unicode"

	"github.com/boergens/gotypst/layout"
	"golang.org/x/text/unicode/bidi"
)

// Cost represents the cost of a line or inline layout.
type Cost = float64

// Cost parameters.
// We choose higher costs than the Knuth-Plass paper (which would be 50) because
// it hyphenates way too eagerly in Typst otherwise.
const (
	DefaultHyphCost Cost = 135.0
	DefaultRuntCost Cost = 100.0
)

// Other parameters.
const (
	MinRatio       = -1.0
	MinApproxRatio = -0.5
	BoundEps       = 1e-3
)

// Zero width space character.
const ZWS = '\u200B'

// Breakpoint represents a line break opportunity.
type Breakpoint int

const (
	// BreakpointNormal is a normal break opportunity (e.g., after a space).
	BreakpointNormal Breakpoint = iota
	// BreakpointMandatory is a mandatory breakpoint (after '\n' or at end).
	BreakpointMandatory
)

// HyphenBreakpoint represents a hyphenation opportunity with char counts.
type HyphenBreakpoint struct {
	// Before is the number of chars before the hyphen in the word.
	Before uint8
	// After is the number of chars after the hyphen in the word.
	After uint8
}

// BreakpointInfo holds breakpoint information including hyphenation data.
type BreakpointInfo struct {
	Type   Breakpoint
	Hyphen *HyphenBreakpoint
}

// Normal returns a normal breakpoint info.
func Normal() BreakpointInfo {
	return BreakpointInfo{Type: BreakpointNormal}
}

// Mandatory returns a mandatory breakpoint info.
func Mandatory() BreakpointInfo {
	return BreakpointInfo{Type: BreakpointMandatory}
}

// Hyphen returns a hyphen breakpoint info.
func Hyphen(before, after uint8) BreakpointInfo {
	return BreakpointInfo{
		Type:   BreakpointNormal,
		Hyphen: &HyphenBreakpoint{Before: before, After: after},
	}
}

// IsHyphen returns true if this is a hyphenation breakpoint.
func (b BreakpointInfo) IsHyphen() bool {
	return b.Hyphen != nil
}

// IsMandatory returns true if this is a mandatory breakpoint.
func (b BreakpointInfo) IsMandatory() bool {
	return b.Type == BreakpointMandatory
}

// Trim determines how to trim the end of a line.
// It's an invariant that Layout <= Shaping.
type Trim struct {
	// Layout is the position up to which text affects layout.
	// Text in Layout..Shaping is shaped but has zero advance.
	Layout int
	// Shaping is the position up to which text is shaped.
	Shaping int
}

// UniformTrim creates a Trim with equal layout and shaping positions.
func UniformTrim(pos int) Trim {
	return Trim{Layout: pos, Shaping: pos}
}

// TrimLine trims a line before the given breakpoint.
func (b BreakpointInfo) TrimLine(start int, line string) Trim {
	if b.IsHyphen() {
		// Trim nothing for hyphen breaks.
		return UniformTrim(start + len(line))
	}

	if b.IsMandatory() {
		// Trim linebreaks for mandatory breaks.
		trimmed := trimMandatoryBreaks(line)
		return UniformTrim(start + len(trimmed))
	}

	// Normal break: trim trailing whitespace for layout but keep for shaping.
	trimmed := trimTrailingWhitespace(line)
	return Trim{
		Layout:  start + len(trimmed),
		Shaping: start + len(line),
	}
}

// trimTrailingWhitespace trims whitespace and ZWS from the end of a string.
func trimTrailingWhitespace(s string) string {
	runes := []rune(s)
	end := len(runes)
	for end > 0 && (unicode.IsSpace(runes[end-1]) || runes[end-1] == ZWS) {
		end--
	}
	return string(runes[:end])
}

// trimMandatoryBreaks trims mandatory line break characters from the end.
func trimMandatoryBreaks(s string) string {
	runes := []rune(s)
	end := len(runes)
	for end > 0 {
		c := runes[end-1]
		if c == '\n' || c == '\r' || c == '\u0085' || c == '\u2028' || c == '\u2029' {
			end--
		} else {
			break
		}
	}
	return string(runes[:end])
}

// Linebreak breaks the text into lines.
func Linebreak(p *Preparation, width layout.Abs) []Line {
	switch p.Config.Linebreaks {
	case layout.LinebreaksSimple:
		return linebreakSimple(p, width)
	case layout.LinebreaksOptimized:
		return linebreakOptimized(p, width)
	default:
		return linebreakSimple(p, width)
	}
}

// linebreakSimple performs line breaking in simple first-fit style.
// It builds lines greedily, always taking the longest possible line.
func linebreakSimple(p *Preparation, width layout.Abs) []Line {
	lines := make([]Line, 0, 16)
	start := 0
	var last *struct {
		line Line
		end  int
	}

	breakpointsFn(p, func(end int, bp BreakpointInfo) {
		// Compute the line and its size.
		var pred *Line
		if len(lines) > 0 {
			pred = &lines[len(lines)-1]
		}
		attempt := makeLine(p, start, end, bp, pred)

		// If the line doesn't fit anymore, push the last fitting attempt
		// and rebuild from there.
		if !width.Fits(attempt.Width) && last != nil {
			lines = append(lines, last.line)
			start = last.end
			attempt = makeLine(p, start, end, bp, &lines[len(lines)-1])
			last = nil
		}

		// Finish the line if mandatory break or already doesn't fit.
		if bp.IsMandatory() || !width.Fits(attempt.Width) {
			lines = append(lines, attempt)
			start = end
			last = nil
		} else {
			last = &struct {
				line Line
				end  int
			}{attempt, end}
		}
	})

	if last != nil {
		lines = append(lines, last.line)
	}

	return lines
}

// linebreakOptimized performs line breaking using Knuth-Plass algorithm.
func linebreakOptimized(p *Preparation, width layout.Abs) []Line {
	metrics := computeCostMetrics(p)

	// Get upper bound from approximate pass.
	upperBound := linebreakOptimizedApproximate(p, width, metrics)

	// Use upper bound for bounded optimization.
	return linebreakOptimizedBounded(p, width, metrics, upperBound)
}

// entry is an entry in the dynamic programming table.
type entry struct {
	pred  int
	total Cost
	line  Line
	end   int
}

// linebreakOptimizedBounded performs Knuth-Plass with upper bound pruning.
func linebreakOptimizedBounded(p *Preparation, width layout.Abs, metrics *CostMetrics, upperBound Cost) []Line {
	// Dynamic programming table.
	table := []entry{{pred: 0, total: 0.0, line: EmptyLine(), end: 0}}

	active := 0
	prevEnd := 0

	breakpointsFn(p, func(end int, bp BreakpointInfo) {
		var best *entry

		// Lower bound for cost of all following line attempts.
		var lineLowerBound *Cost

		for predIndex := active; predIndex < len(table); predIndex++ {
			pred := &table[predIndex]
			start := pred.end
			unbreakable := prevEnd == start

			// Skip if minimum cost exceeds bound.
			if lineLowerBound != nil && pred.total+*lineLowerBound > upperBound+BoundEps {
				continue
			}

			// Build the line.
			attempt := makeLine(p, start, end, bp, &pred.line)

			// Determine cost and ratio.
			lineRatio, lineCost := ratioAndCost(p, metrics, width, &pred.line, &attempt, bp, unbreakable)

			// Adjust active set for overfull lines.
			if lineRatio < metrics.minRatio && active == predIndex {
				active++
			}

			// Total cost including predecessors.
			total := pred.total + lineCost

			// Set lower bound if line is underfull.
			if lineRatio > 0.0 && lineLowerBound == nil && !attempt.HasNegativeWidthItems() {
				lineLowerBound = &lineCost
			}

			// Skip if cost exceeds bound.
			if total > upperBound+BoundEps {
				continue
			}

			// Take if better than current best.
			if best == nil || best.total >= total {
				best = &entry{pred: predIndex, total: total, line: attempt, end: end}
			}
		}

		// Reset active for mandatory breaks.
		if bp.IsMandatory() {
			active = len(table)
		}

		if best != nil {
			table = append(table, *best)
		}
		prevEnd = end
	})

	// Retrace the best path.
	lines := make([]Line, 0, 16)
	idx := len(table) - 1

	// Sanity check - should cover full text.
	if table[idx].end != len(p.Text) {
		// Fallback to unbounded if bound was faulty.
		return linebreakOptimizedBounded(p, width, metrics, math.Inf(1))
	}

	for idx != 0 {
		lines = append(lines, table[idx].line)
		idx = table[idx].pred
	}

	// Reverse to get correct order.
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	return lines
}

// approxEntry is an entry in the approximate DP table.
type approxEntry struct {
	pred        int
	total       Cost
	end         int
	unbreakable bool
	breakpoint  BreakpointInfo
}

// linebreakOptimizedApproximate runs Knuth-Plass with approximate metrics.
func linebreakOptimizedApproximate(p *Preparation, width layout.Abs, metrics *CostMetrics) Cost {
	estimates := computeEstimates(p)

	table := []approxEntry{{
		pred:       0,
		total:      0.0,
		end:        0,
		breakpoint: Mandatory(),
	}}

	active := 0
	prevEnd := 0

	breakpointsFn(p, func(end int, bp BreakpointInfo) {
		var best *approxEntry

		for predIndex := active; predIndex < len(table); predIndex++ {
			pred := &table[predIndex]
			start := pred.end
			unbreakable := prevEnd == start

			// Approximate justification check.
			justify := p.Config.Justify && !bp.IsMandatory()

			// Check for consecutive dashes.
			consecutiveDash := pred.breakpoint.IsHyphen() && bp.IsHyphen()

			// Estimate line metrics.
			trimmedEnd := start + len(trimTrailingWhitespace(p.Text[start:end]))
			hyphenWidth := layout.Abs(0)
			if bp.IsHyphen() {
				hyphenWidth = metrics.approxHyphenWidth
			}

			lineRatio := rawRatio(
				p,
				width,
				estimates.widths.estimate(start, trimmedEnd)+hyphenWidth,
				estimates.stretchability.estimate(start, trimmedEnd),
				estimates.shrinkability.estimate(start, trimmedEnd),
				estimates.justifiables.estimate(start, trimmedEnd),
			)

			lineCost := rawCost(metrics, bp, lineRatio, justify, unbreakable, consecutiveDash, true)

			if lineRatio < metrics.minRatio && active == predIndex {
				active++
			}

			total := pred.total + lineCost

			if best == nil || best.total >= total {
				best = &approxEntry{
					pred:        predIndex,
					total:       total,
					end:         end,
					unbreakable: unbreakable,
					breakpoint:  bp,
				}
			}
		}

		if bp.IsMandatory() {
			active = len(table)
		}

		if best != nil {
			table = append(table, *best)
		}
		prevEnd = end
	})

	// Retrace and compute exact cost.
	indices := make([]int, 0, 16)
	idx := len(table) - 1
	for idx != 0 {
		indices = append(indices, idx)
		idx = table[idx].pred
	}

	pred := EmptyLine()
	start := 0
	var exact Cost

	for i := len(indices) - 1; i >= 0; i-- {
		idx := indices[i]
		e := table[idx]

		attempt := makeLine(p, start, e.end, e.breakpoint, &pred)
		ratio, lineCost := ratioAndCost(p, metrics, width, &pred, &attempt, e.breakpoint, e.unbreakable)

		// If approximation produces invalid layout, bail with infinite bound.
		if ratio < metrics.minRatio {
			return math.Inf(1)
		}

		pred = attempt
		start = e.end
		exact += lineCost
	}

	return exact
}

// ratioAndCost computes the stretch ratio and cost of a line.
func ratioAndCost(p *Preparation, metrics *CostMetrics, availableWidth layout.Abs, pred, attempt *Line, bp BreakpointInfo, unbreakable bool) (float64, Cost) {
	ratio := rawRatio(
		p,
		availableWidth,
		attempt.Width,
		attempt.Stretchability(),
		attempt.Shrinkability(),
		attempt.Justifiables(),
	)

	hasDash := pred.Dash != 0 && attempt.Dash != 0
	cost := rawCost(metrics, bp, ratio, attempt.Justify, unbreakable, hasDash, false)

	return ratio, cost
}

// rawRatio determines the stretch ratio for a line.
func rawRatio(p *Preparation, availableWidth, lineWidth, stretchability, shrinkability layout.Abs, justifiables int) float64 {
	delta := availableWidth - lineWidth

	// Handle floating point errors.
	if delta.ApproxEq(0) {
		delta = 0
	}

	// Determine adjustability.
	var adjustability layout.Abs
	if delta >= 0 {
		adjustability = stretchability
	} else {
		adjustability = shrinkability
	}
	if adjustability < 0 {
		adjustability = 0
	}

	ratio := float64(delta) / float64(adjustability)

	// Handle NaN (often from zero delta and zero adjustability).
	if math.IsNaN(ratio) {
		ratio = 0.0
	}

	// Handle over-stretching with justifiables.
	if ratio > 1.0 {
		j := justifiables
		if j < 1 {
			j = 1
		}
		extraStretch := float64(delta-adjustability) / float64(j)
		ratio = 1.0 + extraStretch/(float64(p.Config.FontSize)/2.0)
	}

	// Clamp ratio.
	if ratio < MinRatio-1.0 {
		ratio = MinRatio - 1.0
	}
	if ratio > 10.0 {
		ratio = 10.0
	}

	return ratio
}

// rawCost computes the cost of a line.
func rawCost(metrics *CostMetrics, bp BreakpointInfo, ratio float64, justify, unbreakable, consecutiveDash, approx bool) Cost {
	// Determine badness.
	var badness Cost
	minRat := metrics.minRatio
	if approx {
		minRat = metrics.minApproxRatio
	}

	if ratio < minRat {
		// Overfull line.
		badness = 1_000_000.0
	} else if !bp.IsMandatory() || justify || ratio < 0.0 {
		// Justified or needs shrinking.
		badness = 100.0 * math.Pow(math.Abs(ratio), 3)
	} else {
		badness = 0.0
	}

	// Compute penalties.
	var penalty Cost

	// Penalize runts.
	if unbreakable && bp.IsMandatory() {
		penalty += metrics.runtCost
	}

	// Penalize hyphenation.
	if bp.Hyphen != nil {
		const limit uint8 = 5
		steps := saturatingSub(limit, bp.Hyphen.Before) + saturatingSub(limit, bp.Hyphen.After)
		extra := 0.15 * float64(steps)
		penalty += (1.0 + extra) * metrics.hyphCost
	}

	// Penalize consecutive dashes.
	if consecutiveDash {
		penalty += metrics.hyphCost
	}

	// Knuth-Plass formula: (1 + badness + penalty)^2
	return math.Pow(1.0+badness+penalty, 2)
}

func saturatingSub(a, b uint8) uint8 {
	if b >= a {
		return 0
	}
	return a - b
}

// CostMetrics holds resolved metrics for cost computation.
type CostMetrics struct {
	minRatio         float64
	minApproxRatio   float64
	approxHyphenWidth layout.Abs
	hyphCost         Cost
	runtCost         Cost
}

// computeCostMetrics computes shared metrics for optimization.
func computeCostMetrics(p *Preparation) *CostMetrics {
	minRatio := 0.0
	minApproxRatio := 0.0
	if p.Config.Justify {
		minRatio = MinRatio
		minApproxRatio = MinApproxRatio
	}

	return &CostMetrics{
		minRatio:         minRatio,
		minApproxRatio:   minApproxRatio,
		approxHyphenWidth: layout.Em(0.33).At(p.Config.FontSize),
		hyphCost:         DefaultHyphCost * p.Config.Costs.Hyphenation,
		runtCost:         DefaultRuntCost * p.Config.Costs.Runt,
	}
}

// Estimates holds cumulative arrays for quick metric estimation.
type Estimates struct {
	widths         *CumulativeVec[layout.Abs]
	stretchability *CumulativeVec[layout.Abs]
	shrinkability  *CumulativeVec[layout.Abs]
	justifiables   *CumulativeVec[int]
}

// computeEstimates computes estimations for approximate Knuth-Plass.
func computeEstimates(p *Preparation) *Estimates {
	cap := len(p.Text)

	widths := newCumulativeVec[layout.Abs](cap)
	stretchability := newCumulativeVec[layout.Abs](cap)
	shrinkability := newCumulativeVec[layout.Abs](cap)
	justifiables := newCumulativeVec[int](cap)

	for _, pi := range p.Items {
		if ti, ok := pi.Item.(*TextItem); ok && ti.shaped != nil {
			for _, g := range ti.shaped.Glyphs {
				byteLen := g.Range.Len()
				stretch := (g.Adjustability.StretchLeft + g.Adjustability.StretchRight).At(g.Size)
				shrink := (g.Adjustability.ShrinkLeft + g.Adjustability.ShrinkRight).At(g.Size)
				widths.push(byteLen, g.XAdvance.At(g.Size))
				stretchability.push(byteLen, stretch)
				shrinkability.push(byteLen, shrink)
				just := 0
				if g.IsJustifiable {
					just = 1
				}
				justifiables.push(byteLen, just)
			}
		} else {
			widths.push(pi.Range.Len(), pi.Item.NaturalWidth())
		}

		widths.adjust(pi.Range.End)
		stretchability.adjust(pi.Range.End)
		shrinkability.adjust(pi.Range.End)
		justifiables.adjust(pi.Range.End)
	}

	return &Estimates{
		widths:         widths,
		stretchability: stretchability,
		shrinkability:  shrinkability,
		justifiables:   justifiables,
	}
}

// CumulativeVec is an accumulative array of a metric.
type CumulativeVec[T Numeric] struct {
	total  T
	summed []T
}

// Numeric constraint for CumulativeVec.
type Numeric interface {
	~int | ~float64
}

func newCumulativeVec[T Numeric](capacity int) *CumulativeVec[T] {
	cv := &CumulativeVec[T]{
		summed: make([]T, 0, capacity),
	}
	var zero T
	cv.summed = append(cv.summed, zero)
	return cv
}

func (c *CumulativeVec[T]) adjust(length int) {
	for len(c.summed) < length {
		c.summed = append(c.summed, c.total)
	}
}

func (c *CumulativeVec[T]) push(byteLen int, metric T) {
	c.total = c.total + metric
	for i := 0; i < byteLen; i++ {
		c.summed = append(c.summed, c.total)
	}
}

func (c *CumulativeVec[T]) estimate(start, end int) T {
	return c.get(end) - c.get(start)
}

func (c *CumulativeVec[T]) get(index int) T {
	if index == 0 {
		var zero T
		return zero
	}
	if index-1 < len(c.summed) {
		return c.summed[index-1]
	}
	return c.total
}

// breakpointsFn calls f for all possible line break points.
func breakpointsFn(p *Preparation, f func(end int, bp BreakpointInfo)) {
	text := p.Text

	// Single breakpoint at end for empty text.
	if len(text) == 0 {
		f(0, Mandatory())
		return
	}

	hyphenate := p.Config.Hyphenate == nil || *p.Config.Hyphenate

	runes := []rune(text)
	runeOffsets := computeRuneOffsets(text)

	last := 0

	for i, r := range runes {
		offset := runeOffsets[i]
		nextOffset := len(text)
		if i+1 < len(runeOffsets) {
			nextOffset = runeOffsets[i+1]
		}

		// Check for line break opportunities.
		bp := classifyBreakpoint(r, i == len(runes)-1)
		if bp == nil {
			continue
		}

		// Hyphenate between last and current breakpoint.
		if hyphenate && last < offset {
			segment := text[last:offset]
			hyphenateSegment(p, last, segment, f)
		}

		f(nextOffset, *bp)
		last = nextOffset
	}
}

// computeRuneOffsets returns byte offsets for each rune.
func computeRuneOffsets(s string) []int {
	offsets := make([]int, 0, len(s))
	offset := 0
	for _, r := range s {
		offsets = append(offsets, offset)
		offset += len(string(r))
	}
	return offsets
}

// classifyBreakpoint determines if a character is a break opportunity.
func classifyBreakpoint(r rune, isLast bool) *BreakpointInfo {
	if isLast {
		bp := Mandatory()
		return &bp
	}

	// Mandatory breaks.
	if r == '\n' || r == '\r' || r == '\u0085' || r == '\u2028' || r == '\u2029' {
		bp := Mandatory()
		return &bp
	}

	// Normal breaks after spaces and certain punctuation.
	if unicode.IsSpace(r) {
		bp := Normal()
		return &bp
	}

	// Use bidi class for additional break opportunities.
	props, _ := bidi.LookupRune(r)
	switch props.Class() {
	case bidi.WS, bidi.S, bidi.B:
		bp := Normal()
		return &bp
	}

	return nil
}

// hyphenateSegment generates hyphenation breakpoints within a segment.
func hyphenateSegment(p *Preparation, offset int, segment string, f func(end int, bp BreakpointInfo)) {
	// Simple word detection: only alphabetic characters.
	runes := []rune(segment)
	allAlphabetic := true
	for _, r := range runes {
		if !unicode.IsLetter(r) {
			allAlphabetic = false
			break
		}
	}

	if !allAlphabetic || len(runes) < 4 {
		return
	}

	// Simple hyphenation: break at syllable boundaries.
	// This is a simplified version - full hyphenation would use a dictionary.
	count := len(runes)
	for i := 2; i < count-2; i++ {
		// Simple heuristic: break between vowel-consonant or consonant-vowel.
		if shouldHyphenate(runes, i) {
			byteOffset := offset
			for j := 0; j < i; j++ {
				byteOffset += len(string(runes[j]))
			}
			f(byteOffset, Hyphen(uint8(i), uint8(count-i)))
		}
	}
}

// shouldHyphenate is a simple heuristic for hyphenation points.
func shouldHyphenate(runes []rune, pos int) bool {
	if pos < 1 || pos >= len(runes) {
		return false
	}

	// Check for vowel-consonant transitions.
	prev := runes[pos-1]
	curr := runes[pos]

	prevVowel := isVowel(prev)
	currVowel := isVowel(curr)

	// Hyphenate between vowel and consonant.
	return prevVowel && !currVowel
}

func isVowel(r rune) bool {
	r = unicode.ToLower(r)
	return r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u' ||
		r == 'á' || r == 'é' || r == 'í' || r == 'ó' || r == 'ú' ||
		r == 'ä' || r == 'ö' || r == 'ü'
}

// makeLine creates a line from a range.
func makeLine(p *Preparation, start, end int, bp BreakpointInfo, pred *Line) Line {
	if start >= end || start >= len(p.Text) {
		return EmptyLine()
	}

	full := p.Text[start:end]

	// Determine if line should be justified.
	justify := strings.HasSuffix(full, "\u2028") ||
		(p.Config.Justify && !bp.IsMandatory())

	// Process dashes.
	var dash Dash
	if bp.IsHyphen() || strings.HasSuffix(full, "\u00AD") {
		dash = DashSoft
	} else if strings.HasSuffix(full, "-") {
		dash = DashHard
	} else if strings.HasSuffix(full, "–") || strings.HasSuffix(full, "—") {
		dash = DashOther
	}

	// Trim the line.
	trim := bp.TrimLine(start, full)

	// Collect items for the line.
	items := collectLineItems(p, start, end, trim)

	// Compute width.
	var width layout.Abs
	for _, item := range items {
		width += item.NaturalWidth()
	}

	return Line{
		Items:   items,
		Width:   width,
		Justify: justify,
		Dash:    dash,
	}
}

// collectLineItems collects items for a line in the given range.
func collectLineItems(p *Preparation, start, end int, trim Trim) []Item {
	var items []Item

	for _, pi := range p.Items {
		// Skip items completely before the range.
		if pi.Range.End <= start {
			continue
		}
		// Stop at items completely after the range.
		if pi.Range.Start >= end {
			break
		}

		// Handle items that overlap the range.
		if pi.Range.Start < start || pi.Range.End > end {
			// For partial overlaps, we'd need to reshape text.
			// For now, include the full item.
			items = append(items, pi.Item)
		} else {
			items = append(items, pi.Item)
		}
	}

	return items
}
