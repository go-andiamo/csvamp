package csvamp

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func buildSetter[T any](currentPath []int, fld reflect.StructField) (func(t *T, val string, quoted bool, defEmpties bool, record []string) error, error) {
	fk := fld.Type.Kind()
	if fk == reflect.Ptr {
		return buildPtrSetter[T](currentPath, fld)
	}
	if isUnmarshalerCsvType(fld.Type) {
		return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
			v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
			// create pointer to the field...
			ptr := reflect.New(fld.Type)
			// call UnmarshalCSV on the pointer...
			u := ptr.Interface().(CsvUnmarshaler)
			if err := u.UnmarshalCSV(val, record); err != nil {
				return err
			}
			// assign dereferenced result back to field...
			v.Set(ptr.Elem())
			return nil
		}, nil
	} else if isUnmarshalerQuotedCsvType(fld.Type) {
		return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
			v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
			// create pointer to the field...
			ptr := reflect.New(fld.Type)
			// call UnmarshalQuotedCSV on the pointer...
			u := ptr.Interface().(CsvQuotedUnmarshaler)
			if err := u.UnmarshalQuotedCSV(val, quoted, record); err != nil {
				return err
			}
			// assign dereferenced result back to field...
			v.Set(ptr.Elem())
			return nil
		}, nil
	} else if isUnmarshalerTextType(fld.Type) {
		return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
			v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
			// create pointer to the field...
			ptr := reflect.New(fld.Type)
			// call UnmarshalText on the pointer...
			u := ptr.Interface().(encoding.TextUnmarshaler)
			if err := u.UnmarshalText([]byte(val)); err != nil {
				return err
			}
			// assign dereferenced result back to field...
			v.Set(ptr.Elem())
			return nil
		}, nil
	}
	switch fk {
	case reflect.Bool:
		return setterBool[T](currentPath), nil
	case reflect.Int:
		return setterInt[T](currentPath, 0), nil
	case reflect.Int8:
		return setterInt[T](currentPath, 8), nil
	case reflect.Int16:
		return setterInt[T](currentPath, 16), nil
	case reflect.Int32:
		return setterInt[T](currentPath, 32), nil
	case reflect.Int64:
		return setterInt[T](currentPath, 64), nil
	case reflect.Uint:
		return setterUint[T](currentPath, 0), nil
	case reflect.Uint8:
		return setterUint[T](currentPath, 8), nil
	case reflect.Uint16:
		return setterUint[T](currentPath, 16), nil
	case reflect.Uint32:
		return setterUint[T](currentPath, 32), nil
	case reflect.Uint64:
		return setterUint[T](currentPath, 64), nil
	case reflect.Float32:
		return setterFloat[T](currentPath, 32), nil
	case reflect.Float64:
		return setterFloat[T](currentPath, 64), nil
	case reflect.String:
		return setterString[T](currentPath), nil
	case reflect.Slice:
		if fld.Type.Elem().Kind() == reflect.String {
			return setterSliceString[T](currentPath), nil
		}
	}
	return nil, fmt.Errorf("struct field unsupported type: %s", fk.String())
}

func setterBool[T any](currentPath []int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		if defEmpties && val == "" {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetBool(false)
		} else if b, err := strconv.ParseBool(val); err != nil {
			return fmt.Errorf("cannot convert value %q to bool", val)
		} else {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetBool(b)
		}
		return nil
	}
}

func setterInt[T any](currentPath []int, bitSize int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		if defEmpties && val == "" {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetInt(0)
		} else if v, err := strconv.ParseInt(val, 10, bitSize); err != nil {
			if bitSize == 0 {
				return fmt.Errorf("cannot convert value %q to int", val)
			} else {
				return fmt.Errorf("cannot convert value %q to int%d", val, bitSize)
			}
		} else {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetInt(v)
		}
		return nil
	}
}

func setterUint[T any](currentPath []int, bitSize int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		if defEmpties && val == "" {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetUint(0)
		} else if v, err := strconv.ParseUint(val, 10, bitSize); err != nil {
			if bitSize == 0 {
				return fmt.Errorf("cannot convert value %q to uint", val)
			} else {
				return fmt.Errorf("cannot convert value %q to uint%d", val, bitSize)
			}
		} else {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetUint(v)
		}
		return nil
	}
}

func setterFloat[T any](currentPath []int, bitSize int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		if defEmpties && val == "" {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetFloat(0)
		} else if v, err := strconv.ParseFloat(val, bitSize); err != nil {
			return fmt.Errorf("cannot convert value %q to float%d", val, bitSize)
		} else {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetFloat(v)
		}
		return nil
	}
}

func setterString[T any](currentPath []int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		reflect.ValueOf(t).Elem().FieldByIndex(currentPath).SetString(val)
		return nil
	}
}

func setterSliceString[T any](currentPath []int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		if val == "" {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).Set(reflect.ValueOf([]string{}))
		} else {
			reflect.ValueOf(t).Elem().FieldByIndex(currentPath).Set(reflect.ValueOf(strings.Split(val, ",")))
		}
		return nil
	}
}

func buildPtrSetter[T any](currentPath []int, fld reflect.StructField) (func(t *T, val string, quoted bool, defEmpties bool, record []string) error, error) {
	if isUnmarshalerCsvType(fld.Type) {
		return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
			v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
			if val == "" && !quoted {
				v.Set(reflect.Zero(v.Type()))
				return nil
			}
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			u := v.Interface().(CsvUnmarshaler)
			return u.UnmarshalCSV(val, record)
		}, nil
	} else if isUnmarshalerQuotedCsvType(fld.Type) {
		return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
			v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
			if val == "" && !quoted {
				v.Set(reflect.Zero(v.Type()))
				return nil
			}
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			u := v.Interface().(CsvQuotedUnmarshaler)
			return u.UnmarshalQuotedCSV(val, quoted, record)
		}, nil
	} else if isUnmarshalerTextType(fld.Type) {
		return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
			v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
			if val == "" && !quoted {
				v.Set(reflect.Zero(v.Type()))
				return nil
			}
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			u := v.Interface().(encoding.TextUnmarshaler)
			return u.UnmarshalText([]byte(val))
		}, nil
	}
	fk := fld.Type.Elem().Kind()
	switch fk {
	case reflect.Bool:
		return setterPtrBool[T](currentPath), nil
	case reflect.Int:
		return setterPtrInt[T](currentPath, 0), nil
	case reflect.Int8:
		return setterPtrInt[T](currentPath, 8), nil
	case reflect.Int16:
		return setterPtrInt[T](currentPath, 16), nil
	case reflect.Int32:
		return setterPtrInt[T](currentPath, 32), nil
	case reflect.Int64:
		return setterPtrInt[T](currentPath, 64), nil
	case reflect.Uint:
		return setterPtrUint[T](currentPath, 0), nil
	case reflect.Uint8:
		return setterPtrUint[T](currentPath, 8), nil
	case reflect.Uint16:
		return setterPtrUint[T](currentPath, 16), nil
	case reflect.Uint32:
		return setterPtrUint[T](currentPath, 32), nil
	case reflect.Uint64:
		return setterPtrUint[T](currentPath, 64), nil
	case reflect.Float32:
		return setterPtrFloat[T](currentPath, 32), nil
	case reflect.Float64:
		return setterPtrFloat[T](currentPath, 64), nil
	case reflect.String:
		return setterPtrString[T](currentPath), nil
	}
	return nil, fmt.Errorf("struct field unsupported type: *%s", fk.String())
}

func setterPtrBool[T any](currentPath []int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
		if val == "" {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		if b, err := strconv.ParseBool(val); err != nil {
			return fmt.Errorf("cannot convert value %q to bool", val)
		} else {
			ptr := reflect.New(v.Type().Elem())
			ptr.Elem().SetBool(b)
			v.Set(ptr)
		}
		return nil
	}
}

func setterPtrInt[T any](currentPath []int, bitSize int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
		if val == "" {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		if i, err := strconv.ParseInt(val, 10, bitSize); err != nil {
			if bitSize == 0 {
				return fmt.Errorf("cannot convert value %q to int", val)
			} else {
				return fmt.Errorf("cannot convert value %q to int%d", val, bitSize)
			}
		} else {
			ptr := reflect.New(v.Type().Elem())
			ptr.Elem().SetInt(i)
			v.Set(ptr)
		}
		return nil
	}
}

func setterPtrUint[T any](currentPath []int, bitSize int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
		if val == "" {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		if i, err := strconv.ParseUint(val, 10, bitSize); err != nil {
			if bitSize == 0 {
				return fmt.Errorf("cannot convert value %q to uint", val)
			} else {
				return fmt.Errorf("cannot convert value %q to uint%d", val, bitSize)
			}
		} else {
			ptr := reflect.New(v.Type().Elem())
			ptr.Elem().SetUint(i)
			v.Set(ptr)
		}
		return nil
	}
}

func setterPtrFloat[T any](currentPath []int, bitSize int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
		if val == "" {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		if f, err := strconv.ParseFloat(val, bitSize); err != nil {
			return fmt.Errorf("cannot convert value %q to float%d", val, bitSize)
		} else {
			ptr := reflect.New(v.Type().Elem())
			ptr.Elem().SetFloat(f)
			v.Set(ptr)
		}
		return nil
	}
}

func setterPtrString[T any](currentPath []int) func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
	return func(t *T, val string, quoted bool, defEmpties bool, record []string) error {
		v := reflect.ValueOf(t).Elem().FieldByIndex(currentPath)
		if val == "" && !quoted {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		ptr := reflect.New(v.Type().Elem())
		ptr.Elem().SetString(val)
		v.Set(ptr)
		return nil
	}
}

var unmarshalerCsvType = reflect.TypeOf((*CsvUnmarshaler)(nil)).Elem()

var unmarshalerQuotedCsvType = reflect.TypeOf((*CsvQuotedUnmarshaler)(nil)).Elem()

var unmarshalerTextType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func isUnmarshalerCsvType(t reflect.Type) bool {
	if t.Kind() != reflect.Ptr {
		return reflect.PointerTo(t).Implements(unmarshalerCsvType)
	} else {
		return t.Implements(unmarshalerCsvType)
	}
}

func isUnmarshalerQuotedCsvType(t reflect.Type) bool {
	if t.Kind() != reflect.Ptr {
		return reflect.PointerTo(t).Implements(unmarshalerQuotedCsvType)
	} else {
		return t.Implements(unmarshalerQuotedCsvType)
	}
}

func isUnmarshalerTextType(t reflect.Type) bool {
	if t.Kind() != reflect.Ptr {
		return reflect.PointerTo(t).Implements(unmarshalerTextType)
	} else {
		return t.Implements(unmarshalerTextType)
	}
}

func isUnmarshalerType(t reflect.Type) bool {
	return isUnmarshalerCsvType(t) || isUnmarshalerQuotedCsvType(t) || isUnmarshalerTextType(t)
}
