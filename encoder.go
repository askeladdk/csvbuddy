package csvbuddy

import (
	"bytes"
	"io"
	"reflect"
)

// Encoder writes and encodes CSV records to an output stream.
type Encoder struct {
	writer     io.Writer
	writerFunc WriterFunc
	mapFunc    MapFunc
	skipHeader bool
}

// NewEncoder creates a new Encoder.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer:     w,
		writerFunc: NewWriter,
		mapFunc:    func(_, v string) string { return v },
	}
}

// Encode encodes a slice of structs to CSV text format.
// The value of v must be a pointer to a slice of structs.
func (e *Encoder) Encode(v interface{}) (err error) {
	var vv reflect.Value
	if vv, err = valueOf(v); err != nil {
		return
	}

	structType := innerTypeOf(vv.Type(), reflect.Ptr, reflect.Slice, reflect.Struct)
	if structType == nil {
		return ErrInvalidArgument
	}

	var header []string
	if header, err = headerOf(structType); err != nil {
		return
	}

	var fields []structField
	var indices []int
	if fields, err = structFieldsOf(structType); err != nil {
		return
	} else if indices, err = headerIndices(header, fields); err != nil {
		return
	}

	w := e.writerFunc(e.writer)

	if !e.skipHeader {
		if err = w.Write(header); err != nil {
			return
		}
	}

	record := make([]string, len(header))
	slice := vv.Elem()
	for i := 0; i < slice.Len(); i++ {
		structval := slice.Index(i)
		for j := 0; j < len(indices); j += 2 {
			field := fields[indices[j+1]]
			fieldval := structval.FieldByIndex(field.Index)
			var value string
			if value, err = field.Encode(fieldval, &field); err != nil {
				return
			}
			record[indices[j]] = e.mapFunc(field.Name, value)
		}

		if err = w.Write(record); err != nil {
			return
		}
	}

	// special case for csv.Writer because Flush does not return an error
	if csvw, ok := w.(interface {
		Flush()
		Error() error
	}); ok {
		csvw.Flush()
		return csvw.Error()
	} else if flusher, ok := w.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}

	return nil
}

// SetMapFunc causes the Encoder to call fn on every field before a record is written.
func (e *Encoder) SetMapFunc(fn MapFunc) { e.mapFunc = fn }

// SetWriterFunc customizes how records are encoded.
// The default value is NewWriter.
func (e *Encoder) SetWriterFunc(fn WriterFunc) { e.writerFunc = fn }

// SkipHeader causes the Encoder to not write the CSV header.
func (e *Encoder) SkipHeader() { e.skipHeader = true }

// Marshal encodes a slice of structs to a byte slice of CSV text format.
// The CSV will be comma-separated and have a header.
func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	if err := NewEncoder(&b).Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
