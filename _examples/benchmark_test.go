package benchmark

import (
	"encoding/csv"
	"github.com/go-andiamo/csvamp"
	csv2 "github.com/go-andiamo/csvamp/csv"
	"io"
	"strconv"
	"strings"
	"testing"
)

type Record struct {
	FirstName string
	LastName  string
	Age       int
}

var mapper = csvamp.MustNewMapper[Record](csvamp.DefaultEmptyValues(true))

const (
	sample = `First name,Last name,Age
"Frodo",Baggins,50
Samwise,Gamgee,38
"Aragorn",Elessar,
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`
	sampleNoHeader = `"Frodo",Baggins,50
Samwise,Gamgee,38
"Aragorn",Elessar,
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`
	expectCount = 5
)

func BenchmarkCsvamp(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// don't include reader creation in benchmark...
		b.StopTimer()
		r := mapper.Reader(strings.NewReader(sample), nil, csv2.ReuseRecord(true))
		b.StartTimer()
		count := 0
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			_ = record
			count++
		}
		if count != expectCount {
			panic("Incorrect expected records count")
		}
	}
}

func BenchmarkManual(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// don't include reader creation in benchmark...
		b.StopTimer()
		r := csv.NewReader(strings.NewReader(sample))
		r.ReuseRecord = true
		b.StartTimer()
		count := 0
		// read past the header...
		_, _ = r.Read()
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			row := Record{}
			row.FirstName = record[0]
			row.LastName = record[1]
			if record[2] != "" {
				row.Age, err = strconv.Atoi(record[2])
				if err != nil {
					panic(err)
				}
			}
			count++
		}
		if count != expectCount {
			panic("Incorrect expected records count")
		}
	}
}

func BenchmarkCsvamp_NoHeader(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// don't include reader creation in benchmark...
		b.StopTimer()
		r := mapper.Reader(strings.NewReader(sampleNoHeader), nil, csv2.NoHeader(true), csv2.ReuseRecord(true))
		b.StartTimer()
		count := 0
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			_ = record
			count++
		}
		if count != expectCount {
			panic("Incorrect expected records count")
		}
	}
}

func BenchmarkManual_NoHeader(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// don't include reader creation in benchmark...
		b.StopTimer()
		r := csv.NewReader(strings.NewReader(sampleNoHeader))
		r.ReuseRecord = true
		b.StartTimer()
		count := 0
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			row := Record{}
			row.FirstName = record[0]
			row.LastName = record[1]
			if record[2] != "" {
				row.Age, err = strconv.Atoi(record[2])
				if err != nil {
					panic(err)
				}
			}
			count++
		}
		if count != expectCount {
			panic("Incorrect expected records count")
		}
	}
}

type Record2 struct {
	Age       int    `csv:"Age"`
	FirstName string `csv:"First name"`
	LastName  string `csv:"Last name"`
}

var mapper2 = csvamp.MustNewMapper[Record2](csvamp.DefaultEmptyValues(true))

func BenchmarkCsvamp_Named(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// don't include reader creation in benchmark...
		b.StopTimer()
		r := mapper2.Reader(strings.NewReader(sample), nil, csv2.ReuseRecord(true))
		b.StartTimer()
		count := 0
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			_ = record
			count++
		}
		if count != expectCount {
			panic("Incorrect expected records count")
		}
	}
}

func BenchmarkManual_Named(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// don't include reader creation in benchmark...
		b.StopTimer()
		r := csv.NewReader(strings.NewReader(sample))
		r.ReuseRecord = true
		b.StartTimer()
		// read and map the headers to indexes (as csvamp does)...
		hdrs, err := r.Read()
		if err != nil {
			panic(err)
		}
		hdrMap := make(map[string]int, len(hdrs))
		for idx, h := range hdrs {
			hdrMap[h] = idx
		}
		count := 0
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			row := Record{}
			if idx, ok := hdrMap["First name"]; ok && idx >= 0 && idx < len(record) {
				row.FirstName = record[idx]
			}
			if idx, ok := hdrMap["Last name"]; ok && idx >= 0 && idx < len(record) {
				row.LastName = record[idx]
			}
			if idx, ok := hdrMap["Age"]; ok && idx >= 0 && idx < len(record) {
				if age := record[idx]; age != "" {
					row.Age, err = strconv.Atoi(age)
					if err != nil {
						panic(err)
					}
				}
			}
			count++
		}
		if count != expectCount {
			panic("Incorrect expected records count")
		}
	}
}

func BenchmarkCsvamp_Named_NoHeader(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// don't include reader creation in benchmark...
		b.StopTimer()
		r := mapper2.Reader(strings.NewReader(sampleNoHeader), nil, csv2.NoHeader(true), csv2.ReuseRecord(true)).
			SupplyHeaders([]string{"First name", "Last name", "Age"})
		b.StartTimer()
		count := 0
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			_ = record
			count++
		}
		if count != expectCount {
			panic("Incorrect expected records count")
		}
	}
}

func BenchmarkManual_Named_NoHeader(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// don't include reader creation in benchmark...
		b.StopTimer()
		r := csv.NewReader(strings.NewReader(sampleNoHeader))
		r.ReuseRecord = true
		hdrMap := map[string]int{
			"First name": 0,
			"Last name":  1,
			"Age":        2,
		}
		b.StartTimer()
		count := 0
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			row := Record{}
			if idx, ok := hdrMap["First name"]; ok && idx >= 0 && idx < len(record) {
				row.FirstName = record[idx]
			}
			if idx, ok := hdrMap["Last name"]; ok && idx >= 0 && idx < len(record) {
				row.LastName = record[idx]
			}
			if idx, ok := hdrMap["Age"]; ok && idx >= 0 && idx < len(record) {
				if age := record[idx]; age != "" {
					row.Age, err = strconv.Atoi(age)
					if err != nil {
						panic(err)
					}
				}
			}
			count++
		}
		if count != expectCount {
			panic("Incorrect expected records count")
		}
	}
}
