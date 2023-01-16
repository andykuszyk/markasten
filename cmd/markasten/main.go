package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
				tagName := strings.Replace(tag, "`", "", -1)
				if files, ok := filesByTags[tagName]; ok {
					filesByTags[tagName] = append(files, dirEntry.Name())
				} else {
					filesByTags[tagName] = []string{dirEntry.Name()}
				}
			}
		}
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	_, err = io.WriteString(outputFile, "# Index\n")
	if err != nil {
		panic(err)
	}

	var sortedTags []string
	for tag, _ := range filesByTags {
		sortedTags = append(sortedTags, tag)
	}
	sort.Strings(sortedTags)

	for _, tag := range sortedTags {
		files := filesByTags[tag]
		_, err = io.WriteString(outputFile, fmt.Sprintf("## %s\n", tag))
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			_, err = io.WriteString(outputFile, fmt.Sprintf("- %s\n", f))
			if err != nil {
				panic(err)
			}
		}
		_, err = io.WriteString(outputFile, "\n")
		if err != nil {
			panic(err)
		}
	}

	return nil
}
