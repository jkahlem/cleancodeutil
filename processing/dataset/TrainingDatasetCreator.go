package dataset

import (
	"path/filepath"
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/processing/dataset/methodgeneration"
	"returntypes-langserver/processing/dataset/returntypesvalidation"
)

// Creates a training and an evaluation set.
func CreateTrainingAndEvaluationSet(methodsWithReturnTypesPath, classHierarchyPath string) errors.Error {
	if methods, classes, err := loadMethodsAndClasses(methodsWithReturnTypesPath, classHierarchyPath); err != nil {
		return err
	} else {
		processors := make(DatasetProcessors, len(configuration.Datasets()))
		tree := createPackageTree(classes)
		for i, dataset := range configuration.Datasets() {
			if processor, err := NewProcessor(dataset, configuration.ReturnTypesValidator, configuration.DatasetOutputDir(), tree); err != nil {
				return err
			} else {
				processors[i] = processor
			}
		}

		for _, method := range methods {
			if err := processors.Process(method); err != nil {
				return err
			}
		}
		return processors.Close()
	}
}

// Loads the methods and class data into the creator
func loadMethodsAndClasses(methodsWithReturnTypesPath, classHierarchyPath string) ([]csv.Method, []csv.Class, errors.Error) {
	methodsRecords, err := csv.ReadRecords(methodsWithReturnTypesPath)
	if err != nil {
		return nil, nil, err
	}

	classesRecords, err := csv.ReadRecords(classHierarchyPath)
	if err != nil {
		return nil, nil, err
	}

	for _, defaultLibrary := range configuration.DefaultLibraries() {
		defaultClassesRecords, err := csv.ReadRecords(defaultLibrary)
		if err != nil {
			return nil, nil, err
		}
		classesRecords = append(classesRecords, defaultClassesRecords...)
	}

	return csv.UnmarshalMethod(methodsRecords), csv.UnmarshalClass(classesRecords), nil
}

func createPackageTree(classes []csv.Class) *packagetree.Tree {
	tree := packagetree.New()
	java.FillPackageTreeByCsvClassNodes(&tree, classes)
	return &tree
}

func Train(modelType configuration.ModelType) errors.Error {
	return trainDatasets(modelType, configuration.DatasetOutputDir(), configuration.Datasets())
}

func trainDatasets(modelType configuration.ModelType, path string, datasets []configuration.Dataset) errors.Error {
	for _, dataset := range datasets {
		path := filepath.Join(path, dataset.Name)
		if err := train(modelType, path, dataset); err != nil {
			return err
		}
		if err := trainDatasets(modelType, path, dataset.Subsets); err != nil {
			return err
		}
	}
	return nil
}

func train(modelType configuration.ModelType, path string, dataset configuration.Dataset) errors.Error {
	// Leaving this function frees the memory occupied by the trainer, therefore put into a seperate function
	if trainer, err := getTrainerByModelType(modelType, dataset); err != nil {
		return err
	} else if err := trainer.Train(path); err != nil {
		return err
	}
	return nil
}

func getTrainerByModelType(modelType configuration.ModelType, dataset configuration.Dataset) (base.Trainer, errors.Error) {
	switch modelType {
	case configuration.MethodGenerator:
		return methodgeneration.NewTrainer(dataset), nil
	case configuration.ReturnTypesValidator:
		return returntypesvalidation.NewTrainer(dataset), nil
	default:
		if len(modelType) == 0 {
			return nil, errors.New("Dataset Creation Error", "Could not create dataset: No dataset type specified.")
		}
		return nil, errors.New("Dataset Creation Error", "Could not create dataset: Unsupported dataset type ("+string(modelType)+")")
	}
}

func Evaluate(modelType configuration.ModelType) errors.Error {
	return evaluateDatasets(modelType, configuration.DatasetOutputDir(), configuration.Datasets())
}

func evaluateDatasets(modelType configuration.ModelType, path string, datasets []configuration.Dataset) errors.Error {
	for _, dataset := range datasets {
		path := filepath.Join(path, dataset.Name)
		if err := evaluate(modelType, path, dataset); err != nil {
			return err
		}
		if err := evaluateDatasets(modelType, path, dataset.Subsets); err != nil {
			return err
		}
	}
	return nil
}

func evaluate(modelType configuration.ModelType, path string, dataset configuration.Dataset) errors.Error {
	if evaluator, err := getEvaluatorByModelType(modelType, dataset); err != nil {
		return err
	} else if err := evaluator.Evaluate(path); err != nil {
		return err
	}
	return nil
}

func getEvaluatorByModelType(modelType configuration.ModelType, dataset configuration.Dataset) (base.Evaluator, errors.Error) {
	switch modelType {
	case configuration.MethodGenerator:
		return methodgeneration.NewEvaluator(), nil
	case configuration.ReturnTypesValidator:
		return returntypesvalidation.NewEvaluator(dataset), nil
	default:
		if len(modelType) == 0 {
			return nil, errors.New("Dataset Creation Error", "Could not create dataset: No dataset type specified.")
		}
		return nil, errors.New("Dataset Creation Error", "Could not create dataset: Unsupported dataset type ("+string(modelType)+")")
	}
}
