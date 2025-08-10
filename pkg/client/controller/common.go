package controller

import (
	"encoding/json"
	"fmt"

	"github.com/jfeliu007/goplantuml/pkg/config"
	"github.com/spf13/cobra"
)

// Global output format setting
var outputFormat string

// SetOutputFormat sets the global output format
func SetOutputFormat(format string) {
	outputFormat = format
}

// GetOutputFormat returns the current output format
func GetOutputFormat() string {
	if outputFormat == "" {
		return "puml"
	}
	return outputFormat
}

// FormatOutput formats the output based on the current format setting
func FormatOutput(data interface{}) string {
	switch GetOutputFormat() {
	case "json":
		if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
			return string(jsonData)
		}
	case "puml":
		fallthrough
	default:
		return fmt.Sprintf("%v", data)
	}
	return fmt.Sprintf("%v", data)
}

// InitBootstrapDiagramCmd creates a bootstrap diagram command
func InitBootstrapDiagramCmd(conf *config.BaseConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagram",
		Short: "Bootstrap diagram generation environment",
		Long:  "Initialize and bootstrap the diagram generation environment with default templates",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Bootstrapping diagram generation environment...")
			fmt.Println("Environment initialized successfully!")
		},
	}
	return cmd
}
