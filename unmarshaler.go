package csvamp

// CsvUnmarshaler is the interface implemented by an object that can
// unmarshal a string CSV field representation of itself.
//
// This is effectively the same as encoding.TextUnmarshaler, but provides
// the field as a string along with the entire record
type CsvUnmarshaler interface {
	UnmarshalCSV(val string, record []string) error
}

// CsvQuotedUnmarshaler is the interface implemented by an object that can
// unmarshal a string CSV field representation of itself.
//
// CsvQuotedUnmarshaler is similar to CsvUnmarshaler, except that it informs
// whether the value string (which may be empty) was quoted
type CsvQuotedUnmarshaler interface {
	UnmarshalQuotedCSV(val string, quoted bool, record []string) error
}
