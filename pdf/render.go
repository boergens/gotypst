package pdf

// Renderer provides font and tag management for PDF generation.
// The actual rendering is done by Writer using the transform-based approach
// that matches the Rust reference implementation.
type Renderer struct {
	// FontManager manages CID fonts for the document.
	FontManager *FontManager

	// tagManager handles PDF/UA accessibility tagging.
	tagManager *TagManager

	// activeTagCount tracks the number of active marked content regions.
	activeTagCount int
}

// NewRenderer creates a new PDF renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		FontManager: NewFontManager(),
	}
}

// SetTagManager sets the tag manager for accessibility tagging.
func (r *Renderer) SetTagManager(tm *TagManager) {
	r.tagManager = tm
}

// TagManager returns the current tag manager.
func (r *Renderer) TagManager() *TagManager {
	return r.tagManager
}
