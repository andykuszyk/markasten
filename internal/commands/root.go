package commands

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "markasten",
	}
	rootCmd.AddCommand(newTagsCommand())
	rootCmd.AddCommand(newBacklinksCommand())
	return rootCmd

}
