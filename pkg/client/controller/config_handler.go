package controller

import (
	"fmt"
	"os"
	"strings"

	"github.com/jfeliu007/goplantuml/pkg/client/usecase"
	yamlconfig "github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/jfeliu007/goplantuml/pkg/entity/request"
	"github.com/spf13/cobra"
)

// runWithConfigFile handles execution using YAML configuration file
func runWithConfigFile(cmd *cobra.Command, args []string, configFile string, uc *usecase.PlantUMLUsecase) error {
	// Load configuration from YAML file
	config, err := yamlconfig.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
		os.Exit(1)
	}

	// Override directories from command line args if provided
	directories := config.Directories
	if len(args) > 0 {
		directories = args
	}

	// Validate directories
	validateReq := request.ValidateRequest{
		Directories: directories,
	}
	validationResult := uc.ValidateDirectories(validateReq)
	if !validationResult["valid"].(bool) {
		errors := validationResult["errors"].([]string)
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	// Create request from config
	req := request.GenerateRequest{
		Directories:             directories,
		IgnoredDirectories:      config.IgnoredDirectories,
		Recursive:               config.Recursive,
		OutputPath:              config.Output.File,
		ShowAggregations:        config.RenderingOptions.ShowAggregations,
		HideFields:              config.RenderingOptions.HideFields,
		HideMethods:             config.RenderingOptions.HideMethods,
		HideConnections:         config.RenderingOptions.HideConnections,
		ShowCompositions:        config.RenderingOptions.ShowCompositions,
		ShowImplementations:     config.RenderingOptions.ShowImplementations,
		ShowAliases:             config.RenderingOptions.ShowAliases,
		ShowConnectionLabels:    config.RenderingOptions.ShowConnectionLabels,
		AggregatePrivateMembers: config.RenderingOptions.AggregatePrivateMembers,
		HidePrivateMembers:      config.RenderingOptions.HidePrivateMembers,
		ShowOptionsAsNote:       false, // Not in YAML config for now
		Title:                   config.RenderingOptions.Title,
		Notes:                   config.RenderingOptions.Notes,
		CustomResources:         []string{}, // Can be added to YAML later
		CustomKeywords:          config.CustomKeywords,
	}

	// Allow CLI -o to override YAML output.file
	if outputOverride, _ := cmd.Flags().GetString("output"); outputOverride != "" {
		req.OutputPath = outputOverride
		config.Output.File = outputOverride // keep success message consistent
	}

	// Check version compatibility
	if config.IsV1() {
		fmt.Println("Note: Using v1 compatibility mode")
		// TODO: Implement v1 mode if needed
	} else if config.IsV2() {
		fmt.Println("Note: Using v2 mode")
	}

	// Generate diagram
	result, err := uc.Generate(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating diagram: %v\n", err)
		os.Exit(1)
	}

	// Output result
	if config.Output.File == "" {
		fmt.Print(result)
	} else {
		fmt.Printf("PlantUML diagram generated successfully: %s\n", config.Output.File)
	}

	return nil
}

// runWithFlags handles execution using command line flags (legacy mode)
func runWithFlags(cmd *cobra.Command, args []string, uc *usecase.PlantUMLUsecase) error {
	// Get flags
	recursive, _ := cmd.Flags().GetBool("recursive")
	ignore, _ := cmd.Flags().GetString("ignore")
	showAggregations, _ := cmd.Flags().GetBool("show-aggregations")
	hideFields, _ := cmd.Flags().GetBool("hide-fields")
	hideMethods, _ := cmd.Flags().GetBool("hide-methods")
	hideConnections, _ := cmd.Flags().GetBool("hide-connections")
	showCompositions, _ := cmd.Flags().GetBool("show-compositions")
	showImplementations, _ := cmd.Flags().GetBool("show-implementations")
	showAliases, _ := cmd.Flags().GetBool("show-aliases")
	showConnectionLabels, _ := cmd.Flags().GetBool("show-connection-labels")
	title, _ := cmd.Flags().GetString("title")
	notes, _ := cmd.Flags().GetString("notes")
	output, _ := cmd.Flags().GetString("output")
	showOptionsAsNote, _ := cmd.Flags().GetBool("show-options-as-note")
	aggregatePrivateMembers, _ := cmd.Flags().GetBool("aggregate-private-members")
	hidePrivateMembers, _ := cmd.Flags().GetBool("hide-private-members")
	customResourcesStr, _ := cmd.Flags().GetString("custom-resources")
	customKeywordsStr, _ := cmd.Flags().GetString("custom-keywords")

	// Parse ignored directories
	var ignoredDirs []string
	if ignore != "" {
		ignoredDirs = strings.Split(ignore, ",")
		for i, dir := range ignoredDirs {
			ignoredDirs[i] = strings.TrimSpace(dir)
		}
	}

	// Parse custom resources
	var customResources []string
	if customResourcesStr != "" {
		customResources = strings.Split(customResourcesStr, ",")
		for i, resource := range customResources {
			customResources[i] = strings.TrimSpace(resource)
		}
	}

	// Parse custom keywords from JSON-like format: "Type1:keyword1,keyword2;Type2:keyword3,keyword4"
	customKeywords := make(map[string][]string)
	if customKeywordsStr != "" {
		pairs := strings.Split(customKeywordsStr, ";")
		for _, pair := range pairs {
			parts := strings.Split(pair, ":")
			if len(parts) == 2 {
				resourceType := strings.TrimSpace(parts[0])
				keywordsStr := strings.TrimSpace(parts[1])
				if keywordsStr != "" {
					keywords := strings.Split(keywordsStr, ",")
					for i, keyword := range keywords {
						keywords[i] = strings.TrimSpace(keyword)
					}
					customKeywords[resourceType] = keywords
				}
			}
		}
	}

	// Validate directories
	validateReq := request.ValidateRequest{
		Directories: args,
	}
	validationResult := uc.ValidateDirectories(validateReq)
	if !validationResult["valid"].(bool) {
		errors := validationResult["errors"].([]string)
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	// Create request
	req := request.GenerateRequest{
		Directories:             args,
		IgnoredDirectories:      ignoredDirs,
		Recursive:               recursive,
		OutputPath:              output,
		ShowAggregations:        showAggregations,
		HideFields:              hideFields,
		HideMethods:             hideMethods,
		HideConnections:         hideConnections,
		ShowCompositions:        showCompositions,
		ShowImplementations:     showImplementations,
		ShowAliases:             showAliases,
		ShowConnectionLabels:    showConnectionLabels,
		AggregatePrivateMembers: aggregatePrivateMembers,
		HidePrivateMembers:      hidePrivateMembers,
		ShowOptionsAsNote:       showOptionsAsNote,
		Title:                   title,
		Notes:                   notes,
		CustomResources:         customResources,
		CustomKeywords:          customKeywords,
	}

	// Generate diagram
	result, err := uc.Generate(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating diagram: %v\n", err)
		os.Exit(1)
	}

	// Output result
	if output == "" {
		fmt.Print(result)
	} else {
		fmt.Printf("PlantUML diagram generated successfully: %s\n", output)
	}

	return nil
}
