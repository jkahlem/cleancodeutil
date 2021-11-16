package dataset

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/csv"
	"sort"
)

// Helper type for sorting dataset rows by the same type
type SortDatasetRow []csv.DatasetRow

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
func SplitToTrainingAndEvaluationSet(dataset []csv.DatasetRow, proportion configuration.DatasetProportion) (trainingSet, evaluationSet []csv.DatasetRow) {
	if !isValidProportion(proportion) {
		return
	}

	sort.Sort(SortDatasetRow(dataset))

	trainingSet = make([]csv.DatasetRow, 0)
	evaluationSet = make([]csv.DatasetRow, 0)

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
func splitRowsToTrainingAndEvaluationSet(dataset []csv.DatasetRow, proportion configuration.DatasetProportion) (trainingSet, evaluationSet []csv.DatasetRow) {
	trainingSetSize, evaluationSetSize := calculateDatasetSizes(len(dataset), proportion)
	trainingSet = make([]csv.DatasetRow, trainingSetSize)
	evaluationSet = make([]csv.DatasetRow, evaluationSetSize)
	for i := 0; i < len(dataset); i++ {
		if i < trainingSetSize {
			trainingSet[i] = dataset[i]
		} else {
			evaluationSet[i-trainingSetSize] = dataset[i]
		}
	}
	return
}

// Calculates the size of the training set and the evaluation set using the given proportion settings for the given amount of rows
func calculateDatasetSizes(rowsCount int, proportion configuration.DatasetProportion) (trainingSetSize, evaluationSetSize int) {
	proportionSum := proportion.Training + proportion.Evaluation
	relativeTrainingSetSize := proportion.Training / proportionSum
	trainingSetSize = int(float64(rowsCount) * relativeTrainingSetSize)
	evaluationSetSize = rowsCount - trainingSetSize
	return
}
