package controller

import (
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/client/controller"
	"github.com/jfeliu007/goplantuml/pkg/config"
)

func TestSetOutputFormat(t *testing.T) {
	// Test SetOutputFormat function
	controller.SetOutputFormat("json")

	format := controller.GetOutputFormat()
	if format != "json" {
		t.Errorf("Expected format 'json', got '%s'", format)
	}
}

func TestGetOutputFormat_Default(t *testing.T) {
	// Test GetOutputFormat default value
	controller.SetOutputFormat("")

	format := controller.GetOutputFormat()
	if format != "puml" {
		t.Errorf("Expected default format 'puml', got '%s'", format)
	}
}

func TestFormatOutput(t *testing.T) {
	// Test FormatOutput function
	testData := map[string]interface{}{
		"test": "value",
	}

	// Test JSON format
	controller.SetOutputFormat("json")
	result := controller.FormatOutput(testData)

	if result == "" {
		t.Error("FormatOutput returned empty string")
	}

	// Test default format
	controller.SetOutputFormat("puml")
	result = controller.FormatOutput("test string")

	if result != "test string" {
		t.Errorf("Expected 'test string', got '%s'", result)
	}
}

func TestInitBootstrapDiagramCmd(t *testing.T) {
	// Test InitBootstrapDiagramCmd function
	baseConfig := &config.BaseConfig{}

	cmd := controller.InitBootstrapDiagramCmd(baseConfig)

	if cmd == nil {
		t.Fatal("InitBootstrapDiagramCmd returned nil")
	}

	if cmd.Use != "diagram" {
		t.Errorf("Expected Use to be 'diagram', got '%s'", cmd.Use)
	}
}
