// Package csvbuddy implements a convenient interface for encoding and decoding CSV files.
//
// Only slices of structs can be encoded and decoded because CSV is defined as a list of records.
//
// The following struct field types are supported:
// bool, int*, uint*, float*, complex*, []byte, string,
// encoding.TextMarshaler, encoding.TextUnmarshaler.
// Other values produce an error.
//
// Pointers to any of the above types are interpreted as optional types.
// Optional types are decoded if the parsed field is not an empty string,
// and they are encoded as an empty string if the pointer is nil.
package csvbuddy
