package dataset

import (
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
		// TODO: Load dataset type by configuration
	} else if creator, err := getCreatorByModelType(configuration.UsedModelType(), methods, classes); err != nil {
		return err
	} else {
		return creator.Create()
	}
}

func getCreatorByModelType(modelType configuration.ModelType, methods []csv.Method, classes []csv.Class) (base.Creator, errors.Error) {
	if modelType == configuration.ReturnTypesValidator {
		return returntypesvalidation.New(methods, createPackageTree(classes)), nil
	} else {
		if len(modelType) == 0 {
			return nil, errors.New("Dataset Creation Error", "Could not create dataset: No dataset type specified.")
		}
		return nil, errors.New("Dataset Creation Error", "Could not create dataset: Unsupported dataset type ("+string(modelType)+")")
	}
}

// Creates a training and an evaluation set.
func CreateTrainingAndEvaluationSetByProcessor(methodsWithReturnTypesPath, classHierarchyPath string) errors.Error {
	if methods, classes, err := loadMethodsAndClasses(methodsWithReturnTypesPath, classHierarchyPath); err != nil {
		return err
	} else {
		processors := make(DatasetProcessors, len(configuration.Datasets()))
		tree := createPackageTree(classes)
		for i, dataset := range configuration.Datasets() {
			processors[i] = NewProcessor(dataset, configuration.ReturnTypesValidator, configuration.DatasetOutputDir(), tree)
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
	if trainer, err := getTrainerByModelType(modelType); err != nil {
		return err
	} else {
		return trainer.Train()
	}
}

func getTrainerByModelType(modelType configuration.ModelType) (base.Trainer, errors.Error) {
	switch modelType {
	case configuration.MethodGenerator:
		return methodgeneration.NewTrainer(), nil
	case configuration.ReturnTypesValidator:
		return returntypesvalidation.NewTrainer(), nil
	default:
		if len(modelType) == 0 {
			return nil, errors.New("Dataset Creation Error", "Could not create dataset: No dataset type specified.")
		}
		return nil, errors.New("Dataset Creation Error", "Could not create dataset: Unsupported dataset type ("+string(modelType)+")")
	}
}

func Evaluate() errors.Error {
	if evaluator, err := getEvaluatorByModelType(configuration.ReturnTypesValidator); err != nil {
		return err
	} else {
		return evaluator.Evaluate()
	}
}

func getEvaluatorByModelType(modelType configuration.ModelType) (base.Evaluator, errors.Error) {
	switch modelType {
	case configuration.MethodGenerator:
		return methodgeneration.NewEvaluator(), nil
	case configuration.ReturnTypesValidator:
		return returntypesvalidation.NewEvaluator(), nil
	default:
		if len(modelType) == 0 {
			return nil, errors.New("Dataset Creation Error", "Could not create dataset: No dataset type specified.")
		}
		return nil, errors.New("Dataset Creation Error", "Could not create dataset: Unsupported dataset type ("+string(modelType)+")")
	}
}
