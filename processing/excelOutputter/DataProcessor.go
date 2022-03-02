package excelOutputter

import (
	"path/filepath"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/dataformat/excel"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/services/predictor"
	"strings"
)

type DatasetProcessor struct {
	targetSet         Dataset
	subsetProcessors  []DatasetProcessor
	path              string
	complementChannel *OutputChannel
	outputChannel     *OutputChannel
}

type OutputChannel struct {
	Output chan []string
	Errors chan errors.Error
}

func NewOutputChannel() *OutputChannel {
	return &OutputChannel{
		Output: make(chan []string),
		Errors: make(chan errors.Error),
	}
}

func NewDatasetProcessor(targetSet Dataset, path string) DatasetProcessor {
	processor := DatasetProcessor{
		targetSet:         targetSet,
		subsetProcessors:  make([]DatasetProcessor, 0, len(targetSet.Subsets)),
		complementChannel: NewOutputChannel(),
		outputChannel:     NewOutputChannel(),
	}
	processor.createOutputStreams(path)
	subdir := filepath.Join(path, targetSet.Name)
	for _, subset := range targetSet.Subsets {
		processor.subsetProcessors = append(processor.subsetProcessors, NewDatasetProcessor(subset, subdir))
	}
	return processor
}

func (p *DatasetProcessor) createOutputStreams(path string) errors.Error {
	if !p.targetSet.NoOutput {
		p.createOutputStream(filepath.Join(path, p.targetSet.Name), p.outputChannel)
	}
	if len(p.targetSet.ComplementFilename) > 0 {
		p.createOutputStream(filepath.Join(path, p.targetSet.ComplementFilename), p.complementChannel)
	}
	return nil
}

func (p *DatasetProcessor) createOutputStream(path string, channel *OutputChannel) {
	xlsxFilePath := path + ".xlsx"
	go func() {
		err := excel.Stream().
			FromChannel(channel.Output).
			WithColumnsFromStruct(csv.Method{}).
			InsertColumnsAt(excel.Col(7), "Project", "Notes").
			Transform(addProjectColumn).
			ToFile(xlsxFilePath)
		log.Info("Saved excel file to: %s\n", xlsxFilePath)
		channel.Errors <- err
	}()
}

func (p *DatasetProcessor) process(method csv.Method) {
	if !p.targetSet.NoOutput {
		p.outputChannel.Output <- method.ToRecord()
	}
	acceptedBySubsetProcessor := false
	for i := range p.subsetProcessors {
		if p.subsetProcessors[i].accepts(method) {
			p.subsetProcessors[i].process(method)
			acceptedBySubsetProcessor = true
		}
	}
	if !acceptedBySubsetProcessor && len(p.targetSet.ComplementFilename) > 0 {
		p.complementChannel.Output <- method.ToRecord()
	}
}

func (p *DatasetProcessor) accepts(method csv.Method) bool {
	method.MethodName = strings.ToLower(predictor.SplitMethodNameToSentence(method.MethodName))
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

func (p *DatasetProcessor) close() {
	if !p.targetSet.NoOutput {
		close(p.outputChannel.Output)
		p.checkErrors(p.outputChannel)
	}
	if len(p.targetSet.ComplementFilename) > 0 {
		close(p.complementChannel.Output)
		p.checkErrors(p.complementChannel)
	}
	for i := range p.subsetProcessors {
		p.subsetProcessors[i].close()
	}
}

func (p *DatasetProcessor) checkErrors(channel *OutputChannel) {
	if err := <-channel.Errors; err != nil {
		log.Error(err)
		log.ReportProblemWithError(err, "Could not create excel output")
	}
}
