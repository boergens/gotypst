package pdf

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Object represents a PDF object that can be written to a PDF file.
type Object interface {
	// writeTo writes the object to the writer.
	writeTo(w io.Writer) (int64, error)
}

// Ref represents a reference to an indirect object.
type Ref struct {
	num int
	gen int
}

// NewRef creates a new object reference.
func NewRef(num, gen int) Ref {
	return Ref{num: num, gen: gen}
}

// Num returns the object number.
func (r Ref) Num() int { return r.num }

// Gen returns the generation number.
func (r Ref) Gen() int { return r.gen }

// IsZero returns true if the reference is unset.
func (r Ref) IsZero() bool { return r.num == 0 && r.gen == 0 }

func (r Ref) writeTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "%d %d R", r.num, r.gen)
	return int64(n), err
}

// Null represents the PDF null object.
type Null struct{}

func (Null) writeTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, "null")
	return int64(n), err
}

// Bool represents a PDF boolean.
type Bool bool

func (b Bool) writeTo(w io.Writer) (int64, error) {
	if b {
		n, err := io.WriteString(w, "true")
		return int64(n), err
	}
	n, err := io.WriteString(w, "false")
	return int64(n), err
}

// Int represents a PDF integer.
type Int int64

func (i Int) writeTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, strconv.FormatInt(int64(i), 10))
	return int64(n), err
}

// Real represents a PDF real number.
type Real float64

func (r Real) writeTo(w io.Writer) (int64, error) {
	// Format with up to 6 decimal places, trimming trailing zeros.
	s := strconv.FormatFloat(float64(r), 'f', 6, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-" {
		s = "0"
	}
	n, err := io.WriteString(w, s)
	return int64(n), err
}

// Name represents a PDF name object.
type Name string

func (n Name) writeTo(w io.Writer) (int64, error) {
	// Names start with / and need special character escaping.
	var buf strings.Builder
	buf.WriteByte('/')
	for i := 0; i < len(n); i++ {
		c := n[i]
		// Characters that need escaping: whitespace, delimiters, # sign
		if c < '!' || c > '~' || c == '#' || c == '(' || c == ')' ||
			c == '<' || c == '>' || c == '[' || c == ']' || c == '{' ||
			c == '}' || c == '/' || c == '%' {
			fmt.Fprintf(&buf, "#%02X", c)
		} else {
			buf.WriteByte(c)
		}
	}
	written, err := io.WriteString(w, buf.String())
	return int64(written), err
}

// String represents a PDF string (literal or hexadecimal).
type String struct {
	data []byte
	hex  bool
}

// NewLiteralString creates a literal string.
func NewLiteralString(s string) String {
	return String{data: []byte(s), hex: false}
}

// NewHexString creates a hexadecimal string.
func NewHexString(data []byte) String {
	return String{data: data, hex: true}
}

func (s String) writeTo(w io.Writer) (int64, error) {
	if s.hex {
		return s.writeHex(w)
	}
	return s.writeLiteral(w)
}

func (s String) writeLiteral(w io.Writer) (int64, error) {
	var buf strings.Builder
	buf.WriteByte('(')
	for _, c := range s.data {
		switch c {
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '\t':
			buf.WriteString("\\t")
		case '\b':
			buf.WriteString("\\b")
		case '\f':
			buf.WriteString("\\f")
		case '(':
			buf.WriteString("\\(")
		case ')':
			buf.WriteString("\\)")
		case '\\':
			buf.WriteString("\\\\")
		default:
			if c < 32 || c > 126 {
				fmt.Fprintf(&buf, "\\%03o", c)
			} else {
				buf.WriteByte(c)
			}
		}
	}
	buf.WriteByte(')')
	n, err := io.WriteString(w, buf.String())
	return int64(n), err
}

func (s String) writeHex(w io.Writer) (int64, error) {
	var buf strings.Builder
	buf.WriteByte('<')
	for _, c := range s.data {
		fmt.Fprintf(&buf, "%02X", c)
	}
	buf.WriteByte('>')
	n, err := io.WriteString(w, buf.String())
	return int64(n), err
}

// Array represents a PDF array.
type Array []Object

func (a Array) writeTo(w io.Writer) (int64, error) {
	var total int64
	n, err := io.WriteString(w, "[")
	total += int64(n)
	if err != nil {
		return total, err
	}
	for i, obj := range a {
		if i > 0 {
			n, err = io.WriteString(w, " ")
			total += int64(n)
			if err != nil {
				return total, err
			}
		}
		written, err := obj.writeTo(w)
		total += written
		if err != nil {
			return total, err
		}
	}
	n, err = io.WriteString(w, "]")
	total += int64(n)
	return total, err
}

// Dict represents a PDF dictionary.
type Dict map[Name]Object

func (d Dict) writeTo(w io.Writer) (int64, error) {
	var total int64
	n, err := io.WriteString(w, "<<")
	total += int64(n)
	if err != nil {
		return total, err
	}

	// Sort keys for deterministic output.
	keys := make([]Name, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, k := range keys {
		v := d[k]
		if v == nil {
			continue
		}
		n, err = io.WriteString(w, "\n")
		total += int64(n)
		if err != nil {
			return total, err
		}
		written, err := k.writeTo(w)
		total += written
		if err != nil {
			return total, err
		}
		n, err = io.WriteString(w, " ")
		total += int64(n)
		if err != nil {
			return total, err
		}
		written, err = v.writeTo(w)
		total += written
		if err != nil {
			return total, err
		}
	}
	n, err = io.WriteString(w, "\n>>")
	total += int64(n)
	return total, err
}

// Stream represents a PDF stream object.
type Stream struct {
	dict Dict
	data []byte
}

// NewStream creates a new stream with the given data.
func NewStream(data []byte) *Stream {
	return &Stream{
		dict: make(Dict),
		data: data,
	}
}

// Dict returns the stream dictionary for modification.
func (s *Stream) Dict() Dict {
	return s.dict
}

// SetFilter sets the stream filter.
func (s *Stream) SetFilter(filter Name) {
	s.dict[Name("Filter")] = filter
}

func (s *Stream) writeTo(w io.Writer) (int64, error) {
	var total int64

	// Set length in dictionary.
	s.dict[Name("Length")] = Int(len(s.data))

	// Write dictionary.
	written, err := s.dict.writeTo(w)
	total += written
	if err != nil {
		return total, err
	}

	// Write stream data.
	n, err := io.WriteString(w, "\nstream\n")
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = w.Write(s.data)
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = io.WriteString(w, "\nendstream")
	total += int64(n)
	return total, err
}

// IndirectObject wraps an object as an indirect object.
type IndirectObject struct {
	ref Ref
	obj Object
}

// NewIndirectObject creates a new indirect object.
func NewIndirectObject(ref Ref, obj Object) *IndirectObject {
	return &IndirectObject{ref: ref, obj: obj}
}

// Ref returns the object reference.
func (io *IndirectObject) Ref() Ref {
	return io.ref
}

// Object returns the contained object.
func (io *IndirectObject) Object() Object {
	return io.obj
}

func (iobj *IndirectObject) writeTo(w io.Writer) (int64, error) {
	var total int64
	n, err := fmt.Fprintf(w, "%d %d obj\n", iobj.ref.num, iobj.ref.gen)
	total += int64(n)
	if err != nil {
		return total, err
	}

	written, err := iobj.obj.writeTo(w)
	total += written
	if err != nil {
		return total, err
	}

	n, err = io.WriteString(w, "\nendobj\n")
	total += int64(n)
	return total, err
}
