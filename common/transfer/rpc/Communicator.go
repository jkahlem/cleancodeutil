package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/messages"
	"returntypes-langserver/common/transfer/rpc/jsonrpc"
	"returntypes-langserver/common/utils"
)

const StreamRPCErrorTitle = "RPC Error"

// Object that will be marshalled to a "null" json value when marshalled to json.
// This allows to set fields with the json "omitempty" attribute explicitly to null
// without omitting the field. (Which is for example required for jsonrpc responses
// which must either contain a result field or an error field while the result field
// may be set to null in case of a successful request)
type JsonNULL struct{}

func (null JsonNULL) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

type communicator struct {
	readWriter        messages.Messager
	connection        Connection
	mappings          map[string]*jsonrpc.Function
	idCounter         int
	mutex             sync.Mutex
	recoverMutex      sync.Mutex
	recoverRev        utils.Revision
	recoverRevMutex   sync.Mutex
	responseQueue     responseQueue
	onConnectionError func(Recoverer)
	onRecoverFailed   func(Recoverer)
	isListening       bool
	isUnrecovered     bool
}

// The Communicator handles the communication with an external service
type Communicator interface {
	// registers a method for rpc calls which can be invoked by the client
	RegisterMethod(method string, params string, fn interface{})
	// calls a method using rpc
	Request(string, interface{}) (interface{}, errors.Error)
	// calls a method using rpc
	Notify(string, interface{})
	// waits for calls of the client
	Listen() errors.Error
}

// The method register is used to hold information about the methods which may be invoked
type MethodRegister interface {
	// Registers a method for invocation.
	// The parameter method is the expected method name in rpc (which may differ from the real method name)
	// The parameter params defines the names of each parameter in rpc (seperated by a comma) in the expected order for the function.
	// The parameter fn is the function called for the specified method name
	RegisterMethod(method string, params string, fn interface{})
}

// Tries to recover the connection
type Recoverer interface {
	// Try to recover the connection
	Recover()
}

// Creates a new communicator
func New(connection Connection, readWriter messages.Messager, onConnectionError, onRecoverFailed func(Recoverer)) Communicator {
	mappings := make(map[string]*jsonrpc.Function)
	return &communicator{
		connection:        connection,
		readWriter:        readWriter,
		mappings:          mappings,
		onConnectionError: onConnectionError,
		onRecoverFailed:   onRecoverFailed,
	}
}

// Registers the specified method for requests from the other side of the connection
// If the method specifies return types, they will be sent to the connenction as responses to the request
// If the method specifies an error return type as the last return type, the error will be sent to the connections as an error response (except it's value is nil)
func (s *communicator) RegisterMethod(methodName string, params string, fn interface{}) {
	value := reflect.ValueOf(fn)
	if s.isValidFunc(value) {
		s.mappings[methodName] = &jsonrpc.Function{Fn: value}
		s.mappings[methodName].SetParams(params)
	}
}

// Returns true if fn is a valid function value
func (s *communicator) isValidFunc(fn reflect.Value) bool {
	if !fn.IsValid() || fn.Kind() != reflect.Func {
		return false
	}
	return true
}

// Sends a request message to the remote service. Returns the result of the response or a response error.
func (s *communicator) Request(methodName string, params interface{}) (interface{}, errors.Error) {
	if s.readWriter == nil {
		return nil, errors.New("RPC Error", "No connection set")
	}

	requestId := s.nextId()
	request := jsonrpc.NewRequest(methodName)
	request.Params = params
	request.Id = requestId
	s.log("Send request with method %s and id %v", request.Method, request.Id)
	if err := s.writeJsonMessageToMessager(request); err != nil {
		return nil, err
	}
	return s.awaitResponse(requestId)
}

// Returns the next id for requests from this side
func (s *communicator) nextId() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.idCounter++
	return s.idCounter
}

// Sends a notification message to the remote service.
func (s *communicator) Notify(methodName string, params interface{}) {
	s.log("Send notification for method %s", methodName)
	request := jsonrpc.NewNotification(methodName)
	request.Params = params
	s.writeJsonMessageToMessager(request)
}

// Listens for incoming messages on the connection
func (s *communicator) Listen() errors.Error {
	s.mutex.Lock()
	if s.isListening {
		// no multiple listen routines...
		return nil
	}
	s.isListening = true
	s.mutex.Unlock()
	defer func() {
		s.isListening = false
		s.responseQueue.Close()
		s.log("Stop listening.")
	}()

	for {
		msg, err := s.awaitMessage()
		if err != nil {
			log.Error(err)
			return err
		} else if msg != nil {
			go s.handleMessage(msg)
		}
	}
}

// Handles the message according to the message type
func (s *communicator) handleMessage(msg interface{}) {
	if request, ok := msg.(jsonrpc.Request); ok {
		s.log("Request with method %s received", request.Method)
		result, err := s.invoke(request.Method, request.Params)
		if err != nil {
			s.respond(request.Id, nil, err)
		} else {
			s.respond(request.Id, result, nil)
		}
	} else if notification, ok := msg.(jsonrpc.Notification); ok {
		s.log("Notification with method %s received", notification.Method)
		s.invoke(notification.Method, notification.Params)
	} else if response, ok := msg.(jsonrpc.Response); ok {
		s.log("Response to id %v received", response.Id)
		s.responseQueue.Append(response)
	} else {
		log.Error(errors.New(StreamRPCErrorTitle, "Unexpected rpc message type"))
	}
}

// Sends a response message to the remote service
func (s *communicator) respond(id, result interface{}, err *jsonrpc.ResponseError) errors.Error {
	response := jsonrpc.NewResponse(id)
	response.Error = err
	response.Result = result
	if result == nil && err == nil {
		response.Result = JsonNULL{}
	} else if err != nil {
		s.log("Repond with error %d -> %s", err.Code, err.Message)
	}
	return s.writeJsonMessageToMessager(response)
}

// Waits for a response with the given id.
func (s *communicator) awaitResponse(id int) (interface{}, errors.Error) {
	if response, err := s.responseQueue.PickResponseWithId(id); err != nil {
		return nil, err
	} else if response.Error != nil {
		return nil, errors.Wrap(response.Error, "RPC Error", "Received response containing an error")
	} else {
		return response.Result, nil
	}
}

// Waits for a message from the remote service and returns it
func (s *communicator) awaitMessage() (interface{}, errors.Error) {
	if s.readWriter == nil {
		return nil, errors.New("RPC Error", "No read writer set")
	}
	msg, err := s.readMessageFromMessager()
	if err != nil {
		return nil, err
	}
	rpcMsg, err := jsonrpc.Unmarshal(msg)
	if err != nil {
		responseError := jsonrpc.NewResponseError(jsonrpc.InvalidRequest, "Unmarshalling json object to request failed.")
		return nil, s.respond(JsonNULL{}, nil, &responseError)
	}
	return rpcMsg, nil
}

// Invokes the given methods using the given params
func (s *communicator) invoke(method string, params interface{}) (result interface{}, err *jsonrpc.ResponseError) {
	fn, found := s.mappings[method]
	if !found {
		err := jsonrpc.NewResponseError(jsonrpc.MethodNotFound, "No method with name '"+method+"' was found")
		return nil, &err
	}
	return jsonrpc.Invoke(fn, params)
}

// Writes the json message to the messager. Tries to recover on connection problems if possible.
func (s *communicator) writeJsonMessageToMessager(obj interface{}) errors.Error {
	if err, needsRecover := s.writeJsonMessageToMessager2(obj); err != nil {
		log.Error(err)
		if needsRecover {
			if err := s.recover(); err != nil {
				return err
			} else {
				return s.writeJsonMessageToMessager(obj)
			}
		} else {
			return err
		}
	}
	return nil
}

// Writes the json message to the messager. Tries to recover on connection problems if possible.
func (s *communicator) writeJsonMessageToMessager2(obj interface{}) (errors.Error, bool) {
	if jsonObj, err := json.Marshal(obj); err != nil {
		return errors.Wrap(err, StreamRPCErrorTitle, "Could not write message"), false
	} else if err := s.readWriter.WriteMessage(jsonObj); err != nil {
		if !s.connection.IsConnected() || errors.Is(err, io.ErrClosedPipe) {
			return err, true
		}
		return err, false
	} else {
		log.Print(log.Messager, ">> Write Message:\n%s\n", jsonObj)
	}
	return nil, false
}

// Reads message from messager. Tries to recover on connection problems if possible.
func (s *communicator) readMessageFromMessager() (string, errors.Error) {
	if msg, err, needRecover := s.readMessageFromMessager2(); err != nil {
		if needRecover {
			if err := s.recover(); err != nil {
				return "", err
			} else {
				return s.readMessageFromMessager()
			}
		} else {
			return msg, err
		}
	} else {
		return msg, nil
	}
}

func (s *communicator) readMessageFromMessager2() (string, errors.Error, bool) {
	if msg, err := s.readWriter.ReadMessage(); err != nil {
		if !s.connection.IsConnected() || errors.Is(err, io.ErrClosedPipe) {
			return "", err, true
		}
		return "", err, false
	} else {
		log.Print(log.Messager, "<< Read Message:\n%s\n", msg)
		return msg, nil, false
	}
}

func (s *communicator) waitForRecover() {
	s.recoverRevMutex.Lock()
	currentRev := s.recoverRev
	s.recoverRevMutex.Unlock()
	if currentRev != nil {
		currentRev.WaitUntilOutdated()
	}
}

// Tries to recover the connection in the configured amount of attempts
func (s *communicator) recover() errors.Error {
	s.recoverMutex.Lock()
	defer s.recoverMutex.Unlock()

	if s.connection.IsRecoverable() {
		if s.isUnrecovered {
			return errors.New("Error", "Could not recover connection")
		} else if !s.connection.IsConnected() {
			s.recoverRevMutex.Lock()
			s.recoverRev = utils.NewRevision()
			s.recoverRevMutex.Unlock()

			attempts := configuration.ConnectionReconnectionAttempts()
			for i := 0; i < attempts; i++ {
				retryTime := time.Now().Add(configuration.ConnectionTimeout())
				s.readWriter.Reset()
				if err := s.connection.Connect(); err == nil {
					return nil
				} else {
					var connError *ConnectionError
					if ok := errors.As(err, &connError); ok && !connError.IsRecoverable() {
						s.responseQueue.Close()
						s.isUnrecovered = true
						if s.onConnectionError != nil {
							go s.onConnectionError(s)
						}
						s.recoverRev.SetOutdated()
						return err
					}
					if i+1 < attempts {
						time.Sleep(time.Until(retryTime))
					} else {
						s.responseQueue.Close()
						s.isUnrecovered = true
						if s.onRecoverFailed != nil {
							go s.onRecoverFailed(s)
						}
						s.recoverRev.SetOutdated()
						return errors.New("RPC Error", "Connection can not be recovered")
					}
				}
			}
		} else {
			return nil
		}
	}

	s.responseQueue.Close()
	return errors.New("RPC Error", "Connection can not be recovered")
}

// Try to recover the connection if possible.
func (s *communicator) Recover() {
	if err := s.resetRecoverState(); err != nil {
		return
	} else if err := s.recover(); err == nil && !s.isListening {
		go s.Listen()
	}
}

// Resets the recover state.
func (s *communicator) resetRecoverState() errors.Error {
	s.recoverMutex.Lock()
	defer s.recoverMutex.Unlock()
	if !s.isUnrecovered {
		return errors.New("Error", "Connection is already recovered")
	}
	if s.recoverRev != nil {
		s.recoverRev.SetOutdated()
		log.Info("Recovered!!\n")
	}
	s.responseQueue.Reopen()
	s.isUnrecovered = false
	return nil
}

// Logs on the communicator layer
func (s *communicator) log(format string, args ...interface{}) {
	log.Print(log.Communicator, fmt.Sprintf("[COMMUNICATOR] %s\n", format), args...)
}
