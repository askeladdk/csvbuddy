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
dec := csvbuddy.NewDecoder(w)
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

Read the rest of the [documentation on pkg.go.dev](https://godoc.org/github.com/askeladdk/csvbuddy). It's easy-peasy!

## Performance

Unscientific benchmarks on my laptop suggest that the performance is comparable with [csvutil](https://github.com/jszwec/csvutil).

```
% go test -bench=. -benchmem
goos: darwin
goarch: amd64
pkg: bench_test
cpu: Intel(R) Core(TM) i5-5287U CPU @ 2.90GHz
BenchmarkMarshal/csvutil.Marshal/1_record-4         	  161492	      6878 ns/op	   10395 B/op	      13 allocs/op
BenchmarkMarshal/csvutil.Marshal/10_records-4       	   66846	     18382 ns/op	   11243 B/op	      22 allocs/op
BenchmarkMarshal/csvutil.Marshal/100_records-4      	    9657	    114004 ns/op	   28652 B/op	     113 allocs/op
BenchmarkMarshal/csvutil.Marshal/1000_records-4     	    1112	   1127961 ns/op	  168951 B/op	    1015 allocs/op
BenchmarkMarshal/csvutil.Marshal/10000_records-4    	     100	  11013181 ns/op	 1526410 B/op	   10018 allocs/op
BenchmarkMarshal/csvutil.Marshal/100000_records-4   	       9	 113611871 ns/op	22456304 B/op	  100022 allocs/op
BenchmarkMarshal/gocsv.Marshal/1_record-4           	  307435	      3950 ns/op	    4712 B/op	      31 allocs/op
BenchmarkMarshal/gocsv.Marshal/10_records-4         	   50709	     22645 ns/op	    7145 B/op	     238 allocs/op
BenchmarkMarshal/gocsv.Marshal/100_records-4        	    5860	    208473 ns/op	   40400 B/op	    2309 allocs/op
BenchmarkMarshal/gocsv.Marshal/1000_records-4       	     571	   2064275 ns/op	  339217 B/op	   23011 allocs/op
BenchmarkMarshal/gocsv.Marshal/10000_records-4      	      55	  21258436 ns/op	 3287871 B/op	  230017 allocs/op
BenchmarkMarshal/gocsv.Marshal/100000_records-4     	       5	 207044564 ns/op	40763456 B/op	 2300026 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/1_record-4        	  237073	      5318 ns/op	    5531 B/op	      22 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/10_records-4      	   77901	     15318 ns/op	    6955 B/op	      94 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/100_records-4     	   10000	    118745 ns/op	   30124 B/op	     815 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/1000_records-4    	    1038	   1210561 ns/op	  228026 B/op	    8017 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/10000_records-4   	     100	  11629436 ns/op	 2161485 B/op	   80020 allocs/op
BenchmarkMarshal/csvbuddy.Marshal/100000_records-4  	       9	 126898449 ns/op	28851391 B/op	  800024 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1_record-4     	  176967	      6956 ns/op	    8116 B/op	      31 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10_records-4   	   83502	     14245 ns/op	    9068 B/op	      40 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100_records-4  	   13210	     89361 ns/op	   18524 B/op	     130 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records-4 	    1342	    884145 ns/op	  113889 B/op	    1030 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records-4         	     144	   8460285 ns/op	 1057926 B/op	   10030 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records-4        	      13	  85097120 ns/op	11055532 B/op	  100030 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1_record-4                	  134127	      8231 ns/op	    7555 B/op	      65 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10_records-4              	   37320	     32954 ns/op	   15451 B/op	     320 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100_records-4             	    4459	    267411 ns/op	   92204 B/op	    2843 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1000_records-4            	     460	   2604536 ns/op	  849847 B/op	   28046 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10000_records-4           	      40	  27755352 ns/op	 9028482 B/op	  280055 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100000_records-4          	       4	 286982946 ns/op	95901584 B/op	 2800070 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/1_record-4                	   73222	     16010 ns/op	    9147 B/op	      96 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10_records-4              	   15373	     77384 ns/op	   22093 B/op	     496 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100_records-4             	    1629	    678903 ns/op	  145862 B/op	    4459 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/1000_records-4            	     163	   6691977 ns/op	 1344054 B/op	   44062 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10000_records-4           	      16	  71776105 ns/op	16362816 B/op	  440074 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100000_records-4          	       2	 724755432 ns/op	165428940 B/op	 4400089 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/1_record-4             	  185084	      6842 ns/op	    6419 B/op	      31 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/10_records-4           	   69288	     15790 ns/op	    9971 B/op	      62 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/100_records-4          	   10000	    106489 ns/op	   39795 B/op	     335 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/1000_records-4         	    1102	    989382 ns/op	  298400 B/op	    3038 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/10000_records-4        	      85	  12084431 ns/op	 5869823 B/op	   30049 allocs/op
BenchmarkUnmarshal/csvbuddy.Unmarshal/100000_records-4       	       8	 142464833 ns/op	58081667 B/op	  300060 allocs/op
```

## License

Package csvbuddy is released under the terms of the ISC license.
