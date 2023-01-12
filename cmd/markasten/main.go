package main

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

var (
	inputPath  *string
	outputPath *string
)

func newRootCmd() *cobra.Command {
	tagsCommand := &cobra.Command{
		Use:  "tags",
		RunE: tagsRunFn,
	}
	inputPath = tagsCommand.Flags().StringP("input", "i", "", "The location of the input files")
	outputPath = tagsCommand.Flags().StringP("output", "o", "", "The location of the output files")
	rootCmd := &cobra.Command{
		Use: "markasten",
	}
	rootCmd.AddCommand(tagsCommand)
	return rootCmd

}

func main() {
	rootCmd := newRootCmd()
	rootCmd.Execute()
}

func tagsRunFn(cmd *cobra.Command, args []string) error {
	fmt.Printf("tags called with -i %s and -o %s\n", *inputPath, *outputPath)
	ioutil.WriteFile(*outputPath, []byte("foo"), 0600)
	return nil
}
