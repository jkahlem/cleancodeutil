package utils

import (
	"reflect"
	"returntypes-langserver/common/debug/errors"

	"github.com/mitchellh/mapstructure"
)

type ValueDecoder interface {
	DecodeValue(interface{}) (interface{}, error)
}

func valueDecoderHook(sourceType, destinationType reflect.Type, sourceValue interface{}) (interface{}, error) {
	var valueDecoder ValueDecoder
	valueDecoderType := reflect.TypeOf(&valueDecoder).Elem()
	if destinationType.Implements(valueDecoderType) {
		v := reflect.New(destinationType).Interface()
		if valueDecoder, ok := v.(ValueDecoder); ok {
			return valueDecoder.DecodeValue(sourceValue)
		}
	}
	return sourceValue, nil
}

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
	config.DecodeHook = valueDecoderHook
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return errors.Wrap(err, "Error", "Could not convert map to struct")
	}
	if err := decoder.Decode(input); err != nil {
		return errors.Wrap(err, "Error", "Could not convert map to struct")
	}
	return nil
}
