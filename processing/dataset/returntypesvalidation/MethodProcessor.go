package returntypesvalidation

import (
	"fmt"
	"path/filepath"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/processing/typeclasses"
	"returntypes-langserver/services/predictor"
)

const (
	TrainingSetFileName   = "trainingSet_returntypesvalidation.csv"
	EvaluationSetFileName = "evaluationSet_returntypesvalidation.csv"
	LabelSetFileName      = "labelSet_returntypesvalidation.csv"
)

type Processor struct {
	OutputDir       string
	methodsSet      map[string]ReturnTypes
	Options         configuration.SpecialOptions
	typeClassMapper typeclasses.Mapper
	typeLabelMapper *base.TypeLabelMapper
}

type ReturnTypes map[string]int

func (r ReturnTypes) Put(typeName string) {
	if _, ok := r[typeName]; !ok {
		r[typeName] = 1
	} else {
		r[typeName]++
	}
}

func (r ReturnTypes) MostUsedType() string {
	max, maxType := 0, ""
	for typeName, count := range r {
		if count > max {
			max = count
			maxType = typeName
		}
	}
	return maxType
}

func NewProcessor(outputDir string, options configuration.SpecialOptions, tree *packagetree.Tree) (base.MethodProcessor, errors.Error) {
	if options.DatasetSize.Training == 0 && options.DatasetSize.Evaluation == 0 {
		return nil, errors.New("Dataset error", fmt.Sprintf("invalid/unset dataset size values for set under '%s'", outputDir))
	}

	processor := &Processor{
		OutputDir:  outputDir,
		Options:    options,
		methodsSet: make(map[string]ReturnTypes),
	}
	if options.TypeClasses != "" {
		// TODO: Use typeclasses of the dataset for typeclass mapper ...
		processor.typeClassMapper = typeclasses.New(tree)
		processor.typeLabelMapper = &base.TypeLabelMapper{}
	}
	return processor, nil
}

func (p *Processor) Process(method *csv.Method) (isFiltered bool, err errors.Error) {
	identifier := p.getIdentifier(method)
	if p.Options.TypeClasses != "" && p.typeClassMapper != nil {
		if err := p.mapTypeToTypeClasses(method); err != nil {
			return false, err
		}
	}
	if _, ok := p.methodsSet[identifier]; !ok {
		p.methodsSet[identifier] = make(ReturnTypes)
	}
	p.methodsSet[identifier].Put(method.ReturnType)
	return false, nil
}

func (p *Processor) mapTypeToTypeClasses(method *csv.Method) errors.Error {
	if returnType, err := p.typeClassMapper.MapReturnTypeToTypeClass(method.ReturnType, method.Labels); err != nil {
		return err
	} else {
		method.ReturnType = returnType
	}
	return nil
}

func (p *Processor) getIdentifier(method *csv.Method) string {
	return string(predictor.GetPredictableMethodName(method.MethodName))
}

func (p *Processor) Close() errors.Error {
	// Filtering of same method definitions can only be done at the end to determine the desired return type
	// Also this needs to be done by each processor, as subsets and so on might filter out further methods which
	// leads to different results
	rows, i := make([]csv.ReturnTypesDatasetRow, len(p.methodsSet)), 0
	for methodName, returnTypes := range p.methodsSet {
		rows[i] = csv.ReturnTypesDatasetRow{
			MethodName: methodName,
			TypeLabel:  p.typeLabelMapper.GetLabel(returnTypes.MostUsedType()),
		}
		i++
	}

	trainingSet, evaluationSet := SplitToTrainingAndEvaluationSet(rows, p.Options.DatasetSize)
	if err := csv.WriteCsvRecords(filepath.Join(p.OutputDir, TrainingSetFileName), csv.MarshalReturnTypesDatasetRow(trainingSet)); err != nil {
		return err
	} else if err := csv.WriteCsvRecords(filepath.Join(p.OutputDir, EvaluationSetFileName), csv.MarshalReturnTypesDatasetRow(evaluationSet)); err != nil {
		return err
	} else if err := csv.WriteCsvRecords(filepath.Join(p.OutputDir, LabelSetFileName), csv.MarshalTypeLabel(p.typeLabelMapper.GetMappings())); err != nil {
		return err
	}
	return nil
}
