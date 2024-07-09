package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	tagsInputPath    *string
	tagsOutputPath   *string
	title            *string
	tagsDebugEnabled *bool
	wikiLinks        *bool
	capitalize       *bool
	tagLinks         *bool
	toc              *bool
)

func newTagsCommand() *cobra.Command {
	tagsCommand := &cobra.Command{
		Use:  "tags",
		RunE: tagsRunFn,
	}
	tagsInputPath = tagsCommand.Flags().StringP("input", "i", "", "The location of the input files")
	tagsOutputPath = tagsCommand.Flags().StringP("output", "o", "", "The location of the output files")
	title = tagsCommand.Flags().StringP("title", "t", "Index", "The title of the generated index file")
	tagsDebugEnabled = tagsCommand.Flags().Bool("debug", false, "If set, debug logging will be enabled")
	wikiLinks = tagsCommand.Flags().Bool("wiki-links", false, "If set, links will be generated for a wiki with file extensions excluded")
	capitalize = tagsCommand.Flags().Bool("capitalize", false, "If set, tag names in the generated index will have their first character capitalized.")
	tagLinks = tagsCommand.Flags().Bool("tag-links", false, "If set, links to files in the generated index will be annotated with the list of other tags they have.")
	toc = tagsCommand.Flags().Bool("toc", false, "If set, a table of contents will be generated containing a link to the heading of each tag")
	debugEnabled = tagsDebugEnabled
	return tagsCommand
}

func tagsRunFn(cmd *cobra.Command, args []string) error {
	debug("tags called with -i %s and -o %s\n", *tagsInputPath, *tagsOutputPath)
	inputDirEntires, err := newFullDirEntryList(*tagsInputPath)
	if err != nil {
		panic(err)
	}

	filesByTags := make(map[string][]indexedFile)
	searchResults, err := searchForMarkdownFiles(inputDirEntires, *tagsInputPath)
	if err != nil {
		panic(err)
	}
	for _, dirEntry := range searchResults {
		fileBytes, err := os.ReadFile(dirEntry.Name())
		if err != nil {
			panic(err)
		}

		scrapedTags, title := scrapeTagsAndTitle(fileBytes)
		filesByTags = appendFilesByTags(scrapedTags, filesByTags, title, dirEntry.Name())
	}

	outputFile, err := os.Create(*tagsOutputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	writeOrPanic(outputFile, fmt.Sprintf("# %s\n", *title))

	var sortedTags []string
	for tag, _ := range filesByTags {
		sortedTags = append(sortedTags, tag)
	}
	sort.Strings(sortedTags)

	if *toc {
		writeOrPanic(outputFile, "\n")
		writeOrPanic(outputFile, "---\n")
		writeOrPanic(outputFile, "\n")
		writeOrPanic(outputFile, "## Table of contents\n")
		for _, tag := range sortedTags {
			header := tagToHeader(tag)
			writeOrPanic(outputFile, fmt.Sprintf("- [%s](#%s)\n", header, headerToLink(header)))
		}
		writeOrPanic(outputFile, "\n")
		writeOrPanic(outputFile, "---\n")
		writeOrPanic(outputFile, "\n")
	}

	for n, tag := range sortedTags {
		files := filesByTags[tag]
		writeOrPanic(outputFile, fmt.Sprintf("## %s\n", tagToHeader(tag)))

		countedTitles := countTitles(files)
		for m, f := range files {
			trailingChar := "\n"
			if m == len(files)-1 && n == len(sortedTags)-1 {
				trailingChar = ""
			}
			relativePath := relativeTo(f.fileName, *tagsOutputPath)
			if *wikiLinks {
				relativePath = makeWikiLink(relativePath)
			}
			title := f.title
			if count, ok := countedTitles[f.title]; ok && count > 1 {
				title = relativePath
			}

			line := fmt.Sprintf(
				"- [%s](%s)",
				title,
				relativePath,
			)
			if *tagLinks {
				for _, otherTag := range f.otherTags {
					line = fmt.Sprintf("%s `%s`", line, otherTag)
				}
			}
			line = fmt.Sprintf(
				"%s%s",
				line,
				trailingChar,
			)

			writeOrPanic(outputFile, line)
		}
		if n < len(sortedTags)-1 {
			writeOrPanic(outputFile, "\n")
		}
	}

	return nil
}

func tagToHeader(tag string) string {
	if *capitalize && len(tag) > 0 {
		return fmt.Sprintf("%s%s", strings.ToUpper(tag[0:1]), tag[1:])
	}
	return tag
}

func scrapeTagsAndTitle(fileBytes []byte) ([]string, string) {
	var scrapedTags []string
	title := ""
	lines := strings.Split(string(fileBytes), "\n")
	if firstNonEmptyLine(lines) != "---" {
		debug("first line was %s, no tags detected", lines[0])
		return scrapedTags, title
	}

	foundYaml := false
	finishedYaml := false
	var yamlLines []string
	for _, line := range lines {
		if len(line) > 2 && line[0:2] == "# " {
			title = line[2:]
			break
		}
		if line == "---" {
			if foundYaml {
				// If we had previously found yaml, this line
				// marks the end of the yaml.
				finishedYaml = true
			}
			if !foundYaml {
				// if we haven't found yaml yet, this line
				// marks the beginning of yaml
				foundYaml = true
			}
			continue
		}
		if len(line) == 0 {
			continue
		}
		if foundYaml && !finishedYaml {
			yamlLines = append(yamlLines, line)
		}
	}
	debug("found yaml:")
	for _, line := range yamlLines {
		debug(line)
	}

	fm := frontmatter{}
	err := yaml.Unmarshal([]byte(strings.Join(yamlLines, "\n")), &fm)
	if err != nil {
		debug("error unmarshalling yaml: %s", err)
		return scrapedTags, title
	}
	return fm.Tags, title
}

func appendFilesByTags(scrapedTags []string, filesByTags map[string][]indexedFile, title string, name string) map[string][]indexedFile {
	for _, tagName := range scrapedTags {
		tagName := tagName
		file := indexedFile{
			fileName:  name,
			title:     title,
			otherTags: getOtherTags(scrapedTags, tagName),
		}
		if files, ok := filesByTags[tagName]; ok {
			filesByTags[tagName] = append(files, file)
		} else {
			filesByTags[tagName] = []indexedFile{file}
		}
	}
	return filesByTags
}

func getOtherTags(tags []string, tag string) []string {
	var otherTags []string
	for _, t := range tags {
		if t == tag {
			continue
		}
		otherTags = append(otherTags, t)
	}
	return otherTags
}
