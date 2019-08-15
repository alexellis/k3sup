package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func MakeVersion() *cobra.Command {
	var command = &cobra.Command{
		Use:          "version",
		Short:        "Print the version",
		Example:      `  k3sup version`,
		SilenceUsage: false,
	}
	command.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Welcome to k3sup!\n")
	}
	return command
}
