package lsp

// Client capabilities
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ClientCapabilities struct {
	TextDocument *TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Workspace    *WorkspaceClientCapabilities    `json:"workspace,omitempty"`
	Window       *WindowClientCapabilities       `json:"window,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Synchronization    *TextDocumentSyncClientCapabilities   `json:"synchronization,omitempty"`
	PublishDiagnostics *PublishDiagnosticsClientCapabilities `json:"publishDiagnostics,omitempty"`
	Completion         *CompletionClientCapabilities         `json:"completion,omitempty"`
}

type TextDocumentSyncClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	WillSave            bool `json:"willSave,omitempty"`
	WillSaveWaitUntil   bool `json:"willSaveWaitUntil,omitempty"`
	DidSave             bool `json:"didSave,omitempty"`
}

type PublishDiagnosticsClientCapabilities struct {
	RelatedInformation     bool              `json:"relatedInformation,omitempty"`
	TagSupport             *ClientTagSupport `json:"tagSupport,omitempty"`
	VersionSupport         bool              `json:"versionSupport,omitempty"`
	CodeDescriptionSupport bool              `json:"codeDescriptionSupport,omitempty"`
	DataSupport            bool              `json:"dataSupport,omitempty"`
}

type WorkspaceClientCapabilities struct {
	Configuration bool `json:"configuration,omitempty"`
}

type ClientTagSupport struct {
	ValueSet []DiagnosticTag `json:"valueSet,omitempty"`
}

type CompletionClientCapabilities struct {
	DynamicRegistration bool                                  `json:"dynamicRegistration"`
	CompletionItem      *CompletionItemClientCapabilities     `json:"completionItem,omitempty"`
	CompletionItemKind  *CompletionItemKindClientCapabilities `json:"completionItemKind,omitempty"`
	ContextSupport      bool                                  `json:"contextSupport"`
}

type CompletionItemClientCapabilities struct {
	SnippetSupport          bool           `json:"snippetSupport"`
	CommitCharactersSupport bool           `json:"commitCharactersSupport"`
	DocumentationFormat     NotImplemented `json:"documentationFormat,omitempty"`
	DeprecatedSupport       bool           `json:"deprecatedSupport"`
	PreselectSupport        bool           `json:"preselectSupport"`
	TagSupport              NotImplemented `json:"tagSupport,omitempty"`
	InsertReplaceSupport    bool           `json:"insertReplaceSupport"`
	ResolveSupport          NotImplemented `json:"resolveSupport,omitempty"`
	InsertTextModeSupport   NotImplemented `json:"insertTextModeSupport,omitempty"`
}

type CompletionItemKindClientCapabilities struct {
	ValueSet []CompletionItemKind `json:"valueSet,omitempty"`
}

type WindowClientCapabilities struct {
	WorkDoneProgress bool `json:"workDoneProgress,omitempty"`
}
