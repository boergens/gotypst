package syntax

// SyntaxMode represents the lexer's mode which determines which tokens it produces.
type SyntaxMode uint8

const (
	// ModeMarkup is the default mode for document content.
	ModeMarkup SyntaxMode = iota
	// ModeMath is used within math equations ($...$).
	ModeMath
	// ModeCode is used within code blocks ({...}) and expressions.
	ModeCode
)

// String returns a human-readable name for the syntax mode.
func (m SyntaxMode) String() string {
	switch m {
	case ModeMarkup:
		return "markup"
	case ModeMath:
		return "math"
	case ModeCode:
		return "code"
	default:
		return "unknown"
	}
}
