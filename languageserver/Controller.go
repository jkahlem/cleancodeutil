package languageserver

import (
	"encoding/json"
	"os"
	"returntypes-langserver/common/code/java/parser"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/rpc"
	"returntypes-langserver/common/transfer/rpc/jsonrpc"
	"returntypes-langserver/languageserver/lsp"
	"returntypes-langserver/languageserver/workspace"
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
	register.RegisterMethod(lsp.MethodTextDocument_Completion, "textDocument,position,context,workDoneToken", c.TextDocumentCompletion)
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
	if path, err := lsp.DocumentURIToFilePath(textDocument.URI); err == nil {
		UpdateDiagnostics(path, contentChanges)
		UpdateDocuments(path, contentChanges)
	}
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
func (c *Controller) TextDocumentCompletion(textDocument lsp.TextDocumentIdentifier, position lsp.Position, context lsp.CompletionContext, workDoneToken interface{}) (*lsp.CompletionList, error) {
	list := lsp.CompletionList{
		IsIncomplete: false,
		Items:        []lsp.CompletionItem{},
	}

	if context.TriggerCharacter == "(" && IsMethodGenerationActive() {
		progress := StartProgress("Method autocompletion", "Generate method declaration", workDoneToken)
		defer progress.Close()

		if path, err := lsp.DocumentURIToFilePath(textDocument.URI); err != nil {
			return nil, err
		} else if file := GetFile(path); file != nil {
			doc := file.Document()
			if item, err := c.createMethodDefinitionCompletion(doc, position); err != nil {
				return nil, err
			} else if item != nil {
				list.Items = append(list.Items, *item)
			}
		}
	}

	return &list, nil
}

func (c *Controller) createMethodDefinitionCompletion(doc *workspace.Document, position lsp.Position) (*lsp.CompletionItem, errors.Error) {
	if method, found := c.findMethodAtCursorPosition(doc, position); found && c.canCompleteMethodDefinition(method) {
		return CompleteMethodDefinition(method, doc)
	}
	return nil, nil
}

func (c *Controller) canCompleteMethodDefinition(method Method) bool {
	for _, annotation := range method.Annotations {
		if annotation.Content == "@Override" {
			return false
		}
	}
	return true
}

func (c *Controller) findMethodAtCursorPosition(doc *workspace.Document, cursorPosition lsp.Position) (Method, bool) {
	methods := c.getMethods(parser.Parse(doc.Text()))
	cursorOffset := doc.ToOffset(cursorPosition)
	for _, m := range methods {
		// the range where the cursor might be to track the auto completion
		start, end := m.Name.Range.End, m.RoundBraces.Range.Start+1
		if cursorOffset >= start && cursorOffset <= end {
			return m, true
		}
	}
	return Method{}, false
}

func (c *Controller) getMethods(class *parser.Class) []Method {
	if class == nil {
		return nil
	}
	methods := make([]Method, len(class.Methods))
	for i, method := range class.Methods {
		methods[i] = Method{
			Method:    method,
			ClassName: class.Name.Content,
		}
	}
	for _, subClass := range class.Classes {
		methods = append(methods, c.getMethods(&subClass)...)
	}
	return methods
}

// Extend method with class name for method name generation
type Method struct {
	parser.Method
	ClassName string
}
