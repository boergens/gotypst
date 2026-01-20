package inline

import "testing"

func TestMathClassString(t *testing.T) {
	tests := []struct {
		class MathClass
		want  string
	}{
		{MathClassNone, "none"},
		{MathClassNormal, "normal"},
		{MathClassLarge, "large"},
		{MathClassBinary, "binary"},
		{MathClassRelation, "relation"},
		{MathClassOpening, "opening"},
		{MathClassClosing, "closing"},
		{MathClassPunctuation, "punctuation"},
		{MathClassFence, "fence"},
		{MathClassGlyphVariant, "glyph-variant"},
		{MathClassSpace, "space"},
		{MathClass(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.class.String(); got != tt.want {
				t.Errorf("MathClass.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMathClassIsOperator(t *testing.T) {
	tests := []struct {
		class MathClass
		want  bool
	}{
		{MathClassLarge, true},
		{MathClassBinary, true},
		{MathClassNormal, false},
		{MathClassRelation, false},
		{MathClassOpening, false},
	}

	for _, tt := range tests {
		t.Run(tt.class.String(), func(t *testing.T) {
			if got := tt.class.IsOperator(); got != tt.want {
				t.Errorf("MathClass.IsOperator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMathClassIsDelimiter(t *testing.T) {
	tests := []struct {
		class MathClass
		want  bool
	}{
		{MathClassOpening, true},
		{MathClassClosing, true},
		{MathClassFence, true},
		{MathClassNormal, false},
		{MathClassBinary, false},
	}

	for _, tt := range tests {
		t.Run(tt.class.String(), func(t *testing.T) {
			if got := tt.class.IsDelimiter(); got != tt.want {
				t.Errorf("MathClass.IsDelimiter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMathGlyphFragment(t *testing.T) {
	g := NewMathGlyphFragment(65, 'A', 12.0)

	if g.GlyphID() != 65 {
		t.Errorf("GlyphID() = %v, want %v", g.GlyphID(), 65)
	}
	if g.Char() != 'A' {
		t.Errorf("Char() = %v, want %v", g.Char(), 'A')
	}
	if g.FontSize() != 12.0 {
		t.Errorf("FontSize() = %v, want %v", g.FontSize(), 12.0)
	}
	if g.Class() != MathClassNormal {
		t.Errorf("Class() = %v, want %v", g.Class(), MathClassNormal)
	}
}

func TestMathGlyphFragmentMetrics(t *testing.T) {
	g := NewMathGlyphFragment(65, 'A', 12.0)
	g.SetWidth(6.0)
	g.SetAscent(8.0)
	g.SetDescent(2.0)
	g.SetItalicsCorrection(0.5)

	if g.Width() != 6.0 {
		t.Errorf("Width() = %v, want %v", g.Width(), 6.0)
	}
	if g.Ascent() != 8.0 {
		t.Errorf("Ascent() = %v, want %v", g.Ascent(), 8.0)
	}
	if g.Height() != 8.0 {
		t.Errorf("Height() = %v, want %v", g.Height(), 8.0)
	}
	if g.Descent() != 2.0 {
		t.Errorf("Descent() = %v, want %v", g.Descent(), 2.0)
	}
	if g.Depth() != 2.0 {
		t.Errorf("Depth() = %v, want %v", g.Depth(), 2.0)
	}
	if g.ItalicsCorrection() != 0.5 {
		t.Errorf("ItalicsCorrection() = %v, want %v", g.ItalicsCorrection(), 0.5)
	}
	if g.TotalHeight() != 10.0 {
		t.Errorf("TotalHeight() = %v, want %v", g.TotalHeight(), 10.0)
	}
}

func TestMathGlyphFragmentWithMetrics(t *testing.T) {
	g := NewMathGlyphFragment(65, 'A', 12.0)
	g2 := g.WithMetrics(7.0, 9.0, 3.0, 1.0)

	// Original unchanged
	if g.Width() != 0 {
		t.Errorf("Original Width changed")
	}

	// Clone has new values
	if g2.Width() != 7.0 {
		t.Errorf("Clone Width() = %v, want %v", g2.Width(), 7.0)
	}
	if g2.Ascent() != 9.0 {
		t.Errorf("Clone Ascent() = %v, want %v", g2.Ascent(), 9.0)
	}
	if g2.Descent() != 3.0 {
		t.Errorf("Clone Descent() = %v, want %v", g2.Descent(), 3.0)
	}
	if g2.ItalicsCorrection() != 1.0 {
		t.Errorf("Clone ItalicsCorrection() = %v, want %v", g2.ItalicsCorrection(), 1.0)
	}

	// Preserved fields
	if g2.GlyphID() != 65 {
		t.Errorf("Clone GlyphID() = %v, want %v", g2.GlyphID(), 65)
	}
	if g2.Char() != 'A' {
		t.Errorf("Clone Char() = %v, want %v", g2.Char(), 'A')
	}
}

func TestMathGlyphFragmentFlags(t *testing.T) {
	g := NewMathGlyphFragment(65, 'A', 12.0)

	if g.IsAccent() {
		t.Error("IsAccent() should be false by default")
	}
	if g.IsStretchable() {
		t.Error("IsStretchable() should be false by default")
	}

	g.SetAccent(true)
	g.SetStretchable(true)

	if !g.IsAccent() {
		t.Error("IsAccent() should be true after SetAccent(true)")
	}
	if !g.IsStretchable() {
		t.Error("IsStretchable() should be true after SetStretchable(true)")
	}
}

func TestNewMathFrameFragment(t *testing.T) {
	frame := &FinalFrame{
		Size:     FinalSize{Width: 20.0, Height: 15.0},
		Baseline: 10.0,
	}

	f := NewMathFrameFragment(frame)

	if f.Frame() != frame {
		t.Error("Frame() should return the original frame")
	}
	if f.Class() != MathClassNone {
		t.Errorf("Class() = %v, want %v", f.Class(), MathClassNone)
	}
	if f.Width() != 20.0 {
		t.Errorf("Width() = %v, want %v", f.Width(), 20.0)
	}
	if f.Ascent() != 10.0 {
		t.Errorf("Ascent() = %v, want %v", f.Ascent(), 10.0)
	}
	if f.Height() != 10.0 {
		t.Errorf("Height() = %v, want %v", f.Height(), 10.0)
	}
	if f.Descent() != 5.0 {
		t.Errorf("Descent() = %v, want %v", f.Descent(), 5.0)
	}
	if f.Depth() != 5.0 {
		t.Errorf("Depth() = %v, want %v", f.Depth(), 5.0)
	}
	if f.TotalHeight() != 15.0 {
		t.Errorf("TotalHeight() = %v, want %v", f.TotalHeight(), 15.0)
	}
}

func TestNewMathFrameFragmentNil(t *testing.T) {
	f := NewMathFrameFragment(nil)

	if f.Frame() == nil {
		t.Error("Frame() should not be nil even for nil input")
	}
	if f.Width() != 0 {
		t.Errorf("Width() = %v, want %v", f.Width(), 0)
	}
}

func TestMathFrameFragmentSetters(t *testing.T) {
	frame := &FinalFrame{
		Size:     FinalSize{Width: 20.0, Height: 15.0},
		Baseline: 10.0,
	}
	f := NewMathFrameFragment(frame)

	f.SetClass(MathClassLarge)
	if f.Class() != MathClassLarge {
		t.Errorf("Class() = %v, want %v", f.Class(), MathClassLarge)
	}

	f.SetItalicsCorrection(2.0)
	if f.ItalicsCorrection() != 2.0 {
		t.Errorf("ItalicsCorrection() = %v, want %v", f.ItalicsCorrection(), 2.0)
	}

	f.SetBaseline(12.0, 8.0)
	if f.Ascent() != 12.0 {
		t.Errorf("Ascent() = %v, want %v", f.Ascent(), 12.0)
	}
	if f.Descent() != 8.0 {
		t.Errorf("Descent() = %v, want %v", f.Descent(), 8.0)
	}
}

func TestNewMathSpaceFragment(t *testing.T) {
	s := NewMathSpaceFragment(5.0)

	if s.Class() != MathClassSpace {
		t.Errorf("Class() = %v, want %v", s.Class(), MathClassSpace)
	}
	if s.Width() != 5.0 {
		t.Errorf("Width() = %v, want %v", s.Width(), 5.0)
	}
	if s.Height() != 0 {
		t.Errorf("Height() = %v, want %v", s.Height(), 0)
	}
	if s.Depth() != 0 {
		t.Errorf("Depth() = %v, want %v", s.Depth(), 0)
	}
	if s.Ascent() != 0 {
		t.Errorf("Ascent() = %v, want %v", s.Ascent(), 0)
	}
	if s.Descent() != 0 {
		t.Errorf("Descent() = %v, want %v", s.Descent(), 0)
	}
	if s.ItalicsCorrection() != 0 {
		t.Errorf("ItalicsCorrection() = %v, want %v", s.ItalicsCorrection(), 0)
	}
}

func TestMathLinebreakFragment(t *testing.T) {
	l := &MathLinebreakFragment{}

	if l.Class() != MathClassNone {
		t.Errorf("Class() = %v, want %v", l.Class(), MathClassNone)
	}
	if l.Width() != 0 {
		t.Errorf("Width() = %v, want %v", l.Width(), 0)
	}
	if l.Height() != 0 {
		t.Errorf("Height() = %v, want %v", l.Height(), 0)
	}
	if l.ItalicsCorrection() != 0 {
		t.Errorf("ItalicsCorrection() = %v, want %v", l.ItalicsCorrection(), 0)
	}
}

func TestMathAlignFragment(t *testing.T) {
	a := &MathAlignFragment{}

	if a.Class() != MathClassNone {
		t.Errorf("Class() = %v, want %v", a.Class(), MathClassNone)
	}
	if a.Width() != 0 {
		t.Errorf("Width() = %v, want %v", a.Width(), 0)
	}
	if a.Height() != 0 {
		t.Errorf("Height() = %v, want %v", a.Height(), 0)
	}
	if a.ItalicsCorrection() != 0 {
		t.Errorf("ItalicsCorrection() = %v, want %v", a.ItalicsCorrection(), 0)
	}
}

func TestMathFragmentInterface(t *testing.T) {
	// Verify all fragment types implement MathFragment
	var _ MathFragment = &MathGlyphFragment{}
	var _ MathFragment = &MathFrameFragment{}
	var _ MathFragment = &MathSpaceFragment{}
	var _ MathFragment = &MathLinebreakFragment{}
	var _ MathFragment = &MathAlignFragment{}
}

func TestClassifyMathChar(t *testing.T) {
	tests := []struct {
		char rune
		want MathClass
	}{
		// Binary operators
		{'+', MathClassBinary},
		{'-', MathClassBinary},
		{'×', MathClassBinary},
		{'÷', MathClassBinary},
		// Relation symbols
		{'=', MathClassRelation},
		{'<', MathClassRelation},
		{'>', MathClassRelation},
		{'≤', MathClassRelation},
		{'≠', MathClassRelation},
		// Large operators
		{'∑', MathClassLarge},
		{'∏', MathClassLarge},
		{'∫', MathClassLarge},
		// Opening delimiters
		{'(', MathClassOpening},
		{'[', MathClassOpening},
		{'{', MathClassOpening},
		// Closing delimiters
		{')', MathClassClosing},
		{']', MathClassClosing},
		{'}', MathClassClosing},
		// Punctuation
		{',', MathClassPunctuation},
		{';', MathClassPunctuation},
		// Fence
		{'|', MathClassFence},
		// Normal (default)
		{'A', MathClassNormal},
		{'x', MathClassNormal},
		{'1', MathClassNormal},
		{'α', MathClassNormal},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			if got := ClassifyMathChar(tt.char); got != tt.want {
				t.Errorf("ClassifyMathChar(%q) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

func TestMathSpacing(t *testing.T) {
	// Test spacing between various math classes at script level 0
	tests := []struct {
		left, right MathClass
		wantNonZero bool
	}{
		// Binary operator spacing
		{MathClassNormal, MathClassBinary, true},
		{MathClassBinary, MathClassNormal, true},
		// Relation spacing
		{MathClassNormal, MathClassRelation, true},
		{MathClassRelation, MathClassNormal, true},
		// No spacing after opening
		{MathClassOpening, MathClassNormal, false},
		{MathClassOpening, MathClassBinary, false},
		// Large operator spacing
		{MathClassNormal, MathClassLarge, true},
		{MathClassLarge, MathClassNormal, true},
	}

	for _, tt := range tests {
		name := tt.left.String() + "-" + tt.right.String()
		t.Run(name, func(t *testing.T) {
			space := MathSpacing(tt.left, tt.right, 0)
			if tt.wantNonZero && space == 0 {
				t.Errorf("MathSpacing(%v, %v, 0) = 0, want non-zero", tt.left, tt.right)
			}
			if !tt.wantNonZero && space != 0 {
				t.Errorf("MathSpacing(%v, %v, 0) = %v, want 0", tt.left, tt.right, space)
			}
		})
	}
}

func TestMathSpacingScriptLevel(t *testing.T) {
	// At script level > 0, spacing should be 0
	left := MathClassNormal
	right := MathClassBinary

	if space := MathSpacing(left, right, 1); space != 0 {
		t.Errorf("MathSpacing at script level 1 = %v, want 0", space)
	}
	if space := MathSpacing(left, right, 2); space != 0 {
		t.Errorf("MathSpacing at script level 2 = %v, want 0", space)
	}
}
