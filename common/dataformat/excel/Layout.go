package excel

import (
	"returntypes-langserver/common/debug/errors"

	"github.com/xuri/excelize/v2"
)

type Layout struct {
	Columns []Column
	// Style for the header row
	HeaderStyle Style
	// Style on each even row
	EvenRowStyle Style
	// Style on each odd row. (This begins from the first row under the header)
	OddRowStyle Style
}

type Column struct {
	Header   string
	Width    float64
	Hide     bool
	Markdown bool
}

type Style struct {
	Bold            bool
	BackgroundColor string
	FontColor       string
	styleId         int
	file            *excelize.File
}

func (s *Style) ToExcelStyle(file *excelize.File) (int, errors.Error) {
	if file == nil {
		return -1, errors.New("Excel error", "Cannot create style for not-existent file")
	} else if s.file != nil && s.file.Path == file.Path {
		return s.styleId, nil
	}
	destStyle := excelize.Style{
		Font:   &excelize.Font{Color: "#000000", Size: 12},
		Border: []excelize.Border{s.border("top"), s.border("left"), s.border("bottom"), s.border("right")},
	}
	if s.Bold {
		destStyle.Font.Bold = true
	}
	if s.FontColor != "" {
		destStyle.Font.Color = s.FontColor
	}
	if s.BackgroundColor != "" {
		destStyle.Fill = excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{s.BackgroundColor},
		}
	}
	id, err := file.NewStyle(&destStyle)
	if err != nil {
		return -1, errors.Wrap(err, "Excel Error", "Could not create style for excel file")
	}
	s.styleId = id
	s.file = file
	return id, nil
}

func (s *Style) border(dir string) excelize.Border {
	return excelize.Border{
		Type:  dir,
		Color: "#ACACAC",
		Style: 1,
	}
}

func ApplyLayout(layout Layout, excelFile *excelize.File, sheetName string) {
	for i, col := range layout.Columns {
		colId := getColumnIdentifier(i)
		excelFile.SetColWidth(sheetName, colId, colId, col.Width)
	}
	excelize.NewFile().NewStyle(&excelize.Style{})
}

func DefaultLayout() Layout {
	return Layout{
		HeaderStyle: Style{
			Bold:            true,
			BackgroundColor: "#C6E0B4",
		},
		EvenRowStyle: Style{
			BackgroundColor: "#D8D8D8",
		},
		OddRowStyle: Style{
			BackgroundColor: "#FFFFFF",
		},
	}
}

func getHeaderStringsFromLayout(layout Layout) []string {
	header := make([]string, 0, len(layout.Columns))
	for _, col := range layout.Columns {
		header = append(header, col.Header)
	}
	return header
}
