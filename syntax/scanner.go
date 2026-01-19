package syntax

import (
	"unicode/utf8"
)

// Scanner is a string iterator with peek/eat capabilities.
// It tracks a cursor position and provides methods for consuming characters.
type Scanner struct {
	text   string
	cursor int
}

// NewScanner creates a new scanner for the given text.
func NewScanner(text string) *Scanner {
	return &Scanner{text: text, cursor: 0}
}

// String returns the underlying text being scanned.
func (s *Scanner) String() string {
	return s.text
}

// Cursor returns the current position in the text.
func (s *Scanner) Cursor() int {
	return s.cursor
}

// Jump sets the cursor to the given position.
func (s *Scanner) Jump(pos int) {
	if pos < 0 {
		pos = 0
	} else if pos > len(s.text) {
		pos = len(s.text)
	}
	s.cursor = pos
}

// Advance moves the cursor forward by the given number of bytes.
func (s *Scanner) Advance(by int) {
	s.Jump(s.cursor + by)
}

// Done returns true if the scanner has reached the end of the text.
func (s *Scanner) Done() bool {
	return s.cursor >= len(s.text)
}

// Peek returns the next rune without consuming it.
// Returns 0 if at end.
func (s *Scanner) Peek() rune {
	if s.cursor >= len(s.text) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(s.text[s.cursor:])
	return r
}

// Scout looks at a rune at a relative offset from the cursor.
// Positive offsets look ahead, negative offsets look behind.
// Returns 0 if the position is out of bounds.
func (s *Scanner) Scout(offset int) rune {
	if offset == 0 {
		return s.Peek()
	}
	if offset > 0 {
		// Look ahead
		pos := s.cursor
		for i := 0; i < offset; i++ {
			if pos >= len(s.text) {
				return 0
			}
			_, size := utf8.DecodeRuneInString(s.text[pos:])
			pos += size
		}
		if pos >= len(s.text) {
			return 0
		}
		r, _ := utf8.DecodeRuneInString(s.text[pos:])
		return r
	}
	// Look behind (negative offset)
	pos := s.cursor
	for i := 0; i < -offset; i++ {
		if pos <= 0 {
			return 0
		}
		_, size := utf8.DecodeLastRuneInString(s.text[:pos])
		pos -= size
	}
	if pos < 0 {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(s.text[pos:])
	return r
}

// Eat consumes and returns the next rune.
// Returns 0 if at end.
func (s *Scanner) Eat() rune {
	if s.cursor >= len(s.text) {
		return 0
	}
	r, size := utf8.DecodeRuneInString(s.text[s.cursor:])
	s.cursor += size
	return r
}

// Uneat moves back one rune.
func (s *Scanner) Uneat() {
	if s.cursor <= 0 {
		return
	}
	_, size := utf8.DecodeLastRuneInString(s.text[:s.cursor])
	s.cursor -= size
}

// EatIf consumes the next rune if it matches the given rune.
// Returns true if consumed.
func (s *Scanner) EatIf(r rune) bool {
	if s.Peek() == r {
		s.Eat()
		return true
	}
	return false
}

// EatIfStr consumes the string if it matches at the current position.
// Returns true if consumed.
func (s *Scanner) EatIfStr(str string) bool {
	if s.At(str) {
		s.cursor += len(str)
		return true
	}
	return false
}

// EatWhile consumes runes while the predicate returns true.
// Returns the consumed string.
func (s *Scanner) EatWhile(pred func(rune) bool) string {
	start := s.cursor
	for !s.Done() {
		r := s.Peek()
		if !pred(r) {
			break
		}
		s.Eat()
	}
	return s.text[start:s.cursor]
}

// EatUntil consumes runes until the predicate returns true.
// Returns the consumed string.
func (s *Scanner) EatUntil(pred func(rune) bool) string {
	start := s.cursor
	for !s.Done() {
		r := s.Peek()
		if pred(r) {
			break
		}
		s.Eat()
	}
	return s.text[start:s.cursor]
}

// EatNewline consumes a newline character (handles \r\n).
// Returns true if a newline was consumed.
func (s *Scanner) EatNewline() bool {
	ate := s.EatIf('\n') || s.EatIf('\r') || s.EatIf('\x0B') ||
		s.EatIf('\x0C') || s.EatIfStr("\u0085") ||
		s.EatIfStr("\u2028") || s.EatIfStr("\u2029")
	if ate && s.Before()[len(s.Before())-1] == '\r' {
		s.EatIf('\n')
	}
	return ate
}

// At checks if the current position starts with the given string.
func (s *Scanner) At(str string) bool {
	if s.cursor+len(str) > len(s.text) {
		return false
	}
	return s.text[s.cursor:s.cursor+len(str)] == str
}

// AtRune checks if the current position matches a rune predicate.
func (s *Scanner) AtRune(pred func(rune) bool) bool {
	if s.Done() {
		return false
	}
	return pred(s.Peek())
}

// AtAny checks if the current position matches any of the given runes.
func (s *Scanner) AtAny(runes ...rune) bool {
	if s.Done() {
		return false
	}
	r := s.Peek()
	for _, target := range runes {
		if r == target {
			return true
		}
	}
	return false
}

// AtAnyStr checks if the current position matches any of the given strings.
func (s *Scanner) AtAnyStr(strs ...string) bool {
	for _, str := range strs {
		if s.At(str) {
			return true
		}
	}
	return false
}

// Before returns the text before the cursor.
func (s *Scanner) Before() string {
	return s.text[:s.cursor]
}

// After returns the text after the cursor.
func (s *Scanner) After() string {
	return s.text[s.cursor:]
}

// From returns the text from the given position to the cursor.
func (s *Scanner) From(start int) string {
	if start < 0 {
		start = 0
	}
	if start > s.cursor {
		return ""
	}
	return s.text[start:s.cursor]
}

// To returns the text from the cursor to the given position.
func (s *Scanner) To(end int) string {
	if end > len(s.text) {
		end = len(s.text)
	}
	if s.cursor > end {
		return ""
	}
	return s.text[s.cursor:end]
}

// Get returns a substring of the text.
func (s *Scanner) Get(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s.text) {
		end = len(s.text)
	}
	if start >= end {
		return ""
	}
	return s.text[start:end]
}

// Clone creates a copy of the scanner with the same position.
func (s *Scanner) Clone() *Scanner {
	return &Scanner{text: s.text, cursor: s.cursor}
}
