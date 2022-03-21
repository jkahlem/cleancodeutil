package workspace

import (
	"returntypes-langserver/languageserver/lsp"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ExampleCode = `
package com.example;

public class Example {
	public String name;
	
	public String getName() {
		return name;
	}
	
	public void setName(String name) {
		this.name = name;
	}
}`

func TestDocumentTextInsertion(t *testing.T) {
	// given
	doc := NewDocument(ExampleCode)
	methodToInsert := `

	public void printName() {
		System.out.println(this.name);
	}`
	insertPosition := lsp.Position{
		Line:      12,
		Character: 2,
	}

	// when
	doc.ApplyChange(lsp.TextDocumentContentChangeEvent{
		Text:  methodToInsert,
		Range: &lsp.Range{Start: insertPosition, End: insertPosition},
	})

	// then
	expected := `
package com.example;

public class Example {
	public String name;
	
	public String getName() {
		return name;
	}
	
	public void setName(String name) {
		this.name = name;
	}

	public void printName() {
		System.out.println(this.name);
	}
}`
	assert.Equal(t, expected, doc.Text())
}

func TestDocumentTextRemoval(t *testing.T) {
	// given
	doc := NewDocument(ExampleCode)
	removalRange := &lsp.Range{ // removal of "setName" method
		Start: lsp.Position{ // After last curly brace of "getName" method
			Line:      8,
			Character: 2,
		},
		End: lsp.Position{ // After last curly brace of "setName" method
			Line:      12,
			Character: 2,
		},
	}

	// when
	doc.ApplyChange(lsp.TextDocumentContentChangeEvent{
		Text:  "",
		Range: removalRange,
	})

	// then
	expected := `
package com.example;

public class Example {
	public String name;
	
	public String getName() {
		return name;
	}
}`
	assert.Equal(t, expected, doc.Text())
}

func TestDocumentTextReplacement(t *testing.T) {
	// given
	doc := NewDocument(ExampleCode)
	removalRange := &lsp.Range{ // removal of "setName" method
		Start: lsp.Position{ // After last curly brace of "getName" method
			Line:      8,
			Character: 2,
		},
		End: lsp.Position{ // After last curly brace of "setName" method
			Line:      12,
			Character: 2,
		},
	}
	methodToInsert := `

	public void printName() {
		System.out.println(this.name);
	}`

	// when
	doc.ApplyChange(lsp.TextDocumentContentChangeEvent{
		Text:  methodToInsert,
		Range: removalRange,
	})

	// then
	expected := `
package com.example;

public class Example {
	public String name;
	
	public String getName() {
		return name;
	}

	public void printName() {
		System.out.println(this.name);
	}
}`
	assert.Equal(t, expected, doc.Text())
}

func TestDocumentMultipleTextInsertions(t *testing.T) {
	// given
	doc := NewDocument(ExampleCode)

	// First event: add field declaration
	fieldToInsert := `
	public String description;`
	fieldInsertEvent := Insert(fieldToInsert, lsp.Position{
		Line:      4,
		Character: 24,
	})

	// Second event: add getter for field
	methodToInsert := `

	public String getDescription() {
		return description;
	}`
	methodInsertEvent := Insert(methodToInsert, lsp.Position{
		Line:      13, // Because one line is inserted before, need to add one to the position to get after "setName"
		Character: 2,
	})

	// when
	doc.ApplyChanges([]lsp.TextDocumentContentChangeEvent{
		fieldInsertEvent,
		methodInsertEvent,
	})

	// then
	expected := `
package com.example;

public class Example {
	public String name;
	
	public String getName() {
		return name;
	}
	
	public void setName(String name) {
		this.name = name;
	}

	public void printName() {
		System.out.println(this.name);
	}
}`
	assert.Equal(t, expected, doc.Text())
}

// Helpers
func Insert(text string, position lsp.Position) lsp.TextDocumentContentChangeEvent {
	return lsp.TextDocumentContentChangeEvent{
		Text: text,
		Range: &lsp.Range{
			Start: position,
			End:   position,
		},
	}
}
