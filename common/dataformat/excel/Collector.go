package excel

import (
	"os"
	"path/filepath"
	"returntypes-langserver/common/debug/errors"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Collector interface {
	ApplyLayout(Layout) errors.Error
	Write([]string, *Style) errors.Error
	Close() errors.Error
}

type file struct {
	excelFile    *excelize.File
	closed       bool
	index        int
	layout       Layout
	streamWriter *excelize.StreamWriter
	sheet        string
}

func newFileCollector(outputPath string) Collector {
	f := excelize.NewFile()
	f.Path = outputPath
	f.SetActiveSheet(f.NewSheet(DefaultSheetName))
	return newFileCollectorByFileAndSheet(f, DefaultSheetName)
}

func newFileCollectorByExcelFile(excelFile *excelize.File) Collector {
	return &file{
		excelFile: excelFile,
	}
}

func newFileCollectorByFileAndSheet(excelFile *excelize.File, sheet string) Collector {
	if excelFile.GetSheetIndex(sheet) == -1 {
		excelFile.NewSheet(sheet)
	}
	return &file{
		excelFile: excelFile,
		sheet:     sheet,
	}
}

func (w *file) ApplyLayout(layout Layout) errors.Error {
	if w.excelFile == nil {
		return w.cancelWithError(errors.New("Excel Error", "Cannot apply layout: Output excel file does not exist"))
	} else if w.closed {
		return errors.New("Excel Error", "Cannot apply layout: File is already closed")
	} else if err := w.checkStreamWriter(); err != nil {
		return err
	}
	for i, col := range layout.Columns {
		if err := w.applyColumnWidth(col, i); err != nil {
			return errors.Wrap(err, "Excel Error", "Cannot apply layout")
		}
	}
	w.layout = layout
	return nil
}

func (w *file) applyColumnWidth(col Column, zeroIndexedCol int) error {
	if col.Width > 0 {
		if err := w.streamWriter.SetColWidth(zeroIndexedCol+1, zeroIndexedCol+1, col.Width); err != nil {
			return err
		}
	}
	/*if col.Hide {
		if err := w.excelFile.SetColVisible(w.sheet, colId, false); err != nil {
			return err
		}
	}*/
	return nil
}

func (w *file) Write(record []string, style *Style) errors.Error {
	if w.excelFile == nil {
		return w.cancelWithError(errors.New("Excel Error", "Cannot write row: Output excel file does not exist"))
	} else if w.closed {
		return errors.New("Excel Error", "Cannot write row: File is already closed")
	} else if style == nil {
		return errors.New("Excel Error", "Cannot write row: style is nil.")
	} else if err := w.checkStreamWriter(); err != nil {
		return err
	} else if styleId, err := style.ToExcelStyle(w.excelFile); err != nil {
		return err
	} else {
		w.addRowToExcelFile(w.index, styleId, record...)
		w.index++
		return nil
	}
}

func (w *file) checkStreamWriter() errors.Error {
	if w.streamWriter == nil {
		if sw, err := w.excelFile.NewStreamWriter(w.sheet); err != nil {
			return errors.Wrap(err, "Excel Error", "Could not open file stream")
		} else {
			w.streamWriter = sw
		}
	}
	return nil
}

func (w *file) Close() errors.Error {
	if w.closed || w.excelFile == nil {
		return nil
	} else {
		if err := w.streamWriter.Flush(); err != nil {
			return errors.Wrap(err, "Excel Error", "Could not flush stream")
		}
		w.closed = true
		if err := os.MkdirAll(filepath.Dir(w.excelFile.Path), 0777); err != nil {
			return errors.Wrap(err, "Excel Error", "Could not create directories")
		} else if err := w.excelFile.Save(); err != nil {
			return errors.Wrap(err, "Excel Error", "Could not save excel file")
		}
		return nil
	}
}

func (w *file) cancelWithError(err errors.Error) errors.Error {
	w.closed = true
	if w.excelFile != nil {
		w.excelFile.Close()
	}
	return err
}

// Adds the given row to the excel file. rowIndex should be the zero-based index of the row.
func (w *file) addRowToExcelFile(rowIndex, styleId int, values ...string) errors.Error {
	cells := make([]interface{}, 0, len(values))
	for colIndex, value := range values {
		/*if len(value) > 0 {
			cell := getCellIdentifier(colIndex, rowIndex)
			if w.layout.Columns[colIndex].Markdown {
				if err := w.excelFile.SetCellRichText(w.sheet, cell, w.parseMarkdown(value)); err != nil {
					return errors.Wrap(err, "Excel Error", "Could not add row to excel file for %s (value: %v)", cell, value)
				}
			} else if err := w.excelFile.SetCellValue(w.sheet, cell, value); err != nil {
				return errors.Wrap(err, "Excel Error", "Could not add row to excel file for %s (value: %v)", cell, value)
			}
		}*/
		cellStyle := styleId
		if len(w.layout.Columns) > colIndex && w.layout.Columns[colIndex].Hide {
			// Because the stream writer implementation from the excelize package does not support setting column visibility,
			// we just set a different style where font colour = background colour, so it does not look like there is something, as it distracts...
			if id, err := w.layout.HiddenStyle.ToExcelStyle(w.excelFile); err == nil {
				cellStyle = id
			}
		}
		cells = append(cells, excelize.Cell{
			Value:   value,
			StyleID: cellStyle,
		})
	}
	if err := w.streamWriter.SetRow(getCellIdentifier(0, rowIndex), cells); err != nil {
		return errors.Wrap(err, "Excel Error", "Could not write row %d", rowIndex)
	}
	return nil
	//return w.applyRowStyle(rowIndex, len(values), styleId)
}

// Very naive markdown parser that only parses for bold text
func (w *file) parseMarkdown(value string) []excelize.RichTextRun {
	isBold := false
	richText := make([]excelize.RichTextRun, 0)
	for _, part := range strings.Split(value, "**") {
		richTextPart := excelize.RichTextRun{Text: part}
		if isBold {
			richTextPart.Font = &excelize.Font{
				Bold: true,
			}
		}
		richText = append(richText, richTextPart)
		isBold = !isBold
	}
	return richText
}

func (w *file) applyRowStyle(rowIndex, valuesLength, styleId int) errors.Error {
	if err := w.excelFile.SetRowStyle(w.sheet, rowIndex+1, rowIndex+1, styleId); err != nil {
		return errors.Wrap(err, "Excel Error", "Cannot apply row style")
	}
	// The row style does somehow not apply to cells which were set (even if called before setting cell values)
	// therefore we need to set the cell's style seperately..
	startCell, endCell := getCellIdentifier(0, rowIndex), getCellIdentifier(valuesLength, rowIndex)
	if err := w.excelFile.SetCellStyle(w.sheet, startCell, endCell, styleId); err != nil {
		return errors.Wrap(err, "Excel Error", "Cannot apply row style")
	}
	return nil
}

type sliceCollector struct {
	slice *[][]string
}

func newSliceCollector(slice *[][]string) Collector {
	return &sliceCollector{
		slice: slice,
	}
}

func (c *sliceCollector) Write(record []string, style *Style) errors.Error {
	if c.slice == nil {
		return errors.New("Excel Error", "Could not write to nil pointer slice.")
	}
	*c.slice = append(*c.slice, record)
	return nil
}

func (c *sliceCollector) Close() errors.Error {
	return nil
}

func (c *sliceCollector) ApplyLayout(layout Layout) errors.Error {
	return nil
}
