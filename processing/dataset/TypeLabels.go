package dataset

import (
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
)

// For mapping type names to their labels (numbers/categories) in the dataset and vice versa
type TypeLabelMapper struct {
	mappings        map[string]int
	reverseMappings map[int]string
	labelCounter    int
}

// Imports labels from a csv file
func (m *TypeLabelMapper) LoadFromFile(labelFile string) errors.Error {
	records, err := csv.ReadRecords(labelFile)
	if err != nil {
		return err
	}

	labels := csv.UnmarshalTypeLabel(records)
	for _, label := range labels {
		m.AddLabel(label)
	}
	return nil
}

// Adds a label to the mappings
func (m *TypeLabelMapper) AddLabel(label csv.TypeLabel) {
	if m.mappings == nil {
		m.mappings = make(map[string]int)
		m.reverseMappings = make(map[int]string)
	}

	m.mappings[label.Name] = label.Label
	m.reverseMappings[label.Label] = label.Name
	if label.Label >= m.labelCounter {
		m.labelCounter = label.Label + 1
	}
}

// Gets the label number for a type name. If a label does not exist, a new one will be created
func (m *TypeLabelMapper) GetLabel(typeName string) int {
	if label, ok := m.mappings[typeName]; ok {
		return label
	}
	m.AddLabel(csv.TypeLabel{
		Name:  typeName,
		Label: m.labelCounter,
	})
	return m.mappings[typeName]
}

// The mappings in a slice
func (m *TypeLabelMapper) GetMappings() []csv.TypeLabel {
	labels := make([]csv.TypeLabel, len(m.mappings))
	i := 0
	for name, label := range m.mappings {
		labels[i].Name = name
		labels[i].Label = label
		i++
	}
	return labels
}

// Returns the name of a type for a label
func (m *TypeLabelMapper) GetTypeName(label int) (typeName string, ok bool) {
	typeName, ok = m.reverseMappings[label]
	return
}

// Exports the mappings to the given file
func (m *TypeLabelMapper) WriteMappings(outputPath string) errors.Error {
	mappings := m.GetMappings()
	records := make([][]string, len(mappings))
	for i, row := range mappings {
		records[i] = row.ToRecord()
	}
	return csv.WriteCsvRecords(outputPath, records)
}
