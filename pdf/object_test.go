package pdf

import (
	"bytes"
	"testing"
)

func TestNull(t *testing.T) {
	var buf bytes.Buffer
	n := Null{}
	_, err := n.writeTo(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "null" {
		t.Errorf("got %q, want %q", got, "null")
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		val  Bool
		want string
	}{
		{true, "true"},
		{false, "false"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		_, err := tt.val.writeTo(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := buf.String(); got != tt.want {
			t.Errorf("Bool(%v) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		val  Int
		want string
	}{
		{0, "0"},
		{42, "42"},
		{-100, "-100"},
		{1234567890, "1234567890"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		_, err := tt.val.writeTo(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := buf.String(); got != tt.want {
			t.Errorf("Int(%v) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestReal(t *testing.T) {
	tests := []struct {
		val  Real
		want string
	}{
		{0.0, "0"},
		{3.14159, "3.14159"},
		{-2.5, "-2.5"},
		{100.0, "100"},
		{0.001, "0.001"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		_, err := tt.val.writeTo(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := buf.String(); got != tt.want {
			t.Errorf("Real(%v) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		val  Name
		want string
	}{
		{"Type", "/Type"},
		{"Page", "/Page"},
		{"Font", "/Font"},
		{"A#B", "/A#23B"},      // # needs escaping
		{"A B", "/A#20B"},      // space needs escaping
		{"Name()", "/Name#28#29"}, // parens need escaping
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		_, err := tt.val.writeTo(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := buf.String(); got != tt.want {
			t.Errorf("Name(%q) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestLiteralString(t *testing.T) {
	tests := []struct {
		val  string
		want string
	}{
		{"Hello", "(Hello)"},
		{"Hello\nWorld", "(Hello\\nWorld)"},
		{"(nested)", "(\\(nested\\))"},
		{"back\\slash", "(back\\\\slash)"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		s := NewLiteralString(tt.val)
		_, err := s.writeTo(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := buf.String(); got != tt.want {
			t.Errorf("LiteralString(%q) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestHexString(t *testing.T) {
	tests := []struct {
		val  []byte
		want string
	}{
		{[]byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}, "<48656C6C6F>"},
		{[]byte{0x00, 0xFF}, "<00FF>"},
		{[]byte{}, "<>"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		s := NewHexString(tt.val)
		_, err := s.writeTo(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := buf.String(); got != tt.want {
			t.Errorf("HexString(%v) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestArray(t *testing.T) {
	tests := []struct {
		arr  Array
		want string
	}{
		{Array{}, "[]"},
		{Array{Int(1), Int(2), Int(3)}, "[1 2 3]"},
		{Array{Name("Type"), Name("Page")}, "[/Type /Page]"},
		{Array{Real(0), Real(0), Real(612), Real(792)}, "[0 0 612 792]"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		_, err := tt.arr.writeTo(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := buf.String(); got != tt.want {
			t.Errorf("Array = %q, want %q", got, tt.want)
		}
	}
}

func TestDict(t *testing.T) {
	dict := Dict{
		Name("Type"): Name("Page"),
		Name("Count"): Int(1),
	}
	var buf bytes.Buffer
	_, err := dict.writeTo(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	// Keys are sorted alphabetically.
	if !bytes.Contains([]byte(got), []byte("/Count 1")) {
		t.Errorf("Dict missing Count: %q", got)
	}
	if !bytes.Contains([]byte(got), []byte("/Type /Page")) {
		t.Errorf("Dict missing Type: %q", got)
	}
}

func TestRef(t *testing.T) {
	ref := NewRef(1, 0)
	var buf bytes.Buffer
	_, err := ref.writeTo(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := buf.String(), "1 0 R"; got != want {
		t.Errorf("Ref = %q, want %q", got, want)
	}
}

func TestStream(t *testing.T) {
	data := []byte("BT /F1 12 Tf 100 700 Td (Hello) Tj ET")
	stream := NewStream(data)
	var buf bytes.Buffer
	_, err := stream.writeTo(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	if !bytes.Contains([]byte(got), []byte("/Length 37")) {
		t.Errorf("Stream missing Length: %q", got)
	}
	if !bytes.Contains([]byte(got), []byte("stream\n")) {
		t.Errorf("Stream missing stream keyword: %q", got)
	}
	if !bytes.Contains([]byte(got), []byte("\nendstream")) {
		t.Errorf("Stream missing endstream: %q", got)
	}
	if !bytes.Contains([]byte(got), data) {
		t.Errorf("Stream missing data: %q", got)
	}
}

func TestIndirectObject(t *testing.T) {
	ref := NewRef(5, 0)
	obj := Dict{Name("Type"): Name("Page")}
	iobj := NewIndirectObject(ref, obj)

	var buf bytes.Buffer
	_, err := iobj.writeTo(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	if !bytes.HasPrefix([]byte(got), []byte("5 0 obj\n")) {
		t.Errorf("IndirectObject wrong start: %q", got)
	}
	if !bytes.HasSuffix([]byte(got), []byte("\nendobj\n")) {
		t.Errorf("IndirectObject wrong end: %q", got)
	}
}
