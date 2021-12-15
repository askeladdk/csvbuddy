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

	e.SetHeader([]string{"B", "A", "C"})

	if err := e.Encode(&data); err != nil {
		t.Fatal(err)
	}

	if b.String() != "B,A,C\n1,abc,\"0,1\"\n2,def,\"0,2\"\n3,ghi,\"0,3\"\n" {
		t.Fatal(b.String())
	}
}
