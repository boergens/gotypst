package layout

import (
	"math"
	"testing"
)

func TestAbsConstants(t *testing.T) {
	// Verify unit conversions
	if Mm*25.4 < In-0.001 || Mm*25.4 > In+0.001 {
		t.Errorf("25.4mm should equal 1in: got %v", Mm*25.4)
	}

	if Cm*2.54 < In-0.001 || Cm*2.54 > In+0.001 {
		t.Errorf("2.54cm should equal 1in: got %v", Cm*2.54)
	}

	if In != 72*Pt {
		t.Errorf("1in should equal 72pt: got %v", In)
	}
}

func TestAbsMethods(t *testing.T) {
	a := Abs(10)
	b := Abs(20)

	if a.Min(b) != 10 {
		t.Errorf("Min(10, 20) = %v, expected 10", a.Min(b))
	}

	if a.Max(b) != 20 {
		t.Errorf("Max(10, 20) = %v, expected 20", a.Max(b))
	}

	neg := Abs(-5)
	if neg.Abs() != 5 {
		t.Errorf("Abs(-5) = %v, expected 5", neg.Abs())
	}

	if a.Clamp(15, 25) != 15 {
		t.Errorf("Clamp(10, 15, 25) = %v, expected 15", a.Clamp(15, 25))
	}
}

func TestPoint(t *testing.T) {
	p := Point{X: 10, Y: 20}
	q := Point{X: 5, Y: 15}

	sum := p.Add(q)
	if sum.X != 15 || sum.Y != 35 {
		t.Errorf("Add: expected (15, 35), got (%v, %v)", sum.X, sum.Y)
	}

	diff := p.Sub(q)
	if diff.X != 5 || diff.Y != 5 {
		t.Errorf("Sub: expected (5, 5), got (%v, %v)", diff.X, diff.Y)
	}

	scaled := p.Scale(2)
	if scaled.X != 20 || scaled.Y != 40 {
		t.Errorf("Scale: expected (20, 40), got (%v, %v)", scaled.X, scaled.Y)
	}
}

func TestSize(t *testing.T) {
	s := Size{Width: 100, Height: 200}

	if s.AspectRatio() != 0.5 {
		t.Errorf("AspectRatio: expected 0.5, got %v", s.AspectRatio())
	}

	if !s.Contains(Point{X: 50, Y: 100}) {
		t.Error("Size should contain point inside")
	}

	if s.Contains(Point{X: 150, Y: 100}) {
		t.Error("Size should not contain point outside")
	}

	zero := Size{}
	if !zero.IsZero() {
		t.Error("Zero size should be zero")
	}
}

func TestRatioResolve(t *testing.T) {
	r := Ratio(0.5) // 50%
	whole := Abs(200)

	result := r.Resolve(whole)
	if result != 100 {
		t.Errorf("50%% of 200 = %v, expected 100", result)
	}
}

func TestRelative(t *testing.T) {
	// 10pt + 25%
	rel := Relative{Abs: 10, Rel: 0.25}
	whole := Abs(100)

	result := rel.Resolve(whole)
	if result != 35 {
		t.Errorf("10pt + 25%% of 100 = %v, expected 35", result)
	}
}

func TestTransformIdentity(t *testing.T) {
	id := Identity()

	if !id.IsIdentity() {
		t.Error("Identity should be identity")
	}

	p := Point{X: 10, Y: 20}
	result := id.Apply(p)
	if result.X != 10 || result.Y != 20 {
		t.Errorf("Identity transform should not change point: got (%v, %v)", result.X, result.Y)
	}
}

func TestTransformTranslate(t *testing.T) {
	tr := Translate(5, 10)

	p := Point{X: 10, Y: 20}
	result := tr.Apply(p)
	if result.X != 15 || result.Y != 30 {
		t.Errorf("Translate(5, 10) of (10, 20) = (%v, %v), expected (15, 30)", result.X, result.Y)
	}
}

func TestTransformScale(t *testing.T) {
	sc := Scale(2, 3)

	p := Point{X: 10, Y: 20}
	result := sc.Apply(p)
	if result.X != 20 || result.Y != 60 {
		t.Errorf("Scale(2, 3) of (10, 20) = (%v, %v), expected (20, 60)", result.X, result.Y)
	}
}

func TestTransformRotate(t *testing.T) {
	rot := Rotate(math.Pi / 2) // 90 degrees

	p := Point{X: 10, Y: 0}
	result := rot.Apply(p)

	// After 90 degree rotation, (10, 0) should become approximately (0, 10)
	if result.X > 0.001 || result.X < -0.001 {
		t.Errorf("X after 90 degree rotation should be ~0, got %v", result.X)
	}
	if result.Y < 9.999 || result.Y > 10.001 {
		t.Errorf("Y after 90 degree rotation should be ~10, got %v", result.Y)
	}
}

func TestTransformComposition(t *testing.T) {
	tr := Translate(10, 20)
	sc := Scale(2, 2)

	// Scale then translate
	combined := sc.Then(tr)

	p := Point{X: 5, Y: 5}
	result := combined.Apply(p)

	// First scale: (5, 5) -> (10, 10)
	// Then translate: (10, 10) -> (20, 30)
	if result.X != 20 || result.Y != 30 {
		t.Errorf("Combined transform: expected (20, 30), got (%v, %v)", result.X, result.Y)
	}
}

func TestSides(t *testing.T) {
	sides := SidesSplat(Abs(10))

	if sides.Left != 10 || sides.Right != 10 || sides.Top != 10 || sides.Bottom != 10 {
		t.Error("SidesSplat should set all sides equal")
	}
}

func TestCorners(t *testing.T) {
	corners := CornersSplat(Abs(5))

	if corners.TopLeft != 5 || corners.TopRight != 5 ||
		corners.BottomLeft != 5 || corners.BottomRight != 5 {
		t.Error("CornersSplat should set all corners equal")
	}
}

func TestDir(t *testing.T) {
	if !DirLTR.IsHorizontal() {
		t.Error("LTR should be horizontal")
	}
	if !DirRTL.IsHorizontal() {
		t.Error("RTL should be horizontal")
	}
	if DirTTB.IsHorizontal() {
		t.Error("TTB should not be horizontal")
	}
	if DirBTT.IsHorizontal() {
		t.Error("BTT should not be horizontal")
	}

	if !DirLTR.IsPositive() {
		t.Error("LTR should be positive")
	}
	if DirRTL.IsPositive() {
		t.Error("RTL should not be positive")
	}
	if !DirTTB.IsPositive() {
		t.Error("TTB should be positive")
	}
	if DirBTT.IsPositive() {
		t.Error("BTT should not be positive")
	}
}

func TestFrame(t *testing.T) {
	frame := NewFrame(Size{Width: 100, Height: 200})

	if frame.Size.Width != 100 || frame.Size.Height != 200 {
		t.Error("Frame size not set correctly")
	}

	if len(frame.Items) != 0 {
		t.Error("New frame should have no items")
	}
}

func TestFramePush(t *testing.T) {
	frame := NewFrame(Size{Width: 100, Height: 100})

	subframe := NewFrame(Size{Width: 50, Height: 50})
	frame.PushFrame(Point{X: 10, Y: 20}, subframe)

	if len(frame.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(frame.Items))
	}
}

func TestFrameTranslate(t *testing.T) {
	frame := NewFrame(Size{Width: 100, Height: 100})

	subframe := NewFrame(Size{Width: 50, Height: 50})
	frame.PushFrame(Point{X: 10, Y: 20}, subframe)

	frame.Translate(Point{X: 5, Y: 10})

	pos, ok := frame.Items[0].(*PositionedItem)
	if !ok {
		t.Fatal("Expected PositionedItem")
	}

	if pos.Pos.X != 15 || pos.Pos.Y != 30 {
		t.Errorf("After translate: expected (15, 30), got (%v, %v)", pos.Pos.X, pos.Pos.Y)
	}
}

func TestFragment(t *testing.T) {
	frag := Fragment{
		NewFrame(Size{Width: 100, Height: 200}),
		NewFrame(Size{Width: 100, Height: 200}),
	}

	str := frag.String()
	if str != "Fragment(2 frames)" {
		t.Errorf("Fragment.String() = %q, expected \"Fragment(2 frames)\"", str)
	}
}

func TestRegion(t *testing.T) {
	regions := Regions{
		Current: Region{
			Size: Size{Width: 595, Height: 842}, // A4 size
			Full: true,
		},
	}

	if regions.IsEmpty() {
		t.Error("Regions should not be empty")
	}

	first := regions.First()
	if first.Size.Width != 595 {
		t.Errorf("Expected A4 width, got %v", first.Size.Width)
	}
}
