package main

import (
	"testing"
)

func TestTags(t *testing.T) {
	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"tags", "-i", "foo", "-o", "bar"})
	rootCmd.Execute()
}
