package foundations

import (
	"testing"
)

// --- StrReplace Tests ---

func TestStrReplace(t *testing.T) {
	tests := []struct {
		name        string
		s           Value
		pattern     Value
		replacement Value
		count       Value
		want        Value
		wantErr     bool
	}{
		// Basic replacement
		{"replace all", Str("hello world"), Str("o"), Str("0"), nil, Str("hell0 w0rld"), false},
		{"replace none found", Str("hello"), Str("x"), Str("y"), nil, Str("hello"), false},
		{"replace empty pattern", Str("abc"), Str(""), Str("-"), nil, Str("-a-b-c-"), false},
		{"replace with empty", Str("hello"), Str("l"), Str(""), nil, Str("heo"), false},

		// With count
		{"replace count 1", Str("aaa"), Str("a"), Str("b"), Int(1), Str("baa"), false},
		{"replace count 2", Str("aaa"), Str("a"), Str("b"), Int(2), Str("bba"), false},
		{"replace count 0", Str("aaa"), Str("a"), Str("b"), Int(0), Str("aaa"), false},
		{"replace count exceeds", Str("aaa"), Str("a"), Str("b"), Int(10), Str("bbb"), false},

		// UTF-8
		{"replace utf8", Str("日本語"), Str("本"), Str("X"), nil, Str("日X語"), false},
		{"replace emoji", Str("hello 👋 world 👋"), Str("👋"), Str("✋"), nil, Str("hello ✋ world ✋"), false},

		// Errors
		{"non-string input", Int(123), Str("1"), Str("x"), nil, nil, true},
		{"non-string pattern", Str("hello"), Int(1), Str("x"), nil, nil, true},
		{"non-string replacement", Str("hello"), Str("l"), Int(1), nil, nil, true},
		{"negative count", Str("hello"), Str("l"), Str("x"), Int(-1), nil, true},
		{"non-int count", Str("hello"), Str("l"), Str("x"), Str("1"), nil, true},

		// None count
		{"none count", Str("aaa"), Str("a"), Str("b"), None, Str("bbb"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrReplace(tt.s, tt.pattern, tt.replacement, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrReplace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("StrReplace() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- StrSplit Tests ---

func TestStrSplit(t *testing.T) {
	tests := []struct {
		name    string
		s       Value
		pattern Value
		want    *Array
		wantErr bool
	}{
		// Split on pattern
		{"split comma", Str("a,b,c"), Str(","), NewArray(Str("a"), Str("b"), Str("c")), false},
		{"split space", Str("a b c"), Str(" "), NewArray(Str("a"), Str("b"), Str("c")), false},
		{"split multi-char", Str("a::b::c"), Str("::"), NewArray(Str("a"), Str("b"), Str("c")), false},
		{"split not found", Str("abc"), Str(","), NewArray(Str("abc")), false},
		{"split empty result", Str(",a,"), Str(","), NewArray(Str(""), Str("a"), Str("")), false},

		// Split on empty string (by character)
		{"split empty to chars", Str("abc"), Str(""), NewArray(Str("a"), Str("b"), Str("c")), false},
		{"split utf8 chars", Str("日本"), Str(""), NewArray(Str("日"), Str("本")), false},

		// Split on whitespace (nil/none pattern)
		{"split whitespace nil", Str("a b  c"), nil, NewArray(Str("a"), Str("b"), Str("c")), false},
		{"split whitespace none", Str("  a  b  "), None, NewArray(Str("a"), Str("b")), false},
		{"split tabs newlines", Str("a\tb\nc"), nil, NewArray(Str("a"), Str("b"), Str("c")), false},

		// UTF-8
		{"split utf8 pattern", Str("日-本-語"), Str("-"), NewArray(Str("日"), Str("本"), Str("語")), false},

		// Empty string
		{"split empty string", Str(""), Str(","), NewArray(Str("")), false},
		{"split empty on empty", Str(""), Str(""), NewArray(), false},

		// Errors
		{"non-string input", Int(123), Str(","), nil, true},
		{"non-string pattern", Str("a,b"), Int(1), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrSplit(tt.s, tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrSplit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotArr, ok := got.(*Array)
				if !ok {
					t.Errorf("StrSplit() returned %T, want *Array", got)
					return
				}
				if !equal(gotArr, tt.want) {
					t.Errorf("StrSplit() = %v, want %v", gotArr, tt.want)
				}
			}
		})
	}
}

// --- StrTrim Tests ---

func TestStrTrim(t *testing.T) {
	tests := []struct {
		name    string
		s       Value
		pattern Value
		at      Value
		repeat  Value
		want    Value
		wantErr bool
	}{
		// Trim whitespace (default)
		{"trim whitespace both", Str("  hello  "), nil, nil, nil, Str("hello"), false},
		{"trim whitespace start", Str("  hello  "), nil, Str("start"), nil, Str("hello  "), false},
		{"trim whitespace end", Str("  hello  "), nil, Str("end"), nil, Str("  hello"), false},
		{"trim tabs newlines", Str("\t\nhello\n\t"), nil, nil, nil, Str("hello"), false},

		// Trim pattern
		{"trim pattern both", Str("xxhelloxx"), Str("x"), nil, nil, Str("hello"), false},
		{"trim pattern start", Str("xxhello"), Str("x"), Str("start"), nil, Str("hello"), false},
		{"trim pattern end", Str("helloxx"), Str("x"), Str("end"), nil, Str("hello"), false},

		// Trim multi-char pattern
		{"trim multi both", Str("ababhelloabab"), Str("ab"), nil, nil, Str("hello"), false},
		{"trim multi once", Str("ababhello"), Str("ab"), Str("start"), Bool(false), Str("abhello"), false},

		// Repeat false
		{"trim once both", Str("xxhelloxx"), Str("x"), nil, Bool(false), Str("xhellox"), false},

		// UTF-8
		{"trim utf8", Str("日日hello日日"), Str("日"), nil, nil, Str("hello"), false},

		// None pattern means whitespace
		{"trim none pattern", Str("  hi  "), None, nil, nil, Str("hi"), false},

		// Errors
		{"non-string input", Int(123), Str("x"), nil, nil, nil, true},
		{"non-string pattern", Str("hello"), Int(1), nil, nil, nil, true},
		{"invalid at", Str("hello"), Str("x"), Str("middle"), nil, nil, true},
		{"non-string at", Str("hello"), Str("x"), Int(1), nil, nil, true},
		{"non-bool repeat", Str("hello"), Str("x"), nil, Int(1), nil, true},

		// None at
		{"none at", Str("xxhelloxx"), Str("x"), None, nil, Str("hello"), false},
		// None repeat
		{"none repeat", Str("xxhelloxx"), Str("x"), nil, None, Str("hello"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrTrim(tt.s, tt.pattern, tt.at, tt.repeat)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrTrim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("StrTrim() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- StrNormalize Tests ---

func TestStrNormalize(t *testing.T) {
	// Test string with combining character: "é" can be e + combining accent
	composed := "café"               // é as single character (U+00E9)
	decomposed := "cafe\u0301"       // e + combining acute accent (U+0301)

	tests := []struct {
		name    string
		s       Value
		form    Value
		want    Value
		wantErr bool
	}{
		// NFC (default) - composed
		{"nfc default", Str(decomposed), nil, Str(composed), false},
		{"nfc explicit", Str(decomposed), Str("nfc"), Str(composed), false},
		{"nfc already composed", Str(composed), Str("nfc"), Str(composed), false},

		// NFD - decomposed
		{"nfd", Str(composed), Str("nfd"), Str(decomposed), false},

		// NFKC - compatibility composed
		{"nfkc", Str("ﬁ"), Str("nfkc"), Str("fi"), false},

		// NFKD - compatibility decomposed
		{"nfkd", Str("ﬁ"), Str("nfkd"), Str("fi"), false},

		// Case insensitive form
		{"uppercase NFC", Str(decomposed), Str("NFC"), Str(composed), false},
		{"mixed case nFc", Str(decomposed), Str("nFc"), Str(composed), false},

		// None form
		{"none form", Str(decomposed), None, Str(composed), false},

		// ASCII (no change)
		{"ascii nfc", Str("hello"), Str("nfc"), Str("hello"), false},

		// Errors
		{"non-string input", Int(123), Str("nfc"), nil, true},
		{"non-string form", Str("hello"), Int(1), nil, true},
		{"unknown form", Str("hello"), Str("xyz"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrNormalize(tt.s, tt.form)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrNormalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("StrNormalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- StrRev Tests ---

func TestStrRev(t *testing.T) {
	tests := []struct {
		name    string
		s       Value
		want    Value
		wantErr bool
	}{
		// Basic
		{"reverse basic", Str("hello"), Str("olleh"), false},
		{"reverse empty", Str(""), Str(""), false},
		{"reverse single", Str("a"), Str("a"), false},
		{"reverse palindrome", Str("abba"), Str("abba"), false},

		// UTF-8
		{"reverse utf8", Str("日本語"), Str("語本日"), false},
		{"reverse emoji", Str("👋🌍"), Str("🌍👋"), false},
		{"reverse mixed", Str("a日b"), Str("b日a"), false},

		// Errors
		{"non-string input", Int(123), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrRev(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrRev() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("StrRev() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- StrRepeat Tests ---

func TestStrRepeat(t *testing.T) {
	tests := []struct {
		name    string
		s       Value
		count   Value
		want    Value
		wantErr bool
	}{
		// Basic
		{"repeat 3", Str("ab"), Int(3), Str("ababab"), false},
		{"repeat 1", Str("x"), Int(1), Str("x"), false},
		{"repeat 0", Str("hello"), Int(0), Str(""), false},
		{"repeat empty", Str(""), Int(5), Str(""), false},

		// UTF-8
		{"repeat utf8", Str("日"), Int(3), Str("日日日"), false},
		{"repeat emoji", Str("👋"), Int(2), Str("👋👋"), false},

		// Errors
		{"non-string input", Int(123), Int(3), nil, true},
		{"non-int count", Str("x"), Str("3"), nil, true},
		{"negative count", Str("x"), Int(-1), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrRepeat(tt.s, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrRepeat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equal(got, tt.want) {
				t.Errorf("StrRepeat() = %v, want %v", got, tt.want)
			}
		})
	}
}
