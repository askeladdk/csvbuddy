package csvbuddy

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func checkKindOf(t *testing.T, have reflect.Type, expect reflect.Kind) {
	if k := kindOf(have); k != expect {
		t.Error(have, k)
	}
}

func checkValueOf(t *testing.T, have interface{}, expect error) {
	if _, err := valueOf(have); err != expect {
		t.Error(have, err)
	}
}

func TestKindOf(t *testing.T) {
	type myInt int

	checkKindOf(t, reflect.TypeOf(int(0)), reflect.Int64)
	checkKindOf(t, reflect.TypeOf(int8(0)), reflect.Int64)
	checkKindOf(t, reflect.TypeOf(int16(0)), reflect.Int64)
	checkKindOf(t, reflect.TypeOf(int32(0)), reflect.Int64)
	checkKindOf(t, reflect.TypeOf(int64(0)), reflect.Int64)
	checkKindOf(t, reflect.TypeOf(myInt(0)), reflect.Int64)

	checkKindOf(t, reflect.TypeOf(byte(0)), reflect.Uint64)
	checkKindOf(t, reflect.TypeOf(uint(0)), reflect.Uint64)
	checkKindOf(t, reflect.TypeOf(uint8(0)), reflect.Uint64)
	checkKindOf(t, reflect.TypeOf(uint16(0)), reflect.Uint64)
	checkKindOf(t, reflect.TypeOf(uint32(0)), reflect.Uint64)
	checkKindOf(t, reflect.TypeOf(uint64(0)), reflect.Uint64)

	checkKindOf(t, reflect.TypeOf(true), reflect.Bool)
	checkKindOf(t, reflect.TypeOf(false), reflect.Bool)

	checkKindOf(t, reflect.TypeOf(complex64(0i)), reflect.Complex128)
	checkKindOf(t, reflect.TypeOf(complex128(0i)), reflect.Complex128)

	checkKindOf(t, reflect.TypeOf(float32(0)), reflect.Float64)
	checkKindOf(t, reflect.TypeOf(float64(0)), reflect.Float64)

	checkKindOf(t, reflect.TypeOf(""), reflect.String)

	checkKindOf(t, reflect.TypeOf((*int)(nil)), reflect.Ptr)
	checkKindOf(t, reflect.TypeOf((*error)(nil)), reflect.Invalid)

	checkKindOf(t, reflect.TypeOf([]byte{}), byteSlice)
	checkKindOf(t, reflect.TypeOf(json.RawMessage{}), byteSlice)

	checkKindOf(t, reflect.TypeOf(time.Time{}), textUnmarshaler)
}

func TestValueOf(t *testing.T) {
	checkValueOf(t, nil, ErrInvalidType)
	checkValueOf(t, (*int)(nil), ErrInvalidType)
	checkValueOf(t, new(int), nil)
}

func TestInnerTypeOf(t *testing.T) {
	var p1 *[]int

	if tp := innerTypeOf(reflect.TypeOf(p1), reflect.Ptr, reflect.Slice, reflect.Int); tp == nil {
		t.Error("p1")
	}

	var p2 *int
	if tp := innerTypeOf(reflect.TypeOf(p2), reflect.Ptr, reflect.Slice, reflect.Int); tp != nil {
		t.Error("p2")
	}

	var p3 *[]int32
	if tp := innerTypeOf(reflect.TypeOf(p3), reflect.Ptr, reflect.Slice, reflect.Int); tp != nil {
		t.Error("p3")
	}
}

func TestStructFieldsOf(t *testing.T) {
	var x struct {
		A int `csv:"a"`
		b int
		C *int
		D int `csv:"-"`
	}

	expected := []structField{
		{
			Index: []int{0},
			Kind:  reflect.Int64,
			Name:  "a",
		},
		{
			Index: []int{2},
			Kind:  reflect.Ptr,
			Name:  "C",
		},
	}

	fields, err := structFieldsOf(reflect.TypeOf(x))
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, fields) {
		t.Error("not equal")
	}
}

func TestStructFieldsOfDup(t *testing.T) {
	var x struct {
		A int  `csv:"a"`
		B *int `csv:"a"`
	}

	if _, err := structFieldsOf(reflect.TypeOf(x)); err == nil {
		t.Error("should detect duplicate field name")
	}
}

func TestStructFieldsOfWrongType(t *testing.T) {
	var x struct {
		A error
	}

	if _, err := structFieldsOf(reflect.TypeOf(x)); err == nil {
		t.Error("should detect invalid field type")
	}
}

func TestHeader(t *testing.T) {
	var x struct {
		A int `csv:"a"`
		b int
		C *int
		D int `csv:"-"`
	}

	header := MustHeader(x)
	expected := []string{"a", "C"}
	if !reflect.DeepEqual(header, expected) {
		t.Error("should be equal")
	}

	if _, err := Header(0); err != ErrInvalidType {
		t.Error("should error")
	}
}
