package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	backlinksFindInputPath    *string
	backlinksFindOutputPath   *string
	backlinksFindDebugEnabled *bool
	LinkRegexp                = regexp.MustCompile(`\[.*\]\(.*\)`)
)

func newBacklinksCommand() *cobra.Command {
	backlinkCommand := &cobra.Command{
		Use: "backlinks",
	}
	findCommand := &cobra.Command{
		Use:  "find",
		RunE: backlinkFindRunFn,
	}
	backlinksFindDebugEnabled = findCommand.Flags().Bool("debug", false, "If set, debug logging will be enabled")
	backlinksFindInputPath = findCommand.Flags().StringP("input", "i", "", "The location of the input files")
	backlinksFindOutputPath = findCommand.Flags().StringP("output", "o", "", "The location of the output file")
	backlinkCommand.AddCommand(findCommand)
	debugEnabled = backlinksFindDebugEnabled
	return backlinkCommand
}

func backlinkFindRunFn(cmd *cobra.Command, args []string) error {
	debug("backlink find called with -i %s and -o %s\n", *backlinksFindInputPath, *backlinksFindOutputPath)
	inputDirEntires, err := newFullDirEntryList(*backlinksFindInputPath)
	if err != nil {
		panic(err)
	}

	backlinksByFile := make(map[string][]string)
	searchResults, err := searchForMarkdownFiles(inputDirEntires, *backlinksFindInputPath)
	if err != nil {
		panic(err)
	}
	for _, dirEntry := range searchResults {
		fileBytes, err := os.ReadFile(dirEntry.Name())
		if err != nil {
			panic(err)
		}
		backlinksByFile[dirEntry.Name()] = scrapeBacklinks(fileBytes)
	}

	outputFile, err := os.Create(*backlinksFindOutputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	for fileName, backlinks := range backlinksByFile {
		writeOrPanic(outputFile, fmt.Sprintf("%s:\n", relativeTo(fileName, *backlinksFindInputPath)))
		for _, backlink := range backlinks {
			writeOrPanic(outputFile, fmt.Sprintf("  - %s\n", relativeTo(backlink, *backlinksFindInputPath)))
		}
	}
	return nil
}

func scrapeBacklinks(fileBytes []byte) []string {
	lines := strings.Split(string(fileBytes), "\n")
	var backlinks []string
	for _, line := range lines {
		matches := LinkRegexp.FindAllString(line, -1)
		if matches == nil {
			continue
		}
		for _, match := range matches {
			backlinks = append(backlinks, match)
		}
	}
	return backlinks
}
