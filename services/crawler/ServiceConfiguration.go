package crawler

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/messages"
	"returntypes-langserver/common/transfer/rpc"
)

//go:generate go run ../serviceGenerator

func serviceConfiguration() rpc.ServiceConfiguration {
	conn := connection{}
	messager := messages.NewJson(&conn)
	return rpc.ServiceConfiguration{
		Connection: &conn,
		Controller: &Controller{},
		Messager:   messager,
		OnInterfaceCreationError: func(err errors.Error) {
			log.Error(err)
		},
	}
}

type Proxy struct {
	// Gets the content of a code file as a XML object
	GetFileContent func(path string, options Options) (string, errors.Error) `rpcmethod:"getFileContent" rpcparams:"path,options"`
	// Gets the content of all code files in a directory as one xml object
	GetDirectoryContents func(path string, options Options) (string, errors.Error) `rpcmethod:"getDirectoryContents" rpcparams:"path,options"`
}
