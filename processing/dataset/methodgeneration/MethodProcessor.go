package methodgeneration

import (
	"path/filepath"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/processing/typeclasses"
	"returntypes-langserver/services/predictor"
	"strings"
)

const (
	TrainingSetFileName   = "methodgeneration_trainingSet.csv"
	EvaluationSetFileName = "methodgeneration_evaluationSet.csv"
)

type Processor struct {
	OutputDir       string
	rows            []csv.MethodGenerationDatasetRow
	methods         utils.StringSet
	Options         configuration.SpecialOptions
	typeClassMapper typeclasses.Mapper
	skip            bool
	files           map[string][]string // TODO: find some way to pass the files/store them globally
}

func NewProcessor(outputDir string, options configuration.SpecialOptions, tree *packagetree.Tree) (base.MethodProcessor, errors.Error) {
	processor := &Processor{
		OutputDir: outputDir,
		Options:   options,
		methods:   make(utils.StringSet),
	}
	if utils.FileExists(processor.trainingFilePath()) {
		processor.skip = true
	}
	if records, err := csv.NewFileReader(configuration.FileContextTypesOutputPath()).ReadFileContextTypesRecords(); err != nil {
		return nil, err
	} else {
		processor.files = make(map[string][]string)
		for _, record := range records {
			processor.files[record.FilePath] = record.ContextTypes
		}
	}
	if options.TypeClasses != nil {
		if typeClassMapper, err := typeclasses.New(tree, options.TypeClasses); err != nil {
			return nil, err
		} else {
			processor.typeClassMapper = typeClassMapper
		}
	}
	return processor, nil
}

func (p *Processor) CanBeSkipped() bool {
	return p.skip
}

func (p *Processor) Process(method *csv.Method) (isFiltered bool, err errors.Error) {
	if p.Options.FilterDuplicates {
		identifier := p.getIdentifier(method)
		if p.methods.Has(identifier) {
			return true, nil
		} else {
			p.methods.Put(identifier)
		}
	}
	if p.Options.TypeClasses != nil && p.typeClassMapper != nil {
		if err := p.mapTypeToTypeClasses(method); err != nil {
			return false, err
		}
	}
	p.rows = append(p.rows, p.mapMethodToDatasetRow(method))
	return false, nil
}

func (p *Processor) mapMethodToDatasetRow(method *csv.Method) csv.MethodGenerationDatasetRow {
	parameters := method.Parameters
	if csv.IsEmptyList(parameters) {
		parameters = []string{"void"}
	}
	datasetRow := csv.MethodGenerationDatasetRow{
		ClassName:    method.ClassName,
		MethodName:   string(predictor.GetPredictableMethodName(method.MethodName)),
		ReturnType:   utils.GetStringExtension(method.ReturnType, "."),
		Parameters:   parameters,
		ContextTypes: p.getContextTypes(method.FilePath),
	}
	return datasetRow
}

func (p *Processor) mapTypeToTypeClasses(method *csv.Method) errors.Error {
	if returnType, err := p.typeClassMapper.MapReturnTypeToTypeClass(method.ReturnType, method.Labels); err != nil {
		return err
	} else {
		method.ReturnType = returnType
	}

	if parameters, err := p.mapParameterTypesToTypeClasses(method.Parameters); err != nil {
		return err
	} else {
		method.Parameters = parameters
	}
	return nil
}

// Maps the parameters to have a type class instead of the type name ...
func (p *Processor) mapParameterTypesToTypeClasses(parameters []string) ([]string, errors.Error) {
	if csv.IsEmptyList(parameters) {
		return nil, nil
	}

	results := make([]string, len(parameters))
	for i, parameter := range parameters {
		// splitted has for each element the pattern "<type> <name>"
		splitted := strings.Split(parameter, " ")
		// TODO: Method labels for parameter types? (e.g. array type for array type parameters ...)
		if typeClass, err := p.typeClassMapper.MapParameterTypeToTypeClass(splitted[0], nil); err != nil {
			return nil, err
		} else {
			splitted[0] = typeClass
			results[i] = strings.Join(splitted, " ")
		}
	}
	return results, nil
}

func (p *Processor) getIdentifier(method *csv.Method) string {
	return strings.Join([]string{method.ClassName, method.MethodName}, ".")
}

func (p *Processor) Close() errors.Error {
	log.Info("Close dataset file at %s\n", p.trainingFilePath())
	proportion := p.Options.DatasetSize
	trainingSetSize, _ := utils.FitProportions(proportion.Training, proportion.Evaluation, len(p.rows))
	trainingSet, evaluationSet := p.rows[:trainingSetSize], p.rows[trainingSetSize:]

	if err := csv.NewFileWriter(p.trainingFilePath()).WriteMethodGenerationDatasetRowRecords(trainingSet); err != nil {
		return err
	} else if err := csv.NewFileWriter(p.OutputDir, EvaluationSetFileName).WriteMethodGenerationDatasetRowRecords(evaluationSet); err != nil {
		return err
	}
	return nil
}

func (p *Processor) trainingFilePath() string {
	return filepath.Join(p.OutputDir, TrainingSetFileName)
}

func (p *Processor) getContextTypes(filePath string) []string {
	if val, ok := p.files[filePath]; ok {
		return val
	}
	/*selector := p.tree.Select(className)
	node := selector.Get()
	for node != nil {
		if fileNode, ok := node.(*java.CodeFile); ok {
			if fileNode.FilePath == p.currentFile {
				return p.currentFileContextTypes
			}
			// Cache context types per file
			p.currentFile = fileNode.FilePath
			p.currentFileContextTypes = p.getContextTypesOfFileNode(fileNode)
			return p.currentFileContextTypes
		} else {
			node = node.ParentNode()
		}
	}*/
	return nil
}

/*
func (p *Processor) getContextTypesOfFileNode(fileNode *java.CodeFile) []string {
	types := make([]string, len(fileNode.Imports)+len(fileNode.Classes))
	for i, importType := range fileNode.Imports {
		types[i] = utils.GetStringExtension(importType.ImportPath, ".")
	}
	for _, class := range fileNode.Classes {
		types = append(types, p.getContextTypesOfClassNode(class)...)
	}
	return nil
}

func (p *Processor) getContextTypesOfClassNode(classNode *java.Class) []string {
	types := make([]string, 1, len(classNode.Classes)+1)
	types[0] = classNode.ClassName
	for _, class := range classNode.Classes {
		types = append(types, p.getContextTypesOfClassNode(class)...)
	}
	return types
}
*/
