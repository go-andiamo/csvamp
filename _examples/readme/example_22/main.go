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
Legolas,Greenleaf,not a number
Gandalf,The Grey,not a number`

	eh := &errorHandler{}
	r := mapper.Reader(strings.NewReader(data), nil).WithErrorHandler(eh)

	rows := make([]Record, 0)
	for _, row := range r.Iterator() {
		rows = append(rows, row)
	}
	fmt.Printf("Rows: %+v\n", rows)
	fmt.Printf("Errors: %+v\n", eh)
}

type errorHandler struct {
	errs  []string
	lines []int
}

func (eh *errorHandler) Handle(err error, line int) error {
	eh.errs = append(eh.errs, err.Error())
	eh.lines = append(eh.lines, line)
	return nil
}
