package metrics

type DataGroup struct {
	Candidate  *Sentence
	References []*Sentence
}

type Sentence struct {
	tokens []string
	ngrams map[int]Ngram
	sgrams map[int]Ngram
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
