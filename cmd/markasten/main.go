package main

import (
	"github.com/andykuszyk/markasten/internal/commands"
)

func main() {
	rootCmd := commands.NewRootCmd()
	rootCmd.Execute()
}
