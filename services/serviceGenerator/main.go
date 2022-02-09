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
