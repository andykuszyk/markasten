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
		basicTagsExtraLineBreaks(),
		tagsWithExtraSpacing(),
		tagsWithSpacesNumbersAndSpecialCharacters(),
		tagsWithUnclosedTag(),
		tagsWithCustomTitle(),
		tagsWithFilesInSubDirectories(),
		// TODO: files in sub directories with same names/titles
		// TODO: files in nested subdirectories, e.g. foo/bar/spam.md
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputDir := writeFiles(t, tc.inputFiles, "markasten-input")
			expectedOutputDir := writeFiles(t, tc.outputFiles, "markasten-expected-output")
			expectedOutputFilePath := filepath.Join(expectedOutputDir, tc.outputFiles[0].name)
			actualOutputDir, err := os.MkdirTemp("", "markasten-actual-output")
			require.NoError(t, err)
			actualOutputFilePath := filepath.Join(actualOutputDir, tc.outputFiles[0].name)

			rootCmd := newRootCmd()
			args := []string{
				"tags",
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
					"`foo` `spam`",
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
					"`bar` `eggs` `spam`",
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
					"`foo` `spam`",
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
					"`bar` `eggs` `spam`",
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

func tagsWithExtraSpacing() testCase {
	return testCase{
		name: "tags with extra spacing",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					"   `foo` `spam`  ",
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
					"`bar`   `eggs`   `spam`",
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

func tagsWithSpacesNumbersAndSpecialCharacters() testCase {
	return testCase{
		name: "tags with spaces, numbers, and special characters",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					" `foo bar` `spam-eggs` `green_h@m` ",
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

func tagsWithUnclosedTag() testCase {
	return testCase{
		name: "tags with unclosed tag",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"---",
					" `foo` `bar",
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
					"## foo",
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
					"`foo`",
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
					"`foo` `bar` `spam`",
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
					"`foo` `bar` `spam`",
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
					"`foo` `bar` `spam`",
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
