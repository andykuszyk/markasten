package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andykuszyk/markasten/internal/commands"

	"github.com/stretchr/testify/require"
)

func TestBacklinksFind(t *testing.T) {
	for _, tc := range []testCase{
		basicBacklinksFind(),
		basicBacklinksFindWithMultipleFiles(),
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputDir := writeFiles(t, tc.inputFiles, "markasten-input")
			expectedOutputDir := writeFiles(t, tc.outputFiles, "markasten-expected-output")
			expectedOutputFilePath := filepath.Join(expectedOutputDir, tc.outputFiles[0].name)
			actualOutputFilePath := filepath.Join(inputDir, tc.outputFiles[0].name)

			rootCmd := commands.NewRootCmd()
			args := []string{
				"backlinks",
				"find",
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

type linkRegexpTestCase struct {
	input   string
	matches []string
}

func TestLinkRegexp(t *testing.T) {
	for _, tc := range []linkRegexpTestCase{
		{
			input:   "foo bar spam",
			matches: []string{},
		},
		{
			input:   "foo [bar](spam)",
			matches: []string{"[bar](spam)"},
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			matches := commands.LinkRegexp.FindAllString(tc.input, -1)
			require.ElementsMatch(t, tc.matches, matches)
		})
	}
}

func basicBacklinksFind() testCase {
	return testCase{
		name: "basic backlinks find",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"# Foo",
					"Foo mentions [bar](./bar.md)",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"# Bar",
					"Bar is mentioned by foo.",
				},
			},
		},
		outputFiles: []file{
			{
				name: "backlinks.yml",
				contents: []string{
					"foo.md:",
					"  - bar.md",
				},
			},
		},
	}
}

func basicBacklinksFindWithMultipleFiles() testCase {
	return testCase{
		name: "backlinks find with multiple files",
		inputFiles: []file{
			{
				name: "foo.md",
				contents: []string{
					"# Foo",
					"Foo mentions [bar](./bar.md)",
				},
			},
			{
				name: "spam.md",
				contents: []string{
					"# Spam",
					"Spam mentions [bar](bar.md)",
				},
			},
			{
				name: "bar.md",
				contents: []string{
					"# Bar",
					"Bar is mentioned by foo.",
				},
			},
		},
		outputFiles: []file{
			{
				name: "backlinks.yml",
				contents: []string{
					"foo.md:",
					"  - bar.md",
					"spam.md:",
					"  - bar.md",
				},
			},
		},
	}
}
