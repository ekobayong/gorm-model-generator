package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gorm-model-generator",
	Short: "A CLI tool to generate GORM models from an existing database",
	Long:  `gorm-model-generator is a CLI tool that connects to your database, inspects the schema, and generates Go structs with GORM tags.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be defined here
}
