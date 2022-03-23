package workspace

import (
	"returntypes-langserver/common/utils"
	"returntypes-langserver/languageserver/lsp"
	"strings"
)

type Document struct {
	content []string
}

func NewDocument(content string) Document {
	doc := Document{}
	doc.SetText(content)
	return doc
}

func (doc *Document) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	for _, change := range changes {
		doc.ApplyChange(change)
	}
}

func (doc *Document) ApplyChange(change lsp.TextDocumentContentChangeEvent) {
	if change.Text != "" && change.Range == nil {
		// Special case: The whole text was replaced
		doc.SetText(change.Text)
		return
	}
	if !change.Range.IsEmpty() {
		doc.remove(*change.Range)
	}
	if change.Text != "" {
		doc.insert(change.Range.Start, change.Text)
	}
}

func (doc *Document) remove(r lsp.Range) {
	if len(doc.content) == 0 {
		return
	}
	r = doc.boundRange(r)
	if r.Start.Line == r.End.Line {
		// no line removal if on the same line
		line := doc.content[r.Start.Line]
		if len(line) == 0 {
			return
		}
		doc.content[r.Start.Line] = line[:r.Start.Character] + line[r.End.Character:]
	} else {
		endline := doc.content[r.End.Line][r.End.Character:]
		doc.content[r.Start.Line] = doc.content[r.Start.Line][:r.Start.Character] + endline
		if len(doc.content) != r.End.Line+1 {
			copy(doc.content[r.Start.Line+1:], doc.content[r.End.Line+1:])
		}
		lineDifference := r.End.Line - r.Start.Line
		doc.content = doc.content[:len(doc.content)-lineDifference]
	}
}

func (doc *Document) insert(pos lsp.Position, text string) {
	if len(doc.content) == 0 || len(doc.content) == 1 && doc.content[0] == "" {
		doc.SetText(text)
		return
	}
	line := doc.content[pos.Line]
	pos = doc.boundPosition(pos)
	linesToInsert := strings.Split(text, "\n")
	if len(linesToInsert) == 1 {
		doc.content[pos.Line] = line[:pos.Character] + linesToInsert[0] + line[pos.Character:]
	} else {
		beforeCursor, afterCursor := line[:pos.Character], line[pos.Character:]
		insertLineCount := len(linesToInsert)
		linesToInsert[0] = beforeCursor + linesToInsert[0]
		linesToInsert[insertLineCount-1] = afterCursor + linesToInsert[insertLineCount-1]
		oldContent := doc.content
		newLength := len(oldContent) - 1 + len(linesToInsert)

		if newLength > cap(doc.content) {
			doc.content = make([]string, newLength, newLength+64)
			copy(doc.content, oldContent[:pos.Line])
		} else {
			doc.content = doc.content[:newLength]
		}
		if pos.Line+1 < len(oldContent) {
			copy(doc.content[pos.Line+len(linesToInsert):], oldContent[pos.Line+1:])
		}
		copy(doc.content[pos.Line:], linesToInsert)
	}
}

func (doc *Document) boundRange(r lsp.Range) lsp.Range {
	r.Start = doc.boundPosition(r.Start)
	r.End = doc.boundPosition(r.End)
	return r
}

func (doc *Document) boundPosition(position lsp.Position) lsp.Position {
	position.Line = utils.BoundIndex(position.Line, len(doc.content))
	if len(doc.content) != 0 {
		position.Character = utils.BoundInside(position.Character, 0, len(doc.content[position.Line]))
	}
	return position
}

func (doc *Document) SetText(text string) {
	doc.content = strings.Split(text, "\n")
}

func (doc *Document) Text() string {
	return strings.Join(doc.content, "\n")
}

// Converts a position (line:col) to the string offset on the full document string
func (doc *Document) ToOffset(pos lsp.Position) int {
	offset := 0
	for i, line := range doc.content {
		if i < pos.Line {
			offset += len(line) + 1 // add one character for the line break
		} else {
			offset += pos.Character
			break
		}
	}
	return offset
}

// Converts a string/byte offset of the code to a position (line:col)
func (doc *Document) ToPosition(offset int) lsp.Position {
	pos := lsp.Position{}
	for i, line := range doc.content {
		lineLength := len(line) + 1
		if lineLength < offset {
			offset -= lineLength
		} else {
			pos.Line = i
			pos.Character = offset
			break
		}
	}
	return pos
}
