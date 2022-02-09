package jsonrpc

import (
	"reflect"
	"strings"
)

type Function struct {
	Fn     reflect.Value
	params []Parameter
}

type Parameter struct {
	Name string
	Type reflect.Type
}

// Sets the parameters of a function using the rpcparams format and the parameter types of the function.
func (f *Function) SetParams(params string) {
	if len(params) == 0 {
		f.params = make([]Parameter, 0)
		return
	}

	splitted := strings.Split(params, ",")
	pars := make([]Parameter, len(splitted))
	fnType := f.Fn.Type()
	for i := 0; i < len(pars) && i < fnType.NumIn(); i++ {
		pars[i] = Parameter{
			Name: splitted[i],
			Type: fnType.In(i),
		}
	}
	f.params = pars
}

func (f *Function) Params() []Parameter {
	if f.params == nil {
		f.params = make([]Parameter, 0)
	}
	return f.params
}
