package charts

import (
	"fmt"
	"returntypes-langserver/common/utils"

	"github.com/go-echarts/go-echarts/v2/opts"
)

// Creates a new table
func NewTable() *Table {
	return &Table{}
}

// Represents a table chart
type Table struct {
	Headings []string
	Records  [][]string
	ChartID  string
	Title    string
}

// Sets the title of the table
func (t *Table) SetTitle(title string) {
	t.Title = title
}

// Sets the headings of the table
func (t *Table) SetHeadings(headings ...string) {
	t.Headings = headings
}

// Adds a row to the table
func (t *Table) AddRow(args ...interface{}) {
	if len(args) == 0 || len(args) != len(t.Headings) {
		return
	}
	record := make([]string, len(args))
	for i, data := range args {
		record[i] = fmt.Sprintf("%v", data)
	}
	t.addRecord(record)
}

func (t *Table) addRecord(record []string) {
	t.Records = append(t.Records, record)
}

// Function for implementing the Charter interface
func (t *Table) Type() string {
	return "table"
}

// Function for implementing the Charter interface
func (t *Table) GetAssets() opts.Assets {
	return opts.Assets{}
}

// Function for implementing the Charter interface
func (t *Table) Validate() {
	// nothing to validate
	t.ChartID = utils.NewUuid()
}
