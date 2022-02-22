package excelOutputter

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"strings"
)

type PatternType string

const (
	Wildcard PatternType = "wildcard"
	RegExp   PatternType = "regexp"
)

type Configuration struct {
	Datasets []Dataset `json:"datasets"`
}

type Dataset struct {
	Name             string    `json:"name"`
	Filter           Filter    `json:"filter"`
	NoOutput         bool      `json:"noOutput"`
	Subsets          []Dataset `json:"subsets"`
	LeftoverFilename string    `json:"leftoverFilename"`
}

type Filter struct {
	Includes *FilterConfiguration `json:"include"`
	Excludes *FilterConfiguration `json:"exclude"`
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
	return f.checkPatterns(f.Method, method.MethodName) &&
		f.checkPatternsOnTargetList(f.Modifier, method.Modifier) &&
		f.checkPatternsOnTargetList(f.Parameter, method.Parameters) &&
		f.checkPatternsOnTargetList(f.Label, method.Labels) &&
		f.checkPatterns(f.ReturnType, method.ReturnType) &&
		f.checkPatterns(f.ClassName, method.ClassName)
}

func (f *FilterConfiguration) checkPatterns(patterns []Pattern, target string) bool {
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

func (f *FilterConfiguration) checkPatternsOnTargetList(patterns []Pattern, targets []string) bool {
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
	//return f.checkPatterns(patterns, strings.Join(targets, PatternDelimiter))
}

type Pattern struct {
	Pattern string      `json:"pattern"`
	Type    PatternType `json:"type"`
	matcher Matcher
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
		} else if len(jsonObj) != 2 {
			return fmt.Errorf("expected pattern object to have only 2 fields ('pattern' and 'type')")
		}
	} else {
		return fmt.Errorf("unsupported pattern: %v", v)
	}
	return p.buildMatcher()
}

func (p *Pattern) unmarshalPattern(jsonObj map[string]interface{}) error {
	if pattern, ok := jsonObj["pattern"].(string); ok && pattern != "" {
		p.Pattern = pattern
		return nil
	} else {
		return fmt.Errorf("unsupported pattern: %v", jsonObj["pattern"])
	}
}

func (p *Pattern) unmarshalType(jsonObj map[string]interface{}) error {
	if typ, ok := jsonObj["type"].(string); ok && (typ == string(RegExp) || typ == string(Wildcard)) {
		p.Type = PatternType(typ)
		return nil
	} else {
		return fmt.Errorf("unsupported type: %v", jsonObj["type"])
	}
}

// Returns true if str fulfills this pattern.
func (p *Pattern) Match(str string) bool {
	if p.matcher == nil {
		return false
	}
	return p.matcher.Match([]byte(str))
}

func (p *Pattern) buildMatcher() error {
	pattern := p.Pattern
	if p.Type == Wildcard {
		lowerCase := strings.ToLower(p.Pattern)
		// for simple patterns, use strings library as it is faster
		if test(lowerCase, "^\\*[a-z0-9]+$") {
			p.matcher = SuffixMatcher(lowerCase[1:])
			return nil
		} else if test(lowerCase, "^[a-z0-9]+\\*$") {
			p.matcher = PrefixMatcher(lowerCase[:len(p.Pattern)-1])
			return nil
		} else if test(lowerCase, "^\\*[a-z0-9]+\\*$") {
			p.matcher = ContainingMatcher(lowerCase[1 : len(p.Pattern)-1])
			return nil
		} else if !strings.ContainsAny(lowerCase, "?*") {
			p.matcher = EqualityMatcher(lowerCase)
			return nil
		}
		pattern = p.wildcardToRegex(lowerCase)
	}
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	p.matcher = reg
	return nil
}

type Matcher interface {
	Match([]byte) bool
}

func (p *Pattern) wildcardToRegex(wildcard string) string {
	wildcard = strings.ReplaceAll(wildcard, "?", ".")
	wildcard = strings.ReplaceAll(wildcard, "*", ".*")
	return fmt.Sprintf("(^|%s)%s", PatternDelimiter, wildcard)
}

func test(s, expr string) bool {
	r, err := regexp.Compile(expr)
	if err != nil {
		return false
	}
	return r.MatchString(s)
}

type SuffixMatcher string
type PrefixMatcher string
type ContainingMatcher string
type EqualityMatcher string

func (suffix SuffixMatcher) Match(target []byte) bool {
	return strings.HasSuffix(string(target), string(suffix))
}

func (prefix PrefixMatcher) Match(target []byte) bool {
	return strings.HasPrefix(string(target), string(prefix))
}

func (substr ContainingMatcher) Match(target []byte) bool {
	return strings.Contains(string(target), string(substr))
}

func (substr EqualityMatcher) Match(target []byte) bool {
	return string(target) == string(substr)
}

func LoadConfiguration(filepath string) (Configuration, errors.Error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "Excel Output Error", "Could not load configuration.")
	}
	return LoadConfigurationFromJson(contents)
}

func LoadConfigurationFromJson(contents []byte) (Configuration, errors.Error) {
	var config Configuration
	if err := utils.UnmarshalJSONStrict(contents, &config); err != nil {
		return Configuration{}, errors.Wrap(err, "Excel Output Error", "Could not parse configuration.")
	}

	return config, nil
}
