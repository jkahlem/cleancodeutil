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

// Loads records from a csv file and passes it to the stream
type csvLoader struct {
	err         errors.Error
	sliceLoader Loader
}

func newCsvLoader(path string) *csvLoader {
	records, err := csv.ReadRecords(path)
	return &csvLoader{
		err:         err,
		sliceLoader: newSliceLoader(records),
	}
}

func (l *csvLoader) Load() ([]string, errors.Error) {
	if l.err != nil {
		return nil, l.err
	}
	return l.sliceLoader.Load()
}

type sliceLoader struct {
	records      [][]string
	currentIndex int
}

func newSliceLoader(records [][]string) Loader {
	return &sliceLoader{
		records:      records,
		currentIndex: -1,
	}
}

func (l *sliceLoader) Load() ([]string, errors.Error) {
	l.currentIndex++
	if l.currentIndex >= len(l.records) {
		return nil, nil
	}
	return l.records[l.currentIndex], nil
}

type channelLoader struct {
	ch           chan []string
	currentIndex int
}

func newChannelLoader(ch chan []string) Loader {
	return &channelLoader{
		ch: ch,
	}
}

func (l *channelLoader) Load() ([]string, errors.Error) {
	record, isOpen := <-l.ch
	if record == nil && isOpen {
		// As long as the channel is not closed, pass an empty record if nil is passed.
		record = make([]string, 0)
	}
	return record, nil
}
