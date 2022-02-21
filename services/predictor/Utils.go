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
	re := regexp.MustCompile(`((^|[^a-zA-Z0-9])([A-Z]+|[a-z])[a-z]*|[A-Z]+[a-z]*|\d+([a-zA-Z][a-z]*)?)`)
	words := re.FindAllString(name, -1)

	// Worst case: Each word in words contains two words
	sentence := make([]string, len(words)*2)
	offset := 0

WordLoop:
	for i, word := range words {
		thisWord := word

		if unicode.IsDigit(rune(thisWord[0])) {
			runes := []rune(thisWord)
			thisDigit := thisWord
			pos := i + offset
			isFollowedByWord := false
			for j, char := range runes {
				if !unicode.IsDigit(char) {
					thisDigit = string(runes[:j])
					sentence[pos+1] = string(runes[j:])
					offset++
					isFollowedByWord = true
					break
				}
			}
			if thisDigit == "2" && isFollowedByWord {
				thisDigit = "to"
			} else if thisDigit == "4" && isFollowedByWord {
				thisDigit = "for"
			}
			sentence[pos] = thisDigit
		} else {
			// trim trailing underscores
			if !unicode.IsLetter(rune(thisWord[0])) {
				thisWord = thisWord[1:]
			}

			runes := []rune(thisWord)
			for j, char := range runes {
				// if the first lower char is not the first/second lower char in the word
				// then thisWord is a combination of an abbreviation and a word (like "RGBTo" in "RGBToHSL")
				// In the case that the lower char is an 's' (like "URLs"), then it is probably a plural-'s'. (or might be a word like "As"...)
				// Otherwise, split before the last uppercase char
				if unicode.IsLower(char) && char != 's' {
					if j > 1 {
						sentence[i+offset] = string(runes[:j-1])
						sentence[i+offset+1] = string(runes[j-1:])
						offset++
						continue WordLoop
					}
					break
				}
			}
			sentence[i+offset] = thisWord
		}
	}

	sentence = sentence[:len(words)+offset]
	return strings.Join(sentence, " ")
}

// Returns true if a predictor script is configured
func isPredictorScriptSet() bool {
	return len(configuration.PredictorScriptPath()) > 0
}
