package main

import (
	"fmt"
	"github.com/go-andiamo/csvamp"
	"strings"
	"time"
)

type Record struct {
	FirstName string
	LastName  string
	DOB       time.Time
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
	const data = `First name,Last name,Date Of Birth
Frodo,Baggins,2968-09-22T00:00:00Z
Samwise,Gamgee,2980-04-06T00:00:00Z
Aragorn,Elessar,2931-03-01T00:00:00Z`

	r := mapper.Reader(strings.NewReader(data), nil)
	err := r.Iterate(func(record Record) (bool, error) {
		fmt.Printf("Name: %s %s\n", record.FirstName, record.LastName)
		fmt.Printf(" DOB: %s\n", time.Time(record.DOB).Format("Mon, 02 Jan 2006"))
		return true, nil
	})
	if err != nil {
		panic(err)
	}
}
