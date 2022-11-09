package csvbuddy

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// ErrInvalidArgument signals that an interface{} argument is of an invalid type.
var ErrInvalidArgument = errors.New("csv: interface{} argument is of an invalid type")

// Decoder reads and decodes CSV records from an input stream.
type Decoder struct {
	reader                io.Reader
	readerFunc            ReaderFunc
	mapFunc               MapFunc
	disallowUnknownFields bool
	disallowShortFields   bool
	skipHeader            bool
}

// Unmarshal decodes a byte slice as a CSV to a slice of structs.
// The CSV is expected to be comma-separated and have a header.
func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

// NewDecoder returns a Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader:     r,
		readerFunc: NewReader,
		mapFunc:    func(_, v string) string { return v },
	}
}

// Decode decodes a CSV as a slice of structs and stores it in v.
// The value of v must be a pointer to a slice of structs.
func (d *Decoder) Decode(v interface{}) error {
	var vv reflect.Value
	vv, err := valueOf(v)
	if err != nil {
		return err
	}

	structType := innerTypeOf(vv.Type(), reflect.Ptr, reflect.Slice, reflect.Struct)
	if structType == nil {
		return ErrInvalidArgument
	}

	slice := reflect.MakeSlice(vv.Elem().Type(), 0, 0) // make([]T)

	var state decodeState

	if err := state.init(d, structType); err != nil {
		return err
	}

	for {
		value, err := state.next()
		if err == io.EOF {
			vv.Elem().Set(slice) // *v = slice
			return nil
		} else if err != nil {
			return err
		}

		slice = reflect.Append(slice, value.Elem()) // append(slice, *value)
	}
}

// DisallowUnknownFields causes the Decoder to raise an error
// if a record has more columns than struct fields.
func (d *Decoder) DisallowUnknownFields() { d.disallowUnknownFields = true }

// DisallowShortFields causes the Decoder to raise an error
// if a record has fewer columns than struct fields.
func (d *Decoder) DisallowShortFields() { d.disallowShortFields = true }

// SetMapFunc causes the Decoder to call fn on every field before type conversion.
// Use this to clean wrongly formatted values.
func (d *Decoder) SetMapFunc(fn MapFunc) { d.mapFunc = fn }

// SkipHeader causes the Decoder to not parse the first
// record as the header but to derive it from the struct tags.
// Use this to read headerless CSVs.
func (d *Decoder) SkipHeader() { d.skipHeader = true }

// SetReaderFunc customizes how records are decoded.
// The default value is NewReader.
func (d *Decoder) SetReaderFunc(fn ReaderFunc) { d.readerFunc = fn }

// DecoderIterator decodes one row at a time
// to enable parsing of large files without
// having to read them entirely into memory.
type DecoderIterator struct {
	state decodeState
	vv    reflect.Value
	err   error
}

// Err returns the most recent non-EOF error.
func (d *DecoderIterator) Err() error {
	return d.err
}

// Scan parses the next row and stores the result in the value passed into Decoder.Iterate.
// It returns false when an error has occurred or it reached EOF.
// After Scan returns false, Err will return the error that caused it to stop.
// If Scan stopped because it has reached EOF, Err will return nil.
func (d *DecoderIterator) Scan() bool {
	var vn reflect.Value
	vn, d.err = d.state.next()
	if errors.Is(d.err, io.EOF) {
		d.err = nil
		return false
	} else if d.err != nil {
		return false
	}

	d.vv.Elem().Set(vn.Elem()) // *vv = *vn
	return true
}

// Iterate returns a DecoderIterator that decodes each row into v,
// which must be a pointer to a struct.
func (d *Decoder) Iterate(v interface{}) (*DecoderIterator, error) {
	var vv reflect.Value
	vv, err := valueOf(v)
	if err != nil {
		return nil, err
	}

	structType := innerTypeOf(vv.Type(), reflect.Ptr, reflect.Struct)
	if structType == nil {
		return nil, ErrInvalidArgument
	}

	var iter DecoderIterator

	if err := iter.state.init(d, structType); err != nil {
		return nil, err
	}

	iter.vv = vv
	return &iter, nil
}

type decodeState struct {
	*Decoder
	r          Reader
	structType reflect.Type
	fields     []structField
	indices    []int
}

func (s *decodeState) init(d *Decoder, structType reflect.Type) error {
	r := d.readerFunc(d.reader)

	header, err := getHeader(structType, r, d.skipHeader)
	if err != nil {
		return err
	}

	fields, indices, err := headerFieldsIndices(structType, header)
	if err != nil {
		return err
	}

	s.Decoder = d
	s.r = r
	s.structType = structType
	s.fields = fields
	s.indices = indices
	return nil
}

func (s *decodeState) next() (reflect.Value, error) {
	record, err := s.r.Read()

	if err == io.EOF {
		return reflect.Value{}, err
	} else if err != nil {
		return reflect.Value{}, fmt.Errorf("csv: %w", err)
	}

	// disallow records that have more or fewer columns than struct fields
	if (s.disallowShortFields && len(record) < len(s.fields)) || (s.disallowUnknownFields && len(record) > len(s.fields)) {
		line, _ := fieldPos(s.r, 0)
		return reflect.Value{}, fmt.Errorf("csv: %w", &csv.ParseError{
			StartLine: line,
			Line:      line,
			Column:    1,
			Err:       csv.ErrFieldCount,
		})
	}

	structval := reflect.New(s.structType) // new(T)

	// loop through every (column index, struct field index) pair
	for i := 0; i < len(s.indices); i += 2 {
		var value string
		var fieldidx int
		// get column if it is within range
		if fieldidx = s.indices[i]; fieldidx < len(record) {
			value = record[fieldidx]
		}
		// get the struct field
		field := s.fields[s.indices[i+1]]
		fieldval := structval.Elem().FieldByIndex(field.Index)
		// clean the value string and type convert it
		value = s.mapFunc(field.Name, value)
		if err = field.Decode(fieldval, value); err != nil {
			if fieldidx >= len(record) {
				fieldidx = 0
			}
			line, column := fieldPos(s.r, fieldidx)
			return reflect.Value{}, fmt.Errorf("csv: %w", &csv.ParseError{
				StartLine: line,
				Line:      line,
				Column:    column,
				Err:       err,
			})
		}
	}

	return structval, nil
}

func getHeader(structType reflect.Type, r Reader, skipHeader bool) ([]string, error) {
	if skipHeader {
		return headerOf(structType)
	}

	return r.Read()
}

func headerFieldsIndices(structType reflect.Type, header []string) (fields []structField, indices []int, err error) {
	if fields, err = structFieldsOf(structType); err != nil {
		return
	} else if indices, err = headerIndices(header, fields); err != nil {
		return
	}
	return
}
