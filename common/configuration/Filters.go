package configuration

import (
	"encoding/json"
	"fmt"
	"regexp"
	"returntypes-langserver/common/utils"
	"strings"
)

type Filter struct {
	Includes FilterConfigurations `json:"include,omitempty"`
	Excludes FilterConfigurations `json:"exclude,omitempty"`
}

type FilterConfigurations []FilterConfiguration

func (f *FilterConfigurations) UnmarshalJSON(data []byte) error {
	var v FilterConfiguration
	if err := json.Unmarshal(data, &v); err == nil {
		*f = append(*f, v)
		return nil
	}
	var valueSlice []FilterConfiguration
	if err := json.Unmarshal(data, &valueSlice); err == nil {
		*f = valueSlice
		return nil
	} else {
		return err
	}
}

func (f FilterConfigurations) DecodeValue(value interface{}) (interface{}, error) {
	var v FilterConfiguration
	if err := utils.DecodeMapToStructStrict(value, &v); err == nil {
		return []FilterConfiguration{v}, nil
	}
	var valueSlice []FilterConfiguration
	if err := utils.DecodeMapToStructStrict(value, &valueSlice); err == nil {
		return valueSlice, nil
	}
	return value, nil
}

type FilterConfiguration struct {
	Method     []Pattern             `json:"method"`
	Modifier   []Pattern             `json:"modifier"`
	Parameter  []Pattern             `json:"parameter"`
	Label      []Pattern             `json:"label"`
	ReturnType []Pattern             `json:"returntype"`
	ClassName  []Pattern             `json:"classname"`
	FilePath   []Pattern             `json:"filePath"`
	AnyOf      []FilterConfiguration `json:"anyOf"`
	AllOf      []FilterConfiguration `json:"allOf"`
}

type Matcher interface {
	Match([]byte) bool
}

type Pattern struct {
	Pattern string      `json:"pattern"`
	Type    PatternType `json:"type"`
	Min     int         `json:"min"`
	Max     int         `json:"max"`
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
		if err := p.unmarshalType(jsonObj); err != nil {
			return err
		}
		p.unmarshalPattern(jsonObj)
		p.unmarshalMinMax(jsonObj)
	} else {
		return fmt.Errorf("unsupported pattern: %v", v)
	}
	return p.buildMatcher()
}

func (p Pattern) DecodeValue(value interface{}) (interface{}, error) {
	if pattern, ok := value.(string); ok {
		p.Pattern = pattern
		p.Type = Wildcard
		return p, p.buildMatcher()
	} else if jsonObj, ok := value.(map[string]interface{}); ok {
		if err := p.unmarshalType(jsonObj); err != nil {
			return value, err
		}
		p.unmarshalPattern(jsonObj)
		p.unmarshalMinMax(jsonObj)
		return p, p.buildMatcher()
	}
	return value, nil
}

func (p *Pattern) unmarshalType(jsonObj map[string]interface{}) error {
	if typ, ok := jsonObj["type"].(string); ok && p.isSupportedType(typ) {
		p.Type = PatternType(typ)
		return nil
	} else {
		return fmt.Errorf("unsupported type: %v", jsonObj["type"])
	}
}

func (p *Pattern) unmarshalPattern(jsonObj map[string]interface{}) {
	if pattern, ok := jsonObj["pattern"].(string); ok && pattern != "" {
		p.Pattern = pattern
	}
}

func (p *Pattern) unmarshalMinMax(jsonObj map[string]interface{}) {
	if typedValue, ok := jsonObj["min"].(float64); ok {
		p.Min = int(typedValue)
	}
	if typedValue, ok := jsonObj["max"].(float64); ok {
		p.Max = int(typedValue)
	}
}

func (p *Pattern) isSupportedType(typ string) bool {
	return utils.ContainsString([]string{string(RegExp), string(Wildcard), string(Length), string(Counter)}, typ)
}

// Returns true if str fulfills this pattern.
func (p *Pattern) Match(str string) bool {
	if p.matcher == nil {
		if err := p.buildMatcher(); err != nil || p.matcher == nil {
			return false
		}
	}
	return p.matcher.Match([]byte(str))
}

func (p *Pattern) buildMatcher() error {
	if (p.Type == Wildcard || p.Type == RegExp) && (p.Min != 0 || p.Max != 0) {
		return fmt.Errorf("unexpected min/max values for %s matcher", p.Type)
	} else if (p.Type == Wildcard || p.Type == RegExp || p.Type == Counter) && p.Pattern == "" {
		return fmt.Errorf("no pattern defined for %s matcher", p.Type)
	}

	pattern := p.Pattern
	switch p.Type {
	case Wildcard:
		// for simple patterns, use strings library as it is faster
		if utils.TestString(pattern, "^\\*[a-zA-Z0-9]+$") {
			p.matcher = utils.SuffixMatcher(pattern[1:])
			return nil
		} else if utils.TestString(pattern, "^[a-zA-Z0-9]+\\*$") {
			p.matcher = utils.PrefixMatcher(pattern[:len(pattern)-1])
			return nil
		} else if utils.TestString(pattern, "^\\*[a-zA-Z0-9]+\\*$") {
			p.matcher = utils.ContainingMatcher(pattern[1 : len(pattern)-1])
			return nil
		} else if !strings.ContainsAny(pattern, "?*") {
			p.matcher = utils.EqualityMatcher(pattern)
			return nil
		}
		pattern = p.wildcardToRegex(pattern)
	case Length:
		if p.Pattern != "" {
			return fmt.Errorf("unexpected value for field 'pattern' on length matcher")
		} else if p.Max == 0 && p.Min == 0 {
			return fmt.Errorf("no min or max boundaries set for length matcher")
		} else if p.Max > 0 && p.Min > p.Max {
			return fmt.Errorf("minimum value (%d) is greater than maximum value (%d).", p.Min, p.Max)
		}
		p.matcher = LengthMatcher{
			Min: p.Min,
			Max: p.Max,
		}
		return nil
	case Counter:
		if p.Max > 0 && p.Min > p.Max {
			return fmt.Errorf("minimum value (%d) is greater than maximum value (%d).", p.Min, p.Max)
		} else if p.Max == 0 && p.Min == 0 {
			return fmt.Errorf("no min or max boundaries set for counter matcher")
		}
		p.matcher = LengthMatcher{
			Min: p.Min,
			Max: p.Max,
		}
		return nil
	}
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	p.matcher = reg
	return nil
}

func (p *Pattern) wildcardToRegex(wildcard string) string {
	wildcard = strings.ReplaceAll(wildcard, "?", ".")
	wildcard = strings.ReplaceAll(wildcard, "*", ".*")
	return fmt.Sprintf("^%s$", wildcard)
}

type LengthMatcher struct {
	Min int
	Max int
}

func (m LengthMatcher) Match(contents []byte) bool {
	if m.Max > 0 && len(contents) > m.Max {
		return false
	}
	return len(contents) >= m.Min
}

type CountMatcher struct {
	Pattern string
	Min     int
	Max     int
}

func (m CountMatcher) Match(contents []byte) bool {
	count := strings.Count(string(contents), m.Pattern)
	if m.Max > 0 && count > m.Max {
		return false
	}
	return count >= m.Min
}
