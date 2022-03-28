package returntypesvalidation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
)

type Evaluator struct {
	labels        [][]string
	evaluationSet []csv.ReturnTypesDatasetRow
	Dataset       configuration.Dataset
}

func NewEvaluator(dataset configuration.Dataset) base.Evaluator {
	return &Evaluator{Dataset: dataset}
}

func (e *Evaluator) Evaluate(path string) errors.Error {
	if e.isEvaluationResultPresent() {
		return nil
	}
	if err := e.loadData(path); err != nil {
		return err
	}
	if result, err := predictor.OnDataset(e.Dataset).EvaluateReturnTypes(mapToMethod(e.evaluationSet), e.labels); err != nil {
		return err
	} else {
		log.Info("Evaluation result:\n- Accuracy Score: %g\n- Eval loss: %g\n- F1 Score: %g\n- MCC: %g\n", result.AccScore, result.EvalLoss, result.F1Score, result.MCC)
		return e.saveEvaluationResult(result)
	}
}

func (e *Evaluator) loadData(path string) errors.Error {
	// Load csv data
	if labels, err := csv.NewFileReader(filepath.Join(path, LabelSetFileName)).ReadAllRecords(); err != nil {
		return err
	} else {
		e.labels = labels
	}
	if evaluationSet, err := csv.NewFileReader(filepath.Join(path, EvaluationSetFileName)).ReadReturnTypesDatasetRowRecords(); err != nil {
		return err
	} else {
		limit := e.Dataset.SpecialOptions.MaxEvaluationRows
		if limit <= 0 || limit > len(evaluationSet) {
			limit = len(evaluationSet)
		}
		e.evaluationSet = evaluationSet[:limit]
	}
	return nil
}

func (e *Evaluator) saveEvaluationResult(msg predictor.Evaluation) errors.Error {
	// Write the evaluation result in a json file
	if file, err := os.Create(configuration.EvaluationResultOutputPath()); err != nil {
		return errors.Wrap(err, "Error", "Could not save evaluation result")
	} else {
		defer file.Close()
		if err := json.NewEncoder(file).Encode(msg); err != nil {
			return errors.Wrap(err, "Error", "Could not save evaluation result")
		}
	}
	return nil
}

func (e *Evaluator) isEvaluationResultPresent() bool {
	return utils.FileExists(configuration.EvaluationResultOutputPath())
}
