package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"returntypes-langserver/common/code/generator"
	"strings"
	"text/template"
)

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
			buildUnmarshallerCode(structType, outputFile)
		}
	}
}

type UnmarshallerAttributes struct {
	TypeName string
	Fields   []FieldAttributes
}

type FieldAttributes struct {
	TypeName string
	Name     string
}

func buildUnmarshallerCode(s generator.Struct, outputFile io.Writer) {
	attr := UnmarshallerAttributes{
		TypeName: s.Name,
		Fields:   make([]FieldAttributes, len(s.Fields)),
	}
	for i, field := range s.Fields {
		attr.Fields[i] = FieldAttributes{
			Name:     field.Name,
			TypeName: field.Type.Code(),
		}
		if strings.HasPrefix(attr.Fields[i].TypeName, "[]") && attr.Fields[i].TypeName != "[]string" {
			log.Fatalf("Unsupported type: %s", attr.Fields[i].TypeName)
		}
	}
	funcs := template.FuncMap{
		"isIntegerType": isIntegerType,
		"typeError":     typeError,
	}
	if tmpl, err := template.New("boilerplate").Funcs(funcs).Parse(UnmarshalTemplate); err != nil {
		log.Fatal(err)
	} else if err := tmpl.Execute(outputFile, attr); err != nil {
		log.Fatal(err)
	}
}

func isIntegerType(str string) bool {
	switch str {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	}
	return false
}

func typeError(typeName string) (string, error) {
	return "", fmt.Errorf("Unsupported type: %s", typeName)
}

const UnmarshalTemplate = `
func Unmarshal{{.TypeName}}(record []string) ({{.TypeName}}, errors.Error) {
	result := {{.TypeName}}{}
	if len(record) < {{len .Fields}} {
		return result, errors.New(CsvErrorTitle, "Could not unmarshal to {{.TypeName}}: Expected {{len .Fields}} fields but got record with %d fields.", len(record))
	}

{{- range $i, $e := .Fields}}
	{{- if eq .TypeName "[]string"}}
	result.{{.Name}} = SplitList(record[{{$i}}])
	{{- else if isIntegerType .TypeName}}
	if val, err := strconv.Atoi(record[{{$i}}]); err != nil {
		return result, errors.Wrap(err, CsvErrorTitle, "Could not unmarshal to {{.TypeName}}: Expected integer value but got '%s'", record[{{$i}}])
	} else {
		{{- if eq .TypeName "int"}}
		result.{{.Name}} = val
		{{- else}}
		result.{{.Name}} = {{.TypeName}}(val)
		{{- end}}
	}
	{{- else if eq .TypeName "string"}}
	result.{{.Name}} = record[{{$i}}]
	{{- else if eq .TypeName "bool"}}
	if val, err := strconv.ParseBool(record[{{$i}}]); err != nil {
		return result, errors.Wrap(err, CsvErrorTitle, "Could not unmarshal to {{.TypeName}}: Expected boolean value, but got '%s'", record[{{$i}}])
	} else {
		result.{{.Name}} = val
	}
	{{- else}}
		{{typeError .TypeName}}
	{{- end}}
{{- end}}
	return result, nil
}

func (s {{.TypeName}}) ToRecord() []string {
	record := make([]string, {{len .Fields}})
	{{- range $i, $e := .Fields}}
	{{- if eq .TypeName "[]string"}}
	record[{{$i}}] = MakeList(s.{{.Name}})
	{{- else if isIntegerType .TypeName}}
	record[{{$i}}] = fmt.Sprintf("%d", s.{{.Name}})
	{{- else if eq .TypeName "string"}}
	record[{{$i}}] = s.{{.Name}}
	{{- else if eq .TypeName "bool"}}
	record[{{$i}}] = strconv.FormatBool(s.{{.Name}})
	{{- else}}
		{{typeError .TypeName}}
	{{- end}}
	{{- end}}
	return record
}

func Marshal{{.TypeName}}(records []{{.TypeName}}) [][]string {
	result := make([][]string, len(records))
	for i := range records {
		result[i] = records[i].ToRecord()
	}
	return result
}

func (r *Reader) Read{{.TypeName}}Records() ([]{{.TypeName}}, errors.Error) {
	defer r.Close()
	rows := make([]{{.TypeName}}, 0, 8)
	for {
		if record, err := r.ReadRecord(); err != nil {
			if err.Is(errors.EOF) {
				return rows, nil
			}
			return nil, err
		} else if unmarshalled, err := Unmarshal{{.TypeName}}(record); err != nil {
			return nil, err
		} else {
			rows = append(rows, unmarshalled)
		}
	}
}

func (w *Writer) Write{{.TypeName}}Records(rows []{{.TypeName}}) errors.Error {
	defer w.Close()
	for _, row := range rows {
		if err := w.WriteRecord(row.ToRecord()); err != nil {
			w.err = err
			return err
		}
	}
	
	if w.destination.Flush(); w.destination.Error() != nil {
		return errors.Wrap(w.destination.Error(), CsvErrorTitle, "Could not write to csv output file")
	}
	return nil
}
`

const Imports = `
import (
	"fmt"
	"returntypes-langserver/common/debug/errors"
	"strconv"
)`
