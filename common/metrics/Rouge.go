package metrics

func RougeN(candidate *Sentence, references []*Sentence, n int) (precision, recall float64) {
	referenceNgrams := make([]Ngram, len(references))
	for i, ref := range references {
		referenceNgrams[i] = ref.Ngram(n)
	}

	return computeScoreForNgrams(candidate.Ngram(n), referenceNgrams)
}

func computeScoreForNgrams(candidateNgrams Ngram, referenceNgrams []Ngram) (precision, recall float64) {
	for i := range referenceNgrams {
		overlapping := countOverlappingWords(candidateNgrams, referenceNgrams[i])
		p, r := calculatePrecisionRecall(overlapping, lenf(candidateNgrams), lenf(referenceNgrams[i]))
		precision += p
		recall += r
	}
	return precision, recall
}

// Computes rouge-s score for the given sentencens based on "skip-ngrams" (word pairs which allow n gaps between the two words.)
func RougeS(candidate *Sentence, references []*Sentence, n int) (precision, recall float64) {
	referenceSkipGrams := make([]Ngram, len(references))
	for i, ref := range references {
		referenceSkipGrams[i] = ref.Sgram(n)
	}

	return computeScoreForNgrams(candidate.Sgram(n), referenceSkipGrams)
}

// Computes rouge-l score for the given sentences (based on longest common subsequence)
func RougeL(candidate *Sentence, references []*Sentence) (precision, recall float64) {
	candidateTokenized := candidate.Tokens()
	for _, ref := range references {
		refTokenized := ref.Tokens()
		p, r := calculatePrecisionRecall(getLcsLength(candidateTokenized, refTokenized), lenf(candidateTokenized), lenf(refTokenized))
		precision += p
		recall += r
	}
	return precision, recall
}

func calculatePrecisionRecall(overlappingWordsCount, totalWordsCandidate, totalWordsReference float64) (float64, float64) {
	return overlappingWordsCount / totalWordsCandidate, overlappingWordsCount / totalWordsReference
}

func lenf(s []string) float64 {
	return float64(len(s))
}
