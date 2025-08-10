package config

import (
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/config"
)

func TestYamlConfig_IsV1_IsV2(t *testing.T) {
	testCases := []struct {
		name       string
		version    string
		expectedV1 bool
		expectedV2 bool
	}{
		{
			name:       "Version v1",
			version:    "v1",
			expectedV1: true,
			expectedV2: false,
		},
		{
			name:       "Version v2",
			version:    "v2",
			expectedV1: false,
			expectedV2: true,
		},
		{
			name:       "Empty version (defaults to v2)",
			version:    "",
			expectedV1: false,
			expectedV2: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &config.Config{
				Version: tc.version,
			}

			if config.IsV1() != tc.expectedV1 {
				t.Errorf("Expected IsV1() to return %v for version '%s'", tc.expectedV1, tc.version)
			}

			if config.IsV2() != tc.expectedV2 {
				t.Errorf("Expected IsV2() to return %v for version '%s'", tc.expectedV2, tc.version)
			}
		})
	}
}
