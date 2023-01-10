package main

import (
	"github.com/spf13/cobra"
)

var (
	inputPath *string
	outputPath *string
)

func main() {
	tagsCommand := &cobra.Command{
		Use: "tags",
		RunE: tagsRunFn,
	}
	inputPath = tagsCommand.Flags().StringP("input", "i", "", "The location of the input files")
	outputPath = tagsCommand.Flags().StringP("output", "o", "", "The location of the output files")
	rootCmd := &cobra.Command{
		Use: "markasten",
	}
	rootCmd.AddCommand(tagsCommand)
	rootCmd.Execute()
}

func tagsRunFn(cmd *cobra.Command, args []string) error {
	return nil
}
