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

func TestDocumentTextInsertionInline(t *testing.T) {
	// given
	doc := NewDocument(ExampleCode)
	insertion := Insert(`final `, lsp.Position{ //	public *String name;
		Line:      4,
		Character: 8,
	})

	// when
	doc.ApplyChange(insertion)

	// then
	expected := `
package com.example;

public class Example {
	public final String name;

	public String getName() {
		return name;
	}

	public void setName(String name) {
		this.name = name;
	}
}`
	assert.Equal(t, expected, doc.Text())
}

func TestTextInsertionIntoEmptyDocument(t *testing.T) {
	// given
	doc := NewDocument("")
	insertion := Insert(ExampleCode, lsp.Position{
		Line:      0,
		Character: 0,
	})

	// when
	doc.ApplyChange(insertion)

	// then
	assert.Equal(t, ExampleCode, doc.Text())
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

func TestDocumentTextRemovalInline(t *testing.T) {
	// given
	doc := NewDocument(ExampleCode)
	removalRange := &lsp.Range{ // removal of "Name" in "getName" method
		Start: lsp.Position{ // public get*Name() {
			Line:      6,
			Character: 18,
		},
		End: lsp.Position{ // public getName*() {
			Line:      6,
			Character: 22,
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

	public String get() {
		return name;
	}

	public void setName(String name) {
		this.name = name;
	}
}`
	assert.Equal(t, expected, doc.Text())
}

func TestDocumentTextFullRemoval(t *testing.T) {
	// given
	doc := NewDocument(ExampleCode)
	removalRange := &lsp.Range{
		Start: lsp.Position{
			Line:      0,
			Character: 0,
		},
		End: lsp.Position{
			Line:      100, // bounds to last line/character as limits are exceeded
			Character: 100,
		},
	}

	// when
	doc.ApplyChange(lsp.TextDocumentContentChangeEvent{
		Text:  "",
		Range: removalRange,
	})

	// then
	assert.Equal(t, "", doc.Text())
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
	public String description;

	public String getName() {
		return name;
	}

	public void setName(String name) {
		this.name = name;
	}

	public String getDescription() {
		return description;
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
