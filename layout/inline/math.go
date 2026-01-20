// Package inline provides inline/paragraph layout including text shaping.
package inline

// MathConstants contains positioning constants for math layout.
// These are based on typical OpenType MATH table values.
type MathConstants struct {
	// ScriptPercentScaleDown is the scale factor for scripts (e.g., 0.7 = 70%).
	ScriptPercentScaleDown float64
	// ScriptScriptPercentScaleDown is the scale for nested scripts.
	ScriptScriptPercentScaleDown float64
	// SuperscriptShiftUp is the baseline shift for superscripts (in em).
	SuperscriptShiftUp Em
	// SuperscriptShiftUpCramped is the shift in cramped style.
	SuperscriptShiftUpCramped Em
	// SubscriptShiftDown is the baseline shift for subscripts (in em).
	SubscriptShiftDown Em
	// SuperscriptBottomMin is the minimum distance from superscript bottom to baseline.
	SuperscriptBottomMin Em
	// SubscriptTopMax is the maximum height of subscript above baseline.
	SubscriptTopMax Em
	// SubSuperscriptGapMin is the minimum gap between sub and superscript.
	SubSuperscriptGapMin Em
	// UpperLimitGapMin is the minimum gap between nucleus and upper limit.
	UpperLimitGapMin Em
	// UpperLimitBaselineRiseMin is the minimum rise of upper limit baseline.
	UpperLimitBaselineRiseMin Em
	// LowerLimitGapMin is the minimum gap between nucleus and lower limit.
	LowerLimitGapMin Em
	// LowerLimitBaselineDropMin is the minimum drop of lower limit baseline.
	LowerLimitBaselineDropMin Em
}

// DefaultMathConstants returns default math constants.
// These approximate typical values from Computer Modern and Latin Modern fonts.
func DefaultMathConstants() MathConstants {
	return MathConstants{
		ScriptPercentScaleDown:       0.7,
		ScriptScriptPercentScaleDown: 0.5,
		SuperscriptShiftUp:           0.45,
		SuperscriptShiftUpCramped:    0.35,
		SubscriptShiftDown:           0.25,
		SuperscriptBottomMin:         0.25,
		SubscriptTopMax:              0.8,
		SubSuperscriptGapMin:         0.2,
		UpperLimitGapMin:             0.15,
		UpperLimitBaselineRiseMin:    0.3,
		LowerLimitGapMin:             0.15,
		LowerLimitBaselineDropMin:    0.6,
	}
}

// MathContext holds context for math layout operations.
type MathContext struct {
	// Constants contains positioning constants.
	Constants MathConstants
	// FontSize is the current font size.
	FontSize Abs
	// Cramped indicates cramped style (tighter superscript positioning).
	Cramped bool
}

// NewMathContext creates a new math context with defaults.
func NewMathContext(fontSize Abs) *MathContext {
	return &MathContext{
		Constants: DefaultMathConstants(),
		FontSize:  fontSize,
		Cramped:   false,
	}
}

// ScriptFontSize returns the font size for scripts.
func (ctx *MathContext) ScriptFontSize() Abs {
	return ctx.FontSize * Abs(ctx.Constants.ScriptPercentScaleDown)
}

// ScriptScriptFontSize returns the font size for nested scripts.
func (ctx *MathContext) ScriptScriptFontSize() Abs {
	return ctx.FontSize * Abs(ctx.Constants.ScriptScriptPercentScaleDown)
}

// LayoutMathScript lays out a base with superscript and/or subscript.
// It returns a FinalFrame containing the positioned base and scripts.
func LayoutMathScript(
	ctx *MathContext,
	base *FinalFrame,
	superscript *FinalFrame,
	subscript *FinalFrame,
	primes int,
) *FinalFrame {
	if base == nil {
		return nil
	}

	c := ctx.Constants

	// Calculate base dimensions
	baseWidth := base.Size.Width
	baseHeight := base.Size.Height
	baseBaseline := base.Baseline

	// Script positioning calculations
	var superShift, subShift Abs
	scriptX := baseWidth

	// Add prime marks width if present
	primeWidth := Abs(0)
	if primes > 0 {
		// Each prime is approximately 0.3em wide
		primeWidth = Abs(float64(primes) * 0.3 * float64(ctx.FontSize))
	}

	// Calculate superscript position
	if superscript != nil {
		superscriptHeight := superscript.Size.Height
		superscriptBaseline := superscript.Baseline

		// Shift up from baseline
		if ctx.Cramped {
			superShift = c.SuperscriptShiftUpCramped.At(ctx.FontSize)
		} else {
			superShift = c.SuperscriptShiftUp.At(ctx.FontSize)
		}

		// Ensure minimum distance from superscript bottom to baseline
		bottomMin := c.SuperscriptBottomMin.At(ctx.FontSize)
		superBottom := baseBaseline - superShift + superscriptHeight - superscriptBaseline
		if superBottom < bottomMin {
			// Raise superscript to meet minimum
			superShift += bottomMin - superBottom
		}
	}

	// Calculate subscript position
	if subscript != nil {
		subscriptBaseline := subscript.Baseline

		// Shift down from baseline
		subShift = c.SubscriptShiftDown.At(ctx.FontSize)

		// Ensure subscript top doesn't go above threshold
		topMax := c.SubscriptTopMax.At(ctx.FontSize)
		subTop := baseBaseline + subShift - subscriptBaseline
		if subTop > topMax {
			// Lower subscript to meet maximum
			subShift += subTop - topMax
		}
	}

	// If both scripts present, ensure minimum gap
	if superscript != nil && subscript != nil {
		superscriptHeight := superscript.Size.Height
		superscriptBaseline := superscript.Baseline
		subscriptBaseline := subscript.Baseline

		superBottom := baseBaseline - superShift + superscriptHeight - superscriptBaseline
		subTop := baseBaseline + subShift - subscriptBaseline

		gap := subTop - superBottom
		minGap := c.SubSuperscriptGapMin.At(ctx.FontSize)

		if gap < minGap {
			// Split the difference
			adjust := (minGap - gap) / 2
			superShift += adjust
			subShift += adjust
		}
	}

	// Calculate total dimensions
	totalWidth := baseWidth + primeWidth
	var scriptWidth Abs
	if superscript != nil && superscript.Size.Width > scriptWidth {
		scriptWidth = superscript.Size.Width
	}
	if subscript != nil && subscript.Size.Width > scriptWidth {
		scriptWidth = subscript.Size.Width
	}
	totalWidth += scriptWidth

	// Calculate height bounds
	top := baseBaseline
	bottom := baseHeight - baseBaseline

	if superscript != nil {
		superTop := superShift + superscript.Baseline
		if superTop > top {
			top = superTop
		}
	}

	if subscript != nil {
		subBottom := subShift + subscript.Size.Height - subscript.Baseline
		if subBottom > bottom {
			bottom = subBottom
		}
	}

	totalHeight := top + bottom
	baseline := top

	// Create the output frame
	output := &FinalFrame{
		Size:     FinalSize{Width: totalWidth, Height: totalHeight},
		Baseline: baseline,
	}

	// Position base (centered on baseline)
	baseY := baseline - baseBaseline
	output.PushFrame(FinalPoint{X: 0, Y: baseY}, base)

	// Position scripts
	scriptX += primeWidth

	if superscript != nil {
		superY := baseline - superShift - superscript.Baseline
		output.PushFrame(FinalPoint{X: scriptX, Y: superY}, superscript)
	}

	if subscript != nil {
		subY := baseline + subShift - subscript.Baseline
		output.PushFrame(FinalPoint{X: scriptX, Y: subY}, subscript)
	}

	// Add prime marks as text if present
	if primes > 0 {
		// Prime marks are positioned at superscript height
		primeStr := ""
		for i := 0; i < primes; i++ {
			primeStr += "′" // Unicode prime character
		}

		// Create a simple prime item - in production this would be properly shaped
		// For now, we rely on the prime being part of the base or superscript content
		// The prime width is already accounted for in the total width
	}

	return output
}

// LayoutMathLimits lays out an operator with limits above and below.
// This is used for operators like ∑, ∏, ∫ in display style.
func LayoutMathLimits(
	ctx *MathContext,
	nucleus *FinalFrame,
	upper *FinalFrame,
	lower *FinalFrame,
) *FinalFrame {
	if nucleus == nil {
		return nil
	}

	c := ctx.Constants

	// Calculate nucleus dimensions
	nucleusWidth := nucleus.Size.Width
	nucleusHeight := nucleus.Size.Height
	nucleusBaseline := nucleus.Baseline

	// Determine the widest element for centering
	maxWidth := nucleusWidth
	if upper != nil && upper.Size.Width > maxWidth {
		maxWidth = upper.Size.Width
	}
	if lower != nil && lower.Size.Width > maxWidth {
		maxWidth = lower.Size.Width
	}

	// Calculate positions
	var upperGap, lowerGap Abs
	var totalHeight Abs
	var baseline Abs

	nucleusY := Abs(0)

	if upper != nil {
		upperGap = c.UpperLimitGapMin.At(ctx.FontSize)
		upperHeight := upper.Size.Height
		nucleusY = upperHeight + upperGap
	}

	if lower != nil {
		lowerGap = c.LowerLimitGapMin.At(ctx.FontSize)
	}

	// Calculate total height
	totalHeight = nucleusY + nucleusHeight
	if lower != nil {
		totalHeight += lowerGap + lower.Size.Height
	}

	// Baseline is at the nucleus baseline
	baseline = nucleusY + nucleusBaseline

	// Create the output frame
	output := &FinalFrame{
		Size:     FinalSize{Width: maxWidth, Height: totalHeight},
		Baseline: baseline,
	}

	// Center and position nucleus
	nucleusX := (maxWidth - nucleusWidth) / 2
	output.PushFrame(FinalPoint{X: nucleusX, Y: nucleusY}, nucleus)

	// Position upper limit (centered above nucleus)
	if upper != nil {
		upperX := (maxWidth - upper.Size.Width) / 2
		upperY := nucleusY - upperGap - upper.Size.Height
		output.PushFrame(FinalPoint{X: upperX, Y: upperY}, upper)
	}

	// Position lower limit (centered below nucleus)
	if lower != nil {
		lowerX := (maxWidth - lower.Size.Width) / 2
		lowerY := nucleusY + nucleusHeight + lowerGap
		output.PushFrame(FinalPoint{X: lowerX, Y: lowerY}, lower)
	}

	return output
}

// IsLargeOperator returns true if the symbol should use limits positioning.
// These are operators that traditionally have limits above/below in display style.
func IsLargeOperator(symbol string) bool {
	switch symbol {
	case "∑", "∏", "∐", // Summation, product, coproduct
		"⋀", "⋁", "⋂", "⋃", // Logical and set operators
		"⨁", "⨂", "⨀", // Circle operators
		"lim", "sup", "inf", "max", "min", // Named limits
		"limsup", "liminf":
		return true
	default:
		return false
	}
}
