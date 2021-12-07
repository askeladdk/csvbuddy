package csvbuddy

import (
	"bytes"
	"encoding"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// ErrInvalidType signals that an interface{} is of an invalid type.
var ErrInvalidType = errors.New("csv: interface{} contains invalid value")

// DecodingError tells where in the source CSV a decoding error occurred.
type DecodingError struct {
	// Line is the line number.
	Line int
	// Column is the column number.
	Column int

	wrapped error
}

// Error implements the error interface.
func (err *DecodingError) Error() string {
	if err.Line > 0 {
		if err.Column > 0 {
			return fmt.Sprintf("csv: decoding error on line %d, column %d: %s", err.Line, err.Column, err.wrapped)
		}
		return fmt.Sprintf("csv: decoding error on line %d: %s", err.Line, err.wrapped)
	}
	return fmt.Sprintf("csv: decoding error: %s", err.wrapped)
}

// Unwrap implements the errors.Unwrap interface.
func (err *DecodingError) Unwrap() error { return err.wrapped }

func newDecodingError(r Reader, field int, err error) error {
	var e DecodingError
	e.Line, e.Column = fieldPos(r, field)
	e.wrapped = err
	return &e
}

// Decoder reads and decodes CSV records from an input stream.
type Decoder struct {
	reader          io.Reader
	readerFunc      func(io.Reader) Reader
	cleanerFunc     func(column, field string) string
	header          []string
	errors          []error
	skipInvalidRows bool
	noUnknownFields bool
}

// NewDecoder returns a Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader:      r,
		readerFunc:  NewReader,
		cleanerFunc: func(_, field string) string { return field },
	}
}

func (d *Decoder) mapHeader(cr Reader, fields []structField) (indices []int, err error) {
	var header []string

	if len(d.header) > 0 {
		header = d.header
	} else if header, err = cr.Read(); err != nil {
		return
	}

	// check for duplicate header names
	names := map[string]struct{}{}
	for _, h := range header {
		if _, exists := names[h]; exists {
			return nil, fmt.Errorf("duplicate header name '%s'", h)
		}
		names[h] = struct{}{}
	}

	// for each column in header, find index of struct field with matching name
	for i, col := range header {
		for j, field := range fields {
			if field.Name == col {
				indices = append(indices, i, j)
			}
		}
	}

	return
}

func decodeField(v reflect.Value, kind reflect.Kind, name string, value string) (err error) {
	switch kind {
	case reflect.Bool:
		var x bool
		if x, err = strconv.ParseBool(value); err == nil {
			v.SetBool(x)
		}
	case reflect.Complex128:
		var x complex128
		if x, err = strconv.ParseComplex(value, 128); err == nil {
			v.SetComplex(x)
		}
	case reflect.Float64:
		var x float64
		if x, err = strconv.ParseFloat(value, 64); err == nil {
			v.SetFloat(x)
		}
	case reflect.Int64:
		var x int64
		if x, err = strconv.ParseInt(value, 10, 64); err == nil {
			v.SetInt(x)
		}
	case reflect.Ptr:
		if value != "" {
			t := v.Type().Elem()
			v.Set(reflect.New(t))
			err = decodeField(v.Elem(), kindOf(t), name, value)
		}
	case reflect.String:
		v.SetString(value)
	case reflect.Uint64:
		var x uint64
		if x, err = strconv.ParseUint(value, 10, 64); err == nil {
			v.SetUint(x)
		}
	case byteSlice:
		v.SetBytes([]byte(value))
	case textUnmarshaler:
		err = v.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
	default:
		err = fmt.Errorf("cannot decode field '%s'", name)
	}
	return
}

func (d *Decoder) decodeFunc(structType reflect.Type, fn func(reflect.Value) error) error {
	var err error

	// parse the struct data type
	var fields []structField
	if fields, err = structFieldsOf(structType); err != nil {
		return fmt.Errorf("csv: %w", err)
	}

	cr := d.readerFunc(d.reader)

	// parse the header
	var indices []int
	if indices, err = d.mapHeader(cr, fields); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("csv: %w", err)
	}

loop:
	for {
		// parse the next record
		var record []string
		if record, err = cr.Read(); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("csv: %w", err)
		}

		// disallow records that have more columns than struct fields
		if d.noUnknownFields && len(record) > len(fields) {
			err = newDecodingError(cr, 0, csv.ErrFieldCount)
			if !d.skipInvalidRows {
				return err
			}
			d.errors = append(d.errors, err)
			continue loop
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
			value = d.cleanerFunc(field.Name, value)
			if err = decodeField(fieldval, field.Kind, field.Name, value); err != nil {
				if fieldidx >= len(record) {
					fieldidx = len(record) - 1
				}
				err = newDecodingError(cr, fieldidx, err)
				if !d.skipInvalidRows {
					return err
				}
				d.errors = append(d.errors, err)
				continue loop
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
		return ErrInvalidType
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
		return ErrInvalidType
	}

	valueType := innerTypeOf(ft.In(0), reflect.Ptr, reflect.Struct)
	if valueType == nil {
		return ErrInvalidType
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
func (d *Decoder) DisallowUnknownFields() { d.noUnknownFields = true }

// Errors returns the collected decoding errors if SkipInvalidRows is enabled.
func (d *Decoder) Errors() []error { return d.errors }

// SetCleanerFunc causes the Decoder to call fn on every field before type conversion.
// Use this to clean wrongly formatted values.
func (d *Decoder) SetCleanerFunc(fn func(column, value string) string) { d.cleanerFunc = fn }

// SetHeader causes the Decoder to use the specified slice as the CSV header
// instead of interpreting the first record as the header.
// Use this to read headerless CSVs.
func (d *Decoder) SetHeader(h []string) { d.header = h }

// SetReaderFunc customizes how records are parsed.
// The default value is NewReader.
func (d *Decoder) SetReaderFunc(fn func(io.Reader) Reader) { d.readerFunc = fn }

// SkipInvalidRows causes the Decoder to ignore errors when reading records.
// Any errors that occur are collected and can be retrieved by calling Errors.
func (d *Decoder) SkipInvalidRows() { d.skipInvalidRows = true }

// Unmarshal decodes a byte slice as a CSV to a slice of structs.
// The CSV is expected to be comma-separated and have a header.
func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}
