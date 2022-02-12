package excel

import (
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
)

type Loader interface {
	// Loads a record and returns it. If no records are available (e.g. reached end of file),
	// it returns no record and no error. (Both values are nil)
	Load() ([]string, errors.Error)
}

type csvLoader struct {
	filepath     string
	records      [][]string
	currentIndex int
}

func newCsvLoader(path string) *csvLoader {
	return &csvLoader{
		filepath: path,
	}
}

func (l *csvLoader) Load() ([]string, errors.Error) {
	if l.records == nil {
		if records, err := csv.ReadRecords(l.filepath); err != nil {
			return nil, err
		} else {
			l.records = records
			l.currentIndex = -1
		}
	}
	l.currentIndex++
	if l.currentIndex >= len(l.records) {
		return nil, nil
	}
	return l.records[l.currentIndex], nil
}
