package excelOutputter

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"strings"
)

type PatternType string

const (
	Wildcard PatternType = "wildcard"
	RegEx    PatternType = "regex"
)

type Configuration struct {
	Datasets []Dataset `json:"datasets"`
}

type Dataset struct {
	Name             string    `json:"name"`
	Filter           Filter    `json:"filter"`
	ConvertNumbers   bool      `json:"convertNumbers"`
	Subsets          []Dataset `json:"subsets"`
	LeftoverFilename string    `json:"leftoverFilename"`
}

type Filter struct {
	Includes *FilterConfiguration `json:"includes"`
	Excludes *FilterConfiguration `json:"excludes"`
}

type FilterConfiguration struct {
	Method     []Pattern `json:"method"`
	Modifier   []Pattern `json:"modifier"`
	Parameter  []Pattern `json:"parameter"`
	Label      []Pattern `json:"label"`
	ReturnType []Pattern `json:"returntype"`
	ClassName  []Pattern `json:"classname"`
}

const PatternDelimiter = ","

func (f *FilterConfiguration) appliesOn(method csv.Method) bool {
	return f.checkPatterns(f.Method, method.MethodName) ||
		f.checkPatternsOnTargetList(f.Modifier, method.Modifier) ||
		f.checkPatternsOnTargetList(f.Parameter, method.Parameters) ||
		f.checkPatternsOnTargetList(f.Label, method.Labels) ||
		f.checkPatterns(f.ReturnType, method.ReturnType) ||
		f.checkPatterns(f.ClassName, method.ClassName)
}

func (f *FilterConfiguration) checkPatterns(patterns []Pattern, target string) bool {
	for i := range patterns {
		if patterns[i].Match(target) {
			return true
		}
	}
	return false
}

func (f *FilterConfiguration) checkPatternsOnTargetList(patterns []Pattern, targets []string) bool {
	/*for _, target := range targets {
		for i := range patterns {
			if patterns[i].Match(target) {
				return true
			}
		}
	}
	return false*/
	return f.checkPatterns(patterns, strings.Join(targets, PatternDelimiter))
}

type Pattern struct {
	Pattern      string      `json:"pattern"`
	Type         PatternType `json:"type"`
	regexPattern *regexp.Regexp
}

func (p *Pattern) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if pattern, ok := v.(string); ok {
		p.Pattern = pattern
		p.Type = Wildcard
	} else if jsonObj, ok := v.(map[string]interface{}); ok {
		if err := p.unmarshalPattern(jsonObj); err != nil {
			return err
		} else if err := p.unmarshalType(jsonObj); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported pattern")
	}
	return p.buildRegex()
}

func (p *Pattern) unmarshalPattern(jsonObj map[string]interface{}) error {
	if pattern, ok := jsonObj["pattern"].(string); ok && pattern != "" {
		p.Pattern = pattern
		return nil
	} else {
		return fmt.Errorf("unsupported pattern")
	}
}

func (p *Pattern) unmarshalType(jsonObj map[string]interface{}) error {
	if typ, ok := jsonObj["type"].(PatternType); ok && (typ == RegEx || typ == Wildcard) {
		p.Type = typ
		return nil
	} else {
		return fmt.Errorf("unsupported type")
	}
}

// Returns true if str fulfills this pattern.
func (p *Pattern) Match(str string) bool {
	if p.regexPattern == nil {
		return false
	}
	return p.regexPattern.Match([]byte(str))
}

func (p *Pattern) buildRegex() error {
	pattern := p.Pattern
	if p.Type == Wildcard {
		pattern = p.wildcardToRegex(p.Pattern)
	}
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	p.regexPattern = reg
	return nil
}

func (p *Pattern) wildcardToRegex(wildcard string) string {
	wildcard = strings.ReplaceAll(wildcard, "?", ".")
	wildcard = strings.ReplaceAll(wildcard, "*", ".*")
	return fmt.Sprintf("(^|%s)%s", PatternDelimiter, wildcard)
}

func LoadConfiguration(filepath string) (Configuration, errors.Error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "Excel Output Error", "Could not load configuration.")
	}

	var config Configuration
	if err := json.Unmarshal(contents, &config); err != nil {
		return Configuration{}, errors.Wrap(err, "Excel Output Error", "Could not parse configuration.")
	}

	return config, nil
}
