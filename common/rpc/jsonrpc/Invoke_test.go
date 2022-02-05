package jsonrpc

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeFunctionSimple(t *testing.T) {
	// given
	called := false
	FunctionSimple := func() { called = true }

	// when
	_, err := Invoke(funcOf(FunctionSimple, ""), nil)

	// then
	assert.Nil(t, err)
	assert.True(t, called)
}

func TestInvokeFunctionWithSingleParameter(t *testing.T) {
	// given
	called := false
	par1val := 0
	FunctionWithSingleParameter := func(par1 int) {
		called = true
		par1val = par1
	}

	// when
	_, err := Invoke(funcOf(FunctionWithSingleParameter, "par1"), Params(10))

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.Equal(t, 10, par1val)
}

func TestInvokeFunctionWithSingleParameterByName(t *testing.T) {
	// given
	called := false
	par1val := 0
	FunctionWithSingleParameter := func(par1 int) {
		called = true
		par1val = par1
	}
	paramsMapJson := `{"par1":10}`
	var mapObj interface{}
	json.Unmarshal([]byte(paramsMapJson), &mapObj)

	// when
	_, err := Invoke(funcOf(FunctionWithSingleParameter, "par1"), mapObj)

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.Equal(t, 10, par1val)
}

func TestInvokeFunctionWithMultipleParameter(t *testing.T) {
	// given
	called := false
	par1val, par2val := 0, ""
	FunctionWithMultipleParameter := func(par1 int, par2 string) {
		called = true
		par1val, par2val = par1, par2
	}

	// when
	_, err := Invoke(funcOf(FunctionWithMultipleParameter, "par1,par2"), Params(10, "testtest"))

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.Equal(t, 10, par1val)
	assert.Equal(t, "testtest", par2val)
}

func TestInvokeFunctionWithComplexParameter(t *testing.T) {
	// given
	called := false
	par1val := TestStruct{}
	FunctionWithComplexParameter := func(par1 TestStruct) {
		called = true
		par1val = par1
	}
	structAsJson := `{"id":5,"information":{"name":"testtest"},"optional":{"name":"test"}}`
	var structObj interface{}
	json.Unmarshal([]byte(structAsJson), &structObj)

	// when
	_, err := Invoke(funcOf(FunctionWithComplexParameter, "par1"), Params(structObj))

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.Equal(t, 5, par1val.Id)
	assert.Equal(t, "testtest", par1val.Information.Name)
	assert.Equal(t, "test", par1val.Optional.Name)
}

func TestInvokeFunctionWithPointerParameter(t *testing.T) {
	// given
	called := false
	var par1val, par2val *TestStruct
	FunctionWithPointerParameter := func(par1 *TestStruct, par2 *TestStruct) {
		called = true
		par1val = par1
		par2val = par2
	}
	structAsJson := `{"id":5,"information":{"name":"testtest"}}`
	var structObj interface{}
	json.Unmarshal([]byte(structAsJson), &structObj)

	// when
	_, err := Invoke(funcOf(FunctionWithPointerParameter, "par1,par2"), Params(nil, structObj))

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.Nil(t, par1val)
	assert.NotNil(t, par2val)
	assert.Equal(t, 5, par2val.Id)
	assert.Equal(t, "testtest", par2val.Information.Name)
	assert.Nil(t, par2val.Optional)
}

func TestInvokeFunctionWithReturnValue(t *testing.T) {
	// given
	called := false
	FunctionWithReturnValue := func() int {
		called = true
		return 10
	}

	// when
	result, err := Invoke(funcOf(FunctionWithReturnValue, ""), nil)

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.Equal(t, 10, result)
}

func TestInvokeFunctionWithMultipleReturnValues(t *testing.T) {
	// given
	called := false
	FunctionWithMultipleReturnValues := func() (int, string) {
		called = true
		return 10, "testtest"
	}

	// when
	result, err := Invoke(funcOf(FunctionWithMultipleReturnValues, ""), nil)
	asSlice, isSlice := result.([]interface{})

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.True(t, isSlice)
	assert.Len(t, asSlice, 2)
	assert.Equal(t, 10, asSlice[0])
	assert.Equal(t, "testtest", asSlice[1])
}

func TestInvokeFunctionWithErrorReturnValue(t *testing.T) {
	// given
	called := false
	FunctionWithErrorReturnValue := func() error {
		called = true
		return errors.New("errormsg")
	}

	// when
	result, err := Invoke(funcOf(FunctionWithErrorReturnValue, ""), nil)

	// then
	assert.True(t, called)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestInvokeFunctionWithMixedSingleReturnValue(t *testing.T) {
	// given
	called1 := false
	called2 := false
	FunctionReturningError := func() (int, error) {
		called1 = true
		return 0, errors.New("errormsg")
	}
	FunctionReturningInt := func() (int, error) {
		called2 = true
		return 10, nil
	}

	// when
	_, err1 := Invoke(funcOf(FunctionReturningError, ""), nil)
	result, err2 := Invoke(funcOf(FunctionReturningInt, ""), nil)

	// then
	assert.True(t, called1)
	assert.NotNil(t, err1)
	assert.True(t, called2)
	assert.Nil(t, err2)
	assert.Equal(t, 10, result)
}

func TestInvokeFunctionWithMixedMultipleReturnValue(t *testing.T) {
	// given
	called1 := false
	called2 := false
	FunctionReturningError := func() (int, string, error) {
		called1 = true
		return 0, "empty", errors.New("errormsg")
	}
	FunctionReturningValues := func() (int, string, error) {
		called2 = true
		return 10, "testtest", nil
	}

	// when
	_, err1 := Invoke(funcOf(FunctionReturningError, ""), nil)
	result, err2 := Invoke(funcOf(FunctionReturningValues, ""), nil)
	asSlice, isSlice := result.([]interface{})

	// then
	assert.True(t, called1)
	assert.NotNil(t, err1)
	assert.True(t, called2)
	assert.Nil(t, err2)
	assert.True(t, isSlice)
	assert.Len(t, asSlice, 2)
	assert.Equal(t, 10, asSlice[0])
	assert.Equal(t, "testtest", asSlice[1])
}

func TestInvokeFunctionWithComplexReturnValue(t *testing.T) {
	// given
	called := false
	FunctionWithComplexReturnValue := func() TestStruct {
		called = true
		return TestStruct{
			Id: 5,
			Information: TestStructNested{
				Name: "testtest",
			},
		}
	}

	// when
	result, err := Invoke(funcOf(FunctionWithComplexReturnValue, ""), nil)
	asStruct, isStruct := result.(TestStruct)

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.True(t, isStruct)
	assert.Equal(t, 5, asStruct.Id)
	assert.Equal(t, "testtest", asStruct.Information.Name)
}

func TestInvokeFunctionWithPointerReturnValue(t *testing.T) {
	// given
	called := false
	FunctionWithComplexReturnValue := func() *TestStruct {
		called = true
		return &TestStruct{
			Id: 5,
			Information: TestStructNested{
				Name: "testtest",
			},
		}
	}

	// when
	result, err := Invoke(funcOf(FunctionWithComplexReturnValue, ""), nil)
	asStructPtr, isTypeOk := result.(*TestStruct)

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.True(t, isTypeOk)
	assert.NotNil(t, asStructPtr)
	assert.Equal(t, 5, asStructPtr.Id)
	assert.Equal(t, "testtest", asStructPtr.Information.Name)
	assert.Nil(t, asStructPtr.Optional)
}

func TestInvokeFunctionWithNilPointerReturnValue(t *testing.T) {
	// given
	called := false
	FunctionWithComplexReturnValue := func() *TestStruct {
		called = true
		return nil
	}

	// when
	result, err := Invoke(funcOf(FunctionWithComplexReturnValue, ""), nil)
	asNilPtr, isTypeOk := result.(*TestStruct)

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	assert.True(t, isTypeOk)
	assert.Nil(t, asNilPtr)
}

// Test relevant structures

type TestStruct struct {
	Id          int                       `mapstructure:"id"`
	Information TestStructNested          `mapstructure:"information"`
	Optional    *TestStructNestedOptional `mapstructure:"optional"`
}

type TestStructNested struct {
	Name string `mapstructure:"name"`
}

type TestStructNestedOptional struct {
	Name string `mapstructure:"name"`
}

// Helper functions

func funcOf(fn interface{}, params string) *Function {
	f := &Function{
		Fn: reflect.ValueOf(fn),
	}
	f.SetParams(params)
	return f
}

func Params(params ...interface{}) []interface{} {
	return params
}
