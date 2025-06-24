package csvamp

import (
	"errors"
	"fmt"
	"github.com/go-andiamo/csvamp/csv"
	"io"
)

// ReaderContext is the interface used to actually read structs from CSV
//
// A reader context is obtained from Mapper.Reader or Mapper.ReaderContext
type ReaderContext[T any] interface {
	// Read reads the next CSV line as a struct or returns error io.EOF
	Read() (T, error)
	// ReadAll reads all CSV lines as structs
	ReadAll() ([]T, error)
	// Iterate iterates over CSV lines and calls the provided function with the read struct
	//
	// Iteration continues until the end of the CSV or when the provided function returns false or an error
	Iterate(fn func(T) (bool, error)) error
	// WithErrorHandler sets the error handler - which can be used to track errors during ReadAll and Iterate
	//
	// Setting an error handler means that errors are reported but don't necessarily halt further reading
	WithErrorHandler(eh ErrorHandler) ReaderContext[T]
	// SupplyHeaders enables CSV headers to be manually supplied
	//
	// Sometimes your csv may not have headers, or you may have already read (and normalised) them
	SupplyHeaders(headers []string) ReaderContext[T]
}

// ErrorHandler is an interface that can be used with ReaderContext.WithErrorHandler
type ErrorHandler interface {
	// Handle handles the error - if it returns the error, then further processing stops
	Handle(err error, line int) error
}

type readerContext[T any] struct {
	reader         *csv.Reader
	mapper         *mapper[T]
	postProcessor  func(row *T) error
	csvHeadersRead bool
	csvHeaders     map[string]int
	csvHeadersErr  error
	errorHandler   ErrorHandler
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
				if err = fn(&t, v, rc.reader.FieldQuoted(i), rc.mapper.defaultEmptyValues, record); err != nil {
					return t, err
				}
			}
		}
		if len(rc.mapper.csvFieldNames) > 0 {
			if err = rc.checkCsvHeaders(); err == nil {
				l := len(record)
				for name, fn := range rc.mapper.csvFieldNames {
					if idx, ok := rc.csvHeaders[name]; ok && idx >= 0 && idx < l {
						if err = fn(&t, record[idx], rc.reader.FieldQuoted(idx), rc.mapper.defaultEmptyValues, record); err != nil {
							return t, err
						}
					} else if !rc.mapper.ignoreUnknownFieldNames {
						return t, fmt.Errorf("csv header %q not present", name)
					}
				}
			}
		}
		if rc.postProcessor != nil && err == nil {
			err = rc.postProcessor(&t)
		}
	}
	return t, err
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
			err = rc.handleError(err)
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
			err = rc.handleError(err)
			continue
		}
		contd, err = fn(t)
		err = rc.handleError(err)
	}
	return err
}

func (rc *readerContext[T]) WithErrorHandler(eh ErrorHandler) ReaderContext[T] {
	rc.errorHandler = eh
	return rc
}

func (rc *readerContext[T]) SupplyHeaders(headers []string) ReaderContext[T] {
	rc.csvHeadersRead = true
	rc.csvHeadersErr = nil
	rc.csvHeaders = make(map[string]int, len(headers))
	for i, h := range headers {
		rc.csvHeaders[h] = i
	}
	return rc
}

func (rc *readerContext[T]) checkCsvHeaders() error {
	if !rc.csvHeadersRead {
		rc.csvHeadersRead = true
		if hdrs, has := rc.reader.Header(); has {
			rc.csvHeaders = make(map[string]int, len(hdrs))
			for i, h := range hdrs {
				rc.csvHeaders[h] = i
			}
		} else {
			rc.csvHeadersErr = errors.New("csv headers not present")
		}
	}
	return rc.csvHeadersErr
}

func (rc *readerContext[T]) handleError(err error) error {
	if err == nil {
		return nil
	} else if rc.errorHandler == nil {
		return &ReaderError{
			Line: rc.reader.CurrentLine(),
			Err:  err,
		}
	}
	return rc.errorHandler.Handle(err, rc.reader.CurrentLine())
}

// ReaderError is the wrapped error returned from ReaderContext.ReadAll / ReaderContext.Iterate
type ReaderError struct {
	Line int
	Err  error
}

func (e *ReaderError) Error() string {
	return fmt.Sprintf("line %d: %v", e.Line, e.Err)
}

func (e *ReaderError) Unwrap() error {
	return e.Err
}
