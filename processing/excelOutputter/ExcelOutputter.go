// Currently no better name. This step targets outputting existing data to excel files using loaded data.
package excelOutputter

import (
	"fmt"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"strings"
)

func CreateOutput() errors.Error {
	log.Info("Write output to excel file ...\n")

	methods, err := csv.NewFileReader(configuration.MethodsWithReturnTypesOutputPath()).ReadMethodRecords()
	if err != nil {
		return err
	}

	log.Info("Write records...\n")
	createOutputOnMethods(methods, configuration.MethodsWithReturnTypesExcelOutputDir(), configuration.ExcelSets())

	return nil
}

func createOutputOnMethods(methods []csv.Method, path string, sets []configuration.ExcelSet) {
	processors := make([]DatasetProcessor, 0, len(sets))
	for _, dataset := range sets {
		processors = append(processors, NewDatasetProcessor(dataset, configuration.MethodsWithReturnTypesExcelOutputDir()))
	}
	for recordIndex, method := range methods {
		if (recordIndex+1)%100 == 0 {
			log.Info("Write record %d of %d\n", recordIndex+1, len(methods))
		}
		method = unqualifyTypeNames(method)
		for i := range processors {
			if !processors[i].accepts(method) {
				continue
			}
			processors[i].process(method)
		}
	}
	for i := range processors {
		processors[i].close()
	}
}

func unqualifyTypeNames(method csv.Method) csv.Method {
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
	return method
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
