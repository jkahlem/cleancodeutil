package returntypesvalidation

import (
	"encoding/json"
	"os"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
)

type Trainer struct {
	labels        [][]string
	trainingSet   [][]string
	evaluationSet [][]string
	err           errors.Error
}

func NewTrainer() base.Trainer {
	return &Trainer{}
}

func (t *Trainer) Train() errors.Error {
	if err := t.loadData(); err != nil {
		return err
	}

	// Train the predictor
	if msg, err := predictor.TrainReturnTypes(t.labels, t.trainingSet, t.evaluationSet); err != nil {
		return err
	} else {
		log.Info("Evaluation result:\n- Accuracy Score: %g\n- Eval loss: %g\n- F1 Score: %g\n- MCC: %g\n", msg.AccScore, msg.EvalLoss, msg.F1Score, msg.MCC)
		return t.saveEvaluationResult(msg)
	}
}

func (t *Trainer) loadData() errors.Error {
	// Load csv data
	t.labels = t.loadRecords(configuration.DatasetLabelsOutputPath())
	t.trainingSet = t.loadRecords(configuration.TrainingSetOutputPath())
	t.evaluationSet = t.loadRecords(configuration.EvaluationSetOutputPath())
	err := t.err
	t.err = nil
	return err
}

func (t *Trainer) loadRecords(path string) [][]string {
	if t.err != nil {
		return nil
	}

	result, err := csv.ReadRecords(path)
	t.err = err
	return result
}

func (t *Trainer) saveEvaluationResult(msg predictor.Evaluation) errors.Error {
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
