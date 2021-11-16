package diagnostics

import (
	"fmt"
	"testing"

	"returntypes-langserver/languageserver/lsp"

	"github.com/stretchr/testify/assert"
)

func TestUpdateTextOnLineInsertion(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 38))
	diagnostic2 := CreateDiagnostic(RangeFrom(20, 15).To(20, 25), RangeFrom(20, 27).To(20, 38))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1, diagnostic2})

	// when
	container.UpdateText(CreateInsertionEvent("a new line\n", At(15, 1)))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 2)
	assertRange(t, RangeFrom(10, 15).To(10, 25), updatedDiagnostics[0].ReturnTypeRange)
	assertRange(t, RangeFrom(10, 27).To(10, 38), updatedDiagnostics[0].MethodNameRange)
	assertRange(t, RangeFrom(21, 15).To(21, 25), updatedDiagnostics[1].ReturnTypeRange)
	assertRange(t, RangeFrom(21, 27).To(21, 38), updatedDiagnostics[1].MethodNameRange)
}

func TestUpdateTextOnInsertionInSameLineBetweenDiagnostic(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 30).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})
	textToInsert := "inserted text inline"

	// when
	container.UpdateText(CreateInsertionEvent(textToInsert, At(10, 27)))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 1)
	assertRange(t, RangeFrom(10, 15).To(10, 25), updatedDiagnostics[0].ReturnTypeRange)
	assertRange(t, RangeFrom(10, 30+len(textToInsert)).To(10, 42+len(textToInsert)), updatedDiagnostics[0].MethodNameRange)
}

func TestUpdateTextOnInsertionCollidingWithDiagnosticWithWrappingWhitespace(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})
	textToInsert := " inserted text wrapped in whitespace "

	// when
	container.UpdateText(CreateInsertionEvent(textToInsert, At(10, 27)))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 1)
	assertRange(t, RangeFrom(10, 15).To(10, 25), updatedDiagnostics[0].ReturnTypeRange)
	assertRange(t, RangeFrom(10, 27+len(textToInsert)).To(10, 42+len(textToInsert)), updatedDiagnostics[0].MethodNameRange)
}

func TestUpdateTextOnInsertionOverlappingDiagnostic(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})
	textToInsert := "insertedInsideReturnType"

	// when
	container.UpdateText(CreateInsertionEvent(textToInsert, At(10, 20)))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 0)
}

func TestUpdateTextOnSimpleRemoval(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 38))
	diagnostic2 := CreateDiagnostic(RangeFrom(20, 15).To(20, 25), RangeFrom(20, 27).To(20, 38))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1, diagnostic2})

	// when
	container.UpdateText(CreateRemovalEvent(RangeFrom(15, 1).To(16, 1)))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 2)
	assertRange(t, RangeFrom(10, 15).To(10, 25), updatedDiagnostics[0].ReturnTypeRange)
	assertRange(t, RangeFrom(10, 27).To(10, 38), updatedDiagnostics[0].MethodNameRange)
	assertRange(t, RangeFrom(19, 15).To(19, 25), updatedDiagnostics[1].ReturnTypeRange)
	assertRange(t, RangeFrom(19, 27).To(19, 38), updatedDiagnostics[1].MethodNameRange)
}

func TestUpdateTextOnRemovalInSameLineBetweenDiagnostic(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 30).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})

	// when
	container.UpdateText(CreateRemovalEvent(RangeFrom(10, 27).To(10, 28)))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 1)
	assertRange(t, RangeFrom(10, 15).To(10, 25), updatedDiagnostics[0].ReturnTypeRange)
	assertRange(t, RangeFrom(10, 29).To(10, 41), updatedDiagnostics[0].MethodNameRange)
}

func TestUpdateTextOnRemovalCollidingWithDiagnostic(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})

	// when
	container.UpdateText(CreateRemovalEvent(RangeFrom(10, 26).To(10, 27)))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 0)
}

func TestUpdateTextOnRemovalOverlappingWithDiagnostic(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})

	// when
	container.UpdateText(CreateRemovalEvent(RangeFrom(10, 20).To(10, 21)))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 0)
}

func TestUpdateTextOnTextReplacement(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 38))
	diagnostic2 := CreateDiagnostic(RangeFrom(20, 15).To(20, 25), RangeFrom(20, 27).To(20, 38))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1, diagnostic2})

	// when
	container.UpdateText(CreateReplacementEvent(RangeFrom(15, 1).To(16, 1), "New text with \n two new lines instead \n of the one line previously there."))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 2)
	assertRange(t, RangeFrom(10, 15).To(10, 25), updatedDiagnostics[0].ReturnTypeRange)
	assertRange(t, RangeFrom(10, 27).To(10, 38), updatedDiagnostics[0].MethodNameRange)
	assertRange(t, RangeFrom(21, 15).To(21, 25), updatedDiagnostics[1].ReturnTypeRange)
	assertRange(t, RangeFrom(21, 27).To(21, 38), updatedDiagnostics[1].MethodNameRange)
}

func TestUpdateTextOnTextReplacementInSameLineBetweenDiagnostic(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 30).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})
	textToInsert := " new text inserted between return type and method name "
	lengthToReplace := 1

	// when
	container.UpdateText(CreateReplacementEvent(RangeFrom(10, 27).To(10, 27+lengthToReplace), textToInsert))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 1)
	assertRange(t, RangeFrom(10, 15).To(10, 25), updatedDiagnostics[0].ReturnTypeRange)
	assertRange(t, RangeFrom(10, 30+len(textToInsert)-lengthToReplace).To(10, 42+len(textToInsert)-lengthToReplace), updatedDiagnostics[0].MethodNameRange)
}

func TestUpdateTextOnTextReplacementCollidingWithDiagnostics(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})
	textToInsert := " text to insert collding with method name but wrapped in whitespaces "
	lengthToReplace := 1

	// when
	container.UpdateText(CreateReplacementEvent(RangeFrom(10, 26).To(10, 26+lengthToReplace), textToInsert))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 1)
	assertRange(t, RangeFrom(10, 15).To(10, 25), updatedDiagnostics[0].ReturnTypeRange)
	assertRange(t, RangeFrom(10, 27+len(textToInsert)-lengthToReplace).To(10, 42+len(textToInsert)-lengthToReplace), updatedDiagnostics[0].MethodNameRange)
}

func TestUpdateTextOnTextReplacementOverlappingWithDiagnostics(t *testing.T) {
	// given
	container := DiagnosticContainer{}
	diagnostic1 := CreateDiagnostic(RangeFrom(10, 15).To(10, 25), RangeFrom(10, 27).To(10, 42))
	container.SetDiagnostics([]ExpectedReturnTypeDiagnostic{diagnostic1})

	// when
	container.UpdateText(CreateReplacementEvent(RangeFrom(10, 20).To(10, 21), "insideReturnType"))
	updatedDiagnostics := container.Diagnostics()

	// then
	assert.Len(t, updatedDiagnostics, 0)
}

// Helper structures

type rangeBuilder struct {
	from lsp.Position
	to   lsp.Position
}

func At(line, char int) lsp.Range {
	return RangeFrom(line, char).To(line, char)
}

func RangeFrom(line, char int) *rangeBuilder {
	builder := &rangeBuilder{}
	return builder.From(line, char)
}

func (r *rangeBuilder) From(line, char int) *rangeBuilder {
	r.from = lsp.Position{
		Line:      line,
		Character: char,
	}
	return r
}

func (r *rangeBuilder) To(line, char int) lsp.Range {
	r.to = lsp.Position{
		Line:      line,
		Character: char,
	}
	return lsp.Range{
		Start: r.from,
		End:   r.to,
	}
}

// Helper functions

func assertRange(t *testing.T, expected, actual lsp.Range) {
	if expected.Start.Line != actual.Start.Line || expected.Start.Character != actual.Start.Character ||
		expected.End.Line != actual.End.Line || expected.End.Character != actual.End.Character {
		assert.Fail(t, fmt.Sprintf("Expected range %s but is a range %s", formatRange(expected), formatRange(actual)))
	}
}

func formatRange(r lsp.Range) string {
	return fmt.Sprintf("from %d:%d to %d:%d", r.Start.Line, r.Start.Character, r.End.Line, r.End.Character)
}

func CreateInsertionEvent(text string, at lsp.Range) lsp.TextDocumentContentChangeEvent {
	return CreateChangeEventWith(text, at)
}

func CreateRemovalEvent(r lsp.Range) lsp.TextDocumentContentChangeEvent {
	return CreateChangeEventWith("", r)
}

func CreateReplacementEvent(rangeToReplace lsp.Range, textToInsertInstead string) lsp.TextDocumentContentChangeEvent {
	return CreateChangeEventWith(textToInsertInstead, rangeToReplace)
}

func CreateChangeEventWith(text string, r lsp.Range) lsp.TextDocumentContentChangeEvent {
	return lsp.TextDocumentContentChangeEvent{
		Text:  text,
		Range: &r,
	}
}

func CreateDiagnostic(returnTypeRange, methodRange lsp.Range) ExpectedReturnTypeDiagnostic {
	return ExpectedReturnTypeDiagnostic{
		MethodNameRange:    methodRange,
		ReturnTypeRange:    returnTypeRange,
		ExpectedReturnType: "void",
	}
}
