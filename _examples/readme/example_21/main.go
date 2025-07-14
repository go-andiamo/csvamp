package main

import (
	"fmt"
	"github.com/go-andiamo/csvamp"
	"strings"
)

type Record struct {
	FirstName string
	LastName  string
	Age       int
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
	const data = `First name,Last name,Age
Frodo,Baggins,50
Samwise,Gamgee,38
Aragorn,Elessar,87
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

	r := mapper.Reader(strings.NewReader(data), nil)

	rows := make([]Record, 0)
	for _, row := range r.Iterator() {
		rows = append(rows, row)
	}
	fmt.Printf("%+v\n", rows)
}
