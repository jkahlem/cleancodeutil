package methodgeneration

import (
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/metrics"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
	"sort"
	"strings"
)

const GeneratedMethodsFile = "methodgeneration_generatedMethods.csv"
const ScoreOutputFile = "methodgeneration_scores.txt"
const ExampleOutputFile = "methodgeneration_examples.txt"
const ResultOutputFile = "methodgeneration_result.xlsx"

type Evaluator struct {
	Dataset      configuration.Dataset
	resultWriter *EvaluationResultWriter
}

type Method struct {
	Name                string
	ExpectedDefinition  *metrics.Sentence
	GeneratedDefinition *metrics.Sentence
	Method              predictor.Method
}

func NewEvaluator(dataset configuration.Dataset) base.Evaluator {
	return &Evaluator{
		Dataset: dataset,
	}
}

func (e *Evaluator) Evaluate(path string) errors.Error {
	log.Info("Evaluate dataset %s\n", e.Dataset.Name())
	if err := e.evaluateCheckpoint(path, ""); err != nil {
		return err
	}
	if e.isCheckpointEvaluationActive() {
		log.Info("Evaluate checkpoints for each %s\n", e.Dataset.EvaluateOn)
		checkpoints, err := predictor.OnDataset(e.Dataset).GetCheckpoints(predictor.MethodGenerator)
		if err != nil {
			return err
		}
		checkpoints = e.reduceCheckpoints(checkpoints)
		for _, checkpoint := range checkpoints {
			if e.Dataset.EvaluateOn == configuration.Step || strings.Contains(checkpoint, "epoch") {
				log.Info("Evaluate checkpoint: %s\n", checkpoint)
				if err := e.evaluateCheckpoint(path, checkpoint); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Reduces the checkpoint array to contain only one checkpoint for the same step. It happens, that the used transformer library
// saves checkpoints for steps and epochs at the same time separately (e.g. checkpoint-1800 and checkpoint-1800-epoch-1) if they
// overlap. This function will check for these cases and removes the entries without the epoch.
func (e *Evaluator) reduceCheckpoints(checkpoints []string) []string {
	checkpointMap := make(map[string]string)
	for _, checkpoint := range checkpoints {
		if !strings.HasPrefix(checkpoint, "checkpoint-") {
			continue
		}
		splitted := strings.Split(checkpoint, "-")
		stepCount := splitted[1]
		if val, ok := checkpointMap[stepCount]; !ok || len(val) < len(checkpoint) {
			checkpointMap[stepCount] = checkpoint
		}
	}

	output := make([]string, 0, len(checkpointMap))
	for _, val := range checkpointMap {
		output = append(output, val)
	}
	sort.Strings(output)
	return output
}

func (e *Evaluator) isCheckpointEvaluationActive() bool {
	return e.Dataset.EvaluateOn == configuration.Epoch || e.Dataset.EvaluateOn == configuration.Step
}

func (e *Evaluator) evaluateCheckpoint(path, checkpoint string) errors.Error {
	checkpointPath := path
	if checkpoint != "" {
		checkpointPath = filepath.Join(path, checkpoint)
	}
	if e.isEvaluationResultPresent(checkpointPath) {
		return nil
	}
	if writer, err := NewResultWriter(filepath.Join(checkpointPath, e.Dataset.Name()+ResultOutputFile)); err != nil {
		return err
	} else {
		e.resultWriter = writer
	}
	evalset, err := e.getEvaluationSetConfig()
	if err != nil {
		return err
	}

	if err := e.evaluateMethods(path, checkpoint, evalset); err != nil {
		return err
	}
	if err := e.writeScoreOutput(checkpointPath, evalset); err != nil {
		return err
	}
	if err := e.writeExampleOutput(checkpointPath, evalset); err != nil {
		return err
	}
	if err := e.resultWriter.Close(); err != nil {
		return err
	}
	return nil
}

func (e *Evaluator) loadEvaluationSet(path string) ([]predictor.Method, errors.Error) {
	evaluationSet, err := csv.NewFileReader(path, EvaluationSetFileName).ReadMethodGenerationDatasetRowRecords()
	if err != nil {
		return nil, err
	}
	limit := e.Dataset.PreprocessingOptions.MaxEvaluationRows
	if limit <= 0 || limit > len(evaluationSet) {
		limit = len(evaluationSet)
	}
	methods, err := mapToMethods(evaluationSet[:limit])
	if err != nil {
		return nil, err
	}
	return methods, nil
}

func (e *Evaluator) generateMethodDefinitions(methods []predictor.Method, checkpoint string) ([]Method, errors.Error) {
	contexts := make([]predictor.MethodContext, len(methods))
	for i, method := range methods {
		contexts[i] = method.Context
	}

	predicted, err := predictor.OnCheckpoint(e.Dataset, checkpoint).GenerateMethods(contexts)
	if len(predicted) != len(methods) {
		return nil, errors.New("Predictor error", "Expected %d methods to be generated but got %d.", len(methods), len(predicted))
	}
	e.formatMethods(contexts, predicted)

	outputMethods := make([]Method, len(predicted))
	for i := range predicted {
		outputMethods[i] = e.parseOutputToMethod(predictor.Method{
			Context: contexts[i],
			Values:  predicted[i][0], //TODO Multiple Suggestions
		}, methods[i].Values)
	}
	return outputMethods, err
}

func (e *Evaluator) formatMethods(contexts []predictor.MethodContext, values [][]predictor.MethodValues) {
	options := configuration.SentenceFormattingOptions{
		MethodName:    true,
		ParameterName: true,
		TypeName:      true,
	}
	predictor.FormatContexts(contexts, options)
	predictor.FormatValues(values, options)
}

func (e *Evaluator) parseOutputToMethod(method predictor.Method, expectedValues predictor.MethodValues) Method {
	return Method{
		Name:                string(method.Context.MethodName),
		ExpectedDefinition:  e.joinParameters(expectedValues),
		GeneratedDefinition: e.joinParameters(method.Values),
		Method:              method,
	}
}

func (e *Evaluator) joinParameters(values predictor.MethodValues) *metrics.Sentence {
	str := ""
	for i, par := range values.Parameters {
		if i > 0 {
			str += " [psp] "
		}
		par.Name = e.formatString(par.Name)
		par.Type = e.formatString(par.Type)
		str += par.String()
	}
	str += " [rsp] " + e.formatString(values.ReturnType)
	return metrics.NewSentence(str)
}

func (e *Evaluator) formatString(str string) string {
	return strings.ToLower(predictor.SplitMethodNameToSentence(str))
}

func (e *Evaluator) getEvaluationSetConfig() (*EvaluationSet, errors.Error) {
	set, err := e.buildEvaluationSet(configuration.EvaluationSet{
		Subsets: configuration.EvaluationSubsets(),
	})
	return &set, err
}

func (e *Evaluator) buildEvaluationSet(setConfiguration configuration.EvaluationSet) (EvaluationSet, errors.Error) {
	set := EvaluationSet{
		Subsets:  make([]EvaluationSet, 0),
		Filter:   setConfiguration.Filter,
		Name:     setConfiguration.Name,
		Examples: setConfiguration.Examples,
	}
	if err := set.initRater(setConfiguration.Metrics); err != nil {
		return set, err
	}

	for _, subsetConfig := range setConfiguration.Subsets {
		subset, err := e.buildEvaluationSet(subsetConfig)
		if err != nil {
			return set, err
		}
		set.Subsets = append(set.Subsets, subset)
	}
	return set, nil
}

func (e *Evaluator) isEvaluationResultPresent(path string) bool {
	return utils.FileExists(filepath.Join(path, e.Dataset.Name()+ResultOutputFile))
}

func (e *Evaluator) evaluateMethods(path, checkpoint string, evalset *EvaluationSet) errors.Error {
	methods, err := e.getGeneratedMethodsForEvaluationSet(path, checkpoint)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	for _, m := range methods {
		evalset.AddMethod(m)
	}
	return e.resultWriter.WriteMethods(methods)
}

func (e *Evaluator) getGeneratedMethodsForEvaluationSet(path, checkpoint string) ([]Method, errors.Error) {
	set, err := e.loadEvaluationSet(path)
	if err != nil {
		return nil, err
	}

	methods, err := e.generateMethodDefinitions(set, checkpoint)
	if err != nil {
		return nil, err
	}
	return methods, nil
}

func (e *Evaluator) writeScoreOutput(path string, evalset *EvaluationSet) errors.Error {
	return e.resultWriter.WriteScores(evalset)
}

func (e *Evaluator) writeExampleOutput(path string, evalset *EvaluationSet) errors.Error {
	examples := evalset.GetExampleMethods()
	if len(examples) == 0 {
		return nil
	}
	generated, err := predictor.OnDataset(e.Dataset).GenerateMethods(examples)
	if err != nil {
		return err
	}

	return e.resultWriter.WriteExamples(examples, generated)
}
