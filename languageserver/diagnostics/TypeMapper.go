package diagnostics

import (
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/java"
	"returntypes-langserver/services/predictor"
)

// Maps types for predictions
type TypeMapper struct {
	mappings predictor.MethodTypeMap
}

// Creates a prediction map with the expected return type for each method
func (mapper *TypeMapper) CreatePredictionMappings(methods []*java.Method) (predictor.MethodTypeMap, errors.Error) {
	mapper.mappings = make(predictor.MethodTypeMap)
	mapper.createEmptyEntries(methods)
	if err := mapper.predictTypeMappings(); err != nil {
		return nil, err
	}

	return mapper.mappings, nil
}

func (mapper *TypeMapper) createEmptyEntries(methods []*java.Method) {
	for _, method := range methods {
		key := predictor.GetPredictableMethodName(method.MethodName)
		mapper.mappings[key] = ""
	}
}

func (mapper *TypeMapper) predictTypeMappings() errors.Error {
	return predictor.PredictToMap(mapper.mappings)
}
