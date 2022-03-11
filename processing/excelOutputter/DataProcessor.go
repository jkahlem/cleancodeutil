package excelOutputter

import (
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/dataformat/excel"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/services/predictor"
	"strings"
)

type DatasetProcessor struct {
	targetSet         configuration.ExcelSet
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

func NewDatasetProcessor(targetSet configuration.ExcelSet, path string) DatasetProcessor {
	return newDatasetProcessorInternal(targetSet, path, false)
}

func newDatasetProcessorInternal(targetSet configuration.ExcelSet, path string, needAllSubprocessors bool) DatasetProcessor {
	processor := DatasetProcessor{
		targetSet:        targetSet,
		subsetProcessors: make([]DatasetProcessor, 0, len(targetSet.Subsets)),
	}
	processor.createOutputStreams(path)
	subdir := filepath.Join(path, targetSet.Name)
	if !needAllSubprocessors {
		needAllSubprocessors = processor.complementChannel != nil
	}
	for _, subset := range targetSet.Subsets {
		subsetProcessor := newDatasetProcessorInternal(subset, subdir, needAllSubprocessors)
		if !needAllSubprocessors && !subsetProcessor.hasOutput() {
			// filter out subset processors which do not create output so they don't take up resources
			continue
		}
		processor.subsetProcessors = append(processor.subsetProcessors, subsetProcessor)
	}
	return processor
}

func (p *DatasetProcessor) hasOutput() bool {
	for i := range p.subsetProcessors {
		if p.subsetProcessors[i].hasOutput() {
			return true
		}
	}
	return p.outputChannel != nil || p.complementChannel != nil
}

func (p *DatasetProcessor) createOutputStreams(path string) errors.Error {
	if !p.targetSet.NoOutput {
		p.outputChannel = NewOutputChannel()
		if ok := p.createOutputStream(filepath.Join(path, p.targetSet.Name), p.outputChannel); !ok {
			p.outputChannel = nil
		}
	}
	if p.targetSet.ComplementFilename != "" {
		p.complementChannel = NewOutputChannel()
		if ok := p.createOutputStream(filepath.Join(path, p.targetSet.ComplementFilename), p.complementChannel); !ok {
			p.complementChannel = nil
		}
	}
	return nil
}

func (p *DatasetProcessor) createOutputStream(path string, channel *OutputChannel) bool {
	xlsxFilePath := path + ".xlsx"
	if utils.FileExists(xlsxFilePath) {
		return false
	}
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
	return true
}

func (p *DatasetProcessor) process(method csv.Method) {
	if p.outputChannel != nil {
		p.outputChannel.Output <- method.ToRecord()
	}
	acceptedBySubsetProcessor := false
	for i := range p.subsetProcessors {
		if p.subsetProcessors[i].accepts(method) {
			p.subsetProcessors[i].process(method)
			acceptedBySubsetProcessor = true
		}
	}
	if !acceptedBySubsetProcessor && p.complementChannel != nil {
		p.complementChannel.Output <- method.ToRecord()
	}
}

func (p *DatasetProcessor) accepts(method csv.Method) bool {
	method.MethodName = strings.ToLower(predictor.SplitMethodNameToSentence(method.MethodName))
	return csv.IsMethodIncluded(method, p.targetSet.Filter)
}

func (p *DatasetProcessor) close() {
	if p.outputChannel != nil {
		close(p.outputChannel.Output)
		p.checkErrors(p.outputChannel)
	}
	if p.complementChannel != nil {
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
