package csvbuddy

import (
	"encoding"
	"encoding/csv"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type uppercase string

func (u *uppercase) UnmarshalText(b []byte) error {
	*(*string)(u) = strings.ToUpper(string(b))
	return nil
}

var _ encoding.TextUnmarshaler = (*uppercase)(nil)

type teststruc struct {
	A []byte     `csv:"bytes"`
	B bool       `csv:"bool"`
	C complex128 `csv:"complex"`
	F float64    `csv:"float"`
	I int        `csv:"int"`
	O *int       `csv:"optional"`
	S string     `csv:"string"`
	T uppercase  `csv:"uppercase"`
	U uint       `csv:"uint"`
}

func TestDecode(t *testing.T) {
	testdata := strings.Join([]string{
		"bool,bytes,complex,float,int,optional,string,uint,uppercase",
		"true,hello,1+1i,3.1415,-173,0,hello world,1337,gopher",
	}, "\n")

	var data []teststruc
	if err := Unmarshal([]byte(testdata), &data); err != nil {
		t.Fatal(err)
	}

	expect := []teststruc{
		{[]byte("hello"), true, 1 + 1i, 3.1415, -173, new(int), "hello world", "GOPHER", 1337},
	}

	if !reflect.DeepEqual(data, expect) {
		t.Error("should be equal")
	}
}

func TestDecodeFunc(t *testing.T) {
	testdata := strings.Join([]string{
		"bool,bytes,complex,float,int,optional,string,uint,uppercase",
		"true,hello,1+1i,3.1415,-173,0,hello world,1337,gopher",
	}, "\n")

	var data []teststruc
	if err := NewDecoder(strings.NewReader(testdata)).DecodeFunc(func(s *teststruc) error {
		data = append(data, *s)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	expect := []teststruc{
		{[]byte("hello"), true, 1 + 1i, 3.1415, -173, new(int), "hello world", "GOPHER", 1337},
	}

	if !reflect.DeepEqual(data, expect) {
		t.Error("should be equal")
	}
}

func TestDecodeHeaderless(t *testing.T) {
	testdata := "true,hello,1+1i,3.1415,-173,0,hello world,1337,gopher"

	var data []teststruc

	d := NewDecoder(strings.NewReader(testdata))
	d.SetHeader(strings.Split("bool,bytes,complex,float,int,optional,string,uint,uppercase", ","))

	if err := d.Decode(&data); err != nil {
		t.Fatal(err)
	}

	expect := []teststruc{
		{[]byte("hello"), true, 1 + 1i, 3.1415, -173, new(int), "hello world", "GOPHER", 1337},
	}

	if !reflect.DeepEqual(data, expect) {
		t.Error("should be equal")
	}
}

func TestDecodeDisallowUnknownFields(t *testing.T) {
	type struc struct {
		B bool
		S string
	}

	testdata := "true,hello,1"

	var data []struc

	d := NewDecoder(strings.NewReader(testdata))
	d.DisallowUnknownFields()
	d.SetHeader(MustHeader(struc{}))

	if err := d.Decode(&data); !errors.Is(err, csv.ErrFieldCount) {
		t.Fatal("should be error")
	}
}

func TestDecodeSkipInvalidRows(t *testing.T) {
	type struc struct {
		B bool
		S string
	}

	testdata := "2,hello\nfalse,world"

	var data []struc

	d := NewDecoder(strings.NewReader(testdata))
	d.SkipInvalidRows()
	d.SetHeader(MustHeader(struc{}))

	if err := d.Decode(&data); err != nil {
		t.Fatal(err)
	}

	expect := []struc{{false, "world"}}
	if !reflect.DeepEqual(data, expect) {
		t.Fatal("not equal")
	}

	if len(d.Errors()) != 1 || !errors.Is(d.Errors()[0], strconv.ErrSyntax) {
		t.Fatal("expected NumError")
	}
}

func TestDecodeTooFewFields(t *testing.T) {
	type struc struct {
		B bool
		S string
	}

	testdata := "true"

	var data []struc

	d := NewDecoder(strings.NewReader(testdata))
	d.SetHeader(MustHeader(struc{}))

	if err := d.Decode(&data); !errors.Is(err, csv.ErrFieldCount) {
		t.Fatal("should be error")
	}
}
