package text

import (
	"testing"

	"github.com/boergens/gotypst/eval"
)

func TestHighlightHooks(t *testing.T) {
	hooks := NewHighlightHooks()
	if hooks == nil {
		t.Fatal("NewHighlightHooks returned nil")
	}
}

func TestHighlightHooksRegister(t *testing.T) {
	hooks := NewHighlightHooks()
	highlighter := NewSimpleKeywordHighlighter()

	hooks.Register(highlighter)

	// Should have highlighters for the supported languages
	for _, lang := range highlighter.SupportedLanguages() {
		if !hooks.HasHighlighter(lang) {
			t.Errorf("expected highlighter for %s after registration", lang)
		}
	}
}

func TestHighlightHooksUnregister(t *testing.T) {
	hooks := NewHighlightHooks()
	highlighter := NewSimpleKeywordHighlighter()

	hooks.Register(highlighter)
	hooks.Unregister("go")

	if hl := hooks.GetHighlighter("go"); hl != nil {
		t.Error("expected nil highlighter for 'go' after unregister")
	}

	// Other languages should still work
	if !hooks.HasHighlighter("python") {
		t.Error("expected highlighter for python to still exist")
	}
}

func TestHighlightHooksDefault(t *testing.T) {
	hooks := NewHighlightHooks()
	hooks.RegisterDefault(&NoOpHighlighter{})

	// Unknown language should use default
	if !hooks.HasHighlighter("unknown-lang") {
		t.Error("expected default highlighter to match unknown language")
	}

	spans := hooks.Highlight("code", "unknown-lang")
	if len(spans) != 1 || spans[0].Text != "code" {
		t.Error("expected default highlighter to return text as-is")
	}
}

func TestNoOpHighlighter(t *testing.T) {
	h := NoOpHighlighter{}

	spans := h.Highlight("func main() {}", "go")
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Text != "func main() {}" {
		t.Errorf("expected text 'func main() {}', got %q", spans[0].Text)
	}
	if spans[0].Style != (HighlightStyle{}) {
		t.Error("expected empty style from NoOpHighlighter")
	}

	langs := h.SupportedLanguages()
	if langs != nil {
		t.Errorf("expected nil supported languages, got %v", langs)
	}
}

func TestSimpleKeywordHighlighter(t *testing.T) {
	h := NewSimpleKeywordHighlighter()

	tests := []struct {
		code string
		lang string
		wantSpans int
	}{
		{code: "func main() {}", lang: "go", wantSpans: 2}, // "func", " main() {}"
		{code: "def hello():", lang: "python", wantSpans: 2}, // "def", " hello():"
		{code: "plain text", lang: "go", wantSpans: 1}, // no keywords, single span
		{code: "return nil", lang: "go", wantSpans: 3}, // "return", " ", "nil"
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			spans := h.Highlight(tt.code, tt.lang)
			if len(spans) != tt.wantSpans {
				t.Errorf("Highlight(%q, %q) = %d spans, want %d",
					tt.code, tt.lang, len(spans), tt.wantSpans)
				for i, s := range spans {
					t.Logf("  span[%d] = %q (style: %+v)", i, s.Text, s.Style)
				}
			}
		})
	}
}

func TestSimpleKeywordHighlighterKeywordDetection(t *testing.T) {
	h := NewSimpleKeywordHighlighter()

	// Test Go keywords
	spans := h.Highlight("func return if else for range", "go")

	// Count highlighted spans (those with non-empty style)
	highlighted := 0
	for _, s := range spans {
		if s.Style.Color != "" || s.Style.Bold {
			highlighted++
		}
	}

	// Should have 6 highlighted keywords
	if highlighted != 6 {
		t.Errorf("expected 6 highlighted keywords, got %d", highlighted)
	}
}

func TestSimpleKeywordHighlighterSupportedLanguages(t *testing.T) {
	h := NewSimpleKeywordHighlighter()
	langs := h.SupportedLanguages()

	expected := map[string]bool{
		"go":         true,
		"python":     true,
		"javascript": true,
		"rust":       true,
	}

	for _, lang := range langs {
		if !expected[lang] {
			t.Errorf("unexpected supported language: %s", lang)
		}
		delete(expected, lang)
	}

	for lang := range expected {
		t.Errorf("expected supported language not found: %s", lang)
	}
}

func TestHighlightRawElement(t *testing.T) {
	hooks := NewHighlightHooks()
	RegisterBuiltinHighlighters(hooks)

	element := &eval.RawElement{
		Text:  "func main() {}",
		Lang:  "go",
		Block: true,
	}

	spans := hooks.HighlightRawElement(element)
	if spans == nil {
		t.Fatal("expected spans, got nil")
	}
	if len(spans) == 0 {
		t.Error("expected non-empty spans")
	}

	// Verify the text is reconstructed correctly
	var reconstructed string
	for _, s := range spans {
		reconstructed += s.Text
	}
	if reconstructed != element.Text {
		t.Errorf("reconstructed text = %q, want %q", reconstructed, element.Text)
	}
}

func TestHighlightRawElementNoLang(t *testing.T) {
	hooks := NewHighlightHooks()
	RegisterBuiltinHighlighters(hooks)

	element := &eval.RawElement{
		Text:  "plain text",
		Lang:  "",
		Block: false,
	}

	spans := hooks.HighlightRawElement(element)
	if spans != nil {
		t.Errorf("expected nil spans for raw element without lang, got %v", spans)
	}
}

func TestDefaultHighlightHooks(t *testing.T) {
	// DefaultHighlightHooks should be initialized with built-in highlighters
	if DefaultHighlightHooks == nil {
		t.Fatal("DefaultHighlightHooks is nil")
	}

	// Should support common languages
	if !DefaultHighlightHooks.HasHighlighter("go") {
		t.Error("DefaultHighlightHooks should support 'go'")
	}
	if !DefaultHighlightHooks.HasHighlighter("python") {
		t.Error("DefaultHighlightHooks should support 'python'")
	}
}

func TestIsWordChar(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{' ', false},
		{'\n', false},
		{'\t', false},
		{'(', false},
		{')', false},
		{'{', false},
		{'}', false},
		{'.', false},
		{',', false},
	}

	for _, tt := range tests {
		if got := isWordChar(tt.r); got != tt.want {
			t.Errorf("isWordChar(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}
