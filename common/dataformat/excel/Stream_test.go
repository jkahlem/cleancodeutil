package excel

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
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
	channel := NewChannel()
	records := [][]string{
		{"0", "1"},
		{"X", "Y", "Z"},
		{"A", "B", "C"},
	}

	// when
	go func() {
		channel.PutError(Stream().FromChannel(channel).WithStaticHeaders("Col0", "Col1", "Col2").ToSlice(&destination))
	}()

	channel.PutRecord(records[0])
	channel.PutRecord(records[1])
	channel.PutRecord(records[2])
	channel.Close()

	// then
	assert.NoError(t, channel.NextError())
	assert.Len(t, destination, 4)
	utils.AssertStringSlice(t, destination[0], "Col0", "Col1", "Col2")
}

func TestCursorChart(t *testing.T) {
	chart := Chart{
		ChartBase: ChartBase{
			Type: "col",
			Title: &Title{
				Name: "Tokens per parameter list",
			},
			Format: &Format{
				XScale:          1.0,
				YScale:          1.0,
				XOffset:         15,
				YOffset:         10,
				PrintObj:        true,
				LockAspectRatio: false,
				Locked:          false,
			},
			VaryColors: false,
			PlotArea: &PlotArea{
				ShowBubbleSize:  true,
				ShowCatName:     false,
				ShowLeaderLines: false,
				ShowPercent:     true,
				ShowSeriesName:  false,
				ShowVal:         true,
			},
		},
		Series: []Series{
			{
				Categories: []interface{}{"0", "1", "2"},
				Values:     []interface{}{100, 50, 30},
			},
		},
	}

	f := excelize.NewFile()
	f.Path = "test2.xlsx"
	f.NewSheet("Sheet1")
	c := NewCursor(f, "Sheet1")
	c.WriteValues([][]interface{}{{"a", "b", "c"}, {}, {"1", "2", "3"}, {chart}, {"x", "y", "z"}})
	SaveFile(f)
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
