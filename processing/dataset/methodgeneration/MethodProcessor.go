package methodgeneration

import (
	"path/filepath"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/processing/typeclasses"
	"returntypes-langserver/services/predictor"
	"strings"
)

type Processor struct {
	OutputDir       string
	rows            []csv.MethodGenerationDatasetRow
	methods         utils.StringSet
	Options         configuration.SpecialOptions
	typeClassMapper typeclasses.Mapper
}

func NewProcessor(outputDir string, options configuration.SpecialOptions, tree *packagetree.Tree) base.MethodProcessor {
	processor := &Processor{
		OutputDir: outputDir,
		Options:   options,
		methods:   make(utils.StringSet),
	}
	if options.TypeClasses != "" {
		// TODO: Use typeclasses of the dataset for typeclass mapper ...
		processor.typeClassMapper = typeclasses.New(tree)
	}
	return processor
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
	if p.Options.TypeClasses != "" && p.typeClassMapper != nil {
		if err := p.mapTypeToTypeClasses(method); err != nil {
			return false, err
		}
	}
	p.rows = append(p.rows, p.mapMethodToDatasetRow(method))
	return false, nil
}

func (p *Processor) mapMethodToDatasetRow(method *csv.Method) csv.MethodGenerationDatasetRow {
	datasetRow := csv.MethodGenerationDatasetRow{
		ClassName:  method.ClassName,
		MethodName: string(predictor.GetPredictableMethodName(method.MethodName)),
		Parameters: method.Parameters,
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
	// TODO: Split evaluation set?
	csv.WriteCsvRecords(filepath.Join(p.OutputDir, "dataset.csv"), csv.MarshalMethodGenerationDatasetRow(p.rows))
	return nil
}
