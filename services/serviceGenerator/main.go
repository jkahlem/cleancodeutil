// Generator for services
//
// Generates boilerplate code for services:
// - Proxy functions/facades are build based on a provided class "Proxy". The calls on the proxy functions can be done by calling them on "remote()"
//   e.g. remote().DoSomething()   (so the client provides a method "DoSomething")
// - Generates a "service facade" where the service may be called from outside of the package. Also manages the service creation on any call (the service is a singleton).
//   The struct which will be used as service needs an "annotation", so a comment
//     // @ServiceGenerator:ServiceDefinition
//   at the closing '}' bracket of the struct definition. There might be only one service definition per package.
//   If a method of the service should not be generated/exposed to outside the package, prepend the following
//   comment annotation to the function:
//     // @ServiceGenerator:IgnoreMethod
//   If the service should be created based on an interface (helpful if mocks exist), then this interface should be annotated as follows:
//     // @ServiceGenerator:ServiceInterfaceDefinition
//   (The service definition is still required).
//   A mock might also be defined using
//     // @ServiceGenerator:ServiceMockDefinition
//   Each service should have a file with a "serviceConfiguration()" function, returning a rpc.ServiceConfiguration, which will be used for creating the service.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"returntypes-langserver/common/code/generator"
)

type ServiceFacadeTemplateAttributes struct {
	ExposedServiceType string
	ActualServiceType  string
	MockServiceType    string
	Methods            []FunctionData
}

func main() {
	parsedData := ParsePackage()

	proxyBody := ""
	if proxyStruct, exists := findProxyStruct(parsedData.Structs); exists {
		proxyBody = buildProxy(proxyStruct)
	} else {
		fmt.Println("No proxy found")
	}
	serviceFacadeBody := buildServiceFacade(parsedData)

	if outputFile, err := os.Create(filepath.Join(filepath.Dir(generator.CurrentFile()), "generated.go")); err != nil {
		log.Fatal(err)
	} else {
		fmt.Fprint(outputFile, generator.HeaderNote)
		fmt.Fprintf(outputFile, "package %s\n", parsedData.Package)
		fmt.Fprint(outputFile, Imports, proxyBody, serviceFacadeBody, InterfaceDef)
	}
}

type ParsedData struct {
	Structs    []generator.Struct
	Interfaces []generator.Interface
	Functions  []generator.FunctionDeclaration
	Package    string
}

func ParsePackage() ParsedData {
	ctx, err := generator.ParsePackage(filepath.Dir(generator.CurrentFile()), "generated.go")
	if err != nil {
		log.Fatal(err)
	}
	return ParsedData{
		Structs:    ctx.ParseStructs(),
		Interfaces: ctx.ParseInterfaces(),
		Functions:  ctx.ParseFunctionDeclarations(),
		Package:    ctx.Package(),
	}
}

const Imports = `
import (
	"io"
	"reflect"
	"sync"

	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/transfer/messages"
	"returntypes-langserver/common/transfer/rpc"
	"returntypes-langserver/common/transfer/rpc/jsonrpc"
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
