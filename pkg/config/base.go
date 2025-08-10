package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BaseConfig contains basic configuration for goplantuml
type BaseConfig struct {
	// Output settings
	OutputFormat string
	OutputPath   string

	// Analysis settings
	Recursive          bool
	IgnoredDirectories []string
	Directories        []string

	// Rendering options
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
	ShowOptionsAsNote       bool

	// Diagram metadata
	Title string
	Notes string

	// Custom resource patterns
	CustomResources []string

	// Custom keyword patterns for function categorization
	CustomKeywords map[string][]string
}

// NewBaseConfig creates a new BaseConfig with default values
func NewBaseConfig() *BaseConfig {
	return &BaseConfig{
		OutputFormat:            "puml",
		OutputPath:              "",
		Recursive:               false,
		IgnoredDirectories:      []string{},
		Directories:             []string{},
		ShowAggregations:        false,
		HideFields:              false,
		HideMethods:             false,
		HideConnections:         false,
		ShowCompositions:        false,
		ShowImplementations:     false,
		ShowAliases:             false,
		ShowConnectionLabels:    false,
		AggregatePrivateMembers: false,
		HidePrivateMembers:      false,
		ShowOptionsAsNote:       false,
		Title:                   "",
		Notes:                   "",
	}
}

// LoadFromEnv loads configuration from environment variables
func (c *BaseConfig) LoadFromEnv() {
	if v := os.Getenv("GOPLANTUML_OUTPUT_FORMAT"); v != "" {
		c.OutputFormat = v
	}
	if v := os.Getenv("GOPLANTUML_OUTPUT_PATH"); v != "" {
		c.OutputPath = v
	}
	if v := os.Getenv("GOPLANTUML_RECURSIVE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.Recursive = b
		}
	}
	if v := os.Getenv("GOPLANTUML_IGNORE"); v != "" {
		c.IgnoredDirectories = strings.Split(v, ",")
	}
	if v := os.Getenv("GOPLANTUML_TITLE"); v != "" {
		c.Title = v
	}
	if v := os.Getenv("GOPLANTUML_NOTES"); v != "" {
		c.Notes = v
	}
}

// ValidateDirectories validates that all specified directories exist
func (c *BaseConfig) ValidateDirectories() error {
	for _, dir := range c.Directories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", dir)
		}
		if fi, err := os.Stat(dir); err == nil && !fi.IsDir() {
			return fmt.Errorf("path is not a directory: %s", dir)
		}
	}
	return nil
}

// GetAbsolutePaths converts relative paths to absolute paths
func (c *BaseConfig) GetAbsolutePaths() error {
	// Convert directories to absolute paths
	for i, dir := range c.Directories {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for directory %s: %w", dir, err)
		}
		c.Directories[i] = abs
	}

	// Convert ignored directories to absolute paths
	for i, dir := range c.IgnoredDirectories {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for ignored directory %s: %w", dir, err)
		}
		c.IgnoredDirectories[i] = abs
	}

	return nil
}

// Validate performs overall validation of the configuration
func (c *BaseConfig) Validate() error {
	if len(c.Directories) == 0 {
		return fmt.Errorf("at least one directory must be specified")
	}

	if err := c.ValidateDirectories(); err != nil {
		return err
	}

	if err := c.GetAbsolutePaths(); err != nil {
		return err
	}

	return nil
}
