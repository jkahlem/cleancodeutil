package methodgeneration

import (
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
)

type Trainer struct {
	Dataset configuration.Dataset
}

func NewTrainer(dataset configuration.Dataset) base.Trainer {
	return &Trainer{
		Dataset: dataset,
	}
}

func (t *Trainer) Train(path string) errors.Error {
	trainingSet, err := csv.ReadRecords(filepath.Join(path, TrainingSetFileName))
	if err != nil {
		return err
	}
	methods, err := mapToMethods(csv.UnmarshalMethodGenerationDatasetRow(trainingSet))
	if err != nil {
		return err
	}

	// Train the predictor
	return predictor.OnDataset(t.Dataset).TrainMethods(methods)
}
