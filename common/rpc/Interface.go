// This package implements structures for handling messages using remote procedure call.
//
// A RPC interface in the context of this package is defined by:
// 1.) A connection which is used for communicating with the client
// 2.) A message protocol which is used to exchange single rpc messages with the client
// 3.) An optional server which defines the incoming interface
// 4.) An optional proxy (/client stub) wrapped in a proxy facade which defines the outgoing interface
//
// When creating an interface, the server will register it's invokable methods including the expected name and parameters.
// These are required to support method names with special characters and passing parameters using the by-name parameter structures.
//
// If a proxy facade is defined, the functions of the underlying proxy will be automatically implemented to make the requests
// to the client and return the result if expected and received. The proxy facade is some boilerplate code which ensures that
// calling the method won't end up in accessing nil pointers.
//
// A proxy facade should always contain one exported field with the proxy type and ```rpcproxy``` tag. The contents of this tag is irrelevant.
// The proxy should contain exported fields with function type defining the number and types of parameters and the expected return type.
// The return types of these functions can have the following structures:
// 1.) The function has no return type - in this case, the caller is not interested in any states or results of the function call, therefore
//     it is handled as a notification. There will also be no feedback even if the notification gets lost due to other errors.
// 2.) The function has only one return type defining an error type - in this case the call is handled as a request which does not expect any
//     results except for the state that the call was processed successfully or not.
//     It is not possible to define a function with only one return type which is not an error type.
// 3.) The function has exactly two return types, one defining any possible exported type and one defining an error type in exactly this order.
//     The first type will be the expected return type.
// Functions expecting more than two return types are not supported.
//
// Functions in the proxy definitions should be tagged with the rpcmethod tag and rpcparams tag.
// - The ```rpcmethod``` tag represents the name of the method defined by the client.
// - The ```rpcparams``` tag represents the name of each of the parameters defined by the client. The parameters should be seperated by a comma.
//   These parameters are in the same order as the parameter types in the function definition.
//   (Currently, parameters containing a comma in their name are not supported as there is no need for it at the moment)
//
// When an interface is created (using finalize), it will immediately try to setup a connection to the service and listens to it.
package rpc

import (
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/log"
	"returntypes-langserver/common/messages"
)

type _interface struct {
	communicator Communicator
	controller   Controller
	connection   Connection
	proxyFacade  interface{}
}

type InterfaceBuilder struct {
	i                 _interface
	err               errors.Error
	messager          messages.Messager
	onConnectionError func(Recoverer)
	onRecoverFailed   func(Recoverer)
}

type Interface interface {
	ProxyFacade() interface{}
	Connection() Connection
	Controller() Controller
}

type Controller interface {
	RegisterMethods(MethodRegister)
}

type Connection interface {
	Connect() errors.Error
	IsConnected() bool
	Close() errors.Error
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	IsRecoverable() bool
}

type ConnectionError struct {
	message     string
	wrappedErr  error
	recoverable bool
}

func (err *ConnectionError) Unwrap() error {
	return err.wrappedErr
}

func (err *ConnectionError) Error() string {
	return err.message
}

func (err *ConnectionError) IsRecoverable() bool {
	return err.recoverable
}

// Creates a connection error
func NewConnectionError(wrappedErr error, msg string, recoverable bool) *ConnectionError {
	return &ConnectionError{
		wrappedErr:  wrappedErr,
		message:     msg,
		recoverable: recoverable,
	}
}

// Creates an interface on the given connection with the given message protocol.
func CreateInterfaceOnConnection(connection Connection, messager messages.Messager) *InterfaceBuilder {
	builder := InterfaceBuilder{
		i: _interface{
			connection: connection,
		},
		messager: messager,
	}
	return &builder
}

// Adds a proxy facade (/client stub) to the interface representing the outgoing interface.
func (builder *InterfaceBuilder) WithProxyFacade(proxyFacadePtrUnwrapped interface{}) *InterfaceBuilder {
	if builder.err != nil || proxyFacadePtrUnwrapped == nil {
		return builder
	}

	if proxyFacadePtr, err := MakeProxyFacade(proxyFacadePtrUnwrapped, &builder.i); err != nil {
		builder.err = err
		return builder
	} else {
		builder.i.proxyFacade = proxyFacadePtr
		return builder
	}
}

// Adds a server to the interface representing the incoming interface.
func (builder *InterfaceBuilder) WithController(controller Controller) *InterfaceBuilder {
	if builder.err != nil {
		return builder
	}

	builder.i.controller = controller
	return builder
}

// The function will be called if a unrecoverable connection error occurs (e.g. an invalid address format)
// This means, the connection recover process should not be started until the connection configurations are changed
func (builder *InterfaceBuilder) OnConnectionError(fn func(Recoverer)) *InterfaceBuilder {
	builder.onConnectionError = fn
	return builder
}

// The function will be called if the connection recover process failed and may be restarted using recoverer.Recover()
func (builder *InterfaceBuilder) OnRecoverFailed(fn func(Recoverer)) *InterfaceBuilder {
	builder.onRecoverFailed = fn
	return builder
}

// Finalizes the interface and starts to communicate with the service.
func (builder *InterfaceBuilder) Finalize() (Interface, errors.Error) {
	if builder.err != nil {
		return nil, builder.err
	} else if builder.i.connection == nil {
		return nil, errors.New("RPC Error", "Cannot create an interface without connection")
	} else if builder.messager == nil {
		return nil, errors.New("RPC Error", "Cannot create an interface without messager")
	} else {
		// creates communicator
		builder.i.communicator = New(builder.i.connection, builder.messager, builder.onConnectionError, builder.onRecoverFailed)
	}

	builder.i.connection.Connect()
	if builder.i.controller != nil {
		builder.i.controller.RegisterMethods(builder.i.communicator)
	}
	go func() {
		if err := builder.i.communicator.Listen(); err != nil {
			log.Error(err)
		}
	}()
	return &builder.i, nil
}

func (i *_interface) ProxyFacade() interface{} {
	return i.proxyFacade
}

func (i *_interface) Connection() Connection {
	return i.connection
}

func (i *_interface) Controller() Controller {
	return i.controller
}
