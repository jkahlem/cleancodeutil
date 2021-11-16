package jsonrpc

type MessageBase struct {
	JsonRPC string `json:"jsonrpc" mapstructure:"jsonrpc"`
}

type Request struct {
	MessageBase `mapstructure:",squash"`
	Id          interface{} `json:"id" mapstructure:"id"`
	Method      string      `json:"method" mapstructure:"method"`
	Params      interface{} `json:"params" mapstructure:"params"`
}

type Response struct {
	MessageBase `mapstructure:",squash"`
	Id          interface{}    `json:"id" mapstructure:"id"`
	Result      interface{}    `json:"result,omitempty" mapstructure:"result"`
	Error       *ResponseError `json:"error,omitempty" mapstructure:"error"`
}

type Notification struct {
	MessageBase `mapstructure:",squash"`
	Method      string      `json:"method" mapstructure:"method"`
	Params      interface{} `json:"params" mapstructure:"params"`
}

func NewRequest(method string) Request {
	request := Request{
		Method: method,
	}
	request.JsonRPC = "2.0"
	return request
}

func NewResponse(id interface{}) Response {
	response := Response{
		Id: id,
	}
	response.JsonRPC = "2.0"
	return response
}

func NewNotification(method string) Notification {
	notification := Notification{
		Method: method,
	}
	notification.JsonRPC = "2.0"
	return notification
}
