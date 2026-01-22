// Data value types for Typst.
// Str, Bytes, Label, Decimal types.

package foundations

import "math/big"

// Str represents a string value.
// Note: Str methods are in str.go
type Str string

func (Str) Type() Type         { return TypeStr }
func (v Str) Display() Content { return Content{} }
func (v Str) Clone() Value     { return v }
func (Str) isValue()           {}

// BytesValue represents a sequence of bytes.
type BytesValue []byte

func (BytesValue) Type() Type         { return TypeBytes }
func (v BytesValue) Display() Content { return Content{} }
func (v BytesValue) Clone() Value {
	if v == nil {
		return BytesValue(nil)
	}
	clone := make([]byte, len(v))
	copy(clone, v)
	return BytesValue(clone)
}
func (BytesValue) isValue() {}

// LabelValue represents a label for referencing content.
type LabelValue string

func (LabelValue) Type() Type         { return TypeLabel }
func (v LabelValue) Display() Content { return Content{} }
func (v LabelValue) Clone() Value     { return v }
func (LabelValue) isValue()           {}

// DecimalValue represents an arbitrary-precision decimal number.
type DecimalValue struct {
	Value *big.Rat
}

func (DecimalValue) Type() Type         { return TypeDecimal }
func (v DecimalValue) Display() Content { return Content{} }
func (v DecimalValue) Clone() Value {
	if v.Value == nil {
		return DecimalValue{}
	}
	return DecimalValue{Value: new(big.Rat).Set(v.Value)}
}
func (DecimalValue) isValue() {}

// VersionValue represents a semantic version.
type VersionValue struct {
	Major int
	Minor int
	Patch int
}

func (VersionValue) Type() Type         { return TypeVersion }
func (v VersionValue) Display() Content { return Content{} }
func (v VersionValue) Clone() Value     { return v }
func (VersionValue) isValue()           {}
