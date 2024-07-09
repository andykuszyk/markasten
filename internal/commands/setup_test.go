package commands_test

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

func TestMain(m *testing.M) {
	m.Run()
}
