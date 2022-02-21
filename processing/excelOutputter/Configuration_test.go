package excelOutputter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatcherBuilding(t *testing.T) {
	// given
	filter := FilterConfiguration{}
	raw := `{"method": ["open*", "*to*", "*end"]}`

	// when
	err := json.Unmarshal([]byte(raw), &filter)

	// then
	assert.NoError(t, err)
	assert.Len(t, filter.Method, 3)

	m1, ok := filter.Method[0].matcher.(PrefixMatcher)
	assert.True(t, ok)
	assert.True(t, m1.Match([]byte("open the door")))
	assert.False(t, m1.Match([]byte("not open")))

	m2, ok := filter.Method[1].matcher.(ContainingMatcher)
	assert.True(t, ok)
	assert.True(t, m2.Match([]byte("a to b")))
	assert.True(t, m2.Match([]byte("to string")))
	assert.True(t, m2.Match([]byte("rename to")))

	m3, ok := filter.Method[2].matcher.(SuffixMatcher)
	assert.True(t, ok)
	assert.True(t, m3.Match([]byte("has to end")))
	assert.False(t, m3.Match([]byte("do not end it")))
}

func TestConfigurationLoading(t *testing.T) {
	// given
	raw := `
	{
		"datasets": [{
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

	// when
	config, err := LoadConfigurationFromJson([]byte(raw))

	// then
	assert.NoError(t, err)
	assert.Equal(t, "save*", config.Datasets[0].Subsets[0].Filter.Includes.Method[0].Pattern)
}
