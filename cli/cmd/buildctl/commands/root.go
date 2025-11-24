package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "buildctl",
	Short: "Build Assistant CLI",
	Long:  "CLI tool for interacting with Build Assistant system",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(eventCmd)
	rootCmd.AddCommand(artifactCmd)
	rootCmd.AddCommand(containerCmd)
}

