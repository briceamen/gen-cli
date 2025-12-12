package generator

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Manifest tracks known SDK methods and their generation status
type Manifest struct {
	Version    int                        `toml:"version"`
	SDKVersion string                     `toml:"sdk_version"`
	Services   map[string]ManifestService `toml:"services"`
}

// ManifestService represents a service in the manifest
type ManifestService struct {
	Methods []ManifestMethod `toml:"methods"`
}

// ManifestParam represents a method parameter stored in the manifest
type ManifestParam struct {
	Name string `toml:"name,omitempty"`
	Type string `toml:"type"`
}

// UnmarshalTOML supports both the legacy string format (just the type) and the
// new table format (name + type).
func (p *ManifestParam) UnmarshalTOML(data interface{}) error {
	switch v := data.(type) {
	case string:
		p.Type = v
		return nil
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok {
			p.Name = name
		}
		if typ, ok := v["type"].(string); ok {
			p.Type = typ
		}
		return nil
	default:
		return nil
	}
}

// ManifestMethod represents a method in the manifest
type ManifestMethod struct {
	Name      string          `toml:"name"`
	Params    []ManifestParam `toml:"params"`
	Returns   string          `toml:"returns"`
	Generated bool            `toml:"generated"`
}

// NewManifest creates a new empty manifest
func NewManifest() *Manifest {
	return &Manifest{
		Version:  1,
		Services: make(map[string]ManifestService),
	}
}

// LoadManifest loads a manifest from a TOML file
func LoadManifest(path string) (*Manifest, error) {
	var manifest Manifest
	if _, err := toml.DecodeFile(path, &manifest); err != nil {
		return nil, err
	}
	if manifest.Services == nil {
		manifest.Services = make(map[string]ManifestService)
	}
	return &manifest, nil
}

// Save writes the manifest to a TOML file
func (m *Manifest) Save(path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(m)
}

// HasMethod checks if a method is already in the manifest
func (m *Manifest) HasMethod(serviceName, methodName string) bool {
	service, ok := m.Services[serviceName]
	if !ok {
		return false
	}
	for _, method := range service.Methods {
		if method.Name == methodName {
			return true
		}
	}
	return false
}

// AddServices adds parsed services to the manifest
func (m *Manifest) AddServices(services []Service) {
	for _, svc := range services {
		ms := m.Services[svc.Name]
		for _, method := range svc.Methods {
			if !m.HasMethod(svc.Name, method.Name) {
				params := make([]ManifestParam, len(method.Params))
				for i, p := range method.Params {
					params[i] = ManifestParam{Name: p.Name, Type: p.Type}
				}
				returns := ""
				for _, r := range method.Returns {
					if !r.IsError {
						returns = r.Type
						break
					}
				}
				ms.Methods = append(ms.Methods, ManifestMethod{
					Name:      method.Name,
					Params:    params,
					Returns:   returns,
					Generated: true,
				})
			}
		}
		m.Services[svc.Name] = ms
	}
}

// MethodsToGenerate converts the manifest into generator Methods, keeping only
// entries marked as Generated.
func (m *Manifest) MethodsToGenerate() map[string][]Method {
	result := make(map[string][]Method)
	for serviceName, svc := range m.Services {
		for _, method := range svc.Methods {
			if !method.Generated {
				continue
			}
			result[serviceName] = append(result[serviceName], method.toMethod())
		}
	}
	return result
}

// MethodsToGenerateSet returns a set of "ServiceName.MethodName" keys for methods
// marked as Generated. This is used to filter parsed SDK methods.
func (m *Manifest) MethodsToGenerateSet() map[string]bool {
	result := make(map[string]bool)
	for serviceName, svc := range m.Services {
		for _, method := range svc.Methods {
			if method.Generated {
				result[serviceName+"."+method.Name] = true
			}
		}
	}
	return result
}

// EnsureParamNames fills missing parameter names using the same inference
// logic as the parser. This avoids writing entries with empty names when
// persisting a legacy manifest.
func (m *Manifest) EnsureParamNames() {
	for svcName, svc := range m.Services {
		updated := false
		for mi := range svc.Methods {
			for pi := range svc.Methods[mi].Params {
				if svc.Methods[mi].Params[pi].Name == "" {
					svc.Methods[mi].Params[pi].Name = inferParamName(svc.Methods[mi].Params[pi].Type, pi)
					updated = true
				}
			}
		}
		if updated {
			m.Services[svcName] = svc
		}
	}
}

func (m ManifestMethod) toMethod() Method {
	var params []Param
	for i, p := range m.Params {
		name := p.Name
		if name == "" {
			name = inferParamName(p.Type, i)
		}
		params = append(params, Param{
			Name: name,
			Type: p.Type,
		})
	}

	var returns []Return
	if m.Returns != "" {
		returns = append(returns, Return{Type: m.Returns})
	}
	// Assume methods always return error as the last return value; this keeps
	// downstream renderer inference logic intact.
	returns = append(returns, Return{Type: "error", IsError: true})

	return Method{
		Name:    m.Name,
		Params:  params,
		Returns: returns,
	}
}
