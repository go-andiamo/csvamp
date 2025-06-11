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
	const sample1 = `First name,Last name,Age
Frodo,Baggins,50
Samwise,Gamgee,38
Aragorn,Elessar,87
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`
	const sample2 = `Age,First name,Last name
50,Frodo,Baggins
38,Samwise,Gamgee
87,Aragorn,Elessar
2931,Legolas,Greenleaf
24000,Gandalf,The Grey`

	recs, err := mapper.Reader(strings.NewReader(sample1), nil).ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", recs)

	m2, err := mapper.Adapt(false, csvamp.OverrideMappings{
		{
			FieldName:    "FirstName",
			CsvFieldName: "First name",
		},
		{
			FieldName:    "LastName",
			CsvFieldName: "Last name",
		},
		{
			FieldName:    "Age",
			CsvFieldName: "Age",
		},
	})
	if err != nil {
		panic(err)
	}
	recs, err = m2.Reader(strings.NewReader(sample2), nil).ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", recs)
}
