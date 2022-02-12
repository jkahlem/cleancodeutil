package excel

import (
	"fmt"
	"returntypes-langserver/common/debug/errors"

	"github.com/xuri/excelize/v2"
)

type Collector interface {
	ApplyLayout(Layout) errors.Error
	Write([]string, *Style) errors.Error
	Close() errors.Error
}

type file struct {
	excelFile *excelize.File
	closed    bool
	index     int
}

func newFileCollector(outputPath string) Collector {
	f := excelize.NewFile()
	f.Path = outputPath
	f.SetActiveSheet(f.NewSheet(DefaultSheetName))
	return &file{
		excelFile: f,
	}
}

func newFileCollectorByExcelFile(excelFile *excelize.File) Collector {
	return &file{
		excelFile: excelFile,
	}
}

func (w *file) ApplyLayout(layout Layout) errors.Error {
	if w.excelFile == nil {
		return w.cancelWithError(errors.New("Excel Error", "Cannot apply layout: Output excel file does not exist"))
	} else if w.closed {
		return errors.New("Excel Error", "Cannot apply layout: File is already closed")
	}
	for i, col := range layout.Columns {
		colId := getColumnIdentifier(i)
		if err := w.applyColumnLayout(col, colId); err != nil {
			return errors.Wrap(err, "Excel Error", "Cannot apply layout")
		}
	}
	return nil
}

func (w *file) applyColumnLayout(col Column, colId string) error {
	if col.Width > 0 {
		if err := w.excelFile.SetColWidth(DefaultSheetName, colId, colId, col.Width); err != nil {
			return err
		}
	}
	if col.Hide {
		if err := w.excelFile.SetColVisible(DefaultSheetName, colId, false); err != nil {
			return err
		}
	}
	return nil
}

func (w *file) Write(record []string, style *Style) errors.Error {
	if w.excelFile == nil {
		return w.cancelWithError(errors.New("Excel Error", "Cannot write row: Output excel file does not exist"))
	} else if w.closed {
		return errors.New("Excel Error", "Cannot write row: File is already closed")
	} else if style == nil {
		return errors.New("Excel Error", "Cannot write row: style is nil.")
	} else if styleId, err := style.ToExcelStyle(w.excelFile); err != nil {
		return err
	} else {
		w.addRowToExcelFile(w.index, styleId, record...)
		w.index++
		return nil
	}
}

func (w *file) Close() errors.Error {
	if w.closed || w.excelFile == nil {
		return nil
	} else {
		w.closed = true
		if err := w.excelFile.Save(); err != nil {
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
	for colIndex, value := range values {
		if len(value) > 0 {
			cell := getCellIdentifier(colIndex, rowIndex)
			if err := w.excelFile.SetCellValue(DefaultSheetName, cell, value); err != nil {
				return errors.Wrap(err, "Excel Error", fmt.Sprintf("Could not add row to excel file for %s (value: %v)", cell, value))
			}
		}
	}
	return w.applyRowStyle(rowIndex, len(values), styleId)
}

func (w *file) applyRowStyle(rowIndex, valuesLength, styleId int) errors.Error {
	if err := w.excelFile.SetRowStyle(DefaultSheetName, rowIndex+1, rowIndex+1, styleId); err != nil {
		return errors.Wrap(err, "Excel Error", "Cannot apply row style")
	}
	// The row style does somehow not apply to cells which were set (even if called before setting cell values)
	// therefore we need to set the cell's style seperately..
	startCell, endCell := getCellIdentifier(0, rowIndex), getCellIdentifier(valuesLength, rowIndex)
	if err := w.excelFile.SetCellStyle(DefaultSheetName, startCell, endCell, styleId); err != nil {
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
