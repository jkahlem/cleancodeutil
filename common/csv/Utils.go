// Provides the CSV structures used in the application and utility functions
// for working with csv files/structures.
package csv

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/errors"
)

const CsvErrorTitle string = "CSV Error"

// Creates a Writer for csv files using the seperator defined in the configuration.
func NewWriter(w io.Writer) *csv.Writer {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = configuration.CsvSeperator()
	return csvWriter
}

// Creates a Reader for csv files using the seperator defined in the configuration.
func NewReader(r io.Reader) *csv.Reader {
	csvReader := csv.NewReader(r)
	csvReader.Comma = configuration.CsvSeperator()
	return csvReader
}

// Turns the array into a string using the csv list seperator defined in the configuration.
func MakeList(str []string) string {
	return strings.Join(str, configuration.CsvListSeperator())
}

// Turns a string into an array using the csv list seperator defined in the configuration.
func SplitList(str string) []string {
	return strings.Split(str, configuration.CsvListSeperator())
}

// Reads a whole csv file and returning all it's records.
func ReadRecords(input string) ([][]string, errors.Error) {
	csvFile, err := os.Open(input)
	if err != nil {
		return nil, errors.Wrap(err, CsvErrorTitle, "Could not open CSV file")
	}
	defer csvFile.Close()

	csvReader := NewReader(csvFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, CsvErrorTitle, "Could not read CSV file")
	}

	return records, nil
}

// Writes the records to the file at the given path.
// The file (and directory) will be created if it does not exist.
func WriteCsvRecords(path string, records [][]string) errors.Error {
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return errors.Wrap(err, CsvErrorTitle, "Could not save CSV file")
	}
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, CsvErrorTitle, "Could not save CSV file")
	}
	defer file.Close()

	if err := WriteRecordsToTarget(file, records); err != nil {
		return err
	}
	return nil
}

// Writes the records to the given file.
func WriteRecordsToTarget(target io.Writer, records [][]string) errors.Error {
	writer := NewWriter(target)
	if err := writer.WriteAll(records); err != nil {
		return errors.Wrap(err, CsvErrorTitle, "Could not save data to CSV file")
	}
	writer.Flush()
	return nil
}

// Returns true if a list value of a csv record is empty
func IsEmptyList(list []string) bool {
	return len(list) == 0 || len(list) == 1 && len(list[0]) == 0
}
