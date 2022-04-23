package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"returntypes-langserver/common/code/generator"
	"strings"
	"text/template"
)

type Schema struct {
	Path         string
	QuotedSchema string
}

type TemplateAttributes struct {
	Schemas []Schema
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal(fmt.Errorf("Need the root path of the schema files as argument, but got %d arguments", len(os.Args)-1))
	}
	ctx, err := generator.ParseFile(generator.CurrentFile())
	if err != nil {
		log.Fatal(err)
	}
	consts := ctx.ParseConsts()
	rootPath := filepath.Clean(filepath.Join(filepath.Dir(generator.CurrentFile()), os.Args[1]))

	schemas := make([]Schema, 0, len(consts))

	for _, cst := range consts {
		if !strings.HasSuffix(cst.Name, "SchemaPath") {
			continue
		}
		relPath := strings.Trim(cst.Value.Code(), `"`)
		filePath := filepath.FromSlash(path.Join(rootPath, relPath))

		schemaContents, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		schemas = append(schemas, Schema{
			Path:         relPath,
			QuotedSchema: fmt.Sprintf("`%s`", string(schemaContents)),
		})
	}

	if outputFile, err := os.Create(path.Join(path.Dir(generator.CurrentFile()), "schemaContents.go")); err != nil {
		log.Fatal(err)
	} else {
		fmt.Fprint(outputFile, generator.HeaderNote)
		fmt.Fprintf(outputFile, "package %s\n", ctx.Package())
		attr := TemplateAttributes{
			Schemas: schemas,
		}

		if tmpl, err := template.New("boilerplate").Parse(SchemaTemplate); err != nil {
			log.Fatal(err)
		} else if err := tmpl.Execute(outputFile, attr); err != nil {
			log.Fatal(err)
		}
	}
}

const SchemaTemplate = `
var SchemaMap map[string]string

func initSchemaMap() {
	SchemaMap = make(map[string]string)

	{{- range $i, $e := .Schemas}}
	SchemaMap["{{ .Path }}"] = {{ .QuotedSchema }}
	{{- end}}
}

func getSchemaMap() map[string]string {
	if SchemaMap == nil {
		initSchemaMap()
	}
	return SchemaMap
}
`
