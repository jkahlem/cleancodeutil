package languageserver

import (
	"returntypes-langserver/languageserver/lsp"
	"returntypes-langserver/languageserver/workspace"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindMethodAtCursor(t *testing.T) {
	// given
	c := Controller{}
	doc := workspace.NewDocument(`package com.example;

public class Example {
	public String name;

	public String getName() {
		return name;
	}

	public void newMethod()

	public void setName(String name) {
		this.name = name;
	}
}`)
	positionInsideBraces := lsp.Position{ // public void newMethod(*)
		Line:      9,
		Character: 23,
	}
	positionAfterBraces := lsp.Position{ // public void newMethod()*
		Line:      9,
		Character: 24,
	}

	// when
	methodWhenInside, foundWhenInside := c.findMethodAtCursorPosition(&doc, positionInsideBraces)
	_, foundWhenAfter := c.findMethodAtCursorPosition(&doc, positionAfterBraces)

	// then
	assert.True(t, foundWhenInside)
	assert.Equal(t, "newMethod", methodWhenInside.Name.Content)
	assert.False(t, foundWhenAfter)
}
