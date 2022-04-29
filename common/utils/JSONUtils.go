package utils

import (
	"encoding/json"
	"returntypes-langserver/common/debug/errors"
	"strings"
)

func UnmarshalJSONStrict(data []byte, v interface{}) errors.Error {
	decoder := json.NewDecoder(&ByteReader{
		bytes: data,
	})
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return errors.Wrap(err, "JSON Error", "Could not unmarshal JSON")
	}
	return nil
}

type ByteReader struct {
	bytes []byte
}

func (r *ByteReader) Read(p []byte) (int, error) {
	lengthToRead := len(r.bytes)
	if lengthToRead > len(p) {
		lengthToRead = len(p)
	}
	copy(p, r.bytes[:lengthToRead])
	r.bytes = r.bytes[lengthToRead:]
	return lengthToRead, nil
}

// Helper to unmarshal static json strings to maps, especially useful for tests
func MustUnmarshalJsonToMap(raw string) map[string]interface{} {
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		panic(err)
	}
	return v
}

// Adds all values in destination if they are set in source and unset in destination
func AddUnsettedValues(source, destination map[string]interface{}) {
	for key, value := range source {
		if _, ok := destination[key]; !ok {
			destination[key] = value
		}
	}
}

// Gets a value inside a (json) object. To access fields of nested objects, these objects must be of
// type map[string]interface{} (which is usually true for unmarshalled json objects).
// The identifier is a list of the fields to access concatenated with a dot. Fields containing a dot
// in their name are not supported / cannot be accessed.
// If the value of any field defined in the identifier is nil, the return value will also be nil and
// no error is returned.
func GetNestedValueOfMap(object map[string]interface{}, identifier string) (interface{}, errors.Error) {
	splitted := strings.Split(identifier, ".")
	if value, ok := object[splitted[0]]; ok && value != nil {
		if len(splitted) == 1 {
			return value, nil
		} else if subMap, ok := value.(map[string]interface{}); ok && subMap != nil {
			return GetNestedValueOfMap(subMap, strings.Join(splitted[1:], "."))
		} else {
			return nil, errors.New("Error", "Nested value is not an object")
		}
	}
	return nil, nil
}

func SetNestedValueOfMap(object map[string]interface{}, identifier string, value interface{}) errors.Error {
	if object == nil {
		return errors.New("Error", "Cannot set value in nil map.")
	}

	splitted := strings.Split(identifier, ".")
	if len(splitted) == 1 {
		object[splitted[0]] = value
		return nil
	}
	if value, ok := object[splitted[0]]; !ok || value == nil {
		object[splitted[0]] = make(map[string]interface{})
	}

	if subMap, ok := object[splitted[0]].(map[string]interface{}); !ok {
		return errors.New("Error", "Cannot set field of value which is not of type map[string]interface{}")
	} else {
		return SetNestedValueOfMap(subMap, strings.Join(splitted[1:], "."), value)
	}
}

func DeleteNestedFieldOfMap(object map[string]interface{}, identifier string) errors.Error {
	if object == nil {
		return errors.New("Error", "Cannot set value in nil map.")
	}

	splitted := strings.Split(identifier, ".")
	if len(splitted) == 1 {
		delete(object, splitted[0])
		return nil
	}
	if value, ok := object[splitted[0]]; !ok || value == nil {
		object[splitted[0]] = make(map[string]interface{})
	}

	if subMap, ok := object[splitted[0]].(map[string]interface{}); !ok {
		return errors.New("Error", "Cannot set field of value which is not of type map[string]interface{}")
	} else {
		return DeleteNestedFieldOfMap(subMap, strings.Join(splitted[1:], "."))
	}
}
