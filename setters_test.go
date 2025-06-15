package csvamp

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestBuildSetter(t *testing.T) {
	testCases := []struct {
		sample    any
		expectErr bool
	}{
		{
			sample: struct {
				Foo MyString
			}{},
		},
		{
			sample: struct {
				Foo *MyString
			}{},
		},
		{
			sample: struct {
				Foo MyText
			}{},
		},
		{
			sample: struct {
				Foo *MyText
			}{},
		},
		{
			sample: struct {
				Foo bool
			}{},
		},
		{
			sample: struct {
				Foo *bool
			}{},
		},
		{
			sample: struct {
				Foo int
			}{},
		},
		{
			sample: struct {
				Foo *int
			}{},
		},
		{
			sample: struct {
				Foo int8
			}{},
		},
		{
			sample: struct {
				Foo *int8
			}{},
		},
		{
			sample: struct {
				Foo int16
			}{},
		},
		{
			sample: struct {
				Foo *int16
			}{},
		},
		{
			sample: struct {
				Foo int32
			}{},
		},
		{
			sample: struct {
				Foo *int32
			}{},
		},
		{
			sample: struct {
				Foo int64
			}{},
		},
		{
			sample: struct {
				Foo *int64
			}{},
		},
		{
			sample: struct {
				Foo uint
			}{},
		},
		{
			sample: struct {
				Foo *uint
			}{},
		},
		{
			sample: struct {
				Foo uint8
			}{},
		},
		{
			sample: struct {
				Foo *uint8
			}{},
		},
		{
			sample: struct {
				Foo uint16
			}{},
		},
		{
			sample: struct {
				Foo *uint16
			}{},
		},
		{
			sample: struct {
				Foo uint32
			}{},
		},
		{
			sample: struct {
				Foo *uint32
			}{},
		},
		{
			sample: struct {
				Foo uint64
			}{},
		},
		{
			sample: struct {
				Foo *uint64
			}{},
		},
		{
			sample: struct {
				Foo float32
			}{},
		},
		{
			sample: struct {
				Foo *float32
			}{},
		},
		{
			sample: struct {
				Foo float64
			}{},
		},
		{
			sample: struct {
				Foo *float64
			}{},
		},
		{
			sample: struct {
				Foo string
			}{},
		},
		{
			sample: struct {
				Foo *string
			}{},
		},
		{
			sample: struct {
				Foo []string
			}{},
		},
		{
			sample: struct {
				Foo []int
			}{},
			expectErr: true,
		},
		{
			sample: struct {
				Foo struct{}
			}{},
			expectErr: true,
		},
		{
			sample: struct {
				Foo *struct{}
			}{},
			expectErr: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			vo := reflect.TypeOf(tc.sample)
			fld := vo.Field(0)
			fn, err := buildSetter[any]([]int{0}, fld)
			if tc.expectErr {
				require.Error(t, err)
				require.Nil(t, fn)
			} else {
				require.NoError(t, err)
				require.NotNil(t, fn)
			}
		})
	}
}

func TestSetterBool(t *testing.T) {
	type testStruct struct {
		Foo bool
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "true", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, true, tc.Foo)

	err = fn(&tc, "not a bool", false, false, nil)
	require.Error(t, err)

	err = fn(&tc, "", false, false, nil)
	require.Error(t, err)
	tc.Foo = true
	err = fn(&tc, "", false, true, nil)
	require.NoError(t, err)
	require.False(t, tc.Foo)
}

func TestSetterPtrBool(t *testing.T) {
	type testStruct struct {
		Foo *bool
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "true", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, true, *tc.Foo)

	err = fn(&tc, "not a bool", false, false, nil)
	require.Error(t, err)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Nil(t, tc.Foo)
}

func TestSetterInt(t *testing.T) {
	type testStruct struct {
		Foo int
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, 1, tc.Foo)

	err = fn(&tc, "not an int", false, false, nil)
	require.Error(t, err)

	err = fn(&tc, "", false, false, nil)
	require.Error(t, err)
	tc.Foo = 1
	err = fn(&tc, "", false, true, nil)
	require.NoError(t, err)
	require.Equal(t, 0, tc.Foo)
}

func TestSetterPtrInt(t *testing.T) {
	type testStruct struct {
		Foo *int
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, 1, *tc.Foo)

	err = fn(&tc, "not an int", false, false, nil)
	require.Error(t, err)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Nil(t, tc.Foo)
}

func TestSetterInt16(t *testing.T) {
	type testStruct struct {
		Foo int16
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, int16(1), tc.Foo)

	err = fn(&tc, "not an int", false, false, nil)
	require.Error(t, err)
}

func TestSetterPtrInt16(t *testing.T) {
	type testStruct struct {
		Foo *int16
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, int16(1), *tc.Foo)

	err = fn(&tc, "not an int", false, false, nil)
	require.Error(t, err)
}

func TestSetterUint(t *testing.T) {
	type testStruct struct {
		Foo uint
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, uint(1), tc.Foo)

	err = fn(&tc, "not an int", false, false, nil)
	require.Error(t, err)

	err = fn(&tc, "", false, false, nil)
	require.Error(t, err)
	tc.Foo = 1
	err = fn(&tc, "", false, true, nil)
	require.NoError(t, err)
	require.Equal(t, uint(0), tc.Foo)
}

func TestSetterPtrUint(t *testing.T) {
	type testStruct struct {
		Foo *uint
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, uint(1), *tc.Foo)

	err = fn(&tc, "not an int", false, false, nil)
	require.Error(t, err)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Nil(t, tc.Foo)
}

func TestSetterUint16(t *testing.T) {
	type testStruct struct {
		Foo uint16
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, uint16(1), tc.Foo)

	err = fn(&tc, "not an int", false, false, nil)
	require.Error(t, err)
}

func TestSetterPtrUint16(t *testing.T) {
	type testStruct struct {
		Foo *uint16
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, uint16(1), *tc.Foo)

	err = fn(&tc, "not an int", false, false, nil)
	require.Error(t, err)
}

func TestSetterFloat(t *testing.T) {
	type testStruct struct {
		Foo float64
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1.1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, 1.1, tc.Foo)

	err = fn(&tc, "not a number", false, false, nil)
	require.Error(t, err)

	err = fn(&tc, "", false, false, nil)
	require.Error(t, err)
	tc.Foo = 1
	err = fn(&tc, "", false, true, nil)
	require.NoError(t, err)
	require.Equal(t, float64(0), tc.Foo)
}

func TestSetterPtrFloat(t *testing.T) {
	type testStruct struct {
		Foo *float64
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "1.1", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, 1.1, *tc.Foo)

	err = fn(&tc, "not a number", false, false, nil)
	require.Error(t, err)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Nil(t, tc.Foo)
}

func TestSetterString(t *testing.T) {
	type testStruct struct {
		Foo string
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, "foo", tc.Foo)
}

func TestSetterPtrString(t *testing.T) {
	type testStruct struct {
		Foo *string
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, "foo", *tc.Foo)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Nil(t, tc.Foo)

	err = fn(&tc, "", true, false, nil)
	require.NoError(t, err)
	require.NotNil(t, tc.Foo)
	require.Equal(t, "", *tc.Foo)
}

func TestSetterSliceString(t *testing.T) {
	type testStruct struct {
		Foo []string
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo,bar", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"foo", "bar"}, tc.Foo)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Empty(t, tc.Foo)
}

func TestSetterUnmarshalCSV(t *testing.T) {
	type testStruct struct {
		Foo MyString
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, MyString("foo"), tc.Foo)
}

func TestSetterPtrUnmarshalCSV(t *testing.T) {
	type testStruct struct {
		Foo *MyString
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, MyString("foo"), *tc.Foo)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Nil(t, tc.Foo)

	err = fn(&tc, "", true, false, nil)
	require.NoError(t, err)
	require.NotNil(t, tc.Foo)
	require.Equal(t, MyString(""), *tc.Foo)
}

func TestSetterUnmarshalCSV_Error(t *testing.T) {
	type testStruct struct {
		Foo MyBadString
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.Error(t, err)
}

func TestSetterPtrUnmarshalCSV_Error(t *testing.T) {
	type testStruct struct {
		Foo *MyBadString
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.Error(t, err)
}

func TestSetterUnmarshalText(t *testing.T) {
	type testStruct struct {
		Foo MyText
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, MyText("foo"), tc.Foo)
}

func TestSetterPtrUnmarshalText(t *testing.T) {
	type testStruct struct {
		Foo *MyText
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, MyText("foo"), *tc.Foo)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Nil(t, tc.Foo)

	err = fn(&tc, "", true, false, nil)
	require.NoError(t, err)
	require.NotNil(t, tc.Foo)
	require.Equal(t, MyText(""), *tc.Foo)
}

func TestSetterUnmarshalText_Errors(t *testing.T) {
	type testStruct struct {
		Foo MyBadText
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.Error(t, err)
}

func TestSetterPtrUnmarshalText_Errors(t *testing.T) {
	type testStruct struct {
		Foo *MyBadText
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.Error(t, err)
}

func TestSetterUnmarshalQuotedCSV(t *testing.T) {
	type testStruct struct {
		Foo MyQuotedString
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, MyQuotedString("foo"), tc.Foo)
	err = fn(&tc, "foo", true, false, nil)
	require.NoError(t, err)
	require.Equal(t, MyQuotedString("\"foo\""), tc.Foo)
}

func TestSetterPtrUnmarshalQuotedCSV(t *testing.T) {
	type testStruct struct {
		Foo *MyQuotedString
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.NoError(t, err)
	require.Equal(t, MyQuotedString("foo"), *tc.Foo)

	err = fn(&tc, "", false, false, nil)
	require.NoError(t, err)
	require.Nil(t, tc.Foo)

	err = fn(&tc, "", true, false, nil)
	require.NoError(t, err)
	require.NotNil(t, tc.Foo)
	require.Equal(t, MyQuotedString(`""`), *tc.Foo)
}

func TestSetterUnmarshalQuotedCSV_Error(t *testing.T) {
	type testStruct struct {
		Foo MyBadQuotedString
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.Error(t, err)
}

func TestSetterPtrUnmarshalQuotedCSV_Error(t *testing.T) {
	type testStruct struct {
		Foo *MyBadQuotedString
	}
	tc := testStruct{}
	fld := reflect.TypeOf(tc).Field(0)
	fn, err := buildSetter[testStruct]([]int{0}, fld)
	require.NoError(t, err)
	err = fn(&tc, "foo", false, false, nil)
	require.Error(t, err)
}

type MyString string

func (my *MyString) UnmarshalCSV(s string, record []string) error {
	*my = MyString(s)
	return nil
}

type MyBadString string

func (my *MyBadString) UnmarshalCSV(s string, record []string) error {
	return errors.New("fooey")
}

type MyText string

func (my *MyText) UnmarshalText(text []byte) error {
	*my = MyText([]byte(text))
	return nil
}

type MyBadText string

func (my *MyBadText) UnmarshalText(text []byte) error {
	return errors.New("fooey")
}

type MyQuotedString string

func (my *MyQuotedString) UnmarshalQuotedCSV(s string, quoted bool, record []string) error {
	if quoted {
		*my = MyQuotedString(`"` + s + `"`)
	} else {
		*my = MyQuotedString(s)
	}
	return nil
}

type MyBadQuotedString string

func (my *MyBadQuotedString) UnmarshalQuotedCSV(s string, quoted bool, record []string) error {
	return errors.New("fooey")
}
