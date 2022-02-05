// Provides the CSV structures used in the application and utility functions
// for working with csv files/structures.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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

// Unmarshals a record using the given type. The order of the struct's fields is used for the value assignment.
// Panics if type is not a struct type.
func unmarshal(record []string, targetType reflect.Type) (reflect.Value, errors.Error) {
	value := reflect.New(targetType).Elem()
	for i := 0; i < targetType.NumField(); i++ {
		if len(record) < i {
			break
		}

		fieldDefinition := targetType.Field(i)
		switch fieldDefinition.Type.Kind() {
		case reflect.String:
			value.Field(i).SetString(record[i])

		case reflect.Slice, reflect.Array:
			if fieldDefinition.Type.Elem().Kind() != reflect.String {
				return value, errors.New(CsvErrorTitle, "Mapping slices/arrays which do not have string type is currently not supported.")
			}
			value.Field(i).Set(reflect.ValueOf(SplitList(record[i])))

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if result, err := strconv.ParseInt(record[i], 10, 32); err != nil {
				return value, errors.Wrap(err, CsvErrorTitle, "Could not unmarshal csv data")
			} else {
				value.Field(i).SetInt(result)
			}

		default:
			return value, errors.New(CsvErrorTitle, fmt.Sprintf("Mapping to the type of field %s (%s) is currently not supported.", fieldDefinition.Name, fieldDefinition.PkgPath))
		}
	}
	return value, nil
}

// Marshals a struct to a record by looking at the types of the struct's field. The order is also determined by the struct's field order.
// Panics if the value is not a struct.
func marshal(target reflect.Value) ([]string, errors.Error) {
	targetType := target.Type()
	record := make([]string, targetType.NumField())
	for i := 0; i < targetType.NumField(); i++ {
		fieldDefinition := targetType.Field(i)

		switch fieldDefinition.Type.Kind() {
		case reflect.String:
			record[i] = target.Field(i).String()

		case reflect.Slice, reflect.Array:
			if casted, ok := (target.Field(i).Interface()).([]string); ok {
				record[i] = MakeList(casted)
			} else {
				return nil, errors.New(CsvErrorTitle, "Mapping slices/arrays which do not have string type is currently not supported.")
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			record[i] = fmt.Sprintf("%d", target.Field(i).Int())

		default:
			return nil, errors.New(CsvErrorTitle, fmt.Sprintf("Mapping to the type of field %s (%s) is currently not supported.", fieldDefinition.Name, fieldDefinition.PkgPath))
		}
	}
	return record, nil
}
