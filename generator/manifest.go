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

// ManifestMethod represents a method in the manifest
type ManifestMethod struct {
	Name      string   `toml:"name"`
	Params    []string `toml:"params"`
	Returns   string   `toml:"returns"`
	Generated bool     `toml:"generated"`
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
func (m *Manifest) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

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
				params := make([]string, len(method.Params))
				for i, p := range method.Params {
					params[i] = p.Type
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
