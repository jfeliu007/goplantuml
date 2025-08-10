package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/jfeliu007/goplantuml/pkg/client/usecase"
	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/jfeliu007/goplantuml/pkg/entity/request"
	"github.com/spf13/cobra"
)

// ClientForLegacyModeV1 provides backward compatibility with the original CLI interface (v1)
func ClientForLegacyModeV1(conf *config.BaseConfig) {
	rootCmd := &cobra.Command{
		Use:   "goplantuml [directories...]",
		Short: "Generate PlantUML diagrams from Go source code (v1)",
		Long: `GoPlantUML v1 generates PlantUML class diagrams from Go source code.

This tool analyzes your Go packages and creates visual representations showing:
- Struct definitions and their fields  
- Method signatures
- Interface implementations
- Type compositions and aggregations
- Package relationships

Usage examples:
  goplantuml ./src
  goplantuml -recursive ./src ./lib
  goplantuml -output diagram.puml ./src`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			uc := usecase.NewPlantUMLUsecase(conf)

			// Get all flags
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

			// Parse ignored directories
			var ignoredDirs []string
			if ignore != "" {
				ignoredDirs = strings.Split(ignore, ",")
				for i, dir := range ignoredDirs {
					ignoredDirs[i] = strings.TrimSpace(dir)
				}
			}

			// Validate directories
			validateReq := request.ValidateRequest{
				Directories: args,
			}
			validationResult := uc.ValidateDirectories(validateReq)
			if !validationResult["valid"].(bool) {
				fmt.Printf("usage:\\ngoplantuml <DIR>\\nDIR Must be a valid directory\\n")
				errors := validationResult["errors"].([]string)
				for _, err := range errors {
					fmt.Fprintf(os.Stderr, "Error: %v\\n", err)
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
			}

			// Generate diagram
			result, err := uc.Generate(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\\n", err)
				os.Exit(1)
			}

			// Output result to stdout if no output file specified
			if output == "" {
				fmt.Print(result)
			}
		},
	}

	// Add all original flags for backward compatibility
	rootCmd.Flags().Bool("recursive", false, "walk all directories recursively")
	rootCmd.Flags().String("ignore", "", "comma separated list of folders to ignore")
	rootCmd.Flags().Bool("show-aggregations", false, "renders public aggregations even when -hide-connections is used (do not render by default)")
	rootCmd.Flags().Bool("hide-fields", false, "hides fields")
	rootCmd.Flags().Bool("hide-methods", false, "hides methods")
	rootCmd.Flags().Bool("hide-connections", false, "hides all connections in the diagram")
	rootCmd.Flags().Bool("show-compositions", false, "Shows compositions even when -hide-connections is used")
	rootCmd.Flags().Bool("show-implementations", false, "Shows implementations even when -hide-connections is used")
	rootCmd.Flags().Bool("show-aliases", false, "Shows aliases even when -hide-connections is used")
	rootCmd.Flags().Bool("show-connection-labels", false, "Shows labels in the connections to identify the connections types (e.g. extends, implements, aggregates, alias of")
	rootCmd.Flags().String("title", "", "Title of the generated diagram")
	rootCmd.Flags().String("notes", "", "Comma separated list of notes to be added to the diagram")
	rootCmd.Flags().String("output", "", "output file path. If omitted, then this will default to standard output")
	rootCmd.Flags().Bool("show-options-as-note", false, "Show a note in the diagram with the none evident options ran with this CLI")
	rootCmd.Flags().Bool("aggregate-private-members", false, "Show aggregations for private members. Ignored if -show-aggregations is not used.")
	rootCmd.Flags().Bool("hide-private-members", false, "Hide private fields and methods")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\\n", err)
		os.Exit(1)
	}
}
