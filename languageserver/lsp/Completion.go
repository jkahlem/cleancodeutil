package lsp

type CompletionTriggerKind int

const (
	Invoked                         CompletionTriggerKind = 1
	TriggerCharacter                CompletionTriggerKind = 2
	TriggerForIncompleteCompletions CompletionTriggerKind = 3
)

type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items,omitempty"`
}

type InsertTextFormat int

// TODO: no prefixes?
const (
	ITF_PlainText InsertTextFormat = 1
	ITF_Snippet   InsertTextFormat = 2
)

type InsertReplaceEdit struct {
	NewText string `json:"newText"`
	// Range if insert is requested
	Insert Range `json:"insert"`
	// Range if replace is requested
	Replace Range `json:"replace"`
}

type InsertTextMode int

const (
	AsIs              InsertTextMode = 1
	AdjustIndentation InsertTextMode = 2
)

type CompletionItemKind int

const (
	Text          CompletionItemKind = 1
	Method        CompletionItemKind = 2
	Function      CompletionItemKind = 3
	Constructor   CompletionItemKind = 4
	Field         CompletionItemKind = 5
	Variable      CompletionItemKind = 6
	Class         CompletionItemKind = 7
	Interface     CompletionItemKind = 8
	Module        CompletionItemKind = 9
	Property      CompletionItemKind = 10
	Unit          CompletionItemKind = 11
	Value         CompletionItemKind = 12
	Enum          CompletionItemKind = 13
	Keyword       CompletionItemKind = 14
	Snippet       CompletionItemKind = 15
	Color         CompletionItemKind = 16
	File          CompletionItemKind = 17
	Reference     CompletionItemKind = 18
	Folder        CompletionItemKind = 19
	EnumMember    CompletionItemKind = 20
	Constant      CompletionItemKind = 21
	Struct        CompletionItemKind = 22
	Event         CompletionItemKind = 23
	Operator      CompletionItemKind = 24
	TypeParameter CompletionItemKind = 25
)

type CompletionItemTag int

// TODO: no prefixes ...
const CIT_Deprecated CompletionItemTag = 1

type CompletionItem struct {
	Label               string              `json:"label"`
	Kind                CompletionItemKind  `json:"kind,omitempty"`
	Tags                []CompletionItemTag `json:"tags,omitempty"`
	Detail              string              `json:"detail,omitempty"`
	Documentation       string              `json:"documentation,omitempty"`
	Preselect           bool                `json:"preselect,omitempty"`
	SortText            string              `json:"sortText,omitempty"`
	FilterText          string              `json:"filterText,omitempty"`
	InsertText          string              `json:"insertText,omitempty"`
	InsertTextFormat    *InsertTextFormat   `json:"insertTextFormat,omitempty"`
	InsertTextMode      *InsertTextMode     `json:"insertTextMode,omitempty"`
	TextEdit            *InsertReplaceEdit  `json:"textEdit,omitempty"`
	AdditionalTextEdits Unsupported         `json:"additionalTextEdits,omitempty"`
	CommitCharacters    []string            `json:"commitCharacters,omitempty"`
	Command             Unsupported         `json:"command,omitempty"`
	Data                interface{}         `json:"data,omitempty"`
}
