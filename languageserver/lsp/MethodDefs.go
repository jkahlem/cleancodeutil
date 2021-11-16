package lsp

// The implemented methods of the language server protocol
const (
	MethodInitialize  = "initialize"
	MethodInitialized = "initialized"
	MethodShutdown    = "shutdown"
	MethodExit        = "exit"

	MethodTextDocument_DidOpen   = "textDocument/didOpen"
	MethodTextDocument_DidChange = "textDocument/didChange"
	MethodTextDocument_DidClose  = "textDocument/didClose"
	MethodTextDocument_DidSave   = "textDocument/didSave"

	MethodWorkspace_DidCreate              = "workspace/didCreateFiles"
	MethodWorkspace_DidRename              = "workspace/didRenameFiles"
	MethodWorkspace_DidDelete              = "workspace/didDeleteFiles"
	MethodWorkspace_DidChangeConfiguration = "workspace/didChangeConfiguration"
	MethodWorkspace_ExecuteCommand         = "workspace/executeCommand"
)
