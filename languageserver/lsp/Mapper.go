package lsp

// Maps strings to configuration items with the specified section
func MapConfigurationItems(sections ...string) []ConfigurationItem {
	destination := make([]ConfigurationItem, 0, len(sections))
	for _, section := range sections {
		destination = append(destination, MapConfigurationItem(section))
	}
	return destination
}

func MapConfigurationItem(section string) ConfigurationItem {
	return ConfigurationItem{
		Section: section,
	}
}
