package main

import (
	"os"

	"github.com/alexellis/k3sup/cmd"
	"github.com/spf13/cobra"
)

func main() {

	cmdInstall := cmd.MakeInstall()
	cmdVersion := cmd.MakeVersion()
	cmdJoin := cmd.MakeJoin()
	cmdUpdate := cmd.MakeUpdate()

	printk3supASCIIArt := cmd.PrintK3supASCIIArt

	var rootCmd = &cobra.Command{
		Use: "k3sup",
		Run: func(cmd *cobra.Command, args []string) {
			printk3supASCIIArt()
			cmd.Help()
		},
	}

	rootCmd.AddCommand(cmdInstall)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(cmdJoin)
	rootCmd.AddCommand(cmdUpdate)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
