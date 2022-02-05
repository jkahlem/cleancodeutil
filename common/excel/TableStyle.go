package excel

import "github.com/xuri/excelize/v2"

type Layout struct {
	Columns []Column
}

type Column struct {
	Header string
	Width  float64
}

func ApplyLayout(layout Layout, excelFile *excelize.File, sheetName string) {
	for i, col := range layout.Columns {
		colId := getColumnIdentifier(i)
		excelFile.SetColWidth(sheetName, colId, colId, col.Width)
	}
}

type layoutBuilder struct {
	layout *Layout
}

func NewLayout() *layoutBuilder {
	return &layoutBuilder{
		layout: &Layout{},
	}
}

func (b *layoutBuilder) WithColumns(headers ...string) *layoutBuilder {
	cols := make([]Column, 0, len(headers))
	for _, header := range headers {
		cols = append(cols, Column{
			Header: header,
		})
	}
	b.layout.Columns = append(b.layout.Columns, cols...)
	return b
}

func (b *layoutBuilder) Build() Layout {
	return *b.layout
}
