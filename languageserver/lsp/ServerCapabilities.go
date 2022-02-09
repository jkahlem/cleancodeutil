package lsp

type TextDocumentSyncKind int

const (
	SyncKindNone        TextDocumentSyncKind = 0
	SyncKindFull        TextDocumentSyncKind = 1
	SyncKindIncremental TextDocumentSyncKind = 2
)

type ServerCapabilities struct {
	TextDocumentSync       *TextDocumentSyncOptions     `json:"textDocumentSync,omitempty"`
	Workspace              *WorkspaceServerCapabilities `json:"workspace,omitempty"`
	ExecuteCommandProvider *ExecuteCommandOptions       `json:"executeCommandProvider,omitempty"`
	CompletionProvider     *CompletionOptions           `json:"completionProvider,omitempty"`
}

type TextDocumentSyncOptions struct {
	OpenClose         bool                 `json:"openClose,omitempty"`
	Change            TextDocumentSyncKind `json:"change,omitempty"`
	WillSave          bool                 `json:"willSave,omitempty"`
	WillSaveWaitUntil bool                 `json:"willSaveWaitUntil,omitempty"`
	// bool or SaveOptions object
	Save interface{} `json:"save,omitempty"`
}

type SaveOptions struct {
	IncludeText bool `json:"includeText,omitempty"`
}

type WorkspaceServerCapabilities struct {
	WorkspaceFolders *WorkspaceFoldersServerCapabilities `json:"workspaceFolders,omitempty"`
	FileOperations   *FileOperationsServerCapabilities   `json:"fileOperations,omitempty"`
}

type WorkspaceFoldersServerCapabilities struct {
	Supported bool `json:"supported,omitempty"`

	// string or bool. if string, its treated as id (see specification)
	ChangeNotifications interface{} `json:"changeNotifications,omitempty"`
}

type FileOperationsServerCapabilities struct {
	DidCreate  *FileOperationRegistrationOptions `json:"didCreate,omitempty"`
	WillCreate *FileOperationRegistrationOptions `json:"willCreate,omitempty"`
	DidRename  *FileOperationRegistrationOptions `json:"didRename,omitempty"`
	WillRename *FileOperationRegistrationOptions `json:"willRename,omitempty"`
	DidDelete  *FileOperationRegistrationOptions `json:"didDelete,omitempty"`
	WillDelete *FileOperationRegistrationOptions `json:"willDelete,omitempty"`
}

type FileOperationRegistrationOptions struct {
	Filters []FileOperationFilter `json:"filters"`
}

type FileOperationFilter struct {
	Scheme  string               `json:"scheme,omitempty"`
	Pattern FileOperationPattern `json:"pattern"`
}

type FileOperationPattern struct {
	Glob    string                       `json:"glob"`
	Matches FileOperationPatternKind     `json:"matches,omitempty"`
	Options *FileOperationPatternOptions `json:"options,omitempty"`
}

type FileOperationPatternOptions struct {
	IgnoreCase bool `json:"ignoreCase,omitempty"`
}

type FileOperationPatternKind string

const (
	FOPatternFile   FileOperationPatternKind = "file"
	FOPatternFolder FileOperationPatternKind = "folder"
)

type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

type CompletionOptions struct {
	WorkDoneProgress    bool     `json:"workDoneProgress"` // TODO: make this extendable ?
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	AllCommitCharacters []string `json:"allCommitCharacters,omitempty"`
	ResolveProvider     bool     `json:"resolveProvider"`
}
