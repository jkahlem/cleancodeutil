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
	Range              Range                          `json:"range" mapstructure:"range"`
	Message            string                         `json:"message" mapstructure:"message"`
	Severity           DiagnosticSeverity             `json:"severity,omitempty" mapstructure:"severity,omitempty"`
	Code               interface{}                    `json:"code,omitempty" mapstructure:"code,omitempty"` // string or integer
	CodeDescription    *CodeDescription               `json:"codeDescription,omitempty" mapstructure:"codeDescription,omitempty"`
	Source             string                         `json:"source,omitempty" mapstructure:"source,omitempty"`
	Tags               []DiagnosticTag                `json:"tags,omitempty" mapstructure:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty" mapstructure:"relatedInformation,omitempty"`
	Data               interface{}                    `json:"data,omitempty" mapstructure:"data,omitempty"`
}

type CodeDescription struct {
	HRef URI `json:"href" mapstructure:"href"`
}

type DiagnosticRelatedInformation struct {
	Location Location `json:"location" mapstructure:"location"`
	Message  string   `json:"message" mapstructure:"message"`
}
