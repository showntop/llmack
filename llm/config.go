package llm

import (
	"os"
)

// ConfigSupplier ...
type ConfigSupplier interface {
	Get(key string) any
}

// ConfigSupplierFunc ...
type ConfigSupplierFunc func(key string) any

// WithConfigSupplier ...
func WithConfigSupplier(c ConfigSupplier) {
	Config = c
}

// Config default cache
var Config ConfigSupplier = &envConfig{}

type envConfig struct {
	Value map[string]any
}

// Get ...
func (c *envConfig) Get(key string) any {
	return os.Getenv(key)
}

// WithConfigs ...
func WithConfigs(c map[string]any) {
	Config = &mapConfig{Value: c}
}

type mapConfig struct {
	Value map[string]any
}

// Get ...
func (c *mapConfig) Get(key string) any {
	return c.Value[key]
}

// WithSingleConfig ...
func WithSingleConfig(c any) {
	Config = &anyConfig{Value: c}
}

// anyConfig for any type config
type anyConfig struct {
	Value any
}

func (c *anyConfig) Get(key string) any {
	return c.Value
}
