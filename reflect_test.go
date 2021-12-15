package csvbuddy

import (
	"reflect"
	"testing"
)

func ptrTo(i interface{}) uintptr {
	return reflect.ValueOf(i).Pointer()
}

func checkValueOf(t *testing.T, have interface{}, expect error) {
	if _, err := valueOf(have); err != expect {
		t.Error(have, err)
	}
}

func TestValueOf(t *testing.T) {
	checkValueOf(t, nil, ErrInvalidArgument)
	checkValueOf(t, (*int)(nil), ErrInvalidArgument)
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
			// Decode: intDecoder,
			Name: "a",
			Base: 10,
			Prec: -1,
			Fmt:  'f',
		},
		{
			Index: []int{2},
			// Decode: textDecoder,
			Name: "C",
			Base: 10,
			Prec: -1,
			Fmt:  'f',
		},
	}

	fields, err := structFieldsOf(reflect.TypeOf(x))
	if err != nil {
		t.Error(err)
	}

	// stupid workaround for DeepEqual being unable to compare funcs
	// even though they are comparable by address
	if ptrTo(fields[0].Decode) != ptrTo(int64Decoder) {
		t.Fatal("not equal")
	} else if ptrTo(fields[1].Decode) != ptrTo(int64PtrDecoder) {
		t.Fatal("not equal")
	} else if ptrTo(fields[0].Encode) != ptrTo(intEncoder) {
		t.Fatal("not equal")
	} else if ptrTo(fields[1].Encode) != ptrTo(intPtrEncoder) {
		t.Fatal("not equal")
	}
	fields[0].Decode = nil
	fields[1].Decode = nil
	fields[0].Encode = nil
	fields[1].Encode = nil

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

func TestValueDecoders(t *testing.T) {
	bsval := []byte("byteSlice")
	bval := true
	cval := 1 - 2i
	fval := 3.14159
	ival := -1337
	uval := uint(1337)
	sval := "hello world"
	tval := uppercase("HELLO WORLD")
	testCases := []struct {
		String   string
		Type     reflect.Type
		Expected interface{}
	}{
		{"byteSlice", byteSliceType, []byte("byteSlice")},
		{"true", reflect.TypeOf(true), true},
		{"1-2.3i", reflect.TypeOf(complex128(0)), 1 - 2.3i},
		{"1-2.3i", reflect.TypeOf(complex64(0)), complex64(1 - 2.3i)},
		{"3.14159", reflect.TypeOf(float32(0)), float32(3.14159)},
		{"3.14159", reflect.TypeOf(float64(0)), 3.14159},
		{"-42", reflect.TypeOf(int8(0)), int8(-42)},
		{"-1337", reflect.TypeOf(int16(0)), int16(-1337)},
		{"-1337", reflect.TypeOf(int32(0)), int32(-1337)},
		{"-1337", reflect.TypeOf(int64(0)), int64(-1337)},
		{"-1337", reflect.TypeOf(int(0)), int(-1337)},
		{"hello world", reflect.TypeOf(""), "hello world"},
		{"hello world", reflect.TypeOf(uppercase("")), uppercase("HELLO WORLD")},
		{"42", reflect.TypeOf(uint8(0)), uint8(42)},
		{"1337", reflect.TypeOf(uint16(0)), uint16(1337)},
		{"1337", reflect.TypeOf(uint32(0)), uint32(1337)},
		{"1337", reflect.TypeOf(uint64(0)), uint64(1337)},
		{"1337", reflect.TypeOf(uint(0)), uint(1337)},
		// optional types
		{"byteSlice", reflect.PtrTo(byteSliceType), &bsval},
		{"true", reflect.TypeOf((*bool)(nil)), &bval},
		{"1-2i", reflect.TypeOf((*complex128)(nil)), &cval},
		{"3.14159", reflect.TypeOf((*float64)(nil)), &fval},
		{"-1337", reflect.TypeOf((*int)(nil)), &ival},
		{"hello world", reflect.TypeOf((*string)(nil)), &sval},
		{"hello world", reflect.TypeOf((*uppercase)(nil)), &tval},
		{"1337", reflect.TypeOf((*uint)(nil)), &uval},
	}

	field := structField{Base: 10, Prec: -1, Fmt: 'f'}
	for _, testCase := range testCases {
		v := reflect.New(testCase.Type)
		if decoder, err := mapValueDecoder(testCase.Type, ""); err != nil {
			t.Error(testCase.String, err)
		} else if err := decoder(v.Elem(), testCase.String, &field); err != nil {
			t.Error(testCase.String, err)
		} else if !reflect.DeepEqual(v.Elem().Interface(), testCase.Expected) {
			t.Error(testCase.String, "not equal", v.Elem(), testCase.Expected)
		}
	}
}

func TestValueEncoders(t *testing.T) {
	testCases := []struct {
		Expected interface{}
		Type     reflect.Type
		Value    interface{}
	}{
		{"byteSlice", byteSliceType, []byte("byteSlice")},
		{"true", reflect.TypeOf(true), true},
		{"(1-2.3i)", reflect.TypeOf(complex128(0)), 1 - 2.3i},
		{"(1-2.3i)", reflect.TypeOf(complex64(0)), complex64(1 - 2.3i)},
		{"3.14159", reflect.TypeOf(float32(0)), float32(3.14159)},
		{"3.14159", reflect.TypeOf(float64(0)), 3.14159},
		{"-57", reflect.TypeOf(int8(0)), int8(-57)},
		{"-1337", reflect.TypeOf(int16(0)), int16(-1337)},
		{"-1337", reflect.TypeOf(int32(0)), int32(-1337)},
		{"-1337", reflect.TypeOf(int64(0)), int64(-1337)},
		{"-1337", reflect.TypeOf(int(0)), int(-1337)},
		{"hello world", reflect.TypeOf(""), "hello world"},
		// {"hello world", reflect.TypeOf(uppercase("")), &tval},
		{"57", reflect.TypeOf(uint8(0)), uint8(57)},
		{"1337", reflect.TypeOf(uint16(0)), uint16(1337)},
		{"1337", reflect.TypeOf(uint32(0)), uint32(1337)},
		{"1337", reflect.TypeOf(uint64(0)), uint64(1337)},
		{"1337", reflect.TypeOf(uint(0)), uint(1337)},
		// optional types
		{"byteSlice", reflect.PtrTo(byteSliceType), []byte("byteSlice")},
		{"true", reflect.TypeOf((*bool)(nil)), true},
		{"(1-2.3i)", reflect.TypeOf((*complex128)(nil)), 1 - 2.3i},
		{"3.14159", reflect.TypeOf((*float64)(nil)), 3.14159},
		{"-1337", reflect.TypeOf((*int)(nil)), -1337},
		{"hello world", reflect.TypeOf((*string)(nil)), "hello world"},
		// {"HELLO WORLD", reflect.TypeOf((*uppercase)(nil)), uppercase("HELLO WORLD")},
		{"1337", reflect.TypeOf((*uint)(nil)), uint(1337)},
	}

	field := structField{Base: 10, Prec: -1, Fmt: 'f'}
	for _, testCase := range testCases {
		if encoder, err := mapValueEncoder(testCase.Type, ""); err != nil {
			t.Error(testCase.Type, err)
		} else if val, err := encoder(reflect.ValueOf(testCase.Value), &field); err != nil {
			t.Error(testCase.Type, err)
		} else if val != testCase.Expected {
			t.Error(testCase.Type, "not equal", val, testCase.Expected)
		}
	}
}

func TestParseTag(t *testing.T) {
	testcases := []struct {
		Tag  string
		Name string
		Base int
		Prec int
		Fmt  byte
	}{
		{"", "", 10, -1, 'f'},
		{",", "", 10, -1, 'f'},
		{"my_field", "my_field", 10, -1, 'f'},
		{",base=5", "", 5, -1, 'f'},
		{"-,base=3", "-", 3, -1, 'f'},
		{"my_field,base=16", "my_field", 16, -1, 'f'},
		{"my_field,base=16,base=2", "my_field", 2, -1, 'f'},
		{"my_field,something,,base=", "my_field", 10, -1, 'f'},
		{"my_field,prec=5,fmt=G", "my_field", 10, 5, 'G'},
	}
	for _, testcase := range testcases {
		name, base, prec, ffmt := parseTag(testcase.Tag)
		if name != testcase.Name {
			t.Error(testcase.Tag, name, "!=", testcase.Name)
		}
		if base != testcase.Base {
			t.Error(testcase.Tag, base, "!=", testcase.Base)
		}
		if prec != testcase.Prec {
			t.Error(testcase.Tag, prec, "!=", testcase.Prec)
		}
		if ffmt != testcase.Fmt {
			t.Error(testcase.Tag, ffmt, "!=", testcase.Fmt)
		}
	}
}

func TestEmbeddedStruct(t *testing.T) {
	type A struct{ A int }
	type B struct{ B int }

	numbers := []struct {
		A
		B B `csv:",inline"`
		C int
	}{
		{A: A{A: 1101}, B: B{B: -100}, C: 1},
		{A: A{A: 1001}, B: B{B: -200}, C: 2},
	}
	text, err := Marshal(&numbers)
	if err != nil {
		t.Fatal(err)
		return
	}
	if string(text) != "A,B,C\n1101,-100,1\n1001,-200,2\n" {
		t.Fatal(string(text))
	}
}
