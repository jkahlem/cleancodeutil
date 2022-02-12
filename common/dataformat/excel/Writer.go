package excel

import (
	"fmt"
	"reflect"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

type StreamWriter interface {
	Write([]string) errors.Error
	BuildLayout(Layout) errors.Error
	SetWriter(StreamWriter)
}

// Writes records considering the order of the struct layout
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
	} else if layout, err := w.buildLayoutByStruct(w.format); err != nil {
		return err
	} else {
		return w.writer.BuildLayout(layout)
	}
}

func (w *structFormatWriter) buildLayoutByStruct(structType interface{}) (Layout, errors.Error) {
	reflected := utils.UnwrapType(reflect.TypeOf(structType))
	if reflected == nil {
		return EmptyLayout(), errors.New("Excel Error", "Could not identify struct type")
	} else if reflected.Kind() != reflect.Struct {
		return EmptyLayout(), errors.New("Excel Error", "Expected a struct type passed.")
	}

	header := make([]string, 0, reflected.NumField())
	for i := 0; i < reflected.NumField(); i++ {
		header = append(header, reflected.Field(i).Tag.Get(ExcelHeaderTag))
	}
	return NewLayout().WithColumns(header...).Build(), nil
}

func (w *structFormatWriter) SetWriter(writer StreamWriter) {
	w.writer = writer
}

type columnSwapper struct {
	i, j   int
	writer StreamWriter
}

func newColumnSwapper(i, j int) StreamWriter {
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
	inserted := record[:w.positionToInsert]
	inserted = append(inserted, w.emptySpace...)
	inserted = append(inserted, record[w.positionToInsert:]...)
	return w.writer.Write(inserted)
}

func (w *columnInserter) BuildLayout(layout Layout) errors.Error {
	pos := int(w.positionToInsert)
	if w.writer == nil {
		return nil
	} else if !(pos >= 0 && pos <= len(layout.Columns)) {
		return errors.New("Excel Error", fmt.Sprintf("Column insertion position exceeds bounds. (Is %d, needs to be between 0 and %d)", w.positionToInsert, len(layout.Columns)))
	}
	columns := layout.Columns[:w.positionToInsert]
	for _, header := range w.columnsToInsert {
		columns = append(columns, Column{Header: header})
	}
	columns = append(columns, layout.Columns[w.positionToInsert:]...)
	return w.writer.BuildLayout(layout)
}

func (w *columnInserter) SetWriter(writer StreamWriter) {
	w.writer = writer
}

type RecordTransformer func([]string) []string

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
