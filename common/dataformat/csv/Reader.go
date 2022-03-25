package csv

import (
	"encoding/csv"
	"io"
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

func NewFileReader(path string) *Reader {
	file, err := utils.CreateFile(path)
	if err != nil {
		return &Reader{
			err: errors.Wrap(err, CsvErrorTitle, "Could not save CSV file"),
		}
	}
	return &Reader{
		source: NewProjectCsvReader(file),
		closer: file,
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

func (r *Reader) Close() {
	if r.closer != nil {
		r.closer.Close()
	}
}
