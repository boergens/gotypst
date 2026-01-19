package syntax

import (
	"testing"
)

func TestSpanDetached(t *testing.T) {
	span := Detached()

	if !span.IsDetached() {
		t.Error("Detached span should report IsDetached() == true")
	}

	if span.Id() != nil {
		t.Errorf("Detached span should have nil id, got %v", span.Id())
	}

	if span.RawFileId() != 0 {
		t.Errorf("Detached span should have raw file id 0, got %d", span.RawFileId())
	}

	if _, _, ok := span.Range(); ok {
		t.Error("Detached span should not have a range")
	}
}

func TestSpanNumberEncoding(t *testing.T) {
	id := FileIdFromRaw(5)
	span, ok := SpanFromNumber(id, 10)

	if !ok {
		t.Fatal("SpanFromNumber should succeed for valid number")
	}

	if span.RawFileId() != 5 {
		t.Errorf("Expected raw file id 5, got %d", span.RawFileId())
	}

	gotId := span.Id()
	if gotId == nil {
		t.Fatal("Expected non-nil file id")
	}
	if gotId.IntoRaw() != 5 {
		t.Errorf("Expected file id 5, got %d", gotId.IntoRaw())
	}

	if span.Number() != 10 {
		t.Errorf("Expected number 10, got %d", span.Number())
	}

	if _, _, ok := span.Range(); ok {
		t.Error("Numbered span should not have a range")
	}
}

func TestSpanNumberInvalidRange(t *testing.T) {
	id := FileIdFromRaw(1)

	// Test number too low (< 2)
	_, ok := SpanFromNumber(id, 0)
	if ok {
		t.Error("SpanFromNumber should fail for number 0")
	}

	_, ok = SpanFromNumber(id, 1)
	if ok {
		t.Error("SpanFromNumber should fail for number 1")
	}

	// Test number at boundary (should succeed)
	_, ok = SpanFromNumber(id, 2)
	if !ok {
		t.Error("SpanFromNumber should succeed for number 2")
	}

	// Test number too high (>= 2^47)
	_, ok = SpanFromNumber(id, 1<<47)
	if ok {
		t.Error("SpanFromNumber should fail for number >= 2^47")
	}
}

func TestSpanRangeEncoding(t *testing.T) {
	id := FileIdFromRaw(65535) // u16::MAX

	testCases := []struct {
		start, end int
	}{
		{0, 0},
		{177, 233},
		{0, 8388607},       // 0 to max (2^23-1)
		{8388606, 8388607}, // near max
	}

	for _, tc := range testCases {
		span := SpanFromRange(id, tc.start, tc.end)

		if span.RawFileId() != 65535 {
			t.Errorf("Range span: expected raw file id 65535, got %d", span.RawFileId())
		}

		start, end, ok := span.Range()
		if !ok {
			t.Errorf("Range span %d..%d should have a range", tc.start, tc.end)
			continue
		}

		if start != tc.start || end != tc.end {
			t.Errorf("Expected range %d..%d, got %d..%d", tc.start, tc.end, start, end)
		}
	}
}

func TestSpanRangeSaturation(t *testing.T) {
	id := FileIdFromRaw(1)
	maxVal := (1 << 23) - 1 // 8388607 (max value that fits in 23 bits)

	// Test saturation of values exceeding max
	span := SpanFromRange(id, maxVal+1000, maxVal+2000)

	start, end, ok := span.Range()
	if !ok {
		t.Fatal("Range span should have a range")
	}

	if start != maxVal {
		t.Errorf("Start should be saturated to %d, got %d", maxVal, start)
	}

	if end != maxVal {
		t.Errorf("End should be saturated to %d, got %d", maxVal, end)
	}
}

func TestSpanOr(t *testing.T) {
	id := FileIdFromRaw(1)
	attached, _ := SpanFromNumber(id, 10)
	detached := Detached()

	// Detached.Or(attached) should return attached
	result := detached.Or(attached)
	if result.IsDetached() {
		t.Error("Detached.Or(attached) should return attached span")
	}

	// attached.Or(detached) should return attached
	result = attached.Or(detached)
	if result.IsDetached() {
		t.Error("attached.Or(detached) should return attached span")
	}
}

func TestFindSpan(t *testing.T) {
	id := FileIdFromRaw(1)
	attached, _ := SpanFromNumber(id, 10)
	detached := Detached()

	// Empty slice returns detached
	result := FindSpan([]Span{})
	if !result.IsDetached() {
		t.Error("FindSpan of empty slice should return detached")
	}

	// All detached returns detached
	result = FindSpan([]Span{detached, detached})
	if !result.IsDetached() {
		t.Error("FindSpan of all detached should return detached")
	}

	// Finds first non-detached
	result = FindSpan([]Span{detached, attached, detached})
	if result.IsDetached() {
		t.Error("FindSpan should find attached span")
	}
	if result.Number() != 10 {
		t.Errorf("Expected number 10, got %d", result.Number())
	}
}

func TestSpanned(t *testing.T) {
	id := FileIdFromRaw(1)
	span, _ := SpanFromNumber(id, 100)

	// Test NewSpanned
	s := NewSpanned("hello", span)
	if s.V != "hello" {
		t.Errorf("Expected value 'hello', got %q", s.V)
	}
	if s.Span != span {
		t.Error("Span mismatch")
	}

	// Test SpannedDetached
	d := SpannedDetached("world")
	if d.V != "world" {
		t.Errorf("Expected value 'world', got %q", d.V)
	}
	if !d.Span.IsDetached() {
		t.Error("SpannedDetached should have detached span")
	}

	// Test Map
	intSpan := NewSpanned(5, span)
	doubled := intSpan.Map(func(x int) int { return x * 2 })
	if doubled.V != 10 {
		t.Errorf("Expected mapped value 10, got %d", doubled.V)
	}
	if doubled.Span != span {
		t.Error("Map should preserve span")
	}
}

func TestSpanRawRoundtrip(t *testing.T) {
	id := FileIdFromRaw(123)
	original, _ := SpanFromNumber(id, 456)

	raw := original.Raw()
	restored := SpanFromRaw(raw)

	if restored.RawFileId() != original.RawFileId() {
		t.Error("Raw roundtrip should preserve file id")
	}

	if restored.Number() != original.Number() {
		t.Error("Raw roundtrip should preserve number")
	}
}

func TestSpanString(t *testing.T) {
	// Detached span
	d := Detached()
	if d.String() != "Span(detached)" {
		t.Errorf("Unexpected detached string: %s", d.String())
	}

	// Numbered span
	id := FileIdFromRaw(1)
	n, _ := SpanFromNumber(id, 42)
	expected := "Span(file=1, number=42)"
	if n.String() != expected {
		t.Errorf("Expected %q, got %q", expected, n.String())
	}

	// Range span
	r := SpanFromRange(id, 10, 20)
	expected = "Span(file=1, range=10..20)"
	if r.String() != expected {
		t.Errorf("Expected %q, got %q", expected, r.String())
	}
}
