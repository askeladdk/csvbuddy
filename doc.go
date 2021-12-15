// Package csvbuddy implements a convenient interface for encoding and decoding CSV files.
//
// Only slices of structs can be encoded and decoded because CSV is defined as a list of records.
//
// Every exported struct field is interpreted as a CSV column.
// Struct fields are automatically mapped by name to a CSV column.
// By default the column name is the same as the field name.
// Use the "csv" tag to encode/decode to a different column.
// The "csv" tag can also take parameters for integer base,
// floating point precision and format.
//
//  // AStruct is an example CSV struct.
//  type AStruct struct {
//      Name    string  `csv:"name"`
//      Hex     uint    `csv:"addr,base=16"`
//      Flt     float64 `csv:"flt,prec=6,fmt=E"`
//      Ignored int     `csv:"-"`
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
