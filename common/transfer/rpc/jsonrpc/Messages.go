package jsonrpc

type MessageBase struct {
	JsonRPC string `json:"jsonrpc"`
}

type Request struct {
	MessageBase `json:",squash"`
	Id          interface{} `json:"id"`
	Method      string      `json:"method"`
	Params      interface{} `json:"params"`
}

type Response struct {
	MessageBase `json:",squash"`
	Id          interface{}    `json:"id"`
	Result      interface{}    `json:"result,omitempty"`
	Error       *ResponseError `json:"error,omitempty"`
}

type Notification struct {
	MessageBase `json:",squash"`
	Method      string      `json:"method"`
	Params      interface{} `json:"params"`
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
