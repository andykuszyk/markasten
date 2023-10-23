package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	inputPath    *string
	outputPath   *string
	title        *string
	debugEnabled *bool
	wikiLinks    *bool
	capitalize   *bool
	tagLinks     *bool
	toc          *bool
	re           = regexp.MustCompile("`[^`]+`")
)

func newRootCmd() *cobra.Command {
	tagsCommand := &cobra.Command{
		Use:  "tags",
		RunE: tagsRunFn,
	}
	inputPath = tagsCommand.Flags().StringP("input", "i", "", "The location of the input files")
	outputPath = tagsCommand.Flags().StringP("output", "o", "", "The location of the output files")
	title = tagsCommand.Flags().StringP("title", "t", "Index", "The title of the generated index file")
	debugEnabled = tagsCommand.Flags().Bool("debug", false, "If set, debug logging will be enabled")
	wikiLinks = tagsCommand.Flags().Bool("wiki-links", false, "If set, links will be generated for a wiki with file extensions excluded")
	capitalize = tagsCommand.Flags().Bool("capitalize", false, "If set, tag names in the generated index will have their first character capitalized.")
	tagLinks = tagsCommand.Flags().Bool("tag-links", false, "If set, links to files in the generated index will be annotated with the list of other tags they have.")
	toc = tagsCommand.Flags().Bool("toc", false, "If set, a table of contents will be generated containing a link to the heading of each tag")
	rootCmd := &cobra.Command{
		Use: "markasten",
	}
	rootCmd.AddCommand(tagsCommand)
	return rootCmd

}

func debug(format string, v ...any) {
	if *debugEnabled {
		log.Printf(format, v...)
	}
}

func main() {
	rootCmd := newRootCmd()
	rootCmd.Execute()
}

type indexedFile struct {
	fileName  string
	title     string
	otherTags []string
}

func writeOrPanic(file *os.File, text string) {
	_, err := io.WriteString(file, text)
	if err != nil {
		panic(err)
	}
}

func tagsRunFn(cmd *cobra.Command, args []string) error {
	debug("tags called with -i %s and -o %s\n", *inputPath, *outputPath)
	inputDirEntires, err := newFullDirEntryList(*inputPath)
	if err != nil {
		panic(err)
	}

	filesByTags := make(map[string][]indexedFile)
	searchResults, err := searchForMarkdownFiles(inputDirEntires, *inputPath)
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

	outputFile, err := os.Create(*outputPath)
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
			relativePath := relativeTo(f.fileName, *outputPath)
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
func relativeTo(filePath string, relativeToPath string) string {
	dir := filepath.Dir(relativeToPath)
	relative, err := filepath.Rel(dir, filePath)
	if err != nil {
		return filePath
	}
	return relative
}

func firstNonEmptyLine(lines []string) string {
	for _, line := range lines {
		if len(line) > 0 {
			return line
		}
	}
	return ""
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

type frontmatter struct {
	Tags []string
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

type fullDirEntry struct {
	dirEntry   fs.DirEntry
	parentPath string
}

func (d fullDirEntry) IsDir() bool {
	return d.dirEntry.IsDir()
}

func (d fullDirEntry) Name() string {
	return filepath.Join(d.parentPath, d.dirEntry.Name())
}

func (d fullDirEntry) IsDotFile() bool {
	return len(d.dirEntry.Name()) > 0 && d.dirEntry.Name()[0:1] == "."
}

func newFullDirEntryList(root string) ([]fullDirEntry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var fullDirEntries []fullDirEntry
	for _, entry := range entries {
		fullDirEntries = append(fullDirEntries, fullDirEntry{
			dirEntry:   entry,
			parentPath: root,
		})
	}
	return fullDirEntries, nil
}

func searchForMarkdownFiles(dirEntries []fullDirEntry, root string) ([]fullDirEntry, error) {
	debug("searching %s", root)
	var entries []fullDirEntry
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDotFile() {
			continue
		}
		if dirEntry.IsDir() {
			debug("found sub directory %s", dirEntry.dirEntry.Name())
			subEntries, err := newFullDirEntryList(dirEntry.Name())
			if err != nil {
				return nil, err
			}
			searchResults, err := searchForMarkdownFiles(subEntries, dirEntry.Name())
			if err != nil {
				return nil, err
			}
			for _, subEntry := range searchResults {
				entries = append(entries, subEntry)
			}
		} else {
			debug("found file %s", dirEntry.Name())
			entries = append(entries, dirEntry)
		}
	}
	return entries, nil
}

func countTitles(files []indexedFile) map[string]int {
	titleCounts := make(map[string]int)
	for _, file := range files {
		titleCounts[file.title]++
	}
	return titleCounts

}

func makeWikiLink(path string) string {
	return path[:len(path)-len(filepath.Ext(path))]
}

func headerToLink(header string) string {
	// Some special characters produce invalid heading links in GitHub.
	// In this case, a heading of "foo:bar" requires a link of "foobar",
	// and a heading of "foo bar" requires a link of "foo-bar".
	return strings.ReplaceAll(strings.ReplaceAll(header, ":", ""), " ", "-")
}
