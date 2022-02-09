// This package implements structures and functionalities for the JSON-RPC 2.0 Specification (https://www.jsonrpc.org/specification)
package jsonrpc

import (
	"encoding/json"

	"returntypes-langserver/common/debug/errors"

	"github.com/mitchellh/mapstructure"
)

const JSONRPCErrorTitle = "JSONRPC Error"
const MediaType = "application/json"

// Unmarshals a string of an JSONRPC object in JSON and maps it to the right message structure.
// The types of some message attributes (like data, result, params) need to be mapped seperately.
// However, unmarshalling batch requests is not supported (as it is not required in the context of this program).
func Unmarshal(raw string) (interface{}, errors.Error) {
	obj := make(map[string]interface{})
	err := json.Unmarshal([]byte(raw), &obj)
	if err != nil {
		return nil, errors.Wrap(err, JSONRPCErrorTitle, "Could not unmarshal JSON")
	}
	return distinguishMessage(obj)
}

// Distinguishes the message type by looking at its structure and returns it as the searched structure.
func distinguishMessage(message map[string]interface{}) (interface{}, errors.Error) {
	if message == nil {
		return nil, errors.New(JSONRPCErrorTitle, "Message is nil")
	}

	if _, ok := message["id"]; ok {
		if _, ok := message["method"]; ok {
			// requests requires id & method
			return mapRequest(message)
		} else {
			_, hasResult := message["result"]
			_, hasError := message["error"]
			if (hasResult && !hasError) || (!hasResult && hasError) {
				// responses requires id and either a result or an error
				return mapResponse(message)
			}
		}
	} else if _, ok := message["method"]; ok {
		// notifications requires method (no id field)
		return mapNotification(message)
	}

	return nil, errors.New(JSONRPCErrorTitle, "Unsupported message type")
}

func mapRequest(message map[string]interface{}) (interface{}, errors.Error) {
	request := Request{}
	if err := mapstructure.Decode(message, &request); err != nil {
		return nil, errors.Wrap(err, JSONRPCErrorTitle, "Could not map request")
	}
	return request, nil
}

func mapResponse(message map[string]interface{}) (interface{}, errors.Error) {
	response := Response{}
	if err := mapstructure.Decode(message, &response); err != nil {
		return nil, errors.Wrap(err, JSONRPCErrorTitle, "Could not map response")
	}
	return response, nil
}

func mapNotification(message map[string]interface{}) (interface{}, errors.Error) {
	notification := Notification{}
	if err := mapstructure.Decode(message, &notification); err != nil {
		return nil, errors.Wrap(err, JSONRPCErrorTitle, "Could not map notification")
	}
	return notification, nil
}
