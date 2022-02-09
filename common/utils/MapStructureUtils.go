package utils

import "github.com/mitchellh/mapstructure"

// Like mapstructure's Decode method, but reuses the "json" tag on the struct fields
func DecodeMapToStruct(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   output,
		TagName:  "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
