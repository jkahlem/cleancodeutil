package csv

import (
	"fmt"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"strconv"
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

func UnmarshalMethodSummarizationData(record []string) (MethodSummarizationData, errors.Error) {
	result := MethodSummarizationData{}
	result.Name = record[0]
	if val, err := strconv.Atoi(record[1]); err != nil {
		return result, errors.Wrap(err, "CSV", "Could not unmarshal to MethodSummarizationData: Expected integer value but got '%s'", record[1])
	} else {
		result.Occurences = val
	}
	if val, err := UnmarshalMethodSummarizationReturnType(SplitList(record[1])); err != nil {
		return result, err
	} else {
		result.ReturnTypes = val
	}
	return result, nil
}

func (s MethodSummarizationData) ToRecord() []string {
	record := make([]string, 3)
	record[0] = s.Name
	record[1] = fmt.Sprintf("%d", s.Occurences)
	pairs := make([]string, 0, len(s.ReturnTypes))
	for _, returnType := range s.ReturnTypes {
		pairs = append(pairs, fmt.Sprintf("%s=%d", returnType.Name, returnType.Count))
	}
	record[2] = MakeList(pairs)
	return record
}

func MarshalMethodSummarizationData(records []MethodSummarizationData) [][]string {
	result := make([][]string, len(records))
	for i := range records {
		result[i] = records[i].ToRecord()
	}
	return result
}

func (r *Reader) ReadMethodSummarizationDataRecords() ([]MethodSummarizationData, errors.Error) {
	defer r.Close()
	rows := make([]MethodSummarizationData, 0, 8)
	for {
		if record, err := r.ReadRecord(); err != nil {
			if err.Is(errors.EOF) {
				return rows, nil
			}
			return nil, err
		} else if unmarshalled, err := UnmarshalMethodSummarizationData(record); err != nil {
			return nil, err
		} else {
			rows = append(rows, unmarshalled)
		}
	}
}

func (w *Writer) WriteMethodSummarizationDataRecords(rows []MethodSummarizationData) errors.Error {
	defer w.Close()
	for _, row := range rows {
		if err := w.WriteRecord(row.ToRecord()); err != nil {
			w.err = err
			return err
		}
	}
	return nil
}

func UnmarshalMethodSummarizationReturnType(valuePairList []string) ([]MethodSummarizationReturnType, errors.Error) {
	returnTypes := make([]MethodSummarizationReturnType, 0, len(valuePairList))
	for _, pair := range valuePairList {
		if key, stringValue, ok := utils.KeyValueByEqualSign(pair); ok {
			if intVal, err := strconv.Atoi(stringValue); err != nil {
				return returnTypes, errors.Wrap(err, "CSV", "Could not unmarshal to MethodSummarizationData: Expected integer value but got '%s'", stringValue)
			} else {
				returnTypes = append(returnTypes, MethodSummarizationReturnType{
					Name:  key,
					Count: intVal,
				})
			}
		}
	}
	return returnTypes, nil
}

func parseInt(raw string, strict bool) int {
	result, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		wrappedErr := errors.Wrap(err, CsvErrorTitle, "Could not unmarshal csv data")
		if strict {
			log.ReportProblemWithError(wrappedErr, "An error occured while unmarshalling data")
		} else {
			log.Error(wrappedErr)
			log.ReportProblem("An error occured while unmarshalling data")
		}
	}
	return int(result)
}
