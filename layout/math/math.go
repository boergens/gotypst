package math

// MathContext provides the context for math layout operations.
type MathContext struct {
	// Engine provides access to the compilation engine.
	Engine *Engine

	// Locator tracks element positions.
	Locator *SplitLocator

	// Region is the layout region.
	Region Region

	// FontsStack holds the current font stack.
	FontsStack []*Font

	// Fragments holds the accumulated math fragments.
	Fragments []MathFragment
}

// NewMathContext creates a new math context.
func NewMathContext(engine *Engine, locator *SplitLocator, base Size, font *Font) *MathContext {
	return &MathContext{
		Engine:     engine,
		Locator:    locator,
		Region:     NewRegion(base, Splat(false)),
		FontsStack: []*Font{font},
		Fragments:  nil,
	}
}

// Font returns the current base font.
func (ctx *MathContext) Font() *Font {
	if len(ctx.FontsStack) == 0 {
		return nil
	}
	return ctx.FontsStack[len(ctx.FontsStack)-1]
}

// Push adds a fragment to the context.
func (ctx *MathContext) Push(fragment MathFragment) {
	ctx.Fragments = append(ctx.Fragments, fragment)
}

// PushSpace adds a space fragment.
func (ctx *MathContext) PushSpace(amount Abs) {
	ctx.Push(&SpaceFragment{Amount: amount})
}

// PushLinebreak adds a linebreak fragment.
func (ctx *MathContext) PushLinebreak() {
	ctx.Push(&LinebreakFragment{})
}

// PushAlign adds an alignment fragment.
func (ctx *MathContext) PushAlign() {
	ctx.Push(&AlignFragment{})
}

// PushTag adds a tag fragment.
func (ctx *MathContext) PushTag(tag Tag) {
	ctx.Push(&TagFragment{Tag: tag})
}

// Extend adds multiple fragments to the context.
func (ctx *MathContext) Extend(fragments []MathFragment) {
	ctx.Fragments = append(ctx.Fragments, fragments...)
}

// LayoutIntoFragments lays out a math item and returns the resulting fragments.
func (ctx *MathContext) LayoutIntoFragments(item *MathItem, styles *StyleChain) ([]MathFragment, error) {
	start := len(ctx.Fragments)
	if err := ctx.LayoutIntoSelf(item, styles); err != nil {
		return nil, err
	}
	result := make([]MathFragment, len(ctx.Fragments)-start)
	copy(result, ctx.Fragments[start:])
	ctx.Fragments = ctx.Fragments[:start]
	return result, nil
}

// LayoutIntoFragment lays out a math item and returns a single fragment.
func (ctx *MathContext) LayoutIntoFragment(item *MathItem, styles *StyleChain) (MathFragment, error) {
	fragments, err := ctx.LayoutIntoFragments(item, styles)
	if err != nil {
		return nil, err
	}

	if len(fragments) == 1 {
		return fragments[0], nil
	}

	// Check if all fragments are text-like
	textLike := true
	for _, f := range fragments {
		if f.MathSize() != nil && !f.IsTextLike() {
			textLike = false
			break
		}
	}

	// Combine fragments into a single frame
	frame := FragmentsIntoFrame(fragments, styles)
	props := DefaultMathProperties()
	if item.Styles != nil {
		// Use item styles if available
	}

	ff := NewFrameFragment(props, styles.ResolveTextSize(), frame)
	ff.TextLike = textLike
	return ff, nil
}

// LayoutIntoSelf lays out a math item, accumulating fragments in the context.
func (ctx *MathContext) LayoutIntoSelf(item *MathItem, styles *StyleChain) error {
	outerStyles := item.Styles
	if outerStyles == nil {
		outerStyles = styles
	}

	for _, child := range item.Children {
		childStyles := child.Styles
		if childStyles == nil {
			childStyles = outerStyles
		}

		// Check if font changed
		if childStyles != outerStyles {
			// TODO: Check font family changes and push new font
		}

		if err := layoutRealized(child, ctx, childStyles); err != nil {
			return err
		}
	}

	return nil
}

// layoutRealized lays out a single math item.
func layoutRealized(item *MathItem, ctx *MathContext, styles *StyleChain) error {
	// Handle non-component items first
	switch item.Kind {
	case MathKindSpacing:
		ctx.PushSpace(item.SpacingAmount)
		return nil
	case MathKindSpace:
		spaceWidth := ctx.Font().Math().SpaceWidth.Resolve(styles.ResolveTextSize())
		ctx.PushSpace(spaceWidth)
		return nil
	case MathKindLinebreak:
		ctx.PushLinebreak()
		return nil
	case MathKindAlign:
		ctx.PushAlign()
		return nil
	case MathKindTag:
		ctx.PushTag(item.Tag)
		return nil
	}

	// Handle component items
	props := item.Props
	if props == nil {
		props = DefaultMathProperties()
	}

	// Insert left spacing
	if props.LSpace != nil {
		width := props.LSpace.RelativeTo(styles.ResolveTextSize())
		frag := &SpaceFragment{Amount: width}

		// Skip alignment points when placing spacing on the left
		inserted := false
		for i := len(ctx.Fragments) - 1; i >= 0; i-- {
			if !IsIgnorant(ctx.Fragments[i]) {
				if _, ok := ctx.Fragments[i].(*AlignFragment); ok {
					// Insert before the alignment point
					ctx.Fragments = append(ctx.Fragments[:i], append([]MathFragment{frag}, ctx.Fragments[i:]...)...)
					inserted = true
				}
				break
			}
		}
		if !inserted {
			ctx.Push(frag)
		}
	}

	// Dispatch based on item kind
	var err error
	switch item.Kind {
	case MathKindBox:
		err = layoutBox(item, ctx, styles, props)
	case MathKindExternal:
		err = layoutExternal(item, ctx, styles, props)
	case MathKindGlyph:
		err = layoutGlyph(item, ctx, styles, props)
	case MathKindCancel:
		err = LayoutCancel(item, ctx, styles, props)
	case MathKindRadical:
		err = LayoutRadical(item, ctx, styles, props)
	case MathKindLine:
		err = LayoutLine(item, ctx, styles, props)
	case MathKindAccent:
		err = LayoutAccent(item, ctx, styles, props)
	case MathKindScripts:
		err = LayoutScripts(item, ctx, styles, props)
	case MathKindPrimes:
		err = LayoutPrimes(item, ctx, styles, props)
	case MathKindTable:
		err = LayoutTable(item, ctx, styles, props)
	case MathKindFraction:
		err = LayoutFraction(item, ctx, styles, props)
	case MathKindSkewedFraction:
		err = LayoutSkewedFraction(item, ctx, styles, props)
	case MathKindText:
		err = LayoutText(item, ctx, styles, props)
	case MathKindFenced:
		err = LayoutFenced(item, ctx, styles, props)
	case MathKindGroup:
		fragment, err := ctx.LayoutIntoFragment(item, styles)
		if err != nil {
			return err
		}
		italics := fragment.ItalicsCorrection()
		accentTop, accentBottom := fragment.AccentAttach()
		ff := NewFrameFragment(props, styles.ResolveTextSize(), fragment.IntoFrame())
		ff.ItalicsCorr = italics
		ff.AccentAttachTop = accentTop
		ff.AccentAttachBottom = accentBottom
		ctx.Push(ff)
	}

	if err != nil {
		return err
	}

	// Insert right spacing
	if props.RSpace != nil {
		width := props.RSpace.RelativeTo(styles.ResolveTextSize())
		ctx.PushSpace(width)
	}

	return nil
}

// layoutBox lays out a box item.
func layoutBox(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// TODO: Implement box layout via inline layout
	frame := NewSoftFrame(Size{})
	ctx.Push(NewFrameFragment(props, styles.ResolveTextSize(), frame))
	return nil
}

// layoutExternal lays out an external content item.
func layoutExternal(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// TODO: Implement external layout
	frame := NewSoftFrame(Size{})
	if !frame.HasBaseline() {
		axis := ctx.Font().Math().AxisHeight.Resolve(styles.ResolveTextSize())
		frame.SetBaseline(frame.Height()/2.0 + axis)
	}
	ctx.Push(NewFrameFragment(props, styles.ResolveTextSize(), frame))
	return nil
}

// layoutGlyph lays out a glyph item.
func layoutGlyph(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// TODO: Implement full glyph layout with shaping
	return nil
}

// MathItem represents a math layout item.
type MathItem struct {
	Kind          MathKind
	Styles        *StyleChain
	Props         *MathProperties
	Children      []*MathItem
	SpacingAmount Abs
	Tag           Tag
	Text          string
}

// MathKind represents the kind of math item.
type MathKind int

const (
	MathKindSpacing MathKind = iota
	MathKindSpace
	MathKindLinebreak
	MathKindAlign
	MathKindTag
	MathKindBox
	MathKindExternal
	MathKindGlyph
	MathKindCancel
	MathKindRadical
	MathKindLine
	MathKindAccent
	MathKindScripts
	MathKindPrimes
	MathKindTable
	MathKindFraction
	MathKindSkewedFraction
	MathKindText
	MathKindFenced
	MathKindGroup
)

// StyleChain represents a chain of styles.
type StyleChain struct {
	// TODO: Implement style chain
	fontSize Abs
}

// ResolveTextSize returns the resolved text size.
func (s *StyleChain) ResolveTextSize() Abs {
	if s == nil || s.fontSize == 0 {
		return 12.0 // Default font size in points
	}
	return s.fontSize
}

// Get returns a style value. Placeholder for now.
func (s *StyleChain) Get(key string) interface{} {
	return nil
}

// Engine represents the compilation engine.
type Engine struct {
	// TODO: Implement engine
}

// SplitLocator tracks element positions.
type SplitLocator struct {
	// TODO: Implement locator
}

// Split creates a child locator.
func (l *SplitLocator) Split() *SplitLocator {
	return &SplitLocator{}
}

// Next returns the next locator position.
func (l *SplitLocator) Next(span interface{}) *SplitLocator {
	return &SplitLocator{}
}

// FragmentsIntoFrame converts a slice of fragments into a single frame.
func FragmentsIntoFrame(fragments []MathFragment, styles *StyleChain) *Frame {
	if len(fragments) == 0 {
		return NewSoftFrame(Size{})
	}

	// Calculate dimensions
	var width Abs
	var ascent, descent Abs
	for _, f := range fragments {
		width += f.Width()
		if a := f.Ascent(); a > ascent {
			ascent = a
		}
		if d := f.Descent(); d > descent {
			descent = d
		}
	}

	frame := NewSoftFrame(Size{X: width, Y: ascent + descent})
	frame.SetBaseline(ascent)

	// Position fragments
	var x Abs
	for _, f := range fragments {
		y := ascent - f.Ascent()
		frame.PushFrame(Point{X: x, Y: y}, f.IntoFrame())
		x += f.Width()
	}

	return frame
}

// LayoutEquationInline lays out an inline equation.
func LayoutEquationInline(
	elem *EquationElem,
	engine *Engine,
	locator *SplitLocator,
	styles *StyleChain,
	region Size,
) ([]InlineItem, error) {
	font := getFont(styles)
	if font == nil {
		return nil, &SourceError{Message: "no font could be found"}
	}

	splitLocator := locator.Split()
	ctx := NewMathContext(engine, splitLocator, region, font)

	// TODO: Resolve equation and layout
	item := &MathItem{Kind: MathKindGroup}
	fragments, err := ctx.LayoutIntoFragments(item, styles)
	if err != nil {
		return nil, err
	}

	var items []InlineItem
	if len(fragments) == 0 {
		// Empty equation should have a height
		items = append(items, &InlineFrame{Frame: NewSoftFrame(Size{})})
	} else {
		frame := FragmentsIntoFrame(fragments, styles)
		items = append(items, &InlineFrame{Frame: frame})
	}

	return items, nil
}

// LayoutEquationBlock lays out a block equation.
func LayoutEquationBlock(
	elem *EquationElem,
	engine *Engine,
	locator *SplitLocator,
	styles *StyleChain,
	regions *Regions,
) (*Fragment, error) {
	font := getFont(styles)
	if font == nil {
		return nil, &SourceError{Message: "no font could be found"}
	}

	splitLocator := locator.Split()
	ctx := NewMathContext(engine, splitLocator, regions.Base(), font)

	// TODO: Resolve equation and layout
	item := &MathItem{Kind: MathKindGroup}
	fragments, err := ctx.LayoutIntoFragments(item, styles)
	if err != nil {
		return nil, err
	}

	frame := FragmentsIntoFrame(fragments, styles)
	return NewFragment([]*Frame{frame}), nil
}

// getFont gets the current font from styles.
func getFont(styles *StyleChain) *Font {
	// TODO: Implement font lookup from styles
	return &Font{}
}

// EquationElem represents an equation element.
type EquationElem struct {
	Block     bool
	Numbering interface{}
	Body      interface{}
}

// SourceError represents a source-located error.
type SourceError struct {
	Span    Span
	Message string
}

func (e *SourceError) Error() string {
	return e.Message
}

// Placeholder layout functions that will be implemented in separate files

// LayoutCancel lays out a cancel item.
func LayoutCancel(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in cancel.go
	return nil
}

// LayoutRadical lays out a radical item.
func LayoutRadical(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in radical.go
	return nil
}

// LayoutLine lays out a line item.
func LayoutLine(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in line.go
	return nil
}

// LayoutAccent lays out an accent item.
func LayoutAccent(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in accent.go
	return nil
}

// LayoutScripts lays out a scripts item.
func LayoutScripts(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in scripts.go
	return nil
}

// LayoutPrimes lays out a primes item.
func LayoutPrimes(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in scripts.go
	return nil
}

// LayoutTable lays out a table item.
func LayoutTable(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in table.go
	return nil
}

// LayoutFraction lays out a fraction item.
func LayoutFraction(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in fraction.go
	return nil
}

// LayoutSkewedFraction lays out a skewed fraction item.
func LayoutSkewedFraction(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in fraction.go
	return nil
}

// LayoutText lays out a text item.
func LayoutText(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in text.go
	return nil
}

// LayoutFenced lays out a fenced item.
func LayoutFenced(item *MathItem, ctx *MathContext, styles *StyleChain, props *MathProperties) error {
	// Implemented in fenced.go
	return nil
}

// FragmentFont returns the font and size for a fragment.
func FragmentFont(f MathFragment, ctx *MathContext, styles *StyleChain) (*Font, Abs) {
	if gf, ok := f.(*GlyphFragment); ok {
		return gf.Item.Font, gf.Item.Size
	}
	fontSize := styles.ResolveTextSize()
	return ctx.Font(), fontSize
}
