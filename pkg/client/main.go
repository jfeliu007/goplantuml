package client

import (
	"github.com/jfeliu007/goplantuml/pkg/client/controller"
	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/spf13/cobra"
)

// InitRootCmdForGoPlantUMLV2 creates the root command for goplantuml v2
func InitRootCmdForGoPlantUMLV2() *cobra.Command {
	var output string
	var rootCmd = &cobra.Command{
		Use:   "goplantuml",
		Short: "Generate PlantUML diagrams from Go source code (v2)",
		Long: `GoPlantUML v2 is a tool for generating PlantUML class diagrams from Go source code.

It analyzes your Go packages and creates visual representations showing:
- Struct definitions and their fields
- Method signatures
- Interface implementations
- Type compositions and aggregations
- Package relationships

The generated diagrams help visualize code architecture and dependencies.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			controller.SetOutputFormat(output)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "puml", "Output format: puml|json")
	return rootCmd
}

// ClientForGoPlantUMLV2 sets up the complete CLI for goplantuml v2
func ClientForGoPlantUMLV2(conf *config.BaseConfig) {
	rootCmd := InitRootCmdForGoPlantUMLV2()
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Add bootstrap command
	bootstrapCmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap and initialize goplantuml environment",
		Long:  "Bootstrap and initialize goplantuml environment with default settings",
	}
	bootstrapCmd.AddCommand(controller.InitBootstrapDiagramCmd(conf))
	rootCmd.AddCommand(bootstrapCmd)

	// Add generate command directly
	generateCmd := controller.InitGenerateCmd(conf)
	rootCmd.AddCommand(generateCmd)

	// Add validate command directly
	validateCmd := controller.InitValidateCmd(conf)
	rootCmd.AddCommand(validateCmd)

	// Add analyze command directly
	analyzeCmd := controller.InitAnalyzeCmd(conf)
	rootCmd.AddCommand(analyzeCmd)

	// Execute the root command
	rootCmd.Execute()
}
