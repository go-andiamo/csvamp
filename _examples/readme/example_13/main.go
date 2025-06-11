package main

import (
	"fmt"
	"github.com/go-andiamo/csvamp"
	"strings"
)

type Record struct {
	FirstName string
	LastName  string
	Age       *int
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
	const data = `First name,Last name,Age
Frodo,Baggins,50
Samwise,Gamgee,38
Aragorn,Elessar,87
Legolas,Greenleaf,
Gandalf,The Grey,`

	r := mapper.Reader(strings.NewReader(data), nil)
	err := r.Iterate(func(record Record) (bool, error) {
		fmt.Printf("Name: %s %s\n", record.FirstName, record.LastName)
		if record.Age != nil {
			fmt.Printf("Age: %d\n", *record.Age)
		} else {
			fmt.Printf("Age: %s\n", "(unknown)")
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}
}
