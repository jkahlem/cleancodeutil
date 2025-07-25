package predictor

import (
	"regexp"
	"strings"
	"unicode"
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
	tokens := re.FindAllString(name, -1)

	// Sentence is a slice of each word in the final sentence. Each token in tokens can consist of max. 2 words.
	sentence := make([]string, len(tokens)*2)
	// Offset is incremented if a token consist of two words
	offset := 0

WordLoop:
	for i, token := range tokens {
		thisToken := token

		if unicode.IsDigit(rune(thisToken[0])) {
			runes := []rune(thisToken)
			thisDigit := thisToken
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
			}
			sentence[pos] = thisDigit
		} else {
			// trim trailing underscores
			if !unicode.IsLetter(rune(thisToken[0])) {
				thisToken = thisToken[1:]
			}

			runes := []rune(thisToken)
			for j, char := range runes {
				// if the first lower char is not the first/second lower char in the word
				// then thisWord is a combination of an abbreviation and a word (like "RGBTo" in "RGBToHSL")
				// In the case that the lower char is an 's' (like "URLs"), then it is probably a plural-'s'. (or might be a word like "As"...)
				// Otherwise, split before the last uppercase char
				if unicode.IsLower(char) {
					if j > 1 && char != 's' {
						// Split the token into two words and increment the offset
						sentence[i+offset] = string(runes[:j-1])
						sentence[i+offset+1] = string(runes[j-1:])
						offset++
						continue WordLoop
					}
					break
				}
			}
			sentence[i+offset] = thisToken
		}
	}

	sentence = sentence[:len(tokens)+offset]
	return strings.Join(sentence, " ")
}
