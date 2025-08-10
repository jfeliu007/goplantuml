package config

import (
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/config"
)

func TestYamlConfig_ToRepositoryDiagramOptions(t *testing.T) {
	// Test v2 options conversion
	v2Config := &config.Config{
		Version: "v2",
		RenderingOptions: config.RenderingOptions{
			ShowAggregations:        true,
			HideFields:              false,
			HideMethods:             true,
			ShowCompositions:        true,
			ShowImplementations:     true,
			ShowAliases:             true,
			ShowConnectionLabels:    false,
			AggregatePrivateMembers: true,
			HidePrivateMembers:      false,
			Title:                   "Test Diagram",
			Notes:                   "Test Notes",
		},
	}

	options := v2Config.ToRepositoryDiagramOptions()

	// For v2, it should return DiagramOptions struct
	diagramOptions, ok := options.(config.DiagramOptions)
	if !ok {
		t.Fatalf("Expected DiagramOptions for v2 config, got %T", options)
	}

	// Verify option values
	if !diagramOptions.ShowAggregations {
		t.Error("Expected ShowAggregations to be true")
	}

	if diagramOptions.HideFields {
		t.Error("Expected HideFields to be false")
	}

	if !diagramOptions.HideMethods {
		t.Error("Expected HideMethods to be true")
	}

	if diagramOptions.Title != "Test Diagram" {
		t.Errorf("Expected title 'Test Diagram', got '%s'", diagramOptions.Title)
	}

	if diagramOptions.Notes != "Test Notes" {
		t.Errorf("Expected notes 'Test Notes', got '%s'", diagramOptions.Notes)
	}
}
