// Generator for services
//
// Generates a lot of boilerplate code for services and writes them to one file (which were before manually managed in multiple files).
// TODO: ServiceFacade generation by "annotations", e.g.
//   type predictor struct {} // Service
//   type Predictor interface { // Service interface
//
//   }
// Structs with line comment == "Service" define the service to use (which should be singleton)
// If an interface with the "Service interface" comment exist, use this one as:
// - return type in the getServiceSingleton() method
// - reference for the exported service methods. (Documentation should also be defined in the interface)
// otherwise do it with the struct.
// CHECK if there is not exactly one struct with service annotation, same for interface. Mocks should be different cases.
// Maybe prefix these generator annotations. As they are really not the best way to do it. something like @ServiceGenerator:ServiceDefinition or something.

// TODO: Remove/Rework the above comment. (As they are just some notes for creating the generator.)
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"returntypes-langserver/common/generator"
	"strings"
	"text/template"

	"github.com/fatih/structtag"
)

func main() {
	ctx, err := generator.ParsePackage(filepath.Dir(generator.CurrentFile()))
	if err != nil {
		log.Fatal(err)
	}
	proxyBody := ""
	structs := ctx.ParseStructs()
	if proxyStruct, exists := findProxyStruct(structs); exists {
		proxyBody = buildProxy(proxyStruct)
	} else {
		fmt.Println("No proxy found")
	}

	if outputFile, err := os.Create(filepath.Join(filepath.Dir(generator.CurrentFile()), "generated.go")); err != nil {
		log.Fatal(err)
	} else {
		fmt.Fprint(outputFile, generator.HeaderNote)
		fmt.Fprintf(outputFile, "package %s\n", ctx.Package())
		fmt.Fprint(outputFile, Imports, proxyBody, InterfaceDef)
	}
}

func findProxyStruct(structs []generator.Struct) (generator.Struct, bool) {
	for _, s := range structs {
		if s.Name == "Proxy" {
			return s, true
		}
	}
	return generator.Struct{}, false
}

func buildProxy(proxyStruct generator.Struct) string {
	outputCode := strings.Builder{}
	outputCode.WriteString(ProxyFacadeDef)
	for _, field := range proxyStruct.Fields {
		if fnType, ok := field.Type.FunctionType(); ok {
			if err := validateFunction(fnType, field); err != nil {
				log.Fatal(err)
			}
			fnData := FunctionData{
				FunctionName:  field.Name,
				Documentation: commentEachLine(field.Documentation),
				Parameters:    mapParametersToNameTypePairs(fnType.In),
				Result:        mapParametersToNameTypePairs(fnType.Out),
			}
			if tmpl, err := template.New("boilerplate").Parse(ProxyFacadeFunctionTemplate); err != nil {
				log.Fatal(err)
			} else if err := tmpl.Execute(&outputCode, fnData); err != nil {
				log.Fatal(err)
			}
		}
	}
	outputCode.WriteString(ProxyFacadeValidateFnDef)
	return outputCode.String()
}

func validateFunction(fnType generator.FunctionType, field generator.StructField) error {
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

func mapParametersToNameTypePairs(parameters []generator.Parameter) []NameTypePair {
	result := make([]NameTypePair, 0, len(parameters))
	for _, par := range parameters {
		result = append(result, mapParameterToNameTypePair(par))
	}
	return result
}

func mapParameterToNameTypePair(par generator.Parameter) NameTypePair {
	return NameTypePair{
		Name: par.Name,
		Type: par.Type.Code(),
	}
}

func commentEachLine(documentation string) string {
	lines := strings.Split(documentation, "\n")
	if len(lines) == 0 {
		return ""
	} else {
		lines = lines[0 : len(lines)-1]
	}
	for i := range lines {
		lines[i] = "// " + lines[i]
	}
	return strings.Join(lines, "\n")
}

type FunctionData struct {
	Documentation string
	FunctionName  string
	Parameters    []NameTypePair
	Result        []NameTypePair
}

type NameTypePair struct {
	Name string
	Type string
}

const ProxyFacadeDef = "type ProxyFacade struct {\n\tProxy Proxy `rpcproxy:\"true\"`\n}\n\n"

const ProxyFacadeFunctionTemplate = `{{.Documentation}}
func (p *ProxyFacade) {{.FunctionName}}({{range $i, $e := .Parameters}}{{if $i}}, {{end}}{{.Name}} {{.Type}}{{end}}) ({{range $i, $e := .Result}}{{if $i}}, {{end}}{{.Name}} {{.Type}}{{end}}) {
	if err := p.validate(p.Proxy.{{.FunctionName}}); err != nil {
		return nil, err
	}
	return p.Proxy.{{.FunctionName}}({{range $i, $e := .Parameters}}{{if $i}}, {{end}}{{.Name}}{{end}})
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

const Imports = `
import (
	"io"
	"reflect"
	"sync"

	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/log"
	"returntypes-langserver/common/messages"
	"returntypes-langserver/common/rpc"
	"returntypes-langserver/common/rpc/jsonrpc"
)

`

const InterfaceDef = `
var interfaceSingleton rpc.Interface
var interfaceMutex sync.Mutex

// Returns a proxy which can be used to communicate with the client.
func remote() *ProxyFacade {
	if ifc := getInterface(); ifc != nil && ifc.ProxyFacade() != nil {
		if facade, ok := ifc.ProxyFacade().(*ProxyFacade); ok {
			return facade
		}
	}
	return &ProxyFacade{}
}

// Returns the service connection
func serviceConnection() io.ReadWriter {
	if getInterface() != nil {
		return getInterface().Connection()
	}
	return nil
}

// Returns the used service interface
func getInterface() rpc.Interface {
	interfaceMutex.Lock()
	defer interfaceMutex.Unlock()

	if interfaceSingleton == nil {
		serviceConfig := serviceConfiguration()
		if ifc, err := rpc.BuildInterfaceFromServiceConfiguration(serviceConfig, &ProxyFacade{}); err != nil {
			if serviceConfig.OnInterfaceCreationError != nil {
				serviceConfig.OnInterfaceCreationError(err)
			}
		} else {
			interfaceSingleton = ifc
		}
	}
	return interfaceSingleton
}`
