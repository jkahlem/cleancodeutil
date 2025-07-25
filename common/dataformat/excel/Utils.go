package excel

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

const DefaultSheetName = "Sheet"
const ExcelHeaderTag = "excel"

// Returns the identifier for an excel cell by zero-based index.
//   GetCellIdentifier(0, 0) // "A1"
//   GetCellIdentifier(26, 100) // "AA101"
func GetCellIdentifier(colIndex, rowIndex int) string {
	return fmt.Sprintf("%s%d", GetColumnIdentifier(colIndex), rowIndex+1)
}

// Returns the identifier for an excel column with the given (zero-based) index, e.g. 0 -> "A", 1 -> "B", ..., 25 -> "Z", 26 -> "AA", 27 -> "AB" etc.
func GetColumnIdentifier(index int) string {
	chr := string(rune((index % 26) + int('A')))
	if index >= 26 {
		return GetColumnIdentifier((index-(index%26))/26-1) + chr
	}
	return chr
}

type Markdown string

func MarkdownToRichText(value Markdown) []excelize.RichTextRun {
	isBold := false
	richText := make([]excelize.RichTextRun, 0)
	for _, part := range strings.Split(string(value), "**") {
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
