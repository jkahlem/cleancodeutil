package returntypesvalidation

import (
	"fmt"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/processing/typeclasses"
	"returntypes-langserver/services/predictor"
	"strings"
)

// The Dataset Creator uses a type class mapper to create dataset rows out of a methods array.
type creator struct {
	typeLabelMapper *base.TypeLabelMapper
	typeClassMapper typeclasses.Mapper
	methods         []csv.Method
	err             errors.Error
}

// Creates a new dataset creator
func New(methods []csv.Method, tree *packagetree.Tree) base.Creator {
	return &creator{
		methods:         methods,
		typeClassMapper: typeclasses.New(tree),
		typeLabelMapper: &base.TypeLabelMapper{},
	}
}

// Loads return types/class hierarchy data from the given files and creates datasets from it
func (c *creator) Create() errors.Error {
	datasetReturnTypes := c.getDatasetRows()
	trainingSet, evaluationSet := c.splitDataset(datasetReturnTypes)
	c.saveDatasets(trainingSet, evaluationSet)
	c.saveLabelMappings()
	return c.err
}

// Creates dataset rows from the loaded methods
func (c *creator) getDatasetRows() []csv.DatasetRow {
	c.mapMethodReturnTypesToTypeClasses()
	c.filterMethodsToRelevantMethods()
	return c.convertMethodsToDatasetRows()
}

func (c *creator) mapMethodReturnTypesToTypeClasses() {
	if c.err != nil {
		return
	}

	for i, method := range c.methods {
		if returnType, err := c.typeClassMapper.MapReturnTypeToTypeClass2(method.ReturnType, method.Labels); err != nil {
			c.err = err
			return
		} else {
			c.methods[i].ReturnType = returnType
		}
	}

	if !configuration.StatisticsSkipCreation() {
		c.createMethodsWithTypeClassesStatistics()
	}
}

func (c *creator) createMethodsWithTypeClassesStatistics() {
	if err := c.writeMethodsWithTypeClasses(); err != nil {
		log.ReportProblem("Could not write data for statistics generation.")
		if configuration.StrictMode() {
			c.err = err
			return
		}
	}
}

func (c *creator) writeMethodsWithTypeClasses() errors.Error {
	records := make([][]string, len(c.methods))
	for i, method := range c.methods {
		records[i] = method.ToRecord()
	}
	return csv.WriteCsvRecords(configuration.MethodsWithTypeClassesOutputPath(), records)
}

// Filters methods to the "relevant" methods for the dataset (no getters/setters etc.)
func (c *creator) filterMethodsToRelevantMethods() {
	if c.err != nil {
		return
	}

	c.logln("filter methods to relevant methods...")
	relevantMethods := base.FilterMethodsByLabels(c.methods)
	summarizedMap := base.CreateMapOfSummarizedMethods(relevantMethods)
	c.methods = base.SummarizeMethodsForReturnTypes(summarizedMap, relevantMethods)
}

// Creates dataset rows of the methods
func (c *creator) convertMethodsToDatasetRows() []csv.DatasetRow {
	if c.err != nil {
		return nil
	}
	c.logln("Create dataset rows")
	rowsReturnTypes := make([]csv.DatasetRow, len(c.methods))
	for i, method := range c.methods {
		returnTypeLabel := c.typeLabelMapper.GetLabel(method.ReturnType)
		rowsReturnTypes[i].MethodName = string(predictor.GetPredictableMethodName(method.MethodName))
		rowsReturnTypes[i].TypeLabel = returnTypeLabel
	}
	return rowsReturnTypes
}

func (c *creator) convertMethodDefinitionToSentence(method csv.Method) (string, string) {
	methodName := string(predictor.GetPredictableMethodName(method.MethodName))
	parameters := make([]string, 0, len(method.Parameters))
	if csv.IsEmptyList(method.Parameters) {
		parameters = append(parameters, "void")
	} else {
		for _, par := range method.Parameters {
			splitted := strings.Split(par, " ")
			parameters = append(parameters, string(predictor.GetPredictableMethodName(splitted[1]))) //append(parameters, fmt.Sprintf("%s %s", splitted[0], string(predictor.GetPredictableMethodName(splitted[1]))))
		}
	}
	return fmt.Sprintf("%s.", methodName), strings.Join(parameters, ", ") + "."
}

// Splits a dataset to a training set and evaluation set
func (c *creator) splitDataset(dataset []csv.DatasetRow) (trainingSet, evaluationSet []csv.DatasetRow) {
	if c.err != nil {
		return
	}
	c.logln("Split dataset")
	trainingSet, evaluationSet = SplitToTrainingAndEvaluationSet(dataset, configuration.DatasetSize())
	return
}

// Saves the dataset to the output path
func (c *creator) saveDatasets(trainingSet, evaluationSet []csv.DatasetRow) {
	c.logln("Save datasets")
	c.writeDataset(configuration.TrainingSetOutputPath(), trainingSet)
	c.writeDataset(configuration.EvaluationSetOutputPath(), evaluationSet)
}

// Saves the type label mappings to the output path
func (c *creator) saveLabelMappings() {
	if c.err != nil {
		return
	}
	if err := c.typeLabelMapper.WriteMappings(configuration.DatasetLabelsOutputPath()); err != nil {
		c.err = err
	}
}

// Writes dataset rows into a csv file
func (c *creator) writeDataset(outputPath string, dataset []csv.DatasetRow) {
	if c.err != nil {
		return
	}
	records := make([][]string, len(dataset))
	for i, row := range dataset {
		records[i] = row.ToRecord()
	}

	if err := csv.WriteCsvRecords(outputPath, records); err != nil {
		c.err = err
	}
}

func (c *creator) logln(str string) {
	if c.err != nil {
		return
	}
	log.Info(str + "\n")
}

func (c *creator) Err() errors.Error {
	return c.err
}
