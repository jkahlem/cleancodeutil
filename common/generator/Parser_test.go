package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	// given
	src := `package somepackage
	type TestStruct struct {
		field1 string
		Field2 func(par1, par2 string, par3 int) (res1, res2 bool, res3 string)
	}`

	// when
	ctx, err := ParseSourceCode(src)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, ctx.fileNode)
	assert.Equal(t, src, ctx.sourceCode)
}

func TestStructParsing(t *testing.T) {
	// given
	ctx, _ := ParseFile("_testFile.go")

	// when
	structs := ctx.ParseStructs()

	// then
	assert.Len(t, structs, 1)

	testStruct := structs[0]
	assert.Equal(t, "TestStruct", testStruct.Name)
	assert.Len(t, testStruct.Fields, 4)

	field1 := testStruct.Fields[0]
	assert.Equal(t, "field1", field1.Name)
	assert.Equal(t, "Multi line\nDocumentation\n", field1.Documentation)
	assert.Equal(t, "string", field1.Type.Code())

	field2 := testStruct.Fields[1]
	assert.Equal(t, "Field2", field2.Name)
	assert.Equal(t, "in line comment\n", field2.LineComment)
	assert.Equal(t, "func(par1, par2 string, par3 int) (res1, res2 bool, res3 string)", field2.Type.Code())

	field3 := testStruct.Fields[2]
	assert.Equal(t, "Field3", field3.Name)
	assert.Equal(t, `tagged:"value"`, field3.Tag)
	assert.Equal(t, "string", field3.Type.Code())

	field4 := testStruct.Fields[3]
	assert.Equal(t, "Field4", field4.Name)
	assert.Equal(t, `tagged:"value"`, field4.Tag)
	assert.Equal(t, "string", field4.Type.Code())
}

func TestFunctionTypeParsing(t *testing.T) {
	// given
	ctx, _ := ParseFile("_testFile.go")
	structs := ctx.ParseStructs()
	// the field type is func(par1, par2 string, par3 int) (res1, res2 bool, res3 string)
	fieldType := structs[0].Fields[1].Type

	// when
	fnType, ok := fieldType.FunctionType()

	// then
	assert.True(t, ok)

	assert.Len(t, fnType.In, 3)
	assert.Equal(t, "par1", fnType.In[0].Name)
	assert.Equal(t, "string", fnType.In[0].Type.Code())
	assert.Equal(t, "par2", fnType.In[1].Name)
	assert.Equal(t, "string", fnType.In[1].Type.Code())
	assert.Equal(t, "par3", fnType.In[2].Name)
	assert.Equal(t, "int", fnType.In[2].Type.Code())

	assert.Len(t, fnType.Out, 3)
	assert.Equal(t, "res1", fnType.Out[0].Name)
	assert.Equal(t, "bool", fnType.Out[0].Type.Code())
	assert.Equal(t, "res2", fnType.Out[1].Name)
	assert.Equal(t, "bool", fnType.Out[1].Type.Code())
	assert.Equal(t, "res3", fnType.Out[2].Name)
	assert.Equal(t, "string", fnType.Out[2].Type.Code())
}
