// Currently no better name. This step targets outputting existing data to excel files using loaded data.
package excelOutputter

import (
	"encoding/json"
	"fmt"
	"os"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/dataformat/excel"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"strings"
)

type PatternType string

const (
	Wildcard PatternType = "wildcard"
	RegEx    PatternType = "regex"
)

type Configuration struct {
	Datasets []Dataset `json:"datasets"`
}

type Dataset struct {
	Name             string    `json:"name"`
	Filter           Filter    `json:"filter"`
	ConvertNumbers   bool      `json:"convertNumbers"`
	Subsets          []Dataset `json:"subsets"`
	LeftoverFilename string    `json:"leftoverFilename"`
}

type Filter struct {
	Includes FilterConfiguration `json:"includes"`
	Excludes FilterConfiguration `json:"excludes"`
}

type FilterConfiguration struct {
	Method     []Pattern `json:"method"`
	Modifier   []Pattern `json:"modifier"`
	Parameter  []Pattern `json:"parameter"`
	Label      []Pattern `json:"label"`
	ReturnType []Pattern `json:"returntype"`
	ClassName  []Pattern `json:"classname"`
}

type Pattern struct {
	Pattern string      `json:"pattern"`
	Type    PatternType `json:"type"`
}

func (p *Pattern) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if pattern, ok := v.(string); ok {
		p.Pattern = pattern
		p.Type = Wildcard
	} else if jsonObj, ok := v.(map[string]interface{}); ok {
		if err := p.unmarshalPattern(jsonObj); err != nil {
			return err
		} else if err := p.unmarshalType(jsonObj); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported pattern")
	}
	return nil
}

func (p *Pattern) unmarshalPattern(jsonObj map[string]interface{}) error {
	if pattern, ok := jsonObj["pattern"].(string); ok && pattern != "" {
		p.Pattern = pattern
		return nil
	} else {
		return fmt.Errorf("unsupported pattern")
	}
}

func (p *Pattern) unmarshalType(jsonObj map[string]interface{}) error {
	if typ, ok := jsonObj["type"].(PatternType); ok && (typ == RegEx || typ == Wildcard) {
		p.Type = typ
		return nil
	} else {
		return fmt.Errorf("unsupported type")
	}
}

func LoadConfiguration(filepath string) (Configuration, errors.Error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "Excel Output Error", "Could not load configuration.")
	}

	var config Configuration
	if err := json.Unmarshal(contents, &config); err != nil {
		return Configuration{}, errors.Wrap(err, "Excel Output Error", "Could not parse configuration.")
	}

	return config, nil
}

func CreateOutput() errors.Error {
	log.Info("Write output to excel file ...\n")

	return excel.ReportingStream().
		FromCSVFile(configuration.MethodsWithReturnTypesOutputPath()).
		WithColumnsFromStruct(csv.Method{}).
		Transform(unqualifyTypeNamesInRecord).
		InsertColumnsAt(excel.Col(7), "Project", "Notes").
		Transform(addProjectColumn).
		ToFile(configuration.MethodsWithReturnTypesExcelOutputPath())
}

func unqualifyTypeNamesInRecord(methodRecord []string) []string {
	method := csv.UnmarshalMethod([][]string{methodRecord})[0]

	for i, exception := range method.Exceptions {
		method.Exceptions[i] = unqualifyTypeName(exception)
	}
	for i, parameter := range method.Parameters {
		par := strings.Split(parameter, " ")
		// Add spaces here so they are present after .ToRecord() conversion
		space := ""
		if i > 0 {
			space = " "
		}
		par[0] = fmt.Sprintf("%s%s", space, unqualifyTypeName(par[0]))
		method.Parameters[i] = strings.Join(par, " ")
	}
	method.ReturnType = unqualifyTypeName(method.ReturnType)

	return method.ToRecord()
}

func unqualifyTypeName(typeName string) string {
	parts := strings.Split(typeName, ".")
	return parts[len(parts)-1]
}

func addProjectColumn(record []string) []string {
	filepath := record[len(record)-1]
	record[7] = strings.Split(filepath, "\\")[0]
	return record
}
