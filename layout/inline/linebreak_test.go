package inline

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestBreakpointInfo(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		bp := Normal()
		if bp.IsHyphen() {
			t.Error("Normal breakpoint should not be hyphen")
		}
		if bp.IsMandatory() {
			t.Error("Normal breakpoint should not be mandatory")
		}
	})

	t.Run("Mandatory", func(t *testing.T) {
		bp := Mandatory()
		if bp.IsHyphen() {
			t.Error("Mandatory breakpoint should not be hyphen")
		}
		if !bp.IsMandatory() {
			t.Error("Mandatory breakpoint should be mandatory")
		}
	})

	t.Run("Hyphen", func(t *testing.T) {
		bp := Hyphen(3, 4)
		if !bp.IsHyphen() {
			t.Error("Hyphen breakpoint should be hyphen")
		}
		if bp.IsMandatory() {
			t.Error("Hyphen breakpoint should not be mandatory")
		}
		if bp.Hyphen.Before != 3 || bp.Hyphen.After != 4 {
			t.Errorf("Hyphen counts wrong: got before=%d, after=%d", bp.Hyphen.Before, bp.Hyphen.After)
		}
	})
}

func TestTrimLine(t *testing.T) {
	tests := []struct {
		name          string
		bp            BreakpointInfo
		start         int
		line          string
		expectLayout  int
		expectShaping int
	}{
		{
			name:          "normal with trailing space",
			bp:            Normal(),
			start:         0,
			line:          "hello ",
			expectLayout:  5, // "hello"
			expectShaping: 6, // "hello "
		},
		{
			name:          "mandatory with newline",
			bp:            Mandatory(),
			start:         0,
			line:          "hello\n",
			expectLayout:  5,
			expectShaping: 5,
		},
		{
			name:          "hyphen trims nothing",
			bp:            Hyphen(3, 3),
			start:         0,
			line:          "hello",
			expectLayout:  5,
			expectShaping: 5,
		},
		{
			name:          "normal no trailing space",
			bp:            Normal(),
			start:         10,
			line:          "world",
			expectLayout:  15,
			expectShaping: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trim := tt.bp.TrimLine(tt.start, tt.line)
			if trim.Layout != tt.expectLayout {
				t.Errorf("Layout: got %d, want %d", trim.Layout, tt.expectLayout)
			}
			if trim.Shaping != tt.expectShaping {
				t.Errorf("Shaping: got %d, want %d", trim.Shaping, tt.expectShaping)
			}
		})
	}
}

func TestCumulativeVec(t *testing.T) {
	cv := newCumulativeVec[int](10)

	cv.push(3, 10)  // bytes 0-2 contribute 10
	cv.push(2, 20)  // bytes 3-4 contribute 20
	cv.adjust(10)   // fill to length 10

	// Test estimates
	tests := []struct {
		start, end int
		want       int
	}{
		{0, 3, 10},
		{0, 5, 30},
		{3, 5, 20},
		{0, 10, 30},
	}

	for _, tt := range tests {
		got := cv.estimate(tt.start, tt.end)
		if got != tt.want {
			t.Errorf("estimate(%d, %d): got %d, want %d", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestCumulativeVecAbs(t *testing.T) {
	cv := newCumulativeVec[layout.Abs](10)

	cv.push(2, layout.Abs(5.0))
	cv.push(3, layout.Abs(7.5))
	cv.adjust(8)

	got := cv.estimate(0, 5)
	want := layout.Abs(12.5)
	if got != want {
		t.Errorf("estimate(0, 5): got %v, want %v", got, want)
	}
}

func TestRawCost(t *testing.T) {
	metrics := &CostMetrics{
		minRatio:       MinRatio,
		minApproxRatio: MinApproxRatio,
		hyphCost:       DefaultHyphCost,
		runtCost:       DefaultRuntCost,
	}

	t.Run("overfull line", func(t *testing.T) {
		cost := rawCost(metrics, Normal(), -2.0, false, false, false, false)
		// Should be very high (1_000_000 badness)
		if cost < 1_000_000 {
			t.Errorf("Overfull line should have high cost, got %v", cost)
		}
	})

	t.Run("perfect fit", func(t *testing.T) {
		cost := rawCost(metrics, Mandatory(), 0.0, false, false, false, false)
		// Should be minimal (only base 1)
		if cost != 1.0 {
			t.Errorf("Perfect fit should have cost 1.0, got %v", cost)
		}
	})

	t.Run("hyphenation penalty", func(t *testing.T) {
		costWithoutHyphen := rawCost(metrics, Normal(), 0.5, false, false, false, false)
		costWithHyphen := rawCost(metrics, Hyphen(3, 3), 0.5, false, false, false, false)
		if costWithHyphen <= costWithoutHyphen {
			t.Error("Hyphenation should add cost")
		}
	})

	t.Run("consecutive dashes penalty", func(t *testing.T) {
		costWithout := rawCost(metrics, Normal(), 0.5, false, false, false, false)
		costWith := rawCost(metrics, Normal(), 0.5, false, false, true, false)
		if costWith <= costWithout {
			t.Error("Consecutive dashes should add cost")
		}
	})

	t.Run("runt penalty", func(t *testing.T) {
		costWithoutRunt := rawCost(metrics, Mandatory(), 0.0, false, false, false, false)
		costWithRunt := rawCost(metrics, Mandatory(), 0.0, false, true, false, false)
		if costWithRunt <= costWithoutRunt {
			t.Error("Runt should add cost")
		}
	})
}

func TestRawRatio(t *testing.T) {
	p := &Preparation{
		Config: &Config{
			FontSize: layout.Abs(12.0),
		},
	}

	t.Run("perfect fit", func(t *testing.T) {
		ratio := rawRatio(p, 100, 100, 10, 10, 5)
		if ratio != 0.0 {
			t.Errorf("Perfect fit should have ratio 0, got %v", ratio)
		}
	})

	t.Run("needs stretching", func(t *testing.T) {
		ratio := rawRatio(p, 100, 90, 20, 10, 5)
		if ratio <= 0 {
			t.Errorf("Underfull line should have positive ratio, got %v", ratio)
		}
	})

	t.Run("needs shrinking", func(t *testing.T) {
		ratio := rawRatio(p, 100, 110, 10, 20, 5)
		if ratio >= 0 {
			t.Errorf("Overfull line should have negative ratio, got %v", ratio)
		}
	})

	t.Run("clamped high", func(t *testing.T) {
		ratio := rawRatio(p, 1000, 10, 1, 1, 0)
		if ratio > 10.0 {
			t.Errorf("Ratio should be clamped to 10, got %v", ratio)
		}
	})
}

func TestBreakpointsFn(t *testing.T) {
	t.Run("empty text", func(t *testing.T) {
		p := &Preparation{
			Text:   "",
			Config: &Config{},
		}

		var breakpoints []BreakpointInfo
		breakpointsFn(p, func(end int, bp BreakpointInfo) {
			breakpoints = append(breakpoints, bp)
		})

		if len(breakpoints) != 1 {
			t.Fatalf("Expected 1 breakpoint for empty text, got %d", len(breakpoints))
		}
		if !breakpoints[0].IsMandatory() {
			t.Error("Single breakpoint should be mandatory")
		}
	})

	t.Run("simple text with spaces", func(t *testing.T) {
		p := &Preparation{
			Text:   "hello world",
			Config: &Config{},
		}

		var ends []int
		breakpointsFn(p, func(end int, bp BreakpointInfo) {
			ends = append(ends, end)
		})

		// Should have breaks after "hello " and at end
		if len(ends) < 2 {
			t.Fatalf("Expected at least 2 breakpoints, got %d", len(ends))
		}
	})

	t.Run("text with newline", func(t *testing.T) {
		p := &Preparation{
			Text:   "hello\nworld",
			Config: &Config{},
		}

		var mandatory int
		breakpointsFn(p, func(end int, bp BreakpointInfo) {
			if bp.IsMandatory() {
				mandatory++
			}
		})

		// Should have mandatory breaks at \n and end
		if mandatory != 2 {
			t.Errorf("Expected 2 mandatory breakpoints, got %d", mandatory)
		}
	})
}

func TestLinebreakSimple(t *testing.T) {
	// Create a simple preparation with some text items
	text := "Hello world this is a test"

	// Create text items for each word
	items := make([]PreparedItem, 0)
	offset := 0
	words := []string{"Hello ", "world ", "this ", "is ", "a ", "test"}
	for _, word := range words {
		shaped := &ShapedText{
			Text: word,
			Size: layout.Abs(12.0),
		}
		// Add glyphs with some width
		for i := range word {
			shaped.Glyphs = append(shaped.Glyphs, ShapedGlyph{
				XAdvance: layout.Em(0.5), // 6pt per char at 12pt font
				Size:     layout.Abs(12.0),
				Range:    Range{Start: i, End: i + 1},
			})
		}
		items = append(items, PreparedItem{
			Range: Range{Start: offset, End: offset + len(word)},
			Item:  &TextItem{shaped: shaped},
		})
		offset += len(word)
	}

	p := &Preparation{
		Text:   text,
		Items:  items,
		Config: &Config{
			Linebreaks: layout.LinebreaksSimple,
			FontSize:   layout.Abs(12.0),
			Costs:      DefaultCosts(),
		},
	}

	// Wide width - should fit on one line
	lines := Linebreak(p, layout.Abs(500))
	if len(lines) != 1 {
		t.Errorf("Wide width: expected 1 line, got %d", len(lines))
	}

	// Narrow width - should wrap
	lines = Linebreak(p, layout.Abs(50))
	if len(lines) < 2 {
		t.Errorf("Narrow width: expected multiple lines, got %d", len(lines))
	}
}

func TestLinebreakOptimized(t *testing.T) {
	text := "Hello world"

	items := make([]PreparedItem, 0)
	shaped := &ShapedText{
		Text: text,
		Size: layout.Abs(12.0),
	}
	for i := range text {
		shaped.Glyphs = append(shaped.Glyphs, ShapedGlyph{
			XAdvance:      layout.Em(0.5),
			Size:          layout.Abs(12.0),
			Range:         Range{Start: i, End: i + 1},
			IsJustifiable: text[i] == ' ',
		})
	}
	items = append(items, PreparedItem{
		Range: Range{Start: 0, End: len(text)},
		Item:  &TextItem{shaped: shaped},
	})

	p := &Preparation{
		Text:   text,
		Items:  items,
		Config: &Config{
			Linebreaks: layout.LinebreaksOptimized,
			FontSize:   layout.Abs(12.0),
			Justify:    true,
			Costs:      DefaultCosts(),
		},
	}

	// Should produce valid output
	lines := Linebreak(p, layout.Abs(100))
	if len(lines) == 0 {
		t.Error("Optimized linebreak produced no lines")
	}
}

func TestSaturatingSub(t *testing.T) {
	tests := []struct {
		a, b uint8
		want uint8
	}{
		{10, 3, 7},
		{3, 10, 0},
		{5, 5, 0},
		{0, 0, 0},
		{255, 0, 255},
	}

	for _, tt := range tests {
		got := saturatingSub(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("saturatingSub(%d, %d): got %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestEmptyLine(t *testing.T) {
	line := EmptyLine()

	if line.Width != 0 {
		t.Errorf("Empty line width should be 0, got %v", line.Width)
	}
	if line.Justify {
		t.Error("Empty line should not be justified")
	}
	if line.Dash != 0 {
		t.Error("Empty line should have no dash")
	}
	if len(line.Items) != 0 {
		t.Errorf("Empty line should have no items, got %d", len(line.Items))
	}
}

func TestIsVowel(t *testing.T) {
	vowels := []rune{'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U', 'ä', 'ö', 'ü'}
	consonants := []rune{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z'}

	for _, v := range vowels {
		if !isVowel(v) {
			t.Errorf("%c should be a vowel", v)
		}
	}

	for _, c := range consonants {
		if isVowel(c) {
			t.Errorf("%c should not be a vowel", c)
		}
	}
}
