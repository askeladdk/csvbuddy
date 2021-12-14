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

## License

Package csvbuddy is released under the terms of the ISC license.
