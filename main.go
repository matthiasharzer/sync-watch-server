package main

import (
	"fmt"
	"os"

	"github.com/matthiasharzer/sync-watch-server/cmd/run"
	"github.com/matthiasharzer/sync-watch-server/cmd/version"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use: "sync-watch-server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCommand.AddCommand(version.Command)
	rootCommand.AddCommand(run.Command)
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
