package excelOutputter

import "returntypes-langserver/common/dataformat/csv"

type DatasetProcessor struct {
	targetSet        Dataset
	subsetProcessors []DatasetProcessor
}

func createDatasetProcessor(targetSet Dataset) DatasetProcessor {
	processor := DatasetProcessor{
		targetSet:        targetSet,
		subsetProcessors: make([]DatasetProcessor, 0, len(targetSet.Subsets)),
	}
	for _, subset := range targetSet.Subsets {
		processor.subsetProcessors = append(processor.subsetProcessors, createDatasetProcessor(subset))
	}
	return processor
}

func (p *DatasetProcessor) process(method csv.Method) {
	// If the process method is called, this means that this processor should already accept the method.
	// Add the method to the output file if one exists.

	// TODO: the passed csv method should be already PREPROCESSED !!!
	// - for example: the method name should be already splitted to a sentence, the number "2" should be already converted to "to" and so on.
	acceptedBySubsetProcessor := false
	for i := range p.subsetProcessors {
		if p.subsetProcessors[i].accepts(method) {
			p.subsetProcessors[i].process(method)
			acceptedBySubsetProcessor = true
		}
	}
	if !acceptedBySubsetProcessor {
		// TODO: add to leftover file.
	}
}

func (p *DatasetProcessor) accepts(method csv.Method) bool {
	filter := p.targetSet.Filter
	if filter.Includes != nil {
		if !filter.Includes.appliesOn(method) {
			return false
		}
	}
	if filter.Excludes != nil {
		if filter.Excludes.appliesOn(method) {
			return false
		}
	}
	return true
}
