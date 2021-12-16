package bench_test

import (
	"bytes"
	"testing"

	"github.com/askeladdk/csvbuddy"
	"github.com/gocarina/gocsv"
	"github.com/jszwec/csvutil"
)

// BenchmarkMarshal is based on
// https://gist.github.com/jszwec/31980321e1852ebb5615a44ccf374f17
func BenchmarkMarshal(b *testing.B) {
	type A struct {
		A int     `csv:"a"`
		B float64 `csv:"b"`
		C string  `csv:"c"`
		D int64   `csv:"d"`
		E int8    `csv:"e"`
		F float32 `csv:"f"`
		G float32 `csv:"g"`
		H float32 `csv:"h"`
		I string  `csv:"i"`
		J int     `csv:"j"`
	}

	fixture := []struct {
		desc    string
		records int
	}{
		{
			desc:    "1 record",
			records: 1,
		},
		{
			desc:    "10 records",
			records: 10,
		},
		{
			desc:    "100 records",
			records: 100,
		},
		{
			desc:    "1000 records",
			records: 1000,
		},
		{
			desc:    "10000 records",
			records: 10000,
		},
		{
			desc:    "100000 records",
			records: 100000,
		},
	}

	tests := []struct {
		desc string
		fn   func([]A, *testing.B)
	}{
		{
			desc: "csvutil.Marshal",
			fn: func(as []A, b *testing.B) {
				if _, err := csvutil.Marshal(as); err != nil {
					b.Error(err)
				}
			},
		},
		{
			desc: "gocsv.Marshal",
			fn: func(as []A, b *testing.B) {
				var buf bytes.Buffer
				if err := gocsv.Marshal(as, &buf); err != nil {
					b.Error(err)
				}
			},
		},
		{
			desc: "csvbuddy.Marshal",
			fn: func(as []A, b *testing.B) {
				if _, err := csvbuddy.Marshal(&as); err != nil {
					b.Error(err)
				}
			},
		},
	}

	a := A{
		A: 1,
		B: 3.14,
		C: "string",
		D: 1,
		E: 1,
		F: 3.13,
		G: 3.14,
		H: 3.14,
		I: "string",
		J: 10,
	}

	for _, t := range tests {
		b.Run(t.desc, func(b *testing.B) {
			for _, f := range fixture {
				b.Run(f.desc, func(b *testing.B) {
					as := make([]A, f.records)
					for i := range as {
						as[i] = a
					}
					for i := 0; i < b.N; i++ {
						t.fn(as, b)
					}
				})
			}
		})
	}
}
