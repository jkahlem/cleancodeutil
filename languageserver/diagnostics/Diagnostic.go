package diagnostics

import (
	"fmt"

	"returntypes-langserver/languageserver/lsp"
)

type ExpectedReturnTypeDiagnostic struct {
	MethodNameRange    lsp.Range
	ReturnTypeRange    lsp.Range
	ExpectedReturnType string
	Version            int
}

// Maps the expected return type diagnostics to lsp diagnostics
func MapExpectedReturnTypeDiagnostics(diagnostics []ExpectedReturnTypeDiagnostic) []lsp.Diagnostic {
	if diagnostics == nil {
		return nil
	}
	lspDiagnostics := make([]lsp.Diagnostic, len(diagnostics))
	for i, diagnostic := range diagnostics {
		lspDiagnostics[i] = MapExpectedReturnTypeDiagnostic(diagnostic)
	}
	return lspDiagnostics
}

// Maps the expected return type diagnostic to a lsp diagnostic
func MapExpectedReturnTypeDiagnostic(diagnostic ExpectedReturnTypeDiagnostic) lsp.Diagnostic {
	return lsp.Diagnostic{
		Severity: lsp.SeverityWarning,
		Message:  fmt.Sprintf("Expected return type: %s", diagnostic.ExpectedReturnType),
		Range: lsp.Range{
			Start: diagnostic.ReturnTypeRange.Start,
			End:   diagnostic.MethodNameRange.End,
		},
	}
}
