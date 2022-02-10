package base

import (
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/processing/typeclasses"
	"returntypes-langserver/services/predictor"
)

type methodNode struct {
	name        string
	returnTypes []returnType
}

type returnType struct {
	name  string
	count int
}

type summarizedMethodsMap map[predictor.PredictableMethodName]*methodNode

// Counts one up on the given return type for this method
func (node *methodNode) CountUpReturnType(name string) {
	if rType := node.findReturnType(name); rType != nil {
		rType.count++
	} else {
		node.returnTypes = append(node.returnTypes, returnType{
			name:  name,
			count: 1,
		})
	}
}

func (node *methodNode) findReturnType(name string) *returnType {
	for i := range node.returnTypes {
		if node.returnTypes[i].name == name {
			return &node.returnTypes[i]
		}
	}
	return nil
}

// Returns the most used return type for this method. In case of a tie, the first return type will be returned
func (node *methodNode) MostUsedReturnType() string {
	maxUsedType := returnType{
		name:  "",
		count: 0,
	}

	for _, returnType := range node.returnTypes {
		if returnType.count > maxUsedType.count {
			maxUsedType = returnType
		}
	}

	return maxUsedType.name
}

// Summarizes methods with the same predictable name to one method containing the most used return type.
// Other data like labels or filepath will get lost
func SummarizeMethodsForReturnTypes(summarizedMethods summarizedMethodsMap, methods []csv.Method) []csv.Method {
	result := make([]csv.Method, 0, len(summarizedMethods))
	statisticsOutput := make([]csv.MethodSummarizationData, 0, len(summarizedMethods))
	for _, node := range summarizedMethods {
		result = append(result, csv.Method{
			MethodName: node.name,
			ReturnType: node.MostUsedReturnType(),
		})

		types, occurences := mapReturnTypes(node.returnTypes)
		statisticsOutput = append(statisticsOutput, csv.MethodSummarizationData{
			Name:        node.name,
			Occurences:  occurences,
			ReturnTypes: types,
		})
	}

	if !configuration.StatisticsSkipCreation() {
		if err := writeSummarizedMethodsData(statisticsOutput); err != nil {
			log.ReportProblem("Could not write data for statistics generation.")
		}
	}
	return result
}

func CreateMapOfSummarizedMethods(methods []csv.Method) summarizedMethodsMap {
	methodMap := make(summarizedMethodsMap)
	for _, method := range methods {
		key := createMethodKey(method)
		if node, isDuplicate := methodMap[key]; isDuplicate {
			node.CountUpReturnType(method.ReturnType)
			continue
		}

		node := methodNode{
			name: method.MethodName,
		}
		node.CountUpReturnType(method.ReturnType)
		methodMap[key] = &node
	}
	return methodMap
}

func createMethodKey(method csv.Method) predictor.PredictableMethodName {
	return predictor.GetPredictableMethodName(method.MethodName)
}

func mapReturnTypes(returnTypes []returnType) (types []csv.MethodSummarizationReturnType, occurences int) {
	types = make([]csv.MethodSummarizationReturnType, 0, len(returnTypes))
	for _, returnType := range returnTypes {
		types = append(types, csv.MethodSummarizationReturnType{
			Name:  returnType.name,
			Count: returnType.count,
		})
		occurences += returnType.count
	}
	return types, occurences
}

func writeSummarizedMethodsData(data []csv.MethodSummarizationData) errors.Error {
	records := make([][]string, len(data))
	for i, method := range data {
		records[i] = method.ToRecord()
	}
	return csv.WriteCsvRecords(configuration.MethodSummarizationDataOutputPath(), records)
}

// Filters methods by their labels (no getters, setters, overridden methods and test code methods)
func FilterMethodsByLabels(methods []csv.Method) []csv.Method {
	result := make([]csv.Method, 0, len(methods))
	for _, method := range methods {
		if method.ReturnType == typeclasses.UnknownType {
			// Filter if return type is not determined.
			// This is not a label, but it's better to do it here aswell than reiterate through the big method list.
			continue
		} else if IsFilterForMethodsLabelsActive(method) {
			continue
		}

		result = append(result, method)
	}

	return result
}

// Returns true if the passed method should be filtered
func IsFilterForMethodsLabelsActive(method csv.Method) bool {
	return (method.HasLabel(string(java.Getter)) && configuration.MethodFilterGetter()) ||
		(method.HasLabel(string(java.Setter)) && configuration.MethodFilterSetter()) ||
		(method.HasLabel(string(java.Override)) && configuration.MethodFilterOverride()) ||
		(method.HasLabel(string(java.TestCode)) && configuration.MethodFilterTestCode())
}
