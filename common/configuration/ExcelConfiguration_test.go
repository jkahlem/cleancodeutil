package configuration

import (
	"encoding/json"
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatcherBuilding(t *testing.T) {
	// given
	filter := FilterConfiguration{}
	raw := `{"method": ["open*", "*to*", "*end", "method"]}`

	// when
	err := json.Unmarshal([]byte(raw), &filter)

	// then
	assert.NoError(t, err)
	assert.Len(t, filter.Method, 4)

	m1, ok := filter.Method[0].matcher.(utils.PrefixMatcher)
	assert.True(t, ok)
	assert.True(t, m1.Match([]byte("open the door")))
	assert.False(t, m1.Match([]byte("not open")))

	m2, ok := filter.Method[1].matcher.(utils.ContainingMatcher)
	assert.True(t, ok)
	assert.True(t, m2.Match([]byte("a to b")))
	assert.True(t, m2.Match([]byte("to string")))
	assert.True(t, m2.Match([]byte("rename to")))

	m3, ok := filter.Method[2].matcher.(utils.SuffixMatcher)
	assert.True(t, ok)
	assert.True(t, m3.Match([]byte("has to end")))
	assert.False(t, m3.Match([]byte("do not end it")))

	m4, ok := filter.Method[3].matcher.(utils.EqualityMatcher)
	assert.True(t, ok)
	assert.True(t, m4.Match([]byte("method")))
	assert.False(t, m4.Match([]byte("a method")))
}

func TestConfigurationLoading(t *testing.T) {
	// given
	raw := `
	{
		"excelSets": [{
        	"name": "default",
        	"filter": {
				"exclude": {
					"label":["testMethod"]
				}
			},
			"noOutput": true,
			"subsets": [
				{
					"name": "saveRemove",
					"filter": {
						"include": {
							"method": ["save*", "remove*"]
						}
					}
				}
			]
		}]
	}`
	config := make(ExcelSetConfiguration, 0)

	// when
	err := config.fromJson([]byte(raw))

	// then
	assert.NoError(t, err)
	assert.Equal(t, "save*", config[0].Subsets[0].Filter.Includes.Method[0].Pattern)
}
