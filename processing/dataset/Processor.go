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
	processor.initializeModelProcessor(modelType, tree)

	for i, subset := range set.Subsets {
		processor.SubProcessors[i] = NewProcessor(subset, modelType, filepath.Join(path, set.Name), tree)
	}
	return processor
}

func (p *DatasetProcessor) initializeModelProcessor(modelType configuration.ModelType, tree *packagetree.Tree) {
	switch modelType {
	case configuration.ReturnTypesValidator:
		p.ModelProcessor = returntypesvalidation.NewProcessor("", p.TargetSet.SpecialOptions, tree)
	case configuration.MethodGenerator:
		p.ModelProcessor = methodgeneration.NewProcessor("", p.TargetSet.SpecialOptions, tree)
	}
	// initialize output streams/folders and so on?
	//
	// Folder structure:
	// {datasetOutputDir}/{datasetName}
	//   /subsets/...
	//   /{modelType}
	//     /evaluation.csv
	//     /dataset.csv
	//     (and other data like labels.csv? at least this is the output location for the one model...)
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
