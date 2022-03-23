package languageserver

import (
	"encoding/json"
	"fmt"
	"os"
	"returntypes-langserver/common/code/java/parser"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/rpc"
	"returntypes-langserver/common/transfer/rpc/jsonrpc"
	"returntypes-langserver/languageserver/lsp"
	"returntypes-langserver/languageserver/workspace"
	"returntypes-langserver/services/predictor"
)

const LanguageServerName string = "returntypes"

type Controller struct {
	initialized bool
	shutdown    bool
}

// Registers methods which may be called from the language client (the IDE).
func (c *Controller) RegisterMethods(register rpc.MethodRegister) {
	// General protocol methods
	register.RegisterMethod(lsp.MethodInitialize, "capabilities,workspaceFolders,rootPath,rootURI", c.Initialize)
	register.RegisterMethod(lsp.MethodShutdown, "", c.Shutdown)
	register.RegisterMethod(lsp.MethodInitialized, "", c.Initialized)
	register.RegisterMethod(lsp.MethodExit, "", c.Exit)

	// Methods on workspace level
	register.RegisterMethod(lsp.MethodWorkspace_DidCreate, "files", c.WorkspaceDidCreate)
	register.RegisterMethod(lsp.MethodWorkspace_DidRename, "files", c.WorkspaceDidRename)
	register.RegisterMethod(lsp.MethodWorkspace_DidDelete, "files", c.WorkspaceDidDelete)
	register.RegisterMethod(lsp.MethodWorkspace_ExecuteCommand, "command,arguments", c.WorkspaceExecuteCommand)
	register.RegisterMethod(lsp.MethodWorkspace_DidChangeConfiguration, "settings", c.WorkspaceDidChangeConfiguration)

	// Methods on text document level
	register.RegisterMethod(lsp.MethodTextDocument_DidOpen, "textDocument", c.TextDocumentDidOpen)
	register.RegisterMethod(lsp.MethodTextDocument_DidChange, "textDocument,contentChanges", c.TextDocumentDidChange)
	register.RegisterMethod(lsp.MethodTextDocument_DidClose, "textDocument", c.TextDocumentDidClose)
	register.RegisterMethod(lsp.MethodTextDocument_DidSave, "textDocument,text", c.TextDocumentDidSave)
	register.RegisterMethod(lsp.MethodTextDocument_Completion, "textDocument,position,context", c.TextDocumentCompletion)
}

// Callable RPC method.
// Will be called by the language client on startup of the language server.
// The language server will check the client capabilities and loads the workspaces the client has opened.
func (c *Controller) Initialize(capabilities lsp.ClientCapabilities, workspaceFolders []lsp.WorkspaceFolder, rootPath, rootURI string) (lsp.InitializeResult, error) {
	if c.initialized {
		responseError := jsonrpc.NewResponseError(jsonrpc.InvalidRequest, "The initialize request may only be sent once.")
		return lsp.InitializeResult{}, responseError
	}

	setClientCapabilities(capabilities)
	if workspaceFolders != nil {
		createVirtualWorkspaces(workspaceFolders)
	} else {
		createVirtualWorkspaces(c.createWorkspaceFolderOfRootInfo(rootPath, rootURI))
	}

	c.initialized = true
	return lsp.InitializeResult{
		Capabilities: Configuration().ServerCapabilities(),
		ServerInfo:   Configuration().ServerInfo(),
	}, nil
}

// The language client may store the workspace info as a seperate path/uri value.
func (c *Controller) createWorkspaceFolderOfRootInfo(rootPath, rootURI string) []lsp.WorkspaceFolder {
	if len(rootURI) == 0 {
		return []lsp.WorkspaceFolder{{Name: rootPath, URI: lsp.FilePathToDocumentURI(rootPath)}}
	} else {
		return []lsp.WorkspaceFolder{{Name: rootURI, URI: lsp.DocumentURI(rootURI)}}
	}
}

// Callable RPC method.
// Will be called by the language client to shutdown the server.
func (c *Controller) Shutdown() {
	c.shutdown = true
}

// Callable RPC method.
// Will be called by the language client after it has initialized (/received the response to initialize)
// The language server will create diagnostics for all project files in this step.
func (c *Controller) Initialized() {
	if err := <-LoadConfiguration(); err != nil {
		log.Error(err)
	}
	if err := <-RegisterDidChangeWorkspaceCapability(); err != nil {
		log.Error(err)
	}
	RefreshDiagnosticsForAllFiles()
}

// Callable RPC method.
// Will be called by the language client to close the server process.
func (c *Controller) Exit() {
	log.Close()
	if c.shutdown {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// Callable RPC method.
// Will be called by the language client if the user opens a (java) file.
func (c *Controller) TextDocumentDidOpen(textDocument lsp.TextDocumentItem) {
	if path, err := lsp.DocumentURIToFilePath(textDocument.URI); err != nil {
		log.Error(err)
	} else {
		AddFileIfNotExists(path, textDocument.Text)
	}
}

// Callable RPC method.
// Will be called by the language client if the user make changes in a (java) file.
func (c *Controller) TextDocumentDidChange(textDocument lsp.VersionedTextDocumentIdentifier, contentChanges []lsp.TextDocumentContentChangeEvent) {
	path, _ := lsp.DocumentURIToFilePath(textDocument.URI)
	UpdateDiagnostics(path, contentChanges)
	UpdateDocuments(path, contentChanges)
}

// Callable RPC method.
// Will be called by the language client if the user closes a (java) file.
// Support is required by the protocol to be able to receive textDocument/didClose notifications.
func (c *Controller) TextDocumentDidClose(textDocument lsp.TextDocumentIdentifier) {
	// do nothing
}

// Callable RPC method.
// Will be called by the language client if the user saves a (java) file.
func (c *Controller) TextDocumentDidSave(textDocument lsp.TextDocumentIdentifier, text string) {
	if path, err := lsp.DocumentURIToFilePath(textDocument.URI); err != nil {
		log.Error(err)
	} else {
		ReloadFile(path)
		RefreshDiagnosticsForFile(path)
	}
}

// Callable RPC method.
// Will be called by the language client if the user creates a (java) file.
func (c *Controller) WorkspaceDidCreate(files []lsp.FileCreate) {
	for _, file := range files {
		if path, err := lsp.DocumentURIToFilePath(lsp.DocumentURI(file.Uri)); err != nil {
			log.Error(err)
		} else {
			AddFileIfNotExists(path, "")
		}
	}
}

// Callable RPC method.
// Will be called by the language client if the user renames a (java) file.
func (c *Controller) WorkspaceDidRename(files []lsp.FileRename) {
	for _, file := range files {
		if oldPath, err := lsp.DocumentURIToFilePath(lsp.DocumentURI(file.OldUri)); err != nil {
			log.Error(err)
		} else if newPath, err := lsp.DocumentURIToFilePath(lsp.DocumentURI(file.NewUri)); err != nil {
			log.Error(err)
		} else {
			RenameFile(oldPath, newPath)
		}
	}
}

// Callable RPC method.
// Will be called by the language client if the user deletes (java) file.
func (c *Controller) WorkspaceDidDelete(files []lsp.FileDelete) {
	for _, file := range files {
		if path, err := lsp.DocumentURIToFilePath(lsp.DocumentURI(file.Uri)); err != nil {
			log.Error(err)
		} else {
			DeleteFile(path)
		}
	}
}

// Callable RPC method
// Will be called by the language client if the language server should execute a predefined command.
func (c *Controller) WorkspaceExecuteCommand(command string, arguments []interface{}) error {
	switch command {
	case CommandRefreshDiagnostics:
		RefreshDiagnosticsForAllFiles()
		break
	case CommandReconnectToPredictor:
		RecoverPredictor()
		break
	}
	return nil
}

// Callable RPC method.
// Will be called by the language client if the user changed configuration settings.
func (c *Controller) WorkspaceDidChangeConfiguration(settings interface{}) error {
	if settingsMap, ok := settings.(map[string]interface{}); ok {
		if asJson, err := json.Marshal(settingsMap[ReturnTypesConfigSection]); err != nil {
			log.Error(errors.Wrap(err, "Server error", "Could not parse client configuration"))
		} else {
			configuration.LoadConfigFromJsonString(string(asJson))
		}

	}
	return nil
}

// Callable RPC method.
// Will be called by the language client if a completion request is triggered (by typing a special character etc..)
func (c *Controller) TextDocumentCompletion(textDocument lsp.TextDocumentIdentifier, position lsp.Position, context lsp.CompletionContext) (*lsp.CompletionList, error) {
	// TODO: Handle completion request...
	// The below implementation is just an example-wise implementation.
	// The example case: user types "completionTest" as method name and then types a opening bracket '('
	// The IDE will autocomplete it as two brackets '()' and will send a request to the language server.
	// The language server will therefore compute something for auto completion, which happens here.
	// the normal "textedit" property will give the main-text edit (which needs to be on the same line !!), while the additionalTextEdits are for
	// edits on other places.
	//
	// Short things short, here is still a lot to do. By the way, the '¥n' in the textEdit text is not working as line breaks,
	// don't know how it is for additionalTextEdits. Maybe multiple-line edits needs multiple additionalTextEdits ...

	// 1. Get file contents
	// 2. Parse file contents to xml <- does this even work for unfinished files?
	// 3. Check if cursor position is actually method beginning
	// 4. Get method name
	// 5. Generate parameter list + return type (-> call predictor)
	// 6. Convert predictor output to completion item & return it

	log.Info("Got textDocument.completion request with char '%s' and kind %v", context.TriggerCharacter, context.TriggerKind)
	list := lsp.CompletionList{
		IsIncomplete: false,
		Items:        []lsp.CompletionItem{},
	}
	if path, err := lsp.DocumentURIToFilePath(textDocument.URI); err != nil {
		return nil, err
	} else if file := GetFile(path); file != nil {
		doc := file.Document()
		if method, found := findMethodAtCursorPosition(doc, position); found {
			// Generate parameter list
			if set, ok := configuration.FindDatasetByReference(configuration.LanguageServerMethodGenerationDataset()); ok {
				value, err := predictor.OnDataset(set).GenerateMethods([]predictor.MethodContext{{
					MethodName: predictor.GetPredictableMethodName(method.Name.Content),
					ClassName:  "Example",
					IsStatic:   false,
				}})
				if err != nil {
					return nil, err
				}

				// convert output to completion item & return it
				insertionRange := lsp.Range{
					Start: doc.ToPosition(method.RoundBraces.Range.Start + 1),
					End:   doc.ToPosition(method.RoundBraces.Range.End),
				}
				item := createCompletionItem(createTextEdit(joinParameterList(value[0].Parameters), insertionRange))
				list.Items = append(list.Items, item)
			}
		}
	}
	return &list, nil
}

func joinParameterList(value []predictor.Parameter) string {
	output := ""
	for i, par := range value {
		if i > 0 {
			output += fmt.Sprintf(", %s %s", par.Type, par.Name)
		} else {
			output += fmt.Sprintf("%s %s", par.Type, par.Name)
		}
	}
	return output
}

func createCompletionItem(textEdits ...lsp.TextEdit) lsp.CompletionItem {
	item := lsp.CompletionItem{
		Label:            "TestAsd",
		Kind:             lsp.Text,
		Preselect:        true,
		InsertTextFormat: lsp.ITF_PlainText,
		InsertTextMode:   lsp.AsIs,
		SortText:         "TestAsd",
		FilterText:       "(TestAsd",
	}
	if len(textEdits) >= 1 {
		item.TextEdit = &textEdits[0]
		item.AdditionalTextEdits = textEdits[1:]
	}

	return item
}

func createTextEdit(text string, r lsp.Range) lsp.TextEdit {
	return lsp.TextEdit{
		NewText: text,
		Range:   r,
	}
}

func createCompletionItemOld(position lsp.Position) lsp.CompletionItem {
	testStr := "completionTest" // completionTest(*)
	// Cursor position ---------------------------^
	return lsp.CompletionItem{
		Label:            "TestAsd",
		Kind:             lsp.Text,
		Preselect:        true,
		InsertTextFormat: lsp.ITF_PlainText,
		InsertTextMode:   lsp.AsIs,
		SortText:         "TestAsd",
		FilterText:       "(TestAsd",
		TextEdit: &lsp.TextEdit{
			NewText: "(int someNumber) {¥n¥treturn 0;¥n}",
			Range: lsp.Range{Start: lsp.Position{
				Line:      position.Line,
				Character: position.Character - 1,
			}, End: lsp.Position{
				Line:      position.Line,
				Character: position.Character + 1,
			}},
		},
		AdditionalTextEdits: []lsp.TextEdit{
			{
				NewText: "public void ",
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      position.Line,
						Character: position.Character - 1 - len(testStr),
					}, End: lsp.Position{
						Line:      position.Line,
						Character: position.Character - 1 - len(testStr),
					},
				},
			},
		},
	}
}

func findMethodAtCursorPosition(doc *workspace.Document, cursorPosition lsp.Position) (parser.Method, bool) {
	methods := parser.ParseMethods(doc.Text())
	cursorOffset := doc.ToOffset(cursorPosition)
	for _, m := range methods {
		// the range where the cursor might be to track the auto completion
		start, end := m.Name.Range.End, m.RoundBraces.Range.Start+1
		if cursorOffset >= start && cursorOffset <= end {
			return m, true
		}
	}
	return parser.Method{}, false
}
