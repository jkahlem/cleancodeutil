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

type ModelType string

const (
	MethodGenerator      ModelType = "MethodGenerator"
	ReturnTypesValidator ModelType = "ReturnTypesValidator"
)

// Creates a training and an evaluation set.
func CreateTrainingAndEvaluationSet(methodsWithReturnTypesPath, classHierarchyPath string) errors.Error {
	if methods, classes, err := loadMethodsAndClasses(methodsWithReturnTypesPath, classHierarchyPath); err != nil {
		return err
		// TODO: Load dataset type by configuration
	} else if creator, err := getCreatorByModelType(MethodGenerator, methods, classes); err != nil {
		return err
	} else {
		return creator.Create()
	}
}

func getCreatorByModelType(modelType ModelType, methods []csv.Method, classes []csv.Class) (base.Creator, errors.Error) {
	if modelType == MethodGenerator {
		return methodgeneration.New(methods, createPackageTree(classes)), nil
	} else if modelType == ReturnTypesValidator {
		return returntypesvalidation.New(methods, createPackageTree(classes)), nil
	} else {
		if len(modelType) == 0 {
			return nil, errors.New("Dataset Creation Error", "Could not create dataset: No dataset type specified.")
		}
		return nil, errors.New("Dataset Creation Error", "Could not create dataset: Unsupported dataset type ("+string(modelType)+")")
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

func Train(modelType ModelType) errors.Error {
	if trainer, err := getTrainerByModelType(modelType); err != nil {
		return err
	} else {
		return trainer.Train()
	}
}

func getTrainerByModelType(modelType ModelType) (base.Trainer, errors.Error) {
	if modelType == MethodGenerator {
		return methodgeneration.NewTrainer(), nil
	} else if modelType == ReturnTypesValidator {
		return returntypesvalidation.NewTrainer(), nil
	} else {
		if len(modelType) == 0 {
			return nil, errors.New("Dataset Creation Error", "Could not create dataset: No dataset type specified.")
		}
		return nil, errors.New("Dataset Creation Error", "Could not create dataset: Unsupported dataset type ("+string(modelType)+")")
	}
}

func Evaluate() errors.Error {
	if evaluator, err := getEvaluatorByModelType(ReturnTypesValidator); err != nil {
		return err
	} else {
		return evaluator.Evaluate()
	}
}

func getEvaluatorByModelType(modelType ModelType) (base.Evaluator, errors.Error) {
	if modelType == MethodGenerator {
		return methodgeneration.NewEvaluator(), nil
	} else if modelType == ReturnTypesValidator {
		return returntypesvalidation.NewEvaluator(), nil
	} else {
		if len(modelType) == 0 {
			return nil, errors.New("Dataset Creation Error", "Could not create dataset: No dataset type specified.")
		}
		return nil, errors.New("Dataset Creation Error", "Could not create dataset: Unsupported dataset type ("+string(modelType)+")")
	}
}
