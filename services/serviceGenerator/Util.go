package main

import (
	"io"
	"log"
	"returntypes-langserver/common/code/generator"
	"strings"
	"text/template"
)

func WriteTemplate(writer io.Writer, tpl string, data interface{}) {
	if tmpl, err := template.New("boilerplate").Funcs(templateFunctions()).Parse(tpl); err != nil {
		log.Fatal(err)
	} else if err := tmpl.Execute(writer, data); err != nil {
		log.Fatal(err)
	}
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

func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"isVariadic": func(value string) bool {
			return strings.HasPrefix(value, "...")
		},
		"asLineComments": asLineComments,
	}
}

func asLineComments(value string) string {
	lines := strings.Split(value, "\n")
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
