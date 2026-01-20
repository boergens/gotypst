package symbols

import (
	"testing"
)

func TestSymbolGet(t *testing.T) {
	sym := newSymbol("arrow", map[string]string{
		"":        "→",
		"r":       "→",
		"l":       "←",
		"r.double": "⇒",
	})

	tests := []struct {
		modifiers []string
		expected  string
	}{
		{nil, "→"},
		{[]string{}, "→"},
		{[]string{"r"}, "→"},
		{[]string{"l"}, "←"},
		{[]string{"r", "double"}, "⇒"},
		{[]string{"nonexistent"}, ""},
	}

	for _, tt := range tests {
		result := sym.Get(tt.modifiers...)
		if result != tt.expected {
			t.Errorf("Get(%v) = %q, want %q", tt.modifiers, result, tt.expected)
		}
	}
}

func TestSymbolDefault(t *testing.T) {
	sym := singleSymbol("alpha", "α")
	if sym.Default() != "α" {
		t.Errorf("Default() = %q, want %q", sym.Default(), "α")
	}

	symNoDefault := newSymbol("paren", map[string]string{
		"l": "(",
		"r": ")",
	})
	if symNoDefault.Default() != "" {
		t.Errorf("Default() = %q, want %q", symNoDefault.Default(), "")
	}
}

func TestSymbolHasVariant(t *testing.T) {
	sym := newSymbol("arrow", map[string]string{
		"":   "→",
		"r":  "→",
		"l":  "←",
	})

	if !sym.HasVariant() {
		t.Error("HasVariant() should be true for default")
	}
	if !sym.HasVariant("r") {
		t.Error("HasVariant('r') should be true")
	}
	if !sym.HasVariant("l") {
		t.Error("HasVariant('l') should be true")
	}
	if sym.HasVariant("t") {
		t.Error("HasVariant('t') should be false")
	}
}

func TestModuleLookup(t *testing.T) {
	// Test the Sym module
	tests := []struct {
		path     string
		expected string
		ok       bool
	}{
		// Greek letters
		{"alpha", "α", true},
		{"beta", "β", true},
		{"gamma", "γ", true},
		{"delta", "δ", true},
		{"pi", "π", true},
		{"pi.alt", "ϖ", true},

		// Uppercase Greek
		{"Alpha", "Α", true},
		{"Delta", "Δ", true},
		{"Pi", "Π", true},

		// Arrows
		{"arrow.r", "→", true},
		{"arrow.l", "←", true},
		{"arrow.t", "↑", true},
		{"arrow.b", "↓", true},
		{"arrow.r.double", "⇒", true},
		{"arrow.l.double", "⇐", true},

		// Math operators
		{"plus", "+", true},
		{"minus", "−", true},
		{"times", "×", true},
		{"div", "÷", true},
		{"eq", "=", true},
		{"eq.not", "≠", true},

		// Logic
		{"forall", "∀", true},
		{"exists", "∃", true},
		{"exists.not", "∄", true},
		{"and", "∧", true},
		{"or", "∨", true},
		{"not", "¬", true},

		// Set theory
		{"emptyset", "∅", true},
		{"in", "∈", true},
		{"in.not", "∉", true},
		{"subset", "⊂", true},
		{"supset", "⊃", true},
		{"union", "∪", true},
		{"inter", "∩", true},

		// Calculus
		{"infinity", "∞", true},
		{"partial", "∂", true},
		{"nabla", "∇", true},
		{"sum", "∑", true},
		{"product", "∏", true},
		{"integral", "∫", true},
		{"integral.double", "∬", true},
		{"integral.triple", "∭", true},

		// Delimiters
		{"paren.l", "(", true},
		{"paren.r", ")", true},
		{"brace.l", "{", true},
		{"brace.r", "}", true},
		{"bracket.l", "[", true},
		{"bracket.r", "]", true},

		// Double-struck
		{"NN", "ℕ", true},
		{"ZZ", "ℤ", true},
		{"QQ", "ℚ", true},
		{"RR", "ℝ", true},
		{"CC", "ℂ", true},

		// Punctuation
		{"dots.h", "…", true},
		{"dots.v", "⋮", true},

		// Nonexistent
		{"nonexistent", "", false},
		{"arrow.nonexistent", "", false},
	}

	for _, tt := range tests {
		result, ok := Sym.Lookup(tt.path)
		if ok != tt.ok {
			t.Errorf("Lookup(%q) ok = %v, want %v", tt.path, ok, tt.ok)
		}
		if result != tt.expected {
			t.Errorf("Lookup(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}

func TestEmojiModuleLookup(t *testing.T) {
	tests := []struct {
		path     string
		expected string
		ok       bool
	}{
		{"fire", "🔥", true},
		{"rocket", "🚀", true},
		{"heart", "❤", true},
		{"heart.blue", "💙", true},
		{"star", "⭐", true},
		{"check", "✔", true},
		{"face.smile", "😊", true},
		{"face.laugh", "😂", true},
		{"hand.wave", "👋", true},
		{"nonexistent", "", false},
	}

	for _, tt := range tests {
		result, ok := Emoji.Lookup(tt.path)
		if ok != tt.ok {
			t.Errorf("Emoji.Lookup(%q) ok = %v, want %v", tt.path, ok, tt.ok)
		}
		if result != tt.expected {
			t.Errorf("Emoji.Lookup(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}

func TestRootModule(t *testing.T) {
	// Test that root module has sym and emoji submodules
	if Root.Submodules["sym"] != Sym {
		t.Error("Root.Submodules['sym'] should be Sym")
	}
	if Root.Submodules["emoji"] != Emoji {
		t.Error("Root.Submodules['emoji'] should be Emoji")
	}

	// Test lookup through root
	result, ok := Root.Lookup("sym.alpha")
	if !ok || result != "α" {
		t.Errorf("Root.Lookup('sym.alpha') = %q, %v, want 'α', true", result, ok)
	}

	result, ok = Root.Lookup("emoji.fire")
	if !ok || result != "🔥" {
		t.Errorf("Root.Lookup('emoji.fire') = %q, %v, want '🔥', true", result, ok)
	}
}

func TestSymModuleNotEmpty(t *testing.T) {
	if len(Sym.Symbols) == 0 {
		t.Error("Sym module should have symbols")
	}

	// Check some expected symbols exist
	expectedSymbols := []string{
		"alpha", "beta", "gamma", "delta", "pi",
		"arrow", "plus", "minus", "times", "div",
		"forall", "exists", "in", "subset",
		"infinity", "sum", "integral",
	}

	for _, name := range expectedSymbols {
		if Sym.Symbols[name] == nil {
			t.Errorf("Sym module should have symbol %q", name)
		}
	}
}

func TestEmojiModuleNotEmpty(t *testing.T) {
	if len(Emoji.Symbols) == 0 {
		t.Error("Emoji module should have symbols")
	}

	// Check some expected emoji exist
	expectedEmoji := []string{
		"fire", "rocket", "heart", "star",
		"face", "hand", "sun", "moon",
	}

	for _, name := range expectedEmoji {
		if Emoji.Symbols[name] == nil {
			t.Errorf("Emoji module should have symbol %q", name)
		}
	}
}

func TestGenderSubmodule(t *testing.T) {
	genderMod := Sym.Submodules["gender"]
	if genderMod == nil {
		t.Fatal("Sym should have gender submodule")
	}

	tests := []struct {
		name     string
		expected string
	}{
		{"female", "♀"},
		{"male", "♂"},
		{"trans", "⚧"},
		{"neuter", "⚲"},
		{"intersex", "⚥"},
	}

	for _, tt := range tests {
		sym := genderMod.Symbols[tt.name]
		if sym == nil {
			t.Errorf("gender submodule should have symbol %q", tt.name)
			continue
		}
		if sym.Default() != tt.expected {
			t.Errorf("gender.%s = %q, want %q", tt.name, sym.Default(), tt.expected)
		}
	}
}

func TestSpecialCharacters(t *testing.T) {
	// Test that control characters and special Unicode work correctly
	tests := []struct {
		path     string
		expected string
	}{
		{"space", " "},
		{"space.nobreak", "\u00A0"},
		{"space.thin", "\u2009"},
		{"zwj", "\u200D"},
		{"zwnj", "\u200C"},
	}

	for _, tt := range tests {
		result, ok := Sym.Lookup(tt.path)
		if !ok {
			t.Errorf("Lookup(%q) should succeed", tt.path)
			continue
		}
		if result != tt.expected {
			t.Errorf("Lookup(%q) = %q (len=%d), want %q (len=%d)",
				tt.path, result, len(result), tt.expected, len(tt.expected))
		}
	}
}

func TestAllVariants(t *testing.T) {
	arrow := Sym.Symbols["arrow"]
	if arrow == nil {
		t.Fatal("arrow symbol should exist")
	}

	variants := arrow.AllVariants()
	if len(variants) == 0 {
		t.Error("arrow should have variants")
	}

	// Check that some expected variants exist
	variantMap := make(map[string]bool)
	for _, v := range variants {
		variantMap[v] = true
	}

	expectedVariants := []string{"r", "l", "t", "b", "r.double", "l.double"}
	for _, v := range expectedVariants {
		if !variantMap[v] {
			t.Errorf("arrow should have variant %q", v)
		}
	}
}
