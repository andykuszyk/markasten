package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type file struct {
	name     string
	contents []string
}

type testCase struct {
	name           string
	additionalArgs []string
	inputFiles     []file
	outputFiles    []file
}

func TestTags(t *testing.T) {
	for _, tc := range []testCase{
		basicTags(),
		basicTagsWithTagLinks(),
		basicTagsWithCapitaliseOption(),
		basicTagsWithWikiLinks(),
		basicTagsExtraLineBreaks(),
		tagsWithSpacesNumbersAndSpecialCharacters(),
		tagsWithCustomTitle(),
		tagsWithFilesInSubDirectories(),
		tagsWithFilesInNestedSubDirectories(),
		tagsWithFilesInSubDirectoriesWithSameNames(),
		tagsWithFilesInDotDirectory(),
		fileWithNoTagsAndBacktickedText(),
		tocFlag(),
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputDir := writeFiles(t, tc.inputFiles, "markasten-input")
			expectedOutputDir := writeFiles(t, tc.outputFiles, "markasten-expected-output")
			expectedOutputFilePath := filepath.Join(expectedOutputDir, tc.outputFiles[0].name)
			actualOutputFilePath := filepath.Join(inputDir, tc.outputFiles[0].name)

			rootCmd := newRootCmd()
			args := []string{
				"tags",
				"--debug",
				"-i",
				inputDir,
				"-o",
				actualOutputFilePath,
			}
			if len(tc.additionalArgs) > 0 {
				for _, a := range tc.additionalArgs {
					args = append(args, a)
				}
			}
			rootCmd.SetArgs(args)
			rootCmd.Execute()

			expectedOutputBytes, err := os.ReadFile(expectedOutputFilePath)
			require.NoError(t, err)

			actualOutputBytes, err := os.ReadFile(actualOutputFilePath)
			require.NoError(t, err)

			require.Equal(t, string(expectedOutputBytes), string(actualOutputBytes))
		})
	}
}

func writeFiles(t *testing.T, files []file, directoryName string) string {
	dir, err := os.MkdirTemp("", directoryName)
	require.NoError(t, err)
	for _, file := range files {
		subDir := filepath.Join(dir, filepath.Dir(file.name))
		fmt.Printf("creating directory %s\n", subDir)
		require.NoError(t, os.MkdirAll(subDir, 0700))
		fmt.Printf("writing file %s\n", filepath.Join(dir, file.name))
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, file.name),
			[]byte(strings.Join(file.contents, "\n")),
			0600,
		))
	}
	return dir
}

func basicTags() testCase {
	return testCase{
		name: "basic tags",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"---",
					"tags:",
					"- bar",
					"- eggs",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## bar",
					"- [Bar](bar.md)",
					"",
					"## eggs",
					"- [Bar](bar.md)",
					"",
					"## foo",
					"- [Foo](foo.md)",
					"",
					"## spam",
					"- [Bar](bar.md)",
					"- [Foo](foo.md)",
				},
			},
		},
	}
}

func basicTagsExtraLineBreaks() testCase {
	return testCase{
		name: "basic tags with extra line breaks",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"",
					"---",
					"tags:",
					"- foo",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"---",
					"tags:",
					"- bar",
					"- eggs",
					"- spam",
					"---",
					"",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## bar",
					"- [Bar](bar.md)",
					"",
					"## eggs",
					"- [Bar](bar.md)",
					"",
					"## foo",
					"- [Foo](foo.md)",
					"",
					"## spam",
					"- [Bar](bar.md)",
					"- [Foo](foo.md)",
				},
			},
		},
	}
}

func tagsWithSpacesNumbersAndSpecialCharacters() testCase {
	return testCase{
		name: "tags with spaces, numbers, and special characters",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo bar",
					"- spam-eggs",
					"- green_h@m",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## foo bar",
					"- [Foo](foo.md)",
					"",
					"## green_h@m",
					"- [Foo](foo.md)",
					"",
					"## spam-eggs",
					"- [Foo](foo.md)",
				},
			},
		},
	}
}

func tagsWithCustomTitle() testCase {
	return testCase{
		name:           "tags with custom title",
		additionalArgs: []string{"-t", "README"},
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# README",
					"## foo",
					"- [Foo](foo.md)",
				},
			},
		},
	}
}

func tagsWithFilesInSubDirectories() testCase {
	return testCase{
		name: "tags with files in sub directories",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar/bar.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
			{
				name: "spam/spam.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Spam",
					"Spam is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## bar",
					"- [Bar](bar/bar.md)",
					"- [Foo](foo.md)",
					"- [Spam](spam/spam.md)",
					"",
					"## foo",
					"- [Bar](bar/bar.md)",
					"- [Foo](foo.md)",
					"- [Spam](spam/spam.md)",
					"",
					"## spam",
					"- [Bar](bar/bar.md)",
					"- [Foo](foo.md)",
					"- [Spam](spam/spam.md)",
				},
			},
		},
	}
}

func tagsWithFilesInNestedSubDirectories() testCase {
	return testCase{
		name: "tags with files in nested sub directories",
		inputFiles: []file{
			{
				name: "eggs/foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar/bar.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
			{
				name: "eggs/spam/spam.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Spam",
					"Spam is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## bar",
					"- [Bar](bar/bar.md)",
					"- [Foo](eggs/foo.md)",
					"- [Spam](eggs/spam/spam.md)",
					"",
					"## foo",
					"- [Bar](bar/bar.md)",
					"- [Foo](eggs/foo.md)",
					"- [Spam](eggs/spam/spam.md)",
					"",
					"## spam",
					"- [Bar](bar/bar.md)",
					"- [Foo](eggs/foo.md)",
					"- [Spam](eggs/spam/spam.md)",
				},
			},
		},
	}
}

func tagsWithFilesInSubDirectoriesWithSameNames() testCase {
	return testCase{
		name: "tags with files in sub directories with same names",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar/bar.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
			{
				name: "spam/eggs.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Bar",
					"Spam is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## bar",
					"- [bar/bar.md](bar/bar.md)",
					"- [Foo](foo.md)",
					"- [spam/eggs.md](spam/eggs.md)",
					"",
					"## foo",
					"- [bar/bar.md](bar/bar.md)",
					"- [Foo](foo.md)",
					"- [spam/eggs.md](spam/eggs.md)",
					"",
					"## spam",
					"- [bar/bar.md](bar/bar.md)",
					"- [Foo](foo.md)",
					"- [spam/eggs.md](spam/eggs.md)",
				},
			},
		},
	}
}

func basicTagsWithWikiLinks() testCase {
	return testCase{
		name: "basic tags with wiki links",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"---",
					"tags:",
					"- bar",
					"- eggs",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
		},
		additionalArgs: []string{"--wiki-links"},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## bar",
					"- [Bar](bar)",
					"",
					"## eggs",
					"- [Bar](bar)",
					"",
					"## foo",
					"- [Foo](foo)",
					"",
					"## spam",
					"- [Bar](bar)",
					"- [Foo](foo)",
				},
			},
		},
	}
}

func tagsWithFilesInDotDirectory() testCase {
	return testCase{
		name: "tags with files in sub directories",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: ".bar/bar.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
			{
				name: "spam/spam.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- bar",
					"- spam",
					"---",
					"",
					"# Spam",
					"Spam is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## bar",
					"- [Foo](foo.md)",
					"- [Spam](spam/spam.md)",
					"",
					"## foo",
					"- [Foo](foo.md)",
					"- [Spam](spam/spam.md)",
					"",
					"## spam",
					"- [Foo](foo.md)",
					"- [Spam](spam/spam.md)",
				},
			},
		},
	}
}

func basicTagsWithCapitaliseOption() testCase {
	return testCase{
		name:           "basic tags with capitalise option",
		additionalArgs: []string{"--capitalize"},
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"---",
					"tags:",
					"- bar",
					"- eggs",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## Bar",
					"- [Bar](bar.md)",
					"",
					"## Eggs",
					"- [Bar](bar.md)",
					"",
					"## Foo",
					"- [Foo](foo.md)",
					"",
					"## Spam",
					"- [Bar](bar.md)",
					"- [Foo](foo.md)",
				},
			},
		},
	}
}

func fileWithNoTagsAndBacktickedText() testCase {
	return testCase{
		name: "file with no tags and backticked text",
		inputFiles: []file{
			{
				name: "home.md",
				contents: []string{
					"Foo is `about` something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"# Bar",
					"Bar is about `something`, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"",
				},
			},
		},
	}
}

func basicTagsWithTagLinks() testCase {
	return testCase{
		name:           "basic tags with tag links",
		additionalArgs: []string{"--tag-links"},
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"---",
					"tags:",
					"- bar",
					"- eggs",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"## bar",
					"- [Bar](bar.md) `eggs` `spam`",
					"",
					"## eggs",
					"- [Bar](bar.md) `bar` `spam`",
					"",
					"## foo",
					"- [Foo](foo.md) `spam`",
					"",
					"## spam",
					"- [Bar](bar.md) `bar` `eggs`",
					"- [Foo](foo.md) `foo`",
				},
			},
		},
	}
}

func tocFlag() testCase {
	return testCase{
		name:           "table of contents",
		additionalArgs: []string{"--toc"},
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"tags:",
					"- foo",
					"- spam",
					"---",
					"",
					"# Foo",
					"Foo is about something, similar to [bar](./bar.md).",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"---",
					"tags:",
					"- bar",
					"- eggs",
					"- spam",
					"---",
					"",
					"# Bar",
					"Bar is about something, similar to [foo](./foo.md).",
				},
			},
		},
		outputFiles: []file{
			{
				name: "index.md",
				contents: []string{
					"# Index",
					"",
					"---",
					"",
					"## Table of contents",
					"- [bar](#bar)",
					"- [eggs](#eggs)",
					"- [foo](#foo)",
					"- [spam](#spam)",
					"",
					"---",
					"",
					"## bar",
					"- [Bar](bar.md)",
					"",
					"## eggs",
					"- [Bar](bar.md)",
					"",
					"## foo",
					"- [Foo](foo.md)",
					"",
					"## spam",
					"- [Bar](bar.md)",
					"- [Foo](foo.md)",
				},
			},
		},
	}
}
