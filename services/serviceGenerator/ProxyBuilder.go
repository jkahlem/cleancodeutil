package main

import (
	"fmt"
	"log"
	"returntypes-langserver/common/generator"
	"strings"

	"github.com/fatih/structtag"
)

const ProxyStructName = "Proxy"

func buildProxy(proxyStruct generator.Struct) string {
	outputCode := strings.Builder{}
	outputCode.WriteString(ProxyFacadeDef)
	for _, field := range proxyStruct.Fields {
		if fnType, ok := field.Type.FunctionType(); ok {
			if err := validateProxyFunction(fnType, field); err != nil {
				log.Fatal(err)
			}
			fnData := FunctionData{
				FunctionName:  field.Name,
				Documentation: field.Documentation,
				Parameters:    mapParametersToNameTypePairs(fnType.In),
				Result:        mapParametersToNameTypePairs(fnType.Out),
			}
			WriteTemplate(&outputCode, ProxyFacadeFunctionTemplate, fnData)
		}
	}
	outputCode.WriteString(ProxyFacadeValidateFnDef)
	return outputCode.String()
}

func findProxyStruct(structs []generator.Struct) (generator.Struct, bool) {
	for _, s := range structs {
		if s.Name == ProxyStructName {
			return s, true
		}
	}
	return generator.Struct{}, false
}

func validateProxyFunction(fnType generator.FunctionType, field generator.StructField) error {
	if tags, err := structtag.Parse(field.Tag); err != nil {
		return err
	} else if _, err := tags.Get("rpcmethod"); err != nil {
		return fmt.Errorf("The required `rpcmethod` tag was not found for function %s.", field.Name)
	} else if paramsTag, err := tags.Get("rpcparams"); err != nil {
		return fmt.Errorf("The required `rpcparams` tag was not found for function %s.", field.Name)
	} else if pars := strings.Split(paramsTag.Value(), ","); len(pars) != len(fnType.In) {
		return fmt.Errorf("Function %s defines %d parameters but the tag defines %d parameters.", field.Name, len(fnType.In), len(pars))
	}
	return nil
}

const ProxyFacadeDef = "type ProxyFacade struct {\n\tProxy Proxy `rpcproxy:\"true\"`\n}\n\n"

const ProxyFacadeFunctionTemplate = `{{asLineComments .Documentation}}
func (p *ProxyFacade) {{.FunctionName}}({{range $i, $e := .Parameters}}{{if $i}}, {{end}}{{.Name}} {{.Type}}{{end}}) ({{range $i, $e := .Result}}{{if $i}}, {{end}}{{.Name}} {{.Type}}{{end}}) {
	if err := p.validate(p.Proxy.{{.FunctionName}}); err != nil {
		return {{range $i, $e := .Result}}{{if $i}}, {{end}}{{if eq .Type "string"}}""{{else if eq .Type "errors.Error"}}err{{else}}nil{{end}}{{end}}
	}
	{{if ne 0 (len .Result)}}return {{end}}p.Proxy.{{.FunctionName}}({{range $i, $e := .Parameters}}{{if $i}}, {{end}}{{.Name}}{{if isVariadic .Type}}...{{end}}{{end}})
}

`

const ProxyFacadeValidateFnDef = `func (p *ProxyFacade) validate(fn interface{}) errors.Error {
	fnVal := reflect.ValueOf(fn)
	if !fnVal.IsValid() || fnVal.IsZero() {
		return errors.New("RPC Error", "Interface function does not exist")
	}
	return nil
}
`
