package foundations

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// String transformation functions
// Translated from foundations/str.rs

// StrReplace replaces occurrences of a pattern in a string.
// If count is nil, all occurrences are replaced.
// If count is provided, only the first count occurrences are replaced.
func StrReplace(s, pattern, replacement Value, count Value) (Value, error) {
	str, ok := s.(Str)
	if !ok {
		return nil, &OpError{Message: "replace: expected string, got " + s.Type()}
	}

	pat, ok := pattern.(Str)
	if !ok {
		return nil, &OpError{Message: "replace: pattern must be a string, got " + pattern.Type()}
	}

	repl, ok := replacement.(Str)
	if !ok {
		return nil, &OpError{Message: "replace: replacement must be a string, got " + replacement.Type()}
	}

	// Handle count parameter
	n := -1 // -1 means replace all
	if count != nil {
		if _, isNone := count.(NoneValue); !isNone {
			cnt, ok := count.(Int)
			if !ok {
				return nil, &OpError{Message: "replace: count must be an integer, got " + count.Type()}
			}
			if cnt < 0 {
				return nil, &OpError{Message: "replace: count must be non-negative"}
			}
			n = int(cnt)
		}
	}

	result := strings.Replace(string(str), string(pat), string(repl), n)
	return Str(result), nil
}

// StrSplit splits a string at occurrences of a pattern.
// If pattern is nil or none, splits on whitespace.
// If pattern is empty string, splits into individual grapheme clusters.
func StrSplit(s, pattern Value) (Value, error) {
	str, ok := s.(Str)
	if !ok {
		return nil, &OpError{Message: "split: expected string, got " + s.Type()}
	}

	// Determine what to split on
	var parts []string

	if pattern == nil {
		// Default: split on whitespace
		parts = strings.Fields(string(str))
	} else if _, isNone := pattern.(NoneValue); isNone {
		// Split on whitespace
		parts = strings.Fields(string(str))
	} else {
		pat, ok := pattern.(Str)
		if !ok {
			return nil, &OpError{Message: "split: pattern must be a string, got " + pattern.Type()}
		}

		if pat == "" {
			// Empty string: split into grapheme clusters (approximated by runes)
			// For proper grapheme cluster handling, we'd need a segmentation library
			// For now, split by runes (Unicode code points)
			parts = make([]string, 0, utf8.RuneCountInString(string(str)))
			for _, r := range string(str) {
				parts = append(parts, string(r))
			}
		} else {
			parts = strings.Split(string(str), string(pat))
		}
	}

	// Convert to array of Str values
	items := make([]Value, len(parts))
	for i, p := range parts {
		items[i] = Str(p)
	}

	return &Array{items: items}, nil
}

// StrTrim removes a pattern from the start and/or end of a string.
// at: "start", "end", or nil (both)
// repeat: if true (default), removes all occurrences; if false, removes once
func StrTrim(s, pattern Value, at Value, repeat Value) (Value, error) {
	str, ok := s.(Str)
	if !ok {
		return nil, &OpError{Message: "trim: expected string, got " + s.Type()}
	}

	// Determine what to trim
	var trimFunc func(string) string
	var trimStart, trimEnd bool = true, true

	// Parse "at" parameter
	if at != nil {
		if _, isNone := at.(NoneValue); !isNone {
			atStr, ok := at.(Str)
			if !ok {
				return nil, &OpError{Message: "trim: 'at' must be a string, got " + at.Type()}
			}
			switch string(atStr) {
			case "start":
				trimStart, trimEnd = true, false
			case "end":
				trimStart, trimEnd = false, true
			default:
				return nil, &OpError{Message: "trim: 'at' must be \"start\" or \"end\", got " + string(atStr)}
			}
		}
	}

	// Parse "repeat" parameter (default true)
	doRepeat := true
	if repeat != nil {
		if _, isNone := repeat.(NoneValue); !isNone {
			rep, ok := repeat.(Bool)
			if !ok {
				return nil, &OpError{Message: "trim: 'repeat' must be a boolean, got " + repeat.Type()}
			}
			doRepeat = bool(rep)
		}
	}

	// Determine trim pattern
	if pattern == nil || func() bool { _, isNone := pattern.(NoneValue); return isNone }() {
		// Default: trim whitespace
		if trimStart && trimEnd {
			trimFunc = strings.TrimSpace
		} else if trimStart {
			trimFunc = func(s string) string {
				return strings.TrimLeftFunc(s, unicode.IsSpace)
			}
		} else {
			trimFunc = func(s string) string {
				return strings.TrimRightFunc(s, unicode.IsSpace)
			}
		}
	} else {
		pat, ok := pattern.(Str)
		if !ok {
			return nil, &OpError{Message: "trim: pattern must be a string, got " + pattern.Type()}
		}
		patStr := string(pat)

		if doRepeat {
			// Repeated trim
			if trimStart && trimEnd {
				trimFunc = func(s string) string {
					for strings.HasPrefix(s, patStr) {
						s = strings.TrimPrefix(s, patStr)
					}
					for strings.HasSuffix(s, patStr) {
						s = strings.TrimSuffix(s, patStr)
					}
					return s
				}
			} else if trimStart {
				trimFunc = func(s string) string {
					for strings.HasPrefix(s, patStr) {
						s = strings.TrimPrefix(s, patStr)
					}
					return s
				}
			} else {
				trimFunc = func(s string) string {
					for strings.HasSuffix(s, patStr) {
						s = strings.TrimSuffix(s, patStr)
					}
					return s
				}
			}
		} else {
			// Single trim
			if trimStart && trimEnd {
				trimFunc = func(s string) string {
					s = strings.TrimPrefix(s, patStr)
					s = strings.TrimSuffix(s, patStr)
					return s
				}
			} else if trimStart {
				trimFunc = func(s string) string {
					return strings.TrimPrefix(s, patStr)
				}
			} else {
				trimFunc = func(s string) string {
					return strings.TrimSuffix(s, patStr)
				}
			}
		}
	}

	result := trimFunc(string(str))
	return Str(result), nil
}

// StrNormalize normalizes a string to the specified Unicode normal form.
// form: "nfc" (default), "nfd", "nfkc", or "nfkd"
func StrNormalize(s, form Value) (Value, error) {
	str, ok := s.(Str)
	if !ok {
		return nil, &OpError{Message: "normalize: expected string, got " + s.Type()}
	}

	// Determine normalization form (default NFC)
	formStr := "nfc"
	if form != nil {
		if _, isNone := form.(NoneValue); !isNone {
			f, ok := form.(Str)
			if !ok {
				return nil, &OpError{Message: "normalize: form must be a string, got " + form.Type()}
			}
			formStr = strings.ToLower(string(f))
		}
	}

	var normalizer norm.Form
	switch formStr {
	case "nfc":
		normalizer = norm.NFC
	case "nfd":
		normalizer = norm.NFD
	case "nfkc":
		normalizer = norm.NFKC
	case "nfkd":
		normalizer = norm.NFKD
	default:
		return nil, &OpError{
			Message: "normalize: unknown form \"" + formStr + "\"",
			Hint:    "expected \"nfc\", \"nfd\", \"nfkc\", or \"nfkd\"",
		}
	}

	result := normalizer.String(string(str))
	return Str(result), nil
}

// StrRev reverses a string.
// Operates on Unicode code points (runes), preserving UTF-8 encoding.
func StrRev(s Value) (Value, error) {
	str, ok := s.(Str)
	if !ok {
		return nil, &OpError{Message: "rev: expected string, got " + s.Type()}
	}

	// Convert to runes, reverse, convert back
	runes := []rune(string(str))
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return Str(string(runes)), nil
}

// StrRepeat repeats a string n times.
// Note: This is also available via the * operator (str * int).
func StrRepeat(s, count Value) (Value, error) {
	str, ok := s.(Str)
	if !ok {
		return nil, &OpError{Message: "repeat: expected string, got " + s.Type()}
	}

	n, ok := count.(Int)
	if !ok {
		return nil, &OpError{Message: "repeat: count must be an integer, got " + count.Type()}
	}

	if n < 0 {
		return nil, &OpError{Message: "repeat: count must be non-negative"}
	}

	if n == 0 {
		return Str(""), nil
	}

	result := strings.Repeat(string(str), int(n))
	return Str(result), nil
}
