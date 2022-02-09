package csv

import (
	"fmt"
	"strings"
)

// For structures which cannot be easily marshalled

type MethodSummarizationData struct {
	Name        string                          `json:"name"`
	Occurences  int                             `json:"occurences"`
	ReturnTypes []MethodSummarizationReturnType `json:"returnTypes"`
}

type MethodSummarizationReturnType struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func UnmarshalMethodSummarizationData(records [][]string) []MethodSummarizationData {
	data := make([]MethodSummarizationData, len(records))
	for i, record := range records {
		data[i].Name = record[0]
		data[i].Occurences = parseInt(record[1], false)
		data[i].ReturnTypes = UnmarshalMethodSummarizationReturnType(SplitList(record[2]))
	}
	return data
}

func UnmarshalMethodSummarizationReturnType(valuePairList []string) []MethodSummarizationReturnType {
	returnTypes := make([]MethodSummarizationReturnType, 0, len(valuePairList))
	for _, pair := range valuePairList {
		splitted := strings.Split(pair, "=")
		if len(splitted) != 2 {
			continue
		}

		returnTypes = append(returnTypes, MethodSummarizationReturnType{
			Name:  splitted[0],
			Count: parseInt(splitted[1], false),
		})
	}
	return returnTypes
}

func (data MethodSummarizationData) ToRecord() []string {
	pairs := make([]string, 0, len(data.ReturnTypes))
	for _, returnType := range data.ReturnTypes {
		pairs = append(pairs, fmt.Sprintf("%s=%d", returnType.Name, returnType.Count))
	}
	return []string{
		data.Name,
		fmt.Sprintf("%d", data.Occurences),
		MakeList(pairs),
	}
}
