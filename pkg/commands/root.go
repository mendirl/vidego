package commands

import (
	"github.com/spf13/cobra"
)

func Execute() error {
	var rootCmd = &cobra.Command{
		Use: "dedup-cli",
	}

	commands := []*cobra.Command{
		newPersistCommand(),
		newDedupCommand(),
		newFilterCommand(),
		newPutbackCommand(),
	}

	for command := range commands {
		rootCmd.AddCommand(commands[command])
	}

	return rootCmd.Execute()
}
