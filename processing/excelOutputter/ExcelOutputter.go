// Currently no better name. This step targets outputting existing data to excel files using loaded data.
package excelOutputter

import (
	"fmt"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/dataformat/excel"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"strings"
)

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
