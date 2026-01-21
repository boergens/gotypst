// Exporting Typst documents to PDF.
//
// This file corresponds to typst-pdf/src/lib.rs

package pdf

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/boergens/gotypst/layout/pages"
)

// PDF exports a document into a PDF file.
//
// Returns the raw bytes making up the PDF file.
// This is the main entry point for PDF export, corresponding to
// the pdf() function in typst-pdf/src/lib.rs.
func PDF(document *pages.PagedDocument, options *Options) ([]byte, error) {
	return convert(document, options)
}

// convert handles the actual PDF conversion.
// This corresponds to convert::convert() in Rust.
func convert(document *pages.PagedDocument, options *Options) ([]byte, error) {
	// Use defaults if options is nil
	if options == nil {
		options = DefaultOptions()
	}

	// Create writer based on tagging preference
	var w *Writer
	if options.Tagged {
		w = NewTaggedWriter()
	} else {
		w = NewWriter()
	}

	// TODO: Apply page ranges filtering if options.PageRanges is set
	// TODO: Apply standards validation

	// Write to buffer
	var buf bytesBuffer
	if err := w.Write(document, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// bytesBuffer is a simple bytes.Buffer wrapper for io.Writer.
type bytesBuffer struct {
	data []byte
}

func (b *bytesBuffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesBuffer) Bytes() []byte {
	return b.data
}

// Options contains settings for PDF export.
// This corresponds to PdfOptions in typst-pdf/src/lib.rs.
type Options struct {
	// Ident is a string that uniquely and stably identifies the document.
	// It should not change between compilations of the same document.
	// If nil (Auto), a hash of the document's title and author is used instead.
	//
	// If an Ident is given, the hash of it will be used to create a PDF
	// document identifier (the identifier itself is not leaked).
	Ident *string

	// Timestamp is the creation timestamp of the document.
	// It will only be used if document date is auto.
	Timestamp *Timestamp

	// PageRanges specifies which ranges of pages should be exported.
	// When nil, all pages are exported.
	PageRanges *PageRanges

	// Standards is a list of PDF standards that Typst will enforce conformance with.
	Standards Standards

	// Tagged indicates whether to produce tagged PDF output.
	// By default (true), even when not producing a PDF/UA-1 document,
	// a tagged PDF document is written to provide a baseline of accessibility.
	// Set to false to disable tagged PDF (e.g., to reduce document size).
	Tagged bool
}

// DefaultOptions returns the default PDF export options.
func DefaultOptions() *Options {
	return &Options{
		Ident:      nil, // Auto
		Timestamp:  nil,
		PageRanges: nil,
		Standards:  DefaultStandards(),
		Tagged:     true,
	}
}

// IsPDFUA returns whether the current export mode is PDF/UA-1.
func (o *Options) IsPDFUA() bool {
	return o.Standards.Validator() == ValidatorUA1
}

// Timestamp represents a datetime with timezone information.
// This corresponds to Timestamp in typst-pdf/src/metadata.rs.
type Timestamp struct {
	// Time is the datetime value.
	Time time.Time
	// Timezone is the timezone offset information.
	Timezone *Timezone
}

// NewTimestamp creates a new timestamp from a time.Time value.
func NewTimestamp(t time.Time) *Timestamp {
	_, offset := t.Zone()
	return &Timestamp{
		Time: t,
		Timezone: &Timezone{
			OffsetMinutes: offset / 60,
		},
	}
}

// Timezone represents timezone offset information.
// This corresponds to Timezone in typst-pdf/src/metadata.rs.
type Timezone struct {
	// OffsetMinutes is the UTC offset in minutes.
	// Positive values are east of UTC, negative are west.
	OffsetMinutes int
}

// PageRanges specifies which pages to include in the export.
// This corresponds to PageRanges in typst-library.
type PageRanges struct {
	// Ranges is a list of page range specifications.
	Ranges []PageRange
}

// PageRange represents a single page range.
type PageRange struct {
	// Start is the first page (1-indexed, inclusive).
	Start int
	// End is the last page (1-indexed, inclusive).
	// If 0, extends to the end of the document.
	End int
}

// Contains returns whether the given page number (1-indexed) is included.
func (pr *PageRanges) Contains(page int) bool {
	if pr == nil || len(pr.Ranges) == 0 {
		return true // All pages included
	}
	for _, r := range pr.Ranges {
		if page >= r.Start && (r.End == 0 || page <= r.End) {
			return true
		}
	}
	return false
}

// Standards encapsulates a list of compatible PDF standards.
// This corresponds to PdfStandards in typst-pdf/src/lib.rs.
type Standards struct {
	version   PDFVersion
	validator Validator
}

// DefaultStandards returns the default PDF standards configuration (PDF 1.7).
func DefaultStandards() Standards {
	return Standards{
		version:   PDFVersion17,
		validator: ValidatorNone,
	}
}

// NewStandards validates a list of PDF standards for compatibility
// and returns their encapsulated representation.
// This corresponds to PdfStandards::new() in Rust.
func NewStandards(list []Standard) (Standards, error) {
	var version *PDFVersion
	var validator *Validator

	setVersion := func(v PDFVersion) error {
		if version != nil {
			return fmt.Errorf("PDF cannot conform to %s and %s at the same time",
				version.String(), v.String())
		}
		version = &v
		return nil
	}

	setValidator := func(v Validator) error {
		if validator != nil {
			return errors.New("Typst currently only supports one PDF substandard at a time")
		}
		validator = &v
		return nil
	}

	for _, standard := range list {
		switch standard {
		case StandardV14:
			if err := setVersion(PDFVersion14); err != nil {
				return Standards{}, err
			}
		case StandardV15:
			if err := setVersion(PDFVersion15); err != nil {
				return Standards{}, err
			}
		case StandardV16:
			if err := setVersion(PDFVersion16); err != nil {
				return Standards{}, err
			}
		case StandardV17:
			if err := setVersion(PDFVersion17); err != nil {
				return Standards{}, err
			}
		case StandardV20:
			if err := setVersion(PDFVersion20); err != nil {
				return Standards{}, err
			}
		case StandardA1b:
			if err := setValidator(ValidatorA1B); err != nil {
				return Standards{}, err
			}
		case StandardA1a:
			if err := setValidator(ValidatorA1A); err != nil {
				return Standards{}, err
			}
		case StandardA2b:
			if err := setValidator(ValidatorA2B); err != nil {
				return Standards{}, err
			}
		case StandardA2u:
			if err := setValidator(ValidatorA2U); err != nil {
				return Standards{}, err
			}
		case StandardA2a:
			if err := setValidator(ValidatorA2A); err != nil {
				return Standards{}, err
			}
		case StandardA3b:
			if err := setValidator(ValidatorA3B); err != nil {
				return Standards{}, err
			}
		case StandardA3u:
			if err := setValidator(ValidatorA3U); err != nil {
				return Standards{}, err
			}
		case StandardA3a:
			if err := setValidator(ValidatorA3A); err != nil {
				return Standards{}, err
			}
		case StandardA4:
			if err := setValidator(ValidatorA4); err != nil {
				return Standards{}, err
			}
		case StandardA4f:
			if err := setValidator(ValidatorA4F); err != nil {
				return Standards{}, err
			}
		case StandardA4e:
			if err := setValidator(ValidatorA4E); err != nil {
				return Standards{}, err
			}
		case StandardUA1:
			if err := setValidator(ValidatorUA1); err != nil {
				return Standards{}, err
			}
		}
	}

	// Build configuration based on what was specified
	result := Standards{}
	switch {
	case version != nil && validator != nil:
		if !isCompatible(*version, *validator) {
			return Standards{}, fmt.Errorf("%s is not compatible with %s",
				version.String(), validator.String())
		}
		result.version = *version
		result.validator = *validator
	case version != nil:
		result.version = *version
		result.validator = ValidatorNone
	case validator != nil:
		result.version = defaultVersionForValidator(*validator)
		result.validator = *validator
	default:
		result.version = PDFVersion17
		result.validator = ValidatorNone
	}

	return result, nil
}

// Version returns the PDF version.
func (s Standards) Version() PDFVersion {
	return s.version
}

// Validator returns the PDF validator/substandard.
func (s Standards) Validator() Validator {
	return s.validator
}

// String returns a debug representation of the standards.
func (s Standards) String() string {
	return "PdfStandards(..)"
}

// PDFVersion represents a PDF version.
type PDFVersion int

const (
	PDFVersion14 PDFVersion = iota
	PDFVersion15
	PDFVersion16
	PDFVersion17
	PDFVersion20
)

// String returns the version string (e.g., "1.7").
func (v PDFVersion) String() string {
	switch v {
	case PDFVersion14:
		return "1.4"
	case PDFVersion15:
		return "1.5"
	case PDFVersion16:
		return "1.6"
	case PDFVersion17:
		return "1.7"
	case PDFVersion20:
		return "2.0"
	default:
		return "1.7"
	}
}

// Validator represents a PDF substandard validator.
type Validator int

const (
	ValidatorNone Validator = iota
	ValidatorA1B
	ValidatorA1A
	ValidatorA2B
	ValidatorA2U
	ValidatorA2A
	ValidatorA3B
	ValidatorA3U
	ValidatorA3A
	ValidatorA4
	ValidatorA4F
	ValidatorA4E
	ValidatorUA1
)

// String returns the validator name.
func (v Validator) String() string {
	switch v {
	case ValidatorNone:
		return "none"
	case ValidatorA1B:
		return "PDF/A-1b"
	case ValidatorA1A:
		return "PDF/A-1a"
	case ValidatorA2B:
		return "PDF/A-2b"
	case ValidatorA2U:
		return "PDF/A-2u"
	case ValidatorA2A:
		return "PDF/A-2a"
	case ValidatorA3B:
		return "PDF/A-3b"
	case ValidatorA3U:
		return "PDF/A-3u"
	case ValidatorA3A:
		return "PDF/A-3a"
	case ValidatorA4:
		return "PDF/A-4"
	case ValidatorA4F:
		return "PDF/A-4f"
	case ValidatorA4E:
		return "PDF/A-4e"
	case ValidatorUA1:
		return "PDF/UA-1"
	default:
		return "unknown"
	}
}

// isCompatible checks if a PDF version is compatible with a validator.
func isCompatible(version PDFVersion, validator Validator) bool {
	switch validator {
	case ValidatorNone:
		return true
	case ValidatorA1B, ValidatorA1A:
		// PDF/A-1 requires PDF 1.4
		return version == PDFVersion14
	case ValidatorA2B, ValidatorA2U, ValidatorA2A:
		// PDF/A-2 requires PDF 1.7 or earlier
		return version <= PDFVersion17
	case ValidatorA3B, ValidatorA3U, ValidatorA3A:
		// PDF/A-3 requires PDF 1.7 or earlier
		return version <= PDFVersion17
	case ValidatorA4, ValidatorA4F, ValidatorA4E:
		// PDF/A-4 requires PDF 2.0
		return version == PDFVersion20
	case ValidatorUA1:
		// PDF/UA-1 requires PDF 1.7 or earlier
		return version <= PDFVersion17
	default:
		return false
	}
}

// defaultVersionForValidator returns the default PDF version for a validator.
func defaultVersionForValidator(validator Validator) PDFVersion {
	switch validator {
	case ValidatorA1B, ValidatorA1A:
		return PDFVersion14
	case ValidatorA4, ValidatorA4F, ValidatorA4E:
		return PDFVersion20
	default:
		return PDFVersion17
	}
}

// Standard represents a PDF standard that Typst can enforce conformance with.
// This corresponds to PdfStandard enum in typst-pdf/src/lib.rs.
type Standard int

const (
	// StandardV14 is PDF 1.4.
	StandardV14 Standard = iota
	// StandardV15 is PDF 1.5.
	StandardV15
	// StandardV16 is PDF 1.6.
	StandardV16
	// StandardV17 is PDF 1.7.
	StandardV17
	// StandardV20 is PDF 2.0.
	StandardV20
	// StandardA1b is PDF/A-1b.
	StandardA1b
	// StandardA1a is PDF/A-1a.
	StandardA1a
	// StandardA2b is PDF/A-2b.
	StandardA2b
	// StandardA2u is PDF/A-2u.
	StandardA2u
	// StandardA2a is PDF/A-2a.
	StandardA2a
	// StandardA3b is PDF/A-3b.
	StandardA3b
	// StandardA3u is PDF/A-3u.
	StandardA3u
	// StandardA3a is PDF/A-3a.
	StandardA3a
	// StandardA4 is PDF/A-4.
	StandardA4
	// StandardA4f is PDF/A-4f.
	StandardA4f
	// StandardA4e is PDF/A-4e.
	StandardA4e
	// StandardUA1 is PDF/UA-1.
	StandardUA1
)

// String returns the standard name for serialization.
func (s Standard) String() string {
	switch s {
	case StandardV14:
		return "1.4"
	case StandardV15:
		return "1.5"
	case StandardV16:
		return "1.6"
	case StandardV17:
		return "1.7"
	case StandardV20:
		return "2.0"
	case StandardA1b:
		return "a-1b"
	case StandardA1a:
		return "a-1a"
	case StandardA2b:
		return "a-2b"
	case StandardA2u:
		return "a-2u"
	case StandardA2a:
		return "a-2a"
	case StandardA3b:
		return "a-3b"
	case StandardA3u:
		return "a-3u"
	case StandardA3a:
		return "a-3a"
	case StandardA4:
		return "a-4"
	case StandardA4f:
		return "a-4f"
	case StandardA4e:
		return "a-4e"
	case StandardUA1:
		return "ua-1"
	default:
		return ""
	}
}

// ParseStandard parses a standard name string into a Standard value.
func ParseStandard(s string) (Standard, error) {
	switch s {
	case "1.4":
		return StandardV14, nil
	case "1.5":
		return StandardV15, nil
	case "1.6":
		return StandardV16, nil
	case "1.7":
		return StandardV17, nil
	case "2.0":
		return StandardV20, nil
	case "a-1b":
		return StandardA1b, nil
	case "a-1a":
		return StandardA1a, nil
	case "a-2b":
		return StandardA2b, nil
	case "a-2u":
		return StandardA2u, nil
	case "a-2a":
		return StandardA2a, nil
	case "a-3b":
		return StandardA3b, nil
	case "a-3u":
		return StandardA3u, nil
	case "a-3a":
		return StandardA3a, nil
	case "a-4":
		return StandardA4, nil
	case "a-4f":
		return StandardA4f, nil
	case "a-4e":
		return StandardA4e, nil
	case "ua-1":
		return StandardUA1, nil
	default:
		return 0, fmt.Errorf("unknown PDF standard: %s", s)
	}
}

// WritePDF is a convenience function that writes a PDF to an io.Writer.
// This is similar to the existing Export function but uses Options.
func WritePDF(document *pages.PagedDocument, options *Options, out io.Writer) error {
	data, err := PDF(document, options)
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}
