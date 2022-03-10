package typeclasses

import "returntypes-langserver/common/configuration"

// Contains all configured type classes
type TypeClassConfiguration struct {
	// The configured type classes
	Classes []configuration.TypeClass `json:"classes"`
	// A type class which is defined as an array type
	ArrayType *configuration.TypeClass
	// A type class which is defined as a chain method type
	ChainMethodType *configuration.TypeClass
	// A type class which contains java.lang.Object
	DefaultType *configuration.TypeClass
}
