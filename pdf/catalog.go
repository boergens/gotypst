package pdf

import (
	"fmt"
	"time"
)

// Catalog represents the PDF document catalog (root object).
type Catalog struct {
	writer   *Writer
	ref      Ref
	pages    Ref
	outlines Ref
	names    Ref
	dests    Ref
	metadata Ref
	markInfo Dict
	lang     string
	pageMode PageMode
	pageLayout PageLayout
	viewerPrefs Dict
}

// PageMode specifies how the document should be displayed when opened.
type PageMode string

const (
	PageModeNone       PageMode = "UseNone"
	PageModeOutlines   PageMode = "UseOutlines"
	PageModeThumbs     PageMode = "UseThumbs"
	PageModeFullScreen PageMode = "FullScreen"
	PageModeOC         PageMode = "UseOC"
	PageModeAttachments PageMode = "UseAttachments"
)

// PageLayout specifies the page layout when the document is opened.
type PageLayout string

const (
	PageLayoutSinglePage     PageLayout = "SinglePage"
	PageLayoutOneColumn      PageLayout = "OneColumn"
	PageLayoutTwoColumnLeft  PageLayout = "TwoColumnLeft"
	PageLayoutTwoColumnRight PageLayout = "TwoColumnRight"
	PageLayoutTwoPageLeft    PageLayout = "TwoPageLeft"
	PageLayoutTwoPageRight   PageLayout = "TwoPageRight"
)

// NewCatalog creates a new document catalog.
func NewCatalog(w *Writer) *Catalog {
	return &Catalog{
		writer: w,
		ref:    w.Alloc(),
	}
}

// Ref returns the catalog reference.
func (c *Catalog) Ref() Ref {
	return c.ref
}

// SetPages sets the page tree reference.
func (c *Catalog) SetPages(ref Ref) {
	c.pages = ref
}

// SetOutlines sets the document outlines (bookmarks) reference.
func (c *Catalog) SetOutlines(ref Ref) {
	c.outlines = ref
}

// SetNames sets the name dictionary reference.
func (c *Catalog) SetNames(ref Ref) {
	c.names = ref
}

// SetDests sets the destinations dictionary reference.
func (c *Catalog) SetDests(ref Ref) {
	c.dests = ref
}

// SetMetadata sets the metadata stream reference.
func (c *Catalog) SetMetadata(ref Ref) {
	c.metadata = ref
}

// SetLang sets the document language tag (e.g., "en-US").
func (c *Catalog) SetLang(lang string) {
	c.lang = lang
}

// SetPageMode sets how the document should be displayed when opened.
func (c *Catalog) SetPageMode(mode PageMode) {
	c.pageMode = mode
}

// SetPageLayout sets the page layout when the document is opened.
func (c *Catalog) SetPageLayout(layout PageLayout) {
	c.pageLayout = layout
}

// SetMarked sets whether the document contains marked content.
func (c *Catalog) SetMarked(marked bool) {
	if c.markInfo == nil {
		c.markInfo = make(Dict)
	}
	c.markInfo[Name("Marked")] = Bool(marked)
}

// SetViewerPreference sets a viewer preference.
func (c *Catalog) SetViewerPreference(key Name, value Object) {
	if c.viewerPrefs == nil {
		c.viewerPrefs = make(Dict)
	}
	c.viewerPrefs[key] = value
}

// Finish writes the catalog to the PDF.
func (c *Catalog) Finish() {
	dict := Dict{
		Name("Type"): Name("Catalog"),
	}

	if !c.pages.IsZero() {
		dict[Name("Pages")] = c.pages
	}
	if !c.outlines.IsZero() {
		dict[Name("Outlines")] = c.outlines
	}
	if !c.names.IsZero() {
		dict[Name("Names")] = c.names
	}
	if !c.dests.IsZero() {
		dict[Name("Dests")] = c.dests
	}
	if !c.metadata.IsZero() {
		dict[Name("Metadata")] = c.metadata
	}
	if c.lang != "" {
		dict[Name("Lang")] = NewLiteralString(c.lang)
	}
	if c.pageMode != "" {
		dict[Name("PageMode")] = Name(c.pageMode)
	}
	if c.pageLayout != "" {
		dict[Name("PageLayout")] = Name(c.pageLayout)
	}
	if len(c.markInfo) > 0 {
		dict[Name("MarkInfo")] = c.markInfo
	}
	if len(c.viewerPrefs) > 0 {
		dict[Name("ViewerPreferences")] = c.viewerPrefs
	}

	c.writer.Write(c.ref, dict)
	c.writer.SetCatalog(c.ref)
}

// Info represents the document information dictionary.
type Info struct {
	writer       *Writer
	ref          Ref
	title        string
	author       string
	subject      string
	keywords     string
	creator      string
	producer     string
	creationDate time.Time
	modDate      time.Time
}

// NewInfo creates a new document info dictionary.
func NewInfo(w *Writer) *Info {
	return &Info{
		writer: w,
		ref:    w.Alloc(),
	}
}

// Ref returns the info dictionary reference.
func (i *Info) Ref() Ref {
	return i.ref
}

// SetTitle sets the document title.
func (i *Info) SetTitle(title string) {
	i.title = title
}

// SetAuthor sets the document author.
func (i *Info) SetAuthor(author string) {
	i.author = author
}

// SetSubject sets the document subject.
func (i *Info) SetSubject(subject string) {
	i.subject = subject
}

// SetKeywords sets the document keywords.
func (i *Info) SetKeywords(keywords string) {
	i.keywords = keywords
}

// SetCreator sets the application that created the document.
func (i *Info) SetCreator(creator string) {
	i.creator = creator
}

// SetProducer sets the PDF producer.
func (i *Info) SetProducer(producer string) {
	i.producer = producer
}

// SetCreationDate sets the document creation date.
func (i *Info) SetCreationDate(t time.Time) {
	i.creationDate = t
}

// SetModDate sets the document modification date.
func (i *Info) SetModDate(t time.Time) {
	i.modDate = t
}

// Finish writes the info dictionary to the PDF.
func (i *Info) Finish() {
	dict := make(Dict)

	if i.title != "" {
		dict[Name("Title")] = NewLiteralString(i.title)
	}
	if i.author != "" {
		dict[Name("Author")] = NewLiteralString(i.author)
	}
	if i.subject != "" {
		dict[Name("Subject")] = NewLiteralString(i.subject)
	}
	if i.keywords != "" {
		dict[Name("Keywords")] = NewLiteralString(i.keywords)
	}
	if i.creator != "" {
		dict[Name("Creator")] = NewLiteralString(i.creator)
	}
	if i.producer != "" {
		dict[Name("Producer")] = NewLiteralString(i.producer)
	}
	if !i.creationDate.IsZero() {
		dict[Name("CreationDate")] = NewLiteralString(formatPDFDate(i.creationDate))
	}
	if !i.modDate.IsZero() {
		dict[Name("ModDate")] = NewLiteralString(formatPDFDate(i.modDate))
	}

	i.writer.Write(i.ref, dict)
	i.writer.SetInfo(i.ref)
}

// formatPDFDate formats a time as a PDF date string.
// PDF date format: D:YYYYMMDDHHmmSSOHH'mm'
func formatPDFDate(t time.Time) string {
	// Get timezone offset.
	_, offset := t.Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	return t.Format("D:20060102150405") + sign +
		padInt(hours, 2) + "'" + padInt(minutes, 2) + "'"
}

func padInt(n, width int) string {
	return fmt.Sprintf("%0*d", width, n)
}
