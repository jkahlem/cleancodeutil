package statistics

// Counts the usages of each return type
type ReturnTypeUsageCounter struct {
	Counts map[string]int `json:"counts"`
}

// Adds the usage of a given type
func (usage *ReturnTypeUsageCounter) AddUsage(typeName string) {
	if usage.Counts == nil {
		usage.Counts = make(map[string]int)
	}

	if _, exists := usage.Counts[typeName]; exists {
		usage.Counts[typeName]++
	} else {
		usage.Counts[typeName] = 1
	}
}
