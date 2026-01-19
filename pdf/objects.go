package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Object represents a PDF object.
type Object interface {
	writeTo(w io.Writer) error
}

// Ref represents an indirect object reference.
type Ref struct {
	ID  int
	Gen int
}

// writeTo writes the reference in PDF format.
func (r Ref) writeTo(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%d %d R", r.ID, r.Gen)
	return err
}

// Name represents a PDF name object.
type Name string

// writeTo writes the name in PDF format.
func (n Name) writeTo(w io.Writer) error {
	// Escape special characters in the name
	var buf bytes.Buffer
	buf.WriteByte('/')
	for i := 0; i < len(n); i++ {
		c := n[i]
		if c < '!' || c > '~' || c == '#' || c == '/' || c == '(' || c == ')' ||
			c == '<' || c == '>' || c == '[' || c == ']' || c == '{' || c == '}' ||
			c == '%' {
			fmt.Fprintf(&buf, "#%02X", c)
		} else {
			buf.WriteByte(c)
		}
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// String represents a PDF string object.
type String string

// writeTo writes the string in PDF format (literal string).
func (s String) writeTo(w io.Writer) error {
	var buf bytes.Buffer
	buf.WriteByte('(')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\', '(', ')':
			buf.WriteByte('\\')
			buf.WriteByte(c)
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '\t':
			buf.WriteString("\\t")
		default:
			buf.WriteByte(c)
		}
	}
	buf.WriteByte(')')
	_, err := w.Write(buf.Bytes())
	return err
}

// HexString represents a PDF hexadecimal string.
type HexString []byte

// writeTo writes the hex string in PDF format.
func (h HexString) writeTo(w io.Writer) error {
	_, err := fmt.Fprintf(w, "<%X>", []byte(h))
	return err
}

// Int represents a PDF integer.
type Int int

// writeTo writes the integer in PDF format.
func (i Int) writeTo(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%d", i)
	return err
}

// Real represents a PDF real number.
type Real float64

// writeTo writes the real number in PDF format.
func (r Real) writeTo(w io.Writer) error {
	s := strconv.FormatFloat(float64(r), 'f', -1, 64)
	// Remove trailing zeros after decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	_, err := w.Write([]byte(s))
	return err
}

// Bool represents a PDF boolean.
type Bool bool

// writeTo writes the boolean in PDF format.
func (b Bool) writeTo(w io.Writer) error {
	if b {
		_, err := w.Write([]byte("true"))
		return err
	}
	_, err := w.Write([]byte("false"))
	return err
}

// Null represents the PDF null object.
type Null struct{}

// writeTo writes the null object.
func (Null) writeTo(w io.Writer) error {
	_, err := w.Write([]byte("null"))
	return err
}

// Array represents a PDF array.
type Array []Object

// writeTo writes the array in PDF format.
func (a Array) writeTo(w io.Writer) error {
	if _, err := w.Write([]byte("[")); err != nil {
		return err
	}
	for i, obj := range a {
		if i > 0 {
			if _, err := w.Write([]byte(" ")); err != nil {
				return err
			}
		}
		if err := obj.writeTo(w); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("]"))
	return err
}

// Dict represents a PDF dictionary.
type Dict map[Name]Object

// writeTo writes the dictionary in PDF format.
func (d Dict) writeTo(w io.Writer) error {
	if _, err := w.Write([]byte("<<")); err != nil {
		return err
	}
	for key, val := range d {
		if err := key.writeTo(w); err != nil {
			return err
		}
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		if err := val.writeTo(w); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(">>"))
	return err
}

// Stream represents a PDF stream object.
type Stream struct {
	Dict Dict
	Data []byte
}

// writeTo writes the stream in PDF format.
func (s Stream) writeTo(w io.Writer) error {
	// Add Length to dictionary
	dict := make(Dict)
	for k, v := range s.Dict {
		dict[k] = v
	}
	dict[Name("Length")] = Int(len(s.Data))

	if err := dict.writeTo(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\nstream\n")); err != nil {
		return err
	}
	if _, err := w.Write(s.Data); err != nil {
		return err
	}
	_, err := w.Write([]byte("\nendstream"))
	return err
}

// Compress compresses the stream data using zlib/FlateDecode.
func (s *Stream) Compress() error {
	if s.Dict == nil {
		s.Dict = make(Dict)
	}

	// Check if already compressed
	if _, ok := s.Dict[Name("Filter")]; ok {
		return nil
	}

	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(s.Data); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}

	s.Data = buf.Bytes()
	s.Dict[Name("Filter")] = Name("FlateDecode")
	return nil
}

// IndirectObject wraps an object with its reference information.
type IndirectObject struct {
	Ref    Ref
	Object Object
}

// writeTo writes the indirect object definition.
func (o IndirectObject) writeTo(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%d %d obj\n", o.Ref.ID, o.Ref.Gen); err != nil {
		return err
	}
	if err := o.Object.writeTo(w); err != nil {
		return err
	}
	_, err := w.Write([]byte("\nendobj\n"))
	return err
}
