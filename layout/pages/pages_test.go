package pages

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestLayoutDocumentEmpty(t *testing.T) {
	engine := &Engine{}
	content := &Content{}
	styles := StyleChain{}

	doc, err := LayoutDocument(engine, content, styles)
	if err != nil {
		t.Fatalf("LayoutDocument failed: %v", err)
	}

	if doc == nil {
		t.Fatal("LayoutDocument returned nil")
	}

	// Empty content should produce at least one empty page
	if len(doc.Pages) == 0 {
		t.Error("Expected at least one page for empty content")
	}
}

func TestLayoutDocumentNilContent(t *testing.T) {
	engine := &Engine{}
	styles := StyleChain{}

	doc, err := LayoutDocument(engine, nil, styles)
	if err != nil {
		t.Fatalf("LayoutDocument failed with nil content: %v", err)
	}

	if doc == nil {
		t.Fatal("LayoutDocument returned nil")
	}
}

func TestCollectEmpty(t *testing.T) {
	locator := &Locator{Current: 0}
	splitLocator := locator.Split()
	styles := StyleChain{}

	items := Collect(nil, splitLocator, styles)

	// Empty children should produce a single empty run
	if len(items) != 1 {
		t.Fatalf("Expected 1 item for empty children, got %d", len(items))
	}

	if _, ok := items[0].(RunItem); !ok {
		t.Error("Expected RunItem for empty children")
	}
}

func TestCollectWithPagebreak(t *testing.T) {
	locator := &Locator{Current: 0}
	splitLocator := locator.Split()
	styles := StyleChain{}

	children := []Pair{
		{Element: &PagebreakElem{Weak: false}, Styles: styles},
	}

	items := Collect(children, splitLocator, styles)

	// Strong pagebreak should produce an empty run before it
	if len(items) < 1 {
		t.Fatal("Expected at least 1 item")
	}
}

func TestCollectWithWeakPagebreak(t *testing.T) {
	locator := &Locator{Current: 0}
	splitLocator := locator.Split()
	styles := StyleChain{}

	children := []Pair{
		{Element: &PagebreakElem{Weak: true}, Styles: styles},
	}

	items := Collect(children, splitLocator, styles)

	// Weak pagebreak alone should still produce items
	if len(items) == 0 {
		t.Error("Expected items for weak pagebreak")
	}
}

func TestCollectWithParityPagebreak(t *testing.T) {
	locator := &Locator{Current: 0}
	splitLocator := locator.Split()
	styles := StyleChain{}

	evenParity := ParityEven
	children := []Pair{
		{Element: &PagebreakElem{Weak: true, To: &evenParity}, Styles: styles},
	}

	items := Collect(children, splitLocator, styles)

	// Should have a parity item
	hasParity := false
	for _, item := range items {
		if _, ok := item.(ParityItem); ok {
			hasParity = true
			break
		}
	}

	if !hasParity {
		t.Error("Expected ParityItem for pagebreak with To field")
	}
}

func TestCollectWithTags(t *testing.T) {
	locator := &Locator{Current: 0}
	splitLocator := locator.Split()
	styles := StyleChain{}

	// Tags followed by a non-boundary pagebreak should produce a TagsItem
	children := []Pair{
		{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 1}}, Styles: styles},
		{Element: &TagElem{Tag: Tag{Kind: TagEnd, Location: 1}}, Styles: styles},
		{Element: &PagebreakElem{Weak: false, Boundary: false}, Styles: styles},
	}

	items := Collect(children, splitLocator, styles)

	// Tags before a non-boundary pagebreak should produce a TagsItem
	hasTags := false
	for _, item := range items {
		if _, ok := item.(TagsItem); ok {
			hasTags = true
			break
		}
	}

	if !hasTags {
		t.Error("Expected TagsItem for tags before pagebreak")
	}
}

func TestCollectTagsAloneAtEnd(t *testing.T) {
	locator := &Locator{Current: 0}
	splitLocator := locator.Split()
	styles := StyleChain{}

	// Tags alone at end with staged_empty_page=true get included in the Run
	// because of edge case handling in collect
	children := []Pair{
		{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 1}}, Styles: styles},
	}

	items := Collect(children, splitLocator, styles)

	// Should produce items (either Run or Tags depending on edge case)
	if len(items) == 0 {
		t.Error("Expected at least one item")
	}
}

func TestParityMatches(t *testing.T) {
	tests := []struct {
		parity    Parity
		pageCount int
		expected  bool
	}{
		{ParityEven, 0, false}, // 0 pages is even, don't need to add
		{ParityEven, 1, true},  // 1 page is odd, need to add
		{ParityEven, 2, false}, // 2 pages is even, don't need to add
		{ParityOdd, 0, true},   // 0 pages is even, need to add
		{ParityOdd, 1, false},  // 1 page is odd, don't need to add
		{ParityOdd, 2, true},   // 2 pages is even, need to add
	}

	for _, tt := range tests {
		result := tt.parity.Matches(tt.pageCount)
		if result != tt.expected {
			t.Errorf("Parity(%v).Matches(%d) = %v, want %v",
				tt.parity, tt.pageCount, result, tt.expected)
		}
	}
}

func TestBindingSwap(t *testing.T) {
	tests := []struct {
		binding  Binding
		pageNum  int
		expected bool
	}{
		{BindingLeft, 0, false},  // First page (0-indexed), no swap
		{BindingLeft, 1, true},   // Second page, swap
		{BindingLeft, 2, false},  // Third page, no swap
		{BindingRight, 0, true},  // First page, swap
		{BindingRight, 1, false}, // Second page, no swap
		{BindingRight, 2, true},  // Third page, swap
	}

	for _, tt := range tests {
		result := tt.binding.Swap(tt.pageNum)
		if result != tt.expected {
			t.Errorf("Binding(%v).Swap(%d) = %v, want %v",
				tt.binding, tt.pageNum, result, tt.expected)
		}
	}
}

func TestManualPageCounter(t *testing.T) {
	counter := NewManualPageCounter()

	if counter.Physical() != 0 {
		t.Errorf("Initial physical = %d, want 0", counter.Physical())
	}
	if counter.Logical() != 1 {
		t.Errorf("Initial logical = %d, want 1", counter.Logical())
	}

	counter.Step()

	if counter.Physical() != 1 {
		t.Errorf("After step physical = %d, want 1", counter.Physical())
	}
	if counter.Logical() != 2 {
		t.Errorf("After step logical = %d, want 2", counter.Logical())
	}
}

func TestCounterUpdateSet(t *testing.T) {
	update := CounterUpdateSet{Value: 42}
	result := update.Apply(1)
	if result != 42 {
		t.Errorf("CounterUpdateSet.Apply(1) = %d, want 42", result)
	}

	// Verify it ignores current value
	result = update.Apply(100)
	if result != 42 {
		t.Errorf("CounterUpdateSet.Apply(100) = %d, want 42", result)
	}
}

func TestCounterUpdateStep(t *testing.T) {
	update := CounterUpdateStep{}
	result := update.Apply(1)
	if result != 2 {
		t.Errorf("CounterUpdateStep.Apply(1) = %d, want 2", result)
	}

	result = update.Apply(99)
	if result != 100 {
		t.Errorf("CounterUpdateStep.Apply(99) = %d, want 100", result)
	}
}

func TestManualPageCounterVisitWithPageCounterUpdate(t *testing.T) {
	counter := NewManualPageCounter()

	// Create a frame with a page counter update tag
	frame := Hard(layout.Size{Width: 100, Height: 100})
	frame.Push(layout.Point{X: 0, Y: 0}, TagItem{
		Tag: Tag{
			Kind:     TagStart,
			Location: 1,
			Elem: &CounterUpdateElem{
				Key:    CounterKeyPage,
				Update: CounterUpdateSet{Value: 10},
			},
		},
	})

	// Initial logical should be 1
	if counter.Logical() != 1 {
		t.Errorf("Initial logical = %d, want 1", counter.Logical())
	}

	// Visit the frame to process counter updates
	err := counter.Visit(&frame)
	if err != nil {
		t.Fatalf("Visit failed: %v", err)
	}

	// Logical should now be 10
	if counter.Logical() != 10 {
		t.Errorf("After visit logical = %d, want 10", counter.Logical())
	}

	// Physical should be unchanged
	if counter.Physical() != 0 {
		t.Errorf("Physical should be unchanged, got %d, want 0", counter.Physical())
	}
}

func TestManualPageCounterVisitWithNestedFrame(t *testing.T) {
	counter := NewManualPageCounter()

	// Create a nested frame structure
	innerFrame := Hard(layout.Size{Width: 50, Height: 50})
	innerFrame.Push(layout.Point{X: 0, Y: 0}, TagItem{
		Tag: Tag{
			Kind:     TagStart,
			Location: 1,
			Elem: &CounterUpdateElem{
				Key:    CounterKeyPage,
				Update: CounterUpdateSet{Value: 5},
			},
		},
	})

	outerFrame := Hard(layout.Size{Width: 100, Height: 100})
	outerFrame.PushFrame(layout.Point{X: 25, Y: 25}, innerFrame)

	err := counter.Visit(&outerFrame)
	if err != nil {
		t.Fatalf("Visit failed: %v", err)
	}

	// Logical should be updated from nested frame
	if counter.Logical() != 5 {
		t.Errorf("After visit logical = %d, want 5", counter.Logical())
	}
}

func TestManualPageCounterVisitIgnoresNonPageCounters(t *testing.T) {
	counter := NewManualPageCounter()

	// Create a frame with a non-page counter update
	frame := Hard(layout.Size{Width: 100, Height: 100})
	frame.Push(layout.Point{X: 0, Y: 0}, TagItem{
		Tag: Tag{
			Kind:     TagStart,
			Location: 1,
			Elem: &CounterUpdateElem{
				Key:    CounterKeyFigure, // Not a page counter
				Update: CounterUpdateSet{Value: 99},
			},
		},
	})

	err := counter.Visit(&frame)
	if err != nil {
		t.Fatalf("Visit failed: %v", err)
	}

	// Logical should remain unchanged (figure counter was ignored)
	if counter.Logical() != 1 {
		t.Errorf("Logical should be unchanged, got %d, want 1", counter.Logical())
	}
}

func TestManualPageCounterVisitIgnoresEndTags(t *testing.T) {
	counter := NewManualPageCounter()

	// Create a frame with an end tag containing a counter update
	frame := Hard(layout.Size{Width: 100, Height: 100})
	frame.Push(layout.Point{X: 0, Y: 0}, TagItem{
		Tag: Tag{
			Kind:     TagEnd, // End tag should be ignored
			Location: 1,
			Elem: &CounterUpdateElem{
				Key:    CounterKeyPage,
				Update: CounterUpdateSet{Value: 99},
			},
		},
	})

	err := counter.Visit(&frame)
	if err != nil {
		t.Fatalf("Visit failed: %v", err)
	}

	// Logical should remain unchanged (end tags are ignored)
	if counter.Logical() != 1 {
		t.Errorf("Logical should be unchanged, got %d, want 1", counter.Logical())
	}
}

func TestManualPageCounterMultipleUpdates(t *testing.T) {
	counter := NewManualPageCounter()

	// Create a frame with multiple counter updates
	frame := Hard(layout.Size{Width: 100, Height: 100})

	// First update: set to 5
	frame.Push(layout.Point{X: 0, Y: 0}, TagItem{
		Tag: Tag{
			Kind:     TagStart,
			Location: 1,
			Elem: &CounterUpdateElem{
				Key:    CounterKeyPage,
				Update: CounterUpdateSet{Value: 5},
			},
		},
	})

	// Second update: step (should become 6)
	frame.Push(layout.Point{X: 0, Y: 10}, TagItem{
		Tag: Tag{
			Kind:     TagStart,
			Location: 2,
			Elem: &CounterUpdateElem{
				Key:    CounterKeyPage,
				Update: CounterUpdateStep{},
			},
		},
	})

	// Third update: set to 100
	frame.Push(layout.Point{X: 0, Y: 20}, TagItem{
		Tag: Tag{
			Kind:     TagStart,
			Location: 3,
			Elem: &CounterUpdateElem{
				Key:    CounterKeyPage,
				Update: CounterUpdateSet{Value: 100},
			},
		},
	})

	err := counter.Visit(&frame)
	if err != nil {
		t.Fatalf("Visit failed: %v", err)
	}

	// Final logical should be 100
	if counter.Logical() != 100 {
		t.Errorf("After multiple updates logical = %d, want 100", counter.Logical())
	}
}

func TestFrameOperations(t *testing.T) {
	frame := Hard(layout.Size{Width: 100, Height: 200})

	if frame.Width() != 100 {
		t.Errorf("Width = %v, want 100", frame.Width())
	}
	if frame.Height() != 200 {
		t.Errorf("Height = %v, want 200", frame.Height())
	}

	// Test Push
	frame.Push(layout.Point{X: 10, Y: 20}, TagItem{Tag: Tag{Kind: TagStart, Location: 1}})
	if len(frame.Items) != 1 {
		t.Errorf("Items count = %d, want 1", len(frame.Items))
	}

	// Test PushFrame
	inner := Hard(layout.Size{Width: 50, Height: 50})
	frame.PushFrame(layout.Point{X: 25, Y: 75}, inner)
	if len(frame.Items) != 2 {
		t.Errorf("Items count = %d, want 2", len(frame.Items))
	}
}

func TestLayoutPageRun(t *testing.T) {
	engine := &Engine{}
	locator := Locator{Current: 0}
	styles := StyleChain{}

	pages, err := LayoutPageRun(engine, nil, locator, styles)
	if err != nil {
		t.Fatalf("LayoutPageRun failed: %v", err)
	}

	if len(pages) == 0 {
		t.Error("Expected at least one page")
	}
}

func TestLayoutBlankPage(t *testing.T) {
	engine := &Engine{}
	locator := Locator{Current: 0}
	styles := StyleChain{}

	page, err := LayoutBlankPage(engine, locator, styles)
	if err != nil {
		t.Fatalf("LayoutBlankPage failed: %v", err)
	}

	if page == nil {
		t.Error("Expected non-nil page")
	}
}

func TestFinalize(t *testing.T) {
	engine := &Engine{}
	counter := NewManualPageCounter()
	tags := []Tag{}

	layouted := LayoutedPage{
		Inner:    Hard(layout.Size{Width: 500, Height: 700}),
		Margin:   Sides[layout.Abs]{Left: 50, Top: 50, Right: 50, Bottom: 50},
		Binding:  BindingLeft,
		TwoSided: false,
	}

	page, err := Finalize(engine, counter, &tags, layouted)
	if err != nil {
		t.Fatalf("Finalize failed: %v", err)
	}

	if page == nil {
		t.Fatal("Expected non-nil page")
	}

	// Check page dimensions (content + margins)
	expectedWidth := layout.Abs(600)  // 500 + 50 + 50
	expectedHeight := layout.Abs(800) // 700 + 50 + 50

	if page.Frame.Width() != expectedWidth {
		t.Errorf("Page width = %v, want %v", page.Frame.Width(), expectedWidth)
	}
	if page.Frame.Height() != expectedHeight {
		t.Errorf("Page height = %v, want %v", page.Frame.Height(), expectedHeight)
	}

	if page.Number != 1 {
		t.Errorf("Page number = %d, want 1", page.Number)
	}
}

func TestFinalizeTwoSided(t *testing.T) {
	engine := &Engine{}
	tags := []Tag{}

	// Test left-bound, second page (should swap margins)
	counter := NewManualPageCounter()
	counter.Step() // Move to page 2

	layouted := LayoutedPage{
		Inner:    Hard(layout.Size{Width: 500, Height: 700}),
		Margin:   Sides[layout.Abs]{Left: 70, Top: 50, Right: 30, Bottom: 50},
		Binding:  BindingLeft,
		TwoSided: true,
	}

	page, err := Finalize(engine, counter, &tags, layouted)
	if err != nil {
		t.Fatalf("Finalize failed: %v", err)
	}

	// Margins should be swapped for second page with left binding
	// The inner frame should be positioned with the swapped margins
	if len(page.Frame.Items) == 0 {
		t.Error("Expected items in frame")
	}
}

func TestSidesSum(t *testing.T) {
	sides := Sides[layout.Abs]{Left: 10, Top: 20, Right: 30, Bottom: 40}
	sum := sides.SumByAxis()

	if sum.Width != 40 { // 10 + 30
		t.Errorf("Sum width = %v, want 40", sum.Width)
	}
	if sum.Height != 60 { // 20 + 40
		t.Errorf("Sum height = %v, want 60", sum.Height)
	}
}

func TestLocatorSplit(t *testing.T) {
	locator := &Locator{Current: 100}
	split := locator.Split()

	loc1 := split.Next(nil)
	if loc1.Current != 101 {
		t.Errorf("First next = %d, want 101", loc1.Current)
	}

	loc2 := split.Next(nil)
	if loc2.Current != 102 {
		t.Errorf("Second next = %d, want 102", loc2.Current)
	}

	relayout := split.Relayout()
	if relayout.Current != 102 {
		t.Errorf("Relayout = %d, want 102", relayout.Current)
	}
}

// TestMigrateUnterminatedTags tests the tag migration algorithm
func TestMigrateUnterminatedTags(t *testing.T) {
	styles := StyleChain{}

	// Test case 1: Mixed terminated and unterminated tags
	t.Run("MixedTerminatedAndUnterminated", func(t *testing.T) {
		// Tag1 (location 1) is unterminated - should migrate
		// Tag2 (location 2) is terminated (has both start and end) - should stay
		children := []Pair{
			{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 1}}, Styles: styles}, // 0: unterminated start
			{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 2}}, Styles: styles}, // 1: terminated start
			{Element: &TagElem{Tag: Tag{Kind: TagEnd, Location: 2}}, Styles: styles},   // 2: terminated end
			{Element: &PagebreakElem{Weak: false}, Styles: styles},                      // 3: pagebreak
		}

		// The algorithm should reorder to: [term-start, term-end, pagebreak, unterm-start]
		newEnd := migrateUnterminatedTags(children, 0, 3)

		// Should return position after the tags that stay (before pagebreak)
		if newEnd != 2 {
			t.Errorf("Expected newEnd=2, got %d", newEnd)
		}

		// Verify reordering: first two should be the terminated pair
		if te, ok := children[0].Element.(*TagElem); !ok || te.Tag.Location != 2 || te.Tag.Kind != TagStart {
			t.Errorf("Expected terminated start tag at position 0, got %v", children[0].Element)
		}
		if te, ok := children[1].Element.(*TagElem); !ok || te.Tag.Location != 2 || te.Tag.Kind != TagEnd {
			t.Errorf("Expected terminated end tag at position 1, got %v", children[1].Element)
		}

		// Position 2 should be the pagebreak
		if _, ok := children[2].Element.(*PagebreakElem); !ok {
			t.Errorf("Expected pagebreak at position 2, got %v", children[2].Element)
		}

		// Position 3 should be the unterminated start tag
		if te, ok := children[3].Element.(*TagElem); !ok || te.Tag.Location != 1 || te.Tag.Kind != TagStart {
			t.Errorf("Expected unterminated start tag at position 3, got %v", children[3].Element)
		}
	})

	// Test case 2: All tags terminated - nothing should migrate
	t.Run("AllTerminated", func(t *testing.T) {
		children := []Pair{
			{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 1}}, Styles: styles},
			{Element: &TagElem{Tag: Tag{Kind: TagEnd, Location: 1}}, Styles: styles},
			{Element: &PagebreakElem{Weak: false}, Styles: styles},
		}

		newEnd := migrateUnterminatedTags(children, 0, 2)

		// All tags stay, so newEnd should be at the original position
		if newEnd != 2 {
			t.Errorf("Expected newEnd=2, got %d", newEnd)
		}
	})

	// Test case 3: All tags unterminated - all should migrate
	t.Run("AllUnterminated", func(t *testing.T) {
		children := []Pair{
			{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 1}}, Styles: styles},
			{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 2}}, Styles: styles},
			{Element: &PagebreakElem{Weak: false}, Styles: styles},
		}

		newEnd := migrateUnterminatedTags(children, 0, 2)

		// No tags stay, so newEnd should be 0 (tagStart)
		if newEnd != 0 {
			t.Errorf("Expected newEnd=0, got %d", newEnd)
		}

		// Pagebreak should now be at position 0
		if _, ok := children[0].Element.(*PagebreakElem); !ok {
			t.Errorf("Expected pagebreak at position 0, got %v", children[0].Element)
		}
	})

	// Test case 4: No trailing tags - nothing to do
	t.Run("NoTrailingTags", func(t *testing.T) {
		children := []Pair{
			{Element: &PagebreakElem{Weak: false}, Styles: styles},
		}

		newEnd := migrateUnterminatedTags(children, 0, 0)

		// Nothing changes
		if newEnd != 0 {
			t.Errorf("Expected newEnd=0, got %d", newEnd)
		}
	})

	// Test case 5: Content followed by tags
	t.Run("ContentFollowedByTags", func(t *testing.T) {
		// Simulate: content, then tags, then pagebreak
		// Using a different ContentElement for "content" - we need a non-tag, non-pagebreak
		textElem := &textContentElem{}
		children := []Pair{
			{Element: textElem, Styles: styles},                                         // 0: content
			{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 1}}, Styles: styles}, // 1: unterminated
			{Element: &PagebreakElem{Weak: false}, Styles: styles},                      // 2: pagebreak
		}

		newEnd := migrateUnterminatedTags(children, 0, 2)

		// Tag at position 1 should migrate, content stays at 0
		// newEnd should be 1 (after content, before pagebreak which moves)
		if newEnd != 1 {
			t.Errorf("Expected newEnd=1, got %d", newEnd)
		}

		// Content stays at 0
		if _, ok := children[0].Element.(*textContentElem); !ok {
			t.Errorf("Expected text content at position 0")
		}

		// Pagebreak moves to position 1
		if _, ok := children[1].Element.(*PagebreakElem); !ok {
			t.Errorf("Expected pagebreak at position 1, got %v", children[1].Element)
		}

		// Unterminated tag moves to position 2
		if te, ok := children[2].Element.(*TagElem); !ok || te.Tag.Location != 1 {
			t.Errorf("Expected unterminated tag at position 2, got %v", children[2].Element)
		}
	})

	// Test case 6: Multiple pagebreaks
	t.Run("MultiplePagebreaks", func(t *testing.T) {
		children := []Pair{
			{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 1}}, Styles: styles}, // unterminated
			{Element: &PagebreakElem{Weak: false}, Styles: styles},
			{Element: &PagebreakElem{Weak: true}, Styles: styles},
		}

		newEnd := migrateUnterminatedTags(children, 0, 1)

		// Tag should migrate after both pagebreaks
		if newEnd != 0 {
			t.Errorf("Expected newEnd=0, got %d", newEnd)
		}

		// First two positions should be pagebreaks
		if _, ok := children[0].Element.(*PagebreakElem); !ok {
			t.Errorf("Expected pagebreak at position 0")
		}
		if _, ok := children[1].Element.(*PagebreakElem); !ok {
			t.Errorf("Expected pagebreak at position 1")
		}

		// Tag should be at position 2
		if _, ok := children[2].Element.(*TagElem); !ok {
			t.Errorf("Expected tag at position 2")
		}
	})
}

// textContentElem is a simple content element for testing
type textContentElem struct{}

func (*textContentElem) isContentElement() {}

// TestCollectWithMixedTags tests collection with complex tag scenarios
func TestCollectWithMixedTags(t *testing.T) {
	locator := &Locator{Current: 0}
	splitLocator := locator.Split()
	styles := StyleChain{}

	// Test: unterminated tag followed by pagebreak should produce
	// a Run for the tag (after migration) in the next run
	children := []Pair{
		{Element: &TagElem{Tag: Tag{Kind: TagStart, Location: 1}}, Styles: styles},
		{Element: &PagebreakElem{Weak: false}, Styles: styles},
		{Element: &TagElem{Tag: Tag{Kind: TagEnd, Location: 1}}, Styles: styles},
	}

	items := Collect(children, splitLocator, styles)

	// Should have items (exact count depends on implementation details)
	if len(items) == 0 {
		t.Error("Expected at least one item")
	}

	// Verify we have at least one RunItem
	hasRun := false
	for _, item := range items {
		if _, ok := item.(RunItem); ok {
			hasRun = true
			break
		}
	}
	if !hasRun {
		t.Error("Expected at least one RunItem")
	}
}
