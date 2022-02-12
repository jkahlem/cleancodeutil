package excel

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStructWithHeaders struct {
	Name string `excel:"NAME"`
	// No tag defined: use empty header
	FieldWithEmptyHeader string
	Number               int    `excel:"number"`
	Text                 string `excel:"Text"`
}

type InfoCaptor struct {
	layout Layout
}

func (w *InfoCaptor) Write(record []string) errors.Error {
	return nil
}

func (w *InfoCaptor) BuildLayout(layout Layout) errors.Error {
	w.layout = layout
	return nil
}

func (w *InfoCaptor) SetWriter(writer StreamWriter) {}

func TestBuildHeaderByStruct(t *testing.T) {
	// given
	captor := &InfoCaptor{}
	w := newStructFormatWriter(TestStructWithHeaders{})
	w.SetWriter(captor)

	// when
	err := w.BuildLayout(EmptyLayout())

	// then
	assert.NoError(t, err)
	utils.AssertStringSlice(t, getHeaderStringsFromLayout(captor.layout), "NAME", "", "number", "Text")
}
