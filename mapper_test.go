package csvamp

import (
	"errors"
	"github.com/go-andiamo/csvamp/csv"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNewMapper(t *testing.T) {
	t.Run("Regular", func(t *testing.T) {
		type testStruct struct {
			Line       int      `csv:"[line]"`
			Raw        []string `csv:"[raw]"`
			RawData    []byte   `csv:"[rawData]"`
			Foo        string
			Bar        string
			unexported string
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)
		rm, ok := m.(*mapper[testStruct])
		require.True(t, ok)
		require.NotNil(t, rm.lineMapper)
		require.NotNil(t, rm.rawMapper)
		require.NotNil(t, rm.rawDataMapper)
		require.Equal(t, 3, rm.fieldIndex)
		require.Len(t, rm.csvFieldIndices, 2)
		_, ok = rm.csvFieldIndices[1]
		require.True(t, ok)
		_, ok = rm.csvFieldIndices[2]
		require.True(t, ok)
		require.Len(t, rm.csvFieldNames, 0)
		require.Len(t, rm.fieldMappings, 2)
		require.Equal(t, 1, rm.fieldMappings["Foo"])
		require.Equal(t, 2, rm.fieldMappings["Bar"])
		require.NotPanics(t, func() {
			_ = MustNewMapper[testStruct]()
		})
	})
	t.Run("Embedded", func(t *testing.T) {
		type Embedded struct {
			Line       int `csv:"[line]"`
			unexported string
			Bar        string `csv:"bar"`
		}
		type testStruct struct {
			Raw        []string `csv:"[raw]"`
			RawData    []byte   `csv:"[rawData]"`
			unexported string
			Foo        string
			Embedded
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)
		rm, ok := m.(*mapper[testStruct])
		require.True(t, ok)
		require.NotNil(t, rm.lineMapper)
		require.NotNil(t, rm.rawMapper)
		require.NotNil(t, rm.rawDataMapper)
		require.Len(t, rm.fieldMappings, 2)
		require.Equal(t, 1, rm.fieldMappings["Foo"])
		require.Equal(t, "bar", rm.fieldMappings["Bar"])
		require.Len(t, rm.fieldIndices, 5)
		require.Equal(t, []int{0}, rm.fieldIndices["Raw"])
		require.Equal(t, []int{1}, rm.fieldIndices["RawData"])
		require.Equal(t, []int{3}, rm.fieldIndices["Foo"])
		require.Equal(t, []int{4, 0}, rm.fieldIndices["Line"])
		require.Equal(t, []int{4, 2}, rm.fieldIndices["Bar"])
		require.NotPanics(t, func() {
			_ = MustNewMapper[testStruct]()
		})
	})
}

func TestNewMapper_Errors(t *testing.T) {
	t.Run("Not a struct", func(t *testing.T) {
		_, err := NewMapper[string]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected a struct but got ")
		require.Panics(t, func() {
			_ = MustNewMapper[string]()
		})
	})
	t.Run("Bad option", func(t *testing.T) {
		type testStruct struct {
			Foo string
		}
		_, err := NewMapper[testStruct]("not a valid option")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown option type:")
	})
	t.Run("Bad field type for [line]", func(t *testing.T) {
		type testStruct struct {
			Line string `csv:"[line]"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected to be int")
	})
	t.Run("Bad field type for [raw]", func(t *testing.T) {
		type testStruct struct {
			Raw string `csv:"[raw]"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected to be slice of strings")
	})
	t.Run("Bad field type for [rawData]", func(t *testing.T) {
		type testStruct struct {
			RawData int `csv:"[rawData]"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected to be slice of bytes or string")
	})
	t.Run("Bad field type for [rawData] in anonymous", func(t *testing.T) {
		type EmbeddedStruct struct {
			RawData int `csv:"[rawData]"`
		}
		type testStruct struct {
			EmbeddedStruct
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected to be slice of bytes or string")
	})
	t.Run("Bad field index", func(t *testing.T) {
		type testStruct struct {
			Foo string `csv:"[0]"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid csv field index")
	})
	t.Run("Bad field index (non-numeric)", func(t *testing.T) {
		type testStruct struct {
			Foo string `csv:"[??]"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid csv field index")
	})
	t.Run("Unsupported struct field type (by index)", func(t *testing.T) {
		type testStruct struct {
			Foo struct{} `csv:"[1]"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "struct field unsupported type:")
	})
	t.Run("Unsupported struct field type (by implied index)", func(t *testing.T) {
		type testStruct struct {
			Foo struct{}
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "struct field unsupported type:")
	})
	t.Run("Unsupported struct field type (by name)", func(t *testing.T) {
		type testStruct struct {
			Foo struct{} `csv:"Foo"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "struct field unsupported type:")
	})
	t.Run("Duplicate field index", func(t *testing.T) {
		type testStruct struct {
			Foo string `csv:"[1]"`
			Bar string `csv:"[1]"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "field with csv index ")
		require.Contains(t, err.Error(), " already mapped")
	})
	t.Run("Duplicate field index (mix specified and implied index)", func(t *testing.T) {
		type testStruct struct {
			Foo string
			Bar string `csv:"[1]"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "field with csv index ")
		require.Contains(t, err.Error(), " already mapped")
	})
	t.Run("Duplicate field name", func(t *testing.T) {
		type testStruct struct {
			Foo string `csv:"Foo"`
			Bar string `csv:"Foo"`
		}
		_, err := NewMapper[testStruct]()
		require.Error(t, err)
		require.Contains(t, err.Error(), "field with csv name ")
		require.Contains(t, err.Error(), " already mapped")
	})
}

func TestMapper_Read(t *testing.T) {
	type testStruct struct {
		Line    int      `csv:"[line]"`
		Raw     []string `csv:"[raw]"`
		RawData string   `csv:"[rawData]"`
		Foo     string
		Bar     string
		Baz     string
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	r := csv.NewReader(strings.NewReader(data))
	ctx := m.ReaderContext(r, nil)

	result, err := ctx.Read()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 2, result.Line)
	require.Equal(t, []string{"Aaa", "Bbb", "Ccc"}, result.Raw)
	require.Equal(t, "Aaa,\"Bbb\",Ccc\n", result.RawData)
	require.Equal(t, "Aaa", result.Foo)
	require.Equal(t, "Bbb", result.Bar)
	require.Equal(t, "Ccc", result.Baz)
}

func TestMapper_Read_WithIOReader(t *testing.T) {
	type testStruct struct {
		Line    int      `csv:"[line]"`
		Raw     []string `csv:"[raw]"`
		RawData string   `csv:"[rawData]"`
		Foo     string
		Bar     string
		Baz     string
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	ctx := m.Reader(strings.NewReader(data), nil)

	result, err := ctx.Read()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 2, result.Line)
	require.Equal(t, []string{"Aaa", "Bbb", "Ccc"}, result.Raw)
	require.Equal(t, "Aaa,\"Bbb\",Ccc\n", result.RawData)
	require.Equal(t, "Aaa", result.Foo)
	require.Equal(t, "Bbb", result.Bar)
	require.Equal(t, "Ccc", result.Baz)
}

func TestMapper_Read_RawDataBytes(t *testing.T) {
	type testStruct struct {
		Line    int      `csv:"[line]"`
		Raw     []string `csv:"[raw]"`
		RawData []byte   `csv:"[rawData]"`
		Foo     string
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	r := csv.NewReader(strings.NewReader(data))
	ctx := m.ReaderContext(r, nil)

	result, err := ctx.Read()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 2, result.Line)
	require.Equal(t, []string{"Aaa", "Bbb", "Ccc"}, result.Raw)
	require.Equal(t, "Aaa,\"Bbb\",Ccc\n", string(result.RawData))
}

func TestMapper_Read_WithPostProcessor(t *testing.T) {
	type testStruct struct {
		Line int `csv:"[line]"`
		Foo  string
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	r := csv.NewReader(strings.NewReader(data))
	ctx := m.ReaderContext(r, func(row *testStruct) error {
		row.Line = 0 - row.Line
		return nil
	})

	result, err := ctx.Read()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, -2, result.Line)
}

func TestMapper_Read_Errors(t *testing.T) {
	t.Run("UnmarshalCSV fails for indexed field", func(t *testing.T) {
		type testStruct struct {
			Foo MyBadString `csv:"[1]"`
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc`
		r := csv.NewReader(strings.NewReader(data))
		ctx := m.ReaderContext(r, nil)

		_, err = ctx.Read()
		require.Error(t, err)
		require.Equal(t, "fooey", err.Error())
	})
	t.Run("UnmarshalCSV fails for named field", func(t *testing.T) {
		type testStruct struct {
			Foo MyBadString `csv:"Foo"`
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc`
		r := csv.NewReader(strings.NewReader(data))
		ctx := m.ReaderContext(r, nil)

		_, err = ctx.Read()
		require.Error(t, err)
		require.Equal(t, "fooey", err.Error())
	})
	t.Run("Named field index out of range", func(t *testing.T) {
		type testStruct struct {
			Foo MyBadString `csv:"Baz"`
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Foo,Bar,Baz
Aaa,"Bbb"`
		r := csv.NewReader(strings.NewReader(data))
		r.FieldsPerRecord = -1
		ctx := m.ReaderContext(r, nil)

		_, err = ctx.Read()
		require.Error(t, err)
		require.Contains(t, err.Error(), "csv field index ")
		require.Contains(t, err.Error(), " out of range in record")
	})
}

func TestMapper_ReadAll(t *testing.T) {
	type testStruct struct {
		Line    int      `csv:"[line]"`
		Raw     []string `csv:"[raw]"`
		RawData string   `csv:"[rawData]"`
		Foo     string
		Bar     string
		Baz     string
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	r := csv.NewReader(strings.NewReader(data))
	ctx := m.ReaderContext(r, nil)

	result, err := ctx.ReadAll()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result, 2)

	require.Equal(t, 2, result[0].Line)
	require.Equal(t, []string{"Aaa", "Bbb", "Ccc"}, result[0].Raw)
	require.Equal(t, "Aaa,\"Bbb\",Ccc\n", result[0].RawData)
	require.Equal(t, "Aaa", result[0].Foo)
	require.Equal(t, "Bbb", result[0].Bar)
	require.Equal(t, "Ccc", result[0].Baz)
	require.Equal(t, 3, result[1].Line)
	require.Equal(t, []string{"Ddd", "Eee", "Fff"}, result[1].Raw)
	require.Equal(t, "Ddd,Eee,Fff", result[1].RawData)
	require.Equal(t, "Ddd", result[1].Foo)
	require.Equal(t, "Eee", result[1].Bar)
	require.Equal(t, "Fff", result[1].Baz)
}

func TestMapper_ReadAll_Errors(t *testing.T) {
	type testStruct struct {
		Line    int      `csv:"[line]"`
		Raw     []string `csv:"[raw]"`
		RawData string   `csv:"[rawData]"`
		Foo     string
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	r := csv.NewReader(strings.NewReader(data))
	ctx := m.ReaderContext(r, func(row *testStruct) error {
		return errors.New("fooey")
	})
	_, err = ctx.ReadAll()
	require.Error(t, err)
}

func TestMapper_Iterate(t *testing.T) {
	type testStruct struct {
		Line    int      `csv:"[line]"`
		Raw     []string `csv:"[raw]"`
		RawData string   `csv:"[rawData]"`
		Foo     string
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	r := csv.NewReader(strings.NewReader(data))
	ctx := m.ReaderContext(r, nil)

	result := make([]testStruct, 0)
	err = ctx.Iterate(func(v testStruct) (bool, error) {
		result = append(result, v)
		return true, nil
	})
	require.NoError(t, err)
	require.Len(t, result, 2)

	require.Equal(t, 2, result[0].Line)
	require.Equal(t, []string{"Aaa", "Bbb", "Ccc"}, result[0].Raw)
	require.Equal(t, "Aaa,\"Bbb\",Ccc\n", result[0].RawData)
	require.Equal(t, 3, result[1].Line)
	require.Equal(t, []string{"Ddd", "Eee", "Fff"}, result[1].Raw)
	require.Equal(t, "Ddd,Eee,Fff", result[1].RawData)
}

func TestMapper_Iterate_Errors(t *testing.T) {
	type testStruct struct {
		Line int `csv:"[line]"`
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	r := csv.NewReader(strings.NewReader(data))
	ctx := m.ReaderContext(r, func(row *testStruct) error {
		return errors.New("fooey")
	})

	result := make([]testStruct, 0)
	err = ctx.Iterate(func(v testStruct) (bool, error) {
		result = append(result, v)
		return true, nil
	})
	require.Error(t, err)
}

func TestMapper_Read_Unmarshaler(t *testing.T) {
	t.Run("Non-Pointer", func(t *testing.T) {
		type testStruct struct {
			Foo MyString
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
		r := csv.NewReader(strings.NewReader(data))
		ctx := m.ReaderContext(r, nil)

		result, err := ctx.Read()
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, MyString("Aaa"), result.Foo)
	})
	t.Run("Pointer", func(t *testing.T) {
		type testStruct struct {
			Foo *MyString
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
		r := csv.NewReader(strings.NewReader(data))
		ctx := m.ReaderContext(r, nil)

		result, err := ctx.Read()
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, MyString("Aaa"), *result.Foo)
	})
}

func TestMapper_Read_NamedField(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		type testStruct struct {
			Foo string `csv:"Foo"`
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc`
		r := csv.NewReader(strings.NewReader(data))
		ctx := m.ReaderContext(r, nil)

		result, err := ctx.Read()
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Aaa", result.Foo)
	})
	t.Run("Csv header not present", func(t *testing.T) {
		type testStruct struct {
			Foo string `csv:"Foo"`
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Bar,Baz
Aaa,"Bbb"`
		r := csv.NewReader(strings.NewReader(data))
		ctx := m.ReaderContext(r, nil)

		_, err = ctx.Read()
		require.Error(t, err)
		require.Contains(t, err.Error(), "csv header ")
		require.Contains(t, err.Error(), " not present")
	})
	t.Run("Csv header not present - ignored", func(t *testing.T) {
		type testStruct struct {
			Foo string `csv:"Foo"`
		}
		m, err := NewMapper[testStruct](IgnoreUnknownFieldNames(true))
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Bar,Baz
Aaa,"Bbb"`
		r := csv.NewReader(strings.NewReader(data))
		ctx := m.ReaderContext(r, nil)

		result, err := ctx.Read()
		require.NoError(t, err)
		require.Equal(t, "", result.Foo)
	})
	t.Run("No Headers", func(t *testing.T) {
		type testStruct struct {
			Foo string `csv:"Foo"`
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Foo,Bar,Baz`
		r := csv.NewReader(strings.NewReader(data))
		r.NoHeader = true
		ctx := m.ReaderContext(r, nil)

		_, err = ctx.Read()
		require.Error(t, err)
		require.Contains(t, err.Error(), "csv headers not present")
	})
}

func TestMapper_Adapt(t *testing.T) {
	type Embedded struct {
		Line       int `csv:"[line]"`
		unexported string
		Bar        string `csv:"bar"`
	}
	type testStruct struct {
		Raw        []string `csv:"[raw]"`
		RawData    []byte   `csv:"[rawData]"`
		unexported string
		Foo        string
		Baz        string   `csv:"[2]"`
		BadType    struct{} `csv:"-"`
		Embedded
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	orm := m.(*mapper[testStruct])
	t.Run("Clear", func(t *testing.T) {
		mm, err := m.Adapt(true, nil)
		require.NoError(t, err)
		rm := mm.(*mapper[testStruct])
		require.Empty(t, rm.fieldMappings)
		require.Empty(t, rm.csvFieldIndices)
		require.Empty(t, rm.csvFieldNames)
	})
	t.Run("Bad option", func(t *testing.T) {
		_, err := m.Adapt(true, nil, "not a valid option")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown option type:")
	})
	t.Run("Clone but change option", func(t *testing.T) {
		mm, err := m.Adapt(false, nil, IgnoreUnknownFieldNames(true))
		require.NoError(t, err)
		rm := mm.(*mapper[testStruct])
		require.True(t, rm.ignoreUnknownFieldNames)
		require.False(t, orm.ignoreUnknownFieldNames)
		require.Equal(t, len(orm.fieldMappings), len(rm.fieldMappings))
		require.Equal(t, len(orm.csvFieldIndices), len(rm.csvFieldIndices))
		require.Equal(t, len(orm.csvFieldNames), len(rm.csvFieldNames))
	})
	t.Run("Unknown struct field", func(t *testing.T) {
		_, err := m.Adapt(true, OverrideMappings{
			{
				FieldName: "A",
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "field ")
		require.Contains(t, err.Error(), " not found")
	})
	t.Run("Remove field indexed", func(t *testing.T) {
		mm, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:     "Baz",
				CsvFieldIndex: -1,
			},
		})
		require.NoError(t, err)
		rm := mm.(*mapper[testStruct])
		require.NotEqual(t, len(orm.fieldMappings), len(rm.fieldMappings))
		require.NotEqual(t, len(orm.csvFieldIndices), len(rm.csvFieldIndices))
		require.Equal(t, len(orm.csvFieldNames), len(rm.csvFieldNames))
	})
	t.Run("Remove field named", func(t *testing.T) {
		mm, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:    "Bar",
				CsvFieldName: "-bar",
			},
		})
		require.NoError(t, err)
		rm := mm.(*mapper[testStruct])
		require.NotEqual(t, len(orm.fieldMappings), len(rm.fieldMappings))
		require.Equal(t, len(orm.csvFieldIndices), len(rm.csvFieldIndices))
		require.NotEqual(t, len(orm.csvFieldNames), len(rm.csvFieldNames))
	})
	t.Run("Re-map named field to indexed", func(t *testing.T) {
		mm, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:     "Bar",
				CsvFieldIndex: 5,
			},
		})
		require.NoError(t, err)
		rm := mm.(*mapper[testStruct])
		require.Equal(t, len(orm.fieldMappings), len(rm.fieldMappings))
		require.NotEqual(t, len(orm.csvFieldIndices), len(rm.csvFieldIndices))
		require.NotEqual(t, len(orm.csvFieldNames), len(rm.csvFieldNames))
	})
	t.Run("Re-map indexed field to new index", func(t *testing.T) {
		mm, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:     "Baz",
				CsvFieldIndex: 5,
			},
		})
		require.NoError(t, err)
		rm := mm.(*mapper[testStruct])
		require.Equal(t, len(orm.fieldMappings), len(rm.fieldMappings))
		require.Equal(t, len(orm.csvFieldIndices), len(rm.csvFieldIndices))
		require.Equal(t, len(orm.csvFieldNames), len(rm.csvFieldNames))
	})
	t.Run("Re-map indexed field to named", func(t *testing.T) {
		mm, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:    "Baz",
				CsvFieldName: "baz",
			},
		})
		require.NoError(t, err)
		rm := mm.(*mapper[testStruct])
		require.Equal(t, len(orm.fieldMappings), len(rm.fieldMappings))
		require.NotEqual(t, len(orm.csvFieldIndices), len(rm.csvFieldIndices))
		require.NotEqual(t, len(orm.csvFieldNames), len(rm.csvFieldNames))
	})
	t.Run("Re-map named field to new name", func(t *testing.T) {
		mm, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:    "Bar",
				CsvFieldName: "bar2",
			},
		})
		require.NoError(t, err)
		rm := mm.(*mapper[testStruct])
		require.Equal(t, len(orm.fieldMappings), len(rm.fieldMappings))
		require.Equal(t, len(orm.csvFieldIndices), len(rm.csvFieldIndices))
		require.Equal(t, len(orm.csvFieldNames), len(rm.csvFieldNames))
	})
	t.Run("Add index mapping for bad field type", func(t *testing.T) {
		_, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:     "BadType",
				CsvFieldIndex: 1,
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "struct field unsupported type:")
	})
	t.Run("Add name mapping for bad field type", func(t *testing.T) {
		_, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:    "BadType",
				CsvFieldName: "whoops",
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "struct field unsupported type:")
	})
}
