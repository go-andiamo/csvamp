package main

import (
	"fmt"
	"github.com/go-andiamo/csvamp"
	"github.com/go-andiamo/csvamp/csv"
	"strings"
)

type Record struct {
	Age       int    `csv:"Age"`
	LastName  string `csv:"Last name"`
	FirstName string `csv:"First name"`
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
	const data = `Frodo,Baggins,50
Samwise,Gamgee,38
Aragorn,Elessar,87
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

	// csv.NoHeader(true) indicates to the reader that the CSV has no header line...
	r := mapper.Reader(strings.NewReader(data), nil, csv.NoHeader(true)).
		SupplyHeaders([]string{"First name", "Last name", "Age"})
	recs, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", recs)
}
