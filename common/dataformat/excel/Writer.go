package excel

import (
	"reflect"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"strconv"
	"strings"
)

type StreamWriter interface {
	Write([]string) errors.Error
	BuildLayout(Layout) errors.Error
	SetWriter(StreamWriter)
}

// Writes headers considering the struct layout
type structFormatWriter struct {
	format interface{}
	writer StreamWriter
}

func newStructFormatWriter(structType interface{}) StreamWriter {
	return &structFormatWriter{
		format: structType,
	}
}

func (w *structFormatWriter) Write(record []string) errors.Error {
	if w.writer == nil {
		return nil
	} else {
		return w.writer.Write(record)
	}
}

func (w *structFormatWriter) BuildLayout(layout Layout) errors.Error {
	if w.writer == nil {
		return nil
	} else if l, err := w.buildLayoutByStruct(layout, w.format); err != nil {
		return err
	} else {
		return w.writer.BuildLayout(l)
	}
}

func (w *structFormatWriter) buildLayoutByStruct(layout Layout, structType interface{}) (Layout, errors.Error) {
	reflected := utils.UnwrapType(reflect.TypeOf(structType))
	if reflected == nil {
		return layout, errors.New("Excel Error", "Could not identify struct type")
	} else if reflected.Kind() != reflect.Struct {
		return layout, errors.New("Excel Error", "Expected a struct type passed.")
	}

	header := make([]Column, 0, reflected.NumField())
	for i := 0; i < reflected.NumField(); i++ {
		tag := reflected.Field(i).Tag.Get(ExcelHeaderTag)
		header = append(header, w.buildColumn(tag))
	}
	layout.Columns = header
	return layout, nil
}

const (
	ColumnWidthAttr    = "width"
	ColumnHideAttr     = "hide"
	ColumnMarkdownAttr = "markdown"
)

func (w *structFormatWriter) buildColumn(tag string) Column {
	splitted := strings.Split(tag, ",")
	col := Column{
		Header: splitted[0],
	}
	for _, attribute := range splitted[1:] {
		if key, value, ok := utils.KeyValueByEqualSign(attribute); ok {
			switch key {
			case ColumnWidthAttr:
				if parsed, err := strconv.ParseFloat(value, 64); err == nil {
					col.Width = parsed
				}
			case ColumnHideAttr:
				if value == "true" {
					col.Hide = true
				}
			case ColumnMarkdownAttr:
				if value == "true" {
					col.Markdown = true
				}
			}
		}
	}
	return col
}

func (w *structFormatWriter) SetWriter(writer StreamWriter) {
	w.writer = writer
}

// Writes headers by the defined order
type staticFormatWriter struct {
	header []string
	writer StreamWriter
}

func newStaticFormatWriter(header []string) StreamWriter {
	return &staticFormatWriter{
		header: header,
	}
}

func (w *staticFormatWriter) Write(record []string) errors.Error {
	if w.writer == nil {
		return nil
	} else {
		return w.writer.Write(record)
	}
}

func (w *staticFormatWriter) BuildLayout(layout Layout) errors.Error {
	if w.writer == nil {
		return nil
	}
	layout.Columns = make([]Column, 0, len(w.header))
	for _, header := range w.header {
		layout.Columns = append(layout.Columns, Column{Header: header})
	}
	return w.writer.BuildLayout(layout)
}

func (w *staticFormatWriter) SetWriter(writer StreamWriter) {
	w.writer = writer
}

// Swaps the defined columns
type columnSwapper struct {
	i, j   Col
	writer StreamWriter
}

func newColumnSwapper(i, j Col) StreamWriter {
	return &columnSwapper{
		i: i,
		j: j,
	}
}

func (w *columnSwapper) Write(record []string) errors.Error {
	if w.writer == nil {
		return nil
	}
	t := record[w.i]
	record[w.i] = record[w.j]
	record[w.j] = t
	return w.writer.Write(record)
}

func (w *columnSwapper) BuildLayout(layout Layout) errors.Error {
	if w.writer == nil {
		return nil
	}
	t := layout.Columns[w.i]
	layout.Columns[w.i] = layout.Columns[w.j]
	layout.Columns[w.j] = t
	return w.writer.BuildLayout(layout)
}

func (w *columnSwapper) SetWriter(writer StreamWriter) {
	w.writer = writer
}

// Inserts columns at the given position
type columnInserter struct {
	columnsToInsert  []string
	emptySpace       []string
	positionToInsert Col
	writer           StreamWriter
}

func newColumnInserter(positionToInsert Col, columns ...string) StreamWriter {
	return &columnInserter{
		columnsToInsert:  columns,
		emptySpace:       make([]string, len(columns)),
		positionToInsert: positionToInsert,
	}
}

func (w *columnInserter) Write(record []string) errors.Error {
	if w.writer == nil {
		return nil
	}
	inserted := make([]string, w.positionToInsert, len(record)+len(w.emptySpace))
	copy(inserted, record[:w.positionToInsert])
	inserted = append(inserted, w.emptySpace...)
	inserted = append(inserted, record[w.positionToInsert:]...)
	return w.writer.Write(inserted)
}

func (w *columnInserter) BuildLayout(layout Layout) errors.Error {
	pos := int(w.positionToInsert)
	if w.writer == nil {
		return nil
	} else if pos < 0 {
		return errors.New("Excel Error", "Column insertion position: Expected value greater than or equal to 0 but got %d", w.positionToInsert)
	} else if pos > len(layout.Columns) {
		for i := len(layout.Columns); i < pos; i++ {
			layout.Columns = append(layout.Columns, Column{Header: ""})
		}
	}
	columns := make([]Column, w.positionToInsert, len(layout.Columns)+len(w.columnsToInsert))
	copy(columns, layout.Columns[:w.positionToInsert])
	for _, header := range w.columnsToInsert {
		columns = append(columns, Column{Header: header})
	}
	if len(layout.Columns[w.positionToInsert:]) > 0 {
		columns = append(columns, layout.Columns[w.positionToInsert:]...)
	}
	layout.Columns = columns
	return w.writer.BuildLayout(layout)
}

func (w *columnInserter) SetWriter(writer StreamWriter) {
	w.writer = writer
}

type RecordTransformer func([]string) []string

// Transforms the data of a record before passing it further through the stream
type transformerWriter struct {
	transformer RecordTransformer
	writer      StreamWriter
}

func newTransformer(transformer RecordTransformer) StreamWriter {
	return &transformerWriter{
		transformer: transformer,
	}
}

func (w *transformerWriter) Write(record []string) errors.Error {
	if w.writer == nil {
		return nil
	} else if w.transformer == nil {
		return w.writer.Write(record)
	}
	return w.writer.Write(w.transformer(record))
}

func (w *transformerWriter) BuildLayout(layout Layout) errors.Error {
	if w.writer == nil {
		return nil
	}
	return w.writer.BuildLayout(layout)
}

func (w *transformerWriter) SetWriter(writer StreamWriter) {
	w.writer = writer
}

type RecordComparer func(a, b []string) bool
type Merger func(new, old []string) []string

// Merges two records lists together
type merger struct {
	oldData   [][]string
	comparer  RecordComparer
	writer    StreamWriter
	mergeFunc Merger
}

func newMerger(oldData [][]string, comparer RecordComparer, mergeFunc Merger) StreamWriter {
	m := &merger{
		comparer:  comparer,
		oldData:   oldData,
		mergeFunc: mergeFunc,
	}
	return m
}

func (w *merger) Write(record []string) errors.Error {
	if w.comparer == nil {
		return errors.New("Excel Error", "No comparer function defined for merging process")
	} else if w.writer == nil {
		return nil
	}
	w.trimOldData()
	oldRecord := w.findRecordInOldData(record)
	if oldRecord != nil {
		return w.writer.Write(w.mergeFunc(record, oldRecord))
	}
	return w.writer.Write(record)
}

func (w *merger) findRecordInOldData(record []string) []string {
	for i, oldRecord := range w.oldData {
		if oldRecord != nil && w.comparer(record, oldRecord) {
			w.oldData[i] = nil
			return oldRecord
		}
	}
	return nil
}

func (w *merger) trimOldData() {
	for i, r := range w.oldData {
		if r != nil {
			w.oldData = w.oldData[i:]
			break
		}
	}
}

func (w *merger) BuildLayout(layout Layout) errors.Error {
	if w.writer == nil {
		return nil
	}
	return w.writer.BuildLayout(layout)
}

func (w *merger) SetWriter(writer StreamWriter) {
	w.writer = writer
}

var Nothing StreamWriter = &Chain{}

// Does nothing and is only used for chaining
type Chain struct {
	writer StreamWriter
}

func (w *Chain) SetWriter(writer StreamWriter) {
	w.writer = writer
}

func (w *Chain) Write(record []string) errors.Error {
	if w.writer == nil {
		return nil
	}
	return w.writer.Write(record)
}

func (w *Chain) BuildLayout(layout Layout) errors.Error {
	if w.writer == nil {
		return nil
	}
	return w.writer.BuildLayout(layout)
}
