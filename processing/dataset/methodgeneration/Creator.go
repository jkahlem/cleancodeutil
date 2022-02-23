package methodgeneration

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

type creator struct {
	typeClassMapper typeclasses.Mapper
	methods         []csv.Method
	err             errors.Error
}

// Creates a new dataset creator
func New(methods []csv.Method, tree *packagetree.Tree) base.Creator {
	return &creator{
		methods:         methods,
		typeClassMapper: typeclasses.New(tree),
	}
}

// Loads return types/class hierarchy data from the given files and creates datasets from it
func (c *creator) Create() errors.Error {
	datasetMethods := c.getDatasetRows()
	// TODO: Split into evaluation sets?
	c.saveDataset(datasetMethods)
	return c.err
}

// Creates dataset rows from the loaded methods
func (c *creator) getDatasetRows() []csv.MethodGenerationDatasetRow {
	c.mapMethodReturnTypesToTypeClasses()
	c.filterMethodsToRelevantMethods()
	return c.convertMethodsToDatasetRows()
}

// Map parameter types to type classes (if no type assignments used?)
func (c *creator) mapMethodReturnTypesToTypeClasses() {
	if c.err != nil {
		return
	}

	// TODO: Check for configuration if type classes or the raw types should be used.
	for i, method := range c.methods {
		if parameters, err := c.mapParameterTypesToTypeClasses(method.Parameters); err != nil {
			c.err = err
			return
		} else {
			c.methods[i].Parameters = parameters
		}
	}
}

// Maps the parameters to have a type class instead of the type name ...
func (c *creator) mapParameterTypesToTypeClasses(parameters []string) ([]string, errors.Error) {
	if csv.IsEmptyList(parameters) {
		return nil, nil
	}
	results := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		// splitted has for each element the pattern "<type> <name>"
		splitted := strings.Split(parameter, " ")
		// TODO: Method labels for parameter types? (e.g. array type for array type parameters ...)
		if typeClass, err := c.typeClassMapper.MapParameterTypeToTypeClass(splitted[0], nil); err != nil {
			return nil, err
		} else {
			splitted[0] = typeClass
			results = append(results, strings.Join(splitted, " "))
		}
	}
	return results, nil
}

// Filters methods to the "relevant" methods for the dataset (no getters/setters etc.)
func (c *creator) filterMethodsToRelevantMethods() {
	if c.err != nil {
		return
	}

	c.logln("filter methods to relevant methods...")
	c.methods = base.FilterMethodsByLabels(c.methods)
}

// Creates dataset rows of the methods
func (c *creator) convertMethodsToDatasetRows() []csv.MethodGenerationDatasetRow {
	if c.err != nil {
		return nil
	}
	c.logln("Create dataset rows")
	rowsParameters := make([]csv.MethodGenerationDatasetRow, 0, len(c.methods))
	context := "string, int, float, enum, object, boolean"
	for _, method := range c.methods {
		name, pars := c.convertMethodDefinitionToSentence(method)
		row := csv.MethodGenerationDatasetRow{
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
				if !strings.Contains(context, parType) {
					ctx = fmt.Sprintf("%s, %s", context, parType)
				}
				row2 := csv.MethodGenerationDatasetRow{
					Prefix: "type assignment",
					// the "name" variable has currently already a dot '.' at the end. So no need to add it another time ...
					MethodName: fmt.Sprintf("method: %s name: %s. context: %s.", name, parName, ctx), // input_text
					Parameters: parType,                                                              // target_text
				}
				rowsParameters = append(rowsParameters, row2)
			}
		}
	}
	return rowsParameters
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

// Saves the dataset to the output path
func (c *creator) saveDataset(trainingSet []csv.MethodGenerationDatasetRow) {
	c.logln("Save methods datasets")
	c.writeDataset(configuration.MethodsTrainingSetOutputPath(), trainingSet)
}

// Writes dataset rows into a csv file
func (c *creator) writeDataset(outputPath string, dataset []csv.MethodGenerationDatasetRow) {
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
