package csv

import "returntypes-langserver/common/configuration"

func IsMethodIncluded(method Method, filter configuration.Filter) bool {
	if filter.Includes != nil {
		if !isAnyFilterFulfilled(method, filter.Includes) {
			return false
		}
	}
	if filter.Excludes != nil {
		if isAnyFilterFulfilled(method, filter.Excludes) {
			return false
		}
	}
	return true
}

func isAnyFilterFulfilled(method Method, filters configuration.FilterConfigurations) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if isFilterFulfilled(method, filter) {
			return true
		}
	}
	return false
}

func areAllFiltersFulfilled(method Method, filters configuration.FilterConfigurations) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if !isFilterFulfilled(method, filter) {
			return false
		}
	}
	return true
}

func isFilterFulfilled(method Method, f configuration.FilterConfiguration) bool {
	return checkPatterns(f.Method, method.MethodName) &&
		checkPatternsOnTargetList(f.Modifier, method.Modifier) &&
		checkPatternsOnTargetList(f.Parameter, method.Parameters) &&
		checkPatternsOnTargetList(f.Label, method.Labels) &&
		checkPatterns(f.ReturnType, method.ReturnType) &&
		checkPatterns(f.ClassName, method.ClassName) &&
		isAnyFilterFulfilled(method, f.AnyOf) &&
		areAllFiltersFulfilled(method, f.AllOf)
}

func checkPatterns(patterns []configuration.Pattern, target string) bool {
	if len(patterns) == 0 {
		return true
	}
	for i := range patterns {
		if patterns[i].Match(target) {
			return true
		}
	}
	return false
}

func checkPatternsOnTargetList(patterns []configuration.Pattern, targets []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, target := range targets {
		for i := range patterns {
			if patterns[i].Match(target) {
				return true
			}
		}
	}
	return false
}
