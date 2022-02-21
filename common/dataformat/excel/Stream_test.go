package excel

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildHeaderByStruct(t *testing.T) {
	// given
	captor := &InfoCaptor{}
	w := newStructFormatWriter(TestStructWithHeaders{})
	w.SetWriter(captor)

	// when
	err := w.BuildLayout(DefaultLayout())

	// then
	assert.NoError(t, err)
	utils.AssertStringSlice(t, getHeaderStringsFromLayout(captor.layout), "NAME", "", "number", "Text")
}

func TestColumnInsert(t *testing.T) {
	// given
	destination := make([][]string, 0)

	// when
	err := Stream().FromSlice(ABCRow()).WithStaticHeaders("Col0", "Col1", "Col2").InsertColumnsAt(Col(1), "Empty1", "Empty2").ToSlice(&destination)

	// then
	assert.NoError(t, err)

	header := destination[0]
	row := destination[1]
	utils.AssertStringSlice(t, header, "Col0", "Empty1", "Empty2", "Col1", "Col2")
	utils.AssertStringSlice(t, row, "A", "", "", "B", "C")
}

func TestColumnSwap(t *testing.T) {
	// given
	destination := make([][]string, 0)

	// when
	err := Stream().FromSlice(ABCRow()).WithStaticHeaders("Col0", "Col1", "Col2").Swap(Col(1), Col(2)).ToSlice(&destination)

	// then
	assert.NoError(t, err)

	header := destination[0]
	row := destination[1]
	utils.AssertStringSlice(t, header, "Col0", "Col2", "Col1")
	utils.AssertStringSlice(t, row, "A", "C", "B")
}

func TestColumnTransform(t *testing.T) {
	// given
	destination := make([][]string, 0)
	transformer := func(record []string) []string {
		record[1] = "ASD"
		return record
	}

	// when
	err := Stream().FromSlice(ABCRow()).WithStaticHeaders("Col0", "Col1", "Col2").Transform(transformer).ToSlice(&destination)

	// then
	assert.NoError(t, err)

	header := destination[0]
	row := destination[1]
	utils.AssertStringSlice(t, header, "Col0", "Col1", "Col2")
	utils.AssertStringSlice(t, row, "A", "ASD", "C")
}

// This "test" is more of a debugging function to actually generate excel files
func TestExcelFileSaving(t *testing.T) {
	// given
	records := [][]string{
		{"0", "1"},
		{"X", "Y", "Z"},
		{"A", "B", "C"},
	}

	// when
	err := Stream().FromSlice(records).WithStaticHeaders("Header 1", "Header 2", "Header 3").ToFile("test.xlsx")

	// then
	assert.NoError(t, err)
}

// This "test" is more of a debugging function to actually generate excel files
func TestExcelFileSavingByStruct(t *testing.T) {
	// given
	records := [][]string{
		{"0", "1"},
		{"X", "Y", "Z"},
		{"A", "B", "C"},
	}

	// when
	err := Stream().FromSlice(records).WithColumnsFromStruct(TestStructWithHeaders{}).ToFile("test.xlsx")

	// then
	assert.NoError(t, err)
}

func TestChannelLoading(t *testing.T) {
	// given
	destination := make([][]string, 0)
	channel := make(chan []string)
	records := [][]string{
		{"0", "1"},
		{"X", "Y", "Z"},
		{"A", "B", "C"},
	}

	go func() {
		channel <- records[0]
		channel <- records[1]
		channel <- records[2]
		close(channel)
	}()

	// when
	err := Stream().FromChannel(channel).WithStaticHeaders("Col0", "Col1", "Col2").ToSlice(&destination)

	// then
	assert.NoError(t, err)
	assert.Len(t, destination, 4)
	utils.AssertStringSlice(t, destination[0], "Col0", "Col1", "Col2")
}

/*-- Unit test helper --*/

func ABCRow() [][]string {
	return [][]string{
		{"A", "B", "C"},
	}
}

type TestStructWithHeaders struct {
	Name string `excel:"NAME,hide=true"`
	// No tag defined: use empty header
	FieldWithEmptyHeader string
	Number               int    `excel:"number,width=50"`
	Text                 string `excel:"Text"`
}

type InfoCaptor struct {
	layout Layout
}

func (w *InfoCaptor) Write(record []string) errors.Error {
	return nil
}

func (w *InfoCaptor) BuildLayout(layout Layout) errors.Error {
	w.layout = layout
	return nil
}

func (w *InfoCaptor) SetWriter(writer StreamWriter) {}
