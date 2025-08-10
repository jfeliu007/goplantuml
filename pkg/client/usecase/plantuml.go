package usecase

import (
	"fmt"
	"os"

	"github.com/jfeliu007/goplantuml/pkg/client/repository"
	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/jfeliu007/goplantuml/pkg/entity/request"
)

// PlantUMLUsecase handles PlantUML diagram generation business logic
type PlantUMLUsecase struct {
	config *config.BaseConfig
	repo   repository.PlantUMLRepository
}

// NewPlantUMLUsecase creates a new PlantUMLUsecase
func NewPlantUMLUsecase(conf *config.BaseConfig) *PlantUMLUsecase {
	return &PlantUMLUsecase{
		config: conf,
		repo:   repository.NewPlantUMLRepository(),
	}
}

// Generate generates a PlantUML diagram based on the request
func (u *PlantUMLUsecase) Generate(req request.GenerateRequest) (string, error) {
	// Convert request to repository options
	options := repository.DiagramOptions{
		ShowAggregations:        req.ShowAggregations,
		HideFields:              req.HideFields,
		HideMethods:             req.HideMethods,
		HideConnections:         req.HideConnections,
		ShowCompositions:        req.ShowCompositions,
		ShowImplementations:     req.ShowImplementations,
		ShowAliases:             req.ShowAliases,
		ShowConnectionLabels:    req.ShowConnectionLabels,
		AggregatePrivateMembers: req.AggregatePrivateMembers,
		HidePrivateMembers:      req.HidePrivateMembers,
		Title:                   req.Title,
		Notes:                   req.Notes,
		CustomResources:         req.CustomResources,
		CustomKeywords:          req.CustomKeywords,
	}

	// Generate diagram using repository
	result, err := u.repo.GenerateDiagram(req.Directories, req.IgnoredDirectories, req.Recursive, options)
	if err != nil {
		return "", fmt.Errorf("failed to generate diagram: %w", err)
	}

	// Write to file if output path is specified
	if req.OutputPath != "" {
		if err := u.writeToFile(req.OutputPath, result); err != nil {
			return "", fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return result, nil
}

// GenerateFromConfig generates a PlantUML diagram using the configured settings
func (u *PlantUMLUsecase) GenerateFromConfig() (string, error) {
	req := request.GenerateRequest{
		Directories:             u.config.Directories,
		IgnoredDirectories:      u.config.IgnoredDirectories,
		Recursive:               u.config.Recursive,
		OutputPath:              u.config.OutputPath,
		OutputFormat:            u.config.OutputFormat,
		ShowAggregations:        u.config.ShowAggregations,
		HideFields:              u.config.HideFields,
		HideMethods:             u.config.HideMethods,
		HideConnections:         u.config.HideConnections,
		ShowCompositions:        u.config.ShowCompositions,
		ShowImplementations:     u.config.ShowImplementations,
		ShowAliases:             u.config.ShowAliases,
		ShowConnectionLabels:    u.config.ShowConnectionLabels,
		AggregatePrivateMembers: u.config.AggregatePrivateMembers,
		HidePrivateMembers:      u.config.HidePrivateMembers,
		ShowOptionsAsNote:       u.config.ShowOptionsAsNote,
		Title:                   u.config.Title,
		Notes:                   u.config.Notes,
	}

	return u.Generate(req)
}

// writeToFile writes content to a file
func (u *PlantUMLUsecase) writeToFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// ValidateDirectories validates that directories exist and are accessible
func (u *PlantUMLUsecase) ValidateDirectories(req request.ValidateRequest) map[string]interface{} {
	result := map[string]interface{}{
		"valid":       true,
		"errors":      []string{},
		"checked":     len(req.Directories),
		"directories": req.Directories,
	}

	// Use repository to validate directories
	if err := u.repo.ValidateDirectories(req.Directories); err != nil {
		result["valid"] = false
		result["errors"] = []string{err.Error()}
	}

	return result
}

// AnalyzeStructure analyzes Go source code structure without generating diagrams
func (u *PlantUMLUsecase) AnalyzeStructure(req request.AnalyzeRequest) map[string]interface{} {
	// Use repository to analyze structure
	analysis, err := u.repo.AnalyzeCodeStructure(req.Directories, req.Recursive)
	if err != nil {
		return map[string]interface{}{
			"analyzed": false,
			"error":    err.Error(),
		}
	}

	return map[string]interface{}{
		"analyzed":    true,
		"directories": req.Directories,
		"recursive":   req.Recursive,
		"packages":    analysis.Packages,
		"structs":     analysis.Structs,
		"interfaces":  analysis.Interfaces,
		"summary":     analysis.Summary,
	}
}
