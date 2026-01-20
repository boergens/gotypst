package foundations

import (
	"testing"
)

func TestStrLen(t *testing.T) {
	tests := []struct {
		name  string
		input Str
		want  Int
	}{
		{"empty", "", 0},
		{"ascii", "hello", 5},
		{"unicode", "héllo", 5},
		{"emoji", "👋🏽", 1}, // Emoji with skin tone modifier is one grapheme cluster
		{"combining", "e\u0301", 1}, // e + combining acute accent = one grapheme cluster
		{"mixed", "a👨‍👩‍👧b", 3}, // a + family emoji + b
		{"chinese", "你好", 2},
		{"newlines", "a\nb\nc", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StrLen(tt.input)
			if got != tt.want {
				t.Errorf("StrLen(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestStrIsEmpty(t *testing.T) {
	tests := []struct {
		input Str
		want  Bool
	}{
		{"", true},
		{"a", false},
		{" ", false},
		{"👋", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := StrIsEmpty(tt.input)
			if got != tt.want {
				t.Errorf("StrIsEmpty(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStrFirst(t *testing.T) {
	tests := []struct {
		name  string
		input Str
		want  Value
	}{
		{"empty", "", None},
		{"single", "a", Str("a")},
		{"multiple", "abc", Str("a")},
		{"emoji", "👋hello", Str("👋")},
		{"combining", "e\u0301bc", Str("e\u0301")}, // e + combining acute
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StrFirst(tt.input)
			if !valuesEqual(got, tt.want) {
				t.Errorf("StrFirst(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStrLast(t *testing.T) {
	tests := []struct {
		name  string
		input Str
		want  Value
	}{
		{"empty", "", None},
		{"single", "a", Str("a")},
		{"multiple", "abc", Str("c")},
		{"emoji at end", "hello👋", Str("👋")},
		{"combining at end", "abc\u0301", Str("c\u0301")}, // c + combining acute
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StrLast(tt.input)
			if !valuesEqual(got, tt.want) {
				t.Errorf("StrLast(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStrAt(t *testing.T) {
	tests := []struct {
		name    string
		input   Str
		index   Int
		want    Value
		wantErr bool
	}{
		{"first", "hello", 0, Str("h"), false},
		{"middle", "hello", 2, Str("l"), false},
		{"last", "hello", 4, Str("o"), false},
		{"negative", "hello", -1, Str("o"), false},
		{"negative second", "hello", -2, Str("l"), false},
		{"out of bounds positive", "hello", 5, nil, true},
		{"out of bounds negative", "hello", -6, nil, true},
		{"emoji", "a👋b", 1, Str("👋"), false},
		{"empty string", "", 0, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrAt(tt.input, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrAt(%q, %d) error = %v, wantErr %v", tt.input, tt.index, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !valuesEqual(got, tt.want) {
				t.Errorf("StrAt(%q, %d) = %v, want %v", tt.input, tt.index, got, tt.want)
			}
		})
	}
}

func TestStrSlice(t *testing.T) {
	intPtr := func(i Int) *Int { return &i }

	tests := []struct {
		name    string
		input   Str
		start   Int
		end     *Int
		want    Str
		wantErr bool
	}{
		{"full slice no end", "hello", 0, nil, "hello", false},
		{"from start", "hello", 0, intPtr(3), "hel", false},
		{"from middle", "hello", 2, intPtr(4), "ll", false},
		{"to end", "hello", 2, nil, "llo", false},
		{"negative start", "hello", -3, nil, "llo", false},
		{"negative end", "hello", 0, intPtr(-2), "hel", false},
		{"both negative", "hello", -3, intPtr(-1), "ll", false},
		{"inverted range", "hello", 3, intPtr(1), "", false},
		{"single char", "hello", 1, intPtr(2), "e", false},
		{"empty result", "hello", 2, intPtr(2), "", false},
		{"emoji", "a👋b", 0, intPtr(2), "a👋", false},
		{"start beyond length", "hello", 10, nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrSlice(tt.input, tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrSlice(%q, %d, %v) error = %v, wantErr %v", tt.input, tt.start, tt.end, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotStr, ok := got.(Str)
				if !ok {
					t.Errorf("StrSlice(%q, %d, %v) returned non-Str: %T", tt.input, tt.start, tt.end, got)
					return
				}
				if gotStr != tt.want {
					t.Errorf("StrSlice(%q, %d, %v) = %q, want %q", tt.input, tt.start, tt.end, gotStr, tt.want)
				}
			}
		})
	}
}

func TestStrClusters(t *testing.T) {
	tests := []struct {
		name  string
		input Str
		want  []Str
	}{
		{"empty", "", nil},
		{"ascii", "abc", []Str{"a", "b", "c"}},
		{"emoji", "a👋b", []Str{"a", "👋", "b"}},
		{"combining", "e\u0301", []Str{"e\u0301"}},
		{"family emoji", "👨‍👩‍👧", []Str{"👨‍👩‍👧"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StrClusters(tt.input)
			if len(tt.want) == 0 {
				if got.Len() != 0 {
					t.Errorf("StrClusters(%q) = %d items, want 0", tt.input, got.Len())
				}
				return
			}
			if got.Len() != len(tt.want) {
				t.Errorf("StrClusters(%q) = %d items, want %d", tt.input, got.Len(), len(tt.want))
				return
			}
			for i, w := range tt.want {
				item := got.At(i)
				s, ok := item.(Str)
				if !ok {
					t.Errorf("StrClusters(%q)[%d] = %T, want Str", tt.input, i, item)
					continue
				}
				if s != w {
					t.Errorf("StrClusters(%q)[%d] = %q, want %q", tt.input, i, s, w)
				}
			}
		})
	}
}

func TestStrCodepoints(t *testing.T) {
	tests := []struct {
		name  string
		input Str
		want  []Str
	}{
		{"empty", "", nil},
		{"ascii", "abc", []Str{"a", "b", "c"}},
		{"emoji", "👋", []Str{"👋"}},
		{"combining separate", "e\u0301", []Str{"e", "\u0301"}}, // Two codepoints
		{"chinese", "你好", []Str{"你", "好"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StrCodepoints(tt.input)
			if len(tt.want) == 0 {
				if got.Len() != 0 {
					t.Errorf("StrCodepoints(%q) = %d items, want 0", tt.input, got.Len())
				}
				return
			}
			if got.Len() != len(tt.want) {
				t.Errorf("StrCodepoints(%q) = %d items, want %d", tt.input, got.Len(), len(tt.want))
				return
			}
			for i, w := range tt.want {
				item := got.At(i)
				s, ok := item.(Str)
				if !ok {
					t.Errorf("StrCodepoints(%q)[%d] = %T, want Str", tt.input, i, item)
					continue
				}
				if s != w {
					t.Errorf("StrCodepoints(%q)[%d] = %q, want %q", tt.input, i, s, w)
				}
			}
		})
	}
}

// valuesEqual compares two Values for equality in tests.
func valuesEqual(a, b Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case NoneValue:
		_, ok := b.(NoneValue)
		return ok
	case Str:
		bv, ok := b.(Str)
		return ok && av == bv
	case Int:
		bv, ok := b.(Int)
		return ok && av == bv
	default:
		return false
	}
}
