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
}

const Imports = `
import (
	"reflect"
	"returntypes-langserver/common/log"
)`

var marshallerTemplate = `
func (t {{.TypeName}}) ToRecordTEST() []string {
	if record, err := marshal(reflect.ValueOf(t)); err != nil {
		log.Error(err)
		log.ReportProblem("An error occured while marshalling data")
		return nil
	} else {
		return record
	}
}

func Unmarshal{{.TypeName}}TEST(records [][]string) []{{.TypeName}} {
	typ := reflect.TypeOf({{.TypeName}}{})
	result := make([]{{.TypeName}}, 0, len(records))
	for _, record := range records {
		if unmarshalled, err := unmarshal(record,  typ); err != nil {
			log.Error(err)
			log.ReportProblem("An error occured while unmarshalling data")
		} else if c, ok := (unmarshalled.Interface()).({{.TypeName}}); ok {
			result = append(result, c)
		}
	}
	return result
}
`
