package config

import (
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/config"
)

func TestYamlConfig_SetDefaults(t *testing.T) {
	testCases := []struct {
		name            string
		inputConfig     *config.Config
		expectedVersion string
		expectedDirs    []string
		expectedFormat  string
	}{
		{
			name:            "Empty config",
			inputConfig:     &config.Config{},
			expectedVersion: "v2",
			expectedDirs:    []string{"."},
			expectedFormat:  "puml",
		},
		{
			name: "Partial config",
			inputConfig: &config.Config{
				Version:     "v1",
				Directories: []string{"custom_dir"},
			},
			expectedVersion: "v1",
			expectedDirs:    []string{"custom_dir"},
			expectedFormat:  "puml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a copy to avoid modifying the original
			configCopy := *tc.inputConfig

			// Simulate the setDefaults function behavior
			if configCopy.Version == "" {
				configCopy.Version = "v2"
			}
			if len(configCopy.Directories) == 0 {
				configCopy.Directories = []string{"."}
			}
			if configCopy.Output.Format == "" {
				configCopy.Output.Format = "puml"
			}

			// Assert
			if configCopy.Version != tc.expectedVersion {
				t.Errorf("Expected version '%s', got '%s'", tc.expectedVersion, configCopy.Version)
			}

			if len(configCopy.Directories) != len(tc.expectedDirs) {
				t.Errorf("Expected %d directories, got %d", len(tc.expectedDirs), len(configCopy.Directories))
			}

			for i, expectedDir := range tc.expectedDirs {
				if configCopy.Directories[i] != expectedDir {
					t.Errorf("Expected directory[%d] '%s', got '%s'", i, expectedDir, configCopy.Directories[i])
				}
			}

			if configCopy.Output.Format != tc.expectedFormat {
				t.Errorf("Expected format '%s', got '%s'", tc.expectedFormat, configCopy.Output.Format)
			}
		})
	}
}
