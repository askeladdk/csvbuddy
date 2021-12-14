package csvbuddy

import (
	"encoding"
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

type valueDecoder func(reflect.Value, string, *structField) error

func boolDecoder(v reflect.Value, s string, _ *structField) (err error) {
	var x bool
	if x, err = strconv.ParseBool(s); err == nil {
		v.SetBool(x)
	}
	return
}

func complexDecoder(bitSize int) valueDecoder {
	return func(v reflect.Value, s string, _ *structField) (err error) {
		var x complex128
		if x, err = strconv.ParseComplex(s, bitSize); err == nil {
			v.SetComplex(x)
		}
		return
	}
}

func floatDecoder(bitSize int) valueDecoder {
	return func(v reflect.Value, s string, _ *structField) (err error) {
		var x float64
		if x, err = strconv.ParseFloat(s, bitSize); err == nil {
			v.SetFloat(x)
		}
		return
	}
}

var (
	complex64Decoder  = complexDecoder(64)
	complex128Decoder = complexDecoder(128)
	float32Decoder    = floatDecoder(32)
	float64Decoder    = floatDecoder(64)
)

func intDecoder(v reflect.Value, s string, field *structField) (err error) {
	var x int64
	if x, err = strconv.ParseInt(s, field.Base, 64); err == nil {
		v.SetInt(x)
	}
	return
}

func ptrDecoder(decode valueDecoder) valueDecoder {
	return func(v reflect.Value, s string, field *structField) (err error) {
		if s == "" {
			return
		}
		for v.Type().Kind() == reflect.Ptr {
			t := v.Type().Elem()
			v.Set(reflect.New(t))
			v = v.Elem()
		}
		return decode(v, s, field)
	}
}

var (
	byteSlicePtrDecoder  = ptrDecoder(byteSliceDecoder)
	boolPtrDecoder       = ptrDecoder(boolDecoder)
	complex64PtrDecoder  = ptrDecoder(complex64Decoder)
	complex128PtrDecoder = ptrDecoder(complex128Decoder)
	float32PtrDecoder    = ptrDecoder(float32Decoder)
	float64PtrDecoder    = ptrDecoder(float64Decoder)
	intPtrDecoder        = ptrDecoder(intDecoder)
	stringPtrDecoder     = ptrDecoder(stringDecoder)
	textPtrDecoder       = ptrDecoder(textDecoder)
	uintPtrDecoder       = ptrDecoder(uintDecoder)
)

func stringDecoder(v reflect.Value, s string, _ *structField) (err error) {
	v.SetString(s)
	return
}

func uintDecoder(v reflect.Value, s string, field *structField) (err error) {
	var x uint64
	if x, err = strconv.ParseUint(s, field.Base, 64); err == nil {
		v.SetUint(x)
	}
	return
}

func byteSliceDecoder(v reflect.Value, s string, _ *structField) (err error) {
	v.SetBytes([]byte(s))
	return
}

func textDecoder(v reflect.Value, s string, _ *structField) (err error) {
	return v.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
}

func mapValueDecoder(t reflect.Type, name string) (valueDecoder, error) {
	if reflect.PtrTo(t).Implements(textUnmarshalerType) {
		return textDecoder, nil
	}

	switch t.Kind() {
	case reflect.Bool:
		return boolDecoder, nil
	case reflect.Complex64:
		return complex64Decoder, nil
	case reflect.Complex128:
		return complex128Decoder, nil
	case reflect.Float32:
		return float32Decoder, nil
	case reflect.Float64:
		return float64Decoder, nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return intDecoder, nil
	case reflect.Ptr:
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if reflect.PtrTo(t).Implements(textUnmarshalerType) {
			return textPtrDecoder, nil
		}
		switch t.Kind() {
		case reflect.Bool:
			return boolPtrDecoder, nil
		case reflect.Complex64:
			return complex64PtrDecoder, nil
		case reflect.Complex128:
			return complex128PtrDecoder, nil
		case reflect.Float32:
			return float32PtrDecoder, nil
		case reflect.Float64:
			return float64PtrDecoder, nil
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			return intPtrDecoder, nil
		case reflect.String:
			return stringPtrDecoder, nil
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			return uintPtrDecoder, nil
		}
		if t.ConvertibleTo(byteSliceType) {
			return byteSlicePtrDecoder, nil
		}
		return nil, fmt.Errorf("cannot decode field '%s'", name)
	case reflect.String:
		return stringDecoder, nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return uintDecoder, nil
	}

	if t.ConvertibleTo(byteSliceType) {
		return byteSliceDecoder, nil
	}

	return nil, fmt.Errorf("cannot decode field '%s'", name)
}

type valueEncoder func(reflect.Value, *structField) (string, error)

func boolEncoder(v reflect.Value, _ *structField) (string, error) {
	return strconv.FormatBool(v.Bool()), nil
}

func complexEncoder(bitSize int) valueEncoder {
	return func(v reflect.Value, field *structField) (string, error) {
		return strconv.FormatComplex(v.Complex(), 'f', -1, bitSize), nil
	}
}

func floatEncoder(bitSize int) valueEncoder {
	return func(v reflect.Value, field *structField) (string, error) {
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil
	}
}

var (
	complex64Encoder  = complexEncoder(64)
	complex128Encoder = complexEncoder(128)
	float32Encoder    = floatEncoder(32)
	float64Encoder    = floatEncoder(64)
)

func intEncoder(v reflect.Value, field *structField) (string, error) {
	return strconv.FormatInt(v.Int(), field.Base), nil
}

func ptrEncoder(encode valueEncoder) valueEncoder {
	return func(v reflect.Value, field *structField) (string, error) {
		for v.Type().Kind() == reflect.Ptr {
			if v.IsNil() {
				return "", nil
			}
			v = v.Elem()
		}
		return encode(v, field)
	}
}

var (
	byteSlicePtrEncoder  = ptrEncoder(byteSliceEncoder)
	boolPtrEncoder       = ptrEncoder(boolEncoder)
	complex64PtrEncoder  = ptrEncoder(complex64Encoder)
	complex128PtrEncoder = ptrEncoder(complex128Encoder)
	float32PtrEncoder    = ptrEncoder(float32Encoder)
	float64PtrEncoder    = ptrEncoder(float64Encoder)
	intPtrEncoder        = ptrEncoder(intEncoder)
	stringPtrEncoder     = ptrEncoder(stringEncoder)
	textPtrEncoder       = ptrEncoder(textEncoder)
	uintPtrEncoder       = ptrEncoder(uintEncoder)
)

func stringEncoder(v reflect.Value, _ *structField) (string, error) {
	return v.String(), nil
}

func uintEncoder(v reflect.Value, field *structField) (string, error) {
	return strconv.FormatUint(v.Uint(), field.Base), nil
}

func byteSliceEncoder(v reflect.Value, _ *structField) (string, error) {
	return string(v.Bytes()), nil
}

func textEncoder(v reflect.Value, _ *structField) (s string, err error) {
	text, err := v.Addr().Interface().(encoding.TextMarshaler).MarshalText()
	return string(text), err
}

func mapValueEncoder(t reflect.Type, name string) (valueEncoder, error) {
	if reflect.PtrTo(t).Implements(textMarshalerType) {
		return textEncoder, nil
	}

	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder, nil
	case reflect.Complex64:
		return complex64Encoder, nil
	case reflect.Complex128:
		return complex128Encoder, nil
	case reflect.Float32:
		return float32Encoder, nil
	case reflect.Float64:
		return float64Encoder, nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return intEncoder, nil
	case reflect.Ptr:
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if reflect.PtrTo(t).Implements(textMarshalerType) {
			return textPtrEncoder, nil
		}
		switch t.Kind() {
		case reflect.Bool:
			return boolPtrEncoder, nil
		case reflect.Complex64:
			return complex64PtrEncoder, nil
		case reflect.Complex128:
			return complex128PtrEncoder, nil
		case reflect.Float32:
			return float32PtrEncoder, nil
		case reflect.Float64:
			return float64PtrEncoder, nil
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			return intPtrEncoder, nil
		case reflect.String:
			return stringPtrEncoder, nil
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			return uintPtrEncoder, nil
		}
		if t.ConvertibleTo(byteSliceType) {
			return byteSlicePtrEncoder, nil
		}
		return nil, fmt.Errorf("cannot encode field '%s'", name)
	case reflect.String:
		return stringEncoder, nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return uintEncoder, nil
	}

	if t.ConvertibleTo(byteSliceType) {
		return byteSliceEncoder, nil
	}

	return nil, fmt.Errorf("cannot encode field '%s'", name)
}

type structField struct {
	Index  []int
	Name   string
	Base   int
	Decode valueDecoder
	Encode valueEncoder
}

func valueOf(i interface{}) (v reflect.Value, err error) {
	if i == nil {
		err = ErrInvalidArgument
	} else if v = reflect.ValueOf(i); v.IsNil() {
		err = ErrInvalidArgument
	}
	return
}

func parseTag(tag string) (name string, base int) {
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
		// parse integer base base=N
		if strings.HasPrefix(val, "base=") {
			base, _ = strconv.Atoi(val[5:])
		}
	}
	return
}

func structFieldsOf(t reflect.Type) ([]structField, error) {
	structLock.Lock()
	defer structLock.Unlock()
	if fields, exists := structCache[t]; exists {
		return fields, nil
	}

	var fields []structField
	names := map[string]struct{}{}

	for i := 0; i < t.NumField(); i++ {
		if field := t.Field(i); field.IsExported() {
			var name string
			var base int
			if tag, ok := field.Tag.Lookup("csv"); ok {
				name, base = parseTag(tag)
			}
			if name == "" {
				name = field.Name
			}
			if base == 0 {
				base = 10
			}
			if name != "-" {
				if _, exists := names[name]; exists {
					return nil, fmt.Errorf("duplicate field name '%s'", name)
				}
				decoder, err := mapValueDecoder(field.Type, name)
				if err != nil {
					return nil, err
				}
				encoder, err := mapValueEncoder(field.Type, name)
				if err != nil {
					return nil, err
				}
				fields = append(fields, structField{
					Index:  field.Index,
					Name:   name,
					Base:   base,
					Decode: decoder,
					Encode: encoder,
				})
				names[name] = struct{}{}
			}
		}
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
	if fields, err := structFieldsOf(t); err != nil {
		return nil, err
	} else {
		header := make([]string, len(fields))
		for i, field := range fields {
			header[i] = field.Name
		}
		return header, nil
	}
}
