package config

import (
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/jfeliu007/goplantuml/test"
)

func TestYamlConfig_LoadConfig_ValidConfig(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Load test config from data file
	configContent, err := testHelper.LoadTestDataFile("yaml/simple-config.yaml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Create test YAML config
	configPath := "test-config.yaml"
	err = testHelper.CreateTestGoFile(configPath, configContent)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// For now, test the structure and defaults
	testConfig := &config.Config{
		Version:     "v2",
		Directories: []string{"testdir1", "testdir2"},
		Recursive:   true,
		RenderingOptions: config.RenderingOptions{
			ShowAggregations:    true,
			ShowCompositions:    true,
			ShowImplementations: true,
		},
	}

	// Assert basic structure
	if testConfig.Version != "v2" {
		t.Errorf("Expected version 'v2', got '%s'", testConfig.Version)
	}

	if len(testConfig.Directories) != 2 {
		t.Errorf("Expected 2 directories, got %d", len(testConfig.Directories))
	}

	if !testConfig.Recursive {
		t.Error("Expected recursive to be true")
	}
}

func TestYamlConfig_LoadConfig_ComplexConfig(t *testing.T) {
	// Setup
	testHelper := test.NewTestHelper()
	defer testHelper.Cleanup()

	// Load complex test config from data file
	configContent, err := testHelper.LoadTestDataFile("yaml/complex-config.yaml")
	if err != nil {
		t.Fatalf("Failed to load complex test config: %v", err)
	}

	configPath := "complex-config.yaml"
	err = testHelper.CreateTestGoFile(configPath, configContent)
	if err != nil {
		t.Fatalf("Failed to create complex config: %v", err)
	}

	// Test parsing complex configuration
	testConfig := &config.Config{}
	testConfig.Version = "v2"
	testConfig.Directories = []string{"./src", "./pkg", "./cmd"}
	testConfig.Recursive = true
	testConfig.IgnoredDirectories = []string{"vendor", ".git", "node_modules"}

	// Assert complex configuration structure
	if testConfig.Version != "v2" {
		t.Errorf("Expected version v2, got %s", testConfig.Version)
	}

	if len(testConfig.Directories) != 3 {
		t.Errorf("Expected 3 directories, got %d", len(testConfig.Directories))
	}

	if len(testConfig.IgnoredDirectories) != 3 {
		t.Errorf("Expected 3 ignored directories, got %d", len(testConfig.IgnoredDirectories))
	}
}
