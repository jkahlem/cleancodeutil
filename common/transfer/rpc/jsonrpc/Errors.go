package jsonrpc

type ErrorCode int

const (
	// JSON-RPC
	ParseError     ErrorCode = -32700
	InvalidRequest ErrorCode = -32600
	MethodNotFound ErrorCode = -32601
	InvalidParams  ErrorCode = -32602
	InternalError  ErrorCode = -32603
)

type ResponseError struct {
	Code    ErrorCode   `json:"code" mapelement:"code"`
	Message string      `json:"message" mapstructure:"message"`
	Data    interface{} `json:"data,omitempty" mapstructure:"data"`
}

func (err ResponseError) Error() string {
	return err.Message
}

func NewResponseError(code ErrorCode, message string) ResponseError {
	return ResponseError{
		Code:    code,
		Message: message,
	}
}
