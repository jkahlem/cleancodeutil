package rouge

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouge(t *testing.T) {
	score := ComputeN(sentence("the cat was found under the bed"), [][]string{sentence("the cat was under the bed")}, 1)
	fmt.Println(score)
}

func TestLcs(t *testing.T) {
	lcs := getLcsLength(sentence("the cat was found under the bed"), sentence("the under cat was under the bed"))
	assert.Equal(t, float64(6), lcs)
}

func sentence(str string) []string {
	return strings.Split(str, " ")
}
