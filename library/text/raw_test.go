package text

import (
	"testing"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/syntax"
)

func TestRawFunc(t *testing.T) {
	rawFunc := RawFunc()
	if rawFunc.Func == nil {
		t.Fatal("RawFunc returned nil Func")
	}
	if rawFunc.Func.Name == nil || *rawFunc.Func.Name != "raw" {
		t.Error("RawFunc should have name 'raw'")
	}
}

func TestRawImpl(t *testing.T) {
	tests := []struct {
		name     string
		args     *eval.Args
		wantText string
		wantLang string
		wantBlock bool
		wantErr  bool
	}{
		{
			name: "simple text",
			args: makeArgs(
				positionalArg(eval.StrValue("hello world")),
			),
			wantText: "hello world",
			wantLang: "",
			wantBlock: false,
		},
		{
			name: "text with lang",
			args: makeArgs(
				positionalArg(eval.StrValue("fn main() {}")),
				namedArg("lang", eval.StrValue("rust")),
			),
			wantText: "fn main() {}",
			wantLang: "rust",
			wantBlock: false,
		},
		{
			name: "block code",
			args: makeArgs(
				positionalArg(eval.StrValue("print('hello')")),
				namedArg("lang", eval.StrValue("python")),
				namedArg("block", eval.True),
			),
			wantText: "print('hello')",
			wantLang: "python",
			wantBlock: true,
		},
		{
			name: "multiline code",
			args: makeArgs(
				positionalArg(eval.StrValue("line1\nline2\nline3")),
			),
			wantText: "line1\nline2\nline3",
			wantLang: "",
			wantBlock: false,
		},
		{
			name: "missing text argument",
			args: makeArgs(),
			wantErr: true,
		},
		{
			name: "wrong type for text",
			args: makeArgs(
				positionalArg(eval.IntValue(42)),
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rawImpl(nil, tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content, ok := result.(eval.ContentValue)
			if !ok {
				t.Fatalf("expected ContentValue, got %T", result)
			}
			if len(content.Content.Elements) != 1 {
				t.Fatalf("expected 1 element, got %d", len(content.Content.Elements))
			}
			raw, ok := content.Content.Elements[0].(*eval.RawElement)
			if !ok {
				t.Fatalf("expected RawElement, got %T", content.Content.Elements[0])
			}
			if raw.Text != tt.wantText {
				t.Errorf("text = %q, want %q", raw.Text, tt.wantText)
			}
			if raw.Lang != tt.wantLang {
				t.Errorf("lang = %q, want %q", raw.Lang, tt.wantLang)
			}
			if raw.Block != tt.wantBlock {
				t.Errorf("block = %v, want %v", raw.Block, tt.wantBlock)
			}
		})
	}
}

func TestRawLang(t *testing.T) {
	tests := []struct {
		element *eval.RawElement
		want    eval.Value
	}{
		{
			element: &eval.RawElement{Lang: "python"},
			want:    eval.StrValue("python"),
		},
		{
			element: &eval.RawElement{Lang: ""},
			want:    eval.None,
		},
	}

	for _, tt := range tests {
		result := RawLang(tt.element)
		if result != tt.want {
			t.Errorf("RawLang(%+v) = %v, want %v", tt.element, result, tt.want)
		}
	}
}

func TestRawText(t *testing.T) {
	element := &eval.RawElement{Text: "hello world"}
	result := RawText(element)
	want := eval.StrValue("hello world")
	if result != want {
		t.Errorf("RawText() = %v, want %v", result, want)
	}
}

func TestRawBlock(t *testing.T) {
	tests := []struct {
		element *eval.RawElement
		want    eval.Value
	}{
		{element: &eval.RawElement{Block: true}, want: eval.True},
		{element: &eval.RawElement{Block: false}, want: eval.False},
	}

	for _, tt := range tests {
		result := RawBlock(tt.element)
		if result != tt.want {
			t.Errorf("RawBlock(%+v) = %v, want %v", tt.element, result, tt.want)
		}
	}
}

func TestRawLines(t *testing.T) {
	tests := []struct {
		text  string
		count int
	}{
		{text: "single line", count: 1},
		{text: "line1\nline2", count: 2},
		{text: "line1\nline2\nline3", count: 3},
		{text: "", count: 1}, // Empty string produces one empty line
		{text: "\n", count: 2}, // Trailing newline produces empty line
	}

	for _, tt := range tests {
		element := &eval.RawElement{Text: tt.text}
		result := RawLines(element)
		arr, ok := result.(eval.ArrayValue)
		if !ok {
			t.Fatalf("RawLines() returned %T, want ArrayValue", result)
		}
		if len(arr) != tt.count {
			t.Errorf("RawLines(%q) has %d lines, want %d", tt.text, len(arr), tt.count)
		}
	}
}

// Helper functions for creating test arguments

func makeArgs(args ...eval.Arg) *eval.Args {
	return &eval.Args{
		Span:  syntax.Detached(),
		Items: args,
	}
}

func positionalArg(v eval.Value) eval.Arg {
	return eval.Arg{
		Span:  syntax.Detached(),
		Name:  nil,
		Value: syntax.Spanned[eval.Value]{V: v, Span: syntax.Detached()},
	}
}

func namedArg(name string, v eval.Value) eval.Arg {
	n := name
	return eval.Arg{
		Span:  syntax.Detached(),
		Name:  &n,
		Value: syntax.Spanned[eval.Value]{V: v, Span: syntax.Detached()},
	}
}
