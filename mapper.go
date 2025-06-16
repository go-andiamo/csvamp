package csvamp

import (
	"fmt"
	"github.com/go-andiamo/csvamp/csv"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const (
	csvTagName    = "csv"
	csvTagLine    = "[line]"
	csvTagRaw     = "[raw]"
	csvTagRawData = "[rawData]"
)

// Mapper is an interface for mapping structs onto CSV
//
// Use NewMapper / MustNewMapper to create a new Mapper
type Mapper[T any] interface {
	// Reader returns a reader context for the mapper using the provided io.Reader
	//
	// the postProcessor func, if provided, can be used to validate (or modify) the struct after it has been read
	//
	// the options can be any of csv.Comma, csv.Comment, csv.FieldsPerRecord, csv.LazyQuotes, csv.TrimLeadingSpace or csv.NoHeader
	Reader(r io.Reader, postProcessor func(row *T) error, options ...any) ReaderContext[T]
	// ReaderContext returns a reader context for the mapper using the provided csv.Reader
	//
	// the postProcessor func, if provided, can be used to validate (or modify) the struct after it has been read
	ReaderContext(r *csv.Reader, postProcessor func(row *T) error) ReaderContext[T]
	// Adapt creates a new Mapper from this mapper with struct field to CSV fields overridden
	//
	// Options from the original mapper are preserved unless overridden by the provided options
	Adapt(clear bool, mappings OverrideMappings, options ...any) (Mapper[T], error)
	// Mappings returns the current effective struct field to CSV field mappings
	Mappings() OverrideMappings
}

// NewMapper creates a new struct to CSV Mapper for the specified generic struct type
func NewMapper[T any](options ...any) (Mapper[T], error) {
	result := &mapper[T]{}
	if err := result.setOptions(options...); err != nil {
		return nil, err
	}
	if err := result.mapStruct(); err != nil {
		return nil, err
	}
	return result, nil
}

// MustNewMapper is the same as NewMapper - except that it panics in case of error
func MustNewMapper[T any](options ...any) Mapper[T] {
	m, err := NewMapper[T](options...)
	if err != nil {
		panic(err)
	}
	return m
}

type mapper[T any] struct {
	ignoreUnknownFieldNames bool
	defaultEmptyValues      bool
	lineMapper              func(t *T, r *csv.Reader)
	rawMapper               func(t *T, r []string)
	rawDataMapper           func(t *T, r []byte)
	csvFieldIndices         map[int]func(t *T, val string, quoted bool, defEmpties bool, record []string) error
	csvFieldNames           map[string]func(t *T, val string, quoted bool, defEmpties bool, record []string) error
	fieldMappings           map[string]any // int value is csv index, string value is csv header
	fieldIndices            map[string][]int
	fieldIndex              int // used only while inspecting struct fields
}

func (m *mapper[T]) setOptions(options ...any) error {
	for _, o := range options {
		if o != nil {
			switch option := o.(type) {
			case IgnoreUnknownFieldNames:
				m.ignoreUnknownFieldNames = bool(option)
			case DefaultEmptyValues:
				m.defaultEmptyValues = bool(option)
			default:
				return fmt.Errorf("unknown option type: %T", option)
			}
		}
	}
	return nil
}

func (m *mapper[T]) Reader(r io.Reader, postProcessor func(row *T) error, csvOptions ...any) ReaderContext[T] {
	return m.ReaderContext(csv.NewReader(r, csvOptions...), postProcessor)
}

func (m *mapper[T]) ReaderContext(r *csv.Reader, postProcessor func(row *T) error) ReaderContext[T] {
	return &readerContext[T]{
		reader:        r,
		mapper:        m,
		postProcessor: postProcessor,
	}
}

func (m *mapper[T]) Adapt(clear bool, mappings OverrideMappings, options ...any) (Mapper[T], error) {
	result := &mapper[T]{
		ignoreUnknownFieldNames: m.ignoreUnknownFieldNames,
		lineMapper:              m.lineMapper,
		rawMapper:               m.rawMapper,
		rawDataMapper:           m.rawDataMapper,
		fieldIndices:            m.fieldIndices,
		csvFieldIndices:         make(map[int]func(t *T, val string, quoted bool, defEmpties bool, record []string) error),
		csvFieldNames:           make(map[string]func(t *T, val string, quoted bool, defEmpties bool, record []string) error),
		fieldMappings:           make(map[string]any),
	}
	if err := result.setOptions(options...); err != nil {
		return nil, err
	}
	if !clear {
		result.csvFieldIndices = cloneMap(m.csvFieldIndices)
		result.csvFieldNames = cloneMap(m.csvFieldNames)
		result.fieldMappings = cloneMap(m.fieldMappings)
	}
	var t T
	to := reflect.TypeOf(t)
	for _, mapping := range mappings {
		fieldPath, ok := m.fieldIndices[mapping.FieldName]
		if !ok {
			return nil, fmt.Errorf("field %q not found", mapping.FieldName)
		}
		switch {
		case mapping.CsvFieldIndex < 0:
			// remove index mapping...
			delete(result.csvFieldIndices, 0-mapping.CsvFieldIndex)
			delete(result.fieldMappings, mapping.FieldName)
		case strings.HasPrefix(mapping.CsvFieldName, "-"):
			// remove name mapping...
			delete(result.csvFieldNames, strings.TrimPrefix(mapping.CsvFieldName, "-"))
			delete(result.fieldMappings, mapping.FieldName)
		case mapping.CsvFieldIndex > 0:
			// re-map by index...
			if exm, ok := result.fieldMappings[mapping.FieldName]; ok {
				switch k := exm.(type) {
				case int:
					delete(result.csvFieldIndices, k)
				case string:
					delete(result.csvFieldNames, k)
				}
			}
			result.fieldMappings[mapping.FieldName] = mapping.CsvFieldIndex
			var err error
			fld := to.FieldByIndex(fieldPath)
			if result.csvFieldIndices[mapping.CsvFieldIndex], err = buildSetter[T](fieldPath, fld); err != nil {
				return nil, err
			}
		case mapping.CsvFieldName != "":
			// re-map by name...
			if exm, ok := result.fieldMappings[mapping.FieldName]; ok {
				switch k := exm.(type) {
				case int:
					delete(result.csvFieldIndices, k)
				case string:
					delete(result.csvFieldNames, k)
				}
			}
			result.fieldMappings[mapping.FieldName] = mapping.CsvFieldName
			var err error
			fld := to.FieldByIndex(fieldPath)
			if result.csvFieldNames[mapping.CsvFieldName], err = buildSetter[T](fieldPath, fld); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func (m *mapper[T]) Mappings() OverrideMappings {
	type temp struct {
		indices []int
		mapping OverrideMapping
	}
	list := make([]temp, 0, len(m.fieldIndices))
	for k, v := range m.fieldIndices {
		if fmv, ok := m.fieldMappings[k]; ok {
			switch mvt := fmv.(type) {
			case int:
				list = append(list, temp{
					indices: v,
					mapping: OverrideMapping{
						FieldName:     k,
						CsvFieldIndex: mvt,
					},
				})
			case string:
				list = append(list, temp{
					indices: v,
					mapping: OverrideMapping{
						FieldName:    k,
						CsvFieldName: mvt,
					},
				})
			}
		}
	}
	sort.Slice(list, func(i, j int) (less bool) {
		a, b := list[i].indices, list[j].indices
		for x := 0; x < len(a) && x < len(b); x++ {
			if a[x] != b[x] {
				less = a[x] < b[x]
				break
			}
		}
		return less
	})
	result := make(OverrideMappings, len(list))
	for i, v := range list {
		result[i] = v.mapping
	}
	return result
}

func cloneMap[K comparable, V any](original map[K]V) map[K]V {
	clone := make(map[K]V, len(original))
	for k, v := range original {
		clone[k] = v
	}
	return clone
}

func (m *mapper[T]) mapStruct() error {
	m.fieldIndex = 1
	m.csvFieldNames = make(map[string]func(t *T, val string, quoted bool, defEmpties bool, record []string) error)
	m.csvFieldIndices = make(map[int]func(t *T, val string, quoted bool, defEmpties bool, record []string) error)
	m.fieldMappings = make(map[string]any)
	m.fieldIndices = make(map[string][]int)
	var t T
	rt := reflect.TypeOf(t)
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct but got %T", t)
	}
	return m.visitStructFields(rt, nil, nil)
}

func (m *mapper[T]) visitStructFields(rt reflect.Type, fieldPath []int, namePath []string) (err error) {
	for i := 0; i < rt.NumField(); i++ {
		fld := rt.Field(i)
		if !fld.IsExported() {
			continue
		}
		currentPath := append(fieldPath, i)
		if fld.Type.Kind() == reflect.Struct {
			if fld.Anonymous {
				if err = m.visitStructFields(fld.Type, currentPath, namePath); err != nil {
					return err
				}
				continue
			} else if !isUnmarshalerType(fld.Type) {
				if _, ok := fld.Tag.Lookup(csvTagName); ok {
					return fmt.Errorf("nested struct field cannot have %q tag", csvTagName)
				}
				if err = m.visitStructFields(fld.Type, currentPath, append(namePath, fld.Name)); err != nil {
					return err
				}
				continue
			}
		}
		fldName := strings.Join(append(namePath, fld.Name), ".")
		m.fieldIndices[fldName] = append([]int{}, currentPath...)
		if tag, ok := fld.Tag.Lookup(csvTagName); ok {
			switch tag {
			case csvTagLine:
				if fld.Type.Kind() != reflect.Int {
					return fmt.Errorf("field with %q expected to be int (field name: %q)", csvTagLine, fldName)
				}
				m.lineMapper = func(t *T, r *csv.Reader) {
					ln, _ := r.FieldPos(1)
					reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetInt(int64(ln))
				}
			case csvTagRaw:
				if fld.Type.Kind() != reflect.Slice || fld.Type.Elem().Kind() != reflect.String {
					return fmt.Errorf("field with %q expected to be slice of strings (field name: %q)", csvTagRaw, fldName)
				}
				m.rawMapper = func(t *T, r []string) {
					reflect.ValueOf(t).Elem().FieldByIndex(currentPath).Set(reflect.ValueOf(r))
				}
			case csvTagRawData:
				if fld.Type.Kind() == reflect.Slice && fld.Type.Elem().Kind() == reflect.Uint8 {
					m.rawDataMapper = func(t *T, r []byte) {
						reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetBytes(r)
					}
				} else if fld.Type.Kind() == reflect.String {
					m.rawDataMapper = func(t *T, r []byte) {
						reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetString(string(r))
					}
				} else {
					return fmt.Errorf("field with %q expected to be slice of bytes or string (field name: %q)", csvTagRawData, fldName)
				}
			default:
				if strings.HasPrefix(tag, "[") && strings.HasSuffix(tag, "]") {
					// specified by index
					tag = tag[1 : len(tag)-1]
					if idx, err := strconv.Atoi(tag); err == nil && idx > 0 {
						if _, exists := m.csvFieldIndices[idx]; exists {
							return fmt.Errorf("field with csv index %d already mapped  (field name: %q)", idx, fldName)
						}
						if m.csvFieldIndices[idx], err = buildSetter[T](currentPath, fld); err != nil {
							return err
						}
						m.fieldMappings[fldName] = idx
						// following fields follow this index
						m.fieldIndex = idx + 1
					} else {
						return fmt.Errorf("invalid csv field index [%s] (field name: %q)", tag, fldName)
					}
				} else if tag != "-" && tag != "" {
					// specified by name
					if _, exists := m.csvFieldNames[tag]; exists {
						return fmt.Errorf("field with csv name %q already mapped  (field name: %q)", tag, fldName)
					}
					if m.csvFieldNames[tag], err = buildSetter[T](currentPath, fld); err != nil {
						return err
					}
					m.fieldMappings[fldName] = tag
				}
			}
		} else {
			if m.csvFieldIndices[m.fieldIndex], err = buildSetter[T](currentPath, fld); err != nil {
				return err
			}
			m.fieldMappings[fldName] = m.fieldIndex
			m.fieldIndex++
		}
	}
	return nil
}
