package flow

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestUnbreakablePod(t *testing.T) {
	tests := []struct {
		name       string
		width      layout.Sizing
		height     layout.Sizing
		inset      layout.Sides[layout.Abs]
		base       layout.Size
		wantWidth  layout.Abs
		wantHeight layout.Abs
		wantExpandX bool
		wantExpandY bool
	}{
		{
			name:       "auto sizing uses full base",
			width:      layout.Auto,
			height:     layout.Auto,
			inset:      layout.Sides[layout.Abs]{},
			base:       layout.Size{Width: 100, Height: 200},
			wantWidth:  100,
			wantHeight: 200,
			wantExpandX: false, // Auto doesn't expand
			wantExpandY: false,
		},
		{
			name:       "rel sizing resolves against base",
			width:      layout.NewRelSizing(layout.RelRatio(0.5)),
			height:     layout.NewRelSizing(layout.RelAbs(50)),
			inset:      layout.Sides[layout.Abs]{},
			base:       layout.Size{Width: 100, Height: 200},
			wantWidth:  50,
			wantHeight: 50,
			wantExpandX: true, // Rel sizing expands
			wantExpandY: true,
		},
		{
			name:       "inset reduces available space",
			width:      layout.Auto,
			height:     layout.Auto,
			inset:      layout.Sides[layout.Abs]{Left: 10, Top: 20, Right: 10, Bottom: 20},
			base:       layout.Size{Width: 100, Height: 200},
			wantWidth:  80,  // 100 - 10 - 10
			wantHeight: 160, // 200 - 20 - 20
			wantExpandX: false,
			wantExpandY: false,
		},
		{
			name:       "fr sizing uses full base",
			width:      layout.NewFr(1),
			height:     layout.NewFr(2),
			inset:      layout.Sides[layout.Abs]{},
			base:       layout.Size{Width: 100, Height: 200},
			wantWidth:  100,
			wantHeight: 200,
			wantExpandX: true, // Fr expands
			wantExpandY: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := UnbreakablePod(tt.width, tt.height, tt.inset, StyleChain{}, tt.base)

			if !pod.Size.Width.ApproxEq(tt.wantWidth) {
				t.Errorf("width = %v, want %v", pod.Size.Width, tt.wantWidth)
			}
			if !pod.Size.Height.ApproxEq(tt.wantHeight) {
				t.Errorf("height = %v, want %v", pod.Size.Height, tt.wantHeight)
			}
			if pod.Expand.X != tt.wantExpandX {
				t.Errorf("expand.X = %v, want %v", pod.Expand.X, tt.wantExpandX)
			}
			if pod.Expand.Y != tt.wantExpandY {
				t.Errorf("expand.Y = %v, want %v", pod.Expand.Y, tt.wantExpandY)
			}
		})
	}
}

func TestDistribute(t *testing.T) {
	tests := []struct {
		name        string
		height      layout.Abs
		regionH     layout.Abs
		backlog     []layout.Abs
		wantFirst   layout.Abs
		wantBacklog []layout.Abs
	}{
		{
			name:        "zero height",
			height:      0,
			regionH:     100,
			backlog:     nil,
			wantFirst:   0,
			wantBacklog: []layout.Abs{},
		},
		{
			name:        "fits in first region",
			height:      50,
			regionH:     100,
			backlog:     nil,
			wantFirst:   50,
			wantBacklog: []layout.Abs{},
		},
		{
			name:        "exactly fills first region",
			height:      100,
			regionH:     100,
			backlog:     nil,
			wantFirst:   100,
			wantBacklog: []layout.Abs{},
		},
		{
			name:        "overflows to backlog",
			height:      150,
			regionH:     100,
			backlog:     []layout.Abs{100},
			wantFirst:   100,
			wantBacklog: []layout.Abs{50},
		},
		{
			name:        "fills multiple regions",
			height:      250,
			regionH:     100,
			backlog:     []layout.Abs{100, 100},
			wantFirst:   100,
			wantBacklog: []layout.Abs{100, 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regions := &layout.Regions{
				Size:    layout.Size{Width: 100, Height: tt.regionH},
				Full:    tt.regionH,
				Backlog: tt.backlog,
			}

			buf := make([]layout.Abs, 0, 4)
			first, backlog := Distribute(tt.height, regions, &buf)

			if !first.ApproxEq(tt.wantFirst) {
				t.Errorf("first = %v, want %v", first, tt.wantFirst)
			}

			if len(backlog) != len(tt.wantBacklog) {
				t.Errorf("backlog length = %d, want %d", len(backlog), len(tt.wantBacklog))
				return
			}

			for i, got := range backlog {
				if !got.ApproxEq(tt.wantBacklog[i]) {
					t.Errorf("backlog[%d] = %v, want %v", i, got, tt.wantBacklog[i])
				}
			}
		})
	}
}

func TestBreakablePod(t *testing.T) {
	tests := []struct {
		name       string
		width      layout.Sizing
		height     layout.Sizing
		regionH    layout.Abs
		wantWidth  layout.Abs
		wantHeight layout.Abs
	}{
		{
			name:       "auto height inherits region",
			width:      layout.Auto,
			height:     layout.Auto,
			regionH:    100,
			wantWidth:  200,
			wantHeight: 100,
		},
		{
			name:       "rel height distributes",
			width:      layout.Auto,
			height:     layout.NewRelSizing(layout.RelAbs(50)),
			regionH:    100,
			wantWidth:  200,
			wantHeight: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regions := &layout.Regions{
				Size: layout.Size{Width: 200, Height: tt.regionH},
				Full: tt.regionH,
			}
			inset := layout.Sides[layout.Abs]{}
			buf := make([]layout.Abs, 0, 2)

			pod := BreakablePod(tt.width, tt.height, inset, StyleChain{}, regions, &buf)

			if !pod.Size.Width.ApproxEq(tt.wantWidth) {
				t.Errorf("width = %v, want %v", pod.Size.Width, tt.wantWidth)
			}
			if !pod.Size.Height.ApproxEq(tt.wantHeight) {
				t.Errorf("height = %v, want %v", pod.Size.Height, tt.wantHeight)
			}
		})
	}
}

func TestLayoutSingleBlock(t *testing.T) {
	// Test basic block layout
	elem := &BlockElem{
		Width:  layout.Auto,
		Height: layout.Auto,
	}

	region := layout.Region{
		Size: layout.Size{Width: 100, Height: 200},
	}

	frame, err := LayoutSingleBlock(elem, &Engine{}, Locator{}, StyleChain{}, region)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if frame == nil {
		t.Fatal("expected frame, got nil")
	}

	// Frame should be Hard kind for explicit blocks
	if frame.Kind() != layout.FrameKindHard {
		t.Errorf("kind = %v, want FrameKindHard", frame.Kind())
	}
}

func TestLayoutMultiBlock(t *testing.T) {
	// Test multi-region block layout
	elem := &BlockElem{
		Width:  layout.Auto,
		Height: layout.Auto,
	}

	regions := &layout.Regions{
		Size:    layout.Size{Width: 100, Height: 200},
		Full:    200,
		Backlog: []layout.Abs{200},
	}

	fragment, err := LayoutMultiBlock(elem, &Engine{}, Locator{}, StyleChain{}, regions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fragment == nil {
		t.Fatal("expected fragment, got nil")
	}

	// Should have at least one frame
	if fragment.Len() == 0 {
		t.Error("expected at least one frame")
	}
}

func TestGrow(t *testing.T) {
	// Create a simple frame
	frame := layout.NewFrame(layout.Size{Width: 80, Height: 160})

	// Apply inset
	inset := layout.Sides[layout.Abs]{Left: 10, Top: 20, Right: 10, Bottom: 20}
	grown := grow(frame, inset)

	// Check new size
	if !grown.Width().ApproxEq(100) { // 80 + 10 + 10
		t.Errorf("grown width = %v, want 100", grown.Width())
	}
	if !grown.Height().ApproxEq(200) { // 160 + 20 + 20
		t.Errorf("grown height = %v, want 200", grown.Height())
	}
}

func TestResolveInset(t *testing.T) {
	inset := layout.Sides[layout.Rel]{
		Left:   layout.RelRatio(0.1),  // 10% of width
		Top:    layout.RelAbs(20),     // 20 absolute
		Right:  layout.RelRatio(0.1),  // 10% of width
		Bottom: layout.RelAbs(20),     // 20 absolute
	}
	base := layout.Size{Width: 100, Height: 200}

	resolved := resolveInset(inset, StyleChain{}, base)

	if !resolved.Left.ApproxEq(10) {
		t.Errorf("left = %v, want 10", resolved.Left)
	}
	if !resolved.Top.ApproxEq(20) {
		t.Errorf("top = %v, want 20", resolved.Top)
	}
	if !resolved.Right.ApproxEq(10) {
		t.Errorf("right = %v, want 10", resolved.Right)
	}
	if !resolved.Bottom.ApproxEq(20) {
		t.Errorf("bottom = %v, want 20", resolved.Bottom)
	}
}
