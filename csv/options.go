package csv

// Comma is an option that can be used for NewReader to set Reader.Comma
type Comma rune

// Comment is an option that can be used for NewReader to set Reader.Comment
type Comment rune

// FieldsPerRecord is an option that can be used for NewReader to set Reader.FieldsPerRecord
type FieldsPerRecord int

// LazyQuotes is an option that can be used for NewReader to set Reader.LazyQuotes
type LazyQuotes bool

// TrimLeadingSpace is an option that can be used for NewReader to set Reader.TrimLeadingSpace
type TrimLeadingSpace bool

// NoHeader is an option that can be used for NewReader to set Reader.NoHeader
type NoHeader bool

type ReuseRecord bool
