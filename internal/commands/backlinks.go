package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	backlinksFindInputPath    *string
	backlinksFindOutputPath   *string
	backlinksFindDebugEnabled *bool
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
		backlinks, err := scrapeBacklinks(fileBytes)
		if err != nil {
			panic(err)
		}
		backlinksByFile[dirEntry.Name()] = backlinks
	}

	outputFile, err := os.Create(*backlinksFindOutputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	writeOrPanic(outputFile, `foo.md:
  - bar.md`)
	return nil
}

func scrapeBacklinks(_ []byte) ([]string, error) {
	return nil, nil
}
