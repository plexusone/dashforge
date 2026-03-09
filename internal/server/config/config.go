// Package config provides configuration loading for Dashforge Server.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the complete server configuration.
type Config struct {
	// Server configuration
	Server ServerConfig `yaml:"server"`

	// Database connections
	Databases []DatabaseConfig `yaml:"databases"`

	// Authentication configuration
	Auth AuthConfig `yaml:"auth"`

	// Storage for dashboards
	Storage StorageConfig `yaml:"storage"`

	// Authorization configuration
	AuthZ AuthZConfig `yaml:"authz"`
}

// AuthZConfig holds authorization settings.
type AuthZConfig struct {
	// Mode is "simple" or "spicedb"
	Mode string `yaml:"mode"`

	// SpiceDB endpoint (e.g., "localhost:50051")
	SpiceDBEndpoint string `yaml:"spicedb_endpoint"`
	SpiceDBEndpointEnv string `yaml:"spicedb_endpoint_env"`

	// SpiceDB preshared key
	SpiceDBToken string `yaml:"spicedb_token"`
	SpiceDBTokenEnv string `yaml:"spicedb_token_env"`

	// Insecure connection (for development)
	SpiceDBInsecure bool `yaml:"spicedb_insecure"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port            int    `yaml:"port"`
	Host            string `yaml:"host"`
	ReadTimeout     string `yaml:"readTimeout"`
	WriteTimeout    string `yaml:"writeTimeout"`
	ShutdownTimeout string `yaml:"shutdownTimeout"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // postgres, mysql, sqlite
	URL      string `yaml:"url"`
	URLEnv   string `yaml:"urlEnv"` // Environment variable name for URL
	MaxConns int    `yaml:"maxConns"`
	ReadOnly bool   `yaml:"readOnly"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Provider string        `yaml:"provider"` // local, github, google, oidc
	Local    LocalAuth     `yaml:"local,omitempty"`
	OAuth    OAuthConfig   `yaml:"oauth,omitempty"`
	JWT      JWTConfig     `yaml:"jwt"`
}

// LocalAuth holds local authentication settings.
type LocalAuth struct {
	// AllowSignup allows new user registration
	AllowSignup bool `yaml:"allowSignup"`
}

// OAuthConfig holds OAuth provider settings.
type OAuthConfig struct {
	ClientID     string `yaml:"clientId"`
	ClientIDEnv  string `yaml:"clientIdEnv"`
	ClientSecret string `yaml:"clientSecret"`
	ClientSecretEnv string `yaml:"clientSecretEnv"`
	AuthURL      string `yaml:"authUrl"`
	TokenURL     string `yaml:"tokenUrl"`
	Scopes       []string `yaml:"scopes"`
}

// JWTConfig holds JWT settings.
type JWTConfig struct {
	Secret       string `yaml:"secret"`
	SecretEnv    string `yaml:"secretEnv"`
	ExpiresIn    string `yaml:"expiresIn"`
	RefreshIn    string `yaml:"refreshIn"`
}

// StorageConfig holds dashboard storage settings.
type StorageConfig struct {
	Type string `yaml:"type"` // file, database
	Path string `yaml:"path"` // for file storage
}

// Load reads configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Expand environment variables
	data = []byte(os.ExpandEnv(string(data)))

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Apply defaults
	cfg.applyDefaults()

	// Resolve environment variable references
	cfg.resolveEnvVars()

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.ReadTimeout == "" {
		c.Server.ReadTimeout = "30s"
	}
	if c.Server.WriteTimeout == "" {
		c.Server.WriteTimeout = "60s"
	}
	if c.Server.ShutdownTimeout == "" {
		c.Server.ShutdownTimeout = "10s"
	}
	if c.Storage.Type == "" {
		c.Storage.Type = "file"
	}
	if c.Storage.Path == "" {
		c.Storage.Path = "./dashboards"
	}
}

func (c *Config) resolveEnvVars() {
	// Resolve database URLs from environment
	for i := range c.Databases {
		if c.Databases[i].URLEnv != "" && c.Databases[i].URL == "" {
			c.Databases[i].URL = os.Getenv(c.Databases[i].URLEnv)
		}
	}

	// Resolve OAuth credentials from environment
	if c.Auth.OAuth.ClientIDEnv != "" && c.Auth.OAuth.ClientID == "" {
		c.Auth.OAuth.ClientID = os.Getenv(c.Auth.OAuth.ClientIDEnv)
	}
	if c.Auth.OAuth.ClientSecretEnv != "" && c.Auth.OAuth.ClientSecret == "" {
		c.Auth.OAuth.ClientSecret = os.Getenv(c.Auth.OAuth.ClientSecretEnv)
	}

	// Resolve JWT secret from environment
	if c.Auth.JWT.SecretEnv != "" && c.Auth.JWT.Secret == "" {
		c.Auth.JWT.Secret = os.Getenv(c.Auth.JWT.SecretEnv)
	}

	// Resolve SpiceDB settings from environment
	if c.AuthZ.SpiceDBEndpointEnv != "" && c.AuthZ.SpiceDBEndpoint == "" {
		c.AuthZ.SpiceDBEndpoint = os.Getenv(c.AuthZ.SpiceDBEndpointEnv)
	}
	if c.AuthZ.SpiceDBTokenEnv != "" && c.AuthZ.SpiceDBToken == "" {
		c.AuthZ.SpiceDBToken = os.Getenv(c.AuthZ.SpiceDBTokenEnv)
	}
}

// IsSpiceDB returns true if SpiceDB authorization is configured.
func (c *Config) IsSpiceDB() bool {
	return c.AuthZ.Mode == "spicedb" && c.AuthZ.SpiceDBEndpoint != ""
}

// Default returns a default configuration.
func Default() *Config {
	cfg := &Config{}
	cfg.applyDefaults()
	return cfg
}
