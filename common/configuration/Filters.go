package configuration

import (
	"encoding/json"
	"fmt"
	"regexp"
	"returntypes-langserver/common/utils"
	"strings"
)

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

type Matcher interface {
	Match([]byte) bool
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
		// for simple patterns, use strings library as it is faster
		if utils.TestString(pattern, "^\\*[a-zA-Z0-9]+$") {
			p.matcher = utils.SuffixMatcher(pattern[1:])
			return nil
		} else if utils.TestString(pattern, "^[a-zA-Z0-9]+\\*$") {
			p.matcher = utils.PrefixMatcher(pattern[:len(p.Pattern)-1])
			return nil
		} else if utils.TestString(pattern, "^\\*[a-zA-Z0-9]+\\*$") {
			p.matcher = utils.ContainingMatcher(pattern[1 : len(p.Pattern)-1])
			return nil
		} else if !strings.ContainsAny(pattern, "?*") {
			p.matcher = utils.EqualityMatcher(pattern)
			return nil
		}
		pattern = p.wildcardToRegex(p.Pattern)
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
