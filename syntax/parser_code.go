package syntax

// code parses the contents of a code block.
func code(p *Parser, stopSet SyntaxSet) {
	m := p.marker()
	codeExprs(p, stopSet)
	p.wrap(m, Code)
}

// codeExprs parses a sequence of code expressions.
func codeExprs(p *Parser, stopSet SyntaxSet) {
	if !stopSet.Contains(End) {
		stopSet = stopSet.Add(End)
	}
	if p.depth >= MaxDepth {
		p.depthCheckError(&stopSet)
		return
	}

	for !p.atSet(stopSet) {
		p.withNLMode(NLContextualContinue, func(p *Parser) {
			if !p.atSet(CodeExprSet) {
				p.unexpected()
				return
			}
			codeExpr(p)
			if !p.atSet(stopSet) && !p.eatIf(Semicolon) {
				p.expected("semicolon or line break")
				if p.at(Label) {
					p.hint("labels can only be applied in markup mode")
					p.hint("try wrapping your code in a markup block (`[ ]`)")
				}
			}
		})
	}
}

// embeddedCodeExpr parses an atomic code expression embedded in markup or math.
func embeddedCodeExpr(p *Parser) {
	p.enterModes(ModeCode, NLStop, func(p *Parser) {
		p.assert(Hash)
		if p.hadTrivia() || p.end() {
			p.expected("expression")
			return
		}

		stmt := p.atSet(StmtSet)
		at := p.atSet(AtomicCodeExprSet)
		codeExprPrec(p, true, 0)

		// Consume error for things like `#12p` or `#"abc\"`.
		if !at {
			p.unexpected()
		}

		semi := (stmt || p.directlyAt(Semicolon)) && p.eatIf(Semicolon)

		if stmt && !semi && !p.end() && !p.at(RightBracket) {
			p.expected("semicolon or line break")
		}
	})
}

// codeExpr parses a single code expression.
func codeExpr(p *Parser) {
	codeExprPrec(p, false, 0)
}

// codeExprPrec parses a code expression with at least the given precedence.
func codeExprPrec(p *Parser, atomic bool, minPrec int) {
	cleanup := p.increaseDepth()
	if cleanup == nil {
		return
	}
	defer cleanup()

	m := p.marker()
	if !atomic && p.atSet(UnaryOpSet) {
		op := UnOpFromSyntaxKind(p.current())
		if op != nil {
			p.eat()
			codeExprPrec(p, atomic, op.Precedence())
			p.wrap(m, Unary)
		} else {
			codePrimary(p, atomic)
		}
	} else {
		codePrimary(p, atomic)
	}

	for {
		if p.directlyAt(LeftParen) || p.directlyAt(LeftBracket) {
			args(p)
			p.wrap(m, FuncCall)
			continue
		}

		atFieldOrMethod := p.directlyAt(Dot) && p.peekIsIdent()

		if atomic && !atFieldOrMethod {
			break
		}

		if p.eatIf(Dot) {
			p.expect(Ident)
			p.wrap(m, FieldAccess)
			continue
		}

		var binop *BinOp
		if p.atSet(BinaryOpSet) {
			op := BinOpFromSyntaxKind(p.current())
			if op >= 0 {
				binop = &op
			}
		} else if minPrec <= BinOpNotIn.Precedence() && p.eatIf(Not) {
			if p.at(In) {
				op := BinOpNotIn
				binop = &op
			} else {
				p.expected("keyword `in`")
				break
			}
		}

		if binop != nil {
			prec := binop.Precedence()
			if prec < minPrec {
				break
			}

			if binop.Assoc() == AssocLeft {
				prec++
			}

			p.eat()
			codeExprPrec(p, false, prec)
			p.wrap(m, Binary)
			continue
		}

		break
	}
}

// peekIsIdent checks if the next token after the current is an identifier.
func (p *Parser) peekIsIdent() bool {
	// Clone the lexer to peek
	savedCursor := p.lexer.Cursor()
	savedMode := p.lexer.Mode()
	kind, _ := p.lexer.Next()
	p.lexer.Jump(savedCursor)
	p.lexer.SetMode(savedMode)
	return kind == Ident
}

// codePrimary parses a primary expression in code.
func codePrimary(p *Parser, atomic bool) {
	m := p.marker()
	switch p.current() {
	case Ident:
		p.eat()
		if !atomic && p.at(Arrow) {
			p.wrap(m, Params)
			p.assert(Arrow)
			codeExpr(p)
			p.wrap(m, Closure)
		}

	case Underscore:
		if !atomic {
			p.eat()
			if p.at(Arrow) {
				p.wrap(m, Params)
				p.eat()
				codeExpr(p)
				p.wrap(m, Closure)
			} else if p.eatIf(Eq) {
				codeExpr(p)
				p.wrap(m, DestructAssignment)
			} else {
				p.nodes[m].Expected("expression")
			}
		} else {
			p.expected("expression")
		}

	case LeftBrace:
		codeBlock(p)
	case LeftBracket:
		contentBlock(p)
	case LeftParen:
		exprWithParen(p, atomic)
	case Dollar:
		equation(p)
	case Let:
		letBinding(p)
	case Set:
		setRule(p)
	case Show:
		showRule(p)
	case Context:
		contextual(p, atomic)
	case If:
		conditional(p)
	case While:
		whileLoop(p)
	case For:
		forLoop(p)
	case Import:
		moduleImport(p)
	case Include:
		moduleInclude(p)
	case Break:
		breakStmt(p)
	case Continue:
		continueStmt(p)
	case Return:
		returnStmt(p)

	case Raw:
		p.eat() // Raw is handled entirely in the Lexer.

	case None, Auto, Int, Float, Bool, Numeric, Str, Label:
		p.eat()

	default:
		p.expected("expression")
	}
}

// block parses a content or code block.
func block(p *Parser) {
	switch p.current() {
	case LeftBracket:
		contentBlock(p)
	case LeftBrace:
		codeBlock(p)
	default:
		p.expected("block")
	}
}

// codeBlock parses a code block: `{ let x = 1; x + 2 }`.
func codeBlock(p *Parser) {
	m := p.marker()
	p.enterModes(ModeCode, NLContinue, func(p *Parser) {
		p.assert(LeftBrace)
		code(p, SyntaxSetOf(RightBrace, RightBracket, RightParen, End))
		p.expectClosingDelimiter(m, RightBrace)
	})
	p.wrap(m, CodeBlock)
}

// contentBlock parses a content block: `[*Hi* there!]`.
func contentBlock(p *Parser) {
	m := p.marker()
	p.enterModes(ModeMarkup, NLContinue, func(p *Parser) {
		p.assert(LeftBracket)
		markup(p, true, true, SyntaxSetOf(RightBracket, End))
		p.expectClosingDelimiter(m, RightBracket)
	})
	p.wrap(m, ContentBlock)
}

// letBinding parses a let binding: `let x = 1`.
func letBinding(p *Parser) {
	m := p.marker()
	p.assert(Let)

	m2 := p.marker()
	closure := false
	other := false

	if p.eatIf(Ident) {
		if p.directlyAt(LeftParen) {
			params(p)
			closure = true
		}
	} else {
		pattern(p, false, make(map[string]bool), "")
		other = true
	}

	// The condition for parsing an initializer expression:
	// - If closure or other (destructuring pattern), we REQUIRE an '='
	// - Otherwise (simple identifier), we OPTIONALLY consume '='
	// Only parse the expression if '=' was found
	var hasEq bool
	if closure || other {
		hasEq = p.expect(Eq)
	} else {
		hasEq = p.eatIf(Eq)
	}

	if hasEq {
		codeExpr(p)
	}

	if closure {
		p.wrap(m2, Closure)
	}

	p.wrap(m, LetBinding)
}

// setRule parses a set rule: `set text(...)`.
func setRule(p *Parser) {
	m := p.marker()
	p.assert(Set)

	m2 := p.marker()
	p.expect(Ident)
	for p.eatIf(Dot) {
		p.expect(Ident)
		p.wrap(m2, FieldAccess)
	}

	args(p)
	if p.eatIf(If) {
		codeExpr(p)
	}
	p.wrap(m, SetRule)
}

// showRule parses a show rule: `show heading: it => emph(it.body)`.
func showRule(p *Parser) {
	m := p.marker()
	p.assert(Show)
	m2 := p.beforeTrivia()

	if !p.at(Colon) {
		codeExpr(p)
	}

	if p.eatIf(Colon) {
		codeExpr(p)
	} else {
		p.expectedAt(m2, "colon")
	}

	p.wrap(m, ShowRule)
}

// contextual parses a contextual expression: `context text.lang`.
func contextual(p *Parser, atomic bool) {
	m := p.marker()
	p.assert(Context)
	codeExprPrec(p, atomic, 0)
	p.wrap(m, Contextual)
}

// conditional parses an if-else conditional: `if x { y } else { z }`.
func conditional(p *Parser) {
	m := p.marker()
	p.assert(If)
	codeExpr(p)
	block(p)
	if p.eatIf(Else) {
		if p.at(If) {
			conditional(p)
		} else {
			block(p)
		}
	}
	p.wrap(m, Conditional)
}

// whileLoop parses a while loop: `while x { y }`.
func whileLoop(p *Parser) {
	m := p.marker()
	p.assert(While)
	codeExpr(p)
	block(p)
	p.wrap(m, WhileLoop)
}

// forLoop parses a for loop: `for x in y { z }`.
func forLoop(p *Parser) {
	m := p.marker()
	p.assert(For)

	seen := make(map[string]bool)
	pattern(p, false, seen, "")

	if p.at(Comma) {
		node := p.eatAndGet()
		node.Unexpected()
		node.Hint("destructuring patterns must be wrapped in parentheses")
		if p.atSet(PatternSet) {
			pattern(p, false, seen, "")
		}
	}

	p.expect(In)
	codeExpr(p)
	block(p)
	p.wrap(m, ForLoop)
}

// moduleImport parses a module import: `import "utils.typ": a, b, c`.
func moduleImport(p *Parser) {
	m := p.marker()
	p.assert(Import)
	codeExpr(p)
	if p.eatIf(As) {
		p.expect(Ident)
	}

	if p.eatIf(Colon) {
		if p.at(LeftParen) {
			p.withNLMode(NLContinue, func(p *Parser) {
				m2 := p.marker()
				p.assert(LeftParen)
				importItems(p)
				p.expectClosingDelimiter(m2, RightParen)
			})
		} else if !p.eatIf(Star) {
			importItems(p)
		}
	}

	p.wrap(m, ModuleImport)
}

// importItems parses items to import from a module.
func importItems(p *Parser) {
	m := p.marker()
	for !p.current().IsTerminator() {
		itemMarker := p.marker()
		if !p.eatIf(Ident) {
			p.unexpected()
		}

		// Nested import path: `a.b.c`
		for p.eatIf(Dot) {
			p.expect(Ident)
		}

		p.wrap(itemMarker, ImportItemPath)

		// Rename imported item.
		if p.eatIf(As) {
			p.expect(Ident)
			p.wrap(itemMarker, RenamedImportItem)
		}

		if !p.current().IsTerminator() {
			p.expect(Comma)
		}
	}

	p.wrap(m, ImportItems)
}

// moduleInclude parses a module include: `include "chapter1.typ"`.
func moduleInclude(p *Parser) {
	m := p.marker()
	p.assert(Include)
	codeExpr(p)
	p.wrap(m, ModuleInclude)
}

// breakStmt parses a break from a loop: `break`.
func breakStmt(p *Parser) {
	m := p.marker()
	p.assert(Break)
	p.wrap(m, LoopBreak)
}

// continueStmt parses a continue in a loop: `continue`.
func continueStmt(p *Parser) {
	m := p.marker()
	p.assert(Continue)
	p.wrap(m, LoopContinue)
}

// returnStmt parses a return from a function: `return`, `return x + 1`.
func returnStmt(p *Parser) {
	m := p.marker()
	p.assert(Return)
	if p.atSet(CodeExprSet) {
		codeExpr(p)
	}
	p.wrap(m, FuncReturn)
}

// exprWithParen parses an expression starting with a parenthesis.
func exprWithParen(p *Parser, atomic bool) {
	if atomic {
		parenthesizedOrArrayOrDict(p)
		return
	}

	// Try to restore from memo or create a checkpoint.
	key, checkpoint, needsParse := p.restoreMemoOrCheckpoint()
	if !needsParse {
		return
	}
	prevLen := checkpoint.nodeLen

	// First attempt: parse as parenthesized, array, or dict.
	kind := parenthesizedOrArrayOrDict(p)

	// Check if we need to backtrack.
	if p.at(Arrow) {
		p.restore(checkpoint)
		m := p.marker()
		params(p)
		if !p.expect(Arrow) {
			return
		}
		codeExpr(p)
		p.wrap(m, Closure)
	} else if p.at(Eq) && kind != Parenthesized {
		p.restore(checkpoint)
		m := p.marker()
		destructuringOrParenthesized(p, true, make(map[string]bool))
		if !p.expect(Eq) {
			return
		}
		codeExpr(p)
		p.wrap(m, DestructAssignment)
	} else {
		return
	}

	// Memoize result if we backtracked.
	p.memoizeParsedNodes(key, prevLen)
}

// GroupState holds state for array/dictionary parsing.
type GroupState struct {
	count          int
	maybeJustParens bool
	kind           *SyntaxKind
	seen           map[string]bool
}

// parenthesizedOrArrayOrDict parses parenthesized expr, array, or dict.
func parenthesizedOrArrayOrDict(p *Parser) SyntaxKind {
	state := GroupState{
		count:          0,
		maybeJustParens: true,
		kind:           nil,
		seen:           make(map[string]bool),
	}

	m := p.marker()
	p.withNLMode(NLContinue, func(p *Parser) {
		p.assert(LeftParen)
		if p.eatIf(Colon) {
			k := Dict
			state.kind = &k
		}

		for !p.current().IsTerminator() {
			if !p.atSet(ArrayOrDictItemSet) {
				p.unexpected()
				continue
			}

			arrayOrDictItem(p, &state)
			state.count++

			if !p.current().IsTerminator() && p.expect(Comma) {
				state.maybeJustParens = false
			}
		}

		p.expectClosingDelimiter(m, RightParen)
	})

	var kind SyntaxKind
	if state.maybeJustParens && state.count == 1 {
		kind = Parenthesized
	} else if state.kind != nil {
		kind = *state.kind
	} else {
		kind = Array
	}

	p.wrap(m, kind)
	return kind
}

// arrayOrDictItem parses a single item in an array or dictionary.
func arrayOrDictItem(p *Parser, state *GroupState) {
	m := p.marker()

	if p.eatIf(Dots) {
		// Spread item: `..item`.
		codeExpr(p)
		p.wrap(m, Spread)
		state.maybeJustParens = false
		return
	}

	codeExpr(p)

	if p.eatIf(Colon) {
		// Named/keyed pair: `name: item` or `"key": item`.
		codeExpr(p)

		node := p.nodes[m]
		var pairKind SyntaxKind
		if node.Kind() == Ident {
			pairKind = Named
		} else {
			pairKind = Keyed
		}

		// Check for duplicate keys.
		var key string
		if node.Kind() == Ident || node.Kind() == Str {
			key = node.Text()
			if state.seen[key] {
				node.ConvertToError("duplicate key: " + key)
			}
			state.seen[key] = true
		}

		p.wrap(m, pairKind)
		state.maybeJustParens = false

		if state.kind != nil && *state.kind == Array {
			p.nodes[m].Expected("expression")
		} else {
			k := Dict
			state.kind = &k
		}
	} else {
		// Positional item.
		if state.kind != nil && *state.kind == Dict {
			p.nodes[m].Expected("named or keyed pair")
		} else {
			k := Array
			state.kind = &k
		}
	}
}

// args parses a function call's argument list.
func args(p *Parser) {
	if !p.directlyAt(LeftParen) && !p.directlyAt(LeftBracket) {
		p.expected("argument list")
		if p.at(LeftParen) || p.at(LeftBracket) {
			p.hint("there may not be any spaces before the argument list")
		}
	}

	m := p.marker()
	if p.at(LeftParen) {
		m2 := p.marker()
		p.withNLMode(NLContinue, func(p *Parser) {
			p.assert(LeftParen)

			seen := make(map[string]bool)
			for !p.current().IsTerminator() {
				if !p.atSet(ArgSet) {
					p.unexpected()
					continue
				}

				arg(p, seen)

				if !p.current().IsTerminator() {
					p.expect(Comma)
				}
			}

			p.expectClosingDelimiter(m2, RightParen)
		})
	}

	for p.directlyAt(LeftBracket) {
		contentBlock(p)
	}

	p.wrap(m, Args)
}

// arg parses a single argument in an argument list.
func arg(p *Parser, seen map[string]bool) {
	m := p.marker()

	// Spread argument: `..args`.
	if p.eatIf(Dots) {
		codeExpr(p)
		p.wrap(m, Spread)
		return
	}

	// Normal positional argument or argument name.
	wasAtExpr := p.atSet(CodeExprSet)
	text := p.currentText()
	codeExpr(p)

	// Named argument: `thickness: 12pt`.
	if p.eatIf(Colon) {
		// Recover from bad argument name.
		if wasAtExpr {
			if p.nodes[m].Kind() != Ident {
				p.nodes[m].Expected("identifier")
			} else if seen[text] {
				p.nodes[m].ConvertToError("duplicate argument: " + text)
			}
			seen[text] = true
		}

		codeExpr(p)
		p.wrap(m, Named)
	}
}

// params parses a closure's parameters: `(x, y)`.
func params(p *Parser) {
	m := p.marker()
	p.withNLMode(NLContinue, func(p *Parser) {
		p.assert(LeftParen)

		seen := make(map[string]bool)
		sink := false

		for !p.current().IsTerminator() {
			if !p.atSet(ParamSet) {
				p.unexpected()
				continue
			}

			param(p, seen, &sink)

			if !p.current().IsTerminator() {
				p.expect(Comma)
			}
		}

		p.expectClosingDelimiter(m, RightParen)
	})
	p.wrap(m, Params)
}

// param parses a single parameter in a parameter list.
func param(p *Parser, seen map[string]bool, sink *bool) {
	m := p.marker()

	// Argument sink: `..sink`.
	if p.eatIf(Dots) {
		if p.atSet(PatternLeafSet) {
			patternLeaf(p, false, seen, "parameter")
		}
		p.wrap(m, Spread)
		if *sink {
			p.nodes[m].ConvertToError("only one argument sink is allowed")
		}
		*sink = true
		return
	}

	// Normal positional parameter or parameter name.
	wasAtPat := p.atSet(PatternSet)
	pattern(p, false, seen, "parameter")

	// Named parameter: `thickness: 12pt`.
	if p.eatIf(Colon) {
		// Recover from bad parameter name.
		if wasAtPat && p.nodes[m].Kind() != Ident {
			p.nodes[m].Expected("identifier")
		}

		codeExpr(p)
		p.wrap(m, Named)
	}
}

// pattern parses a binding or reassignment pattern.
func pattern(p *Parser, reassignment bool, seen map[string]bool, dupe string) {
	cleanup := p.increaseDepth()
	if cleanup == nil {
		return
	}
	defer cleanup()

	switch p.current() {
	case Underscore:
		p.eat()
	case LeftParen:
		destructuringOrParenthesized(p, reassignment, seen)
	default:
		patternLeaf(p, reassignment, seen, dupe)
	}
}

// destructuringOrParenthesized parses a destructuring or parenthesized pattern.
func destructuringOrParenthesized(p *Parser, reassignment bool, seen map[string]bool) {
	sink := false
	count := 0
	maybeJustParens := true

	m := p.marker()
	p.withNLMode(NLContinue, func(p *Parser) {
		p.assert(LeftParen)

		for !p.current().IsTerminator() {
			if !p.atSet(DestructuringItemSet) {
				p.unexpected()
				continue
			}

			destructuringItem(p, reassignment, seen, &maybeJustParens, &sink)
			count++

			if !p.current().IsTerminator() && p.expect(Comma) {
				maybeJustParens = false
			}
		}

		p.expectClosingDelimiter(m, RightParen)
	})

	if maybeJustParens && count == 1 && !sink {
		p.wrap(m, Parenthesized)
	} else {
		p.wrap(m, Destructuring)
	}
}

// destructuringItem parses an item in a destructuring pattern.
func destructuringItem(p *Parser, reassignment bool, seen map[string]bool, maybeJustParens *bool, sink *bool) {
	m := p.marker()

	// Destructuring sink: `..rest`.
	if p.eatIf(Dots) {
		if p.atSet(PatternLeafSet) {
			patternLeaf(p, reassignment, seen, "")
		}
		p.wrap(m, Spread)
		if *sink {
			p.nodes[m].ConvertToError("only one destructuring sink is allowed")
		}
		*sink = true
		return
	}

	// Normal positional pattern or destructuring key.
	wasAtPat := p.atSet(PatternSet)

	// Check if this is a named destructuring item.
	checkpoint := p.checkpoint()
	if p.eatIf(Ident) && p.at(Colon) {
		// Named destructuring item.
	} else {
		p.restore(checkpoint)
		pattern(p, reassignment, seen, "")
	}

	// Named destructuring item.
	if p.eatIf(Colon) {
		// Recover from bad named destructuring.
		if wasAtPat && p.nodes[m].Kind() != Ident {
			p.nodes[m].Expected("identifier")
		}

		pattern(p, reassignment, seen, "")
		p.wrap(m, Named)
		*maybeJustParens = false
	}
}

// patternLeaf parses a leaf in a pattern.
func patternLeaf(p *Parser, reassignment bool, seen map[string]bool, dupe string) {
	if p.current().IsKeyword() {
		p.eatAndGet().Expected("pattern")
		return
	} else if !p.atSet(PatternLeafSet) {
		p.expected("pattern")
		return
	}

	m := p.marker()
	text := p.currentText()

	// Parse an atomic expression for better error recovery.
	codeExprPrec(p, true, 0)

	if !reassignment {
		node := p.nodes[m]
		if node.Kind() == Ident {
			if seen[text] {
				dupeName := "binding"
				if dupe != "" {
					dupeName = dupe
				}
				node.ConvertToError("duplicate " + dupeName + ": " + text)
			}
			seen[text] = true
		} else {
			node.Expected("pattern")
		}
	}
}

// UnOpFromSyntaxKind converts a SyntaxKind to a UnOp pointer.
func UnOpFromSyntaxKind(kind SyntaxKind) *UnOp {
	switch kind {
	case Plus:
		op := UnOpPos
		return &op
	case Minus:
		op := UnOpNeg
		return &op
	case Not:
		op := UnOpNot
		return &op
	}
	return nil
}

// Precedence returns the precedence of a unary operator.
func (op UnOp) Precedence() int {
	return 7 // All unary operators have the same precedence
}
