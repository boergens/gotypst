// Context for Typst evaluation.
// Translated from typst-library/src/foundations/context.rs

package foundations

// Context holds data that is contextually made available to code.
//
// Contextual functions and expressions require the presence of certain
// pieces of context to be evaluated. This includes things like `text.lang`,
// `measure`, or `counter(heading).get()`.
//
// Matches Rust's Context struct in context.rs.
type Context struct {
	// Location is the location in the document.
	Location *Location

	// Styles are the active styles.
	Styles *StyleChain
}

// NewContext creates an empty context.
func NewContext() *Context {
	return &Context{
		Location: nil,
		Styles:   nil,
	}
}

// NewContextWith creates a context with the given location and styles.
func NewContextWith(location *Location, styles *StyleChain) *Context {
	return &Context{
		Location: location,
		Styles:   styles,
	}
}

// GetLocation returns the location, or an error if not available.
func (c *Context) GetLocation() (*Location, error) {
	if c == nil || c.Location == nil {
		return nil, &ContextError{Message: "can only be used when context is known"}
	}
	return c.Location, nil
}

// GetStyles returns the styles, or an error if not available.
func (c *Context) GetStyles() (*StyleChain, error) {
	if c == nil || c.Styles == nil {
		return nil, &ContextError{Message: "can only be used when context is known"}
	}
	return c.Styles, nil
}

// ContextError is returned when context is not available.
type ContextError struct {
	Message string
}

func (e *ContextError) Error() string {
	return e.Message
}
