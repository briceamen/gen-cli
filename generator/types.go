package generator

// Service represents a parsed SDK service interface
type Service struct {
	Name    string
	Methods []Method
}

// Method represents a method in a service interface
type Method struct {
	Name       string
	Params     []Param
	Returns    []Return
	HasContext bool
}

// Param represents a method parameter
type Param struct {
	Name string
	Type string
}

// Return represents a method return type
type Return struct {
	Type    string
	IsError bool
}

// StructField represents a field in a struct (for expanding opts structs)
type StructField struct {
	Name     string
	Type     string
	JSONTag  string
	Optional bool
}

// ParsedStruct represents a parsed struct definition
type ParsedStruct struct {
	Name   string
	Fields []StructField
}
