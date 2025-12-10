package generator

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Spec represents the generated command specification
type Spec struct {
	Version  int                    `toml:"version"`
	Commands map[string]CommandSpec `toml:"commands"`
}

// CommandSpec represents a single command in the spec
type CommandSpec struct {
	Service  string   `toml:"service"`
	Method   string   `toml:"method"`
	Use      string   `toml:"use"`
	Flags    []string `toml:"flags"`
	Returns  string   `toml:"returns"`
	Renderer string   `toml:"renderer"`
}

// GenerateSpec generates a TOML spec file for the new methods
func GenerateSpec(newMethods map[string][]Method, outputPath string) (err error) {
	spec := Spec{
		Version:  1,
		Commands: make(map[string]CommandSpec),
	}

	for serviceName, methods := range newMethods {
		prefix := strings.TrimSuffix(serviceName, "Service")

		for _, method := range methods {
			use := strings.TrimPrefix(method.Name, prefix)
			use = toKebabCase(use)

			commandKey := toKebabCase(prefix) + "-" + use

			var flags []string
			for _, param := range method.Params {
				flags = append(flags, toKebabCase(param.Name))
			}

			returnType := ""
			renderer := "success"
			for _, ret := range method.Returns {
				if !ret.IsError {
					returnType = ret.Type
					if strings.HasPrefix(returnType, "[]") {
						renderer = "table"
					} else if strings.HasPrefix(returnType, "*") {
						renderer = "detail"
					}
					break
				}
			}

			spec.Commands[commandKey] = CommandSpec{
				Service:  serviceName,
				Method:   method.Name,
				Use:      use,
				Flags:    flags,
				Returns:  returnType,
				Renderer: renderer,
			}
		}
	}

	specPath := filepath.Join(outputPath, "spec.toml")
	f, err := os.Create(specPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(spec)
}
