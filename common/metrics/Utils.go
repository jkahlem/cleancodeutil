package metrics

import (
	"math"
	"strings"
)

type Ngram []string

func FScore(precision, recall, beta float64) float64 {
	b2 := beta * beta
	return (1 + b2) * (precision * recall) / ((b2 * precision) + recall)
}

func getNgrams(target []string, n int) Ngram {
	if n <= 0 {
		panic("invalid n-gram value")
	} else if n == 1 {
		return target
	} else if n >= len(target) {
		return []string{strings.Join(target, " ")}
	}
	result := make([]string, len(target)-n)
	for i := 0; i < len(target)-n; i++ {
		g := make([]string, n)
		for j := 0; j < n; j++ {
			g[j] = target[i+j]
		}
		result[i] = strings.Join(g, " ")
	}
	return result
}

func countOverlappingWords(candidate, reference Ngram) float64 {
	var n float64 = 0
	for _, word := range candidate {
		for i, refWord := range reference {
			if word == refWord {
				n++
				// TODO: this might actually change the contents of the input (for n = 1)... maybe accept only strings and split them by whitespaces?
				reference[i] = ""
				break
			}
		}
	}
	return n
}

func getSkipGrams(target []string, n int) Ngram {
	if len(target) < 2 {
		panic("Cannot build skip grams for sentence with lesser than 2 words")
	}
	result := make([]string, 0, len(target))
	for i, word := range target {
		for j := i + 1; j <= n; j++ {
			result = append(result, word+" "+target[j])
		}
	}
	return result
}

func getLcsLength(candidate, reference []string) float64 {
	m, n := len(candidate)+1, len(reference)+1
	mat := make([][]int, m)
	for i := 0; i < len(mat); i++ {
		mat[i] = make([]int, n)
	}

	for i := 1; i < m; i++ {
		for j := 1; j < n; j++ {
			if candidate[i-1] == reference[j-1] {
				mat[i][j] = mat[i-1][j-1] + 1
			} else {
				mat[i][j] = max(mat[i][j-1], mat[i-1][j])
			}
		}
	}
	return float64(mat[m-1][n-1])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type GeometricMean struct {
	count float64
	sum   float64
}

func (g GeometricMean) Add(value float64) GeometricMean {
	g.count++
	g.sum += math.Log(value)
	return g
}

func (g GeometricMean) Value() float64 {
	return math.Exp(g.sum / g.count)
}

func tokenizeSentence(str string) []string {
	return strings.Split(str, " ")
}
