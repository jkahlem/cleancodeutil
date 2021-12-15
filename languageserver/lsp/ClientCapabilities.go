package lsp

// Client capabilities
type ClientInfo struct {
	Name    string `json:"name" mapstructure:"name"`
	Version string `json:"version,omitempty" mapstructure:"version,omitempty"`
}

type ClientCapabilities struct {
	TextDocument *TextDocumentClientCapabilities `json:"textDocument,omitempty" mapstructure:"textDocument,omitempty"`
	Workspace    *WorkspaceClientCapabilities    `json:"workspace,omitempty" mapstructure:"workspace,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Synchronization    *TextDocumentSyncClientCapabilities   `json:"synchronization,omitempty" mapstructure:"synchronization,omitempty"`
	PublishDiagnostics *PublishDiagnosticsClientCapabilities `json:"publishDiagnostics,omitempty" mapstructure:"publishDiagnostics,omitempty"`
	Completion         *CompletionClientCapabilities         `json:"completion,omitempty" mapstructure:"completion,omitempty"`
}

type TextDocumentSyncClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty" mapstructure:"dynamicRegistration,omitempty"`
	WillSave            bool `json:"willSave,omitempty" mapstructure:"willSave,omitempty"`
	WillSaveWaitUntil   bool `json:"willSaveWaitUntil,omitempty" mapstructure:"willSaveWaitUntil,omitempty"`
	DidSave             bool `json:"didSave,omitempty" mapstructure:"didSave,omitempty"`
}

type PublishDiagnosticsClientCapabilities struct {
	RelatedInformation     bool              `json:"relatedInformation,omitempty" mapstructure:"relatedInformation,omitempty"`
	TagSupport             *ClientTagSupport `json:"tagSupport,omitempty" mapstructure:"tagSupport,omitempty"`
	VersionSupport         bool              `json:"versionSupport,omitempty" mapstructure:"versionSupport,omitempty"`
	CodeDescriptionSupport bool              `json:"codeDescriptionSupport,omitempty" mapstructure:"codeDescriptionSupport,omitempty"`
	DataSupport            bool              `json:"dataSupport,omitempty" mapstructure:"dataSupport,omitempty"`
}

type WorkspaceClientCapabilities struct {
	Configuration bool `json:"configuration,omitempty" mapstructure:"configuration,omitempty"`
}

type ClientTagSupport struct {
	ValueSet []DiagnosticTag `json:"valueSet,omitempty" mapstructure:"valueSet,omitempty"`
}

type CompletionClientCapabilities struct {
	DynamicRegistration bool                                  `json:"dynamicRegistration" mapstructure:"dynamicRegistration"`
	CompletionItem      *CompletionItemClientCapabilities     `json:"completionItem,omitempty" mapstructure:"completionItem,omitempty"`
	CompletionItemKind  *CompletionItemKindClientCapabilities `json:"completionItemKind,omitempty" mapstructure:"completionItemKind,omitempty"`
	ContextSupport      bool                                  `json:"contextSupport" mapstructure:"contextSupport"`
}

type CompletionItemClientCapabilities struct {
	SnippetSupport          bool           `json:"snippetSupport" mapstructure:"snippetSupport"`
	CommitCharactersSupport bool           `json:"commitCharactersSupport" mapstructure:"commitCharactersSupport"`
	DocumentationFormat     NotImplemented `json:"documentationFormat,omitempty" mapstructure:"documentationFormat,omitempty"`
	DeprecatedSupport       bool           `json:"deprecatedSupport" mapstructure:"deprecatedSupport"`
	PreselectSupport        bool           `json:"preselectSupport" mapstructure:"preselectSupport"`
	TagSupport              NotImplemented `json:"tagSupport,omitempty" mapstructure:"tagSupport,omitempty"`
	InsertReplaceSupport    bool           `json:"insertReplaceSupport" mapstructure:"insertReplaceSupport"`
	ResolveSupport          NotImplemented `json:"resolveSupport,omitempty" mapstructure:"resolveSupport,omitempty"`
	InsertTextModeSupport   NotImplemented `json:"insertTextModeSupport,omitempty" mapstructure:"insertTextModeSupport,omitempty"`
}

type CompletionItemKindClientCapabilities struct {
	ValueSet []CompletionItemKind `json:"valueSet,omitempty" mapstructure:"valueSet,omitempty"`
}
