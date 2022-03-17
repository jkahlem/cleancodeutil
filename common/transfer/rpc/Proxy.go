package rpc

import (
	"fmt"
	"reflect"
	"strings"

	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

const ProxyMethodTag = "rpcmethod"
const ProxyMethodParamsTag = "rpcparams"
const ProxyTag = "rpcproxy"

type MethodDefinition struct {
	Type      reflect.Type
	Name      string
	Params    []string
	Interface _interface
}

// Implements the functionality for the proxy inside the given proxy facade
func MakeProxyFacade(proxyFacadePtrUnwrapped interface{}, _interface *_interface) (proxyFacadePtr interface{}, err errors.Error) {
	if proxyFacade, err := getProxyFacade(proxyFacadePtrUnwrapped); err != nil {
		return nil, err
	} else if proxy, err := getProxy(proxyFacade); err != nil {
		return nil, err
	} else if err := makeProxy(proxy, _interface); err != nil {
		return nil, err
	} else {
		// save the proxy facade as a pointer to the struct
		return proxyFacade.Addr().Interface(), nil
	}
}

// Unwraps the proxy facade from the pointer
func getProxyFacade(proxyFacadePtrUnwrapped interface{}) (reflect.Value, errors.Error) {
	proxyFacadePtr := reflect.ValueOf(proxyFacadePtrUnwrapped)
	if err := checkProxyFacade(proxyFacadePtr); err != nil {
		return reflect.ValueOf(nil), err
	}
	proxyFacade := proxyFacadePtr.Elem()
	return proxyFacade, nil
}

// Returns the proxy inside the given proxy facade
func getProxy(proxyFacade reflect.Value) (reflect.Value, errors.Error) {
	proxy, err := getProxyFieldOfFacade(proxyFacade)
	if err != nil {
		return reflect.ValueOf(nil), err
	} else if err := checkProxy(proxy); err != nil {
		return reflect.ValueOf(nil), err
	}
	return proxy, nil
}

// Implements the proxy functionality
func makeProxy(proxy reflect.Value, _interface *_interface) errors.Error {
	for i := 0; i < proxy.NumField(); i++ {
		field := proxy.Type().Field(i)
		methodDef := MethodDefinition{
			Type:   field.Type,
			Name:   field.Tag.Get(ProxyMethodTag),
			Params: splitParams(field.Tag.Get(ProxyMethodParamsTag)),
		}
		if len(methodDef.Name) == 0 {
			return errors.New("RPC Error", fmt.Sprintf("Field '%s' of proxy has no method name.", field.Name))
		}
		if err := checkProxyMethod(methodDef); err != nil {
			return err
		}

		proxy.Field(i).Set(makeProxyMethod(methodDef, _interface))
	}
	return nil
}

// Splits the rpcparams by the comma
func splitParams(params string) []string {
	var paramsSplitted []string
	if len(params) == 0 {
		paramsSplitted = make([]string, 0)
	} else {
		paramsSplitted = strings.Split(params, ",")
	}
	return paramsSplitted
}

// Checks if the value is a valid proxy facade value
func checkProxyFacade(facade reflect.Value) errors.Error {
	if !facade.IsValid() {
		return errors.New("RPC Error", "Invalid proxy facade value")
	} else if !utils.IsPtrTo(facade, reflect.Struct) {
		return errors.New("RPC Error", "Expected a pointer to a struct for proxy facade")
	}
	return nil
}

// Returns the field of the proxy facade which has the "rpcproxy" tag
func getProxyFieldOfFacade(facade reflect.Value) (reflect.Value, errors.Error) {
	facadeType := facade.Type()
	for i := 0; i < facadeType.NumField(); i++ {
		field := facadeType.Field(i)
		if !facade.Field(i).CanSet() {
			continue
		} else if _, exist := field.Tag.Lookup(ProxyTag); exist {
			return facade.Field(i), nil
		}
	}
	return reflect.ValueOf(nil), errors.New("RPC Error", "No exported proxy field with the rpcproxy tag found in proxy facade")
}

// Checks if the value is a valid proxy
func checkProxy(proxy reflect.Value) errors.Error {
	if !proxy.IsValid() {
		return errors.New("RPC Error", "Invalid proxy value")
	} else if proxy.Kind() != reflect.Struct {
		return errors.New("RPC Error", "Expected a struct for proxy")
	}
	return nil
}

// Checks if fn is a valid function definition for the proxy / rpc
func checkProxyMethod(methodDef MethodDefinition) errors.Error {
	if methodDef.Type.Kind() != reflect.Func {
		return errors.New("RPC Error", fmt.Sprintf("Expected function type but got %s.", methodDef.Type.Name()))
	} else if methodDef.Type.NumIn() != len(methodDef.Params) {
		errmsg := fmt.Sprintf("Function expects %d parameters, but %d are defined", methodDef.Type.NumIn(), len(methodDef.Params))
		return errors.New("RPC Error", errmsg)
	} else if methodDef.Type.NumOut() > 0 && !utils.IsErrorType(lastOut(methodDef.Type)) {
		return errors.New("RPC Error", "A proxy method for requests should always return an error type")
	} else if methodDef.Type.NumOut() > 2 {
		return errors.New("RPC Error", "A proxy method should have a maximum of two return types.")
	}
	return nil
}

// Returns the last return type of the given function. Panics if fn is not of kind Func
func lastOut(fn reflect.Type) reflect.Type {
	return fn.Out(fn.NumOut() - 1)
}

// Creates a method for the proxy which executes the rpc request/notification sending when called
func makeProxyMethod(methodDef MethodDefinition, _interface *_interface) reflect.Value {
	fn := reflect.MakeFunc(methodDef.Type, func(args []reflect.Value) []reflect.Value {
		arguments := mapArgumentsToParamsMap(args, methodDef.Params)
		if isNotificationMethod(methodDef.Type) {
			if _interface.communicator != nil {
				_interface.communicator.Notify(methodDef.Name, arguments)
			}
			return nil
		} else {
			results, err := processRequest(methodDef.Name, arguments, _interface)
			return mapResultsToSlice(results, err, methodDef)
		}
	})
	return fn
}

// Creates a map of string -> argument value pairs using the rpcparams definition of the given method (for by-name parameter structure)
func mapArgumentsToParamsMap(reflectedArgs []reflect.Value, params []string) map[string]interface{} {
	arguments := make(map[string]interface{})
	for j, v := range reflectedArgs {
		arguments[params[j]] = v.Interface()
	}
	return arguments
}

// Returns true if the method is a notification method
func isNotificationMethod(fn reflect.Type) bool {
	return fn.NumOut() == 0
}

// makes a request for the given method
func processRequest(rpcMethodName string, arguments interface{}, _interface *_interface) (interface{}, errors.Error) {
	if _interface.communicator != nil {
		return _interface.communicator.Request(rpcMethodName, arguments)
	} else {
		return nil, errors.New("RPC Error", "No stream")
	}
}

// Maps the results of a request to a slice of reflect.Values.
// The slice may contain an error if defined.
// Note: Results must not be formatted by-name or by-position. The result value can be any kind of value, which is then the return value.
//       The only other case is, if the response contains an error (and therefore no result property). There is no by-name or by-position specified for result.
func mapResultsToSlice(results interface{}, err errors.Error, methodDef MethodDefinition) []reflect.Value {
	expectedReturnType := methodDef.Type.Out(0)
	returnValues := createZeroValueReturnValues(methodDef.Type)
	returnValue := reflect.Zero(expectedReturnType)

	resultsValue := reflect.ValueOf(results)
	if resultsValue.IsValid() {
		if v, err2 := utils.CastValueToTypeIfPossible(resultsValue, expectedReturnType); err2 != nil {
			err = errors.Wrap(err2, "RPC Error", fmt.Sprintf("Error at return parameter for method %s", methodDef.Name))
		} else {
			returnValue = v
		}
	}

	if utils.IsErrorType(methodDef.Type.Out(0)) {
		if err != nil {
			returnValues[0] = reflect.ValueOf(err)
		}
	} else {
		if returnValue.IsValid() {
			returnValues[0] = returnValue
		}
		if err != nil && methodDef.Type.NumOut() > 1 {
			returnValues[1] = reflect.ValueOf(err)
		}
	}
	return returnValues
}

// Creates a slice of the zero values of each expected type for the return values of the given method
func createZeroValueReturnValues(methodToImplement reflect.Type) []reflect.Value {
	returnValues := make([]reflect.Value, methodToImplement.NumOut())
	// set each value of the return values to its zero value
	for j := 0; j < len(returnValues); j++ {
		returnValues[j] = reflect.Zero(methodToImplement.Out(j))
	}
	return returnValues
}
