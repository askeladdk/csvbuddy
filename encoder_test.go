package csvbuddy

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	data := []struct {
		A string
		B int
		C float64
	}{
		{"abc", 1, 0.1},
		{"def", 2, 0.2},
		{"ghi", 3, 0.3},
	}

	var b bytes.Buffer

	e := NewEncoder(&b)

	e.SetMapFunc(func(name, value string) string {
		if name == "C" {
			value = strings.Replace(value, ".", ",", 1)
		}
		return value
	})

	if err := e.Encode(&data); err != nil {
		t.Fatal(err)
	}

	if b.String() != "A,B,C\nabc,1,\"0,1\"\ndef,2,\"0,2\"\nghi,3,\"0,3\"\n" {
		t.Fatal(b.String())
	}
}
