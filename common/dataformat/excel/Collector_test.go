package excel

import (
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestColumnIdentifierGeneration(t *testing.T) {
	assert.Equal(t, "A", getColumnIdentifier(0))
	assert.Equal(t, "Z", getColumnIdentifier(25))
	assert.Equal(t, "AA", getColumnIdentifier(26))
	assert.Equal(t, "AB", getColumnIdentifier(27))
	assert.Equal(t, "BA", getColumnIdentifier(52))
}

func TestExcelFileGeneration(t *testing.T) {
	// given
	records := [][]string{
		{"Header 1", "Header 2", "Header 3"},
		{"0", "1"},
		{"X", "Y", "Z"},
	}
	file := excelize.NewFile()
	file.SetActiveSheet(file.NewSheet(DefaultSheetName))
	collector := newFileCollectorByExcelFile(file)
	style := Style{}

	// when
	collector.Write(records[0], style)
	collector.Write(records[1], style)
	collector.Write(records[2], style)
	actualRows, _ := file.GetRows(DefaultSheetName)

	// then
	assert.Equal(t, 3, len(actualRows))
	utils.AssertStringSlice(t, actualRows[0], "Header 1", "Header 2", "Header 3")
	utils.AssertStringSlice(t, actualRows[1], "0", "1")
	utils.AssertStringSlice(t, actualRows[2], "X", "Y", "Z")
}
