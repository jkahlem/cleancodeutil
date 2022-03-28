// Provides the CSV structures used in the application and utility functions
// for working with csv files/structures.
package csv

import (
	"encoding/csv"
	"io"
	"strings"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
)

const CsvErrorTitle string = "CSV Error"

// Creates a Writer for csv files using the seperator defined in the configuration.
func NewProjectCsvWriter(w io.Writer) *csv.Writer {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = configuration.CsvSeperator()
	return csvWriter
}

// Creates a Reader for csv files using the seperator defined in the configuration.
func NewProjectCsvReader(r io.Reader) *csv.Reader {
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

// Writes the records to the given file.
func WriteRecordsToTarget(target io.Writer, records [][]string) errors.Error {
	writer := NewProjectCsvWriter(target)
	if err := writer.WriteAll(records); err != nil {
		return errors.Wrap(err, CsvErrorTitle, "Could not save data to CSV file")
	}
	writer.Flush()
	return nil
}

// Returns true if a list value of a csv record is empty
func IsEmptyList(list []string) bool {
	return len(list) == 0 || len(list) == 1 && list[0] == ""
}
