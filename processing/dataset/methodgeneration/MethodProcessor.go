package methodgeneration

import (
	"path/filepath"
	"returntypes-langserver/common/code/java"
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
	datasetRow := csv.MethodGenerationDatasetRow{
		ClassName:    method.ClassName,
		MethodName:   string(predictor.GetPredictableMethodName(method.MethodName)),
		ReturnType:   utils.GetStringExtension(method.ReturnType, "."),
		Parameters:   method.Parameters,
		ContextTypes: p.getContextTypes(*method),
		IsStatic:     utils.ContainsString(method.Modifier, "static"),
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
func (p *Processor) mapParameterTypesToTypeClasses(rawParameters []string) ([]string, errors.Error) {
	if csv.IsEmptyList(rawParameters) {
		return nil, nil
	}

	parameters, err := java.ParseParameterList(rawParameters)
	if err != nil {
		return nil, err
	}
	for i, parameter := range parameters {
		var labels []string
		if parameter.Type.IsArrayType {
			labels = []string{string(java.ArrayType)}
		}
		if typeClass, err := p.typeClassMapper.MapParameterTypeToTypeClass(parameter.Type.TypeName, labels); err != nil {
			return nil, err
		} else {
			parameters[i].Type.TypeName = typeClass
		}
	}
	return java.FormatParameterList(parameters, nil), nil
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

func (p *Processor) getContextTypes(method csv.Method) []string {
	contextTypes := make([]string, 0, 6)
	if val, ok := p.files[method.FilePath]; ok {
		contextTypes = append(contextTypes, val...)
	}
	contextTypes = p.appendContextType(contextTypes, method.ReturnType)
	if parameters, err := java.ParseParameterList(method.Parameters); err != nil {
		for _, par := range parameters {
			contextTypes = p.appendContextType(contextTypes, par.Type.TypeName)
		}
	}
	if len(contextTypes) == 0 {
		return nil
	}
	return contextTypes
}

func (p *Processor) appendContextType(contextTypes []string, typeName string) []string {
	shortened := utils.GetStringExtension(typeName, ".")
	if utils.ContainsString(contextTypes, typeName) {
		contextTypes = append(contextTypes, shortened)
	}
	return contextTypes
}
