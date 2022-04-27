package dataset

import (
	"fmt"
	"path/filepath"
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/metrics"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/processing/dataset/methodgeneration"
	"returntypes-langserver/processing/dataset/returntypesvalidation"
	"returntypes-langserver/services/predictor"
)

// Common dataset processor which "preprocesses" the data (like applying common filters and so on)
type DatasetProcessor struct {
	ModelProcessor base.MethodProcessor
	TargetSet      configuration.Dataset
	SubProcessors  []DatasetProcessor
	tree           *packagetree.Tree
}

// Helper type which passes the calls to all slice elements
type DatasetProcessors []DatasetProcessor

func (p *DatasetProcessors) Process(method csv.Method) errors.Error {
	for i := range *p {
		if err := (*p)[i].Process(method); err != nil {
			return err
		}
	}
	return nil
}

func (p *DatasetProcessors) Close() errors.Error {
	for i := range *p {
		if err := (*p)[i].Close(); err != nil {
			return err
		}
	}
	return nil
}

func NewProcessor(set configuration.Dataset, modelType configuration.ModelType, path string, tree *packagetree.Tree) (DatasetProcessor, errors.Error) {
	path = filepath.Join(path, set.Name())
	processor := DatasetProcessor{
		TargetSet:     set,
		SubProcessors: make([]DatasetProcessor, 0, len(set.Subsets)),
	}
	if err := processor.initializeModelProcessor(modelType, path, tree); err != nil {
		return processor, err
	}

	for _, subset := range set.Subsets {
		if subprocessor, err := NewProcessor(inheritOptions(set, subset), modelType, path, tree); err != nil {
			return processor, err
		} else if !subprocessor.CanBeSkipped() {
			processor.SubProcessors = append(processor.SubProcessors, subprocessor)
		}
	}
	return processor, nil
}

func (p *DatasetProcessor) CanBeSkipped() bool {
	if !configuration.SkipIfOutputExists() {
		return false
	}
	if p.ModelProcessor != nil && !p.ModelProcessor.CanBeSkipped() {
		return false
	}
	for i := range p.SubProcessors {
		if !p.SubProcessors[i].CanBeSkipped() {
			return false
		}
	}
	return true
}

func (p *DatasetProcessor) initializeModelProcessor(modelType configuration.ModelType, path string, tree *packagetree.Tree) (err errors.Error) {
	var processor base.MethodProcessor
	switch modelType {
	case configuration.ReturnTypesValidator:
		processor, err = returntypesvalidation.NewProcessor(path, p.TargetSet.SpecialOptions, tree)
	case configuration.MethodGenerator:
		processor, err = methodgeneration.NewProcessor(path, p.TargetSet.SpecialOptions, tree)
	}
	p.ModelProcessor = processor
	return err
}

func (p *DatasetProcessor) Process(method csv.Method) errors.Error {
	if included, err := p.isIncluded(method); !included {
		return err
	}

	if p.ModelProcessor != nil {
		if isFiltered, err := p.ModelProcessor.Process(&method); err != nil {
			return err
		} else if isFiltered {
			return nil
		}
	}

	for i := range p.SubProcessors {
		if err := p.SubProcessors[i].Process(method); err != nil {
			return err
		}
	}
	return nil
}

func (p *DatasetProcessor) Close() errors.Error {
	if p.ModelProcessor != nil {
		if err := p.ModelProcessor.Close(); err != nil {
			return err
		}
	}
	for i := range p.SubProcessors {
		if err := p.SubProcessors[i].Close(); err != nil {
			return err
		}
	}
	return nil
}

func (p *DatasetProcessor) isIncluded(method csv.Method) (bool, errors.Error) {
	if !csv.IsMethodIncluded(method, p.TargetSet.Filter) {
		return false, nil
	}
	if p.TargetSet.SpecialOptions.MaxTokensPerOutputSequence != 0 {
		parameters, err := java.ParseParameterList(method.Parameters)
		if err == nil {
			outputSequence := p.getOutputSequence(parameters, method.ReturnType)
			tokens := metrics.TokenizeSentence(predictor.SplitMethodNameToSentence(outputSequence))
			if len(tokens) > p.TargetSet.SpecialOptions.MaxTokensPerOutputSequence {
				return false, nil
			}
		} else {
			return false, err
		}
	}
	return true, nil
}

func (p *DatasetProcessor) getOutputSequence(parameters []java.Parameter, returnType string) string {
	output := ""
	for i, par := range parameters {
		if i > 0 {
			output += ", "
		}
		output += fmt.Sprintf("%s - %s", par.Type.TypeName, par.Name)
	}
	return output + ". $ " + returnType
}

// Copies options from parent to child which should be inherited by the child
func inheritOptions(parent, child configuration.Dataset) configuration.Dataset {
	if child.SpecialOptions.DatasetSize.Training == 0 && child.SpecialOptions.DatasetSize.Evaluation == 0 {
		child.SpecialOptions.DatasetSize = parent.SpecialOptions.DatasetSize
	}
	return child
}
