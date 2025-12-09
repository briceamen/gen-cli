package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
)

const (
	authDataVersionV2  = "2.0"
	authDataVersionV21 = "2.1"
	defaultAuthHost    = "auth.scalingo.com"
)

// AuthConfig holds the authentication configuration loaded from file
type AuthConfig struct {
	AuthDataVersion   string          `json:"auth_data_version" mapstructure:"auth_data_version"`
	LastUpdate        time.Time       `json:"last_update" mapstructure:"last_update"`
	AuthConfigPerHost json.RawMessage `json:"auth_config_data" mapstructure:"auth_config_data"`
}

// CredentialsData holds the token and user info for a specific host
type CredentialsData struct {
	Tokens *UserToken `json:"tokens" mapstructure:"tokens"`
	User   *UserInfo  `json:"user" mapstructure:"user"`
}

// UserToken holds the API token
type UserToken struct {
	Token string `json:"token" mapstructure:"token"`
}

// UserInfo holds basic user info
type UserInfo struct {
	ID       string `json:"id" mapstructure:"id"`
	Username string `json:"username" mapstructure:"username"`
	Email    string `json:"email" mapstructure:"email"`
}

// ConfigPerHost maps hosts to their credentials
type ConfigPerHost map[string]CredentialsData

// Config holds the CLI configuration
type Config struct {
	v        *viper.Viper
	AuthFile string
	Region   string
}

// C is the global config instance
var C *Config

func init() {
	C = New()
}

// New creates a new Config instance
func New() *Config {
	c := &Config{
		v: viper.New(),
	}
	c.init()
	return c
}

func (c *Config) init() {
	home := homeDir()
	if home == "" {
		panic("The HOME environment variable must be defined")
	}

	configDir := filepath.Join(home, ".config", "scalingo")
	c.AuthFile = filepath.Join(configDir, "auth")

	// Check for env override
	if envAuth := os.Getenv("SCALINGO_AUTH_FILE"); envAuth != "" {
		c.AuthFile = envAuth
	}

	// Region from env
	c.Region = os.Getenv("SCALINGO_REGION")
	if c.Region == "" {
		// Try to read from config file
		configPath := filepath.Join(configDir, "config.json")
		c.v.SetConfigFile(configPath)
		c.v.SetConfigType("json")
		if err := c.v.ReadInConfig(); err == nil {
			c.Region = c.v.GetString("region")
		}
	}
}

// LoadAuth loads authentication from the auth file
func (c *Config) LoadAuth() (string, error) {
	// First check for env var override
	if token := os.Getenv("SCALINGO_API_TOKEN"); token != "" {
		return token, nil
	}

	// Read auth file using viper
	v := viper.New()
	v.SetConfigFile(c.AuthFile)
	v.SetConfigType("json")

	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not authenticated: run 'scalingo login' first or set SCALINGO_API_TOKEN env var")
		}
		return "", fmt.Errorf("failed to read auth file: %w", err)
	}

	// Get auth version
	version := v.GetString("auth_data_version")
	if version != authDataVersionV2 && version != authDataVersionV21 {
		return "", fmt.Errorf("unsupported auth version: %s", version)
	}

	// Get auth config per host - this is stored as json.RawMessage
	// We need to read the raw data and unmarshal it
	authConfigRaw := v.Get("auth_config_data")
	if authConfigRaw == nil {
		return "", fmt.Errorf("no auth_config_data in auth file")
	}

	// The auth_config_data is a map[string]interface{} when viper reads it
	configMap, ok := authConfigRaw.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid auth_config_data format")
	}

	// Try auth.scalingo.com first (v2.1), then api.scalingo.com (v2.0)
	hosts := []string{defaultAuthHost, "api.scalingo.com"}
	for _, host := range hosts {
		hostData, ok := configMap[host]
		if !ok {
			continue
		}

		hostMap, ok := hostData.(map[string]interface{})
		if !ok {
			continue
		}

		tokensData, ok := hostMap["tokens"]
		if !ok {
			continue
		}

		tokensMap, ok := tokensData.(map[string]interface{})
		if !ok {
			continue
		}

		token, ok := tokensMap["token"].(string)
		if ok && token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no valid token found in auth file")
}

// GetRegion returns the configured region
func (c *Config) GetRegion() string {
	return c.Region
}

func homeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
