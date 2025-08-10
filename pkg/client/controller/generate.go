package controller

import (
	"fmt"
	"os"

	"github.com/jfeliu007/goplantuml/pkg/client/usecase"
	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/jfeliu007/goplantuml/pkg/entity/request"
	"github.com/spf13/cobra"
)

// InitGenerateCmd creates a generate command for PlantUML diagram generation
func InitGenerateCmd(conf *config.BaseConfig) *cobra.Command {
	uc := usecase.NewPlantUMLUsecase(conf)

	generateCmd := &cobra.Command{
		Use:   "generate [directories...]",
		Short: "Generate PlantUML diagram from Go source code",
		Long: `Generate PlantUML class diagrams from Go source code.
		
You can specify one or more directories to analyze. The tool will parse
the Go source code and generate a PlantUML diagram showing the structure
and relationships of your code.`,
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// Check if config file is specified
			configFile, _ := cmd.Flags().GetString("config")
			if configFile != "" {
				runWithConfigFile(cmd, args, configFile, uc)
				return
			}

			// Fallback to traditional flag-based execution
			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, "Error: At least one directory must be specified when not using config file\n")
				os.Exit(1)
			}

			runWithFlags(cmd, args, uc)
		},
	}

	// Add flags
	generateCmd.Flags().StringP("config", "c", "", "Path to YAML configuration file")
	generateCmd.Flags().BoolP("recursive", "r", false, "Walk all directories recursively")
	generateCmd.Flags().String("ignore", "", "Comma separated list of folders to ignore")
	generateCmd.Flags().Bool("show-aggregations", false, "Renders public aggregations even when -hide-connections is used")
	generateCmd.Flags().Bool("hide-fields", false, "Hides fields")
	generateCmd.Flags().Bool("hide-methods", false, "Hides methods")
	generateCmd.Flags().Bool("hide-connections", false, "Hides all connections in the diagram")
	generateCmd.Flags().Bool("show-compositions", false, "Shows compositions even when -hide-connections is used")
	generateCmd.Flags().Bool("show-implementations", false, "Shows implementations even when -hide-connections is used")
	generateCmd.Flags().Bool("show-aliases", false, "Shows aliases even when -hide-connections is used")
	generateCmd.Flags().Bool("show-connection-labels", false, "Shows labels in the connections")
	generateCmd.Flags().StringP("title", "t", "", "Title of the generated diagram")
	generateCmd.Flags().StringP("notes", "n", "", "Comma separated list of notes to be added to the diagram")
	generateCmd.Flags().StringP("output", "o", "", "Output file path. If omitted, outputs to stdout")
	generateCmd.Flags().Bool("show-options-as-note", false, "Show a note in the diagram with the rendering options")
	generateCmd.Flags().Bool("aggregate-private-members", false, "Show aggregations for private members")
	generateCmd.Flags().Bool("hide-private-members", false, "Hide private fields and methods")
	generateCmd.Flags().String("custom-resources", "", "Comma separated list of custom resource patterns for function categorization")
	generateCmd.Flags().String("custom-keywords", "", "Custom keywords in format 'ResourceType:keyword1,keyword2;AnotherType:keyword3,keyword4' (e.g., 'User:GetUser,CreateUser;Order:ProcessOrder')")

	return generateCmd
}

// InitValidateCmd creates a validate command for checking directory structure
func InitValidateCmd(conf *config.BaseConfig) *cobra.Command {
	uc := usecase.NewPlantUMLUsecase(conf)

	validateCmd := &cobra.Command{
		Use:   "validate [directories...]",
		Short: "Validate Go source directories",
		Long:  "Validate that the specified directories exist and contain Go source files.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
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
			fmt.Println("All directories are valid!")
		},
	}

	return validateCmd
}

// InitAnalyzeCmd creates an analyze command for exploring code structure
func InitAnalyzeCmd(conf *config.BaseConfig) *cobra.Command {
	analyzeCmd := &cobra.Command{
		Use:   "analyze [directories...]",
		Short: "Analyze Go source code structure",
		Long:  "Analyze Go source code and provide information about the structure without generating diagrams.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Analyzing Go source code structure...")
			// TODO: Implement analysis functionality
			fmt.Printf("Directories to analyze: %v\n", args)
		},
	}

	return analyzeCmd
}
