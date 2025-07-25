package excel

import (
	"encoding/json"
	"fmt"
	"returntypes-langserver/common/debug/errors"

	"github.com/xuri/excelize/v2"
)

type Cursor struct {
	file    *excelize.File
	sheet   string
	x       int
	y       int
	err     errors.Error
	styleId int
}

func NewCursor(file *excelize.File, sheet string) *Cursor {
	return &Cursor{
		file:  file,
		sheet: sheet,
	}
}

// Moves the cursor to the given absolute position
func (c *Cursor) SetPosition(x, y int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	c.x, c.y = x, y
}

// Moves the cursor to the given relative position
func (c *Cursor) Move(x, y int) {
	c.SetPosition(c.x+x, c.y+y)
}

// Writes the values for one row. The cursor does not move during this process.
func (c *Cursor) WriteRowValues(values ...interface{}) errors.Error {
	if c.file == nil || c.err != nil {
		return c.err
	}
	for i, val := range values {
		if err := c.setCellValue(i, 0, val); err != nil {
			return err
		}
	}
	if err := c.applyStyle(c.x, c.y, len(values)-1, 0); err != nil {
		return err
	}
	return nil
}

// Writes the passed values to the excel file from the current position. The cursor moves to the next row under the written rows
func (c *Cursor) WriteValues(values [][]interface{}) errors.Error {
	if c.file == nil || c.err != nil {
		return c.err
	}
	for _, row := range values {
		for x, val := range row {
			if err := c.setCellValue(x, 0, val); err != nil {
				return err
			}
		}
		if err := c.applyStyle(c.x, c.y, len(row)-1, 0); err != nil {
			return err
		}
		c.Move(0, 1)
	}
	return nil
}

func (c *Cursor) setCellValue(x, y int, value interface{}) errors.Error {
	targetCell := GetCellIdentifier(c.x+x, c.y+y)
	if str, ok := value.(Markdown); ok {
		if err := c.file.SetCellRichText(c.sheet, targetCell, MarkdownToRichText(str)); err != nil {
			c.err = errors.Wrap(err, "Excel", "Could not write cell at position %s (%d, %d)", targetCell, c.x+x, c.y+y)
			return c.err
		}
	} else if chart, ok := value.(Chart); ok {
		if err := c.WriteChart(chart); err != nil {
			return err
		}
	} else if err := c.file.SetCellValue(c.sheet, targetCell, value); err != nil {
		c.err = errors.Wrap(err, "Excel", "Could not write cell at position %s (%d, %d)", targetCell, c.x+x, c.y+y)
		return c.err
	}
	return nil
}

func (c *Cursor) applyStyle(sx, sy, wdt, hgt int) errors.Error {
	if c.styleId > 0 && wdt >= 0 && hgt >= 0 {
		if err := c.file.SetCellStyle(c.sheet, GetCellIdentifier(sx, sy), GetCellIdentifier(sx+wdt, sy+hgt), c.styleId); err != nil {
			c.err = errors.Wrap(err, "Excel", "Could not apply style")
			return c.err
		}
	}
	return nil
}

// Applys the current style for the are from the current position to (current.x + toX, current.y + toY)
func (c *Cursor) ApplyStyle(toX, toY int) errors.Error {
	if c.styleId == 0 {
		return errors.New("Excel", "No style set")
	} else if toX < 0 || toY < 0 {
		return errors.New("Excel", "Negative coordinates not supported: %d, %d", toX, toY)
	}

	return c.applyStyle(c.x, c.y, toX, toY)
}

// Sets a style to use during writing
func (c *Cursor) SetStyle(styleId int) {
	c.styleId = styleId
}

// Returns the zero-based x position
func (c *Cursor) X() int {
	return c.x
}

// Returns the zero-based y position
func (c *Cursor) Y() int {
	return c.y
}

// Returns the identifier for the current column
func (c *Cursor) Col() string {
	return GetColumnIdentifier(c.x)
}

// Returns the current row number
func (c *Cursor) Row() string {
	return fmt.Sprintf("%d", c.y+1)
}

// Returns the current cell
func (c *Cursor) Cell() string {
	return GetCellIdentifier(c.x, c.y)
}

// Errors which occurred during the cursor operatios
func (c *Cursor) Error() errors.Error {
	return c.err
}

const DefaultChartHeight = 290

// Writes a chart at the current position. The cursor moves under the chart (vertically-only) during this process.
func (c *Cursor) WriteChart(chart Chart) errors.Error {
	if c.file == nil || c.err != nil {
		return c.err
	}

	if chartRaw, err := c.createRawChart(chart); err != nil {
		return err
	} else if err := c.addChart(chartRaw); err != nil {
		return err
	}
	for height := 0.0; height < DefaultChartHeight; c.Move(0, 1) {
		if hgt, err := c.file.GetRowHeight(c.sheet, c.y+1); err != nil {
			return errors.Wrap(err, "Excel", "Could not calculate chart height")
		} else {
			height += hgt
		}
	}
	return nil
}

func (c *Cursor) addChart(chart ChartRaw) errors.Error {
	chartJson, err := json.Marshal(chart)
	if err != nil {
		c.err = errors.Wrap(err, "Excel", "Could not create chart")
		return c.err
	}
	if err := c.file.AddChart(c.sheet, c.Cell(), string(chartJson)); err != nil {
		c.err = errors.Wrap(err, "Excel", "Could not create chart")
		return c.err
	}
	return nil
}

func (c *Cursor) createRawChart(chart Chart) (ChartRaw, errors.Error) {
	chartRaw := ChartRaw{
		ChartBase: chart.ChartBase,
		Series:    make([]SeriesRaw, len(chart.Series)),
	}
	for i, series := range chart.Series {
		seriesRaw, err := c.writeSeries(series)
		if err != nil {
			return chartRaw, err
		}
		chartRaw.Series[i] = seriesRaw
	}
	return chartRaw, nil
}

func (c *Cursor) writeSeries(series Series) (SeriesRaw, errors.Error) {
	raw := SeriesRaw{}
	values := make([][]interface{}, 0, 3)
	if series.Name != "" {
		raw.Name = c.dataRange(c.x, c.y, 0, 0)
		values = append(values, []interface{}{series.Name})
	}
	if len(series.Categories) > 0 {
		raw.Categories = c.dataRange(c.x, c.y+len(values), len(series.Categories)-1, 0)
		values = append(values, series.Categories)
	}
	if len(series.Values) > 0 {
		raw.Values = c.dataRange(c.x, c.y+len(values), len(series.Values)-1, 0)
		values = append(values, series.Values)
	}
	if err := c.WriteValues(values); err != nil {
		return raw, err
	}
	return raw, nil
}

func (c *Cursor) dataRange(fromX, fromY, toX, toY int) string {
	return fmt.Sprintf("'%s'!$%s$%d:$%s$%d", c.sheet, GetColumnIdentifier(fromX), fromY+1, GetColumnIdentifier(fromX+toX), fromY+toY+1)
}
