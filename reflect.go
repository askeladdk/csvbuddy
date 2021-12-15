package csvbuddy

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var (
	byteSliceType       = reflect.TypeOf((*[]byte)(nil)).Elem()
	errorType           = reflect.TypeOf((*error)(nil)).Elem()
	textMarshalerType   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	structCache = map[reflect.Type][]structField{}
	structLock  sync.Mutex
)

type converter interface {
	Decode(reflect.Value, string) error
	Encode(reflect.Value) (string, error)
}

type boolCodec struct{}

func (b *boolCodec) Decode(v reflect.Value, s string) (err error) {
	var x bool
	if x, err = strconv.ParseBool(s); err == nil {
		v.SetBool(x)
	}
	return
}

func (b *boolCodec) Encode(v reflect.Value) (string, error) {
	return strconv.FormatBool(v.Bool()), nil
}

type complexCodec struct {
	BitSize int
	Prec    int
	Fmt     byte
}

func (c *complexCodec) Decode(v reflect.Value, s string) (err error) {
	var x complex128
	if x, err = strconv.ParseComplex(s, c.BitSize); err == nil {
		v.SetComplex(x)
	}
	return
}

func (c *complexCodec) Encode(v reflect.Value) (string, error) {
	return strconv.FormatComplex(v.Complex(), c.Fmt, c.Prec, c.BitSize), nil
}

type floatCodec struct {
	BitSize int
	Prec    int
	Fmt     byte
}

func (c *floatCodec) Decode(v reflect.Value, s string) (err error) {
	var x float64
	if x, err = strconv.ParseFloat(s, c.BitSize); err == nil {
		v.SetFloat(x)
	}
	return
}

func (c *floatCodec) Encode(v reflect.Value) (string, error) {
	return strconv.FormatFloat(v.Float(), c.Fmt, c.Prec, c.BitSize), nil
}

type intCodec struct {
	BitSize int
	Base    int
}

func (c *intCodec) Decode(v reflect.Value, s string) (err error) {
	var x int64
	if x, err = strconv.ParseInt(s, c.Base, c.BitSize); err == nil {
		v.SetInt(x)
	}
	return
}

func (c *intCodec) Encode(v reflect.Value) (string, error) {
	return strconv.FormatInt(v.Int(), c.Base), nil
}

type uintCodec struct {
	BitSize int
	Base    int
}

func (c *uintCodec) Decode(v reflect.Value, s string) (err error) {
	var x uint64
	if x, err = strconv.ParseUint(s, c.Base, c.BitSize); err == nil {
		v.SetUint(x)
	}
	return
}

func (c *uintCodec) Encode(v reflect.Value) (string, error) {
	return strconv.FormatUint(v.Uint(), c.Base), nil
}

type stringCodec struct{}

func (c *stringCodec) Decode(v reflect.Value, s string) (err error) {
	v.SetString(s)
	return
}

func (c *stringCodec) Encode(v reflect.Value) (string, error) {
	return v.String(), nil
}

type textCodec struct{}

func (c *textCodec) Decode(v reflect.Value, s string) (err error) {
	if tu, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
		return tu.UnmarshalText([]byte(s))
	}
	return errors.New("value does not implement encoding.TextUnmarshaler")
}

func (c *textCodec) Encode(v reflect.Value) (s string, err error) {
	if tm, ok := v.Addr().Interface().(encoding.TextMarshaler); ok {
		text, err := tm.MarshalText()
		return string(text), err
	}
	return "", errors.New("value does not implement encoding.TextMarshaler")
}

type byteSliceCodec struct{}

func (c *byteSliceCodec) Decode(v reflect.Value, s string) (err error) {
	v.SetBytes([]byte(s))
	return
}

func (c *byteSliceCodec) Encode(v reflect.Value) (string, error) {
	return string(v.Bytes()), nil
}

type ptrCodec struct {
	converter
}

func (c *ptrCodec) Decode(v reflect.Value, s string) (err error) {
	if s == "" {
		return
	}
	for v.Type().Kind() == reflect.Ptr {
		t := v.Type().Elem()
		v.Set(reflect.New(t))
		v = v.Elem()
	}
	return c.converter.Decode(v, s)
}

func (c *ptrCodec) Encode(v reflect.Value) (string, error) {
	for v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	return c.converter.Encode(v)
}

func implementsTextMarshaler(t reflect.Type) bool {
	ptr := reflect.PtrTo(t)
	return ptr.Implements(textUnmarshalerType) || ptr.Implements(textMarshalerType)
}

func newValueConverter(t reflect.Type, name string, base, prec int, ffmt byte) (converter, error) {
	if implementsTextMarshaler(t) {
		return &textCodec{}, nil
	}

	switch t.Kind() {
	case reflect.Bool:
		return &boolCodec{}, nil
	case reflect.Complex64:
		return &complexCodec{64, prec, ffmt}, nil
	case reflect.Complex128:
		return &complexCodec{128, prec, ffmt}, nil
	case reflect.Float32:
		return &floatCodec{32, prec, ffmt}, nil
	case reflect.Float64:
		return &floatCodec{64, prec, ffmt}, nil
	case reflect.Int8:
		return &intCodec{8, base}, nil
	case reflect.Int16:
		return &intCodec{16, base}, nil
	case reflect.Int32:
		return &intCodec{32, base}, nil
	case reflect.Int64, reflect.Int:
		return &intCodec{64, base}, nil
	case reflect.Ptr:
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if implementsTextMarshaler(t) {
			return &ptrCodec{&textCodec{}}, nil
		}
		switch t.Kind() {
		case reflect.Bool:
			return &ptrCodec{&boolCodec{}}, nil
		case reflect.Complex64:
			return &ptrCodec{&complexCodec{64, prec, ffmt}}, nil
		case reflect.Complex128:
			return &ptrCodec{&complexCodec{128, prec, ffmt}}, nil
		case reflect.Float32:
			return &ptrCodec{&floatCodec{32, prec, ffmt}}, nil
		case reflect.Float64:
			return &ptrCodec{&floatCodec{64, prec, ffmt}}, nil
		case reflect.Int8:
			return &ptrCodec{&intCodec{8, base}}, nil
		case reflect.Int16:
			return &ptrCodec{&intCodec{16, base}}, nil
		case reflect.Int32:
			return &ptrCodec{&intCodec{32, base}}, nil
		case reflect.Int64, reflect.Int:
			return &ptrCodec{&intCodec{64, base}}, nil
		case reflect.String:
			return &ptrCodec{&stringCodec{}}, nil
		case reflect.Uint8:
			return &ptrCodec{&uintCodec{8, base}}, nil
		case reflect.Uint16:
			return &ptrCodec{&uintCodec{16, base}}, nil
		case reflect.Uint32:
			return &ptrCodec{&uintCodec{32, base}}, nil
		case reflect.Uint64, reflect.Uint:
			return &ptrCodec{&uintCodec{64, base}}, nil
		}
		if t.ConvertibleTo(byteSliceType) {
			return &ptrCodec{&byteSliceCodec{}}, nil
		}
		return nil, fmt.Errorf("cannot decode field '%s'", name)
	case reflect.String:
		return &stringCodec{}, nil
	case reflect.Uint8:
		return &uintCodec{8, base}, nil
	case reflect.Uint16:
		return &uintCodec{16, base}, nil
	case reflect.Uint32:
		return &uintCodec{32, base}, nil
	case reflect.Uint64, reflect.Uint:
		return &uintCodec{64, base}, nil
	}

	if t.ConvertibleTo(byteSliceType) {
		return &byteSliceCodec{}, nil
	}

	return nil, fmt.Errorf("cannot decode field '%s'", name)
}

type structField struct {
	Index     []int  // struct field index
	Name      string // column name
	converter        // value converter
}

func valueOf(i interface{}) (v reflect.Value, err error) {
	if i == nil {
		err = ErrInvalidArgument
	} else if v = reflect.ValueOf(i); v.IsNil() {
		err = ErrInvalidArgument
	}
	return
}

func parseTag(tag string) (name string, base, prec int, fmt byte) {
	base, prec, fmt = 10, -1, 'f'
	// parse the name
	if i := strings.IndexByte(tag, ','); i == -1 {
		name = tag
		return
	} else {
		name, tag = tag[:i], tag[i+1:]
	}
	// parse the other parameters
	var val string
	for tag != "" {
		if i := strings.IndexByte(tag, ','); i == -1 {
			val, tag = tag, ""
		} else {
			val, tag = tag[:i], tag[i+1:]
		}
		switch {
		case strings.HasPrefix(val, "base="): // integer base
			if n, err := strconv.Atoi(val[5:]); err == nil {
				base = n
			}
		case strings.HasPrefix(val, "prec="): // floating point precision
			if n, err := strconv.Atoi(val[5:]); err == nil {
				prec = n
			}
		case strings.HasPrefix(val, "fmt="): // floating point format
			if len(val) >= 4 {
				if c := val[4]; strings.IndexByte("beEfgGxX", c) >= 0 {
					fmt = c
				}
			}
		}
	}
	return
}

func appendStructFields(t reflect.Type, index []int, fields *[]structField, names *map[string]struct{}) error {
	for i := 0; i < t.NumField(); i++ {
		if field := t.Field(i); field.IsExported() {
			// check for inline struct
			if field.Type.Kind() == reflect.Struct {
				if field.Anonymous || strings.Contains(field.Tag.Get("csv"), ",inline") {
					if err := appendStructFields(field.Type, append(append([]int{}, index...), i), fields, names); err != nil {
						return err
					}
					continue
				}
			}

			var name string
			var base, prec int
			var ffmt byte
			base, prec, ffmt = 10, -1, 'f'
			if tag, ok := field.Tag.Lookup("csv"); ok {
				name, base, prec, ffmt = parseTag(tag)
			}
			if name == "" {
				name = field.Name
			}
			if name != "-" {
				if _, exists := (*names)[name]; exists {
					return fmt.Errorf("duplicate field name '%s'", name)
				}
				codec, err := newValueConverter(field.Type, name, base, prec, ffmt)
				if err != nil {
					return err
				}
				*fields = append(*fields, structField{
					Index:     append(append([]int{}, index...), field.Index...),
					Name:      name,
					converter: codec,
				})
				(*names)[name] = struct{}{}
			}
		}
	}
	return nil
}

func structFieldsOf(t reflect.Type) ([]structField, error) {
	structLock.Lock()
	defer structLock.Unlock()
	if fields, exists := structCache[t]; exists {
		return fields, nil
	}

	var fields []structField
	names := map[string]struct{}{}

	if err := appendStructFields(t, nil, &fields, &names); err != nil {
		return nil, err
	}

	structCache[t] = fields
	return fields, nil
}

func innerTypeOf(t reflect.Type, kinds ...reflect.Kind) reflect.Type {
	for i := 0; i < len(kinds)-1; i++ {
		if t.Kind() != kinds[i] {
			return nil
		}
		t = t.Elem()
	}
	if t.Kind() != kinds[len(kinds)-1] {
		return nil
	}
	return t
}

func headerIndices(header []string, fields []structField) (indices []int, err error) {
	// check for duplicate header names
	names := map[string]struct{}{}
	for _, h := range header {
		if _, exists := names[h]; exists {
			return nil, fmt.Errorf("duplicate header name '%s'", h)
		}
		names[h] = struct{}{}
	}

	// for every column in header, find index of struct field with matching name
	for i, col := range header {
		for j, field := range fields {
			if field.Name == col {
				indices = append(indices, i, j)
			}
		}
	}

	return
}

func headerOf(t reflect.Type) ([]string, error) {
	fields, err := structFieldsOf(t)
	if err != nil {
		return nil, err
	}
	header := make([]string, len(fields))
	for i, field := range fields {
		header[i] = field.Name
	}
	return header, nil
}

// Header returns the header of v, which must be a pointer to a slice of structs.
func Header(v interface{}) ([]string, error) {
	t := innerTypeOf(reflect.TypeOf(v), reflect.Ptr, reflect.Slice, reflect.Struct)
	return headerOf(t)
}
