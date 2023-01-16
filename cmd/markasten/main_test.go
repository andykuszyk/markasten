package main

import (
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
	name        string
	inputFiles  []file
	outputFiles []file
}

func TestTags(t *testing.T) {
	for _, tc := range []testCase{
		{
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
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputDir := writeFiles(t, tc.inputFiles, "markasten-input")
			expectedOutputDir := writeFiles(t, tc.outputFiles, "markasten-expected-output")
			expectedOutputFilePath := filepath.Join(expectedOutputDir, tc.outputFiles[0].name)
			actualOutputDir, err := os.MkdirTemp("", "markasten-actual-output")
			require.NoError(t, err)
			actualOutputFilePath := filepath.Join(actualOutputDir, tc.outputFiles[0].name)

			rootCmd := newRootCmd()
			rootCmd.SetArgs([]string{
				"tags",
				"-i",
				inputDir,
				"-o",
				actualOutputFilePath,
			})
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
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, file.name),
			[]byte(strings.Join(file.contents, "\n")),
			0600,
		))
	}
	return dir
}
