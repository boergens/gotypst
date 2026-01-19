// Package syntax provides the parser for Typst documents.
//
// This file implements a recursive descent parser with memoization for efficient
// parsing of Typst source code. The parser handles three syntax modes: Markup,
// Math, and Code.
package syntax

// MaxDepth is the maximum nesting depth for expressions.
const MaxDepth = 256

// Parse parses a source file as top-level markup.
func Parse(text string) *SyntaxNode {
	p := NewParser(text, 0, ModeMarkup)
	markupExprs(p, true, SyntaxSetOf(End))
	return p.finishInto(Markup)
}

// ParseCode parses top-level code.
func ParseCode(text string) *SyntaxNode {
	p := NewParser(text, 0, ModeCode)
	codeExprs(p, SyntaxSetOf(End))
	return p.finishInto(Code)
}

// ParseMath parses top-level math.
func ParseMath(text string) *SyntaxNode {
	p := NewParser(text, 0, ModeMath)
	mathExprs(p, SyntaxSetOf(End))
	return p.finishInto(Math)
}

// AtNewline describes how to proceed with parsing when at a newline.
type AtNewline int

const (
	// NLContinue continues at newlines.
	NLContinue AtNewline = iota
	// NLStop stops at any newline.
	NLStop
	// NLContextualContinue continues only if there is a continuation with else or dot.
	NLContextualContinue
	// NLStopParBreak stops only at a paragraph break, not normal newlines.
	NLStopParBreak
	// NLRequireColumn requires that the token's column be >= a column.
	// Values >= 1000000 indicate NLRequireColumn with the column being value - 1000000.
	NLRequireColumn
)

// requireColumn creates an AtNewline that requires the given column.
func requireColumn(col int) AtNewline {
	return AtNewline(1000000 + col)
}

// isRequireColumn returns true if this is a NLRequireColumn mode.
func (m AtNewline) isRequireColumn() bool {
	return m >= 1000000
}

// column returns the required column for NLRequireColumn modes.
func (m AtNewline) column() int {
	if m.isRequireColumn() {
		return int(m - 1000000)
	}
	return 0
}

// stopAt returns true if parsing should stop at the given newline.
func (m AtNewline) stopAt(newline *Newline, kind SyntaxKind) bool {
	if newline == nil {
		return false
	}
	switch m {
	case NLContinue:
		return false
	case NLStop:
		return true
	case NLContextualContinue:
		return kind != Else && kind != Dot
	case NLStopParBreak:
		return newline.parbreak
	default:
		if m.isRequireColumn() {
			minCol := m.column()
			if newline.column < 0 {
				return false // Column not tracked, continue parsing
			}
			return newline.column <= minCol
		}
		return false
	}
}

// Newline holds information about newlines in a group of trivia.
type Newline struct {
	// column is the column of the start of the next token in its line.
	// -1 means the column is not tracked.
	column int
	// parbreak is true if any newlines were paragraph breaks.
	parbreak bool
}

// Token represents a single token returned from the lexer with cached info.
type Token struct {
	// kind is the SyntaxKind of the current token.
	kind SyntaxKind
	// node is the SyntaxNode of the current token, ready to be eaten.
	node *SyntaxNode
	// nTrivia is the number of preceding trivia before this token.
	nTrivia int
	// newline holds info about newlines in preceding trivia, or nil if none.
	newline *Newline
	// start is the index into text of the token's start.
	start int
	// prevEnd is the index into text of the previous token's end.
	prevEnd int
}

// Marker represents a node's position in the parser. Used for wrapping.
type Marker int

// MemoKey is the type for memoization keys (text index).
type MemoKey int

// PartialState holds state needed to restore the parser's current token and lexer.
type PartialState struct {
	cursor  int
	lexMode SyntaxMode
	token   Token
}

// Checkpoint holds a full checkpoint of the parser state.
type Checkpoint struct {
	nodeLen int
	state   PartialState
}

// MemoArena provides efficient parser backtracking similar to packrat parsing.
type MemoArena struct {
	// arena holds previously parsed nodes (to reduce allocations).
	arena []*SyntaxNode
	// memoMap maps text index to parsed nodes and parser state.
	memoMap map[MemoKey]struct {
		nodes []*SyntaxNode
		state PartialState
	}
}

// Parser manages parsing a stream of tokens into a tree of SyntaxNodes.
type Parser struct {
	// text is the source text shared with the lexer.
	text string
	// lexer is the lexer over the source text.
	lexer *Lexer
	// nlMode is the newline mode: whether to insert a temporary end at newlines.
	nlMode AtNewline
	// token is the current token under inspection, not yet in nodes.
	token Token
	// balanced tracks if the parser has matched delimiters.
	balanced bool
	// nodes holds the concrete syntax tree of previously parsed text.
	nodes []*SyntaxNode
	// memo is used for efficient parser backtracking.
	memo *MemoArena
	// depth is the current expression nesting depth.
	depth int
}

// NewParser creates a new parser starting from the given text offset and mode.
func NewParser(text string, offset int, mode SyntaxMode) *Parser {
	lexer := NewLexer(text, mode)
	lexer.Jump(offset)
	nlMode := NLContinue
	nodes := make([]*SyntaxNode, 0, 64)
	token := lex(&nodes, lexer, nlMode)
	return &Parser{
		text:     text,
		lexer:    lexer,
		nlMode:   nlMode,
		token:    token,
		balanced: true,
		nodes:    nodes,
		memo: &MemoArena{
			memoMap: make(map[MemoKey]struct {
				nodes []*SyntaxNode
				state PartialState
			}),
		},
		depth: 0,
	}
}

// finish consumes the parser and returns the parsed nodes.
func (p *Parser) finish() []*SyntaxNode {
	return p.nodes
}

// finishInto consumes the parser and returns a single top-level node.
func (p *Parser) finishInto(kind SyntaxKind) *SyntaxNode {
	return Inner(kind, p.finish())
}

// current returns the kind of the next token to be eaten.
func (p *Parser) current() SyntaxKind {
	return p.token.kind
}

// at returns true if the current token is the given kind.
func (p *Parser) at(kind SyntaxKind) bool {
	return p.token.kind == kind
}

// atSet returns true if the current token is in the given set.
func (p *Parser) atSet(set SyntaxSet) bool {
	return set.Contains(p.token.kind)
}

// end returns true if at the end of the token stream.
func (p *Parser) end() bool {
	return p.at(End)
}

// directlyAt returns true if at the given kind with no preceding trivia.
func (p *Parser) directlyAt(kind SyntaxKind) bool {
	return p.token.kind == kind && !p.hadTrivia()
}

// hadTrivia returns true if the current token had preceding trivia.
func (p *Parser) hadTrivia() bool {
	return p.token.nTrivia > 0
}

// hadNewline returns true if the current token had a newline in its trivia.
func (p *Parser) hadNewline() bool {
	return p.token.newline != nil
}

// currentColumn returns the column of the current token's start.
func (p *Parser) currentColumn() int {
	if p.token.newline != nil && p.token.newline.column >= 0 {
		return p.token.newline.column
	}
	return p.lexer.Column(p.token.start)
}

// currentText returns the current token's text.
func (p *Parser) currentText() string {
	return p.text[p.token.start:p.currentEnd()]
}

// currentStart returns the offset of the current token's start.
func (p *Parser) currentStart() int {
	return p.token.start
}

// currentEnd returns the offset of the current token's end.
func (p *Parser) currentEnd() int {
	return p.lexer.Cursor()
}

// prevEnd returns the offset of the previous token's end.
func (p *Parser) prevEnd() int {
	return p.token.prevEnd
}

// marker returns a marker pointing to the current position.
func (p *Parser) marker() Marker {
	return Marker(len(p.nodes))
}

// beforeTrivia returns a marker pointing to the first trivia before the current token.
func (p *Parser) beforeTrivia() Marker {
	return Marker(len(p.nodes) - p.token.nTrivia)
}

// nodeAt returns the node at the given marker.
func (p *Parser) nodeAt(m Marker) *SyntaxNode {
	return p.nodes[m]
}

// nodeAtMut returns the node at the given marker for mutation.
func (p *Parser) nodeAtMut(m Marker) *SyntaxNode {
	return p.nodes[m]
}

// eatAndGet eats the current node and returns a reference for mutation.
func (p *Parser) eatAndGet() *SyntaxNode {
	offset := len(p.nodes)
	p.eat()
	return p.nodes[offset]
}

// eatIf eats the token if at the given kind. Returns true if eaten.
func (p *Parser) eatIf(kind SyntaxKind) bool {
	if p.at(kind) {
		p.eat()
		return true
	}
	return false
}

// assert asserts that we are at the given kind and eats it.
func (p *Parser) assert(kind SyntaxKind) {
	if p.token.kind != kind {
		panic("parser assertion failed: expected " + kind.Name())
	}
	p.eat()
}

// convertAndEat converts the current token's kind and eats it.
func (p *Parser) convertAndEat(kind SyntaxKind) {
	p.token.node.ConvertToKind(kind)
	p.eat()
}

// eat eats the current token by saving it to nodes and moving to the next.
func (p *Parser) eat() {
	p.nodes = append(p.nodes, p.token.node)
	p.token = lex(&p.nodes, p.lexer, p.nlMode)
}

// flushTrivia detaches parsed trivia nodes from this token.
func (p *Parser) flushTrivia() {
	p.token.nTrivia = 0
	p.token.prevEnd = p.token.start
}

// wrap wraps nodes from a marker to before the current token in a new node.
func (p *Parser) wrap(from Marker, kind SyntaxKind) {
	to := int(p.beforeTrivia())
	fromIdx := int(from)
	if fromIdx > to {
		fromIdx = to
	}

	// Get children to wrap
	children := make([]*SyntaxNode, to-fromIdx)
	copy(children, p.nodes[fromIdx:to])

	// Get trailing nodes (trivia)
	trailing := make([]*SyntaxNode, len(p.nodes)-to)
	copy(trailing, p.nodes[to:])

	// Reconstruct: [nodes before from] + [wrapped node] + [trailing]
	p.nodes = p.nodes[:fromIdx]
	p.nodes = append(p.nodes, Inner(kind, children))
	p.nodes = append(p.nodes, trailing...)
}

// wrapError wraps nodes from a marker in an error node.
func (p *Parser) wrapError(from Marker, message string) {
	to := int(p.beforeTrivia())
	fromIdx := int(from)
	if fromIdx > to {
		fromIdx = to
	}
	var text string
	for i := fromIdx; i < to; i++ {
		text += p.nodes[i].IntoText()
	}
	errNode := ErrorNode(NewSyntaxError(message), text)
	newNodes := make([]*SyntaxNode, fromIdx+1+len(p.nodes)-to)
	copy(newNodes[:fromIdx], p.nodes[:fromIdx])
	newNodes[fromIdx] = errNode
	copy(newNodes[fromIdx+1:], p.nodes[to:])
	p.nodes = newNodes
}

// enterModes parses within the given syntax mode for subsequent tokens.
func (p *Parser) enterModes(mode SyntaxMode, stop AtNewline, f func(*Parser)) {
	previous := p.lexer.Mode()
	p.lexer.SetMode(mode)
	p.withNLMode(stop, f)
	if mode != previous {
		p.lexer.SetMode(previous)
		p.lexer.Jump(p.token.prevEnd)
		p.nodes = p.nodes[:len(p.nodes)-p.token.nTrivia]
		p.token = lex(&p.nodes, p.lexer, p.nlMode)
	}
}

// withNLMode parses within the given newline mode.
func (p *Parser) withNLMode(mode AtNewline, f func(*Parser)) {
	previous := p.nlMode
	p.nlMode = mode
	f(p)
	p.nlMode = previous
	if p.token.newline != nil && mode != previous {
		actualKind := p.token.node.Kind()
		if p.nlMode.stopAt(p.token.newline, actualKind) {
			p.token.kind = End
		} else {
			p.token.kind = actualKind
		}
	}
}

// lex moves the lexer forward and prepares the current token.
func lex(nodes *[]*SyntaxNode, lexer *Lexer, nlMode AtNewline) Token {
	prevEnd := lexer.Cursor()
	start := prevEnd
	kind, node := lexer.Next()
	nTrivia := 0
	hadNewline := false
	parbreak := false

	for kind.IsTrivia() {
		hadNewline = hadNewline || lexer.Newline()
		parbreak = parbreak || kind == Parbreak
		nTrivia++
		*nodes = append(*nodes, node)
		start = lexer.Cursor()
		kind, node = lexer.Next()
	}

	var newline *Newline
	if hadNewline {
		col := -1
		if lexer.Mode() == ModeMarkup {
			col = lexer.Column(start)
		}
		newline = &Newline{column: col, parbreak: parbreak}
		if nlMode.stopAt(newline, kind) {
			kind = End
		}
	}

	return Token{
		kind:    kind,
		node:    node,
		nTrivia: nTrivia,
		newline: newline,
		start:   start,
		prevEnd: prevEnd,
	}
}

// Memoization methods

// memoizeParsedNodes stores parsed nodes in the memo map.
func (p *Parser) memoizeParsedNodes(key MemoKey, prevLen int) {
	checkpoint := p.checkpoint()
	nodeLen := checkpoint.nodeLen
	memoNodes := make([]*SyntaxNode, nodeLen-prevLen)
	copy(memoNodes, p.nodes[prevLen:nodeLen])
	p.memo.memoMap[key] = struct {
		nodes []*SyntaxNode
		state PartialState
	}{
		nodes: memoNodes,
		state: checkpoint.state,
	}
}

// restoreMemoOrCheckpoint tries to load a memoized result.
// Returns (key, checkpoint, true) if no memo found, (_, _, false) if restored from memo.
func (p *Parser) restoreMemoOrCheckpoint() (MemoKey, Checkpoint, bool) {
	key := MemoKey(p.currentStart())
	if memo, ok := p.memo.memoMap[key]; ok {
		p.nodes = append(p.nodes, memo.nodes...)
		p.restorePartial(memo.state)
		return 0, Checkpoint{}, false
	}
	return key, p.checkpoint(), true
}

// restore restores the parser to a checkpoint.
func (p *Parser) restore(checkpoint Checkpoint) {
	p.nodes = p.nodes[:checkpoint.nodeLen]
	p.restorePartial(checkpoint.state)
}

// restorePartial restores the token and lexer state.
func (p *Parser) restorePartial(state PartialState) {
	p.lexer.Jump(state.cursor)
	p.lexer.SetMode(state.lexMode)
	p.token = state.token
}

// checkpoint saves a checkpoint of the parser state.
func (p *Parser) checkpoint() Checkpoint {
	return Checkpoint{
		nodeLen: len(p.nodes),
		state: PartialState{
			cursor:  p.lexer.Cursor(),
			lexMode: p.lexer.Mode(),
			token:   p.token,
		},
	}
}

// Error handling methods

// expect consumes the given kind or produces an error.
func (p *Parser) expect(kind SyntaxKind) bool {
	if p.at(kind) {
		p.eat()
		return true
	}
	if kind == Ident && p.token.kind.IsKeyword() {
		p.trimErrors()
		p.eatAndGet().Expected(kind.Name())
	} else {
		p.balanced = p.balanced && !kind.IsGrouping()
		p.expected(kind.Name())
	}
	return false
}

// expectClosingDelimiter consumes the closing delimiter or marks the opener as error.
func (p *Parser) expectClosingDelimiter(open Marker, kind SyntaxKind) {
	if !p.eatIf(kind) {
		p.nodes[open].ConvertToError("unclosed delimiter")
	}
}

// expected produces an error that the given thing was expected.
func (p *Parser) expected(thing string) {
	if !p.afterError() {
		p.expectedAt(p.beforeTrivia(), thing)
	}
}

// afterError returns true if the last non-trivia node is an error.
func (p *Parser) afterError() bool {
	m := p.beforeTrivia()
	return int(m) > 0 && p.nodes[m-1].Kind().IsError()
}

// expectedAt produces an error at the given marker position.
func (p *Parser) expectedAt(m Marker, thing string) {
	errNode := ErrorNode(NewSyntaxError("expected "+thing), "")
	// Insert error node at position m
	p.nodes = append(p.nodes[:m], append([]*SyntaxNode{errNode}, p.nodes[m:]...)...)
}

// hint adds a hint to a trailing error.
func (p *Parser) hint(h string) {
	m := p.beforeTrivia()
	if int(m) > 0 {
		p.nodes[m-1].Hint(h)
	}
}

// unexpected consumes the next token and produces an unexpected error.
func (p *Parser) unexpected() {
	p.trimErrors()
	p.balanced = p.balanced && !p.token.kind.IsGrouping()
	p.eatAndGet().Unexpected()
}

// trimErrors removes trailing zero-length error nodes.
func (p *Parser) trimErrors() {
	end := int(p.beforeTrivia())
	start := end
	for start > 0 && p.nodes[start-1].Kind().IsError() && p.nodes[start-1].IsEmpty() {
		start--
	}
	if start < end {
		p.nodes = append(p.nodes[:start], p.nodes[end:]...)
	}
}

// Depth checking methods

// checkDepthUntil checks if max depth has been exceeded.
// Returns the parser if ok, nil if depth exceeded.
func (p *Parser) checkDepthUntil(stopSet SyntaxSet) *Parser {
	if p.depth < MaxDepth {
		return p
	}
	p.depthCheckError(&stopSet)
	return nil
}

// increaseDepth increases the depth and returns a cleanup function.
// Returns nil if depth exceeded.
func (p *Parser) increaseDepth() func() {
	if p.depth < MaxDepth {
		p.depth++
		return func() { p.depth-- }
	}
	p.depthCheckError(nil)
	return nil
}

// depthCheckError generates an error for exceeded depth.
func (p *Parser) depthCheckError(stopSet *SyntaxSet) {
	m := p.marker()

	balance := 0
	savedNLMode := p.nlMode
	p.nlMode = NLContinue
	for {
		if p.atSet(SyntaxSetOf(LeftBracket, LeftBrace, LeftParen)) {
			balance++
		} else if p.atSet(SyntaxSetOf(RightBracket, RightBrace, RightParen)) {
			balance--
			if balance < 0 {
				balance = 0
			}
		}
		p.eat()

		atStop := stopSet == nil || p.atSet(*stopSet)
		if (balance == 0 && atStop) || p.end() {
			break
		}
	}
	p.nlMode = savedNLMode

	p.wrapError(m, "maximum parsing depth exceeded")
}
