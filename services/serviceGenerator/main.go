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
	"unicode"

	"github.com/fatih/structtag"
)

type ServiceFacadeTemplateAttributes struct {
	ExportedServiceType string
	ActualServiceType   string
	MockServiceType     string
	Methods             []FunctionData
}

func main() {
	ctx, err := generator.ParsePackage(filepath.Dir(generator.CurrentFile()), "generated.go")
	if err != nil {
		log.Fatal(err)
	}
	proxyBody := ""
	structs := ctx.ParseStructs()
	interfaces := ctx.ParseInterfaces()
	functions := ctx.ParseFunctionDeclarations()
	if proxyStruct, exists := findProxyStruct(structs); exists {
		proxyBody = buildProxy(proxyStruct)
	} else {
		fmt.Println("No proxy found")
	}
	serviceFacadeBody := buildServiceFacade(structs, interfaces, functions)

	if outputFile, err := os.Create(filepath.Join(filepath.Dir(generator.CurrentFile()), "generated.go")); err != nil {
		log.Fatal(err)
	} else {
		fmt.Fprint(outputFile, generator.HeaderNote)
		fmt.Fprintf(outputFile, "package %s\n", ctx.Package())
		fmt.Fprint(outputFile, Imports)
		fmt.Fprint(outputFile, proxyBody)
		fmt.Fprint(outputFile, serviceFacadeBody)
		fmt.Fprint(outputFile, InterfaceDef)
	}
}

func buildServiceFacade(structs []generator.Struct, interfaces []generator.Interface, functions []generator.FunctionDeclaration) string {
	outputCode := strings.Builder{}
	if serviceStruct, exists := findServiceStruct(structs); exists {
		serviceAttributes := ServiceFacadeTemplateAttributes{
			ExportedServiceType: "*" + serviceStruct.Name,
			ActualServiceType:   serviceStruct.Name,
		}
		if mock, exists := findServiceMock(structs); exists {
			serviceAttributes.MockServiceType = mock.Name
		}

		if serviceInterface, exists := findServiceInterface(interfaces); exists {
			fmt.Println("Service interface found")
			// For service interfaces, export the interface as type
			serviceAttributes.ExportedServiceType = serviceInterface.Name
			for _, method := range serviceInterface.Methods {
				fnType, ok := method.Type.FunctionType()
				if !ok {
					continue
				}
				serviceAttributes.Methods = append(serviceAttributes.Methods, FunctionData{
					FunctionName:  method.Name,
					Documentation: commentEachLine(method.Documentation),
					Parameters:    mapParametersToNameTypePairs(fnType.In),
					Result:        mapParametersToNameTypePairs(fnType.Out),
				})
			}
		} else {
			fmt.Println("No interface defined - build service facade by function declarations")
			// build by function declarations
			for _, function := range functions {
				if function.ReceiverType != serviceAttributes.ExportedServiceType || !isExportedName(function.Name) {
					continue
				}
				fnType, ok := function.Type.FunctionType()
				if !ok {
					continue
				}
				serviceAttributes.Methods = append(serviceAttributes.Methods, FunctionData{
					FunctionName:  function.Name,
					Documentation: commentEachLine(function.Documentation),
					Parameters:    mapParametersToNameTypePairs(fnType.In),
					Result:        mapParametersToNameTypePairs(fnType.Out),
				})
			}
		}
		if tmpl, err := template.New("boilerplate").Parse(ServiceFacadeTemplate); err != nil {
			log.Fatal(err)
		} else if err := tmpl.Execute(&outputCode, serviceAttributes); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(fmt.Errorf("No service found"))
	}
	return outputCode.String()
}

// Checks if the name/identifier is exported (accessible from outside the package)
func isExportedName(name string) bool {
	firstChar := name[0]
	return firstChar != '_' && unicode.IsLetter(rune(firstChar)) && !unicode.IsLower(rune(firstChar))
}

func findProxyStruct(structs []generator.Struct) (generator.Struct, bool) {
	for _, s := range structs {
		if s.Name == "Proxy" {
			return s, true
		}
	}
	return generator.Struct{}, false
}

func findServiceStruct(structs []generator.Struct) (service generator.Struct, found bool) {
	for _, s := range structs {
		if strings.Contains(s.LineComment, "@ServiceGenerator:ServiceDefinition") {
			if found {
				log.Fatal(fmt.Errorf("Multiple service definitions found. Service definitions should be unique per package."))
				return
			}
			service = s
			found = true
		}
	}
	return
}

func findServiceMock(structs []generator.Struct) (generator.Struct, bool) {
	for _, s := range structs {
		if strings.Contains(s.LineComment, "@ServiceGenerator:ServiceMockDefinition") {
			return s, true
		}
	}
	return generator.Struct{}, false
}
func findServiceInterface(interfaces []generator.Interface) (generator.Interface, bool) {
	for _, i := range interfaces {
		fmt.Println("->", i.LineComment, i.Documentation)
		if strings.Contains(i.LineComment, "@ServiceGenerator:ServiceInterfaceDefinition") {
			return i, true
		}
	}
	return generator.Interface{}, false
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

const ServiceFacadeTemplate = `
var singleton {{.ExportedServiceType}}
var singletonMutex sync.Mutex

func getSingleton() {{.ExportedServiceType}} {
	singletonMutex.Lock()
	defer singletonMutex.Unlock()

	if singleton == nil {
		singleton = createSingleton()
	}
	return singleton
}

func createSingleton() {{.ExportedServiceType}} {
	{{if .MockServiceType}}
	if serviceConfiguration().UseMock {
		log.Info("Setup {{.ExportedServiceType}} service using mock...\n")
		return &{{.MockServiceType}}{}
	}
	{{end}}
	return &{{.ActualServiceType}}{}
}

{{range .Methods}}
{{.Documentation}}
func {{.FunctionName}}({{range $i, $e := .Parameters}}{{if $i}}, {{end}}{{.Name}} {{.Type}}{{end}}) ({{range $i, $e := .Result}}{{if $i}}, {{end}}{{.Name}} {{.Type}}{{end}}) {
	return getSingleton().{{.FunctionName}}({{range $i, $e := .Parameters}}{{if $i}}, {{end}}{{.Name}}{{end}})
}
{{end}}
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
