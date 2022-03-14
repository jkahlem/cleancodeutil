package rouge

import "strings"

type ngram []string

func ComputeN(candidate []string, references [][]string, n int) float64 {
	candidateNgrams := getNgrams(candidate, n)
	referenceNgrams := make([]ngram, len(references))
	for i, ref := range references {
		referenceNgrams[i] = getNgrams(ref, n)
	}

	return computeScoreForNgrams(candidateNgrams, referenceNgrams)
}

func computeScoreForNgrams(candidateNgrams ngram, referenceNgrams []ngram) float64 {
	var bestScore float64 = 0
	for i := range referenceNgrams {
		overlapping := countOverlappingWords(candidateNgrams, referenceNgrams[i])
		f := calculateRougeFscore(overlapping, lenf(candidateNgrams), lenf(referenceNgrams[i]), 1)
		if f > bestScore {
			bestScore = f
		}
	}
	return bestScore
}

func fscore(precision, recall, beta float64) float64 {
	b2 := beta * beta
	return (1 + b2) * (precision * recall) / ((b2 * precision) + recall)
}

func getNgrams(target []string, n int) ngram {
	if n <= 0 {
		panic("invalid n-gram value")
	} else if n == 1 {
		return target
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

func countOverlappingWords(candidate, reference ngram) float64 {
	var n float64 = 0
	for _, word := range candidate {
		for _, refWord := range reference {
			if word == refWord {
				n++
				break
			}
		}
	}
	return n
}

// Computes rouge-s score for the given sentencens based on "skip-ngrams" (word pairs which allow n gaps between the two words.)
func ComputeS(candidate []string, references [][]string, n int) float64 {
	candidateSkipGrams := getSkipGrams(candidate, n)
	referenceSkipGrams := make([]ngram, len(references))
	for i, ref := range references {
		referenceSkipGrams[i] = getSkipGrams(ref, n)
	}

	return computeScoreForNgrams(candidateSkipGrams, referenceSkipGrams)
}

func getSkipGrams(target []string, n int) ngram {
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

// Computes rouge-l score for the given sentences (based on longest common subsequence)
func ComputeL(candidate []string, references [][]string) float64 {
	var bestScore float64
	for _, ref := range references {
		f := calculateRougeFscore(getLcsLength(candidate, ref), lenf(candidate), lenf(ref), 1)
		if f > bestScore {
			bestScore = f
		}
	}
	return 0
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

func calculateRougeFscore(overlappingWordsCount, totalWordsCandidate, totalWordsReference, beta float64) float64 {
	precision := overlappingWordsCount / totalWordsCandidate
	recall := overlappingWordsCount / totalWordsReference
	return fscore(precision, recall, 1)
}

func lenf(s []string) float64 {
	return float64(len(s))
}
