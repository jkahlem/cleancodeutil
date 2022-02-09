package dataset

import (
	"fmt"
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing/typeclasses"
	"returntypes-langserver/services/predictor"
	"strings"
)

// The Dataset Creator uses a type class mapper to create dataset rows out of a methods array.
type creator struct {
	typeLabelMapper    *TypeLabelMapper
	typeClassMapper    typeclasses.Mapper
	tree               packagetree.Tree
	methods            []csv.Method
	methodsReturnTypes []csv.Method
	methodsParameters  []csv.Method
	classes            []csv.Class
	err                errors.Error
}

// Creates a new dataset creator
func NewCreator() *creator {
	c := creator{}
	return &c
}

// Loads return types/class hierarchy data from the given files and creates datasets from it
func (c *creator) CreateTrainingAndEvaluationSet(methodsWithReturnTypesPath, classHierarchyPath string) {
	c.loadMethodsAndClasses(methodsWithReturnTypesPath, classHierarchyPath)
	datasetReturnTypes, datasetMethods := c.getDatasetRows()
	trainingSet, evaluationSet := c.splitDataset(datasetReturnTypes)
	//trainingSetMethods , evaluationSetMethods := c.splitDataset(datasetMethods)
	c.saveDatasets(trainingSet, evaluationSet)
	c.saveDatasetsMethods(datasetMethods, nil)
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
func (c *creator) getDatasetRows() ([]csv.DatasetRow, []csv.DatasetRow2) {
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

	if methods, err := c.typeClassMapper.MapMethodsTypesToTypeClass(c.methods); err != nil {
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
	summarizedMap := CreateMapOfSummarizedMethods(relevantMethods)
	c.methodsReturnTypes = SummarizeMethodsForReturnTypes(summarizedMap, relevantMethods)
	c.methodsParameters = SummarizeMethodsForParameters(summarizedMap, relevantMethods)
}

// Creates dataset rows of the methods
func (c *creator) convertMethodsToDatasetRows() ([]csv.DatasetRow, []csv.DatasetRow2) {
	if c.err != nil {
		return nil, nil
	}
	c.logln("Create dataset rows")
	rowsReturnTypes := make([]csv.DatasetRow, len(c.methodsReturnTypes))
	for i, method := range c.methodsReturnTypes {
		returnTypeLabel := c.typeLabelMapper.GetLabel(method.ReturnType)
		rowsReturnTypes[i].MethodName = string(predictor.GetPredictableMethodName(method.MethodName))
		rowsReturnTypes[i].TypeLabel = returnTypeLabel
	}
	rowsParameters := make([]csv.DatasetRow2, 0, len(c.methodsParameters))
	context := "string, int, float, enum, object, boolean"
	for _, method := range c.methodsParameters {
		name, pars := c.convertMethodDefinitionToSentence(method)
		row := csv.DatasetRow2{
			Prefix:     "generate parameters",
			MethodName: name,
			Parameters: pars,
		}
		rowsParameters = append(rowsParameters, row)
		if !csv.IsEmptyList(method.Parameters) {
			for _, par := range method.Parameters {
				splitted := strings.Split(par, " ")
				parType, parName := string(predictor.GetPredictableMethodName(splitted[0])), string(predictor.GetPredictableMethodName(splitted[1]))
				ctx := context
				if strings.Index(context, parType) == -1 {
					ctx = fmt.Sprintf("%s, %s", context, parType)
				}
				// TOOD: don't need a special fallback, as type assignment should for example pick "object" if it does not know any better thing
				row2 := csv.DatasetRow2{
					Prefix: "type assignment",
					// the "name" variable has currently already a dot '.' at the end. So no need to add it another time ...
					MethodName: fmt.Sprintf("method: %s name: %s. context: %s.", name, parName, ctx), // input_text
					Parameters: parType,                                                              // target_text
				}
				rowsParameters = append(rowsParameters, row2)
			}
		}
	}
	return rowsReturnTypes, rowsParameters
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

// Saves the dataset to the output path
func (c *creator) saveDatasetsMethods(trainingSet, evaluationSet []csv.DatasetRow2) {
	c.logln("Save methods datasets")
	c.writeDataset2(configuration.MethodsTrainingSetOutputPath(), trainingSet)
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

// Writes dataset rows into a csv file
func (c *creator) writeDataset2(outputPath string, dataset []csv.DatasetRow2) {
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
