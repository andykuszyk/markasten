package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	inputDirEntires, err := os.ReadDir(*inputPath)
	if err != nil {
		panic(err)
	}

	filesByTags := make(map[string][]string)
	for _, dirEntry := range inputDirEntires {
		if dirEntry.IsDir() {
			// TODO scan file tree recursively.
			continue
		}
		filePath := filepath.Join(*inputPath, dirEntry.Name())
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}
		lines := strings.Split(string(fileBytes), "\n")
		for n, line := range lines {
			if n > 2 {
				break
			}
			if line == "---" {
				continue
			}
			tags := strings.Split(line, " ")
			for _, tag := range tags {
				if files, ok := filesByTags[tag]; ok {
					filesByTags[tag] = append(files, dirEntry.Name())
				} else {
					filesByTags[tag] = []string{dirEntry.Name()}
				}
			}
		}
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	for tag, files := range filesByTags {
		io.WriteString(outputFile, fmt.Sprintf("## %s\n", tag))
		for _, f := range files {
			io.WriteString(outputFile, fmt.Sprintf("- %s\n", f))
		}
		io.WriteString(outputFile, "\n")
	}

	return nil
}
