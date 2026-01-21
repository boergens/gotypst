package syntax

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Lexer is an iterator over a source code string which returns tokens.
type Lexer struct {
	s       *Scanner
	mode    SyntaxMode
	newline bool
	err     *SyntaxError
}

// NewLexer creates a new lexer with the given mode.
func NewLexer(text string, mode SyntaxMode) *Lexer {
	return &Lexer{
		s:       NewScanner(text),
		mode:    mode,
		newline: false,
		err:     nil,
	}
}

// Mode returns the current lexing mode.
func (l *Lexer) Mode() SyntaxMode {
	return l.mode
}

// SetMode changes the lexing mode.
func (l *Lexer) SetMode(mode SyntaxMode) {
	l.mode = mode
}

// Cursor returns the current position in the string.
func (l *Lexer) Cursor() int {
	return l.s.Cursor()
}

// Jump sets the cursor to the given position.
func (l *Lexer) Jump(index int) {
	l.s.Jump(index)
}

// Newline returns true if the last token contained a newline.
func (l *Lexer) Newline() bool {
	return l.newline
}

// Column returns the number of characters until the most recent newline from an index.
func (l *Lexer) Column(index int) int {
	s := l.s.Clone()
	s.Jump(index)
	count := 0
	for i := len(s.Before()) - 1; i >= 0; i-- {
		r := rune(s.Before()[i])
		if IsNewline(r) {
			break
		}
		count++
	}
	return count
}

// error creates a syntax error and returns the Error kind.
func (l *Lexer) error(message string) SyntaxKind {
	l.err = NewSyntaxError(message)
	return Error
}

// hint adds a hint to the current error.
func (l *Lexer) hint(message string) {
	if l.err != nil {
		l.err.AddHint(message)
	}
}

// Next returns the next token in the text.
func (l *Lexer) Next() (SyntaxKind, *SyntaxNode) {
	l.err = nil
	start := l.s.Cursor()
	l.newline = false

	c := l.s.Eat()
	var kind SyntaxKind

	switch {
	case c == 0:
		kind = End
	case IsSpace(c, l.mode):
		kind = l.whitespace(start, c)
	case c == '#' && start == 0 && l.s.EatIf('!'):
		kind = l.shebang()
	case c == '/' && l.s.EatIf('/'):
		kind = l.lineComment()
	case c == '/' && l.s.EatIf('*'):
		kind = l.blockComment()
	case c == '*' && l.s.EatIf('/'):
		kind = l.error("unexpected end of block comment")
		l.hint("consider escaping the `*` with a backslash or opening the block comment with `/*`")
	case c == '`' && l.mode != ModeMath:
		return l.raw()
	default:
		switch l.mode {
		case ModeMarkup:
			kind = l.markup(start, c)
		case ModeMath:
			var node *SyntaxNode
			kind, node = l.math(start, c)
			if node != nil {
				return kind, node
			}
		case ModeCode:
			kind = l.code(start, c)
		}
	}

	text := l.s.From(start)
	var node *SyntaxNode
	if l.err != nil {
		node = ErrorNode(l.err, text)
		l.err = nil
	} else {
		node = Leaf(kind, text)
	}
	return kind, node
}

// whitespace consumes whitespace characters.
func (l *Lexer) whitespace(start int, c rune) SyntaxKind {
	l.s.EatWhile(func(r rune) bool { return IsSpace(r, l.mode) })
	text := l.s.From(start)

	// Count newlines
	newlines := 0
	if c == ' ' && len(text) == 1 {
		// Optimize for single space
		newlines = 0
	} else {
		newlines = countNewlines(text)
	}

	l.newline = newlines > 0
	if l.mode == ModeMarkup && newlines >= 2 {
		return Parbreak
	}
	return Space
}

func (l *Lexer) shebang() SyntaxKind {
	l.s.EatUntil(IsNewline)
	return Shebang
}

func (l *Lexer) lineComment() SyntaxKind {
	l.s.EatUntil(IsNewline)
	return LineComment
}

func (l *Lexer) blockComment() SyntaxKind {
	state := '_'
	depth := 1

	for {
		c := l.s.Eat()
		if c == 0 {
			break
		}
		switch {
		case state == '*' && c == '/':
			depth--
			if depth == 0 {
				return BlockComment
			}
			state = '_'
		case state == '/' && c == '*':
			depth++
			state = '_'
		default:
			state = c
		}
	}

	return BlockComment
}

// markup handles tokens in markup mode.
func (l *Lexer) markup(start int, c rune) SyntaxKind {
	switch c {
	case '\\':
		return l.backslash()
	case 'h':
		if l.s.EatIfStr("ttp://") {
			return l.link()
		}
		if l.s.EatIfStr("ttps://") {
			return l.link()
		}
		return l.text()
	case '<':
		if l.s.AtRune(IsIDContinue) {
			return l.label()
		}
		return l.text()
	case '@':
		if l.s.AtRune(IsIDContinue) {
			return l.refMarker()
		}
		return l.text()
	case '.':
		if l.s.EatIfStr("..") {
			return Shorthand
		}
		return l.text()
	case '-':
		if l.s.EatIfStr("--") {
			return Shorthand
		}
		if l.s.EatIf('-') {
			return Shorthand
		}
		if l.s.EatIf('?') {
			return Shorthand
		}
		if l.s.AtRune(unicode.IsDigit) {
			return Shorthand
		}
		if l.spaceOrEnd() {
			return ListMarker
		}
		return l.text()
	case '*':
		if !l.inWord() {
			return Star
		}
		return l.text()
	case '_':
		if !l.inWord() {
			return Underscore
		}
		return l.text()
	case '#':
		return Hash
	case '[':
		return LeftBracket
	case ']':
		return RightBracket
	case '\'':
		return SmartQuote
	case '"':
		return SmartQuote
	case '$':
		return Dollar
	case '~':
		return Shorthand
	case ':':
		return Colon
	case '=':
		l.s.EatWhile(func(r rune) bool { return r == '=' })
		if l.spaceOrEnd() {
			return HeadingMarker
		}
		return l.text()
	case '+':
		if l.spaceOrEnd() {
			return EnumMarker
		}
		return l.text()
	case '/':
		if l.spaceOrEnd() {
			return TermMarker
		}
		return l.text()
	}

	if c >= '0' && c <= '9' {
		return l.numbering(start)
	}

	return l.text()
}

func (l *Lexer) backslash() SyntaxKind {
	if l.s.EatIfStr("u{") {
		hex := l.s.EatWhile(func(r rune) bool {
			return (r >= '0' && r <= '9') ||
				(r >= 'a' && r <= 'f') ||
				(r >= 'A' && r <= 'F')
		})
		if !l.s.EatIf('}') {
			return l.error("unclosed Unicode escape sequence")
		}
		val, err := strconv.ParseUint(hex, 16, 32)
		if err != nil || !isValidCodepoint(uint32(val)) {
			return l.error(fmt.Sprintf("invalid Unicode codepoint: %s", hex))
		}
		return Escape
	}

	if l.s.Done() || unicode.IsSpace(l.s.Peek()) {
		return Linebreak
	}
	l.s.Eat()
	return Escape
}

func isValidCodepoint(val uint32) bool {
	return val <= 0x10FFFF && (val < 0xD800 || val > 0xDFFF)
}

// raw handles raw blocks (backtick-delimited code).
func (l *Lexer) raw() (SyntaxKind, *SyntaxNode) {
	start := l.s.Cursor() - 1

	// Count opening backticks
	backticks := 1
	for l.s.EatIf('`') {
		backticks++
	}

	// Special case for ``
	if backticks == 2 {
		nodes := []*SyntaxNode{
			Leaf(RawDelim, "`"),
			Leaf(RawDelim, "`"),
		}
		return Raw, Inner(Raw, nodes)
	}

	// Find end of raw text
	found := 0
	for found < backticks {
		c := l.s.Eat()
		if c == 0 {
			err := NewSyntaxError("unclosed raw text")
			return Error, ErrorNode(err, l.s.From(start))
		}
		if c == '`' {
			found++
		} else {
			found = 0
		}
	}
	end := l.s.Cursor()

	nodes := make([]*SyntaxNode, 0, 3)
	prevStart := start

	pushRaw := func(kind SyntaxKind) {
		text := l.s.Get(prevStart, l.s.Cursor())
		nodes = append(nodes, Leaf(kind, text))
		prevStart = l.s.Cursor()
	}

	// Opening delimiter
	l.s.Jump(start + backticks)
	pushRaw(RawDelim)

	if backticks >= 3 {
		l.blockyRaw(end-backticks, &nodes, &prevStart)
	} else {
		l.inlineRaw(end-backticks, &nodes, &prevStart)
	}

	// Closing delimiter
	l.s.Jump(end)
	pushRaw(RawDelim)

	return Raw, Inner(Raw, nodes)
}

func (l *Lexer) blockyRaw(innerEnd int, nodes *[]*SyntaxNode, prevStart *int) {
	pushRaw := func(kind SyntaxKind) {
		text := l.s.Get(*prevStart, l.s.Cursor())
		*nodes = append(*nodes, Leaf(kind, text))
		*prevStart = l.s.Cursor()
	}

	// Language tag
	tag := l.s.EatUntil(func(r rune) bool {
		return unicode.IsSpace(r) || r == '`'
	})
	if len(tag) > 0 {
		pushRaw(RawLang)
	}

	// Split into lines
	content := l.s.To(innerEnd)
	lines := splitNewlines(content)

	// Determine dedent level
	dedent := -1
	for i, line := range lines {
		if i == 0 {
			continue // Skip first line
		}
		spaces := countLeadingSpaces(line)
		if spaces < len(line) { // Line has non-whitespace
			if dedent < 0 || spaces < dedent {
				dedent = spaces
			}
		}
	}
	// Also consider the last line for dedent
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		spaces := countLeadingSpaces(lastLine)
		if dedent < 0 || spaces < dedent {
			dedent = spaces
		}
	}
	if dedent < 0 {
		dedent = 0
	}

	// Check if last line is all whitespace
	trimLastLine := len(lines) > 0 && isAllWhitespace(lines[len(lines)-1])
	if trimLastLine {
		lines = lines[:len(lines)-1]
	} else if len(lines) > 0 {
		// Trim trailing space if last line ends with backtick
		lastLine := lines[len(lines)-1]
		if strings.HasSuffix(strings.TrimRight(lastLine, " \t"), "`") {
			lines[len(lines)-1] = strings.TrimSuffix(lastLine, " ")
		}
	}

	// Process first line
	if len(lines) > 0 {
		firstLine := lines[0]
		if isAllWhitespace(firstLine) {
			l.s.Advance(len(firstLine))
		} else {
			lineEnd := l.s.Cursor() + len(firstLine)
			if l.s.EatIf(' ') {
				pushRaw(RawTrimmed)
			}
			l.s.Jump(lineEnd)
			pushRaw(Text)
		}
		lines = lines[1:]
	}

	// Process remaining lines
	for _, line := range lines {
		offset := min(dedent, countLeadingSpaces(line))
		l.s.EatNewline()
		l.s.Advance(offset)
		pushRaw(RawTrimmed)
		l.s.Advance(len(line) - offset)
		pushRaw(Text)
	}

	// Add final trimmed if needed
	if l.s.Cursor() < innerEnd {
		l.s.Jump(innerEnd)
		pushRaw(RawTrimmed)
	}
}

func (l *Lexer) inlineRaw(innerEnd int, nodes *[]*SyntaxNode, prevStart *int) {
	pushRaw := func(kind SyntaxKind) {
		text := l.s.Get(*prevStart, l.s.Cursor())
		*nodes = append(*nodes, Leaf(kind, text))
		*prevStart = l.s.Cursor()
	}

	for l.s.Cursor() < innerEnd {
		if IsNewline(l.s.Peek()) {
			pushRaw(Text)
			l.s.EatNewline()
			pushRaw(RawTrimmed)
			continue
		}
		l.s.Eat()
	}
	pushRaw(Text)
}

func (l *Lexer) link() SyntaxKind {
	link, balanced := linkPrefix(l.s.After())
	l.s.Advance(len(link))

	if !balanced {
		return l.error("automatic links cannot contain unbalanced brackets, use the `link` function instead")
	}

	return Link
}

func (l *Lexer) numbering(start int) SyntaxKind {
	l.s.EatWhile(func(r rune) bool { return r >= '0' && r <= '9' })

	read := l.s.From(start)
	if l.s.EatIf('.') && l.spaceOrEnd() {
		if _, err := strconv.ParseUint(read, 10, 64); err == nil {
			return EnumMarker
		}
	}

	return l.text()
}

func (l *Lexer) refMarker() SyntaxKind {
	l.s.EatWhile(IsValidInLabelLiteral)

	// Don't include trailing characters likely to be part of text
	for {
		prev := l.s.Scout(-1)
		if prev == '.' || prev == ':' {
			l.s.Uneat()
		} else {
			break
		}
	}

	return RefMarker
}

func (l *Lexer) label() SyntaxKind {
	label := l.s.EatWhile(IsValidInLabelLiteral)
	if len(label) == 0 {
		return l.error("label cannot be empty")
	}

	if !l.s.EatIf('>') {
		return l.error("unclosed label")
	}

	return Label
}

func (l *Lexer) text() SyntaxKind {
	// Characters that can break text
	breakChars := map[rune]bool{
		' ': true, '\t': true, '\n': true, '\x0b': true, '\x0c': true, '\r': true,
		'\\': true, '/': true, '[': true, ']': true, '~': true, '-': true,
		'.': true, '\'': true, '"': true, '*': true, '_': true, ':': true,
		'h': true, '`': true, '$': true, '<': true, '>': true, '@': true, '#': true,
	}

	for {
		l.s.EatUntil(func(r rune) bool {
			if breakChars[r] {
				return true
			}
			return unicode.IsSpace(r)
		})

		// Check if we should continue with the same text node
		s := l.s.Clone()
		c := s.Eat()
		switch {
		case c == ' ' && unicode.IsLetter(s.Peek()) || unicode.IsDigit(s.Peek()):
			// Continue
		case c == '/' && !s.AtAny('/', '*'):
			// Continue
		case c == '-' && !s.AtAny('-', '?'):
			// Continue
		case c == '.' && !s.At(".."):
			// Continue
		case c == 'h' && !s.At("ttp://") && !s.At("ttps://"):
			// Continue
		case c == '@' && !s.AtRune(IsValidInLabelLiteral):
			// Continue
		default:
			return Text
		}
		l.s = s
	}
}

func (l *Lexer) inWord() bool {
	wordy := func(c rune) bool {
		if c == 0 {
			return false
		}
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
		script := GetScript(c)
		return script != ScriptHan && script != ScriptHiragana &&
			script != ScriptKatakana && script != ScriptHangul
	}
	prev := l.s.Scout(-2)
	next := l.s.Peek()
	return wordy(prev) && wordy(next)
}

func (l *Lexer) spaceOrEnd() bool {
	return l.s.Done() || unicode.IsSpace(l.s.Peek()) ||
		l.s.At("//") || l.s.At("/*")
}

// math handles tokens in math mode.
func (l *Lexer) math(start int, c rune) (SyntaxKind, *SyntaxNode) {
	var kind SyntaxKind

	switch c {
	case '\\':
		kind = l.backslash()
	case '"':
		kind = l.string()
	case '-':
		if l.s.EatIfStr(">>") {
			kind = MathShorthand
		} else if l.s.EatIf('>') {
			kind = MathShorthand
		} else if l.s.EatIfStr("->") {
			kind = MathShorthand
		} else {
			kind = MathShorthand
		}
	case ':':
		if l.s.EatIf('=') {
			kind = MathShorthand
		} else if l.s.EatIfStr(":=") {
			kind = MathShorthand
		} else {
			kind = Colon
		}
	case '!':
		if l.s.EatIf('=') {
			kind = MathShorthand
		} else {
			kind = Bang
		}
	case '.':
		if l.s.EatIfStr("..") {
			kind = MathShorthand
		} else {
			kind = Dot
		}
	case '<':
		switch {
		case l.s.EatIfStr("==>"):
			kind = MathShorthand
		case l.s.EatIfStr("-->"):
			kind = MathShorthand
		case l.s.EatIfStr("--"):
			kind = MathShorthand
		case l.s.EatIfStr("-<"):
			kind = MathShorthand
		case l.s.EatIfStr("->"):
			kind = MathShorthand
		case l.s.EatIfStr("<<"):
			kind = MathShorthand
		case l.s.EatIfStr("<-"):
			kind = MathShorthand
		case l.s.EatIfStr("=>"):
			kind = MathShorthand
		case l.s.EatIfStr("=="):
			kind = MathShorthand
		case l.s.EatIfStr("~~"):
			kind = MathShorthand
		case l.s.EatIf('='):
			kind = MathShorthand
		case l.s.EatIf('<'):
			kind = MathShorthand
		case l.s.EatIf('-'):
			kind = MathShorthand
		case l.s.EatIf('~'):
			kind = MathShorthand
		default:
			kind = Lt
		}
	case '>':
		switch {
		case l.s.EatIfStr("->"):
			kind = MathShorthand
		case l.s.EatIfStr(">>"):
			kind = MathShorthand
		case l.s.EatIf('='):
			kind = MathShorthand
		case l.s.EatIf('>'):
			kind = MathShorthand
		default:
			kind = Gt
		}
	case '=':
		switch {
		case l.s.EatIfStr("=>"):
			kind = MathShorthand
		case l.s.EatIf('>'):
			kind = MathShorthand
		case l.s.EatIf(':'):
			kind = MathShorthand
		default:
			kind = Eq
		}
	case '|':
		switch {
		case l.s.EatIfStr("->"):
			kind = MathShorthand
		case l.s.EatIfStr("=>"):
			kind = MathShorthand
		case l.s.EatIf('|'):
			kind = MathShorthand
		default:
			kind = l.mathText(start, c)
		}
	case '~':
		switch {
		case l.s.EatIfStr("~>"):
			kind = MathShorthand
		case l.s.EatIf('>'):
			kind = MathShorthand
		default:
			kind = MathShorthand
		}
	case '*':
		kind = MathShorthand
	case ',':
		kind = Comma
	case ';':
		kind = Semicolon
	case '#':
		kind = Hash
	case '_':
		kind = Underscore
	case '$':
		kind = Dollar
	case '/':
		kind = Slash
	case '^':
		kind = Hat
	case '&':
		kind = MathAlignPoint
	case '√', '∛', '∜':
		kind = Root
	case '\'':
		l.s.EatWhile(func(r rune) bool { return r == '\'' })
		kind = MathPrimes
	case '(':
		kind = LeftParen
	case ')':
		kind = RightParen
	case '[':
		if l.s.EatIf('|') {
			kind = LeftBrace
		} else {
			kind = LeftBracket
		}
	default:
		mathClass := DefaultMathClass(c)
		switch mathClass {
		case MathClassOpening:
			kind = LeftBrace
		case MathClassClosing:
			if l.s.Scout(-1) == '|' && c == ']' {
				kind = RightBrace
			} else {
				kind = RightBrace
			}
		default:
			// Check for identifiers
			if IsMathIDStart(c) && l.s.AtRune(IsMathIDContinue) {
				l.s.EatWhile(IsMathIDContinue)
				text := l.s.From(start)
				// Check if it's a single grapheme
				graphemes := countGraphemes(text)
				if graphemes == 1 {
					kind = MathText
				} else {
					return l.mathIdentOrField(start)
				}
			} else {
				kind = l.mathText(start, c)
			}
		}
	}

	return kind, nil
}

func (l *Lexer) mathIdentOrField(start int) (SyntaxKind, *SyntaxNode) {
	kind := MathIdent
	node := Leaf(kind, l.s.From(start))

	for {
		ident := l.maybeDotIdent()
		if ident == "" {
			break
		}
		kind = FieldAccess
		children := []*SyntaxNode{
			node,
			Leaf(Dot, "."),
			Leaf(Ident, ident),
		}
		node = Inner(kind, children)
	}

	return kind, node
}

func (l *Lexer) maybeDotIdent() string {
	if l.s.Scout(1) != 0 && IsMathIDStart(l.s.Scout(1)) && l.s.EatIf('.') {
		identStart := l.s.Cursor()
		l.s.Eat()
		l.s.EatWhile(IsMathIDContinue)
		return l.s.From(identStart)
	}
	return ""
}

func (l *Lexer) mathText(start int, c rune) SyntaxKind {
	// Keep numbers and grapheme clusters together
	if unicode.IsDigit(c) {
		l.s.EatWhile(unicode.IsDigit)
		s := l.s.Clone()
		if s.EatIf('.') && len(s.EatWhile(unicode.IsDigit)) > 0 {
			l.s = s
		}
	} else {
		// Eat the rest of the grapheme cluster
		// Simplified: just eat combining marks
		l.s.EatWhile(func(r rune) bool {
			return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r)
		})
	}
	return MathText
}

// MaybeMathNamedArg handles named arguments in math function calls.
func (l *Lexer) MaybeMathNamedArg(start int) *SyntaxNode {
	cursor := l.s.Cursor()
	l.s.Jump(start)

	if l.s.AtRune(IsIDStart) {
		l.s.Eat()
		l.s.EatWhile(IsIDContinue)
		// Check for colon but not := or ::=
		if l.s.At(":") && !l.s.At(":=") && !l.s.At("::=") {
			text := l.s.From(start)
			if text != "_" {
				return Leaf(Ident, text)
			}
			err := NewSyntaxError("expected identifier, found underscore")
			return ErrorNode(err, text)
		}
	}

	l.s.Jump(cursor)
	return nil
}

// MaybeMathSpreadArg handles spread arguments in math function calls.
func (l *Lexer) MaybeMathSpreadArg(start int) *SyntaxNode {
	cursor := l.s.Cursor()
	l.s.Jump(start)

	if l.s.EatIfStr("..") {
		// Only infer spread if not followed by problematic characters
		if !l.spaceOrEnd() && !l.s.AtAny('.', ',', ';', ')', '$') {
			return Leaf(Dots, l.s.From(start))
		}
	}

	l.s.Jump(cursor)
	return nil
}

// code handles tokens in code mode.
func (l *Lexer) code(start int, c rune) SyntaxKind {
	switch c {
	case '<':
		if l.s.AtRune(IsIDContinue) {
			return l.label()
		}
		if l.s.EatIf('=') {
			return LtEq
		}
		return Lt
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return l.number(start, c)
	case '.':
		if l.s.AtRune(func(r rune) bool { return r >= '0' && r <= '9' }) {
			return l.number(start, c)
		}
		if l.s.EatIf('.') {
			return Dots
		}
		return Dot
	case '"':
		return l.string()
	case '=':
		if l.s.EatIf('=') {
			return EqEq
		}
		if l.s.EatIf('>') {
			return Arrow
		}
		return Eq
	case '!':
		if l.s.EatIf('=') {
			return ExclEq
		}
		return Bang
	case '>':
		if l.s.EatIf('=') {
			return GtEq
		}
		return Gt
	case '+':
		if l.s.EatIf('=') {
			return PlusEq
		}
		return Plus
	case '-', '\u2212': // minus sign
		if l.s.EatIf('=') {
			return HyphEq
		}
		return Minus
	case '*':
		if l.s.EatIf('=') {
			return StarEq
		}
		return Star
	case '/':
		if l.s.EatIf('=') {
			return SlashEq
		}
		return Slash
	case '{':
		return LeftBrace
	case '}':
		return RightBrace
	case '[':
		return LeftBracket
	case ']':
		return RightBracket
	case '(':
		return LeftParen
	case ')':
		return RightParen
	case '$':
		return Dollar
	case ',':
		return Comma
	case ';':
		return Semicolon
	case ':':
		return Colon
	}

	if IsIDStart(c) {
		return l.ident(start)
	}

	return l.error(fmt.Sprintf("the character `%c` is not valid in code", c))
}

func (l *Lexer) ident(start int) SyntaxKind {
	l.s.EatWhile(IsIDContinue)
	ident := l.s.From(start)

	// Check if preceded by . or @ (but not ..)
	prev := l.s.Get(0, start)
	if !strings.HasSuffix(prev, ".") && !strings.HasSuffix(prev, "@") || strings.HasSuffix(prev, "..") {
		if kw := keyword(ident); kw != End {
			return kw
		}
	}

	if ident == "_" {
		return Underscore
	}
	return Ident
}

func (l *Lexer) number(start int, firstC rune) SyntaxKind {
	// Handle alternative integer bases
	base := 10
	if firstC == '0' {
		if l.s.EatIf('b') {
			base = 2
		} else if l.s.EatIf('o') {
			base = 8
		} else if l.s.EatIf('x') {
			base = 16
		}
	}

	// Read the initial digits
	// For hex, read all alphanumerics (Rust behavior) so that invalid digits
	// like 'z' in '0x123z' are included in the number, not treated as suffix.
	if base == 16 {
		l.s.EatWhile(func(r rune) bool {
			return (r >= '0' && r <= '9') ||
				(r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z')
		})
	} else {
		l.s.EatWhile(func(r rune) bool { return r >= '0' && r <= '9' })
	}

	// Read floating point digits and exponents
	isFloat := false
	if base == 10 {
		if firstC == '.' {
			isFloat = true
		} else if !l.s.At("..") && !l.s.AtRune(IsIDStart) {
			if l.s.Peek() == '.' && l.s.Scout(1) != '.' {
				s := l.s.Clone()
				s.Eat() // eat the dot
				if !s.AtRune(IsIDStart) {
					l.s.Eat()
					isFloat = true
					l.s.EatWhile(func(r rune) bool { return r >= '0' && r <= '9' })
				}
			}
		}

		// Read the exponent
		if !l.s.At("em") && (l.s.EatIf('e') || l.s.EatIf('E')) {
			isFloat = true
			l.s.EatIf('+')
			l.s.EatIf('-')
			l.s.EatWhile(func(r rune) bool { return r >= '0' && r <= '9' })
		}
	}

	number := l.s.From(start)
	suffix := l.s.EatWhile(func(r rune) bool {
		return (r >= '0' && r <= '9') ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			r == '%'
	})

	// Parse large integer literals as floats
	if base == 10 && !isFloat {
		_, err := strconv.ParseInt(number, 10, 64)
		if err != nil {
			if _, ferr := strconv.ParseFloat(number, 64); ferr == nil {
				isFloat = true
			}
		}
	}

	// Validate suffix
	validSuffix := suffix == "" ||
		suffix == "pt" || suffix == "mm" || suffix == "cm" ||
		suffix == "in" || suffix == "deg" || suffix == "rad" ||
		suffix == "em" || suffix == "fr" || suffix == "%"

	var suffixErr string
	if !validSuffix {
		suffixErr = fmt.Sprintf("invalid number suffix: %s", suffix)
	}

	// Validate number
	var numberErr string
	if isFloat {
		if _, err := strconv.ParseFloat(number, 64); err != nil {
			numberErr = fmt.Sprintf("invalid floating point number: %s", number)
		}
	} else if base != 10 {
		numPart := number
		if len(numPart) > 2 {
			numPart = numPart[2:] // Skip 0b/0o/0x
		}
		_, err := strconv.ParseInt(numPart, base, 64)
		if err != nil {
			baseName := map[int]string{2: "binary", 8: "octal", 16: "hexadecimal"}[base]
			numberErr = fmt.Sprintf("invalid %s number: %s", baseName, number)
		} else if suffix != "" {
			baseName := map[int]string{2: "binary", 8: "octal", 16: "hexadecimal"}[base]
			numberErr = fmt.Sprintf("%s numbers cannot have a suffix", baseName)
		}
	}

	// Return result
	if numberErr != "" && suffixErr != "" {
		kind := l.error(numberErr)
		l.hint(suffixErr)
		return kind
	}
	if numberErr != "" {
		return l.error(numberErr)
	}
	if suffixErr != "" {
		return l.error(suffixErr)
	}

	if isFloat {
		return Float
	}
	if suffix != "" {
		return Numeric
	}
	return Int
}

func (l *Lexer) string() SyntaxKind {
	escaped := false
	l.s.EatUntil(func(c rune) bool {
		stop := c == '"' && !escaped
		escaped = c == '\\' && !escaped
		return stop
	})

	if !l.s.EatIf('"') {
		return l.error("unclosed string")
	}

	return Str
}

// keyword tries to parse an identifier into a keyword.
func keyword(ident string) SyntaxKind {
	switch ident {
	case "none":
		return None
	case "auto":
		return Auto
	case "true", "false":
		return Bool
	case "not":
		return Not
	case "and":
		return And
	case "or":
		return Or
	case "let":
		return Let
	case "set":
		return Set
	case "show":
		return Show
	case "context":
		return Context
	case "if":
		return If
	case "else":
		return Else
	case "for":
		return For
	case "in":
		return In
	case "while":
		return While
	case "break":
		return Break
	case "continue":
		return Continue
	case "return":
		return Return
	case "import":
		return Import
	case "include":
		return Include
	case "as":
		return As
	}
	return End // Not a keyword
}

// Helper functions

func countNewlines(text string) int {
	count := 0
	s := NewScanner(text)
	for {
		c := s.Eat()
		if c == 0 {
			break
		}
		if IsNewline(c) {
			if c == '\r' {
				s.EatIf('\n')
			}
			count++
		}
	}
	return count
}

func splitNewlines(text string) []string {
	var lines []string
	s := NewScanner(text)
	start := 0
	end := 0

	for {
		c := s.Eat()
		if c == 0 {
			break
		}
		if IsNewline(c) {
			if c == '\r' {
				s.EatIf('\n')
			}
			lines = append(lines, text[start:end])
			start = s.Cursor()
		}
		end = s.Cursor()
	}

	lines = append(lines, text[start:])
	return lines
}

func countLeadingSpaces(line string) int {
	count := 0
	for _, r := range line {
		if unicode.IsSpace(r) {
			count++
		} else {
			break
		}
	}
	return count
}

func isAllWhitespace(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func countGraphemes(s string) int {
	// Simplified: count runes as a rough approximation
	return len([]rune(s))
}

func linkPrefix(text string) (string, bool) {
	s := NewScanner(text)
	var brackets []byte

	for !s.Done() {
		c := s.Peek()
		switch {
		case (c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			c == '!' || c == '#' || c == '$' || c == '%' || c == '&' || c == '*' || c == '+' ||
			c == ',' || c == '-' || c == '.' || c == '/' || c == ':' || c == ';' || c == '=' ||
			c == '?' || c == '@' || c == '_' || c == '~' || c == '\'':
			s.Eat()
		case c == '[':
			brackets = append(brackets, '[')
			s.Eat()
		case c == '(':
			brackets = append(brackets, '(')
			s.Eat()
		case c == ']':
			if len(brackets) > 0 && brackets[len(brackets)-1] == '[' {
				brackets = brackets[:len(brackets)-1]
				s.Eat()
			} else {
				goto done
			}
		case c == ')':
			if len(brackets) > 0 && brackets[len(brackets)-1] == '(' {
				brackets = brackets[:len(brackets)-1]
				s.Eat()
			} else {
				goto done
			}
		default:
			goto done
		}
	}

done:
	// Don't include trailing punctuation
	for {
		prev := s.Scout(-1)
		if prev == '!' || prev == ',' || prev == '.' || prev == ':' || prev == ';' || prev == '?' || prev == '\'' {
			s.Uneat()
		} else {
			break
		}
	}

	return s.Before(), len(brackets) == 0
}
