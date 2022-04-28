// Currently no better name. This step targets outputting existing data to excel files using loaded data.
package excelOutputter

import (
	"fmt"
	"path/filepath"
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/dataformat/excel"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/common/utils/progressbar"
	"returntypes-langserver/processing/projects"
	"strings"
)

func CreateOutput(projects []projects.Project) errors.Error {
	log.Info("Write methods to excel files ...\n")

	methods, err := csv.NewFileReader(configuration.MethodsWithReturnTypesOutputPath()).ReadMethodRecords()
	if err != nil {
		return err
	}

	createOutputOnMethods(methods, configuration.MethodsWithReturnTypesExcelOutputDir(), configuration.ExcelSets())
	if configuration.CreateMethodOutputPerProject() {
		createOutputForProjects(methods, projects)
	}

	return nil
}

func createOutputForProjects(methods []csv.Method, projects []projects.Project) {
	log.Info("Write methods per project ...")
	progress := progressbar.StartNew(len(projects))
	defer progress.Finish()

	for _, project := range projects {
		progress.Increment()
		progress.SetOperation(project.Name())

		path := filepath.Join(configuration.MethodsWithReturnTypesExcelOutputDir(), "project-output", project.Name()+".xlsx")
		if utils.FileExists(path) {
			continue
		}

		projectMethods := make([][]string, 0, len(methods))
		for _, m := range methods {
			if strings.Split(m.FilePath, string(filepath.Separator))[0] == project.Name() {
				projectMethods = append(projectMethods, unqualifyTypeNames(m).ToRecord())
			}
		}
		excel.Stream().FromSlice(projectMethods).
			WithColumnsFromStruct(csv.Method{}).
			InsertColumnsAt(excel.Col(7), "Project", "Notes").
			Transform(addProjectColumn).
			ToFile(path)
	}
}

func createOutputOnMethods(methods []csv.Method, path string, sets []configuration.ExcelSet) {
	processors := make([]DatasetProcessor, 0, len(sets))
	for _, dataset := range sets {
		p := NewDatasetProcessor(dataset, path)
		if !p.hasOutput() {
			continue
		}
		processors = append(processors, p)
	}
	if len(processors) == 0 {
		return
	}

	progress := progressbar.StartNew(len(methods))
	defer progress.Finish()

	progress.SetOperation("Write methods")
	for _, method := range methods {
		progress.Increment()

		method = unqualifyTypeNames(method)
		for i := range processors {
			if !processors[i].accepts(method) {
				continue
			}
			processors[i].process(method)
		}
	}

	progress.SetOperation("Save output")
	for i := range processors {
		processors[i].close()
	}
}

func unqualifyTypeNames(method csv.Method) csv.Method {
	if parameters, err := java.ParseParameterList(method.Parameters); err != nil {
		for i, parameter := range parameters {
			// Add spaces here so they are present after the formatting for excel output
			space := ""
			if i > 0 {
				space = " "
			}
			// Directly overwrite parameter formatting to excel file format
			method.Parameters[i] = fmt.Sprintf("%s%s%s", space, parameter.Type.TypeName, parameter.Name)
		}
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
