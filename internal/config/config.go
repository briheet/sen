// Package config loads and validates sen project configuration.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	configType = "toml"
	// DefaultPath is the configuration used when no path is provided.
	DefaultPath = "sen.toml"
)

var (
	errResolveConfigPath  = errors.New("resolve config path")
	errReadConfig         = errors.New("read config")
	errDecodeConfig       = errors.New("decode config")
	errValidateConfig     = errors.New("validate config")
	errConfigPathConflict = errors.New("config path cannot be provided as both an argument and a flag")
)

// ServiceType identifies how a service is presented and managed.
type ServiceType string

const (
	// ServiceTypeServer identifies an application process.
	ServiceTypeServer ServiceType = "server"
	// ServiceTypeKV identifies a key-value datastore.
	ServiceTypeKV ServiceType = "kv"
)

// ServiceLang identifies a server runtime adapter.
type ServiceLang string

const (
	// ServiceLangGo identifies the Go server adapter.
	ServiceLangGo ServiceLang = "go"
	// ServiceLangNode identifies the Node.js server adapter.
	ServiceLangNode ServiceLang = "node"
)

// ServiceProvider identifies an external service implementation.
type ServiceProvider string

const (
	// ServiceProviderRedis identifies a Redis key-value service.
	ServiceProviderRedis ServiceProvider = "redis"
)

// Config describes a sen project and its services.
type Config struct {
	Project  Project   `mapstructure:"project"`
	Services []Service `mapstructure:"services" validate:"min=1,dive"`
}

// Project contains project-wide configuration.
type Project struct {
	Name string `mapstructure:"name" validate:"required"`
}

// Service describes one process or external service.
type Service struct {
	Name      string          `mapstructure:"name" validate:"required"`
	Type      ServiceType     `mapstructure:"type" validate:"required,oneof=server kv"`
	Lang      ServiceLang     `mapstructure:"lang" validate:"omitempty,oneof=go node"`
	Provider  ServiceProvider `mapstructure:"provider" validate:"omitempty,oneof=redis"`
	Path      string          `mapstructure:"path"`
	Address   string          `mapstructure:"address"`
	BuildArgs []string        `mapstructure:"build_args"`
	RunArgs   []string        `mapstructure:"run_args"`
}

var schema = validator.New(validator.WithRequiredStructEnabled())

// ResolvePath selects the CLI path and expands project directories.
func ResolvePath(path string, flagSet bool, args []string) (string, error) {
	if len(args) == 1 {
		if flagSet {
			return "", errConfigPathConflict
		}
		path = args[0]
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, DefaultPath)
	}
	return path, nil
}

// Load reads and validates a TOML configuration file.
func Load(path string) (*Config, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Join(errResolveConfigPath, err)
	}

	reader := viper.New()
	reader.SetConfigFile(path)
	reader.SetConfigType(configType)
	if err := reader.ReadInConfig(); err != nil {
		return nil, errors.Join(errReadConfig, err)
	}
	var result Config
	if err := reader.UnmarshalExact(&result); err != nil {
		return nil, errors.Join(errDecodeConfig, err)
	}
	if err := result.validate(filepath.Dir(path)); err != nil {
		return nil, errors.Join(errValidateConfig, err)
	}
	return &result, nil
}

func (c *Config) validate(baseDir string) error {
	if err := schema.Struct(c); err != nil {
		return err
	}
	if c.Project.Name == "." || c.Project.Name == ".." || filepath.Base(c.Project.Name) != c.Project.Name {
		return errors.New("project name must not be a path")
	}

	names := make(map[string]struct{}, len(c.Services))
	for index := range c.Services {
		service := &c.Services[index]
		if _, exists := names[service.Name]; exists {
			return errors.New("service " + strconv.Quote(service.Name) + " has a duplicate name")
		}
		names[service.Name] = struct{}{}

		switch service.Type {
		case ServiceTypeServer:
			if service.Lang == "" {
				return errors.New("service " + strconv.Quote(service.Name) + " requires lang")
			}
			if service.Provider != "" {
				return errors.New("service " + strconv.Quote(service.Name) + " cannot define provider")
			}
			if strings.TrimSpace(service.Path) == "" {
				return errors.New("service " + strconv.Quote(service.Name) + " requires path")
			}
			if service.Address != "" {
				return errors.New("service " + strconv.Quote(service.Name) + " cannot define address")
			}
			if filepath.IsAbs(service.Path) {
				service.Path = filepath.Clean(service.Path)
			} else {
				service.Path = filepath.Clean(filepath.Join(baseDir, service.Path))
			}
		case ServiceTypeKV:
			if service.Provider == "" {
				return errors.New("service " + strconv.Quote(service.Name) + " requires provider")
			}
			if service.Lang != "" {
				return errors.New("service " + strconv.Quote(service.Name) + " cannot define lang")
			}
			if strings.TrimSpace(service.Address) == "" {
				return errors.New("service " + strconv.Quote(service.Name) + " requires address")
			}
			if service.Path != "" {
				return errors.New("service " + strconv.Quote(service.Name) + " cannot define path")
			}
			if len(service.BuildArgs) != 0 || len(service.RunArgs) != 0 {
				return errors.New("service " + strconv.Quote(service.Name) + " cannot define build_args or run_args")
			}
		default:
			return errors.New("service " + strconv.Quote(service.Name) + " has unsupported type " + strconv.Quote(string(service.Type)))
		}
	}
	return nil
}
