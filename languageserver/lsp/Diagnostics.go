package lsp

type DiagnosticTag int

const (
	Unnecessary DiagnosticTag = 1
	Deprecated  DiagnosticTag = 2
)

type DiagnosticSeverity int

const (
	SeverityError       DiagnosticSeverity = 1
	SeverityWarning     DiagnosticSeverity = 2
	SeverityInformation DiagnosticSeverity = 3
	SeverityHint        DiagnosticSeverity = 4
)

type Diagnostic struct {
	Range              Range                          `json:"range"`
	Message            string                         `json:"message"`
	Severity           DiagnosticSeverity             `json:"severity,omitempty"`
	Code               interface{}                    `json:"code,omitempty"` // string or integer
	CodeDescription    *CodeDescription               `json:"codeDescription,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Tags               []DiagnosticTag                `json:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
	Data               interface{}                    `json:"data,omitempty"`
}

type CodeDescription struct {
	HRef URI `json:"href"`
}

type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}
