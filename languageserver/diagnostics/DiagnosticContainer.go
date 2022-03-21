package diagnostics

import (
	"regexp"
	"strings"

	"returntypes-langserver/languageserver/lsp"
)

// Contains diagnostics for a file.
type DiagnosticContainer struct {
	diagnostics []ExpectedReturnTypeDiagnostic
	version     int
}

func (c *DiagnosticContainer) Version() int {
	return c.version
}

func (c *DiagnosticContainer) Diagnostics() []ExpectedReturnTypeDiagnostic {
	out := make([]ExpectedReturnTypeDiagnostic, len(c.diagnostics))
	copy(out, c.diagnostics)
	return out
}

// Sets a new set of diagnostics and raises the container's version
func (c *DiagnosticContainer) SetDiagnostics(diagnostics []ExpectedReturnTypeDiagnostic) {
	c.diagnostics = diagnostics
	c.version++
}

// Updates the text according to the event. Returns true if the update effects the diagnostics positions, otherwise returns false.
func (c *DiagnosticContainer) UpdatePositions(event lsp.TextDocumentContentChangeEvent) bool {
	remainingDiagnostics := make([]ExpectedReturnTypeDiagnostic, 0, len(c.diagnostics))
	updated := false
	for _, diagnostic := range c.diagnostics {
		if !overlapsDiagnostic(diagnostic, event) {
			applyDiagnosticPositionChange(&diagnostic, event)
			remainingDiagnostics = append(remainingDiagnostics, diagnostic)
		} else {
			updated = true
		}
	}
	c.diagnostics = remainingDiagnostics
	if updated {
		c.version++
	}
	return updated
}

// Applies the position changes of the change event to the method name range and return type range of the diagnostic.
// Returns true if any of these has changed.
func applyDiagnosticPositionChange(diagnostic *ExpectedReturnTypeDiagnostic, event lsp.TextDocumentContentChangeEvent) bool {
	methodChanged := applyRangePositionChange(&diagnostic.MethodNameRange, event)
	returnTypeChanged := applyRangePositionChange(&diagnostic.ReturnTypeRange, event)
	return methodChanged || returnTypeChanged
}

// Applies the position change of the range according to the change event and returns true, if the position of the diagnostic has changed.
func applyRangePositionChange(diagnosticRange *lsp.Range, event lsp.TextDocumentContentChangeEvent) bool {
	// We do not need to care about overlapping ranges and cases where eventRange comes after the diagnostic
	// as overlapping ranges are already filtered out and ranges after the diagnostic range will not change the
	// diagnostics position.
	// We only look at cases where the eventRange (e.g. the selected text or cursor position before text change)
	// comes before the diagnostic.
	eventRange := event.Range
	if diagnosticRange.Start.IsAfter(eventRange.End) || diagnosticRange.Start.IsSame(eventRange.End) {
		// There are three possible text events:
		// - Text removal (Backspace, Cut etc.)
		// - Text insertion (Pressing keys, Pasting etc.)
		// - Text replacement (Select and press keys/paste, Find and Replace etc.)
		//
		// The text removal is always removing a given selection (eventRange) while the text insertion is inserting event.Text to eventRange.Start.
		// Text replacements can be seen as a combination of these two operations.
		startBeforeChanges, endBeforeChanges := diagnosticRange.Start, diagnosticRange.End
		if !eventRange.Start.IsSame(eventRange.End) {
			applyRemoval(diagnosticRange, eventRange)
		}
		if len(event.Text) > 0 {
			applyInsertion(diagnosticRange, eventRange.Start, event.Text)
		}
		return !startBeforeChanges.IsSame(diagnosticRange.Start) || !endBeforeChanges.IsSame(diagnosticRange.End)
	}
	return false
}

func applyRemoval(diagnosticRange, eventRange *lsp.Range) {
	if eventRange.End.Line == diagnosticRange.Start.Line {
		charDiff := eventRange.End.Character - eventRange.Start.Character
		diagnosticRange.Start.Character -= charDiff
		if diagnosticRange.Start.Line == diagnosticRange.End.Line {
			// if the range is only one line, the end will move aswell
			diagnosticRange.End.Character -= charDiff
		}
	}
	lineDiff := eventRange.End.Line - eventRange.Start.Line
	diagnosticRange.Start.Line -= lineDiff
	diagnosticRange.End.Line -= lineDiff
}

func applyInsertion(diagnosticRange *lsp.Range, cursorPosition lsp.Position, text string) {
	lines := strings.Split(text, "\n")
	if diagnosticRange.Start.Line == cursorPosition.Line {
		addedCharsInDiagnosticLine := len(lines[len(lines)-1])
		diagnosticRange.Start.Character += addedCharsInDiagnosticLine
		if diagnosticRange.Start.Line == diagnosticRange.End.Line {
			// if the range is only one line, the end will move aswell
			diagnosticRange.End.Character += addedCharsInDiagnosticLine
		}
	}
	// move the diagnostic for the inserted lines
	diagnosticRange.Start.Line += len(lines) - 1
	diagnosticRange.End.Line += len(lines) - 1
}

func overlapsDiagnostic(diagnostic ExpectedReturnTypeDiagnostic, event lsp.TextDocumentContentChangeEvent) bool {
	if overlapsRange(diagnostic.MethodNameRange, event) || overlapsRange(diagnostic.ReturnTypeRange, event) {
		return true
	}
	return false
}

func overlapsRange(diagnostic lsp.Range, event lsp.TextDocumentContentChangeEvent) bool {
	eventRange := event.Range
	if diagnostic.End.IsBefore(eventRange.Start) || diagnostic.Start.IsAfter(eventRange.End) {
		return false
	}
	// if the inserted content is colliding (so not directly overlapping) with the diagnostic range then:
	// - Check if the char directly neighbouring to the diagnostic is some alpha(numeric) char which indicates a new name or something
	if len(event.Text) > 0 {
		if diagnostic.End.IsSame(eventRange.Start) {
			firstChar := event.Text[0:1]
			return isIdentifierChar(firstChar, false)
		} else if diagnostic.Start.IsSame(eventRange.End) {
			lastChar := event.Text[len(event.Text)-1 : len(event.Text)]
			return isIdentifierChar(lastChar, true)
		}
	} else {
		// Here could be something for checking for wrapping whitespaces at removal,
		// but for this, we need to track the whole text document for changes and apply them (using the range stuff)
		// which is not a simple thing to implement. In the context of this implementation, it is not worth the effort.
	}
	return true
}

func isIdentifierChar(char string, onBeginning bool) bool {
	identifierPattern := "_[a-zA-Z0-9]"
	if onBeginning {
		identifierPattern = "_[a-zA-Z]"
	}
	regexp.MustCompile(identifierPattern).Match([]byte(char))
	return false
}
