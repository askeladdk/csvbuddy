package csvbuddy_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/askeladdk/csvbuddy"
)

func Example_unmarshal() {
	moviesCSV := strings.Join([]string{
		"movie,year of release",
		"The Matrix,1999",
		"Back To The Future,1985",
		"The Terminator,1984",
		"2001: A Space Odyssey,1968",
	}, "\n")

	var movies []struct {
		Name string `csv:"movie"`
		Year int    `csv:"year of release"`
	}

	_ = csvbuddy.Unmarshal([]byte(moviesCSV), &movies)

	for _, movie := range movies {
		fmt.Printf("%s was released in %d.\n", movie.Name, movie.Year)
	}
	// Output:
	// The Matrix was released in 1999.
	// Back To The Future was released in 1985.
	// The Terminator was released in 1984.
	// 2001: A Space Odyssey was released in 1968.
}

func Example_marshal() {
	movies := []struct {
		Name string `csv:"movie"`
		Year int    `csv:"year of release"`
	}{
		{"The Matrix", 1999},
		{"Back To The Future", 1985},
		{"The Terminator", 1984},
		{"2001: A Space Odyssey", 1968},
	}

	text, _ := csvbuddy.Marshal(&movies)

	fmt.Println(string(text))
	// Output:
	// movie,year of release
	// The Matrix,1999
	// Back To The Future,1985
	// The Terminator,1984
	// 2001: A Space Odyssey,1968
}

func Example_dataCleaning() {
	// A messy CSV that is missing a header, uses semicolon delimiter,
	// has numbers with comma decimals, inconsistent capitalization, and stray spaces.
	var messyCSV = strings.Join([]string{
		"Tokyo   ; JP ; 35,6897 ; 139,6922",
		"jakarta ; Id ; -6,2146 ; 106,8451",
		"DELHI   ; in ; 28,6600 ;  77,2300  ",
	}, "\n")

	type city struct {
		Name      string  `csv:"name"`
		Country   string  `csv:"country"`
		Latitude  float32 `csv:"lat"`
		Longitude float32 `csv:"lng"`
	}

	d := csvbuddy.NewDecoder(strings.NewReader(messyCSV))

	// Set the Decoder to use the header derived from the city struct fields.
	d.SkipHeader()

	// Set the CSV reader to delimit on semicolons.
	d.SetReaderFunc(func(r io.Reader) csvbuddy.Reader {
		cr := csv.NewReader(r)
		cr.Comma = ';'
		cr.ReuseRecord = true
		return cr
	})

	// Set the Decoder to clean messy values.
	d.SetMapFunc(func(name, value string) string {
		value = strings.TrimSpace(value)
		switch name {
		case "lat", "lng":
			value = strings.ReplaceAll(value, ",", ".")
		case "name":
			value = strings.Title(strings.ToLower(value)) //nolint
		case "country":
			value = strings.ToUpper(value)
		}
		return value
	})

	// Decode into the cities variable.
	var cities []city
	_ = d.Decode(&cities)

	for _, city := range cities {
		fmt.Printf("%s, %s is located at coordinate (%.4f, %.4f).\n", city.Name, city.Country, city.Latitude, city.Longitude)
	}
	// Output:
	// Tokyo, JP is located at coordinate (35.6897, 139.6922).
	// Jakarta, ID is located at coordinate (-6.2146, 106.8451).
	// Delhi, IN is located at coordinate (28.6600, 77.2300).
}

func Example_floatingPointTags() {
	numbers := []struct {
		N float64 `csv:"number,prec=3,fmt=E"`
	}{{math.Pi}, {100e4}}
	text, _ := csvbuddy.Marshal(&numbers)
	fmt.Println(string(text))
	// Output:
	// number
	// 3.142E+00
	// 1.000E+06
}
