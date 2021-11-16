package predictor

import (
	"regexp"
	"strings"
	"unicode"

	"returntypes-langserver/common/configuration"
)

type PredictableMethodName string

// Converts a string into a predictable method name.
func GetPredictableMethodName(name string) PredictableMethodName {
	sentence := SplitMethodNameToSentence(name)
	return PredictableMethodName(strings.ToLower(sentence))
}

// Splits a method name into it's words and makes a sentence of it.
func SplitMethodNameToSentence(name string) string {
	// go does not support lookahead/lookbehind, so first do a more naive word split
	re := regexp.MustCompile(`((^|[^a-zA-Z])([A-Z]+|[a-z])[a-z]*|[A-Z]+[a-z]*|\d+)`)
	words := re.FindAllString(name, -1)

	// Worst case: Each word in words contains two words
	sentence := make([]string, len(words)*2)
	offset := 0

WordLoop:
	for i, word := range words {
		thisWord := word

		// trim trailing underscores
		if !unicode.IsLetter(rune(thisWord[0])) {
			thisWord = thisWord[1:]
		}

		runes := []rune(thisWord)
		for j, char := range runes {
			// if the first lower char is not the first/second lower char in the word
			// then thisWord is a combination of an abbreviation and a word (like "RGBTo" in "RGBToHSL")
			// in this case, split before the last uppercase char
			if unicode.IsLower(char) {
				if j > 1 {
					sentence[i+offset] = string(runes[0 : j-1])
					sentence[i+offset+1] = string(runes[j-1:])
					offset++
					continue WordLoop
				}
				break
			}
		}
		sentence[i+offset] = thisWord
	}

	sentence = sentence[:len(words)+offset]
	return strings.Join(sentence, " ")
}

// Returns true if a predictor script is configured
func isPredictorScriptSet() bool {
	return len(configuration.PredictorScriptPath()) > 0
}
