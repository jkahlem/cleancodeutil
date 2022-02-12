package excel

import (
	"fmt"
	"returntypes-langserver/common/debug/errors"

	"github.com/xuri/excelize/v2"
)

const DefaultSheetName = "Sheet"
const ExcelHeaderTag = "excel"

// Returns the identifier for an excel cell by zero-based index.
//   getCellIdentifier(0, 0) // "A1"
//   getCellIdentifier(26, 100) // "AA101"
func getCellIdentifier(colIndex, rowIndex int) string {
	return fmt.Sprintf("%s%d", getColumnIdentifier(colIndex), rowIndex+1)
}

// Returns the identifier for an excel column with the given (zero-based) index, e.g. 0 -> "A", 1 -> "B", ..., 25 -> "Z", 26 -> "AA", 27 -> "AB" etc.
func getColumnIdentifier(index int) string {
	chr := string(rune((index % 26) + int('A')))
	if index >= 26 {
		return getColumnIdentifier((index-(index%26))/26-1) + chr
	}
	return chr
}

func FromCSV(string, interface{}) (*excelize.File, errors.Error) {
	// TODO: This is the outdated functionality which should be replaced by the new streams. (Remove this function)
	return nil, errors.New("Error", "Currently not implemented.")
}
