package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	inputPath  *string
	outputPath *string
	title      *string
)

func newRootCmd() *cobra.Command {
	tagsCommand := &cobra.Command{
		Use:  "tags",
		RunE: tagsRunFn,
	}
	inputPath = tagsCommand.Flags().StringP("input", "i", "", "The location of the input files")
	outputPath = tagsCommand.Flags().StringP("output", "o", "", "The location of the output files")
	title = tagsCommand.Flags().StringP("title", "t", "Index", "The title of the generated index file")
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

type indexedFile struct {
	fileName string
	title    string
}

func tagsRunFn(cmd *cobra.Command, args []string) error {
	fmt.Printf("tags called with -i %s and -o %s\n", *inputPath, *outputPath)
	inputDirEntires, err := os.ReadDir(*inputPath)
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile("`[^`]+`")

	filesByTags := make(map[string][]indexedFile)
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

		title := ""
		name := dirEntry.Name()
		var scrapedTags []string
		lines := strings.Split(string(fileBytes), "\n")
		for _, line := range lines {
			if len(line) > 2 && line[0:2] == "# " {
				title = line[2:]
				break
			}
			if line == "---" {
				continue
			}
			if len(line) == 0 {
				continue
			}
			tags := re.FindAllString(line, -1)
			for _, tag := range tags {
				tagName := strings.Replace(tag, "`", "", -1)
				scrapedTags = append(scrapedTags, tagName)
			}
		}
		for _, tagName := range scrapedTags {
			tagName := tagName
			file := indexedFile{
				fileName: name,
				title:    title,
			}
			if files, ok := filesByTags[tagName]; ok {
				filesByTags[tagName] = append(files, file)
			} else {
				filesByTags[tagName] = []indexedFile{file}
			}
		}
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	_, err = io.WriteString(outputFile, fmt.Sprintf("# %s\n", *title))
	if err != nil {
		panic(err)
	}

	var sortedTags []string
	for tag, _ := range filesByTags {
		sortedTags = append(sortedTags, tag)
	}
	sort.Strings(sortedTags)

	for n, tag := range sortedTags {
		files := filesByTags[tag]
		_, err = io.WriteString(outputFile, fmt.Sprintf("## %s\n", tag))
		if err != nil {
			panic(err)
		}
		for m, f := range files {
			trailingChar := "\n"
			if m == len(files)-1 && n == len(sortedTags)-1 {
				trailingChar = ""
			}
			_, err = io.WriteString(outputFile, fmt.Sprintf("- [%s](%s)%s", f.title, f.fileName, trailingChar))
			if err != nil {
				panic(err)
			}
		}
		if n < len(sortedTags)-1 {
			_, err = io.WriteString(outputFile, "\n")
			if err != nil {
				panic(err)
			}
		}
	}

	return nil
}
