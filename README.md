# csvbuddy - convenient CSV codec for Go

[![GoDoc](https://godoc.org/github.com/askeladdk/csvbuddy?status.png)](https://godoc.org/github.com/askeladdk/csvbuddy)
[![Go Report Card](https://goreportcard.com/badge/github.com/askeladdk/csvbuddy)](https://goreportcard.com/report/github.com/askeladdk/csvbuddy)

## Overview

Package csvbuddy implements a convenient interface for encoding and decoding CSV files.

## Install

```
go get -u github.com/askeladdk/csvbuddy
```

## Quickstart

Use `Marshal` and `Unmarshal` to encode and decode slices of structs to and from byte slices. Only slices of structs can be encoded and decoded because CSV is defined as a list of records.

```go
type Person struct {
    Name string `csv:"name"`
    Age  int    `csv:"age"`
}

boys := []Person{
    {"Stan", 10},
    {"Kyle", 10},
    {"Cartman", 10},
    {"Kenny", 10},
    {"Ike", 5},
}

text, _ := csvbuddy.Marshal(&boys)

var boys2 []Person
_ = csvbuddy.Unmarshal(text, &boys2)

fmt.Println(string(text))
fmt.Println(reflect.DeepEqual(&boys, &boys2))
// Output:
// name,age
// Stan,10
// Kyle,10
// Cartman,10
// Kenny,10
// Ike,5
//
// true
```

Use `NewEncoder` and `NewDecoder` to produce and consume output and input streams.

```go
func encodePeople(w io.Writer, people *[]People) error {
    return csvbuddy.NewEncoder(w).Encode(people)
}

func decodePeople(r io.Reader, people *[]People) error {
    return csvbuddy.NewDecoder(r).Decode(people)
}
```

Use the `SetReaderFunc` and `SetWriterFunc` methods if you need more control over CSV parsing and writing. You can also provide a custom parser and writer by implementing the `Reader` and `Writer` interfaces.

```go
enc := csvbuddy.NewEncoder(w)
enc.SetWriterFunc(func(w io.Writer) csvbuddy.Writer {
    cw := csv.NewWriter(w)
    cw.Comma = ';'
    return cw
})
```

```go
dec := csvbuddy.NewDecoder(r)
dec.SetReaderFunc(func(r io.Reader) csvbuddy.Reader {
    cr := csv.NewReader(w)
    cr.Comma = ';'
    cr.ReuseRecord = true
    return cr
})
```

Use the `SetMapFunc` method to perform data cleaning on the fly.

```go
dec := csvbuddy.NewDecoder(w)
dec.SetMapFunc(func(name, value string) string {
    value = strings.TrimSpace(value)
    if name == "age" && value == "" {
        value = "0"
    }
    return value
})
```

Use the `Iterate` method to decode a CSV as a stream of rows. This allows decoding of very large CSV files without having to read it entirely into memory.

```go
dec := csvbuddy.NewDecoder(r)
var row structType

iter, _ := dec.Iterate(&row)

for iter.Scan() {
    fmt.Println(row)
}
```

Read the rest of the [documentation on pkg.go.dev](https://godoc.org/github.com/askeladdk/csvbuddy). It's easy-peasy!

## Performance

Unscientific benchmarks on my laptop suggest that the performance is comparable with [csvutil](https://github.com/jszwec/csvutil).

```
% go test -bench=. -benchmem
goos: darwin
goarch: amd64
pkg: bench_test
cpu: Intel(R) Core(TM) i5-5287U CPU @ 2.90GHz

# Marshal
BenchmarkMarshal/csvutil.Marshal/1_record-4         	  162283	      6300 ns/op	   10395 B/op	      13 allocs/op
BenchmarkMarshal/csvutil.Marshal/10_records-4       	   73813	     16208 ns/op	   11243 B/op	      22 allocs/op
BenchmarkMarshal/csvutil.Marshal/100_records-4      	   10000	    113700 ns/op	   27372 B/op	     113 allocs/op
BenchmarkMarshal/csvutil.Marshal/1000_records-4     	    1069	   1150120 ns/op	  185336 B/op	    1016 allocs/op
BenchmarkMarshal/csvutil.Marshal/10000_records-4    	     100	  11192591 ns/op	 1542795 B/op	   10019 allocs/op
BenchmarkMarshal/csvutil.Marshal/100000_records-4   	      10	 105524150 ns/op	22383758 B/op	  100023 allocs/op
BenchmarkMarshal/gocsv.Marshal/1_record-4           	  261706	      4046 ns/op	    4712 B/op	      31 allocs/op
BenchmarkMarshal/gocsv.Marshal/10_records-4         	   53725	     21879 ns/op	    7145 B/op	     238 allocs/op
BenchmarkMarshal/gocsv.Marshal/100_records-4        	    5115	    208961 ns/op	   39120 B/op	    2309 allocs/op
BenchmarkMarshal/gocsv.Marshal/1000_records-4       	     554	   2113323 ns/op	  355617 B/op	   23012 allocs/op
BenchmarkMarshal/gocsv.Marshal/10000_records-4      	      57	  20734269 ns/op	 3303851 B/op	  230019 allocs/op
BenchmarkMarshal/gocsv.Marshal/100000_records-4     	       5	 240347370 ns/op	40780699 B/op	 2300032 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/1_record-4        	  238672	      5174 ns/op	    5531 B/op	      22 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/10_records-4      	   76998	     15281 ns/op	    6955 B/op	      94 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/100_records-4     	    9872	    123790 ns/op	   28843 B/op	     815 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/1000_records-4    	     994	   1379329 ns/op	  244415 B/op	    8018 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/10000_records-4   	      96	  12226165 ns/op	 2178207 B/op	   80021 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/100000_records-4  	       9	 121619259 ns/op	28867782 B/op	  800025 allocs/op

# Unmarshal
BenchmarkUnmarshal/csvutil.Unmarshal/1_record-4             	  167607	      7193 ns/op	    8132 B/op	      31 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10_records-4           	   78772	     14302 ns/op	    9084 B/op	      40 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100_records-4          	   10000	    104986 ns/op	   18541 B/op	     130 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records-4         	    1317	    835977 ns/op	  113918 B/op	    1030 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records-4         	     140	   8483102 ns/op	 1058253 B/op	   10030 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records-4        	      13	  90634928 ns/op	11056816 B/op	  100031 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1_record-4                	  128088	      8749 ns/op	    7571 B/op	      65 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10_records-4              	   35547	     33419 ns/op	   15467 B/op	     320 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100_records-4             	    3704	    279071 ns/op	   92221 B/op	    2843 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1000_records-4            	     429	   2790427 ns/op	  878584 B/op	   28047 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10000_records-4           	      39	  33458428 ns/op	 9066452 B/op	  280054 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100000_records-4          	       4	 309066282 ns/op	95385456 B/op	 2800067 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/1_record-4                	   66268	     16141 ns/op	    9163 B/op	      96 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10_records-4              	   14227	     80105 ns/op	   22109 B/op	     496 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100_records-4             	    1738	    689075 ns/op	  145894 B/op	    4459 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/1000_records-4            	     171	   6987786 ns/op	 1442679 B/op	   44063 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10000_records-4           	      13	  77296045 ns/op	15585227 B/op	  440072 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100000_records-4          	       2	 756436991 ns/op	164327960 B/op	 4400088 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/1_record-4             	  176460	      5842 ns/op	    6483 B/op	      32 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/10_records-4           	   76069	     17500 ns/op	   10035 B/op	      63 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/100_records-4          	   10000	    106009 ns/op	   39860 B/op	     336 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/1000_records-4         	    1092	   1072939 ns/op	  396786 B/op	    3040 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/10000_records-4        	      79	  12884045 ns/op	 5068202 B/op	   30048 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/100000_records-4       	       8	 132744382 ns/op	56674790 B/op	  300060 allocs/op
```

## License

Package csvbuddy is released under the terms of the ISC license.
