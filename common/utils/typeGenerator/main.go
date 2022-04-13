package main

import (
	"fmt"
	"log"
	"os"
	"returntypes-langserver/common/code/generator"
	"strings"
	"text/template"
)

type TemplateAttributes struct {
	OutputType string
	TargetType string
}

func main() {
	output := ""
	if len(os.Args) < 4 {
		fmt.Println("Unsupported arguments")
		return
	} else {
		kind, outputType, targetType := os.Args[1], os.Args[2], os.Args[3]
		if kind == "stack" {
			output += generateStack(outputType, targetType)
		}
	}

	if content, err := os.ReadFile(generator.CurrentFile()); err != nil {
		log.Fatal(err)
	} else {
		output = string(content) + "\n" + output
		if err := os.WriteFile(generator.CurrentFile(), []byte(output), os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
}

func generateStack(outputType, targetType string) string {
	strBuilder := strings.Builder{}

	if tmpl, err := template.New("boilerplate").Parse(StackDef); err != nil {
		log.Fatal(err)
	} else if err := tmpl.Execute(&strBuilder, TemplateAttributes{
		OutputType: outputType,
		TargetType: targetType,
	}); err != nil {
		log.Fatal(err)
	}

	return strBuilder.String()
}

const StackDef = `
type {{.OutputType}} []{{.TargetType}}

func New{{.OutputType}}() *{{.OutputType}} {
	s := make({{.OutputType}}, 0)
	return &s
}

func (s *{{.OutputType}}) Push(value {{.TargetType}}) {
	*s = append(*s, value)
}

func (s *{{.OutputType}}) Pop() ({{.TargetType}}, bool) {
	if s.IsEmpty() {
		var zero {{.TargetType}}
		return zero, false
	}
	elm := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return elm, true
}

func (s *{{.OutputType}}) Peek() ({{.TargetType}}, bool) {
	if s.IsEmpty() {
		var zero {{.TargetType}}
		return zero, false
	}
	return (*s)[len(*s)-1], true
}

func (s *{{.OutputType}}) IsEmpty() bool {
	return len(*s) == 0
}`
