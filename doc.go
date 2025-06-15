// Package csvamp - Go package for reading CSV directly into structs
//
// Example:
//
//		type MyStruct struct {
//			Line int      `csv:"[line]"` // optionally capture the csv line number
//			Raw  []string `csv:"[raw]"`  // optionally capture the csv record
//			Foo  string   `csv:"[1]"`    // map to indexed csv field (1 based)
//			Bar  string   `csv:"bar"`    // map to named csv field
//		}
//
//		// create the mapper...
//		m, err := csvamp.NewMapper[MyStruct]()
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		const sample = `foo,bar
//	Aaa,Bbb
//	Ccc,Ddd`
//
//		result, err := m.Reader(strings.NewReader(sample), nil).ReadAll()
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("%+v\n", result)
package csvamp
