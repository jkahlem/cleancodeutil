package excel

import (
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStructWithHeaders struct {
	Name string `excel:"NAME"`
	// No tag defined: use empty header
	FieldWithEmptyHeader string
	Number               int    `excel:"number"`
	Text                 string `excel:"Text"`
}

func TestBuildHeaderByStruct(t *testing.T) {
	// when
	layout, err := buildLayoutByStruct(TestStructWithHeaders{})

	// then
	assert.NoError(t, err)
	utils.AssertStringSlice(t, getHeaderStringsFromLayout(layout), "NAME", "", "number", "Text")
}

func TestFromRecords(t *testing.T) {
	// given
	records := [][]string{
		{"0", "1"},
		{"X", "Y", "Z"},
	}
	layout := NewLayout().WithColumns("Header 1", "Header 2", "Header 3").Build()
	sheetName := "Sheet"

	// when
	file, err := fromRecords(records, layout, sheetName)
	actualRows, _ := file.GetRows(sheetName)

	// then
	assert.NoError(t, err)
	assert.Equal(t, 3, len(actualRows))
	utils.AssertStringSlice(t, actualRows[0], "Header 1", "Header 2", "Header 3")
	utils.AssertStringSlice(t, actualRows[1], "0", "1")
	utils.AssertStringSlice(t, actualRows[2], "X", "Y", "Z")
}

func TestColumnIdentifierGeneration(t *testing.T) {
	assert.Equal(t, "A", getColumnIdentifier(0))
	assert.Equal(t, "Z", getColumnIdentifier(25))
	assert.Equal(t, "AA", getColumnIdentifier(26))
	assert.Equal(t, "AB", getColumnIdentifier(27))
	assert.Equal(t, "BA", getColumnIdentifier(52))
}
