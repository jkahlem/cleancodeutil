package git

import (
	"encoding/json"
	"reflect"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"strings"
)

// Contains repository definitions
type RepositoryList struct {
	Repositories []RepositoryDefinition
}

// Defines a repository to clone
type RepositoryDefinition struct {
	Url     string `json:"url"`
	DirName string `json:"dirName"`
}

// Unmarshals the passed raw data of the repository list.
func unmarshalRepositoryList(content []byte) (RepositoryList, errors.Error) {
	list := RepositoryList{}
	jsonObj := make(map[string]interface{})
	if err := json.Unmarshal(content, &jsonObj); err != nil {
		return list, errors.Wrap(err, CloneErrorTitle, "Could not read repository list")
	}
	if raw, ok := jsonObj["repositories"]; ok {
		repositories, err := unmarshalRepositoryListEntries(raw)
		if err != nil {
			return list, err
		}
		list.Repositories = repositories
	}

	return list, nil
}

// Unmarshals entries of a repository list.
func unmarshalRepositoryListEntries(repositories interface{}) ([]RepositoryDefinition, errors.Error) {
	sliceValue := reflect.ValueOf(repositories)
	if sliceValue.Kind() != reflect.Slice {
		return nil, errors.New(CloneErrorTitle, "Could not read repository list. Expected array type for entry 'repositories'.")
	}
	resultSlice := make([]RepositoryDefinition, sliceValue.Len())
	for i := 0; i < sliceValue.Len(); i++ {
		value := sliceValue.Index(i)
		entry, err := unmarshalRepositoryListEntry(value)
		if err != nil {
			log.ReportProblemWithError(err, "Malformed repository list entry detected")
			continue
		}
		resultSlice[i] = entry
	}
	return resultSlice, nil
}

// Unmarshals a repository list entry which may be either a string or a object.
func unmarshalRepositoryListEntry(value reflect.Value) (RepositoryDefinition, errors.Error) {
	value = unwrap(value)
	switch value.Kind() {
	case reflect.String:
		return unmarshalRepositoryListEntryFromString(value.String())
	case reflect.Map:
		return unmarshalRepositoryListEntryFromMap(value)
	default:
		return RepositoryDefinition{}, errors.New(CloneErrorTitle, "Could not read repository list. Entry needs to be of type string or object.")
	}
}

// Unmarshals a repository list entry which is defined as a string
func unmarshalRepositoryListEntryFromString(url string) (RepositoryDefinition, errors.Error) {
	if !isSupportedUrl(url) {
		return RepositoryDefinition{}, errors.New(CloneErrorTitle, "Unsupported repository url: "+url)
	}
	_, dirName := getOwnerAndRepositoryFromURL(url)
	return RepositoryDefinition{
		Url:     url,
		DirName: dirName,
	}, nil
}

// Unmarshals a repository list entry which is defined as an object
func unmarshalRepositoryListEntryFromMap(mapValue reflect.Value) (RepositoryDefinition, errors.Error) {
	urlValue, dirNameValue := getMapEntry(mapValue, "url"), getMapEntry(mapValue, "dirName")
	if urlValue.Kind() != reflect.String {
		return RepositoryDefinition{}, errors.New(CloneErrorTitle, "Could not read repository list. Entry needs a field 'url' of type string.")
	}

	def, err := unmarshalRepositoryListEntryFromString(urlValue.String())
	if err != nil {
		return def, err
	}

	if dirNameValue.Kind() == reflect.String {
		def.DirName = dirNameValue.String()
	}
	return def, nil
}

// Returns true if the url is supported. (Currently only github with https protocol)
func isSupportedUrl(url string) bool {
	return len(url) > 0 && strings.HasPrefix(url, "https://github.com")
}

func getMapEntry(mapValue reflect.Value, key string) reflect.Value {
	return unwrap(mapValue.MapIndex(reflect.ValueOf(key)))
}

func unwrap(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Interface {
		value = value.Elem()
	}
	return value
}
