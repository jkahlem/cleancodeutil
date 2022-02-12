package excel

import (
	"fmt"
	"returntypes-langserver/common/debug/errors"

	"github.com/xuri/excelize/v2"
)

type Collector interface {
	Write([]string, Style) errors.Error
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

func (w *file) Write(record []string, style Style) errors.Error {
	if w.excelFile == nil {
		return w.cancelWithError(errors.New("Excel Error", "Output excel file does not exist"))
	} else if w.closed {
		return errors.New("Excel Error", "File is already closed")
	}

	w.addRowToExcelFile(w.index, record...)
	w.index++
	return nil
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
func (c *file) addRowToExcelFile(rowIndex int, values ...string) errors.Error {
	// excel index starts at 1 so elevate the zero-based index
	rowIndex++
	for index, value := range values {
		if len(value) > 0 {
			if err := c.excelFile.SetCellValue(DefaultSheetName, fmt.Sprintf("%s%d", getColumnIdentifier(index), rowIndex), value); err != nil {
				return errors.Wrap(err, "Excel Error", fmt.Sprintf("Could not add row to excel file for %s%d (value: %v)", getColumnIdentifier(index), rowIndex, value))
			}
		}
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

func (c *sliceCollector) Write(record []string, style Style) errors.Error {
	if c.slice == nil {
		return errors.New("Excel Error", "Could not write to nil pointer slice.")
	}
	*c.slice = append(*c.slice, record)
	return nil
}

func (c *sliceCollector) Close() errors.Error {
	return nil
}
