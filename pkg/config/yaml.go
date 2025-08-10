package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the configuration for GoPlantUML
type Config struct {
	Version            string           `yaml:"version"`
	Directories        []string         `yaml:"directories"`
	IgnoredDirectories []string         `yaml:"ignored_directories"`
	Recursive          bool             `yaml:"recursive"`
	Output             OutputConfig     `yaml:"output"`
	RenderingOptions   RenderingOptions `yaml:"rendering_options"`
	CustomKeywords     CustomKeywords   `yaml:"custom_keywords"`
}

// CustomKeywords represents custom keyword patterns for function categorization
// Each key is a resource type, and the value is a list of keywords for that type
type CustomKeywords map[string][]string

// OutputConfig represents output configuration
type OutputConfig struct {
	File   string `yaml:"file"`
	Format string `yaml:"format"`
}

// RenderingOptions represents rendering options for diagram generation
type RenderingOptions struct {
	ShowAggregations        bool   `yaml:"show_aggregations"`
	HideFields              bool   `yaml:"hide_fields"`
	HideMethods             bool   `yaml:"hide_methods"`
	HideConnections         bool   `yaml:"hide_connections"`
	ShowCompositions        bool   `yaml:"show_compositions"`
	ShowImplementations     bool   `yaml:"show_implementations"`
	ShowAliases             bool   `yaml:"show_aliases"`
	ShowConnectionLabels    bool   `yaml:"show_connection_labels"`
	AggregatePrivateMembers bool   `yaml:"aggregate_private_members"`
	HidePrivateMembers      bool   `yaml:"hide_private_members"`
	Title                   string `yaml:"title"`
	Notes                   string `yaml:"notes"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(configPath string) (*Config, error) {
	// If no path specified, try default locations
	if configPath == "" {
		configPath = findDefaultConfig()
		if configPath == "" {
			return nil, fmt.Errorf("no configuration file found in default locations")
		}
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %v", configPath, err)
	}

	setDefaults(&config)
	return &config, nil
}

// findDefaultConfig looks for default configuration files
func findDefaultConfig() string {
	candidates := []string{
		"etc/goplantuml/goplantuml.yaml", // Default configuration location
		"goplantuml.yaml",
		"goplantuml.yml",
		".goplantuml.yaml",
		".goplantuml.yml",
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := ioutil.ReadFile(filename)
	return err == nil
}

// setDefaults sets default values for the configuration
func setDefaults(config *Config) {
	if config.Version == "" {
		config.Version = "v2"
	}
	if config.Output.Format == "" {
		config.Output.Format = "puml"
	}
}

// toV1Options converts to v1 repository options map
func (c *Config) toV1Options() map[string]interface{} {
	return map[string]interface{}{
		"show_aggregations":         c.RenderingOptions.ShowAggregations,
		"hide_fields":               c.RenderingOptions.HideFields,
		"hide_methods":              c.RenderingOptions.HideMethods,
		"hide_connections":          c.RenderingOptions.HideConnections,
		"show_compositions":         c.RenderingOptions.ShowCompositions,
		"show_implementations":      c.RenderingOptions.ShowImplementations,
		"show_aliases":              c.RenderingOptions.ShowAliases,
		"show_connection_labels":    c.RenderingOptions.ShowConnectionLabels,
		"aggregate_private_members": c.RenderingOptions.AggregatePrivateMembers,
		"hide_private_members":      c.RenderingOptions.HidePrivateMembers,
		"title":                     c.RenderingOptions.Title,
		"notes":                     c.RenderingOptions.Notes,
	}
}

// toV2Options converts to v2 repository options
func (c *Config) toV2Options() DiagramOptions {
	return DiagramOptions{
		ShowAggregations:        c.RenderingOptions.ShowAggregations,
		HideFields:              c.RenderingOptions.HideFields,
		HideMethods:             c.RenderingOptions.HideMethods,
		HideConnections:         c.RenderingOptions.HideConnections,
		ShowCompositions:        c.RenderingOptions.ShowCompositions,
		ShowImplementations:     c.RenderingOptions.ShowImplementations,
		ShowAliases:             c.RenderingOptions.ShowAliases,
		ShowConnectionLabels:    c.RenderingOptions.ShowConnectionLabels,
		AggregatePrivateMembers: c.RenderingOptions.AggregatePrivateMembers,
		HidePrivateMembers:      c.RenderingOptions.HidePrivateMembers,
		Title:                   c.RenderingOptions.Title,
		Notes:                   c.RenderingOptions.Notes,
		CustomKeywords:          c.CustomKeywords,
	}
}

// DiagramOptions represents options for diagram generation
type DiagramOptions struct {
	ShowAggregations        bool
	HideFields              bool
	HideMethods             bool
	HideConnections         bool
	ShowCompositions        bool
	ShowImplementations     bool
	ShowAliases             bool
	ShowConnectionLabels    bool
	AggregatePrivateMembers bool
	HidePrivateMembers      bool
	Title                   string
	Notes                   string
	CustomKeywords          map[string][]string
}

// IsV1 returns true if the configuration is for v1
func (c *Config) IsV1() bool {
	return c.Version == "v1"
}

// IsV2 returns true if the configuration is for v2
func (c *Config) IsV2() bool {
	return c.Version == "v2" || c.Version == ""
}
