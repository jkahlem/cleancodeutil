package languageserver

import "returntypes-langserver/languageserver/lsp"

// Commands which the client can send using the workspace/executeCommand method.
const (
	CommandRefreshDiagnostics   = "returntypes.refreshDiagnostics"
	CommandReconnectToPredictor = "returntypes.predictor.reconnect"
)

type ServerConfiguration struct {
	clientCapabilities lsp.ClientCapabilities
	workspaces         []lsp.WorkspaceFolder
}

var config *ServerConfiguration

// Returns the capabilities of this language server implementation.
func (config *ServerConfiguration) ServerCapabilities() lsp.ServerCapabilities {
	return lsp.ServerCapabilities{
		TextDocumentSync: &lsp.TextDocumentSyncOptions{
			OpenClose: true,
			Change:    lsp.SyncKindIncremental,
			Save: lsp.SaveOptions{
				IncludeText: false,
			},
		},
		Workspace: &lsp.WorkspaceServerCapabilities{
			WorkspaceFolders: &lsp.WorkspaceFoldersServerCapabilities{
				Supported: true,
			},
			FileOperations: &lsp.FileOperationsServerCapabilities{
				DidCreate: &lsp.FileOperationRegistrationOptions{
					Filters: []lsp.FileOperationFilter{lsp.CreateFileOperationFilter("file", lsp.FOPatternFile, "**/*.java")},
				},
				DidRename: &lsp.FileOperationRegistrationOptions{
					Filters: []lsp.FileOperationFilter{lsp.CreateFileOperationFilter("file", lsp.FOPatternFile, "**/*.java")},
				},
				DidDelete: &lsp.FileOperationRegistrationOptions{
					Filters: []lsp.FileOperationFilter{lsp.CreateFileOperationFilter("file", lsp.FOPatternFile, "**/*.java")},
				},
			},
		},
		ExecuteCommandProvider: &lsp.ExecuteCommandOptions{
			Commands: []string{CommandRefreshDiagnostics, CommandReconnectToPredictor},
		},
	}
}

// Returns the server info.
func (config *ServerConfiguration) ServerInfo() *lsp.ServerInfo {
	return &lsp.ServerInfo{
		Name: LanguageServerName,
	}
}

// Returns the textDocument client capabilities.
func (config *ServerConfiguration) TextDocumentClientCapabilities() *lsp.TextDocumentClientCapabilities {
	return config.clientCapabilities.TextDocument
}

// Returns the textDocument.publishDiagnostics client capabilities
func (config *ServerConfiguration) PublishDiagnosticsClientCapabilities() *lsp.PublishDiagnosticsClientCapabilities {
	textDocumentCapabilities := config.TextDocumentClientCapabilities()
	if textDocumentCapabilities == nil {
		return nil
	}
	return textDocumentCapabilities.PublishDiagnostics
}

// Returns the workspace client capabilities
func (config *ServerConfiguration) WorkspaceClientCapabilities() *lsp.WorkspaceClientCapabilities {
	return config.clientCapabilities.Workspace
}

// Returns the workspace.configuration client capabilities
func (config *ServerConfiguration) ConfigurationClientCapabilities() bool {
	workspaceCapabilities := config.WorkspaceClientCapabilities()
	if workspaceCapabilities == nil {
		return false
	}
	return workspaceCapabilities.Configuration
}
