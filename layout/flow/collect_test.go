package flow

import (
	"testing"

	"github.com/boergens/gotypst/eval"
)

func TestCollectEmpty(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, content, FlowModeBlock, styles, locator)
	if len(children) != 0 {
		t.Errorf("expected 0 children, got %d", len(children))
	}
}

func TestCollectNil(t *testing.T) {
	engine := &Engine{}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, nil, FlowModeBlock, styles, locator)
	if len(children) != 0 {
		t.Errorf("expected 0 children for nil content, got %d", len(children))
	}
}

func TestCollectText(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.TextElement{Text: "Hello"},
		},
	}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, content, FlowModeBlock, styles, locator)
	// Text elements don't produce flow children directly (they're inline)
	if len(children) != 0 {
		t.Errorf("expected 0 children for text, got %d", len(children))
	}
}

func TestCollectParagraph(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.ParagraphElement{
				Body: eval.Content{
					Elements: []eval.ContentElement{
						&eval.TextElement{Text: "Hello"},
					},
				},
			},
		},
	}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, content, FlowModeBlock, styles, locator)
	if len(children) != 1 {
		t.Fatalf("expected 1 child for paragraph, got %d", len(children))
	}
	if _, ok := children[0].(*MultiChild); !ok {
		t.Errorf("expected MultiChild for paragraph, got %T", children[0])
	}
}

func TestCollectHeading(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.HeadingElement{
				Depth: 1,
				Content: eval.Content{
					Elements: []eval.ContentElement{
						&eval.TextElement{Text: "Title"},
					},
				},
			},
		},
	}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, content, FlowModeBlock, styles, locator)
	if len(children) != 1 {
		t.Fatalf("expected 1 child for heading, got %d", len(children))
	}
	single, ok := children[0].(*SingleChild)
	if !ok {
		t.Fatalf("expected SingleChild for heading, got %T", children[0])
	}
	if !single.Sticky {
		t.Error("expected heading to be sticky")
	}
}

func TestCollectParbreak(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.ParbreakElement{},
		},
	}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, content, FlowModeBlock, styles, locator)
	if len(children) != 1 {
		t.Fatalf("expected 1 child for parbreak, got %d", len(children))
	}
	if _, ok := children[0].(RelChild); !ok {
		t.Errorf("expected RelChild for parbreak, got %T", children[0])
	}
}

func TestCollectWithStyles(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.ParagraphElement{
				Body: eval.Content{
					Elements: []eval.ContentElement{
						&eval.TextElement{Text: "Test"},
					},
				},
			},
		},
	}
	styles := StyleChain{
		Styles: map[string]interface{}{
			"block.align": "center",
		},
	}

	work := CollectWithStyles(engine, content, FlowModeBlock, styles)
	if work == nil {
		t.Fatal("expected non-nil work")
	}
	if work.Done() {
		t.Error("expected work to not be done initially")
	}
}

func TestLocatorNext(t *testing.T) {
	l := &Locator{Current: 0}

	loc1 := l.Next()
	if loc1 != 1 {
		t.Errorf("expected location 1, got %d", loc1)
	}

	loc2 := l.Next()
	if loc2 != 2 {
		t.Errorf("expected location 2, got %d", loc2)
	}
}

func TestStyleChainGet(t *testing.T) {
	styles := StyleChain{
		Styles: map[string]interface{}{
			"key1": "value1",
		},
	}

	val := styles.Get("key1")
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}

	val2 := styles.Get("nonexistent")
	if val2 != nil {
		t.Errorf("expected nil for nonexistent key, got %v", val2)
	}

	// Test nil styles
	emptyStyles := StyleChain{}
	val3 := emptyStyles.Get("key1")
	if val3 != nil {
		t.Errorf("expected nil for nil styles, got %v", val3)
	}
}

func TestCollectRawBlock(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.RawElement{
				Text:  "code",
				Block: true,
			},
		},
	}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, content, FlowModeBlock, styles, locator)
	if len(children) != 1 {
		t.Fatalf("expected 1 child for block raw, got %d", len(children))
	}
	if _, ok := children[0].(*SingleChild); !ok {
		t.Errorf("expected SingleChild for block raw, got %T", children[0])
	}
}

func TestCollectRawInline(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.RawElement{
				Text:  "code",
				Block: false,
			},
		},
	}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, content, FlowModeBlock, styles, locator)
	// Inline raw elements don't produce flow children
	if len(children) != 0 {
		t.Errorf("expected 0 children for inline raw, got %d", len(children))
	}
}

func TestCollectEquationBlock(t *testing.T) {
	engine := &Engine{}
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.EquationElement{
				Block: true,
			},
		},
	}
	styles := StyleChain{}
	locator := &Locator{}

	children := Collect(engine, content, FlowModeBlock, styles, locator)
	if len(children) != 1 {
		t.Fatalf("expected 1 child for block equation, got %d", len(children))
	}
	single, ok := children[0].(*SingleChild)
	if !ok {
		t.Fatalf("expected SingleChild for block equation, got %T", children[0])
	}
	// Display math should be centered
	if single.Align.X != FixedAlignCenter {
		t.Errorf("expected centered X alignment for display math, got %v", single.Align.X)
	}
}

func TestCollectStack(t *testing.T) {
	engine := &Engine{}

	// Vertical stack
	contentV := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.StackElement{
				Dir: eval.StackTTB,
			},
		},
	}
	styles := StyleChain{}
	locator := &Locator{}

	childrenV := Collect(engine, contentV, FlowModeBlock, styles, locator)
	if len(childrenV) != 1 {
		t.Fatalf("expected 1 child for vertical stack, got %d", len(childrenV))
	}
	if _, ok := childrenV[0].(*MultiChild); !ok {
		t.Errorf("expected MultiChild for vertical stack, got %T", childrenV[0])
	}

	// Horizontal stack
	contentH := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.StackElement{
				Dir: eval.StackLTR,
			},
		},
	}
	locator2 := &Locator{}

	childrenH := Collect(engine, contentH, FlowModeBlock, styles, locator2)
	if len(childrenH) != 1 {
		t.Fatalf("expected 1 child for horizontal stack, got %d", len(childrenH))
	}
	if _, ok := childrenH[0].(*SingleChild); !ok {
		t.Errorf("expected SingleChild for horizontal stack, got %T", childrenH[0])
	}
}
