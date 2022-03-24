// This package contains several structures defined in the Language Server Protocol specification (version 3.16)
// at https://microsoft.github.io/language-server-protocol/specification
// Currently unused types/properties might not be implemented
package lsp

type DocumentURI string
type URI string

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Location struct {
	URI   DocumentURI `json:"uri"`
	Range Range       `json:"range"`
}

type WorkspaceFolder struct {
	URI  DocumentURI `json:"uri"`
	Name string      `json:"name"`
}

type TextDocumentItem struct {
	URI        DocumentURI `json:"uri"`
	LanguageId string      `json:"languageId"`
	Version    int         `json:"version"`
	Text       string      `json:"text"`
}

type TextDocumentIdentifier struct {
	URI DocumentURI `json:"uri"`
}

type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier `json:",squash"`
	Version                int `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text  string `json:"text"`
	Range *Range `json:"range,omitempty"`
}

type FileCreate struct {
	Uri string `json:"uri"`
}

type FileRename struct {
	OldUri string `json:"oldUri"`
	NewUri string `json:"newUri"`
}

type FileDelete struct {
	Uri string `json:"uri"`
}

type MessageType int

const (
	MessageError   MessageType = 1
	MessageWarning MessageType = 2
	MessageInfo    MessageType = 3
	MessageLog     MessageType = 4
)

type MessageActionItem struct {
	Title string `json:"title"`
}

type ConfigurationItem struct {
	ScopeURI DocumentURI `json:"scopeUri,omitempty"`
	Section  string      `json:"section,omitempty"`
}

// A type used for fields which presumably are not used but would need other types to be implemented and therefore
// cost too much effort
type NotImplemented interface{}
