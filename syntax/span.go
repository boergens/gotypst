// Package syntax provides source location tracking types for Typst.
//
// This package is a Go translation of typst-syntax/src/span.rs from the
// original Typst compiler.
package syntax

import (
	"fmt"
)

// FileId identifies a source file. The zero value represents no file (detached).
//
// In the original Typst, FileId is backed by NonZeroU16 and has interning
// capabilities. This Go version uses a simple uint16 wrapper where 0 represents
// the absence of a file. Use a FileRegistry to map FileIds to paths.
type FileId uint16

// NoFile represents a detached/invalid file ID.
const NoFile FileId = 0

// FileIdFromRaw creates a FileId from a raw uint16 value.
// Value 0 creates NoFile (detached).
func FileIdFromRaw(v uint16) FileId {
	return FileId(v)
}

// Raw returns the underlying uint16 value.
func (id FileId) Raw() uint16 {
	return uint16(id)
}

// IsValid returns true if this FileId points to a valid file.
func (id FileId) IsValid() bool {
	return id != NoFile
}

// Span defines a range in a source file.
//
// This is used throughout the compiler to track which source section an
// element stems from or an error applies to.
//
// Spans come in two flavors: Numbered spans and raw range spans.
//
// # Numbered spans
// Typst source files use numbered spans. Rather than using byte ranges,
// which shift a lot as you type, each AST node gets a unique number.
//
// During editing, the span numbers stay mostly stable, even for nodes behind
// an insertion. This is not true for simple ranges as they would shift. Spans
// can be used as inputs to memoized functions without hurting cache
// performance when text is inserted somewhere in the document other than the
// end.
//
// # Raw range spans
// Non-Typst files use raw ranges instead of numbered spans. The maximum
// encodable value for start and end is 2^23. Larger values will be saturated.
//
// Data layout: | 16 bits file id | 48 bits number |
//
// Number encoding:
//   - 0 means detached (no file, no span)
//   - 2..2^47-1 is a numbered span
//   - 2^47..2^48-1 is a raw range span
type Span struct {
	bits uint64
}

// Span constants for bit manipulation.
const (
	// spanFull is the valid range for numbered spans (2..2^47).
	spanFullStart uint64 = 2
	spanFullEnd   uint64 = 1 << 47

	// spanDetached is the value reserved for detached spans.
	spanDetached uint64 = 0

	// Bit layout constants.
	spanNumberBits     = 48
	spanFileIdShift    = spanNumberBits
	spanNumberMask     = (uint64(1) << spanNumberBits) - 1
	spanRangeBase      = spanFullEnd
	spanRangePartBits  = 23
	spanRangePartShift = spanRangePartBits
	spanRangePartMask  = (uint64(1) << spanRangePartBits) - 1
	spanRangePartMax   = uint64(1) << spanRangePartBits
)

// Detached returns a span that does not point into any file.
func Detached() Span {
	return Span{bits: spanDetached}
}

// SpanFromNumber creates a new span from a file id and a number.
// Returns false if number is not in the valid range (2..2^47).
func SpanFromNumber(id FileId, number uint64) (Span, bool) {
	if number < spanFullStart || number >= spanFullEnd {
		return Span{}, false
	}
	return packSpan(id, number), true
}

// SpanFromRange creates a new span from a raw byte range instead of a span number.
// If one of the range's parts exceeds the maximum value (2^23-1), it is saturated.
func SpanFromRange(id FileId, start, end int) Span {
	startU64 := uint64(start)
	endU64 := uint64(end)

	// Saturate to the maximum representable value (23 bits = 8388607)
	if startU64 > spanRangePartMask {
		startU64 = spanRangePartMask
	}
	if endU64 > spanRangePartMask {
		endU64 = spanRangePartMask
	}

	number := (startU64 << spanRangePartShift) | endU64
	return packSpan(id, spanRangeBase+number)
}

// SpanFromRaw constructs a Span from a raw uint64 value.
// Should only be used with values retrieved via Span.Raw().
func SpanFromRaw(v uint64) Span {
	return Span{bits: v}
}

// packSpan packs a file ID and the low bits into a span.
func packSpan(id FileId, low uint64) Span {
	bits := (uint64(id) << spanFileIdShift) | low
	return Span{bits: bits}
}

// IsDetached returns true if the span is detached (doesn't point to any file).
func (s Span) IsDetached() bool {
	return s.bits == spanDetached
}

// Id returns the file id that the span points into.
// Returns NoFile if the span is detached.
func (s Span) Id() FileId {
	fileIdBits := s.bits >> spanFileIdShift
	return FileId(fileIdBits)
}

// Number returns the unique number of the span within its source.
// This is an internal detail used by the compiler.
func (s Span) Number() uint64 {
	return s.bits & spanNumberMask
}

// Range extracts a raw byte range from the span, if it is a raw range span.
// Returns start, end, ok where ok is false if this is not a range span.
func (s Span) Range() (start, end int, ok bool) {
	number := s.Number()
	if number < spanRangeBase {
		return 0, 0, false
	}

	rangeBits := number - spanRangeBase
	start = int(rangeBits >> spanRangePartShift)
	end = int(rangeBits & spanRangePartMask)
	return start, end, true
}

// Raw returns the raw underlying uint64 value.
func (s Span) Raw() uint64 {
	return s.bits
}

// Or returns other if s is detached, and s otherwise.
func (s Span) Or(other Span) Span {
	if s.IsDetached() {
		return other
	}
	return s
}

// String implements fmt.Stringer.
func (s Span) String() string {
	if s.IsDetached() {
		return "Span(detached)"
	}
	if start, end, ok := s.Range(); ok {
		return fmt.Sprintf("Span(file=%d, range=%d..%d)", s.Id(), start, end)
	}
	return fmt.Sprintf("Span(file=%d, number=%d)", s.Id(), s.Number())
}

// FindSpan finds the first non-detached span in the slice.
// Returns Detached() if all spans are detached or the slice is empty.
func FindSpan(spans []Span) Span {
	for _, span := range spans {
		if !span.IsDetached() {
			return span
		}
	}
	return Detached()
}

// Spanned pairs a value with its source code location.
type Spanned[T any] struct {
	// V is the spanned value.
	V T
	// Span is the value's location in source code.
	Span Span
}

// NewSpanned creates a new Spanned from a value and its span.
func NewSpanned[T any](v T, span Span) Spanned[T] {
	return Spanned[T]{V: v, Span: span}
}

// SpannedDetached creates a new Spanned with a detached span.
func SpannedDetached[T any](v T) Spanned[T] {
	return Spanned[T]{V: v, Span: Detached()}
}

// Map transforms the value using the provided function, preserving the span.
func (s Spanned[T]) Map(f func(T) T) Spanned[T] {
	return Spanned[T]{V: f(s.V), Span: s.Span}
}

// String implements fmt.Stringer for Spanned values that implement Stringer.
func (s Spanned[T]) String() string {
	return fmt.Sprintf("%v", s.V)
}
