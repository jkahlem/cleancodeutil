package utils

import (
	"encoding/json"
	"returntypes-langserver/common/debug/errors"
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
