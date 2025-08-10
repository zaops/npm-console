package config

import (
	"fmt"
	"os"
	"path/filepath"

	"npm-console/pkg/logger"
	"npm-console/pkg/utils"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// Application settings
	App AppConfig `yaml:"app" json:"app"`
	
	// Logger configuration
	Logger logger.Config `yaml:"logger" json:"logger"`
	
	// Web server configuration
	Web WebConfig `yaml:"web" json:"web"`
	
	// Package manager configurations
	Managers ManagersConfig `yaml:"managers" json:"managers"`
	
	// Cache settings
	Cache CacheConfig `yaml:"cache" json:"cache"`
}

// AppConfig represents application-level configuration
type AppConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"`
	DataDir     string `yaml:"data_dir" json:"data_dir"`
	ConfigDir   string `yaml:"config_dir" json:"config_dir"`
}

// WebConfig represents web server configuration
type WebConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Host    string `yaml:"host" json:"host"`
	Port    int    `yaml:"port" json:"port"`
	TLS     struct {
		Enabled  bool   `yaml:"enabled" json:"enabled"`
		CertFile string `yaml:"cert_file" json:"cert_file"`
		KeyFile  string `yaml:"key_file" json:"key_file"`
	} `yaml:"tls" json:"tls"`
	CORS struct {
		Enabled        bool     `yaml:"enabled" json:"enabled"`
		AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
		AllowedMethods []string `yaml:"allowed_methods" json:"allowed_methods"`
		AllowedHeaders []string `yaml:"allowed_headers" json:"allowed_headers"`
	} `yaml:"cors" json:"cors"`
}

// ManagersConfig represents package managers configuration
type ManagersConfig struct {
	NPM  ManagerConfig `yaml:"npm" json:"npm"`
	PNPM ManagerConfig `yaml:"pnpm" json:"pnpm"`
	Yarn ManagerConfig `yaml:"yarn" json:"yarn"`
	Bun  ManagerConfig `yaml:"bun" json:"bun"`
}

// ManagerConfig represents individual package manager configuration
type ManagerConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Registry string            `yaml:"registry" json:"registry"`
	Proxy    string            `yaml:"proxy" json:"proxy"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	AutoClean    bool   `yaml:"auto_clean" json:"auto_clean"`
	MaxSize      string `yaml:"max_size" json:"max_size"`
	MaxAge       string `yaml:"max_age" json:"max_age"`
	ScanInterval string `yaml:"scan_interval" json:"scan_interval"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home, _ := utils.GetHomeDir()
	configDir, _ := utils.GetConfigDir()
	
	return &Config{
		App: AppConfig{
			Name:        "npm-console",
			Version:     "1.0.0",
			Environment: "development",
			DataDir:     filepath.Join(home, ".npm-console"),
			ConfigDir:   filepath.Join(configDir, "npm-console"),
		},
		Logger: *logger.DefaultConfig(),
		Web: WebConfig{
			Enabled: true,
			Host:    "localhost",
			Port:    8080,
			TLS: struct {
				Enabled  bool   `yaml:"enabled" json:"enabled"`
				CertFile string `yaml:"cert_file" json:"cert_file"`
				KeyFile  string `yaml:"key_file" json:"key_file"`
			}{
				Enabled: false,
			},
			CORS: struct {
				Enabled        bool     `yaml:"enabled" json:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods" json:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers" json:"allowed_headers"`
			}{
				Enabled:        true,
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"*"},
			},
		},
		Managers: ManagersConfig{
			NPM: ManagerConfig{
				Enabled:  true,
				Registry: "https://registry.npmjs.org/",
				Settings: make(map[string]string),
			},
			PNPM: ManagerConfig{
				Enabled:  true,
				Registry: "https://registry.npmjs.org/",
				Settings: make(map[string]string),
			},
			Yarn: ManagerConfig{
				Enabled:  true,
				Registry: "https://registry.npmjs.org/",
				Settings: make(map[string]string),
			},
			Bun: ManagerConfig{
				Enabled:  true,
				Registry: "https://registry.npmjs.org/",
				Settings: make(map[string]string),
			},
		},
		Cache: CacheConfig{
			AutoClean:    false,
			MaxSize:      "10GB",
			MaxAge:       "30d",
			ScanInterval: "1h",
		},
	}
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()
	
	// Set up viper
	v := viper.New()
	v.SetConfigType("yaml")
	
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config in standard locations
		v.SetConfigName(".npm-console")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
		if configDir, err := utils.GetConfigDir(); err == nil {
			v.AddConfigPath(filepath.Join(configDir, "npm-console"))
		}
	}
	
	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("NPM_CONSOLE")
	
	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}
	
	// Unmarshal into config struct
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate and set defaults
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return config, nil
}

// Save saves the configuration to a file
func (c *Config) Save(configPath string) error {
	if configPath == "" {
		configDir, err := utils.GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}
		configPath = filepath.Join(configDir, "npm-console", "config.yaml")
	}
	
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := utils.MakeDir(dir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// validate validates the configuration
func (c *Config) validate() error {
	// Validate web port
	if c.Web.Port <= 0 || c.Web.Port > 65535 {
		return fmt.Errorf("invalid web port: %d", c.Web.Port)
	}
	
	// Ensure data and config directories exist
	if err := utils.MakeDir(c.App.DataDir); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	if err := utils.MakeDir(c.App.ConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	return nil
}

// GetManagerConfig returns configuration for a specific manager
func (c *Config) GetManagerConfig(manager string) *ManagerConfig {
	switch manager {
	case "npm":
		return &c.Managers.NPM
	case "pnpm":
		return &c.Managers.PNPM
	case "yarn":
		return &c.Managers.Yarn
	case "bun":
		return &c.Managers.Bun
	default:
		return nil
	}
}

// SetManagerRegistry sets the registry for a specific manager
func (c *Config) SetManagerRegistry(manager, registry string) error {
	config := c.GetManagerConfig(manager)
	if config == nil {
		return fmt.Errorf("unknown manager: %s", manager)
	}
	config.Registry = registry
	return nil
}

// SetManagerProxy sets the proxy for a specific manager
func (c *Config) SetManagerProxy(manager, proxy string) error {
	config := c.GetManagerConfig(manager)
	if config == nil {
		return fmt.Errorf("unknown manager: %s", manager)
	}
	config.Proxy = proxy
	return nil
}
