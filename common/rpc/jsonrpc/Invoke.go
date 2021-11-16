package jsonrpc

import (
	"fmt"
	"reflect"

	"returntypes-langserver/common/utils"
)

// Calls the function fn with the given parameters casted to the expected type.
func Invoke(fn *Function, params interface{}) (result interface{}, err *ResponseError) {
	mappedPars, err := prepareParamsForFunctionCall(fn, params)
	if err != nil {
		return nil, err
	}

	out := fn.Fn.Call(mappedPars)

	return mapReturnValues(out)
}

// Maps the incoming parameters to a list of reflect values of each parameter with the expected type.
// Returns an error if the types do not match and are not castable to the expected type.
// This function implements the jsonrpc byName and byPosition params ordering, so if a slice is passed
// the types will be casted by position to the expected type and if a map is passed, the types will be casted
// by name. In the latter case, the order of the return types is determined using the order of Function.params.
func prepareParamsForFunctionCall(fn *Function, params interface{}) ([]reflect.Value, *ResponseError) {
	unwrapped := utils.UnwrapInterface(reflect.ValueOf(params))
	if unwrapped.Kind() == reflect.Slice {
		return mapParamsByPosition(fn, reflect.ValueOf(utils.WrapInSlice(params)))
	} else if unwrapped.Kind() == reflect.Map {
		return mapParamsByName(fn, unwrapped)
	} else if !unwrapped.IsValid() {
		if fn.Fn.Type().NumIn() != 0 {
			err := NewResponseError(InvalidParams, "Function call expects params, but got null.")
			return nil, &err
		}
		// unwrapped is not valid, if params is nil (null), so its just a function call without passing parameters.
		return nil, nil
	} else {
		err := NewResponseError(InvalidParams, "The params must be a structured value either as a slice or as an object")
		return nil, &err
	}
}

// Maps the parameters as reflect values by looking at the position of the incoming parameters.
func mapParamsByPosition(fn *Function, source reflect.Value) ([]reflect.Value, *ResponseError) {
	fnType := fn.Fn.Type()
	destination := createZeroParams(fn)
	if err := checkParamsLengthForFunctionCall(fn, source); err != nil {
		return nil, err
	}
	for i := 0; i < fnType.NumIn(); i++ {
		if value, err := utils.CastValueToTypeIfPossible(source.Index(i), fnType.In(i)); err != nil {
			err := NewResponseError(InvalidParams, fmt.Sprintf("Unexpected parameter type at index %d", i))
			return nil, &err
		} else {
			destination[i] = value
		}
	}
	return destination, nil
}

// Maps the parameters as reflect values by looking at the fields of the parameter.
func mapParamsByName(fn *Function, source reflect.Value) ([]reflect.Value, *ResponseError) {
	destination := createZeroParams(fn)
	keyValuePairs := utils.MakeKeyValuePairChannel(source)
	for pair := range keyValuePairs {
		if pair.Key.Kind() != reflect.String {
			continue
		} else if position, expectedType := getExpectedParameterTypeAndPosition(pair.Key.String(), fn.Params()); expectedType == nil {
			continue
		} else if value, err := utils.CastValueToTypeIfPossible(pair.Value, expectedType); err != nil {
			responseError := NewResponseError(InvalidParams, fmt.Sprintf("Unexpected parameter type for parameter %s", expectedType))
			return nil, &responseError
		} else {
			destination[position] = value
		}
	}
	return destination, nil
}

// Returns the position and type of the given parameter name in the given parameter slice.
func getExpectedParameterTypeAndPosition(paramName string, params []Parameter) (int, reflect.Type) {
	for i, parameter := range params {
		if parameter.Name == paramName {
			return i, parameter.Type
		}
	}
	return -1, nil
}

// Creates a slice of zero values of the expected parameter types.
func createZeroParams(fn *Function) []reflect.Value {
	fnType := fn.Fn.Type()
	params := make([]reflect.Value, fnType.NumIn())
	for i := range params {
		params[i] = reflect.Zero(fnType.In(i))
	}
	return params
}

// Checks if the length of the given parameters matches the length of the expected parameters.
func checkParamsLengthForFunctionCall(fn *Function, params reflect.Value) *ResponseError {
	fnType := fn.Fn.Type()
	if fnType.NumIn() != params.Len() {
		err := NewResponseError(InvalidParams, fmt.Sprintf("Expected %d parameters but got %d", fnType.NumIn(), params.Len()))
		return &err
	}
	return nil
}

// Maps the returned values to:
// - either a slice of return values ([]interface{}) when multiple values are returned (except for the last error)
// - or to the given return value if it's the only one (except for the last error)
// If the last return type of the function is an error type and it's value is not nil, then it will be used as the error resposne
// to the client.
func mapReturnValues(out []reflect.Value) (interface{}, *ResponseError) {
	if len(out) > 0 {
		if remained, err := extractError(out); err != nil {
			return nil, err
		} else if len(remained) > 0 {
			if len(remained) == 1 {
				return remained[0].Interface(), nil
			} else {
				return utils.MapToInterfaceSlice(remained), nil
			}
		}
	}
	return nil, nil
}

// Checks if the last value is an error type and extracts it from the slice.
// Returns a slice without this last error type if it exists. Returns also the extracted error.
func extractError(out []reflect.Value) ([]reflect.Value, *ResponseError) {
	if utils.IsErrorType(out[len(out)-1].Type()) {
		errorVal := out[len(out)-1]
		out = out[:len(out)-1]
		if !errorVal.IsNil() {
			if isResponseErrorType(errorVal.Type()) {
				return nil, asPtrToResponseError(errorVal)
			} else {
				err, _ := errorVal.Interface().(error)
				rpcerr := NewResponseError(InternalError, err.Error())
				return nil, &rpcerr
			}
		}
	}
	return out, nil
}

// Checks if the given error type is a response error.
func isResponseErrorType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	errorType := reflect.TypeOf(ResponseError{})
	return t.AssignableTo(errorType)
}

// Wraps the value as a pointer to a response error if it is not already a pointer.
func asPtrToResponseError(v reflect.Value) *ResponseError {
	if v.Kind() != reflect.Ptr {
		v = v.Addr()
	}
	asPtr, _ := v.Interface().(*ResponseError)
	return asPtr
}
