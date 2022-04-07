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
	records, err := csv.NewFileReader(path).ReadAllRecords()
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
	ch           *Channel
	currentIndex int
}

func newChannelLoader(ch *Channel) Loader {
	return &channelLoader{
		ch: ch,
	}
}

func (l *channelLoader) Load() ([]string, errors.Error) {
	record, isOpen := l.ch.NextRecord()
	if record == nil && isOpen {
		// As long as the channel is not closed, pass an empty record if nil is passed.
		record = make([]string, 0)
	}
	return record, nil
}

type Channel struct {
	input  chan []string
	errors chan errors.Error
}

// Creates a channel with an input and an error channel. This makes writing in asynchronously easier, when having the actual stream inside a go routine.
func NewChannel() *Channel {
	return &Channel{
		input:  make(chan []string),
		errors: make(chan errors.Error),
	}
}

// Returns the next record passed in the channel
func (c *Channel) NextRecord() (record []string, open bool) {
	if c.input == nil {
		return nil, false
	}
	record, open = <-c.input
	return
}

// Puts a record in the channel which is passed through the stream
func (c *Channel) PutRecord(record []string) {
	if c.input == nil {
		c.input = make(chan []string)
	}
	c.input <- record
}

// Puts a new error in the channel
func (c *Channel) PutError(err errors.Error) {
	if c.errors == nil {
		c.errors = make(chan errors.Error)
	}
	c.errors <- err
}

// Returns an error if the stream contains any.
func (c *Channel) NextError() errors.Error {
	if c.errors == nil {
		return nil
	}
	err, isOpen := <-c.errors
	if !isOpen {
		return nil
	}
	return err
}

// Closes only the input channel - it is still possible to put errors.
func (c *Channel) Close() {
	if c.input != nil {
		close(c.input)
	}
}

// A function which returns records. Returning nil indicates, that there are no records left.
type LoaderFunc func() []string

func (l LoaderFunc) Load() ([]string, errors.Error) {
	return l(), nil
}
