package main

import (
	"github.com/alexellis/k3sup/pkg/cmd"
	"github.com/spf13/cobra"
)

func main() {

	cmdInstall := cmd.MakeInstall()

	cmdVersion := cmd.MakeVersion()

	cmdJoin := cmd.MakeJoin()

	printK3sASCIIArt := cmd.PrintK3supASCIIArt

	var rootCmd = &cobra.Command{
		Use: "k3sup",
		Run: func(cmd *cobra.Command, args []string) {
			printK3sASCIIArt()
			cmd.Help()
		},
	}

	rootCmd.AddCommand(cmdInstall)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(cmdJoin)

	rootCmd.Execute()
}
