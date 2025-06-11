package csvamp

import (
	"fmt"
	"github.com/go-andiamo/csvamp/csv"
	"io"
	"reflect"
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
	// the postProcessor func, if provided, is used to modify (or validate) the struct after it has been read
	Reader(r io.Reader, postProcessor func(row *T) error) ReaderContext[T]
	// ReaderContext returns a reader context for the mapper using the provided csv.Reader
	//
	// the postProcessor func, if provided, is used to modify (or validate) the struct after it has been read
	ReaderContext(r *csv.Reader, postProcessor func(row *T) error) ReaderContext[T]
	// Adapt creates a new Mapper from this mapper with struct field to CSV fields overridden
	//
	// Options from the original mapper are preserved unless overridden by the provided options
	Adapt(clear bool, mappings OverrideMappings, options ...any) (Mapper[T], error)
}

// ReaderContext is an interface used to actually read structs from a CSV reader
type ReaderContext[T any] interface {
	// Read reads the next CSV line as a struct or returns error io.EOF
	Read() (T, error)
	// ReadAll reads all CSV lines as structs
	ReadAll() ([]T, error)
	// Iterate iterates over CSV lines and calls the provided function with the read struct
	//
	// Iteration continues until the end of the CSV or when the provided function returns false or an error
	Iterate(fn func(T) (bool, error)) error
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
	lineMapper              func(t *T, r *csv.Reader)
	rawMapper               func(t *T, r []string)
	rawDataMapper           func(t *T, r []byte)
	csvFieldIndices         map[int]func(t *T, val string, record []string) error
	csvFieldNames           map[string]func(t *T, val string, record []string) error
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
			default:
				return fmt.Errorf("unknown option type: %T", option)
			}
		}
	}
	return nil
}

func (m *mapper[T]) Reader(r io.Reader, postProcessor func(row *T) error) ReaderContext[T] {
	return m.ReaderContext(csv.NewReader(r), postProcessor)
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
		csvFieldIndices:         make(map[int]func(t *T, val string, record []string) error),
		csvFieldNames:           make(map[string]func(t *T, val string, record []string) error),
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

func cloneMap[K comparable, V any](original map[K]V) map[K]V {
	clone := make(map[K]V, len(original))
	for k, v := range original {
		clone[k] = v
	}
	return clone
}

func (m *mapper[T]) mapStruct() error {
	m.fieldIndex = 1
	m.csvFieldNames = make(map[string]func(t *T, val string, record []string) error)
	m.csvFieldIndices = make(map[int]func(t *T, val string, record []string) error)
	m.fieldMappings = make(map[string]any)
	m.fieldIndices = make(map[string][]int)
	var t T
	to := reflect.TypeOf(t)
	if to.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct but got %T", t)
	}
	return m.visitStructFields(to, nil)
}

func (m *mapper[T]) visitStructFields(to reflect.Type, fieldPath []int) (err error) {
	for i := 0; i < to.NumField(); i++ {
		fld := to.Field(i)
		if !fld.IsExported() {
			continue
		}
		currentPath := append(fieldPath, i)
		if fld.Anonymous && fld.Type.Kind() == reflect.Struct {
			if err = m.visitStructFields(fld.Type, currentPath); err != nil {
				return err
			}
			continue
		}
		m.fieldIndices[fld.Name] = append([]int{}, currentPath...)
		if tag, ok := fld.Tag.Lookup(csvTagName); ok {
			switch tag {
			case csvTagLine:
				if fld.Type.Kind() != reflect.Int {
					return fmt.Errorf("field with %q expected to be int (field name: %q)", csvTagLine, fld.Name)
				}
				m.lineMapper = func(t *T, r *csv.Reader) {
					ln, _ := r.FieldPos(1)
					reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetInt(int64(ln))
				}
			case csvTagRaw:
				if fld.Type.Kind() != reflect.Slice || fld.Type.Elem().Kind() != reflect.String {
					return fmt.Errorf("field with %q expected to be slice of strings (field name: %q)", csvTagRaw, fld.Name)
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
					return fmt.Errorf("field with %q expected to be slice of bytes or string (field name: %q)", csvTagRawData, fld.Name)
				}
			default:
				if strings.HasPrefix(tag, "[") && strings.HasSuffix(tag, "]") {
					// specified by index
					tag = tag[1 : len(tag)-1]
					if idx, err := strconv.Atoi(tag); err == nil && idx > 0 {
						if _, exists := m.csvFieldIndices[idx]; exists {
							return fmt.Errorf("field with csv index %d already mapped  (field name: %q)", idx, fld.Name)
						}
						if m.csvFieldIndices[idx], err = buildSetter[T](currentPath, fld); err != nil {
							return err
						}
						m.fieldMappings[fld.Name] = idx
						// following fields follow this index
						m.fieldIndex = idx + 1
					} else {
						return fmt.Errorf("invalid csv field index [%s] (field name: %q)", tag, fld.Name)
					}
				} else if tag != "-" && tag != "" {
					// specified by name
					if _, exists := m.csvFieldNames[tag]; exists {
						return fmt.Errorf("field with csv name %q already mapped  (field name: %q)", tag, fld.Name)
					}
					if m.csvFieldNames[tag], err = buildSetter[T](currentPath, fld); err != nil {
						return err
					}
					m.fieldMappings[fld.Name] = tag
				}
			}
		} else {
			if m.csvFieldIndices[m.fieldIndex], err = buildSetter[T](currentPath, fld); err != nil {
				return err
			}
			m.fieldMappings[fld.Name] = m.fieldIndex
			m.fieldIndex++
		}
	}
	return nil
}

type readerContext[T any] struct {
	reader         *csv.Reader
	mapper         *mapper[T]
	postProcessor  func(row *T) error
	csvHeadersRead bool
	csvHeaders     map[string]int
	csvHeadersErr  error
}

func (rc *readerContext[T]) Read() (t T, err error) {
	var record []string
	if record, err = rc.reader.Read(); err == nil {
		if rc.mapper.lineMapper != nil {
			rc.mapper.lineMapper(&t, rc.reader)
		}
		if rc.mapper.rawMapper != nil {
			rc.mapper.rawMapper(&t, record)
		}
		if rc.mapper.rawDataMapper != nil {
			rc.mapper.rawDataMapper(&t, rc.reader.RawRecord())
		}
		for i, v := range record {
			if fn, ok := rc.mapper.csvFieldIndices[i+1]; ok {
				if err = fn(&t, v, record); err != nil {
					return t, err
				}
			}
		}
		if len(rc.mapper.csvFieldNames) > 0 {
			var headers map[string]int
			if headers, err = rc.getCsvHeaders(); err == nil {
				for name, fn := range rc.mapper.csvFieldNames {
					if idx, ok := headers[name]; ok {
						if idx >= 0 && idx < len(record) {
							if err = fn(&t, record[idx], record); err != nil {
								return t, err
							}
						} else {
							return t, fmt.Errorf("csv field index %d (for header %q) out of range in record", idx+1, name)
						}
					} else if !rc.mapper.ignoreUnknownFieldNames {
						return t, fmt.Errorf("csv header %q not present", name)
					}
				}
			}
		}
		if rc.postProcessor != nil {
			err = rc.postProcessor(&t)
		}
	}
	return t, err
}

func (rc *readerContext[T]) getCsvHeaders() (map[string]int, error) {
	if !rc.csvHeadersRead {
		rc.csvHeadersRead = true
		if hdrs, has := rc.reader.Header(); has {
			rc.csvHeaders = make(map[string]int, len(hdrs))
			for i, h := range hdrs {
				rc.csvHeaders[h] = i
			}
		} else {
			rc.csvHeadersErr = fmt.Errorf("csv headers not present")
		}
	}
	return rc.csvHeaders, rc.csvHeadersErr
}

func (rc *readerContext[T]) ReadAll() (result []T, err error) {
	for err == nil {
		var t T
		t, err = rc.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			continue
		}
		result = append(result, t)
	}
	return result, err
}

func (rc *readerContext[T]) Iterate(fn func(T) (bool, error)) (err error) {
	contd := true
	for contd && err == nil {
		var t T
		t, err = rc.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			continue
		}
		contd, err = fn(t)
	}
	return err
}
