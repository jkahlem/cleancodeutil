// This package contains several structures defined in the Language Server Protocol specification (version 3.16)
// at https://microsoft.github.io/language-server-protocol/specification
// Currently unused types/properties might not be implemented
package lsp

type DocumentURI string
type URI string

type Range struct {
	Start Position `json:"start" mapstructure:"start"`
	End   Position `json:"end" mapstructure:"end"`
}

type Position struct {
	Line      int `json:"line" mapstructure:"line"`
	Character int `json:"character" mapstructure:"character"`
}

type Location struct {
	URI   DocumentURI `json:"uri" mapstructure:"uri"`
	Range Range       `json:"range" mapstructure:"range"`
}

type WorkspaceFolder struct {
	URI  DocumentURI `json:"uri" mapstructure:"uri"`
	Name string      `json:"name" mapstructure:"name"`
}

type TextDocumentItem struct {
	URI        DocumentURI `json:"uri" mapstructure:"uri"`
	LanguageId string      `json:"languageId" mapstructure:"languageId"`
	Version    int         `json:"version" mapstructure:"version"`
	Text       string      `json:"text" mapstructure:"text"`
}

type TextDocumentIdentifier struct {
	URI DocumentURI `json:"uri" mapstructure:"uri"`
}

type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier `mapstructure:",squash"`
	Version                int `json:"version" mapstructure:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text        string `json:"text" mapstructure:"text"`
	Range       *Range `json:"range,omitempty" mapstructure:"range,omitempty"`
	RangeLength int    `json:"rangeLength,omitempty" mapstructure:"rangeLength,omitempty"`
}

type FileCreate struct {
	Uri string `json:"uri" mapstructure:"uri"`
}

type FileRename struct {
	OldUri string `json:"oldUri" mapstructure:"oldUri"`
	NewUri string `json:"newUri" mapstructure:"newUri"`
}

type FileDelete struct {
	Uri string `json:"uri" mapstructure:"uri"`
}

type MessageType int

const (
	MessageError   MessageType = 1
	MessageWarning MessageType = 2
	MessageInfo    MessageType = 3
	MessageLog     MessageType = 4
)

type MessageActionItem struct {
	Title string `json:"title" mapstructure:"title"`
}

type ConfigurationItem struct {
	ScopeURI DocumentURI `json:"scopeUri,omitempty" mapstructure:"scopeUri,omitempty"`
	Section  string      `json:"section,omitempty" mapstructure:"section,omitempty"`
}

// A type used for fields which presumably are not used but would need other types to be implemented and therefore
// cost too much effort
type NotImplemented interface{}
