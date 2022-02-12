package excel

import "github.com/xuri/excelize/v2"

type Layout struct {
	Columns []Column
	style   Style
}

type Column struct {
	Header string
	Width  float64
}

type Style struct {
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

func EmptyLayout() Layout {
	return NewLayout().Build()
}

func getHeaderStringsFromLayout(layout Layout) []string {
	header := make([]string, 0, len(layout.Columns))
	for _, col := range layout.Columns {
		header = append(header, col.Header)
	}
	return header
}
