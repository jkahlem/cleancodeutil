package dataset

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
)

// Does more specific processings like filters
type MethodProcessor interface {
	Process(csv.Method) errors.Error
	Close() errors.Error
}

// Common dataset processor which "preprocesses" the data (like applying common filters and so on)
type DatasetProcessor struct {
	ModelProcessor MethodProcessor
	TargetSet      configuration.Dataset
	SubProcessors  []DatasetProcessor
}

func NewProcessor(set configuration.Dataset) DatasetProcessor {
	processor := DatasetProcessor{
		SubProcessors:  make([]DatasetProcessor, len(set.Subsets)),
		ModelProcessor: getModelProcessorForModelType(ReturnTypesValidator),
	}

	// {Perform any initialization here}
	for i, subset := range set.Subsets {
		processor.SubProcessors[i] = NewProcessor(subset)
	}
	return processor
}

func getModelProcessorForModelType(modelType ModelType) MethodProcessor {
	// TODO
	return nil
}

func (p *DatasetProcessor) initialize() {
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
		if err := p.ModelProcessor.Process(method); err != nil {
			return err
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
