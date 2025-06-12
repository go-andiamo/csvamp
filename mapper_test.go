package csvamp

import (
	"github.com/stretchr/testify/require"
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

		om := m.Mappings()
		require.Len(t, om, 2)
		require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, om[0])
		require.Equal(t, OverrideMapping{FieldName: "Bar", CsvFieldIndex: 2}, om[1])

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

		om := m.Mappings()
		require.Len(t, om, 2)
		require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, om[0])
		require.Equal(t, OverrideMapping{FieldName: "Bar", CsvFieldName: "bar"}, om[1])

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
	rm := m.(*mapper[testStruct])
	mappings := m.Mappings()
	require.Len(t, mappings, 3)
	require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, mappings[0])
	require.Equal(t, OverrideMapping{FieldName: "Baz", CsvFieldIndex: 2}, mappings[1])
	require.Equal(t, OverrideMapping{FieldName: "Bar", CsvFieldName: "bar"}, mappings[2])

	t.Run("Clear", func(t *testing.T) {
		m2, err := m.Adapt(true, nil)
		require.NoError(t, err)
		rm2 := m2.(*mapper[testStruct])
		require.Empty(t, rm2.fieldMappings)
		require.Empty(t, rm2.csvFieldIndices)
		require.Empty(t, rm2.csvFieldNames)

		mappings2 := m2.Mappings()
		require.Len(t, mappings2, 0)
	})
	t.Run("Bad option", func(t *testing.T) {
		_, err := m.Adapt(true, nil, "not a valid option")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown option type:")
	})
	t.Run("Clone but change option", func(t *testing.T) {
		m2, err := m.Adapt(false, nil, IgnoreUnknownFieldNames(true))
		require.NoError(t, err)
		rm2 := m2.(*mapper[testStruct])
		require.True(t, rm2.ignoreUnknownFieldNames)
		require.False(t, rm.ignoreUnknownFieldNames)
		require.Equal(t, len(rm.fieldMappings), len(rm2.fieldMappings))
		require.Equal(t, len(rm.csvFieldIndices), len(rm2.csvFieldIndices))
		require.Equal(t, len(rm.csvFieldNames), len(rm2.csvFieldNames))
		require.Equal(t, mappings, m2.Mappings())
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
		m2, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:     "Baz",
				CsvFieldIndex: -1,
			},
		})
		require.NoError(t, err)
		rm2 := m2.(*mapper[testStruct])
		require.NotEqual(t, len(rm.fieldMappings), len(rm2.fieldMappings))
		require.NotEqual(t, len(rm.csvFieldIndices), len(rm2.csvFieldIndices))
		require.Equal(t, len(rm.csvFieldNames), len(rm2.csvFieldNames))

		mappings2 := m2.Mappings()
		require.Len(t, mappings2, 2)
		require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, mappings2[0])
		require.Equal(t, OverrideMapping{FieldName: "Bar", CsvFieldName: "bar"}, mappings2[1])
	})
	t.Run("Remove field named", func(t *testing.T) {
		m2, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:    "Bar",
				CsvFieldName: "-bar",
			},
		})
		require.NoError(t, err)
		rm2 := m2.(*mapper[testStruct])
		require.NotEqual(t, len(rm.fieldMappings), len(rm2.fieldMappings))
		require.Equal(t, len(rm.csvFieldIndices), len(rm2.csvFieldIndices))
		require.NotEqual(t, len(rm.csvFieldNames), len(rm2.csvFieldNames))

		mappings2 := m2.Mappings()
		require.Len(t, mappings2, 2)
		require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, mappings2[0])
		require.Equal(t, OverrideMapping{FieldName: "Baz", CsvFieldIndex: 2}, mappings2[1])
	})
	t.Run("Re-map named field to indexed", func(t *testing.T) {
		m2, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:     "Bar",
				CsvFieldIndex: 5,
			},
		})
		require.NoError(t, err)
		rm2 := m2.(*mapper[testStruct])
		require.Equal(t, len(rm.fieldMappings), len(rm2.fieldMappings))
		require.NotEqual(t, len(rm.csvFieldIndices), len(rm2.csvFieldIndices))
		require.NotEqual(t, len(rm.csvFieldNames), len(rm2.csvFieldNames))

		mappings2 := m2.Mappings()
		require.Len(t, mappings2, 3)
		require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, mappings2[0])
		require.Equal(t, OverrideMapping{FieldName: "Baz", CsvFieldIndex: 2}, mappings2[1])
		require.Equal(t, OverrideMapping{FieldName: "Bar", CsvFieldIndex: 5}, mappings2[2])
	})
	t.Run("Re-map indexed field to new index", func(t *testing.T) {
		m2, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:     "Baz",
				CsvFieldIndex: 5,
			},
		})
		require.NoError(t, err)
		rm2 := m2.(*mapper[testStruct])
		require.Equal(t, len(rm.fieldMappings), len(rm2.fieldMappings))
		require.Equal(t, len(rm.csvFieldIndices), len(rm2.csvFieldIndices))
		require.Equal(t, len(rm.csvFieldNames), len(rm2.csvFieldNames))

		mappings2 := m2.Mappings()
		require.Len(t, mappings2, 3)
		require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, mappings2[0])
		require.Equal(t, OverrideMapping{FieldName: "Baz", CsvFieldIndex: 5}, mappings2[1])
		require.Equal(t, OverrideMapping{FieldName: "Bar", CsvFieldName: "bar"}, mappings2[2])
	})
	t.Run("Re-map indexed field to named", func(t *testing.T) {
		m2, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:    "Baz",
				CsvFieldName: "baz",
			},
		})
		require.NoError(t, err)
		rm2 := m2.(*mapper[testStruct])
		require.Equal(t, len(rm.fieldMappings), len(rm2.fieldMappings))
		require.NotEqual(t, len(rm.csvFieldIndices), len(rm2.csvFieldIndices))
		require.NotEqual(t, len(rm.csvFieldNames), len(rm2.csvFieldNames))

		mappings2 := m2.Mappings()
		require.Len(t, mappings2, 3)
		require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, mappings2[0])
		require.Equal(t, OverrideMapping{FieldName: "Baz", CsvFieldName: "baz"}, mappings2[1])
		require.Equal(t, OverrideMapping{FieldName: "Bar", CsvFieldName: "bar"}, mappings2[2])
	})
	t.Run("Re-map named field to new name", func(t *testing.T) {
		m2, err := m.Adapt(false, OverrideMappings{
			{
				FieldName:    "Bar",
				CsvFieldName: "bar2",
			},
		})
		require.NoError(t, err)
		rm2 := m2.(*mapper[testStruct])
		require.Equal(t, len(rm.fieldMappings), len(rm2.fieldMappings))
		require.Equal(t, len(rm.csvFieldIndices), len(rm2.csvFieldIndices))
		require.Equal(t, len(rm.csvFieldNames), len(rm2.csvFieldNames))

		mappings2 := m2.Mappings()
		require.Len(t, mappings2, 3)
		require.Equal(t, OverrideMapping{FieldName: "Foo", CsvFieldIndex: 1}, mappings2[0])
		require.Equal(t, OverrideMapping{FieldName: "Baz", CsvFieldIndex: 2}, mappings2[1])
		require.Equal(t, OverrideMapping{FieldName: "Bar", CsvFieldName: "bar2"}, mappings2[2])
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
