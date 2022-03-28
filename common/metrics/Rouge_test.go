package metrics

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouge(t *testing.T) {
	s := NewSentence("the cat was found under the bed")
	p, r := RougeN(s, []*Sentence{NewSentence("the cat was under the bed")}, 1)
	fmt.Println(FScore(p, r, 1))
	assert.NotNil(t, s.ngrams[1])
}

func TestLcs(t *testing.T) {
	lcs := getLcsLength(tokenize("the cat was found under the bed"), tokenize("the under cat was under the bed"))
	assert.Equal(t, float64(6), lcs)
}

func tokenize(str string) []string {
	return strings.Split(str, " ")
}
