package dataset

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/csv"
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/java"
	"returntypes-langserver/common/log"
	"returntypes-langserver/common/packagetree"
	"returntypes-langserver/processing/typeclasses"
	"returntypes-langserver/services/predictor"
)

// The Dataset Creator uses a type class mapper to create dataset rows out of a methods array.
type creator struct {
	typeLabelMapper *TypeLabelMapper
	typeClassMapper typeclasses.Mapper
	tree            packagetree.Tree
	methods         []csv.Method
	classes         []csv.Class
	err             errors.Error
}

// Creates a new dataset creator
func NewCreator() *creator {
	c := creator{}
	return &c
}

// Loads return types/class hierarchy data from the given files and creates datasets from it
func (c *creator) CreateTrainingAndEvaluationSet(methodsWithReturnTypesPath, classHierarchyPath string) {
	c.loadMethodsAndClasses(methodsWithReturnTypesPath, classHierarchyPath)
	dataset := c.getDatasetRows()
	trainingSet, evaluationSet := c.splitDataset(dataset)
	c.saveDatasets(trainingSet, evaluationSet)
	c.saveLabelMappings()
}

// Loads the methods and class data into the creator
func (c *creator) loadMethodsAndClasses(methodsWithReturnTypesPath, classHierarchyPath string) {
	c.logln("Load data...")
	methodsRecords, err := csv.ReadRecords(methodsWithReturnTypesPath)
	if err != nil {
		c.err = err
		return
	}

	classesRecords, err := csv.ReadRecords(classHierarchyPath)
	if err != nil {
		c.err = err
		return
	}

	for _, defaultLibrary := range configuration.DefaultLibraries() {
		defaultClassesRecords, err := csv.ReadRecords(defaultLibrary)
		if err != nil {
			c.err = err
			return
		}
		classesRecords = append(classesRecords, defaultClassesRecords...)
	}

	c.methods = csv.UnmarshalMethod(methodsRecords)
	c.classes = csv.UnmarshalClasses(classesRecords)
}

// Creates dataset rows from the loaded methods
func (c *creator) getDatasetRows() []csv.DatasetRow {
	c.createPackageTree()
	c.createTypeClassMapper()
	c.createTypeLabeler()
	c.mapMethodReturnTypesToTypeClasses()
	c.filterMethodsToRelevantMethods()
	return c.convertMethodsToDatasetRows()
}

func (c *creator) createPackageTree() {
	if c.err != nil {
		return
	}

	c.logln("Create package tree...")
	c.tree = packagetree.New()
	java.FillPackageTreeByCsvClassNodes(&c.tree, c.classes)
}

func (c *creator) createTypeClassMapper() {
	c.typeClassMapper = typeclasses.New(&c.tree)
}

func (c *creator) createTypeLabeler() {
	c.typeLabelMapper = &TypeLabelMapper{}
}

func (c *creator) mapMethodReturnTypesToTypeClasses() {
	if c.err != nil {
		return
	}

	if methods, err := c.typeClassMapper.MapReturnTypesToTypeClass(c.methods); err != nil {
		c.err = err
		return
	} else {
		c.methods = methods
	}

	if !configuration.StatisticsSkipCreation() {
		if err := c.writeMethodsWithTypeClasses(); err != nil {
			log.ReportProblem("Could not write data for statistics generation.")
			if configuration.StrictMode() {
				c.err = err
				return
			}
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
	relevantMethods := FilterMethodsByLabels(c.methods)
	c.methods = SummarizeMethods(relevantMethods)
	return
}

// Creates dataset rows of the methods
func (c *creator) convertMethodsToDatasetRows() []csv.DatasetRow {
	if c.err != nil {
		return nil
	}
	c.logln("Create dataset rows")
	rows := make([]csv.DatasetRow, len(c.methods))
	for i, method := range c.methods {
		returnTypeLabel := c.typeLabelMapper.GetLabel(method.ReturnType)
		rows[i].MethodName = string(predictor.GetPredictableMethodName(method.MethodName))
		rows[i].TypeLabel = returnTypeLabel
	}
	return rows
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
