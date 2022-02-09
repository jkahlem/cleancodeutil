package main

import (
	"fmt"
	"log"
	"returntypes-langserver/common/code/generator"
	"strings"
)

const (
	ServiceDefinitionAnnotation          = "@ServiceGenerator:ServiceDefinition"
	ServiceMockDefinitionAnnotation      = "@ServiceGenerator:ServiceMockDefinition"
	ServiceInterfaceDefinitionAnnotation = "@ServiceGenerator:ServiceInterfaceDefinition"
	IgnoreMethodAnnotation               = "@ServiceGenerator:IgnoreMethod"
)

func buildServiceFacade(parsedData ParsedData) string {
	outputCode := strings.Builder{}
	serviceStruct, exists := findServiceStruct(parsedData.Structs)
	if !exists {
		log.Fatal(fmt.Errorf("No service found"))
		return ""
	}
	serviceAttributes := ServiceFacadeTemplateAttributes{
		ExposedServiceType: "*" + serviceStruct.Name,
		ActualServiceType:  serviceStruct.Name,
	}
	if mock, exists := findServiceMock(parsedData.Structs); exists {
		serviceAttributes.MockServiceType = mock.Name
	}

	if serviceInterface, exists := findServiceInterface(parsedData.Interfaces); exists {
		fmt.Println("Service interface found")
		// For service interfaces, expose the interface as type
		serviceAttributes.ExposedServiceType = serviceInterface.Name
		serviceAttributes.Methods = getServiceMethodsByInterfaceDeclaration(serviceInterface)
	} else {
		fmt.Println("No interface defined - build service facade by function declarations")
		// build by function declarations
		serviceAttributes.Methods = getServiceMethodsByFunctionDeclarations(parsedData.Functions, serviceAttributes.ExposedServiceType)
	}

	WriteTemplate(&outputCode, ServiceFacadeTemplate, serviceAttributes)
	return outputCode.String()
}

func getServiceMethodsByInterfaceDeclaration(ifc generator.Interface) []FunctionData {
	methods := make([]FunctionData, 0, len(ifc.Methods))
	for _, method := range ifc.Methods {
		fnType, ok := method.Type.FunctionType()
		if !ok {
			continue
		}
		methods = append(methods, FunctionData{
			FunctionName:  method.Name,
			Documentation: method.Documentation,
			Parameters:    mapParametersToNameTypePairs(fnType.In),
			Result:        mapParametersToNameTypePairs(fnType.Out),
		})
	}
	return methods
}

func getServiceMethodsByFunctionDeclarations(functions []generator.FunctionDeclaration, expectedReceiver string) []FunctionData {
	methods := make([]FunctionData, 0, 1)
	for _, function := range functions {
		if function.ReceiverType != expectedReceiver || strings.Index(function.Documentation, IgnoreMethodAnnotation) != -1 {
			continue
		}
		fnType, ok := function.Type.FunctionType()
		if !ok {
			continue
		}
		methods = append(methods, FunctionData{
			FunctionName:  function.Name,
			Documentation: function.Documentation,
			Parameters:    mapParametersToNameTypePairs(fnType.In),
			Result:        mapParametersToNameTypePairs(fnType.Out),
		})
	}
	return methods
}

func findServiceStruct(structs []generator.Struct) (service generator.Struct, found bool) {
	for _, s := range structs {
		if strings.Contains(s.LineComment, ServiceDefinitionAnnotation) {
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
		if strings.Contains(s.LineComment, ServiceMockDefinitionAnnotation) {
			return s, true
		}
	}
	return generator.Struct{}, false
}

func findServiceInterface(interfaces []generator.Interface) (generator.Interface, bool) {
	for _, i := range interfaces {
		fmt.Println("->", i.LineComment, i.Documentation)
		if strings.Contains(i.LineComment, ServiceInterfaceDefinitionAnnotation) {
			return i, true
		}
	}
	return generator.Interface{}, false
}

const ServiceFacadeTemplate = `
var singleton {{.ExposedServiceType}}
var singletonMutex sync.Mutex

func getSingleton() {{.ExposedServiceType}} {
	singletonMutex.Lock()
	defer singletonMutex.Unlock()

	if singleton == nil {
		singleton = createSingleton()
	}
	return singleton
}

func createSingleton() {{.ExposedServiceType}} {
	{{if .MockServiceType}}
	if serviceConfiguration().UseMock {
		log.Info("Setup {{.ExposedServiceType}} service using mock...\n")
		return &{{.MockServiceType}}{}
	}
	{{end}}
	return &{{.ActualServiceType}}{}
}

{{range .Methods}}
{{asLineComments .Documentation}}
func {{.FunctionName}}({{range $i, $e := .Parameters}}{{if $i}}, {{end}}{{.Name}} {{.Type}}{{end}}) ({{range $i, $e := .Result}}{{if $i}}, {{end}}{{.Name}} {{.Type}}{{end}}) {
	{{if ne 0 (len .Result)}}return {{end}}getSingleton().{{.FunctionName}}({{range $i, $e := .Parameters}}{{if $i}}, {{end}}{{.Name}}{{if isVariadic .Type}}...{{end}}{{end}})
}
{{end}}
`
