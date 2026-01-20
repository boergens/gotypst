package flow

import (
	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/layout"
)

// Collector converts content elements into flow layout children.
// It preprocesses content into the Work structure for distribution.
type Collector struct {
	// engine provides layout context.
	engine *Engine
	// mode determines how content is processed.
	mode FlowMode
	// styles provides style information.
	styles StyleChain
	// locator tracks element locations.
	locator *Locator
	// children accumulates the collected flow children.
	children []Child
	// lastWasSpacing tracks if the last child was spacing.
	lastWasSpacing bool
}

// StyleChain represents a chain of styles for content.
type StyleChain struct {
	// Styles contains the style values.
	Styles map[string]interface{}
}

// Get retrieves a style value.
func (s StyleChain) Get(key string) interface{} {
	if s.Styles == nil {
		return nil
	}
	return s.Styles[key]
}

// Locator tracks element locations for introspection.
type Locator struct {
	// Current is the current location counter.
	Current uint64
}

// Next returns the next location.
func (l *Locator) Next() Location {
	l.Current++
	return Location(l.Current)
}

// NewCollector creates a new content collector.
func NewCollector(engine *Engine, mode FlowMode, styles StyleChain, locator *Locator) *Collector {
	return &Collector{
		engine:  engine,
		mode:    mode,
		styles:  styles,
		locator: locator,
	}
}

// Collect converts content into flow layout children.
// This is the main entry point for content collection.
func Collect(engine *Engine, content *eval.Content, mode FlowMode, styles StyleChain, locator *Locator) []Child {
	c := NewCollector(engine, mode, styles, locator)
	c.collectContent(content)
	return c.children
}

// collectContent processes a Content value.
func (c *Collector) collectContent(content *eval.Content) {
	if content == nil {
		return
	}
	for _, elem := range content.Elements {
		c.collectElement(elem)
	}
}

// collectElement dispatches to the appropriate handler for each element type.
func (c *Collector) collectElement(elem eval.ContentElement) {
	if elem == nil {
		return
	}

	switch e := elem.(type) {
	// Text and inline elements
	case *eval.TextElement:
		c.collectText(e)
	case *eval.SpaceElement:
		c.collectSpace(e)
	case *eval.LinebreakElement:
		c.collectLinebreak(e)

	// Paragraph structure
	case *eval.ParbreakElement:
		c.collectParbreak(e)
	case *eval.ParagraphElement:
		c.collectParagraph(e)

	// Block elements
	case *eval.HeadingElement:
		c.collectHeading(e)
	case *eval.RawElement:
		c.collectRaw(e)

	// List elements
	case *eval.ListElement:
		c.collectList(e)
	case *eval.EnumElement:
		c.collectEnum(e)
	case *eval.TermsElement:
		c.collectTerms(e)

	// Layout elements
	case *eval.StackElement:
		c.collectStack(e)
	case *eval.AlignElement:
		c.collectAlign(e)

	// Styling elements
	case *eval.StrongElement:
		c.collectStrong(e)
	case *eval.EmphElement:
		c.collectEmph(e)

	// Link and reference elements
	case *eval.LinkElement:
		c.collectLink(e)
	case *eval.RefElement:
		c.collectRef(e)

	// Math elements
	case *eval.EquationElement:
		c.collectEquation(e)

	// Smart quotes
	case *eval.SmartQuoteElement:
		c.collectSmartQuote(e)

	default:
		// Unknown element - treat as single unbreakable block
		c.collectUnknown(elem)
	}
}

// collectText handles text elements.
// Text is collected into inline content that will become lines.
func (c *Collector) collectText(elem *eval.TextElement) {
	// Text elements contribute to paragraph content.
	// They will be shaped and broken into lines during inline layout.
	// For now, we mark that we have content (not just spacing).
	c.lastWasSpacing = false
}

// collectSpace handles space elements.
func (c *Collector) collectSpace(elem *eval.SpaceElement) {
	// Spaces contribute to inline content.
	// They don't create flow children by themselves.
	c.lastWasSpacing = true
}

// collectLinebreak handles explicit linebreak elements.
func (c *Collector) collectLinebreak(elem *eval.LinebreakElement) {
	// Linebreaks force a line break within a paragraph.
	// They are handled during inline layout, not flow layout.
	c.lastWasSpacing = false
}

// collectParbreak handles paragraph break elements.
func (c *Collector) collectParbreak(elem *eval.ParbreakElement) {
	// Paragraph breaks create spacing between blocks.
	// Add relative spacing with weakness (can be collapsed).
	spacing := c.getParSpacing()
	if spacing > 0 {
		c.addRelSpacing(spacing, 1) // weakness 1 = collapsible
	}
	c.lastWasSpacing = true
}

// collectParagraph handles paragraph elements.
func (c *Collector) collectParagraph(elem *eval.ParagraphElement) {
	// Paragraphs are multi-child blocks that can break across regions.
	// The actual line breaking happens in inline layout.
	align := c.getParagraphAlignment()
	c.children = append(c.children, &MultiChild{
		Align:  align,
		Sticky: false,
		Alone:  false,
	})
	c.lastWasSpacing = false
}

// collectHeading handles heading elements.
func (c *Collector) collectHeading(elem *eval.HeadingElement) {
	// Headings are sticky blocks - they shouldn't be alone at region bottom.
	align := c.getBlockAlignment()
	c.children = append(c.children, &SingleChild{
		Align:  align,
		Sticky: true, // Headings stick to following content
		Alone:  true, // Can be alone (for chapter-start headings)
	})
	c.lastWasSpacing = false
}

// collectRaw handles raw/code elements.
func (c *Collector) collectRaw(elem *eval.RawElement) {
	if elem.Block {
		// Block-level raw elements are single children.
		align := c.getBlockAlignment()
		c.children = append(c.children, &SingleChild{
			Align:  align,
			Sticky: false,
			Alone:  false,
		})
	}
	// Inline raw elements are handled as part of paragraph content.
	c.lastWasSpacing = false
}

// collectList handles list elements.
func (c *Collector) collectList(elem *eval.ListElement) {
	// Lists are multi-child blocks.
	align := c.getBlockAlignment()
	c.children = append(c.children, &MultiChild{
		Align:  align,
		Sticky: false,
		Alone:  false,
	})
	c.lastWasSpacing = false
}

// collectEnum handles enumeration (numbered list) elements.
func (c *Collector) collectEnum(elem *eval.EnumElement) {
	// Enumerations are multi-child blocks.
	align := c.getBlockAlignment()
	c.children = append(c.children, &MultiChild{
		Align:  align,
		Sticky: false,
		Alone:  false,
	})
	c.lastWasSpacing = false
}

// collectTerms handles term list (definition list) elements.
func (c *Collector) collectTerms(elem *eval.TermsElement) {
	// Terms lists are multi-child blocks.
	align := c.getBlockAlignment()
	c.children = append(c.children, &MultiChild{
		Align:  align,
		Sticky: false,
		Alone:  false,
	})
	c.lastWasSpacing = false
}

// collectStack handles stack layout elements.
func (c *Collector) collectStack(elem *eval.StackElement) {
	// Stacks arrange children along an axis.
	// Vertical stacks (ttb/btt) are multi-child blocks.
	// Horizontal stacks (ltr/rtl) are single-child blocks.
	if elem.Dir == eval.StackTTB || elem.Dir == eval.StackBTT {
		// Vertical stack - potentially breakable
		align := c.getBlockAlignment()
		c.children = append(c.children, &MultiChild{
			Align:  align,
			Sticky: false,
			Alone:  false,
		})
	} else {
		// Horizontal stack - single unbreakable block
		align := c.getBlockAlignment()
		c.children = append(c.children, &SingleChild{
			Align:  align,
			Sticky: false,
			Alone:  false,
		})
	}
	c.lastWasSpacing = false
}

// collectAlign handles alignment elements.
func (c *Collector) collectAlign(elem *eval.AlignElement) {
	// Aligned content modifies the alignment of its body.
	// Collect the body content with modified alignment.
	c.collectContent(&elem.Body)
}

// collectStrong handles strong (bold) elements.
func (c *Collector) collectStrong(elem *eval.StrongElement) {
	// Strong is an inline style - collect its content.
	c.collectContent(&elem.Content)
}

// collectEmph handles emphasis (italic) elements.
func (c *Collector) collectEmph(elem *eval.EmphElement) {
	// Emphasis is an inline style - collect its content.
	c.collectContent(&elem.Content)
}

// collectLink handles link elements.
func (c *Collector) collectLink(elem *eval.LinkElement) {
	// Links are inline elements with just a URL.
	// They don't contribute flow children directly.
	c.lastWasSpacing = false
}

// collectRef handles reference elements.
func (c *Collector) collectRef(elem *eval.RefElement) {
	// References are inline elements.
	c.lastWasSpacing = false
}

// collectEquation handles equation (math) elements.
func (c *Collector) collectEquation(elem *eval.EquationElement) {
	if elem.Block {
		// Display math is a single block.
		align := Axes[FixedAlignment]{X: FixedAlignCenter, Y: FixedAlignStart}
		c.children = append(c.children, &SingleChild{
			Align:  align,
			Sticky: false,
			Alone:  false,
		})
	}
	// Inline math is handled as part of paragraph content.
	c.lastWasSpacing = false
}

// collectSmartQuote handles smart quote elements.
func (c *Collector) collectSmartQuote(elem *eval.SmartQuoteElement) {
	// Smart quotes are inline elements.
	c.lastWasSpacing = false
}

// collectUnknown handles unknown/unsupported elements.
func (c *Collector) collectUnknown(elem eval.ContentElement) {
	// Treat unknown elements as single unbreakable blocks.
	align := c.getBlockAlignment()
	c.children = append(c.children, &SingleChild{
		Align:  align,
		Sticky: false,
		Alone:  false,
	})
	c.lastWasSpacing = false
}

// addRelSpacing adds relative spacing to the children.
func (c *Collector) addRelSpacing(amount layout.Abs, weakness uint8) {
	c.children = append(c.children, RelChild{
		Amount:   Rel{Abs: amount},
		Weakness: weakness,
	})
	c.lastWasSpacing = true
}

// addFrSpacing adds fractional spacing to the children.
func (c *Collector) addFrSpacing(amount layout.Fr, weakness uint8) {
	c.children = append(c.children, FrChild{
		Amount:   amount,
		Weakness: weakness,
	})
	c.lastWasSpacing = true
}

// addTag adds an introspection tag.
func (c *Collector) addTag(tag *Tag) {
	c.children = append(c.children, TagChild{Tag: tag})
}

// addBreak adds a column/page break.
func (c *Collector) addBreak(weak bool) {
	c.children = append(c.children, BreakChild{Weak: weak})
}

// addFlush adds a float flush directive.
func (c *Collector) addFlush() {
	c.children = append(c.children, FlushChild{})
}

// getParSpacing returns the paragraph spacing from styles.
func (c *Collector) getParSpacing() layout.Abs {
	// Default paragraph spacing (approximately 0.65em at 11pt)
	// This should come from styles in a full implementation.
	return layout.Abs(7.15) // ~0.65em at 11pt
}

// getParagraphAlignment returns the alignment for paragraphs.
func (c *Collector) getParagraphAlignment() Axes[FixedAlignment] {
	// Default to start alignment.
	// This should come from styles in a full implementation.
	return Axes[FixedAlignment]{X: FixedAlignStart, Y: FixedAlignStart}
}

// getBlockAlignment returns the default block alignment.
func (c *Collector) getBlockAlignment() Axes[FixedAlignment] {
	// Default to start alignment.
	return Axes[FixedAlignment]{X: FixedAlignStart, Y: FixedAlignStart}
}

// CollectWithStyles collects content with the given style chain.
// Returns preprocessed children ready for distribution.
func CollectWithStyles(engine *Engine, content *eval.Content, mode FlowMode, styles StyleChain) *Work {
	locator := &Locator{}
	children := Collect(engine, content, mode, styles, locator)
	return NewWork(children)
}
