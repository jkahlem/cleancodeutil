package somepackage

type TestStruct struct {
	// Multi line
	// Documentation
	field1         string
	Field2         func(par1, par2 string, par3 int) (res1, res2 bool, res3 string) // in line comment
	Field3, Field4 string                                                           `tagged:"value"`
}

type TestInterface interface {
	// Multi line
	// Documentation
	SampleMethod(string, int) error
}

// Multi line
// Documentation
func (t *TestStruct) SampleMethod(par1 string, par2 int) error {
	return nil
}
