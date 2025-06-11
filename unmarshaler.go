package csvamp

// CsvUnmarshaler is the interface implemented by an object that can
// unmarshal a string CSV field representation of itself.
//
// This is effectively the same as encoding.TextUnmarshaler, but provides
// the field as a string along with the entire record
type CsvUnmarshaler interface {
	UnmarshalCSV(val string, record []string) error
}
