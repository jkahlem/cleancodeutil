package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"returntypes-langserver/common/code/generator"
	"text/template"
)

func NewMain() {
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
