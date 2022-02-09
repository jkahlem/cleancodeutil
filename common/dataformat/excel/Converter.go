package excel

import (
	"fmt"
	"reflect"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"

	"github.com/xuri/excelize/v2"
)

const DefaultSheetName = "Sheet"

// Reads a .csv file from the given path and builds an excel file from it. The structType defines a struct type, which should be used as reference for creating
// the right headers. The struct's type definition can define the header values using the "excel" tag.
func FromCSV(filePath string, structType interface{}) (*excelize.File, errors.Error) {
	if layout, err := buildLayoutByStruct(structType); err != nil {
		return nil, err
	} else {
		return FromCSVWithLayout(filePath, layout)
	}
}

// Reads a .csv file from the given path and builds an excel file from it, while adding the given headers to it.
func FromCSVWithLayout(filePath string, layout Layout) (*excelize.File, errors.Error) {
	if records, err := csv.ReadRecords(filePath); err != nil {
		return nil, err
	} else {
		return fromRecords(records, layout, DefaultSheetName)
	}
}

func buildLayoutByStruct(structType interface{}) (Layout, errors.Error) {
	reflected := utils.UnwrapType(reflect.TypeOf(structType))
	if reflected == nil {
		return Layout{}, errors.New("Excel Error", "Could not identify struct type")
	} else if reflected.Kind() != reflect.Struct {
		return Layout{}, errors.New("Excel Error", "Expected a struct type passed.")
	}

	header := make([]string, 0, reflected.NumField())
	for i := 0; i < reflected.NumField(); i++ {
		header = append(header, reflected.Field(i).Tag.Get("excel"))
	}
	return NewLayout().WithColumns(header...).Build(), nil
}

// Builds an excel file from the given records.
func fromRecords(records [][]string, layout Layout, sheetName string) (*excelize.File, errors.Error) {
	if len(records) == 0 {
		return nil, nil
	}

	output := excelize.NewFile()
	sheet := output.NewSheet(sheetName)
	header := getHeaderStringsFromLayout(layout)
	rowId := 1
	if header != nil {
		if err := addRowToExcelFile(output, sheetName, rowId, header...); err != nil {
			return nil, err
		}
		rowId++
	}
	for i, row := range records {
		if row == nil {
			continue
		} else if err := addRowToExcelFile(output, sheetName, rowId+i, row...); err != nil {
			return nil, err
		}
	}

	output.SetActiveSheet(sheet)
	return output, nil
}

func getHeaderStringsFromLayout(layout Layout) []string {
	header := make([]string, 0, len(layout.Columns))
	for _, col := range layout.Columns {
		header = append(header, col.Header)
	}
	return header
}

// Adds the given row to the excel file.
func addRowToExcelFile(output *excelize.File, sheet string, rowId int, values ...string) errors.Error {
	for index, value := range values {
		if len(value) > 0 {
			if err := output.SetCellValue(sheet, fmt.Sprintf("%s%d", getColumnIdentifier(index), rowId), value); err != nil {
				return errors.Wrap(err, "Excel Error", fmt.Sprintf("Could not add row to excel file for %s%d (value: %v)", getColumnIdentifier(index), rowId, value))
			}
		}
	}
	return nil
}

// Returns the identifier for an excel column with the given (zero-based) index, e.g. 0 -> "A", 1 -> "B", ..., 25 -> "Z", 26 -> "AA", 27 -> "AB" etc.
func getColumnIdentifier(index int) string {
	chr := string(rune((index % 26) + int('A')))
	if index >= 26 {
		return getColumnIdentifier((index-(index%26))/26-1) + chr
	}
	return chr
}
