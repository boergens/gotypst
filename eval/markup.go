// Markup evaluation for Typst.
// Translated from typst-eval/src/markup.rs

package eval

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Markup Evaluation
// ----------------------------------------------------------------------------

// evalMarkup evaluates a stream of markup expressions.
// Matches Rust: fn eval_markup(vm: &mut Vm, exprs: &mut impl Iterator<Item = ast::Expr>) -> SourceResult<Content>
func evalMarkup(vm *Vm, markup *syntax.MarkupNode) (foundations.Value, error) {
	flow := vm.TakeFlow()
	exprs := markup.Exprs()
	seq := make([]foundations.Content, 0, len(exprs))

	i := 0
loop:
	for i < len(exprs) {
		expr := exprs[i]
		i++

		switch e := expr.(type) {
		case *syntax.SetRuleExpr:
			// Evaluate set rule to get styles
			styles, err := evalSetRuleToStyles(vm, e)
			if err != nil {
				return nil, err
			}
			if vm.HasFlow() {
				break loop
			}

			// Evaluate remaining markup as tail
			tailMarkup := syntax.MarkupNodeFromExprs(exprs[i:])
			tail, err := evalMarkup(vm, tailMarkup)
			if err != nil {
				return nil, err
			}
			tailContent := Display(tail)
			seq = append(seq, tailContent.StyledWithMap(styles))
			i = len(exprs) // Mark all consumed

		case *syntax.ShowRuleExpr:
			// Evaluate show rule to get recipe
			recipe, err := evalShowRuleToRecipe(vm, e)
			if err != nil {
				return nil, err
			}
			if vm.HasFlow() {
				break loop
			}

			// Evaluate remaining markup as tail
			tailMarkup := syntax.MarkupNodeFromExprs(exprs[i:])
			tail, err := evalMarkup(vm, tailMarkup)
			if err != nil {
				return nil, err
			}
			tailContent := Display(tail)
			styledContent, err := tailContent.StyledWithRecipe(vm.Engine, vm.Context, recipe)
			if err != nil {
				return nil, err
			}
			seq = append(seq, styledContent)
			i = len(exprs)

		default:
			value, err := evalExpr(vm, expr)
			if err != nil {
				return nil, err
			}

			// Handle labels specially - attach to previous element
			if label, ok := value.(foundations.LabelValue); ok {
				// Find the last non-unlabellable element
				attached := false
				for j := len(seq) - 1; j >= 0; j-- {
					if !seq[j].IsUnlabellable() {
						// Check if already labelled
						if seq[j].Label() != nil {
							vm.Engine.Sink.Warn(foundations.SourceDiagnostic{
								Span:     seq[j].Span(),
								Severity: foundations.SeverityWarning,
								Message:  "content labelled multiple times",
								Hints:    []string{"only the last label is used, the rest are ignored"},
							})
						}
						seq[j] = seq[j].Labelled(label)
						attached = true
						break
					}
				}
				if !attached {
					vm.Engine.Sink.Warn(foundations.SourceDiagnostic{
						Span:     expr.ToUntyped().Span(),
						Severity: foundations.SeverityWarning,
						Message:  "label is not attached to anything",
					})
				}
			} else {
				// Add displayed value to sequence
				content := Display(value).WithSpan(expr.ToUntyped().Span())
				seq = append(seq, content)
			}
		}

		if vm.HasFlow() {
			break loop
		}
	}

	if flow != nil {
		vm.SetFlow(flow)
	}

	return foundations.ContentValue{Content: foundations.Sequence(seq)}, nil
}

// ----------------------------------------------------------------------------
// Text and Whitespace
// ----------------------------------------------------------------------------

// evalText evaluates a text expression.
// Matches Rust: impl Eval for ast::Text
func evalText(_ *Vm, e *syntax.TextExpr) (foundations.Value, error) {
	return foundations.ContentValue{Content: foundations.TextElemPacked(e.Get())}, nil
}

// evalSpace evaluates a space expression.
// Matches Rust: impl Eval for ast::Space
func evalSpace(_ *Vm, _ *syntax.SpaceExpr) (foundations.Value, error) {
	return foundations.ContentValue{Content: foundations.SpaceElemShared()}, nil
}

// evalLinebreak evaluates a linebreak expression.
// Matches Rust: impl Eval for ast::Linebreak
func evalLinebreak(_ *Vm, _ *syntax.LinebreakExpr) (foundations.Value, error) {
	return foundations.ContentValue{Content: foundations.LinebreakElemShared()}, nil
}

// evalParbreak evaluates a paragraph break expression.
// Matches Rust: impl Eval for ast::Parbreak
func evalParbreak(_ *Vm, _ *syntax.ParbreakExpr) (foundations.Value, error) {
	return foundations.ContentValue{Content: foundations.ParbreakElemShared()}, nil
}

// ----------------------------------------------------------------------------
// Escape and Shorthand
// ----------------------------------------------------------------------------

// evalEscape evaluates an escape sequence.
// Matches Rust: impl Eval for ast::Escape
func evalEscape(_ *Vm, e *syntax.EscapeExpr) (foundations.Value, error) {
	return foundations.SymbolValue{Char: e.Get()}, nil
}

// evalShorthand evaluates a shorthand expression.
// Matches Rust: impl Eval for ast::Shorthand
func evalShorthand(_ *Vm, e *syntax.ShorthandExpr) (foundations.Value, error) {
	return foundations.SymbolValue{Char: e.Get()}, nil
}

// ----------------------------------------------------------------------------
// Smart Quotes
// ----------------------------------------------------------------------------

// evalSmartQuote evaluates a smart quote expression.
// Matches Rust: impl Eval for ast::SmartQuote
func evalSmartQuote(_ *Vm, e *syntax.SmartQuoteExpr) (foundations.Value, error) {
	return foundations.ContentValue{Content: foundations.SmartQuoteElemPacked(e.Double())}, nil
}

// ----------------------------------------------------------------------------
// Strong and Emph
// ----------------------------------------------------------------------------

// evalStrong evaluates a strong (bold) expression.
// Matches Rust: impl Eval for ast::Strong
func evalStrong(vm *Vm, e *syntax.StrongExpr) (foundations.Value, error) {
	body := e.Body()
	if body == nil {
		return foundations.ContentValue{}, nil
	}

	// Warn if empty
	if len(body.Exprs()) == 0 {
		vm.Engine.Sink.Warn(foundations.SourceDiagnostic{
			Span:     e.ToUntyped().Span(),
			Severity: foundations.SeverityWarning,
			Message:  "no text within stars",
			Hints:    []string{"using multiple consecutive stars (e.g. **) has no additional effect"},
		})
	}

	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.StrongElemPacked(Display(content))}, nil
}

// evalEmph evaluates an emphasis (italic) expression.
// Matches Rust: impl Eval for ast::Emph
func evalEmph(vm *Vm, e *syntax.EmphExpr) (foundations.Value, error) {
	body := e.Body()
	if body == nil {
		return foundations.ContentValue{}, nil
	}

	// Warn if empty
	if len(body.Exprs()) == 0 {
		vm.Engine.Sink.Warn(foundations.SourceDiagnostic{
			Span:     e.ToUntyped().Span(),
			Severity: foundations.SeverityWarning,
			Message:  "no text within underscores",
			Hints:    []string{"using multiple consecutive underscores (e.g. __) has no additional effect"},
		})
	}

	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.EmphElemPacked(Display(content))}, nil
}

// ----------------------------------------------------------------------------
// Raw
// ----------------------------------------------------------------------------

// evalRaw evaluates a raw (code) expression.
// Matches Rust: impl Eval for ast::Raw
func evalRaw(_ *Vm, e *syntax.RawExpr) (foundations.Value, error) {
	lines := e.Lines()
	lang := e.Lang()
	block := e.Block()

	return foundations.ContentValue{Content: foundations.RawElemPacked(lines, lang, block)}, nil
}

// ----------------------------------------------------------------------------
// Link
// ----------------------------------------------------------------------------

// evalLink evaluates a link expression.
// Matches Rust: impl Eval for ast::Link
func evalLink(_ *Vm, e *syntax.LinkExpr) (foundations.Value, error) {
	url := e.Get()
	return foundations.ContentValue{Content: foundations.LinkElemPacked(url)}, nil
}

// ----------------------------------------------------------------------------
// Label
// ----------------------------------------------------------------------------

// evalLabel evaluates a label expression.
// Matches Rust: impl Eval for ast::Label
func evalLabel(_ *Vm, e *syntax.LabelExpr) (foundations.Value, error) {
	return foundations.LabelValue(e.Get()), nil
}

// ----------------------------------------------------------------------------
// Reference
// ----------------------------------------------------------------------------

// evalRef evaluates a reference expression.
// Matches Rust: impl Eval for ast::Ref
func evalRef(vm *Vm, e *syntax.RefExpr) (foundations.Value, error) {
	target := e.Target()

	var supplement *foundations.Content
	if supp := e.Supplement(); supp != nil {
		suppValue, err := evalMarkup(vm, supp)
		if err != nil {
			return nil, err
		}
		content := Display(suppValue)
		supplement = &content
	}

	return foundations.ContentValue{Content: foundations.RefElemPacked(target, supplement)}, nil
}

// ----------------------------------------------------------------------------
// Heading
// ----------------------------------------------------------------------------

// evalHeading evaluates a heading expression.
// Matches Rust: impl Eval for ast::Heading
func evalHeading(vm *Vm, e *syntax.HeadingExpr) (foundations.Value, error) {
	depth := e.Level()
	body := e.Body()
	if body == nil {
		return foundations.ContentValue{}, nil
	}

	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.HeadingElemPacked(Display(content), depth)}, nil
}

// ----------------------------------------------------------------------------
// List Items
// ----------------------------------------------------------------------------

// evalListItem evaluates a list item expression.
// Matches Rust: impl Eval for ast::ListItem
func evalListItem(vm *Vm, e *syntax.ListItemExpr) (foundations.Value, error) {
	body := e.Body()
	if body == nil {
		return foundations.ContentValue{}, nil
	}

	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.ListItemPacked(Display(content))}, nil
}

// evalEnumItem evaluates an enum item expression.
// Matches Rust: impl Eval for ast::EnumItem
func evalEnumItem(vm *Vm, e *syntax.EnumItemExpr) (foundations.Value, error) {
	body := e.Body()
	if body == nil {
		return foundations.ContentValue{}, nil
	}

	content, err := evalMarkup(vm, body)
	if err != nil {
		return nil, err
	}

	number := e.Number()
	return foundations.ContentValue{Content: foundations.EnumItemPacked(Display(content), number)}, nil
}

// evalTermItem evaluates a term item expression.
// Matches Rust: impl Eval for ast::TermItem
func evalTermItem(vm *Vm, e *syntax.TermItemExpr) (foundations.Value, error) {
	term := e.Term()
	desc := e.Description()

	var termContent, descContent foundations.Content
	if term != nil {
		termValue, err := evalMarkup(vm, term)
		if err != nil {
			return nil, err
		}
		termContent = Display(termValue)
	}
	if desc != nil {
		descValue, err := evalMarkup(vm, desc)
		if err != nil {
			return nil, err
		}
		descContent = Display(descValue)
	}

	return foundations.ContentValue{Content: foundations.TermItemPacked(termContent, descContent)}, nil
}
