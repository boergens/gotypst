package syntax

import "testing"

func TestSyntaxKindValues(t *testing.T) {
	// Verify the first few kinds have correct values
	tests := []struct {
		kind SyntaxKind
		want uint8
	}{
		{End, 0},
		{Error, 1},
		{Shebang, 2},
		{LineComment, 3},
		{BlockComment, 4},
	}
	for _, tt := range tests {
		if uint8(tt.kind) != tt.want {
			t.Errorf("%s = %d, want %d", tt.kind.Name(), tt.kind, tt.want)
		}
	}
}

func TestSyntaxKindIsGrouping(t *testing.T) {
	grouping := []SyntaxKind{LeftBrace, RightBrace, LeftBracket, RightBracket, LeftParen, RightParen}
	notGrouping := []SyntaxKind{End, Error, Plus, Minus, Ident}

	for _, k := range grouping {
		if !k.IsGrouping() {
			t.Errorf("%s.IsGrouping() = false, want true", k.Name())
		}
	}
	for _, k := range notGrouping {
		if k.IsGrouping() {
			t.Errorf("%s.IsGrouping() = true, want false", k.Name())
		}
	}
}

func TestSyntaxKindIsTerminator(t *testing.T) {
	terminators := []SyntaxKind{End, Semicolon, RightBrace, RightParen, RightBracket}
	notTerminators := []SyntaxKind{LeftBrace, LeftParen, Plus, Ident}

	for _, k := range terminators {
		if !k.IsTerminator() {
			t.Errorf("%s.IsTerminator() = false, want true", k.Name())
		}
	}
	for _, k := range notTerminators {
		if k.IsTerminator() {
			t.Errorf("%s.IsTerminator() = true, want false", k.Name())
		}
	}
}

func TestSyntaxKindIsBlock(t *testing.T) {
	blocks := []SyntaxKind{CodeBlock, ContentBlock}
	notBlocks := []SyntaxKind{End, LeftBrace, Ident}

	for _, k := range blocks {
		if !k.IsBlock() {
			t.Errorf("%s.IsBlock() = false, want true", k.Name())
		}
	}
	for _, k := range notBlocks {
		if k.IsBlock() {
			t.Errorf("%s.IsBlock() = true, want false", k.Name())
		}
	}
}

func TestSyntaxKindIsStmt(t *testing.T) {
	stmts := []SyntaxKind{LetBinding, SetRule, ShowRule, ModuleImport, ModuleInclude}
	notStmts := []SyntaxKind{End, Let, Ident, CodeBlock}

	for _, k := range stmts {
		if !k.IsStmt() {
			t.Errorf("%s.IsStmt() = false, want true", k.Name())
		}
	}
	for _, k := range notStmts {
		if k.IsStmt() {
			t.Errorf("%s.IsStmt() = true, want false", k.Name())
		}
	}
}

func TestSyntaxKindIsTrivia(t *testing.T) {
	trivia := []SyntaxKind{Shebang, LineComment, BlockComment, Space, Parbreak}
	notTrivia := []SyntaxKind{End, Text, Ident}

	for _, k := range trivia {
		if !k.IsTrivia() {
			t.Errorf("%s.IsTrivia() = false, want true", k.Name())
		}
	}
	for _, k := range notTrivia {
		if k.IsTrivia() {
			t.Errorf("%s.IsTrivia() = true, want false", k.Name())
		}
	}
}

func TestSyntaxKindIsKeyword(t *testing.T) {
	keywords := []SyntaxKind{Not, And, Or, None, Auto, Let, Set, Show, Context, If, Else, For, In, While, Break, Continue, Return, Import, Include, As}
	notKeywords := []SyntaxKind{End, Ident, Plus, LeftBrace}

	for _, k := range keywords {
		if !k.IsKeyword() {
			t.Errorf("%s.IsKeyword() = false, want true", k.Name())
		}
	}
	for _, k := range notKeywords {
		if k.IsKeyword() {
			t.Errorf("%s.IsKeyword() = true, want false", k.Name())
		}
	}
}

func TestSyntaxKindIsError(t *testing.T) {
	if !Error.IsError() {
		t.Error("Error.IsError() = false, want true")
	}
	if End.IsError() {
		t.Error("End.IsError() = true, want false")
	}
}

func TestSyntaxKindName(t *testing.T) {
	tests := []struct {
		kind SyntaxKind
		want string
	}{
		{End, "end of tokens"},
		{Error, "syntax error"},
		{LeftBrace, "opening brace"},
		{Let, "keyword `let`"},
		{Ident, "identifier"},
	}
	for _, tt := range tests {
		if got := tt.kind.Name(); got != tt.want {
			t.Errorf("%d.Name() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestSyntaxKindString(t *testing.T) {
	// String() should return the same as Name()
	if End.String() != End.Name() {
		t.Errorf("End.String() != End.Name()")
	}
}
