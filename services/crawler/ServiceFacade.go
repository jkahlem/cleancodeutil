package crawler

import (
	"sync"

	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/java"
)

var singleton *crawler
var singletonMutex sync.Mutex

// Gets the content of one java file.
func GetCodeElements(path string, options Options) (java.FileContainer, errors.Error) {
	return getSingleton().GetCodeElements(path, options)
}

// Gets the content of all java files in the specified directory.
func GetCodeElementsOfDirectory(path string, options Options) (java.FileContainer, errors.Error) {
	return getSingleton().GetCodeElementsOfDirectory(path, options)
}

// Gets the content of all java files in the specified directory.
func GetRawCodeElementsOfDirectory(path string, options Options) (string, errors.Error) {
	return getSingleton().GetRawCodeElementsOfDirectory(path, options)
}

func getSingleton() *crawler {
	singletonMutex.Lock()
	defer singletonMutex.Unlock()

	if singleton == nil {
		singleton = createSingleton()
	}
	return singleton
}

func createSingleton() *crawler {
	return &crawler{}
}
