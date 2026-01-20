// Package text provides text-related functions for the Typst standard library.
package text

import (
	"github.com/boergens/gotypst/eval"
)

// HighlightedSpan represents a span of highlighted text with styling.
type HighlightedSpan struct {
	Text  string
	Style HighlightStyle
}

// HighlightStyle represents the style for a highlighted span.
type HighlightStyle struct {
	// Color is the text color (RGB hex like "ff0000")
	Color string
	// Bold indicates if the text should be bold
	Bold bool
	// Italic indicates if the text should be italic
	Italic bool
	// Underline indicates if the text should be underlined
	Underline bool
}

// SyntaxHighlighter is the interface for syntax highlighting implementations.
// Highlighters take raw code text and return highlighted spans.
type SyntaxHighlighter interface {
	// Highlight takes source code and returns highlighted spans.
	// Returns nil if highlighting is not possible or not needed.
	Highlight(code string, lang string) []HighlightedSpan

	// SupportedLanguages returns a list of language identifiers this highlighter supports.
	SupportedLanguages() []string
}

// HighlightHooks manages syntax highlighting hooks.
type HighlightHooks struct {
	// highlighters is a map from language to highlighter
	highlighters map[string]SyntaxHighlighter
	// defaultHighlighter is used when no language-specific highlighter is found
	defaultHighlighter SyntaxHighlighter
}

// NewHighlightHooks creates a new HighlightHooks instance.
func NewHighlightHooks() *HighlightHooks {
	return &HighlightHooks{
		highlighters: make(map[string]SyntaxHighlighter),
	}
}

// Register registers a syntax highlighter for specific languages.
func (h *HighlightHooks) Register(highlighter SyntaxHighlighter) {
	for _, lang := range highlighter.SupportedLanguages() {
		h.highlighters[lang] = highlighter
	}
}

// RegisterDefault registers a default highlighter used when no language-specific one is found.
func (h *HighlightHooks) RegisterDefault(highlighter SyntaxHighlighter) {
	h.defaultHighlighter = highlighter
}

// Unregister removes a highlighter for a specific language.
func (h *HighlightHooks) Unregister(lang string) {
	delete(h.highlighters, lang)
}

// GetHighlighter returns the highlighter for a given language, or nil if none registered.
func (h *HighlightHooks) GetHighlighter(lang string) SyntaxHighlighter {
	if hl, ok := h.highlighters[lang]; ok {
		return hl
	}
	return h.defaultHighlighter
}

// Highlight highlights code using the registered highlighter for the language.
// Returns nil if no highlighter is registered for the language.
func (h *HighlightHooks) Highlight(code string, lang string) []HighlightedSpan {
	highlighter := h.GetHighlighter(lang)
	if highlighter == nil {
		return nil
	}
	return highlighter.Highlight(code, lang)
}

// HasHighlighter returns true if a highlighter is registered for the given language.
func (h *HighlightHooks) HasHighlighter(lang string) bool {
	_, ok := h.highlighters[lang]
	return ok || h.defaultHighlighter != nil
}

// HighlightRawElement highlights a raw element's text.
// Returns the highlighted spans, or nil if no highlighting is available.
func (h *HighlightHooks) HighlightRawElement(element *eval.RawElement) []HighlightedSpan {
	if element.Lang == "" {
		return nil
	}
	return h.Highlight(element.Text, element.Lang)
}

// DefaultHighlightHooks is the global default highlight hooks instance.
// This can be used when a World implementation doesn't provide custom hooks.
var DefaultHighlightHooks = NewHighlightHooks()

// NoOpHighlighter is a highlighter that returns the text as a single unhighlighted span.
// This is useful as a fallback when no real syntax highlighting is available.
type NoOpHighlighter struct{}

// Highlight returns the text as a single unhighlighted span.
func (NoOpHighlighter) Highlight(code string, lang string) []HighlightedSpan {
	return []HighlightedSpan{{Text: code}}
}

// SupportedLanguages returns an empty list since this matches no specific language.
func (NoOpHighlighter) SupportedLanguages() []string {
	return nil
}

// SimpleKeywordHighlighter provides basic keyword highlighting for common languages.
type SimpleKeywordHighlighter struct {
	keywords map[string]map[string]HighlightStyle
}

// NewSimpleKeywordHighlighter creates a new simple keyword highlighter.
func NewSimpleKeywordHighlighter() *SimpleKeywordHighlighter {
	return &SimpleKeywordHighlighter{
		keywords: map[string]map[string]HighlightStyle{
			"go": {
				"func":      {Color: "0000ff", Bold: true},
				"return":    {Color: "0000ff", Bold: true},
				"if":        {Color: "0000ff", Bold: true},
				"else":      {Color: "0000ff", Bold: true},
				"for":       {Color: "0000ff", Bold: true},
				"range":     {Color: "0000ff", Bold: true},
				"package":   {Color: "0000ff", Bold: true},
				"import":    {Color: "0000ff", Bold: true},
				"type":      {Color: "0000ff", Bold: true},
				"struct":    {Color: "0000ff", Bold: true},
				"interface": {Color: "0000ff", Bold: true},
				"var":       {Color: "0000ff", Bold: true},
				"const":     {Color: "0000ff", Bold: true},
				"nil":       {Color: "ff6600", Italic: true},
				"true":      {Color: "ff6600", Italic: true},
				"false":     {Color: "ff6600", Italic: true},
			},
			"python": {
				"def":    {Color: "0000ff", Bold: true},
				"class":  {Color: "0000ff", Bold: true},
				"return": {Color: "0000ff", Bold: true},
				"if":     {Color: "0000ff", Bold: true},
				"else":   {Color: "0000ff", Bold: true},
				"elif":   {Color: "0000ff", Bold: true},
				"for":    {Color: "0000ff", Bold: true},
				"while":  {Color: "0000ff", Bold: true},
				"import": {Color: "0000ff", Bold: true},
				"from":   {Color: "0000ff", Bold: true},
				"try":    {Color: "0000ff", Bold: true},
				"except": {Color: "0000ff", Bold: true},
				"None":   {Color: "ff6600", Italic: true},
				"True":   {Color: "ff6600", Italic: true},
				"False":  {Color: "ff6600", Italic: true},
			},
			"javascript": {
				"function": {Color: "0000ff", Bold: true},
				"return":   {Color: "0000ff", Bold: true},
				"if":       {Color: "0000ff", Bold: true},
				"else":     {Color: "0000ff", Bold: true},
				"for":      {Color: "0000ff", Bold: true},
				"while":    {Color: "0000ff", Bold: true},
				"const":    {Color: "0000ff", Bold: true},
				"let":      {Color: "0000ff", Bold: true},
				"var":      {Color: "0000ff", Bold: true},
				"class":    {Color: "0000ff", Bold: true},
				"null":     {Color: "ff6600", Italic: true},
				"true":     {Color: "ff6600", Italic: true},
				"false":    {Color: "ff6600", Italic: true},
				"undefined": {Color: "ff6600", Italic: true},
			},
			"rust": {
				"fn":     {Color: "0000ff", Bold: true},
				"let":    {Color: "0000ff", Bold: true},
				"mut":    {Color: "0000ff", Bold: true},
				"return": {Color: "0000ff", Bold: true},
				"if":     {Color: "0000ff", Bold: true},
				"else":   {Color: "0000ff", Bold: true},
				"for":    {Color: "0000ff", Bold: true},
				"while":  {Color: "0000ff", Bold: true},
				"loop":   {Color: "0000ff", Bold: true},
				"match":  {Color: "0000ff", Bold: true},
				"use":    {Color: "0000ff", Bold: true},
				"mod":    {Color: "0000ff", Bold: true},
				"pub":    {Color: "0000ff", Bold: true},
				"struct": {Color: "0000ff", Bold: true},
				"impl":   {Color: "0000ff", Bold: true},
				"trait":  {Color: "0000ff", Bold: true},
				"None":   {Color: "ff6600", Italic: true},
				"Some":   {Color: "ff6600", Italic: true},
				"true":   {Color: "ff6600", Italic: true},
				"false":  {Color: "ff6600", Italic: true},
			},
		},
	}
}

// Highlight implements SyntaxHighlighter.
func (h *SimpleKeywordHighlighter) Highlight(code string, lang string) []HighlightedSpan {
	keywords, ok := h.keywords[lang]
	if !ok {
		return []HighlightedSpan{{Text: code}}
	}

	var spans []HighlightedSpan
	current := ""
	word := ""

	for _, r := range code {
		if isWordChar(r) {
			word += string(r)
		} else {
			// Emit accumulated word
			if word != "" {
				if style, isKeyword := keywords[word]; isKeyword {
					if current != "" {
						spans = append(spans, HighlightedSpan{Text: current})
						current = ""
					}
					spans = append(spans, HighlightedSpan{Text: word, Style: style})
				} else {
					current += word
				}
				word = ""
			}
			current += string(r)
		}
	}

	// Emit final word
	if word != "" {
		if style, isKeyword := keywords[word]; isKeyword {
			if current != "" {
				spans = append(spans, HighlightedSpan{Text: current})
				current = ""
			}
			spans = append(spans, HighlightedSpan{Text: word, Style: style})
		} else {
			current += word
		}
	}

	// Emit remaining text
	if current != "" {
		spans = append(spans, HighlightedSpan{Text: current})
	}

	return spans
}

// SupportedLanguages implements SyntaxHighlighter.
func (h *SimpleKeywordHighlighter) SupportedLanguages() []string {
	langs := make([]string, 0, len(h.keywords))
	for lang := range h.keywords {
		langs = append(langs, lang)
	}
	return langs
}

// isWordChar returns true if the rune is a valid word character.
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// RegisterBuiltinHighlighters registers the built-in syntax highlighters.
func RegisterBuiltinHighlighters(hooks *HighlightHooks) {
	hooks.Register(NewSimpleKeywordHighlighter())
	hooks.RegisterDefault(&NoOpHighlighter{})
}

func init() {
	// Register built-in highlighters by default
	RegisterBuiltinHighlighters(DefaultHighlightHooks)
}
