package excel

import (
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

func (c *Cursor) SetPosition(x, y int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	c.x, c.y = x, y
}

func (c *Cursor) Move(x, y int) {
	c.SetPosition(c.x+x, c.y+y)
}

func (c *Cursor) WriteRowValues(values ...interface{}) errors.Error {
	if c.file == nil || c.err != nil {
		return c.err
	}
	for i, val := range values {
		if err := c.file.SetCellValue(c.sheet, getCellIdentifier(c.x+i, c.y), val); err != nil {
			c.err = errors.Wrap(err, "Excel", "Could not write cell at position %s (%d, %d)", getCellIdentifier(c.x+i, c.y), c.x+i, c.y)
			return c.err
		}
	}
	c.applyStyle(c.x, c.y, len(values)-1, 0)
	return c.err
}

func (c *Cursor) WriteStringValues(values [][]string) errors.Error {
	if c.file == nil || c.err != nil {
		return c.err
	}
	for y, row := range values {
		for x, val := range row {
			if err := c.file.SetCellValue(c.sheet, getCellIdentifier(c.x+x, c.y+y), val); err != nil {
				c.err = errors.Wrap(err, "Excel", "Could not write cell at position %s (%d, %d)", getCellIdentifier(c.x+x, c.y+y), c.x+x, c.y+y)
				return c.err
			}
		}
		c.applyStyle(c.x, c.y+y, len(row)-1, 0)
	}
	return c.err
}

func (c *Cursor) applyStyle(sx, sy, wdt, hgt int) {
	if c.styleId > 0 && wdt >= 0 && hgt >= 0 {
		c.file.SetCellStyle(c.sheet, getCellIdentifier(sx, sy), getCellIdentifier(sx+wdt, sy+hgt), c.styleId)
	}
}

func (c *Cursor) SetStyle(styleId int) {
	c.styleId = styleId
}

func (c *Cursor) Error() errors.Error {
	return c.err
}
