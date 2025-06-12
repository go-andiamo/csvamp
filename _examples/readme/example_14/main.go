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
Frodo,Baggins,not a number!
Samwise,Gamgee,38
Aragorn,Elessar,not a number!
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

	eh := &errorHandler{}
	r := mapper.Reader(strings.NewReader(data), nil).WithErrorHandler(eh)
	recs, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", recs)

	fmt.Printf("Found %d errors\n", len(eh.errs))
	for i, err := range eh.errs {
		fmt.Printf("Line %d: %s\n", eh.lines[i], err)
	}
}

type errorHandler struct {
	errs  []error
	lines []int
}

func (eh *errorHandler) Handle(err error, line int) error {
	eh.errs = append(eh.errs, err)
	eh.lines = append(eh.lines, line)
	return nil
}
