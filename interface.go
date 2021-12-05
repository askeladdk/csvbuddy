package csvbuddy

import (
	"encoding/csv"
	"io"
)

// Reader iterates over CSV records.
// A Reader must return io.EOF to signal end of file.
// This interface must be implemented by custom CSV parsers.
type Reader interface {
	Read() ([]string, error)
}

// NewReader returns a new csv.Reader that reads from r.
func NewReader(r io.Reader) Reader {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // fields are checked by Decoder
	cr.ReuseRecord = true
	return cr
}
