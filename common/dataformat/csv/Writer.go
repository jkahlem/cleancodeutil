package csv

import (
	"encoding/csv"
	"io"
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

func NewFileWriter(path string) *Writer {
	file, err := utils.CreateFile(path)
	if err != nil {
		return &Writer{
			err: errors.Wrap(err, CsvErrorTitle, "Could not save CSV file"),
		}
	}
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

func (w *Writer) Close() {
	if w.closer != nil {
		w.closer.Close()
	}
}

// csv.NewWriter(writer).WithSeparator(",").WriteDatasetRows(datasetRows)
// csv.NewFileWriter(writer)
