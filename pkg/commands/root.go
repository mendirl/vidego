package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() error {
	var rootCmd = &cobra.Command{
		Use: "dedup-cli",
	}

	commands := []*cobra.Command{
		newDedupCommand(),
		newFilterCommand(),
		newPutbackCommand(),
		newMoveCommand(),
		newDeleteCommand(),
		newSortCommand(),
	}

	for command := range commands {
		rootCmd.AddCommand(commands[command])
	}

	return rootCmd.Execute()
}

func initConfig(cfgFile string) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("$HOME")
		viper.SetConfigName(".vidego")
	}

	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s", err)
	}
}
