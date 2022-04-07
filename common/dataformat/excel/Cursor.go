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
		if err := c.setCellValue(i, 0, val); err != nil {
			return err
		}
	}
	if err := c.applyStyle(c.x, c.y, len(values)-1, 0); err != nil {
		return err
	}
	return nil
}

func (c *Cursor) WriteStringValues(values [][]string) errors.Error {
	if c.file == nil || c.err != nil {
		return c.err
	}
	for y, row := range values {
		for x, val := range row {
			if err := c.setCellValue(x, y, val); err != nil {
				return err
			}
		}
		if err := c.applyStyle(c.x, c.y+y, len(row)-1, 0); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cursor) setCellValue(x, y int, value interface{}) errors.Error {
	targetCell := getCellIdentifier(c.x+x, c.y+y)
	if str, ok := value.(string); ok {
		if err := c.file.SetCellRichText(c.sheet, targetCell, MarkdownToRichText(str)); err != nil {
			c.err = errors.Wrap(err, "Excel", "Could not write cell at position %s (%d, %d)", targetCell, c.x+x, c.y+y)
			return c.err
		}
	} else if err := c.file.SetCellValue(c.sheet, targetCell, value); err != nil {
		c.err = errors.Wrap(err, "Excel", "Could not write cell at position %s (%d, %d)", targetCell, c.x+x, c.y+y)
		return c.err
	}
	return nil
}

func (c *Cursor) applyStyle(sx, sy, wdt, hgt int) errors.Error {
	if c.styleId > 0 && wdt >= 0 && hgt >= 0 {
		if err := c.file.SetCellStyle(c.sheet, getCellIdentifier(sx, sy), getCellIdentifier(sx+wdt, sy+hgt), c.styleId); err != nil {
			c.err = errors.Wrap(err, "Excel", "Could not apply style")
			return c.err
		}
	}
	return nil
}

func (c *Cursor) ApplyStyle(width, height int) errors.Error {
	if c.styleId == 0 {
		return errors.New("Excel", "No style set")
	} else if width < 0 || height < 0 {
		return errors.New("Excel", "Got negative width/height: %d, %d", width, height)
	}

	return c.applyStyle(c.x, c.y, width, height)
}

func (c *Cursor) SetStyle(styleId int) {
	c.styleId = styleId
}

func (c *Cursor) X() int {
	return c.x
}

func (c *Cursor) Y() int {
	return c.y
}

func (c *Cursor) Error() errors.Error {
	return c.err
}
