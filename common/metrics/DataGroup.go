package metrics

type WordCount map[string]int

type DataGroup struct {
	Candidate  *Sentence
	References []*Sentence
}

type Sentence struct {
	plain       string
	tokens      []string
	ngrams      map[int]Ngram
	sgrams      map[int]Ngram
	ngramsWords map[int]WordCount
}

func NewDataGroup(candidate string, references []string) DataGroup {
	dataGroup := DataGroup{
		Candidate:  NewSentence(candidate),
		References: make([]*Sentence, len(references)),
	}
	for i, str := range references {
		dataGroup.References[i] = NewSentence(str)
	}
	return dataGroup
}

func NewSentence(str string) *Sentence {
	return &Sentence{
		plain:  str,
		tokens: tokenizeSentence(str),
	}
}

func (s *Sentence) Ngram(n int) Ngram {
	if s.ngrams == nil {
		s.ngrams = make(map[int]Ngram)
	} else if val, ok := s.ngrams[n]; ok {
		return val
	}
	val := getNgrams(s.tokens, n)
	s.ngrams[n] = val
	return val
}

// Returns a WordCount value containing mappings between each token and how often it occurs in the sequence
func (s *Sentence) NgramWordCount(n int) WordCount {
	if s.ngramsWords == nil {
		s.ngramsWords = make(map[int]WordCount)
	} else if val, ok := s.ngramsWords[n]; ok {
		return val
	}
	s.ngramsWords[n] = countWords(s.Ngram(n))
	return s.ngramsWords[n]
}

func (s *Sentence) Sgram(n int) Ngram {
	if s.sgrams == nil {
		s.sgrams = make(map[int]Ngram)
	} else if val, ok := s.sgrams[n]; ok {
		return val
	}
	val := getSkipGrams(s.tokens, n)
	s.sgrams[n] = val
	return val
}

func (s *Sentence) Tokens() []string {
	return s.tokens
}

func (s *Sentence) String() string {
	return s.plain
}
