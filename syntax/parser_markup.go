package syntax

// markup parses markup expressions until a stop condition is met.
func markup(p *Parser, atStart bool, wrapTrivia bool, stopSet SyntaxSet) {
	var m Marker
	if wrapTrivia {
		m = p.beforeTrivia()
	} else {
		m = p.marker()
	}
	markupExprs(p, atStart, stopSet)
	if wrapTrivia {
		p.flushTrivia()
	}
	p.wrap(m, Markup)
}

// markupExprs parses a sequence of markup expressions.
func markupExprs(p *Parser, atStart bool, stopSet SyntaxSet) {
	if !stopSet.Contains(End) {
		stopSet = stopSet.Add(End)
	}
	if p.depth >= MaxDepth {
		p.depthCheckError(&stopSet)
		return
	}

	atStart = atStart || p.hadNewline()
	nesting := 0
	// Keep going if we're at a nested right-bracket regardless of stop set.
	for !p.atSet(stopSet) || (nesting > 0 && p.at(RightBracket)) {
		markupExpr(p, atStart, &nesting)
		atStart = p.hadNewline()
	}
}

// markupExpr parses a single markup expression.
func markupExpr(p *Parser, atStart bool, nesting *int) {
	cleanup := p.increaseDepth()
	if cleanup == nil {
		return
	}
	defer cleanup()

	switch p.current() {
	case LeftBracket:
		*nesting++
		p.convertAndEat(Text)
	case RightBracket:
		if *nesting > 0 {
			*nesting--
			p.convertAndEat(Text)
		} else {
			p.unexpected()
			p.hint("try using a backslash escape: \\]")
		}

	case Shebang:
		p.eat()

	case Text, Linebreak, Escape, Shorthand, SmartQuote, Link, Label:
		p.eat()

	case Raw:
		p.eat() // Raw is handled entirely in the Lexer.

	case Hash:
		embeddedCodeExpr(p)
	case Star:
		strong(p)
	case Underscore:
		emph(p)
	case HeadingMarker:
		if atStart {
			heading(p)
		} else {
			p.convertAndEat(Text)
		}
	case ListMarker:
		if atStart {
			listItem(p)
		} else {
			p.convertAndEat(Text)
		}
	case EnumMarker:
		if atStart {
			enumItem(p)
		} else {
			p.convertAndEat(Text)
		}
	case TermMarker:
		if atStart {
			termItem(p)
		} else {
			p.convertAndEat(Text)
		}
	case RefMarker:
		reference(p)
	case Dollar:
		equation(p)

	case Colon:
		p.convertAndEat(Text)

	default:
		p.unexpected()
	}
}

// strong parses strong content: `*Strong*`.
func strong(p *Parser) {
	p.withNLMode(NLStopParBreak, func(p *Parser) {
		m := p.marker()
		p.assert(Star)
		markup(p, false, true, SyntaxSetOf(Star, RightBracket, End))
		p.expectClosingDelimiter(m, Star)
		p.wrap(m, Strong)
	})
}

// emph parses emphasized content: `_Emphasized_`.
func emph(p *Parser) {
	p.withNLMode(NLStopParBreak, func(p *Parser) {
		m := p.marker()
		p.assert(Underscore)
		markup(p, false, true, SyntaxSetOf(Underscore, RightBracket, End))
		p.expectClosingDelimiter(m, Underscore)
		p.wrap(m, Emph)
	})
}

// heading parses a section heading: `= Introduction`.
func heading(p *Parser) {
	p.withNLMode(NLStop, func(p *Parser) {
		m := p.marker()
		p.assert(HeadingMarker)
		markup(p, false, false, SyntaxSetOf(Label, RightBracket, End))
		p.wrap(m, Heading)
	})
}

// listItem parses an item in a bullet list: `- ...`.
func listItem(p *Parser) {
	col := p.currentColumn()
	p.withNLMode(requireColumn(col), func(p *Parser) {
		m := p.marker()
		p.assert(ListMarker)
		markup(p, true, false, SyntaxSetOf(RightBracket, End))
		p.wrap(m, ListItem)
	})
}

// enumItem parses an item in an enumeration: `+ ...` or `1. ...`.
func enumItem(p *Parser) {
	col := p.currentColumn()
	p.withNLMode(requireColumn(col), func(p *Parser) {
		m := p.marker()
		p.assert(EnumMarker)
		markup(p, true, false, SyntaxSetOf(RightBracket, End))
		p.wrap(m, EnumItem)
	})
}

// termItem parses an item in a term list: `/ Term: Details`.
func termItem(p *Parser) {
	col := p.currentColumn()
	p.withNLMode(requireColumn(col), func(p *Parser) {
		m := p.marker()
		p.withNLMode(NLStop, func(p *Parser) {
			p.assert(TermMarker)
			markup(p, false, false, SyntaxSetOf(Colon, RightBracket, End))
		})
		p.expect(Colon)
		markup(p, true, false, SyntaxSetOf(RightBracket, End))
		p.wrap(m, TermItem)
	})
}

// reference parses a reference: `@target`, `@target[..]`.
func reference(p *Parser) {
	m := p.marker()
	p.assert(RefMarker)
	if p.directlyAt(LeftBracket) {
		contentBlock(p)
	}
	p.wrap(m, Ref)
}

// equation parses a mathematical equation: `$x$`, `$ x^2 $`.
func equation(p *Parser) {
	m := p.marker()
	p.enterModes(ModeMath, NLContinue, func(p *Parser) {
		p.assert(Dollar)
		parseMathContent(p, SyntaxSetOf(Dollar, End))
		p.expectClosingDelimiter(m, Dollar)
	})
	p.wrap(m, Equation)
}
