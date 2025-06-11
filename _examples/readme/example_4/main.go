package main

import (
	"fmt"
	"github.com/go-andiamo/csvamp"
	"strings"
)

type Record struct {
	Line      int `csv:"[line]"`
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
	recs, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", recs)
}
