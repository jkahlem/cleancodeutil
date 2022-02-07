// Generator for services
//
// Generates a lot of boilerplate code for services and writes them to one file (which were before manually managed in multiple files)
// It does:
// - Generate a proxy facade with functions which validates the underlying proxy (based on the struct)
//   - Checks if each method of a proxy has the rpcmethod and rpcparams tags
//   - Checks if each method of a proxy has exactly the same amount of parameters in the rpcparams tag as in the rpcmethod tag
//   - For each method: Generates a method on the proxy with exactly the same name/arguments which validates the existence of the proxy and it's functions
//   - Copies also the comments for the generated method
//   * Need to find the proxy struct (File "Proxy.go" -> Struct Name "Proxy")
//   * Need to go through the fields of the proxies including their comments
// - Generate functions for initializing a callable proxy (facade)
//   - basically a "proxy()" method which returns a singleton proxy facade (and initializes it) would be enough. But see the interface stuff for it.
// - Generate functions for the singleton interface / singleton service stuff. This might also result into a rework of the interface stuff.
//   - singleton service: Export only methods which are actually exported from the given class
//     * Need to search for methods belonging to a specific class. Also: How to find the target class?
// - Maybe also generate service Interfaces and the service mock stubs (if a mock does not exist at the moment, so optional.)
//   - Is this really needed, except for the predictor stuff? (How to use the mock??)
// - Controller generation? (Especially, or at least, the register methods stuff. But how to define the actual method names? see language server)
//
// There are no further validations if for example names of generated functions are duplicated or something

// TODO: Remove/Rework the above comment. (As they are just some notes for creating the generator.)
package main

import (
	"fmt"
	"log"
	"returntypes-langserver/common/generator"
	"strings"
	"text/template"

	"github.com/fatih/structtag"
)

type TemplateAttributes struct {
	TypeName string
}

func main() {
	ctx, err := generator.ParseFile(generator.CurrentFile())
	if err != nil {
		log.Fatal(err)
	}
	structs := ctx.ParseStructs()
	fmt.Println(structs)

	// write type name templates to output file
	/*if outputFile, err := os.Create(path.Join(path.Dir(targetFile), "Marshaller.go")); err != nil {
		log.Fatal(err)
	} else {
		fmt.Fprint(outputFile, HeaderNote)
		fmt.Fprintf(outputFile, "package %s\n", srcFileNode.Name.Name)
		fmt.Fprint(outputFile, Imports)
		for _, typeName := range typeNames {
			if tmpl, err := template.New("boilerplate").Parse(marshallerTemplate); err != nil {
				log.Fatal(err)
			} else if err := tmpl.Execute(outputFile, TemplateAttributes{TypeName: typeName}); err != nil {
				log.Fatal(err)
			}
		}
	}*/
	if proxyStruct, exists := findProxyStruct(structs); exists {
		fmt.Println(buildProxy(proxyStruct))
	} else {
		fmt.Println("No proxy found")
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
	"reflect"
	"returntypes-langserver/common/log"
)`
