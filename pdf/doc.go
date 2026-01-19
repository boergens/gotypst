// Package pdf provides PDF export functionality for Typst documents.
//
// This package converts compiled Typst documents into PDF format.
// It implements low-level PDF primitives for building PDF files,
// including object serialization, page tree management, and document structure.
//
// # Document Structure
//
// A PDF document consists of:
//   - Header: PDF version identifier
//   - Body: Indirect objects (dictionaries, streams, arrays, etc.)
//   - Cross-reference table: Object locations for random access
//   - Trailer: Document metadata and root references
//
// # Usage
//
// Create a document, add pages, and write to output:
//
//	doc := pdf.NewDocument(pdf.V1_7)
//	doc.Info().SetTitle("My Document")
//	doc.Info().SetProducer("GoTypst")
//
//	page := doc.AddPage(pdf.A4Width, pdf.A4Height)
//	content := doc.AddContentStream([]byte("BT /F1 12 Tf 100 700 Td (Hello) Tj ET"))
//	page.SetContents(content)
//	page.Finish()
//
//	doc.Finish(outputFile)
//
// # Page Tree
//
// Pages are organized in a balanced tree structure for efficient access.
// The PageTree type manages page allocation and tree construction.
//
// # Object Types
//
// The package provides all PDF object types:
//   - Null, Bool, Int, Real: Primitive types
//   - Name, String: Text types
//   - Array, Dict: Container types
//   - Stream: Binary data with metadata
//   - Ref: Indirect object references
package pdf
