package csvamp

import (
	"errors"
	"github.com/go-andiamo/csvamp/csv"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestReaderContext_Read(t *testing.T) {
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

func TestReaderContext_Read_WithIOReader(t *testing.T) {
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

func TestReaderContext_Read_RawDataBytes(t *testing.T) {
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
	ctx := m.Reader(strings.NewReader(data), nil)

	result, err := ctx.Read()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 2, result.Line)
	require.Equal(t, []string{"Aaa", "Bbb", "Ccc"}, result.Raw)
	require.Equal(t, "Aaa,\"Bbb\",Ccc\n", string(result.RawData))
}

func TestReaderContext_Read_WithPostProcessor(t *testing.T) {
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
	ctx := m.Reader(strings.NewReader(data), func(row *testStruct) error {
		row.Line = 0 - row.Line
		return nil
	})

	result, err := ctx.Read()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, -2, result.Line)
}

func TestReaderContext_Read_Errors(t *testing.T) {
	t.Run("UnmarshalCSV fails for indexed field", func(t *testing.T) {
		type testStruct struct {
			Foo MyBadString `csv:"[1]"`
		}
		m, err := NewMapper[testStruct]()
		require.NoError(t, err)
		require.NotNil(t, m)

		const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc`
		ctx := m.Reader(strings.NewReader(data), nil)

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
		ctx := m.Reader(strings.NewReader(data), nil)

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

func TestReaderContext_ReadAll(t *testing.T) {
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

func TestReaderContext_ReadAll_Errors(t *testing.T) {
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
	ctx := m.Reader(strings.NewReader(data), func(row *testStruct) error {
		return errors.New("fooey")
	})
	_, err = ctx.ReadAll()
	require.Error(t, err)
}

func TestReaderContext_ReadAll_WithErrorHandler(t *testing.T) {
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
	eh := &testErrorHandler{}
	ctx := m.Reader(strings.NewReader(data), func(row *testStruct) error {
		return errors.New("fooey")
	}).WithErrorHandler(eh)
	_, err = ctx.ReadAll()
	require.NoError(t, err)
	require.Len(t, eh.errs, 2)
	require.Len(t, eh.lines, 2)
	require.Equal(t, 2, eh.lines[0])
	require.Equal(t, 3, eh.lines[1])
}

func TestReaderContext_Iterate(t *testing.T) {
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
	ctx := m.Reader(strings.NewReader(data), nil)

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

func TestReaderContext_Iterate_Errors(t *testing.T) {
	type testStruct struct {
		Line int `csv:"[line]"`
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	ctx := m.Reader(strings.NewReader(data), func(row *testStruct) error {
		return errors.New("fooey")
	})

	result := make([]testStruct, 0)
	err = ctx.Iterate(func(v testStruct) (bool, error) {
		result = append(result, v)
		return true, nil
	})
	require.Error(t, err)
}

func TestReaderContext_Iterate_WithErrorHandler(t *testing.T) {
	type testStruct struct {
		Line int `csv:"[line]"`
	}
	m, err := NewMapper[testStruct]()
	require.NoError(t, err)
	require.NotNil(t, m)

	const data = `Foo,Bar,Baz
Aaa,"Bbb",Ccc
Ddd,Eee,Fff`
	eh := &testErrorHandler{}
	ctx := m.Reader(strings.NewReader(data), func(row *testStruct) error {
		return errors.New("fooey")
	}).WithErrorHandler(eh)

	result := make([]testStruct, 0)
	err = ctx.Iterate(func(v testStruct) (bool, error) {
		result = append(result, v)
		return true, nil
	})
	require.NoError(t, err)
	require.Len(t, eh.errs, 2)
	require.Len(t, eh.lines, 2)
	require.Equal(t, 2, eh.lines[0])
	require.Equal(t, 3, eh.lines[1])
}

func TestReaderContext_Read_Unmarshaler(t *testing.T) {
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
		ctx := m.Reader(strings.NewReader(data), nil)

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
		ctx := m.Reader(strings.NewReader(data), nil)

		result, err := ctx.Read()
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, MyString("Aaa"), *result.Foo)
	})
}

func TestReaderContext_Read_NamedField(t *testing.T) {
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
		ctx := m.Reader(strings.NewReader(data), nil)

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
		ctx := m.Reader(strings.NewReader(data), nil)

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

type testErrorHandler struct {
	errs  []error
	lines []int
}

func (h *testErrorHandler) Handle(err error, line int) error {
	h.errs = append(h.errs, err)
	h.lines = append(h.lines, line)
	return nil
}
