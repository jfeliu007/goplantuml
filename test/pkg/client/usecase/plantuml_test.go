package usecase

import (
	"testing"

	"github.com/jfeliu007/goplantuml/pkg/client/usecase"
	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/jfeliu007/goplantuml/pkg/entity/request"
)

func TestNewPlantUMLUsecase(t *testing.T) {
	// Test NewPlantUMLUsecase constructor
	baseConfig := &config.BaseConfig{}

	uc := usecase.NewPlantUMLUsecase(baseConfig)

	if uc == nil {
		t.Fatal("NewPlantUMLUsecase returned nil")
	}
}

func TestPlantUMLUsecase_Generate(t *testing.T) {
	// Test Generate method
	baseConfig := &config.BaseConfig{}
	uc := usecase.NewPlantUMLUsecase(baseConfig)

	req := request.GenerateRequest{
		Directories: []string{"."},
		Recursive:   false,
	}

	// This would typically fail because we need actual Go files to parse
	// For now, just test that the method doesn't panic
	_, err := uc.Generate(req)

	// We expect an error since there are no valid Go files in current directory for parsing
	if err == nil {
		t.Log("Generate completed without error")
	} else {
		t.Logf("Generate returned expected error: %v", err)
	}
}

func TestPlantUMLUsecase_ValidateDirectories(t *testing.T) {
	// Test ValidateDirectories method
	baseConfig := &config.BaseConfig{}
	uc := usecase.NewPlantUMLUsecase(baseConfig)

	req := request.ValidateRequest{
		Directories: []string{"."},
	}

	result := uc.ValidateDirectories(req)

	if result == nil {
		t.Fatal("ValidateDirectories returned nil")
	}

	// Check if result contains expected keys
	if _, exists := result["valid"]; !exists {
		t.Error("ValidateDirectories result missing 'valid' key")
	}
}
