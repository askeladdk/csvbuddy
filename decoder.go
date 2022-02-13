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

// NewDecoder returns a Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader:     r,
		readerFunc: NewReader,
		mapFunc:    func(_, v string) string { return v },
	}
}

func (d *Decoder) decodeFunc(structType reflect.Type, fn func(reflect.Value) error) error {
	var err error

	cr := d.readerFunc(d.reader)

	var header []string
	if d.skipHeader {
		if header, err = headerOf(structType); err != nil {
			return err
		}
	} else if header, err = cr.Read(); err != nil {
		return err
	}

	var fields []structField
	var indices []int
	if fields, err = structFieldsOf(structType); err != nil {
		return err
	} else if indices, err = headerIndices(header, fields); err != nil {
		return err
	}

	for {
		// parse the next record
		var record []string
		if record, err = cr.Read(); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("csv: %w", err)
		}

		// disallow records that have more or fewer columns than struct fields
		if (d.disallowShortFields && len(record) < len(fields)) || (d.disallowUnknownFields && len(record) > len(fields)) {
			line, _ := fieldPos(cr, 0)
			return fmt.Errorf("csv: %w", &csv.ParseError{
				StartLine: line,
				Line:      line,
				Column:    1,
				Err:       csv.ErrFieldCount,
			})
		}

		structval := reflect.New(structType) // new(T)

		// loop through every (column index, struct field index) pair
		for i := 0; i < len(indices); i += 2 {
			var value string
			var fieldidx int
			// get column if it is within range
			if fieldidx = indices[i]; fieldidx < len(record) {
				value = record[fieldidx]
			}
			// get the struct field
			field := fields[indices[i+1]]
			fieldval := structval.Elem().FieldByIndex(field.Index)
			// clean the value string and type convert it
			value = d.mapFunc(field.Name, value)
			if err = field.Decode(fieldval, value); err != nil {
				if fieldidx >= len(record) {
					fieldidx = 1
				}
				line, column := fieldPos(cr, fieldidx)
				return fmt.Errorf("csv: %w", &csv.ParseError{
					StartLine: line,
					Line:      line,
					Column:    column,
					Err:       err,
				})
			}
		}

		if err = fn(structval); err != nil {
			return fmt.Errorf("csv: %w", err)
		}
	}

	return nil
}

// Decode decodes a CSV as a slice of structs and stores it in v.
// The value of v must be a pointer to a slice of structs.
func (d *Decoder) Decode(v interface{}) (err error) {
	var vv reflect.Value
	if vv, err = valueOf(v); err != nil {
		return
	}

	valueType := innerTypeOf(vv.Type(), reflect.Ptr, reflect.Slice, reflect.Struct)
	if valueType == nil {
		return ErrInvalidArgument
	}

	slice := reflect.MakeSlice(vv.Elem().Type(), 0, 0) // make([]T)

	if err = d.decodeFunc(valueType, func(value reflect.Value) error {
		slice = reflect.Append(slice, value.Elem()) // append(slice, *value)
		return nil
	}); err != nil {
		return err
	}

	vv.Elem().Set(slice) // *v = slice
	return
}

// DecodeFunc implements streaming CSV parsing by calling a function on each record.
// The function signature must be func(*YourStruct) error.
func (d *Decoder) DecodeFunc(fn interface{}) (err error) {
	var fv reflect.Value
	if fv, err = valueOf(fn); err != nil {
		return
	}

	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Func || ft.NumIn() != 1 || ft.NumOut() != 1 || ft.Out(0) != errorType {
		return ErrInvalidArgument
	}

	valueType := innerTypeOf(ft.In(0), reflect.Ptr, reflect.Struct)
	if valueType == nil {
		return ErrInvalidArgument
	}

	return d.decodeFunc(valueType, func(value reflect.Value) error {
		in := [1]reflect.Value{value}
		if out := fv.Call(in[:]); !out[0].IsNil() {
			return out[0].Interface().(error)
		}
		return nil
	})
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

// Unmarshal decodes a byte slice as a CSV to a slice of structs.
// The CSV is expected to be comma-separated and have a header.
func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}
