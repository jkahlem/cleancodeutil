package languageserver

import (
	"io"
	"returntypes-langserver/common/log"
	"returntypes-langserver/common/messages"
	"returntypes-langserver/common/rpc"
	"sync"
)

const LSPMediaType = "application/vscode-jsonrpc"
const LSPMIMEType = LSPMediaType + "; charset=utf-8"

var interfaceSingleton *InterfaceWrapper
var interfaceMutex sync.Mutex

// Wraps the interface implementing failsafe methods to avoid nil pointer errors.
type InterfaceWrapper struct {
	ifc rpc.Interface
}

func (I *InterfaceWrapper) ProxyFacade() *ProxyFacade {
	if I.ifc != nil && I.ifc.ProxyFacade() != nil {
		if facade, ok := I.ifc.ProxyFacade().(*ProxyFacade); ok {
			return facade
		}
	}
	return &ProxyFacade{}
}

func (I *InterfaceWrapper) Connection() io.ReadWriter {
	if I.ifc != nil {
		return I.ifc.Connection()
	}
	return nil
}

func getInterface() *InterfaceWrapper {
	interfaceMutex.Lock()
	defer interfaceMutex.Unlock()

	if interfaceSingleton == nil {
		interfaceSingleton = createInterface()
	}
	return interfaceSingleton
}

func createInterface() *InterfaceWrapper {
	conn := connection{}
	messager := messages.NewReadWriter(&conn)
	messager.AcceptMediaType(LSPMediaType)
	messager.SetWritingMimeType(LSPMIMEType)

	controller := Controller{}
	wrapper := InterfaceWrapper{}
	ifc, err := rpc.CreateInterfaceOnConnection(&conn, messager).WithProxyFacade(&ProxyFacade{}).WithController(&controller).Finalize()
	if err != nil {
		log.Error(err)
	} else {
		wrapper.ifc = ifc
	}

	return &wrapper
}
