package csv

import (
	"fmt"
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/log"
	"strconv"
)

// Generate Marshal / Unmarshal methods (-> Marshaller.go)
//go:generate go run ./marshallerGenerator

type Method struct {
	MethodName string
	ReturnType string
	Labels     []string
	FilePath   string
	// Parameters are in this format: "<type> <name>" (seperated by a single space)
	Parameters []string
}

type Class struct {
	ClassName string
	Extends   []string
}

type TypeConversion struct {
	SourceType      string
	DestinationType string
}

type DatasetRow struct {
	MethodName string
	TypeLabel  int
}
type DatasetRow2 struct {
	Prefix     string
	MethodName string
	Parameters string
}

type TypeLabel struct {
	Name  string
	Label int
}

// Returns true if the label is defined in the methods label list
func (method Method) HasLabel(searchedLabel string) bool {
	for _, methodLabel := range method.Labels {
		if methodLabel == searchedLabel {
			return true
		}
	}
	return false
}

func UnmarshalMethod(records [][]string) []Method {
	methods := make([]Method, len(records))
	for i, record := range records {
		methods[i].MethodName = record[0]
		methods[i].ReturnType = record[1]
		methods[i].Labels = SplitList(record[2])
		methods[i].FilePath = record[3]
		if len(record) >= 4 {
			methods[i].Parameters = SplitList(record[4])
		}
	}
	return methods
}

func (method Method) ToRecord() []string {
	return []string{
		method.MethodName,
		method.ReturnType,
		MakeList(method.Labels),
		method.FilePath,
		MakeList(method.Parameters),
	}
}

func UnmarshalTypeConversion(records [][]string) []TypeConversion {
	convs := make([]TypeConversion, len(records))
	for i, record := range records {
		convs[i].SourceType = record[0]
		convs[i].DestinationType = record[1]
	}
	return convs
}

func (conv TypeConversion) ToRecord() []string {
	return []string{
		conv.SourceType,
		conv.DestinationType,
	}
}

func UnmarshalClasses(records [][]string) []Class {
	classes := make([]Class, len(records))
	for i, record := range records {
		classes[i].ClassName = record[0]
		classes[i].Extends = SplitList(record[1])
	}
	return classes
}

func (class Class) ToRecord() []string {
	return []string{
		class.ClassName,
		MakeList(class.Extends),
	}
}

func UnmarshalDatasetRow(records [][]string) []DatasetRow {
	datasetRow := make([]DatasetRow, len(records))
	for i, record := range records {
		datasetRow[i].MethodName = record[0]
		datasetRow[i].TypeLabel = parseInt(record[1], true)
	}
	return datasetRow
}

func (datasetRow DatasetRow) ToRecord() []string {
	return []string{
		string(datasetRow.MethodName),
		fmt.Sprintf("%d", datasetRow.TypeLabel),
	}
}

func UnmarshalDatasetRow2(records [][]string) []DatasetRow2 {
	datasetRow := make([]DatasetRow2, len(records))
	for i, record := range records {
		datasetRow[i].Prefix = record[0]
		datasetRow[i].MethodName = record[1]
		datasetRow[i].Parameters = record[2]
	}
	return datasetRow
}

func (datasetRow DatasetRow2) ToRecord() []string {
	return []string{
		datasetRow.Prefix,
		datasetRow.MethodName,
		datasetRow.Parameters,
	}
}

func UnmarshalTypeLabel(records [][]string) []TypeLabel {
	labels := make([]TypeLabel, len(records))
	for i, record := range records {
		labels[i].Name = record[0]
		labels[i].Label = parseInt(record[1], true)
	}
	return labels
}

func (typeLabel TypeLabel) ToRecord() []string {
	return []string{
		typeLabel.Name,
		fmt.Sprintf("%d", typeLabel.Label),
	}
}

func parseInt(raw string, strict bool) int {
	result, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		wrappedErr := errors.Wrap(err, CsvErrorTitle, "Could not unmarshal csv data")
		if strict {
			log.ReportProblemWithError(wrappedErr, "An error occured while unmarshalling data")
		} else {
			log.Error(wrappedErr)
			log.ReportProblem("An error occured while unmarshalling data")
		}
	}
	return int(result)
}
