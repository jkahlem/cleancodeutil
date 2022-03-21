package returntypesvalidation

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/utils"
	"sort"
)

// Helper type for sorting dataset rows by the same type
type SortDatasetRow []csv.ReturnTypesDatasetRow

func (dataset SortDatasetRow) Len() int {
	return len(dataset)
}

func (dataset SortDatasetRow) Less(a, b int) bool {
	return dataset[a].TypeLabel < dataset[b].TypeLabel
}

func (dataset SortDatasetRow) Swap(a, b int) {
	t := dataset[b]
	dataset[b] = dataset[a]
	dataset[a] = t
}

// Splits a dataset into a training and a evaluation set by the given relation (t:e  ->  training : evaluation)
// The relation does apply on a type level, to prevent for example that the training set has 1000 boolean rows and the evaluation set has none.
func SplitToTrainingAndEvaluationSet(dataset []csv.ReturnTypesDatasetRow, proportion configuration.DatasetProportion) (trainingSet, evaluationSet []csv.ReturnTypesDatasetRow) {
	if !isValidProportion(proportion) {
		return
	}

	sort.Sort(SortDatasetRow(dataset))

	trainingSet = make([]csv.ReturnTypesDatasetRow, 0)
	evaluationSet = make([]csv.ReturnTypesDatasetRow, 0)

	for j, i := 0, 1; i < len(dataset); i++ {
		if dataset[i].TypeLabel != dataset[i-1].TypeLabel || i+1 == len(dataset) {
			tset, eset := splitRowsToTrainingAndEvaluationSet(dataset[j:i+1], proportion)
			trainingSet = append(trainingSet, tset...)
			evaluationSet = append(evaluationSet, eset...)
			j = i
		}
	}
	return
}

// Checks if the proportion is valid.
func isValidProportion(proportion configuration.DatasetProportion) bool {
	if proportion.Training < 0 || proportion.Evaluation < 0 || (proportion.Training == 0 && proportion.Evaluation == 0) {
		return false
	}
	return true
}

// Helper function that splits a part of the dataset (with same type) to the given relation.
func splitRowsToTrainingAndEvaluationSet(dataset []csv.ReturnTypesDatasetRow, proportion configuration.DatasetProportion) (trainingSet, evaluationSet []csv.ReturnTypesDatasetRow) {
	trainingSetSize, evaluationSetSize := utils.FitProportions(proportion.Training, proportion.Evaluation, len(dataset))
	trainingSet = make([]csv.ReturnTypesDatasetRow, trainingSetSize)
	evaluationSet = make([]csv.ReturnTypesDatasetRow, evaluationSetSize)
	for i := 0; i < len(dataset); i++ {
		if i < trainingSetSize {
			trainingSet[i] = dataset[i]
		} else {
			evaluationSet[i-trainingSetSize] = dataset[i]
		}
	}
	return
}
