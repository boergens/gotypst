// Package eval provides variable capture analysis for closures.
// Translated from typst-eval/src/call.rs (CapturesVisitor)

package eval

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// CapturesVisitor determines which variables to capture for a closure.
// It walks the AST and identifies variables that need to be captured
// from outer scopes.
//
// Matches Rust's CapturesVisitor in call.rs.
type CapturesVisitor struct {
	// external is the outer scopes to capture from (may be nil for IDE analysis).
	external *foundations.Scopes
	// internal tracks newly bound variables within the closure.
	internal *foundations.Scopes
	// captures is the resulting scope of captured variables.
	captures *foundations.Scope
	// capturer indicates what kind of construct is capturing (function vs context).
	capturer foundations.Capturer
}

// NewCapturesVisitor creates a new visitor for the given external scopes.
func NewCapturesVisitor(external *foundations.Scopes, capturer foundations.Capturer) *CapturesVisitor {
	return &CapturesVisitor{
		external: external,
		internal: foundations.NewScopes(nil),
		captures: foundations.NewScope(),
		capturer: capturer,
	}
}

// Finish returns the scope of captured variables.
func (v *CapturesVisitor) Finish() *foundations.Scope {
	return v.captures
}

// Visit visits any node and collects all captured variables.
func (v *CapturesVisitor) Visit(node *syntax.SyntaxNode) {
	if node == nil {
		return
	}

	switch node.Kind() {
	case syntax.Ident:
		// Every identifier is a potential variable that we need to capture.
		// Identifiers that shouldn't count as captures because they
		// actually bind a new name are handled below.
		ident := syntax.IdentExprFromNode(node)
		if ident != nil {
			v.capture(ident.Get(), false)
		}

	case syntax.MathIdent:
		// Math identifiers use different lookup rules.
		// Single-letter idents in math are symbols, not variables.
		ident := syntax.MathIdentExprFromNode(node)
		if ident != nil {
			v.capture(ident.Get(), true)
		}

	case syntax.CodeBlock, syntax.ContentBlock:
		// Code and content blocks create a scope.
		v.internal.Enter()
		for _, child := range node.Children() {
			v.Visit(child)
		}
		v.internal.Exit()

	case syntax.FieldAccess:
		// Don't capture the field of a field access, only the target.
		access := syntax.FieldAccessExprFromNode(node)
		if access != nil {
			v.Visit(access.Target().ToUntyped())
		}

	case syntax.Closure:
		// A closure contains parameter bindings, which are bound before the
		// body is evaluated. Care must be taken so that the default values
		// of named parameters cannot access previous parameter bindings.
		expr := syntax.ClosureExprFromNode(node)
		if expr == nil {
			return
		}

		// Visit default values for named parameters (before binding params).
		if params := expr.Params(); params != nil {
			for _, param := range params.Children() {
				if np, ok := param.(*syntax.NamedParam); ok {
					if defExpr := np.Default(); defExpr != nil {
						v.Visit(defExpr.ToUntyped())
					}
				}
			}
		}

		// Enter a new internal scope for the closure.
		v.internal.Enter()

		// Bind the function name for recursion.
		if name := expr.Name(); name != nil {
			v.bind(name)
		}

		// Bind all parameters.
		if params := expr.Params(); params != nil {
			for _, param := range params.Children() {
				switch p := param.(type) {
				case *syntax.PosParam:
					// Positional parameter - bind the name
					if name := p.Name(); name != nil {
						v.bind(name)
					}
				case *syntax.NamedParam:
					// Named parameter - bind the name
					if name := p.Name(); name != nil {
						v.bind(name)
					}
				case *syntax.SinkParam:
					// Sink parameter - bind the name if present
					if name := p.Name(); name != nil {
						v.bind(name)
					}
				case *syntax.DestructuringParam:
					// Destructuring parameter - bind pattern names
					if pattern := p.Pattern(); pattern != nil {
						destPattern := syntax.DestructuringPatternFromNode(pattern.ToUntyped())
						if destPattern != nil {
							for _, ident := range destPattern.Bindings() {
								v.bind(ident)
							}
						}
					}
				case *syntax.PlaceholderParam:
					// Placeholder parameter - nothing to bind
				}
			}
		}

		// Visit the body.
		v.Visit(expr.Body().ToUntyped())
		v.internal.Exit()

	case syntax.LetBinding:
		// A let expression contains a binding, but that binding is only
		// active after the body is evaluated.
		expr := syntax.LetBindingExprFromNode(node)
		if expr == nil {
			return
		}

		// Visit the initializer first.
		if init := expr.Init(); init != nil {
			v.Visit(init.ToUntyped())
		}

		// Then bind the pattern.
		for _, ident := range expr.Bindings() {
			v.bind(ident)
		}

	case syntax.ForLoop:
		// A for loop contains bindings in its pattern. These are
		// active after the iterable is evaluated but before the body.
		expr := syntax.ForLoopExprFromNode(node)
		if expr == nil {
			return
		}

		// Visit the iterable first.
		if iter := expr.Iter(); iter != nil {
			v.Visit(iter.ToUntyped())
		}

		// Enter scope and bind pattern.
		v.internal.Enter()
		if pattern := expr.Pattern(); pattern != nil {
			for _, ident := range pattern.Bindings() {
				v.bind(ident)
			}
		}

		// Visit the body.
		if body := expr.Body(); body != nil {
			v.Visit(body.ToUntyped())
		}
		v.internal.Exit()

	case syntax.ModuleImport:
		// An import contains items, but these are active only after
		// the path is evaluated.
		expr := syntax.ModuleImportExprFromNode(node)
		if expr == nil {
			return
		}

		// Visit the source first.
		if source := expr.Source(); source != nil {
			v.Visit(source.ToUntyped())
		}

		// Then bind the imported items.
		imports := expr.Imports()
		if items, ok := imports.(*syntax.ImportItemsNode); ok {
			for _, item := range items.Items() {
				if ident := item.BoundName(); ident != nil {
					v.bind(ident)
				}
			}
		}
		// For wildcard imports, we don't bind specific names.
		// For no imports (just `import "path"`), we might bind the module name.
		if newName := expr.NewName(); newName != nil {
			v.bind(newName)
		}

	case syntax.Named:
		// Never capture the name part of a named pair, only the value.
		named := syntax.NamedArgFromNode(node)
		if named != nil {
			if expr := named.Expr(); expr != nil {
				v.Visit(expr.ToUntyped())
			}
			return
		}
		// If it's not a NamedArg, traverse children normally.
		for _, child := range node.Children() {
			v.Visit(child)
		}

	default:
		// Everything else is traversed from left to right.
		for _, child := range node.Children() {
			v.Visit(child)
		}
	}
}

// bind adds a new internal variable binding.
// The concrete value does not matter as we only use the scoping
// mechanism of Scopes, not the values themselves.
func (v *CapturesVisitor) bind(ident *syntax.IdentExpr) {
	if ident == nil {
		return
	}
	v.internal.Top().Bind(ident.Get(), foundations.NewBindingDetached(foundations.NoneValue{}))
}

// capture captures a variable if it isn't internal.
func (v *CapturesVisitor) capture(name string, isMath bool) {
	// Skip if already bound internally.
	if v.internal.Get(name) != nil {
		return
	}

	// Look up in external scopes.
	var binding *foundations.Binding
	if v.external != nil {
		if isMath {
			binding = v.external.GetInMath(name)
		} else {
			binding = v.external.Get(name)
		}
	}

	if binding == nil {
		// If external scopes are nil (IDE analysis), create a detached binding.
		if v.external == nil {
			binding = &foundations.Binding{}
			*binding = foundations.NewBindingDetached(foundations.NoneValue{})
		} else {
			// Variable not found in external scopes - nothing to capture.
			return
		}
	}

	// Capture the binding with the appropriate kind.
	captured := binding.Capture(v.capturer)
	v.captures.Bind(name, captured)
}
