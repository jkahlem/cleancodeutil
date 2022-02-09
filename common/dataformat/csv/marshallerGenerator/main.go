package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"returntypes-langserver/common/code/generator"
	"text/template"
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

	if outputFile, err := os.Create(path.Join(path.Dir(generator.CurrentFile()), "marshaller.go")); err != nil {
		log.Fatal(err)
	} else {
		fmt.Fprint(outputFile, generator.HeaderNote)
		fmt.Fprintf(outputFile, "package %s\n", ctx.Package())
		fmt.Fprint(outputFile, Imports)
		for _, structType := range structs {
			if tmpl, err := template.New("boilerplate").Parse(marshallerTemplate); err != nil {
				log.Fatal(err)
			} else if err := tmpl.Execute(outputFile, TemplateAttributes{TypeName: structType.Name}); err != nil {
				log.Fatal(err)
			}
		}
	}
}

const Imports = `
import (
	"reflect"
	"returntypes-langserver/common/debug/log"
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
