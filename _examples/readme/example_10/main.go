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

	err := r.Iterate(func(record Record) (bool, error) {
		fmt.Printf("%+v\n", record)
		return true, nil
	})
	if err != nil {
		panic(err)
	}
}
