package metrics

import (
	"math"
	"regexp"
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

// Counts how often tokens of candidate occurs in reference including clipping (so each reference token maps to max. only one candidate token).
func countOverlappingWords(candidate WordCount, references []WordCount) (candidateCounts, referenceCounts, clippedCount int) {
	tokenCounts := make(WordCount)
	for _, reference := range references {
		for key, value := range reference {
			if val, ok := tokenCounts[key]; !ok || val < value {
				tokenCounts[key] = value
			}
		}
	}
	candidateCounts, referenceCounts, clippedCount = 0, 0, 0
	for key, candidateCount := range candidate {
		if refCount, ok := tokenCounts[key]; ok {
			if candidateCount < refCount {
				clippedCount += candidateCount
			} else {
				clippedCount += refCount
			}
		}
		candidateCounts += candidateCount
	}
	for _, value := range tokenCounts {
		referenceCounts += value
	}
	return candidateCounts, referenceCounts, clippedCount
}

func countWords(ngram Ngram) WordCount {
	counts := make(WordCount)
	for _, word := range ngram {
		if _, ok := counts[word]; ok {
			counts[word]++
		} else {
			counts[word] = 1
		}
	}
	return counts
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

var tokenizer = regexp.MustCompile("([a-zA-Z]+|\\[.+?\\]|-|,)")

func TokenizeSentence(str string) []string {
	return tokenizer.FindAllString(strings.ToLower(str), -1)
	//return strings.Split(str, " ")
}

func sentencesToNgrams(sentences []*Sentence, n int) []Ngram {
	referenceNgrams := make([]Ngram, len(sentences))
	for i, ref := range sentences {
		referenceNgrams[i] = ref.Ngram(n)
	}
	return referenceNgrams
}
