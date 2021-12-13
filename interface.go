package csvbuddy

import (
	"encoding/csv"
	"io"
)

func fieldPos(r Reader, field int) (line int, column int) {
	if fp, ok := r.(interface{ FieldPos(int) (int, int) }); ok {
		return fp.FieldPos(field)
	}
	return 0, 0
}

// Reader parses a CSV input stream to records.
// A Reader must return io.EOF to signal end of file.
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

// Writer writes CSV records.
//
// Writer may optionally support flushing by implementing Flush() error.
type Writer interface {
	Write([]string) error
}

// NewWriter returns a new csv.Writer that writes to w.
func NewWriter(w io.Writer) Writer {
	return csv.NewWriter(w)
}

// MapFunc is a function that replaces a field value by another value.
type MapFunc func(name, value string) string

// ReaderFunc is a function that returns a Reader that reads from an input stream.
type ReaderFunc func(io.Reader) Reader

// WriterFunc is a function that returns a Writer that writes to an output stream.
type WriterFunc func(io.Writer) Writer
