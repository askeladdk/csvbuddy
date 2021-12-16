package bench_test

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/askeladdk/csvbuddy"
	"github.com/gocarina/gocsv"
	"github.com/jszwec/csvutil"
	"github.com/yunabe/easycsv"
)

// BenchmarkUnmarshal is based on
// https://gist.github.com/jszwec/e8515e741190454fa3494bcd3e1f100f
func BenchmarkUnmarshal(b *testing.B) {
	type A struct {
		A int     `csv:"a" name:"a"`
		B float64 `csv:"b" name:"b"`
		C string  `csv:"c" name:"c"`
		D int64   `csv:"d" name:"d"`
		E int8    `csv:"e" name:"e"`
		F float32 `csv:"f" name:"f"`
		G float32 `csv:"g" name:"g"`
		H float32 `csv:"h" name:"h"`
		I string  `csv:"i" name:"i"`
		J int     `csv:"j" name:"j"`
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
		fn   func([]byte, *testing.B)
	}{
		{
			desc: "csvutil.Unmarshal",
			fn: func(data []byte, b *testing.B) {
				var a []A
				if err := csvutil.Unmarshal(data, &a); err != nil {
					b.Error(err)
				}
			},
		},
		{
			desc: "gocsv.Unmarshal",
			fn: func(data []byte, b *testing.B) {
				var a []A
				if err := gocsv.UnmarshalBytes(data, &a); err != nil {
					b.Error(err)
				}
			},
		},
		{
			desc: "easycsv.ReadAll",
			fn: func(data []byte, b *testing.B) {
				r := easycsv.NewReader(bytes.NewReader(data))
				var a []A
				if err := r.ReadAll(&a); err != nil {
					b.Error(err)
				}
			},
		},
		{
			desc: "csvbuddy.Unmarshal",
			fn: func(data []byte, b *testing.B) {
				var a []A
				if err := csvbuddy.Unmarshal(data, &a); err != nil {
					b.Error(err)
				}
			},
		},
	}

	for _, t := range tests {
		b.Run(t.desc, func(b *testing.B) {
			for _, f := range fixture {
				b.Run(f.desc, func(b *testing.B) {
					data := genData(f.records)
					for i := 0; i < b.N; i++ {
						t.fn(data, b)
					}
				})
			}
		})
	}
}

func genData(records int) []byte {
	header := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	record := []string{"1", "2.5", "foo", "6", "7", "8", "9", "10", "bar", "10"}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(header)

	for i := 0; i < records; i++ {
		_ = w.Write(record)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
