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
	Address
}

type Address struct {
	Street string
	Town   string
	County string
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
	const data = `First name,Last name,Age,Street,Town,County
Frodo,Baggins,50,1 Bagshot Row,Hobbiton,The Shire
Samwise,Gamgee,38,2 Bagshot Row,Hobbiton,The Shire
Aragorn,Elessar,87,Royal Quarters,The Citadel,Minas Tirith`

	r := mapper.Reader(strings.NewReader(data), nil)
	recs, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", recs)
}
