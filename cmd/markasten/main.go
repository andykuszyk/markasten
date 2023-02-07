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
)

var (
	inputPath    *string
	outputPath   *string
	title        *string
	debugEnabled *bool
	wikiLinks    *bool
	capitalize   *bool
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
	fileName string
	title    string
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
		header := tag
		if *capitalize && len(tag) > 0 {
			header = fmt.Sprintf("%s%s", strings.ToUpper(tag[0:1]), tag[1:])
		}
		_, err = io.WriteString(outputFile, fmt.Sprintf("## %s\n", header))
		if err != nil {
			panic(err)
		}
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
			_, err = io.WriteString(
				outputFile,
				fmt.Sprintf(
					"- [%s](%s)%s",
					title,
					relativePath,
					trailingChar,
				),
			)
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
	return scrapedTags, title
}

func appendFilesByTags(scrapedTags []string, filesByTags map[string][]indexedFile, title string, name string) map[string][]indexedFile {
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
	return filesByTags
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
