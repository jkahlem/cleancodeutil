package typeclasses

// Contains all configured type classes
type TypeClassConfiguration struct {
	// The configured type classes
	Classes []TypeClass `json:"classes"`
	// A type class which is defined as an array type
	ArrayType *TypeClass
	// A type class which is defined as a chain method type
	ChainMethodType *TypeClass
	// A type class which contains java.lang.Object
	DefaultType *TypeClass
}

type TypeClass struct {
	// The name of the type class
	Label string `json:"label"`
	// If true, then this type class is used as a type for all array types (may be only used for max. of one type class)
	IsArrayType bool `json:"isArrayType"`
	// If true, then this type class is used as a type for all methods which are chain methods (may be only used for max. of one type class)
	IsChainMethodType bool `json:"isChainMethodType"`
	// A list of canonical names of classes/types which belong to this type class including the ones extending or implementing them
	Elements []string `json:"elements"`
	// The color used for this type class for visualization
	Color string `json:"color"`
}
