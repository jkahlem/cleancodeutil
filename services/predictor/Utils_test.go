package predictor

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentenceSplitter(t *testing.T) {
	testStrings := [][]string{
		{"createObject", "create object"},
		{"FindObjectByName", "find object by name"},
		{"parseJSON", "parse json"},
		{"parseURLs", "parse urls"},
		{"URLsToParse", "urls to parse"},
		{"RGBToHSL", "rgb to hsl"},
		{"recognize_snake_Case", "recognize snake case"},
		{"_underscore_beginning", "underscore beginning"},
		{"multiple__underscores____exists", "multiple underscores exists"}}

	for _, pair := range testStrings {
		actual := strings.ToLower(SplitMethodNameToSentence(pair[0]))
		assert.Equal(t, pair[1], actual)
	}
}

func TestNumberToWordConverter(t *testing.T) {
	testStrings := [][]string{{"bytes2string", "bytes to string"},
		{"someMethod2", "some method 2"},
		{"from2022to2023", "from 2022 to 2023"}}

	for _, pair := range testStrings {
		actual := strings.ToLower(SplitMethodNameToSentence(pair[0]))
		assert.Equal(t, pair[1], actual)
	}
}

func TestTest(t *testing.T) {
	rx := regexp.MustCompile(`\d+`)
	assert.True(t, rx.MatchString("asd"))
}
