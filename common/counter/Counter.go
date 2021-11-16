// Counter for debugging/statistics/analytics purposes which can be used globally.
// The same counter can be accessed by using the same identifier, so it can be used everywhere in the code.
package counter

type Counter int

var counterMap map[string]*Counter

// Returns a global counter for the given identifier. Creates a new one if it does not exist.
func For(identifier string) *Counter {
	if counterMap == nil {
		counterMap = make(map[string]*Counter)
	}
	if counter, exists := counterMap[identifier]; !exists {
		var cnt Counter = Counter(0)
		counterMap[identifier] = &cnt
		return counterMap[identifier]
	} else {
		return counter
	}
}

// Counts one up.
func (c *Counter) CountUp() *Counter {
	*c++
	return c
}

// Returns the count.
func (c *Counter) GetCount() int {
	return int(*c)
}

// Sets the count to 0.
func (c *Counter) Reset() *Counter {
	*c = 0
	return c
}
