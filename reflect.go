package csvbuddy

import (
	"encoding"
	"fmt"
	"reflect"
	"sync"
)

var (
	byteSliceType       = reflect.TypeOf((*[]byte)(nil)).Elem()
	errorType           = reflect.TypeOf((*error)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	structCache = map[reflect.Type][]structField{}
	structLock  sync.Mutex
)

const (
	byteSlice reflect.Kind = 1<<16 + iota
	textUnmarshaler
)

type structField struct {
	Index []int
	Kind  reflect.Kind
	Name  string
}

func kindOf(t reflect.Type) reflect.Kind {
	if reflect.PtrTo(t).Implements(textUnmarshalerType) {
		return textUnmarshaler
	}

	switch k := t.Kind(); k {
	case reflect.Bool, reflect.String:
		return k
	case reflect.Complex64, reflect.Complex128:
		return reflect.Complex128
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return reflect.Int64
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return reflect.Uint64
	case reflect.Float32, reflect.Float64:
		return reflect.Float64
	case reflect.Ptr:
		if kindOf(t.Elem()) != reflect.Invalid {
			return reflect.Ptr
		}
	}

	if t.ConvertibleTo(byteSliceType) {
		return byteSlice
	}

	return reflect.Invalid
}

func valueOf(v interface{}) (vv reflect.Value, err error) {
	if v == nil {
		err = ErrInvalidType
	} else if vv = reflect.ValueOf(v); vv.IsNil() {
		err = ErrInvalidType
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
			name := field.Name
			if tag, ok := field.Tag.Lookup("csv"); ok {
				name = tag
			}
			if name != "-" {
				if _, exists := names[name]; exists {
					return nil, fmt.Errorf("duplicate field name '%s'", name)
				}
				kind := kindOf(field.Type)
				if kind == reflect.Invalid {
					return nil, fmt.Errorf("cannot decode field of type '%s'", field.Type)
				}
				fields = append(fields, structField{
					Index: field.Index,
					Kind:  kind,
					Name:  name,
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

// Header inspects a struct and returns a header based on its fields which can be used by Decoder.
// An error is returned if a struct field is of an unsuitable type.
//
//  type A struct {
//    // some fields
//  }
//  header, err := Header(A{})
func Header(v interface{}) ([]string, error) {
	if t := reflect.TypeOf(v); t.Kind() != reflect.Struct {
		return nil, ErrInvalidType
	} else if fields, err := structFieldsOf(t); err != nil {
		return nil, err
	} else {
		header := make([]string, len(fields))
		for i, field := range fields {
			header[i] = field.Name
		}
		return header, nil
	}
}

// MustHeader is like Header except it panics if there is an error.
func MustHeader(v interface{}) []string {
	header, err := Header(v)
	if err != nil {
		panic(err.Error())
	}
	return header
}
