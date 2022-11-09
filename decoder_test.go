package csvbuddy

import (
	"encoding"
	"encoding/csv"
	"errors"
	"math"
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

func (u *uppercase) MarshalText() ([]byte, error) {
	return []byte(*u), nil
}

var _ encoding.TextUnmarshaler = (*uppercase)(nil)
var _ encoding.TextMarshaler = (*uppercase)(nil)

type testStruct struct {
	A []byte     `csv:"bytes"`
	B bool       `csv:"bool"`
	C complex128 `csv:"complex"`
	F float64    `csv:"float"`
	I int        `csv:"int"`
	O *int       `csv:"optional"`
	S string     `csv:"string"`
	T uppercase  `csv:"uppercase"`
	U uint       `csv:"uint,base=16"`
}

func TestDecode(t *testing.T) {
	testdata := strings.Join([]string{
		"bool,bytes,complex,float,int,optional,string,uint,uppercase",
		"true,hello,1+1i,3.1415,-173,0,hello world,1337,gopher",
	}, "\n")

	var data []testStruct
	if err := Unmarshal([]byte(testdata), &data); err != nil {
		t.Fatal(err)
	}

	expect := []testStruct{
		{[]byte("hello"), true, 1 + 1i, 3.1415, -173, new(int), "hello world", "GOPHER", 4919},
	}

	if !reflect.DeepEqual(data, expect) {
		t.Error("should be equal")
	}
}

func TestDecoderIterate(t *testing.T) {
	testdata := strings.Join([]string{
		"bool,bytes,complex,float,int,optional,string,uint,uppercase",
		"true,hello,1+1i,3.1415,-173,0,hello world,1337,gopher",
	}, "\n")

	var data []testStruct
	var row testStruct
	iter, err := NewDecoder(strings.NewReader(testdata)).Iterate(&row)
	if err != nil {
		t.Fatal(err)
	}

	for iter.Scan() {
		data = append(data, row)
	}

	if err := iter.Err(); err != nil {
		t.Fatal(err)
	}

	expect := []testStruct{
		{[]byte("hello"), true, 1 + 1i, 3.1415, -173, new(int), "hello world", "GOPHER", 4919},
	}

	if !reflect.DeepEqual(data, expect) {
		t.Error("should be equal")
	}
}

func TestDecodeHeaderless(t *testing.T) {
	testdata := "hello,true,1+1i,3.1415,-173,0,hello world,gopher,1337"

	var data []testStruct

	d := NewDecoder(strings.NewReader(testdata))
	d.SkipHeader()

	if err := d.Decode(&data); err != nil {
		t.Fatal(err)
	}

	expect := []testStruct{
		{[]byte("hello"), true, 1 + 1i, 3.1415, -173, new(int), "hello world", "GOPHER", 4919},
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
	d.SkipHeader()

	if err := d.Decode(&data); !errors.Is(err, csv.ErrFieldCount) {
		t.Fatal("should be error")
	}
}

func TestDecodeDisallowShortFields(t *testing.T) {
	type struc struct {
		B bool
		S string
	}

	testdata := "true"

	var data []struc

	d := NewDecoder(strings.NewReader(testdata))
	d.DisallowShortFields()
	d.SkipHeader()

	if err := d.Decode(&data); !errors.Is(err, csv.ErrFieldCount) {
		t.Fatal("should be error")
	}
}

func TestDecodeShortFields(t *testing.T) {
	type struc struct {
		B bool
		S string
	}

	testdata := "true\nfalse,yes"

	var data []struc

	d := NewDecoder(strings.NewReader(testdata))
	d.SkipHeader()

	if err := d.Decode(&data); err != nil {
		t.Fatal("should not be error")
	} else if len(data) != 2 || data[0].S != "" || data[1].S != "yes" {
		t.Fatal("parse error")
	}
}

func TestDecodeSyntaxError(t *testing.T) {
	type struc struct {
		B bool
		I int
	}

	testdata := ","

	var data []struc

	d := NewDecoder(strings.NewReader(testdata))
	d.SkipHeader()

	if err := d.Decode(&data); !errors.Is(err, strconv.ErrSyntax) {
		t.Fatal("expected syntax error", err)
	}
}

func TestDecodeOptionalFields(t *testing.T) {
	type struc struct {
		B *bool
		I *int
	}

	testdata := ","

	var data []struc

	d := NewDecoder(strings.NewReader(testdata))
	d.SkipHeader()

	if err := d.Decode(&data); err != nil {
		t.Fatal(err)
	}

	if data[0].B != nil || data[0].I != nil {
		t.Fatal("expected nil")
	}
}

func TestDecoderMapfunc(t *testing.T) {
	type struc struct {
		A int
		B float64
	}

	testdata := "100,1.23\n,n/a"

	var data []struc

	d := NewDecoder(strings.NewReader(testdata))
	d.SkipHeader()
	d.SetMapFunc(func(column, field string) string {
		if column == "A" && field == "" {
			return "0"
		} else if column == "B" && field == "n/a" {
			return "NaN"
		}
		return field
	})

	if err := d.Decode(&data); err != nil {
		t.Fatal(err)
	}

	if data[1].A != 0 || !math.IsNaN(data[1].B) {
		t.Fatal("should be zero and NaN")
	}
}
