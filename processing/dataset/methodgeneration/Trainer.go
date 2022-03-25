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
	if exists, err := predictor.OnDataset(t.Dataset).ModelExists(predictor.MethodGenerator); err != nil {
		return err
	} else if exists {
		// Skip because the model is already trained
		return nil
	}
	trainingSet, err := csv.NewFileReader(filepath.Join(path, TrainingSetFileName)).ReadMethodGenerationDatasetRowRecords()
	if err != nil {
		return err
	}
	limit := t.Dataset.SpecialOptions.MaxTrainingRows
	if limit <= 0 || limit > len(trainingSet) {
		limit = len(trainingSet)
	}
	methods, err := mapToMethods(trainingSet[:limit])
	if err != nil {
		return err
	}

	// Train the predictor
	return predictor.OnDataset(t.Dataset).TrainMethods(methods[:4000])
}
