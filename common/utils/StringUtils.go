package utils

import (
	"reflect"
	"strings"
)

// Explodes slices in a list of arguments to the same level of the other arguments (but not for nested slices).
// This helps when calling variadic functions with both enumerating elements and slices to explode which is not
// supported in go by default. Example:
//
//   // Joins words into one string by a whitespace " "
//   func MakeSentence(words string...) string { /* ... */ }
//
//   otherWords := []string{"in", "mixed", "usage"}
//
//   MakeSentence("Pass", "strings", otherWords...)   // Will result in a compile-time error
//   MakeSentence(ExplodeSlices("Pass", "strings", otherWords...)) // Returns "Pass strings in mixed usage"
func ExplodeSlices(args ...interface{}) []interface{} {
	params := make([]interface{}, 0, len(args))
	for i := range args {
		value := UnwrapInterface(reflect.ValueOf(args[i]))
		if value.Kind() == reflect.Slice {
			for j := 0; j < value.Len(); j++ {
				sliceValue := UnwrapInterface(value.Index(j))
				params = append(params, sliceValue.Interface())
			}
		} else {
			params = append(params, value.Interface())
		}
	}
	return params
}

// Splits the passed string to a key value pair. If the string has no key value pair, ```ok``` will be false.
func KeyValueByEqualSign(raw string) (key, value string, ok bool) {
	splitted := strings.Split(raw, "=")
	if len(splitted) != 2 {
		return "", "", false
	}
	return splitted[0], splitted[1], true
}
