package lsp

import "returntypes-langserver/common/rpc/jsonrpc"

const LSPErrorTitle = "LSP Error"

const (
	// JSON-RPC
	ServerNotInitialized jsonrpc.ErrorCode = -32002
	UnknownErrorCode     jsonrpc.ErrorCode = -32001

	// LSP
	ContentModified  jsonrpc.ErrorCode = -32801
	RequestCancelled jsonrpc.ErrorCode = -32800
)
