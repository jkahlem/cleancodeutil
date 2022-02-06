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
