package methodgeneration

import (
	"fmt"
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
)

const GeneratedMethodsFile = "methodgeneration_generatedMethods.csv"
const ScoreOutputFile = "methodgeneration_scores.txt"
const ExampleOutputFile = "methodgeneration_examples.txt"

type Evaluator struct {
	Dataset configuration.Dataset
}

type Method struct {
	Name                string
	ExpectedDefinition  string
	GeneratedDefinition string
}

func NewEvaluator(dataset configuration.Dataset) base.Evaluator {
	return &Evaluator{
		Dataset: dataset,
	}
}

func (e *Evaluator) Evaluate(path string) errors.Error {
	if e.isEvaluationResultPresent(path) {
		return nil
	}
	evalset := e.getEvaluationSetConfig()

	if err := e.evaluateMethods(path, evalset); err != nil {
		return err
	}
	if err := e.writeScoreOutput(path, evalset); err != nil {
		return err
	}
	if err := e.writeExampleOutput(path, evalset); err != nil {
		return err
	}
	return nil
}

func (e *Evaluator) loadEvaluationSet(path string) ([]predictor.Method, errors.Error) {
	evaluationSet, err := csv.ReadRecords(filepath.Join(path, TrainingSetFileName))
	if err != nil {
		return nil, err
	}
	limit := e.Dataset.SpecialOptions.MaxEvaluationRows
	if limit <= 0 || limit > len(evaluationSet) {
		limit = len(evaluationSet)
	}
	methods, err := mapToMethods(csv.UnmarshalMethodGenerationDatasetRow(evaluationSet[:limit]))
	if err != nil {
		return nil, err
	}
	return methods, nil
}

func (e *Evaluator) generateMethodDefinitions(methods []predictor.Method) ([]Method, errors.Error) {
	contexts := make([]predictor.MethodContext, len(methods))
	for i, method := range methods {
		contexts[i] = method.Context
	}

	predicted, err := predictor.OnDataset(e.Dataset).GenerateMethods(contexts)
	if len(predicted) != len(methods) {
		return nil, errors.New("Predictor error", fmt.Sprintf("Expected %d methods to be generated but got %d.", len(methods), len(predicted)))
	}

	outputMethods := make([]Method, len(predicted))
	for i := range predicted {
		outputMethods[i] = e.parseOutputToMethod(predictor.Method{
			Context: contexts[i],
			Values:  predicted[i],
		}, methods[i].Values)
	}
	return outputMethods, err
}

func (e *Evaluator) parseOutputToMethod(method predictor.Method, expectedValues predictor.MethodValues) Method {
	return Method{
		Name:                string(method.Context.MethodName),
		ExpectedDefinition:  e.joinParameters(expectedValues.Parameters),
		GeneratedDefinition: e.joinParameters(method.Values.Parameters),
	}
}

func (e *Evaluator) joinParameters(parameters []predictor.Parameter) string {
	if len(parameters) == 0 {
		return "void."
	}
	joined := ""
	for i := range parameters {
		joined += fmt.Sprintf("%s %s", parameters[i].Type, parameters[i].Name)
	}
	return joined
}

func (e *Evaluator) getEvaluationSetConfig() *EvaluationSet {
	set := e.buildEvaluationSet(configuration.EvaluationSet{
		Subsets: configuration.EvaluationSubsets(),
	})
	return &set
}

func (e *Evaluator) buildEvaluationSet(setConfiguration configuration.EvaluationSet) EvaluationSet {
	set := EvaluationSet{
		Subsets:  make([]EvaluationSet, 0),
		Filter:   setConfiguration.Filter,
		Name:     setConfiguration.Name,
		Examples: setConfiguration.Examples,
	}
	set.initRater(setConfiguration.Metrics)

	for _, subset := range setConfiguration.Subsets {
		set.Subsets = append(set.Subsets, e.buildEvaluationSet(subset))
	}
	return set
}

func (e *Evaluator) isEvaluationResultPresent(path string) bool {
	return utils.FileExists(filepath.Join(path, GeneratedMethodsFile))
}

func (e *Evaluator) evaluateMethods(path string, evalset *EvaluationSet) errors.Error {
	methods, err := e.getGeneratedMethodsForEvaluationSet(path)
	if err != nil {
		return err
	}

	generatedMethodsFile, err := utils.CreateFile(filepath.Join(path, GeneratedMethodsFile))
	if err != nil {
		return err
	}
	defer generatedMethodsFile.Close()

	for _, m := range methods {
		evalset.AddMethod(m)
		fmt.Fprintf(generatedMethodsFile, "%s;%s;%s\n", m.Name, m.ExpectedDefinition, m.GeneratedDefinition)
	}
	return nil
}

func (e *Evaluator) getGeneratedMethodsForEvaluationSet(path string) ([]Method, errors.Error) {
	set, err := e.loadEvaluationSet(path)
	if err != nil {
		return nil, err
	}

	methods, err := e.generateMethodDefinitions(set)
	if err != nil {
		return nil, err
	}
	return methods, nil
}

func (e *Evaluator) writeScoreOutput(path string, evalset *EvaluationSet) errors.Error {
	scoreFile, err := utils.CreateFile(filepath.Join(path, ScoreOutputFile))
	if err != nil {
		return err
	}
	defer scoreFile.Close()
	fmt.Fprintln(scoreFile, "Print evaluations for: ", path)
	evalset.PrintScore(scoreFile)
	return nil
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

	file, err := utils.CreateFile(filepath.Join(path, ExampleOutputFile))
	if err != nil {
		return err
	}
	defer file.Close()

	for i, methodValue := range generated {
		fmt.Fprintf(file, "%s -> %s\n", examples[i], methodValue)
	}

	return nil
}
