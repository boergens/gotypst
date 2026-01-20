package eval

// ----------------------------------------------------------------------------
// Text Element and Styling
// ----------------------------------------------------------------------------

// TextElement represents styled text content.
type TextElement struct {
	// Text is the text content.
	Text string

	// Font is the font family or families to use.
	// Can be a single font name or a list for fallback.
	Font []string

	// Size is the font size in points. Zero means inherit.
	Size float64

	// Weight is the font weight (100-900). Zero means inherit.
	Weight FontWeight

	// Style is the font style (normal, italic, oblique).
	Style FontStyle

	// Fill is the text fill color. Nil means inherit.
	Fill *Color

	// Stroke is the text stroke. Nil means no stroke.
	Stroke *Stroke

	// Underline indicates whether text should be underlined.
	Underline *TextDecoration

	// Strikethrough indicates whether text should have strikethrough.
	Strikethrough *TextDecoration

	// Overline indicates whether text should have an overline.
	Overline *TextDecoration
}

func (*TextElement) isContentElement() {}

// ----------------------------------------------------------------------------
// Font Properties
// ----------------------------------------------------------------------------

// FontWeight represents font weight values (100-900).
type FontWeight int

const (
	// FontWeightInherit means inherit from parent context.
	FontWeightInherit FontWeight = 0

	// FontWeightThin is weight 100.
	FontWeightThin FontWeight = 100

	// FontWeightExtraLight is weight 200.
	FontWeightExtraLight FontWeight = 200

	// FontWeightLight is weight 300.
	FontWeightLight FontWeight = 300

	// FontWeightRegular is weight 400 (normal).
	FontWeightRegular FontWeight = 400

	// FontWeightMedium is weight 500.
	FontWeightMedium FontWeight = 500

	// FontWeightSemiBold is weight 600.
	FontWeightSemiBold FontWeight = 600

	// FontWeightBold is weight 700.
	FontWeightBold FontWeight = 700

	// FontWeightExtraBold is weight 800.
	FontWeightExtraBold FontWeight = 800

	// FontWeightBlack is weight 900.
	FontWeightBlack FontWeight = 900
)

// FontWeightFromString converts a string to a FontWeight.
func FontWeightFromString(s string) FontWeight {
	switch s {
	case "thin":
		return FontWeightThin
	case "extralight", "extra-light":
		return FontWeightExtraLight
	case "light":
		return FontWeightLight
	case "regular", "normal":
		return FontWeightRegular
	case "medium":
		return FontWeightMedium
	case "semibold", "semi-bold":
		return FontWeightSemiBold
	case "bold":
		return FontWeightBold
	case "extrabold", "extra-bold":
		return FontWeightExtraBold
	case "black":
		return FontWeightBlack
	default:
		return FontWeightInherit
	}
}

// FontStyle represents font style (normal, italic, oblique).
type FontStyle int

const (
	// FontStyleInherit means inherit from parent context.
	FontStyleInherit FontStyle = iota

	// FontStyleNormal is the normal (upright) style.
	FontStyleNormal

	// FontStyleItalic is the italic style.
	FontStyleItalic

	// FontStyleOblique is the oblique (slanted) style.
	FontStyleOblique
)

// FontStyleFromString converts a string to a FontStyle.
func FontStyleFromString(s string) FontStyle {
	switch s {
	case "normal":
		return FontStyleNormal
	case "italic":
		return FontStyleItalic
	case "oblique":
		return FontStyleOblique
	default:
		return FontStyleInherit
	}
}

// ----------------------------------------------------------------------------
// Stroke Properties
// ----------------------------------------------------------------------------

// Stroke represents text stroke properties.
type Stroke struct {
	// Paint is the stroke color or gradient.
	Paint *Color

	// Thickness is the stroke width in points.
	Thickness float64

	// Cap is the line cap style.
	Cap LineCap

	// Join is the line join style.
	Join LineJoin

	// Dash is the optional dash pattern.
	Dash *DashPattern
}

// LineCap represents line cap styles for strokes.
type LineCap int

const (
	// LineCapButt is a flat cap at the endpoint.
	LineCapButt LineCap = iota

	// LineCapRound is a rounded cap.
	LineCapRound

	// LineCapSquare is a square cap extending beyond the endpoint.
	LineCapSquare
)

// LineCapFromString converts a string to a LineCap.
func LineCapFromString(s string) LineCap {
	switch s {
	case "butt":
		return LineCapButt
	case "round":
		return LineCapRound
	case "square":
		return LineCapSquare
	default:
		return LineCapButt
	}
}

// LineJoin represents line join styles for strokes.
type LineJoin int

const (
	// LineJoinMiter is a sharp corner.
	LineJoinMiter LineJoin = iota

	// LineJoinRound is a rounded corner.
	LineJoinRound

	// LineJoinBevel is a beveled corner.
	LineJoinBevel
)

// LineJoinFromString converts a string to a LineJoin.
func LineJoinFromString(s string) LineJoin {
	switch s {
	case "miter":
		return LineJoinMiter
	case "round":
		return LineJoinRound
	case "bevel":
		return LineJoinBevel
	default:
		return LineJoinMiter
	}
}

// DashPattern represents a stroke dash pattern.
type DashPattern struct {
	// Array is the dash array (lengths of dashes and gaps).
	Array []float64

	// Phase is the dash phase offset.
	Phase float64
}

// ----------------------------------------------------------------------------
// Text Decoration
// ----------------------------------------------------------------------------

// TextDecoration represents decoration properties for text.
type TextDecoration struct {
	// Stroke is the decoration line stroke. Nil uses the text fill color.
	Stroke *Stroke

	// Offset is the optional offset from the default position.
	// Positive moves down for underline, up for overline.
	Offset *float64

	// Evade controls whether the decoration evades glyph descenders/ascenders.
	// Applicable for underline and overline.
	Evade bool

	// Background controls whether the decoration is drawn behind the text.
	Background bool
}

// NewUnderline creates a simple underline decoration.
func NewUnderline() *TextDecoration {
	return &TextDecoration{
		Evade: true,
	}
}

// NewStrikethrough creates a simple strikethrough decoration.
func NewStrikethrough() *TextDecoration {
	return &TextDecoration{}
}

// NewOverline creates a simple overline decoration.
func NewOverline() *TextDecoration {
	return &TextDecoration{
		Evade: true,
	}
}

// ----------------------------------------------------------------------------
// Helper Functions
// ----------------------------------------------------------------------------

// NewTextElement creates a new TextElement with just text content.
func NewTextElement(text string) *TextElement {
	return &TextElement{Text: text}
}

// WithFont returns a copy of the TextElement with the specified font.
func (t *TextElement) WithFont(fonts ...string) *TextElement {
	result := *t
	result.Font = fonts
	return &result
}

// WithSize returns a copy of the TextElement with the specified size.
func (t *TextElement) WithSize(size float64) *TextElement {
	result := *t
	result.Size = size
	return &result
}

// WithWeight returns a copy of the TextElement with the specified weight.
func (t *TextElement) WithWeight(weight FontWeight) *TextElement {
	result := *t
	result.Weight = weight
	return &result
}

// WithStyle returns a copy of the TextElement with the specified style.
func (t *TextElement) WithStyle(style FontStyle) *TextElement {
	result := *t
	result.Style = style
	return &result
}

// WithFill returns a copy of the TextElement with the specified fill color.
func (t *TextElement) WithFill(color Color) *TextElement {
	result := *t
	result.Fill = &color
	return &result
}

// WithStroke returns a copy of the TextElement with the specified stroke.
func (t *TextElement) WithStroke(stroke Stroke) *TextElement {
	result := *t
	result.Stroke = &stroke
	return &result
}

// WithUnderline returns a copy of the TextElement with underline decoration.
func (t *TextElement) WithUnderline(deco *TextDecoration) *TextElement {
	result := *t
	result.Underline = deco
	return &result
}

// WithStrikethrough returns a copy of the TextElement with strikethrough decoration.
func (t *TextElement) WithStrikethrough(deco *TextDecoration) *TextElement {
	result := *t
	result.Strikethrough = deco
	return &result
}

// WithOverline returns a copy of the TextElement with overline decoration.
func (t *TextElement) WithOverline(deco *TextDecoration) *TextElement {
	result := *t
	result.Overline = deco
	return &result
}

// HasStyling returns true if the text element has any styling applied.
func (t *TextElement) HasStyling() bool {
	return len(t.Font) > 0 ||
		t.Size != 0 ||
		t.Weight != FontWeightInherit ||
		t.Style != FontStyleInherit ||
		t.Fill != nil ||
		t.Stroke != nil ||
		t.Underline != nil ||
		t.Strikethrough != nil ||
		t.Overline != nil
}

// Clone creates a deep copy of the TextElement.
func (t *TextElement) Clone() *TextElement {
	result := &TextElement{
		Text:   t.Text,
		Size:   t.Size,
		Weight: t.Weight,
		Style:  t.Style,
	}

	if len(t.Font) > 0 {
		result.Font = make([]string, len(t.Font))
		copy(result.Font, t.Font)
	}

	if t.Fill != nil {
		fill := *t.Fill
		result.Fill = &fill
	}

	if t.Stroke != nil {
		stroke := *t.Stroke
		result.Stroke = &stroke
	}

	if t.Underline != nil {
		deco := *t.Underline
		result.Underline = &deco
	}

	if t.Strikethrough != nil {
		deco := *t.Strikethrough
		result.Strikethrough = &deco
	}

	if t.Overline != nil {
		deco := *t.Overline
		result.Overline = &deco
	}

	return result
}
