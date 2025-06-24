# CSVAMP
[![GoDoc](https://godoc.org/github.com/go-andiamo/csvamp?status.svg)](https://pkg.go.dev/github.com/go-andiamo/csvamp)
[![Latest Version](https://img.shields.io/github/v/tag/go-andiamo/csvamp.svg?sort=semver&style=flat&label=version&color=blue)](https://github.com/go-andiamo/csvamp/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-andiamo/csvamp)](https://goreportcard.com/report/github.com/go-andiamo/csvamp)

Read CSVs directly into structs.

---

## Features

- Minimal reflection use
  - Field reflection only at mapper create time
  - Efficient field setters at read time
- Support for common field types: `bool`,`int`,`int8`,`int16`,`int32`,`int64`,`uint`,`uint8`,`uint16`,`uint32`,`uint64`,`float32`,`float64`,`string`
  - and pointers to those types
  - quoted detection on string pointers
- Support for additional types - when they implement `csvamp.CsvUnmarshaler`, `csvamp.CsvQuotedUnmarshaler` or `encoding.TextUnmarshaler`
- Support for embedded structs and nested structs
- Map struct fields to CSV field index or header name (using `csv` tag)
- Adaptable to varying CSVs
- Post processor option for validating and/or finalising struct
- Optional error handler for tracking errors without halting reads

---

## Installation

```bash
go get github.com/go-andiamo/csvamp
```

---

## Examples

<details>
    <summary><strong>1. Basic implied field ordering</strong></summary>

Without specifying any `csv` tags on struct fields, the order of the struct fields implies the CSV field order...

```go
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
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), nil)
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/1A3tsYXcEgP)

</details><br>

<details>
    <summary><strong>2. Explicit field indexes</strong></summary>

You can specify the actual CSV field indexes using the `csv` tag...

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/csvamp"
    "strings"
)

type Record struct {
    Age       int    `csv:"[3]"`
    LastName  string `csv:"[2]"`
    FirstName string `csv:"[1]"`
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
    const data = `First name,Last name,Age
Frodo,Baggins,50
Samwise,Gamgee,38
Aragorn,Elessar,87
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), nil)
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/p9Lc8mMe5Ic)

</details><br>

<details>
    <summary><strong>3. Mapping to CSV headers</strong></summary>

You can use the `csv` tag to map struct fields to explicit CSV headers...

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/csvamp"
    "strings"
)

type Record struct {
    Age       int    `csv:"Age"`
    LastName  string `csv:"Last name"`
    FirstName string `csv:"First name"`
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
    const data = `First name,Last name,Age
Frodo,Baggins,50
Samwise,Gamgee,38
Aragorn,Elessar,87
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), nil)
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/OvowRsJnCUE)

</details><br>

<details>
    <summary><strong>4. Capturing the CSV line number</strong></summary>

You can use a special `csv` tag to capture the CSV line number into the struct...

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/csvamp"
    "strings"
)

type Record struct {
    Line      int `csv:"[line]"`
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
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), nil)
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/ZVrBXm7-VsP)

</details><br>

<details>
    <summary><strong>5. Capturing the CSV record</strong></summary>

You can use a special `csv` tag to capture the CSV record (line) into the struct...

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/csvamp"
    "strings"
)

type Record struct {
    Raw       []string `csv:"[raw]"`
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
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), nil)
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/TBDZCvZBRvv)

</details><br>

<details>
    <summary><strong>6. Adapting the mapper to varying CSVs</strong></summary>

Sometimes, your CSV won't always arrive in the format you're expecting - the mapper can adapt to that...

```go
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
```

[try on go-playground](https://go.dev/play/p/xcNunZOCcL8)

</details><br>

<details>
    <summary><strong>7. Using <code>postProcessor</code> to validate</strong></summary>

You can use the `postProcessor` to validate records...

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/csvamp"
    "strings"
)

type Record struct {
    Line      int `csv:"[line]"`
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
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), func(row *Record) error {
        if row.Age > 200 {
            return fmt.Errorf("age is too high - on line %d", row.Line)
        }
        return nil
    })
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/AFgn9Dy7YAd)

</details><br>

<details>
    <summary><strong>8. Using <code>postProcessor</code> to adjust struct read</strong></summary>

Sometimes, some fields are calculated - you can exclude them from being read using `csv:"-"` and then fill them in using a `postProcessor`...

```go
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
    AgeInDays int `csv:"-"`
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
    const data = `First name,Last name,Age
Frodo,Baggins,50
Samwise,Gamgee,38
Aragorn,Elessar,87
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), func(row *Record) error {
        row.AgeInDays = row.Age * 365
        return nil
    })
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/h0gSN5CZ3Df)

</details><br>

<details>
    <summary><strong>9. Reading one row at a time</strong></summary>

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/csvamp"
    "io"
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
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), nil)

    for {
        record, err := r.Read()
        if err == io.EOF {
            break
        } else if err != nil {
            panic(err)
        }
        fmt.Printf("%+v\n", record)
}
}
```

[try on go-playground](https://go.dev/play/p/Vot_iCJv1MX)

</details><br>

<details>
    <summary><strong>10. Iterating over rows</strong></summary>

```go
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
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    r := mapper.Reader(strings.NewReader(data), nil)

    err := r.Iterate(func(record Record) (bool, error) {
        fmt.Printf("%+v\n", record)
        return true, nil
    })
    if err != nil {
        panic(err)
    }
}
```

[try on go-playground](https://go.dev/play/p/smNxzXmIlel)

</details><br>

<details>
    <summary><strong>11. Using <code>CsvUnmarshaler</code> (e.g. dates)</strong></summary>

CSVs come in all flavours - and dates (which `csvamp` doesn't handle natively) come in varying formats.  Use types that implement `csvamp.CsvUnmarshaler` to resolve this...

```go
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
```

[try on go-playground](https://go.dev/play/p/sCBMdYV2bM4)

</details><br>

<details>
    <summary><strong>12. Utilising <code>encoding.TextUnmarshaler</code> (e.g. dates)</strong></summary>

`csvamp` only supports 'primitive' types - but, fortunately, many additional types support the `encoding.TextUnmarshaler` interface - 
this can be utilised.  The following example utilises the fact that `time.Time` implements the `encoding.TextUnmarshaler` interface...

```go
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
```

[try on go-playground](https://go.dev/play/p/yr6LrgCfOVj)

</details><br>

<details>
    <summary><strong>13. Optional fields? Use pointer types</strong></summary>

When a struct field is a pointer type, `csvamp` treats it as optional - if the corresponding CSV field is empty - it is treated as not there at all...

```go
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
```

[try on go-playground](https://go.dev/play/p/Ppw0ZoBZwwp)

</details><br>

<details>
    <summary><strong>14. Using error handler to track errors</strong></summary>

When using `ReadAll()` (or `Iterate()`) you may not want to halt when an error is encountered... 

```go
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
```

[try on go-playground](https://go.dev/play/p/QLNoDtewIRa)

</details><br>

<details>
    <summary><strong>15. Manually supply CSV headers</strong></summary>

Sometimes, incoming CSV won't have a header line - but your struct fields are mapped to header names.  This can be handled by supplying the headers...

```go
package main

import (
    "fmt"
    "github.com/go-andiamo/csvamp"
    "github.com/go-andiamo/csvamp/csv"
    "strings"
)

type Record struct {
    Age       int    `csv:"Age"`
    LastName  string `csv:"Last name"`
    FirstName string `csv:"First name"`
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
    const data = `Frodo,Baggins,50
Samwise,Gamgee,38
Aragorn,Elessar,87
Legolas,Greenleaf,2931
Gandalf,The Grey,24000`

    // csv.NoHeader(true) indicates to the reader that the CSV has no header line...
    r := mapper.Reader(strings.NewReader(data), nil, csv.NoHeader(true)).
        SupplyHeaders([]string{"First name", "Last name", "Age"})
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/bgTZ12NM11x)

</details><br>

<details>
    <summary><strong>16. Quoted detection on string pointer fields</strong></summary>

Using string pointer fields determines whether the CSV field was quoted or un-quoted for empty string or nil...

```go
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
```

[try on go-playground](https://go.dev/play/p/AIbfIS90PUv)

</details><br>

<details>
    <summary><strong>17. Nested structs</strong></summary>

```go
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
    Address   Address
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
```

[try on go-playground](https://go.dev/play/p/qgPT2hONuZ3)

</details><br>

<details>
    <summary><strong>18. Embedded structs</strong></summary>

```go
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
```

[try on go-playground](https://go.dev/play/p/uzwjWlb7ZdA)

</details><br>

<details>
    <summary><strong>19. Nested structs with unmarshalling</strong></summary>

Sometimes, you may want a nested struct to dissect a single csv field.  If the nested struct implements `CsvUnmarshaler`, `CsvQuotedUnmarshaler` or `encoding.TextUnmarshaler` interface, the struct field is treated as mapped to a single csv field...

```go
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
    Address   Address
}

type Address struct {
    Lines []string
}

func (a *Address) UnmarshalCSV(val string, record []string) error {
    if val != "" {
        a.Lines = strings.Split(val, "\n")
    }
    return nil
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
    const data = `First name,Last name,Age,Address
Frodo,Baggins,50,"1 Bagshot Row
Hobbiton
The Shire"
Samwise,Gamgee,38,"2 Bagshot Row
Hobbiton
The Shire"
Aragorn,Elessar,87,"Royal Quarters
The Citadel
Minas Tirith"`

    r := mapper.Reader(strings.NewReader(data), nil)
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/PhysaOsaro0)

</details><br>

<details>
    <summary><strong>20. Special handling of <code>[]string</code> fields</strong></summary>

csvamp has special handling of `[]string` fields - it treats the quoted csv field as comma separated...

```go
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
    Address   []string
}

var mapper = csvamp.MustNewMapper[Record]()

func main() {
    const data = `First name,Last name,Age,Address
Frodo,Baggins,50,"1 Bagshot Row,Hobbiton,The Shire"
Samwise,Gamgee,38,"2 Bagshot Row,Hobbiton,The Shire"
Aragorn,Elessar,87,"Royal Quarters,The Citadel,Minas Tirith"`

    r := mapper.Reader(strings.NewReader(data), nil)
    recs, err := r.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", recs)
}
```

[try on go-playground](https://go.dev/play/p/1oFekDnz1Lc)

</details><br>
