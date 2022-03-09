package methodgeneration

import (
	"path/filepath"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
)

type Trainer struct{}

func NewTrainer() base.Trainer {
	return &Trainer{}
}

func (t *Trainer) Train(path string) errors.Error {
	trainingSet, err := csv.ReadRecords(filepath.Join(path, TrainingSetFileName))
	if err != nil {
		return err
	}

	// Train the predictor
	if msg, err := predictor.TrainMethods(trainingSet[0:40000], nil); err != nil {
		return err
	} else {
		log.Info("Evaluation result:\n- Accuracy Score: %g\n- Eval loss: %g\n- F1 Score: %g\n- MCC: %g\n", msg.AccScore, msg.EvalLoss, msg.F1Score, msg.MCC)
		return nil
	}
}
