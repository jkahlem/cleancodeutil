package utils

import (
	"fmt"
	"reflect"

	"returntypes-langserver/common/debug/errors"
)

var InvalidValues = errors.ErrorId("Error", "Invalid values")

// Returns true if the type is implementing the error interface.
func IsErrorType(t reflect.Type) bool {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	return t.Implements(errorType)
}

// Unwraps the value as long as it is an interface type. Does nothing, if the value is no interface.
// Panics if value is not valid.
func UnwrapInterface(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Interface {
		value = value.Elem()
	}
	return value
}

// Unwraps a reflected type from any pointers / interfaces "on top" of it and returns it.
// For example having an interface{} which is a pointer to a struct will be unwrapped to the struct type.
// Just returns the type if it is not wrapped and returns nil if type is also nil.
func UnwrapType(typ reflect.Type) reflect.Type {
	if typ == nil {
		return nil
	} else if typ.Kind() == reflect.Interface || typ.Kind() == reflect.Ptr {
		return UnwrapType(typ.Elem())
	}
	return typ
}

// Returns true if value may be nil, so value.IsNil will not panic.
func MayBeNilValue(value reflect.Value) bool {
	if value.IsValid() {
		return MayBeNilKind(value.Kind())
	}
	return false
}

// Returns true if the kind is a nillable kind
func MayBeNilKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	}
	return false
}

// Tries to cast the value to the expected type by:
// - trying to set the nil value/zero value for the expected type if it is nillable
// - trying to cast it straight to the value (so the value type and expected type are the same)
// - trying to convert the type to the value and assign it then (e.g. convert float to int)
// - trying to map the value to a struct if the value is a map and the expectedType is a struct
// - trying to map the value to a slice if the value is an interface{} slice but the expected type is a typed slice (like []string)
//   This will also do a type check for each of the values in the slice.
// - trying to cast the value to the underlying type if expectedType is a pointer type and returns the pointer to the casted value.
// If all of these attempts are failing, it will return an invalid reflect value and an error.
func CastValueToTypeIfPossible(value reflect.Value, expectedType reflect.Type) (reflect.Value, errors.Error) {
	v := UnwrapInterface(value)
	if !v.IsValid() {
		if MayBeNilKind(expectedType.Kind()) {
			return reflect.Zero(expectedType), nil
		} else {
			return reflect.ValueOf(nil), errors.New("Type Error", fmt.Sprintf("Unexpected type: Expected %s but go invalid value.", expectedType))
		}
	}

	if v.Type().AssignableTo(expectedType) {
		return v, nil
	} else if v.Type().ConvertibleTo(expectedType) {
		return v.Convert(expectedType), nil
	} else if mapped, err := MapToStruct(v, expectedType); err == nil || !errors.Is(err, InvalidValues) {
		return mapped, err
	} else if slice, err := CopyToSliceType(v, expectedType); err == nil || !errors.Is(err, InvalidValues) {
		return slice, err
	} else if expectedType.Kind() == reflect.Ptr {
		if casted, err := CastValueToTypeIfPossible(v, expectedType.Elem()); err != nil {
			return casted, err
		} else if !casted.CanAddr() {
			return casted, errors.New("Reflect Error", "Cannot get an address for the casted value")
		} else {
			return casted.Addr(), err
		}
	} else {
		return reflect.ValueOf(nil), errors.New("Type Error", fmt.Sprintf("Unexpected type: Expected %s but got %s.", expectedType, v.Type()))
	}
}

// Converts a slice value to the target type. This is especially for converting an []interface{} slice to a typed slice.
// This function also checks if the source slice is valid and if both parameters are slices, otherwise it returns an error,
// so usages in if-clauses are possible (without a check before).
func CopyToSliceType(sourceSlice reflect.Value, targetType reflect.Type) (reflect.Value, errors.Error) {
	if !sourceSlice.IsValid() || sourceSlice.Kind() != reflect.Slice || targetType.Kind() != reflect.Slice {
		return reflect.Zero(targetType), errors.NewById(InvalidValues)
	}
	destination := reflect.MakeSlice(targetType, sourceSlice.Len(), sourceSlice.Cap())
	targetElemType := targetType.Elem()
	for i := 0; i < sourceSlice.Len(); i++ {
		if value, err := CastValueToTypeIfPossible(sourceSlice.Index(i), targetElemType); err != nil {
			return value, errors.Wrap(err, "Error", fmt.Sprintf("Incompatible types at index %d", i))
		} else {
			destination.Index(i).Set(value)
		}
	}
	return destination, nil
}

// Used to map a map (source) to a target struct (targetType). Does also check if the types matches
// so simple usages in if-clauses are possible (just need to check for the error).
func MapToStruct(source reflect.Value, targetType reflect.Type) (reflect.Value, errors.Error) {
	if !source.IsValid() || source.Kind() != reflect.Map || targetType.Kind() != reflect.Struct {
		return reflect.Zero(targetType), errors.NewById(InvalidValues)
	}
	destination := reflect.New(targetType)
	if err := DecodeMapToStruct(source.Interface(), destination.Interface()); err != nil {
		return reflect.Zero(targetType), errors.Wrap(err, "Error", fmt.Sprintf("Could not map json object to desired structure"))
	}
	return destination.Elem(), nil
}

// Wraps a value in a slice if it is not already a slice.
func WrapInSlice(value interface{}) interface{} {
	v := UnwrapInterface(reflect.ValueOf(value))
	if v.IsValid() && v.Kind() != reflect.Slice && (!MayBeNilValue(v) || !v.IsNil()) {
		slice := make([]interface{}, 1)
		slice[0] = value
		return slice
	}
	return value
}

// Maps a slice of reflect values to an interface{} slice.
func MapToInterfaceSlice(values []reflect.Value) []interface{} {
	destination := make([]interface{}, len(values))
	for i, value := range values {
		destination[i] = value.Interface()
	}
	return destination
}

type KeyValuePair struct {
	Key   reflect.Value
	Value reflect.Value
}

// Creates a key value pair channel of the given map which can be easily iterated over using the range keyword.
// The key/values of each key value pair are not wrapped in interfaces.
func MakeKeyValuePairChannel(mapValue reflect.Value) chan KeyValuePair {
	if mapValue.Kind() != reflect.Map {
		return nil
	}
	pairChannel := make(chan KeyValuePair, len(mapValue.MapKeys()))
	mapIterator := mapValue.MapRange()
	for mapIterator.Next() {
		pairChannel <- KeyValuePair{
			Key:   UnwrapInterface(mapIterator.Key()),
			Value: UnwrapInterface(mapIterator.Value()),
		}
	}
	close(pairChannel)
	return pairChannel
}

// Checks if the valueToCheck is a pointer to a value with the expected kind
func IsPtrTo(valueToCheck reflect.Value, expectedKind reflect.Kind) bool {
	if valueToCheck.Kind() != reflect.Ptr || !valueToCheck.Elem().IsValid() || valueToCheck.Elem().Kind() != expectedKind {
		return false
	}
	return true
}
