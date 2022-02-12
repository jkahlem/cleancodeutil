package excel

import "returntypes-langserver/common/debug/errors"

type Col int

type WriteableStream interface {
	// Performs changes on the data of a record
	Transform(transformer RecordTransformer) WriteableStream
	// Inserts the new columns at the given position (zero-indexed)
	InsertColumnsAt(position Col, columns ...string) WriteableStream
	// Swaps two columns
	Swap(i, j Col) WriteableStream
	// For external writer
	Do(StreamWriter) WriteableStream

	// Writes all records to the given excel file
	ToFile(path string) errors.Error
	// Writes all records to the given slice
	ToSlice(slice *[][]string) errors.Error
	// For external collectors
	To(Collector) errors.Error
}

// A writeable stream which may have a base layout
type BaseLayoutStream interface {
	WriteableStream
	// Sets the base layout for the stream using the given struct
	WithColumnsFromStruct(structType interface{}) WriteableStream
	// Sets the base layout for the stream using the given headers
	WithStaticHeaders(headers ...string) WriteableStream
	// For external writer which define a base layout
	With(StreamWriter) WriteableStream
}

type stream struct {
	loader Loader
	parts  []StreamWriter
}

// Helper class to enforce starting stream creation with a loader.
type streamStart struct{}

// Creates a stream to process data on a stream, starting with loading the data, altering data, columns and so on and writing them to the file.
// Example:
//   excel.Stream().FromCSVFile(path)
//     .FormattedByStruct(csv.Method{})
//     .Transform(TypeUnqualifier)
//     .InsertColumnsAt(excel.Col(1), "Groups", "Labels")
//     .Swap(excel.Col(2), excel.Col(5))
//     .ToFile("path/to/output.xlsx")
func Stream() *streamStart {
	return &streamStart{}
}

/*-- Loader methods --*/

func (s *streamStart) FromCSVFile(path string) BaseLayoutStream {
	return s.From(newCsvLoader(path))
}

func (s *streamStart) FromSlice(records [][]string) BaseLayoutStream {
	return s.From(newSliceLoader(records))
}

func (s *streamStart) From(loader Loader) BaseLayoutStream {
	return &stream{
		loader: loader,
	}
}

/*-- Base layout methods --*/

func (s *stream) WithColumnsFromStruct(structType interface{}) WriteableStream {
	return s.addWriter(newStructFormatWriter(structType))
}

func (s *stream) WithStaticHeaders(headers ...string) WriteableStream {
	return s.addWriter(newStaticFormatWriter(headers))
}

func (s *stream) With(writer StreamWriter) WriteableStream {
	return s.addWriter(writer)
}

/*-- Writer Methods --*/

func (s *stream) Transform(transformer RecordTransformer) WriteableStream {
	return s.addWriter(newTransformer(transformer))
}

func (s *stream) InsertColumnsAt(position Col, columns ...string) WriteableStream {
	return s.addWriter(newColumnInserter(position, columns...))
}

func (s *stream) Swap(i, j Col) WriteableStream {
	return s.addWriter(newColumnSwapper(i, j))
}

func (s *stream) Do(writer StreamWriter) WriteableStream {
	return s.addWriter(writer)
}

func (s *stream) addWriter(writer StreamWriter) WriteableStream {
	s.parts = append(s.parts, writer)
	return s
}

/*-- Collector Methods --*/

func (s *stream) ToFile(path string) errors.Error {
	return s.To(newFileCollector(path))
}

func (s *stream) ToSlice(slice *[][]string) errors.Error {
	return s.To(newSliceCollector(slice))
}

func (s *stream) To(collector Collector) errors.Error {
	defer collector.Close()

	s.connect(collector)
	head := s.parts[0]
	if err := head.BuildLayout(DefaultLayout()); err != nil {
		return errors.Wrap(err, "Excel Error", "Could not build layout")
	}
	for {
		record, err := s.loader.Load()
		if err != nil {
			return errors.Wrap(err, "Excel Error", "An error occured when loading a record from stream")
		} else if record == nil {
			// End when no record is left
			return nil
		} else if err = head.Write(record); err != nil {
			return errors.Wrap(err, "Excel Error", "An error occured while writing an excel row.")
		}
	}
}

func (s *stream) connect(c Collector) {
	for i := len(s.parts) - 2; i >= 0; i-- {
		s.parts[i].SetWriter(s.parts[i+1])
	}
	s.parts[len(s.parts)-1].SetWriter(&collector{collector: c})
}

type collector struct {
	layout        Layout
	collector     Collector
	headerWritten bool
	oddRow        bool
}

func (c *collector) BuildLayout(layout Layout) errors.Error {
	c.layout = layout
	c.collector.ApplyLayout(layout)
	return nil
}

func (c *collector) WriteHeader() errors.Error {
	if c.collector == nil {
		return nil
	}
	header := make([]string, 0, len(c.layout.Columns))
	for _, col := range c.layout.Columns {
		header = append(header, col.Header)
	}
	return c.collector.Write(header, &c.layout.HeaderStyle)
}

func (c *collector) Write(record []string) errors.Error {
	if c.collector == nil {
		return nil
	}
	if !c.headerWritten {
		if err := c.WriteHeader(); err != nil {
			return err
		}
		c.headerWritten = true
	}
	c.oddRow = !c.oddRow
	return c.collector.Write(record, c.getRowStyle())
}

func (c *collector) getRowStyle() *Style {
	if c.oddRow {
		return &c.layout.OddRowStyle
	} else {
		return &c.layout.EvenRowStyle
	}
}

func (c *collector) SetWriter(StreamWriter) {}
