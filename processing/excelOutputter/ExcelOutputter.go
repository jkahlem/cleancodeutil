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

	config, err := LoadConfiguration(configuration.AbsolutePathFromGoProjectDir("datasets.json"))
	if err != nil {
		return err
	}
	log.Info("Dataset configuration loaded\n")
	//log.Info("Dataset processors created\n")
	records, err := csv.ReadRecords(configuration.MethodsWithReturnTypesOutputPath())
	if err != nil {
		return err
	}

	log.Info("Write records...\n")
	createOutputOnRecords(records, configuration.MethodsWithReturnTypesExcelOutputDir(), config)

	return nil
}

func createOutputOnRecords(records [][]string, path string, config Configuration) {
	processors := make([]DatasetProcessor, 0, len(config.Datasets))
	for _, dataset := range config.Datasets {
		processors = append(processors, NewDatasetProcessor(dataset, configuration.MethodsWithReturnTypesExcelOutputDir()))
	}
	for recordIndex, method := range csv.UnmarshalMethod(records) {
		if (recordIndex+1)%100 == 0 {
			log.Info("Write record %d of %d\n", recordIndex+1, len(records))
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

func unqualifyTypeNamesInRecord(methodRecord []string) []string {
	method := csv.UnmarshalMethod([][]string{methodRecord})[0]
	return unqualifyTypeNames(method).ToRecord()
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
