package languageserver

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/messages"
	"returntypes-langserver/common/transfer/rpc"
	"returntypes-langserver/languageserver/lsp"
)

//go:generate go run ../services/serviceGenerator

const LSPMediaType = "application/vscode-jsonrpc"
const LSPMIMEType = LSPMediaType + "; charset=utf-8"

func serviceConfiguration() rpc.ServiceConfiguration {
	conn := connection{}
	messager := messages.NewReadWriter(&conn)
	messager.AcceptMediaType(LSPMediaType)
	messager.SetWritingMimeType(LSPMIMEType)
	return rpc.ServiceConfiguration{
		Connection: &conn,
		Messager:   messager,
		Controller: &Controller{},
		OnInterfaceCreationError: func(err errors.Error) {
			log.Error(err)
		},
	}
}

type Proxy struct {
	PublishDiagnostics func(uri lsp.DocumentURI, diagnostics []lsp.Diagnostic, version int) `rpcmethod:"textDocument/publishDiagnostics" rpcparams:"uri,diagnostics,version"`
	ShowMessage        func(msgType lsp.MessageType, message string)                        `rpcmethod:"window/showMessage" rpcparams:"type,message"`
	ShowMessageRequest func(msgType lsp.MessageType, message string,
		actions []lsp.MessageActionItem) (lsp.MessageActionItem, errors.Error) `rpcmethod:"window/showMessageRequest" rpcparams:"type,message,actions"`
	LogMessage         func(msgType lsp.MessageType, message string)                     `rpcmethod:"window/logMessage" rpcparams:"type,message"`
	GetConfiguration   func(items []lsp.ConfigurationItem) ([]interface{}, errors.Error) `rpcmethod:"workspace/configuration" rpcparams:"items"`
	RegisterCapability func(registrations []lsp.Registration) errors.Error               `rpcmethod:"client/registerCapability" rpcparams:"registrations"`
}
