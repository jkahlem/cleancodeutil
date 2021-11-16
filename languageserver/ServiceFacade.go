package languageserver

import (
	"returntypes-langserver/common/errors"
	"returntypes-langserver/languageserver/lsp"
	"sync"
)

var singleton *languageServer
var singletonMutex sync.Mutex

func getSingleton() *languageServer {
	singletonMutex.Lock()
	defer singletonMutex.Unlock()

	if singleton == nil {
		singleton = createSingleton()
	}
	return singleton
}

func createSingleton() *languageServer {
	return &languageServer{}
}

func Configuration() *ServerConfiguration {
	return getSingleton().Configuration()
}

// Starts the language server.
func Startup() {
	getSingleton().Startup()
}

// Publishes diagnostics to the client.
func PublishDiagnostics(uri string, diagnostics []lsp.Diagnostic, version int) {
	getSingleton().PublishDiagnostics(uri, diagnostics, version)
}

func setClientCapabilities(clientCapabilities lsp.ClientCapabilities) {
	getSingleton().setClientCapabilities(clientCapabilities)
}

// Create virtual workspaces using the given workspace folders.
func createVirtualWorkspaces(workspaces []lsp.WorkspaceFolder) {
	getSingleton().createVirtualWorkspaces(workspaces)
}

// Adds a file on the given path into the virtual workspace if it does not exist there already.
func AddFileIfNotExists(path string) {
	getSingleton().AddFileIfNotExists(path)
}

// Reloads the file on the given path in all virtual workspaces containing it.
func ReloadFile(path string) {
	getSingleton().ReloadFile(path)
}

// Renames the file on the given path in all virtual workspaces containing it.
func RenameFile(oldPath, newPath string) {
	getSingleton().RenameFile(oldPath, newPath)
}

// Deletes the file on the given path in all virtual workspaces containing it.
func DeleteFile(path string) {
	getSingleton().DeleteFile(path)
}

// Updates the diagnostics of the given file in all workspaces containing it.
func UpdateDiagnostics(path string, changes []lsp.TextDocumentContentChangeEvent) {
	getSingleton().UpdateDiagnostics(path, changes)
}

// Refreshes the diagnostics for a given file.
func RefreshDiagnosticsForFile(path string) {
	getSingleton().RefreshDiagnosticsForFile(path)
}

// Refreshes the diagnostics for all files.
func RefreshDiagnosticsForAllFiles() {
	getSingleton().RefreshDiagnosticsForAllFiles()
}

// Shows a message to the user in the IDE.
func ShowMessage(msgType lsp.MessageType, message string) {
	getSingleton().ShowMessage(msgType, message)
}

// Makes a request to the user with possible actions (appears for example as buttons the user can click)
func ShowMessageRequest(msgType lsp.MessageType, message string, actions []Action) {
	getSingleton().ShowMessageRequest(msgType, message, actions)
}

// Logs a message to the IDE.
func LogMessage(msgType lsp.MessageType, message string) {
	getSingleton().LogMessage(msgType, message)
}

// Recovers the predictor connection.
func RecoverPredictor() {
	getSingleton().RecoverPredictor()
}

// Loads the extension configuration of the IDE.
func LoadConfiguration() chan errors.Error {
	return getSingleton().LoadConfiguration()
}

// Registers workspace/didChangeConfiguration for the extension's configuration sections.
// This capability needs to be registered explicitly, otherwise there will be no notifications.
func RegisterDidChangeWorkspaceCapability() chan errors.Error {
	return getSingleton().RegisterDidChangeWorkspaceCapability()
}

// Registers a capability.
func RegisterCapability(registrations ...lsp.Registration) chan errors.Error {
	return getSingleton().RegisterCapability(registrations...)
}
