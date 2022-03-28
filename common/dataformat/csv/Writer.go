package csv

import (
	"encoding/csv"
	"io"
	"path/filepath"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

type Writer struct {
	destination *csv.Writer
	// If closer != nil, then the writer is responsible for closing the file
	closer io.Closer
	err    errors.Error
}

func NewWriter(destination io.Writer) *Writer {
	return &Writer{
		destination: NewProjectCsvWriter(destination),
	}
}

// Creates a new file writer writing to the file on the given path. If multiple path elements are passed, they are joined with filepath.Join.
func NewFileWriter(path ...string) *Writer {
	file := utils.OpenFileLazy(filepath.Join(path...))
	return &Writer{
		destination: NewProjectCsvWriter(file),
		closer:      file,
	}
}

func (w *Writer) WithSeparator(separator rune) *Writer {
	w.destination.Comma = separator
	return w
}

func (w *Writer) WriteRecord(record []string) errors.Error {
	if w.err != nil {
		return w.err
	} else if w.destination == nil {
		return errors.New(CsvErrorTitle, "Destination not defined")
	} else if err := w.destination.Write(record); err != nil {
		return errors.Wrap(err, CsvErrorTitle, "Could not write to csv output file")
	}
	return nil
}

func (w *Writer) WriteAllRecords(records [][]string) errors.Error {
	defer w.Close()
	for _, record := range records {
		if err := w.WriteRecord(record); err != nil {
			w.err = err
			return err
		}
	}
	if w.destination.Flush(); w.destination.Error() != nil {
		return errors.Wrap(w.destination.Error(), CsvErrorTitle, "Could not write to csv output file")
	}
	return nil
}

func (w *Writer) Close() {
	if w.closer != nil {
		w.closer.Close()
	}
}
