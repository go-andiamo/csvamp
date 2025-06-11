package main

import (
	"fmt"
	"github.com/go-andiamo/csvamp"
	"strings"
	"time"
)

type DOB time.Time

func (d *DOB) UnmarshalCSV(val string, record []string) error {
	dt, err := time.Parse("2006-01-02", val)
	if err != nil {
		return err
	}
	*d = DOB(dt)
	return nil
}

type Record struct {
	FirstName string
	LastName  string
	DOB       DOB
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
	const data = `First name,Last name,Date Of Birth
Frodo,Baggins,2968-09-22
Samwise,Gamgee,2980-04-06
Aragorn,Elessar,2931-03-01`

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
