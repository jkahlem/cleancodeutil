package returntypesvalidation

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
)

type Trainer struct {
	labels      [][]string
	trainingSet []csv.ReturnTypesDatasetRow
	Dataset     configuration.Dataset
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
		log.Info("[Returntypes validation] Skip training of dataset '%s' because it is already trained.", t.Dataset.Name())
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
	if labels, err := csv.NewFileReader(path, LabelSetFileName).ReadAllRecords(); err != nil {
		return err
	} else {
		t.labels = labels
	}

	if trainingSet, err := csv.NewFileReader(path, TrainingSetFileName).ReadReturnTypesDatasetRowRecords(); err != nil {
		return err
	} else {
		limit := t.Dataset.SpecialOptions.MaxTrainingRows
		if limit <= 0 || limit > len(trainingSet) {
			limit = len(trainingSet)
		}
		t.trainingSet = trainingSet[:limit]
	}
	return nil
}
