package excel

import (
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestColumnIdentifierGeneration(t *testing.T) {
	assert.Equal(t, "A", GetColumnIdentifier(0))
	assert.Equal(t, "Z", GetColumnIdentifier(25))
	assert.Equal(t, "AA", GetColumnIdentifier(26))
	assert.Equal(t, "AB", GetColumnIdentifier(27))
	assert.Equal(t, "BA", GetColumnIdentifier(52))
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
	collector.Write(records[0], &style)
	collector.Write(records[1], &style)
	collector.Write(records[2], &style)
	collector.Close()
	actualRows, _ := file.GetRows(DefaultSheetName)

	// then
	assert.Equal(t, 3, len(actualRows))
	utils.AssertStringSlice(t, actualRows[0], "Header 1", "Header 2", "Header 3")
	utils.AssertStringSlice(t, actualRows[1], "0", "1")
	utils.AssertStringSlice(t, actualRows[2], "X", "Y", "Z")
}
