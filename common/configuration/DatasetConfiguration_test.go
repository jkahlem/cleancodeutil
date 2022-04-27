package configuration

import (
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionsMerging(t *testing.T) {
	// given
	source := map[string]interface{}{
		"modelOptions": map[string]interface{}{
			"numOfEpochs": 3,
		},
		"specialOptions": map[string]interface{}{
			"maxTrainingRows": 3000,
		},
		"subsets": []interface{}{
			map[string]interface{}{
				"modelOptions": map[string]interface{}{
					"batchSize": 32,
				},
				"specialOptions": map[string]interface{}{
					"filterDuplicates": true,
				},
				"subsets": []interface{}{
					map[string]interface{}{
						"modelOptions": map[string]interface{}{
							"numOfEpochs": 0,
						},
						"specialOptions": map[string]interface{}{
							"maxTrainingRows": 0,
						},
					},
					map[string]interface{}{},
				},
			},
		},
	}

	// when
	var destination Dataset
	err := utils.DecodeMapToStruct(source, &destination)

	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 3, destination.ModelOptions.NumOfEpochs)
		assert.Equal(t, 3000, destination.PreprocessingOptions.MaxTrainingRows)

		// Subset should have the options from parent set and the directly set ones
		subset := destination.Subsets[0]
		assert.Equal(t, 3, subset.ModelOptions.NumOfEpochs)
		assert.Equal(t, 3000, subset.PreprocessingOptions.MaxTrainingRows)
		assert.Equal(t, 32, subset.ModelOptions.BatchSize)
		assert.Equal(t, true, subset.CreationOptions.FilterDuplicates)

		// First subset of subset should overwrite the options from first layer
		subsetOfSubset1 := subset.Subsets[0]
		assert.Equal(t, 0, subsetOfSubset1.ModelOptions.NumOfEpochs)
		assert.Equal(t, 0, subsetOfSubset1.PreprocessingOptions.MaxTrainingRows)
		assert.Equal(t, 32, subsetOfSubset1.ModelOptions.BatchSize)
		assert.Equal(t, true, subsetOfSubset1.CreationOptions.FilterDuplicates)

		// Second subset of subset should completely copy the parent options
		subsetOfSubset2 := subset.Subsets[1]
		assert.Equal(t, 3, subsetOfSubset2.ModelOptions.NumOfEpochs)
		assert.Equal(t, 3000, subsetOfSubset2.PreprocessingOptions.MaxTrainingRows)
		assert.Equal(t, 32, subsetOfSubset2.ModelOptions.BatchSize)
		assert.Equal(t, true, subsetOfSubset2.CreationOptions.FilterDuplicates)
	}
}

func TestStuff(t *testing.T) {
	set := Dataset{
		DatasetBase: DatasetBase{
			NameRaw: "a",
			ModelOptions: ModelOptions{
				NumReturnSequences: 1,
			},
		},
		CreationOptions: DatasetCreationOptions{
			MaxTokensPerOutputSequence: 12,
		},
	}
	alt := DatasetBase{
		NameRaw: "b",
	}

	altSet := set
	altSet.DatasetBase = alt

	assert.Equal(t, altSet.NameRaw, "b")
	assert.Equal(t, altSet.ModelOptions.NumReturnSequences, 1)
	assert.Equal(t, altSet.CreationOptions.MaxTokensPerOutputSequence, 300)
}
