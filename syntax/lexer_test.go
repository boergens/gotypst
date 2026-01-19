package syntax

import (
	"testing"
)

func TestLexerMarkup(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []SyntaxKind
	}{
		{
			name:  "simple text",
			input: "Hello world",
			want:  []SyntaxKind{Text, End},
		},
		{
			name:  "text with space",
			input: "Hello world",
			want:  []SyntaxKind{Text, End},
		},
		{
			name:  "heading",
			input: "= Heading",
			want:  []SyntaxKind{HeadingMarker, Space, Text, End},
		},
		{
			name:  "multiple equals heading",
			input: "== Subheading",
			want:  []SyntaxKind{HeadingMarker, Space, Text, End},
		},
		{
			name:  "list marker",
			input: "- item",
			want:  []SyntaxKind{ListMarker, Space, Text, End},
		},
		{
			name:  "enum marker",
			input: "+ item",
			want:  []SyntaxKind{EnumMarker, Space, Text, End},
		},
		{
			name:  "numbered enum",
			input: "1. first",
			want:  []SyntaxKind{EnumMarker, Space, Text, End},
		},
		{
			name:  "term marker",
			input: "/ term",
			want:  []SyntaxKind{TermMarker, Space, Text, End},
		},
		{
			name:  "strong",
			input: "*bold*",
			want:  []SyntaxKind{Star, Text, Star, End},
		},
		{
			name:  "emphasis",
			input: "_italic_",
			want:  []SyntaxKind{Underscore, Text, Underscore, End},
		},
		{
			name:  "inline code hash",
			input: "#code",
			want:  []SyntaxKind{Hash, Text, End},
		},
		{
			name:  "dollar sign",
			input: "$x$",
			want:  []SyntaxKind{Dollar, Text, Dollar, End},
		},
		{
			name:  "smart quotes",
			input: `"hello"`,
			want:  []SyntaxKind{SmartQuote, Text, SmartQuote, End},
		},
		{
			name:  "escape sequence",
			input: `\*not bold\*`,
			want:  []SyntaxKind{Escape, Text, Escape, End},
		},
		{
			name:  "unicode escape",
			input: `\u{1F600}`,
			want:  []SyntaxKind{Escape, End},
		},
		{
			name:  "linebreak",
			input: "line\\ break",
			want:  []SyntaxKind{Text, Linebreak, Space, Text, End},
		},
		{
			name:  "paragraph break",
			input: "para1\n\npara2",
			want:  []SyntaxKind{Text, Parbreak, Text, End},
		},
		{
			name:  "link http",
			input: "http://example.com",
			want:  []SyntaxKind{Link, End},
		},
		{
			name:  "link https",
			input: "https://example.com",
			want:  []SyntaxKind{Link, End},
		},
		{
			name:  "label",
			input: "<my-label>",
			want:  []SyntaxKind{Label, End},
		},
		{
			name:  "reference",
			input: "@my-ref",
			want:  []SyntaxKind{RefMarker, End},
		},
		{
			name:  "shorthand ellipsis",
			input: "...",
			want:  []SyntaxKind{Shorthand, End},
		},
		{
			name:  "shorthand em dash",
			input: "---",
			want:  []SyntaxKind{Shorthand, End},
		},
		{
			name:  "shorthand en dash",
			input: "--",
			want:  []SyntaxKind{Shorthand, End},
		},
		{
			name:  "shorthand tilde",
			input: "~",
			want:  []SyntaxKind{Shorthand, End},
		},
		{
			name:  "raw inline",
			input: "`code`",
			want:  []SyntaxKind{Raw, End},
		},
		{
			name:  "raw block",
			input: "```rust\nfn main() {}\n```",
			want:  []SyntaxKind{Raw, End},
		},
		{
			name:  "line comment",
			input: "text // comment\nmore",
			want:  []SyntaxKind{Text, Space, LineComment, Space, Text, End},
		},
		{
			name:  "block comment",
			input: "text /* comment */ more",
			want:  []SyntaxKind{Text, Space, BlockComment, Space, Text, End},
		},
		{
			name:  "nested block comment",
			input: "/* outer /* inner */ outer */",
			want:  []SyntaxKind{BlockComment, End},
		},
		{
			name:  "brackets",
			input: "[content]",
			want:  []SyntaxKind{LeftBracket, Text, RightBracket, End},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, ModeMarkup)
			var got []SyntaxKind
			for {
				kind, _ := l.Next()
				got = append(got, kind)
				if kind == End {
					break
				}
			}
			if len(got) != len(tt.want) {
				t.Errorf("got %d tokens, want %d", len(got), len(tt.want))
				t.Errorf("got: %v", got)
				t.Errorf("want: %v", tt.want)
				return
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("token %d: got %v, want %v", i, g, tt.want[i])
				}
			}
		})
	}
}

func TestLexerCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []SyntaxKind
	}{
		{
			name:  "identifier",
			input: "foo",
			want:  []SyntaxKind{Ident, End},
		},
		{
			name:  "underscore identifier",
			input: "_",
			want:  []SyntaxKind{Underscore, End},
		},
		{
			name:  "hyphenated identifier",
			input: "my-func",
			want:  []SyntaxKind{Ident, End},
		},
		{
			name:  "integer",
			input: "42",
			want:  []SyntaxKind{Int, End},
		},
		{
			name:  "float",
			input: "3.14",
			want:  []SyntaxKind{Float, End},
		},
		{
			name:  "float with exponent",
			input: "1.5e10",
			want:  []SyntaxKind{Float, End},
		},
		{
			name:  "numeric with unit",
			input: "12pt",
			want:  []SyntaxKind{Numeric, End},
		},
		{
			name:  "percentage",
			input: "50%",
			want:  []SyntaxKind{Numeric, End},
		},
		{
			name:  "binary number",
			input: "0b1010",
			want:  []SyntaxKind{Int, End},
		},
		{
			name:  "octal number",
			input: "0o755",
			want:  []SyntaxKind{Int, End},
		},
		{
			name:  "hex number",
			input: "0xff",
			want:  []SyntaxKind{Int, End},
		},
		{
			name:  "string",
			input: `"hello"`,
			want:  []SyntaxKind{Str, End},
		},
		{
			name:  "string with escape",
			input: `"hello\nworld"`,
			want:  []SyntaxKind{Str, End},
		},
		{
			name:  "keyword none",
			input: "none",
			want:  []SyntaxKind{None, End},
		},
		{
			name:  "keyword auto",
			input: "auto",
			want:  []SyntaxKind{Auto, End},
		},
		{
			name:  "keyword true",
			input: "true",
			want:  []SyntaxKind{Bool, End},
		},
		{
			name:  "keyword false",
			input: "false",
			want:  []SyntaxKind{Bool, End},
		},
		{
			name:  "keyword let",
			input: "let",
			want:  []SyntaxKind{Let, End},
		},
		{
			name:  "keyword if",
			input: "if",
			want:  []SyntaxKind{If, End},
		},
		{
			name:  "keyword else",
			input: "else",
			want:  []SyntaxKind{Else, End},
		},
		{
			name:  "keyword for",
			input: "for",
			want:  []SyntaxKind{For, End},
		},
		{
			name:  "keyword in",
			input: "in",
			want:  []SyntaxKind{In, End},
		},
		{
			name:  "keyword while",
			input: "while",
			want:  []SyntaxKind{While, End},
		},
		{
			name:  "keyword return",
			input: "return",
			want:  []SyntaxKind{Return, End},
		},
		{
			name:  "keyword import",
			input: "import",
			want:  []SyntaxKind{Import, End},
		},
		{
			name:  "operator ==",
			input: "==",
			want:  []SyntaxKind{EqEq, End},
		},
		{
			name:  "operator !=",
			input: "!=",
			want:  []SyntaxKind{ExclEq, End},
		},
		{
			name:  "operator <=",
			input: "<=",
			want:  []SyntaxKind{LtEq, End},
		},
		{
			name:  "operator >=",
			input: ">=",
			want:  []SyntaxKind{GtEq, End},
		},
		{
			name:  "operator +=",
			input: "+=",
			want:  []SyntaxKind{PlusEq, End},
		},
		{
			name:  "operator -=",
			input: "-=",
			want:  []SyntaxKind{HyphEq, End},
		},
		{
			name:  "operator =>",
			input: "=>",
			want:  []SyntaxKind{Arrow, End},
		},
		{
			name:  "dots operator",
			input: "..",
			want:  []SyntaxKind{Dots, End},
		},
		{
			name:  "braces",
			input: "{}",
			want:  []SyntaxKind{LeftBrace, RightBrace, End},
		},
		{
			name:  "brackets",
			input: "[]",
			want:  []SyntaxKind{LeftBracket, RightBracket, End},
		},
		{
			name:  "parens",
			input: "()",
			want:  []SyntaxKind{LeftParen, RightParen, End},
		},
		{
			name:  "comma",
			input: ",",
			want:  []SyntaxKind{Comma, End},
		},
		{
			name:  "semicolon",
			input: ";",
			want:  []SyntaxKind{Semicolon, End},
		},
		{
			name:  "colon",
			input: ":",
			want:  []SyntaxKind{Colon, End},
		},
		{
			name:  "simple expression",
			input: "let x = 1",
			want:  []SyntaxKind{Let, Space, Ident, Space, Eq, Space, Int, End},
		},
		{
			name:  "function call",
			input: "foo(1, 2)",
			want:  []SyntaxKind{Ident, LeftParen, Int, Comma, Space, Int, RightParen, End},
		},
		{
			name:  "field access not keyword",
			input: "x.let",
			want:  []SyntaxKind{Ident, Dot, Ident, End},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, ModeCode)
			var got []SyntaxKind
			for {
				kind, _ := l.Next()
				got = append(got, kind)
				if kind == End {
					break
				}
			}
			if len(got) != len(tt.want) {
				t.Errorf("got %d tokens, want %d", len(got), len(tt.want))
				t.Errorf("got: %v", got)
				t.Errorf("want: %v", tt.want)
				return
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("token %d: got %v, want %v", i, g, tt.want[i])
				}
			}
		})
	}
}

func TestLexerMath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []SyntaxKind
	}{
		{
			name:  "simple variable",
			input: "x",
			want:  []SyntaxKind{MathText, End},
		},
		{
			name:  "number",
			input: "42",
			want:  []SyntaxKind{MathText, End},
		},
		{
			name:  "float",
			input: "3.14",
			want:  []SyntaxKind{MathText, End},
		},
		{
			name:  "subscript",
			input: "_",
			want:  []SyntaxKind{Underscore, End},
		},
		{
			name:  "superscript",
			input: "^",
			want:  []SyntaxKind{Hat, End},
		},
		{
			name:  "fraction",
			input: "/",
			want:  []SyntaxKind{Slash, End},
		},
		{
			name:  "alignment point",
			input: "&",
			want:  []SyntaxKind{MathAlignPoint, End},
		},
		{
			name:  "primes",
			input: "'''",
			want:  []SyntaxKind{MathPrimes, End},
		},
		{
			name:  "root symbol",
			input: "√",
			want:  []SyntaxKind{Root, End},
		},
		{
			name:  "parens",
			input: "()",
			want:  []SyntaxKind{LeftParen, RightParen, End},
		},
		{
			name:  "shorthand arrow",
			input: "->",
			want:  []SyntaxKind{MathShorthand, End},
		},
		{
			name:  "shorthand double arrow",
			input: "=>",
			want:  []SyntaxKind{MathShorthand, End},
		},
		{
			name:  "shorthand less equal",
			input: "<=",
			want:  []SyntaxKind{MathShorthand, End},
		},
		{
			name:  "shorthand not equal",
			input: "!=",
			want:  []SyntaxKind{MathShorthand, End},
		},
		{
			name:  "hash for code",
			input: "#",
			want:  []SyntaxKind{Hash, End},
		},
		{
			name:  "dollar to exit",
			input: "$",
			want:  []SyntaxKind{Dollar, End},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, ModeMath)
			var got []SyntaxKind
			for {
				kind, _ := l.Next()
				got = append(got, kind)
				if kind == End {
					break
				}
			}
			if len(got) != len(tt.want) {
				t.Errorf("got %d tokens, want %d", len(got), len(tt.want))
				t.Errorf("got: %v", got)
				t.Errorf("want: %v", tt.want)
				return
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("token %d: got %v, want %v", i, g, tt.want[i])
				}
			}
		})
	}
}

func TestLexerErrors(t *testing.T) {
	tests := []struct {
		name    string
		mode    SyntaxMode
		input   string
		wantErr bool
	}{
		{
			name:    "unclosed unicode escape",
			mode:    ModeMarkup,
			input:   `\u{1234`,
			wantErr: true,
		},
		{
			name:    "invalid unicode codepoint",
			mode:    ModeMarkup,
			input:   `\u{FFFFFF}`,
			wantErr: true,
		},
		{
			name:    "unclosed label",
			mode:    ModeMarkup,
			input:   "<label",
			wantErr: true,
		},
		{
			name:    "empty label",
			mode:    ModeMarkup,
			input:   "<>",
			wantErr: false, // This is just text
		},
		{
			name:    "unclosed raw",
			mode:    ModeMarkup,
			input:   "`code",
			wantErr: true,
		},
		{
			name:    "unclosed string",
			mode:    ModeCode,
			input:   `"hello`,
			wantErr: true,
		},
		{
			name:    "invalid number suffix",
			mode:    ModeCode,
			input:   "42xyz",
			wantErr: true,
		},
		{
			name:    "unexpected block comment end",
			mode:    ModeMarkup,
			input:   "*/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, tt.mode)
			hasError := false
			for {
				kind, _ := l.Next()
				if kind == Error {
					hasError = true
				}
				if kind == End {
					break
				}
			}
			if hasError != tt.wantErr {
				t.Errorf("hasError = %v, want %v", hasError, tt.wantErr)
			}
		})
	}
}

func TestLexerModeSwitch(t *testing.T) {
	l := NewLexer("hello #code world", ModeMarkup)

	// First token is text
	kind, _ := l.Next()
	if kind != Text {
		t.Errorf("expected Text, got %v", kind)
	}

	// Space
	kind, _ = l.Next()
	if kind != Space {
		t.Errorf("expected Space, got %v", kind)
	}

	// Hash
	kind, _ = l.Next()
	if kind != Hash {
		t.Errorf("expected Hash, got %v", kind)
	}

	// Switch to code mode
	l.SetMode(ModeCode)

	// "code" should be an identifier now
	kind, _ = l.Next()
	if kind != Ident {
		t.Errorf("expected Ident, got %v", kind)
	}

	// Switch back to markup
	l.SetMode(ModeMarkup)

	// Space
	kind, _ = l.Next()
	if kind != Space {
		t.Errorf("expected Space, got %v", kind)
	}

	// "world" should be text again
	kind, _ = l.Next()
	if kind != Text {
		t.Errorf("expected Text, got %v", kind)
	}
}

func TestLexerNewline(t *testing.T) {
	l := NewLexer("line1\nline2", ModeMarkup)

	kind, _ := l.Next()
	if kind != Text {
		t.Errorf("expected Text, got %v", kind)
	}
	if l.Newline() {
		t.Error("expected no newline after first token")
	}

	kind, _ = l.Next()
	if kind != Space {
		t.Errorf("expected Space, got %v", kind)
	}
	if !l.Newline() {
		t.Error("expected newline after space token")
	}

	kind, _ = l.Next()
	if kind != Text {
		t.Errorf("expected Text, got %v", kind)
	}
}

func TestScanner(t *testing.T) {
	s := NewScanner("hello world")

	// Basic peek and eat
	if s.Peek() != 'h' {
		t.Errorf("expected 'h', got %c", s.Peek())
	}

	if s.Eat() != 'h' {
		t.Errorf("expected to eat 'h'")
	}

	if s.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", s.Cursor())
	}

	// Scout
	if s.Scout(-1) != 'h' {
		t.Errorf("expected scout(-1) to be 'h'")
	}

	if s.Scout(0) != 'e' {
		t.Errorf("expected scout(0) to be 'e'")
	}

	// EatWhile
	word := s.EatWhile(func(r rune) bool { return r >= 'a' && r <= 'z' })
	if word != "ello" {
		t.Errorf("expected 'ello', got %q", word)
	}

	// From
	if s.From(0) != "hello" {
		t.Errorf("expected 'hello', got %q", s.From(0))
	}

	// At
	if !s.At(" world") {
		t.Error("expected to be at ' world'")
	}

	// EatIf
	if !s.EatIf(' ') {
		t.Error("expected to eat space")
	}

	// After
	if s.After() != "world" {
		t.Errorf("expected 'world', got %q", s.After())
	}

	// Uneat
	s.Uneat()
	if s.Peek() != ' ' {
		t.Errorf("expected space after uneat, got %c", s.Peek())
	}
}
