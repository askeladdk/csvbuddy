// Package csvbuddy implements a convenient interface for encoding and decoding CSV files.
//
// Only slices of structs can be encoded and decoded because CSV is defined as a list of records.
//
// Every exported struct field is interpreted as a CSV column.
// Struct fields are automatically mapped by name to a CSV column.
// Use the "csv" struct field tag to customize how each field is marshaled.
//
//  // StructA demonstrates CSV struct field tags.
//  type StructA struct {
//      // The first param is always the name, which can be empty.
//      // Default is the name of the field.
//      Name string `csv:"name"`
//      // Exported fields with name "-" are ignored.
//      Ignored int `csv:"-"`
//      // Use base to set the integer base. Default is 10.
//      Hex uint `csv:"addr,base=16"`
//      // Use prec and fmt to set floating point precision and format. Default is -1 and 'f'.
//      Flt float64 `csv:"flt,prec=6,fmt=E"`
//      // Inline structs with inline tag.
//      // Any csv fields in the inlined struct are also (un)marshaled.
//      // Beware of naming clashes.
//      B StructB `csv:",inline"`
//      // Embedded structs do not need the inline tag.
//      StructC
//  }
//
// The following struct field types are supported:
// bool, int[8, 16, 32, 64], uint[8, 16, 32, 64], float[32, 64], complex[64, 128],
// []byte, string, encoding.TextMarshaler, encoding.TextUnmarshaler.
// Other values produce an error.
//
// Pointers to any of the above types are interpreted as optional types.
// Optional types are decoded if the parsed field is not an empty string,
// and they are encoded as an empty string if the pointer is nil.
package csvbuddy
