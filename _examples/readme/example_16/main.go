package main

import (
	"fmt"
	"github.com/go-andiamo/csvamp"
	"strings"
)

type Record struct {
	Name *string
	Age  int
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
	const data = `Name,Age
"",50
,38`

	r := mapper.Reader(strings.NewReader(data), nil)
	recs, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	// Name for first record should be pointer to empty string
	// Name for second record should be nil
	fmt.Printf("%+v\n", recs)
}
