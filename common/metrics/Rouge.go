package metrics

func RougeN(candidate string, references []string, n int) (precision, recall float64) {
	candidateNgrams := getNgrams(tokenizeSentence(candidate), n)
	referenceNgrams := make([]ngram, len(references))
	for i, ref := range references {
		referenceNgrams[i] = getNgrams(tokenizeSentence(ref), n)
	}

	return computeScoreForNgrams(candidateNgrams, referenceNgrams)
}

func computeScoreForNgrams(candidateNgrams ngram, referenceNgrams []ngram) (precision, recall float64) {
	for i := range referenceNgrams {
		overlapping := countOverlappingWords(candidateNgrams, referenceNgrams[i])
		p, r := calculatePrecisionRecall(overlapping, lenf(candidateNgrams), lenf(referenceNgrams[i]))
		precision += p
		recall += r
	}
	// TODO: Need to get the average of precision/recall?
	return precision, recall
}

// Computes rouge-s score for the given sentencens based on "skip-ngrams" (word pairs which allow n gaps between the two words.)
func RougeS(candidate string, references []string, n int) (precision, recall float64) {
	candidateSkipGrams := getSkipGrams(tokenizeSentence(candidate), n)
	referenceSkipGrams := make([]ngram, len(references))
	for i, ref := range references {
		referenceSkipGrams[i] = getSkipGrams(tokenizeSentence(ref), n)
	}

	return computeScoreForNgrams(candidateSkipGrams, referenceSkipGrams)
}

// Computes rouge-l score for the given sentences (based on longest common subsequence)
func RougeL(candidate string, references []string) (precision, recall float64) {
	candidateTokenized := tokenizeSentence(candidate)
	for _, ref := range references {
		refTokenized := tokenizeSentence(ref)
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
