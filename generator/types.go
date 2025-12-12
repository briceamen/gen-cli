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
	// Hidden indicates this method should not generate a CLI command
	// (typically because it's a helper method used by chaining)
	Hidden bool
}

// Param represents a method parameter
type Param struct {
	Name string
	Type string
	// ChainedFrom indicates this param should be fetched by calling another method first
	// e.g., logsURL param is chained from LogsURL method
	ChainedFrom *ChainedParam
}

// ChainedParam describes how to fetch a parameter value from another method
type ChainedParam struct {
	MethodName   string  // Method to call to get this value (e.g., "LogsURL")
	SourceParams []Param // Parameters needed by the chained method (e.g., "app")
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

// StructBuilder represents code to build an SDK struct from CLI flags
type StructBuilder struct {
	VarName   string               // Go variable name (e.g., "opts")
	TypeName  string               // SDK type name (e.g., "AppsCreateOpts")
	IsPointer bool                 // Whether the param expects a pointer
	Fields    []StructFieldMapping // Fields to populate from CLI flags
}

// StructFieldMapping maps a CLI flag to a struct field
type StructFieldMapping struct {
	FieldName  string // Struct field name (e.g., "Name")
	FlagVar    string // CLI flag variable name (e.g., "name")
	FieldType  string // Go type for conversion (e.g., "string")
	FlagName   string // CLI flag name in kebab-case (e.g., "name")
	IsPointer  bool   // Whether the field type is a pointer
	NeedsDeref bool   // Whether we need to take address of flag value
	Skip       bool   // Whether to skip this field (complex types)
}
