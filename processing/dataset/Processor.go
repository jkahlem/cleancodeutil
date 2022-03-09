package dataset

import (
	"path/filepath"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/processing/dataset/methodgeneration"
	"returntypes-langserver/processing/dataset/returntypesvalidation"
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

func NewProcessor(set configuration.Dataset, modelType configuration.ModelType, path string, tree *packagetree.Tree) DatasetProcessor {
	processor := DatasetProcessor{
		TargetSet:     set,
		SubProcessors: make([]DatasetProcessor, len(set.Subsets)),
	}
	processor.initializeModelProcessor(modelType, path, tree)

	for i, subset := range set.Subsets {
		processor.SubProcessors[i] = NewProcessor(inheritOptions(set, subset), modelType, filepath.Join(path, set.Name), tree)
	}
	return processor
}

func (p *DatasetProcessor) initializeModelProcessor(modelType configuration.ModelType, path string, tree *packagetree.Tree) {
	switch modelType {
	case configuration.ReturnTypesValidator:
		p.ModelProcessor = returntypesvalidation.NewProcessor(path, p.TargetSet.SpecialOptions, tree)
	case configuration.MethodGenerator:
		p.ModelProcessor = methodgeneration.NewProcessor(path, p.TargetSet.SpecialOptions, tree)
	}
}

func (p *DatasetProcessor) Process(method csv.Method) errors.Error {
	if !p.isIncluded(method) {
		return nil
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

func (p *DatasetProcessor) isIncluded(method csv.Method) bool {
	if !csv.IsMethodIncluded(method, p.TargetSet.Filter) {
		return false
	} else if len(method.MethodName) < p.TargetSet.SpecialOptions.MinMethodNameLength {
		return false
	}
	return true
}

// Copies options from parent to child which should be inherited by the child
func inheritOptions(parent, child configuration.Dataset) configuration.Dataset {
	if child.SpecialOptions.DatasetSize.Training == 0 && child.SpecialOptions.DatasetSize.Evaluation == 0 {
		child.SpecialOptions.DatasetSize = parent.SpecialOptions.DatasetSize
	}
	return child
}
