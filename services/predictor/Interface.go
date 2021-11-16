package predictor

import (
	"io"
	"sync"

	"returntypes-langserver/common/log"
	"returntypes-langserver/common/messages"
	"returntypes-langserver/common/rpc"
	"returntypes-langserver/common/rpc/jsonrpc"
)

var interfaceSingleton *InterfaceWrapper
var interfaceMutex sync.Mutex
var recoverFailedEventHandler []func(rpc.Recoverer)
var connectionErrorEventHandler []func(rpc.Recoverer)

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

// Returns the interface wrapper used for the predictor.
func getInterface() *InterfaceWrapper {
	interfaceMutex.Lock()
	defer interfaceMutex.Unlock()

	if interfaceSingleton == nil {
		interfaceSingleton = createInterface()
	}
	return interfaceSingleton
}

// Creates a new interface for the predictor.
func createInterface() *InterfaceWrapper {
	conn := &PredictorConnection{}
	wrapper := InterfaceWrapper{}
	messager := messages.NewReadWriter(conn)
	messager.AcceptMediaType(jsonrpc.MediaType)
	messager.SetWritingMimeType(jsonrpc.MediaType)

	ifc, err := rpc.CreateInterfaceOnConnection(conn, messager).WithProxyFacade(&ProxyFacade{}).
		OnRecoverFailed(func(r rpc.Recoverer) {
			// Call all handlers which are registered with OnRecoverFailed
			for _, fn := range recoverFailedEventHandler {
				if fn != nil {
					fn(r)
				}
			}
		}).
		OnConnectionError(func(r rpc.Recoverer) {
			// Call all handlers which are registered with OnConnectionError
			for _, fn := range connectionErrorEventHandler {
				if fn != nil {
					fn(r)
				}
			}
		}).Finalize()

	if err != nil {
		log.FatalError(err)
	} else {
		wrapper.ifc = ifc
	}

	return &wrapper
}

// Adds a handler for the RecoverFailed event.
func OnRecoverFailed(handler func(rpc.Recoverer)) {
	recoverFailedEventHandler = append(recoverFailedEventHandler, handler)
}

// Adds a handler for the ConnectionError event.
func OnConnectionError(handler func(rpc.Recoverer)) {
	connectionErrorEventHandler = append(connectionErrorEventHandler, handler)
}
