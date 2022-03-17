package returntypesvalidation

import (
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
)

type Trainer struct {
	labels      [][]string
	trainingSet []csv.ReturnTypesDatasetRow
	Dataset     configuration.Dataset
	err         errors.Error
}

func NewTrainer(dataset configuration.Dataset) base.Trainer {
	return &Trainer{
		Dataset: dataset,
	}
}

func (t *Trainer) Train(path string) errors.Error {
	if exists, err := predictor.OnDataset(t.Dataset).ModelExists(predictor.ReturnTypesPrediction); err != nil {
		return err
	} else if exists {
		// Skip because the model is already trained
		return nil
	}
	if err := t.loadData(path); err != nil {
		return err
	}

	// Train the predictor
	return predictor.OnDataset(t.Dataset).TrainReturnTypes(mapToMethod(t.trainingSet), t.labels)
}

func (t *Trainer) loadData(path string) errors.Error {
	// Load csv data
	t.labels = t.loadRecords(filepath.Join(path, LabelSetFileName))
	t.trainingSet = csv.UnmarshalReturnTypesDatasetRow(t.loadRecords(filepath.Join(path, TrainingSetFileName)))
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
