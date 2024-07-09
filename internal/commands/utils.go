package commands

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	re           = regexp.MustCompile("`[^`]+`")
	debugEnabled *bool
)

type frontmatter struct {
	Tags []string
}

func debug(format string, v ...any) {
	if *debugEnabled {
		log.Printf(format, v...)
	} else {
		log.Printf("debug logging is disabled")
	}
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
