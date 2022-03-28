package csv

import (
	"encoding/csv"
	"io"
	"path/filepath"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

type Reader struct {
	source *csv.Reader
	// If closer != nil, then the writer is responsible for closing the file
	closer io.Closer
	err    errors.Error
}

func NewReader(destination io.Reader) *Reader {
	return &Reader{
		source: NewProjectCsvReader(destination),
	}
}

// Creates a new file reader reading the file on the given path. If multiple path elements are passed, they are joined with filepath.Join.
func NewFileReader(path ...string) *Reader {
	file := utils.OpenFileLazy(filepath.Join(path...))
	return &Reader{
		source: NewProjectCsvReader(file),
		closer: utils.OpenFileLazy(filepath.Join(path...)),
	}
}

func (r *Reader) WithSeparator(separator rune) *Reader {
	r.source.Comma = separator
	return r
}

func (r *Reader) ReadRecord() ([]string, errors.Error) {
	if r.err != nil {
		return nil, r.err
	} else if r.source == nil {
		return nil, errors.New(CsvErrorTitle, "Destination not defined")
	} else if record, err := r.source.Read(); err != nil {
		if err == io.EOF {
			return nil, errors.WrapById(err, errors.EOF)
		}
		return nil, errors.Wrap(err, CsvErrorTitle, "Could not read from csv input file")
	} else {
		return record, nil
	}
}

func (r *Reader) ReadAllRecords() ([][]string, errors.Error) {
	defer r.Close()
	rows := make([][]string, 0, 8)
	for {
		if record, err := r.ReadRecord(); err != nil {
			if err.Is(errors.EOF) {
				return rows, nil
			}
			return nil, err
		} else {
			rows = append(rows, record)
		}
	}
}

func (r *Reader) Close() {
	if r.closer != nil {
		r.closer.Close()
	}
}
