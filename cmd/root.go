package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "cm",
		Short: "Config Manager",
	}
)

const (
	moduleApi                = "api"
	moduleDispatcherConsumer = "dispatcher-consumer"
	moduleInventoryConsumer  = "inventory-consumer"
)

func init() {
	runCommand := &cobra.Command{
		Use:   "run",
		Short: "Run config manager",
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				os.Exit(1)
			}
		},
	}

	runCommand.Flags().StringSliceP("module", "m", []string{moduleApi, moduleDispatcherConsumer, moduleInventoryConsumer}, "module(s) to run")
	rootCmd.AddCommand(runCommand)
}

func Execute() error {
	return rootCmd.Execute()
}
