package syntax

import "unicode"

// Precedence constants for math operators.
const (
	mathFuncPrec = 2
	mathRootPrec = 2
)

// parseMathContent parses the contents of a mathematical equation.
func parseMathContent(p *Parser, stopSet SyntaxSet) {
	m := p.marker()
	mathExprs(p, stopSet)
	p.wrap(m, Math)
}

// mathExprs parses a sequence of math expressions.
// Returns the number of expressions parsed (including errors).
func mathExprs(p *Parser, stopSet SyntaxSet) int {
	if !stopSet.Contains(End) {
		stopSet = stopSet.Add(End)
	}
	if p.depth >= MaxDepth {
		p.depthCheckError(&stopSet)
		return 1
	}

	count := 0
	for !p.atSet(stopSet) {
		if p.atSet(MathExprSet) {
			mathExpr(p)
		} else {
			p.unexpected()
		}
		count++
	}
	return count
}

// mathExpr parses a single math expression.
func mathExpr(p *Parser) {
	mathExprPrec(p, 0, NewSyntaxSet())
}

// mathExprPrec parses a math expression with at least the given precedence.
func mathExprPrec(p *Parser, minPrec int, stopSet SyntaxSet) {
	cleanup := p.increaseDepth()
	if cleanup == nil {
		return
	}
	defer cleanup()

	m := p.marker()
	continuable := false

	switch p.current() {
	case Hash:
		embeddedCodeExpr(p)

	case MathIdent, FieldAccess:
		continuable = true
		p.eat()
		// Parse a function call for an identifier or field access.
		if mathFuncPrec >= minPrec && p.directlyAt(LeftParen) {
			mathArgs(p)
			p.wrap(m, FuncCall)
			continuable = false
		}

	case LeftBrace, LeftParen:
		mathDelimited(p)

	case RightBrace:
		if p.currentText() == "|]" {
			p.convertAndEat(MathShorthand)
		} else {
			p.convertAndEat(MathText)
		}

	case Dot, Bang, Comma, Semicolon, RightParen:
		p.convertAndEat(MathText)

	case MathText:
		continuable = isMathAlphabetic(p.currentText())
		p.eat()

	case Linebreak, MathAlignPoint, MathShorthand:
		p.eat()

	case MathPrimes, Escape, Str:
		continuable = true
		p.eat()

	case Root:
		p.eat()
		m2 := p.marker()
		mathExprPrec(p, mathRootPrec, NewSyntaxSet())
		mathUnparen(p, m2)
		p.wrap(m, MathRoot)

	default:
		p.expected("expression")
	}

	// Maybe recognize an implicit function call.
	if continuable &&
		mathFuncPrec >= minPrec &&
		!p.hadTrivia() &&
		p.atSet(SyntaxSetOf(LeftBrace, LeftParen)) {
		mathDelimited(p)
		p.wrap(m, Math)
	}

	// Parse infix and postfix operators.
	for !p.atSet(stopSet) {
		opKind := p.current()
		hadTrivia := p.hadTrivia()
		wrapper, infixAssoc, prec, ok := mathOp(opKind, hadTrivia)
		if !ok || prec < minPrec {
			break
		}

		// Prepare a chaining set for attachment operators.
		var chainSet SyntaxSet
		if wrapper == MathAttach {
			chainSet = SyntaxSetOf(Hat, Underscore).Remove(opKind)
		}

		// Eat the operator itself.
		if opKind == Bang {
			p.convertAndEat(MathText)
		} else {
			p.eat()
		}

		// Slash removes parens from its left operand.
		if wrapper == MathFrac {
			mathUnparen(p, m)
		}

		// Parse the operator's right operand.
		if infixAssoc != nil {
			nextPrec := prec
			if *infixAssoc == AssocLeft {
				nextPrec++
			}
			mRhs := p.marker()
			mathExprPrec(p, nextPrec, chainSet)
			mathUnparen(p, mRhs)
		}

		// Avoid interrupting a chain when initially parsing a prime.
		if !(opKind == MathPrimes && p.atSet(stopSet)) {
			// Parse chained attachment operators as a single attachment.
			for p.atSet(chainSet) {
				chainSet = chainSet.Remove(p.current())
				p.eat()
				mChainRhs := p.marker()
				mathExprPrec(p, prec, chainSet)
				mathUnparen(p, mChainRhs)
			}
		}

		// Finish the operator.
		p.wrap(m, wrapper)
	}
}

// mathOp returns the precedence and wrapper for math operators.
func mathOp(kind SyntaxKind, hadTrivia bool) (wrapper SyntaxKind, assoc *Assoc, prec int, ok bool) {
	switch kind {
	case Slash:
		a := AssocLeft
		return MathFrac, &a, 1, true
	case Underscore:
		a := AssocRight
		return MathAttach, &a, 2, true
	case Hat:
		a := AssocRight
		return MathAttach, &a, 2, true
	case MathPrimes:
		if !hadTrivia {
			return MathAttach, nil, 2, true
		}
	case Bang:
		if !hadTrivia {
			return Math, nil, 3, true
		}
	}
	return 0, nil, 0, false
}

// isMathAlphabetic returns true if text is alphabetic for math implicit calls.
func isMathAlphabetic(text string) bool {
	runes := []rune(text)
	if len(runes) == 1 {
		r := runes[0]
		return unicode.IsLetter(r) || isMathClassAlphabetic(r)
	}
	// Multiple characters - check all are alphabetic.
	for _, r := range runes {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// isMathClassAlphabetic checks if a rune has the math class Alphabetic.
// This is a simplified check - the full implementation would use Unicode math class data.
func isMathClassAlphabetic(r rune) bool {
	// Simplified: treat common math letters as alphabetic
	// In full implementation, this would consult unicode-math-class crate data
	return unicode.IsLetter(r)
}

// mathDelimited parses matched delimiters in math: `[x + y]`.
func mathDelimited(p *Parser) {
	m := p.marker()
	if p.currentText() == "[|" {
		p.convertAndEat(MathShorthand)
	} else {
		p.convertAndEat(MathText)
	}
	mBody := p.marker()
	mathExprs(p, SyntaxSetOf(Dollar, End, RightBrace, RightParen))
	if p.atSet(SyntaxSetOf(RightBrace, RightParen)) {
		p.wrap(mBody, Math)
		if p.currentText() == "|]" {
			p.convertAndEat(MathShorthand)
		} else {
			p.convertAndEat(MathText)
		}
		p.wrap(m, MathDelimited)
	} else {
		// No closing delimiter, just produce a math sequence.
		p.wrap(m, Math)
	}
}

// mathUnparen removes one set of parentheses from a parsed expression.
func mathUnparen(p *Parser, m Marker) {
	if int(m) >= len(p.nodes) {
		return
	}
	node := p.nodes[m]
	if node.Kind() != MathDelimited {
		return
	}

	children := node.Children()
	if len(children) >= 2 {
		first := children[0]
		last := children[len(children)-1]
		if first.Text() == "(" && last.Text() == ")" {
			first.ConvertToKind(LeftParen)
			last.ConvertToKind(RightParen)
			node.ConvertToKind(Math)
		}
	}
}

// mathArgs parses an argument list in math: `(a, b; c, d; size: #50%)`.
func mathArgs(p *Parser) {
	m := p.marker()
	p.assert(LeftParen)

	positional := true
	hasArrays := false

	maybeArrayStart := p.marker()
	seen := make(map[string]bool)
	for !p.atSet(SyntaxSetOf(End, Dollar, RightParen)) {
		positional = mathArg(p, seen)

		switch p.current() {
		case Comma:
			p.eat()
			if !positional {
				maybeArrayStart = p.marker()
			}
		case Semicolon:
			if !positional {
				maybeArrayStart = p.marker()
			}
			// Parse an array: `a, b, c;`.
			p.wrap(maybeArrayStart, Array)
			p.eat()
			maybeArrayStart = p.marker()
			hasArrays = true
		case End, Dollar, RightParen:
			// done
		default:
			p.expected("comma or semicolon")
		}
	}

	// Check if we need to wrap preceding arguments in an array.
	if maybeArrayStart != p.marker() && hasArrays && positional {
		p.wrap(maybeArrayStart, Array)
	}

	p.expectClosingDelimiter(m, RightParen)
	p.wrap(m, Args)
}

// mathArg parses a single argument in a math argument list.
// Returns whether the parsed argument was positional.
func mathArg(p *Parser, seen map[string]bool) bool {
	m := p.marker()
	start := p.currentStart()

	var argKind *SyntaxKind

	// Check for spread argument: `..args`.
	if p.at(Dot) {
		if spreadNode := p.lexer.MaybeMathSpreadArg(start); spreadNode != nil {
			k := Spread
			argKind = &k
			p.token.node = spreadNode
			p.eat()
		}
	}

	// Check for named argument: `thickness: #12pt`.
	if argKind == nil && p.atSet(SyntaxSetOf(MathText, MathIdent, Underscore)) {
		if namedNode := p.lexer.MaybeMathNamedArg(start); namedNode != nil {
			k := Named
			argKind = &k
			text := p.currentText()
			p.token.node = namedNode
			p.eat()
			p.convertAndEat(Colon)
			if seen[text] {
				p.nodes[m].ConvertToError("duplicate argument: " + text)
			}
			seen[text] = true
		}
	}

	// Parse the argument itself.
	mArg := p.marker()
	count := mathExprs(p, SyntaxSetOf(End, Dollar, Comma, Semicolon, RightParen))

	if count == 0 {
		// Named arguments require a value.
		if argKind != nil && *argKind == Named {
			p.expected("expression")
		}
		p.flushTrivia()
	}

	// Wrap math function arguments.
	if count != 1 {
		p.wrap(mArg, Math)
	}

	if argKind != nil {
		p.wrap(m, *argKind)
	}
	return argKind == nil || *argKind != Named
}
