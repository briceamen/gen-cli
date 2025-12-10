package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// ParseSDK parses the go-scalingo SDK and extracts all service interfaces
func ParseSDK(sdkPath string) ([]Service, error) {
	fset := token.NewFileSet()

	// Parse all Go files in the SDK root directory
	pkgs, err := parser.ParseDir(fset, sdkPath, func(fi os.FileInfo) bool {
		// Skip test files and vendor
		return !strings.HasSuffix(fi.Name(), "_test.go") &&
			!strings.HasPrefix(fi.Name(), ".")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	var services []Service
	structs := make(map[string]ParsedStruct)

	// First pass: collect all structs (needed for expanding opts)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								ps := parseStruct(typeSpec.Name.Name, structType)
								structs[ps.Name] = ps
							}
						}
					}
				}
			}
		}
	}

	// Second pass: extract service interfaces
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							// Look for *Service interfaces
							if strings.HasSuffix(typeSpec.Name.Name, "Service") {
								// Skip preview services - they use different clients
								if strings.Contains(typeSpec.Name.Name, "Preview") {
									continue
								}
								if ifaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
									service := parseInterface(typeSpec.Name.Name, ifaceType)
									services = append(services, service)
								}
							}
						}
					}
				}
			}
		}
	}

	return services, nil
}

// ParseSDKWithStructs parses the SDK and returns both services and struct definitions
// This is more efficient than calling ParseSDK and GetStructs separately
func ParseSDKWithStructs(sdkPath string) ([]Service, map[string]ParsedStruct, error) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, sdkPath, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go") &&
			!strings.HasPrefix(fi.Name(), ".")
	}, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	var services []Service
	structs := make(map[string]ParsedStruct)

	// First pass: collect all structs
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								ps := parseStruct(typeSpec.Name.Name, structType)
								structs[ps.Name] = ps
							}
						}
					}
				}
			}
		}
	}

	// Second pass: extract service interfaces
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if strings.HasSuffix(typeSpec.Name.Name, "Service") {
								// Skip preview services - they use different clients
								if strings.Contains(typeSpec.Name.Name, "Preview") {
									continue
								}
								if ifaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
									service := parseInterface(typeSpec.Name.Name, ifaceType)
									services = append(services, service)
								}
							}
						}
					}
				}
			}
		}
	}

	return services, structs, nil
}

// IsExpandableParam checks if a parameter type should be expanded into individual CLI flags
// This includes types ending in Opts, Options, or Params
func IsExpandableParam(paramType string, structs map[string]ParsedStruct) bool {
	baseType := strings.TrimPrefix(paramType, "*")
	if _, ok := structs[baseType]; ok {
		return strings.HasSuffix(baseType, "Opts") ||
			strings.HasSuffix(baseType, "Options") ||
			strings.HasSuffix(baseType, "Params")
	}
	return false
}

// IsPaginationParam checks if a parameter is a pagination options type
func IsPaginationParam(paramType string) bool {
	baseType := strings.TrimPrefix(paramType, "*")
	return baseType == "PaginationOpts"
}

// HasPaginationMeta checks if a method returns PaginationMeta (indicating it supports pagination)
func HasPaginationMeta(method Method) bool {
	for _, ret := range method.Returns {
		if ret.Type == "PaginationMeta" {
			return true
		}
	}
	return false
}

// SupportsPagination checks if a method supports pagination by examining its parameters.
// A method supports pagination if it accepts PaginationOpts.
func SupportsPagination(method Method) bool {
	for _, param := range method.Params {
		if IsPaginationParam(param.Type) {
			return true
		}
	}
	return false
}

// GetPrimaryReturnType returns the first non-error return type
func GetPrimaryReturnType(method Method) string {
	for _, ret := range method.Returns {
		if !ret.IsError && ret.Type != "PaginationMeta" {
			return ret.Type
		}
	}
	return ""
}

// InferRendererType determines the appropriate renderer based on return type
func InferRendererType(returnType string) string {
	switch {
	case returnType == "" || returnType == "error":
		return "success"
	case strings.HasPrefix(returnType, "[]") || strings.HasPrefix(returnType, "[]*"):
		return "table"
	case returnType == "*http.Response":
		return "http"
	default:
		return "detail"
	}
}

func parseInterface(name string, iface *ast.InterfaceType) Service {
	service := Service{
		Name: name,
	}

	if iface.Methods == nil {
		return service
	}

	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue // embedded interface
		}

		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		m := Method{
			Name: method.Names[0].Name,
		}

		// Parse parameters
		if funcType.Params != nil {
			for _, param := range funcType.Params.List {
				paramType := typeToString(param.Type)

				// Check if first param is context
				if paramType == "context.Context" {
					m.HasContext = true
					continue
				}

				// Handle multi-name declarations like "app, addonID string"
				if len(param.Names) > 0 {
					for _, name := range param.Names {
						m.Params = append(m.Params, Param{
							Name: name.Name,
							Type: paramType,
						})
					}
				} else {
					// Generate name from type (e.g., "PaginationOpts" -> "opts")
					paramName := inferParamName(paramType, len(m.Params))
					m.Params = append(m.Params, Param{
						Name: paramName,
						Type: paramType,
					})
				}
			}
		}

		// Parse return types
		if funcType.Results != nil {
			for _, result := range funcType.Results.List {
				retType := typeToString(result.Type)
				m.Returns = append(m.Returns, Return{
					Type:    retType,
					IsError: retType == "error",
				})
			}
		}

		service.Methods = append(service.Methods, m)
	}

	return service
}

func parseStruct(name string, st *ast.StructType) ParsedStruct {
	ps := ParsedStruct{
		Name: name,
	}

	if st.Fields == nil {
		return ps
	}

	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue // embedded field
		}

		sf := StructField{
			Name: field.Names[0].Name,
			Type: typeToString(field.Type),
		}

		// Extract JSON tag
		if field.Tag != nil {
			tag := field.Tag.Value
			if idx := strings.Index(tag, `json:"`); idx != -1 {
				start := idx + 6
				end := strings.Index(tag[start:], `"`)
				if end != -1 {
					jsonTag := tag[start : start+end]
					parts := strings.Split(jsonTag, ",")
					sf.JSONTag = parts[0]
					for _, part := range parts[1:] {
						if part == "omitempty" {
							sf.Optional = true
						}
					}
				}
			}
		}

		ps.Fields = append(ps.Fields, sf)
	}

	return ps
}

// inferParamName generates a parameter name from its type
func inferParamName(paramType string, index int) string {
	// Strip pointer and slice prefixes
	name := strings.TrimPrefix(paramType, "*")
	name = strings.TrimPrefix(name, "[]")

	// Get the last part if there's a package path
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}

	// Convert to camelCase
	if len(name) > 0 {
		name = strings.ToLower(name[:1]) + name[1:]
	}

	// Handle common suffixes
	if strings.HasSuffix(name, "Opts") || strings.HasSuffix(name, "Options") {
		name = "opts"
	} else if strings.HasSuffix(name, "Params") {
		name = "params"
	}

	// If still empty or conflicts, add index
	if name == "" {
		name = fmt.Sprintf("arg%d", index)
	}

	return name
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

// GetStructs returns parsed structs for a given SDK path (used for expanding opts)
func GetStructs(sdkPath string) (map[string]ParsedStruct, error) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, sdkPath, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}

	structs := make(map[string]ParsedStruct)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								ps := parseStruct(typeSpec.Name.Name, structType)
								structs[ps.Name] = ps
							}
						}
					}
				}
			}
		}
	}

	return structs, nil
}
