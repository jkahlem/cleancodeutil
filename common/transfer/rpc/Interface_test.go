package rpc

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/transfer/messages"
	"returntypes-langserver/common/transfer/rpc/jsonrpc"
	"returntypes-langserver/common/utils"

	"github.com/stretchr/testify/assert"
)

const FnDuplicateString = "duplicateString"
const FnCheckStringIsAllUppercase = "checkStringIsAllUppercase"
const FnSubstringFromFirstA = "substringFromFirstA"
const FnDoNothing = "doNothing"
const FnGetStringOfNestedStruct = "getStringOfNestedStruct"
const FnCheckIfNestedValueAndStringAreEqual = "checkIfNestedValueAndStringAreEqual"

func TestInterfaceCreation(t *testing.T) {
	// given
	controller := TestControllerImplementation{}
	connection := NewTestConnection()
	messager := CreateMessager(connection)

	// when
	_, err := CreateInterfaceOnConnection(connection, messager).WithProxyFacade(&TestProxyFacade{}).WithController(&controller).Finalize()

	// then
	assert.NoError(t, err)
}

func TestRequestToExternalMethod(t *testing.T) {
	// given
	connection, ifc := CreateSimpleTestInterface(t)
	responseMessage := CreateResponse(1, "responseString")
	connection.setReceivedContent(responseMessage)
	facade, ok := ifc.ProxyFacade().(*TestProxyFacade)

	// when
	returnedString, err := facade.Proxy.MethodOfExternalService("test")

	// then
	expectedRequest := "Content-Length: 74\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"method":"externalMethod",` +
		`"params":{"str":"test"}` +
		`}`
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, expectedRequest, connection.sentContent())
	assert.Equal(t, "responseString", returnedString)
}

func TestMultipleRequestsToExternalMethod(t *testing.T) {
	// given
	connection, ifc := CreateSimpleTestInterface(t)
	responseMessage := ""
	for i := 0; i < 10; i++ {
		// Create a string of unsorted responses. (The order of responses should not matter)
		requestId := i
		if i%2 == 1 {
			requestId = 10 - i
		}
		responseMessage += CreateResponse(requestId+1, "responseString")
	}
	connection.setReceivedContent(responseMessage)
	facade, _ := ifc.ProxyFacade().(*TestProxyFacade)

	// when
	for i := 0; i < 10; i++ {
		facade.Proxy.MethodOfExternalService("test")
	}

	// then
	expectedRequest := "Content-Length: %d\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":%d,` +
		`"method":"externalMethod",` +
		`"params":{"str":"test"}` +
		`}`
	expectedRequests := ""
	for i := 0; i < 10; i++ {
		expectedRequests += fmt.Sprintf(expectedRequest, 73+len(fmt.Sprintf("%d", i+1)), i+1)
	}
	assert.Equal(t, expectedRequests, connection.sentContent())
}

func TestRequestToExternalMethodWithErrorOnly(t *testing.T) {
	// given
	connection, ifc := CreateSimpleTestInterface(t)
	responseMessage := CreateResponse(1, "")
	facade, ok := ifc.ProxyFacade().(*TestProxyFacade)
	connection.setReceivedContent(responseMessage)

	// when
	err := facade.Proxy.MethodOfExternalServiceWithErrorOnly("test")

	// then
	expectedRequest := "Content-Length: 83\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"method":"externalMethodWithError",` +
		`"params":{"str":"test"}` +
		`}`
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, expectedRequest, connection.sentContent())
}

func TestRequestToExternalMethodReturningError(t *testing.T) {
	// given
	connection, ifc := CreateSimpleTestInterface(t)
	responseMessage := CreateErrorResponse(1, int(jsonrpc.InternalError), "test error message")
	facade, ok := ifc.ProxyFacade().(*TestProxyFacade)
	connection.setReceivedContent(responseMessage)

	// when
	err := facade.Proxy.MethodOfExternalServiceWithErrorOnly("test")

	// then
	assert.Error(t, err)
	assert.True(t, ok)
}

func TestRequestToExternalMethodReturningSlices(t *testing.T) {
	// given
	connection, ifc := CreateSimpleTestInterface(t)
	responseMessage := CreateResponseFromJson(`{"jsonrpc":"2.0","id":1,"result":["foo", "bar"]}`)
	facade, _ := ifc.ProxyFacade().(*TestProxyFacade)
	connection.setReceivedContent(responseMessage)

	// when
	slice, err := facade.Proxy.MethodReturningSlice("test")

	// then
	expectedRequest := "Content-Length: 89\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"method":"externalMethodReturningSlices",` +
		`"params":{"str":"test"}` +
		`}`
	assert.NoError(t, err)
	assert.Equal(t, expectedRequest, connection.sentContent())
	assert.NotNil(t, slice)
	assert.Len(t, slice, 2)
	assert.Equal(t, "foo", slice[0])
	assert.Equal(t, "bar", slice[1])
}

func TestRequestToExternalMethodReturningStruct(t *testing.T) {
	// given
	responseMessage := CreateResponseFromJson(`{"jsonrpc":"2.0","id":1,"result":{"text": "bar"}}`)
	connection, ifc := CreateSimpleTestInterface(t)
	facade, _ := ifc.ProxyFacade().(*TestProxyFacade)
	connection.setReceivedContent(responseMessage)

	// when
	testStruct, err := facade.Proxy.MethodReturningStruct("test")

	// then
	expectedRequest := "Content-Length: 89\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"method":"externalMethodReturningStruct",` +
		`"params":{"str":"test"}` +
		`}`
	assert.NoError(t, err)
	assert.Equal(t, expectedRequest, connection.sentContent())
	assert.Equal(t, "bar", testStruct.Text)
}

func TestRequestToExternalMethodReturningInterfaceSlice(t *testing.T) {
	// given
	responseMessage := CreateResponseFromJson(`{"jsonrpc":"2.0","id":1,"result":[null, 17, "test"]}`)
	connection, ifc := CreateSimpleTestInterface(t)
	facade, _ := ifc.ProxyFacade().(*TestProxyFacade)
	connection.setReceivedContent(responseMessage)

	// when
	slice, err := facade.Proxy.MethodReturningInterfaceSlice("test")

	// then
	expectedRequest := "Content-Length: 97\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"method":"externalMethodReturningInterfaceSlice",` +
		`"params":{"str":"test"}` +
		`}`
	assert.NoError(t, err)
	assert.Equal(t, expectedRequest, connection.sentContent())
	assert.Len(t, slice, 3)
	assert.Nil(t, slice[0])
	number, _ := slice[1].(float64)
	assert.Equal(t, 17, int(number))
	assert.Equal(t, "test", slice[2])
}

func TestNotificationToExternalMethod(t *testing.T) {
	// given
	connection, ifc := CreateSimpleTestInterface(t)
	facade, ok := ifc.ProxyFacade().(*TestProxyFacade)

	// when
	facade.Proxy.MethodForNotifications("test")

	// then
	expectedRequest := "Content-Length: 79\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"method":"externalNotificationMethod",` +
		`"params":{"str":"test"}` +
		`}`
	assert.True(t, ok)
	assert.Equal(t, expectedRequest, connection.sentContent())
}

func TestRequestToController(t *testing.T) {
	// given
	connection, _ := CreateSimpleTestInterface(t)
	requestMessage := CreateRequestWithStringParams(1, FnDuplicateString, "parameter")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 54\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"result":"parameterparameter"` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

func TestRequestToControllerReturningNoError(t *testing.T) {
	// given
	connection, _ := CreateSimpleTestInterface(t)
	requestMessage := CreateRequestWithStringParams(1, FnCheckStringIsAllUppercase, "UPPERCASE")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 38\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"result":null` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

func TestRequestToControllerReturningError(t *testing.T) {
	// given
	connection, _ := CreateSimpleTestInterface(t)
	requestMessage := CreateRequestWithStringParams(1, FnCheckStringIsAllUppercase, "lowercase")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 113\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"error":{"code":-32603,"message":"Test error: The string is not completely uppercased!"}` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

func TestRequestToControllerWithMultipleReturnValues(t *testing.T) {
	// given
	connection, _ := CreateSimpleTestInterface(t)
	requestMessage := CreateRequestWithStringParams(1, FnSubstringFromFirstA, "this is a sentence")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 50\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"result":["a sentence",8]` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

func TestRequestToControllerReturningNothing(t *testing.T) {
	// given
	connection, _ := CreateSimpleTestInterface(t)
	requestMessage := CreateRequest(1, FnDoNothing, "")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 38\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"result":null` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

func TestRequestToControllerWithNestedStruct(t *testing.T) {
	// given
	connection, _ := CreateSimpleTestInterface(t)
	requestMessage := CreateRequestWithNestedStruct(1, FnGetStringOfNestedStruct, "string value")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 48\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"result":"string value"` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

func TestRequestToControllerWithMixedParams(t *testing.T) {
	// given
	connection, _ := CreateSimpleTestInterface(t)
	requestMessage := CreateRequest(1, FnCheckIfNestedValueAndStringAreEqual, `{"str":"text","structure":{"nested":{"value":"text"}}}`)

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 38\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"result":true` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

func TestRequestOnUnstableConnection(t *testing.T) {
	// given
	configuration.LoadConfigFromJsonString(`{"connection":{"timeout":5,"reconnectionAttempts":5}}`)
	connection, _ := CreateUnstableTestInterface(t, 3)
	requestMessage := CreateRequest(1, FnDoNothing, "")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 38\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"result":null` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

func TestRequestOnUnrecoverableConnection(t *testing.T) {
	// given
	configuration.LoadConfigFromJsonString(`{"connection":{"timeout":5,"reconnectionAttempts":5}}`)

	called := make(chan bool, 1)
	onRecoverFailed := func(Recoverer) {
		called <- true
	}
	connection, _ := CreateUnstableTestInterfaceWithEvents(t, 10, nil, onRecoverFailed)

	requestMessage := CreateRequest(1, FnDoNothing, "")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	assert.True(t, <-called)
}

func TestTryRecoverAfterUnrecoverable(t *testing.T) {
	// given
	configuration.LoadConfigFromJsonString(`{"connection":{"timeout":5,"reconnectionAttempts":5}}`)

	onRecoverFailed := func(recoverer Recoverer) {
		recoverer.Recover()
	}
	connection, _ := CreateUnstableTestInterfaceWithEvents(t, 10, nil, onRecoverFailed)

	requestMessage := CreateRequest(1, FnDoNothing, "")

	// when
	connection.setReceivedContent(requestMessage)

	// then
	expectedResponse := "Content-Length: 38\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{` +
		`"jsonrpc":"2.0",` +
		`"id":1,` +
		`"result":null` +
		`}`
	assert.Equal(t, expectedResponse, connection.sentContent())
}

// Test relevant structures

type TestNestedStruct struct {
	Nested *NestedStringValueWrapper `json:"nested,omitempty"`
}

type NestedStringValueWrapper struct {
	Value string `json:"value"`
}

type TestProxy struct {
	MethodOfExternalService              func(string) (string, error)        `rpcmethod:"externalMethod" rpcparams:"str"`
	MethodOfExternalServiceWithErrorOnly func(string) error                  `rpcmethod:"externalMethodWithError" rpcparams:"str"`
	MethodForNotifications               func(string)                        `rpcmethod:"externalNotificationMethod" rpcparams:"str"`
	MethodReturningSlice                 func(string) ([]string, error)      `rpcmethod:"externalMethodReturningSlices" rpcparams:"str"`
	MethodReturningStruct                func(string) (TestStruct, error)    `rpcmethod:"externalMethodReturningStruct" rpcparams:"str"`
	MethodReturningInterfaceSlice        func(string) ([]interface{}, error) `rpcmethod:"externalMethodReturningInterfaceSlice" rpcparams:"str"`
}

type TestProxyFacade struct {
	Proxy TestProxy `rpcproxy:"true"`
}

type TestStruct struct {
	Text string `json:"text"`
}

type TestControllerImplementation struct{}

func (t *TestControllerImplementation) RegisterMethods(rpc MethodRegister) {
	rpc.RegisterMethod(FnDuplicateString, "str", t.DuplicateString)
	rpc.RegisterMethod(FnCheckStringIsAllUppercase, "str", t.CheckStringIsAllUppercase)
	rpc.RegisterMethod(FnSubstringFromFirstA, "str", t.SubstringFromFirstA)
	rpc.RegisterMethod(FnDoNothing, "", t.DoNothing)
	rpc.RegisterMethod(FnGetStringOfNestedStruct, "structure", t.GetStringOfNestedStruct)
	rpc.RegisterMethod(FnCheckIfNestedValueAndStringAreEqual, "structure,str", t.CheckIfNestedValueAndStringAreEqual)
}

func (t *TestControllerImplementation) DuplicateString(str string) (string, error) {
	return fmt.Sprintf("%s%s", str, str), nil
}

func (t *TestControllerImplementation) CheckStringIsAllUppercase(str string) error {
	if str != strings.ToUpper(str) {
		return errors.New("Test error", "The string is not completely uppercased!")
	}
	return nil
}

func (t *TestControllerImplementation) SubstringFromFirstA(str string) (string, int, error) {
	index := strings.Index(strings.ToUpper(str), "A")
	if index == -1 {
		return "", index, errors.New("Test error", "String has no 'a'!")
	}
	return str[index:], index, nil
}

func (t *TestControllerImplementation) DoNothing() {}

func (t *TestControllerImplementation) GetStringOfNestedStruct(structure TestNestedStruct) (string, error) {
	if structure.Nested == nil {
		return "", nil
	}
	return structure.Nested.Value, nil
}

func (t *TestControllerImplementation) CheckIfNestedValueAndStringAreEqual(structure TestNestedStruct, str string) (bool, error) {
	if structure.Nested == nil {
		return false, nil
	}
	return structure.Nested.Value == str, nil
}

// Helper structures

type TestConnection struct {
	in                    []byte
	out                   []byte
	onWrite               chan bool
	mutex                 sync.Mutex
	unrecoverableAttempts int
	recoverable           bool
	establishmentState    utils.Revision
}

func (t *TestConnection) getChan() chan bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.onWrite == nil {
		t.onWrite = make(chan bool, 10)
	}
	return t.onWrite
}

func (t *TestConnection) Read(toRead []byte) (int, error) {
	if t.unrecoverableAttempts > 0 {
		return 0, errors.Wrap(io.ErrClosedPipe, "Connection error", "Could not read")
	}
	t.establishmentState.WaitUntilOutdated()
	if t.in != nil && len(t.in) > 0 {
		n := copy(toRead, t.in)
		t.in = t.in[n:]
		return n, nil
	}
	return 0, io.EOF
}

func (t *TestConnection) Write(toWrite []byte) (int, error) {
	if t.unrecoverableAttempts > 0 {
		return 0, errors.Wrap(io.ErrClosedPipe, "Connection error", "Could not write")
	}
	t.out = append(t.out, toWrite...)
	t.getChan() <- true
	return len(toWrite), nil
}

func (t *TestConnection) sentContent() string {
	<-t.getChan()
	return string(t.out)
}

func (t *TestConnection) setReceivedContent(content string) {
	t.in = []byte(content)
	t.establishmentState.SetOutdated()
}

func (t *TestConnection) Connect() errors.Error {
	if t.unrecoverableAttempts > 0 {
		t.unrecoverableAttempts--
		return errors.Wrap(NewConnectionError(io.ErrClosedPipe, "Could not connect", true), "Connection error", "Could not connect")
	}
	return nil
}

func (t *TestConnection) IsConnected() bool {
	return t.unrecoverableAttempts == 0
}

func (t *TestConnection) Close() errors.Error {
	return nil
}

func (t *TestConnection) IsRecoverable() bool {
	return t.recoverable
}

func (t *TestConnection) setRecoverable(state bool) {
	t.recoverable = state
}

func (t *TestConnection) establishConnectionAfterAttempts(attempts int) {
	t.unrecoverableAttempts = attempts
}

func NewTestConnection() *TestConnection {
	return &TestConnection{
		establishmentState: utils.NewRevision(),
	}
}

// Helper functions

func CreateSimpleTestInterface(t *testing.T) (*TestConnection, Interface) {
	connection := NewTestConnection()
	return CreateTestInterface(t, connection, nil, nil)
}

func CreateUnstableTestInterface(t *testing.T, failingAttemptsCount int) (*TestConnection, Interface) {
	connection := NewTestConnection()
	connection.setRecoverable(true)
	connection.establishConnectionAfterAttempts(failingAttemptsCount)
	return CreateTestInterface(t, connection, nil, nil)
}

func CreateUnstableTestInterfaceWithEvents(t *testing.T, failingAttemptsCount int, onConnectionError, onRecoverFailed func(Recoverer)) (*TestConnection, Interface) {
	connection := NewTestConnection()
	connection.setRecoverable(true)
	connection.establishConnectionAfterAttempts(failingAttemptsCount)
	return CreateTestInterface(t, connection, onConnectionError, func(r Recoverer) {
		onRecoverFailed(r)
	})
}

func CreateTestInterface(t *testing.T, connection *TestConnection, onConnectionError, onRecoverFailed func(Recoverer)) (*TestConnection, Interface) {
	controller := TestControllerImplementation{}
	messager := CreateMessager(connection)
	ifc, err := CreateInterfaceOnConnection(connection, messager).WithProxyFacade(&TestProxyFacade{}).WithController(&controller).
		OnConnectionError(onConnectionError).OnRecoverFailed(onRecoverFailed).Finalize()
	if err != nil {
		panic(err)
	}
	return connection, ifc
}

func CreateResponse(id int, result string) string {
	var content string
	if len(result) > 0 {
		content = fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":"%s"}`, id, result)
	} else {
		content = fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":null}`, id)
	}
	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: %s\r\n\r\n", len(content), jsonrpc.MediaType)
	return header + content
}

func CreateErrorResponse(id, errorCode int, errMsg string) string {
	content := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"error":{"code":%d,"message":"%s"}}`, id, errorCode, errMsg)
	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: %s\r\n\r\n", len(content), jsonrpc.MediaType)
	return header + content
}

func CreateResponseFromJson(jsonContent string) string {
	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: %s\r\n\r\n", len(jsonContent), jsonrpc.MediaType)
	return header + jsonContent
}

func CreateRequestWithStringParams(id int, method, par string) string {
	return CreateRequest(id, method, fmt.Sprintf(`{"str":"%s"}`, par))
}

func CreateRequestWithNestedStruct(id int, method, value string) string {
	return CreateRequest(id, method, fmt.Sprintf(`{"structure":{"nested":{"value":"%s"}}}`, value))
}

func CreateRequest(id int, method, params string) string {
	var content string
	if len(params) > 0 {
		content = fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","id":%d,"params":%s}`, method, id, params)
	} else {
		content = fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","id":%d}`, method, id)
	}
	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: %s\r\n\r\n", len(content), jsonrpc.MediaType)
	return header + content
}

func CreateMessager(conn Connection) messages.Messager {
	messager := messages.NewReadWriter(conn)
	messager.SetWritingMimeType(jsonrpc.MediaType)
	messager.AcceptMediaType(jsonrpc.MediaType)
	return messager
}
