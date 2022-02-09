package git

import (
	"io/ioutil"
	"os"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
)

var loadedRepositoryList *RepositoryList

// Returns the loaded repository list
func GetRepositoryList() RepositoryList {
	if loadedRepositoryList == nil {
		return RepositoryList{}
	}
	return *loadedRepositoryList
}

// Loads a repository list
func LoadRepositoryList() errors.Error {
	if list, err := loadRepositoryListFromFile(configuration.ClonerRepositoryListPath()); err != nil {
		return err
	} else {
		loadedRepositoryList = &list
		return nil
	}
}

// Reads the urls from the git input file without already cloned repositories
func loadRepositoryListFromFile(path string) (RepositoryList, errors.Error) {
	content, err := readRepositoryList(path)
	if err != nil {
		return RepositoryList{}, err
	}

	return unmarshalRepositoryList(content)
}

// Reads the repository list file
func readRepositoryList(path string) ([]byte, errors.Error) {
	file, err := os.Open(configuration.ClonerRepositoryListPath())
	if err != nil {
		return nil, errors.Wrap(err, CloneErrorTitle, "Could not read repository list")
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, CloneErrorTitle, "Could not read repository list")
	}
	return content, nil
}
