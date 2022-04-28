package returntypesvalidation

import (
	"fmt"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/services/predictor"
)

func mapToMethod(rows []csv.ReturnTypesDatasetRow) []predictor.Method {
	outputRows := make([]predictor.Method, len(rows))
	for i := range rows {
		outputRows[i] = predictor.Method{
			Context: predictor.MethodContext{
				MethodName: rows[i].MethodName,
			},
			Values: predictor.MethodValues{
				ReturnType: fmt.Sprintf("%d", rows[i].TypeLabel),
			},
		}
	}
	return outputRows
}
