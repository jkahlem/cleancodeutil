package dataset

import "returntypes-langserver/common/errors"

// Creates a training and an evaluation set.
func CreateTrainingAndEvaluationSet(methodsWithReturnTypesPath, classHierarchyPath string) errors.Error {
	creator := NewCreator()
	creator.CreateTrainingAndEvaluationSet(methodsWithReturnTypesPath, classHierarchyPath)
	return creator.Err()
}
