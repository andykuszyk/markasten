package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestTags(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println(wd)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"tags", "-i", "../../testdata/input/", "-o", "../../index.md"})
	rootCmd.Execute()

	outputBytes, err := ioutil.ReadFile("../../index.md")
	if err != nil {
		panic(err)
	}
	inputBytes, err := ioutil.ReadFile("../../testdata/outputs/tags-index/index.md")
	if err != nil {
		panic(err)
	}
	if string(outputBytes) != string(inputBytes) {
		panic("input and output don't match")
	}
}
