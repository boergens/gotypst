package syntax

// SyntaxError represents an error that occurred during parsing.
type SyntaxError struct {
	Message string
	Hints   []string
}

// NewSyntaxError creates a new syntax error with the given message.
func NewSyntaxError(message string) *SyntaxError {
	return &SyntaxError{
		Message: message,
		Hints:   nil,
	}
}

// Error implements the error interface.
func (e *SyntaxError) Error() string {
	return e.Message
}

// AddHint adds a hint to the error.
func (e *SyntaxError) AddHint(hint string) {
	e.Hints = append(e.Hints, hint)
}
