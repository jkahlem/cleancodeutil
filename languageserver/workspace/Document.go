package workspace

import (
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
	// There are three possible text events:
	// - Text removal (Backspace, Cut etc.)
	// - Text insertion (Pressing keys, Pasting etc.)
	// - Text replacement (Select and press keys/paste, Find and Replace etc.)
	//
	// The text removal is always removing a given selection (eventRange) while the text insertion is inserting event.Text to eventRange.Start.
	// Text replacements can be seen as a combination of these two operations.
	if !change.Range.IsEmpty() {
		// TODO: remove contents between range.start to range.end
	}
	if change.Text != "" {
		// TODO: insert text at position from range.start
	}
	// Nothing else to do
}

func (doc *Document) SetText(text string) {
	doc.content = strings.Split(text, "\n")
}

func (doc *Document) Text() string {
	return strings.Join(doc.content, "\n")
}
