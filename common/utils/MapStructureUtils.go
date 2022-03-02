package utils

import (
	"returntypes-langserver/common/debug/errors"

	"github.com/mitchellh/mapstructure"
)

// Like mapstructure's Decode method, but reuses the "json" tag on the struct fields
func DecodeMapToStruct(input, output interface{}) errors.Error {
	return DecodeMapToStructWithConfiguration(input, output, mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   output,
		TagName:  "json",
	})
}

// Like mapstructure's Decode method, but reuses the "json" tag on the struct fields and erroring on unknown fields
func DecodeMapToStructStrict(input, output interface{}) errors.Error {
	return DecodeMapToStructWithConfiguration(input, output, mapstructure.DecoderConfig{
		Metadata:    nil,
		Result:      output,
		TagName:     "json",
		ErrorUnused: true,
	})
}

func DecodeMapToStructWithConfiguration(input, output interface{}, config mapstructure.DecoderConfig) errors.Error {
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return errors.Wrap(err, "Error", "Could not convert map to struct")
	}
	if err := decoder.Decode(input); err != nil {
		return errors.Wrap(err, "Error", "Could not convert map to struct")
	}
	return nil
}
